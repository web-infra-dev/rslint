package no_anonymous_default_export

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_anonymous_default_export.schema.json
var schemaJSON []byte

// See: https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/no-anonymous-default-export.js
var NoAnonymousDefaultExportRule = rule.Rule{
	Name:   "import/no-anonymous-default-export",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var config map[string]any
		if len(options) > 0 {
			config, _ = options[0].(map[string]any)
		}

		checkExport := func(node *ast.Node) {
			declaration := node
			if node.Kind == ast.KindExportAssignment {
				export := node.AsExportAssignment()
				if export.IsExportEquals {
					return
				}
				declaration = utils.ESTreeRuntimeExpression(export.Expression)
			} else {
				if node.ModifierFlags()&ast.ModifierFlagsExportDefault != ast.ModifierFlagsExportDefault || node.Name() != nil {
					return
				}
			}

			var option, message string
			switch declaration.Kind {
			case ast.KindArrayLiteralExpression:
				option, message = "allowArray", "Assign array to a variable before exporting as module default"
			case ast.KindArrowFunction:
				option, message = "allowArrowFunction", "Assign arrow function to a variable before exporting as module default"
			case ast.KindCallExpression:
				// Only plain import() is excluded. ast.IsImportCall also accepts
				// import.defer(), which upstream treats as an ordinary call.
				if ast.IsOptionalChain(declaration) || declaration.AsCallExpression().Expression.Kind == ast.KindImportKeyword {
					return
				}
				option, message = "allowCallExpression", "Assign call result to a variable before exporting as module default"
			case ast.KindClassDeclaration:
				option, message = "allowAnonymousClass", "Unexpected default export of anonymous class"
			case ast.KindFunctionDeclaration:
				// Bodyless TypeScript declarations are TSDeclareFunction in ESTree.
				if declaration.AsFunctionDeclaration().Body == nil {
					return
				}
				option, message = "allowAnonymousFunction", "Unexpected default export of anonymous function"
			case ast.KindObjectLiteralExpression:
				option, message = "allowObject", "Assign object to a variable before exporting as module default"
			case ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression:
				option, message = "allowLiteral", "Assign literal to a variable before exporting as module default"
			case ast.KindNewExpression:
				option, message = "allowNew", "Assign instance to a variable before exporting as module default"
			default:
				if !utils.IsESTreeLiteralKind(declaration.Kind) {
					return
				}
				option, message = "allowLiteral", "Assign literal to a variable before exporting as module default"
			}

			allowed := option == "allowCallExpression"
			if value, ok := config[option].(bool); ok {
				allowed = value
			}
			if !allowed {
				reportNode := node
				if node.Kind != ast.KindExportAssignment {
					// tsgo includes leading decorators in the declaration range.
					for _, modifier := range node.ModifierNodes() {
						if modifier.Kind == ast.KindExportKeyword {
							reportNode = modifier
							break
						}
					}
				}
				start := scanner.GetTokenPosOfNode(reportNode, ctx.SourceFile, false)
				ctx.ReportRange(core.NewTextRange(start, node.End()), rule.RuleMessage{Description: message})
			}
		}

		return rule.RuleListeners{
			ast.KindExportAssignment:    checkExport,
			ast.KindFunctionDeclaration: checkExport,
			ast.KindClassDeclaration:    checkExport,
		}
	},
}
