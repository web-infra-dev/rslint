package unambiguous

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/unambiguous.js
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/utils/unambiguous.js
var UnambiguousRule = rule.Rule{
	Name:   "import/unambiguous",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		if ctx.LanguageOptions.EffectiveSourceType() != "module" {
			return nil
		}

		// SourceFile has no listener event. Inspect only its top-level body,
		// matching upstream's Import/Export declarations and TSExportAssignment.
		// IsExternalModule also accepts import.meta and compiler module settings;
		// IsExternalModuleIndicator additionally accepts bare TS import aliases.
		var first, last *ast.Node
		for _, statement := range ctx.SourceFile.Statements.Nodes {
			// JSDoc imports and typedefs are synthetic statements in tsgo.
			if utils.IsJSDocSyntaxNode(statement) {
				continue
			}
			if first == nil {
				first = statement
			}
			last = statement
			switch statement.Kind {
			case ast.KindImportDeclaration, ast.KindExportDeclaration, ast.KindExportAssignment:
				return nil
			}
			// tsgo represents ESTree export declaration wrappers as modifiers.
			if ast.HasSyntacticModifier(statement, ast.ModifierFlagsExport) {
				return nil
			}
		}

		// Match Program locations from @typescript-eslint/parser for TS and
		// Espree for JS: Espree trims both ends of a nonempty Program, but
		// includes all trivia when there are no statements.
		var reportRange core.TextRange
		if ctx.SourceFile.ScriptKind == core.ScriptKindJS || ctx.SourceFile.ScriptKind == core.ScriptKindJSX {
			reportRange = ctx.SourceFile.AsNode().Loc
			if first != nil {
				reportRange = core.NewTextRange(scanner.GetTokenPosOfNode(first, ctx.SourceFile, false), last.End())
			}
		} else {
			program := ctx.SourceFile.AsNode()
			reportRange = core.NewTextRange(scanner.GetTokenPosOfNode(program, ctx.SourceFile, false), program.End())
		}
		// Upstream uses a literal message without a message ID.
		ctx.ReportRange(reportRange, rule.RuleMessage{
			Description: "This module could be parsed as a valid script.",
		})
		return nil
	},
}
