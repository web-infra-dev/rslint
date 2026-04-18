package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// skipTransparentKinds matches parentheses + TS type assertions.
const skipTransparentKinds = ast.OEKParentheses | ast.OEKAssertions

// IsCallee checks if a node is the callee of a CallExpression or NewExpression,
// skipping parentheses and TS type assertions between the node and the call.
func IsCallee(node *ast.Node) bool {
	current := node
	parent := current.Parent
	for parent != nil && ast.IsOuterExpression(parent, skipTransparentKinds) {
		current = parent
		parent = current.Parent
	}
	if parent == nil {
		return false
	}
	if ast.IsCallExpression(parent) && parent.AsCallExpression().Expression == current {
		return true
	}
	if parent.Kind == ast.KindNewExpression && parent.AsNewExpression().Expression == current {
		return true
	}
	return false
}

// GetStaticStringValue returns the string value if the node is a string literal
// or a no-substitution template literal. Returns "" if the value cannot be
// statically determined.
func GetStaticStringValue(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text
	}
	return ""
}

// IsNonReferenceIdentifier checks if an identifier is NOT a value reference
// (i.e., it's a declaration name, property key, label, or module specifier name
// rather than a reference to a variable).
func IsNonReferenceIdentifier(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	// Property access name: a.b — `b` is a property key, not a variable reference.
	if parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Name() == node {
		return true
	}

	// Qualified type name: A.B.C (used in types) — right-hand names are not refs.
	if parent.Kind == ast.KindQualifiedName && parent.AsQualifiedName().Right == node {
		return true
	}

	// Meta property: new.target, import.meta — `target`/`meta` are keywords.
	if parent.Kind == ast.KindMetaProperty {
		return true
	}

	// Re-export specifiers: export { x } from 'mod'
	// All identifiers are source module names, not local references.
	if parent.Kind == ast.KindExportSpecifier && isReExportSpecifier(parent) {
		return true
	}

	// ast.IsDeclarationName covers: variable, function, class, parameter,
	// property assignment, method, accessor, enum member, etc.
	if ast.IsDeclarationName(node) {
		// ShorthandPropertyAssignment { x } — x IS a reference to the variable.
		if parent.Kind == ast.KindShorthandPropertyAssignment {
			return false
		}
		// export { x } (no rename, local) — x IS a reference to the local/global variable.
		if parent.Kind == ast.KindExportSpecifier && parent.AsExportSpecifier().PropertyName == nil {
			return false
		}
		return true
	}

	// Property name in destructuring: { x: y } — x is just a key.
	if parent.Kind == ast.KindBindingElement {
		be := parent.AsBindingElement()
		if be.PropertyName != nil && be.PropertyName == node {
			return true
		}
	}

	// Import source name: import { x as y } — x is the source module's export name.
	if parent.Kind == ast.KindImportSpecifier {
		importSpec := parent.AsImportSpecifier()
		if importSpec.PropertyName != nil && importSpec.PropertyName == node {
			return true
		}
	}

	// Label names: label: while(true) { break label; continue label; }
	if parent.Kind == ast.KindLabeledStatement ||
		parent.Kind == ast.KindBreakStatement ||
		parent.Kind == ast.KindContinueStatement {
		return true
	}

	return false
}

// isReExportSpecifier checks if an ExportSpecifier is part of a re-export
// declaration (export { ... } from 'mod').
func isReExportSpecifier(exportSpec *ast.Node) bool {
	// ExportSpecifier → NamedExports → ExportDeclaration
	namedExports := exportSpec.Parent
	if namedExports == nil {
		return false
	}
	exportDecl := namedExports.Parent
	if exportDecl == nil || exportDecl.Kind != ast.KindExportDeclaration {
		return false
	}
	return exportDecl.AsExportDeclaration().ModuleSpecifier != nil
}
