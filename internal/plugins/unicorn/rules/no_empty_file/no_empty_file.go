package no_empty_file

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_empty_file.schema.json
var schemaJSON []byte

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/rules/no-empty-file.js
var NoEmptyFileRule = rule.Rule{
	Name:   "unicorn/no-empty-file",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// Run is called once per file. Only leading, unparenthesized string
		// statements are directives; strings inside ordinary blocks are content.
		statements := ctx.SourceFile.Statements.Nodes
		inPrologue := true
		for _, statement := range statements {
			if inPrologue && ast.IsPrologueDirective(statement) {
				continue
			}
			inPrologue = false
			if !isEmptyStatement(statement) {
				return nil
			}
		}

		text := ctx.SourceFile.Text()
		allowComments := false
		if len(options) > 0 {
			opts, _ := options[0].(map[string]any)
			allowComments, _ = opts["allowComments"].(bool)
		}
		allowCommentOnlyFile := allowComments && len(statements) == 0 && scanner.GetShebang(text) == ""
		comments := ctx.Comments.All()
		var firstDirective *ast.CommentRange
		for _, comment := range comments {
			value := utils.CommentValue(text, comment)
			if comment.Kind == ast.KindSingleLineCommentTrivia &&
				strings.HasPrefix(value, "/") {
				return nil
			}
			if unicornutil.IsDisableOrEnableDirective(comment, value) {
				if firstDirective == nil {
					firstDirective = comment
				}
			} else if allowCommentOnlyFile {
				return nil
			}
		}

		// Report the entire Program, or the first disable/enable comment so
		// suppression also works after leading trivia.
		reportRange := core.NewTextRange(0, len(text))
		// typescript-eslint starts Program after leading trivia, unlike Espree.
		if !ast.IsSourceFileJS(ctx.SourceFile) {
			reportRange = reportRange.WithPos(scanner.SkipTrivia(text, 0))
		}
		if firstDirective != nil {
			reportRange = core.NewTextRange(firstDirective.Pos(), firstDirective.End())
		}
		ctx.ReportRange(reportRange, rule.RuleMessage{
			Id:          "no-empty-file",
			Description: "Empty files are not allowed.",
		})
		return nil
	},
}

func isEmptyStatement(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindEmptyStatement:
		return true
	case ast.KindBlock:
		for _, statement := range node.Statements() {
			if !isEmptyStatement(statement) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
