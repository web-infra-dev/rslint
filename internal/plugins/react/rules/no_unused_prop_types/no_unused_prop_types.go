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
	binding  *ast.Symbol
	declared []*prop
	used     map[string]bool
	anyUsed  bool
	aliases  map[string][]string
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
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
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
	if !ok {
		return "", nil
	}
	if node.Kind == ast.KindComputedPropertyName {
		return name, unwrap(node.AsComputedPropertyName().Expression)
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
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		callee := unwrap(call.Expression)
		if callee != nil && callee.Kind == ast.KindPropertyAccessExpression {
			method := propertyName(callee.AsPropertyAccessExpression().Name())
			if method == "shape" || method == "exact" {
				if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
					return propMap(call.Arguments.Nodes[0], customValidators, prefix)
				}
			}
		}
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

func typeProps(node *ast.Node, aliases map[string][]*ast.Node, seen map[string]bool, prefix string) []*prop {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindParenthesizedType:
		return typeProps(node.AsParenthesizedTypeNode().Type, aliases, seen, prefix)
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
			p.children = typeProps(typ, aliases, seen, fullName)
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
						result = mergeProps(result, typeProps(typ.AsExpressionWithTypeArguments().Expression, aliases, seen, prefix))
					} else if typ != nil && typ.Kind == ast.KindTypeReference {
						result = mergeProps(result, typeProps(typ, aliases, seen, prefix))
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
			p.children = typeProps(typ, aliases, seen, fullName)
			result = append(result, p)
		}
		return result
	case ast.KindTypeReference:
		name := reactutil.EntityNameRightmost(node.AsTypeReferenceNode().TypeName)
		if name == nil || name.Kind != ast.KindIdentifier || seen[name.AsIdentifier().Text] {
			return nil
		}
		if name.AsIdentifier().Text == "ReturnType" {
			args := node.AsTypeReferenceNode().TypeArguments
			if args != nil && len(args.Nodes) > 0 {
				argument := args.Nodes[0]
				if argument.Kind == ast.KindFunctionType {
					return typeProps(argument.AsFunctionTypeNode().Type, aliases, seen, prefix)
				}
				if argument.Kind == ast.KindTypeQuery {
					query := argument.AsTypeQueryNode()
					if query.ExprName != nil && query.ExprName.Kind == ast.KindIdentifier {
						if initializer := reactutil.ResolveIdentifierInitializer(query.ExprName.AsNode(), nil); initializer != nil {
							initializer = unwrap(initializer)
							if initializer.Kind == ast.KindArrowFunction {
								return typePropsFromFunctionBody(initializer.AsArrowFunction().Body, aliases, seen, prefix)
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
			result = mergeProps(result, typeProps(declaration, aliases, seen, prefix))
		}
		return result
	case ast.KindIntersectionType:
		var result []*prop
		for _, part := range node.AsIntersectionTypeNode().Types.Nodes {
			result = mergeProps(result, typeProps(part, aliases, seen, prefix))
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

func typePropsFromFunctionBody(body *ast.Node, aliases map[string][]*ast.Node, seen map[string]bool, prefix string) []*prop {
	body = unwrap(body)
	if body == nil {
		return nil
	}
	if body.Kind == ast.KindObjectLiteralExpression {
		var result []*prop
		for _, member := range body.AsObjectLiteralExpression().Properties.Nodes {
			if member != nil && member.Kind == ast.KindSpreadAssignment {
				spread := unwrap(member.AsSpreadAssignment().Expression)
				if spread != nil && spread.Kind == ast.KindCallExpression {
					call := spread.AsCallExpression()
					callee := unwrap(call.Expression)
					if callee != nil && callee.Kind == ast.KindIdentifier && callee.AsIdentifier().Text == "bindActionCreators" && call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
						result = append(result, typePropsFromFunctionBody(call.Arguments.Nodes[0], aliases, seen, prefix)...)
					}
				}
				continue
			}
			if member == nil || member.Kind != ast.KindPropertyAssignment {
				continue
			}
			name, reportNode := propKey(member.AsPropertyAssignment().Name())
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

func functionComponentType(node *ast.Node) *ast.Node {
	for current := node; current != nil && current.Parent != nil; current = current.Parent {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression, ast.KindTypeAssertionExpression:
			continue
		case ast.KindVariableDeclaration:
			if parent.AsVariableDeclaration().Initializer != current || parent.AsVariableDeclaration().Type == nil {
				return nil
			}
			typ := parent.AsVariableDeclaration().Type
			if typ.Kind == ast.KindTypeReference {
				args := typ.AsTypeReferenceNode().TypeArguments
				if args != nil && len(args.Nodes) > 0 {
					return args.Nodes[0]
				}
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

func declaredProps(node *ast.Node, customValidators []string, wrappers []reactutil.PropWrapperEntry) []*prop {
	node = unwrap(node)
	if node == nil {
		return nil
	}
	if node.Kind == ast.KindCallExpression && reactutil.IsPropWrapperCall(node, wrappers) {
		call := node.AsCallExpression()
		if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
			return declaredProps(call.Arguments.Nodes[0], customValidators, wrappers)
		}
	}
	if node.Kind == ast.KindIdentifier {
		if initializer := reactutil.ResolveIdentifierInitializer(node, nil); initializer != nil && initializer != node {
			return declaredProps(initializer, customValidators, wrappers)
		}
	}
	return propMap(node, customValidators, "")
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

func assignmentComponent(root *ast.Node, components []*component, resolve func(*ast.Node) *ast.Symbol) *component {
	if root == nil || root.Kind != ast.KindIdentifier {
		return nil
	}
	if resolve != nil {
		if binding := resolve(root); binding != nil {
			for _, c := range components {
				if c.binding == binding {
					return c
				}
			}
		}
	}
	// A reference resolver is unavailable for some syntax-only runs. Keep the
	// fallback confined to the same lexical scope so sibling shadowed bindings
	// cannot be associated by text alone.
	for _, c := range components {
		if c.name == root.AsIdentifier().Text && componentScope(c.node) == componentScope(root) {
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
					c := assignmentComponent(propertyAccessRoot(bin.Left), components, resolve)
					if c != nil {
						if len(parts) == 2 || (len(parts) == 3 && parts[1] == "prototype") {
							appendDeclared(c, declaredProps(bin.Right, customValidators, wrappers))
						} else {
							var nestedProps []*prop
							for i := 2; i < len(parts); i++ {
								name := strings.Join(parts[2:i+1], ".")
								nestedProps = append(nestedProps, newProp(parts[i], name, node))
							}
							appendDeclared(c, nestedProps)
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
		c.aliases[name.AsIdentifier().Text] = path
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
	if len(path) == 0 {
		c.anyUsed = true
		return
	}
	for i := 1; i <= len(path); i++ {
		c.used[strings.Join(path[:i], ".")] = true
	}
}

func pathFrom(node *ast.Node, aliases map[string][]string) ([]string, bool) {
	node = unwrap(node)
	if node == nil {
		return nil, false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		path, ok := aliases[node.AsIdentifier().Text]
		return path, ok
	case ast.KindPropertyAccessExpression:
		pa := node.AsPropertyAccessExpression()
		if base := unwrap(pa.Expression); base != nil && base.Kind == ast.KindThisKeyword && propertyName(pa.Name()) == "props" {
			return nil, true
		}
		path, ok := pathFrom(pa.Expression, aliases)
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
		path, ok := pathFrom(ea.Expression, aliases)
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
	if parent != nil && parent.Kind == ast.KindPropertyAssignment && parent.AsPropertyAssignment().Initializer == node {
		return propertyName(parent.AsPropertyAssignment().Name())
	}
	return ""
}

func isLifecycleFunction(node *ast.Node, checkAsyncSafe bool) bool {
	switch lifecycleFunctionName(node) {
	case "constructor", "componentWillReceiveProps", "shouldComponentUpdate", "componentWillUpdate", "componentDidUpdate":
		return true
	case "getDerivedStateFromProps", "getSnapshotBeforeUpdate", "UNSAFE_componentWillReceiveProps", "UNSAFE_componentWillUpdate":
		return checkAsyncSafe
	}
	return false
}

func isSetStateCallback(node *ast.Node) bool {
	if node == nil || node.Parent == nil || node.Parent.Kind != ast.KindCallExpression {
		return false
	}
	callee := unwrap(node.Parent.AsCallExpression().Expression)
	return callee != nil && callee.Kind == ast.KindPropertyAccessExpression && propertyName(callee.AsPropertyAccessExpression().Name()) == "setState"
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
			if setStateCallback && parameterIndex == 0 {
				continue
			}
			isRootComponentParameter := node == c.node && parameterIndex == 0
			isLifecycleParameter := parameterIndex == 0 && isLifecycleFunction(node, checkAsyncSafe)
			isValidatorParameter := parameterIndex == 0 && isPropTypesValidatorCallback(node)
			bindsProps := isRootComponentParameter || isLifecycleParameter || isValidatorParameter || (setStateCallback && parameterIndex == 1)
			switch name.Kind {
			case ast.KindObjectBindingPattern:
				var prefix []string
				if call := node.Parent; call != nil && call.Kind == ast.KindCallExpression {
					callee := unwrap(call.AsCallExpression().Expression)
					if callee != nil && callee.Kind == ast.KindPropertyAccessExpression {
						method := propertyName(callee.AsPropertyAccessExpression().Name())
						if method == "map" || method == "forEach" || method == "filter" {
							if path, ok := pathFrom(callee.AsPropertyAccessExpression().Expression, c.aliases); ok {
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
				if bindsProps {
					c.aliases[name.AsIdentifier().Text] = nil
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
					if propertyName(binding.PropertyName) == "props" {
						bindingPath(binding.Name(), nil, c)
					}
				}
			} else if path, ok := pathFrom(vd.Initializer, c.aliases); ok {
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
			if path, ok := pathFrom(node, c.aliases); ok {
				c.mark(path)
			}
		}
	}
	if node.Kind == ast.KindJsxSpreadAttribute {
		if _, ok := pathFrom(node.AsJsxSpreadAttribute().Expression, c.aliases); ok {
			c.anyUsed = true
		}
	}
	if node.Kind == ast.KindSpreadAssignment {
		if _, ok := pathFrom(node.AsSpreadAssignment().Expression, c.aliases); ok {
			c.anyUsed = true
		}
	}
	if node.Kind == ast.KindSpreadElement {
		if _, ok := pathFrom(node.AsSpreadElement().Expression, c.aliases); ok {
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
	if p.name == "__ANY_KEY__" {
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
		if p.name == "__ANY_KEY__" || slices.Contains(opts.ignore, p.fullName) {
			continue
		}
		if p.never {
			continue
		}
		if p.shape && opts.skipShapeProps {
			continue
		}
		if !c.propUsed(p) {
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
			node: node, name: componentName(node), binding: componentBinding(node, resolveSymbol),
			used: map[string]bool{}, aliases: map[string][]string{},
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
						appendDeclared(c, typeProps(typ, aliases, map[string]bool{}, ""))
					}
					for _, member := range c.node.Members() {
						if member.Kind == ast.KindPropertyDeclaration {
							pd := member.AsPropertyDeclaration()
							name := propertyName(pd.Name())
							if name == "propTypes" && pd.Initializer != nil {
								appendDeclared(c, declaredProps(pd.Initializer, opts.customValidators, propWrappers))
							}
							if name == "props" && pd.Type != nil {
								appendDeclared(c, typeProps(pd.Type, aliases, map[string]bool{}, ""))
							}
						} else if (member.Kind == ast.KindGetAccessor || member.Kind == ast.KindMethodDeclaration) && ast.IsStatic(member) && propertyName(member.Name()) == "propTypes" {
							body := member.Body()
							if body != nil {
								body.ForEachChild(func(n *ast.Node) bool {
									if n.Kind == ast.KindReturnStatement && n.AsReturnStatement().Expression != nil {
										appendDeclared(c, declaredProps(n.AsReturnStatement().Expression, opts.customValidators, propWrappers))
									}
									return false
								})
							}
						}
					}
				case ast.KindObjectLiteralExpression:
					for _, member := range c.node.AsObjectLiteralExpression().Properties.Nodes {
						if propertyName(member.Name()) == "propTypes" && member.Kind == ast.KindPropertyAssignment {
							appendDeclared(c, declaredProps(member.AsPropertyAssignment().Initializer, opts.customValidators, propWrappers))
						}
					}
				}
				{
					// Function parameters are the root aliases for SFCs.
					if isFunctionLike(c.node) {
						for _, parameter := range reactutil.FunctionParameters(c.node) {
							if parameter.Kind != ast.KindParameter {
								continue
							}
							decl := parameter.AsParameterDeclaration()
							bindingPath(decl.Name(), nil, c)
							if decl.Type != nil {
								appendDeclared(c, typeProps(decl.Type, aliases, map[string]bool{}, ""))
							}
						}
						if typ := functionComponentType(c.node); typ != nil {
							appendDeclared(c, typeProps(typ, aliases, map[string]bool{}, ""))
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
