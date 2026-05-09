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
// Parentheses AND TS expression wrappers (`as` / `satisfies` / `<T>x` / `x!`)
// are transparently skipped on both the callee itself and the pragma
// identifier (e.g. `(React).createElement` / `(React as any).createElement`).
// Optional-chain calls (`React?.createElement(...)`) are NOT recognized
// (upstream's `node.callee.object.name` access fails on the OptionalCall
// shape).
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
	return isPragmaFactoryCallCore(callee, pragma, tc, createElementOnly, false)
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
	return isPragmaFactoryCallCore(callee, pragma, tc, createOrCloneElement, true)
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

func isPragmaFactoryCallCore(callee *ast.Node, pragma string, tc *checker.Checker, names pragmaFactoryNames, allowOptionalChain bool) bool {
	if callee == nil {
		return false
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	// `IsCreateElementCall` (the public-named variant used by other rules)
	// historically peels TS expression wrappers off the callee itself —
	// keep that branch intact for backwards compatibility.
	// `IsCreateOrCloneElementCall`, used by `no-array-index-key`, mirrors
	// ESLint's JS-only AST and only skips parentheses on the callee.
	if names == createElementOnly {
		callee = SkipExpressionWrappers(callee)
	} else {
		callee = ast.SkipParentheses(callee)
	}

	// Bare callee: `createElement(arg)` / `cloneElement(arg)` — recognized
	// only when destructured from the pragma module. Mirrors upstream's
	// second branch of `isCreateElement` / `isCreateCloneElement`.
	if callee.Kind == ast.KindIdentifier {
		if !names.matches(callee.AsIdentifier().Text) {
			return false
		}
		return IsDestructuredFromPragmaImport(callee, pragma, tc)
	}

	// Member-access callee: `<pragma>.<name>(arg)`.
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	if !allowOptionalChain && ast.IsOptionalChain(callee) {
		return false
	}
	prop := callee.AsPropertyAccessExpression()
	nameNode := prop.Name()
	if nameNode.Kind != ast.KindIdentifier || !names.matches(nameNode.AsIdentifier().Text) {
		return false
	}
	// Pragma sub-expression: `IsCreateElementCall` historically peels TS
	// wrappers; `IsCreateOrCloneElementCall` only peels parens to match
	// ESLint's JS-only AST exactly.
	var pragmaExpr *ast.Node
	if names == createElementOnly {
		pragmaExpr = SkipExpressionWrappers(prop.Expression)
	} else {
		pragmaExpr = ast.SkipParentheses(prop.Expression)
	}
	return pragmaExpr.Kind == ast.KindIdentifier && pragmaExpr.AsIdentifier().Text == pragma
}
