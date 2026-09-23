package consistent_type_specifier_style

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed consistent_type_specifier_style.schema.json
var schemaJSON []byte

func preferInlineMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferInline",
		Description: "Prefer using inline type specifiers instead of a top-level type-only import.",
	}
}

func preferTopLevelMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferTopLevel",
		Description: "Prefer using a top-level type-only import instead of inline type specifiers.",
	}
}

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/src/rules/consistent-type-specifier-style.js
var ConsistentTypeSpecifierStyleRule = rule.Rule{
	Name:   "import/consistent-type-specifier-style",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// Upstream's runtime uses top-level when the option is omitted, even
		// though its schema and documentation advertise an inline default.
		preferInline := len(options) > 0 && options[0] == "prefer-inline"
		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				declaration := node.AsImportDeclaration()
				if declaration.ImportClause == nil {
					return
				}
				clause := declaration.ImportClause.AsImportClause()
				if clause.NamedBindings == nil || clause.NamedBindings.Kind != ast.KindNamedImports {
					return
				}
				specifiers := clause.NamedBindings.AsNamedImports().Elements.Nodes
				if len(specifiers) == 0 {
					return
				}

				if preferInline {
					if clause.IsTypeOnly() {
						ctx.ReportNodeWithDeferredFixes(node, preferInlineMessage(), func() []rule.RuleFix {
							// Type import attributes such as resolution-mode require import type.
							if declaration.Attributes != nil {
								return nil
							}
							return inlineFixes(ctx.SourceFile, clause, specifiers)
						})
					}
					return
				}
				if clause.IsTypeOnly() {
					return
				}

				var typeSpecifiers []*ast.Node
				var lastValueSpecifier *ast.Node
				for _, specifier := range specifiers {
					if specifier.AsImportSpecifier().IsTypeOnly {
						typeSpecifiers = append(typeSpecifiers, specifier)
					} else {
						lastValueSpecifier = specifier
					}
				}
				if len(typeSpecifiers) == 0 {
					return
				}

				// Every specifier diagnostic carries the same combined fix. Build it
				// once, and only when the consumer requests autofixes.
				var fixes []rule.RuleFix
				fixesBuilt := false
				buildFixes := func() []rule.RuleFix {
					if !fixesBuilt {
						fixes = topLevelFixes(ctx.SourceFile, node, typeSpecifiers, lastValueSpecifier)
						fixesBuilt = true
					}
					return fixes
				}
				if lastValueSpecifier == nil && clause.Name() == nil {
					ctx.ReportNodeWithDeferredFixes(node, preferTopLevelMessage(), buildFixes)
					return
				}
				for _, specifier := range typeSpecifiers {
					ctx.ReportNodeWithDeferredFixes(specifier, preferTopLevelMessage(), buildFixes)
				}
			},
		}
	},
}

func inlineFixes(sourceFile *ast.SourceFile, clause *ast.ImportClause, specifiers []*ast.Node) []rule.RuleFix {
	fixes := make([]rule.RuleFix, 0, len(specifiers)+2)
	if token, ok := utils.TokenAtOrAfter(sourceFile, clause.Pos()); ok && token.Kind == ast.KindTypeKeyword {
		fixes = append(fixes, rule.RuleFixRemoveRange(token.Range()))
	}
	if clause.Name() != nil {
		fixes = append(fixes, rule.RuleFixInsertBefore(sourceFile, clause.Name(), "type "))
	}
	for _, specifier := range specifiers {
		fixes = append(fixes, rule.RuleFixInsertBefore(sourceFile, specifier, "type "))
	}
	return fixes
}

func topLevelFixes(sourceFile *ast.SourceFile, node *ast.Node, typeSpecifiers []*ast.Node, lastValueSpecifier *ast.Node) []rule.RuleFix {
	declaration := node.AsImportDeclaration()
	// Dropping attributes can change resolution, and copying value-import
	// attributes onto import type can produce invalid TypeScript.
	if declaration.Attributes != nil {
		return nil
	}
	text := utils.TrimmedNodeText(sourceFile, node)
	mayHaveComments := strings.Contains(text, "//") || strings.Contains(text, "/*")
	clause := declaration.ImportClause.AsImportClause()
	names := make([]string, 0, len(typeSpecifiers))
	for _, node := range typeSpecifiers {
		specifier := node.AsImportSpecifier()
		imported := node.PropertyNameOrName()
		name := imported.Text()
		// Preserve quoted export names rather than changing the imported binding.
		if imported.Kind == ast.KindStringLiteral {
			name = utils.TrimmedNodeText(sourceFile, imported)
		}
		if imported.Kind == ast.KindStringLiteral || name != specifier.Name().Text() {
			name += " as " + specifier.Name().Text()
		}
		names = append(names, name)
	}
	newImport := "import type {" + strings.Join(names, ", ") + "} from " + utils.TrimmedNodeText(sourceFile, declaration.ModuleSpecifier) + ";"
	if lastValueSpecifier == nil && clause.Name() == nil {
		if mayHaveComments && utils.HasCommentInsideNode(sourceFile, node) {
			return nil
		}
		return []rule.RuleFix{rule.RuleFixReplace(sourceFile, node, newImport)}
	}

	capacity := 2
	if lastValueSpecifier != nil {
		capacity += 2 * len(typeSpecifiers)
	}
	fixes := make([]rule.RuleFix, 0, capacity)
	if lastValueSpecifier != nil {
		for _, specifier := range typeSpecifiers {
			if mayHaveComments && utils.HasCommentInsideNode(sourceFile, specifier) {
				return nil
			}
			if comma, ok := utils.TokenAtOrAfter(sourceFile, specifier.End()); ok && comma.Kind == ast.KindCommaToken {
				fixes = append(fixes, rule.RuleFixRemoveRange(comma.Range()))
			}
			fixes = append(fixes, rule.RuleFixRemove(sourceFile, specifier))
		}
		if comma, ok := utils.TokenAtOrAfter(sourceFile, lastValueSpecifier.End()); ok && comma.Kind == ast.KindCommaToken {
			fixes = append(fixes, rule.RuleFixRemoveRange(comma.Range()))
		}
	} else if comma, ok := utils.TokenAtOrAfter(sourceFile, clause.Name().End()); ok {
		if mayHaveComments && (utils.HasCommentsInRange(sourceFile, core.NewTextRange(comma.End, clause.NamedBindings.End())) ||
			utils.HasCommentInsideNode(sourceFile, clause.NamedBindings)) {
			return nil
		}
		fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(comma.Start, clause.NamedBindings.End())))
	}
	return append(fixes, rule.RuleFixInsertAfter(node, "\n"+newImport))
}
