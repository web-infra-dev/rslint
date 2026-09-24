package exports_last

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/exports-last.js
var ExportsLastRule = rule.Rule{
	Name:   "import/exports-last",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		// SourceFile has no listener event; inspect its body once here.
		statements := ctx.SourceFile.Statements.Nodes
		lastNonExport := len(statements) - 1
		for lastNonExport >= 0 && exportStart(statements[lastNonExport]) != nil {
			lastNonExport--
		}
		for i := range lastNonExport {
			statement := statements[i]
			if start := exportStart(statement); start != nil {
				pos := scanner.GetTokenPosOfNode(start, ctx.SourceFile, false)
				ctx.ReportRange(core.NewTextRange(pos, statement.End()), rule.RuleMessage{
					Description: "Export statements should appear at the end of the file",
				})
			}
		}
		return nil
	},
}

// exportStart recognizes ESTree's three ES export statement kinds. tsgo stores
// exported declarations as modifiers; starting at that modifier also excludes
// decorators before export. TypeScript's export = and export as namespace are
// separate ESTree kinds and do not count as ES exports.
func exportStart(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindExportDeclaration:
		return node
	case ast.KindExportAssignment:
		if node.AsExportAssignment().IsExportEquals {
			return nil
		}
		return node
	}
	for _, modifier := range node.ModifierNodes() {
		if modifier.Kind == ast.KindExportKeyword {
			return modifier
		}
	}
	return nil
}
