package prop_types

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed prop_types.schema.json
var schemaJSON []byte

type options struct {
	ignore, customValidators []string
	skipUndeclared           bool
}
type propType struct {
	children           map[string]propType
	union              []propType
	any, open, invalid bool
}
type use struct {
	node  *ast.Node
	names []string
}
type component struct {
	node          *ast.Node
	binding       *ast.Symbol
	declared      map[string]propType
	used          []use
	declaredBlock bool
	props         string
	destructured  map[string][]string
}

func parseOptions(raw []any) options {
	o := options{}
	if len(raw) == 0 {
		return o
	}
	m, _ := raw[0].(map[string]interface{})
	for _, key := range []string{"ignore", "customValidators"} {
		if a, ok := m[key].([]interface{}); ok {
			for _, v := range a {
				if s, ok := v.(string); ok {
					if key == "ignore" {
						o.ignore = append(o.ignore, s)
					} else {
						o.customValidators = append(o.customValidators, s)
					}
				}
			}
		}
	}
	o.skipUndeclared, _ = m["skipUndeclared"].(bool)
	return o
}

func keyName(n *ast.Node) string {
	if n == nil {
		return ""
	}
	if name, ok := utils.GetStaticPropertyName(n); ok {
		return name
	}
	return ""
}

func elementName(n *ast.Node) string {
	n = unwrap(n)
	if n == nil {
		return ""
	}
	switch n.Kind {
	case ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindNullKeyword,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindRegularExpressionLiteral:
		return keyName(n)
	}
	return ""
}

func unwrap(n *ast.Node) *ast.Node { return reactutil.SkipExpressionWrappers(n) }

func propMap(n *ast.Node, customValidators []string) (map[string]propType, bool) {
	return propMapSeen(n, customValidators, map[*ast.Node]bool{}, true)
}

func propMapSeen(n *ast.Node, customValidators []string, seen map[*ast.Node]bool, topLevel bool) (map[string]propType, bool) {
	n = unwrap(n)
	if n == nil {
		return nil, false
	}
	if n.Kind == ast.KindObjectLiteralExpression {
		out := map[string]propType{}
		for _, p := range n.AsObjectLiteralExpression().Properties.Nodes {
			if p == nil {
				continue
			}
			if p.Kind == ast.KindSpreadAssignment {
				// A shape spread does not make every nested key valid. The
				// top-level opaque-declaration fallback is handled by declared.
				if topLevel {
					out["__ANY_KEY__"] = propType{any: true}
				}
				continue
			}
			var name, value *ast.Node
			switch p.Kind {
			case ast.KindPropertyAssignment:
				pa := p.AsPropertyAssignment()
				name, value = pa.Name(), pa.Initializer
			case ast.KindShorthandPropertyAssignment:
				name, value = p.AsShorthandPropertyAssignment().Name(), p.AsShorthandPropertyAssignment().Name()
			case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
				name = p.Name()
				value = p
			default:
				continue
			}
			k := keyName(name)
			if k == "" {
				out["__ANY_KEY__"] = propType{any: true}
				continue
			}
			validator := validatorTypeSeen(value, customValidators, seen)
			if validator.invalid {
				out["__INVALID_VALIDATOR__"] = validator
				continue
			}
			out[k] = validator
		}
		return out, true
	}
	return nil, false
}

func validatorType(n *ast.Node, customValidators []string) propType {
	return validatorTypeSeen(n, customValidators, map[*ast.Node]bool{})
}

func validatorTypeSeen(n *ast.Node, customValidators []string, seen map[*ast.Node]bool) propType {
	n = unwrap(n)
	if n == nil {
		return propType{any: true}
	}
	if n.Kind == ast.KindIdentifier {
		if seen[n] {
			return propType{invalid: true}
		}
		seen[n] = true
		if initializer := reactutil.ResolveIdentifierInitializer(n, nil); initializer != nil && initializer != n {
			return validatorTypeSeen(initializer, customValidators, seen)
		}
		return propType{open: true}
	}
	if n.Kind == ast.KindPropertyAccessExpression && keyName(n.AsPropertyAccessExpression().Name()) == "isRequired" {
		return validatorTypeSeen(n.AsPropertyAccessExpression().Expression, customValidators, seen)
	}
	if n.Kind != ast.KindCallExpression {
		// Primitive and broad validators (for example PropTypes.object and
		// PropTypes.string) validate any property read beneath the prop.
		return propType{open: true}
	}
	call := n.AsCallExpression()
	callee := unwrap(call.Expression)
	name := ""
	if callee != nil && callee.Kind == ast.KindPropertyAccessExpression {
		name = keyName(callee.AsPropertyAccessExpression().Name())
		if object := unwrap(callee.AsPropertyAccessExpression().Expression); object != nil && object.Kind == ast.KindIdentifier && slices.Contains(customValidators, object.AsIdentifier().Text) {
			return propType{open: true}
		}
	}
	if name == "shape" || name == "exact" {
		if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
			if m, ok := propMapSeen(call.Arguments.Nodes[0], customValidators, seen, false); ok {
				if m["__INVALID_VALIDATOR__"].invalid {
					return propType{invalid: true}
				}
				return propType{children: m}
			}
		}
	}
	if name == "objectOf" || name == "arrayOf" {
		if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
			return propType{children: map[string]propType{"__ANY_KEY__": validatorTypeSeen(call.Arguments.Nodes[0], customValidators, seen)}}
		}
	}
	if name == "oneOfType" && call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
		argument := unwrap(call.Arguments.Nodes[0])
		if argument != nil && argument.Kind == ast.KindArrayLiteralExpression {
			var union []propType
			for _, candidate := range argument.AsArrayLiteralExpression().Elements.Nodes {
				if candidate != nil && candidate.Kind != ast.KindOmittedExpression {
					union = append(union, validatorTypeSeen(candidate, customValidators, seen))
				}
			}
			if len(union) > 0 {
				return propType{union: union}
			}
		}
	}
	// Calls whose structure this port does not model (instanceOf, oneOf, or a
	// dynamic shape argument) are broad validators upstream.
	return propType{open: true}
}

func declared(n *ast.Node, customValidators []string, propWrappers []reactutil.PropWrapperEntry) (map[string]propType, bool) {
	n = unwrap(n)
	if n != nil && n.Kind == ast.KindIdentifier {
		if initializer := reactutil.ResolveIdentifierInitializer(n, nil); initializer != nil && initializer != n {
			return declared(initializer, customValidators, propWrappers)
		}
	}
	if n != nil && n.Kind == ast.KindCallExpression && reactutil.IsPropWrapperCall(n, propWrappers) {
		call := n.AsCallExpression()
		if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
			return declared(call.Arguments.Nodes[0], customValidators, propWrappers)
		}
	}
	m, ok := propMap(n, customValidators)
	return m, ok
}

func setComponentDeclared(c *component, n *ast.Node, customValidators []string, propWrappers []reactutil.PropWrapperEntry) {
	if m, ok := declared(n, customValidators, propWrappers); ok {
		c.declared = m
	} else {
		c.declared = map[string]propType{"__ANY_KEY__": {any: true}}
	}
	c.declaredBlock = true
}

func visibleTypeAlias(use *ast.Node, candidates []*ast.Node) *ast.Node {
	var best *ast.Node
	bestDepth := -1
	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}
		for scope, depth := use, 0; scope != nil; scope, depth = scope.Parent, depth+1 {
			if scope.Kind != ast.KindBlock && scope.Kind != ast.KindSourceFile && scope.Kind != ast.KindModuleBlock {
				continue
			}
			for declarationScope := candidate.Parent; declarationScope != nil; declarationScope = declarationScope.Parent {
				if declarationScope == scope {
					if best == nil || depth < bestDepth {
						best, bestDepth = candidate, depth
					}
					break
				}
			}
		}
	}
	return best
}

func returnTypePropMap(query *ast.Node) (map[string]propType, bool) {
	if query == nil || query.Kind != ast.KindTypeQuery {
		return nil, false
	}
	expr := query.AsTypeQueryNode().ExprName
	if expr == nil || expr.Kind != ast.KindIdentifier {
		return nil, false
	}
	initializer := reactutil.ResolveIdentifierInitializer(expr, nil)
	if initializer == nil {
		return nil, false
	}
	initializer = unwrap(initializer)
	if initializer == nil || (initializer.Kind != ast.KindArrowFunction && initializer.Kind != ast.KindFunctionExpression) {
		return nil, false
	}
	body := unwrap(initializer.Body())
	if body == nil || body.Kind != ast.KindObjectLiteralExpression {
		return nil, false
	}
	return propMap(body, nil)
}

func declaredType(n *ast.Node, aliases map[string][]*ast.Node, aliasesBySymbol map[*ast.Symbol]*ast.Node, depth int) (map[string]propType, bool) {
	if n == nil || depth == 0 {
		return nil, false
	}
	var members *ast.TypeElementList
	switch n.Kind {
	case ast.KindTypeLiteral:
		members = n.AsTypeLiteralNode().Members
	case ast.KindInterfaceDeclaration:
		members = n.AsInterfaceDeclaration().Members
	case ast.KindTypeAliasDeclaration:
		return declaredType(n.AsTypeAliasDeclaration().Type, aliases, aliasesBySymbol, depth-1)
	case ast.KindTypeReference:
		ref := n.AsTypeReferenceNode()
		if ref.TypeName != nil && reactutil.EntityNameRightmost(ref.TypeName) != nil && reactutil.EntityNameRightmost(ref.TypeName).AsIdentifier().Text == "ReturnType" && ref.TypeArguments != nil && len(ref.TypeArguments.Nodes) == 1 {
			if declared, ok := returnTypePropMap(ref.TypeArguments.Nodes[0]); ok {
				return declared, true
			}
		}
		name := reactutil.EntityNameRightmost(n.AsTypeReferenceNode().TypeName)
		if name != nil {
			if symbol := name.Symbol(); symbol != nil {
				if declaration := aliasesBySymbol[symbol]; declaration != nil {
					return declaredType(declaration, aliases, aliasesBySymbol, depth-1)
				}
			}
			if declaration := visibleTypeAlias(n, aliases[name.AsIdentifier().Text]); declaration != nil {
				return declaredType(declaration, aliases, aliasesBySymbol, depth-1)
			}
		}
		// Imported or otherwise unresolved annotations are opaque upstream and
		// must not be treated as if no declaration were present.
		return map[string]propType{"__ANY_KEY__": {any: true}}, true
	case ast.KindIdentifier:
		if symbol := n.Symbol(); symbol != nil {
			if declaration := aliasesBySymbol[symbol]; declaration != nil {
				return declaredType(declaration, aliases, aliasesBySymbol, depth-1)
			}
		}
		if declaration := visibleTypeAlias(n, aliases[n.AsIdentifier().Text]); declaration != nil {
			return declaredType(declaration, aliases, aliasesBySymbol, depth-1)
		}
		return map[string]propType{"__ANY_KEY__": {any: true}}, true
	case ast.KindParenthesizedType:
		return declaredType(n.AsParenthesizedTypeNode().Type, aliases, aliasesBySymbol, depth-1)
	case ast.KindIntersectionType:
		declared := map[string]propType{}
		found := false
		for _, part := range n.AsIntersectionTypeNode().Types.Nodes {
			if members, ok := declaredType(part, aliases, aliasesBySymbol, depth-1); ok {
				for name, prop := range members {
					declared[name] = prop
				}
				found = true
			}
		}
		return declared, found
	case ast.KindUnionType:
		// A property present in any member of a union is a declared prop. This
		// is deliberately merged like an intersection: prop-types only decides
		// whether a read has a declaration, not which discriminated branch is
		// active at runtime.
		declared := map[string]propType{}
		found := false
		for _, part := range n.AsUnionTypeNode().Types.Nodes {
			if members, ok := declaredType(part, aliases, aliasesBySymbol, depth-1); ok {
				mergeDeclared(declared, members)
				found = true
			}
		}
		return declared, found
	default:
		return nil, false
	}
	if members == nil {
		return map[string]propType{}, true
	}
	declared := map[string]propType{}
	for _, member := range members.Nodes {
		if member == nil {
			continue
		}
		switch member.Kind {
		case ast.KindPropertySignature:
			property := member.AsPropertySignatureDeclaration()
			if property != nil {
				if name := keyName(property.Name()); name != "" {
					if property.Type != nil {
						declared[name] = propTypeFromType(property.Type.AsNode(), aliases, aliasesBySymbol, depth-1)
					} else {
						// TypeScript permits a property signature without an annotation.
						// It is still a declared prop, but its nested shape is unknown.
						declared[name] = propType{open: true}
					}
				}
			}
		case ast.KindMethodSignature:
			if name := keyName(member.AsMethodSignatureDeclaration().Name()); name != "" {
				declared[name] = propType{open: true}
			}
		}
	}
	if n.Kind == ast.KindInterfaceDeclaration {
		heritageClauses := n.AsInterfaceDeclaration().HeritageClauses
		if heritageClauses == nil {
			return declared, true
		}
		for _, clause := range heritageClauses.Nodes {
			if clause == nil || clause.Kind != ast.KindHeritageClause {
				continue
			}
			for _, heritage := range clause.AsHeritageClause().Types.Nodes {
				if heritage != nil && heritage.Kind == ast.KindExpressionWithTypeArguments {
					entry := heritage.AsExpressionWithTypeArguments()
					var base map[string]propType
					var ok bool
					if name := reactutil.EntityNameRightmost(entry.Expression); name != nil && name.AsIdentifier().Text == "ReturnType" && entry.TypeArguments != nil && len(entry.TypeArguments.Nodes) == 1 {
						base, ok = returnTypePropMap(entry.TypeArguments.Nodes[0])
					} else {
						base, ok = declaredType(entry.Expression, aliases, aliasesBySymbol, depth-1)
					}
					if ok {
						mergeDeclared(declared, base)
					} else {
						// Imported base interfaces such as React.HTMLAttributes are
						// intentionally opaque, so their additional props are valid.
						declared["__ANY_KEY__"] = propType{any: true}
					}
				}
			}
		}
	}
	return declared, true
}

func propTypeFromType(n *ast.Node, aliases map[string][]*ast.Node, aliasesBySymbol map[*ast.Symbol]*ast.Node, depth int) propType {
	if children, ok := declaredType(n, aliases, aliasesBySymbol, depth); ok {
		return propType{children: children}
	}
	return propType{open: true}
}

func classPropsType(class *ast.Node) *ast.Node {
	if class == nil {
		return nil
	}
	var clauses *ast.HeritageClauseList
	switch class.Kind {
	case ast.KindClassDeclaration:
		clauses = class.AsClassDeclaration().HeritageClauses
	case ast.KindClassExpression:
		clauses = class.AsClassExpression().HeritageClauses
	default:
		return nil
	}
	if clauses == nil {
		return nil
	}
	for _, clause := range clauses.Nodes {
		if clause == nil || clause.Kind != ast.KindHeritageClause {
			continue
		}
		types := clause.AsHeritageClause().Types
		if types == nil {
			continue
		}
		for _, typ := range types.Nodes {
			if typ == nil || typ.Kind != ast.KindExpressionWithTypeArguments {
				continue
			}
			arguments := typ.AsExpressionWithTypeArguments().TypeArguments
			if arguments != nil && len(arguments.Nodes) > 0 {
				return arguments.Nodes[0]
			}
		}
	}
	return nil
}

// extendsComponent accepts the pragma-independent `Foo.Component` form and
// JSDoc's explicit component marker, both of which eslint-plugin-react treats
// as React component classes.
func extendsComponent(class *ast.Node) bool {
	if class == nil {
		return false
	}
	for _, doc := range class.JSDoc(nil) {
		jd := doc.AsJSDoc()
		if jd == nil || jd.Tags == nil {
			continue
		}
		for _, tag := range jd.Tags.Nodes {
			if ast.IsJSDocAugmentsTag(tag) {
				aug := tag.AsJSDocAugmentsTag()
				if aug != nil && jsDocComponentName(aug.ClassName) {
					return true
				}
			}
		}
	}
	var clauses *ast.HeritageClauseList
	if class.Kind == ast.KindClassDeclaration {
		clauses = class.AsClassDeclaration().HeritageClauses
	} else if class.Kind == ast.KindClassExpression {
		clauses = class.AsClassExpression().HeritageClauses
	}
	if clauses == nil {
		return false
	}
	for _, clause := range clauses.Nodes {
		if clause == nil || clause.Kind != ast.KindHeritageClause {
			continue
		}
		for _, typ := range clause.AsHeritageClause().Types.Nodes {
			if typ != nil && typ.Kind == ast.KindExpressionWithTypeArguments && jsDocComponentName(typ.AsExpressionWithTypeArguments().Expression) {
				return true
			}
		}
	}
	return false
}

func jsDocComponentName(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindExpressionWithTypeArguments {
		return jsDocComponentName(node.AsExpressionWithTypeArguments().Expression)
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		return keyName(node.AsPropertyAccessExpression().Name()) == "Component" || keyName(node.AsPropertyAccessExpression().Name()) == "PureComponent"
	}
	if node.Kind == ast.KindIdentifier {
		return node.AsIdentifier().Text == "Component" || node.AsIdentifier().Text == "PureComponent"
	}
	return false
}

func componentVariableType(n *ast.Node) *ast.Node {
	for current := n; current != nil && current.Parent != nil; {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression, ast.KindTypeAssertionExpression, ast.KindCallExpression:
			current = parent
			continue
		case ast.KindVariableDeclaration:
			if parent.AsVariableDeclaration().Initializer == current && parent.AsVariableDeclaration().Type != nil {
				return parent.AsVariableDeclaration().Type.AsNode()
			}
		}
		return nil
	}
	return nil
}

func reactComponentTypeArgument(typeNode *ast.Node) *ast.Node {
	if typeNode == nil || typeNode.Kind != ast.KindTypeReference {
		return nil
	}
	name := reactutil.EntityNameRightmost(typeNode.AsTypeReferenceNode().TypeName)
	if name == nil {
		return nil
	}
	arguments := typeNode.AsTypeReferenceNode().TypeArguments
	if arguments == nil || len(arguments.Nodes) == 0 {
		return nil
	}
	typeName := importedReactTypeName(name)
	// An unqualified component type is React-specific only when it is a
	// named import.  A project-local `FC<Props>` must not suppress this rule.
	if typeName == "" && typeNode.AsTypeReferenceNode().TypeName.Kind == ast.KindIdentifier {
		return nil
	}
	if typeName == "" {
		qualified := typeNode.AsTypeReferenceNode().TypeName
		if qualified == nil || qualified.Kind != ast.KindQualifiedName || qualified.AsQualifiedName().Left == nil || qualified.AsQualifiedName().Left.Kind != ast.KindIdentifier {
			return nil
		}
		qualifier := qualified.AsQualifiedName().Left.AsIdentifier().Text
		if qualifier != "React" {
			root := ast.GetSourceFileOfNode(name)
			if root == nil || !strings.Contains(root.Text(), "* as "+qualifier) || (!strings.Contains(root.Text(), `"react"`) && !strings.Contains(root.Text(), `'react'`)) {
				return nil
			}
		}
		typeName = name.AsIdentifier().Text
	}
	if typeName == "ForwardRefRenderFunction" && len(arguments.Nodes) >= 2 {
		return arguments.Nodes[1]
	}
	if !slices.Contains([]string{"FC", "FunctionComponent", "SFC", "StatelessComponent", "ComponentType", "VFC", "VoidFunctionComponent"}, typeName) {
		return nil
	}
	return arguments.Nodes[0]
}

// importedReactTypeName returns the exported name for a locally aliased React
// type (for example `import { FC as X } from "react"`).
func importedReactTypeName(name *ast.Node) string {
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	localName := name.AsIdentifier().Text
	root := ast.GetSourceFileOfNode(name)
	if root == nil {
		return ""
	}
	var imported string
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n == nil || imported != "" {
			return false
		}
		if n.Kind == ast.KindImportSpecifier {
			specifier := n.AsImportSpecifier()
			if specifier.Name() != nil && specifier.Name().Kind == ast.KindIdentifier && specifier.Name().AsIdentifier().Text == localName {
				for parent := n.Parent; parent != nil; parent = parent.Parent {
					if parent.Kind == ast.KindImportDeclaration {
						module := parent.AsImportDeclaration().ModuleSpecifier
						if module != nil && module.Kind == ast.KindStringLiteral && module.Text() == "react" {
							imported = localName
							if specifier.PropertyName != nil && specifier.PropertyName.Kind == ast.KindIdentifier {
								imported = specifier.PropertyName.AsIdentifier().Text
							}
						}
						break
					}
				}
			}
		}
		n.ForEachChild(visit)
		return false
	}
	root.Node.ForEachChild(visit)
	return imported
}

func forwardRefPropsType(n *ast.Node) *ast.Node {
	for current := n; current != nil && current.Parent != nil; current = current.Parent {
		parent := current.Parent
		if parent.Kind == ast.KindCallExpression {
			call := parent.AsCallExpression()
			if call.Arguments == nil || len(call.Arguments.Nodes) == 0 || unwrap(call.Arguments.Nodes[0]) != unwrap(current) {
				return nil
			}
			callee := unwrap(call.Expression)
			isForwardRef := callee != nil && callee.Kind == ast.KindIdentifier && callee.AsIdentifier().Text == "forwardRef"
			if callee != nil && callee.Kind == ast.KindPropertyAccessExpression {
				isForwardRef = keyName(callee.AsPropertyAccessExpression().Name()) == "forwardRef"
			}
			if !isForwardRef {
				return nil
			}
			arguments := call.TypeArguments
			if arguments != nil && len(arguments.Nodes) >= 2 {
				return arguments.Nodes[1]
			}
			return nil
		}
		if parent.Kind != ast.KindParenthesizedExpression && parent.Kind != ast.KindAsExpression && parent.Kind != ast.KindSatisfiesExpression && parent.Kind != ast.KindNonNullExpression && parent.Kind != ast.KindTypeAssertionExpression {
			return nil
		}
	}
	return nil
}

func componentName(n *ast.Node) string {
	if name := reactutil.BindingIdentifierName(n); name != "" {
		return name
	}
	var suffix []string
	for current := n; current != nil && current.Parent != nil; {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression,
			ast.KindAsExpression,
			ast.KindSatisfiesExpression,
			ast.KindNonNullExpression,
			ast.KindTypeAssertionExpression,
			ast.KindCallExpression:
			current = parent
			continue
		case ast.KindVariableDeclaration:
			if parent.AsVariableDeclaration().Initializer == current {
				name := parent.AsVariableDeclaration().Name()
				if name != nil && name.Kind == ast.KindIdentifier {
					return strings.Join(append([]string{name.AsIdentifier().Text}, suffix...), ".")
				}
			}
		case ast.KindBinaryExpression:
			assignment := parent.AsBinaryExpression()
			if assignment.OperatorToken != nil && assignment.OperatorToken.Kind == ast.KindEqualsToken && assignment.Right == current {
				if name := componentReference(assignment.Left); name != "" {
					return strings.Join(append([]string{name}, suffix...), ".")
				}
			}
		case ast.KindPropertyAssignment:
			assignment := parent.AsPropertyAssignment()
			if assignment.Initializer == current {
				if name := keyName(assignment.Name()); name != "" {
					suffix = append([]string{name}, suffix...)
					current = parent
					continue
				}
			}
		case ast.KindObjectLiteralExpression:
			current = parent
			continue
		}
		return ""
	}
	return ""
}

func componentBinding(n *ast.Node) *ast.Symbol {
	if n == nil {
		return nil
	}
	if n.Kind == ast.KindFunctionDeclaration || n.Kind == ast.KindClassDeclaration {
		return n.Symbol()
	}
	for current := n; current != nil && current.Parent != nil; {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression, ast.KindTypeAssertionExpression, ast.KindCallExpression:
			current = parent
		case ast.KindVariableDeclaration:
			if parent.AsVariableDeclaration().Initializer == current {
				return parent.Symbol()
			}
			return nil
		default:
			return nil
		}
	}
	return nil
}

func setDeclared(m map[string]propType, names []string, value propType) {
	if len(names) == 0 {
		return
	}
	if len(names) == 1 {
		m[names[0]] = value
		return
	}
	p := m[names[0]]
	if p.children == nil {
		p.children = map[string]propType{}
	}
	setDeclared(p.children, names[1:], value)
	m[names[0]] = p
}

func mergeDeclared(dst, src map[string]propType) {
	for name, value := range src {
		if old, ok := dst[name]; ok && old.children != nil && value.children != nil {
			mergeDeclared(old.children, value.children)
			value.children = old.children
		}
		dst[name] = value
	}
}

func addDestructured(c *component, pattern *ast.Node, prefix []string, includeIntermediate ...bool) {
	if c == nil || pattern == nil || pattern.Kind != ast.KindObjectBindingPattern {
		return
	}
	recordIntermediate := len(includeIntermediate) > 0 && includeIntermediate[0]
	for _, e := range pattern.AsBindingPattern().Elements.Nodes {
		if e == nil || e.Kind != ast.KindBindingElement {
			continue
		}
		be := e.AsBindingElement()
		if be.DotDotDotToken != nil {
			continue
		}
		// A dynamic computed key has no statically knowable prop name. ESLint
		// does not turn its local binding into a validation requirement.
		if be.PropertyName != nil && be.PropertyName.Kind == ast.KindComputedPropertyName {
			continue
		}
		key := keyName(be.PropertyName)
		if key == "" {
			key = keyName(be.Name())
		}
		if key == "" || be.Name() == nil {
			continue
		}
		path := append(append([]string{}, prefix...), key)
		// Every direct binding is a prop use. Function-parameter nested patterns
		// also use their intermediate object; local destructuring does not.
		report := be.Name()
		if be.PropertyName != nil {
			report = be.PropertyName
		}
		if be.Name().Kind != ast.KindObjectBindingPattern || recordIntermediate {
			appendUse(c, report, path)
		}
		if be.Name().Kind == ast.KindIdentifier {
			c.destructured[be.Name().AsIdentifier().Text] = path
		} else if be.Name().Kind == ast.KindObjectBindingPattern {
			addDestructured(c, be.Name(), path, recordIntermediate)
		}
	}
}

func addThisPropsDestructured(c *component, pattern *ast.Node) bool {
	if c == nil || pattern == nil || pattern.Kind != ast.KindObjectBindingPattern {
		return false
	}
	for _, e := range pattern.AsBindingPattern().Elements.Nodes {
		if e == nil || e.Kind != ast.KindBindingElement {
			continue
		}
		be := e.AsBindingElement()
		key := keyName(be.PropertyName)
		if key == "" {
			key = keyName(be.Name())
		}
		if key != "props" || be.Name() == nil {
			continue
		}
		switch be.Name().Kind {
		case ast.KindObjectBindingPattern:
			addDestructured(c, be.Name(), nil)
		case ast.KindIdentifier:
			c.destructured[be.Name().AsIdentifier().Text] = nil
		}
		return true
	}
	return false
}

func propsPath(root *ast.Node, names []string, c *component) ([]string, bool) {
	root = unwrap(root)
	if root == nil || c == nil {
		return nil, false
	}
	if root.Kind == ast.KindThisKeyword {
		if len(names) > 0 && names[0] == "props" {
			return names[1:], true
		}
		return nil, false
	}
	if root.Kind != ast.KindIdentifier {
		return nil, false
	}
	if root.AsIdentifier().Text == c.props || isClassPropsParameter(root, c) {
		return names, true
	}
	if prefix, ok := c.destructured[root.AsIdentifier().Text]; ok {
		return append(append([]string{}, prefix...), names...), true
	}
	return nil, false
}

func isClassPropsParameter(ident *ast.Node, c *component) bool {
	if ident == nil || ident.Kind != ast.KindIdentifier || c == nil || (c.node.Kind != ast.KindClassDeclaration && c.node.Kind != ast.KindClassExpression) {
		return false
	}
	for current := ident.Parent; current != nil && current != c.node; current = current.Parent {
		switch current.Kind {
		case ast.KindConstructor:
			return functionFirstParameterName(current) == ident.AsIdentifier().Text
		case ast.KindMethodDeclaration:
			name := keyName(current.Name())
			switch name {
			case "componentWillReceiveProps", "shouldComponentUpdate", "componentWillUpdate", "componentDidUpdate",
				"getDerivedStateFromProps", "getSnapshotBeforeUpdate", "UNSAFE_componentWillReceiveProps", "UNSAFE_componentWillUpdate":
				return functionFirstParameterName(current) == ident.AsIdentifier().Text
			default:
				return false
			}
		case ast.KindArrowFunction, ast.KindFunctionExpression:
			return isSetStatePropsParameter(current, ident)
		case ast.KindFunctionDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			return false
		}
	}
	return false
}

func isSetStatePropsParameter(fn, ident *ast.Node) bool {
	if fn == nil || ident == nil || ident.Kind != ast.KindIdentifier {
		return false
	}
	params := reactutil.FunctionParameters(fn)
	if len(params) < 2 || params[1] == nil || params[1].Kind != ast.KindParameter || functionParameterName(params[1]) != ident.AsIdentifier().Text {
		return false
	}
	return isSetStateCallback(fn)
}

func isSetStateCallback(fn *ast.Node) bool {
	call := reactutil.SkipExpressionWrappersUp(fn)
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	invocation := call.AsCallExpression()
	if invocation.Arguments == nil || len(invocation.Arguments.Nodes) == 0 || unwrap(invocation.Arguments.Nodes[0]) != fn {
		return false
	}
	callee := unwrap(invocation.Expression)
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression || keyName(callee.AsPropertyAccessExpression().Name()) != "setState" {
		return false
	}
	receiver := unwrap(callee.AsPropertyAccessExpression().Expression)
	return receiver != nil && receiver.Kind == ast.KindThisKeyword
}

func functionFirstParameterName(fn *ast.Node) string {
	params := reactutil.FunctionParameters(fn)
	if len(params) == 0 || params[0] == nil || params[0].Kind != ast.KindParameter {
		return ""
	}
	return functionParameterName(params[0])
}

func functionParameterName(param *ast.Node) string {
	name := param.AsParameterDeclaration().Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.AsIdentifier().Text
}

func isGeneratorFunction(n *ast.Node) bool {
	if n == nil {
		return false
	}
	switch n.Kind {
	case ast.KindFunctionDeclaration:
		return n.AsFunctionDeclaration().AsteriskToken != nil
	case ast.KindFunctionExpression:
		return n.AsFunctionExpression().AsteriskToken != nil
	case ast.KindMethodDeclaration:
		return n.AsMethodDeclaration().AsteriskToken != nil
	}
	return false
}

func getterReturn(body *ast.Node) *ast.Node {
	if body == nil || body.Kind != ast.KindBlock {
		return nil
	}
	stmts := body.AsBlock().Statements
	if stmts == nil {
		return nil
	}
	for i := len(stmts.Nodes) - 1; i >= 0; i-- {
		if stmt := stmts.Nodes[i]; stmt != nil && stmt.Kind == ast.KindReturnStatement {
			return stmt.AsReturnStatement().Expression
		}
	}
	return nil
}

func componentFor(node *ast.Node, comps []*component) *component {
	var best *component
	for p := node; p != nil; p = p.Parent {
		for _, c := range comps {
			if c.node == p {
				best = c
				break
			}
		}
		if best != nil {
			return best
		}
	}
	return nil
}

func appendUse(c *component, node *ast.Node, names []string) {
	if c != nil && len(names) > 0 {
		c.used = append(c.used, use{node: node, names: names})
	}
}

// memberReportNode matches the ESTree MemberExpression report range used by
// eslint-plugin-react: diagnostics attach to the terminal property, rather
// than to the complete member chain.
func memberReportNode(n *ast.Node) *ast.Node {
	n = unwrap(n)
	if n == nil {
		return n
	}
	switch n.Kind {
	case ast.KindPropertyAccessExpression:
		return n.AsPropertyAccessExpression().Name()
	case ast.KindElementAccessExpression:
		return n.AsElementAccessExpression().ArgumentExpression
	default:
		return n
	}
}

func memberNames(n *ast.Node) (*ast.Node, []string, bool) {
	var names []string
	cur := unwrap(n)
	var report *ast.Node
	for cur != nil {
		switch cur.Kind {
		case ast.KindPropertyAccessExpression:
			pa := cur.AsPropertyAccessExpression()
			k := keyName(pa.Name())
			if k == "" {
				return nil, nil, false
			}
			names = append([]string{k}, names...)
			report = pa.Name()
			cur = unwrap(pa.Expression)
		case ast.KindElementAccessExpression:
			ea := cur.AsElementAccessExpression()
			k := elementName(ea.ArgumentExpression)
			if k == "" {
				if argument := unwrap(ea.ArgumentExpression); argument != nil && argument.Kind == ast.KindBigIntLiteral {
					return nil, nil, false
				}
				k = "__COMPUTED_PROP__"
			}
			names = append([]string{k}, names...)
			report = ea.ArgumentExpression
			cur = unwrap(ea.Expression)
		default:
			return cur, names, true
		}
	}
	return report, names, false
}

func componentReference(n *ast.Node) string {
	root, names, ok := memberNames(n)
	if !ok || root == nil || root.Kind != ast.KindIdentifier {
		return ""
	}
	parts := append([]string{root.AsIdentifier().Text}, names...)
	return strings.Join(parts, ".")
}

func isAssignmentTarget(n *ast.Node) bool {
	for n != nil && n.Parent != nil {
		parent := n.Parent
		switch parent.Kind {
		case ast.KindPropertyAccessExpression:
			if parent.AsPropertyAccessExpression().Expression == n {
				n = parent
				continue
			}
		case ast.KindElementAccessExpression:
			if parent.AsElementAccessExpression().Expression == n {
				n = parent
				continue
			}
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression, ast.KindTypeAssertionExpression:
			n = parent
			continue
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			return binary.OperatorToken != nil && ast.IsAssignmentOperator(binary.OperatorToken.Kind) && binary.Left == n
		}
		return false
	}
	return false
}

func rootMatches(n *ast.Node, c *component) bool {
	n = unwrap(n)
	if n == nil {
		return false
	}
	if n.Kind == ast.KindThisKeyword {
		return true
	}
	if n.Kind != ast.KindIdentifier {
		return false
	}
	if n.AsIdentifier().Text == c.props || isClassPropsParameter(n, c) {
		return true
	}
	_, ok := c.destructured[n.AsIdentifier().Text]
	return ok
}

func isDeclared(m map[string]propType, names []string) bool {
	if len(names) == 0 {
		return true
	}
	p, ok := m[names[0]]
	if !ok {
		p, ok = m["__ANY_KEY__"]
	}
	if !ok {
		return false
	}
	return propTypeDeclares(p, names[1:])
}

func propTypeDeclares(p propType, remaining []string) bool {
	if p.invalid {
		return false
	}
	if len(remaining) == 0 || p.any || p.open {
		return true
	}
	for _, candidate := range p.union {
		if propTypeDeclares(candidate, remaining) {
			return true
		}
	}
	if p.children == nil {
		return false
	}
	return isDeclared(p.children, remaining)
}

func reportName(names []string) string {
	return strings.ReplaceAll(strings.Join(names, "."), ".__COMPUTED_PROP__", "[]")
}

func isObjectPrototypeProperty(name string) bool {
	switch name {
	case "__defineGetter__", "__defineSetter__", "hasOwnProperty", "__lookupGetter__", "__lookupSetter__",
		"isPrototypeOf", "propertyIsEnumerable", "toString", "valueOf", "__proto__", "toLocaleString", "constructor":
		return true
	}
	return false
}

var PropTypesRule = rule.Rule{
	Name: "react/prop-types", Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, raw []any) rule.RuleListeners {
		o := parseOptions(raw)
		pragma := reactutil.GetReactPragma(ctx.Settings)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)
		wrappers := reactutil.GetComponentWrapperFunctions(ctx.Settings, pragma)
		propWrappers := reactutil.GetPropWrapperFunctions(ctx.Settings)
		typeAliases := map[string][]*ast.Node{}
		typeAliasesBySymbol := map[*ast.Symbol]*ast.Node{}
		var collectTypeAliases ast.Visitor
		collectTypeAliases = func(n *ast.Node) bool {
			if n == nil {
				return false
			}
			switch n.Kind {
			case ast.KindTypeAliasDeclaration:
				if name := n.AsTypeAliasDeclaration().Name(); name != nil && name.Kind == ast.KindIdentifier {
					typeAliases[name.AsIdentifier().Text] = append(typeAliases[name.AsIdentifier().Text], n)
					if symbol := name.Symbol(); symbol != nil {
						typeAliasesBySymbol[symbol] = n
					}
					if symbol := n.Symbol(); symbol != nil {
						typeAliasesBySymbol[symbol] = n
					}
				}
			case ast.KindInterfaceDeclaration:
				if name := n.AsInterfaceDeclaration().Name(); name != nil && name.Kind == ast.KindIdentifier {
					typeAliases[name.AsIdentifier().Text] = append(typeAliases[name.AsIdentifier().Text], n)
					if symbol := name.Symbol(); symbol != nil {
						typeAliasesBySymbol[symbol] = n
					}
					if symbol := n.Symbol(); symbol != nil {
						typeAliasesBySymbol[symbol] = n
					}
				}
			}
			n.ForEachChild(collectTypeAliases)
			return false
		}
		ctx.SourceFile.Node.ForEachChild(collectTypeAliases)
		var comps []*component
		byName := map[string]*component{}
		byBinding := map[*ast.Symbol]*component{}
		var propTypeAssignments []*ast.Node
		applyPropTypeAssignment := func(assignment *ast.Node) {
			binary := assignment.AsBinaryExpression()
			root, names, ok := memberNames(binary.Left)
			if !ok || root == nil || root.Kind != ast.KindIdentifier {
				return
			}
			propTypesIndex := slices.Index(names, "propTypes")
			if propTypesIndex < 0 {
				return
			}
			componentParts := append([]string{root.AsIdentifier().Text}, names[:propTypesIndex]...)
			c := byName[strings.Join(componentParts, ".")]
			if propTypesIndex == 0 {
				symbol := root.Symbol()
				if symbol == nil && ctx.TypeChecker != nil {
					symbol = ctx.TypeChecker.GetSymbolAtLocation(root)
				}
				if bySymbol := byBinding[symbol]; bySymbol != nil {
					c = bySymbol
				}
			}
			if c == nil {
				return
			}
			propNames := names[propTypesIndex+1:]
			if len(propNames) == 0 {
				if m, ok := declared(binary.Right, o.customValidators, propWrappers); ok {
					mergeDeclared(c.declared, m)
				} else {
					c.declared = map[string]propType{"__ANY_KEY__": {any: true}}
				}
			} else {
				setDeclared(c.declared, propNames, validatorType(binary.Right, o.customValidators))
			}
			c.declaredBlock = true
		}
		var walk ast.Visitor
		walk = func(n *ast.Node) bool {
			if n == nil {
				return false
			}
			switch n.Kind {
			case ast.KindClassDeclaration, ast.KindClassExpression:
				if reactutil.ExtendsReactComponent(n, pragma) || extendsComponent(n) {
					c := &component{node: n, declared: map[string]propType{}, destructured: map[string][]string{}}
					comps = append(comps, c)
					c.binding = componentBinding(n)
					if c.binding != nil {
						byBinding[c.binding] = c
					}
					if declared, ok := declaredType(classPropsType(n), typeAliases, typeAliasesBySymbol, 8); ok {
						c.declared = declared
						c.declaredBlock = true
					}
					if name := componentName(n); name != "" {
						byName[name] = c
					}
				}
			case ast.KindObjectLiteralExpression:
				if reactutil.IsCreateReactClassObjectArg(n, pragma, createClass) {
					c := &component{node: n, declared: map[string]propType{}, destructured: map[string][]string{}}
					comps = append(comps, c)
					c.binding = componentBinding(n)
					if c.binding != nil {
						byBinding[c.binding] = c
					}
					if name := componentName(n); name != "" {
						byName[name] = c
					}
				}
			case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration:
				if !isGeneratorFunction(n) && reactutil.IsStatelessReactComponentWithWrappers(n, pragma, ctx.TypeChecker, wrappers) {
					c := &component{node: n, declared: map[string]propType{}, destructured: map[string][]string{}}
					comps = append(comps, c)
					c.binding = componentBinding(n)
					if c.binding != nil {
						byBinding[c.binding] = c
					}
					if name := componentName(n); name != "" {
						byName[name] = c
					}
					ps := n.Parameters()
					if len(ps) > 0 && ps[0].Kind == ast.KindParameter {
						parameter := ps[0].AsParameterDeclaration()
						name := parameter.Name()
						if name != nil && name.Kind == ast.KindIdentifier {
							c.props = name.AsIdentifier().Text
						} else if name != nil && name.Kind == ast.KindObjectBindingPattern {
							addDestructured(c, name, nil, true)
						}
						if parameter.Type != nil {
							if declared, ok := declaredType(parameter.Type.AsNode(), typeAliases, typeAliasesBySymbol, 8); ok {
								c.declared = declared
								c.declaredBlock = true
							}
						} else if declared, ok := declaredType(reactComponentTypeArgument(componentVariableType(n)), typeAliases, typeAliasesBySymbol, 8); ok {
							c.declared = declared
							c.declaredBlock = true
						} else if declared, ok := declaredType(forwardRefPropsType(n), typeAliases, typeAliasesBySymbol, 8); ok {
							c.declared = declared
							c.declaredBlock = true
						}
					}
				}
				if n.Kind == ast.KindMethodDeclaration {
					if c := componentFor(n.Parent, comps); c != nil {
						name := keyName(n.AsMethodDeclaration().Name())
						switch name {
						case "componentWillReceiveProps", "shouldComponentUpdate", "componentWillUpdate", "componentDidUpdate", "getDerivedStateFromProps", "getSnapshotBeforeUpdate", "UNSAFE_componentWillReceiveProps", "UNSAFE_componentWillUpdate":
							params := reactutil.FunctionParameters(n)
							if len(params) > 0 && params[0] != nil && params[0].Kind == ast.KindParameter {
								if pattern := params[0].AsParameterDeclaration().Name(); pattern != nil && pattern.Kind == ast.KindObjectBindingPattern {
									addDestructured(c, pattern, nil)
								}
							}
						}
					}
				}
				// setState's second callback parameter is the current props object.
				// It also appears in arrow functions stored in class fields, which
				// are not MethodDeclarations in TypeScript's AST.
				if c := componentFor(n, comps); c != nil {
					params := reactutil.FunctionParameters(n)
					if len(params) >= 2 && params[1] != nil && params[1].Kind == ast.KindParameter && isSetStateCallback(n) {
						if pattern := params[1].AsParameterDeclaration().Name(); pattern != nil && pattern.Kind == ast.KindObjectBindingPattern {
							addDestructured(c, pattern, nil)
						}
					}
				}
			case ast.KindPropertyAssignment:
				pa := n.AsPropertyAssignment()
				if keyName(pa.Name()) == "propTypes" {
					if c := componentFor(n.Parent, comps); c != nil {
						setComponentDeclared(c, pa.Initializer, o.customValidators, propWrappers)
					}
				}
			case ast.KindPropertyDeclaration:
				pd := n.AsPropertyDeclaration()
				if keyName(pd.Name()) == "props" && !ast.HasSyntacticModifier(n, ast.ModifierFlagsStatic) {
					if c := componentFor(n.Parent, comps); c != nil && pd.Type != nil {
						if declared, ok := declaredType(pd.Type.AsNode(), typeAliases, typeAliasesBySymbol, 8); ok {
							c.declared = declared
							c.declaredBlock = true
						}
					}
				}
				if keyName(pd.Name()) == "propTypes" {
					if c := componentFor(n.Parent, comps); c != nil {
						setComponentDeclared(c, pd.Initializer, o.customValidators, propWrappers)
					}
				}
			case ast.KindGetAccessor:
				ga := n.AsGetAccessorDeclaration()
				if ast.HasSyntacticModifier(n, ast.ModifierFlagsStatic) && keyName(ga.Name()) == "propTypes" {
					if c := componentFor(n.Parent, comps); c != nil {
						if m, ok := declared(getterReturn(ga.Body), o.customValidators, propWrappers); ok {
							c.declared = m
							c.declaredBlock = true
						}
					}
				}
			case ast.KindVariableDeclaration:
				vd := n.AsVariableDeclaration()
				if vd.Name() != nil {
					if root, names, ok := memberNames(vd.Initializer); ok && root != nil {
						if c := componentFor(n, comps); c != nil {
							if root.Kind == ast.KindThisKeyword && len(names) == 0 && vd.Name().Kind == ast.KindObjectBindingPattern {
								addThisPropsDestructured(c, vd.Name())
							}
							if path, ok := propsPath(root, names, c); ok {
								switch vd.Name().Kind {
								case ast.KindObjectBindingPattern:
									addDestructured(c, vd.Name(), path, c.node.Kind == ast.KindClassDeclaration || c.node.Kind == ast.KindClassExpression)
								case ast.KindIdentifier:
									c.destructured[vd.Name().AsIdentifier().Text] = path
								}
							}
						}
					}
				}
			case ast.KindBinaryExpression:
				b := n.AsBinaryExpression()
				if b.OperatorToken != nil && b.OperatorToken.Kind == ast.KindEqualsToken && b.Left != nil && b.Left.Kind == ast.KindPropertyAccessExpression {
					root, names, ok := memberNames(b.Left)
					if ok && root != nil && root.Kind == ast.KindIdentifier {
						propTypesIndex := slices.Index(names, "propTypes")
						if propTypesIndex < 0 {
							break
						}
						propTypeAssignments = append(propTypeAssignments, n)
						applyPropTypeAssignment(n)
					}
				}
			case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
				if isAssignmentTarget(n) {
					break
				}
				root, names, ok := memberNames(n)
				if !ok || root == nil {
					break
				}
				if len(names) == 0 {
					break
				}
				// A nested component may capture its enclosing component's props.
				// Attribute a use to every containing component whose own props
				// binding matches the member root, rather than only the nearest one.
				for _, c := range comps {
					contains := false
					for parent := n; parent != nil; parent = parent.Parent {
						if parent == c.node {
							contains = true
							break
						}
					}
					if contains && rootMatches(root, c) {
						parent := n.Parent
						nestedMember := parent != nil && ((parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == n) || (parent.Kind == ast.KindElementAccessExpression && parent.AsElementAccessExpression().Expression == n))
						// When the root prop is declared, only the complete access is
						// relevant. An undeclared root, on the other hand, is reported
						// at every member level by eslint-plugin-react.
						if nestedMember && isDeclared(c.declared, names[:1]) {
							continue
						}
						if path, ok := propsPath(root, names, c); ok {
							if n.Kind == ast.KindElementAccessExpression {
								// A lone dynamic lookup has no prop name to validate. Once an
								// earlier segment is already missing, reporting a trailing
								// array/object index adds only a duplicate diagnostic.
								argument := unwrap(n.AsElementAccessExpression().ArgumentExpression)
								if (len(path) == 1 && path[0] == "__COMPUTED_PROP__") ||
									(len(path) > 1 && (argument == nil || argument.Kind != ast.KindStringLiteral) && !isDeclared(c.declared, path[:len(path)-1])) {
									continue
								}
							}
							appendUse(c, memberReportNode(n), path)
						}
					}
				}
			}
			n.ForEachChild(walk)
			return false
		}
		ctx.SourceFile.Node.ForEachChild(walk)
		// A component declaration can be hoisted below its propTypes assignment.
		// Replay assignments after collection so declaration order does not decide
		// whether the component is known.
		for _, assignmentNode := range propTypeAssignments {
			applyPropTypeAssignment(assignmentNode)
		}
		for _, c := range comps {
			if len(c.used) == 0 || !c.declaredBlock && o.skipUndeclared {
				continue
			}
			slices.SortFunc(c.used, func(a, b use) int {
				if position := a.node.Pos() - b.node.Pos(); position != 0 {
					return position
				}
				return len(a.names) - len(b.names)
			})
			for _, u := range c.used {
				name := u.names[0]
				if isObjectPrototypeProperty(u.names[len(u.names)-1]) || slices.Contains(o.ignore, name) {
					continue
				}
				if !isDeclared(c.declared, u.names) {
					ctx.ReportNode(u.node, rule.RuleMessage{Id: "missingPropType", Description: fmt.Sprintf("'%s' is missing in props validation", reportName(u.names)), Data: map[string]string{"name": reportName(u.names)}})
				}
			}
		}
		return rule.RuleListeners{}
	},
}
