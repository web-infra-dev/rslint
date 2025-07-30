package no_useless_constructor

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/rslint/internal/rule"
)

func buildNoUselessConstructorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noUselessConstructor",
		Description: "Useless constructor.",
	}
}

func buildRemoveConstructorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "removeConstructor",
		Description: "Remove the constructor.",
	}
}

// Check if method with accessibility is not useless
func checkAccessibility(node *ast.Node) bool {
	if node.Kind != ast.KindConstructor {
		return true
	}

	ctor := node.AsConstructorDeclaration()
	classNode := ctor.Parent

	// Get accessibility modifier
	modifiers := ctor.Modifiers()
	if modifiers != nil {
		for _, mod := range modifiers.Nodes {
			switch mod.Kind {
			case ast.KindProtectedKeyword, ast.KindPrivateKeyword:
				// protected or private constructors are not useless
				return false
			case ast.KindPublicKeyword:
					// public constructors in classes with superClass are not useless
					if classNode != nil && ast.IsClassDeclaration(classNode) {
						classDecl := classNode.AsClassDeclaration()
						if classDecl.HeritageClauses != nil {
							for _, clause := range classDecl.HeritageClauses.Nodes {
								if clause.AsHeritageClause().Token == ast.KindExtendsKeyword {
									return false
								}
							}
						}
					}
				}
			}
		}

	return true
}

// Check if constructor is not useless due to typescript parameter properties and decorators
func checkParams(node *ast.Node) bool {
	if node.Kind != ast.KindConstructor {
		return true
	}

	ctor := node.AsConstructorDeclaration()
	params := ctor.Parameters
	if params == nil {
		return true
	}

	// Check each parameter
	for _, param := range params.Nodes {
		if param.Kind != ast.KindParameter {
			continue
		}

		paramDecl := param.AsParameterDeclaration()

		// Check if it's a parameter property (has accessibility modifier or readonly)
		if paramDecl.Modifiers() != nil {
			for _, mod := range paramDecl.Modifiers().Nodes {
				switch mod.Kind {
				case ast.KindPublicKeyword, ast.KindPrivateKeyword, ast.KindProtectedKeyword, ast.KindReadonlyKeyword:
					return false
				}
			}
		}

		// Check for decorators
		if ast.GetCombinedModifierFlags(param)&ast.ModifierFlagsDecorator != 0 {
			return false
		}
	}

	return true
}

// Check if constructor body is empty or only contains super call
func isConstructorUseless(node *ast.Node) bool {
	if node.Kind != ast.KindConstructor {
		return false
	}

	ctor := node.AsConstructorDeclaration()
	body := ctor.Body
	if body == nil {
		// Constructor without body (overload signature)
		return false
	}

	// Check if constructor extends a class
	classNode := ctor.Parent
	if classNode == nil || !ast.IsClassDeclaration(classNode) {
		return false
	}

	classDecl := classNode.AsClassDeclaration()
	var extendsClause bool
	if classDecl.HeritageClauses != nil {
		for _, clause := range classDecl.HeritageClauses.Nodes {
			if clause.AsHeritageClause().Token == ast.KindExtendsKeyword {
				extendsClause = true
				break
			}
		}
	}

	statements := body.Statements()
	if statements == nil || len(statements) == 0 {
		// Empty constructor body
		// If class extends another class, empty constructor is NOT useless
		// (even though it's a TypeScript error, we shouldn't flag it as useless)
		if extendsClause {
			return false
		}
		return true
	}

	if !extendsClause {
		// No extends clause, so any non-empty constructor is not useless
		return false
	}

	// For classes that extend, check if constructor only calls super with same args
	if len(statements) != 1 {
		return false
	}

	stmt := statements[0]
	if stmt.Kind != ast.KindExpressionStatement {
		return false
	}

	expr := stmt.AsExpressionStatement().Expression
	if expr.Kind != ast.KindCallExpression {
		return false
	}

	callExpr := expr.AsCallExpression()
	if callExpr.Expression.Kind != ast.KindSuperKeyword {
		return false
	}

	// Check if super is called with same arguments
	ctorParams := ctor.Parameters
	superArgs := callExpr.Arguments

	if ctorParams == nil && superArgs == nil {
		return true
	}
	if ctorParams == nil || superArgs == nil {
		return false
	}

	// Special case: super(...arguments) is always useless if parameters are simple
	if len(superArgs.Nodes) == 1 {
		arg := superArgs.Nodes[0]
		if arg.Kind == ast.KindSpreadElement {
			spreadExpr := arg.AsSpreadElement().Expression
			if spreadExpr.Kind == ast.KindIdentifier && 
			   spreadExpr.AsIdentifier().Text == "arguments" {
				// Check if any parameter has complex pattern (destructuring, default values)
				// If so, the constructor is not useless even with super(...arguments)
				for _, param := range ctorParams.Nodes {
					paramDecl := param.AsParameterDeclaration()
					// Check if parameter name is not a simple identifier (e.g., destructured)
					if paramDecl.Name().Kind != ast.KindIdentifier {
						return false
					}
					// Check if parameter has default value
					if paramDecl.Initializer != nil {
						return false
					}
				}
				// All parameters are simple, so super(...arguments) is useless
				return true
			}
		}
	}

	// Check if any parameter has complex pattern (destructuring, default values)
	// If so, the constructor is not useless even with super(...arguments)
	for _, param := range ctorParams.Nodes {
		paramDecl := param.AsParameterDeclaration()
		// Check if parameter name is not a simple identifier (e.g., destructured)
		if paramDecl.Name().Kind != ast.KindIdentifier {
			return false
		}
		// Check if parameter has default value
		if paramDecl.Initializer != nil {
			return false
		}
	}

	// Count non-rest parameters
	normalParamCount := 0
	var restParam *ast.Node
	for _, param := range ctorParams.Nodes {
		if param.AsParameterDeclaration().DotDotDotToken != nil {
			restParam = param
			break
		}
		normalParamCount++
	}

	// If we have rest parameters, check for spread in super call
	if restParam != nil {
		// Check if last argument is spread of rest parameter  
		if len(superArgs.Nodes) == 0 {
			return false
		}

		lastArg := superArgs.Nodes[len(superArgs.Nodes)-1]
		if lastArg.Kind != ast.KindSpreadElement {
			// Check special case: super(...arguments)
			// Only consider it useless if constructor has no parameters
			if len(superArgs.Nodes) == 1 && lastArg.Kind == ast.KindIdentifier &&
				lastArg.AsIdentifier().Text == "arguments" && len(ctorParams.Nodes) == 0 {
				return true
			}
			return false
		}

		spreadExpr := lastArg.AsSpreadElement().Expression
		if spreadExpr.Kind != ast.KindIdentifier {
			return false
		}

		// Check if spread identifier matches rest param name
		restParamName := restParam.AsParameterDeclaration().Name()
		if restParamName.Kind == ast.KindIdentifier {
			if spreadExpr.AsIdentifier().Text != restParamName.AsIdentifier().Text {
				return false
			}
		} else {
			return false
		}

		// Check non-rest args match
		if len(superArgs.Nodes)-1 != normalParamCount {
			return false
		}
	} else {
		// No rest params - check exact match or super(...arguments) with no params
		if len(ctorParams.Nodes) != len(superArgs.Nodes) {
			// Special case: constructor() { super(...arguments); }
			// This should be considered useless when constructor has no parameters
			if len(ctorParams.Nodes) == 0 && len(superArgs.Nodes) == 1 {
				arg := superArgs.Nodes[0]
				if arg.Kind == ast.KindSpreadElement {
					spreadExpr := arg.AsSpreadElement().Expression
					if spreadExpr.Kind == ast.KindIdentifier && 
					   spreadExpr.AsIdentifier().Text == "arguments" {
						return true
					}
				}
			}
			return false
		}
	}

	// Check each argument matches its parameter
	for i := 0; i < normalParamCount && i < len(superArgs.Nodes); i++ {
		param := ctorParams.Nodes[i].AsParameterDeclaration()
		arg := superArgs.Nodes[i]

		// Skip spread arguments (handled above)
		if arg.Kind == ast.KindSpreadElement {
			continue
		}

		if arg.Kind != ast.KindIdentifier {
			return false
		}

		paramName := param.Name()
		if paramName.Kind != ast.KindIdentifier {
			return false
		}

		if arg.AsIdentifier().Text != paramName.AsIdentifier().Text {
			return false
		}
	}

	return true
}

// Helper to get text range for suggestion fix
func getConstructorRange(ctx rule.RuleContext, node *ast.Node) core.TextRange {
	if node.Kind != ast.KindConstructor {
		return core.NewTextRange(node.Pos(), node.End())
	}

	// Find the start of constructor including any leading whitespace on the same line
	start := node.Pos()
	text := ctx.SourceFile.Text()
	
	// Move back to beginning of line
	for start > 0 && text[start-1] != '\n' {
		start--
	}

	// Find end including the closing brace and any trailing whitespace
	end := node.End()
	
	// Skip trailing whitespace and include newline if present
	for end < len(text) && (text[end] == ' ' || text[end] == '\t') {
		end++
	}
	if end < len(text) && text[end] == '\n' {
		end++
	}
	
	return core.NewTextRange(start, end)
}

var NoUselessConstructorRule = rule.Rule{
	Name: "no-useless-constructor",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindConstructor: func(node *ast.Node) {
				// Skip if constructor has accessibility that makes it not useless
				if !checkAccessibility(node) {
					return
				}

				// Skip if constructor has parameter properties or decorators
				if !checkParams(node) {
					return
				}

				// Check if constructor is actually useless
				if !isConstructorUseless(node) {
					return
				}

				// Report with suggestion to remove
				removeRange := getConstructorRange(ctx, node)
				
				ctx.ReportNodeWithSuggestions(node, buildNoUselessConstructorMessage(),
					rule.RuleSuggestion{
						Message: buildRemoveConstructorMessage(),
						FixesArr: []rule.RuleFix{
							rule.RuleFixReplaceRange(removeRange, "  "),
						},
					})
			},
		}
	},
}