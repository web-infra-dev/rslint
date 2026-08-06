// Package unicode_bom implements the core ESLint rule `unicode-bom`: a file
// must begin with a Unicode byte order mark, U+FEFF, under `always`, and must
// not under `never`.
//
// The mark is never part of the text a rule sees — reading a file decodes it
// away, and caller-supplied source is stripped to match — so the question is
// answered by ctx.HasBOM, rslint's SourceCode#hasBOM. Both fixes are the ones
// ESLint writes: writing a mark in is an insertion at the front of the text,
// and cutting one out is a range starting at -1, one position ahead of it.
package unicode_bom

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed unicode_bom.schema.json
var schemaJSON []byte

// parseOptions reports whether the file is required to carry a mark. `never`
// is the default, matching upstream's `defaultOptions`.
func parseOptions(options []any) bool {
	if len(options) == 0 {
		return false
	}
	mode, _ := options[0].(string)
	return mode == "always"
}

// UnicodeBomRule requires or disallows a Unicode BOM.
// https://eslint.org/docs/latest/rules/unicode-bom
var UnicodeBomRule = rule.Rule{
	Name:   "unicode-bom",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		requireBOM := parseOptions(options)
		hasBOM := ctx.HasBOM()

		// The linter never fires a KindSourceFile listener, so report eagerly.
		// The whole check is one property of the file; there is no node to
		// wait for. The zero-width range at offset 0 renders as line 1,
		// column 1 — where ESLint puts it.
		reportRange := core.NewTextRange(0, 0)

		switch {
		case requireBOM && !hasBOM:
			ctx.ReportRangeWithDeferredFixes(reportRange, rule.RuleMessage{
				Id:          "expected",
				Description: "Expected Unicode BOM (Byte Order Mark).",
			}, func() []rule.RuleFix {
				return []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(0, 0), utils.BOM)}
			})
		case !requireBOM && hasBOM:
			ctx.ReportRangeWithDeferredFixes(reportRange, rule.RuleMessage{
				Id:          "unexpected",
				Description: "Unexpected Unicode BOM (Byte Order Mark).",
			}, func() []rule.RuleFix {
				return []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(-1, 0))}
			})
		}

		return rule.RuleListeners{}
	},
}
