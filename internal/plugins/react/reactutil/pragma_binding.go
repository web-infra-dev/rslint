package reactutil

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
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
// `pragma.toLocaleLowerCase()` semantic. `tc` may be nil — when no
// TypeChecker is available the function falls back to a syntax-only scan. It
// checks enclosing block/function bindings before the SourceFile-wide import
// scan so a local binding cannot masquerade as an imported pragma helper.
func IsDestructuredFromPragmaImport(ident *ast.Node, pragma string, tc *checker.Checker) bool {
	if ident == nil || ident.Kind != ast.KindIdentifier {
		return false
	}
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	pragmaLower := ecmascript.StringToLowerCase(pragma)

	if tc == nil {
		if hasEnclosingBinding(ident, ident.AsIdentifier().Text) {
			return false
		}
		return findPragmaBindingByName(getSourceFileNode(ident), ident.AsIdentifier().Text, pragma, pragmaLower)
	}

	symbol := tc.GetSymbolAtLocation(ident)
	if symbol == nil {
		return false
	}

	// Pick the most relevant declaration. Upstream walks `latestDef` —
	// for value bindings ValueDeclaration is the right one; for
	// ImportSpecifier (which has no Initializer of its own), upstream
	// walks `latestDef.parent.type === 'ImportDeclaration'`. We mirror
	// by trying ValueDeclaration first then Declarations[0].
	var decl *ast.Node
	if symbol.ValueDeclaration != nil {
		decl = symbol.ValueDeclaration
	} else if len(symbol.Declarations) > 0 {
		decl = symbol.Declarations[0]
	}
	if decl == nil {
		return false
	}

	// 1) Named import: `import { memo } from 'react'` — declaration is
	//    an ImportSpecifier (or ImportClause for default imports, but
	//    bare callee `memo` won't bind to a default).
	if decl.Kind == ast.KindImportSpecifier {
		// Walk up: ImportSpecifier → NamedImports → ImportClause →
		// ImportDeclaration.
		for p := decl.Parent; p != nil; p = p.Parent {
			if p.Kind == ast.KindImportDeclaration {
				ms := p.AsImportDeclaration().ModuleSpecifier
				if ms != nil && ms.Kind == ast.KindStringLiteral &&
					ms.Text() == pragmaLower {
					return true
				}
				return false
			}
		}
		return false
	}

	// 2) BindingElement (object/array destructure): `const { memo } = React`
	//    → declaration is BindingElement; walk up to VariableDeclaration
	//    and inspect its Initializer.
	if decl.Kind == ast.KindBindingElement {
		varDecl := findEnclosingVariableDeclaration(decl)
		if varDecl == nil {
			return false
		}
		init := varDecl.AsVariableDeclaration().Initializer
		return initializerMatchesPragma(init, pragma, pragmaLower)
	}

	// 3) VariableDeclaration: `const memo = React.memo` /
	//    `const memo = require('react').memo`
	if decl.Kind == ast.KindVariableDeclaration {
		init := decl.AsVariableDeclaration().Initializer
		return initializerMatchesPragma(init, pragma, pragmaLower)
	}

	return false
}

func hasEnclosingBinding(ident *ast.Node, name string) bool {
	if ident == nil || name == "" {
		return false
	}
	for current := ident.Parent; current != nil; current = current.Parent {
		if params := functionParameters(current); params != nil {
			for _, param := range params {
				if param == nil {
					continue
				}
				parameter := param.AsParameterDeclaration()
				if parameter != nil && bindingPatternBindsName(parameter.Name(), name) {
					return true
				}
			}
		}
		switch current.Kind {
		case ast.KindBlock, ast.KindCaseBlock, ast.KindModuleBlock:
			if blockDeclaresBinding(current, name) {
				return true
			}
		}
	}
	return false
}

func functionParameters(node *ast.Node) []*ast.Node {
	if node == nil {
		return nil
	}
	if params := FunctionParameters(node); params != nil {
		return params
	}
	switch node.Kind {
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().Parameters.Nodes
	case ast.KindGetAccessor:
		return node.AsGetAccessorDeclaration().Parameters.Nodes
	case ast.KindSetAccessor:
		return node.AsSetAccessorDeclaration().Parameters.Nodes
	}
	return nil
}

func bindingPatternBindsName(node *ast.Node, name string) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindIdentifier {
		return node.AsIdentifier().Text == name
	}
	if node.Kind == ast.KindObjectBindingPattern {
		return objectBindingPatternBindsName(node, name)
	}
	return false
}

func blockDeclaresBinding(block *ast.Node, name string) bool {
	var found bool
	block.ForEachChild(func(statement *ast.Node) bool {
		if statement == nil {
			return false
		}
		switch statement.Kind {
		case ast.KindVariableStatement:
			declarations := statement.AsVariableStatement().DeclarationList
			if declarations == nil || declarations.AsVariableDeclarationList() == nil {
				return false
			}
			for _, declaration := range declarations.AsVariableDeclarationList().Declarations.Nodes {
				if declaration != nil && bindingPatternBindsName(declaration.Name(), name) {
					found = true
					return true
				}
			}
		case ast.KindFunctionDeclaration, ast.KindClassDeclaration:
			if bindingPatternBindsName(statement.Name(), name) {
				found = true
				return true
			}
		}
		return false
	})
	return found
}

// getSourceFileNode walks up from `node` to its enclosing SourceFile,
// returning it as an `*ast.Node`, or nil when no SourceFile ancestor is
// found (extremely unlikely outside of synthesized nodes).
func getSourceFileNode(node *ast.Node) *ast.Node {
	sf := ast.GetSourceFileOfNode(node)
	if sf == nil {
		return nil
	}
	return sf.AsNode()
}

// findPragmaBindingByName is the syntax-only fallback for
// `IsDestructuredFromPragmaImport` when no TypeChecker is available. It
// scans the SourceFile rooted at `root` for any declaration that
// introduces a binding named `name` whose source is the pragma module:
//
//   - `import { name } from '<pragma>'`
//   - `import { x as name } from '<pragma>'` (renamed import — local
//     binding is `name`)
//   - `const { name } = <pragma>` / `const { name } = require('<pragma>')`
//   - `const name = <pragma>.name` / `const name = require('<pragma>').name`
//   - `const { x: name } = <pragma>` / require — destructure-with-rename
//
// Walks the entire SourceFile rather than tracking lexical scope. This
// is a deliberate trade-off: shadowing in inner scopes (e.g. a deeply
// nested `function memo() {}` overriding a top-level
// `import { memo } from 'react'`) is NOT modeled — but the recognition
// only matters for bare callees that already passed name + non-shadow
// checks at the call site, which makes shadowing edge-cases vanish in
// practice.
func findPragmaBindingByName(root *ast.Node, name string, pragma string, pragmaLower string) bool {
	if root == nil || name == "" {
		return false
	}
	var found bool
	var visit func(n *ast.Node)
	visit = func(n *ast.Node) {
		if found || n == nil {
			return
		}
		switch n.Kind {
		case ast.KindImportDeclaration:
			if importDeclBindsNameFromPragma(n, name, pragmaLower) {
				found = true
				return
			}
		case ast.KindVariableDeclaration:
			if variableDeclBindsNameFromPragma(n, name, pragma, pragmaLower) {
				found = true
				return
			}
		}
		n.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found
		})
	}
	visit(root)
	return found
}

// importDeclBindsNameFromPragma reports whether `decl`
// (an ImportDeclaration) introduces a local binding called `name` from
// the module whose lowercased specifier equals `pragmaLower`. Handles
// both plain (`import { name } from '...'`) and renamed
// (`import { x as name } from '...'`) named imports — the local binding
// is the second identifier, which is what we match against `name`.
func importDeclBindsNameFromPragma(decl *ast.Node, name string, pragmaLower string) bool {
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
	if ic.NamedBindings == nil || ic.NamedBindings.Kind != ast.KindNamedImports {
		// Default import / namespace import don't bind `name` directly.
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
	init = SkipExpressionWrappers(init)
	if init == nil || init.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	pa := init.AsPropertyAccessExpression()
	prop := pa.Name()
	if prop == nil || prop.Kind != ast.KindIdentifier || prop.AsIdentifier().Text != name {
		return false
	}
	obj := SkipExpressionWrappers(pa.Expression)
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
	init = SkipExpressionWrappers(init)

	// `init` is the pragma identifier itself (`= React`).
	if init.Kind == ast.KindIdentifier && init.AsIdentifier().Text == pragma {
		return true
	}

	// `init` is `pragma.something` — `= React.memo`.
	if init.Kind == ast.KindPropertyAccessExpression {
		obj := SkipExpressionWrappers(init.AsPropertyAccessExpression().Expression)
		if obj.Kind == ast.KindIdentifier && obj.AsIdentifier().Text == pragma {
			return true
		}
		// `init` is `require('react').memo` — member access on a
		// require call.
		if obj.Kind == ast.KindCallExpression && isRequireCallOfPragma(obj, pragmaLower) {
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
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	c := call.AsCallExpression()
	callee := SkipExpressionWrappers(c.Expression)
	if callee == nil || callee.Kind != ast.KindIdentifier ||
		callee.AsIdentifier().Text != "require" {
		return false
	}
	if c.Arguments == nil || len(c.Arguments.Nodes) == 0 {
		return false
	}
	arg := SkipExpressionWrappers(c.Arguments.Nodes[0])
	if arg == nil || arg.Kind != ast.KindStringLiteral {
		return false
	}
	return arg.AsStringLiteral().Text == pragmaLower
}
