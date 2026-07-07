package no_default_export

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func preferNamedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferNamed",
		Description: "Prefer named exports.",
	}
}

func noAliasDefaultMessage(local string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "noAliasDefault",
		Description: fmt.Sprintf(
			"Do not alias `%s` as `default`. Just export `%s` itself instead.",
			local,
			local,
		),
	}
}

// See: https://github.com/import-js/eslint-plugin-import/blob/main/src/rules/no-default-export.js
var NoDefaultExportRule = rule.Rule{
	Name: "import/no-default-export",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		reportDefaultExport := func(node *ast.Node) {
			ctx.ReportRange(defaultKeywordRange(ctx.SourceFile, node), preferNamedMessage())
		}

		checkExportDeclaration := func(node *ast.Node) {
			exportDecl := node.AsExportDeclaration()
			if exportDecl == nil ||
				exportDecl.ExportClause == nil ||
				exportDecl.ExportClause.Kind != ast.KindNamedExports {
				return
			}

			namedExports := exportDecl.ExportClause.AsNamedExports()
			if namedExports == nil || namedExports.Elements == nil {
				return
			}

			reportRange := tokenAfterExportKeywordRange(ctx.SourceFile, node)
			for _, specifierNode := range namedExports.Elements.Nodes {
				specifier := specifierNode.AsExportSpecifier()
				if specifier == nil || !ast.ModuleExportNameIsDefault(specifier.Name()) {
					continue
				}

				local := specifier.PropertyName
				if local == nil {
					local = specifier.Name()
				}
				ctx.ReportRange(reportRange, noAliasDefaultMessage(localNameForMessage(local)))
			}
		}

		return rule.RuleListeners{
			ast.KindExportAssignment: func(node *ast.Node) {
				exportAssignment := node.AsExportAssignment()
				if exportAssignment == nil || exportAssignment.IsExportEquals {
					return
				}
				reportDefaultExport(node)
			},
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				if isDefaultExportedDeclaration(node) {
					reportDefaultExport(node)
				}
			},
			ast.KindClassDeclaration: func(node *ast.Node) {
				if isDefaultExportedDeclaration(node) {
					reportDefaultExport(node)
				}
			},
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				if isDefaultExportedDeclaration(node) {
					reportDefaultExport(node)
				}
			},
			ast.KindEnumDeclaration: func(node *ast.Node) {
				if isDefaultExportedDeclaration(node) {
					reportDefaultExport(node)
				}
			},
			ast.KindExportDeclaration: checkExportDeclaration,
		}
	},
}

func isDefaultExportedDeclaration(node *ast.Node) bool {
	flags := node.ModifierFlags()
	return flags&ast.ModifierFlagsExport != 0 && flags&ast.ModifierFlagsDefault != 0
}

func localNameForMessage(node *ast.Node) string {
	if node != nil && (ast.IsIdentifier(node) || node.Kind == ast.KindDefaultKeyword) {
		return node.Text()
	}
	return "undefined"
}

func defaultKeywordRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	token, ok := tokenAfterExportKeyword(sourceFile, node)
	if ok && token.Kind == ast.KindDefaultKeyword {
		return token.Range()
	}

	return utils.TrimNodeTextRange(sourceFile, node)
}

func tokenAfterExportKeywordRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	token, ok := tokenAfterExportKeyword(sourceFile, node)
	if ok {
		return token.Range()
	}

	return utils.TrimNodeTextRange(sourceFile, node)
}

// tokenAfterExportKeyword returns the first non-trivia token after a leading
// `export` keyword. Upstream reports named default aliases on that token.
func tokenAfterExportKeyword(sourceFile *ast.SourceFile, node *ast.Node) (utils.SourceToken, bool) {
	if sourceFile == nil || node == nil {
		return utils.SourceToken{}, false
	}

	exportToken, ok := utils.TokenAtOrAfter(sourceFile, node.Pos())
	if !ok || exportToken.Start >= node.End() || exportToken.Kind != ast.KindExportKeyword {
		return utils.SourceToken{}, false
	}

	next, ok := utils.TokenAtOrAfter(sourceFile, exportToken.End)
	if !ok || next.Start >= node.End() {
		return utils.SourceToken{}, false
	}

	return next, true
}
