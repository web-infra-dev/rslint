package code_path_analysis

import "slices"

type internalData struct {
	used               bool
	loopedPrevSegments []*CodePathSegment
}

// A code path segment.
type CodePathSegment struct {
	id              string             // The identifier of this code path. Rules use it to store additional information of each rule.
	nextSegments    []*CodePathSegment // An array of the next segments.
	prevSegments    []*CodePathSegment // An array of the previous segments.
	allNextSegments []*CodePathSegment // An array of the next segments. This array includes unreachable segments.
	allPrevSegments []*CodePathSegment // An array of the previous segments. This array includes unreachable segments.
	reachable       bool               // A flag which shows this is reachable.
	internal        *internalData      // Internal data.
}

func NewCodePathSegment(id string, allPrevSegments []*CodePathSegment, reachable bool) *CodePathSegment {
	segment := &CodePathSegment{
		id:              id,
		nextSegments:    make([]*CodePathSegment, 0),
		prevSegments:    make([]*CodePathSegment, 0),
		allNextSegments: make([]*CodePathSegment, 0),
		allPrevSegments: allPrevSegments,
		reachable:       reachable,
		internal: &internalData{
			used:               false,
			loopedPrevSegments: make([]*CodePathSegment, 0),
		},
	}

	for _, prevSegment := range segment.allPrevSegments {
		if prevSegment.reachable {
			segment.prevSegments = append(segment.prevSegments, prevSegment)
		}
	}

	return segment
}

// Creates the root segment.
func NewRootCodePathSegment(id string) *CodePathSegment {
	return NewCodePathSegment(id, []*CodePathSegment{} /*allPrevSegments*/, true /*reachable*/)
}

// Creates a segment that follows given segments.
func NewNextCodePathSegment(id string, allPrevSegments []*CodePathSegment) *CodePathSegment {
	reachable := false
	for _, segment := range allPrevSegments {
		if segment.reachable {
			reachable = true
			break
		}
	}
	return NewCodePathSegment(id, flattenUnusedSegments(allPrevSegments), reachable)
}

// Creates an unreachable segment that follows given segments.
func NewUnreachableCodePathSegment(id string, allPrevSegments []*CodePathSegment) *CodePathSegment {
	segment := NewCodePathSegment(id, flattenUnusedSegments(allPrevSegments), false /*reachable*/)

	// In `if (a) return a; foo();` case, the unreachable segment preceded by
	// the return statement is not used but must not be remove.
	markUsed(segment)
	return segment
}

// Creates a segment that follows given segments. This factory method does not connect with `allPrevSegments`. But this inherits `reachable` flag.
func NewDisconnectedCodePathSegment(id string, allPrevSegments []*CodePathSegment) *CodePathSegment {
	isReachable := false
	for _, prevSegment := range allPrevSegments {
		if prevSegment.reachable {
			isReachable = true
			break
		}
	}
	return NewCodePathSegment(id, []*CodePathSegment{}, isReachable)
}

// Checks a given previous segment is coming from the end of a loop.
func (cps *CodePathSegment) IsLoopedPrevSegment(segment *CodePathSegment) bool {
	return slices.Contains(cps.internal.loopedPrevSegments, segment)
}

// Replaces unused segments with the previous segments of each unused segment.
func flattenUnusedSegments(segments []*CodePathSegment) []*CodePathSegment {
	done := make(map[string]bool)
	retv := make([]*CodePathSegment, 0)
	for _, segment := range segments {
		// Ignores duplicated.
		if done[segment.id] {
			continue
		}

		if !segment.internal.used {
			for _, prevSegment := range segment.allPrevSegments {
				if !done[prevSegment.id] {
					done[prevSegment.id] = true
					retv = append(retv, prevSegment)
				}
			}
		} else {
			done[segment.id] = true
			retv = append(retv, segment)
		}
	}
	return retv
}

// Makes a given segment being used.
func markUsed(segment *CodePathSegment) {
	if segment.internal.used {
		return
	}

	segment.internal.used = true

	if segment.reachable {
		for _, prevSegment := range segment.allPrevSegments {
			prevSegment.allNextSegments = append(prevSegment.allNextSegments, segment)
			prevSegment.nextSegments = append(prevSegment.nextSegments, segment)
		}
	} else {
		for _, prevSegment := range segment.allPrevSegments {
			prevSegment.allNextSegments = append(prevSegment.allNextSegments, segment)
		}
	}
}

// Marks a previous segment as looped.
func markPrevSegmentAsLooped(segment *CodePathSegment, prevSegment *CodePathSegment) {
	segment.internal.loopedPrevSegments = append(segment.internal.loopedPrevSegments, prevSegment)
}

// Getter methods for accessing private fields

func (cps *CodePathSegment) ID() string {
	return cps.id
}

func (cps *CodePathSegment) NextSegments() []*CodePathSegment {
	return cps.nextSegments
}

func (cps *CodePathSegment) PrevSegments() []*CodePathSegment {
	return cps.prevSegments
}

func (cps *CodePathSegment) AllNextSegments() []*CodePathSegment {
	return cps.allNextSegments
}

func (cps *CodePathSegment) AllPrevSegments() []*CodePathSegment {
	return cps.allPrevSegments
}

func (cps *CodePathSegment) Reachable() bool {
	return cps.reachable
}
