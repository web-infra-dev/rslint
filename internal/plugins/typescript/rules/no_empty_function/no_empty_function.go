package no_empty_function

import (
	_ "embed"

	"github.com/web-infra-dev/rslint/internal/rule"
	core "github.com/web-infra-dev/rslint/internal/rules/no_empty_function"
)

//go:embed no_empty_function.schema.json
var schemaJSON []byte

// NoEmptyFunctionRule mirrors typescript-eslint's extension of ESLint's core
// no-empty-function rule. Upstream delegates every non-exempt function to the
// core rule; the Go implementation can share the core listener directly
// because it already understands TypeScript parameter properties, decorators,
// and override methods.
var NoEmptyFunctionRule = rule.CreateRule(rule.Rule{
	Name:   "no-empty-function",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return core.RunTSESLint(ctx, normalizeConstructorOptions(options))
	},
})

// normalizeConstructorOptions translates the two legacy option spellings
// retained by typescript-eslint into the camel-case names accepted by ESLint's
// core rule. It returns the original options unless a translation is needed.
func normalizeConstructorOptions(options []any) []any {
	if len(options) == 0 {
		return options
	}
	config, ok := options[0].(map[string]any)
	if !ok {
		return options
	}
	allow, ok := config["allow"].([]any)
	if !ok {
		return options
	}

	var normalizedAllow []any
	for i, item := range allow {
		name, ok := item.(string)
		if !ok {
			continue
		}
		var normalized string
		switch name {
		case "private-constructors":
			normalized = "privateConstructors"
		case "protected-constructors":
			normalized = "protectedConstructors"
		default:
			continue
		}
		if normalizedAllow == nil {
			normalizedAllow = append([]any(nil), allow...)
		}
		normalizedAllow[i] = normalized
	}
	if normalizedAllow == nil {
		return options
	}

	normalizedConfig := make(map[string]any, len(config))
	for key, value := range config {
		normalizedConfig[key] = value
	}
	normalizedConfig["allow"] = normalizedAllow
	normalizedOptions := append([]any(nil), options...)
	normalizedOptions[0] = normalizedConfig
	return normalizedOptions
}
