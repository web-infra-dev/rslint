package static_property_placement

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

//go:embed static_property_placement.schema.json
var schemaJSON []byte

const (
	staticPublicField  = "static public field"
	staticGetter       = "static getter"
	propertyAssignment = "property assignment"
)

var propertyNames = map[string]struct{}{
	"childContextTypes": {},
	"contextTypes":      {},
	"contextType":       {},
	"defaultProps":      {},
	"displayName":       {},
	"propTypes":         {},
}

type options struct {
	defaultPosition string
	overrides       map[string]string
}

func parseOptions(raw []any) options {
	o := options{
		defaultPosition: staticPublicField,
	}
	if len(raw) > 0 {
		if position, ok := raw[0].(string); ok && position != "" {
			o.defaultPosition = position
		}
	}
	if len(raw) > 1 {
		if m, ok := raw[1].(map[string]interface{}); ok {
			o.overrides = make(map[string]string, len(m))
			for name := range propertyNames {
				if position, ok := m[name].(string); ok && position != "" {
					o.overrides[name] = position
				}
			}
		}
	}
	return o
}

func (o options) position(name string) string {
	if position, ok := o.overrides[name]; ok {
		return position
	}
	return o.defaultPosition
}

func propertyName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	node = ast.SkipParentheses(node)
	var name string
	switch node.Kind {
	case ast.KindIdentifier:
		name = node.AsIdentifier().Text
	case ast.KindPrivateIdentifier:
		name = reactutil.IdentifierOrPrivateName(node)
		// The upstream displayName predicate accepts Identifier and Literal
		// property nodes, but not PrivateIdentifier nodes. Other properties
		// still use the generic `.name` lookup and must retain private-name
		// support.
		if name == "displayName" {
			return ""
		}
	case ast.KindStringLiteral:
		// displayName is the one property whose upstream predicate accepts
		// a Literal key. The other predicates read key.name and therefore do
		// not match string-literal keys.
		if node.AsStringLiteral().Text == "displayName" {
			name = "displayName"
		}
	case ast.KindComputedPropertyName:
		expr := ast.SkipParentheses(node.AsComputedPropertyName().Expression)
		if expr == nil {
			return ""
		}
		if expr.Kind == ast.KindIdentifier {
			name = expr.AsIdentifier().Text
		} else if expr.Kind == ast.KindStringLiteral && expr.AsStringLiteral().Text == "displayName" {
			name = "displayName"
		}
	}
	if name == "" {
		return ""
	}
	if name == "getDefaultProps" {
		return "defaultProps"
	}
	if _, ok := propertyNames[name]; !ok {
		return ""
	}
	return name
}

func classMemberName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	if name := propertyName(node.Name()); name != "" {
		return name
	}
	if node.Kind != ast.KindPropertyDeclaration {
		return ""
	}
	property := node.AsPropertyDeclaration()
	if property.Type == nil {
		return ""
	}
	nameNode := property.Name()
	if nameNode == nil {
		return ""
	}
	var name string
	switch nameNode.Kind {
	case ast.KindIdentifier:
		name = nameNode.AsIdentifier().Text
	case ast.KindPrivateIdentifier:
		name = reactutil.IdentifierOrPrivateName(nameNode)
	case ast.KindComputedPropertyName:
		expr := ast.SkipParentheses(nameNode.AsComputedPropertyName().Expression)
		if expr != nil && expr.Kind == ast.KindIdentifier {
			name = expr.AsIdentifier().Text
		}
	}
	// typescript-eslint's Flow compatibility branch also applies to the
	// TypeScript ESTree shape: typed `props` and `context` fields are treated
	// as propTypes and contextTypes respectively.
	switch name {
	case "props":
		return "propTypes"
	case "context":
		return "contextTypes"
	default:
		return ""
	}
}

// Messages are immutable; every diagnostic can reuse their formatted text and
// data instead of allocating the same property names for each source member.
var placementMessages = func() map[string]map[string]rule.RuleMessage {
	messages := make(map[string]map[string]rule.RuleMessage, 3)
	for _, position := range []string{staticPublicField, staticGetter, propertyAssignment} {
		messages[position] = make(map[string]rule.RuleMessage, len(propertyNames))
		for name := range propertyNames {
			id, description := "notStaticClassProp", "'%s' should be declared as a static class property."
			switch position {
			case staticGetter:
				id, description = "notGetterClassFunc", "'%s' should be declared as a static getter class function."
			case propertyAssignment:
				id, description = "declareOutsideClass", "'%s' should be declared outside the class body."
			}
			messages[position][name] = rule.RuleMessage{Id: id, Description: fmt.Sprintf(description, name), Data: map[string]string{"name": name}}
		}
	}
	return messages
}()

type componentResolver struct {
	ctx        rule.RuleContext
	pragma     string
	classes    map[*ast.Node]bool
	scopes     *scope.Manager
	firstChild map[*scope.Scope]*scope.Scope
}

func (r *componentResolver) isReactComponent(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if result, ok := r.classes[node]; ok {
		return result
	}
	result := ((node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindClassExpression) && reactutil.ExtendsReactComponent(node, r.pragma)) || reactutil.IsExplicitReactComponent(node)
	if r.classes == nil {
		r.classes = make(map[*ast.Node]bool)
	}
	r.classes[node] = result
	return result
}

func (r *componentResolver) variable(node *ast.Node, name string) []*scope.Variable {
	// A direct top-level assignment to a singly declared top-level binding
	// cannot reach the child-scope fallback. Reuse the binder's declaration
	// here; nested scopes and merged declarations use the complete model below.
	if parent := ast.WalkUpParenthesizedExpressions(node.Parent); parent != nil {
		statement := ast.WalkUpParenthesizedExpressions(parent.Parent)
		if statement != nil && statement.Kind == ast.KindExpressionStatement && statement.Parent == r.ctx.SourceFile.AsNode() {
			if symbol := r.ctx.SourceFile.Locals[name]; symbol != nil && len(symbol.Declarations) == 1 {
				declaration := symbol.Declarations[0]
				kind, topLevel := scope.DefVariable, false
				switch declaration.Kind {
				case ast.KindVariableDeclaration:
					topLevel = declaration.Parent.Kind == ast.KindVariableDeclarationList && declaration.Parent.Parent.Kind == ast.KindVariableStatement && declaration.Parent.Parent.Parent == r.ctx.SourceFile.AsNode()
				case ast.KindClassDeclaration:
					kind, topLevel = scope.DefClassName, declaration.Parent == r.ctx.SourceFile.AsNode()
				case ast.KindFunctionDeclaration:
					kind, topLevel = scope.DefFunctionName, declaration.Parent == r.ctx.SourceFile.AsNode()
				}
				if topLevel && declaration.Name() != nil && declaration.Name().Kind == ast.KindIdentifier && declaration.Name().Text() == name {
					return []*scope.Variable{{ID: declaration.Name(), DefNode: declaration, Kind: kind}}
				}
			}
		}
	}
	if r.scopes == nil {
		r.scopes = scope.Build(r.ctx.SourceFile, scope.Options{})
		r.firstChild = make(map[*scope.Scope]*scope.Scope)
		for _, s := range r.scopes.Scopes {
			if s.Parent != nil && r.firstChild[s.Parent] == nil {
				r.firstChild[s.Parent] = s
			}
		}
	}
	// React's variable helper also checks the first two child scopes before
	// walking outward. The shared scope model preserves those boundaries.
	for s := r.scopes.Acquire(node); s != nil; s = s.Parent {
		child := s
		for range 3 {
			if child == nil {
				break
			}
			if declarations := child.Declarations(name); len(declarations) != 0 {
				return declarations
			}
			child = r.firstChild[child]
		}
	}
	return nil
}

func (r *componentResolver) related(node *ast.Node) bool {
	// Build the complete ESTree member path. Literal/private keys are omitted,
	// including a trailing ["displayName"], before separating the binding name.
	var buffer [8]string
	path := buffer[:0]
	for current := node; current != nil; {
		var object, property *ast.Node
		switch current.Kind {
		case ast.KindPropertyAccessExpression:
			access := current.AsPropertyAccessExpression()
			object, property = access.Expression, access.Name()
		case ast.KindElementAccessExpression:
			access := current.AsElementAccessExpression()
			object, property = access.Expression, ast.SkipParentheses(access.ArgumentExpression)
		default:
			current = nil
			continue
		}
		if property != nil && property.Kind == ast.KindIdentifier {
			path = append(path, property.Text())
		}
		object = ast.SkipParentheses(object)
		if object != nil && object.Kind == ast.KindIdentifier {
			path = append(path, object.Text())
		}
		current = object
	}
	if len(path) == 0 {
		return false
	}
	slices.Reverse(path)
	declarations := r.variable(node, path[0])
	if len(declarations) == 0 {
		return false
	}
	componentName := strings.Join(path[:len(path)-1], ".")
	// Initializer writes are ESLint references, but RefStore deliberately
	// excludes declaration names. Compare the earliest matching initializer
	// with its source-ordered index without copying or sorting that index.
	var first *ast.Node
	for _, declaration := range declarations {
		if declaration.Kind == scope.DefVariable && declaration.DefNode.Kind == ast.KindVariableDeclaration && declaration.DefNode.AsVariableDeclaration().Initializer != nil {
			id := declaration.ID
			if utils.TrimmedNodeText(r.ctx.SourceFile, id) == componentName && (first == nil || id.Pos() < first.Pos()) {
				first = id
			}
		}
	}
	if componentName != "" {
		for _, reference := range r.ctx.Refs.References(declarations[0].DefNode.Symbol()) {
			if first != nil && reference.Pos() > first.Pos() {
				break
			}
			ref := reference
			if parent := ast.WalkUpParenthesizedExpressions(ref.Parent); parent != nil && (parent.Kind == ast.KindPropertyAccessExpression || parent.Kind == ast.KindElementAccessExpression) {
				ref = parent
			}
			if utils.TrimmedNodeText(r.ctx.SourceFile, ref) == componentName {
				first = ref
				break
			}
		}
	}
	if first != nil {
		var component *ast.Node
		parent := ast.WalkUpParenthesizedExpressions(first.Parent)
		if first.Kind == ast.KindPropertyAccessExpression || first.Kind == ast.KindElementAccessExpression {
			component = binaryRight(parent)
		} else if parent != nil && parent.Kind == ast.KindVariableDeclaration && parent.AsVariableDeclaration().Initializer != nil {
			init := ast.SkipParentheses(parent.AsVariableDeclaration().Initializer)
			if init != nil && init.Kind != ast.KindIdentifier {
				component = init
			}
		}
		if component != nil {
			return r.isReactComponent(ast.SkipParentheses(component))
		}
		// The first textual match is authoritative even when it has no right
		// side: then upstream falls back to the declaration, not a later write.
	}
	for _, declaration := range declarations {
		switch declaration.Kind {
		case scope.DefVariable, scope.DefClassName, scope.DefClassInnerName, scope.DefFunctionName, scope.DefFnExprName:
			node := declaration.DefNode
			if node.Kind == ast.KindBindingElement {
				node = ast.FindAncestorKind(node, ast.KindVariableDeclaration)
			}
			if node != nil && node.Kind == ast.KindVariableDeclaration && node.AsVariableDeclaration().Initializer != nil {
				node = node.AsVariableDeclaration().Initializer
			}
			return r.isReactComponent(resolveComponentPath(node, path[1:]))
		}
	}
	return false
}

func resolveComponentPath(node *ast.Node, path []string) *ast.Node {
	if node == nil {
		return nil
	}
	node = ast.SkipParentheses(node)
	for _, name := range path {
		if node == nil {
			return nil
		}
		// Only object literals have ESTree properties; class and TS wrapper nodes
		// retain their identity when a later path segment is inspected.
		if node.Kind != ast.KindObjectLiteralExpression {
			continue
		}
		var next *ast.Node
		for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
			key := property.Name()
			if key != nil && key.Kind == ast.KindComputedPropertyName {
				key = ast.SkipParentheses(key.AsComputedPropertyName().Expression)
			}
			if key == nil || key.Kind != ast.KindIdentifier || key.Text() != name {
				continue
			}
			if property.Kind == ast.KindPropertyAssignment {
				next = property.AsPropertyAssignment().Initializer
			} else if ast.IsFunctionLike(property) {
				next = property
			}
			break
		}
		if next == nil {
			return nil
		}
		node = ast.SkipParentheses(next)
	}
	return node
}

func binaryRight(node *ast.Node) *ast.Node {
	if node == nil || node.Kind != ast.KindBinaryExpression || utils.IsCommaOperator(node) {
		return nil
	}
	return node.AsBinaryExpression().Right
}

var StaticPropertyPlacementRule = rule.Rule{
	Name:   "react/static-property-placement",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		o := parseOptions(rawOptions)
		resolver := componentResolver{ctx: ctx, pragma: reactutil.GetReactPragmaFromContext(ctx)}
		checkClassMember := func(node *ast.Node) {
			if !utils.IsPlainClassMember(node) {
				return
			}
			isStatic := ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic)
			if node.Kind == ast.KindGetAccessor && !isStatic {
				return
			}
			name := classMemberName(node)
			if name == "" {
				return
			}
			expected := o.position(name)
			if isStatic && ((node.Kind == ast.KindPropertyDeclaration && expected == staticPublicField) || (node.Kind == ast.KindGetAccessor && expected == staticGetter)) {
				return
			}
			if resolver.isReactComponent(reactutil.EnclosingClass(node)) {
				ctx.ReportNode(node, placementMessages[expected][name])
			}
		}
		checkAssignment := func(node *ast.Node) {
			if ast.IsOptionalChain(node) || binaryRight(ast.WalkUpParenthesizedExpressions(node.Parent)) == nil {
				return
			}
			var name string
			if node.Kind == ast.KindPropertyAccessExpression {
				name = propertyName(node.AsPropertyAccessExpression().Name())
			} else {
				name = propertyName(node.AsElementAccessExpression().ArgumentExpression)
			}
			if name == "" {
				return
			}
			expected := o.position(name)
			if expected == propertyAssignment || ast.FindAncestorKind(node.Parent, ast.KindClassDeclaration) != nil {
				return
			}
			if resolver.related(node) {
				ctx.ReportNode(node, placementMessages[expected][name])
			}
		}
		return rule.RuleListeners{
			ast.KindPropertyDeclaration:      checkClassMember,
			ast.KindGetAccessor:              checkClassMember,
			ast.KindPropertyAccessExpression: checkAssignment,
			ast.KindElementAccessExpression:  checkAssignment,
		}
	},
}
