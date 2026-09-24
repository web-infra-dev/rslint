package group_exports

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/group-exports.js
var GroupExportsRule = rule.Rule{
	Name:   "import/group-exports",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		type group struct {
			source    string
			hasSource bool
			typeOnly  bool
			commonJS  bool
		}
		type entry struct {
			node  *ast.Node
			group group
		}
		counts := make(map[group]int)
		var exports []entry
		collect := func(node *ast.Node, key group) {
			counts[key]++
			exports = append(exports, entry{node, key})
		}
		collectDeclaration := func(node *ast.Node) {
			if !ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) || ast.HasSyntacticModifier(node, ast.ModifierFlagsDefault) {
				return
			}
			// Import-equals exports stay in the value group, even with type or declare.
			typeOnly := node.Kind == ast.KindInterfaceDeclaration || node.Kind == ast.KindTypeAliasDeclaration ||
				node.Kind != ast.KindImportEqualsDeclaration && ast.HasSyntacticModifier(node, ast.ModifierFlagsAmbient)
			collect(node, group{typeOnly: typeOnly})
		}

		return rule.RuleListeners{
			ast.KindVariableStatement:       collectDeclaration,
			ast.KindFunctionDeclaration:     collectDeclaration,
			ast.KindClassDeclaration:        collectDeclaration,
			ast.KindInterfaceDeclaration:    collectDeclaration,
			ast.KindTypeAliasDeclaration:    collectDeclaration,
			ast.KindEnumDeclaration:         collectDeclaration,
			ast.KindImportEqualsDeclaration: collectDeclaration,
			ast.KindModuleDeclaration: func(node *ast.Node) {
				// A dotted namespace is one ESTree declaration, not one per segment.
				if node.Parent.Kind != ast.KindModuleDeclaration {
					collectDeclaration(node)
				}
			},
			ast.KindExportDeclaration: func(node *ast.Node) {
				declaration := node.AsExportDeclaration()
				if declaration.ExportClause == nil || declaration.ExportClause.Kind != ast.KindNamedExports {
					return
				}
				key := group{typeOnly: declaration.IsTypeOnly}
				if declaration.ModuleSpecifier != nil {
					key.hasSource = true
					key.source = declaration.ModuleSpecifier.Text()
					// Upstream stores sources in an ordinary object; __proto__ is
					// never an enumerable own property and is not reported.
					if key.source == "__proto__" {
						return
					}
				}
				collect(node, key)
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				if ast.IsAssignmentExpression(node, false) && isCommonJSExport(node.AsBinaryExpression().Left) &&
					!utils.IsDefaultValueInDestructuringAssignment(node) {
					collect(node, group{commonJS: true})
				}
			},
			// The walker visits the file's children, ending with its EOF token.
			rule.ListenerOnExit(ast.KindEndOfFile): func(_ *ast.Node) {
				for _, export := range exports {
					if counts[export.group] < 2 {
						continue
					}
					message := "Multiple named export declarations; consolidate all named exports into a single export declaration"
					if export.group.commonJS {
						message = "Multiple CommonJS exports; consolidate all exports into a single assignment to `module.exports`"
					}
					start := export.node
					// ESTree's export wrapper starts at `export`, after any decorators.
					if modifiers := start.Modifiers(); modifiers != nil {
						for _, modifier := range modifiers.Nodes {
							if modifier.Kind == ast.KindExportKeyword {
								start = modifier
								break
							}
						}
					}
					location := utils.TrimNodeTextRange(ctx.SourceFile, start).WithEnd(export.node.End())
					// Upstream reports literal messages without message IDs or edits.
					ctx.ReportRange(location, rule.RuleMessage{Description: message})
				}
			},
		}
	},
}

func isCommonJSExport(node *ast.Node) bool {
	// Collect property names from right to left, matching upstream's .name
	// checks even for computed identifiers and partial accessor chains.
	chain := make([]string, 0, 4)
	for node = utils.ESTreeRuntimeExpression(node); node != nil && !ast.IsOptionalChain(node); {
		object, property := utils.MemberExpressionParts(node)
		if object == nil {
			break
		}
		// Only two or three names can match; one final root identifier may follow.
		if len(chain) == 3 {
			return false
		}
		name := ""
		property = utils.ESTreeRuntimeExpression(property)
		switch property.Kind {
		case ast.KindIdentifier:
			name = property.Text()
		case ast.KindPrivateIdentifier:
			name = property.Text()[1:]
		}
		chain = append(chain, name)
		object = utils.ESTreeRuntimeExpression(object)
		if object.Kind == ast.KindIdentifier {
			chain = append(chain, object.Text())
			break
		}
		node = object
	}
	length := len(chain)
	return (length == 2 || length == 3) && chain[length-1] == "module" && chain[length-2] == "exports" ||
		length == 2 && chain[1] == "exports"
}
