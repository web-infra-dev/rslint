// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
package no_mismatched_map_key

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var NoMismatchedMapKeyRule = rule.Rule{
	Name:   "unicorn/no-mismatched-map-key",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		a := analyzer{ctx: ctx}
		return rule.RuleListeners{
			ast.KindBinaryExpression:      a.checkGuard,
			ast.KindConditionalExpression: a.checkGuard,
			ast.KindIfStatement:           a.checkGuard,
			ast.KindWhileStatement:        a.checkGuard,
		}
	},
}

type analyzer struct {
	ctx          rule.RuleContext
	keyValues    *utils.StaticStringEvaluator
	computedKeys unicornutil.ComputedKeyEvaluator
	reported     map[*ast.Node]struct{}
}

var hasCallOptions = unicornutil.DotMethodCallOptions{
	Method: "has", ArgumentsLength: new(1), RejectSpreadElement: true,
}

var accessCallOptions = unicornutil.DotMethodCallOptions{
	Methods: []string{"get", "set"}, MinimumArguments: new(1),
}

func (a *analyzer) checkGuard(node *ast.Node) {
	test, branches := guardedBranches(node)
	if test == nil || (branches[0] == nil && branches[1] == nil) {
		return
	}
	test = unwrap(test)
	if test != nil && ast.IsPrefixUnaryExpression(test) {
		prefix := test.AsPrefixUnaryExpression()
		if prefix.Operator == ast.KindExclamationToken {
			test = unwrap(prefix.Operand)
		}
	}
	if test == nil || ast.IsOptionalChain(test) {
		return
	}
	guard, ok := unicornutil.MatchDotMethodCall(test, hasCallOptions)
	if !ok {
		return
	}

	var keys [2]*ast.Node
	accessKeys := keys[:0]
	root := rootIdentifier(guard.Object)
	for _, branch := range branches {
		if branch != nil {
			accessKeys = a.collectAccessKeys(branch, guard.Object, root, accessKeys)
		}
	}
	if len(accessKeys) > 0 && !a.isKnownNonMap(guard.Object) {
		a.reportMismatchedKeys(guard.Call.Arguments()[0], accessKeys)
	}
}

func guardedBranches(node *ast.Node) (*ast.Node, [2]*ast.Node) {
	switch node.Kind {
	case ast.KindIfStatement:
		stmt := node.AsIfStatement()
		branches := [2]*ast.Node{singleStatement(stmt.ThenStatement)}
		if stmt.ElseStatement != nil && !ast.IsIfStatement(stmt.ElseStatement) {
			branches[1] = singleStatement(stmt.ElseStatement)
		}
		return lastAndOperand(stmt.Expression), branches
	case ast.KindWhileStatement:
		stmt := node.AsWhileStatement()
		return lastAndOperand(stmt.Expression), [2]*ast.Node{singleStatement(stmt.Statement)}
	case ast.KindConditionalExpression:
		expr := node.AsConditionalExpression()
		return lastAndOperand(expr.Condition), [2]*ast.Node{expr.WhenTrue, expr.WhenFalse}
	case ast.KindBinaryExpression:
		expr := node.AsBinaryExpression()
		if ast.IsLogicalOrCoalescingBinaryOperator(expr.OperatorToken.Kind) {
			return expr.Left, [2]*ast.Node{expr.Right}
		}
	}
	return nil, [2]*ast.Node{}
}

// Only the last operand of an && test is a guard for the branch.
func lastAndOperand(node *ast.Node) *ast.Node {
	node = utils.ESTreeRuntimeExpression(node)
	for node != nil && ast.IsBinaryExpression(node) {
		expr := node.AsBinaryExpression()
		if expr.OperatorToken.Kind != ast.KindAmpersandAmpersandToken {
			break
		}
		node = utils.ESTreeRuntimeExpression(expr.Right)
	}
	return node
}

func singleStatement(node *ast.Node) *ast.Node {
	if node == nil || !ast.IsBlock(node) {
		return node
	}
	stmts := node.Statements()
	if len(stmts) == 1 {
		return stmts[0]
	}
	return nil
}

func (a *analyzer) collectAccessKeys(branch, receiver, root *ast.Node, keys []*ast.Node) []*ast.Node {
	previous := len(keys)
	hasWrite := walkBranch(branch, func(node *ast.Node) bool {
		return a.isNestedGuard(node, receiver)
	}, func(node *ast.Node) bool {
		if a.writesReceiver(node, receiver, root) {
			return true
		}
		if access, ok := unicornutil.MatchDotMethodCall(node, accessCallOptions); ok &&
			!slices.ContainsFunc(access.Call.Arguments(), ast.IsSpreadElement) &&
			a.sameReference(receiver, access.Object) {
			keys = append(keys, utils.ESTreeRuntimeExpression(access.Call.Arguments()[0]))
		}
		return false
	})
	// A write anywhere in this branch invalidates the guard for every access.
	if hasWrite {
		return keys[:previous]
	}
	return keys
}

func (a *analyzer) reportMismatchedKeys(checkedKey *ast.Node, keys []*ast.Node) {
	if walkBranch(checkedKey, isWrapperType, isRuntimeCall) {
		return
	}
	checkedKey = unwrap(checkedKey)
	var checkedValue any
	var checkedKnown, evaluated bool
	for _, key := range keys {
		if _, alreadyReported := a.reported[key]; alreadyReported || walkBranch(key, isWrapperType, isRuntimeCall) {
			continue
		}
		if !evaluated {
			checkedValue, checkedKnown = a.evaluateKey(checkedKey)
			checkedKnown = checkedKnown && utils.IsPrimitiveStaticValue(checkedValue)
			evaluated = true
		}
		if !a.differentKeys(checkedKey, unwrap(key), checkedValue, checkedKnown) {
			continue
		}
		if a.reported == nil {
			a.reported = make(map[*ast.Node]struct{})
		}
		a.reported[key] = struct{}{}
		a.ctx.ReportNode(key, rule.RuleMessage{
			Id:          "no-mismatched-map-key",
			Description: "Do not check a Map key before accessing a different key.",
		})
	}
}

func (a *analyzer) isNestedGuard(node, receiver *ast.Node) bool {
	var test *ast.Node
	switch node.Kind {
	case ast.KindIfStatement:
		test = node.AsIfStatement().Expression
	case ast.KindConditionalExpression:
		test = node.AsConditionalExpression().Condition
	default:
		return false
	}
	return walkBranch(test, nil, func(node *ast.Node) bool {
		call, ok := unicornutil.MatchDotMethodCall(node, hasCallOptions)
		return ok && a.sameReference(receiver, call.Object)
	})
}

func (a *analyzer) evaluateKey(node *ast.Node) (any, bool) {
	if a.keyValues == nil {
		a.keyValues = utils.NewStaticStringEvaluatorWithReferenceResolver(a.ctx.TypeChecker, a.ctx.SourceFile, a.ctx.Refs)
		a.keyValues.GlobalAccess = a.ctx.Globals.Access
	}
	return a.keyValues.EvalValueIfNoSideEffects(node)
}

func (a *analyzer) differentKeys(left, right *ast.Node, leftValue any, leftKnown bool) bool {
	if leftKnown {
		if rightValue, rightKnown := a.evaluateKey(right); rightKnown {
			if equal, known := utils.StaticValuesSameValueZero(leftValue, rightValue); known {
				return !equal
			}
		}
	}
	return isKeyReference(left) && isKeyReference(right) && !a.sameReference(left, right)
}

func isKeyReference(node *ast.Node) bool {
	return node != nil && !ast.IsOptionalChain(node) &&
		(ast.IsIdentifier(node) || ast.IsAccessExpression(node) || node.Kind == ast.KindThisKeyword || node.Kind == ast.KindSuperKeyword)
}

func isRuntimeCall(node *ast.Node) bool {
	// tsgo represents dynamic import as a call. ESTree uses ImportExpression.
	return ast.IsCallExpression(node) && !ast.IsImportCall(node)
}

const typeScriptWrappers = ast.OEKAssertions | ast.OEKExpressionsWithTypeArguments

func unwrap(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	return ast.SkipOuterExpressions(node, ast.OEKParentheses|typeScriptWrappers)
}

func isWrapperType(node *ast.Node) bool {
	parent := node.Parent
	return parent != nil && ast.IsOuterExpression(parent, typeScriptWrappers) && parent.Expression() != node
}

// Stop at code that does not run in the branch. Object method names may be
// computed, so their keys are still visited while their bodies are skipped.
func walkBranch(node *ast.Node, skip, visit func(*ast.Node) bool) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindArrowFunction, ast.KindFunctionDeclaration, ast.KindFunctionExpression,
		ast.KindClassDeclaration, ast.KindClassExpression:
		return false
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return walkBranch(node.Name(), skip, visit)
	}
	if skip != nil && skip(node) {
		return false
	}
	return visit(node) || node.ForEachChild(func(child *ast.Node) bool { return walkBranch(child, skip, visit) })
}
