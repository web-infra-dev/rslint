package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
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

// RstestPluginManagedCall is a call to one of the members Rstest's build
// rewrites, written on the utilities object: `rs.mock(...)`,
// `rstest.hoisted(...)`, `rs['importActual'](...)`, and so on.
type RstestPluginManagedCall struct {
	// Namespace is the receiver as it is written, "rs" or "rstest".
	Namespace string
	// NamespaceNode is the identifier that names the receiver. A caller whose
	// member has API.ResolvesReceiver set passes it to
	// ReceiverIsLocallyDeclared to find out whether this file's own
	// declaration takes the call away from the rewrite.
	NamespaceNode *ast.Node
	// Member is the property name as it is written.
	Member string
	// MemberNode is the node naming the member: the identifier of a dotted
	// member, or the string literal inside the brackets of a computed one.
	MemberNode *ast.Node
	// API is how the build matches this member, from the shared table.
	API PluginManagedAPI
}

// ParseRstestPluginManagedCall matches a call on the utilities object by the
// syntax at the call site alone. It never resolves a binding and never consults
// the TypeChecker, because the rewrite this describes does not either.
//
// It matches only the members in the plugin-managed table, and matches each
// one the way the build matches it. Everything else on the object — `fn`,
// `spyOn`, `mocked`, the timer and stub helpers — is an ordinary function
// reached through an ordinary binding, so a renamed import still works and a
// rule should resolve the receiver through IsUtilitiesObject rather than parse
// it here.
//
// The receiver is read as written because the rewrite reads it as written:
// `rs.mock('./m')` is rewritten in a file whose only `rs` is a local object
// literal and in a file that re-exports `rs` from a helper module, while the
// same call through a renamed binding — `import { rs as r }`, then
// `r.mock('./m')` — is not rewritten at all and throws when it runs. Resolving
// the receiver would therefore report calls the build leaves alone and stay
// silent on calls it rewrites. The two `*Actual` members are the exception, and
// they carry ResolvesReceiver so their callers can ask.
//
// Parentheses and TypeScript's type-only syntax are transparent on both sides
// of the callee: `as`, `satisfies` and `!` are erased before the rewrite runs,
// and it reads through parentheses, so `(rs as any).mock()`, `rs!.mock()` and
// `(rs.mock)()` are all rewritten like the bare form.
//
// The result is nil for anything else. A caller still has to decide what its
// own API accepts: which member names it cares about, where the call may stand,
// and what arguments it takes are all per-API and stay with the caller. One
// shape deliberately left to callers is the surrounding expression: `void
// rs.importActual('./m')` is rewritten while `void rs.mock('./m')` is not,
// which is a question about where the call stands rather than about its callee.
func ParseRstestPluginManagedCall(node *ast.Node) *RstestPluginManagedCall {
	if node == nil || node.Kind != ast.KindCallExpression {
		return nil
	}
	call := node.AsCallExpression()
	if call == nil {
		return nil
	}

	callee := internalUtils.SkipAssertionsAndParens(call.Expression)
	if callee == nil {
		return nil
	}

	var member string
	var memberNode, namespaceNode *ast.Node
	computed := false
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		// An optional receiver, `rs?.mock('./m')`, is matched for no member.
		if access == nil || access.QuestionDotToken != nil {
			return nil
		}
		name := access.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return nil
		}
		member, memberNode, namespaceNode = name.AsIdentifier().Text, name, access.Expression
	case ast.KindElementAccessExpression:
		access := callee.AsElementAccessExpression()
		if access == nil || access.QuestionDotToken != nil {
			return nil
		}
		// Only a plain quoted string names a member here. A template literal
		// is not matched even without a substitution, and an identifier or a
		// substituted template names something known at run time.
		key := internalUtils.SkipAssertionsAndParens(access.ArgumentExpression)
		if key == nil || key.Kind != ast.KindStringLiteral {
			return nil
		}
		member, memberNode, namespaceNode = key.Text(), key, access.Expression
		computed = true
	default:
		return nil
	}

	api, ok := LookupPluginManagedAPI(member)
	if !ok {
		return nil
	}
	if computed && !api.ReadsComputedMember {
		return nil
	}
	if call.QuestionDotToken != nil && !api.ReadsOptionalCall {
		return nil
	}

	namespaceNode = internalUtils.SkipAssertionsAndParens(namespaceNode)
	if namespaceNode == nil || namespaceNode.Kind != ast.KindIdentifier {
		return nil
	}
	namespace := namespaceNode.AsIdentifier().Text
	if !utilityNamespaces[namespace] {
		return nil
	}

	return &RstestPluginManagedCall{
		Namespace:     namespace,
		NamespaceNode: namespaceNode,
		Member:        member,
		MemberNode:    memberNode,
		API:           api,
	}
}
