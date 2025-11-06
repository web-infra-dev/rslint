package consistent_generic_constructors

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type ConsistentGenericConstructorsOptions struct {
	Style string `json:"style"`
}

// ConsistentGenericConstructorsRule enforces consistent generic specifier style in constructor signatures
var ConsistentGenericConstructorsRule = rule.CreateRule(rule.Rule{
	Name: "consistent-generic-constructors",
	Run:  run,
})

func run(ctx rule.RuleContext, options any) rule.RuleListeners {
	opts := ConsistentGenericConstructorsOptions{
		Style: "constructor", // default
	}

	// Parse options
	if options != nil {
		// Handle array format: ["type-annotation"]
		if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
			if style, ok := optArray[0].(string); ok {
				opts.Style = style
			}
		} else if optsMap, ok := options.(map[string]interface{}); ok {
			if style, exists := optsMap["style"].(string); exists {
				opts.Style = style
			}
		} else if style, ok := options.(string); ok {
			opts.Style = style
		}
	}

	checkNode := func(node *ast.Node, typeAnnotation *ast.Node, initializer *ast.Node, isBindingElement bool) {
		if initializer == nil {
			return
		}

		// Check if initializer is a NewExpression
		if initializer.Kind != ast.KindNewExpression {
			return
		}

		newExpr := initializer.AsNewExpression()
		if newExpr == nil {
			return
		}

		// Check if the callee is a simple identifier
		if newExpr.Expression == nil || newExpr.Expression.Kind != ast.KindIdentifier {
			return
		}

		calleeIdent := newExpr.Expression.AsIdentifier()
		if calleeIdent == nil {
			return
		}

		// Check if type arguments exist on constructor
		hasTypeArgsOnConstructor := newExpr.TypeArguments != nil && len(newExpr.TypeArguments.Nodes) > 0

		// Handle case where there's no type annotation
		if typeAnnotation == nil {
			// In type-annotation mode with type arguments on constructor,
			// we should suggest adding a type annotation
			// UNLESS it's a binding element (like array destructuring), where we can't add a type annotation
			if opts.Style == "type-annotation" && hasTypeArgsOnConstructor && !isBindingElement {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "preferTypeAnnotation",
					Description: "The generic type arguments should be specified as part of the type annotation.",
				})
			}
			// For constructor mode or no type args, no violation
			return
		}

		// Check if the type annotation is a type reference
		if typeAnnotation.Kind != ast.KindTypeReference {
			return
		}

		typeRef := typeAnnotation.AsTypeReference()
		if typeRef == nil {
			return
		}

		// Check if type reference name is an identifier
		if typeRef.TypeName == nil || typeRef.TypeName.Kind != ast.KindIdentifier {
			return
		}

		typeNameIdent := typeRef.TypeName.AsIdentifier()
		if typeNameIdent == nil {
			return
		}

		// Check if the names match
		calleeText := calleeIdent.Text
		typeNameText := typeNameIdent.Text
		if calleeText != typeNameText {
			return
		}

		// Check if type arguments exist on type annotation
		hasTypeArgsOnAnnotation := typeRef.TypeArguments != nil && len(typeRef.TypeArguments.Nodes) > 0

		// If both have type arguments or neither has type arguments, no violation
		if hasTypeArgsOnAnnotation == hasTypeArgsOnConstructor {
			return
		}

		if opts.Style == "constructor" {
			// Prefer constructor style
			if hasTypeArgsOnAnnotation && !hasTypeArgsOnConstructor {
				// Report: type args should be on constructor
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "preferConstructor",
					Description: "The generic type arguments should be specified as part of the constructor type arguments.",
				})
			}
		} else {
			// Prefer type-annotation style
			if hasTypeArgsOnConstructor && !hasTypeArgsOnAnnotation {
				// Report: type args should be on type annotation
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "preferTypeAnnotation",
					Description: "The generic type arguments should be specified as part of the type annotation.",
				})
			}
		}
	}

	return rule.RuleListeners{
		// Variable declarations
		ast.KindVariableDeclaration: func(node *ast.Node) {
			if node.Kind != ast.KindVariableDeclaration {
				return
			}
			varDecl := node.AsVariableDeclaration()
			if varDecl == nil {
				return
			}

			// For destructuring patterns, we need to be careful:
			// - `const {a}: Foo<string> = new Foo()` - has type annotation, should check
			// - `const {a} = new Foo<string>()` - the BindingElement listener handles elements inside
			// - `const [a = new Foo<string>()] = []` - the BindingElement listener handles elements inside
			// Since VariableDeclaration's initializer is the whole RHS (e.g., `[]`), not the BindingElement's initializer,
			// we can check if the name is a binding pattern without type annotation and skip
			if varDecl.Type == nil && varDecl.Name() != nil {
				nameKind := varDecl.Name().Kind
				if nameKind == ast.KindArrayBindingPattern || nameKind == ast.KindObjectBindingPattern {
					return
				}
			}

			checkNode(node, varDecl.Type, varDecl.Initializer, false)
		},

		// Property declarations (class properties, including accessor properties)
		ast.KindPropertyDeclaration: func(node *ast.Node) {
			if node.Kind != ast.KindPropertyDeclaration {
				return
			}
			propDecl := node.AsPropertyDeclaration()
			if propDecl == nil {
				return
			}
			checkNode(node, propDecl.Type, propDecl.Initializer, false)
		},

		// Parameters (for functions, constructors, methods, arrow functions)
		ast.KindParameter: func(node *ast.Node) {
			if node.Kind != ast.KindParameter {
				return
			}
			param := node.AsParameterDeclaration()
			if param == nil {
				return
			}

			// Skip if the name is a binding pattern (destructuring), there's no type annotation,
			// AND there's no initializer on the parameter itself
			// If there's a type annotation, we should check it (e.g., `function foo({a}: Foo<string> = new Foo()) {}`)
			// If there's an initializer on the parameter, we should check it (e.g., `function foo({a} = new Foo<string>()) {}`)
			// Only skip when the BindingElement listener will handle initializers inside the pattern (e.g., `function foo([a = new Foo<string>()]) {}`)
			if param.Type == nil && param.Initializer == nil && param.Name() != nil {
				nameKind := param.Name().Kind
				if nameKind == ast.KindArrayBindingPattern || nameKind == ast.KindObjectBindingPattern {
					return
				}
			}

			checkNode(node, param.Type, param.Initializer, false)
		},

		// Binding elements (for destructuring patterns)
		ast.KindBindingElement: func(node *ast.Node) {
			if node.Kind != ast.KindBindingElement {
				return
			}
			bindingElem := node.AsBindingElement()
			if bindingElem == nil {
				return
			}
			// BindingElement doesn't have a Type field, it can only have an initializer
			checkNode(node, nil, bindingElem.Initializer, true)
		},
	}
}
