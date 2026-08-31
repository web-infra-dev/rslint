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
	destructured  map[string]string
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

func propMap(n *ast.Node) (map[string]propType, bool) {
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
			out[k] = validatorType(value)
		}
		return out, true
	}
	return nil, false
}

func validatorType(n *ast.Node) propType {
	n = unwrap(n)
	if n == nil {
		return propType{any: true}
	}
	if n.Kind == ast.KindPropertyAccessExpression && keyName(n.AsPropertyAccessExpression().Name()) == "isRequired" {
		n = unwrap(n.AsPropertyAccessExpression().Expression)
	}
	if n.Kind != ast.KindCallExpression {
		return propType{}
	}
	call := n.AsCallExpression()
	callee := unwrap(call.Expression)
	name := ""
	if callee != nil && callee.Kind == ast.KindPropertyAccessExpression {
		name = keyName(callee.AsPropertyAccessExpression().Name())
	}
	if name == "shape" || name == "exact" {
		if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
			if m, ok := propMap(call.Arguments.Nodes[0]); ok {
				return propType{children: m}
			}
		}
	}
	if name == "objectOf" || name == "arrayOf" {
		if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
			return propType{open: true, children: map[string]propType{"__ANY_KEY__": validatorType(call.Arguments.Nodes[0])}}
		}
	}
	if name == "oneOfType" && call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
		argument := unwrap(call.Arguments.Nodes[0])
		if argument != nil && argument.Kind == ast.KindArrayLiteralExpression {
			var union []propType
			for _, candidate := range argument.AsArrayLiteralExpression().Elements.Nodes {
				if candidate != nil && candidate.Kind != ast.KindOmittedExpression {
					union = append(union, validatorType(candidate))
				}
			}
			if len(union) > 0 {
				return propType{union: union}
			}
		}
	}
	return propType{}
}

func declared(n *ast.Node) (map[string]propType, bool) {
	m, ok := propMap(n)
	return m, ok
}

func className(n *ast.Node) string { return reactutil.BindingIdentifierName(n) }

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
	return n.Kind == ast.KindIdentifier && (n.AsIdentifier().Text == c.props || c.destructured[n.AsIdentifier().Text] != "")
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
					c := &component{node: n, declared: map[string]propType{}, destructured: map[string]string{}}
					comps = append(comps, c)
					if name := className(n); name != "" {
						byName[name] = c
					}
				}
			case ast.KindObjectLiteralExpression:
				if reactutil.IsCreateReactClassObjectArg(n, pragma, createClass) {
					c := &component{node: n, declared: map[string]propType{}, destructured: map[string]string{}}
					comps = append(comps, c)
				}
			case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration:
				if reactutil.IsStatelessReactComponentWithWrappers(n, pragma, ctx.TypeChecker, wrappers) {
					c := &component{node: n, declared: map[string]propType{}, destructured: map[string]string{}}
					comps = append(comps, c)
					if name := className(n); name != "" {
						byName[name] = c
					}
					ps := n.Parameters()
					if len(ps) > 0 && ps[0].Kind == ast.KindParameter {
						name := ps[0].AsParameterDeclaration().Name()
						if name != nil && name.Kind == ast.KindIdentifier {
							c.props = name.AsIdentifier().Text
						} else if name != nil && name.Kind == ast.KindObjectBindingPattern {
							for _, e := range name.AsBindingPattern().Elements.Nodes {
								if e.Kind == ast.KindBindingElement {
									be := e.AsBindingElement()
									if be.DotDotDotToken != nil {
										continue
									}
									k := keyName(be.PropertyName)
									if k == "" {
										k = keyName(be.Name())
									}
									if k != "" {
										c.destructured[keyName(be.Name())] = k
									}
								}
							}
						}
					}
				}
			case ast.KindPropertyAssignment:
				pa := n.AsPropertyAssignment()
				if keyName(pa.Name()) == "propTypes" {
					if c := componentFor(n.Parent, comps); c != nil {
						if m, ok := declared(pa.Initializer); ok {
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
						if m, ok := declared(pd.Initializer); ok {
							c.declared = m
							c.declaredBlock = true
						} else {
							c.declaredBlock = true
							c.declared = map[string]propType{"__ANY_KEY__": {any: true}}
						}
					}
				}
			case ast.KindBinaryExpression:
				b := n.AsBinaryExpression()
				if b.OperatorToken != nil && b.OperatorToken.Kind == ast.KindEqualsToken && b.Left != nil && b.Left.Kind == ast.KindPropertyAccessExpression {
					pa := b.Left.AsPropertyAccessExpression()
					if keyName(pa.Name()) == "propTypes" {
						if id := unwrap(pa.Expression); id != nil && id.Kind == ast.KindIdentifier {
							if c := byName[id.AsIdentifier().Text]; c != nil {
								if m, ok := declared(b.Right); ok {
									c.declared = m
									c.declaredBlock = true
								} else {
									c.declaredBlock = true
									c.declared = map[string]propType{"__ANY_KEY__": {any: true}}
								}
							}
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
					if root.Kind == ast.KindThisKeyword && len(names) > 0 && names[0] == "props" {
						appendUse(c, n, names[1:])
					} else if root.Kind == ast.KindIdentifier {
						if key, ok := c.destructured[root.AsIdentifier().Text]; ok {
							names = append([]string{key}, names...)
						}
						appendUse(c, n, names)
					}
				}
			case ast.KindIdentifier:
				c := componentFor(n, comps)
				if c == nil {
					break
				}
				memberBase := n.Parent != nil && (n.Parent.Kind == ast.KindPropertyAccessExpression && n.Parent.AsPropertyAccessExpression().Expression == n || n.Parent.Kind == ast.KindElementAccessExpression && n.Parent.AsElementAccessExpression().Expression == n)
				if k, ok := c.destructured[n.AsIdentifier().Text]; ok && n != c.node && !ast.IsPartOfParameterDeclaration(n) && !memberBase {
					appendUse(c, n, []string{k})
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
