package no_standalone_expect

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

//go:embed no_standalone_expect.schema.json
var schemaJSON []byte

type blockType string

const (
	blockTypeTest     blockType = "test"
	blockTypeFunction blockType = "function"
	blockTypeDescribe blockType = "describe"
	blockTypeArrow    blockType = "arrow"
	blockTypeTemplate blockType = "template"
)

type options struct {
	additionalTestBlockFunctions map[string]struct{}
}

type scopeEntry struct {
	kind  blockType
	start int
	end   int
}

func buildUnexpectedExpectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedExpect",
		Description: "Expect must be inside of a test block",
	}
}

func parseOptions(rawOptions []any) options {
	opts := options{additionalTestBlockFunctions: map[string]struct{}{}}
	if len(rawOptions) == 0 {
		return opts
	}

	optsMap, _ := rawOptions[0].(map[string]any)
	list, ok := optsMap["additionalTestBlockFunctions"].([]any)
	if !ok {
		return opts
	}

	for _, item := range list {
		if name, ok := item.(string); ok {
			opts.additionalTestBlockFunctions[name] = struct{}{}
		}
	}

	return opts
}

func getBlockType(block *ast.Node, analysis *rstestUtils.RstestCallAnalysis) blockType {
	fn := block.Parent
	if !testFramework.IsFunction(fn) {
		return ""
	}

	if ast.IsFunctionDeclaration(fn) {
		return blockTypeFunction
	}

	if fn.Parent != nil && fn.Parent.Kind == ast.KindVariableDeclaration {
		return blockTypeFunction
	}

	switch fn.Kind {
	case ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
		return blockTypeFunction
	}

	parent := fn.Parent
	if parent != nil && parent.Kind == ast.KindCallExpression {
		parsed := analysis.ParseFnCall(parent)
		if parsed != nil && parsed.Kind == rstestUtils.RstestFnTypeDescribe {
			return blockTypeDescribe
		}
	}

	return ""
}

func calleeIsTaggedTemplateExpression(node *ast.Node) bool {
	callExpr := node.AsCallExpression()
	return callExpr != nil &&
		callExpr.Expression != nil &&
		callExpr.Expression.Kind == ast.KindTaggedTemplateExpression
}

func getFunctionBodyRange(fn *ast.Node) (int, int, bool) {
	if fn == nil {
		return 0, 0, false
	}

	body := fn.Body()
	if body == nil {
		return 0, 0, false
	}

	return body.Pos(), body.End(), true
}

func getTestScopeRange(node *ast.Node) (int, int, bool) {
	callExpr := node.AsCallExpression()
	if callExpr == nil || callExpr.Arguments == nil {
		return 0, 0, false
	}

	for i := len(callExpr.Arguments.Nodes) - 1; i >= 0; i-- {
		if start, end, ok := getFunctionBodyRange(callExpr.Arguments.Nodes[i]); ok {
			return start, end, true
		}
	}

	return 0, 0, false
}

func getTemplateScopeRange(node *ast.Node) (int, int, bool) {
	callExpr := node.AsCallExpression()
	if callExpr == nil || callExpr.Expression == nil || callExpr.Expression.Kind != ast.KindTaggedTemplateExpression {
		return 0, 0, false
	}

	// The scope has to span the whole call, not just the tagged template that is
	// its callee: what a tagged template registers is the call it returns, so the
	// callback handed to that call is the part an `expect` can sit in.
	return node.Pos(), node.End(), true
}

func isStaticExpectCall(parsed *rstestUtils.ParsedRstestExpectCall) bool {
	if rstestUtils.IsStaticRstestExpectCall(parsed) {
		return true
	}

	// The shared helper deliberately treats every two-member static chain as
	// reportable. Rstest's typed `expect.not.<asymmetric>()` form is still a
	// value constructor rather than an assertion, so allow its known built-ins
	// without widening broken chains such as `expect.resolves.toBe()`.
	return parsed != nil &&
		parsed.Entry == rstestUtils.RstestExpectEntryStatic &&
		len(parsed.Modifiers) == 1 &&
		parsed.Modifiers[0] == "not" &&
		rstestUtils.RSTEST_ASYMMETRIC_MATCHERS[parsed.Matcher]
}

var NoStandaloneExpectRule = rule.Rule{
	Name:   "rstest/no-standalone-expect",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		scopeStack := make([]scopeEntry, 0, 8)

		isCustomTestBlockFunction := func(node *ast.Node) bool {
			callExpr := node.AsCallExpression()
			if callExpr == nil {
				return false
			}
			name := testFramework.CalleeChainName(callExpr.Expression)
			if name == "" {
				return false
			}
			_, ok := opts.additionalTestBlockFunctions[name]
			return ok
		}

		isTestBlockCall := func(node *ast.Node) bool {
			parsed := analysis.ParseFnCall(node)
			return parsed != nil && parsed.Kind == rstestUtils.RstestFnTypeTest || isCustomTestBlockFunction(node)
		}

		currentScopeKind := func() blockType {
			if len(scopeStack) == 0 {
				return ""
			}
			return scopeStack[len(scopeStack)-1].kind
		}

		pushScope := func(kind blockType, start, end int) {
			scopeStack = append(scopeStack, scopeEntry{kind: kind, start: start, end: end})
		}

		popScope := func(kind blockType) {
			if len(scopeStack) == 0 || scopeStack[len(scopeStack)-1].kind != kind {
				return
			}
			scopeStack = scopeStack[:len(scopeStack)-1]
		}

		getContainingBlock := func(node *ast.Node) blockType {
			pos := node.Pos()
			for i := len(scopeStack) - 1; i >= 0; i-- {
				scope := scopeStack[i]
				if scope.start <= pos && pos < scope.end {
					return scope.kind
				}
			}
			return ""
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsedExpect := analysis.ParseExpectCall(node)
				if parsedExpect != nil {
					if isStaticExpectCall(parsedExpect) {
						return
					}

					parent := getContainingBlock(node)
					if parent == "" || parent == blockTypeDescribe {
						ctx.ReportNode(node, buildUnexpectedExpectMessage())
					}
					return
				}

				if isTestBlockCall(node) {
					if start, end, ok := getTestScopeRange(node); ok {
						pushScope(blockTypeTest, start, end)
					}
				}

				if calleeIsTaggedTemplateExpression(node) {
					if start, end, ok := getTemplateScopeRange(node); ok {
						pushScope(blockTypeTemplate, start, end)
					}
				}
			},
			rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
				if calleeIsTaggedTemplateExpression(node) {
					popScope(blockTypeTemplate)
				}

				if isTestBlockCall(node) && currentScopeKind() == blockTypeTest {
					popScope(blockTypeTest)
				}
			},
			ast.KindBlock: func(node *ast.Node) {
				if kind := getBlockType(node, analysis); kind != "" {
					pushScope(kind, node.Pos(), node.End())
				}
			},
			rule.ListenerOnExit(ast.KindBlock): func(node *ast.Node) {
				kind := getBlockType(node, analysis)
				if kind != "" && currentScopeKind() == kind {
					popScope(kind)
				}
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				if node.Parent != nil && node.Parent.Kind != ast.KindCallExpression {
					if start, end, ok := getFunctionBodyRange(node); ok {
						pushScope(blockTypeArrow, start, end)
					}
				}
			},
			rule.ListenerOnExit(ast.KindArrowFunction): func(node *ast.Node) {
				if currentScopeKind() == blockTypeArrow {
					popScope(blockTypeArrow)
				}
			},
		}
	},
}
