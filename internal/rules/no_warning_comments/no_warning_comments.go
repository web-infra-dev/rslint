package no_warning_comments

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
	"github.com/web-infra-dev/rslint/internal/utils/warningcomments"
)

//go:embed no_warning_comments.schema.json
var schemaJSON []byte

// NoWarningCommentsRule disallows specified warning terms in comments.
// https://eslint.org/docs/latest/rules/no-warning-comments
var NoWarningCommentsRule = rule.Rule{
	Name:   "no-warning-comments",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		decoration := strings.Join(opts.Decoration, "")
		warningRegExps := make([]*esregexp.RegExp, len(opts.Terms))
		for i, term := range opts.Terms {
			warningRegExps[i] = warningcomments.Pattern(term, opts.Location, decoration)
		}

		text := ctx.SourceFile.Text()
		for _, comment := range ctx.Comments.All() {
			checkComment(ctx, text, comment, opts.Terms, warningRegExps)
		}

		return rule.RuleListeners{}
	},
}

type ruleOptions struct {
	Terms      []string
	Location   string
	Decoration []string
}

func parseOptions(options []any) ruleOptions {
	opts := ruleOptions{
		Terms:    []string{"todo", "fixme", "xxx"},
		Location: "start",
	}
	if len(options) == 0 {
		return opts
	}
	optsMap, _ := options[0].(map[string]any)
	if arr, ok := optsMap["terms"].([]any); ok {
		terms := make([]string, 0, len(arr))
		for _, v := range arr {
			if s, ok := v.(string); ok {
				terms = append(terms, s)
			}
		}
		opts.Terms = terms
	}
	if loc, ok := optsMap["location"].(string); ok {
		opts.Location = loc
	}
	if arr, ok := optsMap["decoration"].([]any); ok {
		decoration := make([]string, 0, len(arr))
		for _, v := range arr {
			if s, ok := v.(string); ok {
				decoration = append(decoration, s)
			}
		}
		opts.Decoration = decoration
	}
	return opts
}

// selfConfigRegex mirrors upstream's hardcoded self-reference guard: a
// directive comment that mentions this rule's own name is exempt from
// matching, so a comment documenting `no-warning-comments` configuration
// doesn't accidentally trip its own configured terms.
var selfConfigRegex = esregexp.MustCompile(`\bno-warning-comments\b`, "u")

func checkComment(ctx rule.RuleContext, text string, comment *ast.CommentRange, terms []string, warningRegExps []*esregexp.RegExp) {
	value := utils.CommentValue(text, comment)

	if utils.IsDirectiveComment(comment.Kind, ecmascript.StringTrim(value)) && selfConfigRegex.Test(value) {
		return
	}

	for i, re := range warningRegExps {
		if re.Test(value) {
			reportMatch(ctx, comment, terms[i], value)
		}
	}
}

func reportMatch(ctx rule.RuleContext, comment *ast.CommentRange, matchedTerm, value string) {
	display := warningcomments.Display(value)
	ctx.ReportRange(core.NewTextRange(comment.Pos(), comment.End()), rule.RuleMessage{
		Id:          "unexpectedComment",
		Description: fmt.Sprintf("Unexpected '%s' comment: '%s'.", matchedTerm, display),
		Data: map[string]string{
			"matchedTerm": matchedTerm,
			"comment":     display,
		},
	})
}
