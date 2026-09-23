package exhaustive_deps

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/react_hooksutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

//go:embed exhaustive_deps.schema.json
var schemaJSON []byte

type suggestionBuilder func() []rule.RuleSuggestion

type suggestionReporter struct {
	ctx              *rule.RuleContext
	dangerousAutofix bool
}

// newSuggestionReporter is the rule-local adapter between upstream's
// suggestion-first reporting model and rslint's independently demanded edit
// categories. Dangerous autofix mode promotes exactly the first suggestion's
// complete edit set while memoizing the shared build for EditDemandAll.
func newSuggestionReporter(ctx *rule.RuleContext, dangerousAutofix bool) *suggestionReporter {
	return &suggestionReporter{
		ctx:              ctx,
		dangerousAutofix: dangerousAutofix,
	}
}

func (reporter *suggestionReporter) report(
	node *ast.Node,
	description string,
	buildSuggestions suggestionBuilder,
) {
	node = utils.ESTreeRuntimeExpression(node)
	message := rule.RuleMessage{Description: description}
	if buildSuggestions == nil {
		reporter.ctx.ReportNode(node, message)
		return
	}
	if !reporter.dangerousAutofix {
		reporter.ctx.ReportNodeWithDeferredSuggestions(node, message, buildSuggestions)
		return
	}
	reporter.reportDangerous(node, message, buildSuggestions)
}

func (reporter *suggestionReporter) reportDangerous(
	node *ast.Node,
	message rule.RuleMessage,
	buildSuggestions suggestionBuilder,
) {
	var suggestions []rule.RuleSuggestion
	suggestionsBuilt := false
	getSuggestions := func() []rule.RuleSuggestion {
		if !suggestionsBuilt {
			suggestions = buildSuggestions()
			suggestionsBuilt = true
		}
		return suggestions
	}
	reporter.ctx.ReportNodeWithDeferredFixesAndSuggestions(
		node,
		message,
		func() []rule.RuleFix {
			built := getSuggestions()
			if len(built) == 0 || len(built[0].FixesArr) == 0 {
				return nil
			}
			// Mirrors upstream's `problem.fix =
			// problem.suggest[0].fix`: promote the first
			// suggestion's complete edit set (including iterator fixes).
			return built[0].FixesArr
		},
		getSuggestions,
	)
}

func buildDependencyArraySuggestion(
	sf *ast.SourceFile,
	depsNode *ast.Node,
	suggestedDeps []string,
	alphabetized bool,
	optionalChains map[string]bool,
) []rule.RuleSuggestion {
	if alphabetized {
		slices.SortFunc(suggestedDeps, ecmascript.CompareStrings)
	}

	formatted := make([]string, len(suggestedDeps))
	for i, key := range suggestedDeps {
		formatted[i] = formatDependency(key, optionalChains)
	}
	suggestionText := "[" + strings.Join(formatted, ", ") + "]"
	return []rule.RuleSuggestion{{
		Message:  rule.RuleMessage{Description: "Update the dependencies array to be: " + suggestionText},
		FixesArr: []rule.RuleFix{rule.RuleFixReplace(sf, utils.ESTreeRuntimeExpression(depsNode), suggestionText)},
	}}
}

// ExhaustiveDepsRule preserves upstream's lexical scope semantics. All caches
// belong to one file and edits remain lazy for diagnostic-only consumers.
var ExhaustiveDepsRule = rule.Rule{
	Name:   "react-hooks/exhaustive-deps",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options, ctx.Settings)
		caches := &runCaches{ctx: &ctx}
		reporter := newSuggestionReporter(&ctx, opts.EnableDangerousAutofixThisMayCauseInfiniteLoops)
		return rule.RuleListeners{ast.KindCallExpression: func(node *ast.Node) { visitCall(&ctx, opts, caches, node, reporter) }}
	},
}

func visitCall(ctx *rule.RuleContext, opts Options, caches *runCaches, node *ast.Node, reports *suggestionReporter) {
	call := node.AsCallExpression()
	callee := utils.ESTreeCallCallee(call.Expression)
	cbIndex := getReactiveHookCallbackIndex(callee, opts.AdditionalHooks)
	if cbIndex < 0 {
		return
	}
	var args []*ast.Node
	if call.Arguments != nil {
		args = call.Arguments.Nodes
	}
	name := reactiveHookName(callee)
	if cbIndex >= len(args) {
		reports.report(callee, fmt.Sprintf("React Hook %s requires an effect callback. Did you forget to pass a callback to the hook?", name), nil)
		return
	}
	callback := stripAsExpression(args[cbIndex])
	isEffect := react_hooksutil.IsEffectStyleHookName(name)
	var deps *ast.Node
	if cbIndex+1 < len(args) && !hasUndefinedIdentifier(args[cbIndex+1]) {
		deps = utils.ESTreeRuntimeExpression(args[cbIndex+1])
	}
	if cbIndex+1 >= len(args) && isEffect && opts.RequireExplicitEffectDeps {
		reports.report(callee, fmt.Sprintf("React Hook %s always requires dependencies. Please add a dependency array or an explicit `undefined`", name), nil)
	}
	auto := opts.AutoDepsHooks[name]
	noDeps := deps == nil || auto && deps.Kind == ast.KindNullKeyword
	if noDeps && !isEffect {
		if name == "useMemo" || name == "useCallback" {
			reports.report(callee, fmt.Sprintf("React Hook %s does nothing when called with only one argument. Did you forget to pass an array of dependencies?", name), nil)
		}
		return
	}
	switch callback.Kind {
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		visitFunctionWithDependencies(ctx, opts, caches, callback, deps, callee, name, isEffect, reports)
		return
	case ast.KindIdentifier:
		if noDeps {
			return
		}
		if deps.Kind == ast.KindArrayLiteralExpression {
			for _, element := range deps.AsArrayLiteralExpression().Elements.Nodes {
				element = utils.ESTreeRuntimeExpression(element)
				if element != nil && element.Kind == ast.KindIdentifier && element.Text() == callback.Text() {
					return
				}
			}
		}
		// Upstream consults the call's own scope, not enclosing scopes. An
		// unresolved/imported callback is not assumed to be a render closure.
		variable := firstVariable(caches.scopes().Acquire(callback), callback.Text())
		if variable == nil {
			return
		}
		if variable.Kind == scope.DefParameter {
			reports.report(callee, getUnknownDependenciesMessage(name), nil)
			return
		}
		if variable.Kind == scope.DefFunctionName && variable.DefNode.Kind == ast.KindFunctionDeclaration {
			visitFunctionWithDependencies(ctx, opts, caches, variable.DefNode, deps, callee, name, isEffect, reports)
			return
		}
		if decl := variableDeclaration(variable); decl != nil && decl.AsVariableDeclaration().Initializer != nil {
			init := utils.ESTreeRuntimeExpression(decl.AsVariableDeclaration().Initializer)
			if init != nil && (init.Kind == ast.KindArrowFunction || init.Kind == ast.KindFunctionExpression) {
				visitFunctionWithDependencies(ctx, opts, caches, init, deps, callee, name, isEffect, reports)
				return
			}
		}
		reports.report(callee, fmt.Sprintf("React Hook %s has a missing dependency: '%s'. Either include it or remove the dependency array.", name, callback.Text()), func() []rule.RuleSuggestion {
			return buildDependencyArraySuggestion(ctx.SourceFile, deps, []string{callback.Text()}, false, nil)
		})
	default:
		reports.report(callee, getUnknownDependenciesMessage(name), nil)
	}
}

func visitFunctionWithDependencies(ctx *rule.RuleContext, opts Options, caches *runCaches, callback, deps, callee *ast.Node, name string, isEffect bool, reports *suggestionReporter) {
	if isEffect && react_hooksutil.HasAsyncModifier(callback) {
		reports.report(callback, "Effect callbacks are synchronous to prevent race conditions. "+
			"Put the async function inside:\n\n"+
			"useEffect(() => {\n"+
			"  async function fetchData() {\n"+
			"    // You can await here\n"+
			"    const response = await MyAPI.getData(someId);\n"+
			"    // ...\n"+
			"  }\n"+
			"  fetchData();\n"+
			"}, [someId]); // Or [] if effect doesn't need props or state\n\n"+
			"Learn more about data fetching with Hooks: https://react.dev/link/hooks-data-fetching", nil)
	}
	callbackScope := caches.scopes().Acquire(callback)
	// acquire(node) in ESLint defaults to the outer name scope of a named
	// function expression. Its function scope remains a child for traversal.
	if callbackScope.Parent != nil && callbackScope.Parent.Kind == scope.KindFunctionExprName && callbackScope.Parent.Block == callback {
		callbackScope = callbackScope.Parent
	}
	pure := map[*scope.Scope]bool{}
	component := callbackScope.Parent
	for component != nil {
		pure[component] = true
		if component.Kind == scope.KindFunction {
			break
		}
		component = component.Parent
	}
	if component == nil {
		return
	}
	dependencies, optional := caches.gather(callback, callbackScope, component, pure, isEffect)
	stable := map[string]bool{}
	stale := false
	for key := range dependencies.Keys() {
		dep := dependencies.GetOrZero(key)
		if dep.IsStable {
			stable[key] = true
		}
		for _, ref := range dep.Refs {
			if ref.WriteExpr == nil {
				continue
			}
			reports.report(ref.WriteExpr, fmt.Sprintf("Assignments to the '%s' variable from inside React Hook %s will be lost after each render. To preserve the value over time, store it in a useRef Hook and keep the mutable value in the '.current' property. Otherwise, you can move this variable directly inside %s.", key, getCalleeText(ctx.SourceFile, callee), getCalleeText(ctx.SourceFile, callee)), nil)
			stale = true
			break
		}
	}
	if stale {
		return
	}
	if deps == nil {
		if !opts.AutoDepsHooks[name] {
			emitSetStateInsideEffectWarning(caches, callback, callee, name, dependencies, stable, reports)
		}
		return
	}
	if opts.AutoDepsHooks[name] && deps.Kind == ast.KindNullKeyword {
		return
	}
	declared, elements := parseDeclaredDeps(ctx, ctx.SourceFile, callee, deps, dependencies, caches.effectEvents)
	external := map[string]bool{}
	for _, dd := range declared {
		root := utils.ESTreeRuntimeExpression(dd.Node)
		for root != nil {
			if root.Kind == ast.KindPropertyAccessExpression {
				root = utils.ESTreeRuntimeExpression(root.AsPropertyAccessExpression().Expression)
				continue
			}
			if root.Kind == ast.KindElementAccessExpression {
				root = utils.ESTreeRuntimeExpression(root.AsElementAccessExpression().Expression)
				continue
			}
			break
		}
		if ref := caches.referenceAt(root); ref != nil {
			if v := ref.Resolved(); v == nil || !withinScope(v.Scope, component) {
				external[dd.Key] = true
			}
		}
	}
	rec := collectRecommendations(dependencies, declared, stable, external, isEffect)
	if len(rec.Missing)+len(rec.Unnecessary)+len(rec.Duplicate) == 0 {
		emitConstructionWarnings(ctx, caches, callbackScope, component, deps, declared, name, reports)
		flushDeferredDiagnostics(elements, reports)
		return
	}
	message := buildDepDiagnostic(ctx, caches, callbackScope, component, callee, name, rec, declared, external, dependencies, optional)
	reports.report(deps, message, func() []rule.RuleSuggestion {
		suggested := rec.Suggested
		if !isEffect && len(rec.Missing) > 0 {
			suggested = collectRecommendations(dependencies, nil, stable, external, false).Suggested
		}
		return buildDependencyArraySuggestion(ctx.SourceFile, deps, suggested, areDeclaredDepsAlphabetized(declared), optional)
	})
	flushDeferredDiagnostics(elements, reports)
}

func buildDepDiagnostic(ctx *rule.RuleContext, caches *runCaches, callbackScope, component *scope.Scope, callee *ast.Node, name string, rec recommendations, declared []declaredDependency, external map[string]bool, dependencies *dependencyMap, optional map[string]bool) string {
	body := getWarningMessage(rec.Missing, "a", "missing", "include", optional)
	if body == "" {
		body = getWarningMessage(rec.Unnecessary, "an", "unnecessary", "exclude", optional)
	}
	if body == "" {
		body = getWarningMessage(rec.Duplicate, "a", "duplicate", "omit", optional)
	}
	extra := ""
	if len(rec.Unnecessary) > 0 {
		for _, dd := range declared {
			if rec.Unnecessary[dd.Key] && strings.HasSuffix(dd.Key, ".current") {
				extra = fmt.Sprintf(" Mutable values like '%s' aren't valid dependencies because mutating them doesn't re-render the component.", dd.Key)
				break
			}
		}
		if extra == "" {
			for _, dd := range declared {
				if external[dd.Key] {
					if firstVariable(callbackScope, dd.Key) == nil {
						extra = fmt.Sprintf(" Outer scope values like '%s' aren't valid dependencies because mutating them doesn't re-render the component.", dd.Key)
					}
					break
				}
			}
		}
	}
	if extra == "" && rec.Missing["props"] {
		onlyMembers := true
		for _, ref := range dependencies.GetOrZero("props").Refs {
			p := utils.ESTreeParent(ref.Identifier)
			if p == nil || p.Kind != ast.KindPropertyAccessExpression && p.Kind != ast.KindElementAccessExpression {
				onlyMembers = false
				break
			}
		}
		if onlyMembers {
			extra = fmt.Sprintf(" However, 'props' will change when *any* prop changes, so the preferred fix is to destructure the 'props' object outside of the %s call and refer to those specific props inside %s.", name, getCalleeText(ctx.SourceFile, callee))
		}
	}
	if extra == "" {
		for _, key := range rec.MissingOrder {
			used := dependencies.GetOrZero(key)
			v := firstVariable(component, key)
			if used == nil || v == nil || v.Kind != scope.DefParameter || used.Refs[0].Resolved() != v {
				continue
			}
			for _, ref := range used.Refs {
				p := utils.ESTreeParent(ref.Identifier)
				if p != nil && p.Kind == ast.KindCallExpression && utils.ESTreeRuntimeExpression(p.AsCallExpression().Expression) == ref.Identifier {
					extra = fmt.Sprintf(" If '%s' changes too often, find the parent component that defines it and wrap that definition in useCallback.", key)
					break
				}
			}
			if extra != "" {
				break
			}
		}
	}
	if extra == "" {
		extra = setStateRecommendation(caches, rec.MissingOrder, dependencies, component.Block)
	}
	return fmt.Sprintf("React Hook %s has %s%s", getCalleeText(ctx.SourceFile, callee), body, extra)
}

func setStateRecommendation(caches *runCaches, missing []string, dependencies *dependencyMap, component *ast.Node) string {
	for _, key := range missing {
		dep := dependencies.GetOrZero(key)
		if dep == nil {
			continue
		}
		for _, ref := range dep.Refs {
			for current := ref.Identifier.Parent; current != nil && current != component; current = current.Parent {
				if current.Kind != ast.KindCallExpression {
					continue
				}
				callee := utils.ESTreeRuntimeExpression(current.AsCallExpression().Expression)
				state := caches.setStateCallSites[callee]
				if state == nil {
					continue
				}
				setter := callee.Text()
				if state.Kind == ast.KindIdentifier && state.Text() == key {
					// Keep an astral identifier intact; slicing its first UTF-16
					// unit as upstream does would create an unpaired surrogate.
					first, _ := ecmascript.DecodeStringRune(key)
					initial := string(first)
					return fmt.Sprintf(" You can also do a functional update '%s(%s => ...)' if you only need '%s' in the '%s' call.", setter, initial, key, setter)
				}
				if caches.stateVariables[ref.Identifier] {
					return fmt.Sprintf(" You can also replace multiple useState variables with useReducer if '%s' needs the current value of '%s'.", setter, key)
				}
				if ref.Resolved().Kind == scope.DefParameter {
					return fmt.Sprintf(" If '%s' needs the current value of '%s', you can also switch to useReducer instead of useState and read '%s' in the reducer.", setter, key, key)
				}
				break
			}
		}
	}
	return ""
}

func emitSetStateInsideEffectWarning(caches *runCaches, callback, callee *ast.Node, name string, dependencies *dependencyMap, stable map[string]bool, reports *suggestionReporter) {
	for key := range dependencies.Keys() {
		for _, ref := range dependencies.GetOrZero(key).Refs {
			if _, ok := caches.setStateCallSites[ref.Identifier]; !ok {
				continue
			}
			from := ref.From
			for from != nil && from.Kind != scope.KindFunction {
				from = from.Parent
			}
			if from == nil || from.Block != callback {
				continue
			}
			rec := collectRecommendations(dependencies, nil, stable, nil, true)
			text := "[" + strings.Join(rec.Suggested, ", ") + "]"
			reports.report(callee, fmt.Sprintf("React Hook %s contains a call to '%s'. Without a list of dependencies, this can lead to an infinite chain of updates. To fix this, pass %s as a second argument to the %s Hook.", name, key, text, name), func() []rule.RuleSuggestion {
				return []rule.RuleSuggestion{{Message: rule.RuleMessage{Description: "Add dependencies array: " + text}, FixesArr: []rule.RuleFix{rule.RuleFixInsertAfter(callback, ", "+text)}}}
			})
			return
		}
	}
}

func emitConstructionWarnings(ctx *rule.RuleContext, caches *runCaches, callbackScope, component *scope.Scope, deps *ast.Node, declared []declaredDependency, name string, reports *suggestionReporter) {
	line, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(ctx.SourceFile, utils.TrimNodeTextRange(ctx.SourceFile, deps).Pos())
	for _, dd := range declared {
		v := firstVariable(component, dd.Key)
		if v == nil {
			continue
		}
		decl := v.DefNode
		kind := ""
		var init *ast.Node
		if d := variableDeclaration(v); d != nil && d.AsVariableDeclaration().Name().Kind == ast.KindIdentifier {
			decl = d
			init = d.AsVariableDeclaration().Initializer
			kind = constructionType(init)
		} else if v.Kind == scope.DefFunctionName && decl.Kind == ast.KindFunctionDeclaration {
			kind = "function"
		} else if v.Kind == scope.DefClassName && decl.Kind == ast.KindClassDeclaration {
			kind = "class"
		}
		if kind == "" {
			continue
		}
		wrapper, construction := "useMemo", "initialization"
		if kind == "function" {
			wrapper, construction = "useCallback", "definition"
		}
		defaultAdvice := fmt.Sprintf("wrap the %s of '%s' in its own %s() Hook.", construction, v.Name, wrapper)
		outside := caches.usedOutside(v, callbackScope, deps)
		advice := "To fix this, " + defaultAdvice
		if !outside {
			advice = fmt.Sprintf("Move it inside the %s callback. Alternatively, %s", name, defaultAdvice)
		}
		causation := "makes"
		if kind == "conditional" || kind == "logical expression" {
			causation = "could make"
		}
		message := fmt.Sprintf("The '%s' %s %s the dependencies of %s Hook (at line %d) change on every render. %s", v.Name, kind, causation, name, line+1, advice)
		var suggestions suggestionBuilder
		if outside && init != nil && kind == "function" {
			suggestions = func() []rule.RuleSuggestion {
				target := utils.ESTreeRuntimeExpression(init)
				return []rule.RuleSuggestion{{Message: rule.RuleMessage{Description: fmt.Sprintf("Wrap the %s of '%s' in its own %s() Hook.", construction, v.Name, wrapper)}, FixesArr: []rule.RuleFix{rule.RuleFixInsertBefore(ctx.SourceFile, target, "useCallback("), rule.RuleFixInsertAfter(target, ")")}}}
			}
		}
		reports.report(decl, message, suggestions)
	}
}

type elementDiagnostic struct {
	node    *ast.Node
	message string
	// suggestionMessage is non-empty only for the useEffectEvent variant.
	// Its range and edit remain deferred until the runtime requests suggestions
	// or dangerous autofixes.
	suggestionMessage string
}

// parseDeclaredDeps collects declared paths and per-element diagnostics.
// Malformed dependency lists still participate in missing-dependency checks.
func parseDeclaredDeps(
	ctx *rule.RuleContext,
	sf *ast.SourceFile,
	reactiveHook *ast.Node,
	depsNode *ast.Node,
	dependencies *dependencyMap,
	useEffectEventVariables map[*ast.Node]bool,
) (declared []declaredDependency, deferred []elementDiagnostic) {
	depsNodeStripped := utils.ESTreeRuntimeExpression(depsNode)
	if depsNodeStripped.Kind == ast.KindAsExpression {
		depsNodeStripped = utils.ESTreeRuntimeExpression(depsNodeStripped.AsAsExpression().Expression)
	}
	if depsNodeStripped == nil || depsNodeStripped.Kind != ast.KindArrayLiteralExpression {
		// Upstream still emits the missing-dep diagnostic afterwards
		// using an empty declared-deps list.
		// The "not an array literal" error itself is still reported.
		ctx.ReportNode(depsNode, rule.RuleMessage{Description: fmt.Sprintf(
			"React Hook %s was passed a dependency list that is not an array literal. "+
				"This means we can't statically verify whether you've passed the correct dependencies.",
			getCalleeText(sf, reactiveHook),
		)})
		return nil, nil
	}
	arr := depsNodeStripped.AsArrayLiteralExpression()
	if arr.Elements == nil {
		return nil, nil
	}
	for _, el := range arr.Elements.Nodes {
		el = utils.ESTreeRuntimeExpression(el)
		if el == nil || el.Kind == ast.KindOmittedExpression {
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
		if useEffectEventVariables[el] {
			dependencyText := getCalleeText(sf, el)
			deferred = append(deferred, elementDiagnostic{
				node:              el,
				message:           fmt.Sprintf("Functions returned from `useEffectEvent` must not be included in the dependency array. Remove `%s` from the list.", dependencyText),
				suggestionMessage: fmt.Sprintf("Remove the dependency `%s`", dependencyText),
			})
		}

		key, ok := analyzeDepsArrayElement(el, nil)
		if !ok {
			// Literal value or complex expression. Mirrors upstream's full
			// set of `Literal` value kinds — string, number, bigint, null,
			// boolean, regex. Each gets the "literal is not a valid dep"
			// diagnostic; everything else falls through to "complex
			// expression".
			elInner := utils.ESTreeRuntimeExpression(el)
			switch elInner.Kind {
			case ast.KindStringLiteral:
				val := elInner.AsStringLiteral().Text
				var msg string
				if dependencies.GetOrZero(val) != nil {
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
	return declared, deferred
}

// flushDeferredDiagnostics emits the per-element diagnostics queued by
// parseDeclaredDeps after the main dependency diagnostic.
//
// Suggestion construction is routed through the same reporter as the
// top-level dependency diagnostic. This keeps demand gating and dangerous
// autofix promotion identical for every upstream `reportProblem` call.
func flushDeferredDiagnostics(
	deferred []elementDiagnostic,
	suggestionReports *suggestionReporter,
) {
	for _, d := range deferred {
		node := d.node
		if d.suggestionMessage == "" {
			suggestionReports.report(node, d.message, nil)
			continue
		}

		suggestionMessage := d.suggestionMessage
		suggestionReports.report(node, d.message, func() []rule.RuleSuggestion {
			return []rule.RuleSuggestion{{
				Message: rule.RuleMessage{Description: suggestionMessage},
				FixesArr: []rule.RuleFix{
					rule.RuleFixRemoveRange(utils.TrimNodeTextRange(suggestionReports.ctx.SourceFile, node)),
				},
			}}
		})
	}
}
