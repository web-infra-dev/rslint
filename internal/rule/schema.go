package rule

import (
	"bytes"
	"encoding/json"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// CompiledSchema is a compiled JSON Schema for a rule's options array.
type CompiledSchema struct {
	schema *jsonschema.Schema
}

// CompileSchema compiles a rule's options JSON Schema. By convention the
// schema describes the ESLint-style options array as a tuple: a top-level
// `{"type": "array", "items": [...]}` — or, for a rule that takes no options,
// `{"type": "array", "maxItems": 0}` with `items` omitted (draft-04's own
// metaschema requires a non-empty schemaArray, so `"items": []` is itself
// invalid). See the <rule-name>.schema.json convention.
//
// rslint only supports Draft 4 — the draft under which a plain array `items`
// means positional/tuple validation — so this is a convention for our own
// authored schema.json files, not a constraint CompileSchema enforces: an
// explicit `$schema` declaring a newer draft would compile fine, just under
// that draft's (different) `items` semantics.
//
// Each call compiles against its own private jsonschema.Compiler, so a
// `$ref` in one rule's schema can never resolve into another rule's `$defs`
// or resources, even if two schemas happen to reuse the same `$id`.
//
// Rules that take no options should reuse [EmptyArraySchema] instead of calling
// CompileSchema with that boilerplate schema themselves.
func CompileSchema(schemaJSON []byte) (*CompiledSchema, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaJSON))
	if err != nil {
		return nil, err
	}

	c := jsonschema.NewCompiler()
	c.DefaultDraft(jsonschema.Draft4)
	const resourceURL = "schema.json"
	if err := c.AddResource(resourceURL, doc); err != nil {
		return nil, err
	}
	compiled, err := c.Compile(resourceURL)
	if err != nil {
		return nil, err
	}
	return &CompiledSchema{schema: compiled}, nil
}

// MustCompileSchema is like CompileSchema but panics instead of returning an
// error. Intended for package-level CompiledSchema variables initialized at
// startup, in the vein of regexp.MustCompile.
func MustCompileSchema(schemaJSON []byte) *CompiledSchema {
	s, err := CompileSchema(schemaJSON)
	if err != nil {
		panic(err)
	}
	return s
}

// EmptyArraySchema validates that a rule's resolved options array
// (context.options) is empty. Rules that take no options should reference
// this shared schema rather than each compiling their own copy of the same
// schema.json.
var EmptyArraySchema = MustCompileSchema([]byte(`{"type": "array", "maxItems": 0}`))

// Validate fills in a rule's resolved options array (context.options — the
// config array after the severity level) with schema-declared defaults, then
// validates the result against the compiled schema. It returns the options
// with defaults applied, along with any validation error.
//
// Defaults are applied the way ajv's `useDefaults` option does: only plain
// object `properties` and tuple-style array `items` are filled in. Defaults
// are not applied through `$ref` or inside `allOf`/`anyOf`/`oneOf`/`not`,
// since which branch applies (or whether the ref should be followed) is
// ambiguous.
func (s *CompiledSchema) Validate(options any) (any, error) {
	options = applyDefaults(s.schema, options)
	if err := s.schema.Validate(options); err != nil {
		return options, err
	}
	return options, nil
}

// applyDefaults recursively fills missing object properties and array tuple
// items in v with the defaults declared on s. It mutates and returns v (or,
// for a tuple that must grow, a value built from appending onto v) rather
// than copying it — matching ajv's own useDefaults, which also inserts its
// defaults directly into the instance being validated rather than a copy.
//
// A tuple item is filled only when it is genuinely absent (i.e. beyond the
// input's original length), following ajv's own dot-template: every tuple
// position with a declared default is unconditionally considered, so a later
// item's default can be inserted even while an earlier, default-less item is
// still missing — matching ajv, this leaves a hole (nil) at that earlier
// position rather than refusing to fill the later one.
func applyDefaults(s *jsonschema.Schema, v any) any {
	switch val := v.(type) {
	case map[string]any:
		for key, prop := range s.Properties {
			if _, ok := val[key]; !ok {
				if prop.Default == nil {
					continue
				}
				val[key] = normalizeNumbers(*prop.Default)
			}
			val[key] = applyDefaults(prop, val[key])
		}
		return val
	case []any:
		items, ok := s.Items.([]*jsonschema.Schema)
		if !ok {
			return val
		}
		origLen := len(val)
		for i, item := range items {
			if i < origLen {
				val[i] = applyDefaults(item, val[i])
				continue
			}
			if item.Default == nil {
				// No default to contribute at this position: don't grow the
				// array on its account (ajv likewise emits no code at all for
				// a tuple position without a declared default).
				continue
			}
			for len(val) < i {
				val = append(val, nil) // pad any not-yet-visited gap up to i
			}
			val = append(val, applyDefaults(item, normalizeNumbers(*item.Default)))
		}
		return val
	default:
		return v
	}
}

// normalizeNumbers mutates v in place, converting any json.Number leaves to
// float64, and returns it (a bare top-level json.Number is instead returned
// as a new float64, since a number can't be mutated through its interface
// value). CompileSchema decodes schema.json — and its "default" values —
// with UseNumber for precision, but the rest of rslint decodes rule options
// with plain encoding/json, which represents numbers as float64. Without
// this, a defaulted numeric option would have a different Go type than the
// same option supplied by the user.
func normalizeNumbers(v any) any {
	switch val := v.(type) {
	case map[string]any:
		for k, v2 := range val {
			val[k] = normalizeNumbers(v2)
		}
		return val
	case []any:
		for i, v2 := range val {
			val[i] = normalizeNumbers(v2)
		}
		return val
	case json.Number:
		f, _ := val.Float64()
		return f
	default:
		return v
	}
}
