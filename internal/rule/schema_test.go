package rule

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
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

func TestValidateFillsDefaultsIntoCallerOptionsInPlace(t *testing.T) {
	// The public Validate contract config.ValidateRules relies on:
	// defaults land in the caller's own option maps, so the very config value
	// that was validated carries them into linting without extra plumbing —
	// the same way ajv's useDefaults mutates the instance ESLint hands it.
	schema := NewSchema([]byte(`{
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

	options := []any{map[string]any{"strict": false}}
	if err := schema.Validate(options); err != nil {
		t.Fatalf("expected valid options, got error: %v", err)
	}
	want := []any{map[string]any{"allow": []any{}, "strict": false}}
	if !reflect.DeepEqual(options, want) {
		t.Errorf("got %#v, want %#v", options, want)
	}
}

func TestValidateTupleGrowthIsValidationOnly(t *testing.T) {
	// A missing tuple slot's default can't be appended into the caller's
	// slice, so it participates in validation only. This matches observable
	// ESLint behavior: ESLint validates ruleOptions.slice(1) — a shallow copy
	// — so an ajv-grown tuple slot never reaches the rule either.
	schema := NewSchema([]byte(`{
		"type": "array",
		"items": [
			{ "type": "number", "default": 42, "minimum": 41 }
		],
		"minItems": 0,
		"maxItems": 1
	}`))

	options := []any{}
	if err := schema.Validate(options); err != nil {
		t.Fatalf("expected the grown default to validate, got error: %v", err)
	}
	if len(options) != 0 {
		t.Errorf("expected the caller's options to stay empty, got %#v", options)
	}
}

func TestValidateFillsObjectPropertyDefaults(t *testing.T) {
	schema := NewSchema([]byte(`{
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

	// Missing properties get filled; a present property is left untouched.
	got, err := schema.validateWithDefaults([]any{map[string]any{"strict": false}})
	if err != nil {
		t.Fatalf("expected valid options, got error: %v", err)
	}
	want := []any{map[string]any{"allow": []any{}, "strict": false}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestValidateFillsTrailingTupleItemDefaults(t *testing.T) {
	schema := NewSchema([]byte(`{
		"type": "array",
		"items": [
			{ "type": "string" },
			{ "type": "number", "default": 42 }
		],
		"minItems": 0,
		"maxItems": 2
	}`))

	// The second tuple item is missing, so it gets appended from its default.
	got, err := schema.validateWithDefaults([]any{"hello"})
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
	schema := NewSchema([]byte(`{
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

	got, err := schema.validateWithDefaults([]any{map[string]any{}})
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
	schema := NewSchema([]byte(`{
		"type": "array",
		"items": [
			{ "type": "string" },
			{ "type": "number", "default": 42 }
		],
		"minItems": 0,
		"maxItems": 2
	}`))

	got, err := schema.validateWithDefaults([]any{})
	want := []any{nil, float64(42)}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
	// Still invalid overall: index 0 is missing and fails its own "type": "string" check.
	if err == nil {
		t.Error("expected an error since the first tuple item is still missing, got nil")
	}
}

func TestValidateDefaultsDoNotShareStorageAcrossCalls(t *testing.T) {
	// A compound (map/slice) default must be deep-copied before insertion:
	// two calls to Validate on the same schema (e.g. from concurrent
	// goroutines validating distinct options for the same rule) must never
	// see or mutate each other's copy of the default. Regression test for a
	// concurrent-map-write crash/data race caused by inserting the schema's
	// literal default directly instead of a copy.
	//
	// A map is used here (rather than a slice) because mutating it via key
	// assignment always mutates shared storage directly, unlike a slice
	// grown by append, which can reallocate past a zero-capacity default.
	schema := NewSchema([]byte(`{
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

	first, err := schema.validateWithDefaults([]any{map[string]any{}})
	if err != nil {
		t.Fatalf("expected valid options, got error: %v", err)
	}
	first[0].(map[string]any)["config"].(map[string]any)["mutated"] = true

	second, err := schema.validateWithDefaults([]any{map[string]any{}})
	if err != nil {
		t.Fatalf("expected valid options, got error: %v", err)
	}
	secondConfig := second[0].(map[string]any)["config"].(map[string]any)
	if _, ok := secondConfig["mutated"]; ok {
		t.Errorf("expected second call's default to be independent of the first call's mutation, got %#v", secondConfig)
	}
}

func TestValidateConcurrentCompoundDefaultsDoNotRace(t *testing.T) {
	// Regression test for the crash reported against this schema: many
	// goroutines validating distinct options concurrently against a schema
	// with an object-typed default, previously hitting "fatal error:
	// concurrent map writes" (and, under -race, a reported data race) inside
	// applyDefaults -> normalizeNumbers because the compound default was
	// inserted by reference instead of copied.
	schema := NewSchema([]byte(`{
		"type": "array",
		"items": [
			{
				"type": "object",
				"properties": {
					"config": {
						"type": "object",
						"properties": {
							"level": { "type": "number", "default": 1 }
						},
						"default": {}
					}
				},
				"additionalProperties": false
			}
		],
		"minItems": 0,
		"maxItems": 1
	}`))

	var wg sync.WaitGroup
	for i := range 64 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			options := []any{map[string]any{}}
			if err := schema.Validate(options); err != nil {
				t.Errorf("expected valid options, got error: %v", err)
			}
			config := options[0].(map[string]any)["config"].(map[string]any)
			config["level"] = float64(i) // mutate this call's own copy
		}(i)
	}
	wg.Wait()
}

func TestValidateHandlesCyclicPureRefWithoutHanging(t *testing.T) {
	// Two schemas that only $ref each other, with neither ever declaring its
	// own Properties/Items: this is pathological enough that ajv itself
	// can't even compile it (stack overflow trying to inline/resolve the
	// cycle at compile time). applyDefaults's resolveRef must not hang or
	// panic on it either, even though there's no real ajv output to match —
	// it should just give up resolving and leave the value as-is.
	schema := NewSchema([]byte(`{
		"type": "array",
		"definitions": {
			"a": { "$ref": "#/definitions/b" },
			"b": { "$ref": "#/definitions/a" }
		},
		"items": [ { "$ref": "#/definitions/a" } ],
		"minItems": 0,
		"maxItems": 1
	}`))

	// No assertion on validity: this schema never resolves to a concrete
	// type, so either outcome is acceptable — the point is that it returns
	// at all, and doesn't corrupt the input on the way.
	got, _ := schema.validateWithDefaults([]any{map[string]any{}})
	want := []any{map[string]any{}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestValidateHandlesCyclicAllOfWithoutHanging(t *testing.T) {
	// Two definitions whose sole allOf branch $refs back to the other: like
	// TestValidateHandlesCyclicPureRefWithoutHanging, this is pathological
	// enough that the underlying jsonschema library itself rejects it (at
	// Validate time, once it tries to actually check the instance against
	// the cyclic schema) — but defaultSources's own cycle guard must still
	// stop collectDefaultSources from recursing forever while walking it to
	// fill in defaults beforehand.
	schema := NewSchema([]byte(`{
		"type": "array",
		"definitions": {
			"a": { "allOf": [ { "$ref": "#/definitions/b" } ] },
			"b": { "allOf": [ { "$ref": "#/definitions/a" } ] }
		},
		"items": [ { "$ref": "#/definitions/a" } ],
		"minItems": 0,
		"maxItems": 1
	}`))

	// No assertion on validity: the point is that this returns at all, and
	// doesn't corrupt the input on the way.
	got, _ := schema.validateWithDefaults([]any{map[string]any{}})
	want := []any{map[string]any{}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestValidateFillsAllOfBranchDefaults(t *testing.T) {
	// allOf is unambiguous — every branch must hold — unlike anyOf/oneOf, so
	// a default declared on an allOf branch's property is filled in, exactly
	// as if that branch's properties were merged into this level. Confirmed
	// against ajv@6: see allOf_multiple_branches_each_contribute_defaults in
	// testdata/ajv_fixtures.json.
	schema := NewSchema([]byte(`{
		"type": "array",
		"items": [
			{
				"type": "object",
				"allOf": [
					{ "properties": { "a": { "type": "string", "default": "a-default" } } },
					{ "properties": { "b": { "type": "string", "default": "b-default" } } }
				]
			}
		],
		"minItems": 0,
		"maxItems": 1
	}`))

	got, err := schema.validateWithDefaults([]any{map[string]any{}})
	if err != nil {
		t.Fatalf("expected valid options, got error: %v", err)
	}
	want := []any{map[string]any{"a": "a-default", "b": "b-default"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestValidateFillsAdditionalPropertiesSchemaDefaultForExistingKey(t *testing.T) {
	// additionalProperties as a schema never manufactures a brand-new key,
	// but a key already present that's only matched by additionalProperties
	// (not a named property) still gets its own nested defaults filled in.
	// Confirmed against ajv@6.
	schema := NewSchema([]byte(`{
		"type": "array",
		"items": [
			{
				"type": "object",
				"properties": { "known": { "type": "string" } },
				"additionalProperties": {
					"type": "object",
					"properties": { "inner": { "type": "string", "default": "inner-default" } }
				}
			}
		],
		"minItems": 0,
		"maxItems": 1
	}`))

	// "extra" isn't a declared property, only matched via additionalProperties.
	got, err := schema.validateWithDefaults([]any{map[string]any{"known": "x", "extra": map[string]any{}}})
	if err != nil {
		t.Fatalf("expected valid options, got error: %v", err)
	}
	want := []any{map[string]any{"known": "x", "extra": map[string]any{"inner": "inner-default"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}

	// additionalProperties never creates a brand-new key on its own account.
	got, err = schema.validateWithDefaults([]any{map[string]any{"known": "x"}})
	if err != nil {
		t.Fatalf("expected valid options, got error: %v", err)
	}
	want = []any{map[string]any{"known": "x"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestValidateFillsPatternPropertiesDefaultForExistingKey(t *testing.T) {
	// Same idea as additionalProperties, but matched via patternProperties.
	// Confirmed against ajv@6.
	schema := NewSchema([]byte(`{
		"type": "array",
		"items": [
			{
				"type": "object",
				"patternProperties": {
					"^opt_": {
						"type": "object",
						"properties": { "inner": { "type": "string", "default": "pp-default" } }
					}
				}
			}
		],
		"minItems": 0,
		"maxItems": 1
	}`))

	got, err := schema.validateWithDefaults([]any{map[string]any{"opt_x": map[string]any{}}})
	if err != nil {
		t.Fatalf("expected valid options, got error: %v", err)
	}
	want := []any{map[string]any{"opt_x": map[string]any{"inner": "pp-default"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestValidateFillsRefSiblingDefault(t *testing.T) {
	// A literal `default` written directly beside a bare $ref (a "$ref
	// sibling") is honored using that literal value, not the ref target's
	// own top-level default — confirmed against ajv@6, which (unlike the
	// underlying jsonschema library, which discards every sibling keyword
	// next to "$ref" for Draft 4) still applies it.
	schema := NewSchema([]byte(`{
		"type": "array",
		"definitions": {
			"foo": { "type": "string", "default": "ref-target-default" }
		},
		"items": [
			{
				"type": "object",
				"properties": {
					"foo": { "$ref": "#/definitions/foo", "default": "sibling-default" }
				},
				"additionalProperties": false
			}
		],
		"minItems": 0,
		"maxItems": 1
	}`))

	got, err := schema.validateWithDefaults([]any{map[string]any{}})
	if err != nil {
		t.Fatalf("expected valid options, got error: %v", err)
	}
	want := []any{map[string]any{"foo": "sibling-default"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestValidateFillsListStyleItemsDefaultPerElement(t *testing.T) {
	// A list-style (single-schema, non-tuple) `items` schema applies the same
	// schema to every element, so each already-present element gets its own
	// defaults filled in independently — this is the real-world case
	// @graphql-eslint/eslint-plugin's relay-arguments schema hits (a rule's
	// own top-level options schema declared as list-style rather than a
	// tuple), confirmed against ajv@6.
	schema := NewSchema([]byte(`{
		"type": "array",
		"items": [
			{
				"type": "array",
				"items": {
					"type": "object",
					"properties": { "foo": { "type": "string", "default": "d" } }
				}
			}
		],
		"minItems": 0,
		"maxItems": 1
	}`))

	got, err := schema.validateWithDefaults([]any{[]any{map[string]any{}, map[string]any{"foo": "x"}}})
	if err != nil {
		t.Fatalf("expected valid options, got error: %v", err)
	}
	want := []any{[]any{map[string]any{"foo": "d"}, map[string]any{"foo": "x"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestValidateListStyleItemsDoesNotGrowArray(t *testing.T) {
	// Unlike a tuple position with its own literal default, list-style
	// `items` never determines the array's length, so it never pads a
	// too-short array — confirmed against ajv@6.
	schema := NewSchema([]byte(`{
		"type": "array",
		"items": [
			{
				"type": "array",
				"minItems": 2,
				"items": { "type": "string", "default": "d" }
			}
		],
		"minItems": 0,
		"maxItems": 1
	}`))

	got, err := schema.validateWithDefaults([]any{[]any{"a"}})
	want := []any{[]any{"a"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
	if err == nil {
		t.Error("expected an error since the inner array is still short of minItems, got nil")
	}
}

func TestValidateFillsTopLevelListStyleItemsDefault(t *testing.T) {
	// The exact real-world shape: a rule's own top-level options schema is
	// list-style (a single object schema, not a tuple) rather than the usual
	// `{"type": "array", "items": [...]}` tuple convention. An empty `{}`
	// option gets its declared default filled in, satisfying minProperties —
	// confirmed against ajv@6.
	schema := NewSchema([]byte(`{
		"type": "array",
		"maxItems": 1,
		"items": {
			"type": "object",
			"minProperties": 1,
			"properties": { "includeBoth": { "type": "boolean", "default": true } }
		}
	}`))

	got, err := schema.validateWithDefaults([]any{map[string]any{}})
	if err != nil {
		t.Fatalf("expected valid options, got error: %v", err)
	}
	want := []any{map[string]any{"includeBoth": true}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
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
			schema := NewSchema(c.Schema)
			if _, err := schema.Compile(); err != nil {
				t.Fatalf("Compile failed: %v", err)
			}

			got, err := schema.validateWithDefaults(c.Input)
			if gotValid := err == nil; gotValid != c.Valid {
				t.Errorf("valid = %v, want %v (err: %v)", gotValid, c.Valid, err)
			}
			if want := c.Output; !reflect.DeepEqual(got, want) {
				t.Errorf("got %#v, want %#v", got, want)
			}
		})
	}
}
