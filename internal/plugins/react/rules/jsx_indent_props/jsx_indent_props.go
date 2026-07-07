package jsx_indent_props

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// JsxIndentPropsRule is the eslint-plugin-react variant of jsx-indent-props.
var JsxIndentPropsRule = BuildRule("react/jsx-indent-props")

// BuildRule constructs the jsx-indent-props rule registered under name.
//
// Ported from eslint-plugin-react's `jsx-indent-props` rule. The rule walks
// every JsxOpeningElement / JsxSelfClosingElement, computes the expected
// prop indent (either `<element-indent> + indentSize` for the numeric / tab
// modes, or the column of the first prop for `'first'`), and reports each
// attribute whose leading-whitespace count differs.
//
// tsgo↔ESTree shape adjustments (vs the upstream JS rule):
//   - Self-closing `<Foo />` is `JsxSelfClosingElement` with no wrapping
//     `JsxElement`, so both `JsxOpeningElement` and `JsxSelfClosingElement`
//     share the same listener.
//   - Source-line scans (line-start, leading whitespace, first non-WS char)
//     run against `SourceFile.Text()`; line numbers come from the
//     ECMA line map via `scanner.ComputeLineOfPosition`.
//
// Ternary operator handling: upstream maintains a `line.isUsingOperator`
// flag, updated by `getNodeIndent` from each scanned line, that grants the
// next first-in-line prop an extra `indentSize` bump (when `'first'` mode and
// `ignoreTernaryOperator` are both off). We replicate `getNodeIndent`'s
// per-prop side effect with the exact same precedence: a line starting with
// `?`/`:` after whitespace (useOperator) sets the flag and PRECEDES the
// `<`-contains reset (useBracket). That precedence is load-bearing for the
// shape where the first prop shares the opening tag's `?`/`:` line — the line
// both starts with the operator and contains `<`, and upstream keeps the flag
// set so the bump carries to the following props.
func BuildRule(name string) rule.Rule {
	return rule.Rule{
		Name: name,
		Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
			options := rule.UnwrapOptions(_options)
			indentType, indentSize, indentChar, indentIsFirst, ignoreTernaryOperator := parseOptions(options)

			text := ctx.SourceFile.Text()
			lineMap := ctx.SourceFile.ECMALineMap()

			// firstNonWhitespaceCharOnLine returns the first non space / tab
			// character on the line containing pos, or 0 if none.
			firstNonWhitespaceCharOnLine := func(pos int) byte {
				start := reactutil.IndentLineStart(lineMap, pos)
				for i := start; i < len(text); i++ {
					c := text[i]
					if c == ' ' || c == '\t' {
						continue
					}
					if c == '\n' || c == '\r' {
						return 0
					}
					return c
				}
				return 0
			}

			// firstLineContainsBracket reports whether the first source line
			// of node (from line-start of trimmed start to the next newline
			// or the trimmed end, whichever comes first) contains a `<`.
			// Mirrors upstream's `useBracket` check inside `getNodeIndent`,
			// which resets `isUsingOperator` whenever the scanned line
			// includes a `<` (i.e. the prop is on the same line as the
			// element's opening `<`, or carries an inline JSX-literal value).
			firstLineContainsBracket := func(node *ast.Node) bool {
				trimmed := utils.TrimNodeTextRange(ctx.SourceFile, node)
				lineStart := reactutil.IndentLineStart(lineMap, trimmed.Pos())
				end := trimmed.End()
				for i := lineStart; i < end && i < len(text); i++ {
					c := text[i]
					if c == '\n' {
						break
					}
					if c == '<' {
						return true
					}
				}
				return false
			}

			check := func(node *ast.Node) {
				var props []*ast.Node
				switch node.Kind {
				case ast.KindJsxOpeningElement:
					opening := node.AsJsxOpeningElement()
					if opening.Attributes != nil {
						attrs := opening.Attributes.AsJsxAttributes()
						if attrs != nil && attrs.Properties != nil {
							props = attrs.Properties.Nodes
						}
					}
				case ast.KindJsxSelfClosingElement:
					self := node.AsJsxSelfClosingElement()
					if self.Attributes != nil {
						attrs := self.Attributes.AsJsxAttributes()
						if attrs != nil && attrs.Properties != nil {
							props = attrs.Properties.Nodes
						}
					}
				}
				if len(props) == 0 {
					return
				}

				openingTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, node)
				openingStart := openingTrimmed.Pos()

				// isUsingOperator: line containing the opening `<` starts
				// with `?` or `:` after whitespace.
				leadChar := firstNonWhitespaceCharOnLine(openingStart)
				isUsingOperator := leadChar == '?' || leadChar == ':'

				elementIndent := reactutil.IndentLeading(text, lineMap, openingStart, indentChar)

				var propIndent int
				if indentIsFirst {
					// 'first' mode aligns to the visual column of the
					// first prop. Use UTF-16 character column (matches
					// ESLint's `loc.start.column`) instead of byte
					// offset so multi-byte characters preceding the
					// first prop don't inflate the expected indent.
					propIndent = reactutil.NodeStartUTF16Column(ctx.SourceFile, props[0])
				} else {
					propIndent = elementIndent + indentSize
				}

				nestedIndent := propIndent

				for _, prop := range props {
					// Mirror upstream's getNodeIndent side effect, which runs
					// for EVERY prop (before the first-in-line report gate). It
					// updates the ternary-operator state from the prop's first
					// source line, with useOperator (line starts with `?`/`:`
					// after whitespace) taking PRIORITY over useBracket (line
					// contains `<`) — exactly upstream's if/else ordering. The
					// priority matters when the first prop shares the opening
					// tag's `?`/`:` line: that line both starts with the
					// operator and contains `<`, and upstream keeps
					// isUsingOperator set so the bump carries to the next prop.
					propLead := firstNonWhitespaceCharOnLine(utils.TrimNodeTextRange(ctx.SourceFile, prop).Pos())
					currentOperator := false
					if propLead == '?' || propLead == ':' {
						isUsingOperator = true
						currentOperator = true
					} else if firstLineContainsBracket(prop) {
						isUsingOperator = false
					}

					// Bump decision, also run for every prop. A prop whose own
					// line starts with the operator (currentOperator) does not
					// consume the bump; the first following prop that doesn't
					// re-arm the operator does. Applying it clears
					// isUsingOperator so the bump fires at most once.
					if isUsingOperator && !currentOperator && !indentIsFirst && !ignoreTernaryOperator {
						nestedIndent += indentSize
						isUsingOperator = false
					}

					if !reactutil.IsNodeFirstInLine(ctx.SourceFile, prop) {
						continue
					}

					actualIndent := reactutil.NodeStartIndent(ctx.SourceFile, prop, indentChar)
					if actualIndent != nestedIndent {
						reactutil.ReportIndentReplaceLeading(ctx, prop, nestedIndent, actualIndent, indentChar, indentType)
					}
				}
			}

			return rule.RuleListeners{
				ast.KindJsxOpeningElement:     check,
				ast.KindJsxSelfClosingElement: check,
			}
		},
	}
}

// parseOptions parses the rule options. The first positional option may be:
//   - "tab"                     → tab indent, size 1
//   - "first"                   → align with first prop's column
//   - integer N                 → N-space indent
//   - object { indentMode, ignoreTernaryOperator } → use indentMode as
//     above, plus optional ignoreTernaryOperator boolean
//
// Default: 4-space indent, ignoreTernaryOperator off.
func parseOptions(options any) (indentType string, indentSize int, indentChar byte, indentIsFirst bool, ignoreTernaryOperator bool) {
	indentType = "space"
	indentSize = 4
	indentChar = ' '

	var first any
	if options == nil {
		return indentType, indentSize, indentChar, indentIsFirst, ignoreTernaryOperator
	}
	if arr, ok := options.([]interface{}); ok {
		if len(arr) > 0 {
			first = arr[0]
		}
	} else {
		first = options
	}
	if first == nil {
		return indentType, indentSize, indentChar, indentIsFirst, ignoreTernaryOperator
	}

	var indentMode any
	if m, ok := first.(map[string]interface{}); ok {
		indentMode = m["indentMode"]
		if v, ok := m["ignoreTernaryOperator"].(bool); ok {
			ignoreTernaryOperator = v
		}
	} else {
		indentMode = first
	}

	switch v := indentMode.(type) {
	case string:
		switch v {
		case "first":
			indentIsFirst = true
			indentType = "space"
			indentChar = ' '
		case "tab":
			indentType = "tab"
			indentSize = 1
			indentChar = '\t'
		}
	case float64:
		indentType = "space"
		indentSize = int(v)
		indentChar = ' '
	case int:
		indentType = "space"
		indentSize = v
		indentChar = ' '
	}
	return indentType, indentSize, indentChar, indentIsFirst, ignoreTernaryOperator
}
