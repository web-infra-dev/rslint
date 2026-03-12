package jsx_filename_extension

import (
	"fmt"
	"path"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// JsxFilenameExtensionRule restricts file extensions that may contain JSX.
var JsxFilenameExtensionRule = rule.Rule{
	Name: "react/jsx-filename-extension",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Default options
		extensions := []string{".jsx"}
		allow := "always"

		// Parse options
		optsMap := utils.GetOptionsMap(options)
		ignoreFilesWithoutCode := false

		if optsMap != nil {
			if exts, ok := optsMap["extensions"]; ok {
				if extArr, ok := exts.([]interface{}); ok {
					extensions = nil
					for _, e := range extArr {
						if s, ok := e.(string); ok {
							extensions = append(extensions, s)
						}
					}
				}
			}
			if a, ok := optsMap["allow"].(string); ok {
				allow = a
			}
			if v, ok := optsMap["ignoreFilesWithoutCode"].(bool); ok {
				ignoreFilesWithoutCode = v
			}
		}

		isExtensionAllowed := func(ext string) bool {
			for _, e := range extensions {
				if e == ext {
					return true
				}
			}
			return false
		}

		fileName := ctx.SourceFile.FileName()

		// Skip virtual files like <text>
		if strings.HasPrefix(fileName, "<") {
			return rule.RuleListeners{}
		}

		ext := path.Ext(fileName)

		if allow == "as-needed" && isExtensionAllowed(ext) {
			// Need to check if file contains JSX. Walk AST eagerly.
			hasJSX := false
			var walker ast.Visitor
			walker = func(node *ast.Node) bool {
				if hasJSX {
					return true
				}
				switch node.Kind {
				case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
					hasJSX = true
					return true
				}
				node.ForEachChild(walker)
				return hasJSX
			}
			ctx.SourceFile.Node.ForEachChild(walker)

			if !hasJSX {
				if ignoreFilesWithoutCode && (ctx.SourceFile.Statements == nil || len(ctx.SourceFile.Statements.Nodes) == 0) {
					return rule.RuleListeners{}
				}
				ctx.ReportRange(core.NewTextRange(0, 0), rule.RuleMessage{
					Id:          "extensionOnlyForJSX",
					Description: fmt.Sprintf("Only files containing JSX may use the extension '%s'", ext),
				})
			}
			return rule.RuleListeners{}
		}

		if !isExtensionAllowed(ext) {
			// Listen for first JSX node and report
			reported := false
			reportJSX := func(node *ast.Node) {
				if !reported {
					reported = true
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "noJSXWithExtension",
						Description: fmt.Sprintf("JSX not allowed in files with extension '%s'", ext),
					})
				}
			}
			return rule.RuleListeners{
				ast.KindJsxElement:            reportJSX,
				ast.KindJsxSelfClosingElement: reportJSX,
				ast.KindJsxFragment:           reportJSX,
			}
		}

		return rule.RuleListeners{}
	},
}
