package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

// RstestCoreModuleFromRequireCall returns the module an rstest `require` call
// names. TypeScript assertions are erased before runtime, so they are skipped
// on both the call and its argument, matching what IsModuleRequireCallModules
// accepts: unwrapping only parentheses would let `require('@rstest/core') as
// any` pass the check and then read no argument at all, and would leave the
// module name of `require('@rstest/core' as any)` empty.
func RstestCoreModuleFromRequireCall(node *ast.Node) (string, bool) {
	node = internalUtils.SkipAssertionsAndParens(node)
	if node == nil || !testFramework.IsModuleRequireCallModules(node, RstestCoreImportModules) {
		return "", false
	}
	arguments := node.Arguments()
	if len(arguments) == 0 {
		return "", false
	}
	specifier := internalUtils.SkipAssertionsAndParens(arguments[0])
	if specifier == nil {
		return "", false
	}
	return internalUtils.GetStaticStringValue(specifier), true
}

func NamedImportElements(declaration *ast.ImportDeclaration) []*ast.Node {
	if declaration == nil || declaration.ImportClause == nil || declaration.ImportClause.IsTypeOnly() {
		return nil
	}
	clause := declaration.ImportClause.AsImportClause()
	if clause == nil || clause.NamedBindings == nil || clause.NamedBindings.Kind != ast.KindNamedImports {
		return nil
	}
	named := clause.NamedBindings.AsNamedImports()
	if named == nil || named.Elements == nil {
		return nil
	}
	return named.Elements.Nodes
}

func ImportedSpecifierName(element *ast.Node) string {
	specifier := element.AsImportSpecifier()
	if specifier == nil || specifier.IsTypeOnly {
		return ""
	}
	name := specifier.Name()
	if specifier.PropertyName != nil {
		name = specifier.PropertyName
	}
	if name == nil {
		return ""
	}
	return name.Text()
}

// RequireBindingImportedName returns the name a destructured `require` binding
// pulls off the module. An identifier key and a string-literal key name the
// same export, so `{ expect }` and `{ 'expect': local }` both report `expect`.
// A computed key only counts when it is a static string: `{ [expect]: local }`
// reads whatever `expect` holds at runtime and names no known export, so
// treating its text as the key would report a binding the module never
// provides.
func RequireBindingImportedName(element *ast.Node) string {
	binding := element.AsBindingElement()
	if binding == nil || binding.DotDotDotToken != nil {
		return ""
	}
	name := binding.PropertyName
	if name == nil {
		name = binding.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return ""
		}
		return name.AsIdentifier().Text
	}
	if name.Kind == ast.KindComputedPropertyName {
		value, ok := internalUtils.GetStaticStringLiteralValue(ast.SkipParentheses(name.AsComputedPropertyName().Expression))
		if !ok {
			return ""
		}
		return value
	}
	if name.Kind == ast.KindIdentifier {
		return name.AsIdentifier().Text
	}
	if !ast.IsStringLiteralLike(name) {
		return ""
	}
	return internalUtils.GetStaticStringValue(name)
}
