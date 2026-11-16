package no_confusing_void_expression

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_confusing_void_expression.schema.json
var schemaJSON []byte

func buildInvalidVoidExprMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidVoidExpr",
		Description: "Placing a void expression inside another expression is forbidden. Move it to its own statement instead.",
	}
}
func buildInvalidVoidExprArrowMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidVoidExprArrow",
		Description: "Returning a void expression from an arrow function shorthand is forbidden. Please add braces to the arrow function.",
	}
}
func buildInvalidVoidExprArrowWrapVoidMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidVoidExprArrowWrapVoid",
		Description: "Void expressions returned from an arrow function shorthand must be marked explicitly with the `void` operator.",
	}
}
func buildInvalidVoidExprReturnMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidVoidExprReturn",
		Description: "Returning a void expression from a function is forbidden. Please move it before the `return` statement.",
	}
}
func buildInvalidVoidExprReturnLastMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidVoidExprReturnLast",
		Description: "Returning a void expression from a function is forbidden. Please remove the `return` statement.",
	}
}
func buildInvalidVoidExprReturnWrapVoidMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidVoidExprReturnWrapVoid",
		Description: "Void expressions returned from a function must be marked explicitly with the `void` operator.",
	}
}
func buildInvalidVoidExprWrapVoidMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidVoidExprWrapVoid",
		Description: "Void expressions used inside another expression must be moved to its own statement or marked explicitly with the `void` operator.",
	}
}
func buildVoidExprWrapVoidMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "voidExprWrapVoid",
		Description: "Mark with an explicit `void` operator.",
	}
}

type NoConfusingVoidExpressionOptions struct {
	IgnoreArrowShorthand         bool
	IgnoreVoidOperator           bool
	IgnoreVoidReturningFunctions bool
}

func parseOptions(options []any) NoConfusingVoidExpressionOptions {
	opts := NoConfusingVoidExpressionOptions{}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]any)
	if value, ok := optsMap["ignoreArrowShorthand"].(bool); ok {
		opts.IgnoreArrowShorthand = value
	}
	if value, ok := optsMap["ignoreVoidOperator"].(bool); ok {
		opts.IgnoreVoidOperator = value
	}
	if value, ok := optsMap["ignoreVoidReturningFunctions"].(bool); ok {
		opts.IgnoreVoidReturningFunctions = value
	}
	return opts
}

// statementExpression converts an expression to statement position without
// reprinting its contents. Parentheses inside callees/receivers are significant:
// GetLeftmostExpression deliberately stops at those authored parentheses.
func statementExpression(file *ast.SourceFile, expression *ast.Node) (string, ast.Kind) {
	inner := ast.SkipParentheses(expression)
	outerRange := utils.TrimNodeTextRange(file, expression)
	innerRange := utils.TrimNodeTextRange(file, inner)
	text := file.Text()
	leftmost := ast.GetLeftmostExpression(inner, false)
	needsParens := false
	switch leftmost.Kind {
	case ast.KindFunctionExpression, ast.KindClassExpression, ast.KindObjectLiteralExpression:
		needsParens = true
	case ast.KindIdentifier:
		// `let [` is a lexical declaration in statement position, even if
		// the original expression was an element access on an identifier.
		if leftmost.AsIdentifier().Text == "let" {
			needsParens = scanner.ScanTokenAtPosition(file, leftmost.End()) == ast.KindOpenBracketToken
		}
	}

	if expression != inner {
		// Only grouping delimiters and trivia are outside the inner node;
		// a slash in either span belongs to a comment, not a string/regexp.
		hasComment := strings.Contains(text[outerRange.Pos():innerRange.Pos()], "/") ||
			strings.Contains(text[innerRange.End():outerRange.End()], "/")
		if needsParens || hasComment {
			return text[outerRange.Pos():outerRange.End()], ast.KindOpenParenToken
		}
	}

	expressionText := text[innerRange.Pos():innerRange.End()]
	if needsParens {
		return "(" + expressionText + ")", ast.KindOpenParenToken
	}
	return expressionText, scanner.ScanTokenAtPosition(file, innerRange.Pos())
}

func needsStatementSeparator(firstToken ast.Kind) bool {
	switch firstToken {
	case ast.KindOpenParenToken, ast.KindOpenBracketToken,
		ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateHead,
		ast.KindPlusToken, ast.KindMinusToken, ast.KindPlusPlusToken, ast.KindMinusMinusToken,
		ast.KindSlashToken, ast.KindSlashEqualsToken, ast.KindRegularExpressionLiteral,
		ast.KindLessThanToken:
		return true
	}
	return false
}

func returnStatementText(file *ast.SourceFile, statement *ast.Node) string {
	expression := statement.AsReturnStatement().Expression
	expressionText, firstToken := statementExpression(file, expression)
	expressionRange := utils.TrimNodeTextRange(file, expression)
	keyword := scanner.GetRangeOfTokenAtPosition(file, statement.Pos())
	// Remove the keyword and its whitespace, retaining comments on either
	// side of the expression and every line break after a line comment.
	triviaStart := scanner.SkipTriviaEx(file.Text(), keyword.End(), &scanner.SkipTriviaOptions{StopAtComments: true})
	text := file.Text()[triviaStart:expressionRange.Pos()] + expressionText +
		file.Text()[expressionRange.End():statement.End()]
	if file.Text()[statement.End()-1] != ';' {
		text += ";"
	}
	if needsStatementSeparator(firstToken) {
		text = ";" + text
	}
	return text
}

var NoConfusingVoidExpressionRule = rule.CreateRule(rule.Rule{
	Name:             "no-confusing-void-expression",
	Schema:           rule.NewSchema(schemaJSON),
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		canFix := func(node *ast.Node) bool {
			t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node)
			return utils.IsTypeFlagSet(t, checker.TypeFlagsVoidLike)
		}

		findInvalidAncestor := func(node *ast.Node) *ast.Node {
			parent := node

			for {
				node = parent
				parent = parent.Parent

				switch parent.Kind {
				case ast.KindParenthesizedExpression:
					continue
				case ast.KindBinaryExpression:
					n := parent.AsBinaryExpression()
					if ast.IsLogicalOrCoalescingBinaryOperator(n.OperatorToken.Kind) && n.Right == node {
						// e.g. `x && console.log(x)`
						// this is valid only if the next ancestor is valid
						continue
					}
					if n.OperatorToken.Kind == ast.KindCommaToken {
						if n.Left == node {
							return nil
						}
						// tsgo nests comma operators; ESTree flattens an unparenthesized
						// sequence. Only its final expression is a value position.
						outer := parent.Parent
						if ast.IsBinaryExpression(outer) && outer.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken && outer.AsBinaryExpression().Left == parent {
							return nil
						}
					}
				case ast.KindExpressionStatement:
					// e.g. `{ console.log("foo"); }`
					// this is always valid
					return nil
				case ast.KindConditionalExpression:
					n := parent.AsConditionalExpression()
					if n.WhenTrue == node || n.WhenFalse == node {
						// e.g. `cond ? console.log(true) : console.log(false)`
						// this is valid only if the next ancestor is valid
						continue
					}
				case ast.KindArrowFunction:
					if opts.IgnoreArrowShorthand {
						// e.g. `() => console.log("foo")`
						// this is valid with an appropriate option
						return nil
					}
				case ast.KindVoidExpression:
					if opts.IgnoreVoidOperator {
						// e.g. `() => console.log("foo")`
						// this is valid with an appropriate option
						return nil
					}
				}
				break
			}

			// Any other parent is invalid.
			// We can assume a return statement will have an argument.
			return parent
		}

		isVoidReturningFunction := func(functionNode *ast.Node) bool {
			// Game plan:
			//   - If the function node has a type annotation, check if it includes `void`.
			//     - If it does then the function is safe to return `void` expressions in.
			//   - Otherwise, check if the function is a function-expression or an arrow-function.
			//   -   If it is, get its contextual type and bail if we cannot.
			//   - Return based on whether the contextual type includes `void` or not

			returnTypeNode := functionNode.Type()
			if returnTypeNode != nil {
				returnType := checker.Checker_getTypeFromTypeNode(ctx.TypeChecker, returnTypeNode)

				return utils.Some(utils.UnionTypeParts(returnType), utils.IsIntrinsicVoidType)
			}

			if !ast.IsArrowFunction(functionNode) && !ast.IsFunctionExpression(functionNode) {
				return false
			}

			functionType := checker.Checker_getContextualType(ctx.TypeChecker, functionNode, checker.ContextFlagsNone)

			if functionType == nil {
				// getContextualType may return nil for expressions inside JSX attributes.
				// Fall back to getTypeAtLocation on the parent JsxAttribute to get the expected prop type.
				parent := functionNode.Parent
				if ast.IsJsxExpression(parent) {
					parent = parent.Parent
				}
				if ast.IsJsxAttribute(parent) {
					functionType = ctx.TypeChecker.GetTypeAtLocation(parent.AsNode())
				}
				if functionType == nil {
					return false
				}
			}
			return utils.Some(utils.UnionTypeParts(functionType), func(t *checker.Type) bool {
				callSignatures := utils.GetCallSignatures(ctx.TypeChecker, t)

				return utils.Some(callSignatures, func(s *checker.Signature) bool {
					returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, s)

					return utils.Some(utils.UnionTypeParts(returnType), utils.IsIntrinsicVoidType)
				})
			})
		}

		isFinalReturn := func(node *ast.Node) bool {
			block := node.Parent
			if !ast.IsBlock(block) {
				return false
			}

			if !ast.IsFunctionLikeDeclaration(block.Parent) {
				return false
			}

			statements := block.AsBlock().Statements.Nodes

			return len(statements) > 0 && statements[len(statements)-1] == node
		}

		checkExpression := func(node *ast.Node) {
			t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node)
			if !utils.IsTypeFlagSet(t, checker.TypeFlagsVoidLike) {
				return
			}

			invalidAncestor := findInvalidAncestor(node)
			if invalidAncestor == nil {
				// void expression is in valid position
				return
			}

			insertVoidFix := func() rule.RuleFix {
				return rule.RuleFixInsertBefore(ctx.SourceFile, node, "void ")
			}

			if ast.IsArrowFunction(invalidAncestor) {
				if opts.IgnoreVoidReturningFunctions && isVoidReturningFunction(invalidAncestor) {
					return
				}

				if opts.IgnoreVoidOperator {
					ctx.ReportNodeWithFixes(node, buildInvalidVoidExprArrowWrapVoidMessage(), insertVoidFix())
					return
				}

				var fixes []rule.RuleFix
				body := invalidAncestor.Body()
				if !ast.IsBlock(body) && canFix(body) {
					bodyText, _ := statementExpression(ctx.SourceFile, body)
					fixes = []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, body,
						"{ "+bodyText+"; }")}
				}

				ctx.ReportNodeWithFixes(node, buildInvalidVoidExprArrowMessage(), fixes...)
				return
			}

			if ast.IsReturnStatement(invalidAncestor) {
				if opts.IgnoreVoidReturningFunctions {
					functionNode := utils.GetParentFunctionNode(invalidAncestor)

					if functionNode != nil && isVoidReturningFunction(functionNode) {
						return
					}
				}

				if opts.IgnoreVoidOperator {
					ctx.ReportNodeWithFixes(node, buildInvalidVoidExprReturnWrapVoidMessage(), insertVoidFix())
					return
				}

				expr := invalidAncestor.AsReturnStatement().Expression
				finalReturn := isFinalReturn(invalidAncestor)
				var fixes []rule.RuleFix
				// Every value-discarding rewrite needs this guard, including
				// early returns whose other branch may return a non-void value.
				if canFix(expr) {
					replacement := returnStatementText(ctx.SourceFile, invalidAncestor)
					if !finalReturn {
						replacement += " return;"
						if !ast.IsBlock(invalidAncestor.Parent) {
							replacement = "{ " + replacement + " }"
						}
					}
					fixes = []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, invalidAncestor, replacement)}
				}
				if finalReturn {
					ctx.ReportNodeWithFixes(node, buildInvalidVoidExprReturnLastMessage(), fixes...)
				} else {
					ctx.ReportNodeWithFixes(node, buildInvalidVoidExprReturnMessage(), fixes...)
				}
				return
			}

			if opts.IgnoreVoidOperator {
				ctx.ReportNodeWithSuggestions(node, buildInvalidVoidExprWrapVoidMessage(), rule.RuleSuggestion{
					Message:  buildVoidExprWrapVoidMessage(),
					FixesArr: []rule.RuleFix{insertVoidFix()},
				})
				return
			}

			ctx.ReportNode(node, buildInvalidVoidExprMessage())
		}

		return rule.RuleListeners{
			ast.KindAwaitExpression:          checkExpression,
			ast.KindCallExpression:           checkExpression,
			ast.KindTaggedTemplateExpression: checkExpression,
		}
	},
})
