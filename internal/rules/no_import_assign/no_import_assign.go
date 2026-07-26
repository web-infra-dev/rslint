package no_import_assign

import (
	"cmp"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// importedBinding holds information about a single imported binding.
type importedBinding struct {
	name        string
	isNamespace bool
	symbol      *ast.Symbol
}

type importedBindingViolationKind uint8

const (
	importedBindingViolationNone importedBindingViolationKind = iota
	importedBindingViolationDirect
	importedBindingViolationMember
)

type importedBindingViolation struct {
	node    *ast.Node
	binding importedBinding
	kind    importedBindingViolationKind
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

func hasMultipleTopLevelImportedBindings(sourceFile *ast.SourceFile) bool {
	if sourceFile == nil || sourceFile.Statements == nil {
		return false
	}

	count := 0
	for _, statement := range sourceFile.Statements.Nodes {
		if statement.Kind != ast.KindImportDeclaration {
			continue
		}
		forEachImportedBindingName(statement, func(*ast.Node, bool) {
			count++
		})
		if count >= 2 {
			return true
		}
	}
	return false
}

func classifyImportedBindingViolation(
	node *ast.Node,
	binding importedBinding,
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

func reportImportedBindingViolation(ctx *rule.RuleContext, violation importedBindingViolation) {
	switch violation.kind {
	case importedBindingViolationDirect:
		ctx.ReportNode(violation.node, rule.RuleMessage{
			Id:          "readonly",
			Description: "'" + violation.binding.name + "' is read-only.",
		})
	case importedBindingViolationMember:
		ctx.ReportNode(violation.node, rule.RuleMessage{
			Id:          "readonlyMember",
			Description: "The members of '" + violation.binding.name + "' are read-only.",
		})
	}
}

func reportImportedBindingGroup(ctx *rule.RuleContext, bindings []importedBinding) {
	if len(bindings) == 1 {
		binding := bindings[0]
		for _, reference := range ctx.Refs.References(binding.symbol) {
			kind := classifyImportedBindingViolation(reference, binding, ctx)
			if kind != importedBindingViolationNone {
				reportImportedBindingViolation(ctx, importedBindingViolation{
					node:    reference,
					binding: binding,
					kind:    kind,
				})
			}
		}
		return
	}

	var violations []importedBindingViolation
	for _, binding := range bindings {
		for _, reference := range ctx.Refs.References(binding.symbol) {
			kind := classifyImportedBindingViolation(reference, binding, ctx)
			if kind != importedBindingViolationNone {
				violations = append(violations, importedBindingViolation{
					node:    reference,
					binding: binding,
					kind:    kind,
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
		reportImportedBindingViolation(ctx, violation)
	}
}

// walkImportedBindingReferences preserves the pre-RefStore implementation for
// single-binding files, direct RuleContext callers without a Program/Refs, and
// imports without binder symbols.
func walkImportedBindingReferences(ctx *rule.RuleContext, bindings []importedBinding) {
	if ctx.SourceFile == nil {
		return
	}

	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}

		if node.Kind == ast.KindIdentifier {
			for _, binding := range bindings {
				if node.Text() != binding.name || isImportBindingName(node) {
					continue
				}

				// Verify symbol identity when the type checker is available.
				if binding.symbol != nil && ctx.TypeChecker != nil &&
					utils.GetReferenceSymbol(node, ctx.TypeChecker) != binding.symbol {
					continue
				}

				kind := classifyImportedBindingViolation(node, binding, ctx)
				if kind != importedBindingViolationNone {
					reportImportedBindingViolation(ctx, importedBindingViolation{
						node:    node,
						binding: binding,
						kind:    kind,
					})
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

func noImportAssignListeners(
	ctx rule.RuleContext,
	refStoreMode noImportAssignRefStoreMode,
) rule.RuleListeners {
	refStoreModeResolved := refStoreMode != noImportAssignRefStoreAuto
	useRefStore := refStoreMode == noImportAssignRefStoreEnabled && ctx.Refs != nil

	return rule.RuleListeners{
		ast.KindImportDeclaration: func(node *ast.Node) {
			if !refStoreModeResolved {
				// One imported binding needs only one compatibility walk, which
				// avoids materializing RefStore buckets for every unrelated
				// identifier. At two bindings the shared index already wins,
				// and its advantage grows with the number of imports. Without a
				// checker, RefStore is also required for exact shadow resolution.
				useRefStore = ctx.Refs != nil &&
					(ctx.TypeChecker == nil ||
						hasMultipleTopLevelImportedBindings(ctx.SourceFile))
				refStoreModeResolved = true
			}

			if useRefStore {
				bindings := collectImportedBindings(node, importBindingSymbol)
				if len(bindings) == 0 {
					return
				}
				if allBindingsHaveSymbols(bindings) {
					reportImportedBindingGroup(&ctx, bindings)
					return
				}
			}

			bindings := collectImportedBindings(node, func(nameNode *ast.Node) *ast.Symbol {
				return checkerImportBindingSymbol(nameNode, &ctx)
			})
			if len(bindings) != 0 {
				walkImportedBindingReferences(&ctx, bindings)
			}
		},
	}
}

// NoImportAssignRule disallows assigning to imported bindings.
var NoImportAssignRule = rule.Rule{
	Name: "no-import-assign",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return noImportAssignListeners(ctx, noImportAssignRefStoreAuto)
	},
}
