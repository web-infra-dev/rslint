package self_closing_comp

import (
	_ "embed"
	"strings"
	"unicode"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
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
			if strings.TrimSpace(text.Text) == "" && strings.Contains(text.Text, "\n") {
				continue
			}
			return false
		}
		return false
	}
	return true
}

// isComponent returns true if the tag name starts with uppercase or contains dots (member expression).
func isComponent(sourceFile *ast.SourceFile, tagName *ast.Node) bool {
	if tagName == nil {
		return false
	}

	// PropertyAccessExpression (e.g., Foo.Bar) is always a component
	if ast.IsPropertyAccessExpression(tagName) {
		return true
	}

	// For identifiers, check if first character is uppercase
	if tagName.Kind == ast.KindIdentifier {
		text := tagName.AsIdentifier().Text
		if len(text) > 0 {
			firstRune := rune(text[0])
			return unicode.IsUpper(firstRune)
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
		firstRune := rune(text[0])
		return unicode.IsUpper(firstRune)
	}
	return false
}
