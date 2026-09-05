package reactutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// IsCreateElementCall reports whether the callee is `<pragma>.createElement`
// (or, with the WithChecker variant below, bare `createElement` resolved
// to a pragma-destructured binding).
//
// Pass an empty pragma to default to "React"; pass GetReactPragma(ctx.Settings)
// to honor the user's `settings.react.pragma` configuration.
//
// Parentheses are transparently skipped on both the callee and the pragma
// identifier (e.g. `(React).createElement` / `(React.createElement)(...)`),
// matching ESLint's paren flattening. TS expression wrappers
// (`as` / `satisfies` / `<T>x` / `x!`) are NOT skipped, so
// `(React as any).createElement` is NOT recognized — mirroring upstream, where
// `node.callee.object.name` is undefined on a wrapped receiver.
// Optional-chain pragma access (`React?.createElement(...)`, and the optional
// call `React.createElement?.(...)`) IS recognized, matching upstream: espree
// exposes `node.callee` as a `MemberExpression` (optional flag set). The one
// exception is a *parenthesized* optional chain — `(React?.createElement)(...)`
// — where the parens freeze the chain into a `ChainExpression` callee that
// upstream does not match; this variant rejects it too.
//
// This non-checker variant only recognizes the member-access form. To
// recognize bare `createElement(...)` calls (with the
// `isDestructuredFromPragmaImport` gate), use
// `IsCreateElementCallWithChecker`.
func IsCreateElementCall(callee *ast.Node, pragma string) bool {
	return isCreateElementCallCore(callee, pragma, nil)
}

// IsCreateElementCallWithChecker is the import-aware variant. When `tc`
// is non-nil, additionally recognizes bare `createElement(arg)` calls
// where the bare callee resolves to a pragma-destructured binding
// (`import { createElement } from 'react'` /
// `const { createElement } = React` / `const createElement = React.createElement`
// / `const { createElement } = require('react')`). Mirrors upstream
// `isCreateElement`'s second branch byte-for-byte.
func IsCreateElementCallWithChecker(callee *ast.Node, pragma string, tc *checker.Checker) bool {
	return isCreateElementCallCore(callee, pragma, tc)
}

// IsCreateElementCallWithRefs is the file-scope-aware variant. When refs is
// provided, it is authoritative for bare imported createElement calls so
// checker-backed linting cannot resolve a binding from another global script.
func IsCreateElementCallWithRefs(
	callee *ast.Node,
	pragma string,
	tc *checker.Checker,
	refs ReferenceResolver,
) bool {
	return isPragmaFactoryCallCore(callee, pragma, tc, refs, createElementOnly)
}

func isCreateElementCallCore(callee *ast.Node, pragma string, tc *checker.Checker) bool {
	return isPragmaFactoryCallCore(callee, pragma, tc, nil, createElementOnly)
}

// IsCreateOrCloneElementCall reports whether the callee resolves to
// `<pragma>.createElement` / `<pragma>.cloneElement` (configured pragma)
// or — when `tc` is non-nil — a bare `createElement` / `cloneElement`
// identifier imported / destructured from the pragma module. Mirrors
// upstream `eslint-plugin-react`'s `isCreateCloneElement` predicate used
// by `no-array-index-key`, INCLUDING upstream's acceptance of optional
// chains (`React?.cloneElement(...)`) — upstream listens on
// `'CallExpression, OptionalCallExpression'` and gates on
// `node.type === 'MemberExpression' || node.type === 'OptionalMemberExpression'`.
//
// Parens are skipped on the pragma sub-expression so `(React).cloneElement`
// is recognized (ESTree flattens parens). TS-only expression wrappers
// (`as` / `satisfies` / `<T>x` / `x!`) on the pragma identifier are NOT
// skipped — that would over-match relative to ESLint's JS-only AST and
// is a divergence we deliberately avoid.
func IsCreateOrCloneElementCall(callee *ast.Node, pragma string, tc *checker.Checker) bool {
	return isPragmaFactoryCallCore(callee, pragma, tc, nil, createOrCloneElement)
}

type pragmaFactoryNames int

const (
	createElementOnly pragmaFactoryNames = iota
	createOrCloneElement
)

func (k pragmaFactoryNames) matches(name string) bool {
	switch k {
	case createElementOnly:
		return name == "createElement"
	case createOrCloneElement:
		return name == "createElement" || name == "cloneElement"
	}
	return false
}

func isPragmaFactoryCallCore(
	callee *ast.Node,
	pragma string,
	tc *checker.Checker,
	refs ReferenceResolver,
	names pragmaFactoryNames,
) bool {
	if callee == nil {
		return false
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	callee = utils.ESTreeCallCallee(callee)
	if callee == nil {
		return false
	}

	// Bare callee: `createElement(arg)` / `cloneElement(arg)` — recognized
	// only when destructured from the pragma module. Mirrors upstream's
	// second branch of `isCreateElement` / `isCreateCloneElement`.
	if callee.Kind == ast.KindIdentifier {
		if !names.matches(callee.AsIdentifier().Text) {
			return false
		}
		return IsDestructuredFromPragmaImportWithRefs(callee, pragma, tc, refs)
	}

	// Member-access callee: `<pragma>.<name>(arg)` or
	// `<pragma>[name](arg)`. Optional-chain access
	// (`React?.createElement(...)`) is accepted — upstream's `isCreateElement`
	// / `isCreateCloneElement` see `node.callee` as a (possibly optional)
	// MemberExpression and match it just the same. For computed access,
	// upstream reads `node.callee.property.name`, so an identifier argument is
	// also a match while a string literal argument is not.
	var nameNode, pragmaExpr *ast.Node
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		prop := callee.AsPropertyAccessExpression()
		nameNode = prop.Name()
		// JSDoc casts and parentheses are absent from ESTree, while an authored
		// TypeScript wrapper on the receiver remains visible and does not match.
		pragmaExpr = utils.ESTreeRuntimeExpression(prop.Expression)
	case ast.KindElementAccessExpression:
		access := callee.AsElementAccessExpression()
		// ESTree's MemberExpression property has a `name` only when the
		// computed property is an Identifier. String literals such as
		// React["createElement"] therefore remain unmatched.
		nameNode = utils.ESTreeRuntimeExpression(access.ArgumentExpression)
		pragmaExpr = utils.ESTreeRuntimeExpression(access.Expression)
	default:
		return false
	}
	if nameNode == nil || pragmaExpr == nil {
		return false
	}
	if nameNode.Kind != ast.KindIdentifier || !names.matches(nameNode.AsIdentifier().Text) {
		return false
	}
	// JSDoc casts and parentheses are absent from ESTree, while an authored
	// TypeScript wrapper on the receiver remains visible and does not match.
	return pragmaExpr.Kind == ast.KindIdentifier && pragmaExpr.AsIdentifier().Text == pragma
}
