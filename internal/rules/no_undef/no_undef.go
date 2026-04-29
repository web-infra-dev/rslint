package no_undef

import (
	"fmt"
	"regexp"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-undef

type options struct {
	checkTypeof bool
}

func parseOptions(opts any) options {
	result := options{checkTypeof: false}
	optsMap := utils.GetOptionsMap(opts)
	if optsMap != nil {
		if v, ok := optsMap["typeof"].(bool); ok {
			result.checkTypeof = v
		}
	}
	return result
}

var globalCommentPattern = regexp.MustCompile(`/\*\s*global\s+([^*]+)\*/`)
var globalNamePattern = regexp.MustCompile(`(\w+)\s*(?::\s*\w+)?`)

// parseGlobalComments scans source text for /*global ...*/ block comments and
// returns the set of variable names declared in them.
func parseGlobalComments(sourceText string) map[string]bool {
	globals := make(map[string]bool)
	for _, match := range globalCommentPattern.FindAllStringSubmatch(sourceText, -1) {
		if len(match) > 1 {
			for _, nameMatch := range globalNamePattern.FindAllStringSubmatch(match[1], -1) {
				if len(nameMatch) > 1 {
					globals[nameMatch[1]] = true
				}
			}
		}
	}
	return globals
}

var NoUndefRule = rule.Rule{
	Name:             "no-undef",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		// Defense-in-depth: RequiresTypeInfo: true filters this rule out for
		// gap files / inferred-project files, but if a future caller bypasses
		// the filter we still want to no-op rather than nil-deref.
		if ctx.TypeChecker == nil {
			return rule.RuleListeners{}
		}

		// Parse /*global ...*/ comments to find declared globals
		declaredGlobals := parseGlobalComments(ctx.SourceFile.Text())

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				// Skip identifiers that are not value references
				if shouldSkip(node, opts.checkTypeof) {
					return
				}

				// For ShorthandPropertyAssignment ({x}), GetSymbolAtLocation
				// returns the property symbol which always exists. Use
				// GetShorthandAssignmentValueSymbol to check the value reference.
				if node.Parent.Kind == ast.KindShorthandPropertyAssignment {
					valueSym := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(node.Parent)
					if valueSym != nil {
						return
					}
				} else {
					sym := ctx.TypeChecker.GetSymbolAtLocation(node)
					if sym != nil {
						return
					}
				}

				name := node.Text()

				// Skip identifiers declared via /*global*/ comments
				if declaredGlobals[name] {
					return
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "undef",
					Description: fmt.Sprintf("'%s' is not defined.", name),
				})
			},
		}
	},
}

// shouldSkip returns true if the identifier should not be checked for being
// undefined. This includes property names, declaration names, labels, type
// positions, and typeof operands (when checkTypeof is false).
func shouldSkip(node *ast.Node, checkTypeof bool) bool {
	if node.Parent == nil {
		return true
	}

	parent := node.Parent

	// Skip declaration names (var x, function x, class x, import x, etc.)
	// Exception: ShorthandPropertyAssignment names ({x}) are declaration names
	// but also value references, so they must NOT be skipped.
	if ast.IsDeclarationName(node) && parent.Kind != ast.KindShorthandPropertyAssignment {
		return true
	}

	// Skip property names in member access (obj.prop -> skip "prop")
	if parent.Kind == ast.KindPropertyAccessExpression &&
		parent.AsPropertyAccessExpression().Name() == node {
		return true
	}

	// Skip property names in object literals ({key: value} -> skip "key")
	if parent.Kind == ast.KindPropertyAssignment && parent.Name() == node {
		return true
	}

	// Skip property names in binding patterns ({key: newName} destructuring -> skip "key")
	if parent.Kind == ast.KindBindingElement && parent.PropertyName() == node {
		return true
	}

	// Skip the original name in aliased imports (import { Original as Alias } -> skip "Original")
	// The propertyName of an ImportSpecifier is the module's export name, not a local reference.
	if parent.Kind == ast.KindImportSpecifier && parent.PropertyName() == node {
		return true
	}

	// Skip the original name in re-export aliases (export { Original as Alias } from 'module')
	// The propertyName of an ExportSpecifier in a re-export is the source module's export name.
	// Note: without `from`, export { X as Y } refers to local X, so only skip when moduleSpecifier exists.
	if parent.Kind == ast.KindExportSpecifier && parent.PropertyName() == node {
		if isReExport(parent) {
			return true
		}
	}

	// Skip label identifiers
	if parent.Kind == ast.KindLabeledStatement ||
		parent.Kind == ast.KindBreakStatement ||
		parent.Kind == ast.KindContinueStatement {
		return true
	}

	// Skip typeof operands (typeof x -> skip "x" unless checkTypeof is true).
	// Walk through ParenthesizedExpression nodes because typeof (x) and
	// typeof ((x)) are semantically identical to typeof x, but TypeScript's
	// AST inserts ParenthesizedExpression nodes unlike ESTree.
	if !checkTypeof && isTypeOfOperand(node) {
		return true
	}

	// Skip identifiers in type-only positions
	if isInTypeOnlyPosition(node) {
		return true
	}

	// Skip identifiers in qualified names (Namespace.Name in type contexts);
	// QualifiedName is only used in type positions, not value expressions.
	if parent.Kind == ast.KindQualifiedName {
		return true
	}

	// Skip enum member names
	if parent.Kind == ast.KindEnumMember && parent.Name() == node {
		return true
	}

	return false
}

// isInTypeOnlyPosition checks whether the identifier is inside a type-only
// context by walking up the parent chain and matching against explicit
// type-context node kinds.
//
// This follows the same pattern as isInTypeContext in no_unused_vars:
// explicit Kind matching avoids false positives from the overly broad
// ast.IsTypeNode() (which includes ExpressionWithTypeArguments).
//
// Note: AsExpression, TypeAssertionExpression, and SatisfiesExpression are
// intentionally excluded. Their expression operand is a value context; only
// the type annotation part is type-only. Since we walk up from the identifier,
// a value operand passes through these nodes without being misclassified.
func isInTypeOnlyPosition(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		switch current.Kind {
		// Core type constructs — any identifier under these is type-only
		case ast.KindTypeReference,
			ast.KindTypeAliasDeclaration,
			ast.KindInterfaceDeclaration,
			ast.KindTypeParameter,
			ast.KindTypeQuery,
			ast.KindTypeOperator,
			ast.KindIndexedAccessType,
			ast.KindConditionalType,
			ast.KindInferType,
			ast.KindTypeLiteral,
			ast.KindMappedType:
			return true

		// ExpressionWithTypeArguments wraps identifiers in heritage clauses.
		// It is type-only EXCEPT in class extends (which is a value position).
		case ast.KindExpressionWithTypeArguments:
			if !isClassExtendsClause(current) {
				return true
			}
			// class extends — value position, keep walking
		}
		current = current.Parent
	}
	return false
}

// isClassExtendsClause returns true if the ExpressionWithTypeArguments node
// is inside a class (not interface) extends clause.
// isTypeOfOperand checks if the identifier is the sole operand of a typeof
// expression, walking through any ParenthesizedExpression wrappers.
// typeof x      → Identifier.Parent = TypeOfExpression → true
// typeof (x)    → Identifier.Parent = ParenExpr.Parent = TypeOfExpression → true
// typeof (a+b)  → Identifier.Parent = BinaryExpression → false
func isTypeOfOperand(node *ast.Node) bool {
	current := node.Parent
	for current != nil && current.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	return current != nil && current.Kind == ast.KindTypeOfExpression
}

// isReExport checks if an ExportSpecifier belongs to a re-export statement
// (i.e., `export { ... } from 'module'`).
// Parent chain: ExportSpecifier → NamedExports → ExportDeclaration.
func isReExport(exportSpecifier *ast.Node) bool {
	namedExports := exportSpecifier.Parent
	if namedExports == nil {
		return false
	}
	exportDecl := namedExports.Parent
	if exportDecl == nil || exportDecl.Kind != ast.KindExportDeclaration {
		return false
	}
	return exportDecl.AsExportDeclaration().ModuleSpecifier != nil
}

func isClassExtendsClause(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindHeritageClause {
		return false
	}
	clause := parent.AsHeritageClause()
	if clause.Token != ast.KindExtendsKeyword {
		return false
	}
	grandparent := parent.Parent
	return grandparent != nil &&
		(grandparent.Kind == ast.KindClassDeclaration || grandparent.Kind == ast.KindClassExpression)
}
