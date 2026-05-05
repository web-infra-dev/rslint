package reactutil

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/scanner"
)

// DefaultReactPragma is the fallback object name for createElement calls
// when `settings.react.pragma` is not configured, matching eslint-plugin-react.
const DefaultReactPragma = "React"

// hookNameRegex matches the React hook naming convention `^use[A-Z0-9].*$`.
// Mirrors eslint-plugin-react-hooks' `RE_HOOKS` and the `HOOK_REGEXP` used by
// no-unstable-nested-components. Exposed here because every React-flavored
// rule that needs to recognize hook calls re-derives the same predicate.
var hookNameRegex = regexp.MustCompile(`^use[A-Z0-9].*$`)

// IsHookName reports whether `name` matches the React hook naming convention.
// Returns false for empty input.
func IsHookName(name string) bool {
	if name == "" {
		return false
	}
	return hookNameRegex.MatchString(name)
}

// GlobToRegex converts a minimatch-style glob (only `*` and `?` wildcards
// recognized; every other regex metacharacter is literally escaped) into a
// fully anchored regular expression. Mirrors the subset of minimatch
// behavior eslint-plugin-react relies on for `propNamePattern`-style options
// — sufficient for `render*`, `*Renderer`, etc.
//
// Compilation is cached per-pattern; the returned `*regexp.Regexp` is
// safe to share across goroutines.
func GlobToRegex(pattern string) *regexp.Regexp {
	if v, ok := globToRegexCache.Load(pattern); ok {
		// Cache values are always `*regexp.Regexp` — `Store` is the
		// only writer below and writes only this type. The check is
		// kept for `forcetypeassert` lint hygiene; a failed assertion
		// would indicate a programming error elsewhere in this file.
		if re, ok := v.(*regexp.Regexp); ok {
			return re
		}
	}
	var b strings.Builder
	b.Grow(len(pattern) + 4)
	b.WriteByte('^')
	for _, r := range pattern {
		switch r {
		case '*':
			b.WriteString(".*")
		case '?':
			b.WriteByte('.')
		case '.', '+', '(', ')', '|', '^', '$', '{', '}', '[', ']', '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('$')
	re := regexp.MustCompile(b.String())
	globToRegexCache.Store(pattern, re)
	return re
}

// MatchGlob is the fast path equivalent of `GlobToRegex(pattern).MatchString(text)`,
// returning false for empty `text`.
func MatchGlob(text, pattern string) bool {
	if text == "" {
		return false
	}
	return GlobToRegex(pattern).MatchString(text)
}

var globToRegexCache sync.Map

// SkipExpressionWrappers is a paren-and-TS-type-wrapper-transparent variant
// of `ast.SkipParentheses`. It additionally peels back tsgo's TS-only
// expression wrappers that ESLint's ESTree never produces: `as`-expressions,
// `satisfies`-expressions, `<T>x` type assertions, and `x!` non-null
// assertions. Use it whenever a rule must reach the underlying expression
// regardless of whether the source uses any of those wrappers — e.g. when
// matching a callee identifier, a JSX tag base, or a return-statement
// argument that may sit behind a `(x as Foo)`.
func SkipExpressionWrappers(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	for {
		switch node.Kind {
		case ast.KindParenthesizedExpression:
			node = node.AsParenthesizedExpression().Expression
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindSatisfiesExpression:
			node = node.AsSatisfiesExpression().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		case ast.KindTypeAssertionExpression:
			node = node.AsTypeAssertion().Expression
		default:
			return node
		}
	}
}

// SkipExpressionWrappersUp is the parent-walk equivalent of
// `SkipExpressionWrappers`: starting from `node.Parent`, walks up while the
// current parent is a transparent expression wrapper (`()` / `as` /
// `satisfies` / `<T>x` / `x!`) and returns the first non-wrapper ancestor,
// or nil when no such ancestor exists. Mirrors what ESTree implicitly does
// by flattening these wrappers — three sites in this rule used to inline
// the loop; one helper keeps them in lockstep.
func SkipExpressionWrappersUp(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	parent := node.Parent
	for parent != nil {
		switch parent.Kind {
		case ast.KindParenthesizedExpression,
			ast.KindAsExpression,
			ast.KindSatisfiesExpression,
			ast.KindNonNullExpression,
			ast.KindTypeAssertionExpression:
			parent = parent.Parent
			continue
		}
		break
	}
	return parent
}

// IsFirstLetterCapitalized is the exported alias for the package-private
// helper. Mirrors eslint-plugin-react's `lib/util/isFirstLetterCapitalized.js`
// — strips leading underscores then compares `unicode.ToUpper(r) == r`.
// Non-cased characters (CJK, digits, emoji) count as "capitalized" because
// they have no upper/lower mapping. Use this for any parent-name / binding
// capitalization check that needs to align with upstream's component
// detection semantics.
func IsFirstLetterCapitalized(s string) bool {
	return isFirstLetterCapitalized(s)
}

// IsLowercaseFirstLetter is the companion of IsFirstLetterCapitalized that
// matches upstream's exact lowercase-skip predicate from
// `lib/rules/no-unstable-nested-components.js`:
//
//	parentName[0] === parentName[0].toLowerCase()
//
// Notably this is NOT the negation of IsFirstLetterCapitalized: this
// helper does NOT strip leading underscores, so `_Foo` is treated as
// lowercase here (the `_` round-trips through `ToLower`) even though
// IsFirstLetterCapitalized returns true for `_Foo` (after stripping `_`,
// `Foo` is capitalized). Both helpers exist because upstream uses each
// in different code paths — keep them paired.
func IsLowercaseFirstLetter(s string) bool {
	if s == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return false
	}
	return unicode.ToLower(r) == r
}

// ParamListOpenParenPos returns the source position of the `(` that opens
// `node`'s parameter list, or -1 when the position cannot be located.
// Walks tokens after `node.Name().End()` via the scanner — robust against
// type-parameter lists (`<T>(p: T)`) where the `(` is not contiguous with
// the name. Use this when narrowing a diagnostic range on an
// object-literal shorthand method / getter / setter so the report site
// aligns with ESTree's `Property { value: FunctionExpression }` shape
// (FE.loc.start at `(`).
//
// `node` must be a MethodDeclaration / GetAccessor / SetAccessor (or
// anything with a non-nil `Name()`); other inputs return -1.
func ParamListOpenParenPos(sf *ast.SourceFile, node *ast.Node) int {
	if sf == nil || node == nil {
		return -1
	}
	name := node.Name()
	if name == nil {
		return -1
	}
	sc := scanner.GetScannerForSourceFile(sf, name.End())
	for {
		tok := sc.Token()
		if tok == ast.KindEndOfFile {
			return -1
		}
		if tok == ast.KindOpenParenToken {
			return sc.TokenStart()
		}
		sc.Scan()
	}
}

// IsObjectLiteralShorthandFunction reports whether `node` is a
// FunctionLike that, in ESTree, would be the inner FunctionExpression of
// a `Property { value: FunctionExpression }` — i.e. an object-literal
// shorthand method / getter / setter. Useful for callers that want to
// narrow diagnostic ranges to the parameter-list `(` (see
// ParamListOpenParenPos) so positions align with ESTree's reporting shape.
func IsObjectLiteralShorthandFunction(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	switch node.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return node.Parent.Kind == ast.KindObjectLiteralExpression
	}
	return false
}

// IsDestructuredFromPragmaImport mirrors upstream eslint-plugin-react's
// `lib/util/isDestructuredFromPragmaImport.js`: reports whether the
// Identifier `ident` (a bare callee like `memo`) was bound from the
// pragma module. Returns true when ident's local binding originated from
// any of:
//
//   - `import { memo } from 'react'` (named import)
//   - `import { memo as m } from 'react'` (named-import rename — checks
//     the imported name, not the local alias)
//   - `import * as React from 'react'`'s namespace + `const memo = React.memo`
//   - `const { memo } = React` (object destructure of the pragma binding)
//   - `const memo = React.memo` (member access via pragma binding)
//   - `const { memo } = require('react')` (require destructure)
//   - `const memo = require('react').memo` (require member access)
//
// `pragma` is the React pragma name (e.g. "React") — the comparison
// against ImportDeclaration / require argument uses
// `strings.ToLower(pragma)` to match upstream's
// `pragma.toLocaleLowerCase()` semantic. `tc` may be nil — when no
// TypeChecker is available the function falls back to a syntax-only
// SourceFile-wide scan via `findPragmaBindingByName`. That fallback is
// strictly a subset of TC-based resolution (no scope precision) but
// covers the idiomatic top-level pragma-import patterns, keeping the
// observable wrapper-recognition behavior aligned with upstream in
// no-tsconfig modes.
func IsDestructuredFromPragmaImport(ident *ast.Node, pragma string, tc *checker.Checker) bool {
	if ident == nil || ident.Kind != ast.KindIdentifier {
		return false
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	pragmaLower := strings.ToLower(pragma)

	if tc == nil {
		// Syntax-only fallback: walk up to the SourceFile and scan it
		// for any binding that introduces `ident.Text` from the pragma
		// module. This is strictly less precise than TC-based
		// resolution (no scope handling, no shadowing detection) but
		// catches the canonical top-level patterns that account for
		// virtually all real-world React pragma imports.
		return findPragmaBindingByName(getSourceFileNode(ident), ident.AsIdentifier().Text, pragma, pragmaLower)
	}

	symbol := tc.GetSymbolAtLocation(ident)
	if symbol == nil {
		return false
	}

	// Pick the most relevant declaration. Upstream walks `latestDef` —
	// for value bindings ValueDeclaration is the right one; for
	// ImportSpecifier (which has no Initializer of its own), upstream
	// walks `latestDef.parent.type === 'ImportDeclaration'`. We mirror
	// by trying ValueDeclaration first then Declarations[0].
	var decl *ast.Node
	if symbol.ValueDeclaration != nil {
		decl = symbol.ValueDeclaration
	} else if len(symbol.Declarations) > 0 {
		decl = symbol.Declarations[0]
	}
	if decl == nil {
		return false
	}

	// 1) Named import: `import { memo } from 'react'` — declaration is
	//    an ImportSpecifier (or ImportClause for default imports, but
	//    bare callee `memo` won't bind to a default).
	if decl.Kind == ast.KindImportSpecifier {
		// Walk up: ImportSpecifier → NamedImports → ImportClause →
		// ImportDeclaration.
		for p := decl.Parent; p != nil; p = p.Parent {
			if p.Kind == ast.KindImportDeclaration {
				ms := p.AsImportDeclaration().ModuleSpecifier
				if ms != nil && ms.Kind == ast.KindStringLiteral &&
					ms.Text() == pragmaLower {
					return true
				}
				return false
			}
		}
		return false
	}

	// 2) BindingElement (object/array destructure): `const { memo } = React`
	//    → declaration is BindingElement; walk up to VariableDeclaration
	//    and inspect its Initializer.
	if decl.Kind == ast.KindBindingElement {
		varDecl := findEnclosingVariableDeclaration(decl)
		if varDecl == nil {
			return false
		}
		init := varDecl.AsVariableDeclaration().Initializer
		return initializerMatchesPragma(init, pragma, pragmaLower)
	}

	// 3) VariableDeclaration: `const memo = React.memo` /
	//    `const memo = require('react').memo`
	if decl.Kind == ast.KindVariableDeclaration {
		init := decl.AsVariableDeclaration().Initializer
		return initializerMatchesPragma(init, pragma, pragmaLower)
	}

	return false
}

// getSourceFileNode walks up from `node` to its enclosing SourceFile,
// returning it as an `*ast.Node`, or nil when no SourceFile ancestor is
// found (extremely unlikely outside of synthesized nodes).
func getSourceFileNode(node *ast.Node) *ast.Node {
	sf := ast.GetSourceFileOfNode(node)
	if sf == nil {
		return nil
	}
	return sf.AsNode()
}

// findPragmaBindingByName is the syntax-only fallback for
// `IsDestructuredFromPragmaImport` when no TypeChecker is available. It
// scans the SourceFile rooted at `root` for any declaration that
// introduces a binding named `name` whose source is the pragma module:
//
//   - `import { name } from '<pragma>'`
//   - `import { x as name } from '<pragma>'` (renamed import — local
//     binding is `name`)
//   - `const { name } = <pragma>` / `const { name } = require('<pragma>')`
//   - `const name = <pragma>.name` / `const name = require('<pragma>').name`
//   - `const { x: name } = <pragma>` / require — destructure-with-rename
//
// Walks the entire SourceFile rather than tracking lexical scope. This
// is a deliberate trade-off: shadowing in inner scopes (e.g. a deeply
// nested `function memo() {}` overriding a top-level
// `import { memo } from 'react'`) is NOT modeled — but the recognition
// only matters for bare callees that already passed name + non-shadow
// checks at the call site, which makes shadowing edge-cases vanish in
// practice.
func findPragmaBindingByName(root *ast.Node, name string, pragma string, pragmaLower string) bool {
	if root == nil || name == "" {
		return false
	}
	var found bool
	var visit func(n *ast.Node)
	visit = func(n *ast.Node) {
		if found || n == nil {
			return
		}
		switch n.Kind {
		case ast.KindImportDeclaration:
			if importDeclBindsNameFromPragma(n, name, pragmaLower) {
				found = true
				return
			}
		case ast.KindVariableDeclaration:
			if variableDeclBindsNameFromPragma(n, name, pragma, pragmaLower) {
				found = true
				return
			}
		}
		n.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found
		})
	}
	visit(root)
	return found
}

// importDeclBindsNameFromPragma reports whether `decl`
// (an ImportDeclaration) introduces a local binding called `name` from
// the module whose lowercased specifier equals `pragmaLower`. Handles
// both plain (`import { name } from '...'`) and renamed
// (`import { x as name } from '...'`) named imports — the local binding
// is the second identifier, which is what we match against `name`.
func importDeclBindsNameFromPragma(decl *ast.Node, name string, pragmaLower string) bool {
	id := decl.AsImportDeclaration()
	if id.ModuleSpecifier == nil || id.ModuleSpecifier.Kind != ast.KindStringLiteral {
		return false
	}
	if id.ModuleSpecifier.Text() != pragmaLower {
		return false
	}
	if id.ImportClause == nil {
		return false
	}
	ic := id.ImportClause.AsImportClause()
	if ic.NamedBindings == nil || ic.NamedBindings.Kind != ast.KindNamedImports {
		// Default import / namespace import don't bind `name` directly.
		return false
	}
	ni := ic.NamedBindings.AsNamedImports()
	if ni.Elements == nil {
		return false
	}
	for _, spec := range ni.Elements.Nodes {
		// ImportSpecifier.Name() returns the local binding identifier
		// (post-rename in `{ x as y }`). That's what shadows scope and
		// what we should compare against `name`.
		local := spec.Name()
		if local != nil && local.Kind == ast.KindIdentifier && local.AsIdentifier().Text == name {
			return true
		}
	}
	return false
}

// variableDeclBindsNameFromPragma reports whether `decl`
// (a VariableDeclaration) introduces a local binding called `name`
// whose value originates from the pragma module. Recognized shapes:
//
//   - `const name = <pragma>.name` / `const name = require('<pragma>').name`
//   - `const { name } = <pragma>` / `const { name } = require('<pragma>')`
//   - `const { x: name } = <pragma>` / `const { x: name } = require('<pragma>')`
func variableDeclBindsNameFromPragma(decl *ast.Node, name, pragma, pragmaLower string) bool {
	vd := decl.AsVariableDeclaration()
	if vd.Initializer == nil {
		return false
	}
	bindingName := vd.Name()
	if bindingName == nil {
		return false
	}
	switch bindingName.Kind {
	case ast.KindIdentifier:
		// `const name = ...` — local binding is `bindingName.Text`.
		if bindingName.AsIdentifier().Text != name {
			return false
		}
		// Initializer must be `<pragma>.name` or `require('<pragma>').name`.
		return initializerIsPragmaMember(vd.Initializer, name, pragma, pragmaLower)
	case ast.KindObjectBindingPattern:
		// `const { name } = ...` or `const { x: name } = ...`. Element
		// match: an ObjectBindingPattern element introduces `name` if
		// either its propertyName is unset and its bindingName.Text is
		// `name`, OR its bindingName.Text is `name` (the alias side).
		if !objectBindingPatternBindsName(bindingName, name) {
			return false
		}
		return initializerMatchesPragma(vd.Initializer, pragma, pragmaLower)
	}
	return false
}

// objectBindingPatternBindsName reports whether any element of the
// ObjectBindingPattern introduces a local binding called `name`. The
// local binding is the BindingElement.Name() — for `{ x: name }`,
// PropertyName is `x` and Name is `name`; we always compare against
// Name. Nested patterns are not recursed into (they don't apply to
// pragma-import shapes).
func objectBindingPatternBindsName(pat *ast.Node, name string) bool {
	obp := pat.AsBindingPattern()
	if obp == nil || obp.Elements == nil {
		return false
	}
	for _, el := range obp.Elements.Nodes {
		be := el.AsBindingElement()
		local := be.Name()
		if local != nil && local.Kind == ast.KindIdentifier && local.AsIdentifier().Text == name {
			return true
		}
	}
	return false
}

// initializerIsPragmaMember reports whether `init` is `<pragma>.<name>` or
// `require('<pragma>').<name>` — the two member-access shapes that
// introduce a `name` binding pulled from the pragma module without
// going through a destructure pattern.
func initializerIsPragmaMember(init *ast.Node, name, pragma, pragmaLower string) bool {
	init = SkipExpressionWrappers(init)
	if init == nil || init.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	pa := init.AsPropertyAccessExpression()
	prop := pa.Name()
	if prop == nil || prop.Kind != ast.KindIdentifier || prop.AsIdentifier().Text != name {
		return false
	}
	obj := SkipExpressionWrappers(pa.Expression)
	if obj == nil {
		return false
	}
	if obj.Kind == ast.KindIdentifier && obj.AsIdentifier().Text == pragma {
		return true
	}
	if obj.Kind == ast.KindCallExpression && isRequireCallOfPragma(obj, pragmaLower) {
		return true
	}
	return false
}

// findEnclosingVariableDeclaration walks up from a BindingElement to its
// enclosing VariableDeclaration, or nil when not found (e.g. parameter
// bindings, which are not pragma imports).
func findEnclosingVariableDeclaration(node *ast.Node) *ast.Node {
	for p := node.Parent; p != nil; p = p.Parent {
		switch p.Kind {
		case ast.KindVariableDeclaration:
			return p
		case ast.KindParameter, ast.KindFunctionDeclaration,
			ast.KindArrowFunction, ast.KindFunctionExpression,
			ast.KindMethodDeclaration:
			return nil
		}
	}
	return nil
}

// initializerMatchesPragma reports whether the given initializer
// expression evaluates to the pragma binding (or to a property of it).
// Mirrors the four init shapes upstream's helper inspects.
func initializerMatchesPragma(init *ast.Node, pragma, pragmaLower string) bool {
	if init == nil {
		return false
	}
	init = SkipExpressionWrappers(init)

	// `init` is the pragma identifier itself (`= React`).
	if init.Kind == ast.KindIdentifier && init.AsIdentifier().Text == pragma {
		return true
	}

	// `init` is `pragma.something` — `= React.memo`.
	if init.Kind == ast.KindPropertyAccessExpression {
		obj := SkipExpressionWrappers(init.AsPropertyAccessExpression().Expression)
		if obj.Kind == ast.KindIdentifier && obj.AsIdentifier().Text == pragma {
			return true
		}
		// `init` is `require('react').memo` — member access on a
		// require call.
		if obj.Kind == ast.KindCallExpression && isRequireCallOfPragma(obj, pragmaLower) {
			return true
		}
	}

	// `init` is `require('react')` directly (destructure case).
	if init.Kind == ast.KindCallExpression && isRequireCallOfPragma(init, pragmaLower) {
		return true
	}

	return false
}

// isRequireCallOfPragma reports whether `call` is `require('<pragmaLower>')`.
// Upstream's helper checks `callee.name === 'require'` and
// `arguments[0].value === pragma.toLocaleLowerCase()`.
func isRequireCallOfPragma(call *ast.Node, pragmaLower string) bool {
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	c := call.AsCallExpression()
	callee := SkipExpressionWrappers(c.Expression)
	if callee == nil || callee.Kind != ast.KindIdentifier ||
		callee.AsIdentifier().Text != "require" {
		return false
	}
	if c.Arguments == nil || len(c.Arguments.Nodes) == 0 {
		return false
	}
	arg := SkipExpressionWrappers(c.Arguments.Nodes[0])
	if arg == nil || arg.Kind != ast.KindStringLiteral {
		return false
	}
	return arg.AsStringLiteral().Text == pragmaLower
}

// DefaultReactCreateClass is the fallback ES5 factory name when
// `settings.react.createClass` is not configured, matching
// eslint-plugin-react.
const DefaultReactCreateClass = "createReactClass"

// ComponentWrapperEntry describes one user-configured component-wrapping
// call site. Either form is recognized:
//
//   - `{property: "memo", object: "React"}` matches `<object>.<property>(fn)`
//     calls. Empty `object` is treated as `DefaultReactPragma`.
//   - `{property: "memo"}` matches bare `<property>(fn)` calls when `object`
//     is empty.
//
// Mirrors eslint-plugin-react's `settings.componentWrapperFunctions` —
// strings in the user setting expand to `{property: <s>}`, objects pass
// through.
//
// `IsUserConfigured` distinguishes entries the user explicitly added via
// `settings.componentWrapperFunctions` from entries we inject as
// hardcoded defaults (memo / forwardRef, pragma-qualified or bare).
// Upstream's `isDestructuredFromPragmaImport` adds a runtime guard to
// bare default entries — they only match when the bare callee was
// destructure-imported from the pragma module. We have no import
// resolver, so we approximate by matching default bare entries on
// non-optional calls only, and matching user-configured bare entries
// freely (since they don't depend on import resolution).
type ComponentWrapperEntry struct {
	Object           string
	Property         string
	IsUserConfigured bool
}

// DefaultComponentWrappers is the always-on wrapper list every React rule
// uses regardless of `settings.componentWrapperFunctions`. Mirrors upstream:
// `{property: 'memo', object: pragma}`, `{property: 'forwardRef', object: pragma}`,
// plus the bare `memo` / `forwardRef` aliases.
func DefaultComponentWrappers(pragma string) []ComponentWrapperEntry {
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	return []ComponentWrapperEntry{
		{Object: pragma, Property: "memo"},
		{Object: pragma, Property: "forwardRef"},
		{Property: "memo"},
		{Property: "forwardRef"},
	}
}

// GetComponentWrapperFunctions reads `settings.componentWrapperFunctions`
// and merges the user's additions on top of `DefaultComponentWrappers`.
// Accepted shapes per entry:
//
//   - string: "myMemo" → {Property: "myMemo"}
//   - object: {"property": "memo", "object": "React"} →
//     {Object: "React", Property: "memo"}; "object" defaults to empty
//     (bare call) when omitted
//
// Unknown / malformed entries are silently ignored, matching upstream.
func GetComponentWrapperFunctions(settings map[string]interface{}, pragma string) []ComponentWrapperEntry {
	out := DefaultComponentWrappers(pragma)
	if settings == nil {
		return out
	}
	raw, ok := settings["componentWrapperFunctions"]
	if !ok {
		return out
	}
	add := func(v interface{}) {
		switch e := v.(type) {
		case string:
			if e != "" {
				out = append(out, ComponentWrapperEntry{Property: e, IsUserConfigured: true})
			}
		case map[string]interface{}:
			prop, _ := e["property"].(string)
			if prop == "" {
				return
			}
			obj, _ := e["object"].(string)
			out = append(out, ComponentWrapperEntry{Object: obj, Property: prop, IsUserConfigured: true})
		}
	}
	switch r := raw.(type) {
	case []interface{}:
		for _, v := range r {
			add(v)
		}
	default:
		add(r)
	}
	return out
}

// MatchesAnyComponentWrapper reports whether `call` matches any entry in
// `wrappers`, with `fn` as its first argument (paren / TS-wrapper transparent).
// Pass an empty pragma to default to "React"; the pragma is only consulted
// for entries whose `Object` is empty AND need to fall back to the configured
// pragma — but `DefaultComponentWrappers` already pre-fills the pragma form,
// so callers normally shouldn't need to thread pragma through twice.
//
// Optional-chain handling mirrors upstream's `isPragmaComponentWrapper`:
//
//   - Member-level optional (`React?.memo(arg)`) — recognized; Babel
//     emits the callee as MemberExpression with `optional: true` and
//     upstream's `callee.type === 'MemberExpression'` check still passes.
//
//   - Call-level optional (`memo?.(arg)`) on a bare Identifier callee —
//     recognized only against `IsUserConfigured: true` entries.
//     Hardcoded bare-default entries (`memo` / `forwardRef` without
//     pragma object) are upstream-gated by
//     `isDestructuredFromPragmaImport`, which we cannot enforce without
//     an import resolver. Restricting hardcoded bare defaults to
//     non-optional matches keeps us conservative; user wrappers don't
//     need that gate (they're explicit user opt-in).
func MatchesAnyComponentWrapper(call, fn *ast.Node, wrappers []ComponentWrapperEntry) bool {
	return matchesAnyComponentWrapperCore(call, fn, wrappers, "", nil)
}

// MatchesAnyComponentWrapperWithChecker is the import-aware variant.
// When `tc` is non-nil and the callee is a bare Identifier matching a
// hardcoded bare default entry (`{Property: "memo"}` /
// `{Property: "forwardRef"}` from `DefaultComponentWrappers`), the
// callee's binding must additionally be destructured from / imported
// from / required from the pragma module (per
// `IsDestructuredFromPragmaImport`). This precisely mirrors upstream's
//
//	(!wrapperFunction.object ||
//	 (wrapperFunction.object === pragma &&
//	  this.isDestructuredFromPragmaImport(node, node.callee.name)))
//
// gate. Without this, `memo(arrow)` would silently classify when `memo`
// is a user-defined function unrelated to React — leading to over-reports
// where upstream skips. Use this variant whenever a TypeChecker is
// available; otherwise the import-resolution check is skipped (matching
// the non-checker variant's conservative behavior — call-level optional
// rejected for hardcoded bare defaults).
func MatchesAnyComponentWrapperWithChecker(call, fn *ast.Node, wrappers []ComponentWrapperEntry, pragma string, tc *checker.Checker) bool {
	return matchesAnyComponentWrapperCore(call, fn, wrappers, pragma, tc)
}

func matchesAnyComponentWrapperCore(call, fn *ast.Node, wrappers []ComponentWrapperEntry, pragma string, tc *checker.Checker) bool {
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	c := call.AsCallExpression()
	if c.Arguments == nil || len(c.Arguments.Nodes) == 0 {
		return false
	}
	if SkipExpressionWrappers(c.Arguments.Nodes[0]) != fn {
		return false
	}
	callLevelOptional := c.QuestionDotToken != nil
	callee := SkipExpressionWrappers(c.Expression)
	for _, w := range wrappers {
		if w.Property == "" {
			continue
		}
		switch callee.Kind {
		case ast.KindIdentifier:
			if w.Object != "" {
				continue
			}
			if callee.AsIdentifier().Text != w.Property {
				continue
			}
			if w.IsUserConfigured {
				// User-configured bare entry: accept any callee shape
				// (call-level optional included). User entries don't
				// need the pragma-import gate since they're explicit
				// opt-in.
				return true
			}
			// Hardcoded bare default (memo / forwardRef without object):
			// upstream gates with `isDestructuredFromPragmaImport`. We
			// always run that gate — when a TypeChecker is available
			// it resolves the binding precisely, and when not it falls
			// back to a syntax-only SourceFile scan that handles the
			// canonical top-level pragma-import shapes.
			if !IsDestructuredFromPragmaImport(callee, pragma, tc) {
				continue
			}
			return true
		case ast.KindPropertyAccessExpression:
			if w.Object == "" {
				continue
			}
			// Call-level optional on a member callee (`(R.memo)?.()`)
			// is structurally distinct from member-level optional
			// (`R?.memo()` — flag on the PropertyAccess) and upstream
			// also rejects it (`callee.name` undefined on the call's
			// own optional shape). Reject regardless of IsUserConfigured.
			if callLevelOptional {
				continue
			}
			pa := callee.AsPropertyAccessExpression()
			obj := SkipExpressionWrappers(pa.Expression)
			if obj.Kind != ast.KindIdentifier || obj.AsIdentifier().Text != w.Object {
				continue
			}
			name := pa.Name()
			if name == nil || name.Kind != ast.KindIdentifier {
				continue
			}
			if name.AsIdentifier().Text == w.Property {
				return true
			}
		}
	}
	return false
}

// GetReactPragma reads `settings.react.pragma` from the config settings map.
// Returns DefaultReactPragma when the setting is absent, not a string, or empty.
func GetReactPragma(settings map[string]interface{}) string {
	if settings == nil {
		return DefaultReactPragma
	}
	reactSettings, ok := settings["react"].(map[string]interface{})
	if !ok {
		return DefaultReactPragma
	}
	pragma, ok := reactSettings["pragma"].(string)
	if !ok || pragma == "" {
		return DefaultReactPragma
	}
	return pragma
}

// DefaultReactFragment is the fallback fragment name for JSX shorthand
// fragment diagnostics when `settings.react.fragment` is not configured,
// matching eslint-plugin-react.
const DefaultReactFragment = "Fragment"

// GetReactFragmentPragma reads `settings.react.fragment` from the config
// settings map. Returns DefaultReactFragment when the setting is absent,
// not a string, or empty.
func GetReactFragmentPragma(settings map[string]interface{}) string {
	if settings == nil {
		return DefaultReactFragment
	}
	reactSettings, ok := settings["react"].(map[string]interface{})
	if !ok {
		return DefaultReactFragment
	}
	v, ok := reactSettings["fragment"].(string)
	if !ok || v == "" {
		return DefaultReactFragment
	}
	return v
}

// GetReactCreateClass reads `settings.react.createClass` from the config
// settings map. Returns DefaultReactCreateClass when the setting is absent,
// not a string, or empty.
func GetReactCreateClass(settings map[string]interface{}) string {
	if settings == nil {
		return DefaultReactCreateClass
	}
	reactSettings, ok := settings["react"].(map[string]interface{})
	if !ok {
		return DefaultReactCreateClass
	}
	v, ok := reactSettings["createClass"].(string)
	if !ok || v == "" {
		return DefaultReactCreateClass
	}
	return v
}

// reactVersionRe captures the leading major[.minor[.patch]] numeric triple of
// a semver-ish string. Prerelease / build metadata / range qualifiers are
// ignored — matching eslint-plugin-react's `semver.coerce`-like behavior for
// the simple comparisons this package performs.
var reactVersionRe = regexp.MustCompile(`(\d+)(?:\.(\d+))?(?:\.(\d+))?`)

// ParseReactVersion returns the (major, minor, patch) triple of
// `settings.react.version`. When the setting is missing, not a string, empty,
// or not recognizable as a version, it defaults to (999, 999, 999) — matching
// eslint-plugin-react's `getReactVersionFromContext`, which treats an absent
// version as "latest".
func ParseReactVersion(settings map[string]interface{}) (int, int, int) {
	if settings == nil {
		return 999, 999, 999
	}
	reactSettings, ok := settings["react"].(map[string]interface{})
	if !ok {
		return 999, 999, 999
	}
	raw, _ := reactSettings["version"].(string)
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 999, 999, 999
	}
	m := reactVersionRe.FindStringSubmatch(raw)
	if m == nil {
		return 999, 999, 999
	}
	toInt := func(s string) int {
		if s == "" {
			return 0
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return 0
		}
		return n
	}
	return toInt(m[1]), toInt(m[2]), toInt(m[3])
}

// ReactVersionLessThan reports whether `settings.react.version` is strictly
// less than the given major.minor.patch. See ParseReactVersion for the default
// when the setting is missing.
func ReactVersionLessThan(settings map[string]interface{}, major, minor, patch int) bool {
	a, b, c := ParseReactVersion(settings)
	if a != major {
		return a < major
	}
	if b != minor {
		return b < minor
	}
	return c < patch
}

// IsCreateClassCall reports whether the given CallExpression's callee is
// `<createClass>(...)` or `<pragma>.<createClass>(...)`. Parentheses are
// skipped on both the callee and the pragma identifier. Pass the empty string
// for pragma/createClass to fall back to `DefaultReactPragma` /
// `DefaultReactCreateClass`.
func IsCreateClassCall(call *ast.CallExpression, pragma, createClass string) bool {
	if call == nil {
		return false
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	if createClass == "" {
		createClass = DefaultReactCreateClass
	}
	callee := ast.SkipParentheses(call.Expression)
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.AsIdentifier().Text == createClass
	case ast.KindPropertyAccessExpression:
		pa := callee.AsPropertyAccessExpression()
		obj := ast.SkipParentheses(pa.Expression)
		if obj.Kind != ast.KindIdentifier || obj.AsIdentifier().Text != pragma {
			return false
		}
		name := pa.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return false
		}
		return name.AsIdentifier().Text == createClass
	}
	return false
}

// ExtendsReactComponent reports whether `classNode` (a ClassDeclaration or
// ClassExpression) has an `extends` clause referencing `Component` or
// `PureComponent` — either as a bare identifier or qualified by the
// configured pragma (e.g. `React.Component`). Parentheses are skipped. Pass
// the empty string for pragma to default to `DefaultReactPragma`.
//
// NOTE: Matches the name regex used by eslint-plugin-react's
// `componentUtil.isES6Component` (`/^(Pure)?Component$/`). Aliased imports
// (e.g. `import { Component as C }`) are not resolved — same as the upstream
// rule.
func ExtendsReactComponent(classNode *ast.Node, pragma string) bool {
	if classNode == nil {
		return false
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	heritage := ast.GetClassExtendsHeritageElement(classNode)
	if heritage == nil {
		return false
	}
	hc := heritage.AsExpressionWithTypeArguments()
	if hc == nil || hc.Expression == nil {
		return false
	}
	expr := ast.SkipParentheses(hc.Expression)
	switch expr.Kind {
	case ast.KindIdentifier:
		return isComponentName(expr.AsIdentifier().Text)
	case ast.KindPropertyAccessExpression:
		pa := expr.AsPropertyAccessExpression()
		obj := ast.SkipParentheses(pa.Expression)
		if obj.Kind != ast.KindIdentifier || obj.AsIdentifier().Text != pragma {
			return false
		}
		nameNode := pa.Name()
		if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			return false
		}
		return isComponentName(nameNode.AsIdentifier().Text)
	}
	return false
}

func isComponentName(name string) bool {
	return name == "Component" || name == "PureComponent"
}

// ExtendsReactPureComponent reports whether `classNode` (a ClassDeclaration
// or ClassExpression) has an `extends` clause referencing `PureComponent` —
// either as a bare identifier or qualified by the configured pragma (e.g.
// `React.PureComponent`). Parentheses are skipped. Pass the empty string for
// pragma to default to `DefaultReactPragma`.
//
// Mirrors eslint-plugin-react's `componentUtil.isPureComponent`, which uses
// the regex `/^(<pragma>\.)?PureComponent$/` over the rendered extends-clause
// text. Plain `Component` does NOT match (use ExtendsReactComponent for the
// broader detection).
func ExtendsReactPureComponent(classNode *ast.Node, pragma string) bool {
	if classNode == nil {
		return false
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	heritage := ast.GetClassExtendsHeritageElement(classNode)
	if heritage == nil {
		return false
	}
	hc := heritage.AsExpressionWithTypeArguments()
	if hc == nil || hc.Expression == nil {
		return false
	}
	expr := ast.SkipParentheses(hc.Expression)
	switch expr.Kind {
	case ast.KindIdentifier:
		return expr.AsIdentifier().Text == "PureComponent"
	case ast.KindPropertyAccessExpression:
		pa := expr.AsPropertyAccessExpression()
		obj := ast.SkipParentheses(pa.Expression)
		if obj.Kind != ast.KindIdentifier || obj.AsIdentifier().Text != pragma {
			return false
		}
		name := pa.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return false
		}
		return name.AsIdentifier().Text == "PureComponent"
	}
	return false
}

// GetJsxTagBaseIdentifier returns the leftmost Identifier of a JSX tag-name
// node — i.e. the symbol a rule must resolve to classify the tag. Pass the
// tag-name node obtained from `GetJsxTagName` (or directly from
// `JsxOpeningElement.TagName` / `JsxSelfClosingElement.TagName`). Returns nil
// when the tag does not terminate in an Identifier (ThisKeyword base,
// JsxNamespacedName, unknown shape).
//
// Shapes handled:
//
//   - `<Foo />`                 → Identifier("Foo")
//   - `<Foo.Bar />`             → Identifier("Foo")
//   - `<Foo.Bar.Baz />`         → Identifier("Foo")
//   - `<this />` / `<this.X />` → nil (ThisKeyword base)
//   - `<a:b />`                 → nil (JsxNamespacedName — not an identifier
//     reference in any scope)
//   - `<foo-bar />`             → Identifier("foo-bar") (tsgo preserves the
//     hyphenated text verbatim; callers decide whether that's DOM).
func GetJsxTagBaseIdentifier(tagName *ast.Node) *ast.Node {
	if tagName == nil {
		return nil
	}
	switch tagName.Kind {
	case ast.KindIdentifier:
		return tagName
	case ast.KindPropertyAccessExpression:
		base := tagName
		for base.Kind == ast.KindPropertyAccessExpression {
			base = base.AsPropertyAccessExpression().Expression
		}
		if base.Kind == ast.KindIdentifier {
			return base
		}
	}
	return nil
}

// IsInsideReactComponent reports whether `node` is lexically contained within
// a React component — either an ES5 component (object literal passed as an
// argument to `<createClass>(...)` / `<pragma>.<createClass>(...)`) or an
// ES6 component (ClassDeclaration / ClassExpression extending Component or
// PureComponent, optionally qualified by pragma).
//
// Pass the empty string for pragma/createClass to fall back to
// `DefaultReactPragma` / `DefaultReactCreateClass`.
//
// Mirrors eslint-plugin-react's componentUtil:
//
//   - ES6 path: only the nearest enclosing class decides component status
//     (matching `getParentES6Component`'s `while scope.type !== 'class'`).
//     A non-component class does not "pass through" to let an outer component
//     class match — this prevents false positives like a non-React inner
//     class nested inside a React class.
//
//   - ES5 path: `this` / `this.refs` must occur inside some function, whose
//     ObjectExpression parent is the argument to `<createClass>(...)`. We
//     approximate this by requiring that an enclosing function has been
//     passed on the walk up before an ObjectExpression is accepted — which
//     rules out pathological cases like `createReactClass({ x: this.refs.y })`
//     where `this` is not inside any function (ESLint's scope walk returns
//     null for that too).
func IsInsideReactComponent(node *ast.Node, pragma, createClass string) bool {
	return GetEnclosingReactComponent(node, pragma, createClass) != nil
}

// GetEnclosingReactComponent is IsInsideReactComponent's sibling that returns
// the component node itself (the ClassDeclaration / ClassExpression, or the
// ObjectLiteralExpression passed to createReactClass) rather than a bool.
// Returns nil when `node` is not inside a React component. See
// IsInsideReactComponent for the detection rules.
func GetEnclosingReactComponent(node *ast.Node, pragma, createClass string) *ast.Node {
	if node == nil {
		return nil
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	if createClass == "" {
		createClass = DefaultReactCreateClass
	}
	seenNearestClass := false
	seenEnclosingFunction := false
	for p := node.Parent; p != nil; p = p.Parent {
		if ast.IsFunctionLike(p) {
			seenEnclosingFunction = true
		}
		switch p.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			if seenNearestClass {
				// The nearest class already decided ES6 classification;
				// outer classes do not get a second chance (matches ESLint's
				// scope-walk that stops at the first class scope).
				continue
			}
			seenNearestClass = true
			if ExtendsReactComponent(p, pragma) {
				return p
			}
		case ast.KindObjectLiteralExpression:
			if !seenEnclosingFunction {
				continue
			}
			// The ObjectLiteralExpression may be wrapped in one or more
			// ParenthesizedExpressions before reaching the CallExpression
			// (ESTree would flatten these; tsgo preserves them), e.g.
			// `createReactClass(({...}))`. Walk up through paren wrappers
			// to find the actual argument position.
			arg := p
			for arg.Parent != nil && arg.Parent.Kind == ast.KindParenthesizedExpression {
				arg = arg.Parent
			}
			parent := arg.Parent
			if parent == nil || parent.Kind != ast.KindCallExpression {
				continue
			}
			call := parent.AsCallExpression()
			if !isObjectArgumentOf(call, arg) {
				continue
			}
			if IsCreateClassCall(call, pragma, createClass) {
				return p
			}
		}
	}
	return nil
}

func isObjectArgumentOf(call *ast.CallExpression, obj *ast.Node) bool {
	if call.Arguments == nil {
		return false
	}
	for _, arg := range call.Arguments.Nodes {
		if arg == obj {
			return true
		}
	}
	return false
}

// IsCreateReactClassObjectArg reports whether `obj` (an ObjectLiteralExpression)
// is the FIRST argument of a `<createClass>(...)` / `<pragma>.<createClass>(...)`
// call. Parens wrapping `obj` before it reaches the call argument position are
// transparent — tsgo preserves them while ESTree flattens — so
// `createReactClass(({...}))` still matches.
//
// Pass the empty string for pragma / createClass to fall back to
// `DefaultReactPragma` / `DefaultReactCreateClass`. Returns false for any
// non-ObjectLiteralExpression input, for objects in non-argument positions,
// and for calls whose callee is not the configured createClass name.
func IsCreateReactClassObjectArg(obj *ast.Node, pragma, createClass string) bool {
	if obj == nil || obj.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	cur := obj
	for cur.Parent != nil && cur.Parent.Kind == ast.KindParenthesizedExpression {
		cur = cur.Parent
	}
	parent := cur.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	call := parent.AsCallExpression()
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 || call.Arguments.Nodes[0] != cur {
		return false
	}
	return IsCreateClassCall(call, pragma, createClass)
}

// GetEnclosingReactComponentOrStateless is GetEnclosingReactComponent extended
// with eslint-plugin-react's `getParentStatelessComponent` fallback: when no
// enclosing ES6 class / ES5 createReactClass component is found, the nearest
// FunctionLike ancestor that looks like a functional component (capital-cased
// name + returns JSX/null) is returned.
//
// Priority matches upstream's `getParentComponent`:
//
//	getParentES6Component || getParentES5Component || getParentStatelessComponent
//
// so when a mutation node is inside an inner function nested within an outer
// class component, the OUTER class component is returned (preventing the
// inner stateless candidate from masking the class boundary).
//
// Only a restricted subset of upstream's heuristics is implemented — the
// patterns covering production React code: named FunctionDeclaration,
// FunctionExpression / ArrowFunction assigned to a capital-cased
// VariableDeclarator, PropertyAssignment, or ExportAssignment (default export),
// plus function expression in a CallExpression (e.g. React.memo wrapper —
// approximate match). This is intentionally conservative: missed detection
// causes a rule miss, over-detection would cause false-positive reports in
// non-component functions.
func GetEnclosingReactComponentOrStateless(node *ast.Node, pragma, createClass string) *ast.Node {
	if comp := GetEnclosingReactComponent(node, pragma, createClass); comp != nil {
		return comp
	}
	for p := node.Parent; p != nil; p = p.Parent {
		if ast.IsFunctionLike(p) && IsStatelessReactComponent(p, pragma) {
			return p
		}
	}
	return nil
}

// IsStatelessReactComponent reports whether `fn` (a FunctionLike) looks like a
// React functional component. Mirrors eslint-plugin-react's
// `getStatelessComponent` decision tree:
//
//   - FunctionDeclaration — component iff returns JSX/null AND either:
//     (a) its own Identifier is capitalized, OR
//     (b) it is anonymous AND carries the `export default` modifier (ESLint's
//     `!node.id || capitalized(node.id.name)` condition).
//
//   - FunctionExpression / ArrowFunction — component iff returns JSX/null AND
//     either wrapped in a pragma component call OR in an "allowed position"
//     AND the position-specific capitalization check passes:
//
//   - Wrapped in `<pragma>.memo(...)` / `<pragma>.forwardRef(...)` / bare
//     `memo(...)` / bare `forwardRef(...)` — always a component.
//
//   - Allowed positions (VariableDeclarator, AssignmentExpression,
//     PropertyAssignment, ReturnStatement, ExportAssignment, outer
//     ArrowFunction body) gate everything else. A bare IIFE or any other
//     CallExpression argument position is NOT allowed, matching upstream's
//     `isInAllowedPositionForComponent` default-false branch.
//
//   - Within an allowed position, specific capitalization rules apply per
//     upstream: VariableDeclarator/PropertyAssignment use the binding name;
//     `Id = fn` assignments use the LHS Identifier; MemberExpression LHS
//     uses the rightmost property name (with `module.exports = ...` as a
//     special blanket-true case); a named FunctionExpression defers to its
//     own Identifier.
//
// Pass the empty string for `pragma` to default to `DefaultReactPragma`.
//
// This wrapper preserves the historical "no checker" call shape used by
// every other React rule. Pass a non-nil checker via
// `IsStatelessReactComponentWithChecker` to enable Identifier-through-scope
// resolution inside the JSX-return checks (relevant for any input where
// the function returns a name bound elsewhere — `return view` ↔
// `let view = <div/>` etc).
func IsStatelessReactComponent(fn *ast.Node, pragma string) bool {
	return isStatelessReactComponentCore(fn, pragma, nil, nil)
}

// IsStatelessReactComponentWithChecker mirrors IsStatelessReactComponent and
// additionally threads `tc` into every isReturningJSX / isReturningJSXOrNull
// gate inside the decision tree. When `tc` is nil, all behavior matches
// `IsStatelessReactComponent` exactly (local-block initializer scan only).
//
// The pragma-component-wrapper branch (Branch 11) uses the hardcoded
// default wrappers (`memo` / `forwardRef`, pragma-qualified or bare). To
// honor `settings.componentWrapperFunctions` here, use
// `IsStatelessReactComponentWithWrappers` instead.
func IsStatelessReactComponentWithChecker(fn *ast.Node, pragma string, tc *checker.Checker) bool {
	return isStatelessReactComponentCore(fn, pragma, tc, nil)
}

// IsStatelessReactComponentWithWrappers is the variant that consults a
// user-provided wrapper list when classifying the inner function of
// pragma-component-wrapper calls (Branch 11 of the decision tree).
//
// Why this matters: `myMemo(() => null)` with
// `settings.componentWrapperFunctions: ['myMemo']` should classify the
// inner arrow as a stateless component (via the wrapper-arm of upstream's
// `getStatelessComponent`), so that the null-only return correctly
// triggers `isStatelessComponentReturningNull` and the rule SKIPs. With
// the hardcoded variant above, `myMemo` isn't recognized → the arrow
// isn't classified → the null-only skip never fires → the rule reports
// where upstream would not.
//
// Pass `wrappers = nil` for hardcoded defaults; pass the configured
// `GetComponentWrapperFunctions(...)` list to honor user settings.
func IsStatelessReactComponentWithWrappers(fn *ast.Node, pragma string, tc *checker.Checker, wrappers []ComponentWrapperEntry) bool {
	return isStatelessReactComponentCore(fn, pragma, tc, wrappers)
}

// isStatelessReactComponentCore is the shared decision tree. `wrappers`
// nil means "use hardcoded memo/forwardRef defaults" (matching the legacy
// public API); non-nil means "consult this list in Branch 11 instead".
func isStatelessReactComponentCore(fn *ast.Node, pragma string, tc *checker.Checker, wrappers []ComponentWrapperEntry) bool {
	if fn == nil {
		return false
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}

	switch fn.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		// Object-literal shorthand method / accessor. Upstream's Property
		// branch (method && !computed) | (!id && !computed) classifies the
		// inner FE as a component when the property key is a capitalized
		// Identifier AND the function returns strict JSX (isReturningJSX).
		// Setters naturally fail functionReturnsJSX (no return value).
		// Class-body occurrences have a ClassLike parent — NOT
		// ObjectLiteralExpression — and are excluded so they continue to go
		// through the ES6-class path.
		parent := fn.Parent
		if parent == nil || parent.Kind != ast.KindObjectLiteralExpression {
			return false
		}
		name := fn.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return false
		}
		return isFirstLetterCapitalized(name.AsIdentifier().Text) && functionReturnsJSXInternal(fn, false, pragma, tc)
	case ast.KindFunctionDeclaration:
		// Branch: FunctionDeclaration requires isReturningJSXOrNull AND
		// (no id || capitalized). Anonymous FD is only legal as
		// `export default function() {...}`.
		if !functionReturnsJSXInternal(fn, true, pragma, tc) {
			return false
		}
		name := fn.Name()
		if name == nil {
			return ast.GetCombinedModifierFlags(fn)&ast.ModifierFlagsDefault != 0
		}
		return name.Kind == ast.KindIdentifier && isFirstLetterCapitalized(name.AsIdentifier().Text)
	case ast.KindFunctionExpression, ast.KindArrowFunction:
	default:
		return false
	}

	parent := fn.Parent
	if parent == nil {
		return false
	}

	// Derived flags mirroring upstream's local `isPropertyAssignment` /
	// `isModuleExportsAssignment`.
	isMEAssign := false
	isModuleExportsAssign := false
	if parent.Kind == ast.KindBinaryExpression {
		bin := parent.AsBinaryExpression()
		if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken && bin.Right == fn {
			left := ast.SkipParentheses(bin.Left)
			if left.Kind == ast.KindPropertyAccessExpression {
				isMEAssign = true
				pa := left.AsPropertyAccessExpression()
				obj := ast.SkipParentheses(pa.Expression)
				name := pa.Name()
				if obj.Kind == ast.KindIdentifier && obj.AsIdentifier().Text == "module" &&
					name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "exports" {
					isModuleExportsAssign = true
				}
			}
		}
	}

	// Branch 1 — ExportDefault (strict isReturningJSX).
	if parent.Kind == ast.KindExportAssignment {
		return functionReturnsJSXInternal(fn, false, pragma, tc)
	}

	// Branch 2 — VariableDeclarator.
	if parent.Kind == ast.KindVariableDeclaration {
		if !functionReturnsJSXInternal(fn, true, pragma, tc) {
			return false
		}
		binding := parent.AsVariableDeclaration().Name()
		if binding != nil && binding.Kind == ast.KindIdentifier {
			return isFirstLetterCapitalized(binding.AsIdentifier().Text)
		}
		return false
	}

	// Branch 3 — early-reject in ReturnStatement / arrow-expression-body
	// when not strictly returning JSX.
	if parent.Kind == ast.KindReturnStatement ||
		(parent.Kind == ast.KindArrowFunction && parent.AsArrowFunction().Body == fn) {
		if !functionReturnsJSXInternal(fn, false, pragma, tc) {
			return false
		}
	}

	// Branch 4 — AssignmentExpression with non-MemberExpression LHS
	// (handled; Identifier LHS path).
	if parent.Kind == ast.KindBinaryExpression && !isMEAssign {
		bin := parent.AsBinaryExpression()
		if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken && bin.Right == fn {
			if !functionReturnsJSXInternal(fn, true, pragma, tc) {
				return false
			}
			// Named FE defers to its own id (matches upstream's final
			// `if (node.id)` check, which runs before the lowercase-LHS
			// reject in the property-assignment tail).
			if fn.Kind == ast.KindFunctionExpression {
				name := fn.Name()
				if name != nil && name.Kind == ast.KindIdentifier {
					return isFirstLetterCapitalized(name.AsIdentifier().Text)
				}
			}
			left := ast.SkipParentheses(bin.Left)
			if left.Kind == ast.KindIdentifier {
				return isFirstLetterCapitalized(left.AsIdentifier().Text)
			}
			return false
		}
	}

	// Branches 5 & 6 — nested Arrow whose outer Arrow is itself in an
	// AssignmentExpression / PropertyAssignment position.
	if parent.Kind == ast.KindArrowFunction && parent.AsArrowFunction().Body == fn {
		grand := parent.Parent
		if grand != nil && !isMEAssign && functionReturnsJSXInternal(fn, true, pragma, tc) {
			switch grand.Kind {
			case ast.KindBinaryExpression:
				bin := grand.AsBinaryExpression()
				if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken && bin.Right == parent {
					left := ast.SkipParentheses(bin.Left)
					if left.Kind == ast.KindIdentifier {
						return isFirstLetterCapitalized(left.AsIdentifier().Text)
					}
					return false
				}
			case ast.KindPropertyAssignment:
				name := grand.AsPropertyAssignment().Name()
				if name != nil && name.Kind == ast.KindIdentifier {
					return isFirstLetterCapitalized(name.AsIdentifier().Text)
				}
				return false
			}
		}
	}

	// Branches 7 & 8 — inner function in a ReturnStatement whose enclosing
	// function itself sits in an AssignmentExpression / PropertyAssignment
	// position. Upstream first checks the inner FE's own id (if capitalized
	// return it), then walks functionExpr = parent.parent.parent.
	if parent.Kind == ast.KindReturnStatement {
		if fn.Kind == ast.KindFunctionExpression {
			name := fn.Name()
			if name != nil && name.Kind == ast.KindIdentifier && isFirstLetterCapitalized(name.AsIdentifier().Text) {
				return true
			}
		}
		// functionExpr = ReturnStatement.parent (Block) . parent (functionExpr)
		funcExpr := parent.Parent
		if funcExpr != nil {
			funcExpr = funcExpr.Parent
		}
		if funcExpr != nil && funcExpr.Parent != nil && !isMEAssign && functionReturnsJSXInternal(fn, true, pragma, tc) {
			gp := funcExpr.Parent
			switch gp.Kind {
			case ast.KindBinaryExpression:
				bin := gp.AsBinaryExpression()
				if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken && bin.Right == funcExpr {
					left := ast.SkipParentheses(bin.Left)
					if left.Kind == ast.KindIdentifier {
						return isFirstLetterCapitalized(left.AsIdentifier().Text)
					}
					return false
				}
			case ast.KindPropertyAssignment:
				name := gp.AsPropertyAssignment().Name()
				if name != nil && name.Kind == ast.KindIdentifier {
					return isFirstLetterCapitalized(name.AsIdentifier().Text)
				}
				return false
			}
		}
	}

	// Branch 9 — parent has a MemberExpression-style key
	// (e.g. `{ [obj.prop]: fn }` computed key resolving to a member access).
	if parent.Kind == ast.KindPropertyAssignment {
		nameNode := parent.AsPropertyAssignment().Name()
		if nameNode != nil && nameNode.Kind == ast.KindComputedPropertyName {
			keyExpr := ast.SkipParentheses(nameNode.AsComputedPropertyName().Expression)
			if keyExpr.Kind == ast.KindPropertyAccessExpression || keyExpr.Kind == ast.KindElementAccessExpression {
				if !functionReturnsJSXInternal(fn, false, pragma, tc) && !functionReturnsOnlyNull(fn) {
					return false
				}
			}
		}
	}

	// Branch 10 — Property method/no-id + !computed form.
	// In tsgo, the `method: true` arm is handled via the MethodDeclaration
	// path above. Here we handle the `!id && !computed` arm — an anonymous
	// FE/Arrow assigned as a PropertyAssignment initializer with Identifier
	// key. Strict isReturningJSX applies.
	if parent.Kind == ast.KindPropertyAssignment {
		pa := parent.AsPropertyAssignment()
		name := pa.Name()
		isComputed := name != nil && name.Kind == ast.KindComputedPropertyName
		hasId := fn.Kind == ast.KindFunctionExpression && fn.Name() != nil
		if !hasId && !isComputed {
			if name != nil && name.Kind == ast.KindIdentifier {
				if !isFirstLetterCapitalized(name.AsIdentifier().Text) {
					return false
				}
				return functionReturnsJSXInternal(fn, false, pragma, tc)
			}
			return false
		}
	}

	// Branch 11 — pragma component wrapper. tsgo preserves `()`, `as`,
	// `satisfies`, `<T>x`, and `x!` wrappers around the arrow argument
	// (ESTree flattens parens and has no equivalent for the TS-only
	// forms), so we walk up through every such wrapper before looking for
	// the enclosing CallExpression.
	effectiveParent := SkipExpressionWrappersUp(fn)
	if effectiveParent != nil && effectiveParent.Kind == ast.KindCallExpression {
		// When the caller threaded `wrappers`, consult the configured
		// list so user `settings.componentWrapperFunctions` entries
		// (`myMemo`, `MyLib.observer`, etc.) participate in stateless-
		// component classification — which in turn makes
		// `isStatelessComponentReturningNull` fire correctly on
		// null-only inner functions of user-configured wrappers. With
		// `wrappers == nil` we fall back to the hardcoded default
		// (memo / forwardRef, pragma-qualified or bare), preserving
		// every legacy caller's behavior.
		matched := false
		if wrappers != nil {
			matched = MatchesAnyComponentWrapperWithChecker(effectiveParent, fn, wrappers, pragma, tc)
		} else {
			matched = isPragmaComponentWrapperCall(effectiveParent, fn, pragma)
		}
		if matched && functionReturnsJSXInternal(fn, true, pragma, tc) {
			return true
		}
	}

	// Branch 12 — require allowed position AND isReturningJSXOrNull.
	if !isInAllowedPositionForComponent(fn) || !functionReturnsJSXInternal(fn, true, pragma, tc) {
		return false
	}

	// Branch 13 — isParentComponentNotStatelessComponent carve-out.
	if parent.Kind == ast.KindPropertyAssignment {
		name := parent.AsPropertyAssignment().Name()
		if name != nil && name.Kind == ast.KindIdentifier &&
			!isFirstLetterCapitalized(name.AsIdentifier().Text) &&
			len(fn.Parameters()) > 0 {
			return false
		}
	}

	// Branch 14 — `if (node.id) return capitalized(node.id.name)`.
	if fn.Kind == ast.KindFunctionExpression {
		name := fn.Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return isFirstLetterCapitalized(name.AsIdentifier().Text)
		}
	}

	// Branch 15 — isPropertyAssignment (MemberExpression LHS) but not
	// module.exports: reject when rightmost property name is lowercase.
	if isMEAssign && !isModuleExportsAssign {
		bin := parent.AsBinaryExpression()
		left := ast.SkipParentheses(bin.Left)
		if left.Kind == ast.KindPropertyAccessExpression {
			pa := left.AsPropertyAccessExpression()
			name := pa.Name()
			if name != nil && name.Kind == ast.KindIdentifier && !isFirstLetterCapitalized(name.AsIdentifier().Text) {
				return false
			}
		}
	}

	// Branch 16 — Property parent + returns only null ⇒ undefined. Already
	// handled by Branch 10 when !id & !computed (strict isReturningJSX
	// rejects null-only). The final `return node` for anonymous allowed-
	// position fallbacks remains.
	return true
}

// functionReturnsOnlyNull mirrors jsxUtil.isReturningOnlyNull: every
// return statement (at depth ≤ 1) returns the `null` literal, and at
// least one return exists. Arrow expression bodies count as a single
// return. Functions without any returns don't qualify.
func functionReturnsOnlyNull(fn *ast.Node) bool {
	var body *ast.Node
	switch fn.Kind {
	case ast.KindFunctionDeclaration:
		body = fn.AsFunctionDeclaration().Body
	case ast.KindFunctionExpression:
		body = fn.AsFunctionExpression().Body
	case ast.KindArrowFunction:
		af := fn.AsArrowFunction()
		body = af.Body
		if body != nil && body.Kind != ast.KindBlock {
			return ast.SkipParentheses(body).Kind == ast.KindNullKeyword
		}
	case ast.KindMethodDeclaration:
		body = fn.AsMethodDeclaration().Body
	case ast.KindGetAccessor:
		body = fn.AsGetAccessorDeclaration().Body
	}
	if body == nil {
		return false
	}
	sawReturn := false
	allNull := true
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n == nil {
			return false
		}
		switch n.Kind {
		case ast.KindReturnStatement:
			sawReturn = true
			rs := n.AsReturnStatement()
			if rs.Expression == nil || ast.SkipParentheses(rs.Expression).Kind != ast.KindNullKeyword {
				allNull = false
			}
			return false
		case ast.KindFunctionExpression,
			ast.KindFunctionDeclaration,
			ast.KindArrowFunction,
			ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor,
			ast.KindConstructor:
			return false
		}
		n.ForEachChild(visit)
		return false
	}
	visit(body)
	return sawReturn && allNull
}

// isInAllowedPositionForComponent mirrors eslint-plugin-react's
// `utils.isInAllowedPositionForComponent`: only parent node kinds in the
// allow-list may host a stateless functional component. Sequence expressions
// (`a, b`) pass through when `fn` is the last operand. ParenthesizedExpression
// wrappers (which ESTree flattens but tsgo preserves) are transparent so
// `const Hello = (init(), arrow)` — whose comma Sequence sits inside parens —
// still reaches the VariableDeclaration ancestor.
func isInAllowedPositionForComponent(fn *ast.Node) bool {
	parent := skipParenParents(fn)
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration,
		ast.KindPropertyAssignment,
		ast.KindReturnStatement,
		ast.KindExportAssignment,
		ast.KindArrowFunction:
		return true
	case ast.KindBinaryExpression:
		bin := parent.AsBinaryExpression()
		if bin.OperatorToken == nil {
			return false
		}
		switch bin.OperatorToken.Kind {
		case ast.KindEqualsToken:
			// AssignmentExpression — always allowed when `fn` is the RHS.
			return bin.Right == fn
		case ast.KindCommaToken:
			// SequenceExpression — only the last operand inherits its parent's
			// allowed-ness.
			if bin.Right == fn {
				return isInAllowedPositionForComponent(parent)
			}
		}
	}
	return false
}

// skipParenParents walks up through ParenthesizedExpression wrappers and
// returns the first non-paren ancestor of `node`, or nil.
func skipParenParents(node *ast.Node) *ast.Node {
	p := node.Parent
	for p != nil && p.Kind == ast.KindParenthesizedExpression {
		p = p.Parent
	}
	return p
}

// isPragmaComponentWrapperCall reports whether `call` is a React
// component-wrapping call — `<pragma>.memo(fn)` / `<pragma>.forwardRef(fn)` /
// bare `memo(fn)` / bare `forwardRef(fn)` — with `fn` as the first argument.
// Pragma defaults to `DefaultReactPragma` when empty. Mirrors upstream's
// default `wrapperFunctions` entries (`{property: 'memo', object: pragma}`,
// `{property: 'forwardRef', object: pragma}`); the user-configurable
// `settings.componentWrapperFunctions` is NOT honored.
//
// Call-level optional chains (`memo?.(fn)`) are rejected for the same
// reason as `MatchesAnyComponentWrapper` — upstream's
// `isPragmaComponentWrapper` reads `callee.name` on a plain Identifier
// callee, which fails on the OptionalCallExpression / ChainExpression
// shape Babel emits.
func isPragmaComponentWrapperCall(call, fn *ast.Node, pragma string) bool {
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	c := call.AsCallExpression()
	if c.QuestionDotToken != nil {
		return false
	}
	if c.Arguments == nil || len(c.Arguments.Nodes) == 0 {
		return false
	}
	// Paren- and TS-wrapper-transparent argument match: tsgo preserves
	// `()` / `as` / `satisfies` / `<T>x` / `x!` wrappers that ESTree
	// either flattens or doesn't have at all, so `React.memo((fn))` /
	// `React.memo(fn as Foo)` / `React.memo(fn!)` all surface the wrapper
	// as the first argument rather than `fn` itself.
	if SkipExpressionWrappers(c.Arguments.Nodes[0]) != fn {
		return false
	}
	callee := SkipExpressionWrappers(c.Expression)
	switch callee.Kind {
	case ast.KindIdentifier:
		text := callee.AsIdentifier().Text
		return text == "memo" || text == "forwardRef"
	case ast.KindPropertyAccessExpression:
		pa := callee.AsPropertyAccessExpression()
		obj := ast.SkipParentheses(pa.Expression)
		if obj.Kind != ast.KindIdentifier || obj.AsIdentifier().Text != pragma {
			return false
		}
		name := pa.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return false
		}
		text := name.AsIdentifier().Text
		return text == "memo" || text == "forwardRef"
	}
	return false
}

// FunctionReturnsJSXOrNull reports whether the function's body contains a
// `return <jsx/>` / `return null` / `return <pragma>.createElement(...)` at
// depth ≤ 1 (nested functions excluded), OR — for an arrow with expression
// body — whether that expression qualifies under the same rules.
// ConditionalExpression is traversed so `return cond ? <jsx/> : null`
// qualifies.
//
// Identifier returns (`return view` where `view` is bound to a JSX value)
// are resolved structurally via a local block scan. Use
// `FunctionReturnsJSXOrNullWithChecker` for full TypeChecker-based scope
// resolution.
//
// Mirrors upstream jsxUtil.isReturningJSX invoked with default arguments
// (which accept JSX, `null`, and `<pragma>.createElement(...)` returns).
// Pass an empty pragma to default to "React".
func FunctionReturnsJSXOrNull(fn *ast.Node, pragma string) bool {
	return functionReturnsJSXInternal(fn, true, pragma, nil)
}

// FunctionReturnsJSXOrNullWithChecker is the TypeChecker-aware variant of
// FunctionReturnsJSXOrNull. When `tc` is non-nil, Identifier returns are
// resolved through `GetSymbolAtLocation` → `Declarations[0]` →
// `VariableDeclaration.Initializer`, matching upstream's `findVariableByName`
// scope walk semantically (any binding the TS resolver can reach is
// considered, not just bindings in the immediately-enclosing block). When
// `tc` is nil, falls back to the local-block scan.
func FunctionReturnsJSXOrNullWithChecker(fn *ast.Node, pragma string, tc *checker.Checker) bool {
	return functionReturnsJSXInternal(fn, true, pragma, tc)
}

// FunctionReturnsJSX is the strict sibling of FunctionReturnsJSXOrNull:
// a `null` return does NOT qualify on its own. `<pragma>.createElement(...)`
// calls still qualify. Mirrors upstream jsxUtil.isReturningJSX invoked with
// `strict=true, ignoreNull=true`. `<pragma>.createElement(...)` calls still
// qualify. Pass an empty pragma to default to "React".
func FunctionReturnsJSX(fn *ast.Node, pragma string) bool {
	return functionReturnsJSXInternal(fn, false, pragma, nil)
}

// FunctionReturnsJSXWithChecker is the TypeChecker-aware strict variant.
// See FunctionReturnsJSXOrNullWithChecker for the resolution semantics.
func FunctionReturnsJSXWithChecker(fn *ast.Node, pragma string, tc *checker.Checker) bool {
	return functionReturnsJSXInternal(fn, false, pragma, tc)
}

func functionReturnsJSXInternal(fn *ast.Node, acceptNull bool, pragma string, tc *checker.Checker) bool {
	if fn == nil {
		return false
	}
	var body *ast.Node
	switch fn.Kind {
	case ast.KindFunctionDeclaration:
		body = fn.AsFunctionDeclaration().Body
	case ast.KindFunctionExpression:
		body = fn.AsFunctionExpression().Body
	case ast.KindArrowFunction:
		body = fn.AsArrowFunction().Body
		if body != nil && body.Kind != ast.KindBlock {
			return isJSXExpression(body, acceptNull, pragma, tc)
		}
	case ast.KindMethodDeclaration:
		body = fn.AsMethodDeclaration().Body
	case ast.KindGetAccessor:
		body = fn.AsGetAccessorDeclaration().Body
	case ast.KindSetAccessor:
		body = fn.AsSetAccessorDeclaration().Body
	}
	if body == nil {
		return false
	}
	found := false
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if found || n == nil {
			return found
		}
		switch n.Kind {
		case ast.KindReturnStatement:
			rs := n.AsReturnStatement()
			if rs.Expression != nil && isJSXExpression(rs.Expression, acceptNull, pragma, tc) {
				found = true
				return true
			}
		case ast.KindFunctionExpression,
			ast.KindFunctionDeclaration,
			ast.KindArrowFunction,
			ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor,
			ast.KindConstructor:
			return false
		}
		n.ForEachChild(visit)
		return found
	}
	visit(body)
	return found
}

// isJSXExpression reports whether `expr` may evaluate to JSX (or to `null`
// when `acceptNull` is true) on at least one control-flow path. Walks through
// ParenthesizedExpression, TS expression wrappers (`as` / `satisfies` / `<T>x`
// / `x!`), ConditionalExpression and LogicalExpression (NON-strict semantics:
// either side qualifying is enough), comma-sequence right-most operands, and
// optional chains. A `<pragma>.createElement(...)` CallExpression also
// qualifies — upstream's jsxUtil.isReturningJSX treats `createElement` calls
// as JSX returns. An Identifier resolves through its declaring
// VariableDeclaration initializer when present, mirroring upstream's
// `findVariableByName` lookup but limited to const/let initializers within
// the same scope (no re-binding analysis).
//
// Strict semantics note: upstream's jsxUtil.isReturningJSX accepts a
// `strict` parameter that, when true, requires BOTH branches of a
// Conditional / LogicalExpression to qualify. Every call site in upstream
// `Components.js` (rev 7.x) passes `strict=undefined` (falsy = non-strict),
// so the strict mode is effectively unreachable through this rule and the
// no-unstable-nested-components rule itself. We therefore match upstream's
// observable behavior (non-strict for all current consumers) and omit the
// strict parameter; if a future rule needs strict mode it should be added
// then with the corresponding test coverage.
//
// Pass `acceptNull=true` for `isReturningJSXOrNull`-style gates and `false`
// for the strict `isReturningJSX` (ignoreNull=true) gates. Pass `tc` (the
// active TypeChecker) when scope-resolved Identifier lookup is desired;
// pass nil to fall back to a local-block initializer scan.
//
// Identifier-via-initializer resolution is one-step only — matching
// upstream's `isJSXValue → findVariableByName → isJSX(variable)` chain
// where `isJSX` accepts ONLY a JSXElement / JSXFragment node and does not
// recurse. No depth bookkeeping needed because the function does not
// recurse on Identifier; the only recursion sites (Conditional / comma /
// `&&` / `||` / `??`) walk strictly smaller AST subtrees.
func isJSXExpression(expr *ast.Node, acceptNull bool, pragma string, tc *checker.Checker) bool {
	expr = SkipExpressionWrappers(expr)
	if expr == nil {
		return false
	}
	switch expr.Kind {
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
		return true
	case ast.KindNullKeyword:
		return acceptNull
	case ast.KindCallExpression:
		return IsCreateElementCallWithChecker(expr.AsCallExpression().Expression, pragma, tc)
	case ast.KindIdentifier:
		// Upstream's `isJSXValue` for the Identifier case calls
		// `findVariableByName` and then `isJSX(variable)` — and `isJSX`
		// ONLY accepts JSXElement / JSXFragment. It does NOT recurse
		// into ConditionalExpression / LogicalExpression / CallExpression
		// (`createElement`) / nested Identifiers. We mirror that here:
		// resolve the initializer one step, accept iff the resolved node
		// is itself a JSX element/fragment. Anything else returns false.
		init := resolveIdentifierInitializer(expr, tc)
		if init == nil {
			return false
		}
		init = SkipExpressionWrappers(init)
		switch init.Kind {
		case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
			return true
		}
		return false
	case ast.KindConditionalExpression:
		ce := expr.AsConditionalExpression()
		return isJSXExpression(ce.WhenTrue, acceptNull, pragma, tc) || isJSXExpression(ce.WhenFalse, acceptNull, pragma, tc)
	case ast.KindBinaryExpression:
		bin := expr.AsBinaryExpression()
		if bin.OperatorToken == nil {
			return false
		}
		switch bin.OperatorToken.Kind {
		case ast.KindCommaToken:
			return isJSXExpression(bin.Right, acceptNull, pragma, tc)
		case ast.KindAmpersandAmpersandToken,
			ast.KindBarBarToken,
			ast.KindQuestionQuestionToken:
			return isJSXExpression(bin.Left, acceptNull, pragma, tc) || isJSXExpression(bin.Right, acceptNull, pragma, tc)
		}
	}
	return false
}

// resolveIdentifierInitializer returns the value-side AST node that an
// Identifier reference is bound to, or nil when the binding cannot be
// determined.
//
//   - When `tc` is non-nil, asks the TypeChecker for the resolved symbol's
//     ValueDeclaration, then returns that declaration's `.Initializer`
//     (only the const/let/var case — class / function declarations have
//     no `Initializer` and aren't useful for JSX-return resolution). This
//     is upstream-equivalent to `findVariableByName` because the TS
//     resolver already follows the full lexical scope chain.
//
//   - When `tc` is nil, falls back to scanning enclosing Block /
//     SourceFile / ModuleBlock / CaseBlock statements for a
//     `VariableStatement` declaring `name` — catches the common
//     same-block initializer case without scope analysis.
func resolveIdentifierInitializer(ident *ast.Node, tc *checker.Checker) *ast.Node {
	if ident == nil || ident.Kind != ast.KindIdentifier {
		return nil
	}
	if tc != nil {
		if init := resolveIdentifierViaChecker(ident, tc); init != nil {
			return init
		}
		// Fall through: TypeChecker may not resolve the binding (e.g. a
		// type-only declaration in a JS file, or a synthesized symbol).
		// The local-block scan is a strict subset, so trying it as a
		// safety net costs nothing.
	}
	return lookupLocalInitializer(ident)
}

// resolveIdentifierViaChecker resolves an Identifier through the
// TypeChecker. Returns the initializer of the resolved
// VariableDeclaration, or nil when the symbol's value declaration is not a
// VariableDeclaration with an Initializer.
//
// All `checker.Checker` access on rslint runs without `--type-check` is
// nil; this function MUST be defensive even though its callers already
// gate on a non-nil `tc`. The double guard keeps the file safe to call
// from any future site that forgets the gate.
func resolveIdentifierViaChecker(ident *ast.Node, tc *checker.Checker) *ast.Node {
	if tc == nil || ident == nil {
		return nil
	}
	symbol := tc.GetSymbolAtLocation(ident)
	if symbol == nil {
		return nil
	}
	// `ValueDeclaration` is the symbol's primary declaration site; for
	// `const x = <div/>` it's the VariableDeclaration. When ValueDeclaration
	// is absent (interfaces, type aliases, ambient symbols) we explicitly
	// don't try to walk `Declarations` — those don't have a JSX value.
	decl := symbol.ValueDeclaration
	if decl == nil {
		// Fall back to Declarations[0] when ValueDeclaration is missing
		// but the symbol still has a concrete declaration (e.g. some
		// shorthand-property bindings).
		if len(symbol.Declarations) == 0 {
			return nil
		}
		decl = symbol.Declarations[0]
	}
	if decl == nil || decl.Kind != ast.KindVariableDeclaration {
		return nil
	}
	return decl.AsVariableDeclaration().Initializer
}

// lookupLocalInitializer mirrors upstream `variableUtil.findVariableByName`'s
// happy path for the cases this rule cares about: a const/let/var binding
// whose initializer is a JSX-or-createElement expression, declared in the
// same enclosing function/program. Walks lexically up the parent chain
// looking for a Block / SourceFile that contains a `VariableStatement`
// declaring `name` with a non-nil initializer; returns the initializer or
// nil when no such declaration is reachable.
//
// We deliberately do NOT re-implement a full scope manager — the tradeoff
// is that re-bindings (e.g. `let x = <div/>; x = 1; return x`) and
// destructuring patterns are not resolved, matching the conservative subset
// of upstream's behavior that the no-unstable-nested-components rule
// actually exercises in its tests.
func lookupLocalInitializer(ident *ast.Node) *ast.Node {
	if ident == nil || ident.Kind != ast.KindIdentifier {
		return nil
	}
	name := ident.AsIdentifier().Text
	for cur := ident.Parent; cur != nil; cur = cur.Parent {
		switch cur.Kind {
		case ast.KindBlock, ast.KindSourceFile, ast.KindCaseBlock, ast.KindModuleBlock:
			if init := findInitializerInStatements(cur, name); init != nil {
				return init
			}
		}
	}
	return nil
}

// findInitializerInStatements scans the statements of a Block / SourceFile
// (or anything with a `Statements` field exposed via ForEachChild) for a
// `VariableStatement` declaring `name` with a direct initializer.
func findInitializerInStatements(scope *ast.Node, name string) *ast.Node {
	if scope == nil {
		return nil
	}
	var found *ast.Node
	scope.ForEachChild(func(stmt *ast.Node) bool {
		if found != nil || stmt == nil {
			return false
		}
		var declList *ast.Node
		switch stmt.Kind {
		case ast.KindVariableStatement:
			declList = stmt.AsVariableStatement().DeclarationList
		}
		if declList == nil {
			return false
		}
		decls := declList.AsVariableDeclarationList()
		if decls == nil || decls.Declarations == nil {
			return false
		}
		for _, d := range decls.Declarations.Nodes {
			if d == nil || d.Kind != ast.KindVariableDeclaration {
				continue
			}
			vd := d.AsVariableDeclaration()
			if vd.Name() == nil || vd.Name().Kind != ast.KindIdentifier {
				continue
			}
			if vd.Name().AsIdentifier().Text != name {
				continue
			}
			if vd.Initializer != nil {
				found = vd.Initializer
				return true
			}
		}
		return false
	})
	return found
}

// isFirstLetterCapitalized mirrors eslint-plugin-react's helper of the same
// name (`lib/util/isFirstLetterCapitalized.js`). The semantics are:
//
//  1. Strip leading underscores: `_Foo` → "Foo" (so `_Foo` is treated the
//     same as `Foo`, matching upstream's `word.replace(/^_+/, ”)`).
//  2. A character is "capitalized" iff `unicode.ToUpper(r) == r` —
//     equivalent to upstream's `firstLetter.toUpperCase() === firstLetter`.
//
// Step 2 means non-cased characters (CJK, digits, emoji, symbols) all
// count as "capitalized" because they have no upper/lower mapping. This
// matters for non-ASCII identifiers like `function 不稳定组件()` — upstream
// classifies the function as a component (CJK char ≠ lowercase letter),
// and we must do the same to stay output-aligned.
func isFirstLetterCapitalized(s string) bool {
	stripped := strings.TrimLeft(s, "_")
	if stripped == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(stripped)
	if r == utf8.RuneError {
		return false
	}
	return unicode.ToUpper(r) == r
}

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
	if callee == nil {
		return false
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	callee = SkipExpressionWrappers(callee)

	// Bare callee: `createElement(arg)` — recognize when destructured
	// from the pragma module. Mirrors upstream's second branch of
	// `isCreateElement`.
	if callee.Kind == ast.KindIdentifier {
		if callee.AsIdentifier().Text != "createElement" {
			return false
		}
		// Without a TypeChecker we can't resolve the binding; bail to
		// match the conservative non-import-aware path (upstream also
		// returns false when the destructure gate fails).
		return IsDestructuredFromPragmaImport(callee, pragma, tc)
	}

	// Member-access callee: `<pragma>.createElement(arg)`.
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	if ast.IsOptionalChain(callee) {
		return false
	}
	prop := callee.AsPropertyAccessExpression()
	nameNode := prop.Name()
	if nameNode.Kind != ast.KindIdentifier || nameNode.AsIdentifier().Text != "createElement" {
		return false
	}
	pragmaExpr := SkipExpressionWrappers(prop.Expression)
	if pragmaExpr.Kind != ast.KindIdentifier || pragmaExpr.AsIdentifier().Text != pragma {
		return false
	}
	return true
}

// GetJsxPropName returns the display name of a JSX node.
// For JsxAttribute: returns the attribute name (including namespaced names like "foo:bar").
// For JsxSpreadAttribute: returns "spread".
// For Identifier nodes (e.g. tag names): returns the identifier text.
// For unknown nodes: returns "".
func GetJsxPropName(node *ast.Node) string {
	if ast.IsJsxAttribute(node) {
		nameNode := node.AsJsxAttribute().Name()
		if nameNode.Kind == ast.KindIdentifier {
			return nameNode.AsIdentifier().Text
		}
		if nameNode.Kind == ast.KindJsxNamespacedName {
			ns := nameNode.AsJsxNamespacedName()
			return ns.Namespace.AsIdentifier().Text + ":" + ns.Name().AsIdentifier().Text
		}
	}
	if ast.IsJsxSpreadAttribute(node) {
		return "spread"
	}
	if node.Kind == ast.KindIdentifier {
		return node.AsIdentifier().Text
	}
	return ""
}

// GetJsxParentElement returns the JsxOpeningElement or JsxSelfClosingElement that
// owns the given JsxAttribute (or JsxSpreadAttribute), or nil if not applicable.
func GetJsxParentElement(attr *ast.Node) *ast.Node {
	if attr == nil || attr.Parent == nil {
		return nil
	}
	grandParent := attr.Parent.Parent
	if grandParent == nil {
		return nil
	}
	switch grandParent.Kind {
	case ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement:
		return grandParent
	}
	return nil
}

// GetJsxTagName returns the tag-name node of a JsxOpeningElement or
// JsxSelfClosingElement, or nil for other kinds.
func GetJsxTagName(element *ast.Node) *ast.Node {
	if element == nil {
		return nil
	}
	switch element.Kind {
	case ast.KindJsxOpeningElement:
		return element.AsJsxOpeningElement().TagName
	case ast.KindJsxSelfClosingElement:
		return element.AsJsxSelfClosingElement().TagName
	}
	return nil
}

// GetJsxChildren returns the child-node list of a JsxElement or JsxFragment,
// or nil for other kinds (JsxSelfClosingElement has no children) and when the
// container's child list is absent.
func GetJsxChildren(parent *ast.Node) []*ast.Node {
	if parent == nil {
		return nil
	}
	switch parent.Kind {
	case ast.KindJsxElement:
		if parent.AsJsxElement().Children == nil {
			return nil
		}
		return parent.AsJsxElement().Children.Nodes
	case ast.KindJsxFragment:
		if parent.AsJsxFragment().Children == nil {
			return nil
		}
		return parent.AsJsxFragment().Children.Nodes
	}
	return nil
}

// GetJsxElementAttributes returns the attribute nodes of a JsxOpeningElement or
// JsxSelfClosingElement, or nil for other kinds or when the element has no
// attributes. Each returned node is either a JsxAttribute or a JsxSpreadAttribute.
func GetJsxElementAttributes(element *ast.Node) []*ast.Node {
	if element == nil {
		return nil
	}
	var attrs *ast.Node
	switch element.Kind {
	case ast.KindJsxOpeningElement:
		attrs = element.AsJsxOpeningElement().Attributes
	case ast.KindJsxSelfClosingElement:
		attrs = element.AsJsxSelfClosingElement().Attributes
	default:
		return nil
	}
	if attrs == nil {
		return nil
	}
	list := attrs.AsJsxAttributes()
	if list == nil || list.Properties == nil {
		return nil
	}
	return list.Properties.Nodes
}

// GetJsxElementTypeString returns the jsx-ast-utils `elementType(node)`
// equivalent — the dotted / namespaced display string of a JSX tag name as
// an ESTree-compatible source caller would see it. `node` may be either a
// JsxOpeningElement / JsxSelfClosingElement, or a raw tag-name node. Returns
// "" for shapes that don't correspond to a legal React/JSX element type
// (e.g. a computed member access), so callers can treat "" as "not a user
// component".
//
// Supported tag shapes:
//
//   - `<Foo>` / `<foo>`       → "Foo" / "foo"
//   - `<Foo.Bar.Baz>`         → "Foo.Bar.Baz" (PropertyAccessExpression chain)
//   - `<this.Foo>`            → "this.Foo" (ThisKeyword base)
//   - `<ns:Name>`             → "ns:Name" (JsxNamespacedName)
//
// This is AST-driven — interior whitespace or comments in unusual forms
// (e.g. `<Foo . Bar />`) are normalized away, matching jsx-ast-utils.
func GetJsxElementTypeString(node *ast.Node) string {
	tagName := node
	if node != nil {
		if t := GetJsxTagName(node); t != nil {
			tagName = t
		}
	}
	return tagNameString(tagName)
}

func tagNameString(tagName *ast.Node) string {
	if tagName == nil {
		return ""
	}
	switch tagName.Kind {
	case ast.KindIdentifier:
		return tagName.AsIdentifier().Text
	case ast.KindThisKeyword:
		return "this"
	case ast.KindJsxNamespacedName:
		ns := tagName.AsJsxNamespacedName()
		if ns.Namespace == nil || ns.Name() == nil {
			return ""
		}
		if ns.Namespace.Kind != ast.KindIdentifier || ns.Name().Kind != ast.KindIdentifier {
			return ""
		}
		return ns.Namespace.AsIdentifier().Text + ":" + ns.Name().AsIdentifier().Text
	case ast.KindPropertyAccessExpression:
		pa := tagName.AsPropertyAccessExpression()
		base := tagNameString(pa.Expression)
		if base == "" {
			return ""
		}
		nameNode := pa.Name()
		if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			return ""
		}
		return base + "." + nameNode.AsIdentifier().Text
	}
	return ""
}

// IsDOMComponent reports whether a JSX opening/self-closing element refers to
// an intrinsic (DOM) element like <div> or <svg:path>, rather than a user
// component like <Foo> or <Foo.Bar>.
//
// Mirrors ESLint-plugin-react's `jsxUtil.isDOMComponent`: a tag name is
// intrinsic iff `elementType(node)` starts with a lowercase letter (regex
// `/^[a-z]/`). For member-expression tags (`<foo.bar>`, `<this.Foo>`) this
// means the classification is decided by the leftmost base identifier's
// first character — so `<foo.bar>` is DOM (matches ESLint, even though the
// runtime React behavior is "always user component"), while `<Foo.Bar>` is
// a user component.
func IsDOMComponent(element *ast.Node) bool {
	tagName := GetJsxTagName(element)
	if tagName == nil {
		return false
	}
	var text string
	switch tagName.Kind {
	case ast.KindIdentifier:
		text = tagName.AsIdentifier().Text
	case ast.KindJsxNamespacedName:
		ns := tagName.AsJsxNamespacedName()
		if ns.Namespace == nil || ns.Name() == nil {
			return false
		}
		text = ns.Namespace.AsIdentifier().Text + ":" + ns.Name().AsIdentifier().Text
	case ast.KindPropertyAccessExpression:
		// Walk to the leftmost base — its first character decides the
		// classification, matching `/^[a-z]/.test(elementType(node))`.
		base := tagName
		for base.Kind == ast.KindPropertyAccessExpression {
			base = base.AsPropertyAccessExpression().Expression
		}
		switch base.Kind {
		case ast.KindIdentifier:
			text = base.AsIdentifier().Text
		case ast.KindThisKeyword:
			// `<this.Foo>` — jsx-ast-utils' elementType returns "this.Foo",
			// first char is lowercase → DOM per ESLint's regex.
			text = "this"
		default:
			return false
		}
	default:
		return false
	}
	if text == "" {
		return false
	}
	first := text[0]
	return first >= 'a' && first <= 'z'
}

// ComponentMap maps a component tag name (e.g. "a", "Link") to the set of
// attribute names that identify its link target (e.g. ["href"] or ["to"]).
type ComponentMap map[string][]string

// DefaultLinkComponents returns the default link-component map: {"a": ["href"]}.
func DefaultLinkComponents() ComponentMap {
	return ComponentMap{"a": {"href"}}
}

// DefaultFormComponents returns the default form-component map: {"form": ["action"]}.
func DefaultFormComponents() ComponentMap {
	return ComponentMap{"form": {"action"}}
}

// ReadComponentsFromSettings extracts a component-name→attribute-list map
// from `settings.<key>`, matching upstream `util/linkComponents`.
//
// Upstream builds the map via `new Map(DEFAULT.concat(settings[key]).map(…))`,
// where same-key entries use last-wins (replace) semantics. This function
// mirrors that: a settings entry for an already-present component replaces
// the base entry entirely.
//
// Shapes accepted (each entry may appear standalone or as an element of an
// outer array, mirroring upstream's `DEFAULT.concat(settings[key] || [])`):
//
//   - string: "Link"                                    → {Link: [defaultAttr]}
//   - {name, <attrField>}: <attrField> string or []str  → {name: [attr…]}
//
// `attrField` is "linkAttribute" for linkComponents and "formAttribute" for
// formComponents — upstream uses distinct field names for each category
// (`value.linkAttribute` vs `value.formAttribute`), so getting this wrong
// would silently fall back to the default attribute for every custom form
// component the user configures.
func ReadComponentsFromSettings(settings map[string]interface{}, key, attrField, defaultAttr string, base ComponentMap) ComponentMap {
	out := ComponentMap{}
	for k, v := range base {
		out[k] = slices.Clone(v)
	}
	if settings == nil {
		return out
	}
	raw, ok := settings[key]
	if !ok {
		return out
	}
	// addOne mirrors upstream's per-entry mapper inside the Map constructor.
	// Each entry REPLACES any previous entry with the same name (last-wins),
	// matching `new Map([...pairs])` semantics.
	addOne := func(entry interface{}) {
		switch e := entry.(type) {
		case string:
			out[e] = []string{defaultAttr}
		case map[string]interface{}:
			name, _ := e["name"].(string)
			if name == "" {
				return
			}
			var attrs []string
			// Mirrors upstream's `[].concat(value[attrField])` coercion:
			// string → single-element list, array → as-is, missing → empty
			// (which we backfill with the default attribute).
			switch la := e[attrField].(type) {
			case string:
				attrs = []string{la}
			case []interface{}:
				for _, v := range la {
					if s, ok := v.(string); ok {
						attrs = append(attrs, s)
					}
				}
			}
			if len(attrs) == 0 {
				attrs = []string{defaultAttr}
			}
			out[name] = attrs
		}
	}
	// Upstream accepts either a single entry (string/object) or an array of
	// them at `settings[key]`. JS's `[].concat(x)` flattens both into the
	// final list; we mirror that by accepting either shape here.
	switch r := raw.(type) {
	case string:
		addOne(r)
	case map[string]interface{}:
		addOne(r)
	case []interface{}:
		for _, entry := range r {
			addOne(entry)
		}
	}
	return out
}

// IsJsxElementLike reports whether node is a JsxElement or
// JsxSelfClosingElement — the two tsgo kinds that correspond to ESTree's
// single `JSXElement` type.
func IsJsxElementLike(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return ast.IsJsxElement(node) || ast.IsJsxSelfClosingElement(node)
}

// IsJsxLike mirrors eslint-plugin-react's `jsxUtil.isJSX` — true for a JSX
// element (either tag form) or a JSX fragment.
func IsJsxLike(node *ast.Node) bool {
	if node == nil {
		return false
	}
	return IsJsxElementLike(node) || ast.IsJsxFragment(node)
}

// EnclosingClass returns the nearest ClassDeclaration / ClassExpression
// ancestor of `node`, or nil when `node` is at the top level. Used by rules
// that need to test whether a class member belongs to a React component.
func EnclosingClass(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	for p := node.Parent; p != nil; p = p.Parent {
		switch p.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			return p
		}
	}
	return nil
}

// BindingIdentifierName returns the identifier text of a named declaration's
// binding, or "" when the declaration is anonymous, the binding is a pattern
// rather than a bare Identifier, or `n` is nil.
func BindingIdentifierName(n *ast.Node) string {
	if n == nil {
		return ""
	}
	name := n.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.AsIdentifier().Text
}

// FunctionParameters returns the parameter list of a function-like node
// (FunctionDeclaration / FunctionExpression / ArrowFunction). Returns nil
// for nil input or any other node kind. Methods / accessors / constructors
// are intentionally not covered — callers that need them should add the
// kind explicitly to keep this helper a thin shim over the common shapes.
func FunctionParameters(fn *ast.Node) []*ast.Node {
	if fn == nil {
		return nil
	}
	switch fn.Kind {
	case ast.KindFunctionDeclaration:
		fd := fn.AsFunctionDeclaration()
		if fd.Parameters == nil {
			return nil
		}
		return fd.Parameters.Nodes
	case ast.KindFunctionExpression:
		fe := fn.AsFunctionExpression()
		if fe.Parameters == nil {
			return nil
		}
		return fe.Parameters.Nodes
	case ast.KindArrowFunction:
		af := fn.AsArrowFunction()
		if af.Parameters == nil {
			return nil
		}
		return af.Parameters.Nodes
	}
	return nil
}

// FirstParamType returns the type annotation of the first parameter of `fn`
// (a FunctionDeclaration / FunctionExpression / ArrowFunction), or nil when
// the function has no parameters or the first parameter is untyped.
func FirstParamType(fn *ast.Node) *ast.Node {
	params := FunctionParameters(fn)
	if len(params) == 0 {
		return nil
	}
	pd := params[0].AsParameterDeclaration()
	if pd == nil {
		return nil
	}
	return pd.Type
}

// PropWrapperEntry encodes one entry of `settings.propWrapperFunctions`. The
// raw entries can be either a bare string (`"forbidExtraProps"`) or an
// `{object, property}` pair (`{ object: "Object", property: "assign" }` →
// matches `Object.assign(...)`). Both shapes are normalized to this struct.
type PropWrapperEntry struct {
	// Object is the receiver portion of a member-call wrapper (e.g.
	// `"Object"` for `Object.assign`). Empty for bare-identifier wrappers.
	Object string
	// Property is the function name (e.g. `"assign"` for `Object.assign`,
	// or `"forbidExtraProps"` for a bare-identifier wrapper).
	Property string
}

// GetPropWrapperFunctions reads `settings.propWrapperFunctions` from the
// rslint config and returns the parsed entries. Unknown shapes (a non-array
// value, an entry that's neither a string nor a `{object, property}` map,
// an entry with empty `property`) are silently dropped — this matches
// eslint-plugin-react's `propWrapperUtil` permissive parsing.
func GetPropWrapperFunctions(settings map[string]interface{}) []PropWrapperEntry {
	v, ok := settings["propWrapperFunctions"]
	if !ok {
		return nil
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	var out []PropWrapperEntry
	for _, entry := range arr {
		switch t := entry.(type) {
		case string:
			if t != "" {
				if dot := strings.IndexByte(t, '.'); dot > 0 && dot < len(t)-1 {
					// Allow `"Object.assign"` style strings (legacy upstream
					// shape) by splitting on the first dot.
					out = append(out, PropWrapperEntry{Object: t[:dot], Property: t[dot+1:]})
				} else {
					out = append(out, PropWrapperEntry{Property: t})
				}
			}
		case map[string]interface{}:
			obj, _ := t["object"].(string)
			prop, _ := t["property"].(string)
			if prop == "" {
				continue
			}
			out = append(out, PropWrapperEntry{Object: obj, Property: prop})
		}
	}
	return out
}

// IsPropWrapperCall reports whether `call` is a CallExpression whose callee
// matches one of the user-configured `propWrapperFunctions` entries.
//
// Supports:
//   - bare identifier callees: `forbidExtraProps(...)`, `merge(...)` —
//     match an entry with empty `Object`.
//   - dotted-property callees: `Object.assign(...)`, `_.assign(...)` —
//     match an entry whose `Object` and `Property` both equal the receiver
//     and method names respectively.
//
// `call` may be wrapped in parens / TS expression wrappers; the callee is
// unwrapped via `SkipExpressionWrappers`. Anything more complex (computed
// access, optional-chain wrappers around the callee head) is treated as
// not matching.
func IsPropWrapperCall(call *ast.Node, wrappers []PropWrapperEntry) bool {
	if len(wrappers) == 0 || call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	callee := SkipExpressionWrappers(call.AsCallExpression().Expression)
	switch callee.Kind {
	case ast.KindIdentifier:
		name := callee.AsIdentifier().Text
		for _, w := range wrappers {
			if w.Object == "" && w.Property == name {
				return true
			}
		}
	case ast.KindPropertyAccessExpression:
		pa := callee.AsPropertyAccessExpression()
		obj := SkipExpressionWrappers(pa.Expression)
		nameNode := pa.Name()
		if obj.Kind != ast.KindIdentifier || nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			return false
		}
		objText := obj.AsIdentifier().Text
		propText := nameNode.AsIdentifier().Text
		for _, w := range wrappers {
			if w.Object == objText && w.Property == propText {
				return true
			}
		}
	}
	return false
}

// EntityNameRightmost returns the rightmost Identifier of a TypeReference's
// EntityName. For a bare `Foo`, returns `Foo`. For `A.B.C`, returns `C`.
// Returns nil if no identifier can be extracted.
func EntityNameRightmost(name *ast.Node) *ast.Node {
	for name != nil {
		switch name.Kind {
		case ast.KindIdentifier:
			return name
		case ast.KindQualifiedName:
			qn := name.AsQualifiedName()
			if qn == nil || qn.Right == nil {
				return nil
			}
			name = qn.Right
		default:
			return nil
		}
	}
	return nil
}
