package require_default_props

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

//go:embed require_default_props.schema.json
var schemaJSON []byte

type options struct {
	classes, functions       string
	forbidDefaultForRequired bool
}

func parseOptions(raw []any) options {
	o := options{classes: "defaultProps", functions: "defaultProps"}
	if len(raw) == 0 {
		return o
	}
	m, _ := raw[0].(map[string]any)
	if v, ok := m["classes"].(string); ok {
		o.classes = v
	}
	if v, ok := m["functions"].(string); ok {
		o.functions = v
	}
	if v, _ := m["ignoreFunctionalComponents"].(bool); v {
		o.functions = "ignore"
	}
	o.forbidDefaultForRequired, _ = m["forbidDefaultForRequired"].(bool)
	return o
}

type prop struct {
	node     *ast.Node
	required bool
}

type component struct {
	node                              *ast.Node
	fn                                *ast.Node
	binding                           *ast.Symbol
	target                            string
	props                             map[string]prop
	defaults                          map[string]bool
	defaultSource                     *ast.Node
	declared, hasDefaults, unresolved bool
}

type componentKey struct {
	binding *ast.Symbol
	target  string
}

type diagnosticKey struct {
	component  componentKey
	contract   *ast.Node
	fallback   *ast.Node
	id         string
	name       string
	occurrence int
}

type analyzer struct {
	ctx         rule.RuleContext
	lookup      *reactutil.VariableDefinitionLookup
	wrappers    []reactutil.PropWrapperEntry
	components  []*component
	byNode      map[*ast.Node]*component
	assignments []*ast.Node
	byTarget    map[componentKey][]*component
	reported    map[diagnosticKey]struct{}
}

func propertyName(n *ast.Node) string {
	if n == nil {
		return ""
	}
	name, _ := utils.GetStaticPropertyName(n)
	return name
}

// Required runtime props are identified syntactically upstream: aliases of .isRequired do not
// change it, and a string-literal computed member has no property.name.
func isRequired(n *ast.Node) bool {
	n = ast.SkipParentheses(n)
	if n == nil || ast.IsOptionalChain(n) {
		return false
	}
	if n.Kind == ast.KindPropertyAccessExpression {
		return propertyName(n.Name()) == "isRequired"
	}
	if n.Kind == ast.KindElementAccessExpression {
		key := ast.SkipParentheses(n.AsElementAccessExpression().ArgumentExpression)
		return key != nil && key.Kind == ast.KindIdentifier && key.Text() == "isRequired"
	}
	return false
}

func (a *analyzer) resolveObject(n *ast.Node, defaults bool) *ast.Node {
	seen := map[*ast.Node]bool{}
	for n != nil {
		n = ast.SkipParentheses(n)
		if seen[n] {
			return nil
		}
		seen[n] = true
		switch n.Kind {
		case ast.KindIdentifier:
			declaration := a.lookup.First(n, n.Text())
			if declaration == nil || declaration.Kind != ast.KindVariableDeclaration {
				return nil
			}
			n = declaration.AsVariableDeclaration().Initializer
		case ast.KindCallExpression:
			call := n.AsCallExpression()
			callee := ast.SkipParentheses(call.Expression)
			matches := false
			// defaultProps.js compares only the bare callee name with string settings.
			if defaults {
				if callee != nil && callee.Kind == ast.KindIdentifier {
					for _, w := range a.wrappers {
						if w.FromString && w.Raw == callee.Text() {
							matches = true
						}
					}
				}
			} else {
				matches = reactutil.IsPropWrapperCall(n, a.wrappers)
			}
			if !matches || call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
				return nil
			}
			n = call.Arguments.Nodes[0]
		case ast.KindObjectLiteralExpression:
			return n
		default:
			return nil
		}
	}
	return nil
}

func (a *analyzer) addProps(c *component, n *ast.Node) {
	c.declared = true
	n = a.resolveObject(n, false)
	if n == nil {
		return
	}
	for _, member := range n.AsObjectLiteralExpression().Properties.Nodes {
		var key, value *ast.Node
		switch member.Kind {
		case ast.KindPropertyAssignment:
			key, value = member.Name(), member.AsPropertyAssignment().Initializer
		case ast.KindShorthandPropertyAssignment:
			key, value = member.Name(), member.Name()
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			key, value = member.Name(), member
		default:
			continue
		}
		name, _ := reactutil.DeclaredPropKey(key)
		if key != nil && (name != "" || key.Kind == ast.KindStringLiteral) {
			c.props[name] = prop{node: member, required: isRequired(value)}
		}
	}
}

func (a *analyzer) addDefaults(c *component, n *ast.Node, external bool) {
	if c.unresolved {
		return
	}
	n = a.resolveObject(n, true)
	if n == nil {
		if external {
			c.unresolved = true
		}
		return
	}
	c.hasDefaults = true
	c.defaultSource = n
	for _, member := range n.AsObjectLiteralExpression().Properties.Nodes {
		if member.Kind == ast.KindSpreadAssignment {
			c.unresolved = true
			return
		}
		key := member.Name()
		if key == nil {
			continue
		}
		if key.Kind == ast.KindComputedPropertyName {
			key = key.AsComputedPropertyName().Expression
		}
		// Upstream uses raw key text, stripping only surrounding quotes.
		name := utils.TrimmedNodeText(a.ctx.SourceFile, key)
		name = strings.TrimPrefix(strings.TrimPrefix(name, "\""), "'")
		name = strings.TrimSuffix(strings.TrimSuffix(name, "\""), "'")
		c.defaults[name] = true
	}
}

func declarationName(n *ast.Node) string {
	if n != nil && n.Kind == ast.KindComputedPropertyName {
		n = ast.SkipParentheses(n.AsComputedPropertyName().Expression)
	}
	if n != nil && n.Kind == ast.KindIdentifier {
		return n.Text()
	}
	return ""
}

func defaultReturn(body *ast.Node) *ast.Node {
	if body == nil || body.Kind != ast.KindBlock {
		return nil
	}
	if body.AsBlock().Statements == nil {
		return nil
	}
	return reactutil.LastReturnedExpression(body.AsBlock().Statements.Nodes)
}

func lastReturn(body *ast.Node) *ast.Node {
	if body == nil || body.Kind != ast.KindBlock {
		return nil
	}
	statements := body.AsBlock().Statements
	if statements == nil {
		return nil
	}
	for i := len(statements.Nodes) - 1; i >= 0; i-- {
		if n := statements.Nodes[i]; n.Kind == ast.KindReturnStatement {
			return n.AsReturnStatement().Expression
		}
	}
	return nil
}

// Only locally authored types are expanded. Unlike upstream, lexical resolution
// also finds types declared inside functions and accepts inline class props types.
// Memoizing completed type nodes avoids re-expanding diamond-shaped aliases.
func (a *analyzer) typeProps(n *ast.Node, visiting map[*ast.Node]bool, memo map[*ast.Node]map[string]prop) map[string]prop {
	if n == nil {
		return nil
	}
	if result, ok := memo[n]; ok {
		return result
	}
	if visiting[n] {
		return nil
	}
	visiting[n] = true
	defer delete(visiting, n)
	result := map[string]prop{}
	merge := func(child *ast.Node) {
		for name, p := range a.typeProps(child, visiting, memo) {
			result[name] = p
		}
	}
	var members *ast.NodeList
	switch n.Kind {
	case ast.KindTypeLiteral:
		members = n.AsTypeLiteralNode().Members
	case ast.KindInterfaceDeclaration:
		declaration := n.AsInterfaceDeclaration()
		if declaration.HeritageClauses != nil {
			for _, clause := range declaration.HeritageClauses.Nodes {
				if clause.AsHeritageClause().Types != nil {
					for _, typ := range clause.AsHeritageClause().Types.Nodes {
						merge(typ)
					}
				}
			}
		}
		members = declaration.Members
	case ast.KindTypeAliasDeclaration:
		merge(n.AsTypeAliasDeclaration().Type)
	case ast.KindParenthesizedType:
		merge(n.AsParenthesizedTypeNode().Type)
	case ast.KindIntersectionType:
		for _, typ := range n.AsIntersectionTypeNode().Types.Nodes {
			merge(typ)
		}
	case ast.KindExpressionWithTypeArguments:
		merge(n.AsExpressionWithTypeArguments().Expression)
	case ast.KindTypeReference, ast.KindIdentifier:
		name := n
		if n.Kind == ast.KindTypeReference {
			reference := n.AsTypeReferenceNode()
			name = reference.TypeName
			if argument := reactutil.ReactGenericArgument(name, reference.TypeArguments, a.ctx.Refs.Resolve); argument != nil {
				result["children"] = prop{}
				merge(argument)
				break
			}
			if name.Kind == ast.KindIdentifier && name.Text() == "ReturnType" && reference.TypeArguments != nil && len(reference.TypeArguments.Nodes) == 1 {
				argument := reference.TypeArguments.Nodes[0]
				if argument.Kind == ast.KindFunctionType {
					merge(argument.AsFunctionTypeNode().Type)
				}
				if argument.Kind == ast.KindTypeQuery {
					query := argument.AsTypeQueryNode().ExprName
					if query.Kind == ast.KindIdentifier {
						declaration := a.lookup.First(query, query.Text())
						if declaration != nil && declaration.Kind == ast.KindVariableDeclaration {
							fn := ast.SkipParentheses(declaration.AsVariableDeclaration().Initializer)
							if fn != nil && reactutil.IsFunctionLikeForComponent(fn) {
								value := fn.Body()
								if value != nil && value.Kind == ast.KindBlock {
									value = defaultReturn(value)
								}
								value = ast.SkipParentheses(value)
								if value != nil && value.Kind == ast.KindObjectLiteralExpression {
									temporary := &component{props: map[string]prop{}}
									a.addProps(temporary, value)
									for key, p := range temporary.props {
										result[key] = p
									}
								} else if value != nil && value.Kind == ast.KindCallExpression && value.AsCallExpression().TypeArguments != nil {
									for _, typ := range value.AsCallExpression().TypeArguments.Nodes {
										merge(typ)
									}
								}
							}
						}
					}
				}
				break
			}
		}
		if name.Kind == ast.KindIdentifier {
			symbol := a.ctx.Refs.Resolve(name)
			if symbol != nil {
				for _, declaration := range symbol.Declarations {
					if ast.GetSourceFileOfNode(declaration) == a.ctx.SourceFile && (declaration.Kind == ast.KindTypeAliasDeclaration || declaration.Kind == ast.KindInterfaceDeclaration) {
						merge(declaration)
					}
				}
			}
		}
	}
	if members != nil {
		for _, member := range members.Nodes {
			var required bool
			switch member.Kind {
			case ast.KindPropertySignature:
				required = member.AsPropertySignatureDeclaration().QuestionToken() == nil
			case ast.KindMethodSignature:
				required = member.AsMethodSignatureDeclaration().QuestionToken() == nil
			default:
				continue
			}
			name, key := reactutil.DeclaredPropKey(member.Name())
			if key != nil && (name != "" || key.Kind == ast.KindStringLiteral) {
				result[name] = prop{node: member, required: required}
			}
		}
	}
	memo[n] = result
	return result
}

func (a *analyzer) addType(c *component, typ *ast.Node) {
	if typ == nil {
		return
	}
	c.declared = true
	for name, p := range a.typeProps(typ, map[*ast.Node]bool{}, map[*ast.Node]map[string]prop{}) {
		c.props[name] = p
	}
}

func (a *analyzer) initialize(c *component) {
	n := c.node
	switch n.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression:
		a.addType(c, reactutil.ClassPropsType(n))
		for _, member := range n.Members() {
			name := declarationName(member.Name())
			if member.Kind == ast.KindPropertyDeclaration {
				declaration := member.AsPropertyDeclaration()
				if name == "props" {
					a.addType(c, declaration.Type)
				}
				if name == "propTypes" {
					a.addProps(c, declaration.Initializer)
				}
				if ast.IsStatic(member) && (name == "defaultProps" || name == "getDefaultProps") {
					a.addDefaults(c, declaration.Initializer, false)
				}
			} else if member.Kind == ast.KindGetAccessor && ast.IsStatic(member) {
				value := lastReturn(member.Body())
				if name == "propTypes" && value != nil {
					a.addProps(c, value)
				}
				if name == "defaultProps" || name == "getDefaultProps" {
					a.addDefaults(c, defaultReturn(member.Body()), false)
				}
			}
		}
	case ast.KindObjectLiteralExpression:
		for _, member := range n.AsObjectLiteralExpression().Properties.Nodes {
			name := declarationName(member.Name())
			if member.Kind == ast.KindPropertyAssignment {
				value := ast.SkipParentheses(member.AsPropertyAssignment().Initializer)
				if name == "propTypes" {
					a.addProps(c, value)
				}
				if (name == "defaultProps" || name == "getDefaultProps") && value != nil && value.Kind == ast.KindFunctionExpression {
					if returned := ast.SkipParentheses(defaultReturn(value.Body())); returned != nil && returned.Kind == ast.KindObjectLiteralExpression {
						a.addDefaults(c, returned, false)
					}
				}
			} else if member.Kind == ast.KindMethodDeclaration && (name == "defaultProps" || name == "getDefaultProps") {
				if returned := ast.SkipParentheses(defaultReturn(member.Body())); returned != nil && returned.Kind == ast.KindObjectLiteralExpression {
					a.addDefaults(c, returned, false)
				}
			}
		}
	}
	if c.fn != nil {
		typ := reactutil.FunctionComponentType(c.fn, a.ctx.Refs.Resolve)
		params := utils.ESTreeParameters(c.fn)
		if len(params) > 0 && params[0].AsParameterDeclaration().Type != nil {
			typ = params[0].AsParameterDeclaration().Type
		}
		a.addType(c, typ)
	}
}

func (a *analyzer) applyAssignment(n *ast.Node) {
	bin := n.AsBinaryExpression()
	// Unlike upstream, only assignment targets contribute default declarations;
	// reading C.defaultProps on the right does not disable prop checks.
	left := ast.SkipParentheses(bin.Left)
	// Upstream reads MemberExpression.property.name, including computed
	// identifiers, but does not unwrap authored TypeScript assertion receivers.
	var parts []string
	root := left
	for root != nil {
		root = ast.SkipParentheses(root)
		if root.Kind == ast.KindIdentifier {
			parts = append(parts, root.Text())
			break
		}
		var key *ast.Node
		switch root.Kind {
		case ast.KindPropertyAccessExpression:
			key = root.Name()
			root = root.AsPropertyAccessExpression().Expression
		case ast.KindElementAccessExpression:
			key = ast.SkipParentheses(root.AsElementAccessExpression().ArgumentExpression)
			root = root.AsElementAccessExpression().Expression
		default:
			return
		}
		if key == nil || key.Kind != ast.KindIdentifier {
			return
		}
		parts = append(parts, key.Text())
	}
	slices.Reverse(parts)
	if root == nil || root.Kind != ast.KindIdentifier {
		return
	}
	binding := a.ctx.Refs.Resolve(root)
	if binding == nil {
		return
	}

	// A marker can also be part of the component namespace, as in
	// `ns.propTypes.C.propTypes`. Match only markers at the declaration end
	// whose preceding path identifies a detected component. Iterating all
	// matches preserves the rare case where both paths identify components.
	for i, part := range parts {
		if i == 0 || i < len(parts)-2 ||
			(part != "propTypes" && part != "defaultProps" && part != "getDefaultProps") {
			continue
		}
		target := strings.Join(parts[:i], ".")
		for _, c := range a.byTarget[componentKey{binding: binding, target: target}] {
			if i == len(parts)-1 {
				if parts[i] == "propTypes" {
					a.addProps(c, bin.Right)
				} else {
					a.addDefaults(c, bin.Right, true)
				}
			} else {
				name := parts[i+1]
				if parts[i] == "propTypes" {
					c.declared = true
					c.props[name] = prop{node: n, required: isRequired(bin.Right)}
				} else {
					c.hasDefaults = true
					c.defaultSource = n
					c.defaults[name] = true
				}
			}
		}
	}
}

func (a *analyzer) report(c *component, n, contract *ast.Node, occurrence int, id, name string) {
	// Component detection may retain multiple component-producing assignments
	// for one binding and property path. Their external prop/default declarations
	// share one contract node, so report that contract once while keeping distinct
	// inline contracts and repeated binding elements independent.
	key := diagnosticKey{
		component:  componentKey{binding: c.binding, target: c.target},
		contract:   contract,
		id:         id,
		name:       name,
		occurrence: occurrence,
	}
	if contract == nil {
		key.fallback = c.node
	}
	if _, exists := a.reported[key]; exists {
		return
	}
	a.reported[key] = struct{}{}

	descriptions := map[string]string{
		"noDefaultWithRequired":      `propType "%s" is required and should not have a defaultProps declaration.`,
		"shouldHaveDefault":          `propType "%s" is not required, but has no corresponding defaultProps declaration.`,
		"noDefaultPropsWithFunction": "Don’t use defaultProps with function components.",
		"shouldAssignObjectDefault":  `propType "%s" is not required, but has no corresponding default argument value.`,
		"destructureInSignature":     "Must destructure props in the function signature to initialize an optional prop.",
	}
	message := rule.RuleMessage{Id: id, Description: descriptions[id]}
	if strings.Contains(message.Description, "%s") {
		message.Description = fmt.Sprintf(message.Description, name)
		message.Data = map[string]string{"name": name}
	}
	a.ctx.ReportNode(n, message)
}

func (c *component) check(a *analyzer, opts options) {
	// NOTE: Unlike ESLint v7.37.5, wrappers retain their function identity so
	// functions: ignore/defaultArguments also applies to memo and forwardRef.
	function := c.fn != nil
	class := c.node.Kind == ast.KindClassDeclaration || c.node.Kind == ast.KindClassExpression
	if c.unresolved || !c.declared || function && opts.functions == "ignore" || class && opts.classes == "ignore" {
		return
	}
	if function && opts.functions == "defaultArguments" {
		if c.hasDefaults {
			a.report(c, c.node, c.defaultSource, 0, "noDefaultPropsWithFunction", "")
		}
		params := utils.ESTreeParameters(c.fn)
		if len(params) == 0 {
			return
		}
		param := params[0].AsParameterDeclaration()
		if param.Initializer != nil || param.DotDotDotToken != nil {
			return
		}
		name := param.Name()
		if name == nil {
			return
		}
		if name.Kind == ast.KindIdentifier {
			optionalNames := make([]string, 0, len(c.props))
			for propName, p := range c.props {
				if !p.required {
					optionalNames = append(optionalNames, propName)
				}
			}
			if len(optionalNames) > 0 {
				slices.Sort(optionalNames)
				contract := c.props[optionalNames[0]].node
				a.report(c, params[0], contract, 0, "destructureInSignature", "")
			}
		} else if name.Kind == ast.KindObjectBindingPattern {
			occurrences := map[string]int{}
			for _, element := range name.AsBindingPattern().Elements.Nodes {
				if element.Kind != ast.KindBindingElement {
					continue
				}
				binding := element.AsBindingElement()
				if binding.DotDotDotToken != nil {
					continue
				}
				key := binding.PropertyName
				if key == nil {
					key = binding.Name()
				}
				if key != nil && key.Kind == ast.KindComputedPropertyName {
					key = ast.SkipParentheses(key.AsComputedPropertyName().Expression)
				}
				if key == nil || key.Kind != ast.KindIdentifier {
					continue
				}
				p, exists := c.props[key.Text()]
				if !exists {
					continue
				}
				occurrence := occurrences[key.Text()]
				occurrences[key.Text()] = occurrence + 1
				if p.required && binding.Initializer != nil {
					a.report(c, element, p.node, occurrence, "noDefaultWithRequired", key.Text())
				}
				if !p.required && binding.Initializer == nil {
					a.report(c, element, p.node, occurrence, "shouldAssignObjectDefault", key.Text())
				}
			}
		}
		return
	}
	// Sort by declaration position to keep diagnostics deterministic across maps.
	names := make([]string, 0, len(c.props))
	for name := range c.props {
		names = append(names, name)
	}
	slices.SortFunc(names, func(x, y string) int {
		if c.props[x].node == nil {
			if c.props[y].node == nil {
				return strings.Compare(x, y)
			}
			return -1
		}
		if c.props[y].node == nil {
			return 1
		}
		return c.props[x].node.Pos() - c.props[y].node.Pos()
	})
	for _, name := range names {
		p := c.props[name]
		if p.node == nil {
			continue
		}
		if p.required {
			if opts.forbidDefaultForRequired && c.defaults[name] {
				a.report(c, p.node, p.node, 0, "noDefaultWithRequired", name)
			}
			continue
		}
		// NOTE: Unlike ESLint, inherited Object.prototype keys are not defaults.
		if !c.defaults[name] {
			a.report(c, p.node, p.node, 0, "shouldHaveDefault", name)
		}
	}
}

var RequireDefaultPropsRule = rule.Rule{
	Name:   "react/require-default-props",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, raw []any) rule.RuleListeners {
		opts := parseOptions(raw)
		pragma := reactutil.GetReactPragmaFromContext(ctx)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)
		scopes := scopeAnalysis.For(ctx)
		wrappers := reactutil.GetComponentWrapperFunctions(ctx.Settings, pragma)
		a := &analyzer{
			ctx:      ctx,
			lookup:   reactutil.NewVariableDefinitionLookup(ctx, scopes),
			wrappers: reactutil.GetPropWrapperFunctions(ctx.Settings),
			byNode:   map[*ast.Node]*component{},
			byTarget: map[componentKey][]*component{},
			reported: map[diagnosticKey]struct{}{},
		}
		collect := func(n *ast.Node) {
			if reactutil.IsAsyncGeneratorFunction(n) || !reactutil.IsDetectedComponent(n, pragma, createClass, wrappers, ctx.TypeChecker, scopes) {
				return
			}
			original := n
			fn := (*ast.Node)(nil)
			if reactutil.IsFunctionLikeForComponent(n) {
				fn = n
			} else if n.Kind == ast.KindCallExpression {
				call := n.AsCallExpression()
				if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
					candidate := reactutil.SkipExpressionWrappers(call.Arguments.Nodes[0])
					if reactutil.IsFunctionLikeForComponent(candidate) {
						// IsDetectedComponent has already established that the call is
						// a configured component wrapper. Retain its function even when
						// it returns a plain value and is not independently detected.
						fn = candidate
					}
				}
			}
			if fn != nil || n.Kind == ast.KindCallExpression {
				if outer := reactutil.OutermostComponentWrapperCall(n, pragma, wrappers, ctx.TypeChecker, scopes); outer != nil {
					n = outer
				}
			}
			if c := a.byNode[n]; c != nil {
				if fn != nil {
					c.fn = fn
				}
				return
			}
			c := &component{node: n, fn: fn, binding: reactutil.ComponentBinding(original, ctx.Refs.Resolve), target: strings.Join(reactutil.ComponentTarget(original), "."), props: map[string]prop{}, defaults: map[string]bool{}}
			a.byNode[n] = c
			a.components = append(a.components, c)
			key := componentKey{binding: c.binding, target: c.target}
			a.byTarget[key] = append(a.byTarget[key], c)
		}
		listeners := rule.RuleListeners{}
		for _, kind := range []ast.Kind{ast.KindClassDeclaration, ast.KindClassExpression, ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindObjectLiteralExpression, ast.KindCallExpression} {
			listeners[kind] = collect
		}
		listeners[ast.KindBinaryExpression] = func(n *ast.Node) {
			if n.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
				a.assignments = append(a.assignments, n)
			}
		}
		listeners[ast.KindEndOfFile] = func(*ast.Node) {
			// Declaration writes follow source order: an early external assignment
			// must not override a later static declaration for the same prop.
			events := append([]*ast.Node(nil), a.assignments...)
			for _, c := range a.components {
				events = append(events, c.node)
			}
			slices.SortStableFunc(events, func(x, y *ast.Node) int { return x.Pos() - y.Pos() })
			for _, n := range events {
				if c := a.byNode[n]; c != nil {
					a.initialize(c)
				} else {
					a.applyAssignment(n)
				}
			}
			for _, c := range a.components {
				c.check(a, opts)
			}
		}
		return listeners
	},
}
