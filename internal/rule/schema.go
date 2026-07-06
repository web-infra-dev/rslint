package rule

import (
	"bytes"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// CompiledSchema is a compiled JSON Schema for a rule's options array. It
// wraps the underlying jsonschema library type so callers outside this
// package never need to import it directly.
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
// means positional/tuple validation — so authored schema.json files must not
// declare a newer `$schema`.
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

// ValidateOptions validates a rule's resolved options array (context.options —
// the config array after the severity level) against its compiled Schema.
func ValidateOptions(schema *CompiledSchema, options []any) error {
	return schema.schema.Validate(options)
}
