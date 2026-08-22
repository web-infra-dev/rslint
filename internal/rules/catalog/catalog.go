// Package catalog assembles the native rule catalog and derives run-scoped
// catalogs for object-form ESLint plugins.
package catalog

import (
	"sync"

	importPlugin "github.com/web-infra-dev/rslint/internal/plugins/import"
	jestPlugin "github.com/web-infra-dev/rslint/internal/plugins/jest"
	jsxA11yPlugin "github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y"
	promisePlugin "github.com/web-infra-dev/rslint/internal/plugins/promise"
	reactPlugin "github.com/web-infra-dev/rslint/internal/plugins/react"
	reactHooksPlugin "github.com/web-infra-dev/rslint/internal/plugins/react_hooks"
	rstestPlugin "github.com/web-infra-dev/rslint/internal/plugins/rstest"
	typescriptPlugin "github.com/web-infra-dev/rslint/internal/plugins/typescript"
	unicornPlugin "github.com/web-infra-dev/rslint/internal/plugins/unicorn"
	"github.com/web-infra-dev/rslint/internal/rule"
	coreRules "github.com/web-infra-dev/rslint/internal/rules"
)

type nativeRuleCollection struct {
	namespace string
	allRules  func() []rule.Rule
}

// nativeRuleCollections is the single ordered inventory of concrete native
// rule sources. Core rules intentionally remain last to preserve the previous
// assembly order.
var nativeRuleCollections = []nativeRuleCollection{
	{namespace: "@typescript-eslint", allRules: typescriptPlugin.GetAllRules},
	{namespace: "import", allRules: importPlugin.GetAllRules},
	{namespace: "react", allRules: reactPlugin.GetAllRules},
	{namespace: "react-hooks", allRules: reactHooksPlugin.GetAllRules},
	{namespace: "jest", allRules: jestPlugin.GetAllRules},
	{namespace: "rstest", allRules: rstestPlugin.GetAllRules},
	{namespace: "jsx-a11y", allRules: jsxA11yPlugin.GetAllRules},
	{namespace: "promise", allRules: promisePlugin.GetAllRules},
	{namespace: "unicorn", allRules: unicornPlugin.GetAllRules},
	{namespace: "", allRules: coreRules.GetAllRules},
}

var nativeCatalog = sync.OnceValue(func() *rule.Catalog {
	var rules []rule.Rule
	for _, collection := range nativeRuleCollections {
		rules = append(rules, collection.allRules()...)
	}
	return rule.NewCatalog(rules...)
})

// Native returns the immutable catalog of rules implemented in Go. The
// catalog is safe to share across concurrent lint runs.
func Native() *rule.Catalog {
	return nativeCatalog()
}

// WithESLintPlugins derives a catalog containing placeholders for the given
// object-form ESLint plugin rules. Built-in rules take precedence; their names
// are returned in input order so an adapter can report the existing warning.
func WithESLintPlugins(base *rule.Catalog, plugins []rule.ESLintPluginMetadata) (*rule.Catalog, []string) {
	if base == nil {
		panic("base rule catalog is required")
	}
	if len(plugins) == 0 {
		return base, nil
	}

	var placeholders []rule.Rule
	var shadowed []string
	for _, plugin := range plugins {
		if plugin.Prefix == "" {
			continue
		}
		for _, ruleName := range plugin.RuleNames {
			fullName := plugin.Prefix + "/" + ruleName
			if existing, ok := base.Lookup(fullName); ok && !existing.IsEslintPluginRule {
				shadowed = append(shadowed, fullName)
				continue
			}
			placeholders = append(placeholders, rule.Rule{
				Name:               fullName,
				IsEslintPluginRule: true,
				Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
					return rule.RuleListeners{}
				},
			})
		}
	}
	return base.WithRules(placeholders...), shadowed
}
