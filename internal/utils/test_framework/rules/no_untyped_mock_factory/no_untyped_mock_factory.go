package no_untyped_mock_factory

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Config leaves API provenance and factory/options discrimination to the
// framework adapter. The annotation contract and edits are framework-neutral.
type Config struct {
	Name       string
	Candidates func(rule.RuleContext) func(*ast.Node) bool
	Unwrap     func(*ast.Node) *ast.Node
}

func NewRule(config Config) rule.Rule {
	if config.Unwrap == nil {
		config.Unwrap = ast.SkipParentheses
	}
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
			candidate := config.Candidates(ctx)
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					call := node.AsCallExpression()
					if call.Arguments == nil || len(call.Arguments.Nodes) != 2 ||
						(call.TypeArguments != nil && len(call.TypeArguments.Nodes) > 0) {
						return
					}
					factory := config.Unwrap(call.Arguments.Nodes[1])
					if factory != nil && (factory.Kind == ast.KindArrowFunction || factory.Kind == ast.KindFunctionExpression) && factory.Type() != nil {
						return
					}
					if !candidate(node) {
						return
					}
					path := config.Unwrap(call.Arguments.Nodes[0])
					moduleName := "./module-name"
					if path != nil && path.Kind == ast.KindStringLiteral {
						moduleName = utils.TrimmedNodeText(ctx.SourceFile, path)
					}
					message := rule.RuleMessage{
						Id:          "addTypeParameterToModuleMock",
						Description: "Add a type parameter to the mock factory such as `typeof import(" + moduleName + ")`",
						Data:        map[string]string{"moduleName": moduleName},
					}
					ctx.ReportNodeWithDeferredFixes(node, message, func() []rule.RuleFix {
						if path == nil || path.Kind != ast.KindStringLiteral {
							return nil
						}
						// An optional call places its type arguments AFTER `?.`.
						// Keep the actual callee, including parentheses and trivia.
						anchor := call.Expression
						if call.QuestionDotToken != nil {
							anchor = call.QuestionDotToken
						}
						return []rule.RuleFix{rule.RuleFixInsertAfter(anchor, "<typeof import("+moduleName+")>")}
					})
				},
			}
		},
	}
}
