// Package no_useless_escape implements ESLint's `no-useless-escape` rule.
//
// It scans string literals, template literals, and regular-expression literals
// for backslash escapes whose escaped character carries no special meaning, and
// reports each such escape with two suggestions: remove the backslash, or
// double it (`\\X`) to embed an actual backslash.
package no_useless_escape

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-useless-escape
var NoUselessEscapeRule = rule.Rule{
	Name: "no-useless-escape",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		allowed := parseAllowRegexCharacters(options)

		return rule.RuleListeners{
			ast.KindStringLiteral: func(node *ast.Node) {
				if isInJsxLiteralContext(node) {
					return
				}
				checkStringLiteral(ctx, node)
			},
			ast.KindNoSubstitutionTemplateLiteral: func(node *ast.Node) {
				if isInsideTaggedTemplate(node) {
					return
				}
				checkTemplateElement(ctx, node)
			},
			ast.KindTemplateHead: func(node *ast.Node) {
				if isInsideTaggedTemplate(node) {
					return
				}
				checkTemplateElement(ctx, node)
			},
			ast.KindTemplateMiddle: func(node *ast.Node) {
				if isInsideTaggedTemplate(node) {
					return
				}
				checkTemplateElement(ctx, node)
			},
			ast.KindTemplateTail: func(node *ast.Node) {
				if isInsideTaggedTemplate(node) {
					return
				}
				checkTemplateElement(ctx, node)
			},
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				validateRegExp(ctx, node, allowed)
			},
		}
	},
}

func parseAllowRegexCharacters(options any) map[string]bool {
	out := map[string]bool{}
	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		return out
	}
	raw, ok := optsMap["allowRegexCharacters"]
	if !ok {
		return out
	}
	arr, ok := raw.([]interface{})
	if !ok {
		return out
	}
	for _, v := range arr {
		if s, ok := v.(string); ok {
			out[s] = true
		}
	}
	return out
}

// validStringEscape mirrors ESLint's VALID_STRING_ESCAPES — the escape
// characters that produce a different value from their literal form, so the
// backslash is meaningful. Single-character entries are checked via this map;
// linebreak runes are checked separately.
var validStringEscape = map[byte]bool{
	'\\': true, 'n': true, 'r': true, 'v': true, 't': true,
	'b': true, 'f': true, 'u': true, 'x': true,
}

// isLineContinuation reports whether the rune begins a LineTerminatorSequence
// that, when preceded by `\` in a string/template, forms a line continuation.
// ESLint's LINEBREAKS includes \n, \r, U+2028 (LINE SEPARATOR), U+2029
// (PARAGRAPH SEPARATOR).
func isLineContinuation(r rune) bool {
	return r == '\n' || r == '\r' || r == 0x2028 || r == 0x2029
}

// isInJsxLiteralContext reports whether `node` is a string literal that JSX
// passes through verbatim — JSX attribute values, plus the rare cases where
// a string literal appears as a direct child of a JsxElement / JsxFragment
// (legacy JSX-text-in-element shapes). JSX itself does not interpret escape
// sequences (https://facebook.github.io/jsx/), so the `\` is part of the
// authored text.
func isInJsxLiteralContext(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindJsxAttribute, ast.KindJsxElement, ast.KindJsxFragment:
		return true
	}
	return false
}

// isInsideTaggedTemplate reports whether the template element belongs to a
// TaggedTemplateExpression. Tagged tags (e.g. `String.raw`, `myFn`) receive the
// raw escape text via the `raw` array, so removing the backslash would change
// the value the tag function sees.
func isInsideTaggedTemplate(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindNoSubstitutionTemplateLiteral:
		// Parent is the TaggedTemplateExpression directly when the template is
		// the tagged form `tag`literal``.
		return node.Parent != nil && node.Parent.Kind == ast.KindTaggedTemplateExpression
	case ast.KindTemplateHead:
		// Parent is TemplateExpression; grandparent is the TaggedTemplate.
		return node.Parent != nil &&
			node.Parent.Parent != nil &&
			node.Parent.Parent.Kind == ast.KindTaggedTemplateExpression
	case ast.KindTemplateMiddle, ast.KindTemplateTail:
		// Parent is TemplateSpan; grandparent is TemplateExpression;
		// great-grandparent is the TaggedTemplate.
		return node.Parent != nil &&
			node.Parent.Parent != nil &&
			node.Parent.Parent.Parent != nil &&
			node.Parent.Parent.Parent.Kind == ast.KindTaggedTemplateExpression
	}
	return false
}

// isStringLiteralDirective reports whether `node` is a string literal that
// occupies a directive prologue position. ESLint's parser marks directives via
// `node.directive` only at: the SourceFile / ModuleBlock head, or the head of
// a function body (FunctionDeclaration, FunctionExpression, ArrowFunction,
// MethodDeclaration, GetAccessor, SetAccessor, Constructor, ClassStaticBlock).
// A plain `Block` inside `if`/`for`/`while`/etc. is NOT a directive container.
//
// rslint mirrors that: the parent must be an ExpressionStatement, the
// grandparent must be one of the directive containers, AND in the Block case
// the great-grandparent must be a function-like / class-static-block.
func isStringLiteralDirective(sourceFile *ast.SourceFile, node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindExpressionStatement {
		return false
	}
	grandparent := parent.Parent
	if grandparent == nil {
		return false
	}
	var statements []*ast.Node
	switch grandparent.Kind {
	case ast.KindSourceFile:
		statements = sourceFile.Statements.Nodes
	case ast.KindModuleBlock:
		statements = grandparent.AsModuleBlock().Statements.Nodes
	case ast.KindBlock:
		// Only function-body / class-static-block / constructor-body Blocks are
		// directive containers. Bodies of `if` / `for` / `while` / standalone
		// blocks etc. are not.
		container := grandparent.Parent
		if container == nil || !isDirectiveBlockContainer(container) {
			return false
		}
		statements = grandparent.AsBlock().Statements.Nodes
	default:
		return false
	}
	for _, stmt := range statements {
		if stmt == parent {
			return true
		}
		if !ast.IsPrologueDirective(stmt) {
			return false
		}
	}
	return false
}

// isDirectiveBlockContainer reports whether a Block whose parent is `node` is
// a function-body Block (i.e. its first statements are part of a directive
// prologue). Mirrors the set of node kinds that carry a `body: Block` and
// honor "use strict" / other directives in ECMAScript.
func isDirectiveBlockContainer(node *ast.Node) bool {
	if ast.IsFunctionLike(node) {
		return true
	}
	switch node.Kind {
	case ast.KindClassStaticBlockDeclaration:
		return true
	}
	return false
}

// rawStartOf returns the source-text start offset of `node`, skipping any
// leading trivia. For string/template/regex literals this points at the
// opening delimiter (`'`, `"`, `` ` ``, `}`, or `/`).
func rawStartOf(sourceFile *ast.SourceFile, node *ast.Node) int {
	return utils.TrimNodeTextRange(sourceFile, node).Pos()
}

// rawTextOf returns the raw source text of `node` (delimiters included).
func rawTextOf(sourceFile *ast.SourceFile, node *ast.Node) string {
	return scanner.GetSourceTextOfNodeFromSourceFile(sourceFile, node, false)
}

// checkStringLiteral scans a `'…'` / `"…"` literal for useless escapes.
func checkStringLiteral(ctx rule.RuleContext, node *ast.Node) {
	rawStart := rawStartOf(ctx.SourceFile, node)
	raw := rawTextOf(ctx.SourceFile, node)
	if len(raw) < 2 {
		return
	}
	quote := raw[0]
	directive := isStringLiteralDirective(ctx.SourceFile, node)
	scanLiteralEscapes(raw, func(escapeIdx int, escapedRune rune, escapedRuneSize int) {
		if isValidStringEscapeRune(escapedRune) {
			return
		}
		if escapedRuneSize == 1 && byte(escapedRune) == quote {
			return
		}
		reportEscape(ctx, rawStart, escapeIdx, escapedDisplayText(raw, escapeIdx, escapedRune, escapedRuneSize), directive, false)
	})
}

// checkTemplateElement scans a template element (NoSubstitutionTemplateLiteral
// or TemplateHead/Middle/Tail) for useless escapes. Template-specific cases:
//   - `` \` `` is the quote escape and never reported.
//   - `\$` is necessary iff followed by `{` (forms `\${` to suppress
//     interpolation).
//   - `\{` is necessary iff preceded by `$` (forms `$\{` to suppress
//     interpolation; the leading `\$` would have been reported instead in the
//     `\${` case).
//
// Template literals are never directives, so the suggestion message is always
// the standard `removeEscape`.
func checkTemplateElement(ctx rule.RuleContext, node *ast.Node) {
	rawStart := rawStartOf(ctx.SourceFile, node)
	raw := rawTextOf(ctx.SourceFile, node)
	scanLiteralEscapes(raw, func(escapeIdx int, escapedRune rune, escapedRuneSize int) {
		// `\`` is the template quote escape — always valid.
		if escapedRune == '`' {
			return
		}
		unnecessary := !isValidStringEscapeRune(escapedRune)

		switch escapedRune {
		case '$':
			// `\$` is necessary only when followed by `{`.
			afterIdx := escapeIdx + 1 + escapedRuneSize
			unnecessary = afterIdx >= len(raw) || raw[afterIdx] != '{'
		case '{':
			// `\{` is necessary only when preceded by `$`. If preceded by `\$`,
			// the rule reports the `\$` instead (it scans left-to-right and
			// the `\$` is the first useless escape).
			unnecessary = escapeIdx == 0 || raw[escapeIdx-1] != '$'
		}
		if !unnecessary {
			return
		}
		reportEscape(ctx, rawStart, escapeIdx, escapedDisplayText(raw, escapeIdx, escapedRune, escapedRuneSize), false, false)
	})
}

// escapedDisplayText returns the source-text bytes for the escaped character,
// used as the `{character}` placeholder in the diagnostic message. ESLint
// computes this as `match[0].slice(1)` over the JS source string — which for
// astral X gives both surrogate code units (i.e. the full character). Slicing
// the UTF-8 source bytes by `escapedRuneSize` produces the equivalent display
// string: same rendered character, different byte-encoding. The message text
// rendered by terminals / IDEs is identical.
func escapedDisplayText(raw string, escapeIdx int, escapedRune rune, escapedRuneSize int) string {
	return raw[escapeIdx+1 : escapeIdx+1+escapedRuneSize]
}

func isValidStringEscapeRune(r rune) bool {
	if r < 128 && validStringEscape[byte(r)] {
		return true
	}
	return isLineContinuation(r)
}

// scanLiteralEscapes mirrors ESLint's `/\\\D/gu`: it walks `raw` and invokes
// `cb` for every backslash that is followed by a non-digit character. The
// callback receives the index of the backslash, the escaped rune, and the
// rune's UTF-8 byte width. A backslash at the end of input or directly before
// an ASCII digit is consumed without reporting (digit-prefixed escapes are
// hex/unicode/octal continuations, not single-character escapes).
func scanLiteralEscapes(raw string, cb func(escapeIdx int, escapedRune rune, escapedRuneSize int)) {
	i := 0
	for i < len(raw) {
		if raw[i] != '\\' {
			i++
			continue
		}
		if i+1 >= len(raw) {
			return
		}
		next := raw[i+1]
		if next >= '0' && next <= '9' {
			i += 2
			continue
		}
		r, size := utf8.DecodeRuneInString(raw[i+1:])
		if size == 0 {
			return
		}
		cb(i, r, size)
		// Step past the backslash and the escaped rune. ESLint's regex matches
		// `/\\\D/g` so its loop also consumes the escaped char before
		// continuing.
		i += 1 + size
	}
}

// reportEscape emits the diagnostic and the two standard suggestions.
//
// `startOffset` is the byte index of the backslash within the raw text;
// `escapedText` is the source-text representation of the escaped character
// (without the leading backslash). When `directive` is true the remove-escape
// suggestion uses the `removeEscapeDoNotKeepSemantics` message — removing the
// `\` from `"use\ strict"` does not preserve the directive's behavior. When
// `disableEscapeBackslash` is true the second (`\X` → `\\X`) suggestion is
// suppressed; this happens for v-mode escapes whose direct parent in the
// regexpp AST is a `ClassIntersection` / `ClassSubtraction` operator.
func reportEscape(ctx rule.RuleContext, rawStart, startOffset int, escapedText string, directive bool, disableEscapeBackslash bool) {
	rangeStart := rawStart + startOffset
	rangeEnd := rangeStart + 1
	textRange := core.NewTextRange(rangeStart, rangeEnd)

	suggestions := make([]rule.RuleSuggestion, 0, 2)
	removeMsg := rule.RuleMessage{
		Id:          "removeEscape",
		Description: "Remove the `\\`. This maintains the current functionality.",
	}
	if directive {
		removeMsg = rule.RuleMessage{
			Id:          "removeEscapeDoNotKeepSemantics",
			Description: "Remove the `\\` if it was inserted by mistake.",
		}
	}
	suggestions = append(suggestions, rule.RuleSuggestion{
		Message:  removeMsg,
		FixesArr: []rule.RuleFix{rule.RuleFixRemoveRange(textRange)},
	})
	if !disableEscapeBackslash {
		suggestions = append(suggestions, rule.RuleSuggestion{
			Message: rule.RuleMessage{
				Id:          "escapeBackslash",
				Description: "Replace the `\\` with `\\\\` to include the actual backslash character.",
			},
			FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(
				core.NewTextRange(rangeStart, rangeStart),
				"\\",
			)},
		})
	}
	ctx.ReportRangeWithSuggestions(textRange, rule.RuleMessage{
		Id:          "unnecessaryEscape",
		Description: fmt.Sprintf("Unnecessary escape character: \\%s.", escapedText),
	}, suggestions...)
}

// ----------------------------------------------------------------------------
// Regex pattern scanner
// ----------------------------------------------------------------------------

// regexClassFrame tracks a single open `[...]` (or `[^...]`, or v-mode nested
// class) while we walk the pattern. A frame holds enough context to answer
// each of the rule's class-sensitive questions:
//
//   - is the current escape positioned at the very start of this class
//     (relevant to `\^` and `\-`)?
//   - is it at the end (relevant to `\-` outside v-mode)?
//   - is the immediately enclosing class a `ClassIntersection` /
//     `ClassSubtraction` (v-mode `--` / `&&`), which suppresses the
//     `escapeBackslash` suggestion?
type regexClassFrame struct {
	start    int  // byte offset of `[`
	end      int  // byte offset of `]` (exclusive of the bracket itself; end = index after `]`)
	negate   bool // `[^…` form
	hasSetOp bool // contains `--` or `&&` at THIS class's level (v-mode only)
}

// reservedDoublePunctuator mirrors REGEX_CLASS_SET_RESERVED_DOUBLE_PUNCTUATOR
// in the ESLint source: characters that, when doubled inside a v-mode class,
// form a reserved-syntax pair (e.g. `&&`, `++`).
var reservedDoublePunctuator = map[byte]bool{
	'!': true, '#': true, '$': true, '%': true, '&': true,
	'*': true, '+': true, ',': true, '.': true, ':': true,
	';': true, '<': true, '=': true, '>': true, '?': true,
	'@': true, '^': true, '`': true, '~': true,
}

// regexGeneralEscape mirrors REGEX_GENERAL_ESCAPES: characters whose `\X` form
// is meaningful in any regex position (assertions, character-class shorthands,
// hex/unicode escape leaders, octal/backreference digits, the `]` escape).
var regexGeneralEscape = map[byte]bool{
	'\\': true, 'b': true, 'c': true, 'd': true, 'D': true, 'f': true,
	'n': true, 'p': true, 'P': true, 'r': true, 's': true, 'S': true,
	't': true, 'v': true, 'w': true, 'W': true, 'x': true, 'u': true,
	'0': true, '1': true, '2': true, '3': true, '4': true, '5': true,
	'6': true, '7': true, '8': true, '9': true, ']': true,
}

// regexNonClassExtra adds the characters whose `\X` is meaningful outside any
// character class (regex metacharacters), on top of regexGeneralEscape.
var regexNonClassExtra = map[byte]bool{
	'^': true, '/': true, '.': true, '$': true, '*': true, '+': true,
	'?': true, '[': true, '{': true, '}': true, '|': true, '(': true,
	')': true, 'B': true, 'k': true,
}

// regexClassSetExtra adds the characters whose `\X` is meaningful inside a
// v-mode character class, on top of regexGeneralEscape.
var regexClassSetExtra = map[byte]bool{
	'q': true, '/': true, '[': true, '{': true, '}': true,
	'|': true, '(': true, ')': true, '-': true,
}

// regexAllowedNonClass / regexAllowedClassU / regexAllowedClassV are the three
// resolved allow-sets (per ESLint's three context branches), pre-merged at
// init so the hot path doesn't rebuild a map per escape.
var (
	regexAllowedNonClass = mergeByteSets(regexGeneralEscape, regexNonClassExtra)
	regexAllowedClassU   = regexGeneralEscape
	regexAllowedClassV   = mergeByteSets(regexGeneralEscape, regexClassSetExtra)
)

// validateRegExp inspects a /…/flags literal for useless escape characters.
// It mirrors ESLint's regexpp-driven walker but operates on the raw pattern
// text rather than a parsed AST: every check the rule cares about can be
// answered from the byte-level position alone (immediate enclosing class,
// negate flag, presence of `--`/`&&` in that class). The pre-scan in
// preScanClass populates the frame metadata before we reach inner content.
func validateRegExp(ctx rule.RuleContext, node *ast.Node, allowed map[string]bool) {
	rawStart := rawStartOf(ctx.SourceFile, node)
	text := rawTextOf(ctx.SourceFile, node)
	pattern, flagsStr := utils.ExtractRegexPatternAndFlags(text)
	if pattern == "" {
		return
	}
	flags := utils.ParseRegexFlags(flagsStr)
	patternStart := rawStart + 1 // past the opening `/`

	// Match ESLint's "wrap regexpp.parsePattern in try/catch and skip on any
	// error" semantics: if the pattern doesn't parse cleanly, we don't report
	// — `no-invalid-regexp` will surface the syntax issue instead.
	if !patternParses(pattern, flags) {
		return
	}

	var stack []regexClassFrame
	i := 0
	for i < len(pattern) {
		c := pattern[i]
		switch {
		case c == '\\':
			advance, escapedByte, escapedSize, isSingleIdentity := readRegexEscape(pattern, i, flags, len(stack) > 0)
			if isSingleIdentity {
				handleRegexIdentityEscape(ctx, patternStart, pattern, i, escapedByte, escapedSize, stack, flags, allowed)
			}
			i += advance
		case c == '[' && (len(stack) == 0 || flags.UnicodeSets):
			// Open a new class frame at depth 0 always; nested `[` only nests
			// under v-mode. In non-v mode an inner `[` is a literal character.
			frame := preScanClass(pattern, i, flags)
			stack = append(stack, frame)
			i++
			if frame.negate {
				i++ // step past the leading `^`
			}
		case c == ']' && len(stack) > 0:
			stack = stack[:len(stack)-1]
			i++
		case len(stack) > 0 && flags.UnicodeSets && (c == '-' || c == '&') &&
			i+1 < len(pattern) && pattern[i+1] == c:
			// v-mode set operator `--` / `&&` — consumed as one unit.
			i += 2
		default:
			_, w := utf8.DecodeRuneInString(pattern[i:])
			if w == 0 {
				i++
			} else {
				i += w
			}
		}
	}
}

// readRegexEscape classifies the `\…` token at pattern[i].
//
// Returns:
//   - advance: how many bytes the token consumes (always ≥ 2).
//   - escapedByte: the byte right after `\`. Only meaningful for single-byte
//     identity escapes.
//   - escapedSize: byte width of the escaped rune (1 for ASCII, ≥ 2 otherwise).
//   - isSingleIdentity: true iff the token is a `\X` form whose value equals
//     `X` itself (i.e. the backslash carries no meaning beyond escaping `X`).
//     Multi-byte structural escapes (`\xHH`, `\uHHHH`, `\u{H…}`, `\p{…}`,
//     `\q{…}`, `\k<…>`, decimal continuations) return false here so the rule
//     never inspects them.
func readRegexEscape(pattern string, i int, flags utils.RegexFlags, inClass bool) (advance int, escapedByte byte, escapedSize int, isSingleIdentity bool) {
	if i+1 >= len(pattern) {
		return 1, 0, 0, false
	}
	next := pattern[i+1]
	switch next {
	case 'x':
		// `\xHH` — 4 bytes. Otherwise treat as identity escape of `x`.
		if i+3 < len(pattern) && utils.IsHexDigit(pattern[i+2]) && utils.IsHexDigit(pattern[i+3]) {
			return 4, 0, 0, false
		}
	case 'u':
		// `\u{H…}` (uvMode), `\uHHHH`, otherwise identity escape of `u`.
		if flags.UV() && i+2 < len(pattern) && pattern[i+2] == '{' {
			if rel := strings.IndexByte(pattern[i+3:], '}'); rel >= 0 {
				return 3 + rel + 1, 0, 0, false
			}
		}
		if i+5 < len(pattern) && utils.AllHexDigits(pattern[i+2:i+6]) {
			return 6, 0, 0, false
		}
	case 'p', 'P':
		// `\p{…}` / `\P{…}` are unicode property escapes only under uvMode.
		// In non-uvMode this is identity escape of `p` / `P`.
		if flags.UV() && i+2 < len(pattern) && pattern[i+2] == '{' {
			if rel := strings.IndexByte(pattern[i+3:], '}'); rel >= 0 {
				return 3 + rel + 1, 0, 0, false
			}
		}
	// `\q{…}` is intentionally NOT skipped as a unit. ESLint's regexpp parser
	// visits Character nodes inside the q-string body (each alternative is a
	// String containing real Character nodes), so identity escapes like `\.`
	// inside `\q{a\.b}` MUST be flagged. We treat `\q` as an identity escape
	// (allowed in v-mode class via regexClassSetExtra) and let the main walker
	// scan the body bytes — `{`, `|`, alternative chars, and the closing `}`
	// are all plain in our model. Unterminated `\q{` is rejected separately
	// in patternParses so we never reach inconsistent state on broken input.
	case 'k':
		// `\k<name>` — named backreference outside class.
		if !inClass && i+2 < len(pattern) && pattern[i+2] == '<' {
			if rel := strings.IndexByte(pattern[i+3:], '>'); rel >= 0 {
				return 3 + rel + 1, 0, 0, false
			}
		}
	case 'c':
		// `\cX` control escape. Only well-formed when followed by an ASCII
		// letter; otherwise regexpp falls back to identity-escape `c` (under
		// non-u). We treat any `\c<x>` as a structural escape — `c` itself is
		// in regexGeneralEscape so non-letter follow-ups still won't flag.
		if i+2 < len(pattern) {
			return 3, 0, 0, false
		}
	}
	if next >= '0' && next <= '9' {
		// Backreference (outside class) or legacy octal — eat all consecutive
		// digits as one structural escape.
		j := i + 2
		for j < len(pattern) && pattern[j] >= '0' && pattern[j] <= '9' {
			j++
		}
		return j - i, 0, 0, false
	}

	// Single-character identity escape `\X`. Decode the rune to handle
	// multi-byte X (very rare but possible in identity-escape position).
	r, w := utf8.DecodeRuneInString(pattern[i+1:])
	if w == 0 {
		return 1, 0, 0, false
	}
	if w == 1 {
		return 2, byte(r), 1, true
	}
	// Multi-byte identity escape. ESLint's set-membership checks operate on
	// single ASCII chars; a multi-byte X is never "valid" in any of those
	// sets, so treating it as a flagged identity is the safe path. Callers
	// pass escapedSize so we can copy the right number of bytes for the
	// diagnostic text.
	return 1 + w, 0, w, true
}

// handleRegexIdentityEscape classifies a `\X` (single-codepoint identity
// escape) and reports it when ESLint would. It encodes:
//
//   - Set membership: `X` against REGEX_GENERAL_ESCAPES /
//     REGEX_NON_CHARCLASS_ESCAPES / REGEX_CLASSSET_CHARACTER_ESCAPES based on
//     class state and uvMode.
//   - Caret exemption: `\^` at `[`-position is meaningful even under v-mode.
//   - Dash exemption (non-v): `\-` away from class edges is meaningful.
//   - V-mode reserved double punctuators: `\X` is meaningful when sandwiched
//     against a literal `X` neighbour (`\&&`, `&\&`, …) — with a small twist
//     for `\^` adjacent to the class's negate caret.
//   - V-mode set-operation parents: when the class containing the escape has
//     `--` / `&&` at its top level (frame.hasSetOp), the second
//     `escapeBackslash` suggestion is suppressed.
func handleRegexIdentityEscape(
	ctx rule.RuleContext,
	patternStart int,
	pattern string,
	escapeIdx int,
	escapedByte byte,
	escapedSize int,
	stack []regexClassFrame,
	flags utils.RegexFlags,
	allowed map[string]bool,
) {
	if escapedSize == 0 {
		return
	}
	escapedText := pattern[escapeIdx+1 : escapeIdx+1+escapedSize]
	if allowed[escapedText] {
		return
	}
	// Multi-byte identity escapes never match the ASCII-keyed allow sets, so
	// we can fast-path them straight to a flag.
	if escapedSize > 1 {
		reportEscape(ctx, patternStart, escapeIdx, escapedText, false, false)
		return
	}

	inClass := len(stack) > 0
	var allowedSet map[byte]bool
	switch {
	case !inClass:
		allowedSet = regexAllowedNonClass
	case flags.UnicodeSets:
		allowedSet = regexAllowedClassV
	default:
		allowedSet = regexAllowedClassU
	}
	if allowedSet[escapedByte] {
		return
	}

	disableBackslash := false
	if inClass {
		frame := stack[len(stack)-1]
		// `\^` exemption: at the very first character position of the class.
		if escapedByte == '^' {
			if frame.start+1 == escapeIdx {
				return
			}
		}
		if !flags.UnicodeSets {
			if escapedByte == '-' {
				escEnd := escapeIdx + 2
				// Outside v-mode, `\-` is meaningful in the middle of a class
				// — only flag it when adjacent to either edge.
				if frame.start+1 != escapeIdx && escEnd != frame.end-1 {
					return
				}
			}
		} else {
			if reservedDoublePunctuator[escapedByte] {
				escEnd := escapeIdx + 2
				if escEnd < len(pattern) && pattern[escEnd] == escapedByte {
					return
				}
				if escapeIdx > 0 && pattern[escapeIdx-1] == escapedByte {
					if escapedByte != '^' {
						return
					}
					if !frame.negate {
						return
					}
					negateCaretIdx := frame.start + 1
					if negateCaretIdx < escapeIdx-1 {
						return
					}
					// Otherwise the prior char IS the negate caret, so the
					// escape still looks redundant — fall through to flag.
				}
			}
			if frame.hasSetOp {
				disableBackslash = true
			}
		}
	}

	reportEscape(ctx, patternStart, escapeIdx, escapedText, false, disableBackslash)
}

// preScanClass walks a `[...]` opening at pattern[start] to determine its end
// position, whether it's negated (`[^…`), and whether it contains a v-mode
// set operator (`--` / `&&`) at its OWN nesting level (not inside a deeper
// nested class or `\q{…}` body).
//
// `\q{…}` content is skipped here as a unit because we need to find the class
// end, and a `]` literal inside `\q{a]b}` must NOT close the outer class.
// (Outside this scanner, the main walker re-walks `\q{…}` body to flag
// nested escapes — see readRegexEscape for the rationale.)
func preScanClass(pattern string, start int, flags utils.RegexFlags) regexClassFrame {
	frame := regexClassFrame{start: start, end: len(pattern)}
	i := start + 1
	if i < len(pattern) && pattern[i] == '^' {
		frame.negate = true
		i++
	}
	depth := 1
	for i < len(pattern) {
		c := pattern[i]
		switch {
		case c == '\\':
			// Special case: `\q{…}` body can contain a literal `]`. Skip the
			// whole `\q{…}` here so we don't mis-close the class.
			if flags.UnicodeSets && i+2 < len(pattern) && pattern[i+1] == 'q' && pattern[i+2] == '{' {
				if rel := strings.IndexByte(pattern[i+3:], '}'); rel >= 0 {
					i += 3 + rel + 1
					continue
				}
			}
			adv, _, _, _ := readRegexEscape(pattern, i, flags, true)
			i += adv
		case c == '[' && flags.UnicodeSets:
			depth++
			i++
		case c == ']':
			depth--
			i++
			if depth == 0 {
				frame.end = i
				return frame
			}
		case flags.UnicodeSets && depth == 1 && (c == '-' || c == '&') &&
			i+1 < len(pattern) && pattern[i+1] == c:
			frame.hasSetOp = true
			i += 2
		default:
			_, w := utf8.DecodeRuneInString(pattern[i:])
			if w == 0 {
				i++
			} else {
				i += w
			}
		}
	}
	return frame
}

// patternParses returns false for patterns that should be skipped by the rule
// — matching ESLint's wrap-regexpp-in-try/catch behavior. We don't fully
// validate regex syntax (that's `no-invalid-regexp`'s job), but we do reject
// the breakage modes that would let our walker drift away from regexpp's
// semantics:
//
//   - Unterminated character class (no closing `]`).
//   - Trailing `\` with no escaped char.
//   - v-mode `\u{H…}` / `\p{…}` / `\q{…}` / non-class `\k<…>` with no closing
//     `}` or `>`.
//
// Anything else (legal-but-weird patterns) is left to the walker.
func patternParses(pattern string, flags utils.RegexFlags) bool {
	depth := 0
	i := 0
	for i < len(pattern) {
		c := pattern[i]
		switch c {
		case '\\':
			if i+1 >= len(pattern) {
				return false
			}
			next := pattern[i+1]
			// Detect unterminated structural escapes BEFORE the readRegexEscape
			// fast path swallows them as identity escapes.
			switch {
			case flags.UV() && next == 'u' && i+2 < len(pattern) && pattern[i+2] == '{':
				if strings.IndexByte(pattern[i+3:], '}') < 0 {
					return false
				}
			case flags.UV() && (next == 'p' || next == 'P') && i+2 < len(pattern) && pattern[i+2] == '{':
				if strings.IndexByte(pattern[i+3:], '}') < 0 {
					return false
				}
			case flags.UnicodeSets && depth > 0 && next == 'q' && i+2 < len(pattern) && pattern[i+2] == '{':
				rel := strings.IndexByte(pattern[i+3:], '}')
				if rel < 0 {
					return false
				}
				// Skip `\q{…}` as a unit so a `]` inside doesn't unbalance the
				// class depth count.
				i += 3 + rel + 1
				continue
			case depth == 0 && next == 'k' && i+2 < len(pattern) && pattern[i+2] == '<':
				if strings.IndexByte(pattern[i+3:], '>') < 0 {
					return false
				}
			}
			adv, _, _, _ := readRegexEscape(pattern, i, flags, depth > 0)
			i += adv
		case '[':
			// Only nest depth in v-mode (where `[[…]…]` is a class set
			// expression). In non-v mode `[` inside a class is a literal char,
			// not a nested class — incrementing here would consume the wrong
			// `]` as the class closer.
			if depth == 0 || flags.UnicodeSets {
				depth++
			}
			i++
		case ']':
			if depth == 0 {
				// `]` outside class — legal in regex (literal `]`).
				i++
				continue
			}
			depth--
			i++
		default:
			_, w := utf8.DecodeRuneInString(pattern[i:])
			if w == 0 {
				i++
			} else {
				i += w
			}
		}
	}
	return depth == 0
}

func mergeByteSets(sets ...map[byte]bool) map[byte]bool {
	out := make(map[byte]bool, 64)
	for _, s := range sets {
		for k, v := range s {
			if v {
				out[k] = true
			}
		}
	}
	return out
}
