package cfg

// cspell:ignore worklist

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// VariableWrite identifies an assignment to a caller-resolved variable.
type VariableWrite struct {
	Node     *ast.Node
	Variable int
}

type variableEvent struct {
	variable int
	write    *bool
}

type variableBlock = Block[variableEvent]

// DeadWrites returns the reachable assignments whose values are never read
// before being overwritten or leaving root. Reads and writes associate identifier
// nodes with caller-resolved variable indices in [0, variableCount). References
// must belong to root; binding, export and capture policy stays with the caller.
// Repeated input nodes and CFG copies of a write share one result; unreachable
// writes are excluded. The result order is unspecified. Inputs and compiler
// data are not modified.
//
// Restricted single-variable queries reuse binder flow. Other queries use the
// shared CFG's evaluation order and batched backward liveness.
func DeadWrites(root *ast.Node, reads map[*ast.Node]int, writes []VariableWrite, variableCount int) []*ast.Node {
	if len(writes) == 0 {
		return nil
	}
	if variableCount == 1 {
		if dead, ok := deadWritesFromBinder(root, reads, writes); ok {
			return dead
		}
	}
	return deadWritesFromCFG(root, reads, writes, variableCount)
}

func deadWritesFromCFG(root *ast.Node, reads map[*ast.Node]int, writes []VariableWrite, variableCount int) []*ast.Node {
	// Only the last occurrence of a repeated input node is used. All CFG copies
	// of that write point at one flag, so a read from any copy keeps it live.
	indices := make(map[*ast.Node]int, len(writes))
	dead := make([]bool, len(writes))
	for index, write := range writes {
		indices[write.Node] = index
	}
	graph := Build(root, Hooks[variableEvent]{
		Read: func(b *Builder[variableEvent], node *ast.Node) {
			if b.Current().Reachable {
				if variable, ok := reads[node]; ok {
					b.Emit(variableEvent{variable: variable})
				}
			}
		},
		Write: func(b *Builder[variableEvent], node *ast.Node) {
			if b.Current().Reachable {
				if index, ok := indices[node]; ok {
					dead[index] = true
					b.Emit(variableEvent{variable: writes[index].Variable, write: &dead[index]})
				}
			}
		},
	})
	solveWriteLiveness(graph, variableCount)
	var result []*ast.Node
	for index, write := range writes {
		if dead[index] {
			result = append(result, write.Node)
		}
	}
	return result
}

// solveWriteLiveness solves backward liveness for all variables together. Each
// variable occupies one bit, avoiding a scan of every event for every symbol.
func solveWriteLiveness(graph *Graph[variableEvent], variableCount int) {
	words := (variableCount + 63) / 64
	blocks := graph.Blocks
	if len(blocks) == 1 && len(blocks[0].Successors) == 0 {
		markVariableBlockDead(blocks[0], make([]uint64, words))
		return
	}

	size := len(blocks) * words
	storage := make([]uint64, 3*size)
	def, liveIn, liveOut := storage[:size], storage[size:2*size], storage[2*size:]
	offsets, predecessors, queue := buildVariablePredecessors(blocks)
	queued := make([]bool, len(blocks))
	for i, blk := range blocks {
		start := i * words
		// Starting with no live-out values leaves exactly the reads that occur
		// before the first write of their variable (the block's use set).
		for j := len(blk.Events) - 1; j >= 0; j-- {
			e := blk.Events[j]
			word, bit := start+e.variable/64, uint64(1)<<(e.variable%64)
			if e.write == nil {
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
		markVariableBlockDead(blk, liveOut[i*words:(i+1)*words])
	}
}

// markVariableBlockDead visits each event once instead of searching forward from every
// assignment. live starts at the block's exit and is consumed as scratch.
func markVariableBlockDead(blk *variableBlock, live []uint64) {
	for i := len(blk.Events) - 1; i >= 0; i-- {
		e := blk.Events[i]
		word, bit := e.variable/64, uint64(1)<<(e.variable%64)
		if e.write == nil {
			live[word] |= bit
		} else {
			if live[word]&bit != 0 {
				*e.write = false
			}
			live[word] &^= bit
		}
	}
}

// buildVariablePredecessors stores reachable successor edges in reverse as a compact
// adjacency list. scratch is reused as the worklist, with at most one pending
// entry per block.
func buildVariablePredecessors(blocks []*variableBlock) (offsets []int, predecessors []int, scratch []int) {
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
