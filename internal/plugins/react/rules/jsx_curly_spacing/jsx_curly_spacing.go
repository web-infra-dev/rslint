package jsx_curly_spacing

import (
	"regexp"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	spacingNever  = "never"
	spacingAlways = "always"
)

type sideConfig struct {
	when                string
	allowMultiline      bool
	objectLiteralSpaces string
}

type defaults struct {
	when                string
	allowMultiline      bool
	objectLiteralSpaces string
}

func boolValue(v interface{}, fallback bool) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return fallback
}

func stringValue(v interface{}, fallback string) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fallback
}

func nestedMap(m map[string]interface{}, key string) map[string]interface{} {
	if m == nil {
		return nil
	}
	if v, ok := m[key]; ok {
		if nm, ok := v.(map[string]interface{}); ok {
			return nm
		}
	}
	return nil
}

// normalizeConfig mirrors upstream's normalizeConfig: pulls when /
// allowMultiline / spacing.objectLiterals from a config object (or treats
// `true` as "use defaults"), then on lastPass falls back objectLiteralSpaces
// to `when` when neither config nor defaults specify it.
func normalizeConfig(configOrTrue interface{}, d defaults, lastPass bool) sideConfig {
	cfg, _ := configOrTrue.(map[string]interface{})
	when := d.when
	if cfg != nil {
		when = stringValue(cfg["when"], when)
	}
	allowMultiline := d.allowMultiline
	if cfg != nil {
		if v, ok := cfg["allowMultiline"]; ok {
			allowMultiline = boolValue(v, allowMultiline)
		}
	}
	objectLiteralSpaces := d.objectLiteralSpaces
	spacing := nestedMap(cfg, "spacing")
	if spacing != nil {
		if v, ok := spacing["objectLiterals"]; ok {
			objectLiteralSpaces = stringValue(v, objectLiteralSpaces)
		}
	}
	if lastPass && objectLiteralSpaces == "" {
		objectLiteralSpaces = when
	}
	return sideConfig{
		when:                when,
		allowMultiline:      allowMultiline,
		objectLiteralSpaces: objectLiteralSpaces,
	}
}

// parseOptions reproduces upstream's option-shape juggling. Inputs that flow
// in here:
//   - nil / [] → all defaults
//   - "never" / "always" → equivalent to [{ when: "<value>" }]
//   - ["never", { allowMultiline?, spacing? }] → merge secondary into the
//     wrapped { when, ... } object
//   - [ { ... } ] / { ... } → use as-is
//
// `attributes` defaults to true (i.e. always check attributes); `children`
// defaults to false (i.e. don't check children unless explicitly enabled).
func parseOptions(options any) (attrs *sideConfig, children *sideConfig) {
	const (
		defaultWhen           = spacingNever
		defaultAllowMultiline = true
		defaultAttributes     = true
		defaultChildren       = false
	)

	var arr []interface{}
	switch v := options.(type) {
	case []interface{}:
		arr = v
	case map[string]interface{}:
		arr = []interface{}{v}
	case string:
		arr = []interface{}{v}
	}

	originalConfig := map[string]interface{}{}
	if len(arr) > 0 {
		switch first := arr[0].(type) {
		case string:
			if first == spacingAlways || first == spacingNever {
				originalConfig["when"] = first
				if len(arr) > 1 {
					if secondary, ok := arr[1].(map[string]interface{}); ok {
						for k, val := range secondary {
							originalConfig[k] = val
						}
					}
				}
			}
		case map[string]interface{}:
			originalConfig = first
		}
	}

	defaultCfg := normalizeConfig(originalConfig, defaults{
		when:           defaultWhen,
		allowMultiline: defaultAllowMultiline,
	}, false)

	attrs = resolveSideConfig(originalConfig, "attributes", defaultAttributes, defaultCfg)
	children = resolveSideConfig(originalConfig, "children", defaultChildren, defaultCfg)
	return attrs, children
}

// resolveSideConfig mirrors upstream's
//
//	const x = has(orig, 'x') ? orig.x : DEFAULT_X;
//	const xConfig = x ? normalizeConfig(x, defaults, true) : null;
//
// — when the side key is absent we fall back to the rule's default; when
// present we apply JS-truthy semantics (`false`/`0`/`""`/`null` → disabled,
// anything else → enabled). Boolean-true is normalised to "use defaults"
// by passing nil into normalizeConfig.
func resolveSideConfig(originalConfig map[string]interface{}, key string, defaultEnabled bool, defaultCfg sideConfig) *sideConfig {
	raw, present := originalConfig[key]
	enabled := defaultEnabled
	if present {
		enabled = jsTruthy(raw)
	}
	if !enabled {
		return nil
	}
	var cfg interface{}
	if present && !isBool(raw) {
		cfg = raw
	}
	c := normalizeConfig(cfg, defaults(defaultCfg), true)
	return &c
}

func isBool(v interface{}) bool {
	_, ok := v.(bool)
	return ok
}

// jsTruthy mirrors JS's `Boolean(v)` for the value types ESLint options can
// carry through JSON: booleans, strings, numbers, null, arrays, and
// objects. Matters for schema-invalid inputs (e.g. `attributes: 0` or
// `children: null`) where ESLint's schema would reject before the rule
// runs but rslint does not perform JSON-schema validation.
func jsTruthy(v interface{}) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return x != ""
	case float64:
		return x != 0
	case float32:
		return x != 0
	case int:
		return x != 0
	case int32:
		return x != 0
	case int64:
		return x != 0
	}
	return true
}

// bodyScan summarizes the trivia / token layout of a JsxExpression or
// JsxSpreadAttribute body — the source range strictly between `{` and `}`.
//
// Field semantics (mirrors ESLint sourceCode helpers used by upstream):
//
//	secondPos       getTokenAfter(first,  { includeComments: true }).range[0]
//	penultimateEnd  getTokenBefore(last,  { includeComments: true }).range[1]
//	nextRealStart   getTokenAfter(first).range[0]              // skips comments
//	prevRealEnd     getTokenBefore(last).range[1]              // skips comments
//
// IMPORTANT — we only ever look at the trivia (whitespace + comments)
// IMMEDIATELY adjacent to the outer `{` and `}`. We deliberately do NOT
// tokenize the inner expression: the tsgo Scanner is stateless about
// template-literal substitution depth (`${ … }` is not re-scanned
// automatically without parser context), so feeding it the entire body
// would mis-classify
// the `}` of a `${expr}` as a real token and corrupt `penultimateEnd`.
// Hand-written byte scanning at the boundary is correct because:
//   - the leading "thing" can only be whitespace, a `/* … */` block
//     comment, a `// …` line comment, or the first character of the inner
//     expression — there are no template literals to navigate,
//   - symmetric reasoning for the trailing edge.
//
// When the body is empty / all-whitespace, secondPos == innerHigh,
// penultimateEnd == innerLow, nextRealStart == innerHigh,
// prevRealEnd == innerLow.
type bodyScan struct {
	secondPos      int
	penultimateEnd int
	nextRealStart  int
	prevRealEnd    int
}

func scanBraceBody(text string, low, high int) bodyScan {
	res := bodyScan{
		secondPos:      high,
		penultimateEnd: low,
		nextRealStart:  high,
		prevRealEnd:    low,
	}
	if low >= high {
		return res
	}

	// secondPos: skip leading whitespace only.
	res.secondPos = utils.SkipLeadingWhitespace(text, low, high)

	// penultimateEnd: skip trailing whitespace only.
	res.penultimateEnd = utils.SkipTrailingWhitespace(text, low, high)

	// nextRealStart: skip whitespace + leading-edge comments. The leading
	// edge cannot contain string/template literals (those would BE the
	// first token), so naive byte-level comment scanning is sufficient.
	p := low
	for p < high {
		if text[p] < 0x80 {
			if utils.IsTriviaWhitespaceByte(text[p]) {
				p++
				continue
			}
		} else {
			r, size := utf8.DecodeRuneInString(text[p:])
			if size > 0 && r != utf8.RuneError && utils.IsTriviaWhitespaceRune(r) {
				p += size
				continue
			}
		}
		if p+1 < high && text[p] == '/' && text[p+1] == '*' {
			p += 2
			for p+1 < high && (text[p] != '*' || text[p+1] != '/') {
				p++
			}
			if p+1 < high {
				p += 2
			} else {
				p = high
			}
			continue
		}
		if p+1 < high && text[p] == '/' && text[p+1] == '/' {
			p += 2
			for p < high && text[p] != '\n' {
				p++
			}
			continue
		}
		break
	}
	res.nextRealStart = p

	// prevRealEnd: skip trailing whitespace + trailing-edge block comments.
	// We do not reverse-scan line comments — `// … \n }` would require
	// finding the line start and verifying no string straddles it; for the
	// fix path the regex trims trailing whitespace per-line, which yields
	// the same output whether we land before `//` or after the comment
	// text. Aligned with upstream ESLint, which has the same gap.
	p = high
	for p > low {
		if text[p-1] < 0x80 {
			if utils.IsTriviaWhitespaceByte(text[p-1]) {
				p--
				continue
			}
		} else {
			r, size := utf8.DecodeLastRuneInString(text[:p])
			if size > 0 && r != utf8.RuneError && utils.IsTriviaWhitespaceRune(r) {
				p -= size
				continue
			}
		}
		if p-2 >= low && text[p-2] == '*' && text[p-1] == '/' {
			p -= 2
			for p-2 >= low && (text[p-2] != '/' || text[p-1] != '*') {
				p--
			}
			if p-2 >= low {
				p -= 2
			} else {
				p = low
			}
			continue
		}
		break
	}
	res.prevRealEnd = p

	return res
}

var (
	leadingWhitespacePerLine  = regexp.MustCompile(`(?m)^\s+`)
	trailingWhitespacePerLine = regexp.MustCompile(`(?m)\s+$`)
)

// fixByTrimmingWhitespace mirrors upstream's helper. mode is "start" (trim
// leading whitespace per line) or "end" (trim trailing). When `spacing` is
// "always", a single space is added on the trimmed side.
func fixByTrimmingWhitespace(text string, from, to int, mode, spacing string) rule.RuleFix {
	src := text[from:to]
	var replaced string
	if mode == "start" {
		replaced = leadingWhitespacePerLine.ReplaceAllString(src, "")
	} else {
		replaced = trailingWhitespacePerLine.ReplaceAllString(src, "")
	}
	if spacing == spacingAlways {
		if mode == "start" {
			replaced += " "
		} else {
			replaced = " " + replaced
		}
	}
	return rule.RuleFix{
		Text:  replaced,
		Range: core.NewTextRange(from, to),
	}
}

// JsxCurlySpacingRule enforces or disallows whitespace inside JSX braces in
// attribute and child positions. Listens for `JsxExpression` (covers the
// `JSXExpressionContainer` use-cases — both attribute initializers and
// element children) and `JsxSpreadAttribute` (attribute spread).
var JsxCurlySpacingRule = rule.Rule{
	Name: "react/jsx-curly-spacing",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		attrsConfig, childrenConfig := parseOptions(options)

		text := ctx.SourceFile.Text()

		validate := func(node *ast.Node) {
			parent := node.Parent
			if parent == nil {
				return
			}

			var cfg *sideConfig
			switch node.Kind {
			case ast.KindJsxExpression:
				// Children spread `<App>{...arr}</App>` is a JSXSpreadChild
				// node in ESTree — upstream's rule registers a listener for
				// JSXExpressionContainer only, so spread children are not
				// covered. tsgo collapses both shapes into a single
				// JsxExpression kind (distinguished by DotDotDotToken), so
				// skip the spread variant here to keep behavior 1:1 with
				// upstream.
				if je := node.AsJsxExpression(); je != nil && je.DotDotDotToken != nil {
					return
				}
				switch parent.Kind {
				case ast.KindJsxAttribute:
					cfg = attrsConfig
				case ast.KindJsxElement, ast.KindJsxFragment:
					cfg = childrenConfig
				default:
					return
				}
			case ast.KindJsxSpreadAttribute:
				cfg = attrsConfig
			default:
				return
			}
			if cfg == nil {
				return
			}

			trimmed := utils.TrimNodeTextRange(ctx.SourceFile, node)
			start := trimmed.Pos()
			end := trimmed.End()
			if start >= end || text[start] != '{' || text[end-1] != '}' {
				return
			}
			openPos := start
			closePos := end - 1
			innerLow := openPos + 1
			innerHigh := closePos

			scan := scanBraceBody(text, innerLow, innerHigh)

			// Emit position for `{` and `}` reports — anchor at the brace token.
			openRange := core.NewTextRange(openPos, openPos+1)
			closeRange := core.NewTextRange(closePos, closePos+1)

			// `isObjectLiteral` mirrors upstream's `first.value === second.value`
			// — the next inner token is itself a `{`. For empty `{}` (no inner
			// content), `secondPos == innerHigh`, so the next character is `}`,
			// and `isObjectLiteral` is correctly false. For TS wrappers like
			// `(`/`<`/`...`/`!`/`as`/`satisfies`, the first inner character is
			// not `{`, so they correctly use `cfg.when`.
			isObjectLiteral := scan.secondPos < innerHigh && text[scan.secondPos] == '{'
			spacing := cfg.when
			if isObjectLiteral {
				spacing = cfg.objectLiteralSpaces
			}

			hasSpaceAfterOpen := scan.secondPos > innerLow
			hasSpaceBeforeClose := scan.penultimateEnd < innerHigh
			multilineAfterOpen := utils.ContainsLineTerminator(text, innerLow, scan.secondPos)
			multilineBeforeClose := utils.ContainsLineTerminator(text, scan.penultimateEnd, innerHigh)

			reportNoBeginningSpace := func() {
				toLoc := scan.nextRealStart
				if scan.secondPos < scan.nextRealStart {
					// secondPos is the start of a leading comment;
					// upstream trims only up to min(realToken, comment).
					toLoc = scan.secondPos
				}
				ctx.ReportRangeWithFixes(
					openRange,
					rule.RuleMessage{
						Id:          "noSpaceAfter",
						Description: "There should be no space after '{'",
					},
					fixByTrimmingWhitespace(text, innerLow, toLoc, "start", ""),
				)
			}

			reportNoEndingSpace := func() {
				fromLoc := scan.prevRealEnd
				if scan.penultimateEnd > scan.prevRealEnd {
					// penultimateEnd is the end of a trailing comment;
					// upstream trims only from max(realToken, comment).
					fromLoc = scan.penultimateEnd
				}
				ctx.ReportRangeWithFixes(
					closeRange,
					rule.RuleMessage{
						Id:          "noSpaceBefore",
						Description: "There should be no space before '}'",
					},
					fixByTrimmingWhitespace(text, fromLoc, innerHigh, "end", ""),
				)
			}

			reportNoBeginningNewline := func() {
				ctx.ReportRangeWithFixes(
					openRange,
					rule.RuleMessage{
						Id:          "noNewlineAfter",
						Description: "There should be no newline after '{'",
					},
					fixByTrimmingWhitespace(text, innerLow, scan.nextRealStart, "start", spacing),
				)
			}

			reportNoEndingNewline := func() {
				ctx.ReportRangeWithFixes(
					closeRange,
					rule.RuleMessage{
						Id:          "noNewlineBefore",
						Description: "There should be no newline before '}'",
					},
					fixByTrimmingWhitespace(text, scan.prevRealEnd, innerHigh, "end", spacing),
				)
			}

			reportRequiredBeginningSpace := func() {
				ctx.ReportRangeWithFixes(
					openRange,
					rule.RuleMessage{
						Id:          "spaceNeededAfter",
						Description: "A space is required after '{'",
					},
					rule.RuleFix{
						Text:  " ",
						Range: core.NewTextRange(openPos+1, openPos+1),
					},
				)
			}

			reportRequiredEndingSpace := func() {
				ctx.ReportRangeWithFixes(
					closeRange,
					rule.RuleMessage{
						Id:          "spaceNeededBefore",
						Description: "A space is required before '}'",
					},
					rule.RuleFix{
						Text:  " ",
						Range: core.NewTextRange(closePos, closePos),
					},
				)
			}

			switch spacing {
			case spacingAlways:
				if !hasSpaceAfterOpen {
					reportRequiredBeginningSpace()
				} else if !cfg.allowMultiline && multilineAfterOpen {
					reportNoBeginningNewline()
				}
				if !hasSpaceBeforeClose {
					reportRequiredEndingSpace()
				} else if !cfg.allowMultiline && multilineBeforeClose {
					reportNoEndingNewline()
				}
			case spacingNever:
				if multilineAfterOpen {
					if !cfg.allowMultiline {
						reportNoBeginningNewline()
					}
				} else if hasSpaceAfterOpen {
					reportNoBeginningSpace()
				}
				if multilineBeforeClose {
					if !cfg.allowMultiline {
						reportNoEndingNewline()
					}
				} else if hasSpaceBeforeClose {
					reportNoEndingSpace()
				}
			}
		}

		return rule.RuleListeners{
			ast.KindJsxExpression:      validate,
			ast.KindJsxSpreadAttribute: validate,
		}
	},
}
