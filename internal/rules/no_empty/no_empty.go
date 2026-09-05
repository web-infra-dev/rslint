package no_empty

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed no_empty.schema.json
var schemaJSON []byte

// https://eslint.org/docs/latest/rules/no-empty
var NoEmptyRule = rule.Rule{
	Name:   "no-empty",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
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

				// Function bodies and class static blocks are distinct ESTree node
				// types, so ESLint's BlockStatement listener does not inspect them.
				if ast.IsFunctionLikeOrClassStaticBlockDeclaration(node.Parent) {
					return
				}

				// Allow empty catch blocks if option is set
				if opts.allowEmptyCatch && node.Parent != nil && node.Parent.Kind == ast.KindCatchClause {
					return
				}

				text := ctx.SourceFile.Text()
				blockRange := core.NewTextRange(scanner.SkipTrivia(text, node.Pos()), node.End())
				innerRange := bracedInnerRange(blockRange)
				// An empty Block cannot contain strings, templates, regex literals,
				// or JSX, so a comment opener in its inner source is a real comment.
				if hasCommentInside(text, innerRange.Pos(), innerRange.End()) {
					return
				}

				reportEmptyStatement(ctx, blockRange, innerRange, "block")
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

				if caseBlock.Clauses != nil && len(caseBlock.Clauses.Nodes) != 0 {
					return
				}

				text := ctx.SourceFile.Text()
				caseBlockRange := core.NewTextRange(scanner.SkipTrivia(text, switchStmt.CaseBlock.Pos()), switchStmt.CaseBlock.End())
				innerRange := bracedInnerRange(caseBlockRange)
				if hasCommentInside(text, innerRange.Pos(), innerRange.End()) {
					return
				}

				reportEmptyStatement(ctx, caseBlockRange, innerRange, "switch")
			},
		}
	},
}

func reportEmptyStatement(ctx rule.RuleContext, statementRange, innerRange core.TextRange, statementType string) {
	ctx.ReportRangeWithDeferredSuggestions(statementRange, emptyStatementMessage(statementType), func() []rule.RuleSuggestion {
		return []rule.RuleSuggestion{{
			Message: suggestCommentMessage(statementType),
			FixesArr: []rule.RuleFix{
				rule.RuleFixReplaceRange(innerRange, " /* empty */ "),
			},
		}}
	})
}

func emptyStatementMessage(statementType string) rule.RuleMessage {
	description := "Empty block statement."
	if statementType == "switch" {
		description = "Empty switch statement."
	}
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: description,
		Data:        map[string]string{"type": statementType},
	}
}

func suggestCommentMessage(statementType string) rule.RuleMessage {
	description := "Add comment inside empty block statement."
	if statementType == "switch" {
		description = "Add comment inside empty switch statement."
	}
	return rule.RuleMessage{
		Id:          "suggestComment",
		Description: description,
		Data:        map[string]string{"type": statementType},
	}
}

func bracedInnerRange(statementRange core.TextRange) core.TextRange {
	if statementRange.End() <= statementRange.Pos()+1 {
		return core.NewTextRange(statementRange.Pos(), statementRange.Pos())
	}
	return core.NewTextRange(statementRange.Pos()+1, statementRange.End()-1)
}

// hasCommentInside reports whether the half-open source range contains a line
// or block comment opener. Callers pass only the trivia inside a syntactically
// empty braced node, where either opener necessarily denotes a real comment.
func hasCommentInside(text string, pos, end int) bool {
	if pos < 0 || end > len(text) || pos >= end {
		return false
	}

	inner := text[pos:end]
	for searchStart := 0; searchStart < len(inner); {
		offset := strings.IndexByte(inner[searchStart:], '/')
		if offset < 0 {
			return false
		}
		slash := searchStart + offset
		if slash+1 < len(inner) && (inner[slash+1] == '/' || inner[slash+1] == '*') {
			return true
		}
		searchStart = slash + 1
	}
	return false
}

type noEmptyOptions struct {
	allowEmptyCatch bool
}

func parseOptions(options []any) noEmptyOptions {
	result := noEmptyOptions{
		allowEmptyCatch: false,
	}
	if len(options) == 0 {
		return result
	}

	optsMap, _ := options[0].(map[string]any)
	if allow, ok := optsMap["allowEmptyCatch"].(bool); ok {
		result.allowEmptyCatch = allow
	}

	return result
}
