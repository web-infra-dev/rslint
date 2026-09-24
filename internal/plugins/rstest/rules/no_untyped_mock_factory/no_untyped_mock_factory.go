package no_untyped_mock_factory

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_untyped_mock_factory"
)

var NoUntypedMockFactoryRule = shared.NewRule(shared.Config{
	Name:   "rstest/no-untyped-mock-factory",
	Unwrap: utils.SkipAssertionsAndParens,
	Candidates: func(ctx rule.RuleContext) func(*ast.Node) bool {
		factories := map[*ast.Symbol]bool{}
		written := map[*ast.Symbol]bool{}
		return func(node *ast.Node) bool {
			argument, member := rstestUtils.ParseModuleMockFactory(node)
			if argument == nil {
				return false
			}
			path := utils.SkipAssertionsAndParens(node.AsCallExpression().Arguments.Nodes[0])
			// The Promise<T> overload already infers the module shape. The
			// CommonJS APIs accept only strings, so have no such exemption.
			if (member == "mock" || member == "doMock") && path != nil && ast.IsCallExpression(path) && path.AsCallExpression().Expression.Kind == ast.KindImportKeyword {
				return false
			}
			factory := utils.SkipAssertionsAndParens(argument)
			if factory == nil {
				return false
			}
			if ast.IsFunctionExpressionOrArrowFunction(factory) {
				return true
			}
			// The second argument can also be options. Resolve only stable,
			// directly declared functions without types; never guess that an
			// unknown identifier is a factory. Cache repeated symbol checks.
			if factory.Kind == ast.KindIdentifier && ctx.Refs != nil {
				if symbol := ctx.Refs.Resolve(factory); symbol != nil {
					known, cached := factories[symbol]
					if !cached {
						for _, reference := range ctx.Refs.References(symbol) {
							if utils.IsWriteReference(reference) {
								written[symbol] = true
								break
							}
						}
						known = !written[symbol] && declaredFactory(ctx, symbol)
						factories[symbol] = known
					}
					if written[symbol] {
						return false
					}
					if known {
						return true
					}
				}
			}
			return ctx.TypeChecker != nil && len(utils.GetCallSignatures(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(factory))) > 0
		}
	},
})

func declaredFactory(ctx rule.RuleContext, symbol *ast.Symbol) bool {
	if len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if ast.GetSourceFileOfNode(declaration) != ctx.SourceFile {
		return false
	}
	if declaration.Kind == ast.KindVariableDeclaration {
		initializer := utils.SkipAssertionsAndParens(declaration.Initializer())
		return initializer != nil && ast.IsVarConst(declaration) && ast.IsFunctionExpressionOrArrowFunction(initializer)
	}
	return declaration.Kind == ast.KindFunctionDeclaration
}
