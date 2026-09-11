package no_unused_prop_types

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_unused_prop_types.schema.json
var schemaJSON []byte

type options struct {
	ignore           []string
	customValidators []string
	skipShapeProps   bool
}

type prop struct {
	name     string
	fullName string
	node     *ast.Node
	shape    bool
	never    bool
	children []*prop
}

type component struct {
	node     *ast.Node
	name     string
	target   []string
	binding  *ast.Symbol
	declared []*prop
	used     map[string]bool
	anyUsed  bool
	aliases  map[*ast.Symbol][]string
	resolve  func(*ast.Node) *ast.Symbol
}

func parseOptions(raw []any) options {
	o := options{skipShapeProps: true}
	if len(raw) == 0 {
		return o
	}
	m, _ := raw[0].(map[string]interface{})
	if values, ok := m["ignore"].([]interface{}); ok {
		for _, value := range values {
			if s, ok := value.(string); ok {
				o.ignore = append(o.ignore, s)
			}
		}
	}
	if values, ok := m["customValidators"].([]interface{}); ok {
		for _, value := range values {
			if s, ok := value.(string); ok {
				o.customValidators = append(o.customValidators, s)
			}
		}
	}
	if value, ok := m["skipShapeProps"].(bool); ok {
		o.skipShapeProps = value
	}
	return o
}

func unwrap(node *ast.Node) *ast.Node {
	return reactutil.SkipExpressionWrappers(node)
}

func staticName(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	return utils.GetStaticPropertyName(node)
}

func propertyName(node *ast.Node) string {
	name, _ := staticName(node)
	return name
}

func componentName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	if node.Kind == ast.KindMethodDeclaration || node.Kind == ast.KindGetAccessor || node.Kind == ast.KindSetAccessor {
		return propertyName(node.Name())
	}
	if name := reactutil.BindingIdentifierName(node); name != "" {
		return name
	}
	for current := node; current != nil; current = current.Parent {
		parent := current.Parent
		if parent == nil {
			break
		}
		switch parent.Kind {
		case ast.KindVariableDeclaration:
			if parent.AsVariableDeclaration().Initializer == current {
				if name := parent.AsVariableDeclaration().Name(); name != nil && name.Kind == ast.KindIdentifier {
					return name.AsIdentifier().Text
				}
			}
		case ast.KindBinaryExpression:
			bin := parent.AsBinaryExpression()
			if bin.Right == current && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken {
				left := unwrap(bin.Left)
				if left != nil && left.Kind == ast.KindIdentifier {
					return left.AsIdentifier().Text
				}
				if parts := propertyAccessParts(left); len(parts) > 0 {
					return parts[len(parts)-1]
				}
			}
		case ast.KindPropertyAssignment:
			if parent.AsPropertyAssignment().Initializer == current {
				return propertyName(parent.AsPropertyAssignment().Name())
			}
		}
	}
	return ""
}

func isReactClass(node *ast.Node, pragma string) bool {
	return node != nil && (node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindClassExpression) && reactutil.ExtendsReactComponent(node, pragma)
}

func isFunctionLike(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return true
	default:
		return false
	}
}

func propKey(node *ast.Node) (string, *ast.Node) {
	if node == nil {
		return "", nil
	}
	name, ok := staticName(node)
	if node.Kind == ast.KindComputedPropertyName {
		expression := unwrap(node.AsComputedPropertyName().Expression)
		if ok {
			return name, expression
		}
		if expression != nil && expression.Kind == ast.KindIdentifier {
			return expression.AsIdentifier().Text, expression
		}
		return "", nil
	}
	if !ok {
		return "", nil
	}
	return name, node
}

func newProp(name, fullName string, node *ast.Node) *prop {
	return &prop{name: name, fullName: fullName, node: node}
}

func propMap(node *ast.Node, customValidators []string, prefix string) []*prop {
	node = unwrap(node)
	if node == nil || node.Kind != ast.KindObjectLiteralExpression {
		return nil
	}
	var result []*prop
	for _, member := range node.AsObjectLiteralExpression().Properties.Nodes {
		if member == nil {
			continue
		}
		if member.Kind == ast.KindSpreadAssignment {
			result = append(result, newProp("__ANY_KEY__", "__ANY_KEY__", member))
			continue
		}
		var key, value *ast.Node
		switch member.Kind {
		case ast.KindPropertyAssignment:
			key = member.AsPropertyAssignment().Name()
			value = member.AsPropertyAssignment().Initializer
		case ast.KindShorthandPropertyAssignment:
			key = member.AsShorthandPropertyAssignment().Name()
			value = key
		case ast.KindMethodDeclaration:
			// ESTree exposes object methods as method-valued Property nodes.
			key = member.Name()
			value = member
		default:
			continue
		}
		name, reportNode := propKey(key)
		if name == "" {
			result = append(result, newProp("__ANY_KEY__", "__ANY_KEY__", member))
			continue
		}
		fullName := name
		if prefix != "" {
			fullName = prefix + "." + name
		}
		current := newProp(name, fullName, reportNode)
		validator := unwrap(value)
		if validator != nil && validator.Kind == ast.KindPropertyAccessExpression {
			pa := validator.AsPropertyAccessExpression()
			if propertyName(pa.Name()) == "isRequired" {
				validator = unwrap(pa.Expression)
			}
		}
		if validator != nil && validator.Kind == ast.KindCallExpression {
			call := validator.AsCallExpression()
			callee := unwrap(call.Expression)
			method := ""
			if callee != nil && callee.Kind == ast.KindPropertyAccessExpression {
				pa := callee.AsPropertyAccessExpression()
				method = propertyName(pa.Name())
				if object := unwrap(pa.Expression); object != nil && object.Kind == ast.KindIdentifier && slices.Contains(customValidators, object.AsIdentifier().Text) {
					method = "custom"
				}
			}
			switch method {
			case "shape", "exact":
				current.shape = true
				if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
					current.children = propMap(call.Arguments.Nodes[0], customValidators, fullName)
				}
			case "arrayOf", "objectOf":
				if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
					current.children = validatorChildren(call.Arguments.Nodes[0], customValidators, fullName+".*")
					child := unwrap(call.Arguments.Nodes[0])
					if len(current.children) == 0 && child != nil && child.Kind == ast.KindPropertyAccessExpression {
						current.children = []*prop{newProp("*", fullName+".*", call.Arguments.Nodes[0])}
					}
				}
			case "oneOfType":
				if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
					arg := unwrap(call.Arguments.Nodes[0])
					if arg != nil && arg.Kind == ast.KindArrayLiteralExpression {
						for _, candidate := range arg.AsArrayLiteralExpression().Elements.Nodes {
							candidateProp := newProp(name, fullName, reportNode)
							candidateProp.children = validatorChildren(candidate, customValidators, fullName)
							current.children = mergeProps(current.children, candidateProp.children)
						}
					}
				}
			}
		}
		result = append(result, current)
	}
	return result
}

func validatorChildren(node *ast.Node, customValidators []string, prefix string) []*prop {
	node = unwrap(node)
	if node == nil {
		return nil
	}
	if node.Kind == ast.KindPropertyAccessExpression && propertyName(node.Name()) == "isRequired" {
		node = unwrap(node.AsPropertyAccessExpression().Expression)
	}
	if node == nil || node.Kind != ast.KindCallExpression {
		return nil
	}
	call := node.AsCallExpression()
	callee := unwrap(call.Expression)
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression || call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return nil
	}
	access := callee.AsPropertyAccessExpression()
	if object := unwrap(access.Expression); object != nil && object.Kind == ast.KindIdentifier && slices.Contains(customValidators, object.Text()) {
		return nil
	}
	argument := unwrap(call.Arguments.Nodes[0])
	switch propertyName(access.Name()) {
	case "shape", "exact":
		return propMap(argument, customValidators, prefix)
	case "arrayOf", "objectOf":
		return validatorChildren(argument, customValidators, prefix+".*")
	case "oneOfType":
		var result []*prop
		if argument != nil && argument.Kind == ast.KindArrayLiteralExpression {
			for _, candidate := range argument.AsArrayLiteralExpression().Elements.Nodes {
				result = mergeProps(result, validatorChildren(candidate, customValidators, prefix))
			}
		}
		return result
	}
	return nil
}

func mergeProps(left, right []*prop) []*prop {
	for _, candidate := range right {
		found := false
		for _, existing := range left {
			if existing.fullName == candidate.fullName {
				existing.children = mergeProps(existing.children, candidate.children)
				found = true
				break
			}
		}
		if !found {
			left = append(left, candidate)
		}
	}
	return left
}

func typeProps(node *ast.Node, aliases map[string][]*ast.Node, resolve func(*ast.Node) *ast.Symbol, seen map[string]bool, prefix string) []*prop {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindParenthesizedType:
		return typeProps(node.AsParenthesizedTypeNode().Type, aliases, resolve, seen, prefix)
	case ast.KindTypeLiteral:
		var result []*prop
		members := node.AsTypeLiteralNode().Members
		if members == nil {
			return nil
		}
		for _, member := range members.Nodes {
			if member == nil {
				continue
			}
			var name string
			var reportNode *ast.Node
			var typ *ast.Node
			switch member.Kind {
			case ast.KindPropertySignature:
				property := member.AsPropertySignatureDeclaration()
				name, reportNode = propKey(property.Name())
				typ = property.Type
			case ast.KindMethodSignature:
				name, reportNode = propKey(member.AsMethodSignatureDeclaration().Name())
			default:
				continue
			}
			if name == "" {
				continue
			}
			fullName := name
			if prefix != "" {
				fullName = prefix + "." + name
			}
			p := newProp(name, fullName, reportNode)
			p.never = typ != nil && typ.Kind == ast.KindNeverKeyword
			result = append(result, p)
		}
		return result
	case ast.KindInterfaceDeclaration:
		decl := node.AsInterfaceDeclaration()
		if decl == nil {
			return nil
		}
		var result []*prop
		if decl.HeritageClauses != nil {
			for _, clause := range decl.HeritageClauses.Nodes {
				if clause == nil || clause.Kind != ast.KindHeritageClause || clause.AsHeritageClause().Types == nil {
					continue
				}
				for _, typ := range clause.AsHeritageClause().Types.Nodes {
					if typ != nil && typ.Kind == ast.KindExpressionWithTypeArguments {
						result = mergeProps(result, typeProps(typ.AsExpressionWithTypeArguments().Expression, aliases, resolve, seen, prefix))
					} else if typ != nil && typ.Kind == ast.KindTypeReference {
						result = mergeProps(result, typeProps(typ, aliases, resolve, seen, prefix))
					}
				}
			}
		}
		if decl.Members == nil {
			return result
		}
		for _, member := range decl.Members.Nodes {
			if member == nil {
				continue
			}
			var name string
			var reportNode *ast.Node
			var typ *ast.Node
			switch member.Kind {
			case ast.KindPropertySignature:
				property := member.AsPropertySignatureDeclaration()
				name, reportNode = propKey(property.Name())
				typ = property.Type
			case ast.KindMethodSignature:
				name, reportNode = propKey(member.AsMethodSignatureDeclaration().Name())
			default:
				continue
			}
			if name == "" {
				continue
			}
			fullName := name
			if prefix != "" {
				fullName = prefix + "." + name
			}
			p := newProp(name, fullName, reportNode)
			p.never = typ != nil && typ.Kind == ast.KindNeverKeyword
			result = append(result, p)
		}
		return result
	case ast.KindTypeReference:
		if argument := reactGenericArgument(node.AsTypeReferenceNode().TypeName, node.AsTypeReferenceNode().TypeArguments, resolve); argument != nil {
			return typeProps(argument, aliases, resolve, seen, prefix)
		}
		name := reactutil.EntityNameRightmost(node.AsTypeReferenceNode().TypeName)
		if name == nil || name.Kind != ast.KindIdentifier || seen[name.AsIdentifier().Text] {
			return nil
		}
		if name.AsIdentifier().Text == "ReturnType" {
			args := node.AsTypeReferenceNode().TypeArguments
			if args != nil && len(args.Nodes) > 0 {
				argument := args.Nodes[0]
				if argument.Kind == ast.KindFunctionType {
					return typeProps(argument.AsFunctionTypeNode().Type, aliases, resolve, seen, prefix)
				}
				if argument.Kind == ast.KindTypeQuery {
					query := argument.AsTypeQueryNode()
					if query.ExprName != nil && query.ExprName.Kind == ast.KindIdentifier {
						if initializer := reactutil.ResolveIdentifierInitializer(query.ExprName.AsNode(), nil); initializer != nil {
							initializer = unwrap(initializer)
							if isFunctionLike(initializer) {
								return typePropsFromFunctionBody(initializer.Body(), aliases, resolve, seen, prefix)
							}
						}
					}
				}
			}
			return nil
		}
		seen[name.AsIdentifier().Text] = true
		var result []*prop
		for _, declaration := range visibleTypeAliases(node, aliases[name.AsIdentifier().Text]) {
			result = mergeProps(result, typeProps(declaration, aliases, resolve, seen, prefix))
		}
		return result
	case ast.KindIntersectionType:
		var result []*prop
		for _, part := range node.AsIntersectionTypeNode().Types.Nodes {
			result = mergeProps(result, typeProps(part, aliases, resolve, seen, prefix))
		}
		return result
	}
	return nil
}

func visibleTypeAliases(use *ast.Node, candidates []*ast.Node) []*ast.Node {
	bestDepth := -1
	var visible []*ast.Node
	for _, candidate := range candidates {
		declarationScope := typeDeclarationScope(candidate)
		if declarationScope == nil {
			continue
		}
		for scope, depth := use, 0; scope != nil; scope, depth = scope.Parent, depth+1 {
			if scope.Kind != ast.KindBlock && scope.Kind != ast.KindSourceFile && scope.Kind != ast.KindModuleBlock {
				continue
			}
			if declarationScope != scope {
				continue
			}
			if bestDepth < 0 || depth < bestDepth {
				bestDepth, visible = depth, []*ast.Node{candidate}
			} else if depth == bestDepth {
				visible = append(visible, candidate)
			}
			break
		}
	}
	return visible
}

func typeDeclarationScope(node *ast.Node) *ast.Node {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindBlock, ast.KindSourceFile, ast.KindModuleBlock:
			return parent
		}
	}
	return nil
}

func typePropsFromFunctionBody(body *ast.Node, aliases map[string][]*ast.Node, resolve func(*ast.Node) *ast.Symbol, seen map[string]bool, prefix string) []*prop {
	body = unwrap(body)
	if body == nil {
		return nil
	}
	if body.Kind == ast.KindBlock {
		var returned *ast.Node
		var visit ast.Visitor
		visit = func(statement *ast.Node) bool {
			if ast.IsFunctionLike(statement) || ast.IsClassLike(statement) {
				return false
			}
			if statement.Kind == ast.KindReturnStatement {
				returned = statement.AsReturnStatement().Expression
				return false
			}
			statement.ForEachChild(visit)
			return false
		}
		body.ForEachChild(visit)
		return typePropsFromFunctionBody(returned, aliases, resolve, seen, prefix)
	}
	if body.Kind == ast.KindCallExpression {
		var result []*prop
		if arguments := body.AsCallExpression().TypeArguments; arguments != nil {
			for _, argument := range arguments.Nodes {
				result = mergeProps(result, typeProps(argument, aliases, resolve, seen, prefix))
			}
		}
		return result
	}
	if body.Kind == ast.KindObjectLiteralExpression {
		var result []*prop
		for _, member := range body.AsObjectLiteralExpression().Properties.Nodes {
			if member != nil && member.Kind == ast.KindSpreadAssignment {
				result = mergeProps(result, typePropsFromFunctionBody(member.AsSpreadAssignment().Expression, aliases, resolve, seen, prefix))
				continue
			}
			if member == nil || (member.Kind != ast.KindPropertyAssignment && member.Kind != ast.KindShorthandPropertyAssignment && member.Kind != ast.KindMethodDeclaration) {
				continue
			}
			name, reportNode := propKey(member.Name())
			if name == "" {
				continue
			}
			fullName := name
			if prefix != "" {
				fullName = prefix + "." + name
			}
			result = append(result, newProp(name, fullName, reportNode))
		}
		return result
	}
	return nil
}

func collectTypeAliases(root *ast.Node) map[string][]*ast.Node {
	aliases := map[string][]*ast.Node{}
	var walk ast.Visitor
	walk = func(node *ast.Node) bool {
		if node != nil && node.Kind == ast.KindTypeAliasDeclaration {
			decl := node.AsTypeAliasDeclaration()
			if name := decl.Name(); name != nil && name.Kind == ast.KindIdentifier {
				aliases[name.AsIdentifier().Text] = append(aliases[name.AsIdentifier().Text], decl.Type)
			}
		}
		if node != nil && node.Kind == ast.KindInterfaceDeclaration {
			decl := node.AsInterfaceDeclaration()
			if name := decl.Name(); name != nil && name.Kind == ast.KindIdentifier {
				aliases[name.AsIdentifier().Text] = append(aliases[name.AsIdentifier().Text], node)
			}
		}
		node.ForEachChild(walk)
		return false
	}
	root.ForEachChild(walk)
	return aliases
}

func classType(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	heritage := ast.GetClassExtendsHeritageElement(node)
	if heritage == nil {
		return nil
	}
	types := heritage.AsExpressionWithTypeArguments().TypeArguments
	if types == nil || len(types.Nodes) == 0 {
		return nil
	}
	return types.Nodes[0]
}

func reactGenericArgument(name *ast.Node, arguments *ast.NodeList, resolve func(*ast.Node) *ast.Symbol) *ast.Node {
	if name == nil || arguments == nil || resolve == nil {
		return nil
	}
	root, member := name, ""
	switch name.Kind {
	case ast.KindQualifiedName:
		root = name.AsQualifiedName().Left
		member = propertyName(name.AsQualifiedName().Right)
	case ast.KindPropertyAccessExpression:
		root = name.AsPropertyAccessExpression().Expression
		member = propertyName(name.AsPropertyAccessExpression().Name())
	}
	if root.Kind != ast.KindIdentifier {
		return nil
	}
	binding := resolve(root)
	if binding == nil {
		if member == "" || root.Text() != "React" {
			return nil
		}
		return reactGenericArgumentAtIndex(member, arguments)
	}
	imported := ""
	for _, statement := range ast.GetSourceFileOfNode(name).Statements.Nodes {
		if statement.Kind != ast.KindImportDeclaration {
			continue
		}
		declaration := statement.AsImportDeclaration()
		if declaration.ModuleSpecifier == nil || declaration.ModuleSpecifier.Text() != "react" || declaration.ImportClause == nil {
			continue
		}
		clause := declaration.ImportClause.AsImportClause()
		if member != "" {
			if clause.Name() != nil && utils.BindingNameSymbol(clause.Name()) == binding {
				imported = member
			}
			if bindings := clause.NamedBindings; bindings != nil && bindings.Kind == ast.KindNamespaceImport && utils.BindingNameSymbol(bindings.Name()) == binding {
				imported = member
			}
		} else if bindings := clause.NamedBindings; bindings != nil && bindings.Kind == ast.KindNamedImports {
			for _, element := range bindings.AsNamedImports().Elements.Nodes {
				specifier := element.AsImportSpecifier()
				if utils.BindingNameSymbol(specifier.Name()) == binding {
					imported = root.Text()
					if specifier.PropertyName != nil {
						imported = specifier.PropertyName.Text()
					}
				}
			}
		}
	}
	return reactGenericArgumentAtIndex(imported, arguments)
}

func reactGenericArgumentAtIndex(imported string, arguments *ast.NodeList) *ast.Node {
	index := 0
	switch imported {
	case "forwardRef", "ForwardRefRenderFunction":
		index = 1
	case "ComponentProps", "ComponentPropsWithRef", "ComponentPropsWithoutRef", "VFC", "VoidFunctionComponent", "PropsWithChildren", "SFC", "StatelessComponent", "FunctionComponent", "FC":
	default:
		return nil
	}
	if len(arguments.Nodes) <= index {
		return nil
	}
	return arguments.Nodes[index]
}

func functionComponentType(node *ast.Node, resolve func(*ast.Node) *ast.Symbol) *ast.Node {
	for current := node; current != nil && current.Parent != nil; current = current.Parent {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression, ast.KindTypeAssertionExpression:
			continue
		case ast.KindCallExpression:
			call := parent.AsCallExpression()
			callee := unwrap(call.Expression)
			if callee != nil && call.TypeArguments != nil {
				if (callee.Kind == ast.KindIdentifier && callee.Text() == "forwardRef") ||
					(callee.Kind == ast.KindPropertyAccessExpression && propertyName(callee.AsPropertyAccessExpression().Name()) == "forwardRef") {
					return reactGenericArgument(callee, call.TypeArguments, resolve)
				}
			}
			continue
		case ast.KindVariableDeclaration:
			if parent.AsVariableDeclaration().Initializer != current || parent.AsVariableDeclaration().Type == nil {
				return nil
			}
			typ := parent.AsVariableDeclaration().Type
			if typ.Kind == ast.KindTypeReference {
				return reactGenericArgument(typ.AsTypeReferenceNode().TypeName, typ.AsTypeReferenceNode().TypeArguments, resolve)
			}
			return nil
		default:
			return nil
		}
	}
	return nil
}

func appendDeclared(c *component, props []*prop) {
	for _, p := range props {
		if p != nil {
			if declaredContains(c.declared, p.fullName) {
				continue
			}
			c.declared = append(c.declared, p)
		}
	}
}

func declaredContains(props []*prop, fullName string) bool {
	for _, p := range props {
		if p.fullName == fullName || declaredContains(p.children, fullName) {
			return true
		}
	}
	return false
}

func findDeclared(props []*prop, fullName string) *prop {
	for _, p := range props {
		if p.fullName == fullName {
			return p
		}
		if nested := findDeclared(p.children, fullName); nested != nil {
			return nested
		}
	}
	return nil
}

func declaredProps(node *ast.Node, customValidators []string, wrappers []reactutil.PropWrapperEntry, resolve func(*ast.Node) *ast.Symbol) []*prop {
	seen := map[*ast.Node]bool{}
	for node = unwrap(node); node != nil && !seen[node]; node = unwrap(node) {
		seen[node] = true
		if node.Kind == ast.KindCallExpression && reactutil.IsPropWrapperCall(node, wrappers) {
			call := node.AsCallExpression()
			if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
				node = call.Arguments.Nodes[0]
				continue
			}
		}
		if node.Kind == ast.KindIdentifier {
			if symbol := resolve(node); symbol != nil {
				var initializer *ast.Node
				for _, declaration := range symbol.Declarations {
					if declaration.Kind == ast.KindVariableDeclaration {
						if value := declaration.AsVariableDeclaration().Initializer; value != nil {
							initializer = value
						}
					}
				}
				if initializer != nil {
					node = initializer
					continue
				}
			}
		}
		return propMap(node, customValidators, "")
	}
	return nil
}

func propertyAccessParts(node *ast.Node) []string {
	node = unwrap(node)
	if node == nil {
		return nil
	}
	if node.Kind == ast.KindIdentifier {
		return []string{node.AsIdentifier().Text}
	}
	if node.Kind != ast.KindPropertyAccessExpression {
		return nil
	}
	pa := node.AsPropertyAccessExpression()
	parts := propertyAccessParts(pa.Expression)
	if len(parts) == 0 {
		return nil
	}
	name := propertyName(pa.Name())
	if name == "" {
		return nil
	}
	return append(parts, name)
}

func componentBinding(node *ast.Node, resolve func(*ast.Node) *ast.Symbol) *ast.Symbol {
	if node == nil {
		return nil
	}
	if node.Kind == ast.KindFunctionDeclaration || node.Kind == ast.KindClassDeclaration {
		return node.Symbol()
	}
	for current := node; current != nil && current.Parent != nil; {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression,
			ast.KindNonNullExpression, ast.KindTypeAssertionExpression, ast.KindCallExpression:
			current = parent
			continue
		case ast.KindPropertyAssignment:
			if parent.AsPropertyAssignment().Initializer == current && parent.Parent != nil {
				current = parent.Parent
				continue
			}
		case ast.KindObjectLiteralExpression:
			current = parent
			continue
		case ast.KindVariableDeclaration:
			if parent.AsVariableDeclaration().Initializer == current {
				name := parent.AsVariableDeclaration().Name()
				if name != nil && name.Kind == ast.KindIdentifier {
					return utils.BindingNameSymbol(name)
				}
			}
		case ast.KindBinaryExpression:
			assignment := parent.AsBinaryExpression()
			if assignment.OperatorToken != nil && assignment.OperatorToken.Kind == ast.KindEqualsToken && assignment.Right == current {
				if root := propertyAccessRoot(assignment.Left); root != nil && root.Kind == ast.KindIdentifier && resolve != nil {
					return resolve(root)
				}
			}
		}
		return nil
	}
	return nil
}

func componentTarget(node *ast.Node) []string {
	name := componentName(node)
	if name == "" {
		return nil
	}
	target := []string{name}
	for current := node; current != nil && current.Parent != nil; {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression,
			ast.KindNonNullExpression, ast.KindTypeAssertionExpression, ast.KindCallExpression:
			current = parent
			continue
		case ast.KindPropertyAssignment:
			if parent.AsPropertyAssignment().Initializer != current || parent.Parent == nil {
				return target
			}
			member := propertyName(parent.AsPropertyAssignment().Name())
			if member == "" {
				return target
			}
			if len(target) == 0 || target[0] != member {
				target = append([]string{member}, target...)
			}
			current = parent.Parent
			continue
		case ast.KindObjectLiteralExpression:
			current = parent
			continue
		case ast.KindVariableDeclaration:
			if parent.AsVariableDeclaration().Initializer == current {
				if objectName := parent.AsVariableDeclaration().Name(); objectName != nil && objectName.Kind == ast.KindIdentifier {
					if len(target) == 1 && target[0] == objectName.AsIdentifier().Text {
						return target
					}
					return append([]string{objectName.AsIdentifier().Text}, target...)
				}
			}
		case ast.KindBinaryExpression:
			assignment := parent.AsBinaryExpression()
			if assignment.OperatorToken != nil && assignment.OperatorToken.Kind == ast.KindEqualsToken && assignment.Right == current {
				if parts := propertyAccessParts(assignment.Left); len(parts) > 0 {
					return parts
				}
			}
		}
		return target
	}
	return target
}

func componentScope(node *ast.Node) *ast.Node {
	for current := node; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindBlock, ast.KindSourceFile, ast.KindModuleBlock:
			return current
		}
	}
	return nil
}

func propertyAccessRoot(node *ast.Node) *ast.Node {
	node = unwrap(node)
	if node == nil {
		return nil
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		return propertyAccessRoot(node.AsPropertyAccessExpression().Expression)
	}
	return node
}

func assignmentComponent(target []string, root *ast.Node, components []*component, resolve func(*ast.Node) *ast.Symbol) *component {
	if len(target) == 0 {
		return nil
	}
	if root != nil && root.Kind == ast.KindIdentifier && resolve != nil {
		if binding := resolve(root); binding != nil {
			for _, c := range components {
				if c.binding == binding && slices.Equal(c.target, target) {
					return c
				}
			}
		}
	}
	// A reference resolver is unavailable for some syntax-only runs. Keep the
	// fallback confined to the same lexical scope so sibling shadowed bindings
	// cannot be associated by text alone.
	for _, c := range components {
		if slices.Equal(c.target, target) && root != nil && componentScope(c.node) == componentScope(root) {
			return c
		}
	}
	return nil
}

func addAssignmentDeclarations(root *ast.Node, components []*component, customValidators []string, wrappers []reactutil.PropWrapperEntry, resolve func(*ast.Node) *ast.Symbol) {
	var walk ast.Visitor
	walk = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if node.Kind == ast.KindBinaryExpression {
			bin := node.AsBinaryExpression()
			if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken {
				parts := propertyAccessParts(bin.Left)
				if len(parts) >= 2 && (parts[len(parts)-1] == "propTypes" || (len(parts) >= 3 && (parts[1] == "propTypes" || parts[len(parts)-2] == "propTypes"))) {
					propTypesIndex := -1
					for i, part := range parts {
						if part == "propTypes" {
							propTypesIndex = i
							break
						}
					}
					target := parts[:propTypesIndex]
					if propTypesIndex == 2 && parts[1] == "prototype" {
						target = parts[:1]
					}
					c := assignmentComponent(target, propertyAccessRoot(bin.Left), components, resolve)
					if c != nil {
						if propTypesIndex == len(parts)-1 {
							appendDeclared(c, declaredProps(bin.Right, customValidators, wrappers, resolve))
						} else {
							path := parts[propTypesIndex+1:]
							declared := newProp(path[len(path)-1], strings.Join(path, "."), node)
							if len(path) == 1 {
								appendDeclared(c, []*prop{declared})
							} else if parent := findDeclared(c.declared, strings.Join(path[:len(path)-1], ".")); parent != nil {
								parent.children = mergeProps(parent.children, []*prop{declared})
							}
						}
					}
				}
			}
		}
		node.ForEachChild(walk)
		return false
	}
	root.ForEachChild(walk)
}

func bindingPath(name *ast.Node, prefix []string, c *component) {
	if name == nil {
		return
	}
	switch name.Kind {
	case ast.KindIdentifier:
		path := append([]string(nil), prefix...)
		if symbol := utils.BindingNameSymbol(name); symbol != nil {
			c.aliases[symbol] = path
		}
		if len(path) > 0 {
			c.mark(path)
		}
	case ast.KindObjectBindingPattern:
		for _, element := range name.AsBindingPattern().Elements.Nodes {
			if element.Kind != ast.KindBindingElement {
				continue
			}
			be := element.AsBindingElement()
			if be.DotDotDotToken != nil {
				c.anyUsed = true
				continue
			}
			key := be.Name()
			part := ""
			if be.PropertyName != nil {
				if be.PropertyName.Kind == ast.KindComputedPropertyName {
					if _, ok := staticName(be.PropertyName); !ok {
						c.anyUsed = true
						continue
					}
				}
				part = propertyName(be.PropertyName)
			} else if key != nil && key.Kind == ast.KindIdentifier {
				part = key.AsIdentifier().Text
			}
			if part == "" {
				continue
			}
			path := append(append([]string(nil), prefix...), part)
			c.mark(path)
			bindingPath(key, path, c)
		}
	}
}

func (c *component) mark(path []string) {
	for i := 1; i <= len(path); i++ {
		c.used[strings.Join(path[:i], ".")] = true
	}
}

func (c *component) pathFrom(node *ast.Node) ([]string, bool) {
	node = unwrap(node)
	if node == nil {
		return nil, false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		path, ok := c.aliases[c.resolve(node)]
		return path, ok
	case ast.KindPropertyAccessExpression:
		pa := node.AsPropertyAccessExpression()
		if base := unwrap(pa.Expression); base != nil && base.Kind == ast.KindThisKeyword && propertyName(pa.Name()) == "props" {
			return nil, true
		}
		path, ok := c.pathFrom(pa.Expression)
		if !ok {
			return nil, false
		}
		name := propertyName(pa.Name())
		if name == "" {
			return nil, false
		}
		return append(append([]string(nil), path...), name), true
	case ast.KindElementAccessExpression:
		ea := node.AsElementAccessExpression()
		path, ok := c.pathFrom(ea.Expression)
		if !ok {
			return nil, false
		}
		key, ok := staticName(unwrap(ea.ArgumentExpression))
		if !ok {
			key = "*"
		}
		return append(append([]string(nil), path...), key), true
	}
	return nil, false
}

func isVariableInitializer(node *ast.Node) bool {
	if node == nil {
		return false
	}
	cur := node
	for cur.Parent != nil && isExpressionWrapper(cur.Parent) {
		cur = cur.Parent
	}
	return cur.Parent != nil && cur.Parent.Kind == ast.KindVariableDeclaration && cur.Parent.AsVariableDeclaration().Initializer == cur
}

func isExpressionWrapper(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression,
		ast.KindNonNullExpression, ast.KindTypeAssertionExpression:
		return true
	default:
		return false
	}
}

func lifecycleFunctionName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	if node.Kind == ast.KindConstructor {
		return "constructor"
	}
	if node.Kind == ast.KindMethodDeclaration || node.Kind == ast.KindGetAccessor || node.Kind == ast.KindSetAccessor {
		return propertyName(node.Name())
	}
	if node.Kind != ast.KindArrowFunction && node.Kind != ast.KindFunctionExpression {
		return ""
	}
	parent := node.Parent
	for parent != nil && isExpressionWrapper(parent) {
		parent = parent.Parent
	}
	if parent != nil {
		switch parent.Kind {
		case ast.KindPropertyAssignment:
			if parent.AsPropertyAssignment().Initializer == node {
				return propertyName(parent.AsPropertyAssignment().Name())
			}
		case ast.KindPropertyDeclaration:
			if parent.AsPropertyDeclaration().Initializer == node {
				return propertyName(parent.AsPropertyDeclaration().Name())
			}
		}
	}
	return ""
}

func isLifecycleFunction(node *ast.Node, c *component, checkAsyncSafe bool) bool {
	if c == nil || c.node == nil || !isDirectComponentMember(node, c.node) {
		return false
	}
	switch lifecycleFunctionName(node) {
	case "constructor", "componentWillReceiveProps", "shouldComponentUpdate", "componentWillUpdate", "componentDidUpdate":
		return true
	case "getDerivedStateFromProps", "getSnapshotBeforeUpdate", "UNSAFE_componentWillReceiveProps", "UNSAFE_componentWillUpdate":
		return checkAsyncSafe
	}
	return false
}

func isDirectComponentMember(node, componentNode *ast.Node) bool {
	if node == nil || componentNode == nil {
		return false
	}
	member := node
	if node.Kind == ast.KindArrowFunction || node.Kind == ast.KindFunctionExpression {
		member = node.Parent
		for member != nil && isExpressionWrapper(member) {
			member = member.Parent
		}
	}
	return member != nil && member.Parent == componentNode
}

func isSetStateCallback(node *ast.Node) bool {
	if node == nil || node.Parent == nil || node.Parent.Kind != ast.KindCallExpression {
		return false
	}
	call := node.Parent.AsCallExpression()
	callee := unwrap(call.Expression)
	return callee != nil && callee.Kind == ast.KindPropertyAccessExpression && propertyName(callee.AsPropertyAccessExpression().Name()) == "setState"
}

func isSetStateUpdater(node *ast.Node) bool {
	if !isSetStateCallback(node) {
		return false
	}
	arguments := node.Parent.AsCallExpression().Arguments
	return arguments != nil && len(arguments.Nodes) > 0 && arguments.Nodes[0] == node
}

func isPropTypesValidatorCallback(node *ast.Node) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		if ast.IsFunctionLike(current) {
			return false
		}
		switch current.Kind {
		case ast.KindPropertyDeclaration:
			if propertyName(current.AsPropertyDeclaration().Name()) == "propTypes" {
				return true
			}
		case ast.KindPropertyAssignment:
			if propertyName(current.AsPropertyAssignment().Name()) == "propTypes" {
				return true
			}
		case ast.KindBinaryExpression:
			binary := current.AsBinaryExpression()
			if binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
				parts := propertyAccessParts(binary.Left)
				return len(parts) >= 2 && parts[len(parts)-1] == "propTypes"
			}
		}
	}
	return false
}

func (c *component) walkUsage(node *ast.Node, nested map[*ast.Node]bool, checkAsyncSafe bool) {
	if node == nil {
		return
	}
	if node != c.node && nested[node] {
		return
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		parameters := reactutil.FunctionParameters(node)
		for parameterIndex, parameter := range parameters {
			if parameter == nil || parameter.Kind != ast.KindParameter {
				continue
			}
			name := parameter.AsParameterDeclaration().Name()
			if name == nil {
				continue
			}
			setStateCallback := isSetStateCallback(node)
			setStateUpdater := isSetStateUpdater(node)
			if setStateUpdater && parameterIndex == 0 {
				continue
			}
			isRootComponentParameter := node == c.node && parameterIndex == 0
			isLifecycleParameter := parameterIndex == 0 && isLifecycleFunction(node, c, checkAsyncSafe)
			isValidatorParameter := parameterIndex == 0 && isPropTypesValidatorCallback(node)
			bindsProps := isRootComponentParameter || isLifecycleParameter || isValidatorParameter || (setStateUpdater && parameterIndex == 1)
			switch name.Kind {
			case ast.KindObjectBindingPattern:
				var prefix []string
				if call := node.Parent; call != nil && call.Kind == ast.KindCallExpression {
					callee := unwrap(call.AsCallExpression().Expression)
					if callee != nil && callee.Kind == ast.KindPropertyAccessExpression {
						method := propertyName(callee.AsPropertyAccessExpression().Name())
						if method == "map" || method == "forEach" || method == "filter" {
							if path, ok := c.pathFrom(callee.AsPropertyAccessExpression().Expression); ok {
								prefix = append(path, "*")
								bindsProps = true
							}
						}
					}
				}
				if bindsProps {
					bindingPath(name, prefix, c)
				}
			case ast.KindIdentifier:
				// Validator callbacks only contribute destructured paths. Other nested
				// parameters named `props` retain eslint-plugin-react's scoped alias.
				if !isValidatorParameter && ((bindsProps && (!isRootComponentParameter || name.AsIdentifier().Text == "props")) || (name.AsIdentifier().Text == "props" && !setStateCallback)) {
					bindingPath(name, nil, c)
				}
			}
		}
	}
	if node.Kind == ast.KindVariableDeclaration {
		vd := node.AsVariableDeclaration()
		if vd.Initializer != nil {
			initializer := unwrap(vd.Initializer)
			if initializer != nil && initializer.Kind == ast.KindThisKeyword && vd.Name().Kind == ast.KindObjectBindingPattern {
				for _, element := range vd.Name().AsBindingPattern().Elements.Nodes {
					if element.Kind != ast.KindBindingElement {
						continue
					}
					binding := element.AsBindingElement()
					key := binding.PropertyName
					if key == nil {
						key = binding.Name()
					}
					if propertyName(key) == "props" {
						bindingPath(binding.Name(), nil, c)
					}
				}
			} else if path, ok := c.pathFrom(vd.Initializer); ok {
				bindingPath(vd.Name(), path, c)
			}
		}
	}
	if node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression {
		cur := node
		for cur.Parent != nil && isExpressionWrapper(cur.Parent) {
			cur = cur.Parent
		}
		isMemberBase := cur.Parent != nil && (cur.Parent.Kind == ast.KindPropertyAccessExpression || cur.Parent.Kind == ast.KindElementAccessExpression) &&
			((cur.Parent.Kind == ast.KindPropertyAccessExpression && cur.Parent.AsPropertyAccessExpression().Expression == cur) ||
				(cur.Parent.Kind == ast.KindElementAccessExpression && cur.Parent.AsElementAccessExpression().Expression == cur))
		if !isVariableInitializer(node) && !isMemberBase {
			if path, ok := c.pathFrom(node); ok {
				c.mark(path)
			}
		}
	}
	if node.Kind == ast.KindJsxSpreadAttribute {
		if expression := unwrap(node.AsJsxSpreadAttribute().Expression); expression != nil && expression.Kind != ast.KindObjectLiteralExpression {
			c.anyUsed = true
		}
	}
	node.ForEachChild(func(child *ast.Node) bool {
		c.walkUsage(child, nested, checkAsyncSafe)
		return false
	})
}

func matches(pattern, used string) bool {
	pp, up := strings.Split(pattern, "."), strings.Split(used, ".")
	if len(pp) != len(up) {
		return false
	}
	for i := range pp {
		if pp[i] != "*" && pp[i] != up[i] {
			return false
		}
	}
	return true
}

func (c *component) propUsed(p *prop) bool {
	if p.name == "__ANY_KEY__" || p.name == "*" {
		return c.anyUsed || len(c.used) > 0
	}
	if c.anyUsed {
		return true
	}
	if p.shape && len(c.used) > 0 {
		return true
	}
	for used := range c.used {
		if matches(p.fullName, used) {
			return true
		}
	}
	return false
}

func (c *component) report(ctx rule.RuleContext, opts options, props []*prop) {
	for _, p := range props {
		if p == nil || p.node == nil {
			continue
		}
		if p.name == "__ANY_KEY__" {
			continue
		}
		if p.never {
			continue
		}
		if p.shape && opts.skipShapeProps {
			continue
		}
		if !slices.Contains(opts.ignore, p.fullName) && !c.propUsed(p) {
			ctx.ReportNode(p.node, rule.RuleMessage{
				Id:          "unusedPropType",
				Description: fmt.Sprintf("'%s' PropType is defined but prop is never used", p.fullName),
				Data:        map[string]string{"name": p.fullName},
			})
		}
		if len(p.children) > 0 {
			c.report(ctx, opts, p.children)
		}
	}
}

func NoUnusedPropTypesRuleRun(ctx rule.RuleContext, optionsRaw []any) rule.RuleListeners {
	opts := parseOptions(optionsRaw)
	pragma := reactutil.GetReactPragmaFromContext(ctx)
	createClass := reactutil.GetReactCreateClass(ctx.Settings)
	checkAsyncSafe := !reactutil.ReactVersionLessThan(ctx.Settings, 16, 3, 0)
	aliases := collectTypeAliases(ctx.SourceFile.AsNode())
	componentWrappers := reactutil.GetComponentWrapperFunctions(ctx.Settings, pragma)
	propWrappers := reactutil.GetPropWrapperFunctions(ctx.Settings)
	var components []*component
	nested := map[*ast.Node]bool{}
	resolveSymbol := func(identifier *ast.Node) *ast.Symbol {
		if identifier == nil || identifier.Kind != ast.KindIdentifier {
			return nil
		}
		if ctx.Refs != nil {
			return ctx.Refs.Resolve(identifier)
		}
		return utils.GetReferenceSymbol(identifier, ctx.TypeChecker)
	}
	newComponent := func(node *ast.Node) *component {
		c := &component{
			node: node, name: componentName(node), target: componentTarget(node), binding: componentBinding(node, resolveSymbol),
			used: map[string]bool{}, aliases: map[*ast.Symbol][]string{}, resolve: resolveSymbol,
		}
		components = append(components, c)
		nested[node] = true
		return c
	}

	var discover ast.Visitor
	discover = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if isReactClass(node, pragma) {
			newComponent(node)
		}
		if node.Kind == ast.KindCallExpression && reactutil.IsCreateClassCall(node.AsCallExpression(), pragma, createClass) {
			call := node.AsCallExpression()
			if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
				arg := unwrap(call.Arguments.Nodes[0])
				if arg != nil && arg.Kind == ast.KindObjectLiteralExpression {
					newComponent(arg)
				}
			}
		}
		if isFunctionLike(node) && reactutil.IsStatelessReactComponentWithWrappers(node, pragma, ctx.TypeChecker, componentWrappers) {
			newComponent(node)
		}
		node.ForEachChild(discover)
		return false
	}
	ctx.SourceFile.AsNode().ForEachChild(discover)

	return rule.RuleListeners{
		ast.KindEndOfFile: func(*ast.Node) {
			for _, c := range components {
				switch c.node.Kind {
				case ast.KindClassDeclaration, ast.KindClassExpression:
					if typ := classType(c.node); typ != nil {
						appendDeclared(c, typeProps(typ, aliases, resolveSymbol, map[string]bool{}, ""))
					}
					for _, member := range c.node.Members() {
						if member.Kind == ast.KindPropertyDeclaration {
							pd := member.AsPropertyDeclaration()
							name := propertyName(pd.Name())
							if name == "propTypes" && pd.Initializer != nil {
								appendDeclared(c, declaredProps(pd.Initializer, opts.customValidators, propWrappers, resolveSymbol))
							}
							if name == "props" && pd.Type != nil {
								appendDeclared(c, typeProps(pd.Type, aliases, resolveSymbol, map[string]bool{}, ""))
							}
						} else if (member.Kind == ast.KindGetAccessor || member.Kind == ast.KindMethodDeclaration) && ast.IsStatic(member) && propertyName(member.Name()) == "propTypes" {
							body := member.Body()
							if body != nil {
								body.ForEachChild(func(n *ast.Node) bool {
									if n.Kind == ast.KindReturnStatement && n.AsReturnStatement().Expression != nil {
										appendDeclared(c, declaredProps(n.AsReturnStatement().Expression, opts.customValidators, propWrappers, resolveSymbol))
									}
									return false
								})
							}
						}
					}
				case ast.KindObjectLiteralExpression:
					for _, member := range c.node.AsObjectLiteralExpression().Properties.Nodes {
						if propertyName(member.Name()) == "propTypes" && member.Kind == ast.KindPropertyAssignment {
							appendDeclared(c, declaredProps(member.AsPropertyAssignment().Initializer, opts.customValidators, propWrappers, resolveSymbol))
						}
					}
				}
				{
					// Function parameters are the root aliases for SFCs.
					if isFunctionLike(c.node) {
						componentType := functionComponentType(c.node, resolveSymbol)
						parameters := reactutil.FunctionParameters(c.node)
						if len(parameters) > 0 && parameters[0].Kind == ast.KindParameter {
							decl := parameters[0].AsParameterDeclaration()
							if decl.Name().Kind != ast.KindIdentifier || decl.Name().AsIdentifier().Text == "props" {
								bindingPath(decl.Name(), nil, c)
							}
							if decl.Type != nil && (componentType == nil || componentType.Parent == nil || componentType.Parent.Kind != ast.KindCallExpression) {
								componentType = decl.Type
							}
						}
						if componentType != nil {
							appendDeclared(c, typeProps(componentType, aliases, resolveSymbol, map[string]bool{}, ""))
						}
					}
					c.walkUsage(c.node, nested, checkAsyncSafe)
				}
			}
			addAssignmentDeclarations(ctx.SourceFile.AsNode(), components, opts.customValidators, propWrappers, resolveSymbol)
			for _, c := range components {
				c.report(ctx, opts, c.declared)
			}
		},
	}
}

var NoUnusedPropTypesRule = rule.Rule{
	Name:   "react/no-unused-prop-types",
	Schema: rule.NewSchema(schemaJSON),
	Run:    NoUnusedPropTypesRuleRun,
}
