package no_useless_else

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/cfg"
)

type exitAnalysis struct {
	ctx    *rule.RuleContext
	values *utils.StaticStringEvaluator
}

func (a *exitAnalysis) analyze(root *ast.Node) map[*ast.Node]bool {
	options := cfg.BranchAnalysis{
		Truthiness:     a.truthiness,
		Nullish:        a.nullish,
		TerminalCall:   a.isProcessExit,
		SafeRead:       a.safeRead,
		SafeExpression: a.safeExpression,
	}
	exits := make(map[*ast.Node]bool)
	options.IfConsequent = func(node *ast.Node, entered, fallsThrough bool) {
		if !isCandidate(node) {
			return
		}
		exitsHere := !fallsThrough
		if !entered {
			// Upstream also checks the shape of a consequent under a constant
			// false condition, as long as the if statement itself is reachable.
			exitsHere = !cfg.StatementFallsThrough(node.AsIfStatement().ThenStatement, options)
		}
		if previous, exists := exits[node]; exists {
			exitsHere = exitsHere && previous
		}
		exits[node] = exitsHere
	}
	cfg.AnalyzeIfBranches(root, options)
	return exits
}

func (a *exitAnalysis) evaluator() *utils.StaticStringEvaluator {
	if a.values == nil {
		a.values = utils.NewStaticStringEvaluatorWithReferenceResolver(nil, a.ctx.SourceFile, a.ctx.Refs)
		a.values.GlobalAccess = a.ctx.Globals.Access
	}
	return a.values
}

func (a *exitAnalysis) truthiness(node *ast.Node) (bool, bool) {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false, false
	}
	switch node.Kind {
	case ast.KindTrueKeyword:
		return true, true
	case ast.KindFalseKeyword, ast.KindNullKeyword:
		return false, true
	}
	return a.evaluator().EvalControlFlowTruthiness(node)
}

func (a *exitAnalysis) nullish(node *ast.Node) (bool, bool) {
	return a.evaluator().EvalControlFlowNullish(node)
}

func (a *exitAnalysis) safeExpression(node *ast.Node) bool {
	// Spreading can run user code even when the source value is known.
	switch node.Kind {
	case ast.KindSpreadElement, ast.KindSpreadAssignment:
		return false
	}
	_, known := a.evaluator().EvalControlFlowValue(node)
	return known
}

func (a *exitAnalysis) isProcessExit(node *ast.Node) bool {
	callee := utils.SkipAssertionsAndParens(node.AsCallExpression().Expression)
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression || ast.IsOptionalChain(callee) {
		return false
	}
	member := callee.AsPropertyAccessExpression()
	name := member.Name()
	object := utils.ESTreeRuntimeExpression(member.Expression)
	return name != nil && name.Kind == ast.KindIdentifier && name.Text() == "exit" &&
		object != nil && object.Kind == ast.KindIdentifier && object.Text() == "process" &&
		a.ctx.Globals.Access("process") != utils.GlobalAccessOff &&
		a.ctx.Refs.IsGlobalReference(object) &&
		ast.FindAncestorKind(object, ast.KindWithStatement) == nil
}

func (a *exitAnalysis) safeRead(node *ast.Node) bool {
	symbol := a.ctx.Refs.ResolveInFile(node)
	if symbol == nil {
		return false
	}
	decl := symbol.ValueDeclaration
	if decl == nil {
		return false
	}
	// A simple parameter of this function is initialized before the body.
	return decl.Kind == ast.KindParameter && cfg.RootOf(node) == decl.Parent &&
		decl.Parent.Body() != nil && decl.Parent.Body().Pos() <= node.Pos()
}
