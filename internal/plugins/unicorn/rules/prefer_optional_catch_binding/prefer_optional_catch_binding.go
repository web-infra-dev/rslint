// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
package prefer_optional_catch_binding

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	"github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
)

var PreferOptionalCatchBindingRule = rule.Rule{
	Name: "unicorn/prefer-optional-catch-binding", Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{ast.KindCatchClause: func(node *ast.Node) {
			clause := node.AsCatchClause()
			declaration := clause.VariableDeclaration
			if declaration == nil {
				return
			}
			binding := declaration.Name()
			used := false
			utils.CollectBindingNames(binding, func(identifier *ast.Node, _ string) {
				owner := identifier.Parent
				if owner.Initializer() != nil || len(ctx.Refs.References(owner.Symbol())) > 0 {
					used = true
				}
			})
			if used || (ast.IsIdentifier(binding) && hasInitializedRedeclaration(ctx, node, binding.Text())) {
				return
			}
			message := rule.RuleMessage{Id: "without-name", Description: "Remove unused catch binding."}
			if ast.IsIdentifier(binding) {
				message = rule.RuleMessage{Id: "with-name", Description: "Remove unused catch binding `" + binding.Text() + "`.", Data: map[string]string{"name": binding.Text()}}
			}
			bindingRange := utils.GetESTreeBindingIdentifierRange(ctx.SourceFile, binding)
			ctx.ReportRangeWithDeferredFixes(bindingRange, message, func() []rule.RuleFix {
				if utils.HasCommentInsideNode(ctx.SourceFile, declaration) {
					return nil
				}
				open, hasOpen := utils.TokenBeforePosition(ctx.SourceFile, bindingRange.Pos())
				closing, hasClose := utils.TokenAtOrAfter(ctx.SourceFile, bindingRange.End())
				if !hasOpen || !hasClose || open.Kind != ast.KindOpenParenToken || closing.Kind != ast.KindCloseParenToken {
					return nil
				}
				bodyStart := utils.TrimNodeTextRange(ctx.SourceFile, clause.Block).Pos()
				end := ecmascript.SkipLeadingWhitespace(ctx.SourceFile.Text(), closing.End, bodyStart)
				return []rule.RuleFix{rule.RuleFixRemoveRange(open.Range()), rule.RuleFixRemoveRange(bindingRange), rule.RuleFixRemoveRange(core.NewTextRange(closing.Start, end))}
			})
		}}
	},
}

// A var initializer inside a catch can write its parameter even though the
// var declaration itself is hoisted. Binder reference lists exclude that write.
func hasInitializedRedeclaration(ctx rule.RuleContext, clause *ast.Node, name string) bool {
	manager := scopeanalysis.Declarations(ctx)
	for _, lexicalScope := range manager.Scopes {
		for _, variable := range lexicalScope.Declarations(name) {
			if variable.Kind != scope.DefVariable || variable.ID.Pos() < clause.Pos() || variable.ID.End() > clause.End() {
				continue
			}
			declaration := variable.DefNode
			if ast.IsBindingElement(declaration) {
				declaration = utils.EnclosingVariableDeclarationOfBindingElement(declaration)
			}
			if declaration == nil || !ast.IsVariableDeclaration(declaration) {
				continue
			}
			loop := declaration.Parent.Parent
			if declaration.Initializer() == nil && loop.Kind != ast.KindForInStatement && loop.Kind != ast.KindForOfStatement {
				continue
			}
			for current := manager.Acquire(variable.ID); current != nil; current = current.Parent {
				definitions := current.Declarations(name)
				if len(definitions) == 0 {
					continue
				}
				if definitions[0].Kind == scope.DefCatch && definitions[0].Parent == clause {
					return true
				}
				break
			}
		}
	}
	return false
}
