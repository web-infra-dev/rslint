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
		stmt := node.AsIfStatement().ThenStatement
		exitsHere := !fallsThrough
		if !entered {
			// Upstream also checks the shape of a consequent under a constant
			// false condition, as long as the if statement itself is reachable.
			exitsHere = !cfg.StatementFallsThrough(stmt, options)
		}
		exitsHere = exitsHere || a.exitsAfterSafeOperations(stmt)
		if previous, exists := exits[node]; exists {
			exitsHere = exitsHere && previous
		}
		exits[node] = exitsHere
	}
	cfg.AnalyzeIfBranches(root, options)
	options.IfConsequent = nil
	for node, exitsHere := range exits {
		if !exitsHere {
			continue
		}
		stmt := node.AsIfStatement().ThenStatement
		if a.hasUnmodeledCatchPath(stmt, options) {
			exits[node] = false
		}
	}
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
		// Only these immutable language globals are known here. Configured
		// host globals may be accessors.
		switch node.Text() {
		case "undefined", "Infinity", "NaN":
			return ast.FindAncestorKind(node, ast.KindWithStatement) == nil &&
				a.ctx.Globals.Access(node.Text()) == utils.GlobalAccessReadonly &&
				a.ctx.Refs.IsGlobalReference(node)
		}
		return false
	}
	decl := symbol.ValueDeclaration
	if decl == nil {
		return false
	}
	switch decl.Kind {
	case ast.KindParameter:
		return parameterReadIsSafe(node, decl)
	case ast.KindVariableDeclaration:
		if decl.Initializer() == nil || decl.Parent == nil || !ast.IsVarConst(decl.Parent) {
			return false
		}
		stmt := decl.Parent.Parent
		// An earlier const in an enclosing block has been initialized.
		// Text order alone proves nothing across functions or switch clauses.
		return stmt != nil && stmt.Kind == ast.KindVariableStatement &&
			stmt.Flags&ast.NodeFlagsAmbient == 0 && stmt.Parent != nil &&
			(stmt.Parent.Kind == ast.KindBlock || stmt.Parent.Kind == ast.KindSourceFile) &&
			stmt.End() <= node.Pos() && cfg.RootOf(node) == cfg.RootOf(stmt) &&
			ast.FindAncestor(node, func(n *ast.Node) bool { return n == stmt.Parent }) != nil &&
			ast.FindAncestorKind(node, ast.KindWithStatement) == nil
	}
	return false
}

func parameterReadIsSafe(node, declaration *ast.Node) bool {
	// A parameter of this function is initialized before the body.
	function := declaration.Parent
	return function != nil && cfg.RootOf(node) == function && function.Body() != nil &&
		function.Body().Pos() <= node.Pos() && ast.FindAncestorKind(node, ast.KindWithStatement) == nil
}
