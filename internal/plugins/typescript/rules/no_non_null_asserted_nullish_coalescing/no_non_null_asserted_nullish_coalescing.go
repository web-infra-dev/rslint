package no_non_null_asserted_nullish_coalescing

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// findVariableAndCheckAssignment looks for a variable declaration matching identName
// in the enclosing scopes and checks if it has been assigned before the given node.
// Returns (found, hasAssignment).
func findVariableAndCheckAssignment(identName string, node *ast.Node) (bool, bool) {
	current := node.Parent
	for current != nil {
		var statements []*ast.Node
		switch current.Kind {
		case ast.KindBlock, ast.KindSourceFile, ast.KindModuleBlock:
			statements = current.Statements()
		}
		for _, stmt := range statements {
			// Check variable declarations (check all, regardless of position)
			if stmt.Kind == ast.KindVariableStatement {
				declList := stmt.AsVariableStatement().DeclarationList
				if declList != nil {
					for _, decl := range declList.AsVariableDeclarationList().Declarations.Nodes {
						varDecl := decl.AsVariableDeclaration()
						name := varDecl.Name()
						if name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == identName {
							// Found the declaration
							if varDecl.Initializer != nil || varDecl.ExclamationToken != nil {
								return true, true
							}
							// Declaration found but no init/exclamation — check for assignments before node
							hasAssignment := checkAssignmentsBefore(identName, node, current)
							return true, hasAssignment
						}
					}
				}
			}
		}
		current = current.Parent
	}
	// Variable not found in any scope
	return false, false
}

// checkAssignmentsBefore checks if there is an assignment expression (e.g., x = foo())
// for the given identifier that appears before the node in the given scope.
func checkAssignmentsBefore(identName string, node *ast.Node, scope *ast.Node) bool {
	var statements []*ast.Node
	switch scope.Kind {
	case ast.KindBlock, ast.KindSourceFile, ast.KindModuleBlock:
		statements = scope.Statements()
	}
	if statements == nil {
		return false
	}
	for _, stmt := range statements {
		// Stop at statements that are at or after the node
		if stmt.Pos() >= node.Pos() {
			break
		}
		// Check assignment expressions
		if stmt.Kind == ast.KindExpressionStatement {
			expr := stmt.AsExpressionStatement().Expression
			if ast.IsBinaryExpression(expr) {
				binExpr := expr.AsBinaryExpression()
				if binExpr.OperatorToken.Kind == ast.KindEqualsToken {
					left := binExpr.Left
					if left.Kind == ast.KindIdentifier && left.AsIdentifier().Text == identName {
						return true
					}
				}
			}
		}
	}
	return false
}

var NoNonNullAssertedNullishCoalescingRule = rule.CreateRule(rule.Rule{
	Name: "no-non-null-asserted-nullish-coalescing",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		msg := rule.RuleMessage{
			Id:          "noNonNullAssertedNullishCoalescing",
			Description: "The nullish coalescing operator is designed to handle undefined and null - using a non-null assertion is not needed.",
		}
		suggestMsg := rule.RuleMessage{
			Id:          "suggestRemovingNonNull",
			Description: "Remove the non-null assertion.",
		}

		return rule.RuleListeners{
			ast.KindNonNullExpression: func(node *ast.Node) {
				parent := node.Parent
				// Check if parent is BinaryExpression with ?? operator and this node is the left operand
				if parent.Kind != ast.KindBinaryExpression {
					return
				}
				binExpr := parent.AsBinaryExpression()
				if binExpr.OperatorToken.Kind != ast.KindQuestionQuestionToken {
					return
				}
				if binExpr.Left != node {
					return
				}

				// If the expression is an Identifier, check scope for prior assignment
				expression := node.AsNonNullExpression().Expression
				if expression.Kind == ast.KindIdentifier {
					identName := expression.AsIdentifier().Text
					found, hasAssignment := findVariableAndCheckAssignment(identName, node)
					if found && !hasAssignment {
						return // Variable declared but not yet assigned — skip
					}
				}

				// Build suggestion to remove the ! token
				s := scanner.GetScannerForSourceFile(ctx.SourceFile, expression.End())
				fix := rule.RuleFixRemoveRange(s.TokenRange())

				ctx.ReportNodeWithSuggestions(node, msg,
					rule.RuleSuggestion{
						Message:  suggestMsg,
						FixesArr: []rule.RuleFix{fix},
					},
				)
			},
		}
	},
})
