package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func RstestCoreModuleFromRequireCall(node *ast.Node) (string, bool) {
	node = ast.SkipParentheses(node)
	if node == nil || !testFramework.IsModuleRequireCallModules(node, RstestCoreImportModules) {
		return "", false
	}
	arguments := node.Arguments()
	if len(arguments) == 0 {
		return "", false
	}
	specifier := ast.SkipParentheses(arguments[0])
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

func RequireBindingImportedName(element *ast.Node) string {
	binding := element.AsBindingElement()
	if binding == nil || binding.DotDotDotToken != nil {
		return ""
	}
	name := binding.PropertyName
	if name == nil {
		name = binding.Name()
	} else if name.Kind == ast.KindComputedPropertyName {
		name = ast.SkipParentheses(name.AsComputedPropertyName().Expression)
	}
	if name == nil || (name.Kind != ast.KindIdentifier && !ast.IsStringLiteralLike(name)) {
		return ""
	}
	if name.Kind == ast.KindIdentifier {
		return name.AsIdentifier().Text
	}
	return internalUtils.GetStaticStringValue(name)
}
