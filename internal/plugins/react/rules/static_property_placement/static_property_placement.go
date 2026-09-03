package static_property_placement

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
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
		overrides:       map[string]string{},
	}
	if len(raw) > 0 {
		if position, ok := raw[0].(string); ok && position != "" {
			o.defaultPosition = position
		}
	}
	if len(raw) > 1 {
		if m, ok := raw[1].(map[string]interface{}); ok {
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
	var name string
	switch node.Kind {
	case ast.KindIdentifier:
		name = node.AsIdentifier().Text
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

func reportClassMember(name, expected string, node *ast.Node, isStatic bool, ctx rule.RuleContext) {
	if expected == propertyAssignment {
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "declareOutsideClass",
			Description: fmt.Sprintf("'%s' should be declared outside the class body.", name),
			Data:        map[string]string{"name": name},
		})
		return
	}

	if expected == staticPublicField && node.Kind == ast.KindPropertyDeclaration && isStatic {
		return
	}
	if expected == staticGetter && node.Kind == ast.KindGetAccessor && isStatic {
		return
	}

	messageID := "notStaticClassProp"
	description := fmt.Sprintf("'%s' should be declared as a static class property.", name)
	if expected == staticGetter {
		messageID = "notGetterClassFunc"
		description = fmt.Sprintf("'%s' should be declared as a static getter class function.", name)
	}
	ctx.ReportNode(node, rule.RuleMessage{
		Id:          messageID,
		Description: description,
		Data:        map[string]string{"name": name},
	})
}

func reportAssignment(name, expected string, node *ast.Node, ctx rule.RuleContext) {
	if expected == propertyAssignment {
		return
	}
	messageID := "notStaticClassProp"
	description := fmt.Sprintf("'%s' should be declared as a static class property.", name)
	if expected == staticGetter {
		messageID = "notGetterClassFunc"
		description = fmt.Sprintf("'%s' should be declared as a static getter class function.", name)
	}
	ctx.ReportNode(node, rule.RuleMessage{
		Id:          messageID,
		Description: description,
		Data:        map[string]string{"name": name},
	})
}

func isReactClass(node *ast.Node, pragma string) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression:
		return reactutil.ExtendsReactComponent(node, pragma)
	}
	return false
}

func relatedReactClass(ctx rule.RuleContext, receiver *ast.Node, pragma string) bool {
	receiver = ast.SkipParentheses(receiver)
	if ast.IsOptionalChain(receiver) {
		return false
	}
	root, path, ok := componentPath(receiver)
	if !ok || ctx.Refs == nil {
		return false
	}
	symbol := ctx.Refs.Resolve(root)
	if symbol != nil && symbol.ValueDeclaration != nil {
		declaration := symbol.ValueDeclaration
		if isReactClass(resolveComponentPath(declaration, path), pragma) {
			return true
		}
		// A successfully resolved local binding is authoritative. Falling
		// back to a same-name class elsewhere in the file would cross a
		// shadowing boundary and report a non-component binding.
		return false
	}
	// The checker/ref store can leave a declaration unresolved in a JS/TSX
	// file without a usable project symbol. Keep the same-file fallback
	// conservative and name-based, matching the plugin's related-component
	// lookup for ordinary class declarations.
	source := ast.GetSourceFileOfNode(receiver)
	if source == nil {
		return false
	}
	if len(path) != 0 {
		return false
	}
	want := root.AsIdentifier().Text
	found := false
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if found || node == nil {
			return
		}
		if node.Kind == ast.KindClassDeclaration {
			name := node.Name()
			if name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == want && isReactClass(node, pragma) {
				found = true
				return
			}
		}
		node.ForEachChild(func(child *ast.Node) bool { visit(child); return found })
	}
	visit(source.AsNode())
	return found
}

func componentPath(node *ast.Node) (*ast.Node, []string, bool) {
	if node == nil {
		return nil, nil, false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return node, nil, true
	case ast.KindPropertyAccessExpression:
		property := node.AsPropertyAccessExpression()
		if property.Name() == nil || property.Name().Kind != ast.KindIdentifier {
			return nil, nil, false
		}
		root, path, ok := componentPath(property.Expression)
		if !ok {
			return nil, nil, false
		}
		return root, append(path, property.Name().AsIdentifier().Text), true
	default:
		return nil, nil, false
	}
}

func resolveComponentPath(node *ast.Node, path []string) *ast.Node {
	node = reactutil.SkipExpressionWrappers(node)
	if node == nil {
		return nil
	}
	if node.Kind == ast.KindVariableDeclaration {
		node = reactutil.SkipExpressionWrappers(node.AsVariableDeclaration().Initializer)
	}
	for _, name := range path {
		if node == nil || node.Kind != ast.KindObjectLiteralExpression {
			return nil
		}
		var next *ast.Node
		for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
			if property == nil || property.Name() == nil || propertyIdentifierName(property.Name()) != name {
				continue
			}
			if property.Kind == ast.KindPropertyAssignment {
				next = property.AsPropertyAssignment().Initializer
			}
			break
		}
		node = reactutil.SkipExpressionWrappers(next)
	}
	return node
}

func propertyIdentifierName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	if node.Kind == ast.KindIdentifier {
		return node.AsIdentifier().Text
	}
	if node.Kind == ast.KindComputedPropertyName {
		expr := ast.SkipParentheses(node.AsComputedPropertyName().Expression)
		if expr != nil && expr.Kind == ast.KindIdentifier {
			return expr.AsIdentifier().Text
		}
	}
	return ""
}

func isInsideClass(node *ast.Node) bool {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Kind == ast.KindClassDeclaration {
			return true
		}
	}
	return false
}

func assignmentMember(node *ast.Node) (*ast.Node, *ast.Node, bool) {
	if node == nil {
		return nil, nil, false
	}
	if ast.IsOptionalChain(node) {
		return nil, nil, false
	}
	var nameNode *ast.Node
	var receiver *ast.Node
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		property := node.AsPropertyAccessExpression()
		nameNode, receiver = property.Name(), property.Expression
	case ast.KindElementAccessExpression:
		property := node.AsElementAccessExpression()
		nameNode, receiver = property.ArgumentExpression, property.Expression
	default:
		return nil, nil, false
	}
	name := propertyName(nameNode)
	if name == "" {
		return nil, nil, false
	}
	current := node
	parent := current.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		current, parent = parent, parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return nil, nil, false
	}
	binary := parent.AsBinaryExpression()
	if binary.OperatorToken == nil || binary.OperatorToken.Kind == ast.KindCommaToken || binary.Right == nil {
		return nil, nil, false
	}
	return node, receiver, true
}

var StaticPropertyPlacementRule = rule.Rule{
	Name:   "react/static-property-placement",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		o := parseOptions(rawOptions)
		pragma := reactutil.GetReactPragma(ctx.Settings)
		checkClassMember := func(node *ast.Node, isStatic bool) {
			if reactutil.GetParentReactComponentScopeBased(node, pragma, reactutil.GetReactCreateClass(ctx.Settings)) == nil {
				return
			}
			name := classMemberName(node)
			if name == "" {
				return
			}
			// MethodDefinition's upstream listener only handles static getters;
			// ordinary methods and instance accessors are not class properties.
			if node.Kind == ast.KindGetAccessor && !isStatic {
				return
			}
			reportClassMember(name, o.position(name), node, isStatic, ctx)
		}

		return rule.RuleListeners{
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				checkClassMember(node, node.Modifiers() != nil && node.Modifiers().Nodes != nil && hasStaticModifier(node))
			},
			ast.KindGetAccessor: func(node *ast.Node) {
				checkClassMember(node, hasStaticModifier(node))
			},
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				member, receiver, ok := assignmentMember(node)
				if !ok || isInsideClass(node) || !relatedReactClass(ctx, receiver, pragma) {
					return
				}
				name := propertyName(member.AsPropertyAccessExpression().Name())
				reportAssignment(name, o.position(name), member, ctx)
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				member, receiver, ok := assignmentMember(node)
				if !ok || isInsideClass(node) || !relatedReactClass(ctx, receiver, pragma) {
					return
				}
				name := propertyName(member.AsElementAccessExpression().ArgumentExpression)
				reportAssignment(name, o.position(name), member, ctx)
			},
		}
	},
}

func hasStaticModifier(node *ast.Node) bool {
	if node == nil || node.Modifiers() == nil {
		return false
	}
	for _, modifier := range node.Modifiers().Nodes {
		if modifier.Kind == ast.KindStaticKeyword {
			return true
		}
	}
	return false
}
