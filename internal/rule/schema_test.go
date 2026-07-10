package rule

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

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
	if err := schema.Validate([]any{}); err != nil {
		t.Errorf("expected empty options to be valid, got: %v", err)
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

func TestCompileSchemaNoOptions(t *testing.T) {
	// The no-options convention: omit `items` (draft-04's own metaschema
	// requires a non-empty schemaArray, so `"items": []` is itself invalid)
	// and rely on `maxItems: 0` to force an empty options array.
	schemaJSON := []byte(`{"type": "array", "maxItems": 0}`)

	schema, err := CompileSchema(schemaJSON)
	if err != nil {
		t.Fatalf("CompileSchema failed: %v", err)
	}

	if err := schema.Validate([]any{}); err != nil {
		t.Errorf("expected empty options to be valid, got: %v", err)
	}
	if err := schema.Validate([]any{"unexpected"}); err == nil {
		t.Error("expected error when options are passed to a no-options rule, got nil")
	}
}

func TestCompileSchemaInvalidJSON(t *testing.T) {
	if _, err := CompileSchema([]byte("not json")); err == nil {
		t.Error("expected error for invalid schema JSON, got nil")
	}
}

func TestCompileSchemaIsolatesResourcesAcrossCalls(t *testing.T) {
	// Each CompileSchema call must use its own private compiler: a `$ref` in
	// one call's schema must never resolve against a `$defs`/`id` registered
	// by an earlier, unrelated call, even if that id is still "known" from a
	// prior compile.
	_, err := CompileSchema([]byte(`{
		"id": "urn:test-schema-isolation",
		"definitions": { "str": { "type": "string" } }
	}`))
	if err != nil {
		t.Fatalf("CompileSchema failed on first schema: %v", err)
	}

	_, err = CompileSchema([]byte(`{
		"type": "array",
		"items": [ { "$ref": "urn:test-schema-isolation#/definitions/str" } ]
	}`))
	if err == nil {
		t.Error("expected $ref into a different call's schema to fail to resolve, got nil error")
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

func TestCompileSchemaAcceptsPatternWithJavaScriptOnlyRegexSyntax(t *testing.T) {
	// A schema author's own "pattern" is compiled once, at CompileSchema
	// time, against the configured regexp engine. ajv (backed by JS's native
	// RegExp) accepts JavaScript-only syntax like lookbehind; Go's regexp
	// package (RE2), the jsonschema library's default engine, does not and
	// would fail to even compile this schema. CompileSchema wires in
	// jsRegexpEngine specifically so this compiles and matches the way ajv
	// does.
	schema, err := CompileSchema([]byte(`{
		"type": "array",
		"items": [ { "type": "string", "pattern": "(?<=a)b" } ],
		"minItems": 1,
		"maxItems": 1
	}`))
	if err != nil {
		t.Fatalf("CompileSchema failed to compile a lookbehind pattern: %v", err)
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
	schema, err := CompileSchema([]byte(`{
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
	if err != nil {
		t.Fatalf("CompileSchema failed: %v", err)
	}

	if err := schema.Validate([]any{map[string]any{"pattern": "(?<=a)b"}}); err != nil {
		t.Errorf("expected a lookbehind pattern value to satisfy format: regex, got: %v", err)
	}
	if err := schema.Validate([]any{map[string]any{"pattern": "a("}}); err == nil {
		t.Error("expected an unbalanced paren to fail format: regex, got nil error")
	}
}

// ajvFixtureCase mirrors one entry of testdata/ajv_fixtures.json, generated
// by scripts/gen-ajv-fixtures.js by running each {schema, input} pair
// through ajv@6 configured exactly the way ESLint configures it (see
// eslint/lib/shared/ajv.js: useDefaults, missingRefs, etc.) The corpus mixes
// synthetic cases pinning down specific useDefaults edge cases (nested/
// tuple/$ref gaps, allOf vs. anyOf/oneOf exclusion, additionalProperties/
// patternProperties, etc.) with real-world cases sourced from real ESLint /
// typescript-eslint / eslint-plugin-unicorn / eslint-plugin-react /
// eslint-plugin-vue rule option schemas, exercising general validation
// correctness: definitions/$defs + $ref chains, oneOf/anyOf/allOf exclusion,
// tuple vs. list-style items, additionalItems, additionalProperties-as-
// schema, minProperties, required, and uniqueItems. That file documents the
// shape and rationale of each case; run the generator script again to
// regenerate it if ajv's behavior is ever in doubt.
type ajvFixtureCase struct {
	Name   string          `json:"name"`
	Schema json.RawMessage `json:"schema"`
	Input  []any           `json:"input"`
	Output []any           `json:"output"`
	Valid  bool            `json:"valid"`
}

func TestValidateMatchesAjvFixtures(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "ajv_fixtures.json"))
	if err != nil {
		t.Fatalf("failed to read ajv fixtures: %v", err)
	}
	var cases []ajvFixtureCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatalf("failed to parse ajv fixtures: %v", err)
	}
	if len(cases) == 0 {
		t.Fatal("expected at least one ajv fixture case")
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			schema, err := CompileSchema(c.Schema)
			if err != nil {
				t.Fatalf("CompileSchema failed: %v", err)
			}

			err = schema.Validate(c.Input)
			if gotValid := err == nil; gotValid != c.Valid {
				t.Errorf("valid = %v, want %v (err: %v)", gotValid, c.Valid, err)
			}
		})
	}
}
