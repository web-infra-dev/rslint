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

// NOTE: eslint-plugin-jest tracks "am I inside a hook" as a single boolean and
// ignores every call made while it is set. That conflates two separate scopes:
// the boolean is cleared as soon as any nested hook call exits, so the run that
// surrounds the outer hook ends early, and hooks written inside a hook callback
// are compared against whatever index leaked in from the enclosing run.
//
// This body keeps one previousHookIndex per hook frame instead. Entering a hook
// pushes a fresh frame for its callback, exiting pops it, so a run is only ever
// compared against hooks declared at its own nesting level: the surrounding run
// survives nested callbacks, and nested runs are checked on their own terms.
func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			if runtime.Skip {
				return rule.RuleListeners{}
			}

			// previousHookIndexes[len-1] is the run currently being checked; a
			// value of -1 means the run has no hook to compare against yet.
			previousHookIndexes := []int{-1}
			// Hook calls that pushed a frame on entry, so exit pops exactly the
			// frames that were pushed even if parsing is not stable across the
			// two visits.
			enteredHooks := map[*ast.Node]bool{}

			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					top := len(previousHookIndexes) - 1
					parsed := runtime.Parse(node)

					if !testFramework.IsCallOfKind(parsed, testFramework.FnKindHook) {
						// Any other call ends the current run.
						previousHookIndexes[top] = -1
						return
					}

					enteredHooks[node] = true
					previousHookIndexes = append(previousHookIndexes, -1)

					currentHookIndex := testFramework.HookOrderIndex(parsed.Name)
					if currentHookIndex < 0 {
						previousHookIndexes[top] = -1
						return
					}

					if currentHookIndex < previousHookIndexes[top] {
						ctx.ReportNode(
							node,
							buildReorderHooksMessage(
								parsed.Name,
								testFramework.HooksOrder[previousHookIndexes[top]],
							),
						)
						return
					}

					previousHookIndexes[top] = currentHookIndex
				},
				rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
					if enteredHooks[node] {
						delete(enteredHooks, node)
						previousHookIndexes = previousHookIndexes[:len(previousHookIndexes)-1]
						return
					}

					previousHookIndexes[len(previousHookIndexes)-1] = -1
				},
			}
		},
	}
}
