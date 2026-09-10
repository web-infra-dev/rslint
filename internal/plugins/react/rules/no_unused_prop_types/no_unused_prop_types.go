package no_unused_prop_types

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
	children []*prop
}

type component struct {
	node     *ast.Node
	name     string
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
			case "shape", "exact", "custom":
				current.shape = method != "custom"
				if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
					current.children = propMap(call.Arguments.Nodes[0], customValidators, fullName)
				}
			case "arrayOf", "objectOf":
				if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
					current.children = validatorChildren(call.Arguments.Nodes[0], customValidators, fullName+".*")
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

func typeProps(node *ast.Node, aliases map[string]*ast.Node, seen map[string]bool, prefix string) []*prop {
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
			if member == nil || member.Kind != ast.KindPropertySignature {
				continue
			}
			name, reportNode := propKey(member.AsPropertySignatureDeclaration().Name())
			if name == "" {
				continue
			}
			fullName := name
			if prefix != "" {
				fullName = prefix + "." + name
			}
			p := newProp(name, fullName, reportNode)
			p.children = typeProps(member.AsPropertySignatureDeclaration().Type, aliases, seen, fullName)
			result = append(result, p)
		}
		return result
	case ast.KindTypeReference:
		name := reactutil.EntityNameRightmost(node.AsTypeReferenceNode().TypeName)
		if name == nil || name.Kind != ast.KindIdentifier || seen[name.AsIdentifier().Text] {
			return nil
		}
		seen[name.AsIdentifier().Text] = true
		return typeProps(aliases[name.AsIdentifier().Text], aliases, seen, prefix)
	case ast.KindIntersectionType:
		var result []*prop
		for _, part := range node.AsIntersectionTypeNode().Types.Nodes {
			result = mergeProps(result, typeProps(part, aliases, seen, prefix))
		}
		return result
	}
	return nil
}

func collectTypeAliases(root *ast.Node) map[string]*ast.Node {
	aliases := map[string]*ast.Node{}
	var walk ast.Visitor
	walk = func(node *ast.Node) bool {
		if node != nil && node.Kind == ast.KindTypeAliasDeclaration {
			decl := node.AsTypeAliasDeclaration()
			if name := decl.Name(); name != nil && name.Kind == ast.KindIdentifier {
				aliases[name.AsIdentifier().Text] = decl.Type
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

func appendDeclared(c *component, props []*prop) {
	for _, p := range props {
		if p != nil {
			c.declared = append(c.declared, p)
		}
	}
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

func findComponent(components []*component, name string) *component {
	for _, c := range components {
		if c.name == name {
			return c
		}
	}
	return nil
}

func addAssignmentDeclarations(root *ast.Node, components []*component, customValidators []string) {
	var walk ast.Visitor
	walk = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if node.Kind == ast.KindBinaryExpression {
			bin := node.AsBinaryExpression()
			if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken {
				parts := propertyAccessParts(bin.Left)
				if len(parts) >= 2 && (parts[len(parts)-1] == "propTypes" || (len(parts) >= 3 && parts[len(parts)-2] == "propTypes")) {
					c := findComponent(components, parts[0])
					if c != nil {
						if len(parts) == 2 {
							appendDeclared(c, propMap(bin.Right, customValidators, ""))
						} else {
							name := parts[len(parts)-1]
							// Direct `Component.propTypes.name = validator` assignments
							// are uncommon compared with the object form. Keep the
							// declaration itself even when the validator has no object
							// wrapper; nested shape data is collected from the object form.
							c.declared = append(c.declared, newProp(name, name, leftmostProperty(bin.Left)))
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

func leftmostProperty(node *ast.Node) *ast.Node {
	node = unwrap(node)
	if node != nil && node.Kind == ast.KindPropertyAccessExpression {
		return node.AsPropertyAccessExpression().Name()
	}
	return node
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
	return node != nil && node.Parent != nil && node.Parent.Kind == ast.KindVariableDeclaration && node.Parent.AsVariableDeclaration().Initializer == node
}

func (c *component) walkUsage(node *ast.Node, nested map[*ast.Node]bool) {
	if node == nil {
		return
	}
	if node != c.node && nested[node] {
		return
	}
	if node.Kind == ast.KindVariableDeclaration {
		vd := node.AsVariableDeclaration()
		if vd.Initializer != nil {
			if path, ok := pathFrom(vd.Initializer, c.aliases); ok {
				bindingPath(vd.Name(), path, c)
			}
		}
	}
	if node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression {
		isMemberBase := node.Parent != nil && (node.Parent.Kind == ast.KindPropertyAccessExpression || node.Parent.Kind == ast.KindElementAccessExpression) &&
			((node.Parent.Kind == ast.KindPropertyAccessExpression && node.Parent.AsPropertyAccessExpression().Expression == node) ||
				(node.Parent.Kind == ast.KindElementAccessExpression && node.Parent.AsElementAccessExpression().Expression == node))
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
		c.walkUsage(child, nested)
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
	pragma := reactutil.GetReactPragma(ctx.Settings)
	createClass := reactutil.GetReactCreateClass(ctx.Settings)
	aliases := collectTypeAliases(ctx.SourceFile.AsNode())
	var components []*component
	nested := map[*ast.Node]bool{}

	var discover ast.Visitor
	discover = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if isReactClass(node, pragma) {
			c := &component{node: node, name: componentName(node), used: map[string]bool{}, aliases: map[string][]string{}}
			components = append(components, c)
			nested[node] = true
		}
		if node.Kind == ast.KindCallExpression && reactutil.IsCreateClassCall(node.AsCallExpression(), pragma, createClass) {
			call := node.AsCallExpression()
			if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
				arg := unwrap(call.Arguments.Nodes[0])
				if arg != nil && arg.Kind == ast.KindObjectLiteralExpression {
					c := &component{node: arg, name: componentName(arg), used: map[string]bool{}, aliases: map[string][]string{}}
					components = append(components, c)
					nested[arg] = true
				}
			}
		}
		if isFunctionLike(node) && reactutil.IsStatelessReactComponentWithChecker(node, pragma, ctx.TypeChecker) {
			c := &component{node: node, name: componentName(node), used: map[string]bool{}, aliases: map[string][]string{}}
			components = append(components, c)
			nested[node] = true
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
								appendDeclared(c, propMap(pd.Initializer, opts.customValidators, ""))
							}
							if name == "props" && pd.Type != nil {
								appendDeclared(c, typeProps(pd.Type, aliases, map[string]bool{}, ""))
							}
						} else if (member.Kind == ast.KindGetAccessor || member.Kind == ast.KindMethodDeclaration) && ast.IsStatic(member) && propertyName(member.Name()) == "propTypes" {
							body := member.Body()
							if body != nil {
								body.ForEachChild(func(n *ast.Node) bool {
									if n.Kind == ast.KindReturnStatement && n.AsReturnStatement().Expression != nil {
										appendDeclared(c, propMap(n.AsReturnStatement().Expression, opts.customValidators, ""))
									}
									return false
								})
							}
						}
					}
				case ast.KindObjectLiteralExpression:
					for _, member := range c.node.AsObjectLiteralExpression().Properties.Nodes {
						if propertyName(member.Name()) == "propTypes" && member.Kind == ast.KindPropertyAssignment {
							appendDeclared(c, propMap(member.AsPropertyAssignment().Initializer, opts.customValidators, ""))
						}
					}
				}
				if c.name != "" {
					// Function parameters are the root aliases for SFCs.
					if isFunctionLike(c.node) {
						for _, parameter := range reactutil.FunctionParameters(c.node) {
							if parameter.Kind != ast.KindParameter {
								continue
							}
							bindingPath(parameter.AsParameterDeclaration().Name(), nil, c)
						}
					}
					c.walkUsage(c.node, nested)
				}
			}
			addAssignmentDeclarations(ctx.SourceFile.AsNode(), components, opts.customValidators)
			for _, c := range components {
				if c.name == "" {
					continue
				}
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
