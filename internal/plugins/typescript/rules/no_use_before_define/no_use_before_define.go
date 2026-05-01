package no_use_before_define

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// options mirrors the typescript-eslint rule schema.
// See https://typescript-eslint.io/rules/no-use-before-define/#options
type options struct {
	functions            bool
	classes              bool
	variables            bool
	enums                bool
	typedefs             bool
	ignoreTypeReferences bool
	allowNamedExports    bool
}

func parseOptions(rawOptions any) options {
	opts := options{
		functions:            true,
		classes:              true,
		variables:            true,
		enums:                true,
		typedefs:             true,
		ignoreTypeReferences: true,
		allowNamedExports:    false,
	}

	// Handle "nofunc" string option.
	if str, ok := rawOptions.(string); ok {
		if str == "nofunc" {
			opts.functions = false
		}
		return opts
	}

	// Handle array format from JS tests: ["nofunc"] or [{...}]
	if arr, ok := rawOptions.([]interface{}); ok && len(arr) > 0 {
		if str, ok := arr[0].(string); ok {
			if str == "nofunc" {
				opts.functions = false
			}
			return opts
		}
	}

	optsMap := utils.GetOptionsMap(rawOptions)
	if optsMap == nil {
		return opts
	}

	if v, ok := optsMap["functions"].(bool); ok {
		opts.functions = v
	}
	if v, ok := optsMap["classes"].(bool); ok {
		opts.classes = v
	}
	if v, ok := optsMap["variables"].(bool); ok {
		opts.variables = v
	}
	if v, ok := optsMap["enums"].(bool); ok {
		opts.enums = v
	}
	if v, ok := optsMap["typedefs"].(bool); ok {
		opts.typedefs = v
	}
	if v, ok := optsMap["ignoreTypeReferences"].(bool); ok {
		opts.ignoreTypeReferences = v
	}
	if v, ok := optsMap["allowNamedExports"].(bool); ok {
		opts.allowNamedExports = v
	}

	return opts
}

// definitionType categorizes a declaration for option-based filtering.
// Maps to ESLint's scope-manager DefinitionType.
type definitionType int

const (
	defUnknown definitionType = iota
	defVariable
	defFunctionName
	defClassName
	defTypeName      // interface or type alias
	defEnumName      // enum declaration
	defNamespaceName // module/namespace declaration
	defImport
	defCatchClause
	defParameter
)

func getDefinitionType(decl *ast.Node) definitionType {
	if decl == nil {
		return defUnknown
	}
	switch decl.Kind {
	case ast.KindVariableDeclaration, ast.KindBindingElement:
		return defVariable
	case ast.KindFunctionDeclaration:
		return defFunctionName
	case ast.KindClassDeclaration, ast.KindClassExpression:
		return defClassName
	case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
		return defTypeName
	case ast.KindEnumDeclaration:
		return defEnumName
	case ast.KindModuleDeclaration:
		return defNamespaceName
	case ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport, ast.KindImportEqualsDeclaration:
		return defImport
	case ast.KindCatchClause:
		return defCatchClause
	case ast.KindParameter:
		return defParameter
	}
	return defUnknown
}

// isNamedExport checks if the identifier is the local name in an export specifier.
// e.g. the `a` in `export { a }` or `export { a as b }`.
func isNamedExport(node *ast.Node) bool {
	if node.Parent == nil || node.Parent.Kind != ast.KindExportSpecifier {
		return false
	}
	spec := node.Parent.AsExportSpecifier()
	// For `export { a as b }`: PropertyName=a (local), Name()=b (exported).
	if spec.PropertyName != nil {
		return spec.PropertyName == node
	}
	return spec.Name() == node
}

// isExportedAliasName checks if the identifier is the non-local export alias.
// e.g. the `b` in `export { a as b }` — not a real reference.
func isExportedAliasName(node *ast.Node) bool {
	if node.Parent == nil || node.Parent.Kind != ast.KindExportSpecifier {
		return false
	}
	spec := node.Parent.AsExportSpecifier()
	return spec.PropertyName != nil && spec.Name() == node
}

// isTypeReference reports whether the identifier sits in a type-only context.
//
// Composes tsgo helpers that mirror the TypeScript compiler:
//   - IsPartOfTypeNode: covers TypeReference, QualifiedName chains in type
//     nodes, and ExpressionWithTypeArguments in every heritage clause except
//     a class's own `extends` (evaluated as a value at runtime).
//   - IsPartOfTypeQuery: covers identifiers inside `typeof T`.
//
// IsPartOfTypeNode only walks up through the RIGHT side of property access
// chains, so the leftmost identifier of a qualified heritage target (e.g. the
// `ns` in `extends ns.B`) is not caught. The final block handles that by
// walking up through PropertyAccessExpression to its enclosing
// ExpressionWithTypeArguments and reusing
// IsExpressionWithTypeArgumentsInClassExtendsClause to exclude class extends.
func isTypeReference(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	if ast.IsPartOfTypeNode(node) || ast.IsPartOfTypeQuery(node) {
		return true
	}
	current := node.Parent
	for current != nil && current.Kind == ast.KindPropertyAccessExpression {
		current = current.Parent
	}
	return current != nil &&
		ast.IsExpressionWithTypeArguments(current) &&
		!ast.IsExpressionWithTypeArgumentsInClassExtendsClause(current)
}

// isInFunctionTypeScope checks if the identifier is inside a function type
// annotation (e.g. `type F = (x: Foo) => void`). References there live in a
// separate scope and should not be checked.
func isInFunctionTypeScope(node *ast.Node) bool {
	found := ast.FindAncestor(node.Parent, func(n *ast.Node) bool {
		switch n.Kind {
		case ast.KindFunctionType, ast.KindConstructorType:
			return true
		// Stop at real scope boundaries.
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindConstructor,
			ast.KindClassDeclaration, ast.KindClassExpression,
			ast.KindSourceFile:
			return true
		}
		return false
	})
	if found == nil {
		return false
	}
	return found.Kind == ast.KindFunctionType || found.Kind == ast.KindConstructorType
}

// getEnclosingFunctionScope returns the nearest enclosing function-like node
// that creates a new variable scope, or nil for program-level code.
//
// This models ESLint's "variableScope" concept:
//   - Static class field initializers and static blocks run synchronously during
//     class evaluation, so they are part of the outer execution context (skipped).
//   - Instance field initializers are conceptually separate execution contexts.
func getEnclosingFunctionScope(node *ast.Node) *ast.Node {
	current := node.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindConstructor,
			ast.KindGetAccessor, ast.KindSetAccessor:
			return current
		case ast.KindClassStaticBlockDeclaration:
			// Static blocks run during class definition — same execution context.
			current = current.Parent
			continue
		case ast.KindPropertyDeclaration:
			// Static field initializers: same execution context as class definition.
			// Instance field initializers: separate execution context.
			if ast.HasStaticModifier(current) && current.AsPropertyDeclaration().Initializer != nil {
				current = current.Parent
				continue
			}
			return current
		}
		current = current.Parent
	}
	return nil
}

// isFromSeparateExecutionContext returns true when reference and declaration
// live in different function-level scopes.
func isFromSeparateExecutionContext(refNode *ast.Node, declNode *ast.Node) bool {
	return getEnclosingFunctionScope(refNode) != getEnclosingFunctionScope(declNode)
}

// isInRange checks if a source position falls within [node.Pos(), node.End()].
func isInRange(node *ast.Node, location int) bool {
	return node != nil && node.Pos() <= location && location <= node.End()
}

// hasScopeBoundaryBetween returns true if there is a class or function scope
// boundary between the reference node and the declaration node. This matches
// the typescript-eslint `variable.scope !== reference.from` check.
func hasScopeBoundaryBetween(refNode *ast.Node, declNode *ast.Node) bool {
	current := refNode.Parent
	for current != nil && current != declNode {
		switch current.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			return true
		}
		current = current.Parent
	}
	return false
}

// isInsideModuleAugmentation checks if a declaration is inside (or is)
// a `declare module '...' { ... }` augmentation block.
func isInsideModuleAugmentation(node *ast.Node) bool {
	return ast.FindAncestor(node, ast.IsExternalModuleAugmentation) != nil
}

// isClassRefInClassDecorator returns true if the reference appears inside a
// decorator of the class it refers to. Decorators are transpiled after the
// class declaration, so such references are safe.
func isClassRefInClassDecorator(decl *ast.Node, refNode *ast.Node) bool {
	if decl.Kind != ast.KindClassDeclaration {
		return false
	}
	decorators := decl.Decorators()
	if len(decorators) == 0 {
		return false
	}
	refStart := refNode.Pos()
	refEnd := refNode.End()
	for _, deco := range decorators {
		if refStart >= deco.Pos() && refEnd <= deco.End() {
			return true
		}
	}
	return false
}

// isSentinelKind returns true for node kinds that form scope boundaries and
// stop the isEvaluatedDuringInitialization walk (matching ESLint's SENTINEL_TYPE).
func isSentinelKind(kind ast.Kind) bool {
	switch kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
		ast.KindClassDeclaration, ast.KindClassExpression,
		ast.KindArrowFunction, ast.KindCatchClause,
		ast.KindImportDeclaration, ast.KindExportDeclaration,
		// Method-like kinds create separate execution contexts and must stop
		// the isEvaluatedDuringInitialization walk. Without this, parameters
		// inside methods would incorrectly "escape" into enclosing variable
		// initializers. ESLint doesn't need these because in its AST model
		// methods are FunctionExpressions.
		ast.KindMethodDeclaration, ast.KindConstructor,
		ast.KindGetAccessor, ast.KindSetAccessor:
		return true
	}
	return false
}

// isEvaluatedDuringInitialization returns true when the reference is evaluated
// during the initialization of its own declaration. Examples:
//
//	var a = a         // self-referencing initializer
//	var [a = a] = []  // destructuring default
//	for (var a in a)  // for-in/for-of right-hand side
//
// NOTE: Unlike the ESLint core rule, the typescript-eslint version does NOT
// check class body evaluation (extends clause, computed property keys, etc.).
// It relies purely on source-position comparison for class declarations.
// See: https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/src/rules/no-use-before-define.ts
func isEvaluatedDuringInitialization(refNode *ast.Node, decl *ast.Node) bool {
	if isFromSeparateExecutionContext(refNode, decl) {
		return false
	}

	// The typescript-eslint rule's isInInitializer checks
	// `variable.scope !== reference.from` first. If the reference crosses a
	// scope boundary (class, block, etc.) relative to the declaration, it is
	// NOT considered "in the initializer" even when positionally inside the
	// initializer's range. We approximate this by checking whether any
	// class/function/block scope exists between the reference and the decl.
	if hasScopeBoundaryBetween(refNode, decl) {
		return false
	}

	location := refNode.End()

	// Only check variable/binding-element initializers (not class bodies).
	declName := utils.GetDeclarationIdentifier(decl)
	if declName == nil {
		return false
	}
	for node := declName.Parent; node != nil; node = node.Parent {
		switch node.Kind {
		case ast.KindVariableDeclaration:
			varDecl := node.AsVariableDeclaration()
			if varDecl.Initializer != nil && isInRange(varDecl.Initializer, location) {
				return true
			}
			// Check for-in/for-of right-hand side.
			if varDeclList := node.Parent; varDeclList != nil && varDeclList.Parent != nil {
				forStmt := varDeclList.Parent
				if forStmt.Kind == ast.KindForInStatement || forStmt.Kind == ast.KindForOfStatement {
					if isInRange(forStmt.AsForInOrOfStatement().Expression, location) {
						return true
					}
				}
			}
			return false
		case ast.KindBindingElement:
			if init := node.AsBindingElement().Initializer; init != nil && isInRange(init, location) {
				return true
			}
		case ast.KindParameter:
			if init := node.AsParameterDeclaration().Initializer; init != nil && isInRange(init, location) {
				return true
			}
		default:
			if isSentinelKind(node.Kind) {
				return false
			}
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Rule definition
// ---------------------------------------------------------------------------

var NoUseBeforeDefineRule = rule.CreateRule(rule.Rule{
	Name:             "no-use-before-define",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				if !utils.IsDeclarationIdentifier(node) {
					checkIdentifier(ctx, opts, node)
				}
			},
		}
	},
})

func checkIdentifier(ctx rule.RuleContext, opts options, node *ast.Node) {
	if ctx.TypeChecker == nil {
		return
	}

	// The renamed export name in `export { a as b }` is not a reference.
	if isExportedAliasName(node) {
		return
	}

	sym := ctx.TypeChecker.GetSymbolAtLocation(node)
	if sym == nil {
		return
	}

	// For alias symbols (import/export specifiers), resolve through the alias
	// to find the actual local declaration. Also include the alias's own
	// declarations (import specifiers) so that imports serve as the "definition
	// point" even when module augmentation adds later declarations.
	declarations := sym.Declarations
	if sym.Flags&ast.SymbolFlagsAlias != 0 {
		if resolved := ctx.TypeChecker.SkipAlias(sym); resolved != nil && len(resolved.Declarations) > 0 {
			declarations = append(append([]*ast.Node(nil), resolved.Declarations...), sym.Declarations...)
		}
	}
	if len(declarations) == 0 {
		return
	}

	// Find the earliest declaration in this source file, skipping export
	// specifiers (they are the reference site, not a definition) and module
	// augmentation declarations (they appear after the real definition).
	var firstDecl *ast.Node
	for _, decl := range declarations {
		if decl.Kind == ast.KindExportSpecifier {
			continue
		}
		if isInsideModuleAugmentation(decl) {
			continue
		}
		if ast.GetSourceFileOfNode(decl) == ctx.SourceFile {
			if firstDecl == nil || decl.Pos() < firstDecl.Pos() {
				firstDecl = decl
			}
		}
	}
	if firstDecl == nil {
		return
	}

	defType := getDefinitionType(firstDecl)
	declName := utils.GetDeclarationIdentifier(firstDecl)
	if declName == nil {
		return
	}

	// In a QualifiedName chain (A.B.C), only the leftmost name is a real reference.
	if node.Parent != nil && node.Parent.Kind == ast.KindQualifiedName {
		topQN := node.Parent
		for topQN.Parent != nil && topQN.Parent.Kind == ast.KindQualifiedName {
			topQN = topQN.Parent
		}
		leftmost := topQN.AsQualifiedName().Left
		for leftmost.Kind == ast.KindQualifiedName {
			leftmost = leftmost.AsQualifiedName().Left
		}
		if leftmost != node {
			return
		}
		// QualifiedName inside import-equals is a namespace alias, not a use.
		if topQN.Parent != nil && topQN.Parent.Kind == ast.KindImportEqualsDeclaration {
			return
		}
	}

	// Named exports: always check (ignoring other options) unless allowNamedExports.
	// For `export { X }`, we need the earliest local binding of X (the import
	// site or variable declaration), not a resolved alias target that might
	// include module augmentation declarations appearing later in the file.
	if isNamedExport(node) {
		if opts.allowNamedExports {
			return
		}
		localDeclName := findLocalBindingName(ctx, sym)
		if localDeclName != nil && isDefinedBeforeUse(localDeclName, node) {
			return
		}
		if localDeclName == nil && isDefinedBeforeUse(declName, node) {
			return
		}
		reportNode(ctx, node)
		return
	}

	// Defined before use — no violation, unless evaluated during its own initialization.
	if isDefinedBeforeUse(declName, node) &&
		(!isEvaluatedDuringInitialization(node, firstDecl) || node.Parent.Kind == ast.KindTypeReference) {
		return
	}

	// Option-based filtering.
	if !isForbidden(opts, defType, node, firstDecl) {
		return
	}

	if isClassRefInClassDecorator(firstDecl, node) {
		return
	}

	if isInFunctionTypeScope(node) {
		return
	}

	reportNode(ctx, node)
}

// findLocalBindingName walks the alias chain from sym back through imports
// to find the earliest local binding (import specifier / namespace import)
// name node in the current file. Returns nil if no local binding is found.
func findLocalBindingName(ctx rule.RuleContext, sym *ast.Symbol) *ast.Node {
	// Walk through the alias chain to find import-site declarations.
	current := sym
	for current != nil && current.Flags&ast.SymbolFlagsAlias != 0 {
		for _, d := range current.Declarations {
			if ast.GetSourceFileOfNode(d) != ctx.SourceFile {
				continue
			}
			switch d.Kind {
			case ast.KindImportSpecifier, ast.KindNamespaceImport, ast.KindImportClause, ast.KindImportEqualsDeclaration:
				return utils.GetDeclarationIdentifier(d)
			}
		}
		resolved := ctx.TypeChecker.SkipAlias(current)
		if resolved == current || resolved == nil {
			break
		}
		current = resolved
	}
	// Fallback: find any local non-augmentation declaration.
	for _, d := range sym.Declarations {
		if d.Kind == ast.KindExportSpecifier {
			continue
		}
		if ast.GetSourceFileOfNode(d) == ctx.SourceFile {
			return utils.GetDeclarationIdentifier(d)
		}
	}
	return nil
}

// isDefinedBeforeUse compares end positions (matching ESLint behavior).
func isDefinedBeforeUse(declName *ast.Node, refNode *ast.Node) bool {
	return declName.End() <= refNode.End()
}

// isForbidden decides whether a use-before-define should be reported based on
// the rule options and the declaration type.
//
// For classes, variables and enums the option only suppresses cross-scope
// references (different function scope). Same-scope TDZ violations are always
// reported regardless of the option value.
func isForbidden(opts options, defType definitionType, refNode *ast.Node, declNode *ast.Node) bool {
	if opts.ignoreTypeReferences && isTypeReference(refNode) {
		return false
	}

	switch defType {
	case defFunctionName:
		return opts.functions
	case defClassName:
		if isFromSeparateExecutionContext(refNode, declNode) {
			return opts.classes
		}
		return true // same scope — always report (TDZ)
	case defVariable:
		if isFromSeparateExecutionContext(refNode, declNode) {
			return opts.variables
		}
		return true
	case defEnumName:
		if isFromSeparateExecutionContext(refNode, declNode) {
			return opts.enums
		}
		return true
	case defTypeName:
		return opts.typedefs
	}

	return true
}

func reportNode(ctx rule.RuleContext, node *ast.Node) {
	name := ""
	if ast.IsIdentifier(node) {
		name = node.AsIdentifier().Text
	}
	ctx.ReportNode(node, rule.RuleMessage{
		Id:          "noUseBeforeDefine",
		Description: fmt.Sprintf("'%s' was used before it was defined.", name),
	})
}
