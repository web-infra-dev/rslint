package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

// utilityNamespaces are the two names the utilities object is written under.
// `@rstest/core` exports one object twice, as `rstest` and as `rs`
// (packages/core/src/runtime/api/public.ts), and registers both onto
// `globalThis` under `globals: true`.
var utilityNamespaces = map[string]bool{
	"rs":     true,
	"rstest": true,
}

// pluginManagedAPIs are the members of the utilities object that never run as
// written. Each one is a stub that throws
// "[Rstest] <name>() was not transformed by Rstest"
// (packages/core/src/runtime/api/utilities.ts, `createPluginManagedApi`); the
// call that does the work is emitted by Rstest's build, which rewrites the call
// site itself.
//
// The distinction decides how a rule may recognize a call. Everything else on
// the object — `fn`, `spyOn`, `mocked`, the timer and stub helpers — is an
// ordinary function reached through an ordinary binding, so a renamed import
// still works and a rule should resolve the receiver. These fourteen are
// matched by the build on the name written at the call site, so a rule must
// read the receiver instead of resolving it: see ParseRstestUtilityCall.
var pluginManagedAPIs = map[string]bool{
	"mock":            true,
	"mockRequire":     true,
	"doMock":          true,
	"doMockRequire":   true,
	"unmock":          true,
	"doUnmock":        true,
	"unmockRequire":   true,
	"doUnmockRequire": true,
	"importMock":      true,
	"requireMock":     true,
	"importActual":    true,
	"requireActual":   true,
	"resetModules":    true,
	"hoisted":         true,
}

// IsPluginManagedAPI reports whether member names one of the utilities-object
// members Rstest's build rewrites rather than calls.
func IsPluginManagedAPI(member string) bool {
	return pluginManagedAPIs[member]
}

// RstestUtilityCall is a call written as a plain member of the utilities
// object: `rs.mock(...)`, `rstest.hoisted(...)`, and so on.
type RstestUtilityCall struct {
	// Namespace is the receiver as it is written, "rs" or "rstest".
	Namespace     string
	NamespaceNode *ast.Node
	// Member is the property name as it is written.
	Member     string
	MemberNode *ast.Node
	Call       *ast.Node
}

// ParseRstestUtilityCall matches a call on the utilities object by the syntax
// at the call site alone. It never resolves a binding and never consults the
// TypeChecker, because the builds this shape describes do not either.
//
// Rstest's mock transform matches the receiver by the name written in the
// source: `rs.mock('./m')` is rewritten in a file whose only `rs` is a local
// object literal and in a file that re-exports `rs` from a helper module, while
// the same call written through a renamed binding — `import { rs as r }`, then
// `r.mock('./m')` — is not rewritten at all and throws when it runs. Resolving
// the receiver would therefore report calls the build leaves alone and stay
// silent on calls it rewrites.
//
// Parentheses and TypeScript's type-only syntax are transparent on both sides
// of the callee: `as`, `satisfies` and `!` are erased before the transform
// runs, and it reads through parentheses, so `(rs as any).mock()`, `rs!.mock()`
// and `(rs.mock)()` are all rewritten like the bare form. A computed member and
// an optional chain are not matched here; the shapes the build accepts vary by
// API, and the narrower set is common to all of them.
//
// The result is nil for anything else. A caller still has to decide what its
// own API accepts: which member names it cares about, where the call may stand,
// and what arguments it takes are all per-API and stay with the caller.
func ParseRstestUtilityCall(node *ast.Node) *RstestUtilityCall {
	if node == nil || node.Kind != ast.KindCallExpression || ast.IsOptionalChain(node) {
		return nil
	}
	call := node.AsCallExpression()
	if call == nil || call.QuestionDotToken != nil {
		return nil
	}

	callee := internalUtils.SkipAssertionsAndParens(call.Expression)
	if callee == nil ||
		callee.Kind != ast.KindPropertyAccessExpression ||
		ast.IsOptionalChain(callee) {
		return nil
	}
	access := callee.AsPropertyAccessExpression()
	if access == nil || access.QuestionDotToken != nil {
		return nil
	}

	memberNode := access.Name()
	if memberNode == nil || memberNode.Kind != ast.KindIdentifier {
		return nil
	}

	namespaceNode := internalUtils.SkipAssertionsAndParens(access.Expression)
	if namespaceNode == nil || namespaceNode.Kind != ast.KindIdentifier {
		return nil
	}
	namespace := namespaceNode.AsIdentifier().Text
	if !utilityNamespaces[namespace] {
		return nil
	}

	return &RstestUtilityCall{
		Namespace:     namespace,
		NamespaceNode: namespaceNode,
		Member:        memberNode.AsIdentifier().Text,
		MemberNode:    memberNode,
		Call:          node,
	}
}
