// Package scope reconstructs ESLint's lexical scope model on top of the tsgo
// AST. rslint has no eslint-scope equivalent, so rules that need to reason
// about "which declaration does this name belong to" share the scope tree
// built here instead of each re-deriving one.
//
// The model deliberately mirrors eslint-scope / typescript-eslint's
// scope-manager rather than tsgo's binder: block scopes, catch scopes,
// function-expression-name scopes, class scopes, class static blocks, and
// class field initializers are all materialized, and `var` hoisting targets
// the nearest function/module/global scope. Concepts rslint does not expose
// (for example `parserOptions.globalReturn`) remain unmodeled.
package scope

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// DefKind classifies what syntactic construct introduced a binding. It maps to
// eslint-scope's `Definition#type` closely enough for rules to branch on it.
type DefKind int

const (
	DefVariable       DefKind = iota // var/let/const binding, binding element
	DefParameter                     // function parameter
	DefFunctionName                  // FunctionDeclaration name (outer scope)
	DefFnExprName                    // FunctionExpression name (inner scope only)
	DefClassName                     // ClassDeclaration name (outer scope)
	DefClassInnerName                // Class name visible inside class scope
	DefImport                        // Import specifier / default / namespace
	DefCatch                         // Catch parameter
	DefType                          // Interface, type alias
	DefEnumName                      // Enum declaration
	DefEnumMember                    // Member of an enum declaration
	DefNamespaceName                 // Module/namespace declaration
	DefTypeParameter                 // Generic type parameter
)

// declaresValue reports whether the binding exists in value space — whether an
// expression can read it. Mirrors eslint-scope's `Definition#isVariableDefinition`.
func (k DefKind) declaresValue() bool {
	switch k {
	case DefType, DefTypeParameter:
		return false
	}
	return true
}

// declaresType reports whether the binding exists in type space — whether a
// type annotation can name it. Mirrors eslint-scope's
// `Definition#isTypeDefinition`.
func (k DefKind) declaresType() bool {
	switch k {
	case DefClassName, DefClassInnerName, DefImport, DefType,
		DefEnumName, DefEnumMember, DefNamespaceName, DefTypeParameter:
		return true
	}
	return false
}

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
	DeclareModifier  bool // declaration is in an ambient context (`declare` / .d.ts)
	// Anonymous marks a binding declared without an identifier — a
	// string-literal enum member (`enum E { "A" = 1 }`). eslint-scope gives
	// those a variable with an empty `identifiers` list, so they take part in
	// name resolution but rules never report at their declaration site.
	Anonymous bool

	Scope *Scope
}

// Kind identifies what construct a scope belongs to.
type Kind int

const (
	KindGlobal                Kind = iota
	KindFunction                   // function-like bodies & their parameters
	KindFunctionType               // TS function/constructor types and call/construct/method signatures
	KindFunctionExprName           // FunctionExpression's name binding
	KindBlock                      // { ... } / for-init / switch case / enum body
	KindCatch                      // catch clause
	KindClass                      // class body: type parameters & inner class name
	KindModule                     // TS namespace
	KindType                       // TS type alias / interface: type parameters
	KindClassStaticBlock           // `static { ... }`
	KindClassFieldInitializer      // the initializer expression of a class field
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
	// References holds the identifier references that appear directly in this
	// scope. Populated only when [Options.CollectReferences] is set.
	References []*Reference

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
// hoist target — eslint-scope's `Scope#variableScope`, which also models an
// execution context. Class static blocks and class field initializers are
// variable scopes, matching eslint-scope's `class-static-block` and
// `class-field-initializer` scope types.
func (s *Scope) VariableScope() *Scope {
	for current := s; current != nil; current = current.Parent {
		switch current.Kind {
		case KindFunction, KindFunctionType, KindModule, KindGlobal, KindClassStaticBlock, KindClassFieldInitializer:
			return current
		}
	}
	return nil
}

// Reference is one identifier that reads or writes a binding. Declaration
// identifiers are not references — eslint-scope models them as references with
// `init: true`, and every consumer of that flag skips them.
type Reference struct {
	// Identifier is the referencing identifier node.
	Identifier *ast.Node
	// From is the scope the reference is evaluated in.
	From *Scope
	// Declarations holds every declaration of the resolved binding, in
	// declaration order; nil when the name resolves to nothing in this file.
	Declarations []*Variable

	// isValueReference is true when the identifier can name a value — every
	// expression position, plus `typeof X` and `export { x }`.
	isValueReference bool
	// isTypeReference is true when the identifier can name a type — every type
	// position, plus `export { x }`, which exports whichever space `x` lives in.
	isTypeReference bool
}

// Resolved returns the first declaration of the binding this reference
// resolves to — eslint-scope's `reference.resolved.defs[0]` — or nil when the
// reference is unresolved.
func (r *Reference) Resolved() *Variable {
	if len(r.Declarations) == 0 {
		return nil
	}
	return r.Declarations[0]
}

// ResolvedIdentifier returns the first declaration identifier of the binding
// this reference resolves to — eslint-scope's
// reference.resolved.identifiers[0]. A binding may have definitions without
// identifiers (for example a string-literal enum member), so this can differ
// from [Reference.Resolved] and can be nil even when Resolved is not.
func (r *Reference) ResolvedIdentifier() *ast.Node {
	if r == nil {
		return nil
	}
	for _, declaration := range r.Declarations {
		if declaration != nil && !declaration.Anonymous && declaration.ID != nil &&
			declaration.ID.Kind == ast.KindIdentifier {
			return declaration.ID
		}
	}
	return nil
}

// IsValueReference reports whether the identifier can resolve in value space.
// Value and type references are independent: a reference may be both.
func (r *Reference) IsValueReference() bool {
	return r != nil && r.isValueReference
}

// IsTypeReference reports whether the identifier can resolve in type space.
// Value and type references are independent: a reference may be both.
func (r *Reference) IsTypeReference() bool {
	return r != nil && r.isTypeReference
}

// Options selects optional analysis passes.
type Options struct {
	// CollectReferences populates [Scope.References] and [Manager.References].
	// Rules that only inspect declarations leave it off so the extra
	// identifier walk and resolution pass are skipped.
	CollectReferences bool
	// ReferenceNames limits collected references to these exact binding names.
	// A nil map collects every name. Scope construction and declaration
	// collection remain complete, so filtering cannot change how retained
	// references resolve.
	ReferenceNames map[string]struct{}
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
	// References lists every reference in the file, in the order the builder
	// discovered them. Populated only when [Options.CollectReferences] is set.
	References []*Reference
}

// Build analyzes `sf` and returns its scope tree.
func Build(sf *ast.SourceFile, opts Options) *Manager {
	m := &Manager{SourceFile: sf}
	b := &builder{
		manager:           m,
		collectReferences: opts.CollectReferences,
		referenceNames:    opts.ReferenceNames,
	}
	m.Global = b.buildProgram(sf)
	if opts.CollectReferences {
		m.resolveReferences()
	}
	return m
}

// resolveReferences links each reference to the innermost enclosing scope that
// declares its name in a way the reference can bind to.
func (m *Manager) resolveReferences() {
	for _, ref := range m.References {
		name := ref.Identifier.Text()
		for current := ref.From; current != nil; current = current.Parent {
			declarations := current.ByName[name]
			if len(declarations) == 0 || !current.binds(ref, declarations) {
				continue
			}
			ref.Declarations = declarations
			break
		}
	}
}

// binds reports whether `declarations` — every declaration of the reference's
// name in this scope — can resolve `ref`. When it can't, the reference carries
// on to the enclosing scope, as eslint-scope's `delegateToUpperScope` does.
func (s *Scope) binds(ref *Reference, declarations []*Variable) bool {
	if !s.bindsBeforeBody(ref, declarations) {
		return false
	}
	// A name lives in value space, type space, or both: `type X` never answers
	// an expression's read of `X`, and `const X` never answers a type
	// annotation naming `X`.
	for _, declaration := range declarations {
		if (ref.isValueReference && declaration.Kind.declaresValue()) ||
			(ref.isTypeReference && declaration.Kind.declaresType()) {
			return true
		}
	}
	return false
}

// bindsBeforeBody reports whether a reference in a parameter default can see
// these declarations. A default is evaluated before the function body's
// lexical environment exists, so `function f(a = x) { const x = 2; }` reads the
// outer `x` — eslint-scope's `FunctionScope#isValidResolution`.
func (s *Scope) bindsBeforeBody(ref *Reference, declarations []*Variable) bool {
	if s.Kind != KindFunction || s.Block == nil {
		return true
	}
	body := s.Block.Body()
	if body == nil || ref.Identifier.Pos() >= body.Pos() {
		return true
	}
	for _, declaration := range declarations {
		if declaration.ID.Pos() < body.Pos() {
			return true
		}
	}
	return false
}
