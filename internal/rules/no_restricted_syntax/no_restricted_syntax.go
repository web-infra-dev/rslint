package no_restricted_syntax

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoRestrictedSyntaxRule is the rslint port of ESLint's `no-restricted-syntax`.
//
// Configuration mirrors ESLint: the rule accepts a list of strings or
// `{ selector, message? }` objects. Each selector follows the esquery
// grammar (the subset implemented in parser.go). At lint time every node
// is matched against every parsed selector; each match produces one
// diagnostic with the user-supplied message (or a default).
//
// https://eslint.org/docs/latest/rules/no-restricted-syntax
var NoRestrictedSyntaxRule = rule.Rule{
	Name: "no-restricted-syntax",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		entries := parseRuleOptions(options)
		if len(entries) == 0 {
			return rule.RuleListeners{}
		}

		mc := &matchContext{sf: ctx.SourceFile}

		// Build per-kind buckets of (selector, message) entries. Selectors
		// whose `candidate kinds` is the universe land in everyKindBucket
		// and are evaluated on every visit.
		perKind := make(map[ast.Kind][]ruleEntry)
		var everyKind []ruleEntry
		for _, e := range entries {
			ks := candidateKinds(e.compiled)
			if ks.universe {
				everyKind = append(everyKind, e)
				continue
			}
			for k := range ks.kinds {
				perKind[k] = append(perKind[k], e)
			}
		}

		// Merge the universe-of-kinds bucket into every kind we listen on.
		// allInterestingKinds is the set we register for `*` / pure
		// attribute selectors; kinds outside it that already have a
		// per-kind bucket still keep their bucket (the universe matchers
		// just don't apply there).
		if len(everyKind) > 0 {
			for _, k := range allInterestingKinds {
				perKind[k] = append(perKind[k], everyKind...)
			}
		}

		visit := func(node *ast.Node, bucket []ruleEntry) {
			for _, e := range bucket {
				if matches(e.compiled, node, mc) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "restrictedSyntax",
						Description: e.formatMessage(),
					})
				}
			}
		}

		listeners := rule.RuleListeners{}
		for k, bucket := range perKind {
			listeners[k] = func(node *ast.Node) {
				visit(node, bucket)
			}
		}
		return listeners
	},
}

// ruleEntry is a parsed option entry: the original selector string,
// the compiled selector tree, and the message (custom or default).
type ruleEntry struct {
	selector string
	compiled selector
	message  string
}

func (e ruleEntry) formatMessage() string {
	if e.message != "" {
		return e.message
	}
	return fmt.Sprintf("Using '%s' is not allowed.", e.selector)
}

// parseRuleOptions normalises the loosely-typed options value handed to
// the rule into a list of ruleEntry. ESLint accepts a mix of strings and
// `{ selector, message? }` objects; rslint receives a single string, a
// single object, or an []interface{} depending on how config.go unwrapped
// the array. Selectors that fail to parse are silently dropped — ESLint
// rejects the whole config in that case, but for runtime resilience we
// prefer to drop the offending entry over panicking.
func parseRuleOptions(opts any) []ruleEntry {
	if opts == nil {
		return nil
	}
	var entries []ruleEntry
	switch v := opts.(type) {
	case string:
		entries = append(entries, buildEntryFromString(v))
	case map[string]interface{}:
		if e, ok := buildEntryFromObject(v); ok {
			entries = append(entries, e)
		}
	case []interface{}:
		for _, item := range v {
			switch x := item.(type) {
			case string:
				entries = append(entries, buildEntryFromString(x))
			case map[string]interface{}:
				if e, ok := buildEntryFromObject(x); ok {
					entries = append(entries, e)
				}
			}
		}
	}
	out := make([]ruleEntry, 0, len(entries))
	for _, e := range entries {
		if e.compiled != nil {
			out = append(out, e)
		}
	}
	return out
}

func buildEntryFromString(sel string) ruleEntry {
	compiled, err := parseSelector(sel)
	if err != nil {
		return ruleEntry{selector: sel}
	}
	return ruleEntry{selector: sel, compiled: compiled}
}

func buildEntryFromObject(m map[string]interface{}) (ruleEntry, bool) {
	rawSel, _ := m["selector"].(string)
	if rawSel == "" {
		return ruleEntry{}, false
	}
	compiled, err := parseSelector(rawSel)
	if err != nil {
		return ruleEntry{}, false
	}
	msg, _ := m["message"].(string)
	return ruleEntry{selector: rawSel, compiled: compiled, message: msg}, true
}
