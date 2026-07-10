// Package jsx_equals_spacing implements jsx-equals-spacing
// (`react/jsx-equals-spacing`).
//
// The rule enforces or disallows spaces around the `=` sign in JSX attributes:
// under `never` (default) there must be no space on either side of `=`, and
// under `always` a single space is required on both sides. Spread attributes
// (`{...props}`) and valueless attributes (`<App foo />`) are never checked.
//
// Upstream listens on ESTree's `JSXOpeningElement`, which covers both the
// `<App ... />` self-closing form and the `<App ...>` form. tsgo splits these
// into KindJsxSelfClosingElement and KindJsxOpeningElement, so the rule
// registers a listener for each. Spacing is decided the way ESLint's
// `sourceCode.isSpaceBetween` does: a whitespace/newline token anywhere between
// the two operands counts as a space, while the inner bytes of a comment do
// NOT (a comment is a single token, so its content never surfaces as a
// whitespace gap). The diagnostic is reported on the `=` token itself, matching
// upstream's `loc: equalToken.loc.start`.
package jsx_equals_spacing

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// JsxEqualsSpacingRule is the eslint-plugin-react variant.
var JsxEqualsSpacingRule = BuildRule("react/jsx-equals-spacing")

const defaultConfig = "never"

var messages = map[string]string{
	"noSpaceBefore":   "There should be no space before '='",
	"noSpaceAfter":    "There should be no space after '='",
	"needSpaceBefore": "A space is required before '='",
	"needSpaceAfter":  "A space is required after '='",
}

// parseOption accepts every shape rslint's loader can deliver for upstream's
// single-string schema (`enum: ['always', 'never']`, default `never`): nil, a
// bare string (single-option CLI form), or a single-element array (rule-tester
// form). Any value outside the enum falls back to the default — upstream's
// JSON schema rejects those before the rule runs, so that branch is only
// reachable through hand-written rslint configs.
func parseOption(options any) string {
	raw := options
	if arr, ok := options.([]interface{}); ok {
		if len(arr) == 0 {
			return defaultConfig
		}
		raw = arr[0]
	}
	if s, ok := raw.(string); ok && (s == "always" || s == "never") {
		return s
	}
	return defaultConfig
}

// equalsInfo records, for one `name = value` JSX attribute, everything the
// rule needs: the `=` token span, the start of the value node, and whether a
// space (in ESLint's `isSpaceBetween` sense) sits on each side of `=`.
type equalsInfo struct {
	eqPos        int
	eqEnd        int
	valueStart   int
	spacedBefore bool
	spacedAfter  bool
}

// analyzeEquals scans forward from `nameEnd` (the end of the attribute name)
// to locate the `=` token and the start of the value, classifying the spacing
// on both sides. Returns ok=false on malformed input where no `=` / value can
// be found (parser recovery shapes) so the caller skips silently.
//
// Spacing mirrors `sourceCode.isSpaceBetween`: only whitespace/newline trivia
// tokens between the operands count as a space. Comments are skipped as
// opaque tokens, so `foo/*c*/=` is treated as un-spaced while `foo /*c*/=`
// (a real space before the comment) is spaced — exactly as ESLint walks the
// `name → comment → =` token chain looking for a gap. Locating `=` via the
// scanner (rather than a raw `text[i] == '='` search) also keeps a `=`
// character inside a comment from being mistaken for the operator.
func analyzeEquals(s *scanner.Scanner, nameEnd int) (equalsInfo, bool) {
	info := equalsInfo{eqPos: -1, eqEnd: -1, valueStart: -1}
	s.ResetTokenState(nameEnd)
	foundEq := false
	for {
		k := s.Scan()
		if k == ast.KindEndOfFile {
			return equalsInfo{}, false
		}
		start, end := s.TokenStart(), s.TokenEnd()
		isWS := k == ast.KindWhitespaceTrivia || k == ast.KindNewLineTrivia
		isComment := k == ast.KindSingleLineCommentTrivia || k == ast.KindMultiLineCommentTrivia
		if !foundEq {
			if k == ast.KindEqualsToken {
				info.eqPos, info.eqEnd = start, end
				foundEq = true
				continue
			}
			if isWS {
				info.spacedBefore = true
				continue
			}
			if isComment {
				continue
			}
			// A non-trivia token other than `=` before the value — only
			// reachable on malformed source; bail.
			return equalsInfo{}, false
		}
		if isWS {
			info.spacedAfter = true
			continue
		}
		if isComment {
			continue
		}
		// First non-trivia, non-comment token after `=` is the value node's
		// first token (ESLint's `value.range[0]`).
		info.valueStart = start
		return info, true
	}
}

// BuildRule constructs the jsx-equals-spacing rule registered under `name`.
func BuildRule(name string) rule.Rule {
	return rule.Rule{
		Name: name,
		Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
			options := rule.LegacyUnwrapOptions(_options)
			config := parseOption(options)
			sf := ctx.SourceFile
			if sf == nil {
				return rule.RuleListeners{}
			}

			// One scanner reused across every attribute in the file; each
			// analyzeEquals call resets its position via ResetTokenState.
			s := scanner.NewScanner()
			s.SetText(sf.Text())
			s.SetLanguageVariant(sf.LanguageVariant)
			s.SetSkipTrivia(false)

			check := func(node *ast.Node) {
				var attrs []*ast.Node
				switch node.Kind {
				case ast.KindJsxOpeningElement:
					if attributes := node.AsJsxOpeningElement().Attributes; attributes != nil {
						if props := attributes.AsJsxAttributes().Properties; props != nil {
							attrs = props.Nodes
						}
					}
				case ast.KindJsxSelfClosingElement:
					if attributes := node.AsJsxSelfClosingElement().Attributes; attributes != nil {
						if props := attributes.AsJsxAttributes().Properties; props != nil {
							attrs = props.Nodes
						}
					}
				default:
					return
				}

				for _, attr := range attrs {
					// Upstream skips when the attribute is a spread OR has no
					// value: `!(type !== 'JSXSpreadAttribute' && value !== null)`.
					if attr.Kind != ast.KindJsxAttribute {
						continue
					}
					jsxAttr := attr.AsJsxAttribute()
					if jsxAttr.Initializer == nil {
						continue
					}
					nameNode := jsxAttr.Name()
					if nameNode == nil {
						continue
					}
					nameEnd := nameNode.End()

					info, ok := analyzeEquals(s, nameEnd)
					if !ok {
						continue
					}

					reportRange := core.NewTextRange(info.eqPos, info.eqEnd)

					if config == "never" {
						if info.spacedBefore {
							ctx.ReportRangeWithFixes(reportRange, rule.RuleMessage{
								Id:          "noSpaceBefore",
								Description: messages["noSpaceBefore"],
							}, rule.RuleFix{
								Text:  "",
								Range: core.NewTextRange(nameEnd, info.eqPos),
							})
						}
						if info.spacedAfter {
							ctx.ReportRangeWithFixes(reportRange, rule.RuleMessage{
								Id:          "noSpaceAfter",
								Description: messages["noSpaceAfter"],
							}, rule.RuleFix{
								Text:  "",
								Range: core.NewTextRange(info.eqEnd, info.valueStart),
							})
						}
					} else { // "always"
						if !info.spacedBefore {
							ctx.ReportRangeWithFixes(reportRange, rule.RuleMessage{
								Id:          "needSpaceBefore",
								Description: messages["needSpaceBefore"],
							}, rule.RuleFix{
								Text:  " ",
								Range: core.NewTextRange(info.eqPos, info.eqPos),
							})
						}
						if !info.spacedAfter {
							ctx.ReportRangeWithFixes(reportRange, rule.RuleMessage{
								Id:          "needSpaceAfter",
								Description: messages["needSpaceAfter"],
							}, rule.RuleFix{
								Text:  " ",
								Range: core.NewTextRange(info.eqEnd, info.eqEnd),
							})
						}
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
