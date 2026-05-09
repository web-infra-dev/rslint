package reactutil

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
)

// ResolveIdentifierInitializer returns the value-side AST node that an
// Identifier reference is bound to, or nil when the binding cannot be
// determined.
//
//   - When `tc` is non-nil, asks the TypeChecker for the resolved symbol's
//     ValueDeclaration, then returns that declaration's `.Initializer`
//     (only the const/let/var case — class / function declarations have
//     no `Initializer` and aren't useful for JSX-return resolution). This
//     is upstream-equivalent to `findVariableByName` because the TS
//     resolver already follows the full lexical scope chain.
//
//   - When `tc` is nil, falls back to scanning enclosing Block /
//     SourceFile / ModuleBlock / CaseBlock statements for a
//     `VariableStatement` declaring `name` — catches the common
//     same-block initializer case without scope analysis.
func ResolveIdentifierInitializer(ident *ast.Node, tc *checker.Checker) *ast.Node {
	if ident == nil || ident.Kind != ast.KindIdentifier {
		return nil
	}
	if tc != nil {
		if init := resolveIdentifierViaChecker(ident, tc); init != nil {
			return init
		}
		// Fall through: TypeChecker may not resolve the binding (e.g. a
		// type-only declaration in a JS file, or a synthesized symbol).
		// The local-block scan is a strict subset, so trying it as a
		// safety net costs nothing.
	}
	return lookupLocalInitializer(ident)
}

// resolveIdentifierViaChecker resolves an Identifier through the
// TypeChecker. Returns the initializer of the resolved
// VariableDeclaration, or nil when the symbol's value declaration is not a
// VariableDeclaration with an Initializer.
//
// All `checker.Checker` access on rslint runs without `--type-check` is
// nil; this function MUST be defensive even though its callers already
// gate on a non-nil `tc`. The double guard keeps the file safe to call
// from any future site that forgets the gate.
func resolveIdentifierViaChecker(ident *ast.Node, tc *checker.Checker) *ast.Node {
	if tc == nil || ident == nil {
		return nil
	}
	symbol := tc.GetSymbolAtLocation(ident)
	if symbol == nil {
		return nil
	}
	// `ValueDeclaration` is the symbol's primary declaration site; for
	// `const x = <div/>` it's the VariableDeclaration. When ValueDeclaration
	// is absent (interfaces, type aliases, ambient symbols) we explicitly
	// don't try to walk `Declarations` — those don't have a JSX value.
	decl := symbol.ValueDeclaration
	if decl == nil {
		// Fall back to Declarations[0] when ValueDeclaration is missing
		// but the symbol still has a concrete declaration (e.g. some
		// shorthand-property bindings).
		if len(symbol.Declarations) == 0 {
			return nil
		}
		decl = symbol.Declarations[0]
	}
	if decl == nil || decl.Kind != ast.KindVariableDeclaration {
		return nil
	}
	return decl.AsVariableDeclaration().Initializer
}

// lookupLocalInitializer mirrors upstream `variableUtil.findVariableByName`'s
// happy path for the cases this rule cares about: a const/let/var binding
// whose initializer is a JSX-or-createElement expression, declared in the
// same enclosing function/program. Walks lexically up the parent chain
// looking for a Block / SourceFile that contains a `VariableStatement`
// declaring `name` with a non-nil initializer; returns the initializer or
// nil when no such declaration is reachable.
//
// We deliberately do NOT re-implement a full scope manager — the tradeoff
// is that re-bindings (e.g. `let x = <div/>; x = 1; return x`) and
// destructuring patterns are not resolved, matching the conservative subset
// of upstream's behavior that the no-unstable-nested-components rule
// actually exercises in its tests.
func lookupLocalInitializer(ident *ast.Node) *ast.Node {
	if ident == nil || ident.Kind != ast.KindIdentifier {
		return nil
	}
	name := ident.AsIdentifier().Text
	for cur := ident.Parent; cur != nil; cur = cur.Parent {
		switch cur.Kind {
		case ast.KindBlock, ast.KindSourceFile, ast.KindCaseBlock, ast.KindModuleBlock:
			if init := findInitializerInStatements(cur, name); init != nil {
				return init
			}
		}
	}
	return nil
}

// findInitializerInStatements scans the statements of a Block / SourceFile
// (or anything with a `Statements` field exposed via ForEachChild) for a
// `VariableStatement` declaring `name` with a direct initializer.
func findInitializerInStatements(scope *ast.Node, name string) *ast.Node {
	if scope == nil {
		return nil
	}
	var found *ast.Node
	scope.ForEachChild(func(stmt *ast.Node) bool {
		if found != nil || stmt == nil {
			return false
		}
		var declList *ast.Node
		switch stmt.Kind {
		case ast.KindVariableStatement:
			declList = stmt.AsVariableStatement().DeclarationList
		}
		if declList == nil {
			return false
		}
		decls := declList.AsVariableDeclarationList()
		if decls == nil || decls.Declarations == nil {
			return false
		}
		for _, d := range decls.Declarations.Nodes {
			if d == nil || d.Kind != ast.KindVariableDeclaration {
				continue
			}
			vd := d.AsVariableDeclaration()
			if vd.Name() == nil || vd.Name().Kind != ast.KindIdentifier {
				continue
			}
			if vd.Name().AsIdentifier().Text != name {
				continue
			}
			if vd.Initializer != nil {
				found = vd.Initializer
				return true
			}
		}
		return false
	})
	return found
}
