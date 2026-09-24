package test_framework

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

// ResolveLocalFunctionBinding returns the unique same-file function
// implementation referenced by identifier. Overload signatures without a body
// are ignored. Mutable bindings and ambiguous runtime declarations are rejected.
func ResolveLocalFunctionBinding(ctx rule.RuleContext, identifier *ast.Node) *ast.Node {
	if identifier == nil || identifier.Kind != ast.KindIdentifier || ctx.Refs == nil {
		return nil
	}
	return ResolveLocalFunctionSymbol(ctx, ctx.Refs.Resolve(identifier))
}

// ResolveLocalFunctionSymbol is ResolveLocalFunctionBinding for callers that
// have already resolved the binding symbol.
func ResolveLocalFunctionSymbol(ctx rule.RuleContext, symbol *ast.Symbol) *ast.Node {
	if symbol == nil || ctx.SourceFile == nil || ctx.Refs == nil {
		return nil
	}
	for _, reference := range ctx.Refs.References(symbol) {
		if internalUtils.IsWriteReference(reference) {
			return nil
		}
	}

	var implementation *ast.Node
	for _, declaration := range symbol.Declarations {
		if declaration == nil || ast.GetSourceFileOfNode(declaration) != ctx.SourceFile {
			return nil
		}
		var candidate *ast.Node
		switch declaration.Kind {
		case ast.KindFunctionDeclaration:
			if declaration.Body() == nil {
				continue
			}
			candidate = declaration
		case ast.KindVariableDeclaration:
			candidate = internalUtils.SkipAssertionsAndParens(
				declaration.AsVariableDeclaration().Initializer,
			)
			if candidate == nil || !ast.IsFunctionExpressionOrArrowFunction(candidate) {
				return nil
			}
		default:
			return nil
		}
		if implementation != nil {
			return nil
		}
		implementation = candidate
	}
	return implementation
}
