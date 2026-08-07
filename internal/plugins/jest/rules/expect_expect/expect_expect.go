package expect_expect

import (
	_ "embed"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

//go:embed expect_expect.schema.json
var schemaJSON []byte

// Message Builder

func buildErrorNoAssertionsMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noAssertions",
		Description: "Test has no assertions",
	}
}

func isTodoTestCall(jestFn *utils.ParsedJestFnCall) bool {
	if jestFn == nil || jestFn.Kind != utils.JestFnTypeTest {
		return false
	}
	return len(jestFn.Members) > 0 && jestFn.Members[len(jestFn.Members)-1] == "todo"
}

func indexUnchecked(unchecked []*ast.Node, call *ast.Node) int {
	for i, c := range unchecked {
		if c == call {
			return i
		}
	}
	return -1
}

func removeUncheckedCall(unchecked *[]*ast.Node, call *ast.Node) bool {
	if idx := indexUnchecked(*unchecked, call); idx >= 0 {
		*unchecked = slices.Delete(*unchecked, idx, idx+1)
		return true
	}
	return false
}

func clearUncheckedCalls(unchecked *[]*ast.Node, calls []*ast.Node) {
	for _, call := range calls {
		if call == nil || call.Kind != ast.KindCallExpression {
			continue
		}
		removeUncheckedCall(unchecked, call)
	}
}

func trackNamedFunctionTestCall(
	ctx rule.RuleContext,
	callNode *ast.Node,
	callExpr *ast.CallExpression,
	uncheckedByDecl map[*ast.Node][]*ast.Node,
	uncheckedByName map[string][]*ast.Node,
) {
	declNode, fnName := utils.ResolveNamedFunctionCallback(ctx, callExpr)
	switch {
	case declNode != nil:
		uncheckedByDecl[declNode] = append(uncheckedByDecl[declNode], callNode)
	case fnName != "":
		uncheckedByName[fnName] = append(uncheckedByName[fnName], callNode)
	}
}

func checkCallExpressionUsed(
	assertNode *ast.Node,
	unchecked *[]*ast.Node,
	uncheckedByDecl map[*ast.Node][]*ast.Node,
	uncheckedByName map[string][]*ast.Node,
) {
	var ancestors []*ast.Node
	for n := assertNode.Parent; n != nil; n = n.Parent {
		ancestors = append(ancestors, n)
	}

	for i := len(ancestors) - 1; i >= 0; i-- {
		n := ancestors[i]
		if n.Kind == ast.KindFunctionDeclaration {
			decl := n.AsFunctionDeclaration()
			if decl != nil && decl.Name() != nil {
				declNode := decl.AsNode()
				fnName := decl.Name().Text()

				clearUncheckedCalls(unchecked, uncheckedByDecl[declNode])
				delete(uncheckedByDecl, declNode)

				clearUncheckedCalls(unchecked, uncheckedByName[fnName])
				delete(uncheckedByName, fnName)
			}
		}
		if n.Kind == ast.KindCallExpression {
			if removeUncheckedCall(unchecked, n) {
				return
			}
		}
	}
}

var ExpectExpectRule = rule.Rule{
	Name:   "jest/expect-expect",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		parsedOptions := utils.ParseAssertionFunctionOptions(options)
		additionalTestBlocks := parsedOptions.AdditionalTestBlockFunctions
		compiled := utils.CompileAssertFunctionNamePatterns(parsedOptions.AssertFunctionNames)
		var unchecked []*ast.Node
		uncheckedByDecl := map[*ast.Node][]*ast.Node{}
		uncheckedByName := map[string][]*ast.Node{}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				if callExpr == nil {
					return
				}

				calleeName := testFramework.CalleeChainName(callExpr.Expression)
				jestFn := utils.ParseJestFnCall(node, ctx)
				isJestTest := jestFn != nil && jestFn.Kind == utils.JestFnTypeTest
				isExtraBlock := calleeName != "" && slices.Contains(additionalTestBlocks, calleeName)

				if isJestTest || isExtraBlock {
					if isTodoTestCall(jestFn) || strings.HasSuffix(calleeName, ".todo") {
						return
					}
					if isJestTest {
						trackNamedFunctionTestCall(
							ctx,
							node,
							callExpr,
							uncheckedByDecl,
							uncheckedByName,
						)
					}
					unchecked = append(unchecked, node)
					return
				}

				if !utils.MatchesAssertFunctionName(calleeName, compiled) {
					return
				}
				checkCallExpressionUsed(
					node,
					&unchecked,
					uncheckedByDecl,
					uncheckedByName,
				)
			},
			rule.ListenerOnExit(ast.KindEndOfFile): func(node *ast.Node) {
				_ = node
				for _, call := range unchecked {
					ce := call.AsCallExpression()
					if ce != nil && ce.Expression != nil {
						ctx.ReportNode(ce.Expression, buildErrorNoAssertionsMessage())
					}
				}
			},
		}
	},
}
