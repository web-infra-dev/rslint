package no_empty_file

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
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
			if isDisableOrEnableComment(comment, value) {
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

// Only classify the comment here; RuleContext handles suppression. General
// directive helpers also include config/global comments, which this rule allows.
func isDisableOrEnableComment(comment *ast.CommentRange, value string) bool {
	label := ecmascript.StringTrim(value)
	if end := strings.IndexFunc(label, ecmascript.IsWhiteSpaceOrLineTerminator); end >= 0 {
		label = label[:end]
	}
	switch label {
	case "eslint-disable", "eslint-enable", "rslint-disable", "rslint-enable":
		return comment.Kind == ast.KindMultiLineCommentTrivia
	case "eslint-disable-next-line", "rslint-disable-next-line":
		return true
	case "eslint-disable-line", "rslint-disable-line":
		return !strings.ContainsAny(value, ecmascript.LineTerminators)
	default:
		return false
	}
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
