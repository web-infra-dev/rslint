package code_path_analyzer

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
		continueDestSegments: nil,
		brokenForkContext:    state.breakContext.brokenForkContext,
	}
}

func NewLoopContextForDoWhileStatement(state *CodePathState, label string) *LoopContext {
	return &LoopContext{
		upper:               state.loopContext,
		kind:                DoWhileStatement,
		label:               label,
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
		prevSegments:         nil,
		leftSegments:         nil,
		endOfLeftSegments:    nil,
		continueDestSegments: nil,
		brokenForkContext:    state.breakContext.brokenForkContext,
	}
}

// Creates a context object of a loop statement and stacks it.
func (s *CodePathState) PushLoopContext(kind LoopStatementKind, label string) *LoopContext {
	switch kind {
	case WhileStatement:
		return NewLoopContextForWhileStatement(s, label)
	case DoWhileStatement:
		return NewLoopContextForDoWhileStatement(s, label)
	case ForStatement:
		return NewLoopContextForForStatement(s, label)
	case ForInStatement:
		return NewLoopContextForForInStatement(s, label)
	case ForOfStatement:
		return NewLoopContextForForOfStatement(s, label)
	default:
		panic("unknown statement kind")
	}
}
