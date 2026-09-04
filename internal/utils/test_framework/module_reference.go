package test_framework

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

// matchesModule reports whether specifier is one of importModules. The lists
// carry at most a handful of entries, so a linear scan beats a map lookup and
// keeps the callers' package-level lists allocation free.
func matchesModule(specifier string, importModules []string) bool {
	for _, module := range importModules {
		if specifier == module {
			return true
		}
	}
	return false
}

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
	return ResolveFunctionReferenceForModules(
		node,
		localName,
		localNode,
		typeChecker,
		sourceFile,
		[]string{importModule},
	)
}

// ResolveFunctionReferenceForModules is ResolveFunctionReferenceForModule for a
// framework whose API surface is reachable through more than one specifier.
func ResolveFunctionReferenceForModules(
	node *ast.Node,
	localName string,
	localNode *ast.Node,
	typeChecker *checker.Checker,
	sourceFile *ast.SourceFile,
	importModules []string,
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

	return ResolveFunctionIdentifierReferenceModules(
		localName,
		identifier,
		typeChecker,
		sourceFile,
		importModules,
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
	return ResolveFunctionIdentifierReferenceModules(
		localName,
		identifier,
		typeChecker,
		sourceFile,
		[]string{importModule},
	)
}

// ResolveFunctionIdentifierReferenceModules is
// ResolveFunctionIdentifierReference across several equivalent specifiers.
func ResolveFunctionIdentifierReferenceModules(
	localName string,
	identifier *ast.Node,
	typeChecker *checker.Checker,
	sourceFile *ast.SourceFile,
	importModules []string,
) (string, *ast.Node, ReferenceMode) {
	if identifier == nil || identifier.Kind != ast.KindIdentifier || typeChecker == nil {
		return localName, identifier, ReferenceModeGlobal
	}

	symbol := typeChecker.GetSymbolAtLocation(identifier)
	return ResolveFunctionIdentifierReferenceFromSymbolModules(
		localName,
		identifier,
		symbol,
		sourceFile,
		importModules,
	)
}

// ResolveFunctionIdentifierReferenceFromSymbol resolves an identifier using
// its already-looked-up symbol. Callers that perform additional symbol-based
// checks can use this helper to avoid repeating TypeChecker lookups.
func ResolveFunctionIdentifierReferenceFromSymbol(
	localName string,
	identifier *ast.Node,
	symbol *ast.Symbol,
	sourceFile *ast.SourceFile,
	importModule string,
) (string, *ast.Node, ReferenceMode) {
	return ResolveFunctionIdentifierReferenceFromSymbolModules(
		localName,
		identifier,
		symbol,
		sourceFile,
		[]string{importModule},
	)
}

// ResolveFunctionIdentifierReferenceFromSymbolModules resolves an identifier
// against every specifier that reaches the same API surface.
//
// Every specifier is tried against each declaration in one pass, rather than
// re-running the whole resolution per specifier. A declaration that belongs to
// a non-matching module still lives in the linted file, so a per-specifier loop
// would reach the shadowing fallback below on its first miss and report the
// binding as a local declaration before the specifier that actually matches was
// ever tried.
func ResolveFunctionIdentifierReferenceFromSymbolModules(
	localName string,
	identifier *ast.Node,
	symbol *ast.Symbol,
	sourceFile *ast.SourceFile,
	importModules []string,
) (string, *ast.Node, ReferenceMode) {
	if identifier == nil || identifier.Kind != ast.KindIdentifier {
		return localName, identifier, ReferenceModeGlobal
	}
	if symbol == nil {
		return localName, identifier, ReferenceModeGlobal
	}

	hasLocalDeclaration := false
	for _, declaration := range symbol.Declarations {
		if declaration == nil {
			continue
		}

		// A type-only import is erased before runtime, so it neither binds the
		// module's export nor shadows anything: the call is the global
		// registration, whether the `type` sits on the clause or the specifier.
		if ast.IsTypeOnlyImportDeclaration(declaration) {
			continue
		}

		if name, originalNode, ok := resolveModuleImportSpecifier(declaration, importModules); ok {
			return name, originalNode, ReferenceModeImport
		}
		if name, originalNode, ok := resolveModuleRequireBinding(declaration, importModules); ok {
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

// IsModuleNamespaceSymbol reports whether symbol is a namespace import or a
// whole-module require for importModule.
func IsModuleNamespaceSymbol(symbol *ast.Symbol, importModule string) bool {
	return IsModuleNamespaceSymbolModules(symbol, []string{importModule})
}

// IsModuleNamespaceSymbolModules is IsModuleNamespaceSymbol across several
// equivalent specifiers.
func IsModuleNamespaceSymbolModules(symbol *ast.Symbol, importModules []string) bool {
	if symbol == nil {
		return false
	}

	for _, declaration := range symbol.Declarations {
		if declaration == nil {
			continue
		}

		// A type-only import is erased before runtime: it binds a name in the
		// type world only, so nothing reaches the module namespace through it.
		if ast.IsTypeOnlyImportDeclaration(declaration) {
			continue
		}

		if declaration.Kind == ast.KindNamespaceImport {
			importDeclaration := FindImportDeclaration(declaration)
			if importDeclaration != nil &&
				importDeclaration.ModuleSpecifier != nil &&
				matchesModule(importDeclaration.ModuleSpecifier.Text(), importModules) {
				return true
			}
		}

		if declaration.Kind == ast.KindVariableDeclaration {
			variable := declaration.AsVariableDeclaration()
			if variable != nil && IsModuleRequireCallModules(variable.Initializer, importModules) {
				return true
			}
		}

		// `import core = require('m')` binds the module namespace the same way
		// a namespace import does.
		if declaration.Kind == ast.KindImportEqualsDeclaration {
			if isImportEqualsOfModules(declaration, importModules) {
				return true
			}
		}
	}

	return false
}

// isImportEqualsOfModules reports whether declaration is
// `import name = require(module)` for any of importModules.
func isImportEqualsOfModules(declaration *ast.Node, importModules []string) bool {
	importEquals := declaration.AsImportEqualsDeclaration()
	if importEquals == nil || importEquals.ModuleReference == nil ||
		importEquals.ModuleReference.Kind != ast.KindExternalModuleReference {
		return false
	}
	reference := importEquals.ModuleReference.AsExternalModuleReference()
	if reference == nil || reference.Expression == nil {
		return false
	}
	specifier := internalUtils.SkipAssertionsAndParens(reference.Expression)
	if specifier == nil || !ast.IsStringLiteralLike(specifier) {
		return false
	}
	return matchesModule(specifier.Text(), importModules)
}

func resolveModuleImportSpecifier(declaration *ast.Node, importModules []string) (string, *ast.Node, bool) {
	if declaration == nil || declaration.Kind != ast.KindImportSpecifier {
		return "", nil, false
	}

	importDeclaration := FindImportDeclaration(declaration)
	if importDeclaration == nil ||
		importDeclaration.ModuleSpecifier == nil ||
		!matchesModule(importDeclaration.ModuleSpecifier.Text(), importModules) {
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

func resolveModuleRequireBinding(declaration *ast.Node, importModules []string) (string, *ast.Node, bool) {
	if declaration == nil || declaration.Kind != ast.KindBindingElement {
		return "", nil, false
	}

	varDeclaration := internalUtils.EnclosingVariableDeclarationOfBindingElement(declaration)
	if varDeclaration == nil ||
		!IsModuleRequireCallModules(varDeclaration.AsVariableDeclaration().Initializer, importModules) {
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
	return IsModuleRequireCallModules(node, []string{importModule})
}

// IsModuleRequireCallModules reports whether node is a require call for any of
// importModules.
func IsModuleRequireCallModules(node *ast.Node, importModules []string) bool {
	if node == nil {
		return false
	}

	// TypeScript assertions are erased before runtime, so
	// `require('m') as any` still binds the module.
	node = internalUtils.SkipAssertionsAndParens(node)
	// The argument is checked below rather than by IsRequireCall, which would
	// reject `require('m' as any)` before the assertion is skipped.
	if node == nil || !ast.IsRequireCall(node, false /* requireStringLiteralLikeArgument */) {
		return false
	}

	arguments := node.Arguments()
	if len(arguments) == 0 || arguments[0] == nil {
		return false
	}
	specifier := internalUtils.SkipAssertionsAndParens(arguments[0])
	if specifier == nil {
		return false
	}

	switch specifier.Kind {
	case ast.KindStringLiteral:
		return matchesModule(specifier.AsStringLiteral().Text, importModules)
	case ast.KindNoSubstitutionTemplateLiteral:
		return matchesModule(specifier.AsNoSubstitutionTemplateLiteral().Text, importModules)
	default:
		return false
	}
}
