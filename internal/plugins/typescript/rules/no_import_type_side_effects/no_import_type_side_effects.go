package no_import_type_side_effects

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildUseTopLevelQualifierMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useTopLevelQualifier",
		Description: "TypeScript will only remove the inline type specifiers which will leave behind a side effect import at runtime. Convert this to a top-level type qualifier to properly remove the entire import.",
	}
}

var NoImportTypeSideEffectsRule = rule.CreateRule(rule.Rule{
	Name: "no-import-type-side-effects",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				importDecl := node.AsImportDeclaration()
				if importDecl == nil {
					return
				}
				clauseNode := importDecl.ImportClause
				if clauseNode == nil {
					return
				}
				// Mirror upstream's `ImportDeclaration[importKind!="type"]`
				// selector. tsgo's `Node.IsTypeOnly()` returns true exactly
				// when the ImportClause carries the top-level `type` keyword.
				if clauseNode.IsTypeOnly() {
					return
				}
				clause := clauseNode.AsImportClause()
				// A default-binding (`import T, ...`) is a non-type specifier
				// in ESLint's model. Bail out the same way upstream does.
				if clause.Name() != nil {
					return
				}
				namedBindings := clause.NamedBindings
				if namedBindings == nil || !ast.IsNamedImports(namedBindings) {
					return
				}
				namedImports := namedBindings.AsNamedImports()
				if namedImports.Elements == nil {
					return
				}
				elements := namedImports.Elements.Nodes
				// Mirror upstream's `if (node.specifiers.length === 0) return;`.
				// In tsgo, empty named-imports `import {} from 'mod';` lands
				// here with zero elements.
				if len(elements) == 0 {
					return
				}

				for _, specifierNode := range elements {
					if specifierNode.Kind != ast.KindImportSpecifier || !specifierNode.IsTypeOnly() {
						return
					}
				}

				text := ctx.SourceFile.Text()
				fixes := make([]rule.RuleFix, 0, len(elements)+1)
				for _, specifierNode := range elements {
					spec := specifierNode.AsImportSpecifier()
					typeStart := scanner.SkipTrivia(text, specifierNode.Pos())
					var imported *ast.Node
					if spec.PropertyName != nil {
						imported = spec.PropertyName
					} else {
						imported = spec.Name()
					}
					importedStart := scanner.SkipTrivia(text, imported.Pos())
					fixes = append(fixes, rule.RuleFixRemoveRange(
						core.NewTextRange(typeStart, importedStart),
					))
				}

				importTokenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos())
				insertPos := importTokenRange.End()
				fixes = append(fixes, rule.RuleFixReplaceRange(
					core.NewTextRange(insertPos, insertPos),
					" type",
				))

				ctx.ReportNodeWithFixes(node, buildUseTopLevelQualifierMessage(), fixes...)
			},
		}
	},
})
