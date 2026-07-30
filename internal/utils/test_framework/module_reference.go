package test_framework

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

// ResolveFunctionReferenceForModule resolves an imported/required test API
// back to its exported name. Local declarations shadow framework globals.
func ResolveFunctionReferenceForModule(
	node *ast.Node,
	localName string,
	localNode *ast.Node,
	typeChecker *checker.Checker,
	sourceFile *ast.SourceFile,
	importModule string,
) (string, *ast.Node, ReferenceMode) {
	if typeChecker == nil {
		return localName, localNode, ReferenceModeGlobal
	}

	callExpr := node.AsCallExpression()
	if callExpr == nil {
		return localName, localNode, ReferenceModeGlobal
	}

	identifier := ResolveFirstIdentifier(callExpr.Expression)
	if identifier == nil || identifier.Kind != ast.KindIdentifier {
		return localName, localNode, ReferenceModeGlobal
	}

	return ResolveFunctionIdentifierReference(
		localName,
		identifier,
		typeChecker,
		sourceFile,
		importModule,
	)
}

// ResolveFunctionIdentifierReference resolves one identifier as a framework
// global or a named import/require binding from importModule.
func ResolveFunctionIdentifierReference(
	localName string,
	identifier *ast.Node,
	typeChecker *checker.Checker,
	sourceFile *ast.SourceFile,
	importModule string,
) (string, *ast.Node, ReferenceMode) {
	if identifier == nil || identifier.Kind != ast.KindIdentifier || typeChecker == nil {
		return localName, identifier, ReferenceModeGlobal
	}

	symbol := typeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil {
		return localName, identifier, ReferenceModeGlobal
	}

	hasLocalDeclaration := false
	for _, declaration := range symbol.Declarations {
		if declaration == nil {
			continue
		}

		if name, originalNode, ok := resolveModuleImportSpecifier(declaration, importModule); ok {
			return name, originalNode, ReferenceModeImport
		}
		if name, originalNode, ok := resolveModuleRequireBinding(declaration, importModule); ok {
			return name, originalNode, ReferenceModeImport
		}
		if sourceFile != nil && ast.GetSourceFileOfNode(declaration) == sourceFile {
			hasLocalDeclaration = true
		}
	}

	if hasLocalDeclaration {
		return "", nil, ReferenceModeGlobal
	}
	return localName, identifier, ReferenceModeGlobal
}

// FindImportDeclaration returns the import declaration enclosing node.
func FindImportDeclaration(node *ast.Node) *ast.ImportDeclaration {
	for current := node; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindImportDeclaration, ast.KindJSImportDeclaration:
			return current.AsImportDeclaration()
		}
	}
	return nil
}

// IsModuleNamespaceReference reports whether identifier is bound to a namespace
// import or a whole-module require for importModule.
func IsModuleNamespaceReference(
	identifier *ast.Node,
	typeChecker *checker.Checker,
	importModule string,
) bool {
	if identifier == nil || identifier.Kind != ast.KindIdentifier || typeChecker == nil {
		return false
	}

	symbol := typeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil {
		return false
	}

	for _, declaration := range symbol.Declarations {
		if declaration == nil {
			continue
		}

		if declaration.Kind == ast.KindNamespaceImport {
			importDeclaration := FindImportDeclaration(declaration)
			if importDeclaration != nil &&
				importDeclaration.ModuleSpecifier != nil &&
				importDeclaration.ModuleSpecifier.Text() == importModule {
				return true
			}
		}

		if declaration.Kind == ast.KindVariableDeclaration {
			variable := declaration.AsVariableDeclaration()
			if variable != nil && IsModuleRequireCall(variable.Initializer, importModule) {
				return true
			}
		}
	}

	return false
}

func resolveModuleImportSpecifier(declaration *ast.Node, importModule string) (string, *ast.Node, bool) {
	if declaration == nil || declaration.Kind != ast.KindImportSpecifier {
		return "", nil, false
	}

	importDeclaration := FindImportDeclaration(declaration)
	if importDeclaration == nil ||
		importDeclaration.ModuleSpecifier == nil ||
		importDeclaration.ModuleSpecifier.Text() != importModule {
		return "", nil, false
	}

	specifier := declaration.AsImportSpecifier()
	if specifier == nil || specifier.IsTypeOnly {
		return "", nil, false
	}
	if specifier.PropertyName != nil {
		return specifier.PropertyName.Text(), specifier.PropertyName, true
	}
	name := specifier.Name()
	if name == nil {
		return "", nil, false
	}
	return name.Text(), name, true
}

func resolveModuleRequireBinding(declaration *ast.Node, importModule string) (string, *ast.Node, bool) {
	if declaration == nil || declaration.Kind != ast.KindBindingElement {
		return "", nil, false
	}

	varDeclaration := internalUtils.EnclosingVariableDeclarationOfBindingElement(declaration)
	if varDeclaration == nil ||
		!IsModuleRequireCall(varDeclaration.AsVariableDeclaration().Initializer, importModule) {
		return "", nil, false
	}

	binding := declaration.AsBindingElement()
	if binding == nil {
		return "", nil, false
	}
	if binding.PropertyName != nil {
		if name := propertyName(binding.PropertyName); name != "" {
			return name, binding.PropertyName, true
		}
	}
	nameNode := binding.Name()
	if name := propertyName(nameNode); name != "" {
		return name, nameNode, true
	}
	return "", nil, false
}

// IsModuleRequireCall reports whether node is require(importModule).
func IsModuleRequireCall(node *ast.Node, importModule string) bool {
	node = ast.SkipParentheses(node)
	if node == nil || !ast.IsRequireCall(node, true /* requireStringLiteralLikeArgument */) {
		return false
	}

	arguments := node.Arguments()
	if len(arguments) == 0 || arguments[0] == nil {
		return false
	}
	specifier := ast.SkipParentheses(arguments[0])
	if specifier == nil {
		return false
	}

	switch specifier.Kind {
	case ast.KindStringLiteral:
		return specifier.AsStringLiteral().Text == importModule
	case ast.KindNoSubstitutionTemplateLiteral:
		return specifier.AsNoSubstitutionTemplateLiteral().Text == importModule
	default:
		return false
	}
}
