package expect_expect

import (
	"regexp"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalutils "github.com/web-infra-dev/rslint/internal/utils"
)

// Message Builder

func buildErrorNoAssertionsMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noAssertions",
		Description: "Test has no assertions",
	}
}

func parseOptions(options any) ([]string, []string) {
	assertNames := []string{"expect"}
	additional := []string{}

	m := internalutils.GetOptionsMap(options)
	if m == nil {
		return assertNames, additional
	}

	if raw, ok := m["assertFunctionNames"]; ok && raw != nil {
		if arr, ok := raw.([]interface{}); ok && len(arr) > 0 {
			out := make([]string, 0, len(arr))
			for _, v := range arr {
				if s, ok := v.(string); ok {
					out = append(out, s)
				}
			}
			if len(out) > 0 {
				assertNames = out
			}
		}
	}

	if raw, ok := m["additionalTestBlockFunctions"]; ok && raw != nil {
		if arr, ok := raw.([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					additional = append(additional, s)
				}
			}
		}
	}

	return assertNames, additional
}

func compileAssertPatterns(patterns []string) []*regexp.Regexp {
	out := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		out = append(out, compileAssertPattern(p))
	}
	return out
}

func compileAssertPattern(pattern string) *regexp.Regexp {
	segs := strings.Split(pattern, ".")
	parts := make([]string, 0, len(segs))
	for _, s := range segs {
		if s == "**" {
			parts = append(parts, `[a-z0-9\.]*`)
		} else {
			parts = append(parts, strings.ReplaceAll(s, "*", `[a-z0-9]*`))
		}
	}
	joined := strings.Join(parts, `\.`)
	// Segments follow eslint-plugin-jest: only `*` is expanded; other characters (e.g. `\$`)
	// are copied into the Regexp source like JavaScript's RegExp constructor.
	return regexp.MustCompile(`(?i)^(?:` + joined + `)(?:\.|$)`)
}

func matchesAssertName(name string, compiled []*regexp.Regexp) bool {
	if name == "" {
		return false
	}
	for _, re := range compiled {
		if re != nil && re.MatchString(name) {
			return true
		}
	}
	return false
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

func checkCallExpressionUsedFromCalls(unchecked *[]*ast.Node, calls []*ast.Node) {
	for _, call := range calls {
		if call == nil || call.Kind != ast.KindCallExpression {
			continue
		}
		if idx := indexUnchecked(*unchecked, call); idx >= 0 {
			*unchecked = slices.Delete(*unchecked, idx, idx+1)
			return
		}
	}
}

func findJestTestCallsPassingFunction(ctx rule.RuleContext, root *ast.Node, decl *ast.Node) []*ast.Node {
	var res []*ast.Node
	if decl == nil {
		return res
	}

	var declNameNode *ast.Node
	var declName string
	var declSymbol *ast.Symbol
	if decl.Kind == ast.KindFunctionDeclaration {
		fnDecl := decl.AsFunctionDeclaration()
		if fnDecl != nil && fnDecl.Name() != nil {
			declNameNode = fnDecl.Name()
			declName = declNameNode.Text()
			declSymbol = internalutils.GetReferenceSymbol(declNameNode, ctx.TypeChecker)
		}
	}
	if declName == "" {
		return res
	}

	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n == nil {
			return
		}
		if n.Kind == ast.KindCallExpression {
			if jestFn := utils.ParseJestFnCall(n, ctx); jestFn != nil && jestFn.Kind == utils.JestFnTypeTest {
				ce := n.AsCallExpression()
				if ce != nil && ce.Arguments != nil && len(ce.Arguments.Nodes) >= 2 {
					arg1 := ast.SkipParentheses(ce.Arguments.Nodes[1])
					if arg1 == nil || arg1.Kind != ast.KindIdentifier {
						goto walkChildren
					}
					if declSymbol != nil {
						if argSymbol := internalutils.GetReferenceSymbol(arg1, ctx.TypeChecker); argSymbol == declSymbol {
							res = append(res, n)
						}
						goto walkChildren
					}
					if arg1.AsIdentifier().Text == declName {
						res = append(res, n)
					}
				}
			}
		}
	walkChildren:
		n.ForEachChild(func(c *ast.Node) bool {
			walk(c)
			return false
		})
	}
	walk(root)
	return res
}

func checkCallExpressionUsed(ctx rule.RuleContext, assertNode *ast.Node, unchecked *[]*ast.Node, namedFnCallCache map[*ast.Node][]*ast.Node) {
	for n := assertNode.Parent; n != nil; n = n.Parent {
		if n.Kind == ast.KindFunctionDeclaration {
			decl := n.AsFunctionDeclaration()
			if decl != nil && decl.Name() != nil {
				declNode := decl.AsNode()
				calls, ok := namedFnCallCache[declNode]
				if !ok {
					calls = findJestTestCallsPassingFunction(ctx, ctx.SourceFile.AsNode(), declNode)
					namedFnCallCache[declNode] = calls
				}
				checkCallExpressionUsedFromCalls(unchecked, calls)
			}
		}
		if n.Kind == ast.KindCallExpression {
			if idx := indexUnchecked(*unchecked, n); idx >= 0 {
				*unchecked = slices.Delete(*unchecked, idx, idx+1)
				return
			}
		}
	}
}

var ExpectExpectRule = rule.Rule{
	Name: "jest/expect-expect",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		assertNames, additionalTestBlocks := parseOptions(options)
		compiled := compileAssertPatterns(assertNames)
		var unchecked []*ast.Node
		namedFnCallCache := map[*ast.Node][]*ast.Node{}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				callExpr := node.AsCallExpression()
				if callExpr == nil {
					return
				}

				calleeName := utils.CalleeChainName(callExpr.Expression)
				jestFn := utils.ParseJestFnCall(node, ctx)
				isJestTest := jestFn != nil && jestFn.Kind == utils.JestFnTypeTest
				isExtraBlock := calleeName != "" && slices.Contains(additionalTestBlocks, calleeName)

				if isJestTest || isExtraBlock {
					if isJestTest && isTodoTestCall(jestFn) {
						return
					}
					unchecked = append(unchecked, node)
					return
				}

				if !matchesAssertName(calleeName, compiled) {
					return
				}
				checkCallExpressionUsed(ctx, node, &unchecked, namedFnCallCache)
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
