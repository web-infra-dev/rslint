package no_untyped_mock_factory

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_untyped_mock_factory"
)

var NoUntypedMockFactoryRule = shared.NewRule(shared.Config{
	Name:   "rstest/no-untyped-mock-factory",
	Unwrap: utils.SkipAssertionsAndParens,
	CanFixWithoutTypeInfo: func(ctx rule.RuleContext, node *ast.Node) bool {
		utility := rstestUtils.ParseRstestPluginManagedCall(node)
		if utility == nil || ctx.Refs == nil {
			return false
		}
		symbol := ctx.Refs.Resolve(utility.NamespaceNode)
		if symbol == nil {
			return true
		}
		return testFramework.IsNamedESMImportSymbolModules(
			symbol,
			rstestUtils.RstestCoreImportModules,
			[]string{"rs", "rstest"},
		)
	},
	Candidates: func(ctx rule.RuleContext) func(*ast.Node) bool {
		checkedWrites := map[*ast.Symbol]bool{}
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

			hoisted := member == "mock" || member == "mockRequire"
			if factory.Kind == ast.KindIdentifier && ctx.Refs != nil {
				if symbol := ctx.Refs.Resolve(factory); symbol != nil {
					if !checkedWrites[symbol] {
						for _, reference := range ctx.Refs.References(symbol) {
							if utils.IsWriteReference(reference) {
								written[symbol] = true
								break
							}
						}
						checkedWrites[symbol] = true
					}
					if written[symbol] {
						return false
					}
					if result, decided := declaredFactoryAtCall(ctx, symbol, node, member, factory); decided {
						return result
					}
				}
			}

			// A lifted call cannot use an ordinary parameter, import, member,
			// or other typed expression: Rstest evaluates it before those
			// bindings are initialized. Only the explicit hoist-safe cases in
			// declaredFactoryAtCall may reach a lifted mock.
			if hoisted {
				return false
			}
			return isCallable(ctx, factory)
		}
	},
})

func declaredFactoryAtCall(
	ctx rule.RuleContext,
	symbol *ast.Symbol,
	call *ast.Node,
	member string,
	reference *ast.Node,
) (bool, bool) {
	var declarations []*ast.Node
	for _, declaration := range symbol.Declarations {
		if declaration != nil && ast.GetSourceFileOfNode(declaration) == ctx.SourceFile {
			declarations = append(declarations, declaration)
		}
	}
	if len(declarations) == 0 {
		return false, false
	}

	hoisted := member == "mock" || member == "mockRequire"
	recognized := false
	variableDeclarations := 0
	for _, declaration := range declarations {
		if enclosingVariableDeclaration(declaration) != nil {
			variableDeclarations++
		}
	}
	for _, declaration := range declarations {
		if declaration.Kind == ast.KindFunctionDeclaration {
			recognized = true
			// An ambient overload has no runtime value. For a lifted mock the
			// implementation must also be visible at the module's lift target.
			if declaration.Body() != nil && (!hoisted || (declaration.Parent != nil && declaration.Parent.Kind == ast.KindSourceFile)) {
				return true, true
			}
			continue
		}

		variable := enclosingVariableDeclaration(declaration)
		if variable == nil {
			continue
		}
		recognized = true
		// Repeated `var` declarations may each initialize the same binding.
		// Choosing one without modeling their execution order can mistake an
		// options value for a factory, so leave that binding undecided.
		if variableDeclarations != 1 {
			continue
		}
		initializer := utils.SkipAssertionsAndParens(variable.Initializer())
		if initializer == nil {
			continue
		}
		if hoistedInitializerIsAvailable(variable, initializer) {
			if directHoistedFactory(declaration, variable, initializer) || isCallable(ctx, reference) {
				return true, true
			}
			continue
		}
		if hoisted || !variableInitializerDominatesCall(variable, call) {
			continue
		}
		if ast.IsFunctionExpressionOrArrowFunction(initializer) || isCallable(ctx, reference) {
			return true, true
		}
	}
	return false, recognized
}

func enclosingVariableDeclaration(declaration *ast.Node) *ast.Node {
	if declaration.Kind == ast.KindVariableDeclaration {
		return declaration
	}
	if declaration.Kind == ast.KindBindingElement {
		return utils.EnclosingVariableDeclarationOfBindingElement(declaration)
	}
	return nil
}

func hoistedInitializerIsAvailable(variable, initializer *ast.Node) bool {
	declarationList := utils.GetDeclListForSymbolDecl(variable)
	if declarationList == nil || declarationList.Parent == nil ||
		declarationList.Parent.Kind != ast.KindVariableStatement ||
		declarationList.Parent.Parent == nil ||
		declarationList.Parent.Parent.Kind != ast.KindSourceFile {
		return false
	}
	if initializer.Kind != ast.KindCallExpression {
		return false
	}
	parsed := rstestUtils.ParseRstestPluginManagedCall(initializer)
	return parsed != nil && parsed.Member == "hoisted"
}

func directHoistedFactory(declaration, variable, initializer *ast.Node) bool {
	if initializer.Kind != ast.KindCallExpression {
		return false
	}
	call := initializer.AsCallExpression()
	if call.Arguments == nil || len(call.Arguments.Nodes) != 1 {
		return false
	}
	callback := utils.SkipAssertionsAndParens(call.Arguments.Nodes[0])
	if callback == nil || !ast.IsFunctionExpressionOrArrowFunction(callback) {
		return false
	}
	returned := callback.Body()
	if returned == nil {
		return false
	}
	if returned.Kind == ast.KindBlock {
		statements := returned.AsBlock().Statements
		if statements == nil || len(statements.Nodes) != 1 || statements.Nodes[0].Kind != ast.KindReturnStatement {
			return false
		}
		returned = statements.Nodes[0].AsReturnStatement().Expression
	}
	returned = utils.SkipAssertionsAndParens(returned)
	if returned == nil {
		return false
	}
	if variable.Name() != nil && variable.Name().Kind == ast.KindIdentifier {
		return ast.IsFunctionExpressionOrArrowFunction(returned)
	}
	destructured := destructuredHoistedValue(declaration, variable, returned)
	return destructured != nil && ast.IsFunctionExpressionOrArrowFunction(destructured)
}

func destructuredHoistedValue(declaration, variable, returned *ast.Node) *ast.Node {
	if declaration.Kind != ast.KindBindingElement || variable.Name() == nil {
		return nil
	}
	// Only a direct child of the declaration's root pattern shares a path with
	// the returned literal below. A nested binding needs a recursive path walk;
	// using only its leaf key could select an unrelated top-level property.
	if declaration.Parent != variable.Name() {
		return nil
	}
	binding := declaration.AsBindingElement()
	if binding == nil || binding.DotDotDotToken != nil {
		return nil
	}

	switch variable.Name().Kind {
	case ast.KindObjectBindingPattern:
		if returned.Kind != ast.KindObjectLiteralExpression {
			return nil
		}
		keyNode := binding.PropertyName
		if keyNode == nil {
			keyNode = binding.Name()
		}
		key, ok := utils.GetStaticPropertyName(keyNode)
		if !ok {
			return nil
		}
		var value *ast.Node
		for _, property := range returned.AsObjectLiteralExpression().Properties.Nodes {
			if property == nil || property.Kind != ast.KindPropertyAssignment {
				return nil
			}
			assignment := property.AsPropertyAssignment()
			propertyKey, ok := utils.GetStaticPropertyName(assignment.Name())
			if !ok {
				// A computed key may overwrite the property selected above.
				return nil
			}
			if propertyKey == key {
				if value != nil {
					return nil
				}
				value = utils.SkipAssertionsAndParens(assignment.Initializer)
			}
		}
		return value

	case ast.KindArrayBindingPattern:
		if returned.Kind != ast.KindArrayLiteralExpression {
			return nil
		}
		elements := variable.Name().AsBindingPattern().Elements
		values := returned.AsArrayLiteralExpression().Elements
		if elements == nil || values == nil {
			return nil
		}
		for _, value := range values.Nodes {
			if value == nil || value.Kind == ast.KindSpreadElement {
				return nil
			}
		}
		for index, element := range elements.Nodes {
			if element == declaration && index < len(values.Nodes) {
				return utils.SkipAssertionsAndParens(values.Nodes[index])
			}
		}
	}
	return nil
}

func variableInitializerDominatesCall(variable, call *ast.Node) bool {
	declarationList := utils.GetDeclListForSymbolDecl(variable)
	if declarationList == nil || declarationList.Parent == nil || declarationList.Parent.Kind != ast.KindVariableStatement {
		return false
	}
	declarationStatement := declarationList.Parent
	container := declarationStatement.Parent
	if container == nil {
		return false
	}

	callStatement := ast.FindAncestor(call, func(node *ast.Node) bool {
		return node.Kind == ast.KindExpressionStatement
	})
	if callStatement == nil {
		return false
	}
	current := callStatement
	for current.Parent != container {
		current = current.Parent
		if current == nil || ast.IsFunctionLikeOrClassStaticBlockDeclaration(current) {
			return false
		}
	}
	return declarationStatement.End() <= current.Pos()
}

func isCallable(ctx rule.RuleContext, node *ast.Node) bool {
	return ctx.TypeChecker != nil && len(utils.GetCallSignatures(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(node))) > 0
}
