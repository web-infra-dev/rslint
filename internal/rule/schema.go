package rule

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/web-infra-dev/rslint/internal/utils"
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
	c.UseRegexpEngine(jsRegexpEngine)
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

// Validate validates the result against the compiled schema. It returns any
// validation error.
func (s *CompiledSchema) Validate(options any) error {
	return s.schema.Validate(options)
}

// jsRegexpEngine compiles pattern the way ajv itself does: as a JavaScript
// RegExp, not a Go RE2 regexp. It's wired into every CompileSchema's
// Compiler via [jsonschema.Compiler.UseRegexpEngine], so it governs both the
// "pattern" keyword (a schema author's own regex, compiled once against the
// schema) and "format": "regex" (asserting that an instance *value* is
// itself a syntactically valid regex — e.g. a rule option like
// eslint-plugin-vitest's consistent-test-filename `pattern`). The jsonschema
// library's own default engine wraps Go's regexp package (RE2), which
// rejects JavaScript-only regex features such as lookbehind
// (`(?<=...)`/`(?<!...)`) that ajv, backed by JS's native RegExp, accepts —
// silently rejecting an otherwise-legal ESLint config. internal/utils
// already wraps dlclark/regexp2 in ECMAScript mode for exactly this purpose
// (see [utils.JSRegexOptions]: "for ESLint rule options that model
// JavaScript RegExp patterns"), so this reuses it rather than introducing a
// second regex engine.
func jsRegexpEngine(pattern string) (jsonschema.Regexp, error) {
	re, err := utils.CompileRegexp2(pattern, utils.JSRegexOptions)
	if err != nil {
		return nil, err
	}
	return jsRegexp{re}, nil
}

// jsRegexp adapts *regexp2.Regexp to satisfy jsonschema.Regexp, whose
// MatchString returns a bare bool. regexp2.Regexp.MatchString instead
// returns (bool, error) — the error is reserved for runtime failures (e.g. a
// MatchTimeout) rather than compile-time invalidity (already surfaced by
// jsRegexpEngine above), so this treats one the same way every other
// rslint caller of regexp2 does: as no match, via [utils.Regexp2MatchString].
type jsRegexp struct {
	re *regexp2.Regexp
}

func (r jsRegexp) String() string {
	return r.re.String()
}

func (r jsRegexp) MatchString(s string) bool {
	return utils.Regexp2MatchString(r.re, s)
}
