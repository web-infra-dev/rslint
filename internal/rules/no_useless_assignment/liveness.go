package no_useless_assignment

// cspell:ignore worklist

import (
	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/utils/cfg"
)

// This file answers "can the value written here still be read?" over the
// control-flow graph internal/utils/cfg lays out. Reads and writes of the
// tracked variables are recorded as events in evaluation order, so liveness is
// a plain backward dataflow over the graph.

type eventKind uint8

const (
	eventRead eventKind = iota
	eventWrite
)

type event struct {
	kind eventKind
	sym  *ast.Symbol
}

type block = cfg.Block[event]

// writeSite is one place in the graph where an assignment appears. An
// assignment inside a `finally` block appears twice — once per copy of the
// block — and only counts as unused when every site is dead.
type writeSite struct {
	blk   *block
	index int
}

// assignment is one reportable write: the identifier ESLint reports on plus the
// variable it targets.
type assignment struct {
	sym        *ast.Symbol
	identifier *ast.Node
	// Most writes occur at exactly one CFG site. A finally block is copied for
	// each completion path, so retain only those uncommon additional sites in a
	// slice instead of allocating one for every assignment.
	site            writeSite
	additionalSites []writeSite
	dead            bool
	// silent marks an assignment that still overwrites the variable but is
	// never reported, because it sits in a `try` block.
	silent bool
}

// hooks turns the graph builder's reference positions into events. An
// identifier absent from both maps belongs to a variable this code path does
// not track, and a reference in an unreachable block belongs to code that never
// runs — neither says anything about the value an assignment leaves behind.
func hooks(readNodes map[*ast.Node]*ast.Symbol, assignByIdent map[*ast.Node]*assignment) cfg.Hooks[event] {
	return cfg.Hooks[event]{
		Read: func(b *cfg.Builder[event], node *ast.Node) {
			if !b.Current().Reachable {
				return
			}
			if sym, ok := readNodes[node]; ok {
				b.Emit(event{kind: eventRead, sym: sym})
			}
		},
		Write: func(b *cfg.Builder[event], node *ast.Node) {
			if !b.Current().Reachable {
				return
			}
			a, ok := assignByIdent[node]
			if !ok {
				return
			}
			blk, index := b.Emit(event{kind: eventWrite, sym: a.sym})
			site := writeSite{blk: blk, index: index}
			if a.site.blk == nil {
				a.site = site
			} else {
				a.additionalSites = append(a.additionalSites, site)
			}
		},
	}
}

// blockState is one block's scratch during the liveness dataflow of a single
// variable.
type blockState struct {
	use     bool
	def     bool
	liveIn  bool
	liveOut bool
}

// markDeadWrites records, per assignment, whether every site of that assignment
// is dead — no read of the variable is reachable before the value is
// overwritten.
func markDeadWrites(graph *cfg.Graph[event], assignments []*assignment) {
	bySymbol := make(map[*ast.Symbol][]*assignment)
	for _, a := range assignments {
		bySymbol[a.sym] = append(bySymbol[a.sym], a)
	}

	states := make([]blockState, len(graph.Blocks))
	if len(bySymbol) == 1 {
		for sym, symAssignments := range bySymbol {
			computeLiveness(graph.Blocks, states, sym)
			markSymbolDead(symAssignments, states)
		}
	} else {
		offsets, predecessors, queue := buildPredecessors(graph.Blocks)
		for sym, symAssignments := range bySymbol {
			computeLivenessWithWorklist(graph.Blocks, states, sym, offsets, predecessors, queue)
			markSymbolDead(symAssignments, states)
		}
	}
}

func markSymbolDead(assignments []*assignment, states []blockState) {
	for _, a := range assignments {
		if a.site.blk == nil {
			continue
		}
		isDead := !liveAfter(a.sym, a.site, states)
		for _, site := range a.additionalSites {
			if liveAfter(a.sym, site, states) {
				isDead = false
				break
			}
		}
		a.dead = isDead
	}
}

// buildPredecessors stores the reachable successor edges in reverse as a
// compact adjacency list. scratch is reused as the per-symbol work queue.
func buildPredecessors(blocks []*block) (offsets []int, predecessors []int, scratch []int) {
	offsets = make([]int, len(blocks)+1)
	for _, blk := range blocks {
		for _, successor := range blk.Successors {
			if successor.Reachable {
				offsets[successor.Index()+1]++
			}
		}
	}
	for i := 1; i < len(offsets); i++ {
		offsets[i] += offsets[i-1]
	}

	predecessors = make([]int, offsets[len(blocks)])
	scratch = make([]int, len(blocks))
	copy(scratch, offsets[:len(blocks)])
	for predecessor, blk := range blocks {
		for _, successor := range blk.Successors {
			if !successor.Reachable {
				continue
			}
			index := successor.Index()
			predecessors[scratch[index]] = predecessor
			scratch[index]++
		}
	}
	return offsets, predecessors, scratch[:0]
}

// computeLivenessWithWorklist propagates liveness only through blocks whose
// live-in state becomes true. This is equivalent to computeLiveness's monotone
// fixed point, but avoids rescanning the whole graph for every tracked symbol.
func computeLivenessWithWorklist(
	blocks []*block,
	states []blockState,
	sym *ast.Symbol,
	offsets []int,
	predecessors []int,
	queue []int,
) {
	queue = queue[:0]
	for i, blk := range blocks {
		var state blockState
		for _, e := range blk.Events {
			if e.sym != sym {
				continue
			}
			if e.kind == eventRead {
				state.use = true
			} else {
				state.def = true
			}
			break
		}
		state.liveIn = state.use
		states[i] = state
		if state.liveIn {
			queue = append(queue, i)
		}
	}

	for head := 0; head < len(queue); head++ {
		blockIndex := queue[head]
		for _, predecessor := range predecessors[offsets[blockIndex]:offsets[blockIndex+1]] {
			state := &states[predecessor]
			state.liveOut = true
			if !state.liveIn && !state.def {
				state.liveIn = true
				queue = append(queue, predecessor)
			}
		}
	}
}

// computeLiveness runs a backward liveness dataflow for one variable, filling
// states in step with blocks.
func computeLiveness(blocks []*block, states []blockState, sym *ast.Symbol) {
	for i, blk := range blocks {
		var state blockState
		for _, e := range blk.Events {
			if e.sym != sym {
				continue
			}
			if e.kind == eventRead {
				if !state.def {
					state.use = true
				}
			} else if !state.use {
				state.def = true
			}
			if state.use || state.def {
				break
			}
		}
		state.liveIn = state.use
		states[i] = state
	}

	for changed := true; changed; {
		changed = false
		for i := len(blocks) - 1; i >= 0; i-- {
			state := &states[i]
			liveOut := false
			for _, successor := range blocks[i].Successors {
				if !successor.Reachable {
					// The code after an abrupt exit never runs, so a read in
					// it keeps nothing alive.
					continue
				}
				if states[successor.Index()].liveIn {
					liveOut = true
					break
				}
			}
			liveIn := state.use || (!state.def && liveOut)
			if liveOut != state.liveOut || liveIn != state.liveIn {
				state.liveOut = liveOut
				state.liveIn = liveIn
				changed = true
			}
		}
	}
}

// liveAfter reports whether the value written at site can still be read.
func liveAfter(sym *ast.Symbol, site writeSite, states []blockState) bool {
	for _, e := range site.blk.Events[site.index+1:] {
		if e.sym != sym {
			continue
		}
		return e.kind == eventRead
	}
	return states[site.blk.Index()].liveOut
}
