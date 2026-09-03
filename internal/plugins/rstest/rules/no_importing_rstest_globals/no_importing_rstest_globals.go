package no_importing_rstest_globals

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

func noImportMessage(name, module string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noImportingRstestGlobals",
		Description: fmt.Sprintf("Do not import `%s` from `%s`; it is available as a global.", name, module),
		Data:        map[string]string{"name": name, "module": module},
	}
}

func noRequireMessage(name, module string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noRequiringRstestGlobals",
		Description: fmt.Sprintf("Do not require `%s` from `%s`; it is available as a global.", name, module),
		Data:        map[string]string{"name": name, "module": module},
	}
}

func localBindingName(element *ast.Node) string {
	if element == nil {
		return ""
	}
	if element.Kind == ast.KindImportSpecifier {
		specifier := element.AsImportSpecifier()
		if specifier != nil && specifier.Name() != nil {
			return specifier.Name().Text()
		}
	}
	if element.Kind == ast.KindBindingElement {
		binding := element.AsBindingElement()
		if binding != nil && binding.Name() != nil && binding.Name().Kind == ast.KindIdentifier {
			return binding.Name().AsIdentifier().Text
		}
	}
	return ""
}

func isInvocationUse(identifier *ast.Node) bool {
	current := identifier
	for parent := current.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			if parent.AsParenthesizedExpression().Expression != current {
				return false
			}
		case ast.KindNonNullExpression:
			if parent.AsNonNullExpression().Expression != current {
				return false
			}
		case ast.KindAsExpression:
			if parent.AsAsExpression().Expression != current {
				return false
			}
		case ast.KindTypeAssertionExpression:
			if parent.AsTypeAssertion().Expression != current {
				return false
			}
		case ast.KindSatisfiesExpression:
			if parent.AsSatisfiesExpression().Expression != current {
				return false
			}
		case ast.KindPropertyAccessExpression:
			if parent.AsPropertyAccessExpression().Expression != current {
				return false
			}
		case ast.KindElementAccessExpression:
			if parent.AsElementAccessExpression().Expression != current {
				return false
			}
		case ast.KindCallExpression:
			return parent.AsCallExpression().Expression == current
		case ast.KindTaggedTemplateExpression:
			return parent.AsTaggedTemplateExpression().Tag == current
		default:
			return false
		}
		current = parent
	}
	return false
}

func exportUsesName(sourceFile *ast.SourceFile, name string) bool {
	used := false
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil || used {
			return
		}
		if node.Kind == ast.KindExportSpecifier {
			specifier := node.AsExportSpecifier()
			local := specifier.Name()
			if specifier.PropertyName != nil {
				local = specifier.PropertyName
			}
			if local != nil && local.Kind == ast.KindIdentifier && local.AsIdentifier().Text == name {
				used = true
				return
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return used
		})
	}
	visit(sourceFile.AsNode())
	return used
}

func bindingCanBeRemoved(ctx rule.RuleContext, element *ast.Node, importedName string) bool {
	local := localBindingName(element)
	if local == "" || local != importedName || ctx.Refs == nil || exportUsesName(ctx.SourceFile, local) {
		return false
	}
	symbol := element.Symbol()
	if symbol == nil {
		return false
	}
	for _, reference := range ctx.Refs.References(symbol) {
		if !isInvocationUse(reference) {
			return false
		}
	}
	return true
}

// allBindingsRemovable reports whether every binding of the declaration is an
// rstest global that is independently safe to drop. Only then may the whole
// declaration go: a sibling binding that is aliased, or referenced as a value
// rather than invoked, still needs its import.
func allBindingsRemovable(ctx rule.RuleContext, elements []*ast.Node, importedName func(*ast.Node) string) bool {
	for _, element := range elements {
		name := importedName(element)
		if !rstestUtils.IsRstestGlobal(name) || !bindingCanBeRemoved(ctx, element, name) {
			return false
		}
	}
	return true
}

func importFix(ctx rule.RuleContext, statement *ast.Node, declaration *ast.ImportDeclaration, elements []*ast.Node, index int, name string) []rule.RuleFix {
	if !bindingCanBeRemoved(ctx, elements[index], name) {
		return nil
	}
	clause := declaration.ImportClause.AsImportClause()
	if clause.Name() == nil && allBindingsRemovable(ctx, elements, rstestUtils.ImportedSpecifierName) {
		return []rule.RuleFix{rule.RuleFixRemove(ctx.SourceFile, statement)}
	}
	if len(elements) > 1 {
		return []rule.RuleFix{rule.RuleFixRemoveRange(rstestUtils.SpecifierRemovalRange(ctx.SourceFile, elements, index))}
	}
	if clause.Name() != nil && clause.NamedBindings != nil {
		start := internalUtils.TrimNodeTextRange(ctx.SourceFile, clause.Name()).End()
		end := internalUtils.TrimNodeTextRange(ctx.SourceFile, clause.NamedBindings).End()
		return []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(start, end))}
	}
	return nil
}

func enclosingVariableStatement(node *ast.Node) (*ast.Node, *ast.VariableDeclarationList) {
	for current := node; current != nil; current = current.Parent {
		if current.Kind == ast.KindVariableStatement {
			statement := current.AsVariableStatement()
			if statement != nil && statement.DeclarationList != nil {
				return current, statement.DeclarationList.AsVariableDeclarationList()
			}
			return current, nil
		}
	}
	return nil, nil
}

func requireFix(ctx rule.RuleContext, declarationNode *ast.Node, elements []*ast.Node, index int, name string) []rule.RuleFix {
	for _, element := range elements {
		binding := element.AsBindingElement()
		if binding == nil || binding.DotDotDotToken != nil || binding.Initializer != nil {
			return nil
		}
	}
	if !bindingCanBeRemoved(ctx, elements[index], name) {
		return nil
	}
	if allBindingsRemovable(ctx, elements, rstestUtils.RequireBindingImportedName) {
		statement, list := enclosingVariableStatement(declarationNode)
		if statement != nil && list != nil && list.Declarations != nil && len(list.Declarations.Nodes) == 1 {
			return []rule.RuleFix{rule.RuleFixRemove(ctx.SourceFile, statement)}
		}
		return nil
	}
	if len(elements) > 1 {
		return []rule.RuleFix{rule.RuleFixRemoveRange(rstestUtils.SpecifierRemovalRange(ctx.SourceFile, elements, index))}
	}
	return nil
}

var NoImportingRstestGlobalsRule = rule.Rule{
	Name:   "rstest/no-importing-rstest-globals",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				declaration := node.AsImportDeclaration()
				if declaration == nil || declaration.ModuleSpecifier == nil {
					return
				}
				module := declaration.ModuleSpecifier.Text()
				if !rstestUtils.IsRstestCoreImportModule(module) {
					return
				}
				elements := rstestUtils.NamedImportElements(declaration)
				for index, element := range elements {
					name := rstestUtils.ImportedSpecifierName(element)
					if !rstestUtils.IsRstestGlobal(name) {
						continue
					}
					ctx.ReportNodeWithDeferredFixes(element, noImportMessage(name, module), func() []rule.RuleFix {
						return importFix(ctx, node, declaration, elements, index, name)
					})
				}
			},
			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				if declaration == nil || declaration.Name() == nil || declaration.Name().Kind != ast.KindObjectBindingPattern {
					return
				}
				module, ok := rstestUtils.RstestCoreModuleFromRequireCall(declaration.Initializer)
				if !ok {
					return
				}
				pattern := declaration.Name().AsBindingPattern()
				if pattern == nil || pattern.Elements == nil {
					return
				}
				elements := pattern.Elements.Nodes
				for index, element := range elements {
					name := rstestUtils.RequireBindingImportedName(element)
					if !rstestUtils.IsRstestGlobal(name) {
						continue
					}
					ctx.ReportNodeWithDeferredFixes(element, noRequireMessage(name, module), func() []rule.RuleFix {
						return requireFix(ctx, node, elements, index, name)
					})
				}
			},
		}
	},
}
