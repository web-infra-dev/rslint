package code_path_analysis

type ChainContext struct {
	upper              *ChainContext
	countChoiceContext int
}

func NewChainContext(state *CodePathState) *ChainContext {
	return &ChainContext{
		upper:              state.chainContext,
		countChoiceContext: 0,
	}
}

// Push a new `ChainExpression` context to the stack.
// This method is called on entering to each `ChainExpression` node.
// This context is used to count forking in the optional chain then merge them on the exiting from the `ChainExpression` node.
func (s *CodePathState) PushChainContext() {
	s.chainContext = NewChainContext(s)
}

// Pop a `ChainExpression` context from the stack.
// This method is called on exiting from each `ChainExpression` node.
// This merges all forks of the last optional chaining.
func (s *CodePathState) PopChainContext() {
	context := s.chainContext
	s.chainContext = context.upper

	// pop all choice contexts of this.
	for i := context.countChoiceContext; i > 0; i-- {
		s.PopChoiceContext()
	}
}

// Create a choice context for optional access.
// This method is called on entering to each `(Call|Member)Expression[optional=true]` node.
// This creates a choice context as similar to `LogicalExpression[operator="??"]` node.
func (s *CodePathState) MakeOptionalNode() {
	if s.chainContext != nil {
		s.chainContext.countChoiceContext += 1
		s.PushChoiceContext("??", false)
	}
}

// Create a fork.
// This method is called on entering to the `arguments|property` property of each `(Call|Member)Expression` node.
func (s *CodePathState) MakeOptionalRight() {
	if s.chainContext != nil {
		s.MakeLogicalRight()
	}
}
