// Package jsx_first_prop_new_line implements jsx-first-prop-new-line
// (`react/jsx-first-prop-new-line`).
//
// The rule enforces the position of the first prop in a JSX element. Five modes
// select when the first prop must (or must not) sit on its own line:
//
//	always               first prop must always be on a new line
//	never                first prop must never be on a new line
//	multiline            new line required when the opening tag spans >1 line
//	multiline-multiprop  (default) as multiline, but only when there is >1 prop
//	multiprop            new line required when there is >1 prop; with <=1 prop
//	                     on a multiline tag the prop must be on the same line
//
// Upstream listens on ESTree's JSXOpeningElement, which covers both the
// `<App .../>` self-closing form and the `<App ...>` form. tsgo splits these
// into KindJsxSelfClosingElement and KindJsxOpeningElement, so the rule
// registers a listener for each.
package jsx_first_prop_new_line

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// JsxFirstPropNewLineRule is the eslint-plugin-react variant.
var JsxFirstPropNewLineRule = BuildRule("react/jsx-first-prop-new-line")

// BuildRule constructs the jsx-first-prop-new-line rule registered under name.
func BuildRule(name string) rule.Rule {
	return rule.Rule{
		Name: name,
		Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
			options := rule.UnwrapOptions(_options)
			// Default option
			option := "multiline-multiprop"

			// Parse options. The schema is a single string enum
			// ('always' | 'never' | 'multiline' | 'multiline-multiprop' |
			// 'multiprop'), delivered either bare (single-option CLI form) or as
			// a one-element array (rule-tester form). There is no object form, so
			// no map handling is needed.
			if options != nil {
				if optArray, ok := options.([]interface{}); ok && len(optArray) > 0 {
					if s, ok := optArray[0].(string); ok {
						option = s
					}
				} else if s, ok := options.(string); ok {
					option = s
				}
			}

			check := func(node *ast.Node) {
				var props []*ast.Node
				var openingNode *ast.Node

				switch node.Kind {
				case ast.KindJsxOpeningElement:
					opening := node.AsJsxOpeningElement()
					openingNode = node
					attrs := opening.Attributes.AsJsxAttributes()
					if attrs.Properties != nil {
						props = attrs.Properties.Nodes
					}
				case ast.KindJsxSelfClosingElement:
					self := node.AsJsxSelfClosingElement()
					openingNode = node
					attrs := self.Attributes.AsJsxAttributes()
					if attrs.Properties != nil {
						props = attrs.Properties.Nodes
					}
				}

				if len(props) == 0 {
					return
				}

				lineMap := ctx.SourceFile.ECMALineMap()
				text := ctx.SourceFile.Text()
				firstProp := props[0]

				// Use the trimmed position for the opening element
				openingTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, openingNode)
				openingLine := scanner.ComputeLineOfPosition(lineMap, openingTrimmed.Pos())
				firstPropTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, firstProp)
				firstPropLine := scanner.ComputeLineOfPosition(lineMap, firstPropTrimmed.Pos())

				// Determine if element is multiline (opening tag start to opening tag end)
				openingEndLine := scanner.ComputeLineOfPosition(lineMap, openingTrimmed.End())
				isMultiline := openingLine != openingEndLine

				propOnNewLineMsg := rule.RuleMessage{
					Id:          "propOnNewLine",
					Description: "Property should be placed on a new line",
				}
				propOnSameLineMsg := rule.RuleMessage{
					Id:          "propOnSameLine",
					Description: "Property should be placed on the same line as the component declaration",
				}

				// NOTE: Unlike ESLint — which replaces the whole range from the
				// tag name (or type arguments) end to the first prop — both fixes
				// below only consume the whitespace directly adjacent to the prop.
				// A comment or type argument sitting between the tag name and the
				// first prop is therefore preserved rather than deleted.
				//
				// Fix: replace whitespace before prop with newline
				newLineFix := func() rule.RuleFix {
					propStart := firstPropTrimmed.Pos()
					wsStart := propStart
					for wsStart > 0 && (text[wsStart-1] == ' ' || text[wsStart-1] == '\t') {
						wsStart--
					}
					return rule.RuleFix{Text: "\n", Range: core.NewTextRange(wsStart, propStart)}
				}
				// Fix: replace whitespace/newlines before prop with a space.
				// Returns nil (report without autofix) when non-whitespace — e.g.
				// a comment — sits between the tag name and the first prop: wsStart
				// then stops before the component-declaration line, so collapsing
				// only the adjacent whitespace would leave the prop on a lower line
				// and the violation would never clear (the autofix would loop).
				sameLineFix := func() []rule.RuleFix {
					propStart := firstPropTrimmed.Pos()
					wsStart := propStart
					for wsStart > 0 && (text[wsStart-1] == ' ' || text[wsStart-1] == '\t' || text[wsStart-1] == '\n' || text[wsStart-1] == '\r') {
						wsStart--
					}
					if scanner.ComputeLineOfPosition(lineMap, wsStart) != openingLine {
						return nil
					}
					return []rule.RuleFix{{Text: " ", Range: core.NewTextRange(wsStart, propStart)}}
				}

				switch option {
				case "always":
					if openingLine == firstPropLine {
						ctx.ReportNodeWithFixes(firstProp, propOnNewLineMsg, newLineFix())
					}
				case "never":
					if openingLine != firstPropLine {
						ctx.ReportNodeWithFixes(firstProp, propOnSameLineMsg, sameLineFix()...)
					}
				case "multiline":
					if isMultiline && openingLine == firstPropLine {
						ctx.ReportNodeWithFixes(firstProp, propOnNewLineMsg, newLineFix())
					}
				case "multiline-multiprop":
					if isMultiline && len(props) > 1 && openingLine == firstPropLine {
						ctx.ReportNodeWithFixes(firstProp, propOnNewLineMsg, newLineFix())
					}
				case "multiprop":
					if len(props) > 1 && openingLine == firstPropLine {
						ctx.ReportNodeWithFixes(firstProp, propOnNewLineMsg, newLineFix())
					} else if len(props) <= 1 && isMultiline && openingLine != firstPropLine {
						ctx.ReportNodeWithFixes(firstProp, propOnSameLineMsg, sameLineFix()...)
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
