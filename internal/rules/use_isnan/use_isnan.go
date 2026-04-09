package use_isnan

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// unwrapToValue strips parentheses (any depth), resolves at most one level
// of comma expression (taking the last element), then strips parentheses again.
// This matches ESLint's isNaNIdentifier which does:
//   node.type === "SequenceExpression" ? node.expressions.at(-1) : node
// In ESTree parentheses are not AST nodes, but in tsgo they are, so we must
// strip them before and after the comma resolution.
func unwrapToValue(node *ast.Node) *ast.Node {
	stripped := ast.SkipParentheses(node)

	// Resolve one level of comma expression
	if stripped.Kind == ast.KindCommaListExpression {
		children := stripped.Children()
		if children != nil && len(children.Nodes) > 0 {
			return ast.SkipParentheses(children.Nodes[len(children.Nodes)-1])
		}
	}
	if stripped.Kind == ast.KindBinaryExpression {
		binary := stripped.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindCommaToken {
			return ast.SkipParentheses(binary.Right)
		}
	}

	return stripped
}

// isNaNIdentifier checks if a node represents NaN (either as the identifier
// "NaN" or as the member expression "Number.NaN").
// Recursively unwraps parentheses and comma expressions to find the value.
func isNaNIdentifier(node *ast.Node) bool {
	if node == nil {
		return false
	}

	nodeToCheck := unwrapToValue(node)

	// Check for bare NaN identifier
	if nodeToCheck.Kind == ast.KindIdentifier && nodeToCheck.Text() == "NaN" {
		return true
	}

	// Check for Number.NaN / Number['NaN'] / Number?.NaN / Number[`NaN`]
	// Uses AccessExpressionStaticName + AccessExpressionObject to handle
	// both PropertyAccessExpression and ElementAccessExpression uniformly.
	if propName, ok := utils.AccessExpressionStaticName(nodeToCheck); ok && propName == "NaN" {
		obj := utils.AccessExpressionObject(nodeToCheck)
		if obj != nil && obj.Kind == ast.KindIdentifier && obj.Text() == "Number" {
			return true
		}
	}

	return false
}

type useIsNaNOptions struct {
	enforceForSwitchCase bool
	enforceForIndexOf    bool
}

func parseOptions(opts any) useIsNaNOptions {
	result := useIsNaNOptions{
		enforceForSwitchCase: true,
		enforceForIndexOf:    false,
	}

	optsMap := utils.GetOptionsMap(opts)
	if optsMap != nil {
		if val, ok := optsMap["enforceForSwitchCase"]; ok {
			if boolVal, ok := val.(bool); ok {
				result.enforceForSwitchCase = boolVal
			}
		}
		if val, ok := optsMap["enforceForIndexOf"]; ok {
			if boolVal, ok := val.(bool); ok {
				result.enforceForIndexOf = boolVal
			}
		}
	}

	return result
}

// UseIsNaNRule requires calls to isNaN() when checking for NaN
var UseIsNaNRule = rule.Rule{
	Name: "use-isnan",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		// Comparison operators to check
		comparisonOperators := map[ast.Kind]bool{
			ast.KindEqualsEqualsToken:            true, // ==
			ast.KindEqualsEqualsEqualsToken:      true, // ===
			ast.KindExclamationEqualsToken:       true, // !=
			ast.KindExclamationEqualsEqualsToken: true, // !==
			ast.KindGreaterThanToken:             true, // >
			ast.KindGreaterThanEqualsToken:       true, // >=
			ast.KindLessThanToken:                true, // <
			ast.KindLessThanEqualsToken:          true, // <=
		}

		listeners := rule.RuleListeners{
			// Check binary expressions for NaN comparisons
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary == nil || binary.OperatorToken == nil {
					return
				}

				if !comparisonOperators[binary.OperatorToken.Kind] {
					return
				}

				if isNaNIdentifier(binary.Left) || isNaNIdentifier(binary.Right) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "comparisonWithNaN",
						Description: "Use the isNaN function to compare with NaN.",
					})
				}
			},
		}

		// Check switch statements for NaN in discriminant and case clauses
		if opts.enforceForSwitchCase {
			listeners[ast.KindSwitchStatement] = func(node *ast.Node) {
				switchStmt := node.AsSwitchStatement()
				if switchStmt == nil {
					return
				}

				// Check if discriminant is NaN
				if isNaNIdentifier(switchStmt.Expression) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "switchNaN",
						Description: "'switch(NaN)' can never match a case clause. Use Number.isNaN instead of the switch.",
					})
				}

				// Check case clauses for NaN
				if switchStmt.CaseBlock == nil {
					return
				}
				caseBlock := switchStmt.CaseBlock.AsCaseBlock()
				if caseBlock == nil || caseBlock.Clauses == nil {
					return
				}

				for _, clause := range caseBlock.Clauses.Nodes {
					if clause.Kind != ast.KindCaseClause {
						continue
					}
					caseClause := clause.AsCaseOrDefaultClause()
					if caseClause == nil || caseClause.Expression == nil {
						continue
					}
					if isNaNIdentifier(caseClause.Expression) {
						ctx.ReportNode(clause, rule.RuleMessage{
							Id:          "caseNaN",
							Description: "'case NaN' can never match. Use Number.isNaN before the switch.",
						})
					}
				}
			}
		}

		// Check indexOf/lastIndexOf calls with NaN argument
		if opts.enforceForIndexOf {
			listeners[ast.KindCallExpression] = func(node *ast.Node) {
				// Get the callee expression, skipping parentheses for (foo?.indexOf)(NaN)
				callee := ast.SkipParentheses(node.Expression())
				if callee == nil {
					return
				}

				// Extract method name from both dot notation (foo.indexOf) and
				// bracket notation (foo["indexOf"]) using shared utility.
				methodName, ok := utils.AccessExpressionStaticName(callee)
				if !ok {
					return
				}

				if methodName != "indexOf" && methodName != "lastIndexOf" {
					return
				}

				// Check if the first argument is NaN, with at most 2 arguments
				args := node.Arguments()
				if len(args) == 0 || len(args) > 2 {
					return
				}

				if isNaNIdentifier(args[0]) {
					description := strings.Replace(
						"Array prototype method '{{methodName}}' cannot find NaN.",
						"{{methodName}}",
						methodName,
						1,
					)
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "indexOfNaN",
						Description: description,
					})
				}
			}
		}

		return listeners
	},
}
