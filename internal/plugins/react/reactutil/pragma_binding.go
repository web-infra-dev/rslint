package reactutil

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

// IsDestructuredFromPragmaImport mirrors upstream eslint-plugin-react's
// `lib/util/isDestructuredFromPragmaImport.js`: reports whether the
// Identifier `ident` (a bare callee like `memo`) was bound from the
// pragma module. Returns true when ident's local binding originated from
// any of:
//
//   - `import { memo } from 'react'` (named import)
//   - `import { memo as m } from 'react'` (named-import rename — checks
//     the imported name, not the local alias)
//   - `import * as React from 'react'`'s namespace + `const memo = React.memo`
//   - `const { memo } = React` (object destructure of the pragma binding)
//   - `const memo = React.memo` (member access via pragma binding)
//   - `const { memo } = require('react')` (require destructure)
//   - `const memo = require('react').memo` (require member access)
//
// `pragma` is the React pragma name (e.g. "React") — the comparison
// against ImportDeclaration / require argument uses
// `ecmascript.StringToLowerCase(pragma)` to match upstream's
// `pragma.toLocaleLowerCase()` semantic. `tc` may be nil — when no TypeChecker
// is available the function uses rslint's shared lexical scope model so every
// local binding shape is resolved before checking its import origin.
func IsDestructuredFromPragmaImport(ident *ast.Node, pragma string, tc *checker.Checker) bool {
	if ident == nil || ident.Kind != ast.KindIdentifier {
		return false
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	pragmaLower := ecmascript.StringToLowerCase(pragma)

	if tc == nil {
		return sourceOnlyPragmaBinding(ident, pragma, pragmaLower)
	}

	symbol := tc.GetSymbolAtLocation(ident)
	if symbol == nil {
		return false
	}

	return isDestructuredFromPragmaSymbol(symbol, pragma, pragmaLower)
}

func isDestructuredFromPragmaDeclaration(decl *ast.Node, pragma, pragmaLower string) bool {
	if decl == nil {
		return false
	}
	if decl.Kind == ast.KindImportDeclaration {
		return importDeclarationFromPragma(decl, pragmaLower)
	}
	if decl.Kind == ast.KindImportSpecifier || decl.Kind == ast.KindImportClause ||
		decl.Kind == ast.KindNamespaceImport {
		for parent := decl.Parent; parent != nil; parent = parent.Parent {
			if parent.Kind == ast.KindImportDeclaration {
				return importDeclarationFromPragma(parent, pragmaLower)
			}
		}
		return false
	}
	return isDestructuredFromPragmaDefinition(decl, pragma, pragmaLower)
}

func isDestructuredFromPragmaDefinition(decl *ast.Node, pragma, pragmaLower string) bool {
	switch decl.Kind {
	case ast.KindBindingElement:
		varDecl := findEnclosingVariableDeclaration(decl)
		if varDecl == nil {
			return false
		}
		return initializerMatchesPragma(varDecl.AsVariableDeclaration().Initializer, pragma, pragmaLower)
	case ast.KindVariableDeclaration:
		return initializerMatchesPragma(decl.AsVariableDeclaration().Initializer, pragma, pragmaLower)
	}
	return false
}

func importDeclarationFromPragma(decl *ast.Node, pragmaLower string) bool {
	if decl == nil || decl.Kind != ast.KindImportDeclaration {
		return false
	}
	moduleSpecifier := decl.AsImportDeclaration().ModuleSpecifier
	return moduleSpecifier != nil && moduleSpecifier.Kind == ast.KindStringLiteral &&
		moduleSpecifier.Text() == pragmaLower
}

func isDestructuredFromPragmaSymbol(symbol *ast.Symbol, pragma, pragmaLower string) bool {
	if symbol == nil {
		return false
	}

	// This checker-only fallback has no reference scope with which to model
	// ESLint's definitions. Preserve the declaration order supplied by the
	// checker instead of comparing Pos values from potentially different files.
	for i := len(symbol.Declarations) - 1; i >= 0; i-- {
		if declaration := symbol.Declarations[i]; declaration != nil {
			return isDestructuredFromPragmaDeclaration(declaration, pragma, pragmaLower)
		}
	}
	if symbol.ValueDeclaration != nil {
		return isDestructuredFromPragmaDeclaration(symbol.ValueDeclaration, pragma, pragmaLower)
	}
	return false
}

func sourceOnlyPragmaBinding(ident *ast.Node, pragma, pragmaLower string) bool {
	sourceFile := ast.GetSourceFileOfNode(ident)
	if sourceFile == nil {
		return false
	}
	manager := scope.Build(sourceFile, scope.Options{
		CollectReferences: true,
		ReferenceNames:    map[string]struct{}{ident.AsIdentifier().Text: {}},
	})
	for _, reference := range manager.References {
		if reference.Identifier != ident {
			continue
		}
		var declaration *scope.Variable
		for i := len(reference.Declarations) - 1; i >= 0; i-- {
			if reference.Declarations[i] != nil {
				declaration = reference.Declarations[i]
				break
			}
		}
		if declaration == nil {
			return false
		}
		if declaration.Kind == scope.DefImport {
			if declaration.DefNode == nil || declaration.DefNode.Kind != ast.KindImportDeclaration {
				return false
			}
			return importDeclBindsNameFromPragma(declaration.DefNode, ident.AsIdentifier().Text, pragmaLower)
		}
		if declaration.Kind != scope.DefVariable {
			return false
		}
		defNode := declaration.DefNode
		if defNode != nil && defNode.Kind == ast.KindBindingElement {
			defNode = findEnclosingVariableDeclaration(defNode)
		}
		return variableDeclBindsNameFromPragma(defNode, ident.AsIdentifier().Text, pragma, pragmaLower)
	}
	return false
}

// importDeclBindsNameFromPragma reports whether `decl`
// (an ImportDeclaration) introduces a local binding called `name` from
// the module whose lowercased specifier equals `pragmaLower`. Handles
// both plain (`import { name } from '...'`) and renamed
// (`import { x as name } from '...'`) named imports — the local binding
// is the second identifier, which is what we match against `name`.
func importDeclBindsNameFromPragma(decl *ast.Node, name string, pragmaLower string) bool {
	if decl == nil || decl.Kind != ast.KindImportDeclaration {
		return false
	}
	id := decl.AsImportDeclaration()
	if id.ModuleSpecifier == nil || id.ModuleSpecifier.Kind != ast.KindStringLiteral {
		return false
	}
	if id.ModuleSpecifier.Text() != pragmaLower {
		return false
	}
	if id.ImportClause == nil {
		return false
	}
	ic := id.ImportClause.AsImportClause()
	if ic.Name() != nil && ic.Name().Kind == ast.KindIdentifier && ic.Name().AsIdentifier().Text == name {
		return true
	}
	if ic.NamedBindings == nil || ic.NamedBindings.Kind != ast.KindNamedImports {
		if ic.NamedBindings != nil && ic.NamedBindings.Kind == ast.KindNamespaceImport {
			ns := ic.NamedBindings.AsNamespaceImport()
			return ns != nil && ns.Name() != nil && ns.Name().Kind == ast.KindIdentifier &&
				ns.Name().AsIdentifier().Text == name
		}
		return false
	}
	ni := ic.NamedBindings.AsNamedImports()
	if ni.Elements == nil {
		return false
	}
	for _, spec := range ni.Elements.Nodes {
		// ImportSpecifier.Name() returns the local binding identifier
		// (post-rename in `{ x as y }`). That's what shadows scope and
		// what we should compare against `name`.
		local := spec.Name()
		if local != nil && local.Kind == ast.KindIdentifier && local.AsIdentifier().Text == name {
			return true
		}
	}
	return false
}

// variableDeclBindsNameFromPragma reports whether `decl`
// (a VariableDeclaration) introduces a local binding called `name`
// whose value originates from the pragma module. Recognized shapes:
//
//   - `const name = <pragma>.name` / `const name = require('<pragma>').name`
//   - `const { name } = <pragma>` / `const { name } = require('<pragma>')`
//   - `const { x: name } = <pragma>` / `const { x: name } = require('<pragma>')`
func variableDeclBindsNameFromPragma(decl *ast.Node, name, pragma, pragmaLower string) bool {
	vd := decl.AsVariableDeclaration()
	if vd.Initializer == nil {
		return false
	}
	bindingName := vd.Name()
	if bindingName == nil {
		return false
	}
	switch bindingName.Kind {
	case ast.KindIdentifier:
		// `const name = ...` — local binding is `bindingName.Text`.
		if bindingName.AsIdentifier().Text != name {
			return false
		}
		// Initializer must be `<pragma>.name` or `require('<pragma>').name`.
		return initializerIsPragmaMember(vd.Initializer, name, pragma, pragmaLower)
	case ast.KindObjectBindingPattern:
		// `const { name } = ...` or `const { x: name } = ...`. Element
		// match: an ObjectBindingPattern element introduces `name` if
		// either its propertyName is unset and its bindingName.Text is
		// `name`, OR its bindingName.Text is `name` (the alias side).
		if !objectBindingPatternBindsName(bindingName, name) {
			return false
		}
		return initializerMatchesPragma(vd.Initializer, pragma, pragmaLower)
	}
	return false
}

// objectBindingPatternBindsName reports whether any element of the
// ObjectBindingPattern introduces a local binding called `name`. The
// local binding is the BindingElement.Name() — for `{ x: name }`,
// PropertyName is `x` and Name is `name`; we always compare against
// Name. Nested patterns are not recursed into (they don't apply to
// pragma-import shapes).
func objectBindingPatternBindsName(pat *ast.Node, name string) bool {
	obp := pat.AsBindingPattern()
	if obp == nil || obp.Elements == nil {
		return false
	}
	for _, el := range obp.Elements.Nodes {
		be := el.AsBindingElement()
		local := be.Name()
		if local != nil && local.Kind == ast.KindIdentifier && local.AsIdentifier().Text == name {
			return true
		}
	}
	return false
}

// initializerIsPragmaMember reports whether `init` is `<pragma>.<name>` or
// `require('<pragma>').<name>` — the two member-access shapes that
// introduce a `name` binding pulled from the pragma module without
// going through a destructure pattern.
func initializerIsPragmaMember(init *ast.Node, name, pragma, pragmaLower string) bool {
	init = utils.ESTreeRuntimeExpression(init)
	if init == nil || (init.Kind != ast.KindPropertyAccessExpression && init.Kind != ast.KindElementAccessExpression) ||
		ast.IsOptionalChain(init) {
		return false
	}
	var obj *ast.Node
	if init.Kind == ast.KindPropertyAccessExpression {
		pa := init.AsPropertyAccessExpression()
		prop := pa.Name()
		if prop == nil || prop.Kind != ast.KindIdentifier || prop.AsIdentifier().Text != name {
			return false
		}
		obj = pa.Expression
	} else {
		obj = init.AsElementAccessExpression().Expression
	}
	obj = utils.ESTreeRuntimeExpression(obj)
	if obj == nil {
		return false
	}
	if obj.Kind == ast.KindIdentifier && obj.AsIdentifier().Text == pragma {
		return true
	}
	if obj.Kind == ast.KindCallExpression && isRequireCallOfPragma(obj, pragmaLower) {
		return true
	}
	return false
}

// findEnclosingVariableDeclaration walks up from a BindingElement to its
// enclosing VariableDeclaration, or nil when not found (e.g. parameter
// bindings, which are not pragma imports).
func findEnclosingVariableDeclaration(node *ast.Node) *ast.Node {
	for p := node.Parent; p != nil; p = p.Parent {
		switch p.Kind {
		case ast.KindVariableDeclaration:
			return p
		case ast.KindParameter, ast.KindFunctionDeclaration,
			ast.KindArrowFunction, ast.KindFunctionExpression,
			ast.KindMethodDeclaration:
			return nil
		}
	}
	return nil
}

// initializerMatchesPragma reports whether the given initializer
// expression evaluates to the pragma binding (or to a property of it).
// Mirrors the four init shapes upstream's helper inspects.
func initializerMatchesPragma(init *ast.Node, pragma, pragmaLower string) bool {
	if init == nil {
		return false
	}
	init = utils.ESTreeRuntimeExpression(init)

	// `init` is the pragma identifier itself (`= React`).
	if init.Kind == ast.KindIdentifier && init.AsIdentifier().Text == pragma {
		return true
	}

	// `init` is `pragma.something` — `= React.memo`.
	if init.Kind == ast.KindPropertyAccessExpression || init.Kind == ast.KindElementAccessExpression {
		if ast.IsOptionalChain(init) {
			return false
		}
		var obj *ast.Node
		if init.Kind == ast.KindPropertyAccessExpression {
			obj = init.AsPropertyAccessExpression().Expression
		} else {
			obj = init.AsElementAccessExpression().Expression
		}
		obj = utils.ESTreeRuntimeExpression(obj)
		if obj != nil && obj.Kind == ast.KindIdentifier && obj.AsIdentifier().Text == pragma {
			return true
		}
		// `init` is `require('react').memo` — member access on a
		// require call.
		if obj != nil && obj.Kind == ast.KindCallExpression && isRequireCallOfPragma(obj, pragmaLower) {
			return true
		}
	}

	// `init` is `require('react')` directly (destructure case).
	if init.Kind == ast.KindCallExpression && isRequireCallOfPragma(init, pragmaLower) {
		return true
	}

	return false
}

// isRequireCallOfPragma reports whether `call` is `require('<pragmaLower>')`.
// Upstream's helper checks `callee.name === 'require'` and
// `arguments[0].value === pragma.toLocaleLowerCase()`.
func isRequireCallOfPragma(call *ast.Node, pragmaLower string) bool {
	if call == nil || call.Kind != ast.KindCallExpression || ast.IsOptionalChain(call) {
		return false
	}
	c := call.AsCallExpression()
	callee := utils.ESTreeRuntimeExpression(c.Expression)
	if callee == nil || callee.Kind != ast.KindIdentifier ||
		callee.AsIdentifier().Text != "require" {
		return false
	}
	if c.Arguments == nil || len(c.Arguments.Nodes) == 0 {
		return false
	}
	arg := utils.ESTreeRuntimeExpression(c.Arguments.Nodes[0])
	if arg == nil || arg.Kind != ast.KindStringLiteral {
		return false
	}
	return arg.AsStringLiteral().Text == pragmaLower
}
