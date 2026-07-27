package no_useless_constructor

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
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
func checkAccessibility(modifierFlags ast.ModifierFlags, classHasSuperClass bool) bool {
	return modifierFlags&ast.ModifierFlagsNonPublicAccessibilityModifier == 0 &&
		(!classHasSuperClass || modifierFlags&ast.ModifierFlagsPublic == 0)
}

// checkParams returns true if the constructor should be checked (no parameter
// properties or decorators that make it useful).
func checkParams(params []*ast.Node) bool {
	const usefulParameterModifiers = ast.ModifierFlagsParameterPropertyModifier | ast.ModifierFlagsDecorator
	for _, param := range params {
		if param.Kind == ast.KindParameter && param.ModifierFlags()&usefulParameterModifiers != 0 {
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
	// ESTree represents every rest parameter as a RestElement. Its binding
	// pattern is checked separately when pairing it with a spread argument.
	if pd.DotDotDotToken != nil {
		return true
	}
	name := param.Name()
	return name != nil && name.Kind == ast.KindIdentifier
}

// singleSuperCall returns the only super() call in the body.
func singleSuperCall(statements []*ast.Node) *ast.CallExpression {
	if len(statements) != 1 {
		return nil
	}
	stmt := statements[0]
	if stmt.Kind != ast.KindExpressionStatement {
		return nil
	}
	expr := stmt.Expression()
	if expr == nil {
		return nil
	}
	expr = ast.SkipParentheses(expr)
	if expr == nil || !ast.IsSuperCall(expr) {
		return nil
	}
	return expr.AsCallExpression()
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
	expr := ast.SkipParentheses(se.Expression)
	return expr != nil && expr.Kind == ast.KindIdentifier && expr.Text() == "arguments"
}

// isValidIdentifierPair checks if the constructor param and super arg are both identifiers with the same name.
func isValidIdentifierPair(paramName *ast.Node, superArg *ast.Node) bool {
	if paramName == nil || superArg == nil {
		return false
	}
	superArg = ast.SkipParentheses(superArg)
	return paramName.Kind == ast.KindIdentifier &&
		superArg != nil &&
		superArg.Kind == ast.KindIdentifier &&
		paramName.Text() == superArg.Text()
}

// isValidRestSpreadPair checks if the constructor param is a rest param and
// the super arg is a spread element with the same identifier.
func isValidRestSpreadPair(paramName *ast.Node, superArg *ast.Node) bool {
	if superArg == nil || superArg.Kind != ast.KindSpreadElement {
		return false
	}
	se := superArg.AsSpreadElement()
	if se == nil || se.Expression == nil {
		return false
	}
	return isValidIdentifierPair(paramName, se.Expression)
}

// isPassingThrough checks simplicity and 1:1 forwarding in one parameter pass.
func isPassingThrough(params []*ast.Node, args []*ast.Node) bool {
	if len(params) != len(args) {
		return false
	}
	for i := range params {
		pd := params[i].AsParameterDeclaration()
		if pd == nil || pd.Initializer != nil {
			return false
		}
		paramName := params[i].Name()
		if paramName == nil {
			return false
		}
		if pd.DotDotDotToken != nil {
			if !isValidRestSpreadPair(paramName, args[i]) {
				return false
			}
		} else if !isValidIdentifierPair(paramName, args[i]) {
			return false
		}
	}
	return true
}

// isRedundantSuperCall checks if the constructor body is just a redundant super() call.
func isRedundantSuperCall(statements []*ast.Node, params []*ast.Node) bool {
	call := singleSuperCall(statements)
	if call == nil {
		return false
	}
	var args []*ast.Node
	if call.Arguments != nil {
		args = call.Arguments.Nodes
	}
	if !isSpreadArguments(args) {
		return isPassingThrough(params, args)
	}
	for _, param := range params {
		if !isSimpleParam(param) {
			return false
		}
	}
	return true
}

// constructorRanges returns ESLint's constructor-key diagnostic span and the
// full constructor span used by its removal suggestion. It avoids allocating a
// standalone scanner on the normal diagnostic-only path.
func constructorRanges(sourceText string, node *ast.Node) (core.TextRange, core.TextRange) {
	start := scanner.SkipTrivia(sourceText, node.Pos())
	fixRange := core.NewTextRange(start, node.End())
	keyStart := start
	if modifiers := node.Modifiers(); modifiers != nil && len(modifiers.Nodes) != 0 {
		keyStart = scanner.SkipTrivia(sourceText, modifiers.Nodes[len(modifiers.Nodes)-1].End())
	}

	const constructorKeyword = "constructor"
	if keyStart >= 0 &&
		keyStart+len(constructorKeyword) <= node.End() &&
		keyStart+len(constructorKeyword) <= len(sourceText) &&
		sourceText[keyStart:keyStart+len(constructorKeyword)] == constructorKeyword {
		return core.NewTextRange(start, keyStart+len(constructorKeyword)), fixRange
	}

	// A string-literal key whose cooked value is "constructor" is also parsed
	// as a constructor. Find its closing quote without materializing a token.
	if keyStart >= 0 && keyStart < node.End() && keyStart < len(sourceText) &&
		(sourceText[keyStart] == '\'' || sourceText[keyStart] == '"') {
		quote := sourceText[keyStart]
		for pos := keyStart + 1; pos < node.End() && pos < len(sourceText); pos++ {
			switch sourceText[pos] {
			case '\\':
				pos++
			case quote:
				return core.NewTextRange(start, pos+1), fixRange
			}
		}
	}
	return fixRange, fixRange
}

func needsLeadingSemicolon(sourceFile *ast.SourceFile, classNode *ast.Node, node *ast.Node) bool {
	sourceText := sourceFile.Text()
	nextStart := scanner.SkipTrivia(sourceText, node.End())
	if nextStart >= len(sourceText) {
		return false
	}

	var nextToken utils.SourceToken
	switch sourceText[nextStart] {
	case '[':
		nextToken = utils.SourceToken{Kind: ast.KindOpenBracketToken, Start: nextStart, End: nextStart + 1, Text: "["}
	case '*':
		nextToken = utils.SourceToken{Kind: ast.KindAsteriskToken, Start: nextStart, End: nextStart + 1, Text: "*"}
	case 'i':
		// `in` and `instanceof` are the other expression-continuation
		// tokens accepted in a class body. Let the scanner distinguish them
		// from ordinary identifier names only on this rare path.
		var ok bool
		nextToken, ok = utils.TokenAtOrAfter(sourceFile, nextStart)
		if !ok {
			return false
		}
	default:
		return false
	}

	return utils.NeedsClassMemberLeadingSemicolon(
		sourceFile,
		classNode,
		node,
		nextToken,
		utils.ClassMemberLeadingSemicolonOptions{IncludePropertiesWithoutInitializers: true},
	)
}

var NoUselessConstructorRule = rule.CreateRule(rule.Rule{
	Name: "no-useless-constructor",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
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
				modifierFlags := node.ModifierFlags()
				if modifierFlags&ast.ModifierFlagsStatic != 0 ||
					!checkAccessibility(modifierFlags, hasSuper) {
					return
				}

				// TypeScript-specific: skip if params have parameter properties or decorators
				var params []*ast.Node
				if constructor.Parameters != nil {
					params = constructor.Parameters.Nodes
				}
				if !checkParams(params) {
					return
				}

				body := constructor.Body.Statements()

				isUseless := false
				if hasSuper {
					isUseless = isRedundantSuperCall(body, params)
				} else {
					isUseless = len(body) == 0
				}

				if !isUseless {
					return
				}

				reportRange, fixRange := constructorRanges(ctx.SourceFile.Text(), node)
				ctx.ReportRangeWithDeferredSuggestions(
					reportRange,
					buildNoUselessConstructorMessage(),
					func() []rule.RuleSuggestion {
						replacement := ""
						if needsLeadingSemicolon(ctx.SourceFile, classNode, node) {
							replacement = ";"
						}
						return []rule.RuleSuggestion{{
							Message:  buildRemoveConstructorMessage(),
							FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(fixRange, replacement)},
						}}
					},
				)
			},
		}
	},
})
