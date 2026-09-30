package consistent_assert

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "consistent-assert/error"

func normalizeAssertModule(moduleName string) (string, bool) {
	moduleName = strings.TrimPrefix(moduleName, "node:")
	return moduleName, moduleName == "assert" || moduleName == "assert/strict"
}

func assertImportBinding(specifier *ast.Node, moduleName string) *ast.Node {
	if specifier.IsTypeOnly() {
		return nil
	}

	importSpecifier := specifier.AsImportSpecifier()
	imported := importSpecifier.PropertyName
	if imported == nil {
		imported = importSpecifier.Name()
	}
	if imported.Kind != ast.KindIdentifier {
		return nil
	}

	name := imported.AsIdentifier().Text
	if name == "default" || (moduleName == "assert" && name == "strict") {
		return importSpecifier.Name()
	}
	return nil
}

func reportAssertCalls(ctx rule.RuleContext, binding *ast.Node) {
	symbol := utils.BindingNameSymbol(binding)
	for _, reference := range ctx.Refs.References(symbol) {
		parent := utils.ESTreeParent(reference)
		if parent == nil || parent.Kind != ast.KindCallExpression {
			continue
		}
		call := parent.AsCallExpression()
		if utils.ESTreeCallCallee(call.Expression) != reference {
			continue
		}

		name := reference.AsIdentifier().Text
		ctx.ReportNodeWithDeferredFixes(reference, rule.RuleMessage{
			Id:          messageID,
			Description: "Prefer `" + name + ".ok(…)` over `" + name + "(…)`.",
			Data:        map[string]string{"name": name},
		}, func() []rule.RuleFix {
			return []rule.RuleFix{rule.RuleFixReplaceRange(
				core.NewTextRange(reference.End(), reference.End()),
				".ok",
			)}
		})
	}
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/rules/consistent-assert.js
var ConsistentAssertRule = rule.Rule{
	Name:   "unicorn/consistent-assert",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				declaration := node.AsImportDeclaration()
				if declaration.ImportClause == nil || declaration.ImportClause.IsTypeOnly() {
					return
				}

				moduleName, ok := normalizeAssertModule(declaration.ModuleSpecifier.Text())
				if !ok {
					return
				}

				clause := declaration.ImportClause.AsImportClause()
				if defaultImport := clause.Name(); defaultImport != nil {
					reportAssertCalls(ctx, defaultImport)
				}

				if clause.NamedBindings == nil || clause.NamedBindings.Kind != ast.KindNamedImports {
					return
				}
				for _, specifier := range clause.NamedBindings.AsNamedImports().Elements.Nodes {
					if binding := assertImportBinding(specifier, moduleName); binding != nil {
						reportAssertCalls(ctx, binding)
					}
				}
			},
		}
	},
}
