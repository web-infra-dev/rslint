package no_mocks_import

import (
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var mocksDirName = "__mocks__"

// Message Builder

func buildNoManualImportErrorMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noManualImport",
		Description: fmt.Sprintf("Mocks should not be manually imported from a %s directory. Instead use `jest.mock` and import from the original module path", mocksDirName),
	}
}

func isMocksImportPath(path string) bool {
	return slices.Contains(strings.Split(path, "/"), mocksDirName)
}

var NoMocksImportRule = rule.Rule{
	Name: "jest/no-mocks-import",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				if isMocksImportPath(node.AsImportDeclaration().ModuleSpecifier.Text()) {
					ctx.ReportNode(node, buildNoManualImportErrorMessage())
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				arguments := node.Arguments()

				if len(arguments) == 0 {
					return
				}

				firstArg := arguments[0]
				if firstArg != nil && ast.IsStringLiteral(firstArg) && isMocksImportPath(firstArg.Text()) {
					ctx.ReportNode(firstArg, buildNoManualImportErrorMessage())
				}
			},
		}
	},
}
