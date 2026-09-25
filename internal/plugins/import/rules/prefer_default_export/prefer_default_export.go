package prefer_default_export

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

//go:embed prefer_default_export.schema.json
var schemaJSON []byte

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/prefer-default-export.js
// Babel default re-export proposals are outside the parser's supported syntax.
var PreferDefaultExportRule = rule.Rule{
	Name:   "import/prefer-default-export",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		targetAny := false
		if len(options) > 0 {
			if option, ok := options[0].(map[string]any); ok {
				targetAny = option["target"] == "any"
			}
		}

		var state struct {
			count  int
			exempt bool
			start  *ast.Node
			end    int
		}
		checkDeclaration := func(node *ast.Node) {
			var exportKeyword *ast.Node
			for _, modifier := range node.ModifierNodes() {
				// Dotted namespaces contain synthetic export modifiers.
				if modifier.Kind == ast.KindExportKeyword && !ast.NodeIsMissing(modifier) {
					exportKeyword = modifier
					break
				}
			}
			if exportKeyword == nil {
				return
			}
			if ast.HasSyntacticModifier(node, ast.ModifierFlagsDefault) ||
				node.Kind == ast.KindTypeAliasDeclaration || node.Kind == ast.KindInterfaceDeclaration {
				state.exempt = true
				return
			}
			if node.Kind == ast.KindVariableStatement {
				for _, declaration := range node.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
					state.count += countExportBindings(declaration.Name())
				}
			} else {
				state.count++
			}
			// Match the ESTree export wrapper, excluding preceding decorators.
			state.start = exportKeyword
			state.end = node.End()
		}

		return rule.RuleListeners{
			ast.KindExportSpecifier: func(node *ast.Node) {
				if ast.ModuleExportNameIsDefault(node.Name()) {
					state.exempt = true
				} else {
					state.count++
					state.start = node
					state.end = node.End()
				}
			},
			ast.KindExportDeclaration: func(node *ast.Node) {
				clause := node.AsExportDeclaration().ExportClause
				if clause == nil || clause.Kind == ast.KindNamespaceExport {
					state.exempt = true
				}
			},
			ast.KindExportAssignment: func(node *ast.Node) {
				if !node.AsExportAssignment().IsExportEquals {
					state.exempt = true
				}
			},
			ast.KindVariableStatement:       checkDeclaration,
			ast.KindFunctionDeclaration:     checkDeclaration,
			ast.KindClassDeclaration:        checkDeclaration,
			ast.KindInterfaceDeclaration:    checkDeclaration,
			ast.KindTypeAliasDeclaration:    checkDeclaration,
			ast.KindEnumDeclaration:         checkDeclaration,
			ast.KindModuleDeclaration:       checkDeclaration,
			ast.KindImportEqualsDeclaration: checkDeclaration,
			rule.ListenerOnExit(ast.KindEndOfFile): func(_ *ast.Node) {
				if state.exempt || state.count == 0 || (!targetAny && state.count != 1) {
					return
				}
				message := "Prefer default export on a file with single export."
				if targetAny {
					message = "Prefer default export to be present on every file that has export."
				}
				// Upstream reports the last named export, with no message ID or edits.
				// Trim its range only when reporting; most exports never need one.
				start := scanner.GetTokenPosOfNode(state.start, ctx.SourceFile, false)
				ctx.ReportRange(core.NewTextRange(start, state.end), rule.RuleMessage{Description: message})
			},
		}
	},
}

// Upstream counts array holes, rest elements and defaulted patterns as one.
// CollectBindingNames would skip holes and expand those patterns instead.
func countExportBindings(node *ast.Node) int {
	if node == nil || (node.Kind != ast.KindObjectBindingPattern && node.Kind != ast.KindArrayBindingPattern) {
		return 1
	}
	count := 0
	for _, element := range node.AsBindingPattern().Elements.Nodes {
		if element.Kind == ast.KindBindingElement {
			binding := element.AsBindingElement()
			if binding.Initializer == nil && binding.DotDotDotToken == nil {
				count += countExportBindings(binding.Name())
				continue
			}
		}
		count++
	}
	return count
}
