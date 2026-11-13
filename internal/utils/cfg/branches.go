package cfg

import "github.com/microsoft/TypeScript/tsc/shim/ast"

// BranchAnalysis supplies facts about values and calls to a reachability
// analysis. Rules decide which facts are safe to assume. The ordinary code
// path graph retains ESLint's more general, syntax-based model.
type BranchAnalysis struct {
	// A known test or nullish value must be safe to evaluate without throwing.
	Truthiness func(*ast.Node) (value bool, known bool)
	Nullish    func(*ast.Node) (value bool, known bool)
	// TerminalCall recognizes a call whose callee evaluation cannot throw and
	// which stops execution after its arguments have been evaluated.
	TerminalCall func(*ast.Node) bool
	// SafeRead recognizes identifier reads that cannot throw.
	SafeRead func(*ast.Node) bool
	// SafeExpression recognizes expressions whose implicit operations
	// (such as numeric coercion) cannot throw.
	SafeExpression func(*ast.Node) bool
	// IfConsequent visits the consequent of each reachable if. A finally body
	// may run on multiple paths, so a node may be visited more than once.
	IfConsequent func(node *ast.Node, entered, fallsThrough bool)
}

func (b *Builder[E]) possibleExpressionThrow(node *ast.Node) {
	if !b.cur.Reachable || len(b.tryStack) == 0 {
		return
	}
	if b.branches.SafeExpression == nil || !b.branches.SafeExpression(node) {
		b.firstThrowableFork()
	}
}

// AnalyzeIfBranches visits the if branches in one control-flow root.
func AnalyzeIfBranches(root *ast.Node, analysis BranchAnalysis) {
	b := newBuilder(Hooks[struct{}]{}, &analysis)
	b.visitRoot(root)
}

// StatementFallsThrough tests whether an isolated statement can complete
// normally if entered. A jump to an enclosing statement leaves this region.
func StatementFallsThrough(statement *ast.Node, analysis BranchAnalysis) bool {
	analysis.IfConsequent = nil
	b := newBuilder(Hooks[struct{}]{}, &analysis)
	b.statement(statement)
	return b.cur.Reachable
}

func (b *Builder[E]) truthiness(node *ast.Node) (bool, bool) {
	if b.branches != nil && b.branches.Truthiness != nil && b.cur.Reachable {
		return b.branches.Truthiness(node)
	}
	return false, false
}

func (b *Builder[E]) logicalRightOperand(operator ast.Kind, left *ast.Node) (bool, bool) {
	if operator == ast.KindQuestionQuestionToken {
		if b.branches != nil && b.branches.Nullish != nil && b.cur.Reachable {
			return b.branches.Nullish(left)
		}
		return false, false
	}
	truthy, known := b.truthiness(left)
	return truthy == (operator == ast.KindAmpersandAmpersandToken), known
}

// reachJump lets each finally between a jump and its destination run before
// that destination becomes reachable. The callers' jump stack still owns
// label resolution. A finally that exits abruptly replaces the pending jump.
func (b *Builder[E]) reachJump(destination *Block[E], tryDepth int) {
	for i := len(b.tryStack) - 1; i >= tryDepth; i-- {
		frame := b.tryStack[i]
		if frame.hasFinally {
			frame.pendingJumps = append(frame.pendingJumps, pendingJump[E]{destination, tryDepth})
			b.link(b.cur, frame.finallyEntry)
			return
		}
	}
	b.link(b.cur, destination)
}
