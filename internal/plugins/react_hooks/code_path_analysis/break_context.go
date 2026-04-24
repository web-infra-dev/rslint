package code_path_analysis

type BreakContext struct {
	upper             *BreakContext
	breakable         bool
	label             string
	brokenForkContext *ForkContext
}

func NewBreakContext(state *CodePathState, breakable bool, label string) *BreakContext {
	return &BreakContext{
		upper:             state.breakContext,
		breakable:         breakable,
		label:             label,
		brokenForkContext: NewEmptyForkContext(state.forkContext, nil /*forkLeavingPath*/),
	}
}

// Creates new context for BreakStatement.
func (s *CodePathState) PushBreakContext(breakable bool, label string) *BreakContext {
	s.breakContext = NewBreakContext(s, breakable, label)
	return s.breakContext
}

// Removes the top item of the break context stack.
func (s *CodePathState) PopBreakContext() *BreakContext {
	context := s.breakContext
	forkContext := s.forkContext

	s.breakContext = context.upper

	// Process this context here for other than switches and loops.
	if !context.breakable {
		brokenForkContext := context.brokenForkContext

		if !brokenForkContext.IsEmpty() {
			brokenForkContext.Add(forkContext.Head())
			forkContext.ReplaceHead(brokenForkContext.MakeNext(0, -1))
		}
	}

	return context
}

// Makes a path for a `break` statement.
// It registers the head segment to a context of `break`.
// It makes new unreachable segment, then it set the head with the segment.
func (s *CodePathState) MakeBreak(label string) {
	forkContext := s.forkContext

	if !forkContext.IsReachable() {
		return
	}

	context := s.getBreakContext(label)

	if context != nil {
		context.brokenForkContext.Add(forkContext.Head())
	}

	forkContext.ReplaceHead(forkContext.MakeUnreachable(-1, -1))
}

func (s *CodePathState) getBreakContext(label string) *BreakContext {
	context := s.breakContext

	for context != nil {
		if label == "" && context.breakable {
			return context
		} else if context.label == label {
			return context
		}

		context = context.upper
	}

	return nil
}

// Makes a path for a `continue` statement.
//
// It makes a looping path.
// It makes new unreachable segment, then it set the head with the segment.
func (s *CodePathState) MakeContinue(label string) {
	forkContext := s.forkContext

	if !forkContext.IsReachable() {
		return
	}

	context := s.getContinueContext(label)

	if context != nil {
		if context.continueDestSegments != nil {
			s.MakeLooped(forkContext.Head(), context.continueDestSegments)

			// If the context is a for-in/of loop, this effects a break also.
			if context.kind == ForInStatement || context.kind == ForOfStatement {
				context.brokenForkContext.Add(forkContext.Head())
			}
		} else {
			context.continueForkContext.Add(forkContext.Head())
		}
	}
	forkContext.ReplaceHead(forkContext.MakeUnreachable(-1, -1))
}

// Gets a loop-context for a `continue` statement.
func (s *CodePathState) getContinueContext(label string) *LoopContext {
	if label == "" {
		return s.loopContext
	}

	context := s.loopContext
	for context != nil {
		if context.label == label {
			return context
		}
		context = context.upper
	}

	return nil
}

// Makes a path for a `return` statement.
//
// It registers the head segment to a context of `return`.
// It makes new unreachable segment, then it set the head with the segment.
func (s *CodePathState) MakeReturn() {
	forkContext := s.forkContext

	if forkContext.IsReachable() {
		returnCtx := s.getReturnContext()
		if returnCtx != nil {
			returnCtx.returnedForkContext.Add(forkContext.Head())
		} else {
			s.addReturnedSegments(forkContext.Head())
		}
		forkContext.ReplaceHead(forkContext.MakeUnreachable(-1, -1))
	}
}
