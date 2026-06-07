package no_jasmine_globals

import (
	"fmt"

	// cspell:ignore jestutils rslintutils
	"github.com/microsoft/typescript-go/shim/ast"
	jestutils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslintutils "github.com/web-infra-dev/rslint/internal/utils"
)

// Message builders

func buildIllegalGlobalMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "illegalGlobal",
		Description: fmt.Sprintf("Illegal usage of global `%s`, prefer `jest.spyOn`", name),
		Data: map[string]string{
			"global":      name,
			"replacement": "jest.spyOn",
		},
	}
}

func buildIllegalFailMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "illegalFail",
		Description: "Illegal usage of `fail`, prefer throwing an error, or the `done.fail` callback",
	}
}

func buildIllegalPendingMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "illegalPending",
		Description: "Illegal usage of `pending`, prefer explicitly skipping a test using `test.skip`",
	}
}

func buildIllegalMethodMessage(chain, replacement string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "illegalMethod",
		Description: fmt.Sprintf("Illegal usage of `%s`, prefer `%s`", chain, replacement),
		Data: map[string]string{
			"method":      chain,
			"replacement": replacement,
		},
	}
}

func buildIllegalJasmineMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "illegalJasmine",
		Description: "Illegal usage of jasmine global",
	}
}

func isBindingResolved(ident *ast.Node, ctx rule.RuleContext) bool {
	if ctx.TypeChecker == nil {
		return false
	}
	sym := ctx.TypeChecker.GetSymbolAtLocation(ident)
	return sym != nil
}

func jasmineRootCallChain(callee *ast.Node) (chain string, jasmineObj *ast.Node, tail string, entryCount int, ok bool) {
	entries := jestutils.GetJestFnMemberEntries(callee)
	if len(entries) < 2 || entries[0].Name != "jasmine" || entries[0].Node == nil {
		return "", nil, "", 0, false
	}

	rootParent := entries[0].Node.Parent
	if rootParent == nil || rslintutils.AccessExpressionObject(rootParent) != entries[0].Node {
		return "", nil, "", 0, false
	}

	// GetJestFnMemberEntries silently drops unsupported bracket accessors (e.g.
	// numeric indices). Confirm the parsed chain depth matches the AST so
	// jasmine[1].any() does not collapse to jasmine.any.
	expected := len(entries)
	actual := 1
	for n := callee; ; {
		obj := rslintutils.AccessExpressionObject(n)
		if obj == nil {
			break
		}
		actual++
		n = obj
	}
	if actual != expected {
		return "", nil, "", 0, false
	}

	return jestutils.JoinJestFnMemberEntries(entries), entries[0].Node, entries[len(entries)-1].Name, len(entries), true
}

func jasmineAssignedProperty(node *ast.Node) (jasmineObj *ast.Node, propName string, ok bool) {
	if node == nil {
		return nil, "", false
	}

	exp := ast.SkipParentheses(rslintutils.AccessExpressionObject(node))
	if exp == nil || exp.Kind != ast.KindIdentifier || exp.AsIdentifier().Text != "jasmine" {
		return nil, "", false
	}

	propName, ok = rslintutils.AccessExpressionStaticName(node)
	if !ok || propName == "" {
		return nil, "", false
	}

	return exp, propName, true
}

func isEcmaPrimitiveLiteral(kind ast.Kind) bool {
	switch kind {
	case ast.KindNumericLiteral, ast.KindStringLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
		ast.KindBigIntLiteral:
		return true
	default:
		return false
	}
}

func reportJasmineAssignedProperty(node *ast.Node, ctx rule.RuleContext) {
	jasmineObj, propName, ok := jasmineAssignedProperty(node)
	if !ok {
		return
	}
	if isBindingResolved(jasmineObj, ctx) {
		return
	}

	parent := node.Parent
	if parent != nil && rslintutils.AccessExpressionObject(parent) == node {
		return
	}
	if parent == nil || !ast.IsAssignmentExpression(parent, false) {
		if parent != nil && ast.IsCallExpression(parent) && parent.AsCallExpression().Expression == node {
			return
		}
		ctx.ReportNode(node, buildIllegalJasmineMessage())
		return
	}
	bin := parent.AsBinaryExpression()
	if bin == nil || bin.Left != node {
		return
	}

	if propName == "DEFAULT_TIMEOUT_INTERVAL" {
		if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken {
			right := ast.SkipParentheses(bin.Right)
			if right != nil && isEcmaPrimitiveLiteral(right.Kind) {
				ctx.ReportNodeWithFixes(
					node,
					buildIllegalJasmineMessage(),
					rule.RuleFixReplace(
						ctx.SourceFile,
						parent,
						fmt.Sprintf("jest.setTimeout(%s)", right.Text()),
					),
				)
				return
			}
		}
	}

	ctx.ReportNode(node, buildIllegalJasmineMessage())
}

var NoJasmineGlobalsRule = rule.Rule{
	Name: "jest/no-jasmine-globals",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				callee := ast.SkipParentheses(callExpr.Expression)
				if callee == nil {
					return
				}

				switch callee.Kind {
				case ast.KindIdentifier:
					name := callee.AsIdentifier().Text
					switch name {
					case "spyOn", "spyOnProperty", "fail", "pending":
						if isBindingResolved(callee, ctx) {
							return
						}
						switch name {
						case "spyOn", "spyOnProperty":
							ctx.ReportNode(node, buildIllegalGlobalMessage(name))
						case "fail":
							ctx.ReportNode(node, buildIllegalFailMessage())
						case "pending":
							ctx.ReportNode(node, buildIllegalPendingMessage())
						}
					}
					return
				case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
					chain, jasmineObj, tail, entryCount, ok := jasmineRootCallChain(callee)
					if !ok || isBindingResolved(jasmineObj, ctx) {
						return
					}

					switch tail {
					case "any", "anything", "arrayContaining", "objectContaining", "stringMatching":
						if entryCount != 2 {
							ctx.ReportNode(node, buildIllegalJasmineMessage())
							return
						}
						ctx.ReportNodeWithFixes(
							node,
							buildIllegalMethodMessage(chain, "expect."+tail),
							rule.RuleFixReplace(ctx.SourceFile, jasmineObj, "expect"),
						)
					case "addMatchers":
						if entryCount != 2 {
							ctx.ReportNode(node, buildIllegalJasmineMessage())
							return
						}
						ctx.ReportNode(node, buildIllegalMethodMessage(chain, "expect.extend"))
					case "createSpy":
						if entryCount != 2 {
							ctx.ReportNode(node, buildIllegalJasmineMessage())
							return
						}
						ctx.ReportNode(node, buildIllegalMethodMessage(chain, "jest.fn"))
					default:
						ctx.ReportNode(node, buildIllegalJasmineMessage())
					}
				}
			},
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				reportJasmineAssignedProperty(node, ctx)
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				reportJasmineAssignedProperty(node, ctx)
			},
			ast.KindIdentifier: func(node *ast.Node) {
				if node.AsIdentifier().Text != "jasmine" || isBindingResolved(node, ctx) {
					return
				}

				parent := node.Parent
				if parent != nil && (ast.IsPropertyAccessExpression(parent) || ast.IsElementAccessExpression(parent)) {
					return
				}

				ctx.ReportNode(node, buildIllegalJasmineMessage())
			},
		}
	},
}
