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

// runDumpRuleSchemas implements the hidden `--dump-rule-schemas` flag: it
// dumps every registered native rule's name and options JSON Schema as JSON
// on stdout, for packages/rslint/scripts/generate-rule-option-types.mjs to
// compile into TypeScript types via json-schema-to-typescript. Deliberately
// left out of `usage` — it's a build-time tool invocation, not a supported
// end-user CLI mode.
func runDumpRuleSchemas() int {
	entries := collectRuleSchemas()
	if err := json.NewEncoder(os.Stdout).Encode(entries); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}
