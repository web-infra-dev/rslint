package jsx_tag_spacing

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed jsx_tag_spacing.schema.json
var schemaJSON []byte

var messages = map[string]string{
	"selfCloseSlashNoSpace":      "Whitespace is forbidden between `/` and `>`; write `/>`",
	"selfCloseSlashNeedSpace":    "Whitespace is required between `/` and `>`; write `/ >`",
	"closeSlashNoSpace":          "Whitespace is forbidden between `<` and `/`; write `</`",
	"closeSlashNeedSpace":        "Whitespace is required between `<` and `/`; write `< /`",
	"beforeSelfCloseNoSpace":     "A space is forbidden before closing bracket",
	"beforeSelfCloseNeedSpace":   "A space is required before closing bracket",
	"beforeSelfCloseNeedNewline": "A newline is required before closing bracket",
	"afterOpenNoSpace":           "A space is forbidden after opening bracket",
	"afterOpenNeedSpace":         "A space is required after opening bracket",
	"beforeCloseNoSpace":         "A space is forbidden before closing bracket",
	"beforeCloseNeedSpace":       "Whitespace is required before closing bracket",
	"beforeCloseNeedNewline":     "A newline is required before closing bracket",
}

var JsxTagSpacingRule = rule.Rule{
	Name:   "react/jsx-tag-spacing",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		config := map[string]string{
			"closingSlash": "never", "beforeSelfClosing": "always",
			"afterOpening": "never", "beforeClosing": "allow",
		}
		if len(options) > 0 {
			if values, ok := options[0].(map[string]any); ok {
				for key := range config {
					if value, ok := values[key].(string); ok {
						config[key] = value
					}
				}
			}
		}

		sf := ctx.SourceFile
		s := scanner.NewScanner()
		s.SetText(sf.Text())
		s.SetSkipTrivia(false)
		// ESLint's isSpaceBetweenTokens ignores whitespace inside comments.
		spaced := func(left, right core.TextRange) bool {
			s.ResetTokenState(left.End())
			for pos := left.End(); pos < right.Pos(); pos = s.TokenEnd() {
				// Scan skips U+2028/U+2029 even with trivia enabled.
				if ecmascript.SkipLeadingWhitespace(sf.Text(), pos, right.Pos()) != pos {
					return true
				}
				if s.Scan() == ast.KindEndOfFile {
					break
				}
			}
			return false
		}
		report := func(id string, loc, edit core.TextRange, text string) {
			ctx.ReportRangeWithFixes(loc, rule.RuleMessage{Id: id, Description: messages[id]}, rule.RuleFix{Range: edit, Text: text})
		}
		insert := func(pos int) core.TextRange { return core.NewTextRange(pos, pos) }
		checkSlash := func(left, right core.TextRange, prefix string) {
			loc := core.NewTextRange(left.Pos(), right.End())
			switch config["closingSlash"] {
			case "never":
				if spaced(left, right) {
					report(prefix+"NoSpace", loc, core.NewTextRange(left.End(), right.Pos()), "")
				}
			case "always":
				if !spaced(left, right) {
					report(prefix+"NeedSpace", loc, insert(right.Pos()), " ")
				}
			}
		}
		check := func(node *ast.Node) {
			selfClosing := node.Kind == ast.KindJsxSelfClosingElement
			closing := node.Kind == ast.KindJsxClosingElement
			name := reactutil.GetJsxTagName(node)
			if closing {
				name = node.AsJsxClosingElement().TagName
			}
			if name == nil {
				return
			}
			// Only the tag's immediate children are needed for its brackets.
			// Walking every attribute expression repeats work for nested JSX.
			children := utils.GetChildren(node, sf)
			if len(children) < 3 || children[len(children)-1].Kind != ast.KindGreaterThanToken {
				return
			}
			first := utils.TrimNodeTextRange(sf, children[0])
			last := utils.TrimNodeTextRange(sf, children[len(children)-1])
			nameRange := utils.TrimNodeTextRange(sf, name)
			opening := first
			if closing {
				// tsgo stores `</` as one token; ESTree exposes `<` and `/`.
				opening = core.NewTextRange(first.End()-1, first.End())
			}
			if selfClosing {
				checkSlash(utils.TrimNodeTextRange(sf, children[len(children)-2]), last, "selfCloseSlash")
			}
			option := config["afterOpening"]
			if option != "allow" && (option != "allow-multiline" || utils.IsSameLine(sf, opening.Pos(), nameRange.Pos())) {
				loc := core.NewTextRange(opening.Pos(), nameRange.Pos())
				if (option == "never" || option == "allow-multiline") && spaced(opening, nameRange) {
					report("afterOpenNoSpace", loc, core.NewTextRange(opening.End(), nameRange.Pos()), "")
				} else if option == "always" && !spaced(opening, nameRange) {
					report("afterOpenNeedSpace", loc, insert(nameRange.Pos()), " ")
				}
			}
			if closing {
				checkSlash(core.NewTextRange(first.Pos(), first.Pos()+1), opening, "closeSlash")
			}

			// Upstream uses the whole last attribute (or tag name) for self-closing
			// and proportional checks, but the final token for ordinary `>` checks.
			left := nameRange
			if attrs := reactutil.GetJsxElementAttributes(node); len(attrs) > 0 {
				left = utils.TrimNodeTextRange(sf, attrs[len(attrs)-1])
			}
			if selfClosing {
				option = config["beforeSelfClosing"]
			} else {
				option = config["beforeClosing"]
				if option != "allow" && option != "proportional-always" {
					previous, ok := utils.TokenBeforePosition(sf, last.Pos())
					if !ok {
						return
					}
					left = previous.Range()
				}
			}
			if option == "allow" {
				return
			}
			next, ok := utils.TokenAtOrAfter(sf, left.End())
			if !ok {
				return
			}
			right := next.Range()
			multiline := !utils.IsSameLine(sf, first.Pos(), last.End())
			prefix := "beforeClose"
			if selfClosing {
				prefix = "beforeSelfClose"
			}
			if option == "proportional-always" && multiline && utils.IsSameLine(sf, left.End(), right.Pos()) {
				report(prefix+"NeedNewline", insert(left.End()), insert(right.Pos()), "\n")
				return
			}
			linePos := left.Pos()
			if selfClosing {
				linePos = left.End()
			}
			if !utils.IsSameLine(sf, linePos, right.Pos()) {
				return
			}
			adjacent := !spaced(left, right)
			loc := core.NewTextRange(left.End(), right.Pos())
			if selfClosing {
				loc = insert(right.Pos())
			}
			if option == "never" && !adjacent {
				report(prefix+"NoSpace", loc, core.NewTextRange(left.End(), right.Pos()), "")
			} else if (option == "always" || selfClosing && option == "proportional-always") && adjacent {
				report(prefix+"NeedSpace", loc, insert(right.Pos()), " ")
			} else if option == "proportional-always" && !closing && !selfClosing && adjacent == multiline {
				report("beforeCloseNeedSpace", loc, insert(right.Pos()), " ")
			}
		}
		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
			ast.KindJsxClosingElement:     check,
		}
	},
}
