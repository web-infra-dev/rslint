// Package exhaustive_deps checks reactive Hook dependency arrays using rslint's shared lexical scopes.
package exhaustive_deps

import (
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/collections"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"github.com/web-infra-dev/rslint/internal/utils/scope"

	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/react_hooksutil"
	"github.com/web-infra-dev/rslint/internal/utils"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

// getReactiveHookCallbackIndex matches the built-in hooks and bare custom names.
func getReactiveHookCallbackIndex(callee *ast.Node, additionalHooks *esregexp.RegExp) int {
	n := react_hooksutil.StripReactNamespace(callee)
	if n == nil || n.Kind != ast.KindIdentifier {
		// Only bare names and built-in React.<name> calls are recognized.
		return -1
	}
	name := n.AsIdentifier().Text
	switch name {
	case "useEffect", "useLayoutEffect", "useCallback", "useMemo":
		return 0
	case "useImperativeHandle":
		return 1
	}
	// `additionalHooks` only applies to bare-identifier callees, mirroring
	// upstream's `node === calleeNode` gate. `React.useCustomEffect` is
	// intentionally NOT treated as a reactive hook by the additionalHooks
	// path — only the unqualified `useCustomEffect` is.
	if additionalHooks != nil && n == utils.ESTreeRuntimeExpression(callee) && additionalHooks.Test(name) {
		return 0
	}
	return -1
}

// stripAsExpression unwraps only the TypeScript `as` casts supported upstream.
func stripAsExpression(node *ast.Node) *ast.Node {
	for node != nil {
		node = utils.ESTreeRuntimeExpression(node)
		if node.Kind != ast.KindAsExpression {
			return node
		}
		node = node.AsAsExpression().Expression
	}
	return nil
}

func containsNode(ancestor, descendant *ast.Node) bool {
	return react_hooksutil.ContainsNode(ancestor, descendant)
}

// nodeText returns the trimmed source text for `node`.
func nodeText(sf *ast.SourceFile, node *ast.Node) string {
	if node == nil {
		return ""
	}
	return utils.TrimmedNodeText(sf, node)
}

func analyzePropertyChainText(node *ast.Node, optionalChains map[string]bool) (string, bool) {
	return analyzePropertyChain(node, optionalChains, false)
}

func analyzeDepsArrayElement(node *ast.Node, optionalChains map[string]bool) (string, bool) {
	return analyzePropertyChain(node, optionalChains, true)
}

func analyzePropertyChain(node *ast.Node, optionalChains map[string]bool, strict bool) (string, bool) {
	if node == nil {
		return "", false
	}
	n := utils.ESTreeRuntimeExpression(node)
	if strict {
		// Upstream throws on these node kinds -> "complex expression".
		switch n.Kind {
		case ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression:
			return "", false
		}
	} else {
		// Callback path: `as` / `satisfies` are transparent.
		for {
			switch n.Kind {
			case ast.KindAsExpression:
				n = utils.ESTreeRuntimeExpression(n.AsAsExpression().Expression)
				continue
			case ast.KindSatisfiesExpression:
				n = utils.ESTreeRuntimeExpression(n.AsSatisfiesExpression().Expression)
				continue
			}
			break
		}
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
		object, ok := analyzePropertyChain(pae.Expression, optionalChains, strict)
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
	case ast.KindElementAccessExpression:
		// Upstream accepts computed access only through the enclosing
		// ESTree ChainExpression, never as an interior chain link. Unlike
		// tsgo's IsOutermostOptionalChain, another ?. does not end that wrapper.
		if !strict || !ast.IsOptionalChain(n) || n.Parent != nil && ast.IsOptionalChain(n.Parent) && n.Parent.Expression() == n {
			return "", false
		}
		eae := n.AsElementAccessExpression()
		object, ok := analyzePropertyChain(eae.Expression, optionalChains, strict)
		if !ok {
			return "", false
		}
		property, ok := analyzePropertyChain(eae.ArgumentExpression, nil, strict)
		if !ok {
			return "", false
		}
		result := object + "." + property
		if optionalChains != nil {
			markOptionalChain(optionalChains, result, eae.QuestionDotToken != nil)
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

// getDependencyNode stops at method receivers, mutable current, and assertions.
func getDependencyNode(node *ast.Node) *ast.Node {
	cur := utils.ESTreeRuntimeExpression(node)
	for cur != nil {
		p := utils.ESTreeParent(cur)
		if p == nil || p.Kind != ast.KindPropertyAccessExpression {
			break
		}
		member := p.AsPropertyAccessExpression()
		if utils.ESTreeCallCallee(member.Expression) != cur || member.Name().Kind != ast.KindIdentifier || member.Name().Text() == "current" {
			break
		}
		caller := utils.ESTreeParent(p)
		if caller != nil && caller.Kind == ast.KindCallExpression && utils.ESTreeCallCallee(caller.AsCallExpression().Expression) == p {
			break
		}
		cur = p
	}
	if cur != nil && cur.Kind == ast.KindPropertyAccessExpression {
		if assignment, ok := getAssignmentBinaryExpr(utils.ESTreeParent(cur)); ok && utils.ESTreeRuntimeExpression(assignment.Left) == cur {
			return utils.ESTreeRuntimeExpression(cur.AsPropertyAccessExpression().Expression)
		}
	}
	return cur
}

func getAssignmentBinaryExpr(node *ast.Node) (*ast.BinaryExpression, bool) {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return nil, false
	}
	expression := node.AsBinaryExpression()
	return expression, expression.OperatorToken != nil && ast.IsAssignmentOperator(expression.OperatorToken.Kind)
}

// Options holds the parsed rule options.
type Options struct {
	AdditionalHooks                                 *esregexp.RegExp
	EnableDangerousAutofixThisMayCauseInfiniteLoops bool
	RequireExplicitEffectDeps                       bool
	// AutoDepsHooks: experimental_autoDependenciesHooks. When the hook's
	// bare name is in this list, missing deps are inferred ("auto deps")
	// rather than being flagged. Mirrors upstream's same-named option.
	AutoDepsHooks map[string]bool
}

// parseOptions parses the rule's options object.
func parseOptions(options []any, settings map[string]interface{}) Options {
	opts := Options{
		AdditionalHooks: react_hooksutil.AdditionalHooksFromSettings(settings, "additionalEffectHooks"),
	}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]interface{})
	opts.EnableDangerousAutofixThisMayCauseInfiniteLoops, _ = optsMap["enableDangerousAutofixThisMayCauseInfiniteLoops"].(bool)
	opts.RequireExplicitEffectDeps, _ = optsMap["requireExplicitEffectDeps"].(bool)
	if hooks := utils.ToStringSlice(optsMap["experimental_autoDependenciesHooks"]); len(hooks) > 0 {
		opts.AutoDepsHooks = make(map[string]bool, len(hooks))
		for _, h := range hooks {
			opts.AutoDepsHooks[h] = true
		}
	}
	// Mirrors upstream's `rawOptions.additionalHooks` truthiness check: a
	// non-empty rule-level pattern replaces the settings fallback even when
	// it fails to compile; absent or empty keeps the settings-derived value.
	if raw, _ := optsMap["additionalHooks"].(string); raw != "" {
		opts.AdditionalHooks = nil
		if re, err := esregexp.Compile(raw, ""); err == nil {
			opts.AdditionalHooks = re
		}
	}
	return opts
}

// declaredDependency is the parsed form of an entry in the deps array.
type declaredDependency struct {
	Key  string
	Node *ast.Node
}

// dependency is a used reference observed inside the callback body.
type dependencyMap = collections.OrderedMap[string, *dependency]

type dependency struct {
	IsStable bool
	Refs     []depReference
}

type depReference struct {
	*scope.Reference
	WriteExpr   *ast.Node
	DepNodeRoot *ast.Node
}

// dependencyTreeNode mirrors upstream's `DependencyTreeNode`.
type dependencyTreeNode struct {
	IsUsed                 bool
	IsSatisfiedRecursively bool
	IsSubtreeUsed          bool
	Children               collections.OrderedMap[string, *dependencyTreeNode]
}

func newDepTreeNode() *dependencyTreeNode {
	return &dependencyTreeNode{}
}

// getOrCreateNodeByPath mirrors upstream's same-named helper.
func getOrCreateNodeByPath(root *dependencyTreeNode, path string) *dependencyTreeNode {
	node := root
	for key := range strings.SplitSeq(path, ".") {
		child, ok := node.Children.Get(key)
		if !ok {
			child = newDepTreeNode()
			node.Children.Set(key, child)
		}
		node = child
	}
	return node
}

// markAllParentsByPath mirrors upstream's same-named helper.
func markAllParentsByPath(root *dependencyTreeNode, path string, fn func(*dependencyTreeNode)) {
	node := root
	for key := range strings.SplitSeq(path, ".") {
		child, ok := node.Children.Get(key)
		if !ok {
			return
		}
		fn(child)
		node = child
	}
}

// recommendations is the result returned by collectRecommendations.
type recommendations struct {
	Suggested    []string
	Unnecessary  map[string]bool
	Duplicate    map[string]bool
	Missing      map[string]bool
	MissingOrder []string
}

// collectRecommendations mirrors upstream's same-named helper. It walks the
// dependency tree to compute missing / unnecessary / duplicate sets and a
// suggested deps array preserving the original declaration order.
//
// Insertion-ordered maps preserve upstream's scope traversal and tree order.
func collectRecommendations(
	dependencies *dependencyMap,
	declaredDependencies []declaredDependency,
	stableDependencies map[string]bool,
	externalDependencies map[string]bool,
	isEffect bool,
) recommendations {
	depTree := newDepTreeNode()
	for key := range dependencies.Keys() {
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
	missingOrder := []string{}
	satisfying := map[string]bool{}
	var scan func(node *dependencyTreeNode, prefix string)
	scan = func(node *dependencyTreeNode, prefix string) {
		for key, child := range node.Children.Entries() {
			path := prefix + key
			if child.IsSatisfiedRecursively {
				if child.IsSubtreeUsed {
					satisfying[path] = true
				}
				continue
			}
			if child.IsUsed {
				missing[path] = true
				missingOrder = append(missingOrder, path)
				continue
			}
			scan(child, path+".")
		}
	}
	scan(depTree, "")

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
	suggested = append(suggested, missingOrder...)
	return recommendations{
		Suggested:    suggested,
		Unnecessary:  unnecessary,
		Duplicate:    duplicate,
		Missing:      missing,
		MissingOrder: missingOrder,
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
	slices.SortFunc(keys, ecmascript.CompareStrings)
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
		callee = utils.ESTreeRuntimeExpression(callee)
	}
	return nodeText(sf, callee)
}

// areDeclaredDepsAlphabetized mirrors upstream's same-named helper.
func areDeclaredDepsAlphabetized(declared []declaredDependency) bool {
	if len(declared) == 0 {
		return true
	}
	return slices.IsSortedFunc(declared, func(a, b declaredDependency) int {
		return ecmascript.CompareStrings(a.Key, b.Key)
	})
}

// hasUndefinedIdentifier reports whether `node` is a literal `undefined`
// identifier — used to recognize `useEffect(fn, undefined)` as "no deps".
func hasUndefinedIdentifier(node *ast.Node) bool {
	n := utils.ESTreeRuntimeExpression(node)
	return n != nil && n.Kind == ast.KindIdentifier && n.AsIdentifier().Text == "undefined"
}

func isInsideTypePosition(node *ast.Node) bool {
	parent := utils.ESTreeParent(node)
	return parent != nil && (parent.Kind == ast.KindTypeReference || parent.Kind == ast.KindTypeQuery)
}

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
			if ast.IsAssignmentOperator(be.OperatorToken.Kind) {
				if constructionType(be.Right) != "" {
					return "assignment expression"
				}
				return ""
			}
			switch be.OperatorToken.Kind {
			case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
				if constructionType(be.Left) != "" || constructionType(be.Right) != "" {
					return "logical expression"
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
