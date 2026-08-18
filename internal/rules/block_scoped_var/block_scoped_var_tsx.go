package block_scoped_var

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// blockScopedVarReferenceMeaning mirrors RefStore's declaration-space model.
// It stays rule-local because block-scoped-var's performance guardrail limits
// production changes to this package.
func blockScopedVarReferenceMeaning(identifier *ast.Node) ast.SymbolFlags {
	if identifier == nil {
		return 0
	}
	parent := identifier.Parent
	if parent != nil && parent.Kind == ast.KindTypePredicate {
		return ast.SymbolFlagsFunctionScopedVariable
	}
	if parent != nil && parent.Kind == ast.KindExportAssignment {
		return ast.SymbolFlagsValue | ast.SymbolFlagsType | ast.SymbolFlagsNamespace | ast.SymbolFlagsAlias
	}
	if parent != nil && parent.Kind == ast.KindExportSpecifier {
		if ast.IsTypeOnlyImportOrExportDeclaration(parent) {
			return ast.SymbolFlagsType | ast.SymbolFlagsNamespace | ast.SymbolFlagsAlias
		}
		return ast.SymbolFlagsValue | ast.SymbolFlagsType | ast.SymbolFlagsNamespace | ast.SymbolFlagsAlias
	}
	entity := identifier
	for entity.Parent != nil && entity.Parent.Kind == ast.KindQualifiedName {
		entity = entity.Parent
	}
	if entity.Parent != nil && entity.Parent.Kind == ast.KindImportEqualsDeclaration &&
		entity.Parent.AsImportEqualsDeclaration().ModuleReference == entity {
		return ast.SymbolFlagsValue | ast.SymbolFlagsType | ast.SymbolFlagsNamespace | ast.SymbolFlagsAlias
	}
	if ast.IsPartOfTypeQuery(identifier) {
		return ast.SymbolFlagsValue | ast.SymbolFlagsAlias
	}
	if isTypeOnlyDecoratorQualifier(identifier) {
		return ast.SymbolFlagsNamespace | ast.SymbolFlagsAlias
	}
	if ast.IsExpressionNode(identifier) {
		return ast.SymbolFlagsValue | ast.SymbolFlagsAlias
	}
	if identifier.Parent != nil && identifier.Parent.Kind == ast.KindQualifiedName &&
		identifier.Parent.AsQualifiedName().Left == identifier {
		return ast.SymbolFlagsNamespace | ast.SymbolFlagsAlias
	}
	if ast.IsPartOfTypeNode(identifier) {
		return ast.SymbolFlagsType | ast.SymbolFlagsAlias
	}
	return ast.SymbolFlagsValue | ast.SymbolFlagsAlias
}

func isTypeOnlyDecoratorQualifier(identifier *ast.Node) bool {
	entity := identifier
	for entity.Parent != nil && entity.Parent.Kind == ast.KindPropertyAccessExpression {
		access := entity.Parent.AsPropertyAccessExpression()
		if access.Expression != entity {
			break
		}
		entity = entity.Parent
	}
	return entity != identifier && ast.IsPartOfTypeNode(entity)
}

// TypeVisitor intentionally ignores an import type's qualifier. Its type
// arguments are visited separately and therefore remain ordinary references.
func isIgnoredImportTypeQualifier(identifier *ast.Node) bool {
	qualifier := identifier
	for qualifier != nil && qualifier.Parent != nil && qualifier.Parent.Kind == ast.KindQualifiedName {
		qualifier = qualifier.Parent
	}
	if qualifier == nil || qualifier.Parent == nil || qualifier.Parent.Kind != ast.KindImportType {
		return false
	}
	return qualifier.Parent.AsImportTypeNode().Qualifier == qualifier
}

// TypeVisitor defines function-type parameter bindings and visits their type
// annotations, but intentionally skips decorators, binding-pattern right-hand
// nodes, and initializers. Index signatures use a different visitor.
func isIgnoredFunctionTypeParameterReference(identifier *ast.Node) bool {
	for current := identifier.Parent; current != nil; current = current.Parent {
		if current.Kind != ast.KindParameter || current.Parent == nil {
			continue
		}
		switch current.Parent.Kind {
		case ast.KindFunctionType,
			ast.KindConstructorType,
			ast.KindCallSignature,
			ast.KindConstructSignature,
			ast.KindMethodSignature,
			ast.KindIndexSignature:
			return !nodeIsInside(identifier, current.Type())
		}
	}
	return false
}

func (state *blockScopedVarState) addDecoratedClassExpressionReferences() {
	for _, classExpression := range state.classExpressions {
		classSymbol := classExpression.Symbol()
		if classSymbol == nil || classExpression.Parent == nil {
			continue
		}
		for _, decorator := range classExpression.Decorators() {
			utils.VisitDescendants(decorator, func(identifier *ast.Node) bool {
				if identifier.Kind != ast.KindIdentifier || state.ctx.Refs == nil ||
					state.ctx.Refs.ResolveInFile(identifier) != classSymbol {
					return true
				}
				meaning := blockScopedVarReferenceMeaning(identifier)
				targetSymbol := state.resolve(classExpression.Parent, identifier.Text(), meaning)
				state.addOrdinaryReference(state.validReferenceTarget(targetSymbol, identifier, meaning), identifier)
				return true
			})
		}
	}
}

func (state *blockScopedVarState) addNamespaceExportReferences() {
	meaning := ast.SymbolFlagsValue | ast.SymbolFlagsAlias
	for _, identifier := range state.namespaceExports {
		target := state.resolve(identifier, identifier.Text(), meaning)
		state.addOrdinaryReference(state.validReferenceTarget(target, identifier, meaning), identifier)
	}
}

func (state *blockScopedVarState) addExportSpecifierReferences(relevantNames map[string]struct{}) {
	for _, identifier := range state.exportSpecifiers {
		if _, relevant := relevantNames[identifier.Text()]; !relevant {
			continue
		}
		specifier := identifier.Parent
		if specifier == nil || specifier.Kind != ast.KindExportSpecifier ||
			specifier.Parent == nil || specifier.Parent.Parent == nil {
			continue
		}
		exportDeclaration := specifier.Parent.Parent
		meaning := blockScopedVarReferenceMeaning(identifier)
		target := state.resolve(exportDeclaration.Parent, identifier.Text(), meaning)
		group := state.validReferenceTarget(target, identifier, meaning)
		if !groupSupportsMeaning(group, meaning) {
			group = nil
		}
		if group == nil && meaning&(ast.SymbolFlagsType|ast.SymbolFlagsNamespace) != 0 {
			group = state.mergedGroupForTypeTarget(identifier, target)
		}
		state.addOrdinaryReference(group, identifier)
	}
}

// TypeVisitor visits an index signature's parameter and return types without
// defining the index parameter as a scope variable. TypeScript's binder does
// define it, so resolve those visited identifiers from outside the signature.
func (state *blockScopedVarState) addIndexSignatureReferences(relevantNames map[string]struct{}) {
	if state.ctx.Refs == nil {
		return
	}
	for _, signature := range state.indexSignatures {
		parameterSymbols := make(map[*ast.Symbol]struct{})
		for _, parameter := range signature.Parameters() {
			utils.CollectBindingNames(parameter.Name(), func(identifier *ast.Node, _ string) {
				if symbol := utils.BindingNameSymbol(identifier); symbol != nil {
					parameterSymbols[symbol] = struct{}{}
				}
			})
		}
		visitType := func(typeNode *ast.Node) {
			utils.VisitDescendants(typeNode, func(identifier *ast.Node) bool {
				if identifier.Kind != ast.KindIdentifier || utils.IsNonReferenceIdentifier(identifier) {
					return true
				}
				if _, relevant := relevantNames[identifier.Text()]; !relevant {
					return true
				}
				if _, boundToIndexParameter := parameterSymbols[state.ctx.Refs.ResolveInFile(identifier)]; !boundToIndexParameter {
					return true
				}
				meaning := blockScopedVarReferenceMeaning(identifier)
				target := state.resolve(signature.Parent, identifier.Text(), meaning)
				group := state.validReferenceTarget(target, identifier, meaning)
				if group == nil && meaning&(ast.SymbolFlagsType|ast.SymbolFlagsNamespace) != 0 {
					group = state.mergedGroupForTypeTarget(identifier, target)
				}
				state.addOrdinaryReference(group, identifier)
				return true
			})
		}
		for _, parameter := range signature.Parameters() {
			visitType(parameter.Type())
		}
		visitType(signature.Type())
	}
}

// addMergedTypeReferences fills one scope-model gap in RefStore. TSESTree
// keeps value and type definitions in one scope variable, while TypeScript can
// represent an exported declaration with separate local/export symbols.
func (state *blockScopedVarState) addMergedTypeReferences() {
	if state.ctx.Refs == nil {
		return
	}
	implicitTypeGroups := state.implicitTypeGlobalGroups()
	var names map[string]struct{}
	for _, group := range state.groupOrder {
		if len(group.occurrences) == 0 {
			continue
		}
		typeTarget := state.resolve(
			group.host,
			group.name,
			ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias,
		)
		if symbolHasTSETypeDefinitionInHost(typeTarget, group.host) || implicitTypeGroups[group.name] == group {
			if names == nil {
				names = make(map[string]struct{})
			}
			names[group.name] = struct{}{}
		}
	}
	if len(names) == 0 {
		return
	}

	utils.VisitDescendants(state.ctx.SourceFile.AsNode(), func(identifier *ast.Node) bool {
		if identifier.Kind != ast.KindIdentifier || utils.IsNonReferenceIdentifier(identifier) {
			return true
		}
		if _, relevant := names[identifier.Text()]; !relevant {
			return true
		}
		resolved := state.ctx.Refs.ResolveInFile(identifier)
		meaning := blockScopedVarReferenceMeaning(identifier)
		typeOnlyHeritage := isTypeOnlyHeritageReference(identifier)
		if typeOnlyHeritage {
			meaning = ast.SymbolFlagsType | ast.SymbolFlagsNamespace | ast.SymbolFlagsAlias
		} else if meaning&(ast.SymbolFlagsType|ast.SymbolFlagsNamespace) == 0 {
			return true
		}
		target := resolved
		if typeOnlyHeritage {
			target = state.resolve(identifier, identifier.Text(), meaning)
		} else if target == nil {
			target = state.ctx.Refs.ResolveInFileWithMeaning(
				identifier,
				ast.SymbolFlagsType|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias,
			)
		}
		if implicitGroup := implicitTypeGroups[identifier.Text()]; implicitGroup != nil &&
			(target == nil || target == implicitGroup.symbol || state.isDefaultLibrarySymbol(target)) {
			state.addOrdinaryReference(implicitGroup, identifier)
			return true
		}
		if state.groups[resolved] != nil && !typeOnlyHeritage {
			return true
		}
		target = state.repairClassExpressionDecoratorTarget(identifier, target, meaning)
		target = state.preBodyMergedSymbol(identifier, target, meaning)
		group := state.validReferenceTarget(target, identifier, meaning)
		if group == nil {
			group = state.mergedGroupForTypeTarget(identifier, target)
		}
		state.addOrdinaryReference(group, identifier)
		return true
	})
}

func (state *blockScopedVarState) implicitTypeGlobalGroups() map[string]*varGroup {
	if state.ctx.SourceFile == nil ||
		(state.ctx.SourceFile.ScriptKind != core.ScriptKindTS && state.ctx.SourceFile.ScriptKind != core.ScriptKindTSX) ||
		!ast.IsGlobalSourceFile(state.ctx.SourceFile.AsNode()) {
		return nil
	}

	activeTypeGlobals := make(map[string]bool)
	sourceProgram := state.ctx.Program()
	if sourceProgram != nil && state.ctx.TypeChecker != nil {
		utils.AddDefaultLibraryTypeGlobalNames(activeTypeGlobals, sourceProgram, state.ctx.TypeChecker)
	}
	// Without a Program, scope-manager 8.65 seeds its generated esnext
	// catalog. With a Program it derives the catalog from target and, notably,
	// ignores noLib; reuse that same public catalog for the ESNext/noLib case
	// that the checker omits.
	defaultESNext := sourceProgram == nil ||
		(sourceProgram.Options().NoLib.IsTrue() &&
			sourceProgram.Options().GetEmitScriptTarget() == core.ScriptTargetESNext)
	if defaultESNext {
		for _, group := range state.groupOrder {
			if rule.IsDefaultTypeScriptTypeGlobal(group.name) {
				activeTypeGlobals[group.name] = true
			}
		}
	}
	// ESLint finalizes the parser scope with ECMAScript, configured, and
	// inline globals without a declaration-space distinction. These injected
	// names are additive: an authored `off` prevents injection, but cannot
	// remove a type variable the TypeScript parser already supplied.
	injectedGlobals := make(map[string]bool)
	state.ctx.Globals.ApplyTo(injectedGlobals)
	for name := range injectedGlobals {
		activeTypeGlobals[name] = true
	}

	var result map[string]*varGroup
	for _, group := range state.groupOrder {
		if len(group.occurrences) == 0 || group.host != state.ctx.SourceFile.AsNode() ||
			!activeTypeGlobals[group.name] {
			continue
		}
		if result == nil {
			result = make(map[string]*varGroup)
		}
		result[group.name] = group
	}
	return result
}

func (state *blockScopedVarState) isDefaultLibrarySymbol(symbol *ast.Symbol) bool {
	sourceProgram := state.ctx.Program()
	return sourceProgram != nil && utils.IsSymbolFromDefaultLibrary(sourceProgram, symbol)
}

// mergedGroupForTypeTarget joins TypeScript's local/export symbol pair without
// conflating identically named variables from reopened TSModule scopes. Both
// symbols retain the same declaration; the lexical host selects the matching
// scope-manager variable.
func (state *blockScopedVarState) mergedGroupForTypeTarget(identifier *ast.Node, target *ast.Symbol) *varGroup {
	if identifier == nil || target == nil {
		return nil
	}
	for host := utils.FindEnclosingScope(identifier); host != nil; host = utils.FindEnclosingScope(host) {
		canonical := state.canonical[canonicalKey{host: host, name: identifier.Text()}]
		group := state.groups[canonical]
		if group != nil && symbolHasTSETypeDefinitionInHost(target, host) {
			if state.classScopeShadowsGroup(identifier, group) {
				return nil
			}
			return group
		}
	}
	return nil
}

func symbolHasTSETypeDefinitionInHost(symbol *ast.Symbol, host *ast.Node) bool {
	if symbol == nil || host == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration != nil && isTSETypeDefinitionKind(declaration.Kind) &&
			tseDefinitionScopeHost(declaration) == host {
			return true
		}
	}
	return false
}

func tseDefinitionScopeHost(declaration *ast.Node) *ast.Node {
	for current := declaration.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindBlock:
			parent := current.Parent
			if parent != nil && ast.IsFunctionLike(parent) && parent.Body() == current {
				return parent
			}
			if parent != nil && parent.Kind == ast.KindClassStaticBlockDeclaration {
				return parent
			}
			return current
		case ast.KindSourceFile,
			ast.KindModuleBlock,
			ast.KindClassDeclaration,
			ast.KindClassExpression,
			ast.KindClassStaticBlockDeclaration,
			ast.KindForStatement,
			ast.KindForInStatement,
			ast.KindForOfStatement,
			ast.KindSwitchStatement,
			ast.KindCatchClause,
			ast.KindTypeAliasDeclaration,
			ast.KindInterfaceDeclaration,
			ast.KindMappedType,
			ast.KindConditionalType,
			ast.KindFunctionType,
			ast.KindConstructorType,
			ast.KindCallSignature,
			ast.KindConstructSignature,
			ast.KindMethodSignature:
			return current
		}
		if ast.IsFunctionLike(current) {
			return current
		}
	}
	return nil
}

func isTSETypeDefinitionKind(kind ast.Kind) bool {
	switch kind {
	case ast.KindModuleDeclaration,
		ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration,
		ast.KindTypeParameter,
		ast.KindClassDeclaration,
		ast.KindEnumDeclaration,
		ast.KindImportClause,
		ast.KindImportSpecifier,
		ast.KindNamespaceImport,
		ast.KindImportEqualsDeclaration:
		return true
	}
	return false
}

func isTypeOnlyHeritageReference(identifier *ast.Node) bool {
	entity := identifier
	for entity.Parent != nil && entity.Parent.Kind == ast.KindPropertyAccessExpression &&
		entity.Parent.AsPropertyAccessExpression().Expression == entity {
		entity = entity.Parent
	}
	if entity.Parent == nil || entity.Parent.Kind != ast.KindExpressionWithTypeArguments ||
		entity.Parent.AsExpressionWithTypeArguments().Expression != entity {
		return false
	}
	heritage := entity.Parent.Parent
	if heritage == nil || heritage.Kind != ast.KindHeritageClause {
		return false
	}
	return heritage.AsHeritageClause().Token == ast.KindImplementsKeyword ||
		(heritage.Parent != nil && heritage.Parent.Kind == ast.KindInterfaceDeclaration)
}

func (state *blockScopedVarState) addJsxNamespacedReferences() {
	for _, node := range state.jsxNamespacedNames {
		name := node.AsJsxNamespacedName()
		for _, identifier := range []*ast.Node{name.Namespace, name.Name()} {
			targetSymbol := state.resolve(identifier, identifier.Text(), ast.SymbolFlagsValue|ast.SymbolFlagsAlias)
			targetSymbol = state.repairClassExpressionDecoratorTarget(identifier, targetSymbol, ast.SymbolFlagsValue|ast.SymbolFlagsAlias)
			targetSymbol = state.preBodyMergedSymbol(identifier, targetSymbol, ast.SymbolFlagsValue|ast.SymbolFlagsAlias)
			state.addOrdinaryReference(
				state.validReferenceTarget(targetSymbol, identifier, ast.SymbolFlagsValue|ast.SymbolFlagsAlias),
				identifier,
			)
		}
	}
}

func (state *blockScopedVarState) repairClassExpressionDecoratorTarget(identifier *ast.Node, target *ast.Symbol, meaning ast.SymbolFlags) *ast.Symbol {
	for current := identifier.Parent; current != nil; current = current.Parent {
		if current.Kind != ast.KindDecorator || current.Parent == nil || current.Parent.Kind != ast.KindClassExpression {
			continue
		}
		classExpression := current.Parent
		if target == classExpression.Symbol() && classExpression.Parent != nil {
			return state.resolve(classExpression.Parent, identifier.Text(), meaning)
		}
		return target
	}
	return target
}

type jsxTimelineScopeKind uint8

const (
	jsxScopeOther jsxTimelineScopeKind = iota
	jsxScopeFunctionType
	jsxScopeMappedType
	jsxScopeConditionalType
)

type jsxTimelineVariable struct {
	defined         bool
	value           bool
	firstIdentifier *ast.Node
	group           *varGroup
}

type jsxTimelineScope struct {
	upper             *jsxTimelineScope
	variableScope     *jsxTimelineScope
	functionBodyStart int
	kind              jsxTimelineScopeKind
	variables         [2]jsxTimelineVariable
}

type jsxTimelineChannel struct {
	active     bool
	nameIndex  int
	done       bool
	resolved   bool
	pending    *jsxTimelineScope
	identifier *ast.Node
}

type jsxTimeline struct {
	state     *blockScopedVarState
	names     [2]string
	nameCount int
	current   *jsxTimelineScope
	channels  [2]jsxTimelineChannel
}

func (state *blockScopedVarState) addTSXFactoryReferences() {
	if state.ctx.SourceFile.ScriptKind != core.ScriptKindTSX || !state.hasJSX {
		return
	}

	factoryName := "React"
	fragmentName := ""
	if sourceProgram := state.ctx.Program(); sourceProgram != nil {
		options := sourceProgram.Options()
		if options.JsxFactory != "" {
			factoryName = jsxFactoryRoot(options.JsxFactory)
		}
		if options.JsxFragmentFactory != "" {
			fragmentName = jsxFactoryRoot(options.JsxFragmentFactory)
		}
	}

	timeline := jsxTimeline{state: state}
	timeline.activateChannel(0, factoryName)
	timeline.activateChannel(1, fragmentName)
	if !timeline.channels[0].active && !timeline.channels[1].active {
		return
	}

	timeline.pushScope(jsxScopeOther, true) // global
	if ast.IsExternalOrCommonJSModule(state.ctx.SourceFile) {
		timeline.pushScope(jsxScopeOther, true) // module
	}
	timeline.visitStatements(state.ctx.SourceFile.AsNode())
	if ast.IsExternalOrCommonJSModule(state.ctx.SourceFile) {
		timeline.popScope()
	}
	timeline.popScope()
}

func jsxFactoryRoot(factory string) string {
	root, _, _ := strings.Cut(factory, ".")
	return strings.TrimSpace(root)
}

func (timeline *jsxTimeline) activateChannel(channelIndex int, name string) {
	if name == "" || !timeline.state.hasReportableGroup(name) {
		return
	}
	nameIndex := -1
	for index := range timeline.nameCount {
		if timeline.names[index] == name {
			nameIndex = index
			break
		}
	}
	if nameIndex == -1 {
		nameIndex = timeline.nameCount
		timeline.names[nameIndex] = name
		timeline.nameCount++
	}
	timeline.channels[channelIndex] = jsxTimelineChannel{
		active:    true,
		nameIndex: nameIndex,
	}
}

func (state *blockScopedVarState) hasReportableGroup(name string) bool {
	for _, group := range state.groupOrder {
		if group.name == name && len(group.occurrences) != 0 {
			return true
		}
	}
	return false
}

func (timeline *jsxTimeline) pushScope(kind jsxTimelineScopeKind, variableScope bool) {
	scope := &jsxTimelineScope{upper: timeline.current, kind: kind}
	if variableScope || timeline.current == nil {
		scope.variableScope = scope
	} else {
		scope.variableScope = timeline.current.variableScope
	}
	timeline.current = scope
}

func (timeline *jsxTimeline) popScope() {
	scope := timeline.current
	if scope == nil {
		return
	}
	for index := range timeline.channels {
		channel := &timeline.channels[index]
		if !channel.active || channel.resolved || channel.pending != scope {
			continue
		}
		variable := &scope.variables[channel.nameIndex]
		if variable.defined && variable.value {
			if timelineReferenceCannotResolveToFunctionBodyVariable(
				timeline.state.ctx.SourceFile,
				channel.identifier,
				scope,
				variable,
			) {
				channel.pending = scope.upper
				if channel.pending == nil {
					channel.resolved = true
				}
			} else {
				if variable.group != nil {
					variable.group.references = append(variable.group.references, referenceEvent{
						identifier:   channel.identifier,
						multiplicity: 1,
						bindingRange: true,
					})
				}
				channel.resolved = true
				channel.pending = nil
			}
		} else {
			channel.pending = scope.upper
			if channel.pending == nil {
				channel.resolved = true
			}
		}
	}
	timeline.current = scope.upper
}

func timelineReferenceCannotResolveToFunctionBodyVariable(
	sourceFile *ast.SourceFile,
	identifier *ast.Node,
	scope *jsxTimelineScope,
	variable *jsxTimelineVariable,
) bool {
	if scope == nil || scope.functionBodyStart == 0 || identifier == nil ||
		utils.TrimNodeTextRange(sourceFile, identifier).Pos() >= scope.functionBodyStart {
		return false
	}
	return variable.firstIdentifier == nil ||
		utils.TrimNodeTextRange(sourceFile, variable.firstIdentifier).Pos() >= scope.functionBodyStart
}

func (timeline *jsxTimeline) define(identifier *ast.Node, value bool, group *varGroup, target *jsxTimelineScope) {
	if identifier == nil || target == nil {
		return
	}
	nameIndex := timeline.indexOfName(identifier.Text())
	if nameIndex == -1 {
		return
	}
	variable := &target.variables[nameIndex]
	if !variable.defined {
		variable.defined = true
		variable.firstIdentifier = identifier
	}
	variable.value = variable.value || value
	if group != nil {
		variable.group = group
	}
}

func (timeline *jsxTimeline) indexOfName(name string) int {
	for index := range timeline.nameCount {
		if timeline.names[index] == name {
			return index
		}
	}
	return -1
}

func (timeline *jsxTimeline) referenceChannel(channelIndex int) {
	channel := &timeline.channels[channelIndex]
	if !channel.active || channel.done {
		return
	}
	for scope := timeline.current; scope != nil; scope = scope.upper {
		variable := &scope.variables[channel.nameIndex]
		if !variable.defined {
			continue
		}
		channel.done = true
		channel.pending = scope
		channel.identifier = variable.firstIdentifier
		return
	}
}

func (timeline *jsxTimeline) resolvedAllActiveChannels() bool {
	hasActive := false
	for index := range timeline.channels {
		channel := &timeline.channels[index]
		if !channel.active {
			continue
		}
		hasActive = true
		if !channel.resolved {
			return false
		}
	}
	return hasActive
}

func (timeline *jsxTimeline) visit(node *ast.Node) {
	if node == nil || timeline.resolvedAllActiveChannels() {
		return
	}
	switch node.Kind {
	case ast.KindBlock:
		timeline.pushScope(jsxScopeOther, false)
		timeline.visitStatements(node)
		timeline.popScope()

	case ast.KindVariableDeclarationList:
		timeline.visitVariableDeclarationList(node)

	case ast.KindFunctionDeclaration:
		if node.Name() != nil {
			timeline.define(node.Name(), true, nil, timeline.current)
		}
		timeline.visitFunction(node, false, false)

	case ast.KindFunctionExpression:
		if node.Name() != nil {
			timeline.pushScope(jsxScopeOther, false)
			timeline.define(node.Name(), true, nil, timeline.current)
			timeline.visitFunction(node, false, false)
			timeline.popScope()
		} else {
			timeline.visitFunction(node, false, false)
		}

	case ast.KindArrowFunction:
		timeline.visitFunction(node, false, false)

	case ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
		timeline.visitMethod(node)

	case ast.KindClassDeclaration, ast.KindClassExpression:
		timeline.visitClass(node)

	case ast.KindPropertyDeclaration:
		timeline.visitClassProperty(node)

	case ast.KindClassStaticBlockDeclaration:
		timeline.pushScope(jsxScopeOther, true)
		if body := node.AsClassStaticBlockDeclaration().Body; body != nil {
			timeline.visitStatements(body)
		}
		timeline.popScope()

	case ast.KindCatchClause:
		timeline.visitCatchClause(node)

	case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
		timeline.visitFor(node)

	case ast.KindSwitchStatement:
		timeline.visitSwitch(node)

	case ast.KindWithStatement:
		withStatement := node.AsWithStatement()
		timeline.visit(withStatement.Expression)
		timeline.pushScope(jsxScopeOther, false)
		timeline.visit(withStatement.Statement)
		timeline.popScope()

	case ast.KindModuleDeclaration:
		if !ast.IsGlobalScopeAugmentation(node) && node.Name() != nil && node.Name().Kind == ast.KindIdentifier {
			timeline.define(node.Name(), true, nil, timeline.current)
		}
		timeline.pushScope(jsxScopeOther, true)
		timeline.visit(node.AsModuleDeclaration().Body)
		timeline.popScope()

	case ast.KindModuleBlock:
		timeline.visitStatements(node)

	case ast.KindInterfaceDeclaration:
		timeline.visitInterface(node)

	case ast.KindTypeAliasDeclaration:
		timeline.visitTypeAlias(node)

	case ast.KindTypeParameter:
		timeline.visitTypeParameter(node, timeline.current)

	case ast.KindFunctionType, ast.KindConstructorType, ast.KindCallSignature, ast.KindConstructSignature:
		timeline.visitFunctionType(node)

	case ast.KindMethodSignature:
		timeline.visitMethodSignature(node)

	case ast.KindIndexSignature:
		timeline.visitIndexSignature(node)

	case ast.KindConditionalType:
		timeline.visitConditionalType(node)

	case ast.KindMappedType:
		timeline.visitMappedType(node)

	case ast.KindInferType:
		timeline.visitInferType(node)

	case ast.KindEnumDeclaration:
		timeline.visitEnum(node)

	case ast.KindImportDeclaration, ast.KindImportEqualsDeclaration:
		timeline.visitImport(node)

	case ast.KindCallExpression:
		call := node.AsCallExpression()
		timeline.visit(call.Expression)
		for _, argument := range node.Arguments() {
			timeline.visit(argument)
		}
		for _, typeArgument := range node.TypeArguments() {
			timeline.visit(typeArgument)
		}

	case ast.KindNewExpression:
		call := node.AsNewExpression()
		timeline.visit(call.Expression)
		for _, argument := range node.Arguments() {
			timeline.visit(argument)
		}
		for _, typeArgument := range node.TypeArguments() {
			timeline.visit(typeArgument)
		}

	case ast.KindTaggedTemplateExpression:
		tagged := node.AsTaggedTemplateExpression()
		timeline.visit(tagged.Tag)
		timeline.visit(tagged.Template)
		for _, typeArgument := range node.TypeArguments() {
			timeline.visit(typeArgument)
		}

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken != nil && ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			if binary.OperatorToken.Kind == ast.KindEqualsToken {
				left := timeline.visitExpressionTarget(binary.Left)
				if isTimelinePattern(left) {
					timeline.visitPatternRightHandNodes(left)
				} else {
					timeline.visit(left)
				}
			} else {
				// A direct array/object pattern is not a valid compound target.
				// TSESTree still parses it, but Referencer skips its entire left
				// subtree. Wrapped expressions remain ordinary expressions.
				switch binary.Left.Kind {
				case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression:
				default:
					timeline.visit(binary.Left)
				}
			}
			timeline.visit(binary.Right)
		} else {
			timeline.visitChildren(node)
		}

	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		if prefix.Operator == ast.KindPlusPlusToken || prefix.Operator == ast.KindMinusMinusToken {
			argument := timeline.visitExpressionTarget(prefix.Operand)
			if isTimelinePattern(argument) {
				timeline.visitPatternRightHandNodes(argument)
			} else {
				timeline.visitChildren(node)
			}
		} else {
			timeline.visitChildren(node)
		}

	case ast.KindPostfixUnaryExpression:
		argument := timeline.visitExpressionTarget(node.AsPostfixUnaryExpression().Operand)
		if isTimelinePattern(argument) {
			timeline.visitPatternRightHandNodes(argument)
		} else {
			timeline.visitChildren(node)
		}

	case ast.KindImportType:
		for _, typeArgument := range node.TypeArguments() {
			timeline.visit(typeArgument)
		}

	case ast.KindAsExpression:
		expression := node.AsAsExpression()
		timeline.visit(expression.Expression)
		timeline.visit(expression.Type)

	case ast.KindSatisfiesExpression:
		expression := node.AsSatisfiesExpression()
		timeline.visit(expression.Expression)
		timeline.visit(expression.Type)

	case ast.KindTypeAssertionExpression:
		expression := node.AsTypeAssertion()
		timeline.visit(expression.Expression)
		timeline.visit(expression.Type)

	case ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement:
		timeline.referenceChannel(0)
		timeline.visitChildren(node)

	case ast.KindJsxFragment:
		timeline.referenceChannel(0)
		timeline.referenceChannel(1)
		timeline.visitChildren(node)

	default:
		timeline.visitChildren(node)
	}
}

func (timeline *jsxTimeline) visitExpressionTarget(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	for node.Kind == ast.KindParenthesizedExpression {
		node = node.Expression()
	}
	switch node.Kind {
	case ast.KindNonNullExpression:
		node = node.Expression()
	case ast.KindAsExpression:
		expression := node.AsAsExpression()
		timeline.visit(expression.Type)
		node = expression.Expression
	case ast.KindTypeAssertionExpression:
		expression := node.AsTypeAssertion()
		timeline.visit(expression.Type)
		node = expression.Expression
	}
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.Expression()
	}
	return node
}

func isTimelinePattern(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier,
		ast.KindObjectLiteralExpression,
		ast.KindArrayLiteralExpression,
		ast.KindSpreadElement,
		ast.KindSpreadAssignment:
		return true
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		return binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken
	}
	return false
}

func (timeline *jsxTimeline) visitPatternRightHandNodes(node *ast.Node) {
	var rightHandNodes []*ast.Node
	collectTimelinePatternRightHandNodes(node, &rightHandNodes)
	for _, rightHand := range rightHandNodes {
		timeline.visit(rightHand)
	}
}

func collectTimelinePatternRightHandNodes(node *ast.Node, result *[]*ast.Node) {
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindParenthesizedExpression,
		ast.KindNonNullExpression:
		collectTimelinePatternRightHandNodes(node.Expression(), result)

	case ast.KindDecorator:
		return

	case ast.KindAsExpression:
		expression := node.AsAsExpression()
		collectTimelinePatternRightHandNodes(expression.Expression, result)
		collectTimelinePatternRightHandNodes(expression.Type, result)

	case ast.KindSatisfiesExpression:
		expression := node.AsSatisfiesExpression()
		collectTimelinePatternRightHandNodes(expression.Expression, result)
		collectTimelinePatternRightHandNodes(expression.Type, result)

	case ast.KindTypeAssertionExpression:
		expression := node.AsTypeAssertion()
		collectTimelinePatternRightHandNodes(expression.Type, result)
		collectTimelinePatternRightHandNodes(expression.Expression, result)

	case ast.KindObjectLiteralExpression:
		for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
			collectTimelinePatternRightHandNodes(property, result)
		}

	case ast.KindArrayLiteralExpression:
		for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
			collectTimelinePatternRightHandNodes(element, result)
		}

	case ast.KindPropertyAssignment:
		property := node.AsPropertyAssignment()
		if property.Name() != nil && property.Name().Kind == ast.KindComputedPropertyName {
			*result = append(*result, property.Name().AsComputedPropertyName().Expression)
		}
		collectTimelinePatternRightHandNodes(property.Initializer, result)

	case ast.KindShorthandPropertyAssignment:
		property := node.AsShorthandPropertyAssignment()
		if property.ObjectAssignmentInitializer != nil {
			*result = append(*result, property.ObjectAssignmentInitializer)
		}

	case ast.KindSpreadAssignment:
		collectTimelinePatternRightHandNodes(node.AsSpreadAssignment().Expression, result)

	case ast.KindSpreadElement:
		collectTimelinePatternRightHandNodes(node.AsSpreadElement().Expression, result)

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken != nil && ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			collectTimelinePatternRightHandNodes(binary.Left, result)
			*result = append(*result, binary.Right)
		} else {
			node.ForEachChild(func(child *ast.Node) bool {
				collectTimelinePatternRightHandNodes(child, result)
				return false
			})
		}

	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		*result = append(*result, access.ArgumentExpression, access.Expression)

	case ast.KindPropertyAccessExpression:
		*result = append(*result, node.AsPropertyAccessExpression().Expression)

	case ast.KindCallExpression:
		*result = append(*result, node.Arguments()...)
		collectTimelinePatternRightHandNodes(node.AsCallExpression().Expression, result)

	default:
		node.ForEachChild(func(child *ast.Node) bool {
			collectTimelinePatternRightHandNodes(child, result)
			return false
		})
	}
}

func (timeline *jsxTimeline) visitChildren(node *ast.Node) {
	node.ForEachChild(func(child *ast.Node) bool {
		timeline.visit(child)
		return timeline.resolvedAllActiveChannels()
	})
}

func (timeline *jsxTimeline) visitStatements(node *ast.Node) {
	for _, statement := range node.Statements() {
		timeline.visit(statement)
		if timeline.resolvedAllActiveChannels() {
			break
		}
	}
}

func (timeline *jsxTimeline) visitVariableDeclarationList(node *ast.Node) {
	list := node.AsVariableDeclarationList()
	target := timeline.current
	host := utils.FindEnclosingScope(node)
	if utils.IsVarKeyword(node) {
		target = timeline.current.variableScope
	}
	for _, declaration := range list.Declarations.Nodes {
		variable := declaration.AsVariableDeclaration()
		timeline.visitBindingPattern(variable.Name(), target, func(identifier *ast.Node) *varGroup {
			if utils.IsVarKeyword(node) && timeline.indexOfName(identifier.Text()) != -1 {
				canonical := timeline.state.canonical[canonicalKey{host: host, name: identifier.Text()}]
				return timeline.state.groups[canonical]
			}
			return nil
		})
		timeline.visit(variable.Initializer)
		timeline.visit(variable.Type)
	}
}

func (timeline *jsxTimeline) visitBindingPattern(name *ast.Node, target *jsxTimelineScope, groupFor func(*ast.Node) *varGroup) {
	utils.CollectBindingNames(name, func(identifier *ast.Node, _ string) {
		var group *varGroup
		if groupFor != nil {
			group = groupFor(identifier)
		}
		timeline.define(identifier, true, group, target)
	})
	for _, rightHand := range bindingPatternRightHandNodes(name) {
		timeline.visit(rightHand)
	}
}

func bindingPatternRightHandNodes(name *ast.Node) []*ast.Node {
	var result []*ast.Node
	var collect func(*ast.Node)
	collect = func(node *ast.Node) {
		if node == nil {
			return
		}
		switch node.Kind {
		case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
			node.ForEachChild(func(child *ast.Node) bool {
				if child.Kind == ast.KindBindingElement {
					collect(child)
				}
				return false
			})
		case ast.KindBindingElement:
			element := node.AsBindingElement()
			if element.PropertyName != nil && element.PropertyName.Kind == ast.KindComputedPropertyName {
				result = append(result, element.PropertyName.AsComputedPropertyName().Expression)
			}
			collect(element.Name())
			if element.Initializer != nil {
				result = append(result, element.Initializer)
			}
		}
	}
	collect(name)
	return result
}

func (timeline *jsxTimeline) visitFunction(node *ast.Node, method, decoratorsVisited bool) {
	if method && !decoratorsVisited {
		for _, parameter := range node.Parameters() {
			for _, decorator := range parameter.Decorators() {
				timeline.visit(decorator)
			}
		}
	}
	timeline.pushScope(jsxScopeOther, true)
	if body := node.Body(); body != nil {
		timeline.current.functionBodyStart = utils.TrimNodeTextRange(timeline.state.ctx.SourceFile, body).Pos()
	}
	for _, parameter := range node.Parameters() {
		timeline.visitParameter(parameter, method)
	}
	timeline.visit(node.Type())
	for _, typeParameter := range node.TypeParameters() {
		timeline.visitTypeParameter(typeParameter, timeline.current)
	}
	body := node.Body()
	if body != nil && body.Kind == ast.KindBlock {
		timeline.visitStatements(body)
	} else {
		timeline.visit(body)
	}
	timeline.popScope()
}

func (timeline *jsxTimeline) visitParameter(node *ast.Node, method bool) {
	parameter := node.AsParameterDeclaration()
	timeline.visitBindingPattern(parameter.Name(), timeline.current, nil)
	timeline.visit(parameter.Initializer)
	timeline.visit(parameter.Type)
	if !method {
		for _, decorator := range node.Decorators() {
			timeline.visit(decorator)
		}
	}
}

func (timeline *jsxTimeline) visitMethod(node *ast.Node) {
	if name := node.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
		timeline.visit(name.AsComputedPropertyName().Expression)
	}
	for _, parameter := range node.Parameters() {
		for _, decorator := range parameter.Decorators() {
			timeline.visit(decorator)
		}
	}
	timeline.visitFunction(node, true, true)
	for _, decorator := range node.Decorators() {
		timeline.visit(decorator)
	}
}

func (timeline *jsxTimeline) visitClass(node *ast.Node) {
	if node.Kind == ast.KindClassDeclaration && node.Name() != nil {
		timeline.define(node.Name(), true, nil, timeline.current)
	}
	for _, decorator := range node.Decorators() {
		timeline.visit(decorator)
	}
	timeline.pushScope(jsxScopeOther, false)
	if node.Name() != nil {
		timeline.define(node.Name(), true, nil, timeline.current)
	}
	classData := node.ClassLikeData()
	if classData.HeritageClauses != nil {
		for _, clause := range classData.HeritageClauses.Nodes {
			heritage := clause.AsHeritageClause()
			if heritage.Token == ast.KindExtendsKeyword {
				for _, item := range heritage.Types.Nodes {
					timeline.visit(item.AsExpressionWithTypeArguments().Expression)
				}
			}
		}
	}
	for _, typeParameter := range node.TypeParameters() {
		timeline.visitTypeParameter(typeParameter, timeline.current)
	}
	if classData.HeritageClauses != nil {
		for _, clause := range classData.HeritageClauses.Nodes {
			for _, item := range clause.AsHeritageClause().Types.Nodes {
				typeArguments := item.AsExpressionWithTypeArguments().TypeArguments
				if typeArguments != nil {
					for _, typeArgument := range typeArguments.Nodes {
						timeline.visit(typeArgument)
					}
				}
			}
		}
	}
	for _, member := range node.Members() {
		timeline.visit(member)
	}
	timeline.popScope()
}

func (timeline *jsxTimeline) visitClassProperty(node *ast.Node) {
	property := node.AsPropertyDeclaration()
	if name := property.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
		timeline.visit(name.AsComputedPropertyName().Expression)
	}
	if property.Initializer != nil {
		timeline.pushScope(jsxScopeOther, true)
		timeline.visit(property.Initializer)
		timeline.popScope()
	}
	for _, decorator := range node.Decorators() {
		timeline.visit(decorator)
	}
	timeline.visit(property.Type)
}

func (timeline *jsxTimeline) visitCatchClause(node *ast.Node) {
	clause := node.AsCatchClause()
	timeline.pushScope(jsxScopeOther, false)
	if clause.VariableDeclaration != nil {
		timeline.visitBindingPattern(clause.VariableDeclaration.Name(), timeline.current, nil)
	}
	timeline.visit(clause.Block)
	timeline.popScope()
}

func (timeline *jsxTimeline) visitFor(node *ast.Node) {
	switch node.Kind {
	case ast.KindForStatement:
		statement := node.AsForStatement()
		needsScope := statement.Initializer != nil &&
			statement.Initializer.Kind == ast.KindVariableDeclarationList &&
			!utils.IsVarKeyword(statement.Initializer)
		if needsScope {
			timeline.pushScope(jsxScopeOther, false)
		}
		timeline.visit(statement.Initializer)
		timeline.visit(statement.Condition)
		timeline.visit(statement.Incrementor)
		timeline.visit(statement.Statement)
		if needsScope {
			timeline.popScope()
		}

	case ast.KindForInStatement, ast.KindForOfStatement:
		statement := node.AsForInOrOfStatement()
		needsScope := statement.Initializer.Kind == ast.KindVariableDeclarationList &&
			!utils.IsVarKeyword(statement.Initializer)
		if needsScope {
			timeline.pushScope(jsxScopeOther, false)
		}
		if statement.Initializer.Kind == ast.KindVariableDeclarationList {
			timeline.visit(statement.Initializer)
		} else {
			timeline.visitPatternRightHandNodes(statement.Initializer)
		}
		timeline.visit(statement.Expression)
		timeline.visit(statement.Statement)
		if needsScope {
			timeline.popScope()
		}
	}
}

func (timeline *jsxTimeline) visitSwitch(node *ast.Node) {
	statement := node.AsSwitchStatement()
	timeline.visit(statement.Expression)
	timeline.pushScope(jsxScopeOther, false)
	timeline.visit(statement.CaseBlock)
	timeline.popScope()
}

func (timeline *jsxTimeline) visitInterface(node *ast.Node) {
	declaration := node.AsInterfaceDeclaration()
	timeline.define(declaration.Name(), false, nil, timeline.current)
	if declaration.TypeParameters != nil {
		timeline.pushScope(jsxScopeOther, false)
	}
	for _, typeParameter := range node.TypeParameters() {
		timeline.visitTypeParameter(typeParameter, timeline.current)
	}
	if declaration.HeritageClauses != nil {
		for _, heritage := range declaration.HeritageClauses.Nodes {
			timeline.visit(heritage)
		}
	}
	for _, member := range declaration.Members.Nodes {
		timeline.visit(member)
	}
	if declaration.TypeParameters != nil {
		timeline.popScope()
	}
}

func (timeline *jsxTimeline) visitTypeAlias(node *ast.Node) {
	declaration := node.AsTypeAliasDeclaration()
	timeline.define(declaration.Name(), false, nil, timeline.current)
	if declaration.TypeParameters != nil {
		timeline.pushScope(jsxScopeOther, false)
	}
	for _, typeParameter := range node.TypeParameters() {
		timeline.visitTypeParameter(typeParameter, timeline.current)
	}
	timeline.visit(declaration.Type)
	if declaration.TypeParameters != nil {
		timeline.popScope()
	}
}

func (timeline *jsxTimeline) visitTypeParameter(node *ast.Node, target *jsxTimelineScope) {
	declaration := node.AsTypeParameterDeclaration()
	timeline.define(declaration.Name(), false, nil, target)
	timeline.visit(declaration.Constraint)
	timeline.visit(declaration.DefaultType)
}

func (timeline *jsxTimeline) visitFunctionType(node *ast.Node) {
	timeline.pushScope(jsxScopeFunctionType, false)
	for _, typeParameter := range node.TypeParameters() {
		timeline.visitTypeParameter(typeParameter, timeline.current)
	}
	for _, parameter := range node.Parameters() {
		utils.CollectBindingNames(parameter.Name(), func(identifier *ast.Node, _ string) {
			timeline.define(identifier, true, nil, timeline.current)
		})
		timeline.visit(parameter.Type())
	}
	timeline.visit(node.Type())
	timeline.popScope()
}

func (timeline *jsxTimeline) visitMethodSignature(node *ast.Node) {
	if name := node.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
		timeline.visit(name.AsComputedPropertyName().Expression)
	}
	timeline.visitFunctionType(node)
}

func (timeline *jsxTimeline) visitIndexSignature(node *ast.Node) {
	for _, parameter := range node.Parameters() {
		if parameter.Name() != nil && parameter.Name().Kind == ast.KindIdentifier {
			timeline.visit(parameter.Type())
		}
	}
	timeline.visit(node.Type())
}

func (timeline *jsxTimeline) visitConditionalType(node *ast.Node) {
	conditional := node.AsConditionalTypeNode()
	timeline.pushScope(jsxScopeConditionalType, false)
	timeline.visit(conditional.CheckType)
	timeline.visit(conditional.ExtendsType)
	timeline.visit(conditional.TrueType)
	timeline.popScope()
	timeline.visit(conditional.FalseType)
}

func (timeline *jsxTimeline) visitMappedType(node *ast.Node) {
	mapped := node.AsMappedTypeNode()
	timeline.pushScope(jsxScopeMappedType, false)
	typeParameter := mapped.TypeParameter.AsTypeParameterDeclaration()
	timeline.define(typeParameter.Name(), false, nil, timeline.current)
	timeline.visit(typeParameter.Constraint)
	timeline.visit(mapped.NameType)
	timeline.visit(mapped.Type)
	timeline.popScope()
}

func (timeline *jsxTimeline) visitInferType(node *ast.Node) {
	target := timeline.current
	for scope := timeline.current; scope != nil; scope = scope.upper {
		if scope.kind == jsxScopeConditionalType {
			target = scope
			break
		}
		if scope.kind != jsxScopeFunctionType && scope.kind != jsxScopeMappedType {
			break
		}
	}
	typeParameter := node.AsInferTypeNode().TypeParameter.AsTypeParameterDeclaration()
	timeline.define(typeParameter.Name(), false, nil, target)
	timeline.visit(typeParameter.Constraint)
}

func (timeline *jsxTimeline) visitEnum(node *ast.Node) {
	declaration := node.AsEnumDeclaration()
	timeline.define(declaration.Name(), true, nil, timeline.current)
	timeline.pushScope(jsxScopeOther, false)
	for _, member := range declaration.Members.Nodes {
		if member.Name() != nil && member.Name().Kind == ast.KindIdentifier {
			timeline.define(member.Name(), true, nil, timeline.current)
		}
		timeline.visit(member.AsEnumMember().Initializer)
	}
	timeline.popScope()
}

func (timeline *jsxTimeline) visitImport(node *ast.Node) {
	if node.Kind == ast.KindImportEqualsDeclaration {
		if node.Name() != nil {
			timeline.define(node.Name(), true, nil, timeline.current)
		}
		return
	}
	utils.VisitDescendants(node, func(candidate *ast.Node) bool {
		switch candidate.Kind {
		case ast.KindImportClause, ast.KindNamespaceImport, ast.KindImportSpecifier:
			if candidate.Name() != nil {
				timeline.define(candidate.Name(), true, nil, timeline.current)
			}
		}
		return true
	})
}
