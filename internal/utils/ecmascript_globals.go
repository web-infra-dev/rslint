package utils

// ecmaScriptGlobals lists ECMAScript built-in global names used by ESLint core
// rules when an option asks them to consider implicit built-ins.
var ecmaScriptGlobals = map[string]bool{
	"AggregateError":       true,
	"Array":                true,
	"ArrayBuffer":          true,
	"AsyncDisposableStack": true,
	"AsyncIterator":        true,
	"Atomics":              true,
	"BigInt":               true,
	"BigInt64Array":        true,
	"BigUint64Array":       true,
	"Boolean":              true,
	"DataView":             true,
	"Date":                 true,
	"decodeURI":            true,
	"decodeURIComponent":   true,
	"DisposableStack":      true,
	"encodeURI":            true,
	"encodeURIComponent":   true,
	"Error":                true,
	"escape":               true,
	"EvalError":            true,
	"FinalizationRegistry": true,
	"Float32Array":         true,
	"Float64Array":         true,
	"Function":             true,
	"globalThis":           true,
	"Infinity":             true,
	"Int8Array":            true,
	"Int16Array":           true,
	"Int32Array":           true,
	"Intl":                 true,
	"isFinite":             true,
	"isNaN":                true,
	"Iterator":             true,
	"JSON":                 true,
	"Map":                  true,
	"Math":                 true,
	"NaN":                  true,
	"Number":               true,
	"Object":               true,
	"parseFloat":           true,
	"parseInt":             true,
	"Promise":              true,
	"Proxy":                true,
	"RangeError":           true,
	"ReferenceError":       true,
	"Reflect":              true,
	"RegExp":               true,
	"Set":                  true,
	"SharedArrayBuffer":    true,
	"String":               true,
	"SuppressedError":      true,
	"Symbol":               true,
	"SyntaxError":          true,
	"TypeError":            true,
	"Uint8Array":           true,
	"Uint8ClampedArray":    true,
	"Uint16Array":          true,
	"Uint32Array":          true,
	"unescape":             true,
	"URIError":             true,
	"undefined":            true,
	"WeakMap":              true,
	"WeakRef":              true,
	"WeakSet":              true,
}

// IsECMAScriptGlobal reports whether name is an ECMAScript built-in global.
func IsECMAScriptGlobal(name string) bool {
	return ecmaScriptGlobals[name]
}

// AddECMAScriptGlobals copies ECMAScript built-in global names into dst.
func AddECMAScriptGlobals(dst map[string]bool) {
	for name := range ecmaScriptGlobals {
		dst[name] = true
	}
}
