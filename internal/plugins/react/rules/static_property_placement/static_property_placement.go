package static_property_placement

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
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
	name, ok := utils.GetStaticPropertyName(node)
	if !ok {
		return ""
	}
	if _, ok := propertyNames[name]; !ok {
		return ""
	}
	return name
}

func reportClassMember(name, expected string, node *ast.Node, isStatic bool, ctx rule.RuleContext) {
	if expected == propertyAssignment {
		if !isStatic {
			return
		}
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
	if receiver == nil || receiver.Kind != ast.KindIdentifier || ctx.Refs == nil {
		return false
	}
	symbol := ctx.Refs.Resolve(receiver)
	if symbol != nil && symbol.ValueDeclaration != nil {
		declaration := symbol.ValueDeclaration
		if isReactClass(declaration, pragma) {
			return true
		}
		if declaration.Kind == ast.KindVariableDeclaration {
			initializer := declaration.AsVariableDeclaration().Initializer
			if isReactClass(reactutil.SkipExpressionWrappers(initializer), pragma) {
				return true
			}
		}
	}
	// The checker/ref store can leave a declaration unresolved in a JS/TSX
	// file without a usable project symbol. Keep the same-file fallback
	// conservative and name-based, matching the plugin's related-component
	// lookup for ordinary class declarations.
	source := ast.GetSourceFileOfNode(receiver)
	if source == nil {
		return false
	}
	want := receiver.AsIdentifier().Text
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

func isInsideClass(node *ast.Node) bool {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Kind == ast.KindClassDeclaration || parent.Kind == ast.KindClassExpression {
			return true
		}
	}
	return false
}

func assignmentMember(node *ast.Node) (*ast.Node, *ast.Node, bool) {
	if node == nil {
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
	if binary.OperatorToken == nil || !ast.IsAssignmentOperator(binary.OperatorToken.Kind) || binary.Left != current {
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
			name := propertyName(node.Name())
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
