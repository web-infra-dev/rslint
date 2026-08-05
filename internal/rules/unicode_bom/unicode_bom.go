// Package unicode_bom implements the core ESLint rule `unicode-bom`, narrowed
// to its `never` option: a file must not begin with a Unicode byte order mark,
// U+FEFF.
//
// `always` is not supported and not planned, so the schema admits `never`
// alone and rejects anything else.
//
// The mark is never part of the text a rule sees — reading a file decodes it
// away, and caller-supplied source is stripped to match — so the question is
// answered by ctx.HasBOM, rslint's SourceCode#hasBOM. Removing it is the fix
// ESLint writes: a range starting at -1, one position ahead of the text.
package unicode_bom

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed unicode_bom.schema.json
var schemaJSON []byte

// UnicodeBomRule disallows a Unicode BOM.
// https://eslint.org/docs/latest/rules/unicode-bom
var UnicodeBomRule = rule.Rule{
	Name:   "unicode-bom",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// `never` is the only option the schema admits, and it is also the
		// default, so there is nothing to read out of options.
		if !ctx.HasBOM() {
			return rule.RuleListeners{}
		}

		// The linter never fires a KindSourceFile listener, so report eagerly.
		// The whole check is one property of the file; there is no node to
		// wait for. The zero-width range at offset 0 renders as line 1,
		// column 1 — where ESLint puts it.
		ctx.ReportRangeWithDeferredFixes(core.NewTextRange(0, 0), rule.RuleMessage{
			Id:          "unexpected",
			Description: "Unexpected Unicode BOM (Byte Order Mark).",
		}, func() []rule.RuleFix {
			return []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(-1, 0))}
		})

		return rule.RuleListeners{}
	},
}
