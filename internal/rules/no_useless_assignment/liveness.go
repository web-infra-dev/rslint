package no_useless_assignment

import (
	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/cfg"
)

// This file answers "can the value written here still be read?" over the
// control-flow graph internal/cfg lays out. Reads and writes of the tracked
// variables are recorded as events in evaluation order, so liveness is a plain
// backward dataflow over the graph.

type eventKind uint8

const (
	eventRead eventKind = iota
	eventWrite
)

type event struct {
	kind   eventKind
	sym    *ast.Symbol
	assign *assignment
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
	sites      []writeSite
	// silent marks an assignment that still overwrites the variable but is
	// never reported, because it sits in a `try` block.
	silent bool
}

// hooks turns the graph builder's reference positions into events. An
// identifier absent from both maps belongs to a variable this code path does
// not track.
func hooks(readNodes map[*ast.Node]*ast.Symbol, assignByIdent map[*ast.Node]*assignment) cfg.Hooks[event] {
	return cfg.Hooks[event]{
		Read: func(b *cfg.Builder[event], node *ast.Node) {
			if sym, ok := readNodes[node]; ok {
				b.Emit(event{kind: eventRead, sym: sym})
			}
		},
		Write: func(b *cfg.Builder[event], node *ast.Node) {
			a, ok := assignByIdent[node]
			if !ok {
				return
			}
			if blk, index := b.Emit(event{kind: eventWrite, sym: a.sym, assign: a}); blk != nil {
				a.sites = append(a.sites, writeSite{blk: blk, index: index})
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

// deadWrites reports, per assignment, whether every site of that assignment is
// dead — no read of the variable is reachable before the value is overwritten.
func deadWrites(graph *cfg.Graph[event], assignments []*assignment) map[*assignment]bool {
	bySymbol := make(map[*ast.Symbol][]*assignment)
	for _, a := range assignments {
		bySymbol[a.sym] = append(bySymbol[a.sym], a)
	}

	states := make([]blockState, len(graph.Blocks))
	dead := make(map[*assignment]bool, len(assignments))
	for sym, symAssignments := range bySymbol {
		computeLiveness(graph.Blocks, states, sym)
		for _, a := range symAssignments {
			if len(a.sites) == 0 {
				continue
			}
			isDead := true
			for _, site := range a.sites {
				if liveAfter(sym, site, states) {
					isDead = false
					break
				}
			}
			dead[a] = isDead
		}
	}
	return dead
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
