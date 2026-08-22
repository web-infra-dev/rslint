package rule

import (
	"slices"
	"strings"
)

// Catalog is an immutable, name-indexed snapshot of the rules available to a
// lint run. A catalog can be shared by concurrent resolvers and lint workers.
// Use WithRules to derive a new snapshot without changing the original.
type Catalog struct {
	rules                map[string]Rule
	ruleNamesByNamespace map[string][]string
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

func newCatalog(rules map[string]Rule) *Catalog {
	ruleNamesByNamespace := make(map[string][]string)
	for ruleName := range rules {
		namespace := ""
		if separator := strings.LastIndex(ruleName, "/"); separator >= 0 {
			namespace = ruleName[:separator]
		}
		ruleNamesByNamespace[namespace] = append(ruleNamesByNamespace[namespace], ruleName)
	}
	return &Catalog{
		rules:                rules,
		ruleNamesByNamespace: ruleNamesByNamespace,
	}
}

// WithRules returns a catalog containing the receiver's rules plus the given
// rules. The receiver is returned unchanged when rules is empty.
func (c *Catalog) WithRules(rules ...Rule) *Catalog {
	if c == nil {
		panic("rule catalog is required")
	}
	if len(rules) == 0 {
		return c
	}

	byName := make(map[string]Rule, len(c.rules)+len(rules))
	for name, ruleImpl := range c.rules {
		byName[name] = ruleImpl
	}
	for _, ruleImpl := range rules {
		byName[ruleImpl.Name] = ruleImpl
	}
	return newCatalog(byName)
}

// RuleNamesForNamespace returns a copy of the rule names in namespace. Core
// rules use the empty namespace.
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
	rules := make(map[string]Rule, len(c.rules))
	for name, ruleImpl := range c.rules {
		rules[name] = ruleImpl
	}
	return rules
}

// ESLintPluginMetadata describes the rules exposed by one object-form ESLint plugin.
// The live plugin stays in Node; Go only needs the prefix and rule names to
// resolve configured rules and route them to the plugin host.
type ESLintPluginMetadata struct {
	Prefix    string   `json:"prefix"`
	RuleNames []string `json:"ruleNames"`
}
