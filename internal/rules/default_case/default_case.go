package default_case

import (
	_ "embed"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

//go:embed default_case.schema.json
var schemaJSON []byte

// https://eslint.org/docs/latest/rules/default-case
var DefaultCaseRule = rule.Rule{
	Name:   "default-case",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
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

				// No default clause found - check the last comment after the last clause (matching ESLint behavior)
				lastClause := clauses.Clauses.Nodes[len(clauses.Clauses.Nodes)-1]
				lastClauseEnd := lastClause.End()
				switchEnd := node.End()

				var lastComment *ast.CommentRange
				commentRange := utils.GetCommentsInRange(ctx.SourceFile, core.NewTextRange(lastClauseEnd, switchEnd))
				for comment := range commentRange {
					c := comment
					lastComment = &c
				}

				if lastComment != nil {
					commentText := strings.TrimSpace(ctx.SourceFile.Text()[lastComment.Pos():lastComment.End()])
					// Remove comment markers
					if strings.HasPrefix(commentText, "//") {
						commentText = strings.TrimSpace(commentText[2:])
					} else if strings.HasPrefix(commentText, "/*") && strings.HasSuffix(commentText, "*/") {
						commentText = strings.TrimSpace(commentText[2 : len(commentText)-2])
					}

					if opts.commentPattern.Test(commentText) {
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
	commentPattern *esregexp.RegExp
}

// defaultCommentPattern mirrors upstream's DEFAULT_COMMENT_PATTERN
// (/^no default$/iu). A configured commentPattern is compiled without the
// `i` flag, exactly like upstream's `new RegExp(commentPattern, "u")`.
var defaultCommentPattern = esregexp.MustCompile(`^no default$`, "iu")

func parseOptions(opts []any) options {
	result := options{commentPattern: defaultCommentPattern}
	if len(opts) == 0 {
		return result
	}

	m, _ := opts[0].(map[string]any)
	if pattern, ok := m["commentPattern"].(string); ok && pattern != "" {
		// An invalid pattern is rejected by the schema's `format: "regex"`
		// before linting starts, so this compile error is only defensive.
		if compiled, err := esregexp.Compile(pattern, "u"); err == nil {
			result.commentPattern = compiled
		}
	}

	return result
}
