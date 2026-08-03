package no_import_assign

import (
	"cmp"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// importedBinding holds information about a single imported binding.
type importedBinding struct {
	name        string
	isNamespace bool
	nameNode    *ast.Node
	symbol      *ast.Symbol
}

type importedBindingMessageCache struct {
	direct string
	member string
}

type importedBindingViolationKind uint8

const (
	importedBindingViolationNone importedBindingViolationKind = iota
	importedBindingViolationDirect
	importedBindingViolationMember
)

type importedBindingViolation struct {
	node         *ast.Node
	bindingIndex int
	kind         importedBindingViolationKind
}

type importedBindingGroup struct {
	declaration *ast.Node
	bindings    []importedBinding
	violations  []importedBindingViolation
}

type importedBindingTarget struct {
	groupIndex   int
	bindingIndex int
}

type importedBindingTargets struct {
	first      importedBindingTarget
	additional []importedBindingTarget
}

type noImportAssignRefStoreMode uint8

const (
	noImportAssignRefStoreAuto noImportAssignRefStoreMode = iota
	noImportAssignRefStoreDisabled
	noImportAssignRefStoreEnabled
)

func makeImportedBinding(nameNode *ast.Node, isNamespace bool, symbol *ast.Symbol) importedBinding {
	return importedBinding{
		name:        nameNode.Text(),
		isNamespace: isNamespace,
		nameNode:    nameNode,
		symbol:      symbol,
	}
}

// importBindingSymbol returns the binder-owned alias symbol used by ctx.Refs.
func importBindingSymbol(nameNode *ast.Node) *ast.Symbol {
	if nameNode == nil || nameNode.Parent == nil || nameNode.Parent.Name() != nameNode {
		return nil
	}
	return nameNode.Parent.Symbol()
}

func checkerImportBindingSymbol(nameNode *ast.Node, ctx *rule.RuleContext) *ast.Symbol {
	if ctx.TypeChecker == nil {
		return nil
	}
	return ctx.TypeChecker.GetSymbolAtLocation(nameNode)
}

// wellKnownMutationMethods maps global object names to their mutation method names.
var wellKnownMutationMethods = map[string]map[string]bool{
	"Object": {
		"assign":           true,
		"defineProperty":   true,
		"defineProperties": true,
		"freeze":           true,
		"setPrototypeOf":   true,
	},
	"Reflect": {
		"defineProperty": true,
		"deleteProperty": true,
		"set":            true,
		"setPrototypeOf": true,
	},
}

// isArgumentOfWellKnownMutationFunction checks if a given node is the first argument
// of a well-known mutation function such as Object.assign, Object.defineProperty,
// Reflect.set, Reflect.setPrototypeOf, etc.
// It skips the check when Object/Reflect is locally shadowed (e.g. var Object).
func isArgumentOfWellKnownMutationFunction(node *ast.Node, ctx *rule.RuleContext) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	// Parentheses do not change the argument's identity, but TypeScript's AST
	// preserves them between the identifier and the call expression.
	argument := utils.OutermostParenthesizedExpression(node)
	if argument == nil || argument.Parent == nil {
		return false
	}
	parent := argument.Parent

	// The parent must be a CallExpression
	if parent.Kind != ast.KindCallExpression {
		return false
	}

	callExpr := parent.AsCallExpression()
	if callExpr == nil || callExpr.Arguments == nil || len(callExpr.Arguments.Nodes) == 0 {
		return false
	}

	// The node must be the first argument
	if callExpr.Arguments.Nodes[0] != argument {
		return false
	}

	// The callee must be a PropertyAccessExpression like Object.assign or Reflect.set
	// Unwrap parentheses: (Object?.defineProperty)(ns, ...) → Object?.defineProperty
	callee := ast.SkipParentheses(callExpr.Expression)
	if callee == nil {
		return false
	}
	var object *ast.Node
	var methodName string
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		propAccess := callee.AsPropertyAccessExpression()
		if propAccess == nil || propAccess.Expression == nil || propAccess.Name() == nil {
			return false
		}
		object = propAccess.Expression
		methodName = propAccess.Name().Text()
	case ast.KindElementAccessExpression:
		elementAccess := callee.AsElementAccessExpression()
		if elementAccess == nil || elementAccess.Expression == nil ||
			elementAccess.ArgumentExpression == nil {
			return false
		}
		var ok bool
		object = elementAccess.Expression
		methodName, ok = utils.GetStaticStringLiteralValue(
			ast.SkipParentheses(elementAccess.ArgumentExpression),
		)
		if !ok {
			return false
		}
	default:
		return false
	}

	// The object must be a simple identifier (Object or Reflect). Parentheses
	// around that identifier are transparent, as they are in ESTree.
	object = ast.SkipParentheses(object)
	if object == nil || object.Kind != ast.KindIdentifier {
		return false
	}

	objectName := object.Text()

	methods, ok := wellKnownMutationMethods[objectName]
	if !ok {
		return false
	}
	if !methods[methodName] {
		return false
	}

	// If Object/Reflect is locally shadowed, skip the check (same as no-console pattern).
	if ctx.TypeChecker != nil {
		sym := ctx.TypeChecker.GetSymbolAtLocation(object)
		if sym != nil {
			for _, decl := range sym.Declarations {
				declSF := ast.GetSourceFileOfNode(decl)
				if declSF != nil && declSF == ctx.SourceFile {
					return false
				}
			}
		}
	}

	return true
}

// isWrappedInTypeAssertion checks whether a node is wrapped in a type assertion
// (e.g. `as any`, `<any>`, or `!`) before reaching the actual write position.
// When a developer writes `(ns.prop as any) = value`, the `as any` is an intentional
// type-level escape hatch; ESLint does not flag such cases, so we skip them too.
func isWrappedInTypeAssertion(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindParenthesizedExpression:
			current = current.Parent
		case ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindNonNullExpression:
			return true
		default:
			return false
		}
	}
	return false
}

// isMemberExpressionWrite checks if a member expression (PropertyAccess or ElementAccess)
// is a write target: assignment left side, update expression operand, delete operand,
// or for-in/of initializer.
func isMemberExpressionWrite(memberExpr *ast.Node) bool {
	// Type assertion wrappers (as any, <any>, !) indicate intentional bypass — skip.
	if isWrappedInTypeAssertion(memberExpr) {
		return false
	}
	if utils.IsWriteReference(memberExpr) {
		return true
	}
	// IsWriteReference does not handle delete expressions, so check separately.
	deleteTarget := utils.OutermostParenthesizedExpression(memberExpr)
	if deleteTarget != nil && deleteTarget.Parent != nil &&
		deleteTarget.Parent.Kind == ast.KindDeleteExpression {
		return true
	}
	return false
}

// isMemberWrite checks if an identifier is the object in a member-write expression.
// For namespace imports like `import * as ns`, member writes include:
//   - ns.prop = val
//   - ns.prop++
//   - ns["prop"] = val
//   - delete ns.prop
//   - Object.assign(ns, ...)
//   - Object.defineProperty(ns, ...)
//   - Reflect.set(ns, ...) etc.
func isMemberWrite(node *ast.Node, ctx *rule.RuleContext) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	// Parentheses around the namespace object are transparent:
	// `((ns)).prop = value` mutates the same namespace as `ns.prop = value`.
	reference := utils.OutermostParenthesizedExpression(node)
	if reference == nil || reference.Parent == nil {
		return false
	}
	parent := reference.Parent

	// Check ns.prop = val, ns.prop++, ns["prop"] = val, delete ns.prop, etc.
	if parent.Kind == ast.KindPropertyAccessExpression {
		propAccess := parent.AsPropertyAccessExpression()
		if propAccess != nil && propAccess.Expression == reference {
			return isMemberExpressionWrite(parent)
		}
	}

	if parent.Kind == ast.KindElementAccessExpression {
		elemAccess := parent.AsElementAccessExpression()
		if elemAccess != nil && elemAccess.Expression == reference {
			return isMemberExpressionWrite(parent)
		}
	}

	// Check spread into destructuring assignment target: ({...ns} = obj)
	if parent.Kind == ast.KindSpreadAssignment {
		return utils.IsInDestructuringAssignment(parent)
	}

	// Check if the identifier is the first argument of a well-known mutation function
	// e.g., Object.assign(ns, ...), Reflect.set(ns, ...), etc.
	if isArgumentOfWellKnownMutationFunction(node, ctx) {
		return true
	}

	// Check for...in/of: for (ns.prop in ...) or for (ns.prop of ...)
	// These are caught by the PropertyAccessExpression + isMemberExpressionWrite path above
	// since IsWriteReference handles for-in/of initializers.

	return false
}

// isImportBindingName checks if the identifier is a declaration name within an import.
func isImportBindingName(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	parent := node.Parent
	switch parent.Kind {
	case ast.KindImportClause, ast.KindNamespaceImport, ast.KindImportSpecifier:
		return true
	}
	return false
}

func forEachImportedBindingName(
	node *ast.Node,
	visit func(nameNode *ast.Node, isNamespace bool),
) {
	importDecl := node.AsImportDeclaration()
	if importDecl == nil || importDecl.ImportClause == nil {
		return
	}

	importClause := importDecl.ImportClause.AsImportClause()
	if importClause == nil {
		return
	}

	// Default import: import foo from "mod"
	if name := importClause.Name(); name != nil {
		visit(name, false)
	}

	if importClause.NamedBindings == nil {
		return
	}

	switch importClause.NamedBindings.Kind {
	case ast.KindNamespaceImport:
		namespaceImport := importClause.NamedBindings.AsNamespaceImport()
		if namespaceImport != nil && namespaceImport.Name() != nil {
			visit(namespaceImport.Name(), true)
		}
	case ast.KindNamedImports:
		namedImports := importClause.NamedBindings.AsNamedImports()
		if namedImports != nil && namedImports.Elements != nil {
			for _, element := range namedImports.Elements.Nodes {
				importSpecifier := element.AsImportSpecifier()
				if importSpecifier != nil && importSpecifier.Name() != nil {
					visit(importSpecifier.Name(), false)
				}
			}
		}
	}
}

func collectImportedBindings(
	node *ast.Node,
	resolveSymbol func(nameNode *ast.Node) *ast.Symbol,
) []importedBinding {
	var bindings []importedBinding
	forEachImportedBindingName(node, func(nameNode *ast.Node, isNamespace bool) {
		bindings = append(bindings, makeImportedBinding(
			nameNode,
			isNamespace,
			resolveSymbol(nameNode),
		))
	})
	return bindings
}

func countTopLevelImportedBindings(sourceFile *ast.SourceFile, limit int) int {
	if sourceFile == nil || sourceFile.Statements == nil {
		return 0
	}

	count := 0
	for _, statement := range sourceFile.Statements.Nodes {
		if statement.Kind != ast.KindImportDeclaration {
			continue
		}
		forEachImportedBindingName(statement, func(*ast.Node, bool) {
			count++
		})
		if count >= limit {
			return count
		}
	}
	return count
}

func classifyImportedBindingViolation(
	node *ast.Node,
	binding *importedBinding,
	ctx *rule.RuleContext,
) importedBindingViolationKind {
	if utils.IsWriteReference(node) {
		return importedBindingViolationDirect
	}
	if binding.isNamespace && isMemberWrite(node, ctx) {
		return importedBindingViolationMember
	}
	return importedBindingViolationNone
}

func importedBindingMessage(
	binding *importedBinding,
	kind importedBindingViolationKind,
	cache *importedBindingMessageCache,
) rule.RuleMessage {
	switch kind {
	case importedBindingViolationDirect:
		description := cache.direct
		if description == "" {
			description = "'" + binding.name + "' is read-only."
			cache.direct = description
		}
		return rule.RuleMessage{
			Id:          "readonly",
			Description: description,
		}
	case importedBindingViolationMember:
		description := cache.member
		if description == "" {
			description = "The members of '" + binding.name + "' are read-only."
			cache.member = description
		}
		return rule.RuleMessage{
			Id:          "readonlyMember",
			Description: description,
		}
	}
	return rule.RuleMessage{}
}

func reportImportedBindingViolation(
	ctx *rule.RuleContext,
	binding *importedBinding,
	violation importedBindingViolation,
	cache *importedBindingMessageCache,
) {
	if violation.kind == importedBindingViolationNone {
		return
	}
	if cache == nil {
		switch violation.kind {
		case importedBindingViolationDirect:
			ctx.ReportNode(violation.node, rule.RuleMessage{
				Id:          "readonly",
				Description: "'" + binding.name + "' is read-only.",
			})
		case importedBindingViolationMember:
			ctx.ReportNode(violation.node, rule.RuleMessage{
				Id:          "readonlyMember",
				Description: "The members of '" + binding.name + "' are read-only.",
			})
		}
		return
	}
	ctx.ReportNode(
		violation.node,
		importedBindingMessage(binding, violation.kind, cache),
	)
}

func reportImportedBindingGroup(ctx *rule.RuleContext, bindings []importedBinding) {
	if len(bindings) == 1 {
		binding := &bindings[0]
		for _, reference := range ctx.Refs.References(binding.symbol) {
			kind := classifyImportedBindingViolation(reference, binding, ctx)
			if kind != importedBindingViolationNone {
				reportImportedBindingViolation(
					ctx,
					binding,
					importedBindingViolation{
						node: reference,
						kind: kind,
					},
					nil,
				)
			}
		}
		return
	}

	var violations []importedBindingViolation
	for bindingIndex := range bindings {
		binding := &bindings[bindingIndex]
		for _, reference := range ctx.Refs.References(binding.symbol) {
			kind := classifyImportedBindingViolation(reference, binding, ctx)
			if kind != importedBindingViolationNone {
				violations = append(violations, importedBindingViolation{
					node:         reference,
					bindingIndex: bindingIndex,
					kind:         kind,
				})
			}
		}
	}

	// The old implementation walked the source once per import declaration,
	// checking all of that declaration's bindings at each identifier. Preserve
	// its diagnostic order after merging the per-symbol reference lists.
	slices.SortStableFunc(violations, func(a importedBindingViolation, b importedBindingViolation) int {
		return cmp.Compare(a.node.Pos(), b.node.Pos())
	})
	for _, violation := range violations {
		reportImportedBindingViolation(
			ctx,
			&bindings[violation.bindingIndex],
			violation,
			nil,
		)
	}
}

// walkImportedBindingGroups scans the source once against all imported bindings
// collected by the normal listener traversal. It indexes only import names, so
// unrelated identifiers never enter a per-file reference map. Violations are
// bucketed by declaration to retain the rule's historical declaration-grouped
// diagnostic order.
func walkImportedBindingGroups(
	ctx *rule.RuleContext,
	groups []importedBindingGroup,
	resolveReferenceSymbol func(*ast.Node) *ast.Symbol,
) {
	if ctx.SourceFile == nil {
		return
	}

	bindingCount := 0
	for _, group := range groups {
		bindingCount += len(group.bindings)
	}

	targetsByName := make(map[string][]importedBindingTarget, bindingCount)
	for groupIndex := range groups {
		for bindingIndex, binding := range groups[groupIndex].bindings {
			targetsByName[binding.name] = append(
				targetsByName[binding.name],
				importedBindingTarget{
					groupIndex:   groupIndex,
					bindingIndex: bindingIndex,
				},
			)
		}
	}

	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}

		if node.Kind == ast.KindIdentifier {
			targets := targetsByName[node.Text()]
			if len(targets) != 0 && !isImportBindingName(node) {
				var referenceSymbol *ast.Symbol
				if resolveReferenceSymbol != nil {
					referenceSymbol = resolveReferenceSymbol(node)
				}
				for _, target := range targets {
					binding := &groups[target.groupIndex].bindings[target.bindingIndex]
					if binding.symbol != nil && resolveReferenceSymbol != nil &&
						referenceSymbol != binding.symbol {
						continue
					}

					kind := classifyImportedBindingViolation(node, binding, ctx)
					if kind != importedBindingViolationNone {
						groups[target.groupIndex].violations = append(
							groups[target.groupIndex].violations,
							importedBindingViolation{
								node:         node,
								bindingIndex: target.bindingIndex,
								kind:         kind,
							},
						)
					}
				}
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(ctx.SourceFile.AsNode())

	for groupIndex := range groups {
		group := &groups[groupIndex]
		for _, violation := range group.violations {
			reportImportedBindingViolation(
				ctx,
				&group.bindings[violation.bindingIndex],
				violation,
				nil,
			)
		}
	}
}

// walkImportedBindingReferences preserves the single-declaration compatibility
// path used by forced strategy tests and avoids a name index for one binding.
func walkImportedBindingReferences(
	ctx *rule.RuleContext,
	bindings []importedBinding,
	resolveReferenceSymbol func(*ast.Node) *ast.Symbol,
) {
	if ctx.SourceFile == nil {
		return
	}

	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}

		if node.Kind == ast.KindIdentifier {
			for bindingIndex := range bindings {
				binding := &bindings[bindingIndex]
				if node.Text() != binding.name || isImportBindingName(node) {
					continue
				}

				if binding.symbol != nil && resolveReferenceSymbol != nil &&
					resolveReferenceSymbol(node) != binding.symbol {
					continue
				}

				kind := classifyImportedBindingViolation(node, binding, ctx)
				if kind != importedBindingViolationNone {
					reportImportedBindingViolation(
						ctx,
						binding,
						importedBindingViolation{
							node: node,
							kind: kind,
						},
						nil,
					)
				}
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(ctx.SourceFile.AsNode())
}

func allBindingsHaveSymbols(bindings []importedBinding) bool {
	for _, binding := range bindings {
		if binding.symbol == nil {
			return false
		}
	}
	return true
}

func allBindingGroupsHaveSymbols(groups []importedBindingGroup) bool {
	for _, group := range groups {
		if !allBindingsHaveSymbols(group.bindings) {
			return false
		}
	}
	return true
}

func noImportAssignScanListeners(
	ctx *rule.RuleContext,
	refStoreMode noImportAssignRefStoreMode,
) rule.RuleListeners {
	var checkerReferenceSymbol func(*ast.Node) *ast.Symbol
	if ctx.TypeChecker != nil {
		checkerReferenceSymbol = func(node *ast.Node) *ast.Symbol {
			return utils.GetReferenceSymbol(node, ctx.TypeChecker)
		}
	}

	if refStoreMode != noImportAssignRefStoreAuto {
		useRefStore := refStoreMode == noImportAssignRefStoreEnabled && ctx.Refs != nil
		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				if useRefStore {
					bindings := collectImportedBindings(node, importBindingSymbol)
					if len(bindings) == 0 {
						return
					}
					if allBindingsHaveSymbols(bindings) {
						reportImportedBindingGroup(ctx, bindings)
						return
					}
				}

				bindings := collectImportedBindings(node, func(nameNode *ast.Node) *ast.Symbol {
					return checkerImportBindingSymbol(nameNode, ctx)
				})
				if len(bindings) != 0 {
					walkImportedBindingReferences(ctx, bindings, checkerReferenceSymbol)
				}
			},
		}
	}

	resolveBindingSymbol := func(*ast.Node) *ast.Symbol { return nil }
	resolveReferenceSymbol := checkerReferenceSymbol
	if ctx.Refs != nil {
		resolveBindingSymbol = importBindingSymbol
		resolveReferenceSymbol = ctx.Refs.Resolve
	} else if ctx.TypeChecker != nil {
		resolveBindingSymbol = func(nameNode *ast.Node) *ast.Symbol {
			return checkerImportBindingSymbol(nameNode, ctx)
		}
	}

	var firstGroup importedBindingGroup
	var remainingGroups []importedBindingGroup
	bindingCount := 0
	return rule.RuleListeners{
		ast.KindImportDeclaration: func(node *ast.Node) {
			bindings := collectImportedBindings(node, resolveBindingSymbol)
			if len(bindings) != 0 {
				group := importedBindingGroup{bindings: bindings}
				if firstGroup.bindings == nil {
					firstGroup = group
				} else {
					remainingGroups = append(remainingGroups, group)
				}
				bindingCount += len(bindings)
			}
		},
		ast.KindEndOfFile: func(*ast.Node) {
			if bindingCount == 0 {
				return
			}

			// Without a checker, RefStore.Resolve keeps the one-binding path
			// exact without materializing the reverse-reference index.
			if bindingCount == 1 {
				walkImportedBindingReferences(
					ctx,
					firstGroup.bindings,
					resolveReferenceSymbol,
				)
				return
			}

			groups := make([]importedBindingGroup, 1, 1+len(remainingGroups))
			groups[0] = firstGroup
			groups = append(groups, remainingGroups...)

			// Recovered syntax can leave an import name without a binder symbol.
			// In that case, keep the old checker fallback rather than mixing
			// symbol domains or degrading that binding to text-only matching.
			if ctx.Refs != nil && ctx.TypeChecker != nil &&
				!allBindingGroupsHaveSymbols(groups) {
				for groupIndex := range groups {
					for bindingIndex := range groups[groupIndex].bindings {
						binding := &groups[groupIndex].bindings[bindingIndex]
						binding.symbol = checkerImportBindingSymbol(binding.nameNode, ctx)
					}
				}
				resolveReferenceSymbol = checkerReferenceSymbol
			}

			// A single selective scan avoids both repeated walks and RefStore's
			// unrelated-name buckets. RefStore.Resolve provides binder identity
			// without materializing the full reverse-reference index.
			walkImportedBindingGroups(ctx, groups, resolveReferenceSymbol)
		},
	}
}

func collectTopLevelImportedBindingGroups(
	sourceFile *ast.SourceFile,
	resolveBindingSymbol func(*ast.Node) *ast.Symbol,
) ([]importedBindingGroup, int) {
	if sourceFile == nil || sourceFile.Statements == nil {
		return nil, 0
	}

	var groups []importedBindingGroup
	bindingCount := 0
	for _, statement := range sourceFile.Statements.Nodes {
		if statement.Kind != ast.KindImportDeclaration {
			continue
		}
		bindings := collectImportedBindings(statement, resolveBindingSymbol)
		if len(bindings) == 0 {
			continue
		}
		groups = append(groups, importedBindingGroup{
			declaration: statement,
			bindings:    bindings,
		})
		bindingCount += len(bindings)
	}
	return groups, bindingCount
}

func replaceImportedBindingSymbolsWithChecker(
	groups []importedBindingGroup,
	ctx *rule.RuleContext,
) {
	for groupIndex := range groups {
		for bindingIndex := range groups[groupIndex].bindings {
			binding := &groups[groupIndex].bindings[bindingIndex]
			binding.symbol = checkerImportBindingSymbol(binding.nameNode, ctx)
		}
	}
}

func collectImportedBindingViolation(
	ctx *rule.RuleContext,
	groups []importedBindingGroup,
	targets importedBindingTargets,
	node *ast.Node,
	resolveReferenceSymbol func(*ast.Node) *ast.Symbol,
) {
	if isImportBindingName(node) {
		return
	}

	kind := importedBindingViolationNone
	if utils.IsWriteReference(node) {
		kind = importedBindingViolationDirect
	} else {
		for targetIndex := 0; targetIndex <= len(targets.additional); targetIndex++ {
			target := targets.first
			if targetIndex != 0 {
				target = targets.additional[targetIndex-1]
			}
			if groups[target.groupIndex].bindings[target.bindingIndex].isNamespace {
				if isMemberWrite(node, ctx) {
					kind = importedBindingViolationMember
				}
				break
			}
		}
	}
	if kind == importedBindingViolationNone {
		return
	}

	var referenceSymbol *ast.Symbol
	if resolveReferenceSymbol != nil {
		referenceSymbol = resolveReferenceSymbol(node)
	}
	for targetIndex := 0; targetIndex <= len(targets.additional); targetIndex++ {
		target := targets.first
		if targetIndex != 0 {
			target = targets.additional[targetIndex-1]
		}
		binding := &groups[target.groupIndex].bindings[target.bindingIndex]
		if kind == importedBindingViolationMember && !binding.isNamespace {
			continue
		}
		if binding.symbol != nil && resolveReferenceSymbol != nil &&
			referenceSymbol != binding.symbol {
			continue
		}
		groups[target.groupIndex].violations = append(
			groups[target.groupIndex].violations,
			importedBindingViolation{
				node:         node,
				bindingIndex: target.bindingIndex,
				kind:         kind,
			},
		)
	}
}

func forEachIdentifierInWriteTarget(node *ast.Node, visit func(*ast.Node)) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindIdentifier {
		visit(node)
		return false
	}
	// Nested writes and calls receive their own listener callback later in the
	// shared traversal. Leave them to that callback so one identifier cannot
	// be collected twice from overlapping write targets.
	switch node.Kind {
	case ast.KindBinaryExpression:
		if ast.IsAssignmentExpression(node, false) {
			return true
		}
	case ast.KindPrefixUnaryExpression:
		operator := node.AsPrefixUnaryExpression().Operator
		if operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken {
			return true
		}
	case ast.KindPostfixUnaryExpression, ast.KindDeleteExpression, ast.KindCallExpression:
		return true
	}
	deferredNestedWrite := false
	node.ForEachChild(func(child *ast.Node) bool {
		if forEachIdentifierInWriteTarget(child, visit) {
			deferredNestedWrite = true
		}
		return false
	})
	return deferredNestedWrite
}

func noImportAssignIntegratedListeners(
	ctx rule.RuleContext,
	groups []importedBindingGroup,
	bindingCount int,
	resolveBindingSymbol func(*ast.Node) *ast.Symbol,
	resolveReferenceSymbol func(*ast.Node) *ast.Symbol,
	checkerReferenceSymbol func(*ast.Node) *ast.Symbol,
) rule.RuleListeners {
	if ctx.Refs != nil && ctx.TypeChecker != nil &&
		!allBindingGroupsHaveSymbols(groups) {
		replaceImportedBindingSymbolsWithChecker(groups, &ctx)
		resolveBindingSymbol = func(nameNode *ast.Node) *ast.Symbol {
			return checkerImportBindingSymbol(nameNode, &ctx)
		}
		resolveReferenceSymbol = checkerReferenceSymbol
	}

	targetsByName := make(map[string]importedBindingTargets, bindingCount)
	hasNamespaceBinding := false
	for groupIndex := range groups {
		for bindingIndex, binding := range groups[groupIndex].bindings {
			hasNamespaceBinding = hasNamespaceBinding || binding.isNamespace
			targets, exists := targetsByName[binding.name]
			target := importedBindingTarget{
				groupIndex:   groupIndex,
				bindingIndex: bindingIndex,
			}
			if exists {
				targets.additional = append(targets.additional, target)
			} else {
				targets.first = target
			}
			targetsByName[binding.name] = targets
		}
	}

	needsFallbackScan := false
	needsViolationSort := false
	checkIdentifier := func(node *ast.Node) {
		if needsFallbackScan {
			return
		}
		targets, exists := targetsByName[node.Text()]
		if !exists {
			return
		}
		collectImportedBindingViolation(
			&ctx,
			groups,
			targets,
			node,
			resolveReferenceSymbol,
		)
	}
	checkWriteTarget := func(node *ast.Node) {
		if forEachIdentifierInWriteTarget(node, checkIdentifier) {
			needsViolationSort = true
		}
	}

	listeners := rule.RuleListeners{
		ast.KindImportDeclaration: func(node *ast.Node) {
			if node.Parent != nil && node.Parent.Kind == ast.KindSourceFile {
				return
			}
			bindings := collectImportedBindings(node, resolveBindingSymbol)
			if len(bindings) == 0 {
				return
			}
			groups = append(groups, importedBindingGroup{
				declaration: node,
				bindings:    bindings,
			})
			needsFallbackScan = true
		},
		ast.KindBinaryExpression: func(node *ast.Node) {
			if !ast.IsAssignmentExpression(node, false) {
				return
			}
			binary := node.AsBinaryExpression()
			if binary == nil {
				return
			}
			checkWriteTarget(binary.Left)
		},
		ast.KindPrefixUnaryExpression: func(node *ast.Node) {
			prefix := node.AsPrefixUnaryExpression()
			if prefix == nil ||
				(prefix.Operator != ast.KindPlusPlusToken &&
					prefix.Operator != ast.KindMinusMinusToken) {
				return
			}
			checkWriteTarget(prefix.Operand)
		},
		ast.KindPostfixUnaryExpression: func(node *ast.Node) {
			postfix := node.AsPostfixUnaryExpression()
			if postfix == nil ||
				(postfix.Operator != ast.KindPlusPlusToken &&
					postfix.Operator != ast.KindMinusMinusToken) {
				return
			}
			checkWriteTarget(postfix.Operand)
		},
		ast.KindForInStatement: func(node *ast.Node) {
			statement := node.AsForInOrOfStatement()
			if statement != nil {
				checkWriteTarget(statement.Initializer)
			}
		},
		ast.KindForOfStatement: func(node *ast.Node) {
			statement := node.AsForInOrOfStatement()
			if statement != nil {
				checkWriteTarget(statement.Initializer)
			}
		},
		ast.KindEndOfFile: func(*ast.Node) {
			if needsFallbackScan {
				slices.SortStableFunc(groups, func(a, b importedBindingGroup) int {
					return cmp.Compare(a.declaration.Pos(), b.declaration.Pos())
				})
				for groupIndex := range groups {
					groups[groupIndex].violations = nil
				}
				if ctx.Refs != nil && ctx.TypeChecker != nil &&
					!allBindingGroupsHaveSymbols(groups) {
					replaceImportedBindingSymbolsWithChecker(groups, &ctx)
					resolveReferenceSymbol = checkerReferenceSymbol
				}
				walkImportedBindingGroups(&ctx, groups, resolveReferenceSymbol)
				return
			}

			violationCount := 0
			for groupIndex := range groups {
				violationCount += len(groups[groupIndex].violations)
			}
			var messageCaches []importedBindingMessageCache
			if violationCount > bindingCount {
				messageCaches = make([]importedBindingMessageCache, bindingCount)
			}

			cacheOffset := 0
			for groupIndex := range groups {
				group := &groups[groupIndex]
				if needsViolationSort {
					slices.SortStableFunc(
						group.violations,
						func(a, b importedBindingViolation) int {
							return cmp.Compare(a.node.Pos(), b.node.Pos())
						},
					)
				}
				for _, violation := range group.violations {
					var cache *importedBindingMessageCache
					if messageCaches != nil {
						cache = &messageCaches[cacheOffset+violation.bindingIndex]
					}
					reportImportedBindingViolation(
						&ctx,
						&group.bindings[violation.bindingIndex],
						violation,
						cache,
					)
				}
				cacheOffset += len(group.bindings)
			}
		},
	}

	if hasNamespaceBinding {
		listeners[ast.KindDeleteExpression] = func(node *ast.Node) {
			deleteExpression := node.AsDeleteExpression()
			if deleteExpression != nil {
				checkWriteTarget(deleteExpression.Expression)
			}
		}
		listeners[ast.KindCallExpression] = func(node *ast.Node) {
			callExpression := node.AsCallExpression()
			if callExpression == nil || callExpression.Arguments == nil ||
				len(callExpression.Arguments.Nodes) == 0 {
				return
			}
			checkWriteTarget(callExpression.Arguments.Nodes[0])
		}
	}
	return listeners
}

func noImportAssignListeners(
	ctx rule.RuleContext,
	refStoreMode noImportAssignRefStoreMode,
) rule.RuleListeners {
	if refStoreMode != noImportAssignRefStoreAuto {
		return noImportAssignScanListeners(&ctx, refStoreMode)
	}
	topLevelBindingCount := countTopLevelImportedBindings(ctx.SourceFile, 2)
	if topLevelBindingCount < 2 {
		// An ImportDeclaration always contains the literal keyword in source.
		// Avoid registering listeners when even recovered syntax cannot contain
		// one; otherwise keep the compatibility listener for nested imports.
		if topLevelBindingCount == 0 && ctx.SourceFile != nil &&
			!strings.Contains(ctx.SourceFile.Text(), "import") {
			return nil
		}
		if ctx.TypeChecker != nil {
			return noImportAssignScanListeners(&ctx, noImportAssignRefStoreDisabled)
		}
		return noImportAssignScanListeners(&ctx, noImportAssignRefStoreAuto)
	}

	var checkerReferenceSymbol func(*ast.Node) *ast.Symbol
	if ctx.TypeChecker != nil {
		checkerReferenceSymbol = func(node *ast.Node) *ast.Symbol {
			return utils.GetReferenceSymbol(node, ctx.TypeChecker)
		}
	}

	resolveBindingSymbol := func(*ast.Node) *ast.Symbol { return nil }
	resolveReferenceSymbol := checkerReferenceSymbol
	if ctx.Refs != nil {
		resolveBindingSymbol = importBindingSymbol
		resolveReferenceSymbol = ctx.Refs.Resolve
	} else if ctx.TypeChecker != nil {
		resolveBindingSymbol = func(nameNode *ast.Node) *ast.Symbol {
			return checkerImportBindingSymbol(nameNode, &ctx)
		}
	}

	groups, bindingCount := collectTopLevelImportedBindingGroups(
		ctx.SourceFile,
		resolveBindingSymbol,
	)
	return noImportAssignIntegratedListeners(
		ctx,
		groups,
		bindingCount,
		resolveBindingSymbol,
		resolveReferenceSymbol,
		checkerReferenceSymbol,
	)
}

// NoImportAssignRule disallows assigning to imported bindings.
var NoImportAssignRule = rule.Rule{
	Name: "no-import-assign",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return noImportAssignListeners(ctx, noImportAssignRefStoreAuto)
	},
}
