package rule

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
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
	if _, err := schema.Validate([]any{}); err != nil {
		t.Errorf("expected empty options to be valid, got: %v", err)
	}

	// A well-formed options object.
	valid := []any{map[string]any{"allow": []any{"warn", "error"}}}
	if _, err := schema.Validate(valid); err != nil {
		t.Errorf("expected valid options, got: %v", err)
	}

	// `allow` must be an array of strings, not a bare string.
	invalidType := []any{map[string]any{"allow": "warn"}}
	if _, err := schema.Validate(invalidType); err == nil {
		t.Error("expected error for allow as non-array, got nil")
	}

	// Unknown property rejected by additionalProperties: false.
	invalidProp := []any{map[string]any{"unknown": true}}
	if _, err := schema.Validate(invalidProp); err == nil {
		t.Error("expected error for unknown property, got nil")
	}

	// Exceeds maxItems: 1.
	tooMany := []any{
		map[string]any{"allow": []any{"warn"}},
		map[string]any{"allow": []any{"error"}},
	}
	if _, err := schema.Validate(tooMany); err == nil {
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

	if _, err := schema.Validate([]any{}); err != nil {
		t.Errorf("expected empty options to be valid, got: %v", err)
	}
	if _, err := schema.Validate([]any{"unexpected"}); err == nil {
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
	if _, err := EmptyArraySchema.Validate([]any{}); err != nil {
		t.Errorf("expected empty options to be valid, got: %v", err)
	}
	if _, err := EmptyArraySchema.Validate([]any{"unexpected"}); err == nil {
		t.Error("expected error when options are passed to EmptyArraySchema, got nil")
	}
}

func TestValidateFillsObjectPropertyDefaults(t *testing.T) {
	schema, err := CompileSchema([]byte(`{
		"type": "array",
		"items": [
			{
				"type": "object",
				"properties": {
					"allow": {
						"type": "array",
						"items": { "type": "string" },
						"default": []
					},
					"strict": { "type": "boolean", "default": true }
				},
				"additionalProperties": false
			}
		],
		"minItems": 0,
		"maxItems": 1
	}`))
	if err != nil {
		t.Fatalf("CompileSchema failed: %v", err)
	}

	// Missing properties get filled; a present property is left untouched.
	got, err := schema.Validate([]any{map[string]any{"strict": false}})
	if err != nil {
		t.Fatalf("expected valid options, got error: %v", err)
	}
	want := []any{map[string]any{"allow": []any{}, "strict": false}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestValidateFillsTrailingTupleItemDefaults(t *testing.T) {
	schema, err := CompileSchema([]byte(`{
		"type": "array",
		"items": [
			{ "type": "string" },
			{ "type": "number", "default": 42 }
		],
		"minItems": 0,
		"maxItems": 2
	}`))
	if err != nil {
		t.Fatalf("CompileSchema failed: %v", err)
	}

	// The second tuple item is missing, so it gets appended from its default.
	got, err := schema.Validate([]any{"hello"})
	if err != nil {
		t.Fatalf("expected valid options, got error: %v", err)
	}
	want := []any{"hello", float64(42)}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestValidateFillsNestedDefaults(t *testing.T) {
	// A default nested inside another default (object-in-object) must also be
	// filled in, not just the top-level property.
	schema, err := CompileSchema([]byte(`{
		"type": "array",
		"items": [
			{
				"type": "object",
				"properties": {
					"outer": {
						"type": "object",
						"properties": {
							"inner": { "type": "string", "default": "fallback" }
						},
						"additionalProperties": false,
						"default": {}
					}
				},
				"additionalProperties": false
			}
		],
		"minItems": 0,
		"maxItems": 1
	}`))
	if err != nil {
		t.Fatalf("CompileSchema failed: %v", err)
	}

	got, err := schema.Validate([]any{map[string]any{}})
	if err != nil {
		t.Fatalf("expected valid options, got error: %v", err)
	}
	want := []any{map[string]any{"outer": map[string]any{"inner": "fallback"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestValidateFillsTrailingItemCreatesGapWhenEarlierItemHasNoDefault(t *testing.T) {
	// Matches ajv's own useDefaults: every tuple position with a declared
	// default is considered unconditionally, so a later item's default can
	// still be inserted even while an earlier, default-less item is missing.
	// That leaves a hole (nil) at the earlier position rather than refusing
	// to fill the later one.
	schema, err := CompileSchema([]byte(`{
		"type": "array",
		"items": [
			{ "type": "string" },
			{ "type": "number", "default": 42 }
		],
		"minItems": 0,
		"maxItems": 2
	}`))
	if err != nil {
		t.Fatalf("CompileSchema failed: %v", err)
	}

	got, err := schema.Validate([]any{})
	want := []any{nil, float64(42)}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
	// Still invalid overall: index 0 is missing and fails its own "type": "string" check.
	if err == nil {
		t.Error("expected an error since the first tuple item is still missing, got nil")
	}
}

func TestValidateDefaultsShareStorageAcrossCalls(t *testing.T) {
	// Deliberate trade-off: applyDefaults inserts the schema's compiled
	// Default directly rather than deep-copying it, so a compound (map/slice)
	// default is the *same* underlying value across every call to Validate on
	// this schema. Mutating one call's result in place can affect another's.
	// This documents that behavior so it isn't mistaken for a bug later.
	//
	// A map is used here (rather than a slice) because mutating it via key
	// assignment always mutates the shared storage directly; a slice default
	// would only demonstrate this if grown by index-assignment rather than
	// append, since append past a zero-capacity empty-slice default happens
	// to reallocate.
	schema, err := CompileSchema([]byte(`{
		"type": "array",
		"items": [
			{
				"type": "object",
				"properties": {
					"config": { "type": "object", "default": {} }
				},
				"additionalProperties": false
			}
		],
		"minItems": 0,
		"maxItems": 1
	}`))
	if err != nil {
		t.Fatalf("CompileSchema failed: %v", err)
	}

	first, err := schema.Validate([]any{map[string]any{}})
	if err != nil {
		t.Fatalf("expected valid options, got error: %v", err)
	}
	first.([]any)[0].(map[string]any)["config"].(map[string]any)["mutated"] = true

	second, err := schema.Validate([]any{map[string]any{}})
	if err != nil {
		t.Fatalf("expected valid options, got error: %v", err)
	}
	secondConfig := second.([]any)[0].(map[string]any)["config"].(map[string]any)
	if mutated, ok := secondConfig["mutated"]; !ok || mutated != true {
		t.Errorf("expected second call's default to share storage with the first call's mutation, got %#v", secondConfig)
	}
}

func TestValidateHandlesCyclicPureRefWithoutHanging(t *testing.T) {
	// Two schemas that only $ref each other, with neither ever declaring its
	// own Properties/Items: this is pathological enough that ajv itself
	// can't even compile it (stack overflow trying to inline/resolve the
	// cycle at compile time). applyDefaults's resolveRef must not hang or
	// panic on it either, even though there's no real ajv output to match —
	// it should just give up resolving and leave the value as-is.
	schema, err := CompileSchema([]byte(`{
		"type": "array",
		"definitions": {
			"a": { "$ref": "#/definitions/b" },
			"b": { "$ref": "#/definitions/a" }
		},
		"items": [ { "$ref": "#/definitions/a" } ],
		"minItems": 0,
		"maxItems": 1
	}`))
	if err != nil {
		t.Fatalf("CompileSchema failed: %v", err)
	}

	// No assertion on validity: this schema never resolves to a concrete
	// type, so either outcome is acceptable — the point is that it returns
	// at all, and doesn't corrupt the input on the way.
	got, _ := schema.Validate([]any{map[string]any{}})
	want := []any{map[string]any{}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

// ajvFixtureCase mirrors one entry of testdata/ajv_defaults_fixtures.json,
// generated by scripts/gen-ajv-defaults-fixtures.js by running the same
// {schema, input} pair through ajv@6 configured exactly the way ESLint
// configures it (see eslint/lib/shared/ajv.js: useDefaults, missingRefs,
// etc.) That file documents the shape and rationale of each case; run the
// generator script again to regenerate it if ajv's behavior is ever in doubt.
type ajvFixtureCase struct {
	Name   string          `json:"name"`
	Schema json.RawMessage `json:"schema"`
	Input  []any           `json:"input"`
	Output []any           `json:"output"`
	Valid  bool            `json:"valid"`
}

func TestValidateMatchesAjvDefaultsFixtures(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "ajv_defaults_fixtures.json"))
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

			got, err := schema.Validate(c.Input)
			if gotValid := err == nil; gotValid != c.Valid {
				t.Errorf("valid = %v, want %v (err: %v)", gotValid, c.Valid, err)
			}
			if want := any(c.Output); !reflect.DeepEqual(got, want) {
				t.Errorf("got %#v, want %#v", got, want)
			}
		})
	}
}

// TestValidateMatchesAjvRealWorldFixtures runs testdata/ajv_realworld_fixtures.json:
// {schema, input} pairs sourced from real ESLint / typescript-eslint /
// eslint-plugin-unicorn / eslint-plugin-react / eslint-plugin-vue rule
// option schemas (generated by scripts/gen-ajv-realworld-fixtures.js), run
// through the same ESLint-configured ajv@6 instance as the synthetic
// defaults fixtures above. Unlike ajv_defaults_fixtures.json — which
// targets the useDefaults corpus specifically — this corpus exercises
// general validation correctness against schema shapes actually shipped by
// real rules: definitions/$defs + $ref chains, oneOf/anyOf/allOf exclusion,
// tuple vs. list-style items, additionalItems, additionalProperties-as-schema,
// minProperties, required, and uniqueItems.
func TestValidateMatchesAjvRealWorldFixtures(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "ajv_realworld_fixtures.json"))
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

			got, err := schema.Validate(c.Input)
			if gotValid := err == nil; gotValid != c.Valid {
				t.Errorf("valid = %v, want %v (err: %v)", gotValid, c.Valid, err)
			}
			if want := any(c.Output); !reflect.DeepEqual(got, want) {
				t.Errorf("got %#v, want %#v", got, want)
			}
		})
	}
}
