package prefer_hooks_in_order

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type ParsedCall = testFramework.ParsedCall

type Runtime struct {
	Parse func(*ast.Node) *testFramework.ParsedCall
	Skip  bool
}

type Config struct {
	Name    string
	Prepare func(rule.RuleContext) Runtime
}

func buildReorderHooksMessage(currentHook, previousHook string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "reorderHooks",
		Description: "`" + currentHook + "` hooks should be before any `" + previousHook + "` hooks",
		Data: map[string]string{
			"currentHook":  currentHook,
			"previousHook": previousHook,
		},
	}
}

// NOTE: Unlike eslint-plugin-jest's bool-based inHook flag, this shared body
// tracks hook call depth so nested hook callbacks cannot accidentally clear the
// outer hook frame on exit. This preserves call-event continuity across hook
// callback internals instead of ending the outer run early.
func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			if runtime.Skip {
				return rule.RuleListeners{}
			}

			previousHookIndex := -1
			hookCallDepth := 0
			enteredHooks := map[*ast.Node]bool{}
			parsedCalls := map[*ast.Node]*testFramework.ParsedCall{}

			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					parsed := runtime.Parse(node)
					parsedCalls[node] = parsed

					if hookCallDepth > 0 {
						if testFramework.IsCallOfKind(parsed, testFramework.FnKindHook) {
							enteredHooks[node] = true
							hookCallDepth++
						}
						return
					}

					if !testFramework.IsCallOfKind(parsed, testFramework.FnKindHook) {
						previousHookIndex = -1
						return
					}

					enteredHooks[node] = true
					hookCallDepth++

					currentHookIndex := testFramework.HookOrderIndex(parsed.Name)
					if currentHookIndex < 0 {
						previousHookIndex = -1
						return
					}

					if currentHookIndex < previousHookIndex {
						ctx.ReportNode(
							node,
							buildReorderHooksMessage(
								parsed.Name,
								testFramework.HooksOrder[previousHookIndex],
							),
						)
						return
					}

					previousHookIndex = currentHookIndex
				},
				rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
					parsed := parsedCalls[node]
					delete(parsedCalls, node)

					if enteredHooks[node] && testFramework.IsCallOfKind(parsed, testFramework.FnKindHook) {
						delete(enteredHooks, node)
						if hookCallDepth > 0 {
							hookCallDepth--
						}
						return
					}

					if hookCallDepth > 0 {
						return
					}

					previousHookIndex = -1
				},
			}
		},
	}
}
