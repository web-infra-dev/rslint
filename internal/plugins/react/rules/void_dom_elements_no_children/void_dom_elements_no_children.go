package void_dom_elements_no_children

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// voidElements is the set of HTML void elements that cannot have children.
var voidElements = map[string]bool{
	"area":     true,
	"base":     true,
	"br":       true,
	"col":      true,
	"embed":    true,
	"hr":       true,
	"img":      true,
	"input":    true,
	"keygen":   true,
	"link":     true,
	"menuitem": true,
	"meta":     true,
	"param":    true,
	"source":   true,
	"track":    true,
	"wbr":      true,
}

var VoidDomElementsNoChildrenRule = rule.Rule{
	Name: "react/void-dom-elements-no-children",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if !reactutil.IsCreateElementCall(call.Expression) {
					return
				}
				args := call.Arguments
				if args == nil || len(args.Nodes) < 1 {
					return
				}
				// First arg must be a string literal matching a void element
				firstArg := args.Nodes[0]
				if firstArg.Kind != ast.KindStringLiteral {
					return
				}
				elementName := firstArg.AsStringLiteral().Text
				if !voidElements[elementName] {
					return
				}
				// Check for third argument (children) first, before checking second arg type
				if len(args.Nodes) >= 3 {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "noChildrenInVoidEl",
						Description: "Void DOM element <" + elementName + " /> cannot receive children.",
					})
					return
				}
				// Second arg must be an ObjectExpression to check props
				if len(args.Nodes) < 2 || args.Nodes[1].Kind != ast.KindObjectLiteralExpression {
					return
				}
				// Check second argument for children/dangerouslySetInnerHTML props
				{
					obj := args.Nodes[1].AsObjectLiteralExpression()
					if obj.Properties != nil {
						for _, prop := range obj.Properties.Nodes {
							if prop.Kind == ast.KindPropertyAssignment {
								nameNode := prop.AsPropertyAssignment().Name()
								if nameNode != nil && nameNode.Kind == ast.KindIdentifier {
									propName := nameNode.AsIdentifier().Text
									if propName == "children" || propName == "dangerouslySetInnerHTML" {
										ctx.ReportNode(node, rule.RuleMessage{
											Id:          "noChildrenInVoidEl",
											Description: "Void DOM element <" + elementName + " /> cannot receive children.",
										})
										return
									}
								}
							}
						}
					}
				}
			},
			ast.KindJsxElement: func(node *ast.Node) {
				jsxElement := node.AsJsxElement()
				openingElement := jsxElement.OpeningElement.AsJsxOpeningElement()
				tagName := getTagNameText(ctx.SourceFile, openingElement.TagName)

				if !voidElements[tagName] {
					return
				}

				hasChildren := jsxElement.Children != nil && len(jsxElement.Children.Nodes) > 0
				hasChildrenAttr, hasDangerousAttr := hasChildrenOrDangerousJsxAttr(openingElement.Attributes)

				if hasChildren || hasChildrenAttr || hasDangerousAttr {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "noChildrenInVoidEl",
						Description: "Void DOM element <" + tagName + " /> cannot receive children.",
					})
				}
			},
			ast.KindJsxSelfClosingElement: func(node *ast.Node) {
				selfClosing := node.AsJsxSelfClosingElement()
				tagName := getTagNameText(ctx.SourceFile, selfClosing.TagName)

				if !voidElements[tagName] {
					return
				}

				hasChildrenAttr, hasDangerousAttr := hasChildrenOrDangerousJsxAttr(selfClosing.Attributes)

				if hasChildrenAttr || hasDangerousAttr {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "noChildrenInVoidEl",
						Description: "Void DOM element <" + tagName + " /> cannot receive children.",
					})
				}
			},
		}
	},
}

// hasChildrenOrDangerousJsxAttr checks JSX attributes for "children" or "dangerouslySetInnerHTML".
func hasChildrenOrDangerousJsxAttr(attributes *ast.Node) (hasChildren bool, hasDangerous bool) {
	if attributes == nil {
		return
	}
	attrs := attributes.AsJsxAttributes()
	if attrs.Properties == nil {
		return
	}
	for _, prop := range attrs.Properties.Nodes {
		if prop.Kind == ast.KindJsxAttribute {
			attr := prop.AsJsxAttribute()
			name := attr.Name()
			if name != nil && name.Kind == ast.KindIdentifier {
				attrName := name.AsIdentifier().Text
				if attrName == "children" {
					hasChildren = true
				}
				if attrName == "dangerouslySetInnerHTML" {
					hasDangerous = true
				}
			}
		}
	}
	return
}

// getTagNameText extracts the text of a JSX tag name.
func getTagNameText(sourceFile *ast.SourceFile, tagName *ast.Node) string {
	if tagName == nil {
		return ""
	}
	if tagName.Kind == ast.KindIdentifier {
		return tagName.AsIdentifier().Text
	}
	// For member expressions or other complex tag names, extract from source text
	trimmed := utils.TrimNodeTextRange(sourceFile, tagName)
	text := sourceFile.Text()[trimmed.Pos():trimmed.End()]
	return strings.TrimSpace(text)
}
