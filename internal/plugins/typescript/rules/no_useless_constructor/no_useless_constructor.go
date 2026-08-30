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

const usefulParameterModifiers = ast.ModifierFlagsParameterPropertyModifier | ast.ModifierFlagsDecorator

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
	for _, param := range params {
		if param.Kind != ast.KindParameter {
			continue
		}
		modifiers := param.AsParameterDeclaration().Modifiers()
		if modifiers != nil && modifiers.ModifierFlags&usefulParameterModifiers != 0 {
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
	if pd == nil || pd.Initializer != nil {
		return false
	}
	if modifiers := pd.Modifiers(); modifiers != nil && modifiers.ModifierFlags&usefulParameterModifiers != 0 {
		return false
	}
	// ESTree represents every rest parameter as a RestElement. Its binding
	// pattern is checked separately when pairing it with a spread argument.
	if pd.DotDotDotToken != nil {
		return true
	}
	name := pd.Name()
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
	if expr.Kind != ast.KindCallExpression {
		expr = ast.SkipParentheses(expr)
	}
	if expr == nil || expr.Kind != ast.KindCallExpression {
		return nil
	}
	call := expr.AsCallExpression()
	if call.Expression == nil || call.Expression.Kind != ast.KindSuperKeyword {
		return nil
	}
	return call
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
	expr := se.Expression
	if expr.Kind != ast.KindIdentifier {
		expr = ast.SkipParentheses(expr)
	}
	return expr != nil && expr.Kind == ast.KindIdentifier && expr.AsIdentifier().Text == "arguments"
}

// isValidIdentifierPair checks if the constructor param and super arg are both identifiers with the same name.
func isValidIdentifierPair(paramName *ast.Node, superArg *ast.Node) bool {
	if paramName == nil || superArg == nil {
		return false
	}
	if paramName.Kind != ast.KindIdentifier {
		return false
	}
	if superArg.Kind != ast.KindIdentifier {
		superArg = ast.SkipParentheses(superArg)
	}
	return superArg != nil &&
		superArg.Kind == ast.KindIdentifier &&
		paramName.AsIdentifier().Text == superArg.AsIdentifier().Text
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
		param := params[i]
		if param.Kind != ast.KindParameter {
			return false
		}
		pd := param.AsParameterDeclaration()
		if pd == nil || pd.Initializer != nil {
			return false
		}
		if modifiers := pd.Modifiers(); modifiers != nil && modifiers.ModifierFlags&usefulParameterModifiers != 0 {
			return false
		}
		paramName := pd.Name()
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

// isRedundantSuperCall checks if parameters are passed through by the only
// super() call in a derived constructor.
func isRedundantSuperCall(call *ast.CallExpression, params []*ast.Node) bool {
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

func containingClass(node *ast.Node) *ast.Node {
	classNode := node.Parent
	if classNode != nil && (classNode.Kind == ast.KindClassDeclaration || classNode.Kind == ast.KindClassExpression) {
		return classNode
	}
	return ast.GetContainingClass(node)
}

func classHasSuperClass(classNode *ast.Node) bool {
	var clauses *ast.HeritageClauseList
	switch classNode.Kind {
	case ast.KindClassDeclaration:
		clauses = classNode.AsClassDeclaration().HeritageClauses
	case ast.KindClassExpression:
		clauses = classNode.AsClassExpression().HeritageClauses
	default:
		return false
	}
	if clauses == nil {
		return false
	}
	for _, clauseNode := range clauses.Nodes {
		clause := clauseNode.AsHeritageClause()
		if clause.Token == ast.KindExtendsKeyword && clause.Types != nil && len(clause.Types.Nodes) != 0 {
			return true
		}
	}
	return false
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
		utils.ClassMemberLeadingSemicolonOptions{
			IncludePropertiesWithoutInitializers: true,
			IncludePostfixInitializers:           true,
		},
	)
}

var NoUselessConstructorRule = rule.CreateRule(rule.Rule{
	Name:   "no-useless-constructor",
	Schema: rule.EmptyArraySchema,
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

				var statements []*ast.Node
				if statementList := constructor.Body.AsBlock().Statements; statementList != nil {
					statements = statementList.Nodes
				}
				var superCall *ast.CallExpression
				switch len(statements) {
				case 0:
				case 1:
					superCall = singleSuperCall(statements)
					if superCall == nil {
						return
					}
				default:
					return
				}

				classNode := containingClass(node)
				if classNode == nil {
					return
				}

				hasSuper := classHasSuperClass(classNode)
				if hasSuper != (superCall != nil) {
					return
				}

				// TypeScript-specific: skip if accessibility makes constructor useful
				modifierFlags := ast.ModifierFlagsNone
				if modifiers := constructor.Modifiers(); modifiers != nil {
					modifierFlags = modifiers.ModifierFlags
				}
				if modifierFlags&ast.ModifierFlagsStatic != 0 ||
					!checkAccessibility(modifierFlags, hasSuper) {
					return
				}

				var params []*ast.Node
				if constructor.Parameters != nil {
					params = constructor.Parameters.Nodes
				}
				// TypeScript-specific: parameter properties and decorators make a
				// constructor useful. Derived constructors validate those modifiers
				// in the same pass that checks argument forwarding.
				if superCall != nil {
					if !isRedundantSuperCall(superCall, params) {
						return
					}
				} else if !checkParams(params) {
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
