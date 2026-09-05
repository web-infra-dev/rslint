package self_closing_comp

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed self_closing_comp.schema.json
var schemaJSON []byte

var SelfClosingCompRule = rule.Rule{
	Name:   "react/self-closing-comp",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		checkComponent := true
		checkHTML := true

		if len(options) > 0 {
			optsMap, _ := options[0].(map[string]interface{})
			if b, ok := optsMap["component"].(bool); ok {
				checkComponent = b
			}
			if b, ok := optsMap["html"].(bool); ok {
				checkHTML = b
			}
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement: func(node *ast.Node) {
				parent := node.Parent
				if parent == nil || parent.Kind != ast.KindJsxElement {
					return
				}

				jsxElement := parent.AsJsxElement()

				// Check if the element has no meaningful children
				if !isChildrenEmpty(jsxElement) {
					return
				}

				openingElement := node.AsJsxOpeningElement()
				tagName := openingElement.TagName

				isComp := isComponent(ctx.SourceFile, tagName)

				if isComp && !checkComponent {
					return
				}
				if !isComp && !checkHTML {
					return
				}

				// Build the fix: replace from the end of opening tag's `>` to end of closing tag `</tag>` with ` />`
				// The opening element ends at its `>`. We need to find the `>` position.
				// The opening element node ends at the `>` character (inclusive of it).
				// The closing element ends at `>` of `</tag>`.
				openEnd := node.End()    // position after the `>` of opening element
				closeEnd := parent.End() // position after the `>` of closing element

				// Replace from position of `>` (openEnd - 1) to closeEnd with ` />`
				fixRange := core.NewTextRange(openEnd-1, closeEnd)
				fix := rule.RuleFix{
					Text:  " />",
					Range: fixRange,
				}

				ctx.ReportNodeWithFixes(parent, rule.RuleMessage{
					Id:          "notSelfClosing",
					Description: "Empty components are self-closing",
				}, fix)
			},
		}
	},
}

// isChildrenEmpty checks if a JsxElement has no children or only whitespace text children.
func isChildrenEmpty(jsxElement *ast.JsxElement) bool {
	if jsxElement.Children == nil || len(jsxElement.Children.Nodes) == 0 {
		return true
	}
	for _, child := range jsxElement.Children.Nodes {
		if child.Kind == ast.KindJsxText {
			text := child.AsJsxText()
			if text.ContainsOnlyTriviaWhiteSpaces {
				continue
			}
			// Only treat as empty if whitespace contains a newline (matching ESLint behavior)
			if ecmascript.StringTrim(text.Text) == "" && strings.Contains(text.Text, "\n") {
				continue
			}
			return false
		}
		return false
	}
	return true
}

// isComponent returns true if the tag name is not a DOM tag: upstream reads
// `!jsxUtil.isDOMComponent(node)`, and that tests the name against `/^[a-z]/`,
// so anything that does not start with an ASCII lowercase letter — a member
// expression among them — is a component.
func isComponent(sourceFile *ast.SourceFile, tagName *ast.Node) bool {
	if tagName == nil {
		return false
	}

	// PropertyAccessExpression (e.g., Foo.Bar) is always a component
	if ast.IsPropertyAccessExpression(tagName) {
		return true
	}

	// For identifiers, check whether the name reads as a DOM tag
	if tagName.Kind == ast.KindIdentifier {
		text := tagName.AsIdentifier().Text
		if len(text) > 0 {
			return !isDOMTagName(text)
		}
	}

	// For other tag name types, get the text from source
	trimmed := utils.TrimNodeTextRange(sourceFile, tagName)
	text := sourceFile.Text()[trimmed.Pos():trimmed.End()]
	if len(text) > 0 {
		// Contains dot means member expression
		if strings.Contains(text, ".") {
			return true
		}
		return !isDOMTagName(text)
	}
	return false
}

// isDOMTagName mirrors eslint-plugin-react's `COMPAT_TAG_REGEX`, which is
// `/^[a-z]/` — ASCII only, so a name starting with `_`, a digit or a character
// with no case of its own is not a DOM tag.
func isDOMTagName(text string) bool {
	return text[0] >= 'a' && text[0] <= 'z'
}
