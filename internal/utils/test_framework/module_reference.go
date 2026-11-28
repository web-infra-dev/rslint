package test_framework

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
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
	if node == nil || node.Kind != ast.KindCallExpression {
		return "", nil, ReferenceModeGlobal
	}

	callExpr := node.AsCallExpression()
	if callExpr == nil {
		return "", nil, ReferenceModeGlobal
	}

	identifier := ResolveFirstIdentifier(callExpr.Expression)
	if identifier == nil || identifier.Kind != ast.KindIdentifier {
		return "", nil, ReferenceModeGlobal
	}

	if typeChecker == nil {
		return localName, identifier, ReferenceModeGlobal
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
	return resolveIdentifierReferenceFromSymbolModules(
		localName,
		identifier,
		symbol,
		sourceFile,
		importModules,
		false,
	)
}

// ResolveTypeIdentifierReferenceFromSymbolModules resolves an identifier used
// in a type position against every specifier that exports the same type
// surface. Unlike the function-reference resolver, it accepts a type-only
// import, written either as `import type { Mock } from "@rstest/core"` or as
// `import { type Mock } from "@rstest/core"`.
func ResolveTypeIdentifierReferenceFromSymbolModules(
	localName string,
	identifier *ast.Node,
	symbol *ast.Symbol,
	sourceFile *ast.SourceFile,
	importModules []string,
) (string, *ast.Node, ReferenceMode) {
	return resolveIdentifierReferenceFromSymbolModules(
		localName,
		identifier,
		symbol,
		sourceFile,
		importModules,
		true,
	)
}

func resolveIdentifierReferenceFromSymbolModules(
	localName string,
	identifier *ast.Node,
	symbol *ast.Symbol,
	sourceFile *ast.SourceFile,
	importModules []string,
	allowTypeOnlySpecifier bool,
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
		// A caller resolving an identifier in a type position asks for the
		// opposite and opts out.
		if !allowTypeOnlySpecifier && ast.IsTypeOnlyImportDeclaration(declaration) {
			continue
		}

		if name, originalNode, ok := resolveModuleImportSpecifier(declaration, importModules, allowTypeOnlySpecifier); ok {
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

// IsNamedESMImportSymbolModules reports whether symbol is a runtime named ESM
// import of one of exportNames from one of importModules. CommonJS require
// bindings deliberately do not match: TypeScript types require() as any, so a
// caller cannot use this fact to prove that generic call syntax is valid.
func IsNamedESMImportSymbolModules(symbol *ast.Symbol, importModules, exportNames []string) bool {
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil || ast.IsTypeOnlyImportDeclaration(declaration) {
			continue
		}
		name, _, ok := resolveModuleImportSpecifier(declaration, importModules, false)
		if !ok {
			continue
		}
		for _, exportName := range exportNames {
			if name == exportName {
				return true
			}
		}
	}
	return false
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
			if IsImportEqualsOfModules(declaration, importModules) {
				return true
			}
		}
	}

	return false
}

// IsImportEqualsOfModules reports whether declaration is
// `import name = require(module)` for any of importModules.
func IsImportEqualsOfModules(declaration *ast.Node, importModules []string) bool {
	if declaration == nil || declaration.Kind != ast.KindImportEqualsDeclaration {
		return false
	}
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

func resolveModuleImportSpecifier(declaration *ast.Node, importModules []string, allowTypeOnly bool) (string, *ast.Node, bool) {
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
	// A type-only binding names no value at runtime, so a caller resolving a
	// function reference must not accept one. The check covers the inline
	// `type` modifier and the one on the enclosing clause alike.
	if specifier == nil || (ast.IsTypeOnlyImportDeclaration(declaration) && !allowTypeOnly) {
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
	return resolveModuleRequireBindingAtDepth(declaration, importModules, 0)
}

func resolveModuleRequireBindingAtDepth(declaration *ast.Node, importModules []string, depth int) (string, *ast.Node, bool) {
	if declaration == nil || declaration.Kind != ast.KindBindingElement {
		return "", nil, false
	}

	binding := declaration.AsBindingElement()
	if binding == nil || binding.DotDotDotToken != nil {
		return "", nil, false
	}

	varDeclaration := internalUtils.EnclosingVariableDeclarationOfBindingElement(declaration)
	if varDeclaration == nil ||
		varDeclaration.Kind != ast.KindVariableDeclaration ||
		declaration.Parent == nil ||
		declaration.Parent.Kind != ast.KindObjectBindingPattern ||
		declaration.Parent != varDeclaration.Name() ||
		!isModuleRequireCallModules(varDeclaration.AsVariableDeclaration().Initializer, importModules, depth) {
		return "", nil, false
	}

	if binding.PropertyName != nil {
		if name, ok := internalUtils.GetStaticPropertyName(binding.PropertyName); ok && name != "" {
			return name, binding.PropertyName, true
		}
		return "", nil, false
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
	return isModuleRequireCallModules(node, importModules, 0)
}

func isModuleRequireCallModules(node *ast.Node, importModules []string, depth int) bool {
	if node == nil || depth > 4 {
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

	matched := false
	switch specifier.Kind {
	case ast.KindStringLiteral:
		matched = matchesModule(specifier.AsStringLiteral().Text, importModules)
	case ast.KindNoSubstitutionTemplateLiteral:
		matched = matchesModule(specifier.AsNoSubstitutionTemplateLiteral().Text, importModules)
	}
	if !matched {
		return false
	}

	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee == nil || callee.Kind != ast.KindIdentifier || isShadowedByNonModuleRequire(callee, depth) {
		return false
	}
	return true
}

var nodeModuleSpecifiers = []string{"node:module", "module"}

// isShadowedByNonModuleRequire reports whether a `require(...)` callee resolves
// to a local declaration that is NOT Node's module loader. Ambient TypeScript
// declarations (`declare const/function require`) and `const` bindings
// initialized from `createRequire(...)` of `node:module` / `module` preserve
// real module loading; local mocks, mutable bindings, and parameters do not.
func isShadowedByNonModuleRequire(callee *ast.Node, depth int) bool {
	if callee == nil || callee.Kind != ast.KindIdentifier {
		return true
	}
	name := callee.AsIdentifier().Text
	if !internalUtils.IsShadowed(callee, name) {
		return false
	}
	decl, ambient := findInnermostDeclaration(callee, name)
	if decl == nil {
		// IsShadowed was true, so never assume a missing declaration is ambient.
		return true
	}
	if ambient {
		// An ambient (`declare ...`) binding is erased at runtime, so `require`
		// still resolves to the global CommonJS loader.
		return false
	}
	if decl.Kind != ast.KindVariableDeclaration || !isImmutableConstDeclaration(decl) {
		return true
	}
	varDecl := decl.AsVariableDeclaration()
	if varDecl == nil || varDecl.Initializer == nil ||
		varDecl.Name() == nil || varDecl.Name().Kind != ast.KindIdentifier ||
		varDecl.Name().AsIdentifier().Text != name {
		return true
	}
	return !isNodeCreateRequireCall(varDecl.Initializer, depth+1)
}

func isImmutableConstDeclaration(decl *ast.Node) bool {
	return decl != nil &&
		decl.Kind == ast.KindVariableDeclaration &&
		ast.IsVarConst(decl) &&
		!ast.IsVarAwaitUsing(decl)
}

func findInnermostDeclaration(node *ast.Node, name string) (decl *ast.Node, ambient bool) {
	prev := node
	inParameterDecorator := false
	crossedScope := false
	var ambientDecl *ast.Node
	record := func(d *ast.Node, amb bool) bool {
		if d == nil {
			return false
		}
		if !amb {
			return true
		}
		if ambientDecl == nil {
			ambientDecl = d
		}
		return false
	}
	for current := node.Parent; current != nil; prev, current = current, current.Parent {
		if current.Kind == ast.KindParameter {
			inParameterDecorator = prev.Kind == ast.KindDecorator
		}
		switch current.Kind {
		case ast.KindSourceFile:
			if sf := current.AsSourceFile(); sf != nil && sf.Statements != nil {
				if d, amb := findDeclInStatements(sf.Statements.Nodes, name, true); record(d, amb) {
					return d, false
				}
			}
			return ambientDecl, ambientDecl != nil
		case ast.KindBlock:
			if block := current.AsBlock(); block != nil && block.Statements != nil {
				if d, amb := findDeclInStatements(block.Statements.Nodes, name, false); record(d, amb) {
					return d, false
				}
			}
		case ast.KindModuleBlock:
			if block := current.AsModuleBlock(); block != nil && block.Statements != nil {
				if d, amb := findDeclInStatements(block.Statements.Nodes, name, true); record(d, amb) {
					return d, false
				}
			}
		case ast.KindClassStaticBlockDeclaration:
			if sb := current.AsClassStaticBlockDeclaration(); sb != nil && sb.Body != nil && sb.Body.Kind == ast.KindBlock {
				if block := sb.Body.AsBlock(); block != nil && block.Statements != nil {
					if d, amb := findDeclInStatements(block.Statements.Nodes, name, true); record(d, amb) {
						return d, false
					}
				}
			}
		case ast.KindCaseBlock:
			if cb := current.AsCaseBlock(); cb != nil && cb.Clauses != nil {
				for _, clause := range cb.Clauses.Nodes {
					if c := clause.AsCaseOrDefaultClause(); c != nil && c.Statements != nil {
						if d, amb := findDeclInStatements(c.Statements.Nodes, name, false); record(d, amb) {
							return d, false
						}
					}
				}
			}
		case ast.KindCatchClause:
			if cc := current.AsCatchClause(); cc != nil && cc.VariableDeclaration != nil {
				if bindingDeclaresName(cc.VariableDeclaration.Name(), name) {
					return cc.VariableDeclaration, false
				}
			}
		case ast.KindForStatement:
			if fs := current.AsForStatement(); fs != nil && fs.Initializer != nil && fs.Initializer.Kind == ast.KindVariableDeclarationList {
				if d := findDeclInVarDeclList(fs.Initializer.AsVariableDeclarationList(), name); d != nil {
					return d, false
				}
			}
		case ast.KindForInStatement, ast.KindForOfStatement:
			if fs := current.AsForInOrOfStatement(); fs != nil && fs.Initializer != nil && fs.Initializer.Kind == ast.KindVariableDeclarationList {
				if d := findDeclInVarDeclList(fs.Initializer.AsVariableDeclarationList(), name); d != nil {
					return d, false
				}
			}
		case ast.KindClassDeclaration, ast.KindClassExpression:
			if prev.Kind == ast.KindDecorator && crossedScope {
				continue
			}
			if current.Kind == ast.KindClassExpression && prev != current.Name() && prev.Kind != ast.KindDecorator {
				if n := current.Name(); n != nil && n.Kind == ast.KindIdentifier && n.AsIdentifier().Text == name {
					return current, false
				}
			}
			crossedScope = true
		default:
			if IsFunction(current) {
				// Decorators, computed property keys, and closures inside
				// parameter decorators evaluate in the enclosing scope.
				if prev.Kind == ast.KindDecorator || prev == current.Name() ||
					(inParameterDecorator && crossedScope && isDirectParameterNode(current, prev)) {
					continue
				}
				crossedScope = true
				inParam := isDirectParameterNode(current, prev)
				// Check parameters before hoisted body `var` declarations so a
				// parameter binding (e.g. `function run(require) { if (false) var require = ... }`)
				// is never mistaken for an unconditional `createRequire` binding.
				for _, param := range current.Parameters() {
					if param != nil && bindingDeclaresName(param.Name(), name) {
						return param, false
					}
				}
				if !inParam && current.Body() != nil && current.Body().Kind == ast.KindBlock {
					if block := current.Body().AsBlock(); block != nil && block.Statements != nil {
						if d, amb := findHoistedVarOrFnInStatements(block.Statements.Nodes, name, true); record(d, amb) {
							return d, false
						}
					}
				}
				if current.Kind == ast.KindFunctionExpression || current.Kind == ast.KindFunctionDeclaration {
					if fnName := current.Name(); fnName != nil && fnName.Kind == ast.KindIdentifier && fnName.AsIdentifier().Text == name {
						if record(current, ast.HasSyntacticModifier(current, ast.ModifierFlagsAmbient)) {
							return current, false
						}
					}
				}
			}
		}
	}
	return ambientDecl, ambientDecl != nil
}

func isDirectParameterNode(fn *ast.Node, child *ast.Node) bool {
	for _, param := range fn.Parameters() {
		if param == child {
			return true
		}
	}
	return false
}

func findDeclInStatements(statements []*ast.Node, name string, includeHoisted bool) (*ast.Node, bool) {
	var ambientDecl *ast.Node
	for _, stmt := range statements {
		if stmt == nil {
			continue
		}
		isAmb := ast.HasSyntacticModifier(stmt, ast.ModifierFlagsAmbient)
		var matched *ast.Node
		switch stmt.Kind {
		case ast.KindVariableStatement:
			if vs := stmt.AsVariableStatement(); vs != nil && vs.DeclarationList != nil {
				matched = findDeclInVarDeclList(vs.DeclarationList.AsVariableDeclarationList(), name)
			}
		case ast.KindFunctionDeclaration, ast.KindClassDeclaration, ast.KindEnumDeclaration, ast.KindModuleDeclaration:
			if n := stmt.Name(); n != nil && n.Kind == ast.KindIdentifier && n.AsIdentifier().Text == name {
				matched = stmt
			}
		case ast.KindImportEqualsDeclaration:
			if ie := stmt.AsImportEqualsDeclaration(); ie != nil && ie.Name() != nil && ie.Name().Text() == name {
				matched = stmt
				if ie.IsTypeOnly {
					isAmb = true
				}
			}
		case ast.KindImportDeclaration:
			if d, typeOnly := findImportDeclMatchingName(stmt, name); d != nil {
				matched = d
				if typeOnly {
					isAmb = true
				}
			}
		}
		if matched != nil {
			if !isAmb {
				return matched, false
			}
			if ambientDecl == nil {
				ambientDecl = matched
			}
		}
	}
	if d, amb := findHoistedVarOrFnInStatements(statements, name, includeHoisted); d != nil {
		if !amb {
			return d, false
		}
		if ambientDecl == nil {
			ambientDecl = d
		}
	}
	if ambientDecl != nil {
		return ambientDecl, true
	}
	return nil, false
}

// findHoistedVarOrFnInStatements finds `var` declarations (when includeVars is
// true) and unbraced function declarations hoisted out of nested non-function
// statements into the enclosing scope. Only statement containers are descended
// into (never expression subtrees), and braced blocks / `using` / `let` / `const`
// keep lexical bindings block-scoped.
func findHoistedVarOrFnInStatements(statements []*ast.Node, name string, includeVars bool) (*ast.Node, bool) {
	var found *ast.Node
	var foundAmbient bool
	var visit func(*ast.Node, bool)
	visit = func(node *ast.Node, inNestedScope bool) {
		if node == nil || (found != nil && !foundAmbient) {
			return
		}
		switch node.Kind {
		case ast.KindFunctionDeclaration:
			if !inNestedScope {
				if n := node.Name(); n != nil && n.Kind == ast.KindIdentifier && n.AsIdentifier().Text == name {
					isAmb := ast.HasSyntacticModifier(node, ast.ModifierFlagsAmbient)
					if !isAmb || found == nil {
						found = node
						foundAmbient = isAmb
					}
				}
			}
		case ast.KindBlock:
			if !includeVars {
				return
			}
			if block := node.AsBlock(); block != nil && block.Statements != nil {
				for _, s := range block.Statements.Nodes {
					visit(s, true)
					if found != nil && !foundAmbient {
						return
					}
				}
			}
		case ast.KindCaseBlock:
			if !includeVars {
				return
			}
			if cb := node.AsCaseBlock(); cb != nil && cb.Clauses != nil {
				for _, clause := range cb.Clauses.Nodes {
					if c := clause.AsCaseOrDefaultClause(); c != nil && c.Statements != nil {
						for _, s := range c.Statements.Nodes {
							visit(s, true)
							if found != nil && !foundAmbient {
								return
							}
						}
					}
				}
			}
		case ast.KindVariableStatement:
			if !includeVars {
				return
			}
			vs := node.AsVariableStatement()
			if vs == nil || vs.DeclarationList == nil {
				return
			}
			declList := vs.DeclarationList
			if declList.Flags&ast.NodeFlagsBlockScoped != 0 {
				return
			}
			if d := findDeclInVarDeclList(declList.AsVariableDeclarationList(), name); d != nil {
				isAmb := ast.HasSyntacticModifier(node, ast.ModifierFlagsAmbient)
				if !isAmb || found == nil {
					found = d
					foundAmbient = isAmb
				}
			}
		case ast.KindIfStatement:
			if is := node.AsIfStatement(); is != nil {
				visit(is.ThenStatement, inNestedScope)
				visit(is.ElseStatement, inNestedScope)
			}
		case ast.KindLabeledStatement:
			if ls := node.AsLabeledStatement(); ls != nil {
				visit(ls.Statement, inNestedScope)
			}
		case ast.KindWhileStatement:
			if ws := node.AsWhileStatement(); ws != nil {
				visit(ws.Statement, inNestedScope)
			}
		case ast.KindDoStatement:
			if ds := node.AsDoStatement(); ds != nil {
				visit(ds.Statement, inNestedScope)
			}
		case ast.KindWithStatement:
			if ws := node.AsWithStatement(); ws != nil {
				visit(ws.Statement, inNestedScope)
			}
		case ast.KindSwitchStatement:
			if ss := node.AsSwitchStatement(); ss != nil {
				visit(ss.CaseBlock, true)
			}
		case ast.KindTryStatement:
			if ts := node.AsTryStatement(); ts != nil {
				visit(ts.TryBlock, true)
				if ts.CatchClause != nil {
					if cc := ts.CatchClause.AsCatchClause(); cc != nil {
						visit(cc.Block, true)
					}
				}
				visit(ts.FinallyBlock, true)
			}
		case ast.KindForStatement:
			if fs := node.AsForStatement(); fs != nil {
				lexInit := fs.Initializer != nil && fs.Initializer.Kind == ast.KindVariableDeclarationList &&
					fs.Initializer.Flags&ast.NodeFlagsBlockScoped != 0
				if includeVars && fs.Initializer != nil && fs.Initializer.Kind == ast.KindVariableDeclarationList && !lexInit {
					if d := findDeclInVarDeclList(fs.Initializer.AsVariableDeclarationList(), name); d != nil {
						found = d
						foundAmbient = false
						return
					}
				}
				visit(fs.Statement, inNestedScope || lexInit)
			}
		case ast.KindForInStatement, ast.KindForOfStatement:
			if fs := node.AsForInOrOfStatement(); fs != nil {
				lexInit := fs.Initializer != nil && fs.Initializer.Kind == ast.KindVariableDeclarationList &&
					fs.Initializer.Flags&ast.NodeFlagsBlockScoped != 0
				if includeVars && fs.Initializer != nil && fs.Initializer.Kind == ast.KindVariableDeclarationList && !lexInit {
					if d := findDeclInVarDeclList(fs.Initializer.AsVariableDeclarationList(), name); d != nil {
						found = d
						foundAmbient = false
						return
					}
				}
				visit(fs.Statement, inNestedScope || lexInit)
			}
		}
	}
	for _, stmt := range statements {
		visit(stmt, false)
		if found != nil && !foundAmbient {
			return found, false
		}
	}
	return found, foundAmbient
}

func findImportDeclMatchingName(stmt *ast.Node, name string) (*ast.Node, bool) {
	imp := stmt.AsImportDeclaration()
	if imp == nil || imp.ImportClause == nil {
		return nil, false
	}
	clauseTypeOnly := imp.ImportClause.IsTypeOnly()
	clause := imp.ImportClause.AsImportClause()
	if clause == nil {
		return nil, false
	}
	if clause.Name() != nil && clause.Name().Text() == name {
		return stmt, clauseTypeOnly
	}
	if clause.NamedBindings != nil {
		switch clause.NamedBindings.Kind {
		case ast.KindNamespaceImport:
			if ns := clause.NamedBindings.AsNamespaceImport(); ns != nil && ns.Name() != nil && ns.Name().Text() == name {
				return clause.NamedBindings, clauseTypeOnly
			}
		case ast.KindNamedImports:
			if named := clause.NamedBindings.AsNamedImports(); named != nil && named.Elements != nil {
				for _, elem := range named.Elements.Nodes {
					if spec := elem.AsImportSpecifier(); spec != nil && spec.Name() != nil && spec.Name().Text() == name {
						return elem, clauseTypeOnly || spec.IsTypeOnly
					}
				}
			}
		}
	}
	return nil, false
}

func findDeclInVarDeclList(list *ast.VariableDeclarationList, name string) *ast.Node {
	if list == nil || list.Declarations == nil {
		return nil
	}
	for _, decl := range list.Declarations.Nodes {
		if decl != nil && decl.Kind == ast.KindVariableDeclaration && bindingDeclaresName(decl.Name(), name) {
			return decl
		}
	}
	return nil
}

func bindingDeclaresName(binding *ast.Node, name string) bool {
	if binding == nil {
		return false
	}
	found := false
	internalUtils.CollectBindingNames(binding, func(_ *ast.Node, n string) {
		if n == name {
			found = true
		}
	})
	return found
}

func isNodeCreateRequireCall(node *ast.Node, depth int) bool {
	if depth > 6 {
		return false
	}
	node = internalUtils.SkipAssertionsAndParens(node)
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	return isNodeCreateRequireFunction(node.AsCallExpression().Expression, depth)
}

func isNodeCreateRequireFunction(expr *ast.Node, depth int) bool {
	if depth > 6 {
		return false
	}
	expr = internalUtils.SkipAssertionsAndParens(expr)
	if expr == nil {
		return false
	}
	switch expr.Kind {
	case ast.KindIdentifier:
		localName := expr.AsIdentifier().Text
		decl, ambient := findInnermostDeclaration(expr, localName)
		if decl == nil || ambient {
			return false
		}
		switch decl.Kind {
		case ast.KindImportSpecifier:
			imported, _, ok := resolveModuleImportSpecifier(decl, nodeModuleSpecifiers, false)
			return ok && imported == "createRequire"
		case ast.KindVariableDeclaration:
			if !isImmutableConstDeclaration(decl) {
				return false
			}
			varDecl := decl.AsVariableDeclaration()
			if varDecl == nil || varDecl.Initializer == nil {
				return false
			}
			if varDecl.Name() != nil && varDecl.Name().Kind == ast.KindIdentifier && varDecl.Name().AsIdentifier().Text == localName {
				return isNodeCreateRequireFunction(varDecl.Initializer, depth+1)
			}
			// Destructured `const { createRequire } = require('node:module')`
			// or `const { createRequire } = mod` where `mod` is `node:module`.
			if varDecl.Name() != nil && varDecl.Name().Kind == ast.KindObjectBindingPattern {
				if binding := findBindingElementInPattern(varDecl.Name(), localName); binding != nil {
					if imported, _, ok := resolveModuleRequireBindingAtDepth(binding, nodeModuleSpecifiers, depth+1); ok && imported == "createRequire" {
						return true
					}
					if propName, ok := bindingElementStaticPropertyName(binding); ok && propName == "createRequire" {
						init := internalUtils.SkipAssertionsAndParens(varDecl.Initializer)
						if init != nil && init.Kind == ast.KindIdentifier {
							return isNodeModuleNamespaceIdent(init, depth+1)
						}
					}
				}
			}
		}
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		member, ok := internalUtils.AccessExpressionStaticName(expr)
		if !ok || member != "createRequire" {
			return false
		}
		obj := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(expr))
		if obj == nil {
			return false
		}
		if isModuleRequireCallModules(obj, nodeModuleSpecifiers, depth+1) {
			return true
		}
		if obj.Kind != ast.KindIdentifier {
			return false
		}
		return isNodeModuleNamespaceIdent(obj, depth+1)
	}
	return false
}

func bindingElementStaticPropertyName(elem *ast.Node) (string, bool) {
	if elem == nil || elem.Kind != ast.KindBindingElement {
		return "", false
	}
	b := elem.AsBindingElement()
	if b == nil || b.DotDotDotToken != nil {
		return "", false
	}
	if b.PropertyName != nil {
		return internalUtils.GetStaticPropertyName(b.PropertyName)
	}
	if b.Name() != nil && b.Name().Kind == ast.KindIdentifier {
		return b.Name().AsIdentifier().Text, true
	}
	return "", false
}

func isNodeModuleNamespaceIdent(obj *ast.Node, depth int) bool {
	if depth > 6 || obj == nil || obj.Kind != ast.KindIdentifier {
		return false
	}
	name := obj.AsIdentifier().Text
	decl, ambient := findInnermostDeclaration(obj, name)
	if decl == nil || ambient {
		return false
	}
	switch decl.Kind {
	case ast.KindNamespaceImport:
		imp := FindImportDeclaration(decl)
		return imp != nil && imp.ModuleSpecifier != nil && matchesModule(imp.ModuleSpecifier.Text(), nodeModuleSpecifiers)
	case ast.KindImportDeclaration:
		imp := decl.AsImportDeclaration()
		return imp != nil && imp.ModuleSpecifier != nil && matchesModule(imp.ModuleSpecifier.Text(), nodeModuleSpecifiers)
	case ast.KindImportSpecifier:
		imported, _, ok := resolveModuleImportSpecifier(decl, nodeModuleSpecifiers, false)
		return ok && imported == "default"
	case ast.KindImportEqualsDeclaration:
		return IsImportEqualsOfModules(decl, nodeModuleSpecifiers)
	case ast.KindVariableDeclaration:
		if !isImmutableConstDeclaration(decl) {
			return false
		}
		varDecl := decl.AsVariableDeclaration()
		if varDecl == nil || varDecl.Initializer == nil ||
			varDecl.Name() == nil || varDecl.Name().Kind != ast.KindIdentifier ||
			varDecl.Name().AsIdentifier().Text != name {
			return false
		}
		init := internalUtils.SkipAssertionsAndParens(varDecl.Initializer)
		if isModuleRequireCallModules(init, nodeModuleSpecifiers, depth+1) {
			return true
		}
		if init != nil && init.Kind == ast.KindIdentifier {
			return isNodeModuleNamespaceIdent(init, depth+1)
		}
	}
	return false
}

func findBindingElementInPattern(pattern *ast.Node, name string) *ast.Node {
	if pattern == nil || pattern.Kind != ast.KindObjectBindingPattern {
		return nil
	}
	bp := pattern.AsBindingPattern()
	if bp == nil || bp.Elements == nil {
		return nil
	}
	for _, elem := range bp.Elements.Nodes {
		if elem == nil || elem.Kind != ast.KindBindingElement {
			continue
		}
		b := elem.AsBindingElement()
		if b != nil && b.DotDotDotToken == nil && b.Name() != nil && b.Name().Kind == ast.KindIdentifier && b.Name().AsIdentifier().Text == name {
			return elem
		}
	}
	return nil
}

// NamedImportElements returns the specifiers of a named import, or nil for a
// type-only import or one without named bindings.
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

// ImportedSpecifierName returns the export name an import specifier binds, or
// "" for a type-only specifier.
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
