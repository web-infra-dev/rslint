package no_empty

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-empty
var NoEmptyRule = rule.Rule{
	Name: "no-empty",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindBlock: func(node *ast.Node) {
				block := node.AsBlock()
				if block == nil {
					return
				}

				// If the body is not empty, return
				if block.Statements != nil && len(block.Statements.Nodes) > 0 {
					return
				}

				// A function is generally allowed to be empty
				if isFunction(node.Parent) {
					return
				}

				// Allow empty catch blocks if option is set
				if opts.allowEmptyCatch && node.Parent != nil && node.Parent.Kind == ast.KindCatchClause {
					return
				}

				// Allow blocks with comments inside
				// For empty blocks, we check the source text for comment patterns
				// This is safe because there are no statements (so no string literals to confuse)
				if hasCommentInside(ctx.SourceFile.Text(), node.Pos(), node.End()) {
					return
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "unexpected",
					Description: "Empty block statement.",
				})
			},

			ast.KindSwitchStatement: func(node *ast.Node) {
				switchStmt := node.AsSwitchStatement()
				if switchStmt == nil || switchStmt.CaseBlock == nil {
					return
				}

				caseBlock := switchStmt.CaseBlock.AsCaseBlock()
				if caseBlock == nil {
					return
				}

				// Report if the switch has no cases
				if caseBlock.Clauses == nil || len(caseBlock.Clauses.Nodes) == 0 {
					// Allow switch statements with comments inside
					if hasCommentInside(ctx.SourceFile.Text(), switchStmt.CaseBlock.Pos(), switchStmt.CaseBlock.End()) {
						return
					}

					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpected",
						Description: "Empty switch statement.",
					})
				}
			},
		}
	},
}

// hasCommentInside checks if there is a comment between the opening and closing braces
func hasCommentInside(text string, pos, end int) bool {
	if pos < 0 || end > len(text) || pos >= end {
		return false
	}

	bodyText := text[pos:end]
	// Find the opening brace
	openIdx := strings.IndexByte(bodyText, '{')
	if openIdx < 0 {
		return false
	}

	// Check text between { and } for comment patterns
	inner := bodyText[openIdx+1:]
	return strings.Contains(inner, "//") || strings.Contains(inner, "/*")
}

func isFunction(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		return true
	}
	return false
}

type noEmptyOptions struct {
	allowEmptyCatch bool
}

func parseOptions(opts any) noEmptyOptions {
	result := noEmptyOptions{
		allowEmptyCatch: false,
	}

	if opts == nil {
		return result
	}

	var optsMap map[string]interface{}
	if arr, ok := opts.([]interface{}); ok && len(arr) > 0 {
		optsMap, _ = arr[0].(map[string]interface{})
	} else {
		optsMap, _ = opts.(map[string]interface{})
	}

	if optsMap != nil {
		if allow, ok := optsMap["allowEmptyCatch"].(bool); ok {
			result.allowEmptyCatch = allow
		}
	}

	return result
}
