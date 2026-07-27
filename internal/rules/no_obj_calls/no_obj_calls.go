package no_obj_calls

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var nonCallableGlobalNames = [...]string{
	"Atomics",
	"JSON",
	"Math",
	"Reflect",
	"Intl",
	"Temporal",
}

var nonCallableGlobals = func() map[string]bool {
	result := make(map[string]bool, len(nonCallableGlobalNames))
	for _, name := range nonCallableGlobalNames {
		result[name] = true
	}
	return result
}()

var globalObjectNames = [...]string{
	"global",
	"globalThis",
	"self",
	"window",
}

var globalObjects = func() map[string]bool {
	result := make(map[string]bool, len(globalObjectNames))
	for _, name := range globalObjectNames {
		result[name] = true
	}
	return result
}()

func sourceMayUseNonCallableGlobal(sourceFile *ast.SourceFile) bool {
	if sourceFile == nil || sourceFile.Identifiers == nil {
		return true
	}
	for _, name := range nonCallableGlobalNames {
		if _, ok := sourceFile.Identifiers[name]; ok {
			return true
		}
	}
	// Computed global-object access such as globalThis["Math"] does not put
	// "Math" in SourceFile.Identifiers.
	for _, name := range globalObjectNames {
		if _, ok := sourceFile.Identifiers[name]; ok {
			return true
		}
	}
	return false
}

type traceKind uint8

const (
	traceNonCallable traceKind = iota
	traceGlobalObject
)

// trackedValue is the small trace map this rule needs. A global-object value
// accepts one of nonCallableGlobalNames as its next static property; a
// non-callable value accepts only call and construct operations.
type trackedValue struct {
	kind       traceKind
	globalName string
}

type traceWalk struct {
	variables map[*ast.Symbol]bool
	globals   map[string]bool
}

type ruleState struct {
	ctx rule.RuleContext

	// Only watched/global-object identifiers are indexed eagerly. Other names
	// are collected on demand solely for the rare `/* global alias */ alias =
	// JSON` propagation path.
	identifiersByName map[string][]*ast.Node
	indexedNames      map[string]bool
	allNamesIndexed   bool

	// localSymbols contains raw binder symbols. Requested names are indexed
	// lazily: files with many declarations but no tracked alias should not pay
	// for a per-declaration map entry. ctx.Refs also deals in raw symbols, so
	// expanding an alias never needs the TypeChecker.
	localSymbols           []*ast.Symbol
	symbolsByName          map[string][]*ast.Symbol
	indexedSymbolNames     map[string]bool
	symbolLookupCount      int
	allSymbolsIndexed      bool
	resolvedReferenceNames map[string]bool
	referenceSymbols       map[*ast.Node]*ast.Symbol

	propertyEvaluator *utils.StaticStringEvaluator
	calls             map[*ast.Node]string
	additionalCalls   map[*ast.Node][]string
}

var lexicalShadowSentinel = &ast.Symbol{}

func newRuleState(ctx rule.RuleContext) *ruleState {
	state := &ruleState{
		ctx:                    ctx,
		identifiersByName:      make(map[string][]*ast.Node),
		indexedNames:           make(map[string]bool),
		symbolsByName:          make(map[string][]*ast.Symbol),
		indexedSymbolNames:     make(map[string]bool),
		resolvedReferenceNames: make(map[string]bool),
	}
	state.collectFileIndex()
	state.collectCalls()
	return state
}

func (state *ruleState) collectFileIndex() {
	if state.ctx.SourceFile == nil {
		return
	}

	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if symbol := node.Symbol(); isTrackableLocalSymbol(symbol) {
			state.localSymbols = append(state.localSymbols, symbol)
		}

		if node.Kind == ast.KindIdentifier {
			name := node.AsIdentifier().Text
			if (nonCallableGlobals[name] || globalObjects[name]) &&
				!utils.IsNonReferenceIdentifier(node) {
				state.identifiersByName[name] = append(state.identifiersByName[name], node)
				state.indexedNames[name] = true
			}
		}

		node.ForEachChild(visit)
		return false
	}
	visit(state.ctx.SourceFile.AsNode())
}

func isTrackableLocalSymbol(symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	const flags = ast.SymbolFlagsVariable |
		ast.SymbolFlagsFunction |
		ast.SymbolFlagsClass |
		ast.SymbolFlagsEnum |
		ast.SymbolFlagsModule |
		ast.SymbolFlagsAlias
	return symbol.Flags&flags != 0
}

func (state *ruleState) collectCalls() {
	for _, name := range nonCallableGlobalNames {
		if state.isGlobalOff(name) || state.globalIsModified(name) {
			continue
		}
		value := trackedValue{kind: traceNonCallable, globalName: name}
		for _, identifier := range state.identifiersByName[name] {
			if state.isGlobalReference(identifier, name) {
				walk := traceWalk{}
				state.trackExpression(identifier, value, &walk)
			}
		}
	}

	for _, name := range globalObjectNames {
		if state.isGlobalOff(name) || state.globalIsModified(name) {
			continue
		}
		value := trackedValue{kind: traceGlobalObject}
		for _, identifier := range state.identifiersByName[name] {
			if state.isGlobalReference(identifier, name) {
				walk := traceWalk{}
				state.trackExpression(identifier, value, &walk)
			}
		}
	}
}

func (state *ruleState) isGlobalOff(name string) bool {
	declared, ok := state.ctx.Globals[name]
	return ok && !declared
}

func (state *ruleState) globalIsModified(name string) bool {
	for _, identifier := range state.identifiersByName[name] {
		if state.isGlobalReference(identifier, name) && utils.IsWriteReference(identifier) {
			return true
		}
	}
	return false
}

func (state *ruleState) isGlobalReference(identifier *ast.Node, name string) bool {
	if identifier == nil || identifier.Kind != ast.KindIdentifier ||
		utils.IsNonReferenceIdentifier(identifier) {
		return false
	}
	return state.localReferenceSymbol(identifier) == nil
}

func (state *ruleState) localReferenceSymbol(identifier *ast.Node) *ast.Symbol {
	if identifier == nil || identifier.Kind != ast.KindIdentifier {
		return nil
	}
	name := identifier.AsIdentifier().Text
	if !state.resolvedReferenceNames[name] {
		state.resolvedReferenceNames[name] = true
		if state.ctx.Refs != nil {
			for _, symbol := range state.symbolsForName(name) {
				for _, reference := range state.ctx.Refs.References(symbol) {
					if state.referenceSymbols == nil {
						state.referenceSymbols = make(map[*ast.Node]*ast.Symbol)
					}
					state.referenceSymbols[reference] = symbol
				}
			}
		}
	}
	if symbol := state.referenceSymbols[identifier]; symbol != nil {
		return symbol
	}

	// NamespaceModule symbols do not resolve in RefStore's value lookup even
	// though ESLint scope analysis treats namespace declarations as bindings.
	// The same fallback also protects direct/unit callers without a Program.
	if (state.ctx.Refs == nil || len(state.symbolsForName(name)) != 0) &&
		utils.IsShadowed(identifier, name) {
		return lexicalShadowSentinel
	}
	return nil
}

func (state *ruleState) symbolsForName(name string) []*ast.Symbol {
	if state.allSymbolsIndexed {
		return state.symbolsByName[name]
	}
	if state.indexedSymbolNames[name] {
		return state.symbolsByName[name]
	}
	state.indexedSymbolNames[name] = true
	state.symbolLookupCount++

	// A few watched-root lookups are cheaper as flat scans and avoid indexing
	// every declaration in the common case. Assignment-heavy alias chains can
	// request many distinct names; cap those scans, then build the complete
	// name index once to keep the tracker linear.
	if state.symbolLookupCount > 8 {
		state.symbolsByName = make(map[string][]*ast.Symbol)
		for _, symbol := range state.localSymbols {
			state.addSymbolByName(symbol)
		}
		state.allSymbolsIndexed = true
		return state.symbolsByName[name]
	}

	for _, symbol := range state.localSymbols {
		if symbol.Name == name {
			state.addSymbolByName(symbol)
		}
	}
	return state.symbolsByName[name]
}

func (state *ruleState) addSymbolByName(symbol *ast.Symbol) {
	for _, existing := range state.symbolsByName[symbol.Name] {
		if existing == symbol {
			return
		}
	}
	state.symbolsByName[symbol.Name] = append(state.symbolsByName[symbol.Name], symbol)
}

func (state *ruleState) trackExpression(node *ast.Node, value trackedValue, walk *traceWalk) {
	if node == nil {
		return
	}

	// ESTree discards parentheses and represents the remaining constructs
	// below as transparent value-preserving wrappers. tsgo keeps several of
	// them as nodes, so walk upward through the equivalent shape first.
	for node.Parent != nil && isPassThrough(node, node.Parent) {
		node = node.Parent
	}

	parent := node.Parent
	if parent == nil {
		return
	}

	if ast.IsAccessExpression(parent) && utils.AccessExpressionObject(parent) == node {
		if value.kind != traceGlobalObject {
			return
		}
		name, ok := state.accessExpressionStaticName(parent)
		if !ok || !nonCallableGlobals[name] {
			return
		}
		state.trackExpression(parent, trackedValue{
			kind:       traceNonCallable,
			globalName: name,
		}, walk)
		return
	}

	switch parent.Kind {
	case ast.KindCallExpression:
		if parent.AsCallExpression().Expression == node && value.kind == traceNonCallable {
			state.addCall(parent, value.globalName)
		}

	case ast.KindNewExpression:
		if parent.AsNewExpression().Expression == node && value.kind == traceNonCallable {
			state.addCall(parent, value.globalName)
		}

	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary != nil && binary.Right == node &&
			binary.OperatorToken != nil && ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			state.trackAssignmentTarget(binary.Left, value, walk)
			// Assignment expressions themselves evaluate to a value and can be
			// nested in another assignment/declaration.
			state.trackExpression(parent, value, walk)
		}

	case ast.KindVariableDeclaration:
		declaration := parent.AsVariableDeclaration()
		if declaration != nil && declaration.Initializer == node {
			state.trackAssignmentTarget(declaration.Name(), value, walk)
		}

	case ast.KindParameter:
		parameter := parent.AsParameterDeclaration()
		if parameter != nil && parameter.Initializer == node {
			state.trackAssignmentTarget(parameter.Name(), value, walk)
		}

	case ast.KindBindingElement:
		element := parent.AsBindingElement()
		if element != nil && element.Initializer == node {
			state.trackAssignmentTarget(element.Name(), value, walk)
		}

	case ast.KindShorthandPropertyAssignment:
		property := parent.AsShorthandPropertyAssignment()
		if property != nil && property.ObjectAssignmentInitializer == node {
			state.trackAssignmentTarget(property.Name(), value, walk)
		}
	}
}

func (state *ruleState) addCall(node *ast.Node, globalName string) {
	if state.calls == nil {
		state.calls = make(map[*ast.Node]string)
	}
	_, exists := state.calls[node]
	if !exists {
		state.calls[node] = globalName
		return
	}
	if state.additionalCalls == nil {
		state.additionalCalls = make(map[*ast.Node][]string)
	}
	state.additionalCalls[node] = append(state.additionalCalls[node], globalName)
}

func (state *ruleState) accessExpressionStaticName(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	if node.Kind == ast.KindElementAccessExpression {
		access := node.AsElementAccessExpression()
		if access == nil || access.ArgumentExpression == nil {
			return "", false
		}
		if state.propertyEvaluator == nil {
			// ReferenceTracker evaluates computed keys without a scope
			// argument: literal expressions can fold, but identifiers cannot.
			// A nil-checker evaluator has exactly that property.
			state.propertyEvaluator = utils.NewStaticStringEvaluatorWithSourceFile(nil, state.ctx.SourceFile)
		}
		return state.propertyEvaluator.Eval(access.ArgumentExpression)
	}
	return utils.AccessExpressionStaticName(node)
}

func (state *ruleState) staticPropertyName(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	if node.Kind == ast.KindComputedPropertyName {
		computed := node.AsComputedPropertyName()
		if computed == nil || computed.Expression == nil {
			return "", false
		}
		if state.propertyEvaluator == nil {
			state.propertyEvaluator = utils.NewStaticStringEvaluatorWithSourceFile(nil, state.ctx.SourceFile)
		}
		return state.propertyEvaluator.Eval(computed.Expression)
	}
	return utils.GetStaticPropertyName(node)
}

func isPassThrough(node *ast.Node, parent *ast.Node) bool {
	switch parent.Kind {
	case ast.KindParenthesizedExpression:
		return parent.AsParenthesizedExpression().Expression == node
	case ast.KindAsExpression:
		return parent.AsAsExpression().Expression == node
	case ast.KindSatisfiesExpression:
		return parent.AsSatisfiesExpression().Expression == node
	case ast.KindTypeAssertionExpression:
		return parent.AsTypeAssertion().Expression == node
	case ast.KindNonNullExpression:
		return parent.AsNonNullExpression().Expression == node
	case ast.KindPartiallyEmittedExpression:
		return parent.Expression() == node
	case ast.KindExpressionWithTypeArguments:
		return parent.AsExpressionWithTypeArguments().Expression == node
	case ast.KindConditionalExpression:
		conditional := parent.AsConditionalExpression()
		return conditional.WhenTrue == node || conditional.WhenFalse == node
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return false
		}
		switch binary.OperatorToken.Kind {
		case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
			return binary.Left == node || binary.Right == node
		case ast.KindCommaToken:
			return binary.Right == node
		}
	}
	return false
}

func (state *ruleState) trackAssignmentTarget(node *ast.Node, value trackedValue, walk *traceWalk) {
	node = ast.SkipParentheses(node)
	if node == nil {
		return
	}

	switch node.Kind {
	case ast.KindIdentifier:
		state.trackIdentifierVariable(node, value, walk)

	case ast.KindObjectBindingPattern:
		pattern := node.AsBindingPattern()
		if pattern == nil || pattern.Elements == nil || value.kind != traceGlobalObject {
			return
		}
		for _, elementNode := range pattern.Elements.Nodes {
			element := elementNode.AsBindingElement()
			if element == nil || element.DotDotDotToken != nil || element.Name() == nil {
				continue
			}
			propertyName := element.PropertyName
			if propertyName == nil {
				propertyName = element.Name()
			}
			name, ok := state.staticPropertyName(propertyName)
			if !ok || !nonCallableGlobals[name] {
				continue
			}
			state.trackAssignmentTarget(element.Name(), trackedValue{
				kind:       traceNonCallable,
				globalName: name,
			}, walk)
		}

	case ast.KindObjectLiteralExpression:
		if value.kind != traceGlobalObject {
			return
		}
		for _, propertyNode := range node.AsObjectLiteralExpression().Properties.Nodes {
			switch propertyNode.Kind {
			case ast.KindPropertyAssignment:
				property := propertyNode.AsPropertyAssignment()
				name, ok := state.staticPropertyName(property.Name())
				if !ok || !nonCallableGlobals[name] {
					continue
				}
				state.trackAssignmentTarget(property.Initializer, trackedValue{
					kind:       traceNonCallable,
					globalName: name,
				}, walk)
			case ast.KindShorthandPropertyAssignment:
				property := propertyNode.AsShorthandPropertyAssignment()
				name, ok := state.staticPropertyName(property.Name())
				if !ok || !nonCallableGlobals[name] {
					continue
				}
				state.trackAssignmentTarget(property.Name(), trackedValue{
					kind:       traceNonCallable,
					globalName: name,
				}, walk)
			}
		}

	case ast.KindBinaryExpression:
		// Default-value assignment patterns (`x = fallback`) retain the
		// tracked value on their left-hand binding.
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil &&
			binary.OperatorToken.Kind == ast.KindEqualsToken {
			state.trackAssignmentTarget(binary.Left, value, walk)
		}
	}
}

func (state *ruleState) trackIdentifierVariable(identifier *ast.Node, value trackedValue, walk *traceWalk) {
	if symbol := bindingSymbol(identifier); symbol != nil {
		state.trackVariable(symbol, value, walk)
		return
	}
	if symbol := state.localReferenceSymbol(identifier); symbol != nil {
		state.trackVariable(symbol, value, walk)
		return
	}

	name := identifier.AsIdentifier().Text
	if state.isKnownGlobal(name) && state.isGlobalReference(identifier, name) {
		state.trackGlobalVariable(name, value, walk)
	}
}

func bindingSymbol(identifier *ast.Node) *ast.Symbol {
	if identifier == nil || identifier.Kind != ast.KindIdentifier || identifier.Parent == nil {
		return nil
	}
	declaration := identifier.Parent
	if declaration.Name() != identifier {
		return nil
	}
	symbol := declaration.Symbol()
	if !isTrackableLocalSymbol(symbol) {
		return nil
	}
	return symbol
}

func (state *ruleState) trackVariable(symbol *ast.Symbol, value trackedValue, walk *traceWalk) {
	if state.ctx.Refs == nil || symbol == nil ||
		(walk.variables != nil && walk.variables[symbol]) {
		return
	}
	if walk.variables == nil {
		walk.variables = make(map[*ast.Symbol]bool)
	}
	walk.variables[symbol] = true
	defer delete(walk.variables, symbol)

	for _, reference := range state.ctx.Refs.References(symbol) {
		if ast.IsWriteOnlyAccess(reference) {
			continue
		}
		state.trackExpression(reference, value, walk)
	}
}

func (state *ruleState) isKnownGlobal(name string) bool {
	if declared, ok := state.ctx.Globals[name]; ok {
		return declared
	}
	return nonCallableGlobals[name] || globalObjects[name]
}

func (state *ruleState) trackGlobalVariable(name string, value trackedValue, walk *traceWalk) {
	if walk.globals != nil && walk.globals[name] {
		return
	}
	if walk.globals == nil {
		walk.globals = make(map[string]bool)
	}
	walk.globals[name] = true
	defer delete(walk.globals, name)

	for _, reference := range state.identifiersForName(name) {
		if ast.IsWriteOnlyAccess(reference) || !state.isGlobalReference(reference, name) {
			continue
		}
		state.trackExpression(reference, value, walk)
	}
}

func (state *ruleState) identifiersForName(name string) []*ast.Node {
	if state.indexedNames[name] || state.allNamesIndexed || state.ctx.SourceFile == nil {
		return state.identifiersByName[name]
	}

	// Reaching an otherwise unknown name means a configured global is being
	// used as an assignment alias. Collect every not-already-indexed name in
	// this rare fallback walk so a chain of configured globals stays linear.
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier {
			identifierName := node.AsIdentifier().Text
			if !state.indexedNames[identifierName] && !utils.IsNonReferenceIdentifier(node) {
				state.identifiersByName[identifierName] = append(state.identifiersByName[identifierName], node)
			}
		}
		node.ForEachChild(visit)
		return false
	}
	visit(state.ctx.SourceFile.AsNode())
	state.allNamesIndexed = true
	return state.identifiersByName[name]
}

func getCalleeName(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	if node == nil {
		return "undefined"
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		if name := node.AsPropertyAccessExpression().Name(); name != nil {
			return name.Text()
		}
		return "null"
	case ast.KindElementAccessExpression:
		argument := ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		if name, ok := utils.GetStaticExpressionValue(argument); ok {
			return name
		}
		// ESLint's reporting helper is deliberately narrower than
		// ReferenceTracker's computed-key evaluator and returns null for
		// folded/TS-wrapped keys.
		return "null"
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	}
	return "undefined"
}

func (state *ruleState) report(node *ast.Node, globalName string) {
	var callee *ast.Node
	switch node.Kind {
	case ast.KindCallExpression:
		callee = node.AsCallExpression().Expression
	case ast.KindNewExpression:
		callee = node.AsNewExpression().Expression
	default:
		return
	}

	name := getCalleeName(callee)
	if name == globalName {
		state.ctx.ReportNode(node, rule.RuleMessage{
			Id:          "unexpectedCall",
			Description: fmt.Sprintf("'%s' is not a function.", name),
		})
		return
	}
	state.ctx.ReportNode(node, rule.RuleMessage{
		Id:          "unexpectedRefCall",
		Description: fmt.Sprintf("'%s' is reference to '%s', which is not a function.", name, globalName),
	})
}

// https://eslint.org/docs/latest/rules/no-obj-calls
var NoObjCallsRule = rule.Rule{
	Name: "no-obj-calls",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if !sourceMayUseNonCallableGlobal(ctx.SourceFile) {
			return rule.RuleListeners{}
		}

		state := newRuleState(ctx)
		if len(state.calls) == 0 {
			return rule.RuleListeners{}
		}

		report := func(node *ast.Node) {
			globalName, ok := state.calls[node]
			if !ok {
				return
			}
			state.report(node, globalName)
			for _, additional := range state.additionalCalls[node] {
				state.report(node, additional)
			}
		}
		return rule.RuleListeners{
			ast.KindCallExpression: report,
			ast.KindNewExpression:  report,
		}
	},
}
