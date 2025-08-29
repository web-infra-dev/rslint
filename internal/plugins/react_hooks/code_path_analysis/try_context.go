package code_path_analysis

type TryContext struct {
	upper                  *TryContext
	position               string
	hasFinalizer           bool
	returnedForkContext    *ForkContext
	thrownForkContext      *ForkContext
	lastOfTryIsReachable   bool
	lastOfCatchIsReachable bool
}

func NewTryContext(state *CodePathState, hasFinalizer bool) *TryContext {
	var returnedForkContext *ForkContext
	if hasFinalizer {
		returnedForkContext = NewEmptyForkContext(state.forkContext, nil /*forkLeavingPath*/)
	}
	return &TryContext{
		upper:                  state.tryContext,
		position:               "try",
		hasFinalizer:           hasFinalizer,
		returnedForkContext:    returnedForkContext,
		thrownForkContext:      NewEmptyForkContext(state.forkContext, nil /*forkLeavingPath*/),
		lastOfTryIsReachable:   false,
		lastOfCatchIsReachable: false,
	}
}

// Creates a context object of TryStatement and stacks it.
func (s *CodePathState) PushTryContext(hasFinalizer bool) *TryContext {
	return NewTryContext(s, hasFinalizer)
}

// PopTryContext pops the last context of TryStatement and finalizes it.
func (s *CodePathState) PopTryContext() {
	context := s.tryContext
	s.tryContext = context.upper

	if context.position == "catch" {
		// Merges two paths from the `try` block and `catch` block merely.
		s.PopForkContext()
		return
	}

	// The following process is executed only when there is the `finally` block.
	returned := context.returnedForkContext
	thrown := context.thrownForkContext

	if returned.IsEmpty() && thrown.IsEmpty() {
		return
	}

	// Separate head to normal paths and leaving paths.
	headSegments := s.forkContext.Head()
	s.forkContext = s.forkContext.upper

	halfLength := len(headSegments) / 2
	normalSegments := headSegments[:halfLength]
	leavingSegments := headSegments[halfLength:]

	// Forwards the leaving path to upper contexts.
	if !returned.IsEmpty() {
		returnCtx := s.getReturnContext()
		if returnCtx != nil {
			returnCtx.returnedForkContext.Add(leavingSegments)
		}
	}
	if !thrown.IsEmpty() {
		throwCtx := s.getThrowContext()
		if throwCtx != nil {
			throwCtx.thrownForkContext.Add(leavingSegments)
		}
	}

	// Sets the normal path as the next.
	s.forkContext.ReplaceHead(normalSegments)

	// If both paths of the `try` block and the `catch` block are
	// unreachable, the next path becomes unreachable as well.
	if !context.lastOfTryIsReachable && !context.lastOfCatchIsReachable {
		s.forkContext.ReplaceHead(s.forkContext.MakeUnreachable(-1, -1))
	}
}

// Makes a code path segment for a `catch` block.
func (s *CodePathState) MakeCatchBlock() {
	context := s.tryContext
	forkContext := s.forkContext
	thrown := context.thrownForkContext

	// Update state.
	context.position = "catch"
	context.thrownForkContext = NewEmptyForkContext(forkContext, nil)
	context.lastOfTryIsReachable = forkContext.IsReachable()

	// Merge thrown paths.
	thrown.Add(forkContext.Head())
	thrownSegments := thrown.MakeNext(0, -1)

	// Fork to a bypass and the merged thrown path.
	s.PushForkContext(nil /*forkLeavingPath*/)
	s.ForkBypassPath()
	s.forkContext.Add(thrownSegments)
}

// MakeFinallyBlock makes a code path segment for a `finally` block.
//
// In the `finally` block, parallel paths are created. The parallel paths
// are used as leaving-paths. The leaving-paths are paths from `return`
// statements and `throw` statements in a `try` block or a `catch` block.
func (s *CodePathState) MakeFinallyBlock() {
	context := s.tryContext
	forkContext := s.forkContext
	returned := context.returnedForkContext
	thrown := context.thrownForkContext
	headOfLeavingSegments := forkContext.Head()

	// Update state.
	if context.position == "catch" {
		// Merges two paths from the `try` block and `catch` block.
		s.PopForkContext()
		forkContext = s.forkContext
		context.lastOfCatchIsReachable = forkContext.IsReachable()
	} else {
		context.lastOfTryIsReachable = forkContext.IsReachable()
	}
	context.position = "finally"

	if returned.IsEmpty() && thrown.IsEmpty() {
		// This path does not leave.
		return
	}

	// Create a parallel segment from merging returned and thrown.
	// This segment will leave at the end of this finally block.
	segments := forkContext.MakeNext(-1, -1)

	for i := 0; i < forkContext.count; i++ {
		prevSegsOfLeavingSegment := []*CodePathSegment{headOfLeavingSegments[i]}

		for j := 0; j < len(returned.segmentsList); j++ {
			prevSegsOfLeavingSegment = append(prevSegsOfLeavingSegment, returned.segmentsList[j][i])
		}
		for j := 0; j < len(thrown.segmentsList); j++ {
			prevSegsOfLeavingSegment = append(prevSegsOfLeavingSegment, thrown.segmentsList[j][i])
		}

		segments = append(segments, NewNextCodePathSegment(
			s.idGenerator.Next(),
			prevSegsOfLeavingSegment,
		))
	}

	s.PushForkContext(nil /*forkLeavingPath*/)
	s.forkContext.Add(segments)
}

// Makes a code path segment from the first throwable node
// to the `catch` block or the `finally` block.
func (s *CodePathState) MakeFirstThrowablePathInTryBlock() {
	forkContext := s.forkContext

	if !forkContext.IsReachable() {
		return
	}

	context := s.getThrowContext()

	if context == nil ||
		context.position != "try" ||
		!context.thrownForkContext.IsEmpty() {
		return
	}

	context.thrownForkContext.Add(forkContext.Head())
	forkContext.ReplaceHead(forkContext.MakeNext(-1, -1))
}

// Gets a context for a `return` statement.
func (s *CodePathState) getReturnContext() *TryContext {
	context := s.tryContext
	for context != nil {
		if context.hasFinalizer && context.position != "finally" {
			return context
		}
		context = context.upper
	}

	return nil
}

// Gets a context for a `throw` statement.
func (s *CodePathState) getThrowContext() *TryContext {
	context := s.tryContext
	for context != nil {
		if context.position == "try" ||
			(context.hasFinalizer && context.position == "catch") {
			return context
		}
		context = context.upper
	}
	// If no try context found, return nil (this should be handled by caller)
	return nil
}

// Makes the final path.
func (s *CodePathState) MakeFinal() {
	segments := s.currentSegments

	if len(segments) > 0 && segments[0].reachable {
		s.addReturnedSegments(segments)
	}
}

// Makes a path for a `throw` statement.
//
// It registers the head segment to a context of `throw`.
// It makes new unreachable segment, then it set the head with the segment.
func (s *CodePathState) MakeThrow() {
	forkContext := s.forkContext

	if forkContext.IsReachable() {
		throwCtx := s.getThrowContext()
		if throwCtx != nil {
			throwCtx.thrownForkContext.Add(forkContext.Head())
		}
		forkContext.ReplaceHead(forkContext.MakeUnreachable(-1, -1))
	}
}
