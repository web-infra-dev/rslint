package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

// IsUtilitiesObject reports whether receiver names Rstest's utilities object.
//
// This is how the members that are NOT plugin-managed are matched. `fn` is an
// ordinary function on an ordinary object, so it is reached through ordinary
// bindings: a renamed import is the same function, and a local declaration of
// `rs` really is a different object. That is the opposite of how the
// plugin-managed members are matched by ParseRstestPluginManagedCall, which
// reads the receiver as written.
func IsUtilitiesObject(ctx rule.RuleContext, receiver *ast.Node) bool {
	receiver = internalUtils.SkipAssertionsAndParens(receiver)
	if receiver == nil {
		return false
	}

	if receiver.Kind == ast.KindIdentifier {
		// `rs` and `rstest` name the same object
		// (packages/core/src/runtime/api/public.ts), either may be imported
		// under a further name, and under `globals: true` both are on
		// `globalThis` with no import at all. A local declaration of the name
		// shadows both, and resolution reports that by returning no name.
		//
		// ctx.Refs.Resolve places an import and a local declaration from the
		// binder alone, so a file that declares the name is recognized whether
		// or not a TypeChecker is configured.
		name, _, _ := testFramework.ResolveFunctionIdentifierReferenceFromSymbolModules(
			receiver.AsIdentifier().Text,
			receiver,
			ctx.Refs.Resolve(receiver),
			ctx.SourceFile,
			RstestCoreImportModules,
		)
		return name == "rs" || name == "rstest"
	}

	// A namespace import, a whole-module require and `import.meta.rstest` all
	// reach the same object through one more member: `core.rs.fn()`,
	// `import.meta.rstest.rs.fn()`. The last one is the module namespace
	// itself (packages/core/importMeta.d.ts types it as
	// `typeof import('@rstest/core')`), so the utilities object is the `rs` or
	// `rstest` on it rather than the property itself.
	member, namespace := CalledPlainMember(receiver)
	if member == nil || (member.Text() != "rs" && member.Text() != "rstest") {
		return false
	}
	namespace = internalUtils.SkipAssertionsAndParens(namespace)
	if namespace == nil {
		return false
	}
	if isImportMetaRstest(namespace) {
		return true
	}
	if namespace.Kind != ast.KindIdentifier {
		return false
	}
	return testFramework.IsModuleNamespaceSymbolModules(
		ctx.Refs.Resolve(namespace),
		RstestCoreImportModules,
	)
}

// ReceiverIsLocallyDeclared reports whether the receiver name is declared in
// this file by something other than an import.
//
// It answers the question the members carrying PluginManagedAPI.ResolvesReceiver
// ask. Those two are left as the runtime stub when the receiver is a local
// binding: a file that declares `const rs = { importActual: … }` — or takes
// `rs` as a parameter — runs its own object, while a file that reaches `rs`
// through an import runs the rewrite, whichever module the import came from. So
// an import of any kind is not a local declaration here, and every other
// declaration is.
//
// The members without that flag do not ask: the build rewrites them on the name
// as written, so a local declaration of `rs` is bypassed rather than honored.
func ReceiverIsLocallyDeclared(ctx rule.RuleContext, receiver *ast.Node) bool {
	symbol := ctx.Refs.Resolve(receiver)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil || ast.GetSourceFileOfNode(declaration) != ctx.SourceFile {
			continue
		}
		switch declaration.Kind {
		case ast.KindImportSpecifier,
			ast.KindImportClause,
			ast.KindNamespaceImport,
			ast.KindImportEqualsDeclaration:
			continue
		}
		return true
	}
	return false
}

// CalledPlainMember returns the property name a callee reaches its member
// through, and the expression it is read off. Both are nil when the callee is
// anything but a plain dotted member.
//
// A computed member is not matched: it is written to reach a property whose
// name is not a fixed identifier, and reporting the string inside the brackets
// would point at a value rather than at a member. The plugin-managed members
// that do accept a bracketed string are read by ParseRstestPluginManagedCall,
// which knows which ones those are.
func CalledPlainMember(callee *ast.Node) (*ast.Node, *ast.Node) {
	callee = internalUtils.SkipAssertionsAndParens(callee)
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return nil, nil
	}
	access := callee.AsPropertyAccessExpression()
	if access == nil {
		return nil, nil
	}
	name := access.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return nil, nil
	}
	return name, access.Expression
}
