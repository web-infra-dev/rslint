package rule

import (
	"bytes"
	"sync"

	"github.com/dlclark/regexp2"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Schema is a rule's options JSON Schema, compiled lazily on first use. By
// convention the schema describes the ESLint-style options array
// (context.options — the config array after the severity level) as a tuple: a
// top-level `{"type": "array", "items": [...]}` — or, for a rule that takes no
// options, `{"type": "array", "maxItems": 0}` with `items` omitted (draft-04's
// own metaschema requires a non-empty schemaArray, so `"items": []` is itself
// invalid). See the <rule-name>.schema.json convention.
//
// rslint only supports Draft 4 — the draft under which a plain array `items`
// means positional/tuple validation — so this is a convention for our own
// authored schema.json files, not a constraint compilation enforces: an
// explicit `$schema` declaring a newer draft would compile fine, just under
// that draft's (different) `items` semantics.
//
// Compilation is deferred until the schema is first used ([Schema.Compile] /
// [Schema.Validate]) and guarded by a sync.Once, so a schema is compiled at
// most once per process no matter how many goroutines race to use it — and a
// rule that is never enabled never pays for compiling its schema at all. The
// once also lets many rules share a single Schema value: [EmptyArraySchema]
// is referenced by every no-options rule yet compiles a single time.
type Schema struct {
	rawJSON []byte

	once   sync.Once
	schema *jsonschema.Schema
	err    error
}

// NewSchema returns a Schema that will compile rawJSON on first use. rawJSON
// is typically a `//go:embed`-ed <rule-name>.schema.json sitting beside the
// rule's source.
//
// Rules that take no options should reference the shared [EmptyArraySchema]
// instead of constructing their own copy of the same schema.
func NewSchema(rawJSON []byte) *Schema {
	return &Schema{rawJSON: rawJSON}
}

// EmptyArraySchema validates that a rule's resolved options array
// (context.options) is empty. Rules that take no options should reference
// this shared schema rather than each carrying their own copy of the same
// schema.json; the lazy once in [Schema.Compile] means it compiles a single
// time process-wide regardless of how many rules use it.
var EmptyArraySchema = NewSchema([]byte(`{"type": "array", "maxItems": 0}`))

// Compile compiles the schema's raw JSON exactly once and returns the
// memoized result; every subsequent call — from any goroutine — returns the
// same compiled schema (or the same compile error).
//
// Each compilation uses its own private jsonschema.Compiler, so a `$ref` in
// one rule's schema can never resolve into another rule's `$defs` or
// resources, even if two schemas happen to reuse the same `$id`.
func (s *Schema) Compile() (*jsonschema.Schema, error) {
	s.once.Do(func() {
		s.schema, s.err = compileSchemaJSON(s.rawJSON)
	})
	return s.schema, s.err
}

// Validate compiles the schema (once) and validates a rule's resolved
// options array (context.options) against it. It returns the compile error,
// if any, otherwise any validation error. A nil options slice is validated
// as an empty options array, not as JSON null.
//
// The compiled schema is read-only, so Validate is safe for concurrent use.
func (s *Schema) Validate(options []any) error {
	compiled, err := s.Compile()
	if err != nil {
		return err
	}
	if options == nil {
		options = []any{}
	}
	return compiled.Validate(options)
}

// compileSchemaJSON does the actual one-time compilation work behind
// [Schema.Compile].
func compileSchemaJSON(rawJSON []byte) (*jsonschema.Schema, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(rawJSON))
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
	return c.Compile(resourceURL)
}

// jsRegexpEngine compiles pattern the way ajv itself does: as a JavaScript
// RegExp, not a Go RE2 regexp. It's wired into every compilation's Compiler
// via [jsonschema.Compiler.UseRegexpEngine], so it governs both the
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
