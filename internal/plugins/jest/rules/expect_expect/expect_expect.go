package expect_expect

import (
	"regexp"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
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

	m := internalUtils.GetOptionsMap(options)
	if m == nil {
		return assertNames, additional
	}

	if raw, ok := m["assertFunctionNames"]; ok && raw != nil {
		if arr, ok := raw.([]interface{}); ok {
			out := make([]string, 0, len(arr))
			for _, v := range arr {
				if s, ok := v.(string); ok {
					out = append(out, s)
				}
			}
			assertNames = out
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
			parts = append(parts, `[a-zA-Z0-9.]*`)
		} else {
			parts = append(parts, strings.ReplaceAll(s, "*", `[a-zA-Z0-9]*`))
		}
	}
	joined := strings.Join(parts, `\.`)
	// Segments follow eslint-plugin-jest: only `*` is expanded; other characters (e.g. `\$`)
	// are copied into the Regexp source like JavaScript's RegExp constructor.
	re, err := regexp.Compile(`(?i)^(?:` + joined + `)(?:\.|$)`)
	if err != nil {
		return nil
	}
	return re
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
	Name: "jest/expect-expect",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		assertNames, additionalTestBlocks := parseOptions(options)
		compiled := compileAssertPatterns(assertNames)
		var unchecked []*ast.Node
		uncheckedByDecl := map[*ast.Node][]*ast.Node{}
		uncheckedByName := map[string][]*ast.Node{}

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

				if !matchesAssertName(calleeName, compiled) {
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
