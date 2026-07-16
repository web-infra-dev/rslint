// Command dump_rule_schemas registers every native rule and dumps each one's
// name and options JSON Schema as JSON on stdout, straight from
// internal/config.GlobalRuleRegistry — the single source of truth for rule
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

	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

// ruleSchemaEntry is one registered rule's name and raw options schema.
type ruleSchemaEntry struct {
	Name   string          `json:"name"`
	Schema json.RawMessage `json:"schema"`
}

// collectRuleSchemas registers every native rule and returns the name +
// raw schema JSON for each one that declares a Schema (nil-Schema rules —
// not yet migrated to the schema framework — are omitted; the TypeScript
// side falls back to `any[]` for any rule ID it doesn't see here).
func collectRuleSchemas() []ruleSchemaEntry {
	rslintconfig.RegisterAllRules()
	rules := rslintconfig.GlobalRuleRegistry.GetAllRules()

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
