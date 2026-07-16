package main

import (
	"encoding/json"
	"testing"
)

func TestCollectRuleSchemas(t *testing.T) {
	entries := collectRuleSchemas()

	byName := make(map[string]ruleSchemaEntry, len(entries))
	for _, e := range entries {
		byName[e.Name] = e
	}

	// Rules with a real, custom options schema.
	for _, name := range []string{"eqeqeq", "no-console"} {
		e, ok := byName[name]
		if !ok {
			t.Errorf("expected %q to be present with a declared schema", name)
			continue
		}
		var doc any
		if err := json.Unmarshal(e.Schema, &doc); err != nil {
			t.Errorf("rule %s: schema is not valid JSON: %v", name, err)
		}
	}

	// A rule declared with the shared EmptyArraySchema (no options) must
	// still show up — this is exactly the case a pure *.schema.json file
	// scan can't see, since no-debugger has no such file on disk.
	noDebugger, ok := byName["no-debugger"]
	if !ok {
		t.Fatal("expected no-debugger (EmptyArraySchema) to be present")
	}
	if string(noDebugger.Schema) != `{"type": "array", "maxItems": 0}` {
		t.Errorf("no-debugger: unexpected schema %s", noDebugger.Schema)
	}

	// A rule that hasn't been migrated to the schema framework at all
	// (Schema == nil) must be omitted, not present with a null/empty schema.
	if _, ok := byName["no-var"]; ok {
		t.Error("expected no-var (no declared Schema) to be omitted")
	}
}
