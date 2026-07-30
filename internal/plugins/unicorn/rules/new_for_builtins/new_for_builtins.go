package new_for_builtins

import (
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	messageIDEnforce             = "enforce"
	messageIDDisallow            = "disallow"
	messageIDDisallowCallOrNew   = "disallowCallOrNew"
	messageIDErrorDate           = "error-date"
	messageIDSuggestionDate      = "suggestion-date"
	messageDateDescription       = "Use `String(new Date())` instead of `Date()`."
	messageDateSuggestion        = "Switch to `String(new Date())`."
	replacementStringNewDateCall = "String(new Date())"
)

var NewForBuiltinsRule = rule.Rule{
	Name:   "unicorn/new-for-builtins",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if !sourceHasPotentialReference(ctx.SourceFile) {
			return nil
		}

		state := newRuleState(ctx)
		collectNamedLocalRoot := state.collectNamedLocalRoot

		return rule.RuleListeners{
			ast.KindCallExpression:                 state.collectCallExpression,
			ast.KindNewExpression:                  state.collectNewExpression,
			ast.KindVariableDeclaration:            state.collectAliasDeclaration,
			ast.KindClassExpression:                collectNamedLocalRoot,
			ast.KindClassDeclaration:               collectNamedLocalRoot,
			ast.KindFunctionExpression:             collectNamedLocalRoot,
			ast.KindFunctionDeclaration:            collectNamedLocalRoot,
			ast.KindEnumDeclaration:                collectNamedLocalRoot,
			ast.KindModuleDeclaration:              collectNamedLocalRoot,
			rule.ListenerOnExit(ast.KindEndOfFile): state.finish,
		}
	},
}

var enforceNewBuiltins = map[string]bool{
	"Object":                   true,
	"Array":                    true,
	"ArrayBuffer":              true,
	"DataView":                 true,
	"Date":                     true,
	"Function":                 true,
	"Map":                      true,
	"WeakMap":                  true,
	"Set":                      true,
	"WeakSet":                  true,
	"Promise":                  true,
	"RegExp":                   true,
	"SharedArrayBuffer":        true,
	"Proxy":                    true,
	"WeakRef":                  true,
	"FinalizationRegistry":     true,
	"DisposableStack":          true,
	"AsyncDisposableStack":     true,
	"Error":                    true,
	"AggregateError":           true,
	"EvalError":                true,
	"RangeError":               true,
	"ReferenceError":           true,
	"SuppressedError":          true,
	"SyntaxError":              true,
	"TypeError":                true,
	"URIError":                 true,
	"Float16Array":             true,
	"Float32Array":             true,
	"Float64Array":             true,
	"Int8Array":                true,
	"Int16Array":               true,
	"Int32Array":               true,
	"BigInt64Array":            true,
	"BigUint64Array":           true,
	"Uint8Array":               true,
	"Uint16Array":              true,
	"Uint32Array":              true,
	"Uint8ClampedArray":        true,
	"Intl.Collator":            true,
	"Intl.DateTimeFormat":      true,
	"Intl.DisplayNames":        true,
	"Intl.DurationFormat":      true,
	"Intl.ListFormat":          true,
	"Intl.Locale":              true,
	"Intl.NumberFormat":        true,
	"Intl.PluralRules":         true,
	"Intl.RelativeTimeFormat":  true,
	"Intl.Segmenter":           true,
	"Temporal.Duration":        true,
	"Temporal.Instant":         true,
	"Temporal.PlainDate":       true,
	"Temporal.PlainDateTime":   true,
	"Temporal.PlainMonthDay":   true,
	"Temporal.PlainTime":       true,
	"Temporal.PlainYearMonth":  true,
	"Temporal.ZonedDateTime":   true,
	"WebAssembly.Module":       true,
	"WebAssembly.Instance":     true,
	"WebAssembly.Memory":       true,
	"WebAssembly.Table":        true,
	"WebAssembly.Global":       true,
	"WebAssembly.Tag":          true,
	"WebAssembly.Exception":    true,
	"WebAssembly.CompileError": true,
	"WebAssembly.LinkError":    true,
	"WebAssembly.RuntimeError": true,
}

var disallowNewBuiltins = map[string]bool{
	"BigInt":  true,
	"Boolean": true,
	"Number":  true,
	"String":  true,
	"Symbol":  true,
}

var disallowCallOrNewBuiltins = map[string]bool{
	"Temporal.Now":      true,
	"WebAssembly":       true,
	"WebAssembly.JSTag": true,
}

var globalObjectNames = map[string]bool{
	"globalThis": true,
	"global":     true,
	"self":       true,
	"window":     true,
}

// potentialReferenceRoots is the cheap listener-side gate. Only these global
// roots, plus names registered in aliasesByID, can ever produce a watched
// builtin path.
var potentialReferenceRoots = func() map[string]bool {
	roots := make(map[string]bool)
	add := func(names map[string]bool) {
		for name := range names {
			root, _, _ := strings.Cut(name, ".")
			roots[root] = true
		}
	}
	add(enforceNewBuiltins)
	add(disallowNewBuiltins)
	add(disallowCallOrNewBuiltins)
	add(globalObjectNames)
	return roots
}()

// watchedNamespaces lets member calls reject non-constructor paths before
// joining them into a lookup key.
var watchedNamespaces = func() map[string]bool {
	namespaces := make(map[string]bool)
	add := func(names map[string]bool) {
		for name := range names {
			for index := strings.LastIndexByte(name, '.'); index >= 0; index = strings.LastIndexByte(name[:index], '.') {
				namespaces[name[:index]] = true
			}
		}
	}
	add(enforceNewBuiltins)
	add(disallowNewBuiltins)
	add(disallowCallOrNewBuiltins)
	return namespaces
}()

// Descriptions are immutable and depend only on the builtin name. Keep the
// per-diagnostic Data map independent, but avoid repeating fmt work for every
// matching call in large/generated files.
var (
	enforceMessageDescriptions = buildMessageDescriptions(enforceNewBuiltins, func(name string) string {
		return fmt.Sprintf("Use `new %s()` instead of `%s()`.", name, name)
	})
	disallowMessageDescriptions = buildMessageDescriptions(disallowNewBuiltins, func(name string) string {
		return fmt.Sprintf("Use `%s()` instead of `new %s()`.", name, name)
	})
	disallowCallOrNewMessageDescriptions = buildMessageDescriptions(disallowCallOrNewBuiltins, func(name string) string {
		return fmt.Sprintf("`%s` is not a function or constructor.", name)
	})
)

type ruleState struct {
	ctx               rule.RuleContext
	aliases           map[*ast.Node]aliasInfo
	aliasesByID       map[string][]aliasInfo
	collectingAliases bool

	// Alias declarations and possibly aliased calls are collected by the
	// linter's existing traversal, then resolved in source order at EOF.
	aliasDeclarations         []*ast.Node
	pendingExpressions        []*ast.Node
	deferRemainingExpressions bool

	// localRootNames is a cheap file-wide "may be shadowed" gate. Exact
	// binding identity still comes from ctx.Refs at the call site.
	localRootNames map[string]bool
}

type aliasInfo struct {
	binding *ast.Node
	scope   *ast.Node
	path    []string
}

type reference struct {
	name string
}

type referenceSource uint8

const (
	referenceSourceBareGlobal referenceSource = iota
	referenceSourceGlobalObject
	referenceSourceAlias
)

type globalReference struct {
	path   []string
	source referenceSource
}

func newRuleState(ctx rule.RuleContext) *ruleState {
	state := &ruleState{
		ctx:               ctx,
		collectingAliases: true,
	}
	state.collectLocalRootNames()
	return state
}

func (state *ruleState) collectLocalRootNames() {
	if state.ctx.SourceFile == nil {
		return
	}
	// The binder links lexical containers in declaration order. Walking their
	// symbol-table names is much cheaper than walking every AST node.
	collect := func(symbols ast.SymbolTable) {
		for name := range symbols {
			if potentialReferenceRoots[name] {
				if state.localRootNames == nil {
					state.localRootNames = map[string]bool{}
				}
				state.localRootNames[name] = true
			}
		}
	}
	for container := state.ctx.SourceFile.AsNode(); container != nil; {
		collect(container.Locals())
		if symbol := container.Symbol(); symbol != nil {
			collect(symbol.Exports)
		}
		if name := container.Name(); name != nil && name.Kind == ast.KindIdentifier &&
			potentialReferenceRoots[name.Text()] {
			if state.localRootNames == nil {
				state.localRootNames = map[string]bool{}
			}
			state.localRootNames[name.Text()] = true
		}
		data := container.LocalsContainerData()
		if data == nil {
			break
		}
		container = data.NextContainer
	}
}

func (state *ruleState) collectCallExpression(node *ast.Node) {
	call := node.AsCallExpression()
	if call == nil {
		return
	}
	if state.deferRemainingExpressions || state.shouldDeferExpression(call.Expression) {
		state.deferRemainingExpressions = true
		state.pendingExpressions = append(state.pendingExpressions, node)
		return
	}
	state.checkCallExpression(node)
}

func (state *ruleState) collectNewExpression(node *ast.Node) {
	newExpression := node.AsNewExpression()
	if newExpression == nil {
		return
	}
	if state.deferRemainingExpressions || state.shouldDeferExpression(newExpression.Expression) {
		state.deferRemainingExpressions = true
		state.pendingExpressions = append(state.pendingExpressions, node)
		return
	}
	state.checkNewExpression(node)
}

func (state *ruleState) shouldDeferExpression(node *ast.Node) bool {
	root := expressionRootIdentifier(node)
	if root == nil {
		return false
	}
	name := root.AsIdentifier().Text
	// A known global root with no local declaration cannot become an alias or
	// shadow, so keep its original immediate reporting order. Everything else
	// waits until alias collection is complete.
	return !potentialReferenceRoots[name] || state.ctx.Refs == nil || state.localRootNames[name]
}

func (state *ruleState) checkCallExpression(node *ast.Node) {
	call := node.AsCallExpression()
	if call == nil {
		return
	}

	ref, ok := state.referenceFromExpression(call.Expression)
	if !ok {
		return
	}

	if disallowCallOrNewBuiltins[ref.name] {
		state.ctx.ReportNode(node, messageDisallowCallOrNew(ref.name))
		return
	}

	if !enforceNewBuiltins[ref.name] {
		return
	}

	// An optional chain (`Array?.()` or `Intl?.DateTimeFormat()`) can't be
	// rewritten to a `new` expression, which cannot itself be optional.
	if hasOptionalChain(node) || hasOptionalChain(call.Expression) {
		return
	}

	if ref.name == "Object" && isStrictObjectComparison(node) {
		return
	}

	if ref.name == "Date" {
		state.reportDateCall(node, call)
		return
	}

	state.ctx.ReportNodeWithDeferredFixes(node, messageEnforce(ref.name), func() []rule.RuleFix {
		return []rule.RuleFix{state.enforceNewFix(node)}
	})
}

func (state *ruleState) checkNewExpression(node *ast.Node) {
	newExpression := node.AsNewExpression()
	if newExpression == nil {
		return
	}

	ref, ok := state.referenceFromExpression(newExpression.Expression)
	if !ok {
		return
	}

	if disallowCallOrNewBuiltins[ref.name] {
		state.ctx.ReportNode(node, messageDisallowCallOrNew(ref.name))
		return
	}

	if !disallowNewBuiltins[ref.name] {
		return
	}

	message := messageDisallow(ref.name)
	if ref.name == "String" || ref.name == "Boolean" || ref.name == "Number" {
		state.ctx.ReportNode(node, message)
		return
	}

	state.ctx.ReportNodeWithDeferredFixes(node, message, func() []rule.RuleFix {
		return state.newToCallFixes(node, newExpression)
	})
}

func (state *ruleState) reportDateCall(node *ast.Node, call *ast.CallExpression) {
	message := rule.RuleMessage{Id: messageIDErrorDate, Description: messageDateDescription}

	hasArguments := call.Arguments != nil && len(call.Arguments.Nodes) > 0
	if !hasArguments && !utils.HasCommentInsideNode(state.ctx.SourceFile, node) {
		state.ctx.ReportNodeWithDeferredFixes(node, message, func() []rule.RuleFix {
			return []rule.RuleFix{state.dateFix(node)}
		})
		return
	}

	state.ctx.ReportNodeWithDeferredSuggestions(node, message, func() []rule.RuleSuggestion {
		return []rule.RuleSuggestion{{
			Message:  rule.RuleMessage{Id: messageIDSuggestionDate, Description: messageDateSuggestion},
			FixesArr: []rule.RuleFix{state.dateFix(node)},
		}}
	})
}

func (state *ruleState) dateFix(node *ast.Node) rule.RuleFix {
	return rule.RuleFixReplaceRange(
		utils.TrimNodeTextRange(state.ctx.SourceFile, node),
		state.dateReplacement(node),
	)
}

func (state *ruleState) newToCallFixes(node *ast.Node, newExpression *ast.NewExpression) []rule.RuleFix {
	nodeRange := utils.TrimNodeTextRange(state.ctx.SourceFile, node)
	expressionRange := utils.TrimNodeTextRange(state.ctx.SourceFile, newExpression.Expression)
	if nodeRange.Pos() >= expressionRange.Pos() {
		return nil
	}

	source := state.ctx.SourceFile.Text()
	removeEnd := nodeRange.Pos() + len("new")
	for removeEnd < expressionRange.Pos() && isWhitespace(source[removeEnd]) {
		removeEnd++
	}

	fixes := []rule.RuleFix{
		rule.RuleFixRemoveRange(core.NewTextRange(nodeRange.Pos(), removeEnd)),
	}

	insertAfterExpression := ""
	if newExpression.Arguments == nil {
		insertAfterExpression = "()"
	}

	if state.needsReturnOrThrowParentheses(node, nodeRange.Pos(), expressionRange.Pos()) {
		if opening, closing, ok := returnOrThrowParenthesesRanges(state.ctx.SourceFile, node.Parent); ok {
			fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(opening, opening), " ("))
			if closing == expressionRange.End() {
				insertAfterExpression += ")"
			} else {
				fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(closing, closing), ")"))
			}
		}
	}

	if insertAfterExpression != "" {
		fixes = append(fixes, rule.RuleFixReplaceRange(
			core.NewTextRange(expressionRange.End(), expressionRange.End()),
			insertAfterExpression,
		))
	}
	return fixes
}

func (state *ruleState) needsReturnOrThrowParentheses(node *ast.Node, newPos int, expressionPos int) bool {
	if node.Parent == nil || node.Parent.Kind == ast.KindParenthesizedExpression {
		return false
	}
	if node.Parent.Kind != ast.KindReturnStatement && node.Parent.Kind != ast.KindThrowStatement {
		return false
	}
	return !sameLine(state.ctx.SourceFile, newPos, expressionPos)
}

func returnOrThrowParenthesesRanges(sourceFile *ast.SourceFile, statement *ast.Node) (int, int, bool) {
	if sourceFile == nil || statement == nil {
		return 0, 0, false
	}

	statementRange := utils.TrimNodeTextRange(sourceFile, statement)
	keywordLength := len("return")
	if statement.Kind == ast.KindThrowStatement {
		keywordLength = len("throw")
	}

	opening := statementRange.Pos() + keywordLength
	closing := statementRange.End()
	source := sourceFile.Text()
	for pos := closing - 1; pos >= statementRange.Pos(); pos-- {
		if isWhitespace(source[pos]) {
			continue
		}
		if source[pos] == ';' {
			closing = pos
		}
		break
	}
	return opening, closing, true
}

func (state *ruleState) enforceNewFix(node *ast.Node) rule.RuleFix {
	nodeRange := utils.TrimNodeTextRange(state.ctx.SourceFile, node)
	text := "new "
	source := state.ctx.SourceFile.Text()
	if utils.NeedsLeadingSpaceForReplacement(source, nodeRange.Pos(), text) {
		text = " " + text
	}
	return rule.RuleFixReplaceRange(core.NewTextRange(nodeRange.Pos(), nodeRange.Pos()), text)
}

func (state *ruleState) dateReplacement(node *ast.Node) string {
	nodeRange := utils.TrimNodeTextRange(state.ctx.SourceFile, node)
	if utils.NeedsLeadingSpaceForReplacement(state.ctx.SourceFile.Text(), nodeRange.Pos(), replacementStringNewDateCall) {
		return " " + replacementStringNewDateCall
	}
	return replacementStringNewDateCall
}

func (state *ruleState) referenceFromExpression(node *ast.Node) (reference, bool) {
	path, root, ok := state.expressionPath(node)
	if !ok || len(path) == 0 {
		return reference{}, false
	}
	rootName := root.AsIdentifier().Text
	hasAliases := len(state.aliasesByID[rootName]) > 0
	if !potentialReferenceRoots[rootName] && !hasAliases {
		return reference{}, false
	}

	// Without a same-named alias, the syntactic path is already enough to
	// reject member calls such as Promise.resolve() before any scope lookup.
	if !hasAliases {
		directPath := path
		if globalObjectNames[rootName] {
			directPath = path[1:]
		}
		if _, ok := watchedBuiltinName(directPath); !ok {
			return reference{}, false
		}
	}

	global, ok := state.globalReferenceInfoFromPath(path, root)
	if !ok {
		return reference{}, false
	}
	name, ok := watchedBuiltinName(global.path)
	if !ok {
		return reference{}, false
	}
	return reference{name: name}, true
}

func expressionRootIdentifier(node *ast.Node) *ast.Node {
	node = utils.SkipAssertionsAndParens(node)
	for node != nil && ast.IsAccessExpression(node) {
		node = utils.SkipAssertionsAndParens(utils.AccessExpressionObject(node))
	}
	if node == nil || node.Kind != ast.KindIdentifier {
		return nil
	}
	return node
}

func (state *ruleState) collectAliasDeclaration(node *ast.Node) {
	declaration := node.AsVariableDeclaration()
	if declaration == nil || declaration.Name() == nil || declaration.Initializer == nil {
		return
	}
	name := declaration.Name()
	if name.Kind != ast.KindIdentifier && name.Kind != ast.KindObjectBindingPattern {
		return
	}
	if expressionRootIdentifier(declaration.Initializer) == nil {
		return
	}
	state.aliasDeclarations = append(state.aliasDeclarations, node)
}

func (state *ruleState) collectNamedLocalRoot(node *ast.Node) {
	name := node.Name()
	if name == nil || name.Kind != ast.KindIdentifier || !potentialReferenceRoots[name.Text()] {
		return
	}
	if state.localRootNames == nil {
		state.localRootNames = map[string]bool{}
	}
	state.localRootNames[name.Text()] = true
}

func (state *ruleState) finishAliasCollection() {
	if !state.collectingAliases {
		return
	}
	for _, declaration := range state.aliasDeclarations {
		state.collectVariableAlias(declaration)
	}
	state.aliasDeclarations = nil
	state.collectingAliases = false
}

func (state *ruleState) finish(node *ast.Node) {
	_ = node
	state.finishAliasCollection()
	pendingExpressions := state.pendingExpressions
	state.pendingExpressions = nil
	for _, node := range pendingExpressions {
		switch node.Kind {
		case ast.KindCallExpression:
			state.checkCallExpression(node)
		case ast.KindNewExpression:
			state.checkNewExpression(node)
		}
	}
}

func (state *ruleState) collectVariableAlias(node *ast.Node) {
	declaration := node.AsVariableDeclaration()
	if declaration == nil || declaration.Name() == nil || declaration.Initializer == nil {
		return
	}

	baseReference, ok := state.globalReferenceInfo(declaration.Initializer)
	if !ok {
		return
	}

	name := declaration.Name()
	switch name.Kind {
	case ast.KindIdentifier:
		if state.shouldRegisterAlias(baseReference) {
			state.registerAlias(name, baseReference.path, node)
		}
	case ast.KindObjectBindingPattern:
		state.collectBindingPatternAliases(name, baseReference)
	}
}

func (state *ruleState) collectBindingPatternAliases(pattern *ast.Node, baseReference globalReference) {
	bindingPattern := pattern.AsBindingPattern()
	if bindingPattern == nil || bindingPattern.Elements == nil {
		return
	}
	for _, elementNode := range bindingPattern.Elements.Nodes {
		element := elementNode.AsBindingElement()
		if element == nil || element.DotDotDotToken != nil || element.Name() == nil {
			continue
		}

		propertyName := element.PropertyName
		if propertyName == nil {
			propertyName = element.Name()
		}
		staticName, ok := utils.GetStaticPropertyName(propertyName)
		if !ok {
			continue
		}
		reference := globalReference{
			path:   appendPath(baseReference.path, staticName),
			source: baseReference.source,
		}

		switch element.Name().Kind {
		case ast.KindIdentifier:
			if state.shouldRegisterAlias(reference) {
				state.registerAlias(element.Name(), reference.path, elementNode)
			}
		case ast.KindObjectBindingPattern:
			state.collectBindingPatternAliases(element.Name(), reference)
		}
	}
}

func (state *ruleState) shouldRegisterAlias(reference globalReference) bool {
	if len(reference.path) == 0 {
		return false
	}

	// ESLint's ReferenceTracker does not follow aliases that originate from a
	// bare `WebAssembly` read, but it does follow `globalThis.WebAssembly` and
	// destructured aliases from a global object.
	if reference.source == referenceSourceBareGlobal && reference.path[0] == "WebAssembly" {
		return false
	}

	name := pathKey(reference.path)
	return isWatchedBuiltin(name) || isWatchedNamespace(name)
}

func (state *ruleState) registerAlias(binding *ast.Node, path []string, declarations ...*ast.Node) {
	if binding == nil || binding.Kind != ast.KindIdentifier {
		return
	}
	if state.aliases == nil {
		state.aliases = map[*ast.Node]aliasInfo{}
		state.aliasesByID = map[string][]aliasInfo{}
	}

	info := aliasInfo{
		binding: binding,
		scope:   state.aliasScopeForBinding(binding),
		path:    slices.Clone(path),
	}

	state.aliases[binding] = info
	for _, declaration := range declarations {
		if declaration != nil {
			state.aliases[declaration] = info
		}
	}
	state.aliasesByID[binding.AsIdentifier().Text] = append(state.aliasesByID[binding.AsIdentifier().Text], info)
}

func (state *ruleState) globalReferenceInfo(node *ast.Node) (globalReference, bool) {
	path, root, ok := state.expressionPath(node)
	if !ok || len(path) == 0 {
		return globalReference{}, false
	}
	return state.globalReferenceInfoFromPath(path, root)
}

func (state *ruleState) globalReferenceInfoFromPath(path []string, root *ast.Node) (globalReference, bool) {
	// A tracked local alias wins even if it reuses one of the conventional
	// global-object names (for example, `{Array: window}`). ReferenceTracker
	// follows the binding identity, not its spelling.
	if aliasPath, ok := state.aliasPathForIdentifier(root); ok {
		return globalReference{path: appendPath(aliasPath, path[1:]...), source: referenceSourceAlias}, true
	}

	if globalObjectNames[path[0]] {
		if state.isLocalNonAliasIdentifier(root) {
			return globalReference{}, false
		}
		return globalReference{path: slices.Clone(path[1:]), source: referenceSourceGlobalObject}, true
	}

	// A non-global, non-alias root cannot become a watched builtin. This also
	// keeps alias collection from doing shadow work for ordinary initializers.
	if !potentialReferenceRoots[path[0]] {
		return globalReference{}, false
	}

	if state.isLocalNonAliasIdentifier(root) {
		return globalReference{}, false
	}

	return globalReference{path: slices.Clone(path), source: referenceSourceBareGlobal}, true
}

func (state *ruleState) expressionPath(node *ast.Node) ([]string, *ast.Node, bool) {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return nil, nil, false
	}

	if node.Kind == ast.KindIdentifier {
		return []string{node.AsIdentifier().Text}, node, true
	}

	if ast.IsAccessExpression(node) {
		name, ok := accessExpressionStaticName(node)
		if !ok {
			return nil, nil, false
		}
		path, root, ok := state.expressionPath(utils.AccessExpressionObject(node))
		if !ok {
			return nil, nil, false
		}
		return append(path, name), root, true
	}

	return nil, nil, false
}

func accessExpressionStaticName(node *ast.Node) (string, bool) {
	if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		if access == nil || access.ArgumentExpression == nil {
			return "", false
		}
		// Element keys can be wrapped in TS assertions; keep those wrappers
		// transparent while still requiring a compile-time static key.
		return utils.GetStaticExpressionValue(utils.SkipAssertionsAndParens(access.ArgumentExpression))
	}
	return utils.AccessExpressionStaticName(node)
}

func (state *ruleState) aliasPathForIdentifier(node *ast.Node) ([]string, bool) {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || node.Kind != ast.KindIdentifier {
		return nil, false
	}
	name := node.AsIdentifier().Text
	infos := state.aliasesByID[name]
	if len(infos) == 0 {
		return nil, false
	}

	if !state.collectingAliases && state.ctx.Refs != nil {
		if symbol := state.ctx.Refs.Resolve(node); symbol != nil {
			info, ok := state.aliasInfoForBinderSymbol(symbol)
			if !ok {
				return nil, false
			}
			return slices.Clone(info.path), true
		}
	}

	if state.ctx.TypeChecker != nil {
		symbol := utils.GetReferenceSymbol(node, state.ctx.TypeChecker)
		if symbol != nil {
			for _, declaration := range symbol.Declarations {
				if info, ok := state.aliases[declaration]; ok && state.aliasInfoApplies(node, info) {
					return slices.Clone(info.path), true
				}
			}
		}
	}
	for i := len(infos) - 1; i >= 0; i-- {
		if state.aliasInfoApplies(node, infos[i]) {
			return slices.Clone(infos[i].path), true
		}
	}
	return nil, false
}

func aliasBindingSymbol(binding *ast.Node) *ast.Symbol {
	if binding == nil || binding.Kind != ast.KindIdentifier {
		return nil
	}
	declaration := binding.Parent
	if declaration == nil || declaration.Name() != binding {
		return nil
	}
	return declaration.Symbol()
}

func (state *ruleState) aliasInfoForBinderSymbol(symbol *ast.Symbol) (aliasInfo, bool) {
	if symbol == nil {
		return aliasInfo{}, false
	}
	for _, declaration := range symbol.Declarations {
		if info, ok := state.aliases[declaration]; ok {
			return info, true
		}
	}
	for _, info := range state.aliasesByID[symbol.Name] {
		if aliasBindingSymbol(info.binding) == symbol {
			return info, true
		}
	}
	return aliasInfo{}, false
}

func (state *ruleState) aliasInfoApplies(node *ast.Node, info aliasInfo) bool {
	if node == nil || node.Kind != ast.KindIdentifier || info.binding == nil || len(info.path) == 0 {
		return false
	}
	scope := info.scope
	if scope == nil && state.ctx.SourceFile != nil {
		scope = state.ctx.SourceFile.AsNode()
	}
	if scope != nil {
		if !nodeWithin(node, scope) {
			return false
		}
		if aliasHasCloserShadow(node, scope, node.AsIdentifier().Text, info.binding) {
			return false
		}
	}
	return true
}

func (state *ruleState) aliasScopeForBinding(binding *ast.Node) *ast.Node {
	variableDeclaration := ast.FindAncestorKind(binding, ast.KindVariableDeclaration)
	if variableDeclaration == nil {
		if state.ctx.SourceFile != nil {
			return state.ctx.SourceFile.AsNode()
		}
		return nil
	}

	declarationList := ast.FindAncestorKind(variableDeclaration, ast.KindVariableDeclarationList)
	if declarationList != nil && utils.IsVarKeyword(declarationList) {
		return varAliasScope(variableDeclaration)
	}

	if declarationList != nil && declarationList.Parent != nil {
		switch declarationList.Parent.Kind {
		case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
			return declarationList.Parent
		}
	}

	scope := ast.FindAncestor(variableDeclaration, func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindBlock, ast.KindCaseBlock, ast.KindModuleBlock, ast.KindSourceFile:
			return true
		default:
			return false
		}
	})
	if scope != nil {
		return scope
	}
	if state.ctx.SourceFile != nil {
		return state.ctx.SourceFile.AsNode()
	}
	return nil
}

func varAliasScope(node *ast.Node) *ast.Node {
	for current := node; current != nil; current = current.Parent {
		if current.Kind == ast.KindSourceFile || current.Kind == ast.KindModuleBlock {
			return current
		}
		if ast.IsFunctionLikeDeclaration(current) {
			if body := current.Body(); body != nil {
				return body
			}
			return current
		}
	}
	return nil
}

func nodeWithin(node *ast.Node, container *ast.Node) bool {
	return node != nil && container != nil &&
		node.Pos() >= container.Pos() && node.End() <= container.End()
}

// aliasHasCloserShadow mirrors utils.IsNameShadowedBetween, but ignores
// ancestor scopes that contain the alias binding itself. This keeps `var`
// aliases inside blocks from looking like shadows of their own references.
func aliasHasCloserShadow(node *ast.Node, boundary *ast.Node, name string, binding *ast.Node) bool {
	for current := node.Parent; current != nil && current != boundary; current = current.Parent {
		if nodeWithin(binding, current) {
			continue
		}
		if ast.IsFunctionLikeDeclaration(current) && utils.HasShadowingParameter(current, name) {
			return true
		}
		switch current.Kind {
		case ast.KindBlock:
			if utils.HasShadowingDeclaration(current, name) {
				return true
			}
		case ast.KindCatchClause:
			catchClause := current.AsCatchClause()
			if catchClause != nil && catchClause.VariableDeclaration != nil {
				variableDeclaration := catchClause.VariableDeclaration.AsVariableDeclaration()
				if variableDeclaration != nil && variableDeclaration.Name() != nil &&
					utils.HasNameInBindingPattern(variableDeclaration.Name(), name) {
					return true
				}
			}
		case ast.KindForStatement:
			forStatement := current.AsForStatement()
			if forStatement != nil && forStatement.Initializer != nil &&
				forStatement.Initializer.Kind == ast.KindVariableDeclarationList &&
				utils.HasVarDeclListWithName(forStatement.Initializer, name) {
				return true
			}
		case ast.KindForInStatement, ast.KindForOfStatement:
			statement := current.AsForInOrOfStatement()
			if statement != nil && statement.Initializer != nil &&
				statement.Initializer.Kind == ast.KindVariableDeclarationList &&
				utils.HasVarDeclListWithName(statement.Initializer, name) {
				return true
			}
		case ast.KindClassDeclaration, ast.KindClassExpression:
			if n := current.Name(); n != nil && n.Kind == ast.KindIdentifier && n.Text() == name {
				return true
			}
		}
	}
	return false
}

func (state *ruleState) isLocalNonAliasIdentifier(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}
	if _, ok := state.aliasPathForIdentifier(node); ok {
		return false
	}

	name := node.AsIdentifier().Text
	if state.ctx.Refs != nil {
		if !state.localRootNames[name] {
			return false
		}
		// Callee roots are reference-position identifiers. Resolve can
		// return a symbol declared outside this file (globals, cross-file,
		// .d.ts), so a non-nil result alone doesn't imply local — it must
		// additionally be declared in this file.
		symbol := state.ctx.Refs.Resolve(node)
		if symbol != nil {
			// A resolved type-only symbol is authoritative. Falling through
			// would rescan the containing scope for every reference.
			return utils.IsValueSymbolDeclaredInFile(symbol, state.ctx.SourceFile)
		}
		// A namespace holding no value of its own is outside the binder's
		// value lookup, and without a TypeChecker to fall back to Resolve
		// can't see it at all — so keep the syntactic check, which counts
		// every identifier-named namespace. The localRootNames gate above
		// keeps it off the common path.
		return utils.IsShadowed(node, name)
	}

	if state.ctx.TypeChecker != nil && state.ctx.SourceFile != nil {
		symbol := utils.GetReferenceSymbol(node, state.ctx.TypeChecker)
		if utils.IsValueSymbolDeclaredInFile(symbol, state.ctx.SourceFile) {
			return true
		}
	}

	return utils.IsShadowed(node, name)
}

func hasOptionalChain(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	if ast.IsOptionalChain(node) {
		return true
	}
	switch node.Kind {
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		return call != nil && hasOptionalChain(call.Expression)
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		return access != nil && hasOptionalChain(access.Expression)
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		return access != nil && hasOptionalChain(access.Expression)
	case ast.KindNonNullExpression:
		expression := node.AsNonNullExpression()
		return expression != nil && hasOptionalChain(expression.Expression)
	default:
		return false
	}
}

func isStrictObjectComparison(node *ast.Node) bool {
	current := node
	for current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}

	parent := current.Parent
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := parent.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil {
		return false
	}
	if binary.OperatorToken.Kind != ast.KindEqualsEqualsEqualsToken && binary.OperatorToken.Kind != ast.KindExclamationEqualsEqualsToken {
		return false
	}
	return binary.Left == current || binary.Right == current
}

func messageEnforce(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageIDEnforce,
		Description: enforceMessageDescriptions[name],
		Data:        map[string]string{"name": name},
	}
}

func messageDisallow(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageIDDisallow,
		Description: disallowMessageDescriptions[name],
		Data:        map[string]string{"name": name},
	}
}

func messageDisallowCallOrNew(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageIDDisallowCallOrNew,
		Description: disallowCallOrNewMessageDescriptions[name],
		Data:        map[string]string{"name": name},
	}
}

func buildMessageDescriptions(names map[string]bool, build func(string) string) map[string]string {
	descriptions := make(map[string]string, len(names))
	for name := range names {
		descriptions[name] = build(name)
	}
	return descriptions
}

func isWatchedBuiltin(name string) bool {
	return enforceNewBuiltins[name] || disallowNewBuiltins[name] || disallowCallOrNewBuiltins[name]
}

func isWatchedNamespace(name string) bool {
	return watchedNamespaces[name]
}

func watchedBuiltinName(path []string) (string, bool) {
	if len(path) == 0 {
		return "", false
	}
	if len(path) > 1 && !watchedNamespaces[path[0]] {
		return "", false
	}
	name := pathKey(path)
	return name, isWatchedBuiltin(name)
}

func sourceHasPotentialReference(sourceFile *ast.SourceFile) bool {
	if sourceFile == nil || sourceFile.Identifiers == nil {
		return true
	}
	for root := range potentialReferenceRoots {
		if _, ok := sourceFile.Identifiers[root]; ok {
			return true
		}
	}
	return false
}

func pathKey(path []string) string {
	return strings.Join(path, ".")
}

func appendPath(path []string, parts ...string) []string {
	return append(slices.Clone(path), parts...)
}

func sameLine(sourceFile *ast.SourceFile, a int, b int) bool {
	lineStarts := sourceFile.ECMALineMap()
	return scanner.ComputeLineOfPosition(lineStarts, a) ==
		scanner.ComputeLineOfPosition(lineStarts, b)
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '\f' || ch == '\v'
}
