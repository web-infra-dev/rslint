package code_path_analysis

type SwitchContext struct {
	upper               *SwitchContext
	hasCase             bool
	defaultSegments     []*CodePathSegment
	defaultBodySegments []*CodePathSegment
	foundDefault        bool
	lastIsDefault       bool
	countForks          int
}

func NewSwitchContext(state *CodePathState, hasCase bool, label string) *SwitchContext {
	return &SwitchContext{
		upper:               state.switchContext,
		hasCase:             hasCase,
		defaultSegments:     nil,
		defaultBodySegments: nil,
		foundDefault:        false,
		lastIsDefault:       false,
		countForks:          0,
	}
}

// Creates a context object of SwitchStatement and stacks it.
func (s *CodePathState) PushSwitchContext(hasCase bool, label string) {
	s.switchContext = NewSwitchContext(s, hasCase, label)

	s.PushBreakContext(true /*breakable*/, label)
}

// Pops the last context of SwitchStatement and finalizes it.
//
//   - Disposes all forking stack for `case` and `default`.
//   - Creates the next code path segment from `context.brokenForkContext`.
//   - If the last `SwitchCase` node is not a `default` part, creates a path
//     to the `default` body.
func (s *CodePathState) PopSwitchContext() {
	context := s.switchContext

	s.switchContext = context.upper

	forkContext := s.forkContext
	brokenForkContext := s.PopBreakContext().brokenForkContext

	if context.countForks == 0 {
		// When there is only one `default` chunk and there is one or more
		// `break` statements, even if forks are nothing, it needs to merge
		// those.
		if !brokenForkContext.IsEmpty() {
			brokenForkContext.Add(forkContext.MakeNext(-1, -1))
			forkContext.ReplaceHead(brokenForkContext.MakeNext(0, -1))
		}

		return
	}

	lastSegments := forkContext.Head()

	s.ForkBypassPath()
	lastCaseSegments := forkContext.Head()

	// `brokenForkContext` is used to make the next segment.
	// It must add the last segment into `brokenForkContext`.
	brokenForkContext.Add(lastSegments)

	// path which is failed in all case test should be connected to path
	// of `default` chunk.
	if !context.lastIsDefault {
		if context.defaultBodySegments != nil {
			// Remove a link from `default` label to its chunk.
			// It's false route.
			RemoveConnection(context.defaultSegments, context.defaultBodySegments)
			s.MakeLooped(lastCaseSegments, context.defaultBodySegments)
		} else {
			// It handles the last case body as broken if `default` chunk
			// does not exist.
			brokenForkContext.Add(lastCaseSegments)
		}
	}

	// Pops the segment context stack until the entry segment.
	for range context.countForks {
		s.forkContext = s.forkContext.upper
	}

	// Creates a path from all brokenForkContext paths.
	// This is a path after switch statement.
	s.forkContext.ReplaceHead(brokenForkContext.MakeNext(0, -1))
}

// Makes a code path segment for a `SwitchCase` node.
func (s *CodePathState) MakeSwitchCaseBody(isEmpty bool, isDefault bool) {
	context := s.switchContext

	if !context.hasCase {
		return
	}

	// Merge forks.
	// The parent fork context has two segments.
	// Those are from the current case and the body of the previous case.
	parentForkContext := s.forkContext
	forkContext := s.PushForkContext(nil /*forkLeavingPath*/)

	forkContext.Add(parentForkContext.MakeNext(0, -1))

	// Save default chunk info.
	// If the default label is not at the last, we must make a path from
	// the last case to the default chunk.
	if isDefault {
		context.defaultSegments = parentForkContext.Head()
		if isEmpty {
			context.foundDefault = true
		} else {
			context.defaultBodySegments = forkContext.Head()
		}
	} else {
		if !isEmpty && context.foundDefault {
			context.foundDefault = false
			context.defaultBodySegments = forkContext.Head()
		}
	}

	context.lastIsDefault = isDefault
	context.countForks++
}
