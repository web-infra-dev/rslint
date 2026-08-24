package cfg

// PathAnalysis answers graph-shape questions shared by rules that consume an
// ESLint-style code path. Thrown exits are deliberately excluded from final
// paths, matching ESLint analyses that treat thrown segments separately.
//
// All per-block state is indexed by Block.Index, so ordinary queries do not
// allocate or hash block pointers.
type PathAnalysis[E any] struct {
	graph        *Graph[E]
	blocks       []pathBlockState
	cyclePaths   *cyclePaths[E]
	shortestExit int
	finalPaths   uint8
	hasFinal     bool
	hasExit      bool
}

type pathBlockState struct {
	shortest int32
	flags    uint8
	paths    uint8
}

const (
	pathFinal uint8 = 1 << iota
	pathThrown
	pathExit
	pathCyclic
	pathOnEveryFinal
	pathCounting
	pathCountKnown
)

// AnalyzePaths computes reachability-derived path information for graph.
func AnalyzePaths[E any](graph *Graph[E]) *PathAnalysis[E] {
	if graph == nil || len(graph.Blocks) == 0 {
		return &PathAnalysis[E]{}
	}

	blockCount := len(graph.Blocks)
	analysis := &PathAnalysis[E]{
		graph:  graph,
		blocks: make([]pathBlockState, blockCount),
	}
	for _, block := range graph.ThrownBlocks {
		if block != nil && block.Reachable {
			analysis.blocks[block.Index()].flags |= pathThrown | pathExit
			analysis.hasExit = true
		}
	}
	for _, block := range graph.FinalBlocks {
		if block != nil && block.Reachable {
			analysis.blocks[block.Index()].flags |= pathExit
			analysis.hasExit = true
			if analysis.blocks[block.Index()].flags&pathThrown != 0 {
				continue
			}
			analysis.blocks[block.Index()].flags |= pathFinal
			analysis.hasFinal = true
		}
	}

	analysis.cyclePaths = newCyclePaths(graph, analysis.blocks)
	analysis.cyclePaths.countToEnd(0)
	analysis.finalPaths = analysis.countFinalPaths(0)
	analysis.findShortestPaths(graph)
	analysis.findFinalPathDominators(graph)
	return analysis
}

// IsCyclic reports whether block belongs to a reachable graph cycle.
func (a *PathAnalysis[E]) IsCyclic(block *Block[E]) bool {
	if a == nil || block == nil || block.Index() >= len(a.blocks) {
		return false
	}
	if a.cyclePaths != nil {
		a.cyclePaths.countToEnd(block.Index())
	}
	return a.blocks[block.Index()].flags&pathCyclic != 0
}

// IsOnEveryFinalPath reports whether every non-throwing path from the graph
// entry to a final block passes through block. It is true when the graph has no
// non-throwing final path, matching the vacuous path-count comparison used by
// ESLint consumers.
func (a *PathAnalysis[E]) IsOnEveryFinalPath(block *Block[E]) bool {
	return a != nil && block != nil && block.Index() < len(a.blocks) && a.blocks[block.Index()].flags&pathOnEveryFinal != 0
}

// IsExit reports whether block is a reachable final segment, including an
// uncaught thrown segment. ESLint includes thrown segments in codePath's
// finalSegments even though path counting deliberately assigns them zero
// paths.
func (a *PathAnalysis[E]) IsExit(block *Block[E]) bool {
	return a != nil && block != nil && block.Index() < len(a.blocks) && a.blocks[block.Index()].flags&pathExit != 0
}

// ShortestPathFromStart returns the number of blocks on the shortest path from
// the graph entry to block. The second result is false for an unreachable or
// unknown block.
func (a *PathAnalysis[E]) ShortestPathFromStart(block *Block[E]) (int, bool) {
	if a == nil || block == nil || block.Index() >= len(a.blocks) {
		return 0, false
	}
	distance := a.blocks[block.Index()].shortest
	return int(distance), distance != 0
}

// ShortestExitPathFromStart returns the number of blocks on the shortest path
// from the graph entry to any reachable final segment, thrown exits included.
func (a *PathAnalysis[E]) ShortestExitPathFromStart() (int, bool) {
	if a == nil || !a.hasExit {
		return 0, false
	}
	return a.shortestExit, true
}

// HasSingleFinalPath reports whether exactly one non-thrown path reaches a
// final block. The count saturates at two because consumers asking about a
// detached ESLint code-path segment only need to distinguish zero, one, and
// multiple paths.
func (a *PathAnalysis[E]) HasSingleFinalPath() bool {
	return a != nil && a.finalPaths == 1
}

func (a *PathAnalysis[E]) countFinalPaths(index int) uint8 {
	state := &a.blocks[index]
	if state.flags&pathCounting != 0 {
		return 0
	}
	if state.flags&pathCountKnown != 0 {
		return state.paths
	}
	state.flags |= pathCounting
	count := uint8(0)
	if state.flags&pathThrown == 0 {
		if state.flags&pathFinal != 0 {
			count = 1
		} else {
			for _, successor := range a.graph.Blocks[index].Successors {
				if successor == nil || !successor.Reachable {
					continue
				}
				count += a.countFinalPaths(successor.Index())
				if count > 1 {
					count = 2
					break
				}
			}
		}
	}
	state.flags &^= pathCounting
	state.flags |= pathCountKnown
	state.paths = count
	return count
}

func (a *PathAnalysis[E]) findShortestPaths(graph *Graph[E]) {
	var queue []int
	if a.cyclePaths != nil && cap(a.cyclePaths.path) >= len(graph.Blocks) {
		queue = a.cyclePaths.path[:1]
	} else {
		queue = make([]int, 1, len(graph.Blocks))
	}
	queue[0] = 0
	a.blocks[0].shortest = 1
	for head := 0; head < len(queue); head++ {
		index := queue[head]
		distance := a.blocks[index].shortest + 1
		for _, successor := range graph.Blocks[index].Successors {
			if successor == nil || !successor.Reachable {
				continue
			}
			successorIndex := successor.Index()
			if a.blocks[successorIndex].shortest != 0 {
				continue
			}
			a.blocks[successorIndex].shortest = distance
			queue = append(queue, successorIndex)
		}
	}
	if a.cyclePaths != nil {
		a.cyclePaths.path = queue[:0]
	}

	for index := range a.blocks {
		distance := a.blocks[index].shortest
		if distance == 0 {
			continue
		}
		if a.blocks[index].flags&pathExit != 0 && (a.shortestExit == 0 || int(distance) < a.shortestExit) {
			a.shortestExit = int(distance)
		}
	}
}

// cyclePaths mirrors the traversal and caching order used by ESLint's
// countPathsFromStart/countPathsToEnd helpers. That order is observable: once
// a shared successor is cached, a later switch clause may no longer walk the
// back edge that marked an earlier clause as cyclic.
type cyclePaths[E any] struct {
	graph  *Graph[E]
	blocks []pathBlockState
	state  []cycleBlockState
	path   []int
}

type cycleBlockState struct {
	activeAt int32
	flags    uint8
}

const (
	cycleThrown uint8 = 1 << iota
	cycleToKnown
	cycleToEnd
)

func newCyclePaths[E any](graph *Graph[E], blocks []pathBlockState) *cyclePaths[E] {
	blockCount := len(graph.Blocks)
	paths := &cyclePaths[E]{
		graph:  graph,
		blocks: blocks,
		state:  make([]cycleBlockState, blockCount),
		path:   make([]int, 0, blockCount),
	}
	for index := range paths.state {
		paths.state[index].activeAt = -1
	}
	for _, block := range graph.ThrownBlocks {
		if block != nil && block.Reachable {
			paths.state[block.Index()].flags |= cycleThrown
		}
	}
	return paths
}

func (p *cyclePaths[E]) countToEnd(index int) bool {
	return p.count(index, -1)
}

func (p *cyclePaths[E]) count(index int, barrierStart int) bool {
	if active := p.state[index].activeAt; active >= 0 {
		// ESLint records the segment after its loop context as the start of the
		// cycle, then marks only the later segments in pathArray. Excluding the
		// repeated header here gives the same observable result: a for-loop test
		// is not cyclic, while its body and incrementor are.
		cycleStart := int(active) + 1
		cycleEnd := len(p.path)
		if barrierStart > int(active) {
			cycleEnd = barrierStart
		}
		for _, member := range p.path[cycleStart:cycleEnd] {
			p.blocks[member].flags |= pathCyclic
		}
		return false
	}
	if p.state[index].flags&cycleToKnown != 0 {
		return p.state[index].flags&cycleToEnd != 0
	}

	p.state[index].activeAt = int32(len(p.path))
	p.path = append(p.path, index)
	hasPath := false
	// An uncaught throw is not a path to a normal exit. A yield is different:
	// it also has a reachable resumption successor, which must still be walked
	// to discover a loop back edge.
	block := p.graph.Blocks[index]
	traverse := p.state[index].flags&cycleThrown == 0
	if !traverse {
		for _, successor := range block.Successors {
			if successor != nil && successor.Reachable {
				traverse = true
				break
			}
		}
	}
	if traverse {
		hasPath = p.blocks[index].flags&pathFinal != 0
		hasSuccessor := false
		for successorIndex, successor := range block.Successors {
			if successor == nil || !successor.Reachable {
				continue
			}
			hasSuccessor = true
			nextBarrier := barrierStart
			if successorIndex < len(block.cycleBarrier) && block.cycleBarrier[successorIndex] {
				nextBarrier = len(p.path)
			}
			if p.count(successor.Index(), nextBarrier) {
				hasPath = true
			}
		}
		if !hasSuccessor && p.state[index].flags&cycleThrown == 0 {
			hasPath = true
		}
	}
	p.path = p.path[:len(p.path)-1]
	p.state[index].activeAt = -1

	if hasPath || !p.graph.Blocks[index].Reachable {
		p.state[index].flags |= cycleToKnown
		if hasPath {
			p.state[index].flags |= cycleToEnd
		} else {
			p.state[index].flags &^= cycleToEnd
		}
	} else {
		p.state[index].flags &^= cycleToKnown
	}
	return hasPath
}

// findFinalPathDominators adds one synthetic exit reached by every final block
// and computes immediate dominators with the Cooper-Harvey-Kennedy algorithm.
// The real blocks on the synthetic exit's dominator chain are exactly those on
// every non-throwing final path.
func (a *PathAnalysis[E]) findFinalPathDominators(graph *Graph[E]) {
	blockCount := len(graph.Blocks)
	if !a.hasFinal {
		for index, block := range graph.Blocks {
			if block.Reachable {
				a.blocks[index].flags |= pathOnEveryFinal
			}
		}
		return
	}

	exit := blockCount
	predCounts := make([]int, blockCount+2)
	for index, block := range graph.Blocks {
		if !block.Reachable {
			continue
		}
		for _, successor := range block.Successors {
			if successor != nil && successor.Reachable {
				predCounts[successor.Index()+1]++
			}
		}
		if a.blocks[index].flags&pathFinal != 0 {
			predCounts[exit+1]++
		}
	}
	for index := 1; index < len(predCounts); index++ {
		predCounts[index] += predCounts[index-1]
	}
	totalPredecessors := predCounts[len(predCounts)-1]
	nodeCount := blockCount + 1
	work := make([]int, totalPredecessors+nodeCount*4)
	predecessors := work[:totalPredecessors]
	predCursor := work[totalPredecessors : totalPredecessors+nodeCount]
	postOrderStorage := work[totalPredecessors+nodeCount : totalPredecessors+nodeCount*2]
	rpoPosition := work[totalPredecessors+nodeCount*2 : totalPredecessors+nodeCount*3]
	immediate := work[totalPredecessors+nodeCount*3:]
	copy(predCursor, predCounts[:nodeCount])
	for index, block := range graph.Blocks {
		if !block.Reachable {
			continue
		}
		for _, successor := range block.Successors {
			if successor == nil || !successor.Reachable {
				continue
			}
			successorIndex := successor.Index()
			predecessors[predCursor[successorIndex]] = index
			predCursor[successorIndex]++
		}
		if a.blocks[index].flags&pathFinal != 0 {
			predecessors[predCursor[exit]] = index
			predCursor[exit]++
		}
	}

	for index := range rpoPosition {
		rpoPosition[index] = -1
		immediate[index] = -1
	}
	postOrder := postOrderStorage[:0]
	var visit func(int)
	visit = func(index int) {
		rpoPosition[index] = 0
		if index < blockCount {
			for _, successor := range graph.Blocks[index].Successors {
				if successor != nil && successor.Reachable && rpoPosition[successor.Index()] == -1 {
					visit(successor.Index())
				}
			}
			if a.blocks[index].flags&pathFinal != 0 && rpoPosition[exit] == -1 {
				visit(exit)
			}
		}
		postOrder = append(postOrder, index)
	}
	visit(0)

	for left, right := 0, len(postOrder)-1; left < right; left, right = left+1, right-1 {
		postOrder[left], postOrder[right] = postOrder[right], postOrder[left]
	}
	for index, block := range postOrder {
		rpoPosition[block] = index
	}
	immediate[0] = 0
	intersect := func(left int, right int) int {
		for left != right {
			for rpoPosition[left] > rpoPosition[right] {
				left = immediate[left]
			}
			for rpoPosition[right] > rpoPosition[left] {
				right = immediate[right]
			}
		}
		return left
	}

	changed := true
	for changed {
		changed = false
		for _, index := range postOrder[1:] {
			start := predCounts[index]
			end := predCounts[index+1]
			newImmediate := -1
			for _, predecessor := range predecessors[start:end] {
				if immediate[predecessor] == -1 {
					continue
				}
				if newImmediate == -1 {
					newImmediate = predecessor
				} else {
					newImmediate = intersect(predecessor, newImmediate)
				}
			}
			if immediate[index] != newImmediate {
				immediate[index] = newImmediate
				changed = true
			}
		}
	}

	for index := immediate[exit]; index >= 0; index = immediate[index] {
		a.blocks[index].flags |= pathOnEveryFinal
		if index == 0 {
			break
		}
	}
}
