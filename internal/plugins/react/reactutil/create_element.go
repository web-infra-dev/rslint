package reactutil

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
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

func isCreateElementCallCore(callee *ast.Node, pragma string, tc *checker.Checker) bool {
	return isPragmaFactoryCallCore(callee, pragma, tc, createElementOnly)
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
	return isPragmaFactoryCallCore(callee, pragma, tc, createOrCloneElement)
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

func isPragmaFactoryCallCore(callee *ast.Node, pragma string, tc *checker.Checker, names pragmaFactoryNames) bool {
	if callee == nil {
		return false
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	// Only parentheses are transparent on the callee, matching ESLint's JS-only
	// AST (ESTree flattens parens). TS expression wrappers
	// (`as` / `satisfies` / `<T>x` / `x!`) are NOT peeled: upstream reads
	// `node.callee.object.name`, which is undefined on a `TSAsExpression` /
	// `TSNonNullExpression` receiver, so `(React as any).createElement` is not
	// recognized — and neither is it here.
	// A *parenthesized* optional chain — `(React?.createElement)(...)` — has its
	// chain terminated by the parens: ESTree freezes it into a `ChainExpression`,
	// so upstream's `node.callee.type === 'MemberExpression'` check fails and the
	// call is NOT recognized. tsgo keeps the explicit ParenthesizedExpression
	// wrapper (it's lost only if we flatten it), so detect that shape first. A
	// non-optional `(React.createElement)(...)` stays transparent and IS
	// recognized; a bare `React?.createElement(...)` / `React.createElement?.(...)`
	// has no wrapping paren and is likewise recognized — all matching upstream.
	if callee.Kind == ast.KindParenthesizedExpression && ast.IsOptionalChain(ast.SkipParentheses(callee)) {
		return false
	}
	callee = ast.SkipParentheses(callee)

	// Bare callee: `createElement(arg)` / `cloneElement(arg)` — recognized
	// only when destructured from the pragma module. Mirrors upstream's
	// second branch of `isCreateElement` / `isCreateCloneElement`.
	if callee.Kind == ast.KindIdentifier {
		if !names.matches(callee.AsIdentifier().Text) {
			return false
		}
		return IsDestructuredFromPragmaImport(callee, pragma, tc)
	}

	// Member-access callee: `<pragma>.<name>(arg)`. Optional-chain access
	// (`React?.createElement(...)`) is accepted — upstream's `isCreateElement`
	// / `isCreateCloneElement` see `node.callee` as a (possibly optional)
	// MemberExpression and match it just the same.
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	prop := callee.AsPropertyAccessExpression()
	nameNode := prop.Name()
	if nameNode.Kind != ast.KindIdentifier || !names.matches(nameNode.AsIdentifier().Text) {
		return false
	}
	// Pragma sub-expression: only parens are transparent (see above). A TS
	// wrapper on the receiver leaves a non-Identifier node, so it won't match —
	// exactly like ESLint reading `.object.name` off a wrapped receiver.
	pragmaExpr := ast.SkipParentheses(prop.Expression)
	return pragmaExpr.Kind == ast.KindIdentifier && pragmaExpr.AsIdentifier().Text == pragma
}
