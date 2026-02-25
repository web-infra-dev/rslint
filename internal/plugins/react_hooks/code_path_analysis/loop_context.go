package code_path_analysis

type LoopStatementKind = uint

const (
	WhileStatement LoopStatementKind = iota + 1
	DoWhileStatement
	ForStatement
	ForInStatement
	ForOfStatement
)

type LoopContext struct {
	upper                *LoopContext
	kind                 LoopStatementKind
	label                string
	test                 bool
	entrySegments        []*CodePathSegment
	continueDestSegments []*CodePathSegment
	endOfInitSegments    []*CodePathSegment
	testSegments         []*CodePathSegment
	endOfTestSegments    []*CodePathSegment
	updateSegments       []*CodePathSegment
	endOfUpdateSegments  []*CodePathSegment
	leftSegments         []*CodePathSegment
	endOfLeftSegments    []*CodePathSegment
	prevSegments         []*CodePathSegment
	brokenForkContext    *ForkContext
	continueForkContext  *ForkContext
}

func NewLoopContextForWhileStatement(state *CodePathState, label string) *LoopContext {
	return &LoopContext{
		upper:                state.loopContext,
		kind:                 WhileStatement,
		label:                label,
		test:                 false,
		continueDestSegments: nil,
		brokenForkContext:    state.breakContext.brokenForkContext,
	}
}

func NewLoopContextForDoWhileStatement(state *CodePathState, label string) *LoopContext {
	return &LoopContext{
		upper:               state.loopContext,
		kind:                DoWhileStatement,
		label:               label,
		test:                false,
		entrySegments:       nil,
		continueForkContext: NewEmptyForkContext(state.forkContext, nil /*forkLeavingPath*/),
		brokenForkContext:   state.breakContext.brokenForkContext,
	}
}

func NewLoopContextForForStatement(state *CodePathState, label string) *LoopContext {
	return &LoopContext{
		upper:                state.loopContext,
		kind:                 ForStatement,
		label:                label,
		test:                 false,
		endOfInitSegments:    nil,
		testSegments:         nil,
		endOfTestSegments:    nil,
		updateSegments:       nil,
		endOfUpdateSegments:  nil,
		continueDestSegments: nil,
		brokenForkContext:    state.breakContext.brokenForkContext,
	}
}

func NewLoopContextForForInStatement(state *CodePathState, label string) *LoopContext {
	return &LoopContext{
		upper:                state.loopContext,
		kind:                 ForInStatement,
		label:                label,
		test:                 false,
		prevSegments:         nil,
		leftSegments:         nil,
		endOfLeftSegments:    nil,
		continueDestSegments: nil,
		brokenForkContext:    state.breakContext.brokenForkContext,
	}
}

func NewLoopContextForForOfStatement(state *CodePathState, label string) *LoopContext {
	return &LoopContext{
		upper:                state.loopContext,
		kind:                 ForOfStatement,
		label:                label,
		test:                 false,
		prevSegments:         nil,
		leftSegments:         nil,
		endOfLeftSegments:    nil,
		continueDestSegments: nil,
		brokenForkContext:    state.breakContext.brokenForkContext,
	}
}

// Creates a context object of a loop statement and stacks it.
func (s *CodePathState) PushLoopContext(kind LoopStatementKind, label string) {
	s.PushBreakContext(true, label)
	switch kind {
	case WhileStatement:
		s.PushChoiceContext("loop", false)
		s.loopContext = NewLoopContextForWhileStatement(s, label)
	case DoWhileStatement:
		s.PushChoiceContext("loop", false)
		s.loopContext = NewLoopContextForDoWhileStatement(s, label)
	case ForStatement:
		s.PushChoiceContext("loop", false)
		s.loopContext = NewLoopContextForForStatement(s, label)
	case ForInStatement:
		s.loopContext = NewLoopContextForForInStatement(s, label)
	case ForOfStatement:
		s.loopContext = NewLoopContextForForOfStatement(s, label)
	default:
		panic("unknown statement kind")
	}
}

// Pops the last context of a loop statement and finalizes it.
func (s *CodePathState) PopLoopContext() {
	context := s.loopContext

	s.loopContext = context.upper

	forkContext := s.forkContext
	brokenForkContext := s.PopBreakContext().brokenForkContext

	switch context.kind {
	case WhileStatement, ForStatement:
		{
			s.PopChoiceContext()
			s.MakeLooped(forkContext.Head(), context.continueDestSegments)
		}
	case DoWhileStatement:
		{
			choiceContext := s.PopChoiceContext()

			if !choiceContext.processed {
				choiceContext.trueForkContext.Add(forkContext.Head())
				choiceContext.falseForkContext.Add(forkContext.Head())
			}
			if !context.test {
				brokenForkContext.AddAll(choiceContext.falseForkContext)
			}

			// `true` paths go to looping.
			segmentsList := choiceContext.trueForkContext.segmentsList

			for _, segment := range segmentsList {
				s.MakeLooped(segment, context.entrySegments)
			}
		}
	case ForInStatement, ForOfStatement:
		{
			brokenForkContext.Add(forkContext.Head())
			s.MakeLooped(forkContext.Head(), context.leftSegments)
		}
	default:
		panic("unreachable")
	}

	// Go next
	if brokenForkContext.IsEmpty() {
		forkContext.ReplaceHead((forkContext.MakeUnreachable(-1, -1)))
	} else {
		forkContext.ReplaceHead(brokenForkContext.MakeNext(0, -1))
	}
}

// Makes a code path segment for the test part of a WhileStatement.
func (s *CodePathState) MakeWhileTest(test bool) {
	context := s.loopContext
	forkContext := s.forkContext
	testSegments := forkContext.MakeNext(0, -1)

	// Update state.
	context.test = test
	context.continueDestSegments = testSegments
	s.forkContext.ReplaceHead(testSegments)
}

// Makes a code path segment for the body part of a WhileStatement.
func (s *CodePathState) MakeWhileBody() {
	context := s.loopContext
	choiceContext := s.choiceContext
	forkContext := s.forkContext

	if !choiceContext.processed {
		choiceContext.trueForkContext.Add(forkContext.Head())
		choiceContext.falseForkContext.Add(forkContext.Head())
	}

	// Update state.
	if !context.test {
		context.brokenForkContext.AddAll(choiceContext.falseForkContext)
	}
	forkContext.ReplaceHead(choiceContext.trueForkContext.MakeNext(0, -1))
}

// Makes a code path segment for the body part of a DoWhileStatement.
func (s *CodePathState) MakeDoWhileBody() {
	context := s.loopContext
	forkContext := s.forkContext
	bodySegments := forkContext.MakeNext(-1, -1)

	// Update state.
	context.entrySegments = bodySegments
	forkContext.ReplaceHead(bodySegments)
}

// Makes a code path segment for the test part of a DoWhileStatement.
func (s *CodePathState) MakeDoWhileTest(test bool) {
	context := s.loopContext
	forkContext := s.forkContext

	context.test = test

	// Creates paths of `continue` statements.
	if !context.continueForkContext.IsEmpty() {
		context.continueForkContext.Add(forkContext.Head())
		testSegments := context.continueForkContext.MakeNext(0, -1)

		forkContext.ReplaceHead(testSegments)
	}
}

// Makes a code path segment for the test part of a ForStatement.
func (s *CodePathState) MakeForTest(test bool) {
	context := s.loopContext
	forkContext := s.forkContext
	endOfInitSegments := forkContext.Head()
	testSegments := forkContext.MakeNext(-1, -1)

	// Update state.
	context.test = test
	context.endOfInitSegments = endOfInitSegments
	context.continueDestSegments = testSegments
	context.testSegments = testSegments
	forkContext.ReplaceHead(testSegments)
}

// Makes a code path segment for the update part of a ForStatement.
func (s *CodePathState) MakeForUpdate() {
	context := s.loopContext
	choiceContext := s.choiceContext
	forkContext := s.forkContext

	// Make the next paths of the test.
	if context.testSegments != nil {
		finalizeTestSegmentsOfFor(context, choiceContext, forkContext.Head())
	} else {
		context.endOfInitSegments = forkContext.Head()
	}

	// Update state.
	updateSegments := forkContext.MakeDisconnected(-1, -1)

	context.continueDestSegments = updateSegments
	context.updateSegments = updateSegments
	forkContext.ReplaceHead(updateSegments)
}

// Makes a code path segment for the body part of a ForStatement.
func (s *CodePathState) MakeForBody() {
	context := s.loopContext
	choiceContext := s.choiceContext
	forkContext := s.forkContext

	// Update state.
	if context.updateSegments != nil {
		context.endOfUpdateSegments = forkContext.Head()

		// `update` -> `test`
		if context.testSegments != nil {
			s.MakeLooped(context.endOfUpdateSegments, context.testSegments)
		}
	} else if context.testSegments != nil {
		finalizeTestSegmentsOfFor(context, choiceContext, forkContext.Head())
	} else {
		context.endOfInitSegments = forkContext.Head()
	}

	bodySegments := context.endOfTestSegments
	if bodySegments == nil {
		/*
		 * If there is not the `test` part, the `body` path comes from the
		 * `init` part and the `update` part.
		 */
		prevForkContext := NewEmptyForkContext(forkContext, nil)

		prevForkContext.Add(context.endOfInitSegments)
		if context.endOfUpdateSegments != nil {
			prevForkContext.Add(context.endOfUpdateSegments)
		}

		bodySegments = prevForkContext.MakeNext(0, -1)
	}

	if context.continueDestSegments == nil {
		context.continueDestSegments = bodySegments
	}
	forkContext.ReplaceHead(bodySegments)
}

// Makes a code path segment for the left part of a ForInStatement and a ForOfStatement.
func (s *CodePathState) MakeForInOfLeft() {
	context := s.loopContext
	forkContext := s.forkContext
	leftSegments := forkContext.MakeDisconnected(-1, -1)

	// Update state.
	context.prevSegments = forkContext.Head()
	context.leftSegments = leftSegments
	context.continueDestSegments = leftSegments
	forkContext.ReplaceHead(leftSegments)
}

// Makes a code path segment for the right part of a ForInStatement and a ForOfStatement.
func (s *CodePathState) MakeForInOfRight() {
	context := s.loopContext
	forkContext := s.forkContext
	temp := NewEmptyForkContext(forkContext, nil)

	temp.Add(context.prevSegments)
	rightSegments := temp.MakeNext(-1, -1)

	// Update state.
	context.endOfLeftSegments = forkContext.Head()
	forkContext.ReplaceHead(rightSegments)
}

// Makes a code path segment for the body part of a ForInStatement and a ForOfStatement.
func (s *CodePathState) MakeForInOfBody() {
	context := s.loopContext
	forkContext := s.forkContext
	temp := NewEmptyForkContext(forkContext, nil)

	temp.Add(context.endOfLeftSegments)
	bodySegments := temp.MakeNext(-1, -1)

	// Make a path: `right` -> `left`.
	s.MakeLooped(forkContext.Head(), context.leftSegments)

	// Update state.
	context.brokenForkContext.Add(forkContext.Head())
	forkContext.ReplaceHead(bodySegments)
}

func finalizeTestSegmentsOfFor(loopContext *LoopContext, choiceContext *ChoiceContext, head []*CodePathSegment) {
	if !choiceContext.processed {
		choiceContext.trueForkContext.Add(head)
		choiceContext.falseForkContext.Add(head)
		choiceContext.qqForkContext.Add(head)
	}

	if !loopContext.test {
		loopContext.brokenForkContext.AddAll(choiceContext.falseForkContext)
	}
	loopContext.endOfTestSegments = choiceContext.trueForkContext.MakeNext(0, -1)
}
