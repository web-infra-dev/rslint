package block_scoped_var

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/binder"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// https://eslint.org/docs/latest/rules/block-scoped-var
var BlockScopedVarRule = rule.Rule{
	Name:   "block-scoped-var",
	Schema: rule.EmptyArraySchema,
	Run:    run,
}

// A varOccurrence is the one definition ESLint associates with a variable for
// one VariableDeclaration node, paired with the binding context active there.
// A declaration list can bind the same variable more than once, but
// sourceCode.getDeclaredVariables returns that variable only once.
type varOccurrence struct {
	identifier *ast.Node
	scopeRange core.TextRange
}

type referenceEvent struct {
	identifier   *ast.Node
	multiplicity int
	bindingRange bool
	textRange    core.TextRange
}

type varGroup struct {
	symbol            *ast.Symbol
	name              string
	host              *ast.Node
	functionBodyStart int
	mayHaveClassSelf  bool
	occurrences       []varOccurrence
	references        []referenceEvent
}

type pendingSyntheticReference struct {
	target *ast.Symbol
	event  referenceEvent
}

type canonicalKey struct {
	host *ast.Node
	name string
}

type blockScopedVarState struct {
	ctx                        rule.RuleContext
	resolver                   binder.NameResolver
	canonical                  map[canonicalKey]*ast.Symbol
	groups                     map[*ast.Symbol]*varGroup
	groupOrder                 []*varGroup
	ordinarySeen               map[*ast.Node]*varGroup
	pendingSynthetic           []pendingSyntheticReference
	nonVarDeclarations         []*ast.Node
	classExpressions           []*ast.Node
	namespaceExports           []*ast.Node
	exportSpecifiers           []*ast.Node
	indexSignatures            []*ast.Node
	jsxNamespacedNames         []*ast.Node
	functions                  []*ast.Node
	parameterTypeScopes        []*ast.Node
	hasTypeSignatureParameters bool
	hasConditionalInfer        bool
	classNames                 map[string]struct{}
	hasJSX                     bool
}

func run(ctx rule.RuleContext, _ []any) rule.RuleListeners {
	if ctx.SourceFile == nil || !containsVarKeywordCandidate(ctx.SourceFile.Text()) {
		return nil
	}

	options := &core.CompilerOptions{}
	if sourceProgram := ctx.Program(); sourceProgram != nil {
		options = sourceProgram.Options()
	}
	resolver := binder.NameResolver{CompilerOptions: options}
	if ctx.SourceFile.IsBound() && ast.IsGlobalSourceFile(ctx.SourceFile.AsNode()) {
		resolver.Globals = ctx.SourceFile.Locals
	}

	state := &blockScopedVarState{
		ctx:      ctx,
		resolver: resolver,
	}
	listeners := rule.RuleListeners{
		ast.KindVariableDeclarationList:        state.collectVariableDeclarationList,
		rule.ListenerOnExit(ast.KindEndOfFile): state.finish,
	}
	if strings.Contains(ctx.SourceFile.Text(), "class") {
		listeners[ast.KindClassDeclaration] = state.collectClassDeclaration
	}

	// TypeScript's resolver has one known scope-manager mismatch for a named
	// class expression's class-level decorator. Collect only those uncommon
	// nodes, and repair references lazily at EOF.
	if (ctx.SourceFile.ScriptKind == core.ScriptKindTS || ctx.SourceFile.ScriptKind == core.ScriptKindTSX) &&
		strings.Contains(ctx.SourceFile.Text(), "@") {
		listeners[ast.KindClassExpression] = state.collectDecoratedClassExpression
	}
	if ctx.SourceFile.ScriptKind == core.ScriptKindTS || ctx.SourceFile.ScriptKind == core.ScriptKindTSX {
		listeners[ast.KindNamespaceExportDeclaration] = state.collectNamespaceExportDeclaration
		listeners[ast.KindExportSpecifier] = state.collectExportSpecifier
		listeners[ast.KindIndexSignature] = state.collectIndexSignature
	}

	// TSESTree treats both pieces of a namespaced TSX tag as value references;
	// Espree does not. Keep the explicit TSX collection for the rule's indexed
	// ordering pass; addOrdinaryReference deduplicates the shared RefStore view.
	if ctx.SourceFile.ScriptKind == core.ScriptKindTSX {
		listeners[ast.KindJsxOpeningElement] = state.collectJSX
		listeners[ast.KindJsxSelfClosingElement] = state.collectJSX
		listeners[ast.KindJsxFragment] = state.collectJSX
		listeners[ast.KindJsxNamespacedName] = state.collectJsxNamespacedName
	}

	return listeners
}

func (state *blockScopedVarState) collectJSX(_ *ast.Node) {
	state.hasJSX = true
}

func (state *blockScopedVarState) collectClassDeclaration(node *ast.Node) {
	if node.Name() == nil {
		return
	}
	if state.classNames == nil {
		state.classNames = make(map[string]struct{})
	}
	state.classNames[node.Name().Text()] = struct{}{}
}

func containsVarKeywordCandidate(text string) bool {
	for offset := 0; offset < len(text); {
		index := strings.Index(text[offset:], "var")
		if index == -1 {
			return false
		}
		index += offset
		beforeBoundary := index == 0 || !isASCIIIdentifierPart(text[index-1])
		after := index + len("var")
		afterBoundary := after == len(text) || !isASCIIIdentifierPart(text[after])
		if beforeBoundary && afterBoundary {
			return true
		}
		offset = index + len("var")
	}
	return false
}

func isASCIIIdentifierPart(character byte) bool {
	return character >= 'a' && character <= 'z' ||
		character >= 'A' && character <= 'Z' ||
		character >= '0' && character <= '9' ||
		character == '_' || character == '$'
}

func (state *blockScopedVarState) collectVariableDeclarationList(node *ast.Node) {
	if !utils.IsVarKeyword(node) {
		state.nonVarDeclarations = append(state.nonVarDeclarations, node)
		return
	}

	scopeRange := blockScopeRange(state.ctx.SourceFile, node)
	reportable := scopeRange.Pos() != 0 || scopeRange.End() < state.ctx.SourceFile.AsNode().End()
	isForInOrOf := node.Parent != nil &&
		(node.Parent.Kind == ast.KindForInStatement || node.Parent.Kind == ast.KindForOfStatement)
	host := utils.FindEnclosingScope(node)
	var firstOccurrence *varGroup
	var seenOccurrences map[*varGroup]struct{}

	utils.ForEachVariableDeclarationBinding(node, func(declaration *ast.Node, identifier *ast.Node, name string) {
		symbol := state.canonicalSymbol(identifier, name, host)
		if symbol == nil {
			return
		}

		group := state.group(symbol, host, name)

		newOccurrence := firstOccurrence == nil
		if firstOccurrence == nil {
			firstOccurrence = group
		} else if group != firstOccurrence {
			if seenOccurrences == nil {
				seenOccurrences = map[*varGroup]struct{}{firstOccurrence: {}}
			}
			if _, seen := seenOccurrences[group]; !seen {
				seenOccurrences[group] = struct{}{}
				newOccurrence = true
			}
		}
		if newOccurrence {
			if reportable {
				group.occurrences = append(group.occurrences, varOccurrence{
					identifier: identifier,
					scopeRange: scopeRange,
				})
			}
		}

		multiplicity := variableDeclarationWriteMultiplicity(declaration, identifier, isForInOrOf)
		if multiplicity == 0 {
			return
		}
		if target := state.resolve(identifier, name, ast.SymbolFlagsValue|ast.SymbolFlagsAlias); target != nil {
			state.addSyntheticReference(
				target,
				referenceEvent{
					identifier:   identifier,
					multiplicity: multiplicity,
					bindingRange: true,
				},
			)
		}
	})
}

func (state *blockScopedVarState) collectNonVarDeclarationWrites(node *ast.Node, relevantNames map[string]struct{}) {
	isForInOrOf := node.Parent != nil &&
		(node.Parent.Kind == ast.KindForInStatement || node.Parent.Kind == ast.KindForOfStatement)
	utils.ForEachVariableDeclarationBinding(node, func(declaration *ast.Node, identifier *ast.Node, name string) {
		if _, relevant := relevantNames[name]; !relevant {
			return
		}
		multiplicity := variableDeclarationWriteMultiplicity(declaration, identifier, isForInOrOf)
		if multiplicity == 0 {
			return
		}
		if target := state.resolve(identifier, name, ast.SymbolFlagsValue|ast.SymbolFlagsAlias); target != nil {
			state.addSyntheticReference(
				target,
				referenceEvent{
					identifier:   identifier,
					multiplicity: multiplicity,
					bindingRange: true,
				},
			)
		}
	})
}

func (state *blockScopedVarState) addSyntheticReference(target *ast.Symbol, event referenceEvent) {
	if group := state.groups[target]; group != nil {
		group.references = append(group.references, event)
		return
	}
	state.pendingSynthetic = append(state.pendingSynthetic, pendingSyntheticReference{
		target: target,
		event:  event,
	})
}

func (state *blockScopedVarState) canonicalSymbol(identifier *ast.Node, name string, host *ast.Node) *ast.Symbol {
	if host == nil {
		return utils.BindingNameSymbol(identifier)
	}
	if state.canonical == nil {
		state.canonical = make(map[canonicalKey]*ast.Symbol)
	}
	key := canonicalKey{host: host, name: name}
	if symbol, ok := state.canonical[key]; ok {
		return symbol
	}

	symbol := state.resolve(host, name, ast.SymbolFlagsValue|ast.SymbolFlagsAlias)
	if symbol == nil {
		symbol = utils.BindingNameSymbol(identifier)
	}
	state.canonical[key] = symbol
	return symbol
}

func (state *blockScopedVarState) resolve(location *ast.Node, name string, meaning ast.SymbolFlags) *ast.Symbol {
	if location == nil || name == "" {
		return nil
	}
	return state.resolver.Resolve(location, name, meaning, nil, false, false)
}

func (state *blockScopedVarState) group(symbol *ast.Symbol, host *ast.Node, name string) *varGroup {
	if state.groups == nil {
		state.groups = make(map[*ast.Symbol]*varGroup)
	}
	if group := state.groups[symbol]; group != nil {
		return group
	}
	group := &varGroup{symbol: symbol, name: name, host: host}
	if host != nil && ast.IsFunctionLike(host) && host.Body() != nil {
		group.functionBodyStart = utils.TrimNodeTextRange(state.ctx.SourceFile, host.Body()).Pos()
	}
	for _, declaration := range symbol.Declarations {
		if declaration != nil &&
			(declaration.Kind == ast.KindClassDeclaration || declaration.Kind == ast.KindClassExpression) {
			group.mayHaveClassSelf = true
			break
		}
	}
	state.groups[symbol] = group
	state.groupOrder = append(state.groupOrder, group)
	return group
}

func (state *blockScopedVarState) collectDecoratedClassExpression(node *ast.Node) {
	if node.Name() != nil && len(node.Decorators()) != 0 {
		state.classExpressions = append(state.classExpressions, node)
	}
}

func (state *blockScopedVarState) collectNamespaceExportDeclaration(node *ast.Node) {
	if node.Name() != nil {
		state.namespaceExports = append(state.namespaceExports, node.Name())
	}
}

func (state *blockScopedVarState) collectExportSpecifier(node *ast.Node) {
	if utils.IsReExportSpecifier(node) {
		return
	}
	specifier := node.AsExportSpecifier()
	localName := specifier.Name()
	if specifier.PropertyName != nil {
		localName = specifier.PropertyName
	}
	if localName != nil && localName.Kind == ast.KindIdentifier {
		state.exportSpecifiers = append(state.exportSpecifiers, localName)
	}
}

func (state *blockScopedVarState) collectIndexSignature(node *ast.Node) {
	state.indexSignatures = append(state.indexSignatures, node)
}

func (state *blockScopedVarState) collectJsxNamespacedName(node *ast.Node) {
	if node.Parent == nil || node.Parent.Kind == ast.KindJsxAttribute {
		return
	}
	switch node.Parent.Kind {
	case ast.KindJsxOpeningElement, ast.KindJsxClosingElement, ast.KindJsxSelfClosingElement:
		state.jsxNamespacedNames = append(state.jsxNamespacedNames, node)
	}
}

func (state *blockScopedVarState) finish(_ *ast.Node) {
	if len(state.groupOrder) == 0 {
		return
	}
	relevantNames := make(map[string]struct{})
	for _, group := range state.groupOrder {
		if _, hasClass := state.classNames[group.name]; hasClass {
			group.mayHaveClassSelf = true
		}
		if len(group.occurrences) != 0 {
			relevantNames[group.name] = struct{}{}
		}
	}
	if len(relevantNames) == 0 {
		return
	}
	state.hasConditionalInfer = strings.Contains(state.ctx.SourceFile.Text(), "infer")
	state.collectRelevantFunctions(relevantNames)
	for _, declaration := range state.nonVarDeclarations {
		state.collectNonVarDeclarationWrites(declaration, relevantNames)
	}

	for _, pending := range state.pendingSynthetic {
		if group := state.groups[pending.target]; group != nil && len(group.occurrences) != 0 {
			group.references = append(group.references, pending.event)
		}
	}
	state.addParameterReferences()
	state.addMergedTypeReferences()
	state.addTypeParameterParameterReferences(relevantNames)
	state.addPreBodyReferences(relevantNames)
	state.addDecoratedClassExpressionReferences()
	state.addNamespaceExportReferences()
	state.addExportSpecifierReferences(relevantNames)
	state.addIndexSignatureReferences(relevantNames)
	state.addJsxNamespacedReferences()
	state.addTSXFactoryReferences()

	// Query the shared reference index exactly once per canonical variable.
	// Reassign the rare pre-body resolution mismatch to the true upper target.
	if state.ctx.Refs != nil {
		for _, group := range state.groupOrder {
			if len(group.occurrences) == 0 {
				continue
			}
			for _, identifier := range state.ctx.Refs.References(group.symbol) {
				target := group
				if isTypeOnlyHeritageReference(identifier) {
					// RefStore follows TypeScript's value treatment for a
					// qualified heritage root. TSESTree visits it in type space;
					// addMergedTypeReferences already restored real type targets.
					target = nil
				} else if group.mayHaveClassSelf && state.classScopeShadowsGroup(identifier, group) {
					target = nil
				} else if group.functionBodyStart != 0 &&
					utils.TrimNodeTextRange(state.ctx.SourceFile, identifier).Pos() < group.functionBodyStart {
					target = state.validReferenceTarget(
						group.symbol,
						identifier,
						blockScopedVarReferenceMeaning(identifier),
					)
				}
				state.addIndexedReference(target, identifier)
			}
		}
	}

	var diagnostics []diagnostic
	for _, group := range state.groupOrder {
		if len(group.occurrences) == 0 {
			continue
		}
		slices.SortStableFunc(group.references, func(left, right referenceEvent) int {
			return left.identifier.Pos() - right.identifier.Pos()
		})
		for index := range group.references {
			event := &group.references[index]
			if event.bindingRange {
				event.textRange = utils.GetESTreeBindingIdentifierRange(state.ctx.SourceFile, event.identifier)
			} else {
				event.textRange = utils.TrimNodeTextRange(state.ctx.SourceFile, event.identifier)
			}
		}
		for _, occurrence := range group.occurrences {
			state.appendOccurrenceDiagnostics(&diagnostics, group, occurrence)
		}
	}

	slices.SortStableFunc(diagnostics, func(left, right diagnostic) int {
		return left.textRange.Pos() - right.textRange.Pos()
	})
	for _, item := range diagnostics {
		state.ctx.ReportRange(item.textRange, item.message)
	}
}

// collectRelevantFunctions reuses the binder's source-ordered lexical
// container chain. This avoids a listener callback on every function and
// still includes functions whose header was incorrectly bound to a body-only
// declaration rather than to an outer reportable group.
func (state *blockScopedVarState) collectRelevantFunctions(relevantNames map[string]struct{}) {
	for container := state.ctx.SourceFile.AsNode(); container != nil; {
		functionLike := ast.IsFunctionLike(container)
		typeSignature := isTypeSignatureContainer(container)
		hasParameters := false
		hasTypeParameters := false
		if functionLike || typeSignature {
			hasParameters = len(container.Parameters()) != 0
			hasTypeParameters = len(container.TypeParameters()) != 0
		}
		needsFunction := functionLike && container.Body() != nil &&
			(hasParameters || container.Type() != nil || hasTypeParameters)
		needsParameterTypeScope := hasParameters && hasTypeParameters && isParameterBeforeTypeParameterScope(container)
		hasRelevantLocal := false
		if needsFunction || needsParameterTypeScope {
			hasRelevantLocal = symbolTableContainsAny(container.Locals(), relevantNames)
		}
		if typeSignature && hasParameters {
			state.hasTypeSignatureParameters = true
		}
		if hasRelevantLocal && needsParameterTypeScope {
			state.parameterTypeScopes = append(state.parameterTypeScopes, container)
		}
		if hasRelevantLocal && needsFunction {
			state.functions = append(state.functions, container)
		}
		data := container.LocalsContainerData()
		if data == nil {
			break
		}
		container = data.NextContainer
	}
}

func isTypeSignatureContainer(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindFunctionType,
		ast.KindConstructorType,
		ast.KindCallSignature,
		ast.KindConstructSignature,
		ast.KindMethodSignature,
		ast.KindIndexSignature:
		return true
	}
	return false
}

func isParameterBeforeTypeParameterScope(node *ast.Node) bool {
	return node != nil && (ast.IsFunctionLike(node) || isTypeSignatureContainer(node)) &&
		node.Kind != ast.KindIndexSignature
}

func symbolTableContainsAny(symbols ast.SymbolTable, names map[string]struct{}) bool {
	if len(symbols) < len(names) {
		for name := range symbols {
			if _, exists := names[name]; exists {
				return true
			}
		}
		return false
	}
	for name := range names {
		if symbols[name] != nil {
			return true
		}
	}
	return false
}

func (state *blockScopedVarState) addParameterReferences() {
	for _, function := range state.functions {
		for _, parameter := range function.Parameters() {
			utils.CollectBindingNames(parameter.Name(), func(identifier *ast.Node, name string) {
				canonical := state.canonical[canonicalKey{host: function, name: name}]
				group := state.groups[canonical]
				if group == nil {
					return
				}
				multiplicity := parameterWriteMultiplicity(identifier, parameter)
				if multiplicity != 0 {
					group.references = append(group.references, referenceEvent{
						identifier:   identifier,
						multiplicity: multiplicity,
						bindingRange: true,
					})
				}
			})
		}
	}
}

// addPreBodyReferences repairs a tsgo resolver edge case without putting a
// Parameter listener on every file. NameResolver can expose a function-body
// declaration to an identifier in that function's header when another header
// expression requires a scope change. TSESTree rejects such a target unless
// the merged variable has at least one definition before the body.
func (state *blockScopedVarState) addPreBodyReferences(relevantNames map[string]struct{}) {
	if state.ctx.Refs == nil || len(state.functions) == 0 {
		return
	}

	for _, host := range state.functions {
		visit := func(root *ast.Node) {
			utils.VisitDescendants(root, func(identifier *ast.Node) bool {
				if identifier.Kind != ast.KindIdentifier || utils.IsNonReferenceIdentifier(identifier) {
					return true
				}
				if _, relevant := relevantNames[identifier.Text()]; !relevant {
					return true
				}
				symbol := state.ctx.Refs.ResolveInFile(identifier)
				meaning := blockScopedVarReferenceMeaning(identifier)
				if symbol != nil {
					symbol = state.repairClassExpressionDecoratorTarget(identifier, symbol, meaning)
				}
				symbol = state.preBodyMergedSymbol(identifier, symbol, meaning)
				state.addOrdinaryReference(state.validReferenceTarget(symbol, identifier, meaning), identifier)
				return true
			})
		}
		for _, parameter := range host.Parameters() {
			visit(parameter)
		}
		visit(host.Type())
		for _, typeParameter := range host.TypeParameters() {
			visit(typeParameter)
		}
	}
}

// The TSESTree referencer defines every parameter before it visits a
// function's type parameters, even though type parameters appear first in
// source. Redirect value references in constraints/defaults to that parameter
// variable (and its canonical var group, when one exists).
func (state *blockScopedVarState) addTypeParameterParameterReferences(relevantNames map[string]struct{}) {
	for _, scope := range state.parameterTypeScopes {
		parameterNames := make(map[string]struct{})
		for _, parameter := range scope.Parameters() {
			utils.CollectBindingNames(parameter.Name(), func(_ *ast.Node, name string) {
				if _, relevant := relevantNames[name]; relevant {
					parameterNames[name] = struct{}{}
				}
			})
		}
		if len(parameterNames) == 0 {
			continue
		}
		shadowed := make(map[string]int)
		for _, typeParameter := range scope.TypeParameters() {
			state.visitTypeParameterParameterReferences(typeParameter, scope, parameterNames, shadowed)
		}
	}
}

func (state *blockScopedVarState) visitTypeParameterParameterReferences(
	node *ast.Node,
	host *ast.Node,
	parameterNames map[string]struct{},
	shadowed map[string]int,
) {
	if node == nil {
		return
	}

	if isFunctionTypeScope(node) {
		// A method signature's computed key is evaluated outside its function
		// type scope. Its parameters, type parameters, and return type are
		// evaluated after all value parameters have been defined.
		if node.Kind == ast.KindMethodSignature {
			if name := node.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
				state.visitTypeParameterParameterReferences(
					name.AsComputedPropertyName().Expression,
					host,
					parameterNames,
					shadowed,
				)
			}
		}

		var added []string
		for _, parameter := range node.Parameters() {
			utils.CollectBindingNames(parameter.Name(), func(_ *ast.Node, name string) {
				if _, relevant := parameterNames[name]; !relevant {
					return
				}
				shadowed[name]++
				added = append(added, name)
			})
		}
		for _, parameter := range node.Parameters() {
			state.visitTypeParameterParameterReferences(parameter.Type(), host, parameterNames, shadowed)
		}
		for _, typeParameter := range node.TypeParameters() {
			state.visitTypeParameterParameterReferences(typeParameter, host, parameterNames, shadowed)
		}
		state.visitTypeParameterParameterReferences(node.Type(), host, parameterNames, shadowed)
		for _, name := range added {
			shadowed[name]--
		}
		return
	}

	if node.Kind == ast.KindIdentifier && !utils.IsNonReferenceIdentifier(node) {
		name := node.Text()
		if _, parameter := parameterNames[name]; parameter && referenceMeaningIncludesValue(blockScopedVarReferenceMeaning(node)) {
			if shadowed[name] != 0 {
				state.suppressOrdinaryReference(node)
				return
			}
			canonical := state.canonical[canonicalKey{host: host, name: name}]
			if group := state.groups[canonical]; group != nil {
				state.addOrdinaryReference(group, node)
			} else {
				state.suppressOrdinaryReference(node)
			}
		}
		return
	}

	node.ForEachChild(func(child *ast.Node) bool {
		state.visitTypeParameterParameterReferences(child, host, parameterNames, shadowed)
		return false
	})
}

func isFunctionTypeScope(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindFunctionType,
		ast.KindConstructorType,
		ast.KindCallSignature,
		ast.KindConstructSignature,
		ast.KindMethodSignature:
		return true
	}
	return false
}

func referenceMeaningIncludesValue(meaning ast.SymbolFlags) bool {
	requested := meaning &^ ast.SymbolFlagsAlias
	return requested&ast.SymbolFlagsValue == ast.SymbolFlagsValue ||
		requested == ast.SymbolFlagsFunctionScopedVariable
}

func (state *blockScopedVarState) preBodyMergedSymbol(identifier *ast.Node, symbol *ast.Symbol, meaning ast.SymbolFlags) *ast.Symbol {
	if identifier == nil {
		return symbol
	}
	referenceStart := utils.TrimNodeTextRange(state.ctx.SourceFile, identifier).Pos()
	for current := identifier.Parent; current != nil; current = current.Parent {
		if !ast.IsFunctionLike(current) || current.Body() == nil ||
			referenceStart >= utils.TrimNodeTextRange(state.ctx.SourceFile, current.Body()).Pos() {
			continue
		}
		if methodParameterDecoratorForReference(identifier, current) != nil {
			continue
		}
		canonical := state.canonical[canonicalKey{host: current, name: identifier.Text()}]
		group := state.groups[canonical]
		if group == nil || !groupSupportsMeaning(group, meaning) {
			continue
		}
		if symbol == group.symbol || !symbolHasDefinitionInsideNode(symbol, current) {
			return group.symbol
		}
		return symbol
	}
	return symbol
}

func groupSupportsMeaning(group *varGroup, meaning ast.SymbolFlags) bool {
	if group == nil {
		return false
	}
	requested := meaning &^ ast.SymbolFlagsAlias
	return group.symbol.Flags&requested != 0 ||
		requested&ast.SymbolFlagsType == ast.SymbolFlagsType &&
			group.symbol.Flags&ast.SymbolFlagsNamespace != 0
}

func symbolHasDefinitionInsideNode(symbol *ast.Symbol, node *ast.Node) bool {
	if symbol == nil || node == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration != nil && declaration.Pos() >= node.Pos() && declaration.End() <= node.End() {
			return true
		}
	}
	return false
}

func (state *blockScopedVarState) addOrdinaryReference(group *varGroup, identifier *ast.Node) {
	if group == nil || identifier == nil || state.isIgnoredOrdinaryReference(identifier) {
		return
	}
	if state.ordinarySeen == nil {
		state.ordinarySeen = make(map[*ast.Node]*varGroup)
	}
	if _, seen := state.ordinarySeen[identifier]; seen {
		return
	}
	state.ordinarySeen[identifier] = group
	state.appendOrdinaryReference(group, identifier)
}

func (state *blockScopedVarState) addIndexedReference(group *varGroup, identifier *ast.Node) {
	if _, handled := state.ordinarySeen[identifier]; handled {
		return
	}
	if group == nil || identifier == nil || state.isIgnoredOrdinaryReference(identifier) {
		return
	}
	state.appendOrdinaryReference(group, identifier)
}

func (state *blockScopedVarState) suppressOrdinaryReference(identifier *ast.Node) {
	if identifier == nil {
		return
	}
	if state.ordinarySeen == nil {
		state.ordinarySeen = make(map[*ast.Node]*varGroup)
	}
	if _, handled := state.ordinarySeen[identifier]; !handled {
		state.ordinarySeen[identifier] = nil
	}
}

func (state *blockScopedVarState) isIgnoredOrdinaryReference(identifier *ast.Node) bool {
	return isESLintIntrinsicJsxTag(identifier) ||
		isIgnoredImportTypeQualifier(identifier) ||
		isIgnoredCompoundPatternReference(identifier) ||
		state.hasConditionalInfer && isConditionalInferShadowed(identifier) ||
		state.hasTypeSignatureParameters && isIgnoredFunctionTypeParameterReference(identifier) ||
		state.isIgnoredEspreeClosingJsxTag(identifier)
}

// ESTree exposes a direct array/object left-hand side as a Pattern only for a
// simple `=` assignment. For a parser-accepted compound assignment, the TSE
// referencer intentionally skips that malformed pattern subtree altogether.
func isIgnoredCompoundPatternReference(identifier *ast.Node) bool {
	for current := identifier; current != nil && current.Parent != nil; current = current.Parent {
		parent := current.Parent
		if parent.Kind != ast.KindBinaryExpression {
			continue
		}
		binary := parent.AsBinaryExpression()
		if binary.Left != current || binary.OperatorToken == nil ||
			!ast.IsCompoundAssignmentOperator(binary.OperatorToken.Kind) {
			continue
		}
		return current.Kind == ast.KindArrayLiteralExpression ||
			current.Kind == ast.KindObjectLiteralExpression
	}
	return false
}

// TSESTree keeps infer bindings in the surrounding conditional-type scope.
// Resolution happens when that scope closes, so an infer appearing later in
// the check/extends/true region shadows same-named outer variables throughout
// that whole region. The false branch is visited after the scope closes.
func isConditionalInferShadowed(identifier *ast.Node) bool {
	if identifier == nil {
		return false
	}
	if referenceMeaningIncludesValue(blockScopedVarReferenceMeaning(identifier)) {
		return false
	}
	for current := identifier; current != nil && current.Parent != nil; current = current.Parent {
		parent := current.Parent
		if parent.Kind != ast.KindConditionalType {
			continue
		}
		conditional := parent.AsConditionalTypeNode()
		if conditional.FalseType == current {
			continue
		}
		shadowed := false
		utils.VisitDescendants(parent, func(node *ast.Node) bool {
			if shadowed || node.Kind != ast.KindInferType {
				return true
			}
			declaration := node.AsInferTypeNode().TypeParameter.AsTypeParameterDeclaration()
			if declaration.Name().Text() == identifier.Text() && inferTargetsConditional(node, parent) {
				shadowed = true
			}
			return true
		})
		if shadowed {
			return true
		}
	}
	return false
}

func inferTargetsConditional(inferType *ast.Node, target *ast.Node) bool {
	for current := inferType; current != nil && current.Parent != nil; current = current.Parent {
		parent := current.Parent
		if parent.Kind != ast.KindConditionalType {
			continue
		}
		conditional := parent.AsConditionalTypeNode()
		if conditional.FalseType == current {
			continue
		}
		return parent == target
	}
	return false
}

func (state *blockScopedVarState) isIgnoredEspreeClosingJsxTag(identifier *ast.Node) bool {
	if state.ctx.SourceFile.ScriptKind != core.ScriptKindJS &&
		state.ctx.SourceFile.ScriptKind != core.ScriptKindJSX {
		return false
	}
	tagName := identifier
	for tagName.Parent != nil && tagName.Parent.Kind == ast.KindPropertyAccessExpression &&
		tagName.Parent.AsPropertyAccessExpression().Expression == tagName {
		tagName = tagName.Parent
	}
	return tagName.Parent != nil && tagName.Parent.Kind == ast.KindJsxClosingElement
}

func (state *blockScopedVarState) appendOrdinaryReference(group *varGroup, identifier *ast.Node) {
	group.references = append(group.references, referenceEvent{
		identifier:   identifier,
		multiplicity: ordinaryReferenceMultiplicity(identifier),
	})
}

func isESLintIntrinsicJsxTag(identifier *ast.Node) bool {
	if identifier == nil || !ast.IsJsxTagName(identifier) {
		return false
	}
	name := identifier.Text()
	if name == "" || name == "this" {
		return false
	}
	if name[0] < utf8.RuneSelf {
		return name[0] >= 'a' && name[0] <= 'z'
	}
	first, size := utf8.DecodeRuneInString(name)
	if first > 0xffff {
		// JavaScript indexes strings by UTF-16 code unit. An astral tag starts
		// with an uncased high surrogate and is therefore component-like.
		return false
	}
	// ESLint's Node runtime follows Unicode 16.0, while the Go/x-text tables
	// used by this build predate the two new uppercase mappings below.
	if first == '\u019b' || first == '\u0264' {
		return true
	}
	firstText := name[:size]
	return cases.Upper(language.Und).String(firstText) != firstText
}

func (state *blockScopedVarState) validReferenceTarget(symbol *ast.Symbol, identifier *ast.Node, meaning ast.SymbolFlags) *varGroup {
	for symbol != nil {
		host := invalidFunctionBodyResolution(state.ctx.SourceFile, identifier, symbol)
		if host == nil {
			group := state.groups[symbol]
			if state.classScopeShadowsGroup(identifier, group) {
				return nil
			}
			return group
		}
		if host.Kind == ast.KindFunctionExpression && host.Name() != nil && host.Name().Text() == identifier.Text() {
			// FunctionExpressionNameScope sits between the function and its
			// lexical parent; it is never a var group owned by this rule.
			return nil
		}
		symbol = state.resolve(host.Parent, identifier.Text(), meaning)
	}
	return nil
}

// A named class has a second, class-local scope-manager variable for its
// self-reference. TypeScript can merge that name with an outer var symbol, so
// separate references in the class scope from the outer reportable group.
// Class-level decorators are visited before the class scope exists.
func (state *blockScopedVarState) classScopeShadowsGroup(identifier *ast.Node, group *varGroup) bool {
	if identifier == nil || group == nil {
		return false
	}
	for current := identifier.Parent; current != nil; current = current.Parent {
		if current.Kind != ast.KindClassDeclaration && current.Kind != ast.KindClassExpression {
			continue
		}
		if current.Name() == nil || current.Name().Text() != group.name ||
			isInsideClassLevelDecorator(identifier, current) {
			continue
		}
		if !nodeIsInside(group.host, current) {
			return true
		}
	}
	return false
}

func isInsideClassLevelDecorator(node *ast.Node, class *ast.Node) bool {
	for current := node.Parent; current != nil && current != class; current = current.Parent {
		if current.Kind == ast.KindDecorator && current.Parent == class {
			return true
		}
	}
	return false
}

func nodeIsInside(node *ast.Node, ancestor *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		if current == ancestor {
			return true
		}
	}
	return false
}

func invalidFunctionBodyResolution(sourceFile *ast.SourceFile, identifier *ast.Node, symbol *ast.Symbol) *ast.Node {
	if identifier == nil || symbol == nil {
		return nil
	}
	referenceStart := utils.TrimNodeTextRange(sourceFile, identifier).Pos()
	for current := identifier.Parent; current != nil; current = current.Parent {
		if !ast.IsFunctionLike(current) || current.Body() == nil {
			continue
		}
		if decorator := methodParameterDecoratorForReference(identifier, current); decorator != nil {
			if !symbolHasDefinitionInsideNode(symbol, decorator) && symbolHasDefinitionInsideNode(symbol, current) {
				return current
			}
		}
		bodyRange := utils.TrimNodeTextRange(sourceFile, current.Body())
		if referenceStart >= bodyRange.Pos() {
			continue
		}
		if symbolDefinitionsAreInsideRange(sourceFile, symbol, bodyRange) {
			return current
		}
	}
	return nil
}

func methodParameterDecoratorForReference(identifier *ast.Node, function *ast.Node) *ast.Node {
	if identifier == nil || function == nil {
		return nil
	}
	switch function.Kind {
	case ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
	default:
		return nil
	}
	for current := identifier.Parent; current != nil && current != function; current = current.Parent {
		if current.Kind == ast.KindDecorator && current.Parent != nil &&
			current.Parent.Kind == ast.KindParameter && current.Parent.Parent == function {
			return current
		}
	}
	return nil
}

func symbolDefinitionsAreInsideRange(sourceFile *ast.SourceFile, symbol *ast.Symbol, textRange core.TextRange) bool {
	if len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil {
			return false
		}
		definition := declaration.Name()
		if definition == nil {
			definition = declaration
		}
		definitionRange := utils.TrimNodeTextRange(sourceFile, definition)
		if definitionRange.Pos() < textRange.Pos() || definitionRange.End() > textRange.End() {
			return false
		}
	}
	return true
}

func (state *blockScopedVarState) appendOccurrenceDiagnostics(diagnostics *[]diagnostic, group *varGroup, occurrence varOccurrence) {
	var message rule.RuleMessage
	messageReady := false

	for _, event := range group.references {
		if event.textRange.Pos() >= occurrence.scopeRange.Pos() && event.textRange.End() <= occurrence.scopeRange.End() {
			continue
		}
		if !messageReady {
			definitionRange := utils.TrimNodeTextRange(state.ctx.SourceFile, occurrence.identifier)
			lineIndex, characterIndex := scanner.GetECMALineAndUTF16CharacterOfPosition(state.ctx.SourceFile, definitionRange.Pos())
			message = rule.RuleMessage{
				Id: "outOfScope",
				Description: fmt.Sprintf(
					"'%s' declared on line %d column %d is used outside of binding context.",
					group.name,
					lineIndex+1,
					int(characterIndex)+1,
				),
			}
			messageReady = true
		}
		for range event.multiplicity {
			*diagnostics = append(*diagnostics, diagnostic{textRange: event.textRange, message: message})
		}
	}
}

type diagnostic struct {
	textRange core.TextRange
	message   rule.RuleMessage
}

func variableDeclarationWriteMultiplicity(declaration *ast.Node, identifier *ast.Node, isForInOrOf bool) int {
	multiplicity := 0
	if declaration.AsVariableDeclaration().Initializer != nil {
		multiplicity++
	}
	if isForInOrOf {
		multiplicity++
	}
	for current := identifier.Parent; current != nil && current != declaration; current = current.Parent {
		if current.Kind == ast.KindBindingElement && current.AsBindingElement().Initializer != nil {
			multiplicity++
		}
	}
	return multiplicity
}

func parameterWriteMultiplicity(identifier *ast.Node, parameter *ast.Node) int {
	multiplicity := 0
	if parameter.AsParameterDeclaration().Initializer != nil {
		multiplicity++
	}
	for current := identifier.Parent; current != nil && current != parameter; current = current.Parent {
		if current.Kind == ast.KindBindingElement && current.AsBindingElement().Initializer != nil {
			multiplicity++
		}
	}
	return multiplicity
}

func ordinaryReferenceMultiplicity(identifier *ast.Node) int {
	if !utils.IsWriteReference(identifier) {
		return 1
	}

	multiplicity := 1
	for current := identifier; current != nil && current.Parent != nil; current = current.Parent {
		parent := current.Parent
		if parent.Kind == ast.KindComputedPropertyName {
			break
		}

		if parent.Kind == ast.KindShorthandPropertyAssignment {
			assignment := parent.AsShorthandPropertyAssignment()
			if assignment.ObjectAssignmentInitializer == current {
				break
			}
			if assignment.Name() == current && assignment.ObjectAssignmentInitializer != nil &&
				utils.IsInDestructuringAssignment(parent) {
				multiplicity++
			}
			continue
		}

		if parent.Kind == ast.KindBinaryExpression {
			binary := parent.AsBinaryExpression()
			if binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindEqualsToken {
				continue
			}
			if binary.Right == current {
				break
			}
			if binary.Left == current && utils.IsDefaultValueInDestructuringAssignment(parent) {
				multiplicity++
			}
		}
	}
	return multiplicity
}

// blockScopeRange returns the nearest binding context ESLint puts on the
// rule's scope stack for a var declaration.
func blockScopeRange(sourceFile *ast.SourceFile, declarationList *ast.Node) core.TextRange {
	for current := declarationList.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindBlock,
			ast.KindForStatement,
			ast.KindForInStatement,
			ast.KindForOfStatement,
			ast.KindSwitchStatement:
			return utils.TrimNodeTextRange(sourceFile, current)
		case ast.KindSourceFile:
			return core.NewTextRange(0, current.End())
		}
	}
	return core.NewTextRange(0, sourceFile.AsNode().End())
}
