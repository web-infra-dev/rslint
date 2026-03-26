package no_hooks

import (
	"fmt"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message builders

func buildErrorUnexpectedHookMessage(hook string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedHook",
		Description: fmt.Sprintf("Unexpected hook: {{%s}}", hook),
	}
}

var allowedHooks = map[string]bool{
	"beforeEach": true,
	"afterEach":  true,
	"beforeAll":  true,
	"afterAll":   true,
}

type Options struct {
	Allow []string `json:"allow"`
}

func parseAllowList(raw any) []string {
	items, ok := raw.([]interface{})
	if !ok {
		return []string{}
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		s, ok := item.(string)
		if ok && allowedHooks[s] {
			out = append(out, s)
		}
	}
	return out
}

func parseOptions(options any) Options {
	opts := Options{Allow: []string{}}
	if options == nil {
		return opts
	}

	optArray, isArray := options.([]interface{})
	if !isArray || len(optArray) == 0 {
		return opts
	}
	optsMap, ok := optArray[0].(map[string]interface{})
	if !ok {
		return opts
	}

	if raw, ok := optsMap["allow"]; ok {
		opts.Allow = parseAllowList(raw)
	}
	return opts
}

var NoHooksRule = rule.Rule{
	Name: "jest/no-hooks",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx.TypeChecker)
				if jestFnCall == nil || jestFnCall.Kind != utils.JestFnTypeHook {
					return
				}

				if !slices.Contains(opts.Allow, jestFnCall.Name) {
					ctx.ReportNode(node, buildErrorUnexpectedHookMessage(jestFnCall.Name))
				}
			},
		}
	},
}
