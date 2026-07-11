package rule

import (
	"sync"
	"testing"
)

func TestSchemaValidateOptions(t *testing.T) {
	// Mirrors the <rule-name>.schema.json convention: a tuple describing the
	// ESLint-style options array, here a single optional object with an
	// `allow: string[]` property (à la no-console).
	schema := NewSchema([]byte(`{
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
	}`))

	// No options at all: satisfies minItems: 0.
	if err := schema.Validate([]any{}); err != nil {
		t.Errorf("expected empty options to be valid, got: %v", err)
	}

	// A nil slice must count as an empty options array, not JSON null.
	if err := schema.Validate(nil); err != nil {
		t.Errorf("expected nil options to be valid, got: %v", err)
	}

	// A well-formed options object.
	valid := []any{map[string]any{"allow": []any{"warn", "error"}}}
	if err := schema.Validate(valid); err != nil {
		t.Errorf("expected valid options, got: %v", err)
	}

	// `allow` must be an array of strings, not a bare string.
	invalidType := []any{map[string]any{"allow": "warn"}}
	if err := schema.Validate(invalidType); err == nil {
		t.Error("expected error for allow as non-array, got nil")
	}

	// Unknown property rejected by additionalProperties: false.
	invalidProp := []any{map[string]any{"unknown": true}}
	if err := schema.Validate(invalidProp); err == nil {
		t.Error("expected error for unknown property, got nil")
	}

	// Exceeds maxItems: 1.
	tooMany := []any{
		map[string]any{"allow": []any{"warn"}},
		map[string]any{"allow": []any{"error"}},
	}
	if err := schema.Validate(tooMany); err == nil {
		t.Error("expected error for too many options, got nil")
	}
}

func TestSchemaNoOptions(t *testing.T) {
	// The no-options convention: omit `items` (draft-04's own metaschema
	// requires a non-empty schemaArray, so `"items": []` is itself invalid)
	// and rely on `maxItems: 0` to force an empty options array.
	schema := NewSchema([]byte(`{"type": "array", "maxItems": 0}`))

	if err := schema.Validate([]any{}); err != nil {
		t.Errorf("expected empty options to be valid, got: %v", err)
	}
	if err := schema.Validate([]any{"unexpected"}); err == nil {
		t.Error("expected error when options are passed to a no-options rule, got nil")
	}
}

func TestSchemaCompileInvalidJSON(t *testing.T) {
	schema := NewSchema([]byte("not json"))
	if _, err := schema.Compile(); err == nil {
		t.Error("expected error for invalid schema JSON, got nil")
	}
	// The compile error is memoized and surfaces through Validate too.
	if err := schema.Validate([]any{}); err == nil {
		t.Error("expected Validate to return the memoized compile error, got nil")
	}
}

func TestSchemaCompileIsMemoized(t *testing.T) {
	schema := NewSchema([]byte(`{"type": "array", "maxItems": 0}`))
	first, err := schema.Compile()
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	second, err := schema.Compile()
	if err != nil {
		t.Fatalf("second Compile failed: %v", err)
	}
	if first != second {
		t.Error("expected repeated Compile calls to return the same compiled schema")
	}
}

func TestSchemaConcurrentValidate(t *testing.T) {
	// Many rules share EmptyArraySchema and the linter fans out per file, so
	// first use of a schema can race across goroutines; the once must make
	// that safe. Run with -race to make this meaningful.
	schema := NewSchema([]byte(`{"type": "array", "maxItems": 0}`))
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := schema.Validate([]any{}); err != nil {
				t.Errorf("expected empty options to be valid, got: %v", err)
			}
			if err := schema.Validate([]any{"unexpected"}); err == nil {
				t.Error("expected error for non-empty options, got nil")
			}
		}()
	}
	wg.Wait()
}

func TestSchemaIsolatesResourcesAcrossSchemas(t *testing.T) {
	// Each Schema must compile with its own private compiler: a `$ref` in
	// one schema must never resolve against a `$defs`/`id` registered by an
	// earlier, unrelated schema's compilation, even if that id is still
	// "known" from a prior compile.
	_, err := NewSchema([]byte(`{
		"id": "urn:test-schema-isolation",
		"definitions": { "str": { "type": "string" } }
	}`)).Compile()
	if err != nil {
		t.Fatalf("Compile failed on first schema: %v", err)
	}

	_, err = NewSchema([]byte(`{
		"type": "array",
		"items": [ { "$ref": "urn:test-schema-isolation#/definitions/str" } ]
	}`)).Compile()
	if err == nil {
		t.Error("expected $ref into a different schema's resources to fail to resolve, got nil error")
	}
}

func TestEmptyArraySchema(t *testing.T) {
	if err := EmptyArraySchema.Validate([]any{}); err != nil {
		t.Errorf("expected empty options to be valid, got: %v", err)
	}
	if err := EmptyArraySchema.Validate([]any{"unexpected"}); err == nil {
		t.Error("expected error when options are passed to EmptyArraySchema, got nil")
	}
}

func TestSchemaAcceptsPatternWithJavaScriptOnlyRegexSyntax(t *testing.T) {
	// A schema author's own "pattern" is compiled once, at schema-compile
	// time, against the configured regexp engine. ajv (backed by JS's native
	// RegExp) accepts JavaScript-only syntax like lookbehind; Go's regexp
	// package (RE2), the jsonschema library's default engine, does not and
	// would fail to even compile this schema. compileSchemaJSON wires in
	// jsRegexpEngine specifically so this compiles and matches the way ajv
	// does.
	schema := NewSchema([]byte(`{
		"type": "array",
		"items": [ { "type": "string", "pattern": "(?<=a)b" } ],
		"minItems": 1,
		"maxItems": 1
	}`))
	if _, err := schema.Compile(); err != nil {
		t.Fatalf("Compile failed on a lookbehind pattern: %v", err)
	}

	if err := schema.Validate([]any{"ab"}); err != nil {
		t.Errorf("expected \"ab\" to satisfy (?<=a)b, got: %v", err)
	}
	if err := schema.Validate([]any{"cb"}); err == nil {
		t.Error("expected \"cb\" to fail (?<=a)b, got nil error")
	}
}

func TestValidateAcceptsJavaScriptOnlyRegexAsFormatRegexValue(t *testing.T) {
	// format: "regex" asserts that an *instance value* is itself a
	// syntactically valid regex — e.g. a rule option like
	// eslint-plugin-vitest's consistent-test-filename `pattern`, which
	// ESLint users may reasonably write using JavaScript-only syntax like
	// lookbehind. ajv accepts it (it just tries `new RegExp(value)`); Go's
	// regexp package does not, so without jsRegexpEngine this otherwise
	// legal ESLint config would be rejected.
	schema := NewSchema([]byte(`{
		"type": "array",
		"items": [
			{
				"type": "object",
				"properties": { "pattern": { "type": "string", "format": "regex" } },
				"additionalProperties": false
			}
		],
		"minItems": 0,
		"maxItems": 1
	}`))

	if err := schema.Validate([]any{map[string]any{"pattern": "(?<=a)b"}}); err != nil {
		t.Errorf("expected a lookbehind pattern value to satisfy format: regex, got: %v", err)
	}
	if err := schema.Validate([]any{map[string]any{"pattern": "a("}}); err == nil {
		t.Error("expected an unbalanced paren to fail format: regex, got nil error")
	}
}
