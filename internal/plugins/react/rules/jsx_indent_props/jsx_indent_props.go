package jsx_indent_props

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// JsxIndentPropsRule enforces JSX prop indentation.
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
// flag that is set when the line containing the JSX `<` starts with `?` or
// `:` after whitespace. When that flag is on (and `'first'` mode is off and
// `ignoreTernaryOperator` is off), the FIRST prop that sits first-in-line
// receives an extra `indentSize` bump, after which the flag is cleared.
// We replicate that behaviour with a per-element `bumpApplied` boolean and
// a per-prop bracket reset (mirrors upstream's `useBracket` reset inside
// `getNodeIndent` — any prop whose first source line contains `<` cancels
// the pending bump, exactly the same way upstream's per-call side effect
// would).
var JsxIndentPropsRule = rule.Rule{
	Name: "react/jsx-indent-props",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
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
				firstPropTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, props[0])
				propIndent = firstPropTrimmed.Pos() - reactutil.IndentLineStart(lineMap, firstPropTrimmed.Pos())
			} else {
				propIndent = elementIndent + indentSize
			}

			nestedIndent := propIndent
			bumpApplied := false

			for _, prop := range props {
				// Mirror upstream: `getNodeIndent(node)` is called per
				// prop regardless of first-in-line status, and resets
				// `isUsingOperator` when the scanned line contains a
				// `<`. Apply the same reset here so a prop whose first
				// source line includes `<` (e.g. on the element's own
				// `<` line, or carrying an inline JSX-literal value)
				// cancels the bump.
				if firstLineContainsBracket(prop) {
					isUsingOperator = false
				}

				if !reactutil.IsNodeFirstInLine(ctx.SourceFile, prop) {
					continue
				}

				if !bumpApplied && isUsingOperator && !indentIsFirst && !ignoreTernaryOperator {
					nestedIndent += indentSize
					bumpApplied = true
					isUsingOperator = false
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
