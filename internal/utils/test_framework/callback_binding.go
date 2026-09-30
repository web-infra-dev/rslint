package test_framework

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

// LocalFunctionBinding returns the function a callback identifier's binding
// always holds, or nil when that cannot be proven from sourceFile alone.
//
// symbol must be the binder symbol the identifier resolves to through refs. The
// binding qualifies only when it has exactly one declaration in sourceFile, is
// never written, and that declaration is a function declaration or a variable
// whose initializer is a function expression or arrow function (through
// parentheses and TypeScript assertions). Any other binding can hold a function
// the file does not show, so attributing a body to it would be a guess.
func LocalFunctionBinding(
	sourceFile *ast.SourceFile,
	refs *rule.RefStore,
	symbol *ast.Symbol,
) *ast.Node {
	if refs == nil || symbol == nil || len(symbol.Declarations) != 1 {
		return nil
	}
	declaration := symbol.Declarations[0]
	if declaration == nil || ast.GetSourceFileOfNode(declaration) != sourceFile ||
		localFunctionBindingIsWritten(refs, symbol) {
		return nil
	}
	switch declaration.Kind {
	case ast.KindFunctionDeclaration:
		return declaration
	case ast.KindVariableDeclaration:
		initializer := internalUtils.SkipAssertionsAndParens(declaration.AsVariableDeclaration().Initializer)
		if initializer != nil && ast.IsFunctionExpressionOrArrowFunction(initializer) {
			return initializer
		}
	}
	return nil
}

// LocalFunctionImplementation returns the unique same-file implementation a
// callback symbol always denotes. Bodyless overload signatures are ignored;
// mutable bindings and ambiguous runtime declarations are rejected.
func LocalFunctionImplementation(
	sourceFile *ast.SourceFile,
	refs *rule.RefStore,
	symbol *ast.Symbol,
) *ast.Node {
	if refs == nil || symbol == nil {
		return nil
	}
	if localFunctionBindingIsWritten(refs, symbol) {
		return nil
	}

	var implementation *ast.Node
	for _, declaration := range symbol.Declarations {
		if declaration == nil || ast.GetSourceFileOfNode(declaration) != sourceFile {
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

func localFunctionBindingIsWritten(refs *rule.RefStore, symbol *ast.Symbol) bool {
	for _, reference := range refs.References(symbol) {
		if internalUtils.IsWriteReference(reference) {
			return true
		}
	}
	return false
}
