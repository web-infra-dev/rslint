package rule

import "testing"

func TestCompileSchemaAndValidateOptions(t *testing.T) {
	// Mirrors the <rule-name>.schema.json convention: a tuple describing the
	// ESLint-style options array, here a single optional object with an
	// `allow: string[]` property (à la no-console).
	schemaJSON := []byte(`{
		"type": "array",
		"items": [
			{
				"type": "object",
				"properties": {
					"allow": {
						"type": "array",
						"items": { "type": "string" }
					}
				},
				"additionalProperties": false
			}
		],
		"minItems": 0,
		"maxItems": 1
	}`)

	schema, err := CompileSchema(schemaJSON)
	if err != nil {
		t.Fatalf("CompileSchema failed: %v", err)
	}

	// No options at all: satisfies minItems: 0.
	if err := ValidateOptions(schema, []any{}); err != nil {
		t.Errorf("expected empty options to be valid, got: %v", err)
	}

	// A well-formed options object.
	valid := []any{map[string]any{"allow": []any{"warn", "error"}}}
	if err := ValidateOptions(schema, valid); err != nil {
		t.Errorf("expected valid options, got: %v", err)
	}

	// `allow` must be an array of strings, not a bare string.
	invalidType := []any{map[string]any{"allow": "warn"}}
	if err := ValidateOptions(schema, invalidType); err == nil {
		t.Error("expected error for allow as non-array, got nil")
	}

	// Unknown property rejected by additionalProperties: false.
	invalidProp := []any{map[string]any{"unknown": true}}
	if err := ValidateOptions(schema, invalidProp); err == nil {
		t.Error("expected error for unknown property, got nil")
	}

	// Exceeds maxItems: 1.
	tooMany := []any{
		map[string]any{"allow": []any{"warn"}},
		map[string]any{"allow": []any{"error"}},
	}
	if err := ValidateOptions(schema, tooMany); err == nil {
		t.Error("expected error for too many options, got nil")
	}
}

func TestCompileSchemaNoOptions(t *testing.T) {
	// The no-options convention: omit `items` (draft-04's own metaschema
	// requires a non-empty schemaArray, so `"items": []` is itself invalid)
	// and rely on `maxItems: 0` to force an empty options array.
	schemaJSON := []byte(`{"type": "array", "maxItems": 0}`)

	schema, err := CompileSchema(schemaJSON)
	if err != nil {
		t.Fatalf("CompileSchema failed: %v", err)
	}

	if err := ValidateOptions(schema, []any{}); err != nil {
		t.Errorf("expected empty options to be valid, got: %v", err)
	}
	if err := ValidateOptions(schema, []any{"unexpected"}); err == nil {
		t.Error("expected error when options are passed to a no-options rule, got nil")
	}
}

func TestCompileSchemaInvalidJSON(t *testing.T) {
	if _, err := CompileSchema([]byte("not json")); err == nil {
		t.Error("expected error for invalid schema JSON, got nil")
	}
}
