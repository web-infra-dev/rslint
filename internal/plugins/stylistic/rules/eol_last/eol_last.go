package eol_last

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const (
	optionAlways = "always"
	optionNever  = "never"
)

// parseMode unwraps the single positional string option. Accepted shapes:
//
//	['eol-last']            → 'always' (default)
//	['eol-last', 'never']   → 'never'
//	['eol-last', 'always']  → 'always'
//
// rslint's config loader collapses the single trailing option element, so the
// rule may receive either []interface{}{"never"} (Go test / multi-element
// config) or a bare string (CLI single-option config). Anything that does not
// resolve to a known string keeps the upstream-equivalent behavior of leaving
// `mode` un-matched (no diagnostic on either branch); we only short-circuit
// to the default when no value was supplied at all.
func parseMode(options any) string {
	switch v := options.(type) {
	case string:
		return v
	case []interface{}:
		if len(v) > 0 {
			if s, ok := v[0].(string); ok {
				return s
			}
		}
	}
	return optionAlways
}

// EolLastRule enforces or disallows a trailing newline at the end of files.
// https://eslint.style/rules/eol-last
//
// This is a whole-file rule: it does not visit AST nodes. rslint does not fire
// a KindSourceFile listener, so the check runs eagerly inside Run and returns
// an empty listener map.
var EolLastRule = rule.Rule{
	Name: "@stylistic/eol-last",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		mode := parseMode(options)
		text := ctx.SourceFile.Text()
		n := len(text)

		if n == 0 {
			return rule.RuleListeners{}
		}

		endsWithLF := strings.HasSuffix(text, "\n")

		switch mode {
		case optionAlways:
			if !endsWithLF {
				// Diagnostic loc is a zero-width span at src.length, matching
				// upstream's `loc: getLocFromIndex(src.length)`. The renderer
				// resolves position n to the line/column just past the final
				// byte, even when that byte is on the same line.
				ctx.ReportRangeWithFixes(
					core.NewTextRange(n, n),
					rule.RuleMessage{
						Id:          "missing",
						Description: "Newline required at end of file but not found.",
					},
					rule.RuleFix{
						Text:  "\n",
						Range: core.NewTextRange(n, n),
					},
				)
			}
		case optionNever:
			if endsWithLF {
				// Diagnostic span covers exactly the final newline sequence
				// (`\r\n` when present, else `\n`), so end-of-range falls on
				// the start of the next line — matching upstream's
				// `[length-(CRLF?2:1), length]` loc.
				startPos := n - 1
				if strings.HasSuffix(text, "\r\n") {
					startPos = n - 2
				}
				// Fix removes the entire trailing run of `(\r?\n)+`, matching
				// upstream's `/(?:\r?\n)+$/u` greedy strip. This collapses
				// multiple terminators (e.g. `\n\n`, `\r\n\r\n`) in one fix.
				fixStart := trailingNewlineRunStart(text)
				ctx.ReportRangeWithFixes(
					core.NewTextRange(startPos, n),
					rule.RuleMessage{
						Id:          "unexpected",
						Description: "Newline not allowed at end of file.",
					},
					rule.RuleFix{
						Text:  "",
						Range: core.NewTextRange(fixStart, n),
					},
				)
			}
		}

		return rule.RuleListeners{}
	},
}

// trailingNewlineRunStart returns the byte index where the trailing
// `(\r?\n)+` sequence begins. Mirrors upstream's `/(?:\r?\n)+$/u.exec(src)`
// — each step consumes a `\n` and, when immediately preceded by `\r`, also
// consumes that `\r`. Callers must only invoke this when text actually ends
// with `\n`; otherwise it returns len(text) unchanged.
func trailingNewlineRunStart(text string) int {
	pos := len(text)
	for pos > 0 && text[pos-1] == '\n' {
		pos--
		if pos > 0 && text[pos-1] == '\r' {
			pos--
		}
	}
	return pos
}
