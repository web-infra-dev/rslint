package no_duplicate_hooks

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type ParsedCall = testFramework.ParsedCall

type Runtime struct {
	Parse func(*ast.Node) *ParsedCall
}

type Config struct {
	Name    string
	Prepare func(rule.RuleContext) Runtime
}

func duplicateHookMessage(hook string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noDuplicateHook",
		Description: "Duplicate " + hook + " in describe block",
		Data:        map[string]string{"hook": hook},
	}
}

// NewRule owns the lexical describe stack and duplicate counting. Framework
// adapters supply only call provenance and semantic names.
func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			hookContexts := []map[string]int{{}}
			enteredDescribes := map[*ast.Node]bool{}

			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					parsed := runtime.Parse(node)
					if testFramework.IsCallOfKind(parsed, testFramework.FnKindDescribe) {
						enteredDescribes[node] = true
						hookContexts = append(hookContexts, map[string]int{})
						return
					}
					if !testFramework.IsCallOfKind(parsed, testFramework.FnKindHook) {
						return
					}

					current := hookContexts[len(hookContexts)-1]
					current[parsed.Name]++
					if current[parsed.Name] > 1 {
						ctx.ReportNode(node, duplicateHookMessage(parsed.Name))
					}
				},
				rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
					if !enteredDescribes[node] {
						return
					}
					delete(enteredDescribes, node)
					hookContexts = hookContexts[:len(hookContexts)-1]
				},
			}
		},
	}
}
