package no_useless_constructor

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
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

// checkAccessibility returns true if the constructor should be checked (accessibility
// does not make it useful). Returns false (skip) for private/protected constructors,
// and for public constructors in classes that extend another class.
func checkAccessibility(node *ast.Node, classHasSuperClass bool) bool {
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsPrivate) {
		return false
	}
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsProtected) {
		return false
	}
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsPublic) {
		if classHasSuperClass {
			return false
		}
	}
	return true
}

// checkParams returns true if the constructor should be checked (no parameter
// properties or decorators that make it useful).
func checkParams(node *ast.Node, params []*ast.Node) bool {
	for _, param := range params {
		if param.Kind != ast.KindParameter {
			continue
		}
		if ast.IsParameterPropertyDeclaration(param, node) {
			return false
		}
		if ast.HasDecorators(param) {
			return false
		}
	}
	return true
}

// isSimpleParam checks if a parameter is a simple identifier (no destructuring, no default value).
// Rest parameters are considered simple.
func isSimpleParam(param *ast.Node) bool {
	if param.Kind != ast.KindParameter {
		return false
	}
	pd := param.AsParameterDeclaration()
	if pd == nil {
		return false
	}
	// Must not have default value
	if pd.Initializer != nil {
		return false
	}
	// Must be a simple identifier (not destructuring)
	name := param.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return false
	}
	return true
}

// isSingleSuperCall checks if the body consists of exactly one statement: super(...).
func isSingleSuperCall(statements []*ast.Node) bool {
	if len(statements) != 1 {
		return false
	}
	stmt := statements[0]
	if stmt.Kind != ast.KindExpressionStatement {
		return false
	}
	expr := stmt.Expression()
	if expr == nil || expr.Kind != ast.KindCallExpression {
		return false
	}
	call := expr.AsCallExpression()
	if call == nil || call.Expression == nil {
		return false
	}
	return call.Expression.Kind == ast.KindSuperKeyword
}

// isSpreadArguments checks if the arguments are exactly `...arguments`.
func isSpreadArguments(args []*ast.Node) bool {
	if len(args) != 1 {
		return false
	}
	arg := args[0]
	if arg.Kind != ast.KindSpreadElement {
		return false
	}
	se := arg.AsSpreadElement()
	if se == nil || se.Expression == nil {
		return false
	}
	return se.Expression.Kind == ast.KindIdentifier && se.Expression.Text() == "arguments"
}

// isValidIdentifierPair checks if the constructor param and super arg are both identifiers with the same name.
func isValidIdentifierPair(paramName *ast.Node, superArg *ast.Node) bool {
	return paramName.Kind == ast.KindIdentifier &&
		superArg.Kind == ast.KindIdentifier &&
		paramName.Text() == superArg.Text()
}

// isValidRestSpreadPair checks if the constructor param is a rest param and
// the super arg is a spread element with the same identifier.
func isValidRestSpreadPair(param *ast.Node, superArg *ast.Node) bool {
	pd := param.AsParameterDeclaration()
	if pd == nil || pd.DotDotDotToken == nil {
		return false
	}
	if superArg.Kind != ast.KindSpreadElement {
		return false
	}
	se := superArg.AsSpreadElement()
	if se == nil || se.Expression == nil {
		return false
	}
	paramName := param.Name()
	if paramName == nil {
		return false
	}
	return isValidIdentifierPair(paramName, se.Expression)
}

// isPassingThrough checks if constructor params are passed through to super() 1:1.
func isPassingThrough(params []*ast.Node, args []*ast.Node) bool {
	if len(params) != len(args) {
		return false
	}
	for i := range params {
		pd := params[i].AsParameterDeclaration()
		if pd == nil {
			return false
		}
		paramName := params[i].Name()
		if paramName == nil {
			return false
		}
		if pd.DotDotDotToken != nil {
			if !isValidRestSpreadPair(params[i], args[i]) {
				return false
			}
		} else {
			if !isValidIdentifierPair(paramName, args[i]) {
				return false
			}
		}
	}
	return true
}

// isRedundantSuperCall checks if the constructor body is just a redundant super() call.
func isRedundantSuperCall(statements []*ast.Node, params []*ast.Node) bool {
	if !isSingleSuperCall(statements) {
		return false
	}
	// All params must be simple (identifier or rest, no destructuring/defaults)
	for _, p := range params {
		if !isSimpleParam(p) {
			return false
		}
	}
	// Get the super call arguments
	expr := statements[0].Expression()
	call := expr.AsCallExpression()
	var args []*ast.Node
	if call.Arguments != nil {
		args = call.Arguments.Nodes
	}
	return isSpreadArguments(args) || isPassingThrough(params, args)
}

var NoUselessConstructorRule = rule.CreateRule(rule.Rule{
	Name: "no-useless-constructor",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindConstructor: func(node *ast.Node) {
				constructor := node.AsConstructorDeclaration()
				if constructor == nil {
					return
				}

				// No body means it's a declaration (declare class, overload signature, abstract)
				if constructor.Body == nil {
					return
				}

				classNode := ast.GetContainingClass(node)
				if classNode == nil {
					return
				}

				hasSuper := ast.GetExtendsHeritageClauseElement(classNode) != nil

				// TypeScript-specific: skip if accessibility makes constructor useful
				if !checkAccessibility(node, hasSuper) {
					return
				}

				// TypeScript-specific: skip if params have parameter properties or decorators
				var params []*ast.Node
				if constructor.Parameters != nil {
					params = constructor.Parameters.Nodes
				}
				if !checkParams(node, params) {
					return
				}

				body := constructor.Body.Statements()

				isUseless := false
				if hasSuper {
					isUseless = isRedundantSuperCall(body, params)
				} else {
					isUseless = len(body) == 0
				}

				if isUseless {
					ctx.ReportNodeWithSuggestions(node, buildNoUselessConstructorMessage(),
						rule.RuleSuggestion{
							Message:  buildRemoveConstructorMessage(),
							FixesArr: []rule.RuleFix{rule.RuleFixRemove(ctx.SourceFile, node)},
						},
					)
				}
			},
		}
	},
})
