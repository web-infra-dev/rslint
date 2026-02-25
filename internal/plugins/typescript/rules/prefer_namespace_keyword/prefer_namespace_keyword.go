package prefer_namespace_keyword

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildUseNamespaceMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useNamespace",
		Description: "Use 'namespace' instead of 'module' to declare custom TypeScript modules.",
	}
}

var PreferNamespaceKeywordRule = rule.CreateRule(rule.Rule{
	Name: "prefer-namespace-keyword",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindModuleDeclaration: func(node *ast.Node) {
				moduleDecl := node.AsModuleDeclaration()
				if moduleDecl == nil {
					return
				}

				// Skip if the module name is a string literal (external module declaration)
				name := moduleDecl.Name()
				if name == nil {
					return
				}
				if name.Kind == ast.KindStringLiteral {
					return
				}

				// Check if the keyword before the module name is "module"
				if moduleDecl.Keyword != ast.KindModuleKeyword {
					return
				}

				// Find the "module" token to replace it with "namespace"
				s := scanner.GetScannerForSourceFile(ctx.SourceFile, node.Pos())
				for s.TokenStart() < name.Pos() {
					if s.Token() == ast.KindModuleKeyword && s.TokenText() == "module" {
						moduleTokenStart := s.TokenStart()
						moduleTokenEnd := s.TokenEnd()
						ctx.ReportNodeWithFixes(node, buildUseNamespaceMessage(),
							rule.RuleFixReplaceRange(
								node.Loc.WithPos(moduleTokenStart).WithEnd(moduleTokenEnd),
								"namespace",
							),
						)
						return
					}
					s.Scan()
				}
			},
		}
	},
})
