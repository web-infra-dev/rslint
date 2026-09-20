package rule

import (
	"maps"
	"slices"
	"strings"
)

// Catalog is an immutable, name-indexed snapshot of the rules available to a
// lint run. A catalog can be shared by concurrent resolvers and lint workers.
// Deriving a catalog for one ESLint-plugin set does not change the receiver.
type Catalog struct {
	rules                map[string]Rule
	ruleNamesByNamespace map[string][]string
	hasESLintPluginRules bool
}

// NewCatalog builds a rule catalog. When the input contains duplicate names,
// the last rule wins, matching the existing rule-assembly semantics.
func NewCatalog(rules ...Rule) *Catalog {
	byName := make(map[string]Rule, len(rules))
	for _, ruleImpl := range rules {
		byName[ruleImpl.Name] = ruleImpl
	}
	return newCatalog(byName)
}

// Namespace returns a rule's plugin namespace using ESLint's rule-ID syntax.
// Unscoped IDs split at the first slash, leaving any further slashes in the
// rule name. IDs starting with @ split at the last slash to support scoped
// plugin namespaces. Core rules have an empty namespace.
func Namespace(ruleName string) string {
	separator := strings.IndexByte(ruleName, '/')
	if strings.HasPrefix(ruleName, "@") {
		separator = strings.LastIndexByte(ruleName, '/')
	}
	if separator >= 0 {
		return ruleName[:separator]
	}
	return ""
}

func newCatalog(rules map[string]Rule) *Catalog {
	ruleNamesByNamespace := make(map[string][]string)
	hasESLintPluginRules := false
	for ruleName, ruleImpl := range rules {
		namespace := Namespace(ruleName)
		ruleNamesByNamespace[namespace] = append(ruleNamesByNamespace[namespace], ruleName)
		hasESLintPluginRules = hasESLintPluginRules || ruleImpl.IsEslintPluginRule
	}
	for namespace := range ruleNamesByNamespace {
		slices.Sort(ruleNamesByNamespace[namespace])
	}
	return &Catalog{
		rules:                rules,
		ruleNamesByNamespace: ruleNamesByNamespace,
		hasESLintPluginRules: hasESLintPluginRules,
	}
}

// ForESLintPlugins derives a catalog for exactly the supplied object-form
// ESLint plugins. Existing plugin placeholders are replaced, while Go rules
// remain unchanged. A Go rule wins a name collision; shadowed names are
// returned in input order for caller-facing warnings.
func (c *Catalog) ForESLintPlugins(plugins []ESLintPluginMetadata) (*Catalog, []string) {
	if c == nil {
		panic("rule catalog is required")
	}
	if len(plugins) == 0 && !c.hasESLintPluginRules {
		return c, nil
	}

	byName := make(map[string]Rule, len(c.rules))
	for ruleName, ruleImpl := range c.rules {
		if !ruleImpl.IsEslintPluginRule {
			byName[ruleName] = ruleImpl
		}
	}
	var shadowed []string
	for _, plugin := range plugins {
		if plugin.Prefix == "" {
			continue
		}
		for _, ruleName := range plugin.RuleNames {
			fullName := plugin.Prefix + "/" + ruleName
			// A different split can produce the same full ID. Only its parsed
			// namespace may supply the rule, matching ESLint's plugin lookup.
			if Namespace(fullName) != plugin.Prefix {
				continue
			}
			if existing, ok := byName[fullName]; ok {
				if !existing.IsEslintPluginRule {
					shadowed = append(shadowed, fullName)
				}
				continue
			}
			byName[fullName] = Rule{
				Name:               fullName,
				IsEslintPluginRule: true,
				Run:                runESLintPluginPlaceholder,
			}
		}
	}
	return newCatalog(byName), shadowed
}

func runESLintPluginPlaceholder(RuleContext, []any) RuleListeners { return RuleListeners{} }

// RuleNamesForNamespace returns a sorted copy of the rule names in namespace.
// Core rules use the empty namespace.
func (c *Catalog) RuleNamesForNamespace(namespace string) []string {
	if c == nil {
		panic("rule catalog is required")
	}
	return slices.Clone(c.ruleNamesByNamespace[namespace])
}

// Lookup returns the rule available under name.
func (c *Catalog) Lookup(name string) (Rule, bool) {
	if c == nil {
		panic("rule catalog is required")
	}
	ruleImpl, ok := c.rules[name]
	return ruleImpl, ok
}

// AllRules returns a copy of the catalog's name-indexed rules.
func (c *Catalog) AllRules() map[string]Rule {
	if c == nil {
		panic("rule catalog is required")
	}
	return maps.Clone(c.rules)
}

// ESLintPluginMetadata describes the rules exposed by one object-form ESLint plugin.
// The live plugin stays in Node; Go only needs the prefix and rule names to
// resolve configured rules and route them to the plugin host.
type ESLintPluginMetadata struct {
	Prefix    string   `json:"prefix"`
	RuleNames []string `json:"ruleNames"`
}
