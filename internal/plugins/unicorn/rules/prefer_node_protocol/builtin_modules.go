package prefer_node_protocol

// builtinModuleNames is the set of Node.js builtin module names that have a
// `node:`-prefixed form, mirroring the `is-builtin-module` package that
// eslint-plugin-unicorn's prefer-node-protocol depends on. Upstream
// deliberately uses this static list (from `builtin-modules`) instead of Node's
// runtime `module.isBuiltin` so the result is stable across Node.js versions.
//
// Derived from builtin-modules@5.3.0 (builtin-modules.json), the version
// resolved by is-builtin-module@5's `builtin-modules: "^5.0.0"` dependency
// range. The rule only ever inspects specifiers WITHOUT the `node:` prefix and
// rewrites them to `node:<name>`, so this set keeps only the bare names that
// have a valid `node:` counterpart. Names that exist solely under the `node:`
// prefix (e.g. `test`, `sqlite`, `sea`, `quic`, `ffi`, `vfs`) are intentionally
// absent: bare `import "test"` must not be flagged because bare `test` is not a
// builtin module.
var builtinModuleNames = map[string]bool{
	"assert":              true,
	"assert/strict":       true,
	"async_hooks":         true,
	"buffer":              true,
	"child_process":       true,
	"cluster":             true,
	"console":             true,
	"constants":           true,
	"crypto":              true,
	"dgram":               true,
	"diagnostics_channel": true,
	"dns":                 true,
	"dns/promises":        true,
	"domain":              true,
	"events":              true,
	"fs":                  true,
	"fs/promises":         true,
	"http":                true,
	"http2":               true,
	"https":               true,
	"inspector":           true,
	"inspector/promises":  true,
	"module":              true,
	"net":                 true,
	"os":                  true,
	"path":                true,
	"path/posix":          true,
	"path/win32":          true,
	"perf_hooks":          true,
	"process":             true,
	"querystring":         true,
	"readline":            true,
	"readline/promises":   true,
	"repl":                true,
	"stream":              true,
	"stream/consumers":    true,
	"stream/iter":         true,
	"stream/promises":     true,
	"stream/web":          true,
	"string_decoder":      true,
	"timers":              true,
	"timers/promises":     true,
	"tls":                 true,
	"trace_events":        true,
	"tty":                 true,
	"url":                 true,
	"util":                true,
	"util/types":          true,
	"v8":                  true,
	"vm":                  true,
	"wasi":                true,
	"worker_threads":      true,
	"zlib":                true,
	"zlib/iter":           true,
}
