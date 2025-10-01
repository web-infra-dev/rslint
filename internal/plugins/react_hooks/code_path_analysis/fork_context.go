package code_path_analysis

// Manage forking
type ForkContext struct {
	idGenerator  *IdGenerator // idGenerator An identifier generator for segments.
	upper        *ForkContext // upper An upper fork context
	count        int          // count A number of parallel segments
	segmentsList [][]*CodePathSegment
}

func NewForkContext(idGenerator *IdGenerator, upper *ForkContext, count int) *ForkContext {
	return &ForkContext{
		idGenerator:  idGenerator,
		upper:        upper,
		count:        count,
		segmentsList: make([][]*CodePathSegment, 0),
	}
}

func NewEmptyForkContext(parentContext *ForkContext, forkLeavingPath *ForkContext) *ForkContext {
	count := parentContext.count
	if forkLeavingPath != nil {
		count = count * 2
	}
	return NewForkContext(
		parentContext.idGenerator,
		parentContext,
		count,
	)
}

func NewRootForkContext(idgenerator *IdGenerator) *ForkContext {
	context := NewForkContext(idgenerator, nil, 1)

	context.Add([]*CodePathSegment{
		NewRootCodePathSegment(idgenerator.Next()),
	})

	return context
}

// The head segments.
func (fc *ForkContext) Head() []*CodePathSegment {
	if len(fc.segmentsList) == 0 {
		return []*CodePathSegment{}
	}

	return fc.segmentsList[len(fc.segmentsList)-1]
}

// A flag which shows empty.
func (fc *ForkContext) IsEmpty() bool {
	return len(fc.segmentsList) == 0
}

// A flag which shows reachable.
func (fc *ForkContext) IsReachable() bool {
	isReachable := false
	segments := fc.Head()
	for _, segment := range segments {
		if segment.reachable {
			isReachable = true
			break
		}
	}
	return isReachable
}

// Creates new segments from this context.
func (fc *ForkContext) MakeNext(begin int, end int) []*CodePathSegment {
	return fc.makeSegments(begin, end, NewNextCodePathSegment)
}

// Creates new segments from this context. The new segments is always unreachable.
func (fc *ForkContext) MakeUnreachable(begin int, end int) []*CodePathSegment {
	return fc.makeSegments(begin, end, NewUnreachableCodePathSegment)
}

// Creates new segments from this context.
// The new segments don't have connections for previous segments.
// But these inherit the reachable flag from this context.
func (fc *ForkContext) MakeDisconnected(begin int, end int) []*CodePathSegment {
	return fc.makeSegments(begin, end, NewDisconnectedCodePathSegment)
}

// Creates new segments from the specific range of `context.segmentsList`.
//
// When `context.segmentsList` is `[[a, b], [c, d], [e, f]]`, `begin` is `0`, and
// `end` is `-1`, this creates `[g, h]`. This `g` is from `a`, `c`, and `e`.
// This `h` is from `b`, `d`, and `f`.
func (fc *ForkContext) makeSegments(begin int, end int, create func(id string, allPrevSegments []*CodePathSegment) *CodePathSegment) []*CodePathSegment {
	list := fc.segmentsList

	normalizedBegin := begin
	if begin < 0 {
		normalizedBegin = len(list) + begin
	}
	normalizedEnd := end
	if end < 0 {
		normalizedEnd = len(list) + end
	}

	segments := make([]*CodePathSegment, 0)

	for i := range fc.count {
		allPrevSegments := make([]*CodePathSegment, 0)
		for j := normalizedBegin; j <= normalizedEnd; j++ {
			allPrevSegments = append(allPrevSegments, list[j][i])
		}

		segment := create(fc.idGenerator.Next(), allPrevSegments)
		segments = append(segments, segment)
	}

	return segments
}

func (fc *ForkContext) mergeExtraSegments(segments []*CodePathSegment) []*CodePathSegment {
	currentSegments := segments

	for len(segments) > fc.count {
		merged := make([]*CodePathSegment, 0)

		length := len(currentSegments) / 2
		for i := range length {
			segment := NewNextCodePathSegment(
				fc.idGenerator.Next(),
				[]*CodePathSegment{
					currentSegments[i],
					currentSegments[i+length],
				},
			)
			merged = append(merged, segment)
		}

		currentSegments = merged
	}

	return currentSegments
}

// Adds segments into this context. The added segments become the head.
func (fc *ForkContext) Add(segments []*CodePathSegment) {
	fc.segmentsList = append(
		fc.segmentsList,
		fc.mergeExtraSegments(segments),
	)
}

// Replaces the head segments with given segments. The current head segments are removed.
func (fc *ForkContext) ReplaceHead(segments []*CodePathSegment) {
	if len(fc.segmentsList) == 0 {
		fc.Add(segments)
		return
	}

	mergedSegments := fc.mergeExtraSegments(segments)
	fc.segmentsList[len(fc.segmentsList)-1] = mergedSegments
}

// Adds all segments of a given fork context into this context.
func (fc *ForkContext) AddAll(context *ForkContext) {
	source := context.segmentsList

	fc.segmentsList = append(fc.segmentsList, source...)
}

// Clears all segments in this context.
func (fc *ForkContext) Clear() {
	fc.segmentsList = make([][]*CodePathSegment, 0)
}

func removeSegment(segments []*CodePathSegment, target *CodePathSegment) []*CodePathSegment {
	for i, segment := range segments {
		if segment == target {
			return append(segments[:i], segments[i+1:]...)
		}
	}
	return segments
}

// Disconnect given segments.
//
// This is used in a process for switch statements.
// If there is the "default" chunk before other cases, the order is different
// between node's and running's.
func RemoveConnection(prevSegments []*CodePathSegment, nextSegments []*CodePathSegment) {
	for i := range prevSegments {
		prevSegment := prevSegments[i]
		nextSegment := nextSegments[i]

		prevSegment.nextSegments = removeSegment(prevSegment.nextSegments, nextSegment)
		prevSegment.allNextSegments = removeSegment(prevSegment.allNextSegments, nextSegment)
		nextSegment.prevSegments = removeSegment(nextSegment.prevSegments, prevSegment)
		nextSegment.allPrevSegments = removeSegment(nextSegment.allPrevSegments, prevSegment)
	}
}
