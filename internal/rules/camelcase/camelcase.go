package camelcase

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

//go:embed camelcase.schema.json
var schemaJSON []byte

// CamelcaseRule enforces ESLint's deliberately narrow definition of camel
// case: internal underscores are rejected, while leading/trailing underscores
// and all-uppercase names are accepted.
var CamelcaseRule = rule.Rule{
	Name:   "camelcase",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		state := camelcaseState{
			ctx:      ctx,
			opts:     opts,
			reported: make(map[int]struct{}),
		}

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				state.identifiers = append(state.identifiers, node)
			},
			ast.KindPrivateIdentifier: func(node *ast.Node) {
				state.privateIdentifiers = append(state.privateIdentifiers, node)
			},
			rule.ListenerOnExit(ast.KindEndOfFile): func(_ *ast.Node) {
				state.checkFile()
			},
		}
	},
}

type camelcaseOptions struct {
	ignoreDestructuring bool
	ignoreGlobals       bool
	ignoreImports       bool
	properties          bool
	allowedNames        map[string]struct{}
	allowedPatterns     []*esregexp.RegExp
}

func parseOptions(options []any) camelcaseOptions {
	opts := camelcaseOptions{properties: true}
	if len(options) == 0 {
		return opts
	}
	m, _ := options[0].(map[string]any)
	if m == nil {
		return opts
	}
	if value, ok := m["ignoreDestructuring"].(bool); ok {
		opts.ignoreDestructuring = value
	}
	if value, ok := m["ignoreGlobals"].(bool); ok {
		opts.ignoreGlobals = value
	}
	if value, ok := m["ignoreImports"].(bool); ok {
		opts.ignoreImports = value
	}
	if value, ok := m["properties"].(string); ok {
		opts.properties = value != "never"
	}
	if allow, ok := m["allow"].([]any); ok {
		opts.allowedNames = make(map[string]struct{}, len(allow))
		for _, raw := range allow {
			pattern, ok := raw.(string)
			if !ok {
				continue
			}
			opts.allowedNames[pattern] = struct{}{}
			if re, err := esregexp.Compile(pattern, "u"); err == nil {
				opts.allowedPatterns = append(opts.allowedPatterns, re)
			}
		}
	}
	return opts
}

func (o camelcaseOptions) isGoodName(name string) bool {
	body := strings.Trim(name, "_")
	if !strings.Contains(body, "_") || body == ecmascript.StringToUpperCase(body) {
		return true
	}
	if _, ok := o.allowedNames[name]; ok {
		return true
	}
	for _, pattern := range o.allowedPatterns {
		if pattern.Test(name) {
			return true
		}
	}
	return false
}

type bindingKind uint8

const (
	bindingNone bindingKind = iota
	bindingLocal
	bindingImport
)

type camelcaseState struct {
	ctx                rule.RuleContext
	opts               camelcaseOptions
	identifiers        []*ast.Node
	privateIdentifiers []*ast.Node
	reported           map[int]struct{}
}

type bindingCandidate struct {
	identifier *ast.Node
	kind       bindingKind
}

type bindingGroupKey struct {
	scope *scope.Scope
	name  string
}

type bindingGroup struct {
	key          bindingGroupKey
	declarations []*scope.Variable
	hasLocal     bool
	hasImport    bool
}

func (s *camelcaseState) checkFile() {
	badIdentifiers := make([]*ast.Node, 0)
	candidates := make([]bindingCandidate, 0)
	needsScopeAnalysis := false
	for _, identifier := range s.identifiers {
		if s.opts.isGoodName(identifier.AsIdentifier().Text) {
			continue
		}
		badIdentifiers = append(badIdentifiers, identifier)
		kind := runtimeBindingKind(identifier)
		if kind != bindingNone {
			needsScopeAnalysis = true
			candidates = append(candidates, bindingCandidate{identifier: identifier, kind: kind})
		} else if s.needsScopeAnalysis(identifier) {
			needsScopeAnalysis = true
		}
	}

	badReferences := make(map[*ast.Node]struct{})
	bindingReports := make(map[*ast.Node]string)
	initializationReports := make(map[*ast.Node]string)
	var resolvedReferences map[*ast.Node]*scope.Reference
	var manager *scope.Manager
	if needsScopeAnalysis {
		// TypeScript's binder and typescript-eslint's scope manager deliberately
		// disagree for some reference spaces and invalid declaration merges. Use
		// the shared ESLint-shaped graph as the authority for both local binding
		// references and Program's local-vs-global decision.
		analysisOptions := scope.Options{CollectReferences: true}
		// Filtering pays for sparse violations. On dense files, collecting every
		// reference avoids building an equally large name set and its map lookups.
		if len(badIdentifiers)*2 < len(s.identifiers) {
			referenceNames := make(map[string]struct{}, len(badIdentifiers))
			for _, identifier := range badIdentifiers {
				referenceNames[identifier.AsIdentifier().Text] = struct{}{}
			}
			analysisOptions.ReferenceNames = referenceNames
		}
		manager = scope.Build(s.ctx.SourceFile, analysisOptions)
		resolvedReferences = make(map[*ast.Node]*scope.Reference, len(manager.References))
		for _, reference := range manager.References {
			resolvedReferences[reference.Identifier] = reference
		}
	}
	if len(candidates) != 0 {
		s.checkBindings(manager, candidates, badReferences, bindingReports, initializationReports)
	}

	for _, identifier := range badIdentifiers {
		name := identifier.AsIdentifier().Text
		if utils.IsImportAttributeKey(identifier) {
			continue
		}
		if reportName, ok := bindingReports[identifier]; ok {
			s.reportBinding(identifier, reportName)
		}
		if reportName, ok := initializationReports[identifier]; ok {
			s.reportInitializationReference(identifier, reportName)
		}

		if runtimeBindingKind(identifier) != bindingNone {
			continue
		}

		if isPropertyName(identifier) {
			if s.opts.properties {
				s.report(identifier, name, false)
				continue
			}
			// A shorthand object member is both its property key and a
			// reference to the local variable. With properties:"never", only
			// the key half is exempt; keep evaluating the reference half.
			if utils.IsNonReferenceIdentifier(identifier) {
				continue
			}
		}
		if isAssignedPropertyAccessName(identifier) {
			if s.opts.properties {
				s.report(identifier, name, false)
			}
			continue
		}
		if isExportedName(identifier) || isLabel(identifier) {
			s.report(identifier, name, false)
			continue
		}

		if _, ok := badReferences[identifier]; ok {
			s.reportReference(identifier, name)
			continue
		}
		if !isRuntimeReference(identifier) {
			continue
		}
		globalDeclared := s.ctx.Globals.Access(name).IsDeclared()
		if reference := resolvedReferences[identifier]; reference != nil && reference.Resolved() != nil {
			// Program only receives unresolved references. This scope-manager
			// lookup also covers declaration merges which TypeScript rejects but
			// ESTree still models as one local variable.
			continue
		}
		if s.ctx.Refs != nil {
			referenceSpace := scope.ESLintReferenceSpace(identifier)
			meaning := referenceSpace.DeclarationMeaning()
			if referenceSpace.IncludesValue() && s.ctx.Refs.HasImplicitWrapperBinding(name) {
				// A lexical wrapper binding is local but has no authored declaration
				// for the scope graph to record.
				continue
			}
			if globalDeclared && !s.ctx.Refs.IsGlobalNameReference(identifier, name, meaning) {
				// In a global script, scope-manager merges configured globals with
				// same-named authored Program definitions. Any identifier on that
				// merged variable makes Program skip all of its references.
				continue
			}
		}
		if s.opts.ignoreGlobals && globalDeclared {
			continue
		}
		s.reportReference(identifier, name)
	}

	if s.opts.properties {
		for _, identifier := range s.privateIdentifiers {
			if !isPrivatePropertyName(identifier) {
				continue
			}
			name := strings.TrimPrefix(identifier.AsPrivateIdentifier().Text, "#")
			if !s.opts.isGoodName(name) {
				s.report(identifier, name, true)
			}
		}
	}
}

// needsScopeAnalysis mirrors the direct-report branches in checkFile and
// returns true only when the identifier can reach Program's resolution logic.
// This keeps property-only, label-only, and intrinsic-JSX files on the cheap
// syntax path without weakening the authoritative scope result where it is
// semantically required.
func (s *camelcaseState) needsScopeAnalysis(identifier *ast.Node) bool {
	if utils.IsImportAttributeKey(identifier) {
		return false
	}
	if isPropertyName(identifier) {
		return !s.opts.properties && !utils.IsNonReferenceIdentifier(identifier)
	}
	if isAssignedPropertyAccessName(identifier) || isExportedName(identifier) || isLabel(identifier) {
		return false
	}
	return isRuntimeReference(identifier)
}

func (s *camelcaseState) checkBindings(
	manager *scope.Manager,
	candidates []bindingCandidate,
	badReferences map[*ast.Node]struct{},
	bindingReports map[*ast.Node]string,
	initializationReports map[*ast.Node]string,
) {
	if manager == nil {
		return
	}
	candidateKinds := make(map[*ast.Node]bindingKind, len(candidates))
	for _, candidate := range candidates {
		candidateKinds[candidate.identifier] = candidate.kind
	}

	groupsByKey := make(map[bindingGroupKey]*bindingGroup)
	groups := make([]*bindingGroup, 0, len(candidates))
	matched := make(map[*ast.Node]struct{}, len(candidates))

	for _, currentScope := range manager.Scopes {
		for _, declaration := range currentScope.Vars {
			kind, ok := candidateKinds[declaration.ID]
			if !ok {
				continue
			}
			matched[declaration.ID] = struct{}{}
			key := bindingGroupKey{scope: currentScope, name: declaration.Name}
			group := groupsByKey[key]
			if group == nil {
				group = &bindingGroup{
					key:          key,
					declarations: currentScope.Declarations(declaration.Name),
				}
				groupsByKey[key] = group
				groups = append(groups, group)
			}
			switch kind {
			case bindingLocal:
				group.hasLocal = true
			case bindingImport:
				group.hasImport = true
			}
		}
	}

	badGroups := make(map[bindingGroupKey]struct{}, len(groups))
	for _, group := range groups {
		badGroups[group.key] = struct{}{}
		name := group.key.name
		anchor := scopeManagerBindingIdentifier(group.key.scope, group.declarations)
		if anchor != nil {
			if group.hasLocal && (!s.opts.ignoreDestructuring || !equalsOriginalName(anchor)) {
				bindingReports[anchor] = name
			}
			if group.hasImport && (!s.opts.ignoreImports || !equalsOriginalName(anchor)) {
				bindingReports[anchor] = name
			}
		}
		for _, declaration := range group.declarations {
			if declaration == nil || declaration.ID == nil {
				continue
			}
			if group.hasImport && declaration.Kind == scope.DefVariable &&
				hasScopeManagerInitReference(declaration.ID) {
				initializationReports[declaration.ID] = name
			}
		}
	}
	for _, reference := range manager.References {
		resolved := reference.Resolved()
		if resolved == nil {
			continue
		}
		if _, bad := badGroups[bindingGroupKey{scope: resolved.Scope, name: resolved.Name}]; bad {
			badReferences[reference.Identifier] = struct{}{}
		}
	}

	// Every runtime binding kind above is modeled by the shared scope builder.
	// Keep a syntax fallback so a newly parsed binding shape cannot silently
	// disappear before the scope model gains its corresponding definition.
	for _, candidate := range candidates {
		if _, ok := matched[candidate.identifier]; ok {
			continue
		}
		name := candidate.identifier.AsIdentifier().Text
		if (candidate.kind == bindingLocal && (!s.opts.ignoreDestructuring || !equalsOriginalName(candidate.identifier))) ||
			(candidate.kind == bindingImport && (!s.opts.ignoreImports || !equalsOriginalName(candidate.identifier))) {
			bindingReports[candidate.identifier] = name
		}
	}
}

// scopeManagerBindingIdentifier returns Variable#identifiers[0]. The shared
// scope builder records declarations in source order except for runtime
// function scopes: typescript-eslint defines parameters before type parameters,
// while tsgo exposes the type parameters first. That is the only ordering
// exception which can affect a camelcase declaration anchor.
func scopeManagerBindingIdentifier(currentScope *scope.Scope, declarations []*scope.Variable) *ast.Node {
	var first *ast.Node
	var firstParameter *ast.Node
	for _, declaration := range declarations {
		if declaration == nil || declaration.Anonymous || declaration.ID == nil ||
			declaration.ID.Kind != ast.KindIdentifier {
			continue
		}
		identifier := declaration.ID
		if first == nil || identifier.Pos() < first.Pos() {
			first = identifier
		}
		if currentScope.Kind == scope.KindFunction && declaration.Kind == scope.DefParameter &&
			(firstParameter == nil || identifier.Pos() < firstParameter.Pos()) {
			firstParameter = identifier
		}
	}
	if firstParameter != nil {
		return firstParameter
	}
	return first
}

func hasScopeManagerInitReference(identifier *ast.Node) bool {
	for current := identifier.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindBindingElement:
			if current.AsBindingElement().Initializer != nil {
				return true
			}
		case ast.KindVariableDeclaration:
			if current.AsVariableDeclaration().Initializer != nil {
				return true
			}
			declarationList := current.Parent
			return declarationList != nil && declarationList.Kind == ast.KindVariableDeclarationList &&
				declarationList.Parent != nil &&
				(declarationList.Parent.Kind == ast.KindForInStatement || declarationList.Parent.Kind == ast.KindForOfStatement)
		}
	}
	return false
}

func (s *camelcaseState) reportBinding(node *ast.Node, name string) {
	if _, exists := s.reported[node.Pos()]; exists {
		return
	}
	s.reported[node.Pos()] = struct{}{}
	s.ctx.ReportRange(utils.GetESTreeBindingIdentifierRange(s.ctx.SourceFile, node), camelcaseMessage(name, false))
}

func (s *camelcaseState) reportReference(node *ast.Node, name string) {
	if !s.shouldReportReference(node) {
		return
	}
	s.report(node, name, false)
}

func (s *camelcaseState) reportInitializationReference(node *ast.Node, name string) {
	if !s.shouldReportReference(node) {
		return
	}
	s.reportBinding(node, name)
}

func (s *camelcaseState) shouldReportReference(node *ast.Node) bool {
	outer := utils.OutermostParenthesizedExpression(node)
	parent := outer.Parent
	if parent != nil {
		if parent.Kind == ast.KindCallExpression || parent.Kind == ast.KindNewExpression {
			return false
		}
		if isDefaultValue(parent, outer) {
			return false
		}
	}
	if s.opts.ignoreDestructuring && equalsOriginalName(node) {
		return false
	}
	return !utils.IsImportAttributeKey(node)
}

func (s *camelcaseState) report(node *ast.Node, name string, private bool) {
	if _, exists := s.reported[node.Pos()]; exists {
		return
	}
	s.reported[node.Pos()] = struct{}{}
	s.ctx.ReportNode(node, camelcaseMessage(name, private))
}

func camelcaseMessage(name string, private bool) rule.RuleMessage {
	if private {
		return rule.RuleMessage{
			Id:          "notCamelCasePrivate",
			Description: fmt.Sprintf("#%s is not in camel case.", name),
			Data:        map[string]string{"name": name},
		}
	}
	return rule.RuleMessage{
		Id:          "notCamelCase",
		Description: fmt.Sprintf("Identifier '%s' is not in camel case.", name),
		Data:        map[string]string{"name": name},
	}
}

func runtimeBindingKind(node *ast.Node) bindingKind {
	if node == nil || node.Parent == nil || node.Parent.Name() != node {
		return bindingNone
	}
	parent := node.Parent
	switch parent.Kind {
	case ast.KindImportClause, ast.KindImportSpecifier, ast.KindNamespaceImport:
		return bindingImport
	case ast.KindVariableDeclaration:
		return bindingLocal
	case ast.KindParameter:
		owner := parent.Parent
		if owner != nil && ast.IsFunctionLikeDeclaration(owner) && owner.Body() != nil {
			return bindingLocal
		}
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
		if parent.Body() != nil {
			return bindingLocal
		}
	case ast.KindClassDeclaration, ast.KindClassExpression:
		return bindingLocal
	case ast.KindBindingElement:
		owner := ast.WalkUpBindingElementsAndPatterns(parent)
		if owner == nil {
			return bindingNone
		}
		switch owner.Kind {
		case ast.KindVariableDeclaration:
			return bindingLocal
		case ast.KindParameter:
			function := owner.Parent
			if function != nil && ast.IsFunctionLikeDeclaration(function) && function.Body() != nil {
				return bindingLocal
			}
		}
	}
	return bindingNone
}

func equalsOriginalName(node *ast.Node) bool {
	if node == nil || node.Parent == nil || node.Parent.Kind != ast.KindImportSpecifier {
		return equalsDestructuredName(node)
	}
	specifier := node.Parent.AsImportSpecifier()
	if specifier.Name() != node {
		return false
	}
	imported := specifier.PropertyName
	if imported == nil {
		imported = specifier.Name()
	}
	name, ok := utils.GetStaticPropertyName(imported)
	return ok && name == node.AsIdentifier().Text
}

func equalsDestructuredName(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	name := node.AsIdentifier().Text
	if node.Parent.Kind == ast.KindBindingElement {
		binding := node.Parent.AsBindingElement()
		if binding.Name() != node || node.Parent.Parent == nil || node.Parent.Parent.Kind != ast.KindObjectBindingPattern {
			return false
		}
		if binding.DotDotDotToken != nil {
			return false
		}
		propertyName := binding.PropertyName
		if propertyName == nil {
			return true
		}
		property, ok := plainIdentifierName(propertyName)
		return ok && property == name
	}

	outer := utils.OutermostParenthesizedExpression(node)
	parent := outer.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindShorthandPropertyAssignment:
		return parent.Name() == node
	case ast.KindPropertyAssignment:
		assignment := parent.AsPropertyAssignment()
		if assignment.Initializer != outer {
			return false
		}
		property, ok := plainIdentifierName(parent.Name())
		return ok && property == name
	}
	return false
}

func plainIdentifierName(node *ast.Node) (string, bool) {
	if node == nil || node.Kind != ast.KindIdentifier {
		return "", false
	}
	return node.AsIdentifier().Text, true
}

func isDefaultValue(parent, child *ast.Node) bool {
	switch parent.Kind {
	case ast.KindParameter:
		return parent.AsParameterDeclaration().Initializer == child
	case ast.KindBindingElement:
		return parent.AsBindingElement().Initializer == child
	case ast.KindShorthandPropertyAssignment:
		return parent.AsShorthandPropertyAssignment().ObjectAssignmentInitializer == child
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		return binary.OperatorToken.Kind == ast.KindEqualsToken && binary.Right == child &&
			utils.IsInDestructuringAssignment(parent)
	}
	return false
}

func isPropertyName(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || parent.Name() != node {
		return false
	}
	switch parent.Kind {
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		if parent.Parent == nil || parent.Parent.Kind != ast.KindObjectLiteralExpression {
			return parent.Parent != nil &&
				(parent.Parent.Kind == ast.KindClassDeclaration || parent.Parent.Kind == ast.KindClassExpression) &&
				utils.IsPlainClassMember(parent)
		}
		return !utils.IsInDestructuringAssignment(parent)
	case ast.KindPropertyDeclaration:
		return parent.Parent != nil &&
			(parent.Parent.Kind == ast.KindClassDeclaration || parent.Parent.Kind == ast.KindClassExpression) &&
			utils.IsPlainClassMember(parent)
	}
	return false
}

func isPrivatePropertyName(node *ast.Node) bool {
	if node == nil || node.Parent == nil || node.Parent.Name() != node {
		return false
	}
	switch node.Parent.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindPropertyDeclaration:
		owner := node.Parent.Parent
		return owner != nil && (owner.Kind == ast.KindClassDeclaration || owner.Kind == ast.KindClassExpression) &&
			utils.IsPlainClassMember(node.Parent)
	}
	return false
}

func isAssignedPropertyAccessName(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindPropertyAccessExpression ||
		parent.AsPropertyAccessExpression().Name() != node {
		return false
	}

	// Espree erases parentheses, but TypeScript assertion and non-null wrappers
	// remain real ESTree parents. Only parentheses are transparent here.
	target := utils.OutermostParenthesizedExpression(parent)
	owner := target.Parent
	if owner == nil {
		return false
	}
	switch owner.Kind {
	case ast.KindBinaryExpression:
		binary := owner.AsBinaryExpression()
		return binary.Left == target && ast.IsAssignmentOperator(binary.OperatorToken.Kind)
	case ast.KindPropertyAssignment:
		property := owner.AsPropertyAssignment()
		return property.Initializer == target && utils.IsInDestructuringAssignment(owner)
	case ast.KindArrayLiteralExpression:
		return utils.IsInDestructuringAssignment(owner)
	case ast.KindSpreadElement, ast.KindSpreadAssignment:
		return utils.IsInDestructuringAssignment(owner)
	}
	return false
}

func isExportedName(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	switch node.Parent.Kind {
	case ast.KindExportSpecifier, ast.KindNamespaceExport, ast.KindNamespaceExportDeclaration:
		return node.Parent.Name() == node
	}
	return false
}

func isLabel(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	switch node.Parent.Kind {
	case ast.KindLabeledStatement, ast.KindBreakStatement, ast.KindContinueStatement:
		return true
	}
	return false
}

func isRuntimeReference(node *ast.Node) bool {
	if isJsxNamePart(node) {
		// Keep direct, member, namespaced, attribute, and parser-specific
		// opening/closing behavior in the shared scope-manager predicate.
		return scope.IsReferenceIdentifier(node)
	}
	if utils.IsImportTypeSyntax(node) {
		return false
	}
	if isAliasedTypeExportSource(node) {
		return true
	}
	if ast.IsPartOfTypeQuery(node) {
		return isTypeQueryValueReference(node)
	}
	if ast.IsPartOfTypeNode(node) {
		return scope.IsReferenceIdentifier(node)
	}
	return node != nil && !ast.IsPartOfTypeNode(node) && !utils.IsNonReferenceIdentifier(node)
}

// isAliasedTypeExportSource recognizes the local side of
// `export type { local_name as exportedName }`. scope-manager exposes that
// identifier through the Program scope, even though the generic
// non-reference helper excludes type-only export specifiers.
func isAliasedTypeExportSource(node *ast.Node) bool {
	if node == nil || node.Parent == nil || node.Parent.Kind != ast.KindExportSpecifier {
		return false
	}
	specifier := node.Parent.AsExportSpecifier()
	if specifier == nil || specifier.PropertyName != node || specifier.Name() == node {
		return false
	}
	return ast.IsTypeOnlyImportOrExportDeclaration(node.Parent) && !utils.IsReExportSpecifier(node.Parent)
}

func isJsxNamePart(node *ast.Node) bool {
	current := node
	for current != nil && current.Parent != nil {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindPropertyAccessExpression:
			access := parent.AsPropertyAccessExpression()
			if access == nil || (access.Expression != current && access.Name() != current) {
				return false
			}
			current = parent
		case ast.KindJsxNamespacedName:
			current = parent
		case ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement, ast.KindJsxClosingElement:
			return ast.IsJsxTagName(current)
		case ast.KindJsxAttribute:
			return parent.Name() == current
		default:
			return false
		}
	}
	return false
}

// isTypeQueryValueReference recognizes the value-side root of a TypeScript
// type query. In `typeof namespace.member`, only `namespace` is a value
// reference; the remaining qualified-name segments are property names.
func isTypeQueryValueReference(node *ast.Node) bool {
	current := node
	for current != nil && current.Parent != nil && current.Parent.Kind == ast.KindQualifiedName {
		qualified := current.Parent.AsQualifiedName()
		if qualified.Left != current {
			return false
		}
		current = current.Parent
	}
	return current != nil && current.Parent != nil && current.Parent.Kind == ast.KindTypeQuery
}
