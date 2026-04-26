// Package exhaustive_deps implements the rslint port of upstream
// `react-hooks/exhaustive-deps`.
//
// The upstream rule (facebook/react/packages/eslint-plugin-react-hooks)
// leans heavily on ESLint's scope manager: every Identifier reference
// is preresolved to a Variable, references inside a callback can be
// enumerated by walking the callback's Scope.references, and references
// across the entire file get a `from` Scope and `resolved` Variable.
// rslint has no scope manager. Instead we resolve identifiers via the
// TypeChecker (`GetSymbolAtLocation`), then look at the symbol's
// declaration to infer scope membership. When the TypeChecker is
// unavailable (gap files, parse errors), we fall back to a name-based
// best-effort: scan the call's enclosing function for declarations
// matching the identifier text, and treat anything else as external.
//
// The diagnostics intentionally mirror upstream's wording so consumers
// of the JS rule can switch over without rewriting message assertions.
// The autofix and suggestion shapes also mirror upstream — replacing
// the deps-array text wholesale, or inserting a deps array after the
// callback when none was provided.
package exhaustive_deps

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/react_hooksutil"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var _ = fmt.Sprintf // ensure fmt referenced
var _ core.TextRange

// All hook-name / namespace / function-like queries are delegated to
// `react_hooksutil`. This file deliberately keeps zero local copies of
// those predicates so that any semantic change is made once and picked
// up by every rule in the plugin.

// getReactiveHookCallbackIndex returns the index of the callback argument
// for a known reactive hook. Mirrors upstream's same-named helper.
//   - 0 for useEffect / useLayoutEffect / useInsertionEffect / useCallback / useMemo
//   - 1 for useImperativeHandle
//   - 0 for additionalHooks-matching custom hooks
//   - -1 otherwise
func getReactiveHookCallbackIndex(callee *ast.Node, additionalHooks *regexp.Regexp) int {
	n := react_hooksutil.StripReactNamespace(callee)
	if n == nil || n.Kind != ast.KindIdentifier {
		// `additionalHooks` is matched against the full path (e.g. `useFoo` or `Namespace.useFoo`);
		// only top-level Identifier callees pass this fast path. Member expressions
		// other than `React.<x>` skip additionalHooks entirely (matches upstream).
		return -1
	}
	name := n.AsIdentifier().Text
	switch name {
	case "useEffect", "useLayoutEffect", "useInsertionEffect", "useCallback", "useMemo":
		return 0
	case "useImperativeHandle":
		return 1
	}
	// `additionalHooks` only applies to bare-identifier callees, mirroring
	// upstream's `node === calleeNode` gate. `React.useCustomEffect` is
	// intentionally NOT treated as a reactive hook by the additionalHooks
	// path — only the unqualified `useCustomEffect` is.
	if additionalHooks != nil && n == callee && additionalHooks.MatchString(name) {
		return 0
	}
	return -1
}

// Hook-name / namespace / function-like queries are imported from
// `react_hooksutil` — the local helpers below are aliases so that
// the rest of this file reads naturally without sprinkling the
// package name everywhere.
var (
	stripReactNamespace     = react_hooksutil.StripReactNamespace
	isFunctionLikeContainer = react_hooksutil.IsFunctionLikeContainer
	findEnclosingFunction   = react_hooksutil.FindEnclosingFunction
	hasAsyncModifier        = react_hooksutil.HasAsyncModifier
)

// effectNameRegex mirrors upstream's `/Effect($|[^a-z])/g` — see also
// `react_hooksutil.IsEffectStyleHookName`. We keep a local handle so
// the existing call sites can stay terse.
var effectNameRegex = regexp.MustCompile(`Effect($|[^a-z])`)

// stripAsExpression unwraps any `as X` / `<X>...` / `satisfies X` wrapper.
func stripAsExpression(node *ast.Node) *ast.Node {
	for node != nil {
		switch node.Kind {
		case ast.KindAsExpression:
			node = node.AsAsExpression().Expression
		case ast.KindTypeAssertionExpression:
			node = node.AsTypeAssertion().Expression
		case ast.KindSatisfiesExpression:
			node = node.AsSatisfiesExpression().Expression
		case ast.KindParenthesizedExpression:
			node = node.AsParenthesizedExpression().Expression
		case ast.KindNonNullExpression:
			node = node.AsNonNullExpression().Expression
		default:
			return node
		}
	}
	return node
}

// containsNode reports whether `descendant` is inside `ancestor` by node range.
// Returns false when either is nil. Inclusive on both ends.
func containsNode(ancestor, descendant *ast.Node) bool {
	if ancestor == nil || descendant == nil {
		return false
	}
	return descendant.Pos() >= ancestor.Pos() && descendant.End() <= ancestor.End()
}

// nodeText returns the trimmed source text for `node`.
func nodeText(sf *ast.SourceFile, node *ast.Node) string {
	if node == nil {
		return ""
	}
	return utils.TrimmedNodeText(sf, node)
}

// analyzePropertyChainText converts a property chain into a dotted string.
// Returns "" + ok=false on shapes upstream's `analyzePropertyChain` would
// reject (which becomes a "complex expression" / "literal not a valid
// dependency" diagnostic at the call site).
//
// Side effect: when `optionalChains` is non-nil, the returned key is
// recorded as either optional (when the path was first seen via `?.`) or
// required (regular `.`). Mirrors upstream's `analyzePropertyChain` +
// `markNode`.
//
// tsgo-specific: peel ParenthesizedExpression and `as` / `satisfies`
// (mirrors upstream's outer-level `while (init.type === 'TSAsExpression'
// || init.type === 'AsExpression')` loop in visitCallExpression). We do
// NOT peel `NonNullExpression` (`x!`) or `TypeAssertionExpression`
// (`<T>x`) — upstream's `analyzePropertyChain` rejects those, so the
// rule reports them as "complex expression" rather than treating them as
// transparent.
func analyzePropertyChainText(node *ast.Node, optionalChains map[string]bool) (string, bool) {
	if node == nil {
		return "", false
	}
	n := ast.SkipParentheses(node)
	for {
		switch n.Kind {
		case ast.KindAsExpression:
			n = ast.SkipParentheses(n.AsAsExpression().Expression)
			continue
		case ast.KindSatisfiesExpression:
			n = ast.SkipParentheses(n.AsSatisfiesExpression().Expression)
			continue
		}
		break
	}
	switch n.Kind {
	case ast.KindIdentifier:
		name := n.AsIdentifier().Text
		if optionalChains != nil {
			optionalChains[name] = false
		}
		return name, true
	case ast.KindThisKeyword:
		// Reject `this` — upstream's `analyzePropertyChain` only handles
		// Identifier/JSXIdentifier as the leaf; `this` forces a
		// "complex expression" diagnostic on declared deps.
		return "", false
	case ast.KindPropertyAccessExpression:
		pae := n.AsPropertyAccessExpression()
		object, ok := analyzePropertyChainText(pae.Expression, optionalChains)
		if !ok {
			return "", false
		}
		prop := pae.Name()
		if prop == nil || prop.Kind != ast.KindIdentifier {
			return "", false
		}
		result := object + "." + prop.AsIdentifier().Text
		if optionalChains != nil {
			optional := pae.QuestionDotToken != nil
			markOptionalChain(optionalChains, result, optional)
		}
		return result, true
	}
	return "", false
}

// markOptionalChain mirrors upstream's `markNode`: a path is marked optional
// only if every observed access used `?.`; the first non-optional sighting
// pins it to false.
func markOptionalChain(m map[string]bool, key string, optional bool) {
	if optional {
		if _, ok := m[key]; !ok {
			m[key] = true
		}
	} else {
		m[key] = false
	}
}

// getDependencyNode walks up from `node` through enclosing
// PropertyAccessExpression chains, stopping at the deepest receiver still
// useful as a dependency key. Mirrors upstream's `getDependency` exactly:
//
//   - `props` -> `props`
//   - `props.foo` (read) -> `props.foo`
//   - `props.foo.bar` -> `props.foo.bar`
//   - `props.foo.current` -> `props.foo.current` (special: `.current` ends the walk)
//   - `props.foo()` -> `props` (method call: don't recurse, return receiver)
//   - `props.foo.bar()` -> `props.foo` (same)
//   - `props.foo = ...` (LHS) -> `props` (the assignment makes the receiver the dep)
func getDependencyNode(node *ast.Node) *ast.Node {
	cur := node
	for cur != nil && cur.Parent != nil {
		p := cur.Parent
		// `(x).y` — peel ParenthesizedExpression transparently (ESTree
		// flattens this; tsgo preserves it).
		if p.Kind == ast.KindParenthesizedExpression {
			cur = p
			continue
		}
		// NOTE: We deliberately do NOT peel `as`/`satisfies`/`!` here.
		// Upstream's `getDependency` / `analyzePropertyChain` reject
		// these wrappers, so a body reference of `user!.name` resolves
		// to dep key `user` (the receiver walk stops at the NonNull
		// wrapper because parent kind != MemberExpression).
		if p.Kind != ast.KindPropertyAccessExpression {
			break
		}
		pae := p.AsPropertyAccessExpression()
		if pae.Expression != cur {
			break
		}
		// `.current` terminates the upward walk — upstream stops here so
		// `.current` is reported, not the parent reference.
		propName := pae.Name()
		if propName == nil || propName.Kind != ast.KindIdentifier {
			break
		}
		if propName.AsIdentifier().Text == "current" {
			break
		}
		// `.foo()` method call: upstream's condition is
		// `!(parent.parent.type === 'CallExpression' && parent.parent.callee === parent)`.
		// When parent IS the callee of a CallExpression, we DON'T recurse
		// — we drop out of the loop and fall through to the else branch
		// which returns `cur` (the current node, the receiver).
		if p.Parent != nil && p.Parent.Kind == ast.KindCallExpression {
			callExpr := p.Parent.AsCallExpression()
			if callExpr.Expression == p {
				break
			}
		}
		cur = p
	}
	// Assignment LHS: `obj.prop = ...` and `obj['prop'] = ...` both make
	// `obj` the dependency. Mirrors upstream's same branch — the LHS
	// member access doesn't represent a stable read of the property,
	// it's a write through the receiver.
	if cur != nil && cur.Parent != nil {
		switch cur.Kind {
		case ast.KindPropertyAccessExpression:
			if be, ok := getAssignmentBinaryExpr(cur.Parent); ok && be.Left == cur {
				return cur.AsPropertyAccessExpression().Expression
			}
		case ast.KindElementAccessExpression:
			if be, ok := getAssignmentBinaryExpr(cur.Parent); ok && be.Left == cur {
				return cur.AsElementAccessExpression().Expression
			}
		}
	}
	return cur
}

// getAssignmentBinaryExpr returns the BinaryExpression iff `node` is one
// with an assignment-shape operator. Mirrors ESTree's AssignmentExpression
// flag in tsgo's collapsed BinaryExpression model.
func getAssignmentBinaryExpr(node *ast.Node) (*ast.BinaryExpression, bool) {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return nil, false
	}
	be := node.AsBinaryExpression()
	if be.OperatorToken == nil {
		return nil, false
	}
	switch be.OperatorToken.Kind {
	case ast.KindEqualsToken,
		ast.KindPlusEqualsToken, ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken, ast.KindAsteriskAsteriskEqualsToken,
		ast.KindSlashEqualsToken, ast.KindPercentEqualsToken,
		ast.KindLessThanLessThanEqualsToken, ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
		ast.KindAmpersandEqualsToken, ast.KindBarEqualsToken,
		ast.KindCaretEqualsToken,
		ast.KindAmpersandAmpersandEqualsToken, ast.KindBarBarEqualsToken,
		ast.KindQuestionQuestionEqualsToken:
		return be, true
	}
	return nil, false
}

// Options holds the parsed rule options.
type Options struct {
	AdditionalHooks                                 *regexp.Regexp
	EnableDangerousAutofixThisMayCauseInfiniteLoops bool
	RequireExplicitEffectDeps                       bool
	// AutoDepsHooks: experimental_autoDependenciesHooks. When the hook's
	// bare name is in this list, missing deps are inferred ("auto deps")
	// rather than being flagged. Mirrors upstream's same-named option.
	AutoDepsHooks map[string]bool
}

// parseOptions parses the rule's options object. Both array and bare-object
// shapes are supported via `utils.GetOptionsMap`.
func parseOptions(options any, settings map[string]interface{}) Options {
	opts := Options{}
	optsMap := utils.GetOptionsMap(options)
	// `addlSet` tracks whether the rule-level `additionalHooks` field was
	// PRESENT in options (even as empty string). Mirrors upstream's
	// `rawOptions.additionalHooks` truthiness check — falsy values fall
	// back to the settings entry, so we record presence here and only
	// fall back when the rule-level option is absent or empty.
	addlSet := false
	if optsMap != nil {
		if raw, ok := optsMap["additionalHooks"].(string); ok {
			addlSet = raw != ""
			if raw != "" {
				if re, err := regexp.Compile(raw); err == nil {
					opts.AdditionalHooks = re
				}
			}
		}
		if v, ok := optsMap["enableDangerousAutofixThisMayCauseInfiniteLoops"].(bool); ok {
			opts.EnableDangerousAutofixThisMayCauseInfiniteLoops = v
		}
		if v, ok := optsMap["requireExplicitEffectDeps"].(bool); ok {
			opts.RequireExplicitEffectDeps = v
		}
		if raw, ok := optsMap["experimental_autoDependenciesHooks"].([]interface{}); ok {
			opts.AutoDepsHooks = map[string]bool{}
			for _, item := range raw {
				if s, ok := item.(string); ok {
					opts.AutoDepsHooks[s] = true
				}
			}
		}
	}
	// Settings fallback for additionalHooks (matches upstream's
	// `getAdditionalEffectHooksFromSettings`, but only when the rule-level
	// option is absent OR empty). Delegates to react_hooksutil.
	if !addlSet && opts.AdditionalHooks == nil {
		opts.AdditionalHooks = react_hooksutil.AdditionalHooksFromSettings(settings, "additionalHooks")
	}
	return opts
}

// declaredDependency is the parsed form of an entry in the deps array.
type declaredDependency struct {
	Key  string
	Node *ast.Node
}

// dependency is a used reference observed inside the callback body.
type dependency struct {
	IsStable bool
	IsRef    bool   // true if it's `<x>.current` reference
	Refs     []*depReference
	First    *ast.Node // first observed reference identifier (for diagnostics)
}

// depReference is a single observed reference inside the callback body —
// the identifier node and the symbol it resolved to (or nil for fallback).
type depReference struct {
	Identifier   *ast.Node
	Symbol       *ast.Symbol
	WriteExpr    *ast.Node // BinaryExpression representing assignment, when this reference is written
	InCleanup    bool
	DepNodeRoot  *ast.Node
}

// dependencyTreeNode mirrors upstream's `DependencyTreeNode`.
type dependencyTreeNode struct {
	IsUsed                 bool
	IsSatisfiedRecursively bool
	IsSubtreeUsed          bool
	Children               map[string]*dependencyTreeNode
}

func newDepTreeNode() *dependencyTreeNode {
	return &dependencyTreeNode{Children: map[string]*dependencyTreeNode{}}
}

// getOrCreateNodeByPath mirrors upstream's same-named helper.
func getOrCreateNodeByPath(root *dependencyTreeNode, path string) *dependencyTreeNode {
	keys := strings.Split(path, ".")
	node := root
	for _, key := range keys {
		child, ok := node.Children[key]
		if !ok {
			child = newDepTreeNode()
			node.Children[key] = child
		}
		node = child
	}
	return node
}

// markAllParentsByPath mirrors upstream's same-named helper.
func markAllParentsByPath(root *dependencyTreeNode, path string, fn func(*dependencyTreeNode)) {
	keys := strings.Split(path, ".")
	node := root
	for _, key := range keys {
		child, ok := node.Children[key]
		if !ok {
			return
		}
		fn(child)
		node = child
	}
}

// recommendations is the result returned by collectRecommendations.
type recommendations struct {
	Suggested   []string
	Unnecessary map[string]bool
	Duplicate   map[string]bool
	Missing     map[string]bool
}

// collectRecommendations mirrors upstream's same-named helper. It walks the
// dependency tree to compute missing / unnecessary / duplicate sets and a
// suggested deps array preserving the original declaration order.
//
// Missing-dep ordering inside the suggested array uses each dep's first-
// reference source position (matches upstream's Map-insertion order).
func collectRecommendations(
	dependencies map[string]*dependency,
	declaredDependencies []declaredDependency,
	stableDependencies map[string]bool,
	externalDependencies map[string]bool,
	isEffect bool,
) recommendations {
	depTree := newDepTreeNode()
	for key := range dependencies {
		node := getOrCreateNodeByPath(depTree, key)
		node.IsUsed = true
		markAllParentsByPath(depTree, key, func(parent *dependencyTreeNode) {
			parent.IsSubtreeUsed = true
		})
	}
	for _, dd := range declaredDependencies {
		node := getOrCreateNodeByPath(depTree, dd.Key)
		node.IsSatisfiedRecursively = true
	}
	for key := range stableDependencies {
		node := getOrCreateNodeByPath(depTree, key)
		node.IsSatisfiedRecursively = true
	}

	missing := map[string]bool{}
	satisfying := map[string]bool{}
	var scan func(node *dependencyTreeNode, keyToPath func(string) string)
	scan = func(node *dependencyTreeNode, keyToPath func(string) string) {
		// Iterate children in deterministic insertion-style order. Map
		// iteration in Go is non-deterministic, so sort keys.
		keys := make([]string, 0, len(node.Children))
		for k := range node.Children {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			child := node.Children[key]
			path := keyToPath(key)
			if child.IsSatisfiedRecursively {
				if child.IsSubtreeUsed {
					satisfying[path] = true
				}
				continue
			}
			if child.IsUsed {
				missing[path] = true
				continue
			}
			cur := path
			scan(child, func(childKey string) string {
				return cur + "." + childKey
			})
		}
	}
	scan(depTree, func(k string) string { return k })

	suggested := []string{}
	unnecessary := map[string]bool{}
	duplicate := map[string]bool{}
	seenSuggested := map[string]bool{}
	for _, dd := range declaredDependencies {
		if satisfying[dd.Key] {
			if !seenSuggested[dd.Key] {
				seenSuggested[dd.Key] = true
				suggested = append(suggested, dd.Key)
			} else {
				duplicate[dd.Key] = true
			}
		} else {
			if isEffect && !strings.HasSuffix(dd.Key, ".current") && !externalDependencies[dd.Key] {
				if !seenSuggested[dd.Key] {
					seenSuggested[dd.Key] = true
					suggested = append(suggested, dd.Key)
				}
			} else {
				unnecessary[dd.Key] = true
			}
		}
	}
	// Append missing in source-reference order (mirrors upstream's
	// Map-insertion = first-reference order). Falls back to alphabetic
	// when the dependency record carries no `First` position.
	missingKeys := make([]string, 0, len(missing))
	for k := range missing {
		missingKeys = append(missingKeys, k)
	}
	sort.SliceStable(missingKeys, func(i, j int) bool {
		di, dj := dependencies[missingKeys[i]], dependencies[missingKeys[j]]
		var pi, pj int
		if di != nil && di.First != nil {
			pi = di.First.Pos()
		} else {
			pi = -1
		}
		if dj != nil && dj.First != nil {
			pj = dj.First.Pos()
		} else {
			pj = -1
		}
		if pi != pj {
			return pi < pj
		}
		return missingKeys[i] < missingKeys[j]
	})
	suggested = append(suggested, missingKeys...)
	return recommendations{
		Suggested:   suggested,
		Unnecessary: unnecessary,
		Duplicate:   duplicate,
		Missing:     missing,
	}
}

// joinEnglish mirrors upstream's `joinEnglish` ("a", "b", and "c").
func joinEnglish(arr []string) string {
	var sb strings.Builder
	for i, s := range arr {
		sb.WriteString(s)
		if i == 0 && len(arr) == 2 {
			sb.WriteString(" and ")
		} else if i == len(arr)-2 && len(arr) > 2 {
			sb.WriteString(", and ")
		} else if i < len(arr)-1 {
			sb.WriteString(", ")
		}
	}
	return sb.String()
}

// formatDependency mirrors upstream's `formatDependency` — re-inserts `?.`
// for path segments that were always observed under optional access.
func formatDependency(path string, optionalChains map[string]bool) string {
	parts := strings.Split(path, ".")
	var sb strings.Builder
	for i, part := range parts {
		if i != 0 {
			pathSoFar := strings.Join(parts[:i+1], ".")
			if optionalChains[pathSoFar] {
				sb.WriteString("?.")
			} else {
				sb.WriteString(".")
			}
		}
		sb.WriteString(part)
	}
	return sb.String()
}

// getWarningMessage mirrors upstream's same-named helper.
func getWarningMessage(deps map[string]bool, singlePrefix, label, fixVerb string, optionalChains map[string]bool) string {
	if len(deps) == 0 {
		return ""
	}
	keys := make([]string, 0, len(deps))
	for k := range deps {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	formatted := make([]string, len(keys))
	for i, k := range keys {
		formatted[i] = "'" + formatDependency(k, optionalChains) + "'"
	}
	plural := len(keys) > 1
	prefix := ""
	if !plural {
		prefix = singlePrefix + " "
	}
	noun := "dependency"
	if plural {
		noun = "dependencies"
	}
	pronoun := "it"
	if plural {
		pronoun = "them"
	}
	return prefix + label + " " + noun + ": " + joinEnglish(formatted) +
		". Either " + fixVerb + " " + pronoun + " or remove the dependency array."
}

// getCalleeText mirrors upstream's `context.getSourceCode().getText(reactiveHook)`.
//
// tsgo-specific: peel ParenthesizedExpression so messages render
// `useEffect` rather than `(useEffect)` — ESTree never exposes the paren
// wrapper, and upstream's diagnostics use the unwrapped text.
func getCalleeText(sf *ast.SourceFile, callee *ast.Node) string {
	if callee != nil {
		callee = ast.SkipParentheses(callee)
	}
	return nodeText(sf, callee)
}

// areDeclaredDepsAlphabetized mirrors upstream's same-named helper.
func areDeclaredDepsAlphabetized(declared []declaredDependency) bool {
	if len(declared) == 0 {
		return true
	}
	keys := make([]string, len(declared))
	for i, d := range declared {
		keys[i] = d.Key
	}
	sorted := append([]string(nil), keys...)
	sort.Strings(sorted)
	return strings.Join(keys, ",") == strings.Join(sorted, ",")
}

// hasUndefinedIdentifier reports whether `node` is a literal `undefined`
// identifier — used to recognize `useEffect(fn, undefined)` as "no deps".
func hasUndefinedIdentifier(node *ast.Node) bool {
	n := stripAsExpression(node)
	return n != nil && n.Kind == ast.KindIdentifier && n.AsIdentifier().Text == "undefined"
}

// isReferenceIdentifier reports whether the given Identifier appears in a
// value-reference position (as opposed to a property name, label, declaration
// name, etc).
func isReferenceIdentifier(id *ast.Node) bool {
	p := id.Parent
	if p == nil {
		return false
	}
	switch p.Kind {
	case ast.KindPropertyAccessExpression:
		return p.AsPropertyAccessExpression().Expression == id
	case ast.KindElementAccessExpression:
		return p.AsElementAccessExpression().Expression == id ||
			p.AsElementAccessExpression().ArgumentExpression == id
	case ast.KindPropertyAssignment:
		return p.AsPropertyAssignment().Initializer == id
	case ast.KindShorthandPropertyAssignment:
		return p.Name() == id
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return false
	case ast.KindJsxAttribute:
		return false
	case ast.KindVariableDeclaration:
		return p.AsVariableDeclaration().Name() != id
	case ast.KindBindingElement:
		return p.AsBindingElement().Name() != id
	case ast.KindParameter:
		return p.AsParameterDeclaration().Name() != id
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
		ast.KindClassDeclaration, ast.KindClassExpression:
		return p.Name() != id
	case ast.KindLabeledStatement:
		return p.AsLabeledStatement().Label != id
	case ast.KindBreakStatement, ast.KindContinueStatement:
		return false
	case ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport,
		ast.KindExportSpecifier:
		return false
	case ast.KindTypeReference, ast.KindTypeQuery:
		return false
	}
	return true
}

// isInsideTypePosition reports whether `node` is inside a TypeReference /
// TypeQuery / type-only construct. Mirrors upstream's
// `dependencyNode.parent?.type === 'TSTypeQuery' || 'TSTypeReference'`.
func isInsideTypePosition(node *ast.Node) bool {
	cur := node
	for cur != nil {
		switch cur.Kind {
		case ast.KindTypeReference, ast.KindTypeQuery, ast.KindTypePredicate,
			ast.KindTypeLiteral, ast.KindTypeOperator, ast.KindIndexedAccessType,
			ast.KindMappedType, ast.KindConditionalType, ast.KindInferType,
			ast.KindUnionType, ast.KindIntersectionType, ast.KindTupleType,
			ast.KindArrayType, ast.KindLiteralType, ast.KindFunctionType,
			ast.KindConstructorType, ast.KindParenthesizedType:
			return true
		}
		// Stop walking when leaving the expression to avoid mistakenly
		// considering an expression inside a function body as "in a type".
		if isFunctionLikeContainer(cur) {
			return false
		}
		cur = cur.Parent
	}
	return false
}

// isInsideEffectCleanup reports whether `idNode` lies inside a function
// returned from the effect callback (i.e. effect cleanup). Mirrors upstream's
// `isInsideEffectCleanup`.
//
// `callback` is the effect callback function-like (ArrowFunction /
// FunctionExpression). We walk up from `idNode` looking for an inner
// function-like whose immediate parent is a ReturnStatement that itself
// belongs to the effect callback.
func isInsideEffectCleanup(idNode *ast.Node, callback *ast.Node) bool {
	cur := idNode
	for cur != nil {
		if cur == callback {
			return false
		}
		if isFunctionLikeContainer(cur) && cur != callback && cur.Parent != nil {
			// `return () => {}` — parent is ReturnStatement.
			if cur.Parent.Kind == ast.KindReturnStatement {
				retEnclosing := findEnclosingFunction(cur.Parent)
				if retEnclosing != nil && retEnclosing == callback {
					return true
				}
			}
		}
		cur = cur.Parent
	}
	return false
}

// isStableHookValue inspects a VariableDeclaration whose initializer is a
// known hook call (or const literal) and reports whether the binding at
// `bindingId` is one of React's stable identities. Mirrors the union of
// upstream's `isStableKnownHookValue` cases.
//
// `bindingId` is the Identifier of the specific binding being inspected
// — for `const [a, b] = useState()`, this is either `a` or `b`.
//
// The second return value, `isUseEffectEvent`, signals that the binding
// is a useEffectEvent return — used both as "stable" and as the gate for
// emitting the "Functions returned from useEffectEvent must not be included"
// diagnostic on a deps-array entry.
func isStableHookValue(decl *ast.Node, bindingId *ast.Node) (stable bool, isUseEffectEvent bool) {
	if decl == nil || decl.Kind != ast.KindVariableDeclaration {
		return false, false
	}
	vd := decl.AsVariableDeclaration()
	init := vd.Initializer
	if init == nil {
		return false, false
	}
	init = stripAsExpression(init)
	// `const foo = 42` / `'str'` / `null` — primitive const is stable.
	parentDeclList := decl.Parent
	if parentDeclList != nil && parentDeclList.Kind == ast.KindVariableDeclarationList {
		dl := parentDeclList.AsVariableDeclarationList()
		// `const` keyword check. NodeFlagsConst / NodeFlagsLet apply to
		// the declaration list.
		if dl.Flags&ast.NodeFlagsConst != 0 {
			switch init.Kind {
			case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindNullKeyword:
				return true, false
			case ast.KindNoSubstitutionTemplateLiteral:
				return true, false
			}
		}
	}
	if init.Kind != ast.KindCallExpression {
		return false, false
	}
	callee := stripReactNamespace(init.AsCallExpression().Expression)
	if callee == nil || callee.Kind != ast.KindIdentifier {
		return false, false
	}
	calleeName := callee.AsIdentifier().Text

	// The binding identifier site dictates which positions are stable.
	bindingName := vd.Name()
	switch calleeName {
	case "useRef":
		// Only the binding name itself is stable; destructured forms aren't.
		if bindingName != nil && bindingName == bindingId {
			return true, false
		}
	case "useEffectEvent":
		if bindingName != nil && bindingName == bindingId {
			return true, true
		}
	case "useState", "useReducer", "useActionState":
		if bindingName != nil && bindingName.Kind == ast.KindArrayBindingPattern {
			arr := bindingName.AsBindingPattern()
			if arr.Elements != nil && len(arr.Elements.Nodes) == 2 {
				first := arr.Elements.Nodes[0]
				second := arr.Elements.Nodes[1]
				// `setX` / `dispatch` is the stable side.
				if isMatchingBindingElementId(second, bindingId) {
					return true, false
				}
				// state itself (`x`) is dynamic.
				_ = first
			}
		}
	case "useTransition":
		if bindingName != nil && bindingName.Kind == ast.KindArrayBindingPattern {
			arr := bindingName.AsBindingPattern()
			if arr.Elements != nil && len(arr.Elements.Nodes) == 2 {
				second := arr.Elements.Nodes[1]
				if isMatchingBindingElementId(second, bindingId) {
					return true, false
				}
			}
		}
	}
	return false, false
}

// isMatchingBindingElementId reports whether `binding` is a BindingElement
// (or OmittedExpression) whose name Identifier equals `id`. Uses Pos/End
// to compare positions instead of pointer equality, which is fragile when
// the same Node is reachable through multiple paths.
func isMatchingBindingElementId(binding *ast.Node, id *ast.Node) bool {
	if binding == nil || binding.Kind == ast.KindOmittedExpression {
		return false
	}
	if binding.Kind != ast.KindBindingElement {
		return false
	}
	be := binding.AsBindingElement()
	name := be.Name()
	if name == nil || name.Kind != ast.KindIdentifier || id == nil || id.Kind != ast.KindIdentifier {
		return false
	}
	if name == id {
		return true
	}
	return name.Pos() == id.Pos() && name.End() == id.End() &&
		name.AsIdentifier().Text == id.AsIdentifier().Text
}

// isElidedComma checks for `[, foo]`-style elided elements. tsgo represents
// them as KindOmittedExpression rather than ESTree's null.
func isElidedComma(n *ast.Node) bool {
	return n != nil && n.Kind == ast.KindOmittedExpression
}

// constructionType returns a human-readable description for a node that
// would yield a fresh referential identity on every render. Mirrors
// upstream's `getConstructionExpressionType`.
func constructionType(node *ast.Node) string {
	if node == nil {
		return ""
	}
	n := stripAsExpression(node)
	switch n.Kind {
	case ast.KindObjectLiteralExpression:
		return "object"
	case ast.KindArrayLiteralExpression:
		return "array"
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		return "function"
	case ast.KindClassExpression:
		return "class"
	case ast.KindConditionalExpression:
		ce := n.AsConditionalExpression()
		if constructionType(ce.WhenTrue) != "" || constructionType(ce.WhenFalse) != "" {
			return "conditional"
		}
		return ""
	case ast.KindBinaryExpression:
		be := n.AsBinaryExpression()
		if be.OperatorToken != nil {
			switch be.OperatorToken.Kind {
			case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
				if constructionType(be.Left) != "" || constructionType(be.Right) != "" {
					return "logical expression"
				}
				return ""
			case ast.KindEqualsToken,
				ast.KindPlusEqualsToken, ast.KindMinusEqualsToken,
				ast.KindAsteriskEqualsToken, ast.KindSlashEqualsToken,
				ast.KindPercentEqualsToken,
				ast.KindAmpersandAmpersandEqualsToken, ast.KindBarBarEqualsToken,
				ast.KindQuestionQuestionEqualsToken:
				if constructionType(be.Right) != "" {
					return "assignment expression"
				}
				return ""
			}
		}
		return ""
	case ast.KindJsxFragment:
		return "JSX fragment"
	case ast.KindJsxElement, ast.KindJsxSelfClosingElement:
		return "JSX element"
	case ast.KindNewExpression:
		return "object construction"
	case ast.KindRegularExpressionLiteral:
		return "regular expression"
	}
	return ""
}

// classifiedConstruction is one item from `scanForConstructions`.
type classifiedConstruction struct {
	Variable          *ast.Node // the declaration name Identifier
	Decl              *ast.Node // the VariableDeclaration / FunctionDeclaration / ClassDeclaration
	InitNode          *ast.Node // VariableDeclaration's initializer (nil for fn/class decls)
	DepType           string
	IsUsedOutsideHook bool
}

// scanForConstructions mirrors upstream's same-named helper.
//
// For each declared dependency name, find its declaration in the component
// scope and report it as a construction iff the declaration is one of
// `function foo() {}` / `class Foo {}` / `const foo = <something construction-shaped>`.
func scanForConstructions(
	declared []declaredDependency,
	declaredDepsNode *ast.Node,
	componentBody *ast.Node,
	hookCallback *ast.Node,
	tc *checker.Checker,
	sf *ast.SourceFile,
) []classifiedConstruction {
	if componentBody == nil {
		return nil
	}
	out := []classifiedConstruction{}
	for _, dd := range declared {
		// Only care about plain identifier deps (`foo`, not `foo.bar`).
		if strings.Contains(dd.Key, ".") {
			continue
		}
		if dd.Node == nil || dd.Node.Kind != ast.KindIdentifier {
			continue
		}
		// Resolve dd.Node via TypeChecker; fall back to a name walk in
		// the component body.
		decl := resolveDeclaration(tc, dd.Node, dd.Key, componentBody)
		if decl == nil {
			continue
		}
		switch decl.Kind {
		case ast.KindVariableDeclaration:
			vd := decl.AsVariableDeclaration()
			name := vd.Name()
			if name == nil || name.Kind != ast.KindIdentifier {
				continue
			}
			if vd.Initializer == nil {
				continue
			}
			ct := constructionType(vd.Initializer)
			if ct == "" {
				continue
			}
			out = append(out, classifiedConstruction{
				Variable:          name,
				Decl:              decl,
				InitNode:          vd.Initializer,
				DepType:           ct,
				IsUsedOutsideHook: isUsedOutsideHook(name, hookCallback, declaredDepsNode, componentBody),
			})
		case ast.KindFunctionDeclaration:
			out = append(out, classifiedConstruction{
				Variable:          decl.Name(),
				Decl:              decl,
				DepType:           "function",
				IsUsedOutsideHook: isUsedOutsideHook(decl.Name(), hookCallback, declaredDepsNode, componentBody),
			})
		case ast.KindClassDeclaration:
			out = append(out, classifiedConstruction{
				Variable:          decl.Name(),
				Decl:              decl,
				DepType:           "class",
				IsUsedOutsideHook: isUsedOutsideHook(decl.Name(), hookCallback, declaredDepsNode, componentBody),
			})
		}
	}
	return out
}

// resolveDeclaration finds the declaration corresponding to `id`. Prefers
// TypeChecker symbol resolution; falls back to a name-based walk over
// `body`. The fallback walk descends into ObjectBindingPattern /
// ArrayBindingPattern so destructured bindings (including renamed
// `{a: b}` form and nested patterns) and parameter destructure are
// discoverable; it also recognizes BindingElement / Parameter as
// terminal "declaration" nodes when the name matches.
func resolveDeclaration(tc *checker.Checker, id *ast.Node, name string, body *ast.Node) *ast.Node {
	if tc != nil {
		sym := tc.GetSymbolAtLocation(id)
		if sym != nil && len(sym.Declarations) > 0 {
			return sym.Declarations[0]
		}
	}
	if body == nil {
		return nil
	}
	var found *ast.Node
	var visit func(n *ast.Node) bool
	visit = func(n *ast.Node) bool {
		if found != nil {
			return true
		}
		switch n.Kind {
		case ast.KindVariableDeclaration:
			vd := n.AsVariableDeclaration()
			vname := vd.Name()
			if vname != nil {
				if vname.Kind == ast.KindIdentifier && vname.AsIdentifier().Text == name {
					found = n
					return true
				}
				// Destructure pattern — descend into the pattern looking
				// for BindingElement whose binding name matches.
			}
		case ast.KindBindingElement:
			be := n.AsBindingElement()
			bn := be.Name()
			if bn != nil && bn.Kind == ast.KindIdentifier && bn.AsIdentifier().Text == name {
				found = n
				return true
			}
		case ast.KindParameter:
			pd := n.AsParameterDeclaration()
			pn := pd.Name()
			if pn != nil && pn.Kind == ast.KindIdentifier && pn.AsIdentifier().Text == name {
				found = n
				return true
			}
		case ast.KindFunctionDeclaration:
			if fname := n.Name(); fname != nil && fname.Kind == ast.KindIdentifier && fname.AsIdentifier().Text == name {
				found = n
				return true
			}
		case ast.KindClassDeclaration:
			if cname := n.Name(); cname != nil && cname.Kind == ast.KindIdentifier && cname.AsIdentifier().Text == name {
				found = n
				return true
			}
		}
		if isFunctionLikeContainer(n) && n != body {
			return false
		}
		n.ForEachChild(visit)
		return false
	}
	visit(body)
	return found
}

// isUsedOutsideHook reports whether `name` (a declaration's Identifier) has
// any reference outside of the hook callback that is not inside the deps
// array. Mirrors upstream's `isUsedOutsideOfHook`.
//
// `scope` bounds the walk. Pass `componentFn` whenever possible — the
// rule only cares about uses inside the surrounding component or hook;
// fall back to `sf.AsNode()` only when no component scope is available.
// Limiting the walk avoids the O(file_size) per-construction cost
// flagged in PR-808 review.
func isUsedOutsideHook(name *ast.Node, hookCallback *ast.Node, depsNode *ast.Node, scope *ast.Node) bool {
	if name == nil || name.Kind != ast.KindIdentifier || scope == nil {
		return false
	}
	target := name.AsIdentifier().Text
	used := false
	var visit func(n *ast.Node) bool
	visit = func(n *ast.Node) bool {
		if used {
			return true
		}
		if n == hookCallback || n == depsNode {
			// Skip the hook callback and the deps array entirely —
			// references inside them aren't "outside the hook".
			return false
		}
		if n.Kind == ast.KindIdentifier && n != name {
			if n.AsIdentifier().Text == target && isReferenceIdentifier(n) {
				used = true
				return true
			}
		}
		n.ForEachChild(visit)
		return false
	}
	visit(scope)
	return used
}

// isComponentOrHookFunction delegates to the shared classifier.
func isComponentOrHookFunction(fn *ast.Node) bool {
	return react_hooksutil.IsComponentOrHookFn(fn)
}

// getUnknownDependenciesMessage mirrors upstream's same-named helper.
func getUnknownDependenciesMessage(reactiveHookName string) string {
	return fmt.Sprintf(
		"React Hook %s received a function whose dependencies are unknown. Pass an inline function instead.",
		reactiveHookName,
	)
}

// reactiveHookName returns just the hook's bare name (after stripping
// `React.`). Used for `additionalHooks` matching and message rendering.
func reactiveHookName(callee *ast.Node) string {
	n := react_hooksutil.StripReactNamespace(callee)
	if n != nil && n.Kind == ast.KindIdentifier {
		return n.AsIdentifier().Text
	}
	return ""
}
