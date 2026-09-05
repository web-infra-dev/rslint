package id_denylist

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

//go:embed id_denylist.schema.json
var schemaJSON []byte

// https://eslint.org/docs/latest/rules/id-denylist

func parseOptions(options []any) map[string]struct{} {
	denyList := make(map[string]struct{}, len(options))
	for _, option := range options {
		if name, ok := option.(string); ok {
			denyList[name] = struct{}{}
		}
	}
	return denyList
}

var IdDenylistRule = rule.Rule{
	Name:   "id-denylist",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		denyList := parseOptions(options)
		if len(denyList) == 0 {
			return rule.RuleListeners{}
		}
		check := func(node *ast.Node) {
			name := node.Text()
			isPrivate := node.Kind == ast.KindPrivateIdentifier
			if isPrivate {
				// A tsgo PrivateIdentifier carries the `#` in its text, while
				// the denied name is written without it.
				name = strings.TrimPrefix(name, "#")
			}
			if _, restricted := denyList[name]; !restricted {
				return
			}
			if !shouldCheck(ctx, node) {
				return
			}
			if isPrivate {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "restrictedPrivate",
					Description: fmt.Sprintf("Identifier '#%s' is restricted.", name),
				})
				return
			}
			ctx.ReportRange(identifierReportRange(ctx.SourceFile, node), rule.RuleMessage{
				Id:          "restricted",
				Description: fmt.Sprintf("Identifier '%s' is restricted.", name),
			})
		}

		return rule.RuleListeners{
			ast.KindIdentifier:        check,
			ast.KindPrivateIdentifier: check,
			ast.KindImportType: func(node *ast.Node) {
				name, nameRange, ok := importTypeAttributeKeywordRange(ctx.SourceFile, node)
				if !ok {
					return
				}
				if _, restricted := denyList[name]; !restricted {
					return
				}
				ctx.ReportRange(nameRange, rule.RuleMessage{
					Id:          "restricted",
					Description: fmt.Sprintf("Identifier '%s' is restricted.", name),
				})
			},
			ast.KindConstructor: func(node *ast.Node) {
				if _, restricted := denyList["constructor"]; !restricted {
					return
				}
				nameRange, ok := constructorNameRange(ctx.SourceFile, node)
				if !ok {
					return
				}
				ctx.ReportRange(nameRange, rule.RuleMessage{
					Id:          "restricted",
					Description: "Identifier 'constructor' is restricted.",
				})
			},
		}
	},
}

// importTypeAttributeKeywordRange restores the outer `with`/`assert` key of
// an import type's options object. typescript-eslint exposes that key as an
// Identifier, but tsgo stores only its token kind on ImportAttributes and
// emits no child node for the source token. Static and dynamic imports keep
// their existing import-attribute exclusion because this listener is limited
// to ImportType.
func importTypeAttributeKeywordRange(
	sourceFile *ast.SourceFile,
	node *ast.Node,
) (name string, textRange core.TextRange, ok bool) {
	if sourceFile == nil || node == nil || node.Kind != ast.KindImportType {
		return "", core.TextRange{}, false
	}
	importType := node.AsImportTypeNode()
	if importType == nil || importType.Argument == nil || importType.Attributes == nil {
		return "", core.TextRange{}, false
	}
	attributes := importType.Attributes.AsImportAttributes()
	if attributes == nil {
		return "", core.TextRange{}, false
	}
	switch attributes.Token {
	case ast.KindWithKeyword:
		name = "with"
	case ast.KindAssertKeyword:
		name = "assert"
	default:
		return "", core.TextRange{}, false
	}

	start, end := importType.Argument.End(), importType.Attributes.Pos()
	if start < 0 || end <= start || end > len(sourceFile.Text()) {
		return "", core.TextRange{}, false
	}
	s := scanner.GetScannerForSourceFile(sourceFile, start)
	for s.Token() != ast.KindEndOfFile && s.TokenStart() < end {
		if s.Token() == attributes.Token && s.TokenEnd() <= end {
			return name, core.NewTextRange(s.TokenStart(), s.TokenEnd()), true
		}
		s.Scan()
	}
	return "", core.TextRange{}, false
}

// identifierReportRange matches the Identifier range typescript-eslint gives
// core rules. A simple variable or non-rest parameter name owns its optional
// marker and direct type annotation in ESTree; binding-pattern children and
// rest identifiers do not. Initializers remain outside the reported range.
func identifierReportRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	textRange := utils.TrimNodeTextRange(sourceFile, node)
	parent := node.Parent
	if parent == nil || parent.Name() != node {
		return textRange
	}

	var last *ast.Node
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		last = parent.AsVariableDeclaration().Type
	case ast.KindParameter:
		parameter := parent.AsParameterDeclaration()
		if parameter.DotDotDotToken == nil {
			last = parameter.Type
			if last == nil {
				last = parameter.QuestionToken
			}
		}
	}
	if last != nil && !utils.IsInJSDocSyntax(last) && last.End() > textRange.End() {
		return textRange.WithEnd(last.End())
	}
	return textRange
}

// constructorNameRange narrows a constructor declaration's modifier-inclusive
// function head to its keyword, which is the identifier range ESLint reports.
func constructorNameRange(sourceFile *ast.SourceFile, node *ast.Node) (core.TextRange, bool) {
	if sourceFile == nil || node == nil {
		return core.TextRange{}, false
	}
	start := node.Pos()
	if modifiers := node.Modifiers(); modifiers != nil && len(modifiers.Nodes) > 0 {
		start = modifiers.Nodes[len(modifiers.Nodes)-1].End()
	}
	if start < 0 || start >= node.End() || node.End() > len(sourceFile.Text()) {
		return core.TextRange{}, false
	}
	s := scanner.GetScannerForSourceFile(sourceFile, start)
	if s.Token() != ast.KindConstructorKeyword || s.TokenEnd() > node.End() {
		return core.TextRange{}, false
	}
	return core.NewTextRange(s.TokenStart(), s.TokenEnd()), true
}

// shouldCheck determines whether a restricted name in this position is
// reported. It mirrors the upstream rule's shouldCheck.
func shouldCheck(ctx rule.RuleContext, node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	if utils.IsInJSDocSyntax(node) {
		return false
	}

	// Import attributes are defined by environments, so naming conventions
	// shouldn't apply to them.
	if isImportAttributeKey(node) {
		return false
	}

	// A JSX tag or attribute name is a JSXIdentifier upstream, a node type the
	// rule's Identifier listener never visits. tsgo spells both with plain
	// Identifier and PropertyAccessExpression nodes, so they are filtered here
	// instead.
	if isJsxName(node) {
		return false
	}

	// `import.meta`, `new.target`.
	if parent.Kind == ast.KindMetaProperty {
		return false
	}

	// Member access has special rules for checking property names. Read access
	// to a property with a restricted name is allowed, because it can be on an
	// object that user has no control over. Write access isn't allowed, because
	// it potentially creates a new property with a restricted name.
	if parent.Kind == ast.KindPropertyAccessExpression &&
		parent.AsPropertyAccessExpression().Name() == node {
		return isAssignmentTarget(parent)
	}

	// A callee or an argument of a call or a `new` is never checked. tsgo
	// spells `import(specifier, options)` as a CallExpression too, where
	// upstream has a distinct ImportExpression that this exclusion misses.
	callChild := utils.OutermostParenthesizedExpression(node)
	callParent := callChild.Parent
	if callParent == nil {
		return false
	}
	switch callParent.Kind {
	case ast.KindCallExpression:
		if !ast.IsImportCall(callParent) {
			return false
		}
	case ast.KindNewExpression:
		return false
	}

	return !isRenamedImport(node) &&
		!isPropertyNameInDestructuring(node) &&
		!isReferenceToGlobalVariable(ctx, node)
}

// isAssignmentTarget reports whether the given member access is written to,
// either on its own or as an element of a destructuring pattern. Upstream
// recognizes the pattern positions by the ArrayPattern, RestElement,
// ObjectPattern-property and AssignmentPattern node types; tsgo spells all of
// them with the expression syntax they share with a plain array or object
// literal, so each one additionally confirms that it sits in a destructuring
// target. Prefix/postfix `++`/`--` and a bare `for (x.y in z)` target are
// deliberately absent: upstream reports neither.
func isAssignmentTarget(node *ast.Node) bool {
	// ESTree retains a ChainExpression around optional property writes; tsgo
	// flattens it into the access node. Until parentheses terminate that chain,
	// the property is still a read rather than an assignment target upstream.
	if ast.IsOptionalChain(node) {
		return false
	}
	target := node
	for target.Parent != nil && target.Parent.Kind == ast.KindParenthesizedExpression {
		target = target.Parent
	}
	parent := target.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindBinaryExpression:
		// `foo.bar = 1` and `foo.bar += 1`, plus the `obj.bar = baz` default of
		// `({ a: obj.bar = baz } = qux)` that upstream spells AssignmentPattern.
		binary := parent.AsBinaryExpression()
		return ast.IsAssignmentOperator(binary.OperatorToken.Kind) && binary.Left == target
	case ast.KindArrayLiteralExpression, ast.KindSpreadElement, ast.KindSpreadAssignment:
		return isDestructuredFrom(target)
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Initializer == target &&
			isDestructuredFrom(target)
	}
	return false
}

// isDestructuredFrom reports whether node is an element of an array or object
// literal that an enclosing assignment reinterprets as a destructuring pattern.
// Upstream reads this off the node type alone, because ESTree gives a pattern
// its own kinds; tsgo reuses the expression syntax, so the enclosing assignment
// has to be found by walking out. The walk climbs only the nodes a pattern is
// built from and stops at the first parent that consumes the literal as a value
// — `[{ a: obj.b }.c] = d` destructures the result of a member access, not the
// literal the member access reads from. An update expression is not one of
// upstream's arms, so it does not count as an enclosing assignment here.
func isDestructuredFrom(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	assignment := ast.GetAssignmentTarget(node)
	if assignment == nil ||
		assignment.Kind == ast.KindPrefixUnaryExpression ||
		assignment.Kind == ast.KindPostfixUnaryExpression {
		return false
	}
	for current := node; current != nil && current != assignment; current = current.Parent {
		if current.Kind == ast.KindNonNullExpression {
			return false
		}
	}
	return true
}

// isRenamedImport reports whether node is an imported or re-exported name that
// the same specifier renames: the `a` of `import { a as b } from 'mod'` or of
// `export { a as b } from 'mod'`.
func isRenamedImport(node *ast.Node) bool {
	parent := node.Parent
	switch parent.Kind {
	case ast.KindImportSpecifier:
		return parent.AsImportSpecifier().PropertyName == node
	case ast.KindExportSpecifier:
		return parent.AsExportSpecifier().PropertyName == node && utils.IsReExportSpecifier(parent)
	}
	return false
}

// isPropertyNameInDestructuring reports whether node is the source key of an
// object destructuring pattern: the `a` of `const { a: b } = foo` or of
// `({ a: obj.b } = foo)`. A shorthand key is deliberately absent — upstream
// skips its key half but still checks the value half, which is the same node
// here (see isVariableReference).
func isPropertyNameInDestructuring(node *ast.Node) bool {
	parent := node.Parent
	switch parent.Kind {
	case ast.KindBindingElement:
		return parent.AsBindingElement().PropertyName == node
	case ast.KindPropertyAssignment:
		return parent.AsPropertyAssignment().Name() == node &&
			isDestructuredFrom(parent.Parent)
	}
	return false
}

// isReferenceToGlobalVariable reports whether node references a global variable
// that the source code does not declare. Those identifiers are allowed, as it
// is assumed that user has no control over the names of external global
// variables.
func isReferenceToGlobalVariable(ctx rule.RuleContext, node *ast.Node) bool {
	if ctx.Refs == nil || node.Kind != ast.KindIdentifier || node.Parent == nil {
		return false
	}
	// A non-aliased local export has one ts-go node for both the local
	// reference and the exported name. Upstream visits the exported-name role,
	// so a global local target must not hide that report.
	if node.Parent.Kind == ast.KindExportSpecifier &&
		node.Parent.PropertyName() == nil && !utils.IsReExportSpecifier(node.Parent) {
		return false
	}
	// import("pkg").Name names an imported module member rather than an
	// implicit global. Type arguments nested within the ImportType are not
	// covered by this helper and keep their normal reference role.
	if utils.IsImportTypeSyntax(node) {
		return false
	}

	referenceSpace := scope.ESLintReferenceSpace(node)
	if node.Parent.Kind != ast.KindNamespaceExportDeclaration && !isVariableReference(node) {
		return false
	}
	if node.Parent.Kind == ast.KindNamespaceExportDeclaration {
		referenceSpace = scope.ReferenceValue
	}
	name := node.Text()
	meaning := referenceSpace.DeclarationMeaning()
	if !ctx.Refs.IsGlobalNameReference(node, name, meaning) {
		return false
	}
	if ctx.Globals.Access(name).IsDeclared() {
		return true
	}
	// typescript-eslint seeds a default esnext catalog in the type namespace.
	// Keep the exemption space-sensitive: `Record` is external in a type, but
	// the same spelling in a value expression remains authored code.
	return referenceSpace.IncludesType() &&
		rule.IsDefaultTypeScriptTypeGlobal(name)
}

// isVariableReference reports whether node occupies a position that upstream's
// scope analysis records as a variable reference, the only kind of identifier
// that can resolve to a global.
func isVariableReference(node *ast.Node) bool {
	if node.Kind != ast.KindIdentifier {
		return false
	}
	if node.Parent.Kind == ast.KindExportSpecifier && !utils.IsReExportSpecifier(node.Parent) {
		propertyName := node.Parent.PropertyName()
		return propertyName == nil || propertyName == node
	}
	// `{ foo }` is one node here and two upstream: a property key and a value
	// reference. Only inside a destructuring pattern, where the key half is
	// skipped as a destructuring key, does the reference half decide the
	// outcome; in an object literal the key half is checked and resolves to
	// nothing.
	if node.Parent.Kind == ast.KindShorthandPropertyAssignment {
		return isDestructuredFrom(node.Parent.Parent)
	}
	return !utils.IsNonReferenceIdentifier(node)
}

// isImportAttributeKey reports whether node names an import attribute: the key
// of a static `with { type: 'json' }` clause, or any key of the options object
// of a dynamic `import(specifier, { with: { type: 'json' } })`, including keys
// nested under such a key.
func isImportAttributeKey(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	// Static import / re-export.
	if parent.Kind == ast.KindImportAttribute && parent.AsImportAttribute().Name() == node {
		// ImportType options reuse ImportAttribute syntax in ts-go, while
		// upstream exposes those keys as ordinary identifiers.
		return !utils.IsImportTypeSyntax(node)
	}
	// Nested ImportType option keys use object syntax too.
	if utils.IsImportTypeSyntax(node) {
		return false
	}

	// Dynamic import. A computed key is evaluated rather than named, so it is
	// checked like any other expression.
	if node.Kind == ast.KindComputedPropertyName {
		return false
	}
	switch parent.Kind {
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		if parent.Name() != node {
			return false
		}
	default:
		return false
	}
	object := parent.Parent
	if object == nil || object.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	// Upstream sees the object literal directly, because ESTree has no node for
	// a parenthesis: `import('x', ({ type: 'json' }))` passes the same options
	// object as the unparenthesized form.
	for object.Parent != nil && object.Parent.Kind == ast.KindParenthesizedExpression {
		object = object.Parent
	}
	owner := object.Parent
	if owner == nil {
		return false
	}
	if ast.IsImportCall(owner) && !utils.IsImportTypeSyntax(owner) {
		arguments := owner.AsCallExpression().Arguments
		if arguments != nil && len(arguments.Nodes) > 1 && arguments.Nodes[1] == object {
			return true
		}
	}
	// Nested key: `type` in `{ with: { type: 'json' } }` is reached through
	// `with`.
	if owner.Kind == ast.KindPropertyAssignment && owner.AsPropertyAssignment().Initializer == object {
		return isImportAttributeKey(owner.Name())
	}
	return false
}

// isJsxName reports whether node is part of a JSX tag name or attribute name.
func isJsxName(node *ast.Node) bool {
	current := node
	for parent := current.Parent; parent != nil; parent = current.Parent {
		switch parent.Kind {
		case ast.KindPropertyAccessExpression, ast.KindQualifiedName, ast.KindJsxNamespacedName:
			current = parent
		case ast.KindJsxOpeningElement:
			return parent.AsJsxOpeningElement().TagName == current
		case ast.KindJsxSelfClosingElement:
			return parent.AsJsxSelfClosingElement().TagName == current
		case ast.KindJsxClosingElement:
			return parent.AsJsxClosingElement().TagName == current
		case ast.KindJsxAttribute:
			return parent.AsJsxAttribute().Name() == current
		default:
			return false
		}
	}
	return false
}
