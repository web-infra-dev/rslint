package rule

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
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
	// doc is the schema's own raw decoded document. It exists solely so
	// [literalDefault] can recover a literal `default` written directly
	// beside a bare `$ref` — a value the compiled *jsonschema.Schema itself
	// never carries, since the underlying library discards every sibling
	// keyword next to "$ref" for Draft 4.
	doc any
	err error
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

// RawJSON returns the schema's original, not-yet-compiled JSON — the same bytes
// passed to NewSchema. Exposed for tooling that needs the schema text
// itself rather than a compiled validator (e.g. tools/dump-rule-schemas, which
// dumps every registered rule's schema for the TypeScript rule-options
// generator).
func (s *Schema) RawJSON() []byte {
	return s.rawJSON
}

// Compile compiles the schema's raw JSON exactly once and returns the
// memoized result; every subsequent call — from any goroutine — returns the
// same compiled schema (or the same compile error).
//
// Each compilation uses its own private jsonschema.Compiler, so a `$ref` in
// one rule's schema can never resolve into another rule's `$defs` or
// resources, even if two schemas happen to reuse the same `$id`.
func (s *Schema) Compile() (*jsonschema.Schema, error) {
	s.once.Do(func() {
		s.schema, s.doc, s.err = compileSchemaJSON(s.rawJSON)
	})
	return s.schema, s.err
}

// Validate compiles the schema (once), fills a rule's resolved options array
// (context.options — the config array after the severity level) with
// schema-declared defaults, then validates the result against the compiled
// schema. It returns the compile error, if any, otherwise any validation
// error. A nil options slice is validated as an empty options array, not as
// JSON null.
//
// Defaults are applied the way ajv's `useDefaults` option does: object
// `properties` (including keys only matched by `patternProperties` or a
// schema-valued `additionalProperties`, for keys already present) and array
// `items` — both tuple-style (a slot's own default can grow the array) and
// list-style (the same schema applied to every already-present element,
// never growing the array) — are filled in, including through a `$ref` once
// the property/item it's attached to is already present, and through every
// `allOf` branch (unambiguous, since all of them must hold) — see
// [applyDefaults]. Defaults are never applied inside `anyOf`/`oneOf`/`not`,
// since which branch applies (if any) is genuinely ambiguous.
//
// Like ajv — which fills defaults directly into the instance it validates —
// the fill mutates the options' own maps and slices in place, so a defaulted
// property lands in the caller's data. The one exception is a grown array:
// appending a missing tuple slot's default can't reach the caller's slice
// header, so the grown options exist only for the validation itself. That,
// too, matches observable ESLint behavior: ESLint validates a shallow copy
// (`ruleOptions.slice(1)`) of the very array it later re-slices the rule's
// runtime options from, so an ajv-grown tuple slot never reaches the rule
// either — only defaults filled into already-present (shared) objects do.
//
// The compiled schema itself is never mutated, so Validate is safe for
// concurrent use on distinct options values; two goroutines must not
// Validate the very same options value at once.
func (s *Schema) Validate(options []any) error {
	_, err := s.validateWithDefaults(options)
	return err
}

// validateWithDefaults is [Schema.Validate] with access to the
// defaults-applied options value — including a tuple-grown array that
// in-place mutation alone can't hand back.
func (s *Schema) validateWithDefaults(options []any) ([]any, error) {
	compiled, err := s.Compile()
	if err != nil {
		return options, err
	}
	if options == nil {
		options = []any{}
	}
	// applyDefaults returns a value of its input's own kind, so a []any in is
	// always a []any out (possibly grown by tuple defaults).
	applied, _ := applyDefaults(compiled, options, s.doc).([]any)
	if err := compiled.Validate(applied); err != nil {
		return applied, err
	}
	return applied, nil
}

// compileSchemaJSON does the actual one-time compilation work behind
// [Schema.Compile], returning both the compiled schema and the raw decoded
// document it was compiled from (kept for [literalDefault]).
func compileSchemaJSON(rawJSON []byte) (*jsonschema.Schema, any, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(rawJSON))
	if err != nil {
		return nil, nil, err
	}

	c := jsonschema.NewCompiler()
	c.DefaultDraft(jsonschema.Draft4)
	c.UseRegexpEngine(jsRegexpEngine)
	// An absolute URI in a private scheme, so the schema's base URL can never
	// be confused with (or resolve against) a real file path or another
	// resource's URL the way a bare relative "schema.json" could.
	const resourceURL = "rslint:///schema.json"
	if err := c.AddResource(resourceURL, doc); err != nil {
		return nil, nil, err
	}
	compiled, err := c.Compile(resourceURL)
	if err != nil {
		return nil, nil, err
	}
	return compiled, doc, nil
}

// applyDefaults recursively fills missing object properties and array items
// in v with the defaults declared on s. It mutates and returns v (or, for a
// tuple that must grow, a value built from appending onto v) rather than
// copying it — matching ajv's own useDefaults, which also inserts its
// defaults directly into the instance being validated rather than a copy.
//
// A tuple item is filled only when it is genuinely absent (i.e. beyond the
// input's original length), following ajv's own dot-template: every tuple
// position with a declared default is unconditionally considered, so a later
// item's default can be inserted even while an earlier, default-less item is
// still missing — matching ajv, this leaves a hole (nil) at that earlier
// position rather than refusing to fill the later one.
//
// A list-style `items` schema (a single schema applied to every element,
// rather than a tuple array) is different: since it says nothing about how
// many elements there are, it never grows v — confirmed against ajv@6, which
// likewise never pads a too-short array on a list-style `items` schema's
// account (unlike a tuple position with its own literal default, which
// does). It only ever fills defaults into elements v already has.
//
// A property/item schema contributes a default to fill an absent slot only
// via [literalDefault] — a literal `default` written directly on that exact
// schema node (including one written beside a bare `$ref`, i.e. a "$ref
// sibling"), never one pulled in by dereferencing (confirmed against ajv@6:
// a $ref target's own top-level `default` is not pulled through into the
// parent when the parent slot is absent, even across a multi-hop $ref
// chain). But once a slot is already present — either originally, or because
// it was just filled from a literal default at this level — recursing into
// it uses [resolveRef] to follow the schema's own `$ref` chain first. This
// matches ajv compiling `$ref` as a call into the referenced schema's own
// compiled validator: since objects/arrays are mutated by reference, that
// nested call's own useDefaults processing lands directly in the shared data
// being validated here.
//
// A compound (map/slice) literalDefault is always run through [deepCopyJSON]
// before it's inserted: the value literalDefault returns is owned by s — the
// compiled *jsonschema.Schema, shared by every Validate call and every rule
// that reuses the same Schema (e.g. [EmptyArraySchema]) — so inserting it
// directly would let this call's own recursion into the newly filled slot
// (or any later mutation of the validated options) mutate that shared value,
// racing with any other concurrent Validate call that inserts the same
// default.
//
// For objects, every `allOf` branch's own `properties` also contributes at
// this level (see [defaultSources]): unlike `anyOf`/`oneOf`/`not`, `allOf`
// has no ambiguity about which branch(es) apply — all of them must hold —
// and ajv's own codegen inlines allOf branches into the same validation
// function rather than compiling them as a separate call, so their declared
// defaults are applied exactly as if merged into this level (confirmed
// against ajv@6). A key not declared in any source's `properties` can still
// have its own nested defaults filled in when it's already present in v and
// matched by that source's `patternProperties` or a schema-valued
// `additionalProperties` — but such a source never manufactures a new key
// this way, matching ajv: only a schema's own declared `properties` (or,
// symmetrically, a tuple's own `items`) ever create a missing slot.
func applyDefaults(s *jsonschema.Schema, v any, doc any) any {
	switch val := v.(type) {
	case map[string]any:
		for _, src := range defaultSources(s) {
			for key, prop := range src.Properties {
				if _, ok := val[key]; !ok {
					def, ok := literalDefault(prop, doc)
					if !ok {
						continue
					}
					val[key] = normalizeNumbers(deepCopyJSON(def))
				}
				val[key] = applyDefaults(prop, val[key], doc)
			}
			for key, elem := range val {
				if _, declared := src.Properties[key]; declared {
					continue // already handled above
				}
				if propSchema, ok := matchAdditionalOrPattern(src, key); ok {
					val[key] = applyDefaults(propSchema, elem, doc)
				}
			}
		}
		return val
	case []any:
		switch items := resolveRef(s).Items.(type) {
		case []*jsonschema.Schema:
			origLen := len(val)
			for i, item := range items {
				if i < origLen {
					val[i] = applyDefaults(item, val[i], doc)
					continue
				}
				def, ok := literalDefault(item, doc)
				if !ok {
					// No default to contribute at this position: don't grow
					// the array on its account (ajv likewise emits no code
					// at all for a tuple position without a declared
					// default).
					continue
				}
				for len(val) < i {
					val = append(val, nil) // pad any not-yet-visited gap up to i
				}
				val = append(val, applyDefaults(item, normalizeNumbers(deepCopyJSON(def)), doc))
			}
		case *jsonschema.Schema:
			// List-style items: the same schema governs every element, so
			// each element already present gets its own defaults filled in
			// independently — but, matching ajv, `items` never determines
			// the array's length, so a too-short array is never padded on
			// its account (unlike a tuple position with its own literal
			// default, which does grow the array — see the case above).
			for i := range val {
				val[i] = applyDefaults(items, val[i], doc)
			}
		}
		return val
	default:
		return v
	}
}

// defaultSources returns, in order, every schema whose own `properties` (and
// `patternProperties`/`additionalProperties`) should contribute when filling
// in an object's defaults at this level: s itself (after following its
// `$ref` chain via [resolveRef]), plus, recursively, every `allOf` branch's
// own sources. See [applyDefaults] for why `allOf` — unlike
// `anyOf`/`oneOf`/`not` — is folded in this way.
//
// Cycle-guarded the same way [resolveRef] is: a pathological schema built
// entirely out of self-referential `$ref`/`allOf` wrappers that never
// bottoms out in a concrete schema must still terminate.
func defaultSources(s *jsonschema.Schema) []*jsonschema.Schema {
	return collectDefaultSources(s, map[*jsonschema.Schema]bool{})
}

func collectDefaultSources(s *jsonschema.Schema, seen map[*jsonschema.Schema]bool) []*jsonschema.Schema {
	s = resolveRef(s)
	if seen[s] {
		return nil
	}
	seen[s] = true
	sources := []*jsonschema.Schema{s}
	for _, branch := range s.AllOf {
		sources = append(sources, collectDefaultSources(branch, seen)...)
	}
	return sources
}

// matchAdditionalOrPattern returns the schema that governs key on s when key
// isn't declared as one of s's own named `properties` — first checking
// `patternProperties`, then falling back to `additionalProperties` when it's
// itself a schema (as opposed to a bare boolean or absent). Confirmed
// against ajv@6: an already-present key matched only this way still gets its
// own nested defaults filled in, even though [applyDefaults] never uses this
// to manufacture the key itself (see its caller, which only reaches here for
// keys already present in the input's map).
func matchAdditionalOrPattern(s *jsonschema.Schema, key string) (*jsonschema.Schema, bool) {
	for re, propSchema := range s.PatternProperties {
		if re.MatchString(key) {
			return propSchema, true
		}
	}
	if propSchema, ok := s.AdditionalProperties.(*jsonschema.Schema); ok {
		return propSchema, true
	}
	return nil, false
}

// literalDefault returns the value [applyDefaults] should insert into an
// absent slot governed by s, or false if s declares none.
//
// For an ordinary schema this is simply s.Default. But when s is a bare
// `$ref`, the compiled *jsonschema.Schema can never carry a literal default
// written directly beside that `$ref` (a "$ref sibling"): for Draft 4 (and
// earlier), the underlying jsonschema library follows the spec's own rule
// that every sibling keyword next to "$ref" — "default" included — MUST be
// ignored, and stops compiling the object entirely once it sees "$ref" (see
// that library's objcompiler.go, compileDraft4). ajv itself does not honor
// that part of the spec, though: confirmed against ajv@6 configured exactly
// like ESLint, a literal `default` written beside a `$ref` *is* used to fill
// a missing slot — using that literal value, not the ref target's own
// top-level default (which, as documented on [applyDefaults], is never
// pulled through either way). Since the compiled Schema has already lost
// this literal value, literalDefault recovers it by walking s's own
// Location — an absolute URL whose fragment is a JSON Pointer into doc, the
// document originally decoded by [compileSchemaJSON] — back into the raw
// JSON.
func literalDefault(s *jsonschema.Schema, doc any) (any, bool) {
	if s.Default != nil {
		return *s.Default, true
	}
	if s.Ref == nil {
		return nil, false
	}
	node, ok := lookupPointer(doc, locationFragment(s.Location))
	if !ok {
		return nil, false
	}
	obj, ok := node.(map[string]any)
	if !ok {
		return nil, false
	}
	def, ok := obj["default"]
	return def, ok
}

// locationFragment returns the JSON Pointer fragment of a compiled schema's
// Location — an absolute URL of the form "<base>#/json/pointer/segments".
func locationFragment(location string) string {
	if i := strings.IndexByte(location, '#'); i != -1 {
		return location[i+1:]
	}
	return ""
}

// lookupPointer resolves an RFC 6901 JSON Pointer against doc, the document
// originally decoded by [jsonschema.UnmarshalJSON] in [compileSchemaJSON].
func lookupPointer(doc any, pointer string) (any, bool) {
	cur := doc
	if pointer == "" {
		return cur, true
	}
	for _, tok := range strings.Split(strings.TrimPrefix(pointer, "/"), "/") {
		tok = strings.NewReplacer("~1", "/", "~0", "~").Replace(tok)
		switch v := cur.(type) {
		case map[string]any:
			next, ok := v[tok]
			if !ok {
				return nil, false
			}
			cur = next
		case []any:
			idx, err := strconv.Atoi(tok)
			if err != nil || idx < 0 || idx >= len(v) {
				return nil, false
			}
			cur = v[idx]
		default:
			return nil, false
		}
	}
	return cur, true
}

// resolveRef follows s's `$ref` chain, if any, to the schema whose own
// `Properties`/`Items` [applyDefaults] should walk — mirroring how ajv
// compiles a bare `{"$ref": ...}` schema into a call to the referenced
// schema's own compiled validator rather than inlining it. A `$ref` chain
// has exactly one target at each hop, so this is unconditional (unlike
// allOf/anyOf/oneOf/not, which [applyDefaults] deliberately never resolves).
//
// Cycle-guarded for a self-referential chain of pure-$ref wrapper schemas
// that never bottoms out in a concrete Properties/Items schema (e.g. two
// schemas that only $ref each other); a normal recursive schema (say, a
// tree node whose child items $ref back to the node's own object schema)
// terminates in one hop once it reaches that concrete schema, so this never
// affects it. It does not need to guard against recursing forever through
// applyDefaults itself, since that recursion walks v (finite JSON data),
// not s.
func resolveRef(s *jsonschema.Schema) *jsonschema.Schema {
	seen := map[*jsonschema.Schema]bool{}
	for s.Ref != nil && !seen[s] {
		seen[s] = true
		s = s.Ref
	}
	return s
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

// deepCopyJSON returns a copy of v in which every nested map[string]any and
// []any is freshly allocated, so mutating the result can never reach v
// itself. v is always either decoded JSON (map[string]any/[]any/string/
// bool/json.Number/nil) or, for *s.Default, a value produced by unmarshaling
// the schema's own JSON — never a struct or pointer type that would need
// special handling.
//
// applyDefaults must call this on every literalDefault before inserting it
// into the options being validated: literalDefault's result — s.Default or a
// value looked up from doc — is owned by the *Schema, which many goroutines
// share (every rule that reuses a schema, e.g. [EmptyArraySchema], and every
// concurrent Validate call on the same rule's schema). Inserting it directly
// would let one goroutine's applyDefaults recursion (or a later mutation of
// the validated options) write into a map/slice another goroutine is
// simultaneously reading or writing — a data race, and for a map a possible
// fatal concurrent-map-write crash.
func deepCopyJSON(v any) any {
	switch val := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, v2 := range val {
			out[k] = deepCopyJSON(v2)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, v2 := range val {
			out[i] = deepCopyJSON(v2)
		}
		return out
	default:
		return v
	}
}

// normalizeNumbers mutates v in place, converting any json.Number leaves to
// float64, and returns it (a bare top-level json.Number is instead returned
// as a new float64, since a number can't be mutated through its interface
// value). compileSchemaJSON decodes schema.json — and its "default" values —
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
