package default_props_match_prop_types

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	scopeAnalysis "github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
)

//go:embed default_props_match_prop_types.schema.json
var schemaJSON []byte

type component struct {
	node       *ast.Node
	props      map[string]bool
	defaults   map[string]*ast.Node
	order      []string
	openProps  bool
	unresolved bool
}

type componentKey struct {
	name    string
	binding *ast.Symbol
	scope   *ast.Node
}

type declaration struct {
	source, value, target         *ast.Node
	owner                         *component
	name                          string
	member                        string
	typed, individual, unresolved bool
}

type analysis struct {
	ctx      rule.RuleContext
	wrappers []reactutil.PropWrapperEntry
}

// NOTE: Unlike ESLint, static keys use their cooked value and unknown props
// remain opaque. TypeScript annotations follow bindings and do not depend on
// the parameter name. These differences prevent missed diagnostics and false
// positives; see the rule documentation and extras tests.
//
// The rule compares declarations, not the types of default values. Runtime
// validators only contribute their directly authored isRequired suffix.
func required(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || ast.IsOptionalChain(node) {
		return false
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		return node.Name() != nil && node.Name().Text() == "isRequired"
	}
	if node.Kind == ast.KindElementAccessExpression {
		key := utils.ESTreeRuntimeExpression(node.AsElementAccessExpression().ArgumentExpression)
		return key != nil && key.Kind == ast.KindIdentifier && key.Text() == "isRequired"
	}
	return false
}

func (a *analysis) initializer(node *ast.Node) *ast.Node {
	if node == nil || node.Kind != ast.KindIdentifier {
		return nil
	}
	symbol := a.ctx.Refs.Resolve(node)
	if symbol == nil {
		return nil
	}
	for i := len(symbol.Declarations) - 1; i >= 0; i-- {
		decl := symbol.Declarations[i]
		if decl.Kind == ast.KindVariableDeclaration {
			return decl.AsVariableDeclaration().Initializer
		}
	}
	return nil
}

func (a *analysis) object(node *ast.Node, defaults bool, seen map[*ast.Node]bool) *ast.Node {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || seen[node] {
		return nil
	}
	seen[node] = true
	if node.Kind == ast.KindIdentifier {
		// defaultProps resolves a variable's initializer once; propTypes recursively
		// follows aliases, matching the distinct upstream helper contracts.
		value := utils.ESTreeRuntimeExpression(a.initializer(node))
		if defaults {
			return value
		}
		return a.object(value, false, seen)
	}
	if node.Kind == ast.KindCallExpression && !ast.IsOptionalChain(node) {
		call := node.AsCallExpression()
		if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
			return nil
		}
		callee := utils.ESTreeRuntimeExpression(call.Expression)
		matched := false
		if defaults {
			if callee != nil && callee.Kind == ast.KindIdentifier {
				for _, wrapper := range a.wrappers {
					if (wrapper.FromString && wrapper.Raw == callee.Text()) || (!wrapper.FromString && wrapper.Property == callee.Text()) {
						matched = true
					}
				}
			}
		} else {
			matched = reactutil.IsPropWrapperCall(node, a.wrappers)
		}
		if matched {
			return a.object(call.Arguments.Nodes[0], defaults, seen)
		}
	}
	return node
}

func propertyValue(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindPropertyAssignment:
		return node.AsPropertyAssignment().Initializer
	case ast.KindShorthandPropertyAssignment:
		return node.Name()
	case ast.KindPropertyDeclaration:
		return node.AsPropertyDeclaration().Initializer
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return node
	}
	return nil
}

// propertyKey distinguishes an authored name from an unknown computed key.
func propertyKey(key *ast.Node) (string, bool) {
	if key == nil {
		return "", false
	}

	return utils.GetStaticPropertyName(key)
}

func (a *analysis) runtimeProps(node *ast.Node, out map[string]bool) bool {
	node = a.object(node, false, map[*ast.Node]bool{})
	if node == nil || node.Kind != ast.KindObjectLiteralExpression {
		return true
	}
	opaque := false
	for _, prop := range node.AsObjectLiteralExpression().Properties.Nodes {
		if prop.Kind == ast.KindSpreadAssignment {
			opaque = true
			continue
		}
		name, ok := propertyKey(prop.Name())
		if ok {
			out[name] = required(propertyValue(prop))
		} else {
			opaque = true
		}
	}
	return opaque
}

func (a *analysis) typeProps(node *ast.Node, out map[string]bool, seen map[*ast.Node]bool) bool {
	if node == nil || seen[node] {
		return true
	}
	seen[node] = true
	defer delete(seen, node)
	opaque := false
	var members *ast.TypeElementList
	switch node.Kind {
	case ast.KindTypeLiteral:
		members = node.AsTypeLiteralNode().Members
	case ast.KindInterfaceDeclaration:
		decl := node.AsInterfaceDeclaration()
		members = decl.Members
		if decl.HeritageClauses != nil {
			for _, clause := range decl.HeritageClauses.Nodes {
				if clause.AsHeritageClause().Types != nil {
					for _, typ := range clause.AsHeritageClause().Types.Nodes {
						opaque = a.typeProps(typ, out, seen) || opaque
					}
				}
			}
		}
	case ast.KindTypeAliasDeclaration:
		return a.typeProps(node.AsTypeAliasDeclaration().Type, out, seen)
	case ast.KindParenthesizedType:
		return a.typeProps(node.AsParenthesizedTypeNode().Type, out, seen)
	case ast.KindIntersectionType:
		for _, typ := range node.AsIntersectionTypeNode().Types.Nodes {
			opaque = a.typeProps(typ, out, seen) || opaque
		}
		return opaque
	case ast.KindTypeReference:
		ref := node.AsTypeReferenceNode()
		if props := reactutil.ReactGenericArgumentWithAmbient(ref.TypeName, ref.TypeArguments, a.ctx.Refs.Resolve); props != nil {
			opaque = a.typeProps(props, out, seen)
			if _, declared := out["children"]; !declared {
				out["children"] = false
			}
			return opaque
		}
		if ref.TypeName == nil || ref.TypeName.Kind != ast.KindIdentifier {
			return true
		}
		if ref.TypeName.Text() == "ReturnType" && ref.TypeArguments != nil && len(ref.TypeArguments.Nodes) == 1 {
			arg := ref.TypeArguments.Nodes[0]
			if arg.Kind == ast.KindFunctionType {
				return a.typeProps(arg.AsFunctionTypeNode().Type, out, seen)
			}
			if arg.Kind == ast.KindTypeQuery {
				fn := a.initializer(arg.AsTypeQueryNode().ExprName)
				if fn != nil && ast.IsFunctionLike(fn) {
					body := utils.ESTreeRuntimeExpression(fn.Body())
					if body != nil && body.Kind == ast.KindBlock {
						body = reactutil.LastReturnedExpressionInBody(body)
					}
					return a.runtimeProps(body, out)
				}
			}
			return true
		}
		return a.namedTypeProps(ref.TypeName, out, seen)
	case ast.KindExpressionWithTypeArguments:
		return a.namedTypeProps(node.AsExpressionWithTypeArguments().Expression, out, seen)
	default:
		return true
	}
	if members == nil {
		return opaque
	}
	for _, member := range members.Nodes {
		if member.Kind != ast.KindPropertySignature && member.Kind != ast.KindMethodSignature {
			opaque = true
			continue
		}
		name, ok := propertyKey(member.Name())
		if !ok {
			opaque = true
			continue
		}
		if member.Kind == ast.KindPropertySignature {
			out[name] = member.AsPropertySignatureDeclaration().QuestionToken() == nil
		} else {
			out[name] = member.AsMethodSignatureDeclaration().QuestionToken() == nil
		}
	}
	return opaque
}

func (a *analysis) namedTypeProps(name *ast.Node, out map[string]bool, seen map[*ast.Node]bool) bool {
	if name == nil || name.Kind != ast.KindIdentifier {
		return true
	}
	symbol := a.ctx.Refs.Resolve(name)
	if symbol == nil || !utils.IsSymbolDeclaredInFile(symbol, a.ctx.SourceFile) {
		return true
	}
	found, opaque := false, false
	for _, decl := range symbol.Declarations {
		if decl.Kind == ast.KindTypeAliasDeclaration || decl.Kind == ast.KindInterfaceDeclaration {
			found = true
			opaque = a.typeProps(decl, out, seen) || opaque
		}
	}
	return !found || opaque
}

func (a *analysis) defaults(c *component, d declaration) {
	if c.unresolved {
		return
	}
	add := func(name string, node *ast.Node) {
		if _, ok := c.defaults[name]; !ok {
			c.order = append(c.order, name)
		}
		c.defaults[name] = node
	}
	if d.individual {
		add(d.member, d.source)
		return
	}
	node := a.object(d.value, true, map[*ast.Node]bool{})
	if node == nil || node.Kind != ast.KindObjectLiteralExpression {
		c.unresolved = d.unresolved
		return
	}
	for _, prop := range node.AsObjectLiteralExpression().Properties.Nodes {
		if prop.Kind == ast.KindSpreadAssignment {
			c.unresolved = true
			return
		}
	}
	for _, prop := range node.AsObjectLiteralExpression().Properties.Nodes {
		name, ok := propertyKey(prop.Name())
		if !ok {
			c.unresolved = true
			return
		}
		add(name, prop)
	}
}

var DefaultPropsMatchPropTypesRule = rule.Rule{
	Name:   "react/default-props-match-prop-types",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		allowRequired := false
		if len(options) > 0 {
			if opts, ok := options[0].(map[string]any); ok {
				allowRequired, _ = opts["allowRequiredDefaults"].(bool)
			}
		}
		a := analysis{ctx: ctx, wrappers: reactutil.GetPropWrapperFunctions(ctx.Settings)}
		pragma := reactutil.GetReactPragma(ctx.Settings)
		wrappers := reactutil.GetComponentWrapperFunctions(ctx.Settings, pragma)
		scopes := scopeAnalysis.For(ctx)
		byNode := map[*ast.Node]*component{}
		byKey := map[componentKey]*component{}
		var components []*component
		var declarations []declaration
		keyFor := func(name string, binding *ast.Symbol, node *ast.Node) componentKey {
			key := componentKey{name: name, binding: binding}
			if binding == nil {
				key.scope = reactutil.ComponentScope(node)
			}
			return key
		}
		register := func(node *ast.Node) *component {
			c := &component{node: node, props: map[string]bool{}, defaults: map[string]*ast.Node{}}
			components = append(components, c)
			byNode[node] = c
			name := strings.Join(reactutil.ComponentTarget(node), ".")
			binding := reactutil.ComponentBinding(node, ctx.Refs.Resolve)
			if name != "" {
				byKey[keyFor(name, binding, node)] = c
			}
			return c
		}
		visitClass := func(node *ast.Node) {
			if !reactutil.ExtendsReactComponent(node, pragma) {
				return
			}
			c := register(node)
			if typ := reactutil.ClassPropsType(node); typ != nil {
				declarations = append(declarations, declaration{source: node, value: typ, owner: c, typed: true})
			}
		}
		visitFunction := func(node *ast.Node) {
			if reactutil.IsAsyncGeneratorFunction(node) || !reactutil.IsStatelessReactComponentWithWrappers(node, pragma, ctx.TypeChecker, wrappers, scopes) {
				return
			}
			c := register(node)
			params := utils.ESTreeParameters(node)
			if len(params) > 0 {
				typ, forwardRef, typedRef := reactutil.FunctionComponentTypeDetailsWithPragma(node, pragma, ctx.Refs.Resolve)
				if !typedRef {
					typ = params[0].AsParameterDeclaration().Type
					if typ == nil && !forwardRef {
						annotation := reactutil.ComponentTypeAnnotation(node)
						if reactutil.ReactFunctionComponentArgumentWithAmbient(annotation, ctx.Refs.Resolve) != nil {
							typ = annotation
						}
					}
				}
				if typ != nil {
					declarations = append(declarations, declaration{source: node, value: typ, owner: c, typed: true})
				}
			}
		}
		visitMember := func(node *ast.Node) {
			c := byNode[node.Parent]
			if c == nil {
				return
			}
			name := reactutil.StaticPropertyName(node.Name())
			if name == "props" && node.Kind == ast.KindPropertyDeclaration && node.AsPropertyDeclaration().Type != nil {
				if ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic) {
					return
				}
				declarations = append(declarations, declaration{source: node, value: node.AsPropertyDeclaration().Type, owner: c, typed: true})
				return
			}
			if name != "propTypes" && name != "defaultProps" && name != "getDefaultProps" {
				return
			}
			value := propertyValue(node)
			defaults := name != "propTypes"
			if node.Kind == ast.KindGetAccessor {
				if !ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic) {
					return
				}
				if defaults {
					value = reactutil.LastReturnedExpressionInBody(node.Body())
				} else {
					value, _ = reactutil.LastDirectReturnedExpression(node.Body())
				}
			} else if node.Kind == ast.KindPropertyDeclaration {
				if !ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic) {
					return
				}
			} else if defaults {
				if c.node.Kind != ast.KindObjectLiteralExpression || value == nil || (value.Kind != ast.KindFunctionExpression && value.Kind != ast.KindMethodDeclaration) {
					return
				}
				value = reactutil.LastReturnedExpressionInBody(value.Body())
				value = utils.ESTreeRuntimeExpression(value)
				if value == nil || value.Kind != ast.KindObjectLiteralExpression {
					return
				}
			}
			declarations = append(declarations, declaration{source: node, value: value, owner: c, name: name})
		}
		return rule.RuleListeners{
			ast.KindClassDeclaration: visitClass, ast.KindClassExpression: visitClass,
			ast.KindFunctionDeclaration: visitFunction, ast.KindFunctionExpression: visitFunction, ast.KindArrowFunction: visitFunction,
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				if reactutil.IsCreateReactClassObjectArg(node, pragma, reactutil.GetReactCreateClass(ctx.Settings)) {
					register(node)
				}
			},
			ast.KindPropertyAssignment: visitMember, ast.KindPropertyDeclaration: visitMember, ast.KindGetAccessor: visitMember,
			ast.KindMethodDeclaration: func(node *ast.Node) { visitFunction(node); visitMember(node) },
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary.OperatorToken == nil || !ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
					return
				}
				root, names, ok := reactutil.StaticPropMemberNames(binary.Left)
				if !ok || root == nil || root.Kind != ast.KindIdentifier {
					return
				}
				for i, name := range names {
					if name == "__COMPUTED_PROP__" {
						return
					}
					if name != "propTypes" && name != "defaultProps" && name != "getDefaultProps" {
						continue
					}
					if len(names) > i+2 {
						return
					}
					d := declaration{source: node, value: binary.Right, target: root, name: name, unresolved: true}
					d.member = strings.Join(append([]string{root.Text()}, names[:i]...), ".")
					declarations = append(declarations, d)
					return
				}
			},
			ast.KindEndOfFile: func(_ *ast.Node) {
				slices.SortStableFunc(declarations, func(x, y declaration) int { return x.source.Pos() - y.source.Pos() })
				for _, d := range declarations {
					c := d.owner
					if c == nil {
						c = byKey[keyFor(d.member, ctx.Refs.Resolve(d.target), d.target)]
						if c == nil {
							continue
						}
						_, names, _ := reactutil.StaticPropMemberNames(d.source.AsBinaryExpression().Left)
						i := slices.Index(names, d.name)
						if len(names) == i+2 {
							d.individual = true
							d.member = names[i+1]
							if d.member == "__COMPUTED_PROP__" {
								c.unresolved = true
								continue
							}
						}
					}
					if d.typed {
						c.openProps = a.typeProps(d.value, c.props, map[*ast.Node]bool{}) || c.openProps
						continue
					}
					if d.name == "propTypes" {
						if d.individual {
							c.props[d.member] = required(d.value)
						} else {
							c.openProps = a.runtimeProps(d.value, c.props) || c.openProps
						}
					} else {
						a.defaults(c, d)
					}
				}
				for _, c := range components {
					if c.unresolved || len(c.props) == 0 {
						continue
					}
					for _, name := range c.order {
						req, exists := c.props[name]
						if (!exists && c.openProps) || exists && (allowRequired || !req) {
							continue
						}
						id, description := "defaultHasNoType", fmt.Sprintf("defaultProp \"%s\" has no corresponding propTypes declaration.", name)
						if exists {
							id = "requiredHasDefault"
							description = fmt.Sprintf("defaultProp \"%s\" defined for isRequired propType.", name)
						}
						ctx.ReportNode(c.defaults[name], rule.RuleMessage{Id: id, Description: description, Data: map[string]string{"name": name}})
					}
				}
			},
		}
	},
}
