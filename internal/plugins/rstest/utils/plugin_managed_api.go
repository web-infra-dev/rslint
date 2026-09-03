package utils

// PluginManagedAPI describes how Rstest's build matches one member of the
// utilities object at the call site.
//
// The fourteen members listed in pluginManagedAPIs are the ones built by
// `createPluginManagedApi` in packages/core/src/runtime/api/utilities.ts. None
// of them runs as written: each is a stub that throws
// "[Rstest] <name>() was not transformed by Rstest", and the call that does the
// work is emitted by the native plugin, which rewrites the call site during the
// build. So whether such a call means anything is a question about the syntax
// at that call site, not about the value the receiver holds at run time.
//
// The rewrite does not read every member the same way, and the difference is
// not cosmetic: a shape outside what a member's rewrite matches is left as the
// throwing stub. A rule that reports the wider set reports calls that never
// run; a rule that reports the narrower set stays silent on calls that do.
type PluginManagedAPI struct {
	// ReadsComputedMember allows the member to be named by a bracketed string
	// literal, `rs['importActual']('./m')`. A template literal is never
	// matched, for any member, even without a substitution.
	ReadsComputedMember bool
	// ReadsOptionalCall allows an optional call, `rs.importActual?.('./m')`.
	// An optional receiver, `rs?.importActual('./m')`, is matched for no
	// member.
	ReadsOptionalCall bool
	// ResolvesReceiver makes a local declaration of the receiver name shadow
	// the rewrite, so `const rs = { importActual: … }` runs its own object.
	// When false the receiver is matched by the name as written and a local
	// declaration is bypassed: `const rs = { mock: … }` in a file Rstest
	// bundles still has its `rs.mock('./m')` rewritten, and the local object's
	// `mock` never runs.
	ResolvesReceiver bool
}

// pluginManagedAPIs maps each member the build rewrites onto the shapes that
// rewrite matches for it.
//
// The two entries that differ are `importActual` and `requireActual`. They are
// the only members that hand back the real module rather than registering
// something in the mock registry, and the rewrite turns them into an ordinary
// module reference, which is why it can afford to read a wider set of call
// shapes and to respect an ordinary binding. Every other member is rewritten on
// the name as written and nothing else — `importMock` and `requireMock`
// included, despite loading a module the same way.
//
// Measured on @rstest/core 0.11.11 by writing each shape at the top level of a
// test file and reading whether the stub threw, one file per shape:
//
//	rs.importActual('./m')      rewritten    rs.mock('./m')      rewritten
//	rs['importActual']('./m')   rewritten    rs['mock']('./m')   throws
//	rs[`importActual`]('./m')   throws       rs[`mock`]('./m')   throws
//	rs.importActual?.('./m')    rewritten    rs.mock?.('./m')    throws
//	rs?.importActual('./m')     throws       rs?.mock('./m')     throws
var pluginManagedAPIs = map[string]PluginManagedAPI{
	"mock":            {},
	"mockRequire":     {},
	"doMock":          {},
	"doMockRequire":   {},
	"unmock":          {},
	"doUnmock":        {},
	"unmockRequire":   {},
	"doUnmockRequire": {},
	"importMock":      {},
	"requireMock":     {},
	"resetModules":    {},
	"hoisted":         {},

	"importActual":  {ReadsComputedMember: true, ReadsOptionalCall: true, ResolvesReceiver: true},
	"requireActual": {ReadsComputedMember: true, ReadsOptionalCall: true, ResolvesReceiver: true},
}

// LookupPluginManagedAPI returns how the build matches member, and whether
// member is plugin-managed at all.
//
// A rule whose member names come from user configuration cannot know at compile
// time which of the two kinds it will be handed, so it has to ask here for each
// member it is given.
func LookupPluginManagedAPI(member string) (PluginManagedAPI, bool) {
	api, ok := pluginManagedAPIs[member]
	return api, ok
}

// IsPluginManagedAPI reports whether member is rewritten by Rstest's build, and
// so has to be matched by the syntax at the call site rather than by resolving
// the receiver.
//
// Everything else on the utilities object — `fn`, `spyOn`, `mocked`, the timer
// and stub helpers — is an ordinary function reached through an ordinary
// binding, where a renamed import is the same function and a local declaration
// really is a different object. IsUtilitiesObject resolves those.
func IsPluginManagedAPI(member string) bool {
	_, ok := pluginManagedAPIs[member]
	return ok
}
