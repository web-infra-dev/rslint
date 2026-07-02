package no_deprecated_api

// deprecatedInfo describes a single deprecated API entry, mirroring upstream's
// `DeprecatedInfo` object ({ since, removed?, replacedBy }).
//
// replacedBy has three upstream shapes that map to two fields here:
//   - a plain string  -> replacedText (used verbatim, no version filtering)
//   - an array        -> replacedList (filtered by the configured node version)
//   - null            -> both empty (no "Use ... instead" suffix)
type deprecatedInfo struct {
	since        string
	removed      string // "" unless the API was removed (uses the "removed" messageId)
	replacedText string
	replacedList []replaceEntry
}

// replaceEntry is one `{ name, supported }` alternative in a replacedBy array.
type replaceEntry struct {
	name      string
	supported string
}

// traceMap is the Go equivalent of upstream's `DeprecatedInfoTraceMap`: a tree
// keyed by property name, where the READ/CALL/CONSTRUCT slots mark a terminal
// deprecated access of that kind.
type traceMap struct {
	read      *deprecatedInfo
	call      *deprecatedInfo
	construct *deprecatedInfo
	children  map[string]*traceMap
}

// --- deprecatedInfo constructors (keep the data tables below compact) ---

func depNull(since string) *deprecatedInfo { return &deprecatedInfo{since: since} }

func depRemoved(since, removed string) *deprecatedInfo {
	return &deprecatedInfo{since: since, removed: removed}
}

func depStr(since, replaced string) *deprecatedInfo {
	return &deprecatedInfo{since: since, replacedText: replaced}
}

func depList(since string, entries ...replaceEntry) *deprecatedInfo {
	return &deprecatedInfo{since: since, replacedList: entries}
}

func rep(name, supported string) replaceEntry { return replaceEntry{name: name, supported: supported} }

// rawModules mirrors upstream lib/rules/no-deprecated-api.js `rawModules`
// (the CommonJS / ESM module deprecation table) 1:1.
var rawModules = map[string]*traceMap{
	"_linklist":    {read: depNull("5.0.0")},
	"_stream_wrap": {read: depNull("12.0.0")},
	"async_hooks": {children: map[string]*traceMap{
		"currentId": {read: depList("8.2.0", rep("'async_hooks.executionAsyncId()'", "8.1.0"))},
		"triggerId": {read: depStr("8.2.0", "'async_hooks.triggerAsyncId()'")},
	}},
	"buffer": {children: map[string]*traceMap{
		"Buffer": {
			construct: depList("6.0.0", rep("'buffer.Buffer.alloc()'", "5.10.0"), rep("'buffer.Buffer.from()'", "5.10.0")),
			call:      depList("6.0.0", rep("'buffer.Buffer.alloc()'", "5.10.0"), rep("'buffer.Buffer.from()'", "5.10.0")),
		},
		"SlowBuffer": {read: depList("6.0.0", rep("'buffer.Buffer.allocUnsafeSlow()'", "5.12.0"))},
	}},
	"constants": {read: depStr("6.3.0", "'constants' property of each module")},
	"crypto": {children: map[string]*traceMap{
		"_toBuf":            {read: depNull("11.0.0")},
		"Credentials":       {read: depStr("0.12.0", "'tls.SecureContext'")},
		"DEFAULT_ENCODING":  {read: depNull("10.0.0")},
		"createCipher":      {read: depList("10.0.0", rep("'crypto.createCipheriv()'", "0.1.94"))},
		"createCredentials": {read: depList("0.12.0", rep("'tls.createSecureContext()'", "0.11.13"))},
		"createDecipher":    {read: depList("10.0.0", rep("'crypto.createDecipheriv()'", "0.1.94"))},
		"fips":              {read: depList("10.0.0", rep("'crypto.getFips()' and 'crypto.setFips()'", "10.0.0"))},
		"prng":              {read: depList("11.0.0", rep("'crypto.randomBytes()'", "0.5.8"))},
		"pseudoRandomBytes": {read: depList("11.0.0", rep("'crypto.randomBytes()'", "0.5.8"))},
		"rng":               {read: depList("11.0.0", rep("'crypto.randomBytes()'", "0.5.8"))},
	}},
	"domain": {read: depNull("4.0.0")},
	"events": {children: map[string]*traceMap{
		"EventEmitter": {children: map[string]*traceMap{
			"listenerCount": {read: depList("4.0.0", rep("'events.EventEmitter#listenerCount()'", "3.2.0"))},
		}},
		"listenerCount": {read: depList("4.0.0", rep("'events.EventEmitter#listenerCount()'", "3.2.0"))},
	}},
	"freelist": {read: depNull("4.0.0")},
	"fs": {children: map[string]*traceMap{
		"SyncWriteStream": {read: depNull("4.0.0")},
		"exists":          {read: depList("4.0.0", rep("'fs.stat()'", "0.0.2"), rep("'fs.access()'", "0.11.15"))},
		"lchmod":          {read: depNull("0.4.0")},
		"lchmodSync":      {read: depNull("0.4.0")},
	}},
	"http": {children: map[string]*traceMap{
		"createClient": {read: depList("0.10.0", rep("'http.request()'", "0.3.6"))},
	}},
	"module": {children: map[string]*traceMap{
		"Module": {children: map[string]*traceMap{
			"createRequireFromPath": {read: depList("12.2.0", rep("'module.createRequire()'", "12.2.0"))},
			"requireRepl":           {read: depStr("6.0.0", "'require(\"repl\")'")},
			"_debug":                {read: depNull("9.0.0")},
		}},
		"createRequireFromPath": {read: depList("12.2.0", rep("'module.createRequire()'", "12.2.0"))},
		"requireRepl":           {read: depStr("6.0.0", "'require(\"repl\")'")},
		"_debug":                {read: depNull("9.0.0")},
	}},
	"net": {children: map[string]*traceMap{
		"_setSimultaneousAccepts": {read: depNull("12.0.0")},
	}},
	"os": {children: map[string]*traceMap{
		"getNetworkInterfaces": {read: depList("0.6.0", rep("'os.networkInterfaces()'", "0.6.0"))},
		"tmpDir":               {read: depList("7.0.0", rep("'os.tmpdir()'", "0.9.9"))},
	}},
	"path": {children: map[string]*traceMap{
		"_makeLong": {read: depList("9.0.0", rep("'path.toNamespacedPath()'", "9.0.0"))},
	}},
	"process": {children: map[string]*traceMap{
		"EventEmitter": {read: depStr("0.6.0", "'require(\"events\")'")},
		"assert":       {read: depStr("10.0.0", "'require(\"assert\")'")},
		"binding":      {read: depNull("10.9.0")},
		"env": {children: map[string]*traceMap{
			"NODE_REPL_HISTORY_FILE": {read: depStr("4.0.0", "'NODE_REPL_HISTORY'")},
		}},
		"report": {children: map[string]*traceMap{
			"triggerReport": {read: depStr("11.12.0", "'process.report.writeReport()'")},
		}},
	}},
	"punycode": {read: depStr("7.0.0", "'https://www.npmjs.com/package/punycode'")},
	"readline": {children: map[string]*traceMap{
		"codePointAt":              {read: depNull("4.0.0")},
		"getStringWidth":           {read: depNull("6.0.0")},
		"isFullWidthCodePoint":     {read: depNull("6.0.0")},
		"stripVTControlCharacters": {read: depNull("6.0.0")},
	}},
	"repl": {children: map[string]*traceMap{
		"REPLServer":      {read: depStr("22.9.0", "new repl.REPLServer()")},
		"Recoverable":     {read: depStr("22.9.0", "new repl.Recoverable()")},
		"REPL_MODE_MAGIC": {read: depNull("8.0.0")},
		"builtinModules":  {read: depStr("22.16.0", "module.builtinModules")},
	}},
	// safe-buffer.Buffer function/constructor is just a re-export of buffer.Buffer
	// and should be deprecated likewise.
	"safe-buffer": {children: map[string]*traceMap{
		"Buffer": {
			construct: depList("6.0.0", rep("'buffer.Buffer.alloc()'", "5.10.0"), rep("'buffer.Buffer.from()'", "5.10.0")),
			call:      depList("6.0.0", rep("'buffer.Buffer.alloc()'", "5.10.0"), rep("'buffer.Buffer.from()'", "5.10.0")),
		},
		"SlowBuffer": {read: depList("6.0.0", rep("'buffer.Buffer.allocUnsafeSlow()'", "5.12.0"))},
	}},
	"sys": {read: depStr("0.3.0", "'util' module")},
	"timers": {children: map[string]*traceMap{
		"enroll":   {read: depList("10.0.0", rep("'setTimeout()'", "0.0.1"), rep("'setInterval()'", "0.0.1"))},
		"unenroll": {read: depList("10.0.0", rep("'clearTimeout()'", "0.0.1"), rep("'clearInterval()'", "0.0.1"))},
	}},
	"tls": {children: map[string]*traceMap{
		"CleartextStream":     {read: depNull("0.10.0")},
		"CryptoStream":        {read: depList("0.12.0", rep("'tls.TLSSocket'", "0.11.4"))},
		"SecurePair":          {read: depList("6.0.0", rep("'tls.TLSSocket'", "0.11.4"))},
		"convertNPNProtocols": {read: depNull("10.0.0")},
		"createSecurePair":    {read: depList("6.0.0", rep("'tls.TLSSocket'", "0.11.4"))},
		"parseCertString":     {read: depList("8.6.0", rep("'querystring.parse()'", "0.1.25"))},
	}},
	"tty": {children: map[string]*traceMap{
		"setRawMode": {read: depStr("0.10.0", "'tty.ReadStream#setRawMode()' (e.g. 'process.stdin.setRawMode()')")},
	}},
	"url": {children: map[string]*traceMap{
		"parse":   {read: depList("11.0.0", rep("'url.URL' constructor", "6.13.0"))},
		"resolve": {read: depList("11.0.0", rep("'url.URL' constructor", "6.13.0"))},
	}},
	"util": {children: map[string]*traceMap{
		"debug":             {read: depList("0.12.0", rep("'console.error()'", "0.1.100"))},
		"error":             {read: depList("0.12.0", rep("'console.error()'", "0.1.100"))},
		"isArray":           {read: depList("4.0.0", rep("'Array.isArray()'", "0.1.100"))},
		"isBoolean":         {read: depNull("4.0.0")},
		"isBuffer":          {read: depList("4.0.0", rep("'Buffer.isBuffer()'", "0.1.101"))},
		"isDate":            {read: depNull("4.0.0")},
		"isError":           {read: depNull("4.0.0")},
		"isFunction":        {read: depNull("4.0.0")},
		"isNull":            {read: depNull("4.0.0")},
		"isNullOrUndefined": {read: depNull("4.0.0")},
		"isNumber":          {read: depNull("4.0.0")},
		"isObject":          {read: depNull("4.0.0")},
		"isPrimitive":       {read: depNull("4.0.0")},
		"isRegExp":          {read: depNull("4.0.0")},
		"isString":          {read: depNull("4.0.0")},
		"isSymbol":          {read: depNull("4.0.0")},
		"isUndefined":       {read: depNull("4.0.0")},
		"log":               {read: depStr("6.0.0", "a third party module")},
		"print":             {read: depList("0.12.0", rep("'console.log()'", "0.1.100"))},
		"pump":              {read: depList("0.10.0", rep("'stream.Readable#pipe()'", "0.9.4"))},
		"puts":              {read: depList("0.12.0", rep("'console.log()'", "0.1.100"))},
		"_extend":           {read: depList("6.0.0", rep("'Object.assign()'", "4.0.0"))},
	}},
	"vm": {children: map[string]*traceMap{
		"runInDebugContext": {read: depNull("8.0.0")},
	}},
	"zlib": {children: map[string]*traceMap{
		"BrotliCompress":   {call: depStr("22.9.0", "new zlib.BrotliCompress()")},
		"BrotliDecompress": {call: depStr("22.9.0", "new zlib.BrotliDecompress()")},
		"Deflate":          {call: depStr("22.9.0", "new zlib.Deflate()")},
		"DeflateRaw":       {call: depStr("22.9.0", "new zlib.DeflateRaw()")},
		"Gunzip":           {call: depStr("22.9.0", "new zlib.Gunzip()")},
		"Gzip":             {call: depStr("22.9.0", "new zlib.Gzip()")},
		"Inflate":          {call: depStr("22.9.0", "new zlib.Inflate()")},
		"InflateRaw":       {call: depStr("22.9.0", "new zlib.InflateRaw()")},
		"Unzip":            {call: depStr("22.9.0", "new zlib.Unzip()")},
	}},
}

// modules is rawModules extended with `node:`-prefixed aliases for builtins.
var modules = extendTraceMapWithNodePrefix(rawModules)

// globals mirrors upstream's `globals` table (deprecated global variables).
// `process` reuses the module trace map so `process.binding` etc. are detected
// as global member accesses too.
var globals = map[string]*traceMap{
	"Buffer": {
		construct: depList("6.0.0", rep("'Buffer.alloc()'", "5.10.0"), rep("'Buffer.from()'", "5.10.0")),
		call:      depList("6.0.0", rep("'Buffer.alloc()'", "5.10.0"), rep("'Buffer.from()'", "5.10.0")),
	},
	"COUNTER_NET_SERVER_CONNECTION":       {read: depNull("11.0.0")},
	"COUNTER_NET_SERVER_CONNECTION_CLOSE": {read: depNull("11.0.0")},
	"COUNTER_HTTP_SERVER_REQUEST":         {read: depNull("11.0.0")},
	"COUNTER_HTTP_SERVER_RESPONSE":        {read: depNull("11.0.0")},
	"COUNTER_HTTP_CLIENT_REQUEST":         {read: depNull("11.0.0")},
	"COUNTER_HTTP_CLIENT_RESPONSE":        {read: depNull("11.0.0")},
	"GLOBAL":                              {read: depList("6.0.0", rep("'global'", "0.1.27"))},
	"Intl": {children: map[string]*traceMap{
		"v8BreakIterator": {read: depRemoved("7.0.0", "9.0.0")},
	}},
	"require": {children: map[string]*traceMap{
		"extensions": {read: depStr("0.12.0", "compiling them ahead of time")},
	}},
	"root":    {read: depList("6.0.0", rep("'global'", "0.1.27"))},
	"process": rawModules["process"],
}

// nodeBuiltinModules is the set of Node.js builtin module names, used as the
// `module.isBuiltin` equivalent for extendTraceMapWithNodePrefix. Names with a
// leading underscore, removed modules (`sys`, `freelist`), and third-party
// re-exports (`safe-buffer`) are intentionally absent — they get no `node:`
// alias, matching upstream's `isBuiltin` filter.
var nodeBuiltinModules = map[string]bool{
	"assert": true, "assert/strict": true, "async_hooks": true, "buffer": true,
	"child_process": true, "cluster": true, "console": true, "constants": true,
	"crypto": true, "dgram": true, "diagnostics_channel": true, "dns": true,
	"dns/promises": true, "domain": true, "events": true, "fs": true,
	"fs/promises": true, "http": true, "http2": true, "https": true,
	"inspector": true, "inspector/promises": true, "module": true, "net": true,
	"os": true, "path": true, "path/posix": true, "path/win32": true,
	"perf_hooks": true, "process": true, "punycode": true, "querystring": true,
	"readline": true, "readline/promises": true, "repl": true, "stream": true,
	"stream/consumers": true, "stream/promises": true, "stream/web": true,
	"string_decoder": true, "timers": true, "timers/promises": true, "tls": true,
	"trace_events": true, "tty": true, "url": true, "util": true,
	"util/types": true, "v8": true, "vm": true, "wasi": true,
	"worker_threads": true, "zlib": true,
}

// extendTraceMapWithNodePrefix mirrors upstream's util of the same name: for
// every builtin module key it adds a `node:`-prefixed alias pointing at the same
// trace map (so `require('node:buffer')` is treated like `require('buffer')`).
func extendTraceMapWithNodePrefix(src map[string]*traceMap) map[string]*traceMap {
	out := make(map[string]*traceMap, len(src)*2)
	for name, value := range src {
		out[name] = value
		if nodeBuiltinModules[name] {
			out["node:"+name] = value
		}
	}
	return out
}

// unprefixNodeColon removes a leading `node:` from a module name for reporting,
// mirroring upstream's util of the same name.
func unprefixNodeColon(name string) string {
	if len(name) > 5 && name[:5] == "node:" {
		return name[5:]
	}
	return name
}
