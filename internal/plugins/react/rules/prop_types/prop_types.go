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
	children  map[string]propType
	union     []propType
	any, open bool
}
type use struct {
	node  *ast.Node
	names []string
}
type component struct {
	node          *ast.Node
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
		ast.KindBigIntLiteral,
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
				out["__ANY_KEY__"] = propType{any: true}
				continue
			}
			var name, value *ast.Node
			switch p.Kind {
			case ast.KindPropertyAssignment:
				pa := p.AsPropertyAssignment()
				name, value = pa.Name(), pa.Initializer
			case ast.KindShorthandPropertyAssignment:
				name, value = p.AsShorthandPropertyAssignment().Name(), p.AsShorthandPropertyAssignment().Name()
			default:
				continue
			}
			k := keyName(name)
			if k == "" {
				out["__ANY_KEY__"] = propType{any: true}
				continue
			}
			out[k] = validatorType(value, customValidators)
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
			return propType{open: true}
		}
		seen[n] = true
		if initializer := reactutil.ResolveIdentifierInitializer(n, nil); initializer != nil && initializer != n {
			return validatorTypeSeen(initializer, customValidators, seen)
		}
		return propType{open: true}
	}
	if n.Kind == ast.KindPropertyAccessExpression && keyName(n.AsPropertyAccessExpression().Name()) == "isRequired" {
		n = unwrap(n.AsPropertyAccessExpression().Expression)
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
			if m, ok := propMap(call.Arguments.Nodes[0], customValidators); ok {
				return propType{children: m}
			}
		}
	}
	if name == "objectOf" || name == "arrayOf" {
		if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
			return propType{open: true, children: map[string]propType{"__ANY_KEY__": validatorTypeSeen(call.Arguments.Nodes[0], customValidators, seen)}}
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
	return propType{}
}

func declared(n *ast.Node, customValidators []string) (map[string]propType, bool) {
	m, ok := propMap(n, customValidators)
	return m, ok
}

func className(n *ast.Node) string { return reactutil.BindingIdentifierName(n) }

func componentName(n *ast.Node) string {
	if name := className(n); name != "" {
		return name
	}
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
					return name.AsIdentifier().Text
				}
			}
		}
		return ""
	}
	return ""
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

func addDestructured(c *component, pattern *ast.Node, prefix []string) {
	if c == nil || pattern == nil || pattern.Kind != ast.KindObjectBindingPattern {
		return
	}
	for _, e := range pattern.AsBindingPattern().Elements.Nodes {
		if e == nil || e.Kind != ast.KindBindingElement {
			continue
		}
		be := e.AsBindingElement()
		if be.DotDotDotToken != nil {
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
		if be.Name().Kind == ast.KindIdentifier {
			c.destructured[be.Name().AsIdentifier().Text] = path
		} else if be.Name().Kind == ast.KindObjectBindingPattern {
			addDestructured(c, be.Name(), path)
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
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindGetAccessor, ast.KindSetAccessor:
			return false
		}
	}
	return false
}

func functionFirstParameterName(fn *ast.Node) string {
	params := reactutil.FunctionParameters(fn)
	if len(params) == 0 || params[0] == nil || params[0].Kind != ast.KindParameter {
		return ""
	}
	name := params[0].AsParameterDeclaration().Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.AsIdentifier().Text
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
			report = cur
			cur = unwrap(pa.Expression)
		case ast.KindElementAccessExpression:
			ea := cur.AsElementAccessExpression()
			k := elementName(ea.ArgumentExpression)
			if k == "" {
				return nil, nil, false
			}
			names = append([]string{k}, names...)
			report = cur
			cur = unwrap(ea.Expression)
		default:
			return cur, names, true
		}
	}
	return report, names, false
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

var PropTypesRule = rule.Rule{
	Name: "react/prop-types", Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, raw []any) rule.RuleListeners {
		o := parseOptions(raw)
		pragma := reactutil.GetReactPragma(ctx.Settings)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)
		wrappers := reactutil.GetComponentWrapperFunctions(ctx.Settings, pragma)
		var comps []*component
		byName := map[string]*component{}
		var walk ast.Visitor
		walk = func(n *ast.Node) bool {
			if n == nil {
				return false
			}
			switch n.Kind {
			case ast.KindClassDeclaration, ast.KindClassExpression:
				if reactutil.ExtendsReactComponent(n, pragma) {
					c := &component{node: n, declared: map[string]propType{}, destructured: map[string][]string{}}
					comps = append(comps, c)
					if name := componentName(n); name != "" {
						byName[name] = c
					}
				}
			case ast.KindObjectLiteralExpression:
				if reactutil.IsCreateReactClassObjectArg(n, pragma, createClass) {
					c := &component{node: n, declared: map[string]propType{}, destructured: map[string][]string{}}
					comps = append(comps, c)
				}
			case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration:
				if reactutil.IsStatelessReactComponentWithWrappers(n, pragma, ctx.TypeChecker, wrappers) {
					c := &component{node: n, declared: map[string]propType{}, destructured: map[string][]string{}}
					comps = append(comps, c)
					if name := componentName(n); name != "" {
						byName[name] = c
					}
					ps := n.Parameters()
					if len(ps) > 0 && ps[0].Kind == ast.KindParameter {
						name := ps[0].AsParameterDeclaration().Name()
						if name != nil && name.Kind == ast.KindIdentifier {
							c.props = name.AsIdentifier().Text
						} else if name != nil && name.Kind == ast.KindObjectBindingPattern {
							addDestructured(c, name, nil)
						}
					}
				}
			case ast.KindPropertyAssignment:
				pa := n.AsPropertyAssignment()
				if keyName(pa.Name()) == "propTypes" {
					if c := componentFor(n.Parent, comps); c != nil {
						if m, ok := declared(pa.Initializer, o.customValidators); ok {
							c.declared = m
							c.declaredBlock = true
						} else {
							c.declaredBlock = true
							c.declared = map[string]propType{"__ANY_KEY__": {any: true}}
						}
					}
				}
			case ast.KindPropertyDeclaration:
				pd := n.AsPropertyDeclaration()
				if keyName(pd.Name()) == "propTypes" {
					if c := componentFor(n.Parent, comps); c != nil {
						if m, ok := declared(pd.Initializer, o.customValidators); ok {
							c.declared = m
							c.declaredBlock = true
						} else {
							c.declaredBlock = true
							c.declared = map[string]propType{"__ANY_KEY__": {any: true}}
						}
					}
				}
			case ast.KindGetAccessor:
				ga := n.AsGetAccessorDeclaration()
				if ast.HasSyntacticModifier(n, ast.ModifierFlagsStatic) && keyName(ga.Name()) == "propTypes" {
					if c := componentFor(n.Parent, comps); c != nil {
						if m, ok := declared(getterReturn(ga.Body), o.customValidators); ok {
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
									addDestructured(c, vd.Name(), path)
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
					if ok && root != nil && root.Kind == ast.KindIdentifier && len(names) >= 1 && names[0] == "propTypes" {
						if c := byName[root.AsIdentifier().Text]; c != nil {
							if len(names) == 1 {
								if m, ok := declared(b.Right, o.customValidators); ok {
									c.declared = m
								} else {
									c.declared = map[string]propType{"__ANY_KEY__": {any: true}}
								}
							} else {
								setDeclared(c.declared, names[1:], validatorType(b.Right, o.customValidators))
							}
							c.declaredBlock = true
						}
					}
				}
			case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
				root, names, ok := memberNames(n)
				if !ok || root == nil {
					break
				}
				c := componentFor(n, comps)
				if c == nil || len(names) == 0 {
					break
				}
				if rootMatches(root, c) {
					if path, ok := propsPath(root, names, c); ok {
						appendUse(c, n, path)
					}
				}
			case ast.KindIdentifier:
				c := componentFor(n, comps)
				if c == nil {
					break
				}
				memberBase := n.Parent != nil && (n.Parent.Kind == ast.KindPropertyAccessExpression && n.Parent.AsPropertyAccessExpression().Expression == n || n.Parent.Kind == ast.KindElementAccessExpression && n.Parent.AsElementAccessExpression().Expression == n)
				if path, ok := c.destructured[n.AsIdentifier().Text]; ok && n != c.node && !ast.IsPartOfParameterDeclaration(n) && !memberBase {
					appendUse(c, n, path)
				}
			}
			n.ForEachChild(walk)
			return false
		}
		ctx.SourceFile.Node.ForEachChild(walk)
		for _, c := range comps {
			if len(c.used) == 0 || !c.declaredBlock && o.skipUndeclared {
				continue
			}
			seen := map[string]bool{}
			for _, u := range c.used {
				name := u.names[0]
				if slices.Contains(o.ignore, name) || seen[reportName(u.names)] {
					continue
				}
				seen[reportName(u.names)] = true
				if !isDeclared(c.declared, u.names) {
					ctx.ReportNode(u.node, rule.RuleMessage{Id: "missingPropType", Description: fmt.Sprintf("'%s' is missing in props validation", reportName(u.names)), Data: map[string]string{"name": reportName(u.names)}})
				}
			}
		}
		return rule.RuleListeners{}
	},
}
