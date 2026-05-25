package block_spacing

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	spacingAlways = "always"
	spacingNever  = "never"
)

// parseOptions mirrors upstream's single-string schema:
//
//	rule: ['block-spacing']              → always (default)
//	rule: ['block-spacing', 'always']    → always
//	rule: ['block-spacing', 'never']     → never
//
// rslint's config loader collapses a single trailing option element into the
// option directly, so accept either a string or a one-string array.
func parseOptions(raw any) bool {
	switch v := raw.(type) {
	case []interface{}:
		if len(v) > 0 {
			if s, ok := v[0].(string); ok && s == spacingNever {
				return false
			}
		}
	case string:
		if v == spacingNever {
			return false
		}
	}
	return true
}

// nextTokenStart walks forward from `low` skipping ASCII + Unicode whitespace
// and line terminators, stopping at the first non-trivia byte OR at the start
// of a comment. This matches ESLint's
// `sourceCode.getTokenAfter(openBrace, { includeComments: true })` semantics:
// comments count as the next token. Delegates to tsgo's `SkipTriviaEx` with
// `StopAtComments: true` so Unicode whitespace forms (e.g. ` `, ` `)
// are handled identically to the parser. Clamps at `high` so an unterminated
// `{ … ` defensively returns `high` rather than scanning past the brace.
func nextTokenStart(text string, low, high int) int {
	p := scanner.SkipTriviaEx(text, low, &scanner.SkipTriviaOptions{StopAtComments: true})
	if p > high {
		return high
	}
	return p
}

// resolveBraces returns the byte positions of the `{` and `}` that bound a
// Block / CaseBlock node. Returns ok=false when the trimmed source range
// doesn't actually start/end with braces (defensive — parser recovery on
// malformed input).
func resolveBraces(text string, sf *ast.SourceFile, node *ast.Node) (openPos, closePos int, ok bool) {
	trimmed := utils.TrimNodeTextRange(sf, node)
	start := trimmed.Pos()
	end := trimmed.End()
	if start >= end || end > len(text) {
		return 0, 0, false
	}
	if text[start] != '{' || text[end-1] != '}' {
		return 0, 0, false
	}
	return start, end - 1, true
}

// BlockSpacingRule enforces consistent spacing inside open/close block tokens
// when they sit on the same line as their adjacent content token. Ported from
// @stylistic/eslint-plugin's block-spacing.
//
// In the upstream ESTree rule three listeners are registered — BlockStatement,
// StaticBlock, SwitchStatement — and each one filters its first/last brace
// out of the wrapping node. In tsgo the same surface is reached more
// directly:
//
//   - `ast.KindBlock` covers every BlockStatement, every function / arrow /
//     method body, AND the body of a ClassStaticBlockDeclaration (which is
//     itself an `*ast.BlockNode` field, visited as a regular Block child).
//   - `ast.KindCaseBlock` is the `{ case … }` wrapper inside a SwitchStatement.
//     Listening on it directly avoids the upstream pattern of skipping over
//     `switch (e)` to find the first `{`.
//
// Both kinds have their `{` at `trimmed.Pos()` and their `}` at
// `trimmed.End()-1`, so the brace locations come straight out of
// `utils.TrimNodeTextRange`.
var BlockSpacingRule = rule.Rule{
	Name: "@stylistic/block-spacing",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		always := parseOptions(rawOptions)
		text := ctx.SourceFile.Text()

		// Opening-side check runs on listener enter; closing-side runs on
		// exit. For nested blocks this interleaves the reports into source
		// order (outer-open → inner-open → inner-close → outer-close),
		// matching how ESLint sorts diagnostics by location and how
		// array_bracket_spacing.go solves the same problem.
		enter := func(node *ast.Node) {
			openPos, closePos, ok := resolveBraces(text, ctx.SourceFile, node)
			if !ok {
				return
			}
			innerLow := openPos + 1
			innerHigh := closePos

			firstTokenStart := nextTokenStart(text, innerLow, innerHigh)
			// Empty block (only whitespace between the braces) — upstream
			// short-circuits via `firstToken === closeBrace`.
			if firstTokenStart >= innerHigh {
				return
			}

			// Only fire when `{` and the first content token are on the
			// same line — upstream's `isTokenOnSameLine(openBrace,
			// firstToken)` short-circuit.
			if utils.ContainsLineTerminator(text, innerLow, firstTokenStart) {
				return
			}

			// `never` + line-comment-starts-immediately-after-`{` is a
			// dedicated upstream exemption: deleting the leading space
			// before a `//` comment would fuse the comment into the
			// `{`-line and silently change layout, so upstream skips the
			// opening check in this exact shape. We mirror it by peeking
			// at the next two bytes of the comment-bearing token.
			if !always &&
				firstTokenStart+1 < innerHigh &&
				text[firstTokenStart] == '/' &&
				text[firstTokenStart+1] == '/' {
				return
			}

			hasSpace := firstTokenStart > innerLow
			switch {
			case always && !hasSpace:
				ctx.ReportRangeWithFixes(
					core.NewTextRange(openPos, openPos+1),
					rule.RuleMessage{
						Id:          "missing",
						Description: "Requires a space after '{'.",
						Data:        map[string]string{"location": "after", "token": "{"},
					},
					rule.RuleFix{
						Text:  " ",
						Range: core.NewTextRange(firstTokenStart, firstTokenStart),
					},
				)
			case !always && hasSpace:
				ctx.ReportRangeWithFixes(
					core.NewTextRange(innerLow, firstTokenStart),
					rule.RuleMessage{
						Id:          "extra",
						Description: "Unexpected space(s) after '{'.",
						Data:        map[string]string{"location": "after", "token": "{"},
					},
					rule.RuleFix{
						Text:  "",
						Range: core.NewTextRange(innerLow, firstTokenStart),
					},
				)
			}
		}

		exit := func(node *ast.Node) {
			openPos, closePos, ok := resolveBraces(text, ctx.SourceFile, node)
			if !ok {
				return
			}
			innerLow := openPos + 1
			innerHigh := closePos

			firstTokenStart := nextTokenStart(text, innerLow, innerHigh)
			// Empty block: nothing to check on the closing side either —
			// upstream's `firstToken === closeBrace` subsumes the close
			// check.
			if firstTokenStart >= innerHigh {
				return
			}
			lastTokenEnd := utils.SkipTrailingWhitespace(text, innerLow, innerHigh)

			if utils.ContainsLineTerminator(text, lastTokenEnd, innerHigh) {
				return
			}

			hasSpace := lastTokenEnd < innerHigh
			switch {
			case always && !hasSpace:
				ctx.ReportRangeWithFixes(
					core.NewTextRange(closePos, closePos+1),
					rule.RuleMessage{
						Id:          "missing",
						Description: "Requires a space before '}'.",
						Data:        map[string]string{"location": "before", "token": "}"},
					},
					rule.RuleFix{
						Text:  " ",
						Range: core.NewTextRange(lastTokenEnd, lastTokenEnd),
					},
				)
			case !always && hasSpace:
				ctx.ReportRangeWithFixes(
					core.NewTextRange(lastTokenEnd, innerHigh),
					rule.RuleMessage{
						Id:          "extra",
						Description: "Unexpected space(s) before '}'.",
						Data:        map[string]string{"location": "before", "token": "}"},
					},
					rule.RuleFix{
						Text:  "",
						Range: core.NewTextRange(lastTokenEnd, innerHigh),
					},
				)
			}
		}

		return rule.RuleListeners{
			ast.KindBlock:                          enter,
			ast.KindCaseBlock:                      enter,
			rule.ListenerOnExit(ast.KindBlock):     exit,
			rule.ListenerOnExit(ast.KindCaseBlock): exit,
		}
	},
}
