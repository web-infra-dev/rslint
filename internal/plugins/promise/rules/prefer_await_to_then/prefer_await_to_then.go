package prefer_await_to_then

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed prefer-await-to-then.schema.json
var schemaJSON []byte

const skipTransparent = ast.OEKParentheses

type Options struct {
	Strict bool
}

func parseOptions(options []any) Options {
	opts := Options{}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]interface{})
	opts.Strict, _ = optsMap["strict"].(bool)
	return opts
}

var preferAwaitToCallbackMessage = rule.RuleMessage{
	Id:          "preferAwaitToCallback",
	Description: "Prefer await to then()/catch()/finally().",
}

// createsScope reports whether node opens a new eslint-scope scope.
func createsScope(node *ast.Node) bool {
	if ast.IsFunctionLike(node) {
		return true
	}
	switch node.Kind {
	case ast.KindForStatement:
		// eslint-scope only opens a for-scope when the loop head binds with let/const/using.
		fs := node.AsForStatement()
		return fs != nil && fs.Initializer != nil &&
			ast.IsVariableDeclarationList(fs.Initializer) &&
			fs.Initializer.Flags&ast.NodeFlagsBlockScoped != 0
	case ast.KindForInStatement, ast.KindForOfStatement:
		// Same rule as KindForStatement: scope only when let/const/using.
		fs := node.AsForInOrOfStatement()
		return fs != nil && fs.Initializer != nil &&
			ast.IsVariableDeclarationList(fs.Initializer) &&
			fs.Initializer.Flags&ast.NodeFlagsBlockScoped != 0
	case ast.KindBlock,
		ast.KindSwitchStatement, ast.KindCatchClause,
		ast.KindClassDeclaration, ast.KindClassExpression,
		ast.KindWithStatement, ast.KindClassStaticBlockDeclaration,
		ast.KindModuleDeclaration:
		return true
	}
	return false
}

// isTopLevelScoped reports whether node is outside any scope-creating ancestor,
// mirroring ESLint's getScope().block.type === 'Program' check.
func isTopLevelScoped(node *ast.Node) bool {
	for cur := node.Parent; cur != nil; cur = cur.Parent {
		if createsScope(cur) {
			return false
		}
		if cur.Kind == ast.KindSourceFile {
			return true
		}
	}
	return true
}

// isInsideYieldOrAwait reports whether any ancestor of node is an AwaitExpression
// or YieldExpression.
func isInsideYieldOrAwait(node *ast.Node) bool {
	for cur := node.Parent; cur != nil; cur = cur.Parent {
		if cur.Kind == ast.KindAwaitExpression || cur.Kind == ast.KindYieldExpression {
			return true
		}
	}
	return false
}

// isInsideConstructor reports whether any ancestor of node is a constructor declaration.
func isInsideConstructor(node *ast.Node) bool {
	for cur := node.Parent; cur != nil; cur = cur.Parent {
		if cur.Kind == ast.KindConstructor {
			return true
		}
	}
	return false
}

// isCypress reports whether callNode is part of a Cypress cy.* chain.
// It mirrors eslint-plugin-promise's recursive isMemberCallWithObjectName('cy', node) check.
func isCypress(node *ast.Node) bool {
	if node == nil || !ast.IsCallExpression(node) {
		return false
	}
	callee := ast.SkipOuterExpressions(node.AsCallExpression().Expression, skipTransparent)
	if callee == nil {
		return false
	}
	var rawObj *ast.Node
	switch {
	case ast.IsPropertyAccessExpression(callee):
		rawObj = callee.AsPropertyAccessExpression().Expression
	case ast.IsElementAccessExpression(callee):
		rawObj = callee.AsElementAccessExpression().Expression
	default:
		return false
	}
	obj := ast.SkipOuterExpressions(rawObj, skipTransparent)
	if obj == nil {
		return false
	}
	if ast.IsIdentifier(obj) && obj.AsIdentifier().Text == "cy" {
		return true
	}
	return isCypress(obj)
}

var PreferAwaitToThenRule = rule.Rule{
	Name:   "promise/prefer-await-to-then",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callee := ast.SkipOuterExpressions(node.AsCallExpression().Expression, skipTransparent)
				if callee == nil {
					return
				}
				var nameNode *ast.Node
				switch {
				case ast.IsPropertyAccessExpression(callee):
					nameNode = callee.AsPropertyAccessExpression().Name()
				case ast.IsElementAccessExpression(callee):
					nameNode = ast.SkipOuterExpressions(callee.AsElementAccessExpression().ArgumentExpression, skipTransparent)
				default:
					return
				}
				if nameNode == nil || !ast.IsIdentifier(nameNode) {
					return
				}
				propName := nameNode.AsIdentifier().Text
				if propName != "then" && propName != "catch" && propName != "finally" {
					return
				}
				if isTopLevelScoped(node) ||
					(!opts.Strict && isInsideYieldOrAwait(node)) ||
					(!opts.Strict && isInsideConstructor(node)) ||
					isCypress(node) {
					return
				}
				ctx.ReportNode(nameNode, preferAwaitToCallbackMessage)
			},
		}
	},
}
