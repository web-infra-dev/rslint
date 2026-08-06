package rule

import "github.com/web-infra-dev/rslint/internal/utils"

// LatestECMAScriptVersion is the edition selected by ESLint's `latest`
// language option. Keep this and ecmaScriptGlobalIntroducedIn aligned with the
// ESLint version rslint targets (currently ESLint 10.8).
const LatestECMAScriptVersion = 2026

// ecmaScriptGlobalIntroducedIn mirrors ESLint's versioned conf/globals.js
// language sets. It deliberately excludes host globals, TypeScript-only types,
// and proposal names that ESLint has not exposed as runtime globals.
var ecmaScriptGlobalIntroducedIn = map[string]int{
	// ES3. ESLint intentionally includes inherited Object.prototype names.
	"Array":                3,
	"Boolean":              3,
	"constructor":          3,
	"Date":                 3,
	"decodeURI":            3,
	"decodeURIComponent":   3,
	"encodeURI":            3,
	"encodeURIComponent":   3,
	"Error":                3,
	"escape":               3,
	"eval":                 3,
	"EvalError":            3,
	"Function":             3,
	"hasOwnProperty":       3,
	"Infinity":             3,
	"isFinite":             3,
	"isNaN":                3,
	"isPrototypeOf":        3,
	"Math":                 3,
	"NaN":                  3,
	"Number":               3,
	"Object":               3,
	"parseFloat":           3,
	"parseInt":             3,
	"propertyIsEnumerable": 3,
	"RangeError":           3,
	"ReferenceError":       3,
	"RegExp":               3,
	"String":               3,
	"SyntaxError":          3,
	"toLocaleString":       3,
	"toString":             3,
	"TypeError":            3,
	"undefined":            3,
	"unescape":             3,
	"URIError":             3,
	"valueOf":              3,

	// ES5.
	"JSON": 5,

	// ES2015.
	"ArrayBuffer":       2015,
	"DataView":          2015,
	"Float32Array":      2015,
	"Float64Array":      2015,
	"Int16Array":        2015,
	"Int32Array":        2015,
	"Int8Array":         2015,
	"Intl":              2015,
	"Map":               2015,
	"Promise":           2015,
	"Proxy":             2015,
	"Reflect":           2015,
	"Set":               2015,
	"Symbol":            2015,
	"Uint16Array":       2015,
	"Uint32Array":       2015,
	"Uint8Array":        2015,
	"Uint8ClampedArray": 2015,
	"WeakMap":           2015,
	"WeakSet":           2015,

	// ES2017.
	"Atomics":           2017,
	"SharedArrayBuffer": 2017,

	// ES2020.
	"BigInt":         2020,
	"BigInt64Array":  2020,
	"BigUint64Array": 2020,
	"globalThis":     2020,

	// ES2021.
	"AggregateError":       2021,
	"FinalizationRegistry": 2021,
	"WeakRef":              2021,

	// ES2025.
	"Float16Array": 2025,
	"Iterator":     2025,

	// ES2026.
	"AsyncDisposableStack": 2026,
	"DisposableStack":      2026,
	"SuppressedError":      2026,
	"Temporal":             2026,
}

// Globals is the complete, immutable global-variable view exposed to native
// rules. It keeps each authored source separate while answering the effective
// access after applying ESLint's precedence:
//
//	ECMAScript edition < languageOptions.globals < inline /* global */ comments
//
// Its zero value represents the latest ECMAScript edition with no authored
// overrides.
type Globals struct {
	languageOptions    LanguageOptions
	configOverrides    map[string]utils.GlobalAccess
	inlineOverrides    map[string]utils.GlobalAccess
	inlineDeclarations []InlineGlobal
}

// NewGlobals constructs the per-file globals view. All maps and slices are
// treated as read-only and may be shared with the configuration and comment
// stores that produced them.
func NewGlobals(
	languageOptions LanguageOptions,
	configOverrides map[string]utils.GlobalAccess,
	inlineOverrides map[string]utils.GlobalAccess,
	inlineDeclarations []InlineGlobal,
) Globals {
	return Globals{
		languageOptions:    languageOptions,
		configOverrides:    configOverrides,
		inlineOverrides:    inlineOverrides,
		inlineDeclarations: inlineDeclarations,
	}
}

// LanguageAccess returns the globals implicitly supplied by the selected
// language options. ECMAScript globals are the only native source today and
// are read-only in ESLint's scope model. Future sourceType-provided global
// names compose here; wrapper-local bindings belong to the scope/RefStore
// layer instead.
func (g Globals) LanguageAccess(name string) utils.GlobalAccess {
	introduced, ok := ecmaScriptGlobalIntroducedIn[name]
	if ok && introduced <= g.languageOptions.EffectiveECMAVersion() {
		return utils.GlobalAccessReadonly
	}
	return utils.GlobalAccessUnset
}

// IsECMAScriptGlobalName reports whether name belongs to any ECMAScript
// edition known to this ESLint snapshot, independently of the edition selected
// for the current file. Rules whose own domain is the latest intrinsic-name
// set can combine this with Access, which still enforces ecmaVersion and
// authored overrides.
func (Globals) IsECMAScriptGlobalName(name string) bool {
	_, ok := ecmaScriptGlobalIntroducedIn[name]
	return ok
}

// ConfigOverride returns only the access authored in
// languageOptions.globals. It does not include language or inline globals.
func (g Globals) ConfigOverride(name string) utils.GlobalAccess {
	return g.configOverrides[name]
}

// Override returns the final explicitly authored access: an inline setting
// wins over languageOptions.globals. It excludes ECMAScript language globals.
func (g Globals) Override(name string) utils.GlobalAccess {
	if access, ok := g.inlineOverrides[name]; ok {
		return access
	}
	return g.configOverrides[name]
}

// ConfiguredAccess returns the effective access before inline comments are
// applied. This matches ESLint's internal configGlobals value for a name.
func (g Globals) ConfiguredAccess(name string) utils.GlobalAccess {
	if access, ok := g.configOverrides[name]; ok {
		return access
	}
	return g.LanguageAccess(name)
}

// Access returns the final effective access for name. Inline comments win over
// languageOptions.globals, which wins over the selected language edition.
func (g Globals) Access(name string) utils.GlobalAccess {
	if access, ok := g.inlineOverrides[name]; ok {
		return access
	}
	return g.ConfiguredAccess(name)
}

// InlineDeclarations returns every parsed inline global declaration in source
// order, including repeated names and their source ranges. Treat it as
// read-only.
func (g Globals) InlineDeclarations() []InlineGlobal {
	return g.inlineDeclarations
}

// ApplyTo applies this view to dst. Every known ECMAScript name is first set
// or removed according to the selected edition, preventing a TypeScript
// default-library seed from reintroducing a future-edition global. Explicit
// config and inline settings are then applied in precedence order.
func (g Globals) ApplyTo(dst map[string]bool) {
	for name := range ecmaScriptGlobalIntroducedIn {
		if g.Access(name).IsDeclared() {
			dst[name] = true
		} else {
			delete(dst, name)
		}
	}
	for name := range g.configOverrides {
		if g.Access(name).IsDeclared() {
			dst[name] = true
		} else {
			delete(dst, name)
		}
	}
	for name := range g.inlineOverrides {
		if g.Access(name).IsDeclared() {
			dst[name] = true
		} else {
			delete(dst, name)
		}
	}
}
