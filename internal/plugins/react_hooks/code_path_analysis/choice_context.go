package code_path_analysis

// A context for ConditionalExpression, LogicalExpression,
// AssignmentExpression (logical assignments only), IfStatement, WhileStatement,
// DoWhileStatement, or ForStatement.
//
// LogicalExpressions have cases that it goes different paths between the
// true case and the false case.
//
// For Example:
//
//	if (a || b) {
//	    foo();
//	} else {
//	    bar();
//	}
//
// In this case, `b` is evaluated always in the code path of the `else`
// block, but it's not so in the code path of the `if` block.
// So there are 3 paths:
//
//	a -> foo();
//	a -> b -> foo();
//	a -> b -> bar();
type ChoiceContext struct {
	upper             *ChoiceContext
	kind              string
	isForkingAsResult bool
	trueForkContext   *ForkContext
	falseForkContext  *ForkContext
	qqForkContext     *ForkContext
	processed         bool
}

func NewChoiceContext(state *CodePathState, kind string, isForkingAsResult bool) *ChoiceContext {
	return &ChoiceContext{
		upper:             state.choiceContext,
		kind:              kind,
		isForkingAsResult: isForkingAsResult,
		trueForkContext:   NewEmptyForkContext(state.forkContext, nil /*forkLeavingPath*/),
		falseForkContext:  NewEmptyForkContext(state.forkContext, nil /*forkLeavingPath*/),
		qqForkContext:     NewEmptyForkContext(state.forkContext, nil /*forkLeavingPath*/),
		processed:         false,
	}
}

func (s *CodePathState) PushChoiceContext(kind string, isForkingAsResult bool) {
	s.choiceContext = NewChoiceContext(s, kind, isForkingAsResult)
}

// Pops the last choice context and finalizes it.
func (s *CodePathState) PopChoiceContext() *ChoiceContext {
	context := s.choiceContext

	s.choiceContext = context.upper

	forkContext := s.forkContext
	headSegments := forkContext.Head()

	switch context.kind {
	case "&&", "||", "??":
		{
			// If any result were not transferred from child contexts,
			// this sets the head segments to both cases.
			// The head segments are the path of the right-hand operand.
			if !context.processed {
				context.trueForkContext.Add(headSegments)
				context.falseForkContext.Add(headSegments)
				context.qqForkContext.Add(headSegments)
			}

			// Transfers results to upper context if this context is in
			// test chunk.
			if context.isForkingAsResult {
				parentContext := s.choiceContext
				parentContext.trueForkContext.AddAll(context.trueForkContext)
				parentContext.falseForkContext.AddAll(context.falseForkContext)
				parentContext.qqForkContext.AddAll(context.qqForkContext)
				parentContext.processed = true

				return context
			}
		}
	case "test":
		{
			if !context.processed {
				// The head segments are the path of the `if` block here.
				// Updates the `true` path with the end of the `if` block.
				context.trueForkContext.Clear()
				context.trueForkContext.Add(headSegments)
			} else {
				// The head segments are the path of the `else` block here.
				// Updates the `false` path with the end of the `else`
				// block.
				context.falseForkContext.Clear()
				context.falseForkContext.Add(headSegments)
			}
		}
	case "loop":
		{
			// Loops are addressed in popLoopContext().
			// This is called from popLoopContext().

			return context
		}
	default:
		{
			panic("Unreachable")
		}
	}

	// Merges all paths.
	prevForkContext := context.trueForkContext

	prevForkContext.AddAll(context.falseForkContext)
	forkContext.ReplaceHead(prevForkContext.MakeNext(0, -1))

	return context
}

// Makes a code path segment of the right-hand operand of a logical expression.
func (s *CodePathState) MakeLogicalRight() {
	context := s.choiceContext
	forkContext := s.forkContext

	if context.processed {
		// This got segments already from the child choice context.
		// Creates the next path from own true/false fork context.
		var prevForkContext *ForkContext

		switch context.kind {
		case "&&": // if true then go to the right-hand side.
			prevForkContext = context.trueForkContext
		case "||": // if false then go to the right-hand side.
			prevForkContext = context.falseForkContext
		case "??": // Both true/false can short-circuit, so needs the third path to go to the right-hand side. That's qqForkContext.
			prevForkContext = context.qqForkContext
		default:
			panic("Unreachable")
		}

		forkContext.ReplaceHead(prevForkContext.MakeNext(0, -1))
		prevForkContext.Clear()
		context.processed = false
	} else {
		// This did not get segments from the child choice context.
		// So addresses the head segments.
		// The head segments are the path of the left-hand operand.
		switch context.kind {
		case "&&": // the false path can short-circuit.
			context.falseForkContext.Add(forkContext.Head())
		case "||": // the true path can short-circuit.
			context.trueForkContext.Add(forkContext.Head())
		case "??": // both can short-circuit.
			context.trueForkContext.Add(forkContext.Head())
			context.falseForkContext.Add(forkContext.Head())
		default:
			panic("Unreachable")
		}

		forkContext.ReplaceHead(forkContext.MakeNext(-1, -1))
	}
}

// Makes a code path segment of the `if` block.
func (s *CodePathState) MakeIfConsequent() {
	context := s.choiceContext
	forkContext := s.forkContext

	// If any result were not transferred from child contexts,
	// this sets the head segments to both cases.
	// The head segments are the path of the test expression.
	if !context.processed {
		context.trueForkContext.Add(forkContext.Head())
		context.falseForkContext.Add(forkContext.Head())
		context.qqForkContext.Add(forkContext.Head())
	}

	context.processed = false

	// Creates new path from the `true` case.
	forkContext.ReplaceHead(context.trueForkContext.MakeNext(0, -1))
}

// Makes a code path segment of the `else` block.
func (s *CodePathState) MakeIfAlternate() {
	context := s.choiceContext
	forkContext := s.forkContext

	// The head segments are the path of the `if` block.
	// Updates the `true` path with the end of the `if` block.
	context.trueForkContext.Clear()
	context.trueForkContext.Add(forkContext.Head())
	context.processed = true

	// Creates new path from the `false` case.
	forkContext.ReplaceHead(context.falseForkContext.MakeNext(0, -1))
}
