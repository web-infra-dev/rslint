package use_isnan

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// isNaNIdentifier checks if a node represents NaN (either as the identifier
// "NaN" or as the member expression "Number.NaN").
func isNaNIdentifier(node *ast.Node) bool {
	if node == nil {
		return false
	}

	// Check for bare NaN identifier
	if node.Kind == ast.KindIdentifier && node.Text() == "NaN" {
		return true
	}

	// Check for Number.NaN
	if node.Kind == ast.KindPropertyAccessExpression {
		propAccess := node.AsPropertyAccessExpression()
		if propAccess == nil {
			return false
		}
		if propAccess.Expression != nil &&
			propAccess.Expression.Kind == ast.KindIdentifier &&
			propAccess.Expression.Text() == "Number" &&
			propAccess.Name() != nil &&
			propAccess.Name().Text() == "NaN" {
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
				// Get the callee expression
				callee := node.Expression()
				if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
					return
				}

				propAccess := callee.AsPropertyAccessExpression()
				if propAccess == nil || propAccess.Name() == nil {
					return
				}

				methodName := propAccess.Name().Text()
				if methodName != "indexOf" && methodName != "lastIndexOf" {
					return
				}

				// Check if the first argument is NaN
				args := node.Arguments()
				if len(args) == 0 {
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
