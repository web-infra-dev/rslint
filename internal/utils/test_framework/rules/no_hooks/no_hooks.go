package no_hooks

import (
	_ "embed"
	"fmt"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

//go:embed no_hooks.schema.json
var schemaJSON []byte

type Runtime struct {
	Parse func(*ast.Node) *ParsedCall
	Skip  bool
}

type Config struct {
	Name             string
	RequiresTypeInfo bool
	Prepare          func(rule.RuleContext) Runtime
}

type Options struct {
	Allow []string
}

type ParsedCall = testFramework.ParsedCall

func buildUnexpectedHookMessage(hook string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedHook",
		Description: fmt.Sprintf("Unexpected '%s' hook", hook),
	}
}

func parseAllowList(raw any) []string {
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		name, ok := item.(string)
		if ok && testFramework.IsHookName(name) {
			out = append(out, name)
		}
	}
	return out
}

func parseOptions(options []any) Options {
	opts := Options{Allow: []string{}}
	if len(options) == 0 {
		return opts
	}

	optsMap, _ := options[0].(map[string]any)
	if raw, ok := optsMap["allow"]; ok {
		opts.Allow = parseAllowList(raw)
	}
	return opts
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:             config.Name,
		RequiresTypeInfo: config.RequiresTypeInfo,
		Schema:           rule.NewSchema(schemaJSON),
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			if runtime.Skip {
				return rule.RuleListeners{}
			}
			opts := parseOptions(options)

			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					parsed := runtime.Parse(node)
					if parsed == nil || parsed.Kind != testFramework.FnKindHook {
						return
					}
					if slices.Contains(opts.Allow, parsed.Name) {
						return
					}
					ctx.ReportNode(node, buildUnexpectedHookMessage(parsed.Name))
				},
			}
		},
	}
}
