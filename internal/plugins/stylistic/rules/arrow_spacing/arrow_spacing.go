package arrow_spacing

import (
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	msgExpectedBefore   = "expectedBefore"
	msgUnexpectedBefore = "unexpectedBefore"
	msgExpectedAfter    = "expectedAfter"
	msgUnexpectedAfter  = "unexpectedAfter"

	descExpectedBefore   = "Missing space before =>."
	descUnexpectedBefore = "Unexpected space before =>."
	descExpectedAfter    = "Missing space after =>."
	descUnexpectedAfter  = "Unexpected space after =>."
)

type options struct {
	before bool
	after  bool
}

// parseOptions mirrors upstream's `defaultOptions: [{ before: true, after: true }]`.
// Either or both keys may be omitted; missing keys default to `true`.
func parseOptions(raw any) options {
	opts := options{before: true, after: true}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	if b, ok := m["before"].(bool); ok {
		opts.before = b
	}
	if b, ok := m["after"].(bool); ok {
		opts.after = b
	}
	return opts
}

func isWhitespaceByte(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\f', '\v':
		return true
	}
	return false
}

// isWhitespaceRune handles the non-ASCII subset of ECMAScript's `\s` —
// callers (`findBeforeToken` / `findAfterToken`) check ASCII bytes through
// `isWhitespaceByte` on the fast path, so this function only sees runes
// ≥ U+0080. The set is exactly:
//
//   - `WhiteSpace` (spec §12.2): U+00A0 NBSP, U+FEFF ZWNBSP, plus Unicode
//     `Space_Separator` (Zs) — U+1680, U+2000-U+200A, U+202F, U+205F, U+3000.
//   - `LineTerminator` (spec §12.3): U+2028 LS, U+2029 PS.
//
// Deliberately excluded — they are NOT matched by JS `\s` (verified
// empirically against V8 with `/\s/.test('')` / `'​'`):
//
//   - U+0085 Next Line: in Unicode `\p{White_Space}` but NOT in ES
//     `WhiteSpace` (only Zs and a specific allow-list count).
//   - U+200B Zero Width Space: category Cf (Format), not Zs; tsgo's scanner
//     treats it as whitespace internally, but that's a TypeScript-compiler
//     quirk, not the ESLint behavior we mirror.
func isWhitespaceRune(r rune) bool {
	switch r {
	case 0x00A0, // <No-Break Space>
		0x1680, // <Ogham Space Mark>
		0x2000, // <En Quad>
		0x2001, // <Em Quad>
		0x2002, // <En Space>
		0x2003, // <Em Space>
		0x2004, // <Three-Per-Em Space>
		0x2005, // <Four-Per-Em Space>
		0x2006, // <Six-Per-Em Space>
		0x2007, // <Figure Space>
		0x2008, // <Punctuation Space>
		0x2009, // <Thin Space>
		0x200A, // <Hair Space>
		0x2028, // <Line Separator>
		0x2029, // <Paragraph Separator>
		0x202F, // <Narrow No-Break Space>
		0x205F, // <Medium Mathematical Space>
		0x3000, // <Ideographic Space>
		0xFEFF: // <Byte Order Mark / Zero Width No-Break Space>
		return true
	}
	return false
}

// findBeforeToken walks left from `arrowStart` through whitespace, then
// locates the start of the previous token-or-comment. Returns the token range
// `[beforeStart, beforeEnd)` and `isSpaced`, which is true iff any whitespace
// was skipped between the previous token and `=>` (equivalent to ESLint's
// `sourceCode.isSpaceBetween(beforeToken, arrowToken)` with
// `includeComments: true`).
//
// Token kinds recognized:
//   - block comment ending in `*/` — walk back to the matching `/*`
//   - identifier / keyword / numeric literal — walk back across runes that
//     satisfy `scanner.IsIdentifierPart` (Unicode-aware, matches the tsgo
//     tokenizer's `[A-Za-z0-9_$\p{ID_Continue}…]` definition)
//   - single-byte punctuation (paren, brace, bracket, comma, semicolon, etc.)
//
// The byte before `=>` in valid syntax can only be `)`, an identifier-part
// rune, a digit (numeric literal), or `/` (block-comment closer). Strings
// and regex literals cannot appear in that lexical position, so the
// non-token-aware `*/` reverse scan can't be tricked by literal content
// (same invariant as array_bracket_spacing's prevRealEnd).
//
// Known limitation — line comment as beforeToken: for the (rare) shape
// `(a) // c\n=> b`, ESLint's `getTokenBefore(arrow, {includeComments: true})`
// returns the `//` comment and reports at column 5 (the `/`). We don't
// reverse-scan line comments (locating `//` from the right requires line-
// start + string-literal boundary detection, see array_bracket_spacing's
// prevRealEnd note), so the identifier-walk falls onto the last byte of the
// comment body and reports a few columns to the right. The `isSpaced`
// decision is unaffected — only the report column differs in this shape.
func findBeforeToken(text string, arrowStart int) (beforeStart, beforeEnd int, isSpaced bool) {
	p := arrowStart
	for p > 0 {
		// Fast path: ASCII whitespace is one byte. Slow path: decode the
		// rune to catch Unicode whitespace (NBSP, ideographic space,
		// line/paragraph separators, …) — same set ESLint's `\s` matches.
		if text[p-1] < 0x80 {
			if !isWhitespaceByte(text[p-1]) {
				break
			}
			p--
			continue
		}
		r, size := utf8.DecodeLastRuneInString(text[:p])
		if size == 0 || r == utf8.RuneError || !isWhitespaceRune(r) {
			break
		}
		p -= size
	}
	beforeEnd = p
	isSpaced = p < arrowStart

	if beforeEnd == 0 {
		return 0, beforeEnd, isSpaced
	}

	if beforeEnd >= 2 && text[beforeEnd-2] == '*' && text[beforeEnd-1] == '/' {
		q := beforeEnd - 2
		for q >= 1 {
			if text[q-1] == '/' && text[q] == '*' {
				return q - 1, beforeEnd, isSpaced
			}
			q--
		}
		return 0, beforeEnd, isSpaced
	}

	if r, size := utf8.DecodeLastRuneInString(text[:beforeEnd]); size > 0 && r != utf8.RuneError && scanner.IsIdentifierPart(r) {
		q := beforeEnd - size
		for q > 0 {
			r2, sz2 := utf8.DecodeLastRuneInString(text[:q])
			if sz2 == 0 || r2 == utf8.RuneError || !scanner.IsIdentifierPart(r2) {
				break
			}
			q -= sz2
		}
		return q, beforeEnd, isSpaced
	}

	return beforeEnd - 1, beforeEnd, isSpaced
}

// findAfterToken walks right from `arrowEnd` through whitespace, then
// locates the next token-or-comment. Mirrors findBeforeToken in the opposite
// direction; recognizes block / line comments and identifier-shaped tokens.
// Uses Unicode-aware `scanner.IsIdentifierPart` for identifier-token walking
// and the full ECMAScript `LineTerminator` set (LF, CR, LS, PS) for line
// comment termination.
func findAfterToken(text string, arrowEnd int) (afterStart, afterEnd int, isSpaced bool) {
	p := arrowEnd
	for p < len(text) {
		if text[p] < 0x80 {
			if !isWhitespaceByte(text[p]) {
				break
			}
			p++
			continue
		}
		r, size := utf8.DecodeRuneInString(text[p:])
		if size == 0 || r == utf8.RuneError || !isWhitespaceRune(r) {
			break
		}
		p += size
	}
	afterStart = p
	isSpaced = p > arrowEnd

	if afterStart >= len(text) {
		return afterStart, afterStart, isSpaced
	}

	first := text[afterStart]

	if afterStart+1 < len(text) && first == '/' && text[afterStart+1] == '*' {
		q := afterStart + 2
		for q+1 < len(text) && (text[q] != '*' || text[q+1] != '/') {
			q++
		}
		if q+1 < len(text) {
			return afterStart, q + 2, isSpaced
		}
		return afterStart, len(text), isSpaced
	}

	if afterStart+1 < len(text) && first == '/' && text[afterStart+1] == '/' {
		q := afterStart + 2
		for q < len(text) {
			c := text[q]
			if c < 0x80 {
				if c == '\n' || c == '\r' {
					break
				}
				q++
				continue
			}
			// Non-ASCII byte — decode rune to catch U+2028 LS and U+2029 PS,
			// which terminate a single-line comment per ECMAScript §12.3.
			r, size := utf8.DecodeRuneInString(text[q:])
			if r == 0x2028 || r == 0x2029 {
				break
			}
			if size == 0 || r == utf8.RuneError {
				q++ // safety: advance past undecodable bytes
				continue
			}
			q += size
		}
		return afterStart, q, isSpaced
	}

	if r, size := utf8.DecodeRuneInString(text[afterStart:]); size > 0 && r != utf8.RuneError && scanner.IsIdentifierPart(r) {
		q := afterStart + size
		for q < len(text) {
			r2, sz2 := utf8.DecodeRuneInString(text[q:])
			if sz2 == 0 || r2 == utf8.RuneError || !scanner.IsIdentifierPart(r2) {
				break
			}
			q += sz2
		}
		return afterStart, q, isSpaced
	}

	return afterStart, afterStart + 1, isSpaced
}

// findArrowTokenRange returns the byte range of the `=>` token for a single
// arrow / function-type / constructor-type node. For ArrowFunction we go
// through `EqualsGreaterThanToken.Pos()` — that field's `Pos()` includes
// leading trivia, so `GetRangeOfTokenAtPosition` is used to land on the
// actual `=>`. For type nodes tsgo does not expose a direct `=>` field; we
// scan forward from `Parameters.End()` until we hit the arrow token.
func findArrowTokenRange(sf *ast.SourceFile, node *ast.Node) (start, end int, ok bool) {
	switch node.Kind {
	case ast.KindArrowFunction:
		af := node.AsArrowFunction()
		if af == nil || af.EqualsGreaterThanToken == nil {
			return 0, 0, false
		}
		rng := scanner.GetRangeOfTokenAtPosition(sf, af.EqualsGreaterThanToken.Pos())
		if rng.End() <= rng.Pos() {
			return 0, 0, false
		}
		return rng.Pos(), rng.End(), true

	case ast.KindFunctionType:
		ft := node.AsFunctionTypeNode()
		if ft == nil || ft.Parameters == nil {
			return 0, 0, false
		}
		return scanForArrow(sf, ft.Parameters.End())

	case ast.KindConstructorType:
		ct := node.AsConstructorTypeNode()
		if ct == nil || ct.Parameters == nil {
			return 0, 0, false
		}
		return scanForArrow(sf, ct.Parameters.End())
	}
	return 0, 0, false
}

// scanForArrow steps the scanner forward from `fromPos`, returning the first
// `=>` token's range. Used for function-type / constructor-type nodes whose
// AST does not carry a direct `EqualsGreaterThanToken` field. The first arrow
// after the parameter list is always THIS node's arrow because nested
// function types live inside the return-type subtree, which sits past the
// arrow.
func scanForArrow(sf *ast.SourceFile, fromPos int) (start, end int, ok bool) {
	text := sf.Text()
	pos := fromPos
	for pos < len(text) {
		tok := scanner.ScanTokenAtPosition(sf, pos)
		rng := scanner.GetRangeOfTokenAtPosition(sf, pos)
		if tok == ast.KindEqualsGreaterThanToken {
			return rng.Pos(), rng.End(), true
		}
		if rng.End() <= pos {
			return 0, 0, false
		}
		pos = rng.End()
	}
	return 0, 0, false
}

// ArrowSpacingRule enforces consistent spacing before/after the `=>` arrow
// token. Ported from @stylistic/eslint-plugin's arrow-spacing.
//
// Each report carries a tiny autofix: an insertion or a whitespace-range
// removal. Both the before and after checks run on every visit, so a single
// arrow with both kinds of violation (e.g. `a=>a` with the default options)
// emits two independent diagnostics whose fixes compose.
var ArrowSpacingRule = rule.Rule{
	Name: "@stylistic/arrow-spacing",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		text := ctx.SourceFile.Text()

		check := func(node *ast.Node) {
			arrowStart, arrowEnd, ok := findArrowTokenRange(ctx.SourceFile, node)
			if !ok {
				return
			}

			beforeStart, beforeEnd, isSpacedBefore := findBeforeToken(text, arrowStart)
			if opts.before {
				if !isSpacedBefore {
					ctx.ReportRangeWithFixes(
						core.NewTextRange(beforeStart, beforeEnd),
						rule.RuleMessage{Id: msgExpectedBefore, Description: descExpectedBefore},
						rule.RuleFix{Text: " ", Range: core.NewTextRange(arrowStart, arrowStart)},
					)
				}
			} else if isSpacedBefore {
				ctx.ReportRangeWithFixes(
					core.NewTextRange(beforeStart, beforeEnd),
					rule.RuleMessage{Id: msgUnexpectedBefore, Description: descUnexpectedBefore},
					rule.RuleFix{Text: "", Range: core.NewTextRange(beforeEnd, arrowStart)},
				)
			}

			afterStart, afterEnd, isSpacedAfter := findAfterToken(text, arrowEnd)
			if opts.after {
				if !isSpacedAfter {
					ctx.ReportRangeWithFixes(
						core.NewTextRange(afterStart, afterEnd),
						rule.RuleMessage{Id: msgExpectedAfter, Description: descExpectedAfter},
						rule.RuleFix{Text: " ", Range: core.NewTextRange(arrowEnd, arrowEnd)},
					)
				}
			} else if isSpacedAfter {
				ctx.ReportRangeWithFixes(
					core.NewTextRange(afterStart, afterEnd),
					rule.RuleMessage{Id: msgUnexpectedAfter, Description: descUnexpectedAfter},
					rule.RuleFix{Text: "", Range: core.NewTextRange(arrowEnd, afterStart)},
				)
			}
		}

		// Register on exit so nested arrows report inner-first. For
		// `(a = ()=>0)=>1` the inner arrow lives in the outer arrow's
		// parameter initializer, so its `=>` sits earlier in source —
		// upstream's diagnostics emerge in source order via ESLint's final
		// sort step, and the on-exit visit replicates that here without an
		// explicit sort.
		return rule.RuleListeners{
			rule.ListenerOnExit(ast.KindArrowFunction):   check,
			rule.ListenerOnExit(ast.KindFunctionType):    check,
			rule.ListenerOnExit(ast.KindConstructorType): check,
		}
	},
}
