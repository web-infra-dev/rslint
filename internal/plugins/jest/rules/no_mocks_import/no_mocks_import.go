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

func buildNoManualImportErrorMessage(mockFunction string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noManualImport",
		Description: fmt.Sprintf("Mocks should not be manually imported from a %s directory. Instead use `%s` and import from the original module path", mocksDirName, mockFunction),
	}
}

func isMocksImportPath(path string) bool {
	return slices.Contains(strings.Split(path, "/"), mocksDirName)
}

func isStringNode(node *ast.Node) bool {
	return node.Kind == ast.KindStringLiteral || node.Kind == ast.KindNoSubstitutionTemplateLiteral
}

// NewRule creates a no-mocks-import rule for a test framework.
func NewRule(name string, mockFunction string) rule.Rule {
	return rule.Rule{
		Name: name,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			return rule.RuleListeners{
				ast.KindImportDeclaration: func(node *ast.Node) {
					if isMocksImportPath(node.AsImportDeclaration().ModuleSpecifier.Text()) {
						ctx.ReportNode(node, buildNoManualImportErrorMessage(mockFunction))
					}
				},
				ast.KindCallExpression: func(node *ast.Node) {
					callExpr := node.AsCallExpression().Expression
					if callExpr.Kind != ast.KindIdentifier {
						return
					}

					callee := callExpr.AsIdentifier()
					if callee == nil || callee.Text != "require" {
						return
					}

					arguments := node.Arguments()
					if len(arguments) == 0 {
						return
					}

					firstArg := arguments[0]
					if firstArg != nil && isStringNode(firstArg) && isMocksImportPath(firstArg.Text()) {
						ctx.ReportNode(firstArg, buildNoManualImportErrorMessage(mockFunction))
					}
				},
			}
		},
	}
}

var NoMocksImportRule = NewRule("jest/no-mocks-import", "jest.mock")
