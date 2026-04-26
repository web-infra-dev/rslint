package exhaustive_deps

import (
	"fmt"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// runCaches holds per-Run (per-file lint pass) memoization tables.
//
// Several helpers in this rule (`hasExtraWriteToSymbol`,
// `hasCurrentAssignment`, `isUsedOutsideHook`) traverse a chunk of the
// AST to answer a question about a particular symbol or identifier.
// Without caching, those traversals re-run for every hook listener in
// the file — easily O(N×hooks×file_size) on large codebases. Caches
// here key on the stable identity (symbol pointer for the first two,
// declaration-Identifier node for the last) so each query pays the
// walk cost at most once per file.
//
// Lifetime: created once at the top of `Run` and kept alive for the
// duration of that listener registration. The runtime calls listeners
// many times for the same file but always within the same Run scope,
// so map identity is preserved.
type runCaches struct {
	extraWrite    map[*ast.Symbol]bool
	currentAssign map[*ast.Symbol]bool
	usedOutside   map[*ast.Node]bool
}

// ExhaustiveDepsRule is the rslint port of upstream `react-hooks/exhaustive-deps`.
//
// Architectural difference from upstream: instead of relying on ESLint's
// scope manager to enumerate `scope.references`, we walk the callback body
// and resolve each Identifier reference via the TypeChecker. The
// "pure scope" check (does this identifier resolve to a binding declared
// outside the callback but inside the component?) is encoded as
// `containsNode(componentBody, declSite) && !containsNode(callback, declSite)`.
// When the TypeChecker is unavailable, the same check degrades to a
// best-effort name walk over the component body.
//
// Diagnostics intentionally mirror upstream wording for switchover parity.
var ExhaustiveDepsRule = rule.Rule{
	Name: "react-hooks/exhaustive-deps",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options, ctx.Settings)
		sf := ctx.SourceFile
		tc := ctx.TypeChecker

		// Run-scoped caches shared across every hook listener invocation
		// in this file. Each cache is keyed by a stable identity (symbol
		// pointer or AST node) and computed by walking the relevant
		// component scope at most once per key. This avoids the
		// previous O(N×hooks×file_size) re-scan pattern that surfaced
		// in PR-808 review feedback for large codebases.
		caches := &runCaches{
			extraWrite:     map[*ast.Symbol]bool{},
			currentAssign:  map[*ast.Symbol]bool{},
			usedOutside:    map[*ast.Node]bool{},
		}

		report := func(node *ast.Node, msg string) {
			ctx.ReportNode(node, rule.RuleMessage{Description: msg})
		}
		reportWithSuggestions := func(node *ast.Node, msg string, sugs ...rule.RuleSuggestion) {
			if opts.EnableDangerousAutofixThisMayCauseInfiniteLoops && len(sugs) > 0 && len(sugs[0].FixesArr) > 0 {
				// Mirrors upstream's
				// `problem.fix = problem.suggest[0].fix`: promote the
				// first suggestion's first fix only (singular `.fix`,
				// not the entire FixesArr), AND keep the suggestions.
				// `ReportNodeWithFixesAndSuggestions` is the preferred
				// path; harnesses without it fall back to a fix-only
				// report.
				firstFix := []rule.RuleFix{sugs[0].FixesArr[0]}
				if ctx.ReportNodeWithFixesAndSuggestions != nil {
					ctx.ReportNodeWithFixesAndSuggestions(node, rule.RuleMessage{Description: msg}, firstFix, sugs)
					return
				}
				ctx.ReportNodeWithFixes(node, rule.RuleMessage{Description: msg}, firstFix...)
				return
			}
			if len(sugs) == 0 {
				ctx.ReportNode(node, rule.RuleMessage{Description: msg})
				return
			}
			ctx.ReportNodeWithSuggestions(node, rule.RuleMessage{Description: msg}, sugs...)
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				visitCall(ctx, sf, tc, opts, caches, node, report, reportWithSuggestions)
			},
		}
	},
}

// visitCall is the entry point for each CallExpression. It mirrors
// upstream's `visitCallExpression` body — argument resolution, special
// cases for missing callback / missing deps, and dispatch to
// `visitFunctionWithDependencies` when the callback is inline.
func visitCall(
	ctx rule.RuleContext,
	sf *ast.SourceFile,
	tc *checker.Checker,
	opts Options,
	caches *runCaches,
	node *ast.Node,
	report func(*ast.Node, string),
	reportWithSuggestions func(*ast.Node, string, ...rule.RuleSuggestion),
) {
	call := node.AsCallExpression()
	callee := call.Expression
	cbIdx := getReactiveHookCallbackIndex(callee, opts.AdditionalHooks)
	if cbIdx == -1 {
		return
	}
	args := []*ast.Node{}
	if call.Arguments != nil {
		args = call.Arguments.Nodes
	}
	if cbIdx >= len(args) {
		// The hook needs a callback but none is provided.
		hookName := reactiveHookName(callee)
		report(callee, fmt.Sprintf(
			"React Hook %s requires an effect callback. Did you forget to pass a callback to the hook?",
			hookName,
		))
		return
	}
	callback := args[cbIdx]
	hookName := reactiveHookName(callee)
	isEffect := effectNameRegex.MatchString(hookName)
	isAutoDepsHook := opts.AutoDepsHooks[hookName]

	var depsNode *ast.Node
	if cbIdx+1 < len(args) {
		maybe := args[cbIdx+1]
		if !hasUndefinedIdentifier(maybe) {
			depsNode = maybe
		}
	}

	// experimental_autoDependenciesHooks: when configured for this hook,
	// `null` literal as the deps argument means "auto-infer", and a
	// missing argument also means "auto" — either way we skip analysis.
	if isAutoDepsHook {
		if depsNode == nil {
			return
		}
		if stripped := stripAsExpression(depsNode); stripped != nil && stripped.Kind == ast.KindNullKeyword {
			return
		}
	}

	// Missing deps array branch.
	if depsNode == nil {
		if !isEffect {
			// Memo / Callback with no deps array: report.
			if hookName == "useMemo" || hookName == "useCallback" {
				report(callee, fmt.Sprintf(
					"React Hook %s does nothing when called with only one argument. Did you forget to pass an array of dependencies?",
					hookName,
				))
			}
			return
		}
		// Effect with no deps array: still analyze for setState-without-deps
		// infinite-loop warning. Handled inside visitFunctionWithDependencies
		// when depsNode is nil.
		if opts.RequireExplicitEffectDeps {
			report(callee, fmt.Sprintf(
				"React Hook %s always requires dependencies. Please add a dependency array or an explicit `undefined`",
				hookName,
			))
		}
	}

	// Resolve callback to a function-like.
	cb := stripAsExpression(callback)
	switch cb.Kind {
	case ast.KindFunctionExpression, ast.KindArrowFunction:
		visitFunctionWithDependencies(ctx, sf, tc, opts, caches, cb, depsNode, callee, hookName, isEffect, node, reportWithSuggestions)
	case ast.KindIdentifier:
		// useEffect(myFn, deps): try to follow the symbol.
		// Without a deps array we can't fault anything (effect always runs).
		if depsNode == nil {
			return
		}
		// If the deps array literally contains this identifier, accept.
		if depArrayContainsName(depsNode, cb.AsIdentifier().Text) {
			return
		}
		// Resolve via TypeChecker to classify the binding kind. A Parameter
		// (or any binding whose declaration is not a reachable function /
		// variable-with-function-init) yields the upstream
		// `getUnknownDependenciesMessage` diagnostic instead of the
		// "missing dep <callback-name>" autofix branch.
		if tc != nil {
			sym := tc.GetSymbolAtLocation(cb)
			if sym != nil && len(sym.Declarations) > 0 {
				decl := sym.Declarations[0]
				// Walk up: a BindingElement nested in a Parameter (a
				// destructured prop, e.g. `function C({myEffect})`) should
				// be treated as Parameter for the unknown-deps decision.
				isParameter := decl.Kind == ast.KindParameter
				if !isParameter && decl.Kind == ast.KindBindingElement {
					for cur := decl.Parent; cur != nil; cur = cur.Parent {
						if cur.Kind == ast.KindParameter {
							isParameter = true
							break
						}
						if isFunctionLikeContainer(cur) {
							break
						}
					}
				}
				if isParameter {
					report(callee, getUnknownDependenciesMessage(hookName))
					return
				}
				switch decl.Kind {
				case ast.KindFunctionDeclaration:
					visitFunctionWithDependencies(ctx, sf, tc, opts, caches, decl, depsNode, callee, hookName, isEffect, node, reportWithSuggestions)
					return
				case ast.KindVariableDeclaration:
					vd := decl.AsVariableDeclaration()
					if vd.Initializer != nil {
						init := stripAsExpression(vd.Initializer)
						if init != nil && (init.Kind == ast.KindArrowFunction || init.Kind == ast.KindFunctionExpression) {
							visitFunctionWithDependencies(ctx, sf, tc, opts, caches, init, depsNode, callee, hookName, isEffect, node, reportWithSuggestions)
							return
						}
					}
				}
			}
		}
		// Fallback: suggest adding the identifier to the deps array.
		fix := rule.RuleFixReplace(sf, depsNode, "["+cb.AsIdentifier().Text+"]")
		reportWithSuggestions(callee,
			fmt.Sprintf(
				"React Hook %s has a missing dependency: '%s'. Either include it or remove the dependency array.",
				hookName, cb.AsIdentifier().Text,
			),
			rule.RuleSuggestion{
				Message:  rule.RuleMessage{Description: "Update the dependencies array to be: [" + cb.AsIdentifier().Text + "]"},
				FixesArr: []rule.RuleFix{fix},
			},
		)
	default:
		// Any other expression at the callback position (CallExpression,
		// member call, etc.). Mirrors upstream's catch-all `getUnknownDependenciesMessage`.
		report(callee, getUnknownDependenciesMessage(hookName))
	}
}

// depArrayContainsName reports whether `depsNode` is an array literal
// containing the identifier `name` as one of its bare identifier elements.
func depArrayContainsName(depsNode *ast.Node, name string) bool {
	depsNode = stripAsExpression(depsNode)
	if depsNode == nil || depsNode.Kind != ast.KindArrayLiteralExpression {
		return false
	}
	arr := depsNode.AsArrayLiteralExpression()
	if arr.Elements == nil {
		return false
	}
	for _, el := range arr.Elements.Nodes {
		if el == nil {
			continue
		}
		e := ast.SkipParentheses(el)
		if e.Kind == ast.KindIdentifier && e.AsIdentifier().Text == name {
			return true
		}
	}
	return false
}

// visitFunctionWithDependencies mirrors upstream's same-named function:
// it walks the callback body, gathers used dependencies, compares them
// to declared deps, and reports diagnostics.
func visitFunctionWithDependencies(
	ctx rule.RuleContext,
	sf *ast.SourceFile,
	tc *checker.Checker,
	opts Options,
	caches *runCaches,
	callback *ast.Node,
	depsNode *ast.Node,
	reactiveHook *ast.Node,
	hookName string,
	isEffect bool,
	hookCall *ast.Node,
	reportWithSuggestions func(*ast.Node, string, ...rule.RuleSuggestion),
) {
	if isEffect && hasAsyncModifier(callback) {
		ctx.ReportNode(callback, rule.RuleMessage{Description: "Effect callbacks are synchronous to prevent race conditions. " +
			"Put the async function inside:\n\n" +
			"useEffect(() => {\n" +
			"  async function fetchData() {\n" +
			"    // You can await here\n" +
			"    const response = await MyAPI.getData(someId);\n" +
			"    // ...\n" +
			"  }\n" +
			"  fetchData();\n" +
			"}, [someId]); // Or [] if effect doesn't need props or state\n\n" +
			"Learn more about data fetching with Hooks: https://react.dev/link/hooks-data-fetching"})
	}

	callbackBody := getCallbackBody(callback)
	if callbackBody == nil {
		return
	}

	// Component scope = the nearest enclosing function-like that
	// classifies as a React component or custom hook (PascalCase or
	// `use*`). Walk past intermediate anonymous IIFEs / non-component
	// closures (e.g. nested useEffect callbacks) so that a hook defined
	// inside another hook's body still resolves component-scope props
	// correctly. Mirrors upstream's `componentScope = currentScope`
	// loop in `visitFunctionWithDependencies`.
	componentFn := findEnclosingFunction(callback)
	for componentFn != nil && !isComponentOrHookFunction(componentFn) {
		componentFn = findEnclosingFunction(componentFn)
	}
	if componentFn == nil {
		return
	}

	// Tracks `setState` callees and the corresponding state Identifier.
	setStateCallSites := map[*ast.Symbol]*ast.Node{}
	stateVariableSymbols := map[*ast.Symbol]bool{}
	useEffectEventSymbols := map[*ast.Symbol]bool{}

	// Pre-scan the component body: identify all `useState` / `useReducer` /
	// `useTransition` / `useRef` / `useEffectEvent` declarations so we can
	// stamp their "stable identity" on references during the gather phase.
	primeStableSymbols(tc, sf, caches, componentFn, setStateCallSites, stateVariableSymbols, useEffectEventSymbols)

	// Gather references inside the callback body.
	dependencies := map[string]*dependency{}
	optionalChains := map[string]bool{}

	gatherDeps(
		ctx, sf, tc, caches, callback, callbackBody, componentFn,
		setStateCallSites, stateVariableSymbols,
		dependencies, optionalChains,
		isEffect,
	)

	stableDeps := map[string]bool{}
	for key, dep := range dependencies {
		if dep.IsStable {
			stableDeps[key] = true
		}
	}

	// Stale-assignment warnings: assignments to outer-scope variables from
	// inside the hook callback are usually mistakes. Reported in source
	// order. For each dep with at least one write, pick the earliest write
	// position; then sort entries by that position. This matches upstream's
	// observable order (iterates each variable's references in declaration
	// order; first-write wins per variable).
	type staleEntry struct {
		writeExpr *ast.Node
		key       string
	}
	staleByKey := map[string]*ast.Node{}
	for key, dep := range dependencies {
		for _, ref := range dep.Refs {
			if ref.WriteExpr == nil {
				continue
			}
			if existing, ok := staleByKey[key]; !ok || ref.WriteExpr.Pos() < existing.Pos() {
				staleByKey[key] = ref.WriteExpr
			}
		}
	}
	staleEntries := make([]staleEntry, 0, len(staleByKey))
	for key, we := range staleByKey {
		staleEntries = append(staleEntries, staleEntry{we, key})
	}
	sort.Slice(staleEntries, func(i, j int) bool {
		return staleEntries[i].writeExpr.Pos() < staleEntries[j].writeExpr.Pos()
	})
	for _, se := range staleEntries {
		ctx.ReportNode(se.writeExpr, rule.RuleMessage{Description: fmt.Sprintf(
			"Assignments to the '%s' variable from inside React Hook %s "+
				"will be lost after each render. To preserve the value over time, "+
				"store it in a useRef Hook and keep the mutable value in the '.current' property. "+
				"Otherwise, you can move this variable directly inside %s.",
			se.key, getCalleeText(sf, reactiveHook), getCalleeText(sf, reactiveHook),
		)})
	}
	if len(staleEntries) > 0 {
		return
	}

	// No deps array case: only emit the setState-without-deps diagnostic.
	if depsNode == nil {
		emitSetStateInsideEffectWarning(ctx, sf, tc, hookCall, reactiveHook, hookName, callback,
			dependencies, setStateCallSites, stableDeps, optionalChains, reportWithSuggestions)
		return
	}

	// Parse declared deps. Per-element diagnostics are deferred so they can
	// be flushed after the main missing-dep diagnostic — matching upstream's
	// report order.
	declared, deferredElementDiags, depsArrayBad := parseDeclaredDeps(ctx, sf, tc, reactiveHook, hookCall, depsNode,
		dependencies, useEffectEventSymbols, componentFn)
	if depsArrayBad {
		return
	}

	// External dependencies: declared deps that resolve outside componentFn
	// (i.e., globals / imports), used to suppress the "unnecessary" verdict
	// for effects and to drive the "Outer scope values" hint.
	externalDeps := computeExternalDeps(declared, componentFn, tc, sf)

	rec := collectRecommendations(dependencies, declared, stableDeps, externalDeps, isEffect)

	suggestedDeps := rec.Suggested
	problemCount := len(rec.Duplicate) + len(rec.Missing) + len(rec.Unnecessary)
	if problemCount == 0 {
		// No deps issues: check for constructions that change every render.
		emitConstructionWarnings(ctx, sf, tc, callback, depsNode, declared, hookName, opts, reportWithSuggestions)
		flushDeferredDiagnostics(ctx, deferredElementDiags, opts.EnableDangerousAutofixThisMayCauseInfiniteLoops)
		return
	}

	// For non-effect hooks with missing deps, recompute suggestions from scratch.
	if !isEffect && len(rec.Missing) > 0 {
		rec2 := collectRecommendations(dependencies, []declaredDependency{}, stableDeps, externalDeps, isEffect)
		suggestedDeps = rec2.Suggested
	}

	// Alphabetize suggestions iff the originals were alphabetized.
	if areDeclaredDepsAlphabetized(declared) {
		sort.Strings(suggestedDeps)
	}

	// Build the warning message.
	hookCalleeText := getCalleeText(sf, reactiveHook)
	msg := buildDepDiagnostic(hookCalleeText, rec, externalDeps, dependencies,
		hookName, optionalChains, setStateCallSites, stateVariableSymbols, componentFn, tc, callback)

	// Build suggestion text.
	formatted := make([]string, len(suggestedDeps))
	for i, k := range suggestedDeps {
		formatted[i] = formatDependency(k, optionalChains)
	}
	suggestionText := "[" + strings.Join(formatted, ", ") + "]"
	fix := rule.RuleFixReplace(sf, depsNode, suggestionText)
	reportWithSuggestions(depsNode, msg,
		rule.RuleSuggestion{
			Message:  rule.RuleMessage{Description: "Update the dependencies array to be: " + suggestionText},
			FixesArr: []rule.RuleFix{fix},
		},
	)
	// Flush per-element diagnostics AFTER the main report — upstream's
	// observable order is missing-dep first, element errors after.
	flushDeferredDiagnostics(ctx, deferredElementDiags, opts.EnableDangerousAutofixThisMayCauseInfiniteLoops)
}

// buildDepDiagnostic constructs the "React Hook X has a/an missing/unnecessary
// /duplicate dependencies: 'a', 'b'." message body, including any of the
// "Mutable values like ref.current aren't valid"/"Outer scope values like..."
// /setStateRecommendation extra warnings.
func buildDepDiagnostic(
	hookText string,
	rec recommendations,
	externalDeps map[string]bool,
	dependencies map[string]*dependency,
	hookName string,
	optionalChains map[string]bool,
	setStateCallSites map[*ast.Symbol]*ast.Node,
	stateVariableSymbols map[*ast.Symbol]bool,
	componentFn *ast.Node,
	tc *checker.Checker,
	callback *ast.Node,
) string {
	body := getWarningMessage(rec.Missing, "a", "missing", "include", optionalChains)
	if body == "" {
		body = getWarningMessage(rec.Unnecessary, "an", "unnecessary", "exclude", optionalChains)
	}
	if body == "" {
		body = getWarningMessage(rec.Duplicate, "a", "duplicate", "omit", optionalChains)
	}
	extra := ""
	// `ref.current` mutable warning.
	if len(rec.Unnecessary) > 0 {
		var badRef string
		keys := make([]string, 0, len(rec.Unnecessary))
		for k := range rec.Unnecessary {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if strings.HasSuffix(k, ".current") {
				badRef = k
				break
			}
		}
		if badRef != "" {
			extra = fmt.Sprintf(
				" Mutable values like '%s' aren't valid dependencies because mutating them doesn't re-render the component.",
				badRef,
			)
		} else if len(externalDeps) > 0 {
			extKeys := make([]string, 0, len(externalDeps))
			for k := range externalDeps {
				extKeys = append(extKeys, k)
			}
			sort.Strings(extKeys)
			// Mirrors upstream's `!scope.set.has(dep)` gate: when the
			// "external" name is also locally declared inside the callback
			// (e.g. `const fetchData = ...; useEffect(() => { fetchData() }, [fetchData])`),
			// the user clearly meant the local — don't add the outer-scope
			// hint. We don't have ESLint's scope manager, so we check by
			// walking the callback body for a matching declaration name.
			if !nameDeclaredInsideCallback(extKeys[0], callback) {
				extra = fmt.Sprintf(
					" Outer scope values like '%s' aren't valid dependencies because mutating them doesn't re-render the component.",
					extKeys[0],
				)
			}
		}
	}
	// `props.foo()` marks `props` as a dep — the warning is confusing, so
	// upstream appends a clarification. Mirrors upstream's
	// `isPropsOnlyUsedInMembers`.
	if extra == "" && rec.Missing["props"] {
		extra = propsOnlyMembersHint(dependencies, hookName)
	}
	// missingCallbackDep: when the only missing dep is a parameter that's
	// being called as a function, suggest wrapping it in useCallback in the
	// parent component. Mirrors upstream's same-named branch.
	if extra == "" && len(rec.Missing) > 0 {
		extra = missingCallbackDepHint(rec.Missing, dependencies, componentFn, tc)
	}
	// setState recommendation: when a missing dep is read inside a setState
	// call, append the appropriate hint. Mirrors upstream's three forms.
	if extra == "" && len(rec.Missing) > 0 {
		extra = setStateRecommendation(rec.Missing, dependencies, setStateCallSites,
			stateVariableSymbols, componentFn, tc)
	}
	return fmt.Sprintf("React Hook %s has %s%s", hookText, body, extra)
}

// nameDeclaredInsideCallback reports whether `name` is the name of a
// VariableDeclaration / FunctionDeclaration / ClassDeclaration / parameter
// inside `callback`'s body — i.e., it is a "local" rather than an "outer"
// scope value. Used to suppress the "Outer scope values" hint when the
// user clearly meant the local binding.
func nameDeclaredInsideCallback(name string, callback *ast.Node) bool {
	body := getCallbackBody(callback)
	if body == nil {
		return false
	}
	found := false
	var visit func(n *ast.Node) bool
	visit = func(n *ast.Node) bool {
		if found {
			return true
		}
		switch n.Kind {
		case ast.KindVariableDeclaration:
			if vn := n.AsVariableDeclaration().Name(); vn != nil && vn.Kind == ast.KindIdentifier && vn.AsIdentifier().Text == name {
				found = true
				return true
			}
		case ast.KindFunctionDeclaration, ast.KindClassDeclaration:
			if vn := n.Name(); vn != nil && vn.Kind == ast.KindIdentifier && vn.AsIdentifier().Text == name {
				found = true
				return true
			}
		case ast.KindParameter:
			if vn := n.AsParameterDeclaration().Name(); vn != nil && vn.Kind == ast.KindIdentifier && vn.AsIdentifier().Text == name {
				found = true
				return true
			}
		case ast.KindBindingElement:
			if vn := n.AsBindingElement().Name(); vn != nil && vn.Kind == ast.KindIdentifier && vn.AsIdentifier().Text == name {
				found = true
				return true
			}
		}
		// Don't descend into nested function-likes — their bindings don't
		// shadow the outer callback's scope. Exception: the callback itself,
		// which we already entered via getCallbackBody.
		if isFunctionLikeContainer(n) && n != callback {
			return false
		}
		n.ForEachChild(visit)
		return false
	}
	visit(body)
	return found
}

// propsOnlyMembersHint mirrors upstream's `isPropsOnlyUsedInMembers`
// check: when `props` is missing AND every observed reference of `props`
// inside the callback is `props.<member>` (never bare `props`), append
// the destructure-outside-the-hook recommendation.
func propsOnlyMembersHint(dependencies map[string]*dependency, hookName string) string {
	dep, ok := dependencies["props"]
	if !ok {
		return ""
	}
	for _, ref := range dep.Refs {
		id := ref.Identifier
		if id == nil {
			return ""
		}
		p := id.Parent
		if p == nil {
			return ""
		}
		if p.Kind != ast.KindPropertyAccessExpression {
			return ""
		}
		// `props` must be the receiver (object), not the property name.
		if p.AsPropertyAccessExpression().Expression != id {
			return ""
		}
	}
	return fmt.Sprintf(
		" However, 'props' will change when *any* prop changes, so the preferred fix is to destructure the 'props' object outside of the %s call and refer to those specific props inside %s.",
		hookName, hookName,
	)
}

// missingCallbackDepHint mirrors upstream's `missingCallbackDep` branch.
// For each missing dep that:
//   - resolves to a Parameter declaration (i.e., a destructured prop), AND
//   - is called at least once inside the callback body
//
// emit "If 'X' changes too often, find the parent component that defines
// it and wrap that definition in useCallback." Returns "" otherwise.
func missingCallbackDepHint(
	missing map[string]bool,
	dependencies map[string]*dependency,
	componentFn *ast.Node,
	tc *checker.Checker,
) string {
	if tc == nil {
		return ""
	}
	missingKeys := make([]string, 0, len(missing))
	for k := range missing {
		missingKeys = append(missingKeys, k)
	}
	sort.Strings(missingKeys)
	for _, dep := range missingKeys {
		used, ok := dependencies[dep]
		if !ok || len(used.Refs) == 0 {
			continue
		}
		// Symbol must be a Parameter declared in the component scope.
		first := used.Refs[0]
		if first.Symbol == nil || len(first.Symbol.Declarations) == 0 {
			continue
		}
		decl := first.Symbol.Declarations[0]
		if decl.Kind != ast.KindParameter {
			// Could also be BindingElement nested inside a Parameter (destructured).
			cur := decl
			isParam := false
			for cur != nil {
				if cur.Kind == ast.KindParameter {
					isParam = true
					break
				}
				cur = cur.Parent
			}
			if !isParam {
				continue
			}
		}
		// Must be called somewhere.
		isCalled := false
		for _, ref := range used.Refs {
			id := ref.Identifier
			if id.Parent != nil && id.Parent.Kind == ast.KindCallExpression {
				if id.Parent.AsCallExpression().Expression == id {
					isCalled = true
					break
				}
			}
		}
		if !isCalled {
			continue
		}
		return fmt.Sprintf(
			" If '%s' changes too often, find the parent component that defines it and wrap that definition in useCallback.",
			dep,
		)
	}
	return ""
}

// setStateRecommendation mirrors upstream's same-named computation. For
// each missing dep, walks references upward looking for a CallExpression
// whose callee was registered in `setStateCallSites`. When found,
// classifies into three forms:
//
//   - "updater"      — `setX(x + 1)`: missing dep IS the state variable for `setX`
//   - "reducer"      — `setCount(count + increment)`: missing dep IS a state variable for some other setter
//   - "inlineReducer"— `setX(...prop...)`: missing dep is a parameter (prop)
//
// Returns the suffix (with leading space) or "" when no recommendation matches.
func setStateRecommendation(
	missing map[string]bool,
	dependencies map[string]*dependency,
	setStateCallSites map[*ast.Symbol]*ast.Node,
	stateVariableSymbols map[*ast.Symbol]bool,
	componentFn *ast.Node,
	tc *checker.Checker,
) string {
	if tc == nil || componentFn == nil {
		return ""
	}
	missingKeys := make([]string, 0, len(missing))
	for k := range missing {
		missingKeys = append(missingKeys, k)
	}
	sort.Strings(missingKeys)
	for _, dep := range missingKeys {
		used, ok := dependencies[dep]
		if !ok {
			continue
		}
		for _, ref := range used.Refs {
			id := ref.Identifier
			cur := id.Parent
			for cur != nil && cur != componentFn {
				if cur.Kind == ast.KindCallExpression {
					call := cur.AsCallExpression()
					calleeSym := tc.GetSymbolAtLocation(call.Expression)
					if calleeSym != nil {
						if stateForSetter, isSetter := setStateCallSites[calleeSym]; isSetter {
							setterName := ""
							if c := call.Expression; c != nil && c.Kind == ast.KindIdentifier {
								setterName = c.AsIdentifier().Text
							}
							// Form selection.
							if stateForSetter != nil && stateForSetter.Kind == ast.KindIdentifier &&
								stateForSetter.AsIdentifier().Text == dep {
								// updater form: setCount(count + 1)
								// Mirrors upstream's `missingDep.slice(0, 1)` —
								// JS string indexing is over UTF-16 code units;
								// for ASCII state names this matches `dep[:1]`,
								// but for multi-byte names (CJK / emoji /
								// accented Latin) Go's byte slice would
								// truncate. Use rune iteration for the first
								// scalar character.
								initial := dep
								for _, r := range dep {
									initial = string(r)
									break
								}
								return fmt.Sprintf(
									" You can also do a functional update '%s(%s => ...)' if you only need '%s' in the '%s' call.",
									setterName, initial, dep, setterName,
								)
							}
							if ref.Symbol != nil && stateVariableSymbols[ref.Symbol] {
								return fmt.Sprintf(
									" You can also replace multiple useState variables with useReducer if '%s' needs the current value of '%s'.",
									setterName, dep,
								)
							}
							if ref.Symbol != nil && len(ref.Symbol.Declarations) > 0 {
								d := ref.Symbol.Declarations[0]
								// Direct Parameter or BindingElement nested in Parameter
								// (destructured prop) — both count as "from props".
								isParam := d.Kind == ast.KindParameter
								if !isParam && d.Kind == ast.KindBindingElement {
									for cur := d.Parent; cur != nil; cur = cur.Parent {
										if cur.Kind == ast.KindParameter {
											isParam = true
											break
										}
										if isFunctionLikeContainer(cur) {
											break
										}
									}
								}
								if isParam {
									return fmt.Sprintf(
										" If '%s' needs the current value of '%s', you can also switch to useReducer instead of useState and read '%s' in the reducer.",
										setterName, dep, dep,
									)
								}
							}
							break
						}
					}
				}
				cur = cur.Parent
			}
		}
	}
	return ""
}

// computeExternalDeps mirrors upstream's external-dep tracking: declared
// deps whose root identifier resolves outside the component scope.
func computeExternalDeps(declared []declaredDependency, componentFn *ast.Node, tc *checker.Checker, sf *ast.SourceFile) map[string]bool {
	out := map[string]bool{}
	if componentFn == nil {
		return out
	}
	for _, dd := range declared {
		// Walk down the dep node to find the root identifier (the leftmost
		// identifier in `a.b.c`).
		root := dd.Node
		for root != nil && root.Kind == ast.KindPropertyAccessExpression {
			root = ast.SkipParentheses(root.AsPropertyAccessExpression().Expression)
		}
		if root == nil || root.Kind != ast.KindIdentifier {
			continue
		}
		// Hybrid resolution mirroring `processIdentifier`: trust TC when
		// it returns a declaration inside the component; otherwise fall
		// back to an AST walk for a same-named local binding (covers
		// destructure-rename forms whose TC sym resolves to the source
		// type's property in an external file).
		var decl *ast.Node
		if tc != nil {
			sym := tc.GetSymbolAtLocation(root)
			if sym != nil && len(sym.Declarations) > 0 {
				if anyDeclWithinComponent(sym, componentFn) {
					// Pick the first declaration that lies inside componentFn
					// — that's the local binding we care about.
					for _, d := range sym.Declarations {
						if d != nil && containsNode(componentFn, d) {
							decl = d
							break
						}
					}
				}
			}
		}
		if decl == nil {
			decl = resolveDeclaration(nil, root, root.AsIdentifier().Text, componentFn)
		}
		if decl == nil {
			out[dd.Key] = true
			continue
		}
		if !containsNode(componentFn, decl) {
			out[dd.Key] = true
		}
	}
	return out
}

// emitSetStateInsideEffectWarning mirrors upstream's "infinite chain of
// updates" diagnostic for `useEffect(() => { setX(...); })` with no deps.
func emitSetStateInsideEffectWarning(
	ctx rule.RuleContext,
	sf *ast.SourceFile,
	tc *checker.Checker,
	hookCall *ast.Node,
	reactiveHook *ast.Node,
	hookName string,
	callback *ast.Node,
	dependencies map[string]*dependency,
	setStateCallSites map[*ast.Symbol]*ast.Node,
	stableDeps map[string]bool,
	optionalChains map[string]bool,
	reportWithSuggestions func(*ast.Node, string, ...rule.RuleSuggestion),
) {
	var found string
	// Iterate dep keys in source order so the "first setState in callback"
	// detected matches upstream's deterministic walk.
	keys := make([]string, 0, len(dependencies))
	for k := range dependencies {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		dep := dependencies[key]
		if found != "" {
			break
		}
		for _, ref := range dep.Refs {
			if ref.Symbol != nil {
				if _, ok := setStateCallSites[ref.Symbol]; ok {
					// Verify the ref is inside the effect callback directly,
					// not nested.
					encl := findEnclosingFunction(ref.Identifier)
					if encl == callback {
						found = key
						break
					}
				}
			}
		}
	}
	if found == "" {
		return
	}
	// Collect recommendations using stableDeps so setState/dispatch (and
	// other stable identities) are filtered out of the suggested array.
	rec := collectRecommendations(dependencies, []declaredDependency{}, stableDeps, map[string]bool{}, true)
	formatted := make([]string, len(rec.Suggested))
	for i, k := range rec.Suggested {
		formatted[i] = formatDependency(k, optionalChains)
	}
	suggestionText := "[" + strings.Join(formatted, ", ") + "]"
	// Insert AFTER the callback (not after the whole call expression),
	// matching upstream's `fixer.insertTextAfter(node, ...)`. Using
	// hookCall.End() would put the deps array INSIDE the closing `)`.
	fix := rule.RuleFix{
		Text:  ", " + suggestionText,
		Range: callback.Loc.WithPos(callback.End()),
	}
	_ = sf
	_ = tc
	_ = reactiveHook
	reportWithSuggestions(hookCall,
		fmt.Sprintf(
			"React Hook %s contains a call to '%s'. Without a list of dependencies, this can lead to an infinite chain of updates. To fix this, pass %s as a second argument to the %s Hook.",
			hookName, found, suggestionText, hookName,
		),
		rule.RuleSuggestion{
			Message:  rule.RuleMessage{Description: "Add dependencies array: " + suggestionText},
			FixesArr: []rule.RuleFix{fix},
		},
	)
}

// emitConstructionWarnings mirrors upstream's `scanForConstructions`
// branch when declared deps have no missing / unnecessary / duplicate
// problems but some declared deps refer to function/object/etc.
// expressions that change identity every render.
func emitConstructionWarnings(
	ctx rule.RuleContext,
	sf *ast.SourceFile,
	tc *checker.Checker,
	callback *ast.Node,
	depsNode *ast.Node,
	declared []declaredDependency,
	hookName string,
	opts Options,
	reportWithSuggestions func(*ast.Node, string, ...rule.RuleSuggestion),
) {
	componentFn := findEnclosingFunction(callback)
	if componentFn == nil {
		return
	}
	consts := scanForConstructions(declared, depsNode, componentFn, callback, tc, sf)
	depsLine, _ := lineOf(sf, depsNode)
	for _, c := range consts {
		wrapper := "useMemo"
		ctype := "initialization"
		if c.DepType == "function" {
			wrapper = "useCallback"
			ctype = "definition"
		}
		nameText := ""
		if c.Variable != nil && c.Variable.Kind == ast.KindIdentifier {
			nameText = c.Variable.AsIdentifier().Text
		}
		defaultAdvice := fmt.Sprintf(
			"wrap the %s of '%s' in its own %s() Hook.",
			ctype, nameText, wrapper,
		)
		var advice string
		if c.IsUsedOutsideHook {
			advice = "To fix this, " + defaultAdvice
		} else {
			advice = fmt.Sprintf(
				"Move it inside the %s callback. Alternatively, %s",
				hookName, defaultAdvice,
			)
		}
		causation := "makes"
		if c.DepType == "conditional" || c.DepType == "logical expression" {
			causation = "could make"
		}
		msg := fmt.Sprintf(
			"The '%s' %s %s the dependencies of %s Hook (at line %d) change on every render. %s",
			nameText, c.DepType, causation, hookName, depsLine, advice,
		)
		var sugs []rule.RuleSuggestion
		if c.IsUsedOutsideHook && c.Decl != nil && c.Decl.Kind == ast.KindVariableDeclaration && c.DepType == "function" {
			before := "useCallback("
			after := ")"
			if wrapper == "useMemo" {
				before = "useMemo(() => { return "
				after = "; })"
			}
			sugs = []rule.RuleSuggestion{{
				Message: rule.RuleMessage{Description: fmt.Sprintf(
					"Wrap the %s of '%s' in its own %s() Hook.",
					ctype, nameText, wrapper,
				)},
				FixesArr: []rule.RuleFix{
					rule.RuleFixInsertBefore(sf, c.InitNode, before),
					rule.RuleFixInsertAfter(c.InitNode, after),
				},
			}}
		}
		reportWithSuggestions(c.Decl, msg, sugs...)
	}
	_ = opts
}

// lineOf returns the 1-based line number of `node`.
func lineOf(sf *ast.SourceFile, node *ast.Node) (int, error) {
	if node == nil {
		return 0, nil
	}
	tr := utils.TrimNodeTextRange(sf, node)
	line, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, tr.Pos())
	return line + 1, nil
}

// elementDiagnostic is a deferred error for one entry in the deps array.
// Held in a slice so the caller can flush it AFTER the main missing/
// unnecessary diagnostic is emitted, matching upstream's report order.
type elementDiagnostic struct {
	node    *ast.Node
	message string
	// suggestion is non-nil only for the useEffectEvent variant — every
	// other element diagnostic upstream is suggestion-less.
	suggestion *rule.RuleSuggestion
}

// parseDeclaredDeps mirrors upstream's loop over the array literal: for
// each element, normalize via `analyzePropertyChainText`; on failure,
// queue the "complex expression" / "literal not a valid dependency"
// diagnostic. The diagnostics are returned (not emitted) so the caller
// can flush them after the main missing-dep diagnostic — upstream's
// per-element reports come later in the output stream.
//
// Returns `bad=true` to signal that the deps array itself is malformed
// and the caller should bail out (no further missing-dep reports).
func parseDeclaredDeps(
	ctx rule.RuleContext,
	sf *ast.SourceFile,
	tc *checker.Checker,
	reactiveHook *ast.Node,
	hookCall *ast.Node,
	depsNode *ast.Node,
	dependencies map[string]*dependency,
	useEffectEventSymbols map[*ast.Symbol]bool,
	componentFn *ast.Node,
) (declared []declaredDependency, deferred []elementDiagnostic, bad bool) {
	depsNodeStripped := stripAsExpression(depsNode)
	if depsNodeStripped == nil || depsNodeStripped.Kind != ast.KindArrayLiteralExpression {
		// Upstream still emits the missing-dep diagnostic afterwards
		// using an empty declared-deps list, so don't return bad=true here.
		// The "not an array literal" error itself is still reported.
		ctx.ReportNode(depsNode, rule.RuleMessage{Description: fmt.Sprintf(
			"React Hook %s was passed a dependency list that is not an array literal. "+
				"This means we can't statically verify whether you've passed the correct dependencies.",
			getCalleeText(sf, reactiveHook),
		)})
		return nil, nil, false
	}
	arr := depsNodeStripped.AsArrayLiteralExpression()
	if arr.Elements == nil {
		return nil, nil, false
	}
	for _, el := range arr.Elements.Nodes {
		if el == nil || isElidedComma(el) {
			continue
		}
		if el.Kind == ast.KindSpreadElement {
			deferred = append(deferred, elementDiagnostic{
				node: el,
				message: fmt.Sprintf(
					"React Hook %s has a spread element in its dependency array. "+
						"This means we can't statically verify whether you've passed the correct dependencies.",
					getCalleeText(sf, reactiveHook),
				),
			})
			continue
		}
		// useEffectEvent rejection — emitted with a `Remove the dependency` suggestion.
		if tc != nil && el.Kind == ast.KindIdentifier {
			sym := tc.GetSymbolAtLocation(el)
			if sym != nil && useEffectEventSymbols[sym] {
				removeRange := utils.TrimNodeTextRange(sf, el)
				deferred = append(deferred, elementDiagnostic{
					node: el,
					message: fmt.Sprintf(
						"Functions returned from `useEffectEvent` must not be included in the dependency array. Remove `%s` from the list.",
						getCalleeText(sf, el),
					),
					suggestion: &rule.RuleSuggestion{
						Message:  rule.RuleMessage{Description: fmt.Sprintf("Remove the dependency `%s`", getCalleeText(sf, el))},
						FixesArr: []rule.RuleFix{rule.RuleFixRemoveRange(removeRange)},
					},
				})
			}
		}

		key, ok := analyzePropertyChainText(el, nil)
		if !ok {
			// Literal value or complex expression. Mirrors upstream's full
			// set of `Literal` value kinds — string, number, bigint, null,
			// boolean, regex. Each gets the "literal is not a valid dep"
			// diagnostic; everything else falls through to "complex
			// expression".
			elInner := stripAsExpression(el)
			switch elInner.Kind {
			case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
				val := ""
				if elInner.Kind == ast.KindStringLiteral {
					val = elInner.AsStringLiteral().Text
				} else {
					val = elInner.AsNoSubstitutionTemplateLiteral().Text
				}
				var msg string
				if dependencies[val] != nil {
					msg = fmt.Sprintf(
						"The %s literal is not a valid dependency because it never changes. "+
							"Did you mean to include %s in the array instead?",
						getCalleeText(sf, el), val,
					)
				} else {
					msg = fmt.Sprintf(
						"The %s literal is not a valid dependency because it never changes. You can safely remove it.",
						getCalleeText(sf, el),
					)
				}
				deferred = append(deferred, elementDiagnostic{node: el, message: msg})
				continue
			case ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindNullKeyword,
				ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindRegularExpressionLiteral:
				deferred = append(deferred, elementDiagnostic{
					node: el,
					message: fmt.Sprintf(
						"The %s literal is not a valid dependency because it never changes. You can safely remove it.",
						getCalleeText(sf, el),
					),
				})
				continue
			default:
				deferred = append(deferred, elementDiagnostic{
					node: el,
					message: fmt.Sprintf(
						"React Hook %s has a complex expression in the dependency array. "+
							"Extract it to a separate variable so it can be statically checked.",
						getCalleeText(sf, reactiveHook),
					),
				})
				continue
			}
		}
		declared = append(declared, declaredDependency{Key: key, Node: el})
	}
	return declared, deferred, false
}

// flushDeferredDiagnostics emits the per-element diagnostics queued by
// parseDeclaredDeps. Called by the caller AFTER the main missing-dep
// diagnostic, so the output order matches upstream.
//
// `enableDangerous` mirrors upstream's `reportProblem` promotion gate:
// when `enableDangerousAutofixThisMayCauseInfiniteLoops` is on and the
// diagnostic carries a suggestion, the first suggestion's first fix is
// promoted to a top-level autofix while keeping the suggestion array.
// This applies to every `reportProblem` call in upstream — including
// the per-element useEffectEvent rejection — so we plumb the flag here.
func flushDeferredDiagnostics(ctx rule.RuleContext, deferred []elementDiagnostic, enableDangerous bool) {
	for _, d := range deferred {
		if d.suggestion != nil {
			if enableDangerous && len(d.suggestion.FixesArr) > 0 && ctx.ReportNodeWithFixesAndSuggestions != nil {
				firstFix := []rule.RuleFix{d.suggestion.FixesArr[0]}
				ctx.ReportNodeWithFixesAndSuggestions(
					d.node,
					rule.RuleMessage{Description: d.message},
					firstFix,
					[]rule.RuleSuggestion{*d.suggestion},
				)
				continue
			}
			ctx.ReportNodeWithSuggestions(d.node, rule.RuleMessage{Description: d.message}, *d.suggestion)
		} else {
			ctx.ReportNode(d.node, rule.RuleMessage{Description: d.message})
		}
	}
}

// getCallbackBody returns the body of the function-like callback. For
// arrow functions with a non-block body, the expression itself is the
// "body" we walk.
func getCallbackBody(fn *ast.Node) *ast.Node {
	if fn == nil {
		return nil
	}
	switch fn.Kind {
	case ast.KindFunctionDeclaration:
		return fn.AsFunctionDeclaration().Body
	case ast.KindFunctionExpression:
		return fn.AsFunctionExpression().Body
	case ast.KindArrowFunction:
		return fn.AsArrowFunction().Body
	}
	return nil
}

// primeStableSymbols pre-walks the component body to populate the
// `setStateCallSites`, `stateVariableSymbols`, and `useEffectEventSymbols`
// caches. These drive both the stable-known-hook-value classification and
// the setState-without-deps diagnostic.
func primeStableSymbols(
	tc *checker.Checker,
	sf *ast.SourceFile,
	caches *runCaches,
	componentFn *ast.Node,
	setStateCallSites map[*ast.Symbol]*ast.Node,
	stateVariableSymbols map[*ast.Symbol]bool,
	useEffectEventSymbols map[*ast.Symbol]bool,
) {
	if componentFn == nil || tc == nil {
		return
	}
	body := getCallbackBody(componentFn)
	if body == nil {
		return
	}
	var visit func(n *ast.Node) bool
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindVariableDeclaration {
			vd := n.AsVariableDeclaration()
			if vd.Initializer != nil {
				init := stripAsExpression(vd.Initializer)
				if init.Kind == ast.KindCallExpression {
					callee := stripReactNamespace(init.AsCallExpression().Expression)
					if callee != nil && callee.Kind == ast.KindIdentifier {
						name := callee.AsIdentifier().Text
						switch name {
						case "useState":
							recordStateBindings(tc, vd, setStateCallSites, stateVariableSymbols, componentFn, caches)
						case "useReducer", "useActionState":
							// dispatch / setState (second tuple element) is
							// registered too — used by setState-without-deps
							// detection and the setStateRecommendation reducer
							// form. The state variable for these hooks is NOT
							// recorded in `stateVariableSymbols` (mirrors
							// upstream — only useState's first element).
							recordReducerBindings(tc, vd, setStateCallSites, componentFn, caches)
						case "useEffectEvent":
							if bn := vd.Name(); bn != nil && bn.Kind == ast.KindIdentifier {
								if sym := tc.GetSymbolAtLocation(bn); sym != nil {
									useEffectEventSymbols[sym] = true
								}
							}
						}
					}
				}
			}
		}
		if isFunctionLikeContainer(n) && n != componentFn {
			return false
		}
		n.ForEachChild(visit)
		return false
	}
	visit(body)
}

// recordStateBindings populates setStateCallSites and stateVariableSymbols
// for a `const [x, setX] = useState(...)` declaration. Mirrors upstream's
// `writeCount > 1` gate: if the setter binding is reassigned anywhere in
// the source, treat it as no longer stable (don't register).
func recordStateBindings(
	tc *checker.Checker,
	vd *ast.VariableDeclaration,
	setStateCallSites map[*ast.Symbol]*ast.Node,
	stateVariableSymbols map[*ast.Symbol]bool,
	scope *ast.Node,
	caches *runCaches,
) {
	bindingName := vd.Name()
	if bindingName == nil || bindingName.Kind != ast.KindArrayBindingPattern {
		return
	}
	arr := bindingName.AsBindingPattern()
	if arr.Elements == nil || len(arr.Elements.Nodes) != 2 {
		return
	}
	stateEl := arr.Elements.Nodes[0]
	setterEl := arr.Elements.Nodes[1]
	stateName := bindingElementName(stateEl)
	setterName := bindingElementName(setterEl)
	if stateName != nil {
		if sym := tc.GetSymbolAtLocation(stateName); sym != nil {
			stateVariableSymbols[sym] = true
		}
	}
	if setterName != nil {
		if sym := tc.GetSymbolAtLocation(setterName); sym != nil {
			// Skip registration if the setter is reassigned. Upstream's
			// `writeCount > 1` test counts the binding declaration itself
			// as 1 write, so any further assignment makes it unstable.
			if !cachedHasExtraWrite(sym, scope, tc, caches) {
				setStateCallSites[sym] = stateName
			}
		}
	}
}

// cachedHasExtraWrite is the memoized wrapper around
// `hasExtraWriteToSymbol`. Keys on the symbol pointer; result is stable
// across a single Run.
func cachedHasExtraWrite(sym *ast.Symbol, scope *ast.Node, tc *checker.Checker, caches *runCaches) bool {
	if caches == nil {
		return hasExtraWriteToSymbol(sym, scope, tc)
	}
	if v, ok := caches.extraWrite[sym]; ok {
		return v
	}
	v := hasExtraWriteToSymbol(sym, scope, tc)
	caches.extraWrite[sym] = v
	return v
}

// cachedHasCurrentAssignment is the memoized wrapper around
// `hasExtraWriteToSymbol`'s sibling helper for `<sym>.current = ...`.
// Same caching rationale.
func cachedHasCurrentAssignment(sym *ast.Symbol, scope *ast.Node, tc *checker.Checker, caches *runCaches) bool {
	if caches == nil {
		return hasCurrentAssignment(sym, scope, tc)
	}
	if v, ok := caches.currentAssign[sym]; ok {
		return v
	}
	v := hasCurrentAssignment(sym, scope, tc)
	caches.currentAssign[sym] = v
	return v
}

// recordReducerBindings registers the dispatch (second tuple element)
// of `const [state, dispatch] = useReducer(...)` / `useActionState(...)`
// as a stable setter. Mirrors the useState branch but does NOT mark the
// state variable in `stateVariableSymbols` — that map is consulted only
// for the useState-specific "reducer" recommendation form.
func recordReducerBindings(
	tc *checker.Checker,
	vd *ast.VariableDeclaration,
	setStateCallSites map[*ast.Symbol]*ast.Node,
	scope *ast.Node,
	caches *runCaches,
) {
	bindingName := vd.Name()
	if bindingName == nil || bindingName.Kind != ast.KindArrayBindingPattern {
		return
	}
	arr := bindingName.AsBindingPattern()
	if arr.Elements == nil || len(arr.Elements.Nodes) != 2 {
		return
	}
	stateName := bindingElementName(arr.Elements.Nodes[0])
	setterName := bindingElementName(arr.Elements.Nodes[1])
	if setterName != nil {
		if sym := tc.GetSymbolAtLocation(setterName); sym != nil {
			if !cachedHasExtraWrite(sym, scope, tc, caches) {
				setStateCallSites[sym] = stateName
			}
		}
	}
}

// hasExtraWriteToSymbol reports whether the given symbol has any
// non-declaration write reference. Mirrors upstream's `writeCount > 1`
// detection on stable hook bindings (used by `recordStateBindings` to
// disable setter stability under reassignment, and by stableForId for
// `isFunctionWithoutCapturedValues` to disable function stability when
// the function binding itself is later reassigned).
//
// Covers four write shapes:
//   - `setX = ...`            simple assignment
//   - `setX += 1`             compound assignment (`+=`/`-=`/`*=`/...)
//   - `({setX} = obj)` / `[setX] = arr`  destructuring assignment
//   - `setX++` / `--setX`     update expression
//
// Element-access writes like `obj.setX = ...` are NOT counted — those
// touch the property, not the binding.
//
// `scope` bounds the walk. Pass `componentFn` whenever possible (the
// rule only cares about writes within the surrounding component or
// hook); fall back to `sf.AsNode()` if the call site doesn't have a
// component scope handy. Limiting the walk to the component avoids the
// O(file_size) per-call cost that surfaced in PR-808 review.
func hasExtraWriteToSymbol(sym *ast.Symbol, scope *ast.Node, tc *checker.Checker) bool {
	if sym == nil || scope == nil || tc == nil {
		return false
	}
	target := ""
	if len(sym.Declarations) > 0 {
		if n := sym.Declarations[0].Name(); n != nil && n.Kind == ast.KindIdentifier {
			target = n.AsIdentifier().Text
		}
	}
	if target == "" {
		return false
	}
	matchesIdent := func(n *ast.Node) bool {
		if n == nil {
			return false
		}
		// Walk past paren / TS-only expression wrappers so shapes like
		// `(setCount as any)++` reach the inner Identifier.
		x := stripAsExpression(n)
		if x == nil || x.Kind != ast.KindIdentifier || x.AsIdentifier().Text != target {
			return false
		}
		s := tc.GetSymbolAtLocation(x)
		return s != nil && s == sym
	}
	found := false
	var visit func(n *ast.Node) bool
	visit = func(n *ast.Node) bool {
		if found {
			return true
		}
		switch n.Kind {
		case ast.KindBinaryExpression:
			be := n.AsBinaryExpression()
			if be.OperatorToken != nil {
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
					lhs := ast.SkipParentheses(be.Left)
					if matchesIdent(lhs) {
						found = true
						return true
					}
					// Destructuring assignment: `({setX} = obj)` /
					// `[setX] = arr` — LHS is an ObjectLiteral / Array
					// literal that visit-recursion will descend into;
					// the inner Identifier still resolves to `sym` and
					// would be a reference, but here we want only the
					// LHS-binding form (ShorthandPropertyAssignment /
					// element). Walk `lhs`'s children explicitly.
					if walkAssignmentTargetForBinding(lhs, matchesIdent) {
						found = true
						return true
					}
				}
			}
		case ast.KindPrefixUnaryExpression:
			pu := n.AsPrefixUnaryExpression()
			if pu.Operator == ast.KindPlusPlusToken || pu.Operator == ast.KindMinusMinusToken {
				if matchesIdent(pu.Operand) {
					found = true
					return true
				}
			}
		case ast.KindPostfixUnaryExpression:
			pu := n.AsPostfixUnaryExpression()
			if pu.Operator == ast.KindPlusPlusToken || pu.Operator == ast.KindMinusMinusToken {
				if matchesIdent(pu.Operand) {
					found = true
					return true
				}
			}
		}
		n.ForEachChild(visit)
		return false
	}
	visit(scope)
	return found
}

// walkAssignmentTargetForBinding inspects an assignment LHS that is an
// ObjectLiteral (`({x} = ...)`) or ArrayLiteral (`[x] = ...`) destructure
// pattern, looking for a binding-position Identifier matching `match`.
// Returns true on first hit. Property values inside ObjectLiteral are the
// binding targets (e.g., `({a: x} = obj)` binds `x`).
func walkAssignmentTargetForBinding(lhs *ast.Node, match func(*ast.Node) bool) bool {
	if lhs == nil {
		return false
	}
	switch lhs.Kind {
	case ast.KindObjectLiteralExpression:
		ole := lhs.AsObjectLiteralExpression()
		if ole.Properties == nil {
			return false
		}
		for _, p := range ole.Properties.Nodes {
			switch p.Kind {
			case ast.KindShorthandPropertyAssignment:
				name := p.Name()
				if match(name) {
					return true
				}
			case ast.KindPropertyAssignment:
				pa := p.AsPropertyAssignment()
				if walkAssignmentTargetForBinding(pa.Initializer, match) {
					return true
				}
			case ast.KindSpreadAssignment:
				sa := p.AsSpreadAssignment()
				if match(sa.Expression) || walkAssignmentTargetForBinding(sa.Expression, match) {
					return true
				}
			}
		}
	case ast.KindArrayLiteralExpression:
		ale := lhs.AsArrayLiteralExpression()
		if ale.Elements == nil {
			return false
		}
		for _, el := range ale.Elements.Nodes {
			if el == nil {
				continue
			}
			if el.Kind == ast.KindOmittedExpression {
				continue
			}
			if el.Kind == ast.KindSpreadElement {
				se := el.AsSpreadElement()
				if match(se.Expression) || walkAssignmentTargetForBinding(se.Expression, match) {
					return true
				}
				continue
			}
			if match(el) {
				return true
			}
			if walkAssignmentTargetForBinding(el, match) {
				return true
			}
		}
	}
	return false
}

// bindingElementName returns the binding's Identifier name node.
func bindingElementName(be *ast.Node) *ast.Node {
	if be == nil || be.Kind != ast.KindBindingElement {
		return nil
	}
	n := be.AsBindingElement().Name()
	if n == nil || n.Kind != ast.KindIdentifier {
		return nil
	}
	return n
}

// gatherDeps walks every Identifier reference inside the callback body
// and records it as a dependency when the resolved declaration sits in
// the component scope (between the callback and the component body
// boundary).
func gatherDeps(
	ctx rule.RuleContext,
	sf *ast.SourceFile,
	tc *checker.Checker,
	caches *runCaches,
	callback *ast.Node,
	body *ast.Node,
	componentFn *ast.Node,
	setStateCallSites map[*ast.Symbol]*ast.Node,
	stateVariableSymbols map[*ast.Symbol]bool,
	dependencies map[string]*dependency,
	optionalChains map[string]bool,
	isEffect bool,
) {
	// Cache: for each resolved symbol, whether it's "stable known".
	stableSymCache := map[*ast.Symbol]bool{}
	hasExtraWrite := func(s *ast.Symbol) bool {
		return cachedHasExtraWrite(s, componentFn, tc, caches)
	}
	// Forward-declared so isFunctionWithoutCapturedValues can recurse
	// through it for transitive function stability.
	var stableForId func(id *ast.Node, sym *ast.Symbol) bool
	stableForId = func(id *ast.Node, sym *ast.Symbol) bool {
		if sym == nil {
			return false
		}
		if v, ok := stableSymCache[sym]; ok {
			return v
		}
		// Tag the symbol as in-progress to break recursion cycles
		// (mutual recursion between two functions): treat as not-stable
		// while we're still computing.
		stableSymCache[sym] = false
		// Setter / dispatch from useState/useReducer/useTransition.
		if _, ok := setStateCallSites[sym]; ok {
			stableSymCache[sym] = true
			return true
		}
		if len(sym.Declarations) > 0 {
			decl := sym.Declarations[0]
			// BindingElement decls live inside the VariableDeclaration's
			// pattern. Walk up to the declaring VariableDeclaration so
			// `isStableHookValue` can inspect the initializer.
			bindingId := decl.Name()
			vdNode := decl
			if decl.Kind == ast.KindBindingElement {
				vdNode = decl.Parent
				for vdNode != nil && vdNode.Kind != ast.KindVariableDeclaration {
					vdNode = vdNode.Parent
				}
			}
			stable, _ := isStableHookValue(vdNode, bindingId)
			if stable {
				// Mirrors upstream's `writeCount > 1` gate: a stable hook
				// binding that is reassigned anywhere is no longer stable.
				if tc != nil && hasExtraWrite(sym) {
					return false
				}
				stableSymCache[sym] = true
				return true
			}
			// Upstream's `isFunctionWithoutCapturedValues`: a user-defined
			// function declared in component scope is stable iff it
			// doesn't capture any non-stable variables from pure scope.
			// We use the TypeChecker to resolve every Identifier inside
			// the function body to its declaring symbol, then test each
			// captured one against `stableForId` recursively. This mirrors
			// upstream's `fnScope.through` walk (which we don't have access
			// to without ESLint's scope manager).
			//
			// Reassignment guard: if the binding itself is reassigned
			// elsewhere (e.g. `handler = somethingElse`), the resolved
			// function may no longer point at our analyzed body. Mirrors
			// upstream's `writeCount > 1` gate that disables the
			// stable-function classification under reassignment.
			if tc != nil && !hasExtraWrite(sym) {
				if isFunctionWithoutCapturedValues(decl, componentFn, tc, stableForId) {
					stableSymCache[sym] = true
					return true
				}
			}
		}
		return false
	}

	// Cleanup ref.current tracking.
	currentRefsInCleanup := map[string]struct {
		dependency string
		ref        *depReference
	}{}

	var visit func(n *ast.Node) bool
	visit = func(n *ast.Node) bool {
		// Stop at nested function-likes? No — references nested inside
		// callbacks within the effect still need to be analyzed (their
		// captures are part of the effect's deps).
		if n.Kind == ast.KindIdentifier && isReferenceIdentifier(n) {
			processIdentifier(ctx, sf, tc, n, callback, componentFn, isEffect,
				setStateCallSites, stableForId, dependencies, optionalChains, currentRefsInCleanup)
		}
		n.ForEachChild(visit)
		return false
	}
	visit(body)

	// Emit the effect-cleanup ref.current warnings.
	for dep, rec := range currentRefsInCleanup {
		// If the ref is reassigned anywhere (`ref.current = ...`), upstream
		// suppresses the warning.
		if cachedHasCurrentAssignment(rec.ref.Symbol, componentFn, tc, caches) {
			continue
		}
		// `dependencyNode.parent.property` — the `.current` property name.
		// We approximate via the depReference's parent if it points to a
		// property access.
		idNode := rec.ref.Identifier
		if idNode == nil || idNode.Parent == nil || idNode.Parent.Kind != ast.KindPropertyAccessExpression {
			continue
		}
		propName := idNode.Parent.AsPropertyAccessExpression().Name()
		if propName == nil {
			continue
		}
		ctx.ReportNode(propName, rule.RuleMessage{Description: fmt.Sprintf(
			"The ref value '%s.current' will likely have changed by the time this effect cleanup function runs. "+
				"If this ref points to a node rendered by React, copy '%s.current' to a variable inside the effect, "+
				"and use that variable in the cleanup function.",
			dep, dep,
		)})
	}
}

// processIdentifier records a single identifier reference as a dep.
func processIdentifier(
	ctx rule.RuleContext,
	sf *ast.SourceFile,
	tc *checker.Checker,
	id *ast.Node,
	callback *ast.Node,
	componentFn *ast.Node,
	isEffect bool,
	setStateCallSites map[*ast.Symbol]*ast.Node,
	stableForId func(*ast.Node, *ast.Symbol) bool,
	dependencies map[string]*dependency,
	optionalChains map[string]bool,
	currentRefsInCleanup map[string]struct {
		dependency string
		ref        *depReference
	},
) {
	// Resolve.
	var sym *ast.Symbol
	if tc != nil {
		sym = tc.GetSymbolAtLocation(id)
	}
	// Determine if the symbol's declaration is in component scope (between
	// callback exclusive and componentFn inclusive).
	//
	// TC behavior nuance: for a destructured rename like
	// `const { foo: bar } = obj()`, TC may sometimes resolve `bar` to
	// the SOURCE object's property symbol rather than the local
	// BindingElement, so `sym.Declarations` is populated but every
	// declaration sits outside the component (in `obj`'s declared type).
	// We detect that exact "all declarations are outside the component
	// AND none lives inside the callback" condition and fall back to an
	// AST walk for a same-named binding inside the component. If the
	// walk finds nothing, the reference really is external (skip).
	// In any other case (declaration in callback, in component scope,
	// etc.), trust TC's verdict.
	if sym != nil {
		if isInComponentPureScope(sym, callback, componentFn) {
			// Pure scope hit — proceed.
		} else if !anyDeclWithinComponent(sym, componentFn) {
			astDecl := resolveDeclaration(nil, id, id.AsIdentifier().Text, componentFn)
			if astDecl == nil {
				return
			}
			if !containsNode(componentFn, astDecl) || containsNode(callback, astDecl) {
				return
			}
		} else {
			// TC sees declarations inside the component but the strict
			// pure-scope check excluded them (e.g. they're inside the
			// callback). Trust TC: not a dep.
			return
		}
	} else {
		// No symbol — fall back to AST walk only.
		decl := resolveDeclaration(nil, id, id.AsIdentifier().Text, componentFn)
		if decl == nil {
			return
		}
		if !containsNode(componentFn, decl) || containsNode(callback, decl) {
			return
		}
	}

	if isInsideTypePosition(id) {
		return
	}

	// Compute dep root + key.
	depRoot := getDependencyNode(id)
	key, ok := analyzePropertyChainText(depRoot, optionalChains)
	if !ok {
		return
	}

	// Skip references to the callback's own binding. Two patterns to cover:
	//
	//   (a) `const myEffect = () => {...}; useEffect(myEffect, ...)`
	//       — `myEffect`'s declaration has the callback as its initializer.
	//
	//   (b) `const foo = useCallback(() => foo(), [])`
	//       — upstream's check: `def.node.init === node.parent`. Here
	//       `node` is the callback ArrowFunction; `node.parent` is the
	//       enclosing CallExpression (`useCallback(...)`); `def.node.init`
	//       is exactly that CallExpression. So a reference to `foo` from
	//       inside the callback is the binding's own self-reference.
	if sym != nil {
		for _, decl := range sym.Declarations {
			if decl.Kind == ast.KindVariableDeclaration {
				vd := decl.AsVariableDeclaration()
				if vd.Initializer != nil {
					init := stripAsExpression(vd.Initializer)
					if init == callback {
						return
					}
					// Pattern (b): vd.Initializer is the hook call, and the
					// hook call is the callback's parent.
					if callback.Parent != nil && init == callback.Parent {
						return
					}
				}
			}
		}
	}

	// Effect cleanup `.current` tracking.
	inCleanup := false
	if isEffect && id.Parent != nil && id.Parent.Kind == ast.KindPropertyAccessExpression {
		pae := id.Parent.AsPropertyAccessExpression()
		if pae.Expression == id {
			propName := pae.Name()
			if propName != nil && propName.Kind == ast.KindIdentifier && propName.AsIdentifier().Text == "current" {
				if isInsideEffectCleanup(id, callback) {
					inCleanup = true
				}
			}
		}
	}

	// writeExpr if this is the LHS of an assignment.
	var writeExpr *ast.Node
	if id.Parent != nil {
		if be, ok := getAssignmentBinaryExpr(id.Parent); ok && be.Left == id {
			writeExpr = id.Parent
		}
	}

	stable := stableForId(id, sym)
	dep, exists := dependencies[key]
	if !exists {
		dep = &dependency{IsStable: stable, First: id}
		dependencies[key] = dep
	}
	ref := &depReference{
		Identifier:  id,
		Symbol:      sym,
		WriteExpr:   writeExpr,
		InCleanup:   inCleanup,
		DepNodeRoot: depRoot,
	}
	dep.Refs = append(dep.Refs, ref)

	if inCleanup {
		if _, has := currentRefsInCleanup[key]; !has {
			currentRefsInCleanup[key] = struct {
				dependency string
				ref        *depReference
			}{dependency: key, ref: ref}
		}
	}
}

// anyDeclWithinComponent reports whether the symbol has at least one
// declaration whose source range lies inside `componentFn`. Used as the
// gate for the AST-walk fallback in `processIdentifier`: when none of
// the symbol's declarations is local to the component, TC has resolved
// to an external (type-property / cross-module) symbol and we may
// safely consult the AST for a same-named local binding.
func anyDeclWithinComponent(sym *ast.Symbol, componentFn *ast.Node) bool {
	if sym == nil || componentFn == nil {
		return false
	}
	for _, d := range sym.Declarations {
		if d == nil {
			continue
		}
		if containsNode(componentFn, d) {
			return true
		}
	}
	return false
}

// isInComponentPureScope reports whether the symbol's declaration sits
// strictly between the callback (exclusive) and the component body
// (inclusive). External symbols (imports, globals) are excluded.
func isInComponentPureScope(sym *ast.Symbol, callback *ast.Node, componentFn *ast.Node) bool {
	if sym == nil {
		return false
	}
	if len(sym.Declarations) == 0 {
		return false
	}
	for _, decl := range sym.Declarations {
		if decl == nil {
			continue
		}
		if !containsNode(componentFn, decl) {
			continue
		}
		if containsNode(callback, decl) {
			continue
		}
		// Skip type-only declarations.
		switch decl.Kind {
		case ast.KindTypeAliasDeclaration, ast.KindInterfaceDeclaration,
			ast.KindTypeParameter:
			continue
		}
		return true
	}
	return false
}

// hasCurrentAssignment scans `scope` (typically the surrounding component)
// for `<sym>.current = ...`. When TypeChecker is available, compares
// symbols; otherwise falls back to a name match against the symbol's
// binding name. Mirrors upstream's "found a `<ref>.current = ...` write"
// gate that suppresses the cleanup warning when the user is the one
// managing the ref.
//
// `scope` bounds the walk — pass `componentFn` (or `sf.AsNode()` if no
// component scope is available). The previous full-SF walk was an
// O(file_size) per-cleanup-ref hot path that PR-808 review flagged.
func hasCurrentAssignment(sym *ast.Symbol, scope *ast.Node, tc *checker.Checker) bool {
	if sym == nil || scope == nil {
		return false
	}
	targetName := ""
	if len(sym.Declarations) > 0 {
		if n := sym.Declarations[0].Name(); n != nil && n.Kind == ast.KindIdentifier {
			targetName = n.AsIdentifier().Text
		}
	}
	found := false
	var visit func(n *ast.Node) bool
	visit = func(n *ast.Node) bool {
		if found {
			return true
		}
		if n.Kind == ast.KindBinaryExpression {
			be := n.AsBinaryExpression()
			if be.OperatorToken != nil && be.OperatorToken.Kind == ast.KindEqualsToken {
				lhs := ast.SkipParentheses(be.Left)
				if lhs.Kind == ast.KindPropertyAccessExpression {
					pae := lhs.AsPropertyAccessExpression()
					prop := pae.Name()
					if prop != nil && prop.Kind == ast.KindIdentifier && prop.AsIdentifier().Text == "current" {
						obj := ast.SkipParentheses(pae.Expression)
						if obj.Kind == ast.KindIdentifier {
							if tc != nil {
								if s := tc.GetSymbolAtLocation(obj); s != nil && s == sym {
									found = true
									return true
								}
							} else if targetName != "" && obj.AsIdentifier().Text == targetName {
								found = true
								return true
							}
						}
					}
				}
			}
		}
		n.ForEachChild(visit)
		return false
	}
	visit(scope)
	return found
}

// isFunctionWithoutCapturedValues mirrors upstream's same-named helper.
// Given a VariableDeclaration / FunctionDeclaration / BindingElement
// declaring a function-typed binding, returns true iff the function body
// does not capture any non-stable variables from the component pure scope.
//
// Implementation strategy:
//
//   - Resolve every reference Identifier inside the function body via
//     TypeChecker. The resolved symbol's declaration tells us which scope
//     the reference belongs to.
//   - A reference is a "capture" iff its declaration sits in component
//     scope (inside componentFn) but outside the function being tested.
//   - We're stable iff every capture is itself a stable hook value.
//
// `isStable` is the recursive `stableForId` so a chain of pure functions
// (`const a = () => 1; const b = () => a()`) classifies as stable.
//
// Without a TypeChecker we conservatively return false (callers degrade
// to the un-stabilized behavior).
func isFunctionWithoutCapturedValues(
	decl *ast.Node,
	componentFn *ast.Node,
	tc *checker.Checker,
	isStable func(*ast.Node, *ast.Symbol) bool,
) bool {
	if decl == nil || componentFn == nil || tc == nil {
		return false
	}
	// Locate the function-like body associated with this declaration.
	var fnNode *ast.Node
	switch decl.Kind {
	case ast.KindFunctionDeclaration:
		fnNode = decl
	case ast.KindVariableDeclaration:
		init := decl.AsVariableDeclaration().Initializer
		if init == nil {
			return false
		}
		init = stripAsExpression(init)
		switch init.Kind {
		case ast.KindArrowFunction, ast.KindFunctionExpression:
			fnNode = init
		}
	case ast.KindBindingElement:
		init := decl.AsBindingElement().Initializer
		if init == nil {
			return false
		}
		init = stripAsExpression(init)
		switch init.Kind {
		case ast.KindArrowFunction, ast.KindFunctionExpression:
			fnNode = init
		}
	default:
		return false
	}
	if fnNode == nil {
		return false
	}
	// `fnNode` must itself live inside the component scope; otherwise the
	// "captures from pure scope" question is not well-defined.
	if !containsNode(componentFn, fnNode) {
		return false
	}
	body := getCallbackBody(fnNode)
	if body == nil {
		return false
	}
	captured := false
	var visit func(n *ast.Node) bool
	visit = func(n *ast.Node) bool {
		if captured {
			return true
		}
		if n.Kind == ast.KindIdentifier && isReferenceIdentifier(n) {
			sym := tc.GetSymbolAtLocation(n)
			if sym != nil && len(sym.Declarations) > 0 {
				d := sym.Declarations[0]
				// Capture iff declaration is in component scope but
				// outside fnNode itself. Self-reference (the function
				// referencing its own binding) is allowed.
				if containsNode(componentFn, d) && !containsNode(fnNode, d) {
					// Self-reference (recursion) — skip.
					if d == decl {
						return false
					}
					if !isStable(n, sym) {
						captured = true
						return true
					}
				}
			}
		}
		n.ForEachChild(visit)
		return false
	}
	visit(body)
	return !captured
}

