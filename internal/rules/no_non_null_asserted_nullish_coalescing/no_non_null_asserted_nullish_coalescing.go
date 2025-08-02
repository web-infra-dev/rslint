package no_non_null_asserted_nullish_coalescing

import (
	"strings"
	
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/typescript-eslint/rslint/internal/rule"
)

func buildNoNonNullAssertedNullishCoalescingMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noNonNullAssertedNullishCoalescing",
		Description: "The nullish coalescing operator is designed to handle undefined and null - using a non-null assertion is not needed.",
	}
}

func buildSuggestRemovingNonNullMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "suggestRemovingNonNull",
		Description: "Remove the non-null assertion.",
	}
}

var NoNonNullAssertedNullishCoalescingRule = rule.Rule{
	Name: "no-non-null-asserted-nullish-coalescing",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Helper function to check if a variable has assignment before the node
		hasAssignmentBeforeNode := func(identifier *ast.Identifier, node *ast.Node) bool {
			// Get the symbol for the identifier
			symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier.AsNode())
			if symbol == nil {
				// If we can't find the symbol, assume it's assigned (external variable, etc.)
				// This handles cases like `foo! ?? bar` where foo is undefined
				return true
			}

			// Check if it's a value declaration with an initializer or definite assignment
			declarations := symbol.Declarations
			for _, decl := range declarations {
				if ast.IsVariableDeclaration(decl) {
					varDecl := decl.AsVariableDeclaration()
					// Check if it has an initializer or is definitely assigned
					if varDecl.Initializer != nil || (varDecl.ExclamationToken != nil && varDecl.ExclamationToken.Kind == ast.KindExclamationToken) {
						// Check if declaration is before the node
						if varDecl.End() < node.Pos() {
							return true
						}
					}
				}
			}

			// Check for assignment expressions (x = foo()) before this node
			// This is a simplified check - in a full implementation, we'd need to traverse
			// the AST more thoroughly to find all assignments in the current scope
			sourceFile := ctx.SourceFile
			sourceText := sourceFile.Text()
			nodeStart := node.Pos()
			
			// Look for assignment patterns like "x = " before this node
			varName := identifier.Text
			beforeCode := sourceText[:nodeStart]
			
			// Simple regex-like check for assignment pattern
			// Look for patterns like "varName =" or "varName=" in the code before this node
			assignmentPattern := varName + " ="
			assignmentPattern2 := varName + "="
			
			if strings.Contains(beforeCode, assignmentPattern) || strings.Contains(beforeCode, assignmentPattern2) {
				return true
			}
			
			return false
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				if node.Kind != ast.KindBinaryExpression {
					return
				}

				binaryExpr := node.AsBinaryExpression()
				
				// Check if it's a nullish coalescing operator
				if binaryExpr.OperatorToken.Kind != ast.KindQuestionQuestionToken {
					return
				}

				// Check if the left operand is a non-null assertion
				leftOperand := binaryExpr.Left
				if !ast.IsNonNullExpression(leftOperand) {
					return
				}

				nonNullExpr := leftOperand.AsNonNullExpression()
				
				// Special case: if the expression is an identifier, check if it has been assigned
				// Only skip the rule for uninitialized identifiers
				if ast.IsIdentifier(nonNullExpr.Expression) {
					identifier := nonNullExpr.Expression.AsIdentifier()
					if !hasAssignmentBeforeNode(identifier, node) {
						return
					}
				}
				// For non-identifier expressions (foo(), obj.prop, etc.), always trigger the rule

				// Create a range for the exclamation token only (preserve spacing)
				// We need to find the exact position of the '!' character
				sourceText := ctx.SourceFile.Text()
				exprEnd := nonNullExpr.Expression.End()
				leftEnd := leftOperand.End()
				
				// Find the '!' character position by scanning from the expression end
				exclamationStart := exprEnd
				exclamationEnd := leftEnd
				
				// Scan to find the actual '!' character (skip whitespace)
				for i := exprEnd; i < leftEnd; i++ {
					if sourceText[i] == '!' {
						exclamationStart = i
						exclamationEnd = i + 1
						break
					}
				}
				
				exclamationRange := core.NewTextRange(exclamationStart, exclamationEnd)

				// Report the issue with a suggestion
				ctx.ReportNodeWithSuggestions(leftOperand, buildNoNonNullAssertedNullishCoalescingMessage(), rule.RuleSuggestion{
					Message: buildSuggestRemovingNonNullMessage(),
					FixesArr: []rule.RuleFix{
						rule.RuleFixReplaceRange(exclamationRange, ""),
					},
				})
			},
		}
	},
}