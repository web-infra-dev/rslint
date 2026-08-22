// Command dump_rule_schemas loads every native rule and dumps each one's
// name and options JSON Schema as JSON on stdout, straight from
// the native rule catalog — the single source of truth for rule
// IDs/prefixes and declared schemas. It's a build-time tool invocation for
// scripts/generate-rule-option-types.mjs, not part of the rslint CLI surface
// (see cmd/rslint), which is why it's a standalone command rather than a
// flag on that binary.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/web-infra-dev/rslint/internal/rules/catalog"
)

// ruleSchemaEntry is one catalogued rule's name and raw options schema.
type ruleSchemaEntry struct {
	Name   string          `json:"name"`
	Schema json.RawMessage `json:"schema"`
}

// collectRuleSchemas loads every native rule and returns the name +
// raw schema JSON for each one that declares a Schema. Every native rule
// declares one; the nil guard covers rules mounted without a Go schema, such
// as ESLint-plugin placeholders. The TypeScript side falls back to `any[]`
// for any rule ID it doesn't see here.
func collectRuleSchemas() []ruleSchemaEntry {
	rules := catalog.Native().AllRules()

	names := make([]string, 0, len(rules))
	for name := range rules {
		names = append(names, name)
	}
	sort.Strings(names)

	entries := make([]ruleSchemaEntry, 0, len(names))
	for _, name := range names {
		schema := rules[name].Schema
		if schema == nil {
			continue
		}
		entries = append(entries, ruleSchemaEntry{
			Name:   name,
			Schema: json.RawMessage(schema.RawJSON()),
		})
	}
	return entries
}

func main() {
	entries := collectRuleSchemas()
	if err := json.NewEncoder(os.Stdout).Encode(entries); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
