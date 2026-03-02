package default_case

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/default-case
var DefaultCaseRule = rule.Rule{
	Name: "default-case",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindSwitchStatement: func(node *ast.Node) {
				switchStmt := node.AsSwitchStatement()
				if switchStmt == nil {
					return
				}

				caseBlock := switchStmt.CaseBlock
				if caseBlock == nil {
					return
				}

				clauses := caseBlock.AsCaseBlock()
				if clauses == nil || clauses.Clauses == nil {
					return
				}

				// Check if there are any clauses at all
				if len(clauses.Clauses.Nodes) == 0 {
					return
				}

				// Check if a default clause exists
				for _, clause := range clauses.Clauses.Nodes {
					if clause.Kind == ast.KindDefaultClause {
						return
					}
				}

				// No default clause found - check for "no default" comment
				lastClause := clauses.Clauses.Nodes[len(clauses.Clauses.Nodes)-1]

				// Check comments after the last clause
				lastClauseEnd := lastClause.End()
				switchEnd := node.End()

				commentRange := utils.GetCommentsInRange(ctx.SourceFile, core.NewTextRange(lastClauseEnd, switchEnd))
				for comment := range commentRange {
					commentText := strings.TrimSpace(ctx.SourceFile.Text()[comment.Pos():comment.End()])
					// Remove comment markers
					if strings.HasPrefix(commentText, "//") {
						commentText = strings.TrimSpace(commentText[2:])
					} else if strings.HasPrefix(commentText, "/*") && strings.HasSuffix(commentText, "*/") {
						commentText = strings.TrimSpace(commentText[2 : len(commentText)-2])
					}

					if opts.commentPattern.MatchString(commentText) {
						return
					}
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "missingDefaultCase",
					Description: "Expected a default case.",
				})
			},
		}
	},
}

type options struct {
	commentPattern *regexp.Regexp
}

func parseOptions(opts any) options {
	result := options{
		commentPattern: regexp.MustCompile(`(?i)^no default$`),
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
		if pattern, ok := optsMap["commentPattern"].(string); ok && pattern != "" {
			if compiled, err := regexp.Compile(pattern); err == nil {
				result.commentPattern = compiled
			}
		}
	}

	return result
}
