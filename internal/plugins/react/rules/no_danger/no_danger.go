package no_danger

import (
	"path"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var dangerousProps = map[string]bool{
	"dangerouslySetInnerHTML": true,
}

var NoDangerRule = rule.Rule{
	Name: "react/no-danger",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		var customComponentNames []string
		if optsMap := utils.GetOptionsMap(options); optsMap != nil {
			if arr, ok := optsMap["customComponentNames"].([]interface{}); ok {
				for _, item := range arr {
					if name, ok := item.(string); ok {
						customComponentNames = append(customComponentNames, name)
					}
				}
			}
		}

		return rule.RuleListeners{
			ast.KindJsxAttribute: func(node *ast.Node) {
				attrName := reactutil.GetJsxPropName(node)
				if !dangerousProps[attrName] {
					return
				}

				parent := reactutil.GetJsxParentElement(node)
				if parent == nil {
					return
				}

				matched := reactutil.IsDOMComponent(parent)
				if !matched && len(customComponentNames) > 0 {
					functionName := getTagNameText(ctx.SourceFile, reactutil.GetJsxTagName(parent))
					for _, pattern := range customComponentNames {
						// path.Match mirrors minimatch for the simple patterns this rule
						// accepts in practice (`*`, `Foo*`, `*Foo`, exact names). Component
						// names never contain `/`, so the separator restriction is moot.
						if ok, err := path.Match(pattern, functionName); ok && err == nil {
							matched = true
							break
						}
					}
				}
				if !matched {
					return
				}

				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "dangerousProp",
					Description: "Dangerous property '" + attrName + "' found",
				})
			},
		}
	},
}

// getTagNameText returns the full, structurally-correct tag-name string to
// match against `customComponentNames`.
//
// NOTE: Unlike eslint-plugin-react/no-danger, which only flattens one level
// of member access (`${nodeName.object.name}.${nodeName.property.name}` —
// producing `"undefined.C"` for `<A.B.C>`) and passes an Identifier AST node
// (not a string) into minimatch for `<ns:name>` — which throws `f.split is
// not a function` at runtime — rslint reads the raw source text. This gives
// the intuitive names (`"A.B.C"`, `"ns:name"`) so patterns like `["A.B.C"]`,
// `["A.*"]`, or `["svg:path"]` behave as users would expect. Divergence
// documented in no_danger.md.
func getTagNameText(sourceFile *ast.SourceFile, tagName *ast.Node) string {
	if tagName == nil {
		return ""
	}
	if tagName.Kind == ast.KindIdentifier {
		return tagName.AsIdentifier().Text
	}
	trimmed := utils.TrimNodeTextRange(sourceFile, tagName)
	return strings.TrimSpace(sourceFile.Text()[trimmed.Pos():trimmed.End()])
}
