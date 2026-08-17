// Package scope reconstructs ESLint's lexical scope model on top of the tsgo
// AST. rslint has no eslint-scope equivalent, so rules that need to reason
// about "which declaration does this name belong to" share the scope tree
// built here instead of each re-deriving one.
//
// The model deliberately mirrors eslint-scope / typescript-eslint's
// scope-manager rather than tsgo's binder: block scopes, catch scopes,
// function-expression-name scopes, and class scopes are all materialized, and
// `var` hoisting targets the nearest function/module/global scope. Concepts
// rslint does not expose (for example `parserOptions.globalReturn`) remain
// unmodeled.
package scope

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// DefKind classifies what syntactic construct introduced a binding. It maps to
// eslint-scope's `Definition#type` closely enough for rules to branch on it.
type DefKind int

const (
	DefVariable       DefKind = iota // var/let/const binding, binding element, enum member
	DefParameter                     // function parameter
	DefFunctionName                  // FunctionDeclaration name (outer scope)
	DefFnExprName                    // FunctionExpression name (inner scope only)
	DefClassName                     // ClassDeclaration name (outer scope)
	DefClassInnerName                // Class name visible inside class scope
	DefImport                        // Import specifier / default / namespace
	DefCatch                         // Catch parameter
	DefType                          // Interface, type alias
	DefEnumName                      // Enum declaration
	DefNamespaceName                 // Module/namespace declaration
	DefTypeParameter                 // Generic type parameter
)

// Variable is a single declaration site. eslint-scope merges every declaration
// of a name within one scope into a single `Variable` carrying several `defs`;
// here each declaration is its own Variable and the merged view is
// [Scope.Declarations], which returns them in declaration order.
type Variable struct {
	Name    string
	ID      *ast.Node // identifier node of this declaration
	DefNode *ast.Node // declaration node (VariableDeclaration, FunctionDeclaration, ...)
	Parent  *ast.Node // parent of DefNode (for parsed import detection, parameter typing, etc.)
	Kind    DefKind

	IsValueBinding   bool // runtime value vs. type-only
	IsTypeOnlyImport bool // ImportSpecifier with `type` modifier
	DeclareModifier  bool // `declare` modifier (.d.ts handling)

	Scope *Scope
}

// Kind identifies what construct a scope belongs to.
type Kind int

const (
	KindGlobal           Kind = iota
	KindFunction              // function-like bodies & their parameters
	KindFunctionExprName      // FunctionExpression's name binding
	KindBlock                 // { ... } / for-init / switch case / enum body
	KindCatch                 // catch clause
	KindClass                 // class body: type parameters & inner class name
	KindModule                // TS namespace
	KindType                  // TS type alias / interface / function type: type parameters
)

// Scope is one node of the scope tree.
type Scope struct {
	Kind   Kind
	Block  *ast.Node
	Parent *Scope

	// Vars lists every declaration in this scope in declaration order.
	Vars []*Variable
	// ByName groups Vars by binding name, preserving declaration order.
	ByName map[string][]*Variable

	// GlobalAugmentation is true inside a `declare global { ... }` chain.
	GlobalAugmentation bool
}

func newScope(kind Kind, block *ast.Node, parent *Scope) *Scope {
	return &Scope{
		Kind:   kind,
		Block:  block,
		Parent: parent,
		ByName: map[string][]*Variable{},
	}
}

// Add records a declaration in this scope.
func (s *Scope) Add(v *Variable) {
	v.Scope = s
	s.Vars = append(s.Vars, v)
	s.ByName[v.Name] = append(s.ByName[v.Name], v)
}

// Declarations returns every declaration of `name` in this scope, in
// declaration order — eslint-scope's `Variable#defs` for that name.
func (s *Scope) Declarations(name string) []*Variable {
	return s.ByName[name]
}

// VariableScope returns the nearest ancestor (or self) that acts as a `var`
// hoist target — eslint-scope's `Scope#variableScope`: function-like scopes,
// module scopes, and the global scope.
func (s *Scope) VariableScope() *Scope {
	for current := s; current != nil; current = current.Parent {
		switch current.Kind {
		case KindFunction, KindModule, KindGlobal:
			return current
		}
	}
	return nil
}

// Manager owns a built scope tree.
type Manager struct {
	SourceFile *ast.SourceFile
	// Global is the file's outermost scope. rslint collapses ESLint's module
	// scope into it.
	Global *Scope
	// Scopes lists every scope in creation order, which is a pre-order walk of
	// the scope tree.
	Scopes []*Scope
}

// Build analyzes `sf` and returns its scope tree.
func Build(sf *ast.SourceFile) *Manager {
	m := &Manager{SourceFile: sf}
	b := &builder{manager: m}
	m.Global = b.buildProgram(sf)
	return m
}
