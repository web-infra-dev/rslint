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
			container *ast.Node
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
			if !key.commonJS {
				// Unlike upstream, keep each namespace/module body independent.
				// A nil container represents this source file's top-level exports.
				key.container = ast.FindAncestorKind(node.Parent, ast.KindModuleBlock)
			}
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
					// All source names, including __proto__, are ordinary map keys.
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
	node = utils.ESTreeRuntimeExpression(node)
	object, property := utils.MemberExpressionParts(node)
	if object == nil || property == nil || ast.IsOptionalChain(node) || property.Kind == ast.KindPrivateIdentifier {
		return false
	}
	object = utils.ESTreeRuntimeExpression(object)
	// Require a complete public access rooted in an identifier, rather than
	// upstream's property-name suffix matching on arbitrary receivers.
	return ast.IsExportsIdentifier(object) || isModuleExports(node) || isModuleExports(object)
}

func isModuleExports(node *ast.Node) bool {
	if !ast.IsAccessExpression(node) || ast.IsOptionalChain(node) ||
		!ast.IsModuleIdentifier(utils.ESTreeRuntimeExpression(node.Expression())) {
		return false
	}
	// tsgo distinguishes static property names from dynamic and private ones.
	name := ast.GetElementOrPropertyAccessName(node)
	return name != nil && name.Text() == "exports"
}
