package no_useless_assignment

// cspell:ignore worklist

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"

	"github.com/web-infra-dev/rslint/internal/utils/cfg"
)

// Reads and writes are recorded in evaluation order. A nil assignment denotes
// a read; writes point to their syntactic assignment so every copy of a finally
// block can keep the same assignment alive.
type event struct {
	variable   int
	assignment *assignment
}

type block = cfg.Block[event]

type assignment struct {
	variable   int
	identifier *ast.Node
	// A reachable write starts dead. Reading its value at any CFG site clears
	// dead, including when a finally block has multiple completion paths.
	dead bool
	// A try-block write still overwrites the variable but is never reported.
	silent bool
}

// hooks records only references to tracked variables in reachable blocks.
func hooks(readNodes map[*ast.Node]int, assignByIdent map[*ast.Node]*assignment) cfg.Hooks[event] {
	return cfg.Hooks[event]{
		Read: func(b *cfg.Builder[event], node *ast.Node) {
			if !b.Current().Reachable {
				return
			}
			if variable, ok := readNodes[node]; ok {
				b.Emit(event{variable: variable})
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
			a.dead = true
			b.Emit(event{variable: a.variable, assignment: a})
		},
	}
}

// markDeadWrites solves backward liveness for all variables together. Each
// variable occupies one bit, avoiding a scan of every event for every symbol.
func markDeadWrites(graph *cfg.Graph[event], variableCount int) {
	words := (variableCount + 63) / 64
	blocks := graph.Blocks
	if len(blocks) == 1 && len(blocks[0].Successors) == 0 {
		markBlockDead(blocks[0], make([]uint64, words))
		return
	}

	size := len(blocks) * words
	storage := make([]uint64, 3*size)
	def, liveIn, liveOut := storage[:size], storage[size:2*size], storage[2*size:]
	offsets, predecessors, queue := buildPredecessors(blocks)
	queued := make([]bool, len(blocks))
	for i, blk := range blocks {
		start := i * words
		// Starting with no live-out values leaves exactly the reads that occur
		// before the first write of their variable (the block's use set).
		for j := len(blk.Events) - 1; j >= 0; j-- {
			e := blk.Events[j]
			word, bit := start+e.variable/64, uint64(1)<<(e.variable%64)
			if e.assignment == nil {
				liveIn[word] |= bit
			} else {
				def[word] |= bit
				liveIn[word] &^= bit
			}
		}
		for _, bits := range liveIn[start : start+words] {
			if bits != 0 {
				queue = append(queue, i)
				queued[i] = true
				break
			}
		}
	}

	// A block's live-in set only grows. Propagate it to predecessors, retaining
	// each predecessor's use set and killing values it defines. A block can be
	// queued again after it is popped, which is necessary for loop back edges.
	for len(queue) != 0 {
		index := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		queued[index] = false
		for _, predecessor := range predecessors[offsets[index]:offsets[index+1]] {
			changed := false
			for word, bits := range liveIn[index*words : (index+1)*words] {
				at := predecessor*words + word
				liveOut[at] |= bits
				incoming := liveIn[at] | (bits &^ def[at])
				if incoming != liveIn[at] {
					liveIn[at] = incoming
					changed = true
				}
			}
			if changed && !queued[predecessor] {
				queue = append(queue, predecessor)
				queued[predecessor] = true
			}
		}
	}

	for i, blk := range blocks {
		markBlockDead(blk, liveOut[i*words:(i+1)*words])
	}
}

// markBlockDead visits each event once instead of searching forward from every
// assignment. live starts at the block's exit and is consumed as scratch.
func markBlockDead(blk *block, live []uint64) {
	for i := len(blk.Events) - 1; i >= 0; i-- {
		e := blk.Events[i]
		word, bit := e.variable/64, uint64(1)<<(e.variable%64)
		if e.assignment == nil {
			live[word] |= bit
		} else {
			if live[word]&bit != 0 {
				e.assignment.dead = false
			}
			live[word] &^= bit
		}
	}
}

// buildPredecessors stores reachable successor edges in reverse as a compact
// adjacency list. scratch is reused as the worklist, with at most one pending
// entry per block.
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
