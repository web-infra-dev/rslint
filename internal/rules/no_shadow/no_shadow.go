package no_shadow

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// ecmaScriptGlobals lists ECMAScript built-in globals referenced by the
// `builtinGlobals` option. We pair this with a TypeChecker-based default-
// library scan so that environment-specific globals (DOM, Node, …) are also
// picked up when type information is available. The hard-coded list is
// necessary because once a user writes `var Object = 0` at module scope, the
// TypeChecker resolves `Object` to the local binding and the default-library
// match is lost.
var ecmaScriptGlobals = map[string]bool{
	"AggregateError":       true,
	"Array":                true,
	"ArrayBuffer":          true,
	"AsyncDisposableStack": true,
	"AsyncIterator":        true,
	"Atomics":              true,
	"BigInt":               true,
	"BigInt64Array":        true,
	"BigUint64Array":       true,
	"Boolean":              true,
	"DataView":             true,
	"Date":                 true,
	"decodeURI":            true,
	"decodeURIComponent":   true,
	"DisposableStack":      true,
	"encodeURI":            true,
	"encodeURIComponent":   true,
	"Error":                true,
	"escape":               true,
	"EvalError":            true,
	"FinalizationRegistry": true,
	"Float32Array":         true,
	"Float64Array":         true,
	"Function":             true,
	"globalThis":           true,
	"Infinity":             true,
	"Int8Array":            true,
	"Int16Array":           true,
	"Int32Array":           true,
	"Intl":                 true,
	"isFinite":             true,
	"isNaN":                true,
	"Iterator":             true,
	"JSON":                 true,
	"Map":                  true,
	"Math":                 true,
	"NaN":                  true,
	"Number":               true,
	"Object":               true,
	"parseFloat":           true,
	"parseInt":             true,
	"Promise":              true,
	"Proxy":                true,
	"RangeError":           true,
	"ReferenceError":       true,
	"Reflect":              true,
	"RegExp":               true,
	"Set":                  true,
	"SharedArrayBuffer":    true,
	"String":               true,
	"SuppressedError":      true,
	"Symbol":               true,
	"SyntaxError":          true,
	"TypeError":            true,
	"Uint8Array":           true,
	"Uint8ClampedArray":    true,
	"Uint16Array":          true,
	"Uint32Array":          true,
	"unescape":             true,
	"URIError":             true,
	"undefined":            true,
	"WeakMap":              true,
	"WeakRef":              true,
	"WeakSet":              true,
}

// https://eslint.org/docs/latest/rules/no-shadow
//
// Scope semantics are reconstructed by walking the AST and tracking scopes
// directly (rslint has no eslint-scope-equivalent). This covers the common
// cases exercised by the ESLint test suite. Framework-level concepts that
// rslint deliberately does not expose (for example `/*global*/` directive
// comments, `env`/`globals` in languageOptions, `parserOptions.globalReturn`)
// are intentionally not modeled — the rule reports shadowing against
// declarations visible within the file plus, when a type checker is
// available, symbols from the default TypeScript libraries.

type hoistMode int

const (
	hoistFunctions hoistMode = iota
	hoistAll
	hoistNever
	hoistTypes
	hoistFunctionsAndTypes
)

type options struct {
	builtinGlobals                             bool
	hoist                                      hoistMode
	allow                                      map[string]bool
	ignoreOnInitialization                     bool
	ignoreTypeValueShadow                      bool
	ignoreFunctionTypeParameterNameValueShadow bool
}

func defaultOptions() options {
	return options{
		builtinGlobals:         false,
		hoist:                  hoistFunctions,
		allow:                  map[string]bool{},
		ignoreOnInitialization: false,
		ignoreTypeValueShadow:  true,
		ignoreFunctionTypeParameterNameValueShadow: true,
	}
}

// defaultOptionsTSESLint returns the typescript-eslint defaults: identical to
// the ESLint core defaults except `hoist` is `functions-and-types`.
func defaultOptionsTSESLint() options {
	o := defaultOptions()
	o.hoist = hoistFunctionsAndTypes
	return o
}

func parseOptionsWith(raw any, opts options) options {
	// Always copy the allow map: the caller's `opts` may be a long-lived
	// defaults instance shared across rule invocations (e.g. the closure
	// captured by `runWithDefaults`). Mutating in-place would leak state
	// from one source file's lint run to the next.
	src := opts.allow
	opts.allow = make(map[string]bool, len(src)+4)
	for k, v := range src {
		opts.allow[k] = v
	}
	optsMap := utils.GetOptionsMap(raw)
	if optsMap == nil {
		return opts
	}
	if v, ok := optsMap["builtinGlobals"].(bool); ok {
		opts.builtinGlobals = v
	}
	if v, ok := optsMap["hoist"].(string); ok {
		switch v {
		case "all":
			opts.hoist = hoistAll
		case "functions":
			opts.hoist = hoistFunctions
		case "never":
			opts.hoist = hoistNever
		case "types":
			opts.hoist = hoistTypes
		case "functions-and-types":
			opts.hoist = hoistFunctionsAndTypes
		}
	}
	if v, ok := optsMap["ignoreOnInitialization"].(bool); ok {
		opts.ignoreOnInitialization = v
	}
	if v, ok := optsMap["ignoreTypeValueShadow"].(bool); ok {
		opts.ignoreTypeValueShadow = v
	}
	if v, ok := optsMap["ignoreFunctionTypeParameterNameValueShadow"].(bool); ok {
		opts.ignoreFunctionTypeParameterNameValueShadow = v
	}
	if list, ok := optsMap["allow"].([]interface{}); ok {
		for _, item := range list {
			if s, ok := item.(string); ok {
				opts.allow[s] = true
			}
		}
	}
	return opts
}

// ---------------------------------------------------------------------------
// Scope model
// ---------------------------------------------------------------------------

type defKind int

const (
	defVariable       defKind = iota // var/let/const binding, binding element
	defParameter                     // function parameter
	defFunctionName                  // FunctionDeclaration name (outer scope)
	defFnExprName                    // FunctionExpression name (inner scope only)
	defClassName                     // ClassDeclaration name (outer scope)
	defClassInnerName                // Class name visible inside class scope
	defImport                        // Import specifier / default / namespace
	defCatch                         // Catch parameter
	defType                          // Interface, type alias
	defEnumName                      // Enum declaration
	defNamespaceName                 // Module/namespace declaration
	defTypeParameter                 // Generic type parameter
)

type variable struct {
	name    string
	id      *ast.Node // identifier node of this declaration
	defNode *ast.Node // declaration node (VariableDeclaration, FunctionDeclaration, ...)
	parent  *ast.Node // parent of defNode (for parsed import detection, parameter typing, etc.)
	kind    defKind

	isValueBinding   bool // runtime value vs. type-only
	isTypeOnlyImport bool // ImportSpecifier with `type` modifier
	declareModifier  bool // `declare` modifier (.d.ts handling)

	scope *scope
}

type scopeKind int

const (
	scopeGlobal           scopeKind = iota
	scopeFunction                   // function-like bodies & their parameters
	scopeFunctionExprName           // FunctionExpression's name binding
	scopeBlock                      // { ... } / for-init / switch case
	scopeCatch                      // catch clause
	scopeClass                      // class body: type parameters & inner class name
	scopeModule                     // TS namespace
	scopeType                       // TS type alias / interface / function type: type parameters
)

type scope struct {
	kind               scopeKind
	block              *ast.Node
	parent             *scope
	vars               []*variable
	byName             map[string][]*variable
	globalAugmentation bool // true inside `declare global { ... }` chain
}

func newScope(kind scopeKind, block *ast.Node, parent *scope) *scope {
	return &scope{
		kind:   kind,
		block:  block,
		parent: parent,
		byName: map[string][]*variable{},
	}
}

func (s *scope) add(v *variable) {
	v.scope = s
	s.vars = append(s.vars, v)
	s.byName[v.name] = append(s.byName[v.name], v)
}

// variableScope returns the nearest ancestor (or self) that acts as a var
// hoist target: function-like scopes, module scopes, and the global scope.
func (s *scope) variableScope() *scope {
	current := s
	for current != nil {
		switch current.kind {
		case scopeFunction, scopeModule, scopeGlobal:
			return current
		}
		current = current.parent
	}
	return nil
}

// ---------------------------------------------------------------------------
// Scope builder
// ---------------------------------------------------------------------------

type builder struct {
	sourceFile     *ast.SourceFile
	allScopes      []*scope
	builtinGlobals map[string]bool
}

func (b *builder) push(kind scopeKind, block *ast.Node, parent *scope) *scope {
	s := newScope(kind, block, parent)
	if parent != nil && parent.globalAugmentation {
		s.globalAugmentation = true
	}
	b.allScopes = append(b.allScopes, s)
	return s
}

func (b *builder) buildProgram(sf *ast.SourceFile) *scope {
	global := b.push(scopeGlobal, sf.AsNode(), nil)
	if sf.Statements == nil {
		return global
	}
	// First pass: hoist var / function / class / import names into the global scope.
	b.hoistStatements(sf.Statements.Nodes, global)
	// Second pass: walk children that introduce nested scopes.
	for _, stmt := range sf.Statements.Nodes {
		b.visitStatement(stmt, global)
	}
	return global
}

// hoistStatements collects declarations that belong to the enclosing
// variable scope (var / function / class / enum / interface / type /
// namespace / import). Block-scoped declarations (let/const/class inside
// a nested block) are collected in visitBlock.
func (b *builder) hoistStatements(statements []*ast.Node, s *scope) {
	for _, stmt := range statements {
		b.hoistStatement(stmt, s, false)
	}
}

// addNamedDecl records a declaration whose binding name is the Name() child
// of `stmt` (covers function, class, interface, type alias, enum, namespace,
// import-equals). Returns without adding when the statement has no identifier
// name (for example anonymous default exports and ambient-module strings).
func (b *builder) addNamedDecl(stmt *ast.Node, s *scope, kind defKind, isValue bool) {
	n := stmt.Name()
	if n == nil || n.Kind != ast.KindIdentifier {
		return
	}
	isAmbient := ast.HasSyntacticModifier(stmt, ast.ModifierFlagsAmbient)
	// Body-less FunctionDeclarations come in two flavors: overload signatures
	// (no `declare`, there will be a later implementation) and ambient
	// declarations (`declare function foo(): void`). typescript-eslint's scope
	// manager merges all signatures under one Variable, so we only need to
	// register a single binding: skip overload signatures, but keep ambient
	// declarations so that inner scopes detect them as shadowed.
	if kind == defFunctionName && stmt.Body() == nil && !isAmbient {
		return
	}
	s.add(&variable{
		name:            n.Text(),
		id:              n,
		defNode:         stmt,
		parent:          stmt.Parent,
		kind:            kind,
		isValueBinding:  isValue,
		declareModifier: isAmbient,
	})
}

// hoistStatement records the declarations introduced by `stmt` into scope
// `s`. When `blockScope` is true, only block-scoped bindings (let/const/
// class/function/type/enum/namespace) are added — var is assumed to have
// been hoisted to an outer variable scope. Otherwise, every binding kind
// (including var) is added and the function's child statements are scanned
// for nested `var` declarations that also hoist to this scope.
func (b *builder) hoistStatement(stmt *ast.Node, s *scope, blockScope bool) {
	if stmt == nil {
		return
	}
	switch stmt.Kind {
	case ast.KindVariableStatement:
		vs := stmt.AsVariableStatement()
		if vs == nil || vs.DeclarationList == nil {
			return
		}
		isVar := utils.IsVarKeyword(vs.DeclarationList)
		if blockScope && isVar {
			return
		}
		declareMod := ast.HasSyntacticModifier(stmt, ast.ModifierFlagsAmbient)
		b.collectVariableDeclarations(vs.DeclarationList, s, declareMod)
	case ast.KindFunctionDeclaration:
		b.addNamedDecl(stmt, s, defFunctionName, true)
	case ast.KindClassDeclaration:
		b.addNamedDecl(stmt, s, defClassName, true)
	case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
		b.addNamedDecl(stmt, s, defType, false)
	case ast.KindEnumDeclaration:
		b.addNamedDecl(stmt, s, defEnumName, true)
	case ast.KindModuleDeclaration:
		// Ambient `declare module 'str' { ... }` uses a string-literal name and
		// doesn't bind a variable — addNamedDecl skips it automatically.
		b.addNamedDecl(stmt, s, defNamespaceName, true)
	case ast.KindImportDeclaration:
		if !blockScope {
			b.collectImport(stmt, s)
		}
	case ast.KindImportEqualsDeclaration:
		if !blockScope {
			b.addNamedDecl(stmt, s, defImport, true)
		}
	case ast.KindForStatement:
		fs := stmt.AsForStatement()
		if !blockScope && fs != nil && fs.Initializer != nil && fs.Initializer.Kind == ast.KindVariableDeclarationList && utils.IsVarKeyword(fs.Initializer) {
			b.collectVariableDeclarations(fs.Initializer, s, false)
		}
	case ast.KindForInStatement, ast.KindForOfStatement:
		fs := stmt.AsForInOrOfStatement()
		if !blockScope && fs != nil && fs.Initializer != nil && fs.Initializer.Kind == ast.KindVariableDeclarationList && utils.IsVarKeyword(fs.Initializer) {
			b.collectVariableDeclarations(fs.Initializer, s, false)
		}
	}

	if blockScope {
		return
	}
	// For variable-scope hoisting, recurse into wrapper statements to collect
	// nested `var` declarations. Skip statement kinds that form scope
	// boundaries of their own (function / class / module bodies are processed
	// by visit*).
	switch stmt.Kind {
	case ast.KindBlock, ast.KindIfStatement, ast.KindWhileStatement,
		ast.KindDoStatement, ast.KindTryStatement, ast.KindCatchClause,
		ast.KindSwitchStatement, ast.KindCaseClause, ast.KindDefaultClause,
		ast.KindCaseBlock, ast.KindForStatement, ast.KindForInStatement,
		ast.KindForOfStatement, ast.KindLabeledStatement, ast.KindWithStatement,
		ast.KindExpressionStatement, ast.KindReturnStatement, ast.KindThrowStatement:
		stmt.ForEachChild(func(child *ast.Node) bool {
			b.hoistVarOnly(child, s)
			return false
		})
	}
}

// hoistVarOnly walks a subtree and only hoists `var` declarations (into the
// enclosing function/module/global scope). Function-like nodes and ambient
// module declarations terminate the walk.
func (b *builder) hoistVarOnly(node *ast.Node, s *scope) {
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindVariableStatement:
		vs := node.AsVariableStatement()
		if vs != nil && vs.DeclarationList != nil && utils.IsVarKeyword(vs.DeclarationList) {
			declareMod := ast.HasSyntacticModifier(node, ast.ModifierFlagsAmbient)
			b.collectVariableDeclarations(vs.DeclarationList, s, declareMod)
		}
		return
	case ast.KindForStatement:
		fs := node.AsForStatement()
		if fs != nil && fs.Initializer != nil && fs.Initializer.Kind == ast.KindVariableDeclarationList && utils.IsVarKeyword(fs.Initializer) {
			b.collectVariableDeclarations(fs.Initializer, s, false)
		}
	case ast.KindForInStatement, ast.KindForOfStatement:
		fs := node.AsForInOrOfStatement()
		if fs != nil && fs.Initializer != nil && fs.Initializer.Kind == ast.KindVariableDeclarationList && utils.IsVarKeyword(fs.Initializer) {
			b.collectVariableDeclarations(fs.Initializer, s, false)
		}
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
		ast.KindArrowFunction, ast.KindMethodDeclaration,
		ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindClassStaticBlockDeclaration, ast.KindClassDeclaration,
		ast.KindClassExpression, ast.KindModuleDeclaration:
		return
	}
	node.ForEachChild(func(child *ast.Node) bool {
		b.hoistVarOnly(child, s)
		return false
	})
}

// collectVariableDeclarations walks a VariableDeclarationList and adds each
// binding identifier to the given scope. For destructuring patterns, the
// innermost BindingElement is stored as `defNode` so that the rule can reach
// the initializer/default of that specific binding site later.
func (b *builder) collectVariableDeclarations(declList *ast.Node, s *scope, declareMod bool) {
	list := declList.AsVariableDeclarationList()
	if list == nil || list.Declarations == nil {
		return
	}
	for _, decl := range list.Declarations.Nodes {
		vd := decl.AsVariableDeclaration()
		if vd == nil || vd.Name() == nil {
			continue
		}
		utils.CollectBindingNames(vd.Name(), func(id *ast.Node, name string) {
			defNode := decl
			for p := id.Parent; p != nil && p != decl; p = p.Parent {
				if p.Kind == ast.KindBindingElement {
					defNode = p
					break
				}
			}
			s.add(&variable{
				name:            name,
				id:              id,
				defNode:         defNode,
				parent:          defNode.Parent,
				kind:            defVariable,
				isValueBinding:  true,
				declareModifier: declareMod,
			})
		})
	}
}

func (b *builder) collectImport(node *ast.Node, s *scope) {
	importDecl := node.AsImportDeclaration()
	if importDecl == nil || importDecl.ImportClause == nil {
		return
	}
	clause := importDecl.ImportClause.AsImportClause()
	if clause == nil {
		return
	}
	isTypeImport := importDecl.ImportClause.IsTypeOnly()
	// Default import.
	if clause.Name() != nil && clause.Name().Kind == ast.KindIdentifier {
		s.add(&variable{
			name:             clause.Name().Text(),
			id:               clause.Name(),
			defNode:          node,
			parent:           node.Parent,
			kind:             defImport,
			isValueBinding:   !isTypeImport,
			isTypeOnlyImport: isTypeImport,
		})
	}
	if clause.NamedBindings == nil {
		return
	}
	switch clause.NamedBindings.Kind {
	case ast.KindNamespaceImport:
		ns := clause.NamedBindings.AsNamespaceImport()
		if ns != nil && ns.Name() != nil && ns.Name().Kind == ast.KindIdentifier {
			s.add(&variable{
				name:             ns.Name().Text(),
				id:               ns.Name(),
				defNode:          node,
				parent:           node.Parent,
				kind:             defImport,
				isValueBinding:   !isTypeImport,
				isTypeOnlyImport: isTypeImport,
			})
		}
	case ast.KindNamedImports:
		named := clause.NamedBindings.AsNamedImports()
		if named == nil || named.Elements == nil {
			return
		}
		for _, elem := range named.Elements.Nodes {
			if elem == nil {
				continue
			}
			spec := elem.AsImportSpecifier()
			if spec == nil || spec.Name() == nil || spec.Name().Kind != ast.KindIdentifier {
				continue
			}
			specTypeOnly := spec.IsTypeOnly || isTypeImport
			s.add(&variable{
				name:             spec.Name().Text(),
				id:               spec.Name(),
				defNode:          node,
				parent:           node.Parent,
				kind:             defImport,
				isValueBinding:   !specTypeOnly,
				isTypeOnlyImport: specTypeOnly,
			})
		}
	}
}

// visitStatement recurses into nodes that may create nested scopes.
func (b *builder) visitStatement(stmt *ast.Node, parent *scope) {
	if stmt == nil {
		return
	}
	switch stmt.Kind {
	case ast.KindFunctionDeclaration:
		b.visitFunctionLike(stmt, parent)
	case ast.KindClassDeclaration:
		b.visitClass(stmt, parent, false)
	case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
		b.visitTypeDecl(stmt, parent)
	case ast.KindEnumDeclaration:
		b.visitEnum(stmt, parent)
	case ast.KindModuleDeclaration:
		b.visitModuleDecl(stmt, parent)
	case ast.KindBlock:
		b.visitBlock(stmt, parent)
	case ast.KindForStatement:
		b.visitForStatement(stmt, parent)
	case ast.KindForInStatement, ast.KindForOfStatement:
		b.visitForInOrOf(stmt, parent)
	case ast.KindIfStatement:
		ifStmt := stmt.AsIfStatement()
		if ifStmt != nil {
			b.visitExpression(ifStmt.Expression, parent)
			b.visitStatement(ifStmt.ThenStatement, parent)
			b.visitStatement(ifStmt.ElseStatement, parent)
		}
	case ast.KindWhileStatement:
		ws := stmt.AsWhileStatement()
		if ws != nil {
			b.visitExpression(ws.Expression, parent)
			b.visitStatement(ws.Statement, parent)
		}
	case ast.KindDoStatement:
		ds := stmt.AsDoStatement()
		if ds != nil {
			b.visitExpression(ds.Expression, parent)
			b.visitStatement(ds.Statement, parent)
		}
	case ast.KindTryStatement:
		ts := stmt.AsTryStatement()
		if ts != nil {
			b.visitStatement(ts.TryBlock, parent)
			if ts.CatchClause != nil {
				b.visitCatch(ts.CatchClause, parent)
			}
			b.visitStatement(ts.FinallyBlock, parent)
		}
	case ast.KindSwitchStatement:
		sw := stmt.AsSwitchStatement()
		if sw != nil {
			b.visitExpression(sw.Expression, parent)
			b.visitSwitchCases(sw, parent)
		}
	case ast.KindLabeledStatement:
		ls := stmt.AsLabeledStatement()
		if ls != nil {
			b.visitStatement(ls.Statement, parent)
		}
	case ast.KindReturnStatement:
		if rs := stmt.AsReturnStatement(); rs != nil {
			b.visitExpression(rs.Expression, parent)
		}
	case ast.KindThrowStatement:
		if ts := stmt.AsThrowStatement(); ts != nil {
			b.visitExpression(ts.Expression, parent)
		}
	case ast.KindExpressionStatement:
		if es := stmt.AsExpressionStatement(); es != nil {
			b.visitExpression(es.Expression, parent)
		}
	case ast.KindVariableStatement:
		vs := stmt.AsVariableStatement()
		if vs != nil && vs.DeclarationList != nil {
			b.visitVarDeclList(vs.DeclarationList, parent)
		}
	case ast.KindExportAssignment:
		ea := stmt.AsExportAssignment()
		if ea != nil {
			b.visitExpression(ea.Expression, parent)
		}
	case ast.KindExportDeclaration:
		// No new scopes.
	case ast.KindImportDeclaration, ast.KindImportEqualsDeclaration:
		// No new scopes.
	case ast.KindWithStatement:
		ws := stmt.AsWithStatement()
		if ws != nil {
			b.visitExpression(ws.Expression, parent)
			b.visitStatement(ws.Statement, parent)
		}
	default:
		// Catch-all — traverse children.
		stmt.ForEachChild(func(child *ast.Node) bool {
			b.visitStatement(child, parent)
			return false
		})
	}
}

// visitVarDeclList walks a VariableDeclarationList so that any nested
// function/class/conditional-type scopes in type annotations, initializers,
// or destructuring defaults get created.
func (b *builder) visitVarDeclList(declList *ast.Node, parent *scope) {
	list := declList.AsVariableDeclarationList()
	if list == nil || list.Declarations == nil {
		return
	}
	for _, decl := range list.Declarations.Nodes {
		vd := decl.AsVariableDeclaration()
		if vd == nil {
			continue
		}
		if vd.Type != nil {
			b.visitExpression(vd.Type, parent)
		}
		if vd.Initializer != nil {
			b.visitExpression(vd.Initializer, parent)
		}
		if vd.Name() != nil {
			b.visitBindingPattern(vd.Name(), parent)
		}
	}
}

// visitBindingPattern recurses into destructuring defaults that may contain
// function / class expressions creating nested scopes.
func (b *builder) visitBindingPattern(pattern *ast.Node, parent *scope) {
	if pattern == nil {
		return
	}
	pattern.ForEachChild(func(child *ast.Node) bool {
		if child.Kind == ast.KindBindingElement {
			be := child.AsBindingElement()
			if be != nil {
				if be.Initializer != nil {
					b.visitExpression(be.Initializer, parent)
				}
				if be.Name() != nil {
					b.visitBindingPattern(be.Name(), parent)
				}
			}
		}
		return false
	})
}

// visitExpression walks an expression to discover nested function/class
// expressions and other scope-creating constructs.
func (b *builder) visitExpression(expr *ast.Node, parent *scope) {
	if expr == nil {
		return
	}
	switch expr.Kind {
	case ast.KindFunctionExpression, ast.KindArrowFunction:
		b.visitFunctionLike(expr, parent)
		return
	case ast.KindMethodDeclaration, ast.KindConstructor,
		ast.KindGetAccessor, ast.KindSetAccessor:
		// Object-literal shorthand methods (and accessors) also need their
		// own function scope. Class-body members are already processed by
		// visitClass, so we only get here for the object-literal case.
		b.visitFunctionLike(expr, parent)
		return
	case ast.KindClassExpression:
		b.visitClass(expr, parent, true)
		return
	}
	if ast.IsFunctionTypeNode(expr) || ast.IsConstructorTypeNode(expr) ||
		ast.IsCallSignatureDeclaration(expr) || ast.IsConstructSignatureDeclaration(expr) ||
		ast.IsMethodSignatureDeclaration(expr) {
		b.visitFunctionType(expr, parent)
		return
	}
	if expr.Kind == ast.KindConditionalType {
		b.visitConditionalType(expr, parent)
		return
	}
	expr.ForEachChild(func(child *ast.Node) bool {
		b.visitExpression(child, parent)
		return false
	})
}

// visitConditionalType creates a scope for the conditional type so that
// `infer X` clauses in the extends-position introduce a binding visible to
// nested types and to the true branch. ESLint/typescript-eslint reports an
// inner `infer X` shadowing an outer one within the same conditional chain.
func (b *builder) visitConditionalType(node *ast.Node, outer *scope) {
	cond := node.AsConditionalTypeNode()
	if cond == nil {
		return
	}
	condScope := b.push(scopeType, node, outer)
	collectInferTypes(cond.ExtendsType, condScope)
	if cond.CheckType != nil {
		b.visitExpression(cond.CheckType, outer)
	}
	if cond.ExtendsType != nil {
		b.visitExpression(cond.ExtendsType, condScope)
	}
	if cond.TrueType != nil {
		b.visitExpression(cond.TrueType, condScope)
	}
	if cond.FalseType != nil {
		b.visitExpression(cond.FalseType, outer)
	}
}

// collectInferTypes walks `extendsType` and adds each `infer X` binding to
// the conditional-type scope. Nested function/conditional types stop the
// walk because they introduce their own scopes.
func collectInferTypes(node *ast.Node, s *scope) {
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindFunctionType, ast.KindConstructorType, ast.KindConditionalType:
		return
	case ast.KindInferType:
		it := node.AsInferTypeNode()
		if it != nil && it.TypeParameter != nil {
			tp := it.TypeParameter.AsTypeParameterDeclaration()
			if tp != nil && tp.Name() != nil && tp.Name().Kind == ast.KindIdentifier {
				s.add(&variable{
					name:           tp.Name().Text(),
					id:             tp.Name(),
					defNode:        it.TypeParameter,
					parent:         it.TypeParameter.Parent,
					kind:           defTypeParameter,
					isValueBinding: false,
				})
			}
		}
	}
	node.ForEachChild(func(child *ast.Node) bool {
		collectInferTypes(child, s)
		return false
	})
}

// visitFunctionType handles TS function/construct types and call/construct
// signatures. Parameters introduce bindings that may trigger (or be filtered
// out by) `ignoreFunctionTypeParameterNameValueShadow`.
func (b *builder) visitFunctionType(node *ast.Node, outer *scope) {
	s := b.push(scopeFunction, node, outer)
	b.addTypeParameters(node, s)
	b.addParameters(node, s, false)
	node.ForEachChild(func(child *ast.Node) bool {
		b.visitExpression(child, s)
		return false
	})
}

// addTypeParameters records `function f<T>` / `class C<T>` / `type X<T>`
// generic names into `s`. Each type parameter is a type-only binding.
func (b *builder) addTypeParameters(node *ast.Node, s *scope) {
	for _, tp := range node.TypeParameters() {
		if tp == nil {
			continue
		}
		tpDecl := tp.AsTypeParameterDeclaration()
		if tpDecl == nil || tpDecl.Name() == nil || tpDecl.Name().Kind != ast.KindIdentifier {
			continue
		}
		s.add(&variable{
			name:           tpDecl.Name().Text(),
			id:             tpDecl.Name(),
			defNode:        tp,
			parent:         tp.Parent,
			kind:           defTypeParameter,
			isValueBinding: false,
		})
	}
}

// addParameters records the binding identifiers from each parameter of
// `node` into `s`, optionally recursing into each parameter's type
// annotation and initializer so that nested function types create scopes.
func (b *builder) addParameters(node *ast.Node, s *scope, recurseAnnotations bool) {
	for _, param := range node.Parameters() {
		if param == nil {
			continue
		}
		// Skip `this` pseudo-parameter.
		if pn := param.Name(); pn != nil && pn.Kind == ast.KindIdentifier && pn.Text() == "this" {
			continue
		}
		if recurseAnnotations {
			// Parameter decorators (`@dec x: T`) run at class-definition time,
			// in the enclosing class scope (s.parent for methods). Fall back
			// to `s` for standalone functions (decorators there are illegal
			// in TS but shouldn't crash).
			decScope := s
			if s.parent != nil {
				decScope = s.parent
			}
			for _, dec := range param.Decorators() {
				b.visitExpression(dec, decScope)
			}
			if paramDecl := param.AsParameterDeclaration(); paramDecl != nil && paramDecl.Type != nil {
				b.visitExpression(paramDecl.Type, s)
			}
			if init := param.Initializer(); init != nil {
				b.visitExpression(init, s)
			}
		}
		if param.Name() == nil {
			continue
		}
		if recurseAnnotations && param.Name().Kind != ast.KindIdentifier {
			b.visitBindingPattern(param.Name(), s)
		}
		utils.CollectBindingNames(param.Name(), func(id *ast.Node, name string) {
			s.add(&variable{
				name:           name,
				id:             id,
				defNode:        param,
				parent:         param.Parent,
				kind:           defParameter,
				isValueBinding: true,
			})
		})
	}
}

// visitFunctionLike handles FunctionDeclaration, FunctionExpression,
// ArrowFunction, MethodDeclaration, Constructor, Get/SetAccessor.
func (b *builder) visitFunctionLike(node *ast.Node, outer *scope) {
	// Computed method/accessor name (`[expr]`) is evaluated in the enclosing
	// scope (class body or object literal context). Walk it before pushing
	// the function's own scope so that shadows inside the key expression
	// check against the right scope.
	if name := node.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
		if cpn := name.AsComputedPropertyName(); cpn != nil && cpn.Expression != nil {
			b.visitExpression(cpn.Expression, outer)
		}
	}

	// Function expression with a name → intermediate scope holding just
	// the function's own name.
	fnExprScope := outer
	if node.Kind == ast.KindFunctionExpression {
		if n := node.Name(); n != nil && n.Kind == ast.KindIdentifier {
			innerFnName := b.push(scopeFunctionExprName, node, outer)
			innerFnName.add(&variable{
				name:           n.Text(),
				id:             n,
				defNode:        node,
				parent:         node.Parent,
				kind:           defFnExprName,
				isValueBinding: true,
			})
			fnExprScope = innerFnName
		}
	}

	fnScope := b.push(scopeFunction, node, fnExprScope)
	b.addTypeParameters(node, fnScope)
	b.addParameters(node, fnScope, true)
	if retType := node.Type(); retType != nil {
		b.visitExpression(retType, fnScope)
	}

	// Body.
	body := node.Body()
	if body == nil {
		return
	}
	if body.Kind == ast.KindBlock {
		block := body.AsBlock()
		if block != nil && block.Statements != nil {
			b.hoistStatements(block.Statements.Nodes, fnScope)
			for _, stmt := range block.Statements.Nodes {
				b.visitStatement(stmt, fnScope)
			}
		}
	} else {
		// Arrow expression body.
		b.visitExpression(body, fnScope)
	}
}

func (b *builder) visitClass(node *ast.Node, outer *scope, isExpression bool) {
	// Class-level decorators run BEFORE the class is defined — in the outer
	// scope. Any shadows inside them (e.g. `@((t) => { const x = 1; })`) must
	// be checked against outer bindings.
	for _, dec := range node.Decorators() {
		b.visitExpression(dec, outer)
	}

	classScope := b.push(scopeClass, node, outer)
	// Inner scope always holds the class name (for both declarations and
	// expressions — ESLint's scope model duplicates the name here).
	if n := node.Name(); n != nil && n.Kind == ast.KindIdentifier {
		_ = isExpression
		classScope.add(&variable{
			name:           n.Text(),
			id:             n,
			defNode:        node,
			parent:         node.Parent,
			kind:           defClassInnerName,
			isValueBinding: true,
		})
	}
	b.addTypeParameters(node, classScope)

	// Heritage clauses (`extends`, `implements`): expressions here are
	// evaluated when the class is defined and can contain IIFEs/arrows whose
	// body may shadow outer bindings. Walk them inside classScope so that
	// class type parameters remain visible to type arguments.
	node.ForEachChild(func(c *ast.Node) bool {
		if c.Kind == ast.KindHeritageClause {
			c.ForEachChild(func(t *ast.Node) bool {
				b.visitExpression(t, classScope)
				return false
			})
		}
		return false
	})

	for _, member := range node.Members() {
		if member == nil {
			continue
		}
		// Member decorators evaluate when the class is defined; use classScope
		// so that class type parameters remain visible.
		for _, dec := range member.Decorators() {
			b.visitExpression(dec, classScope)
		}
		switch member.Kind {
		case ast.KindMethodDeclaration, ast.KindConstructor,
			ast.KindGetAccessor, ast.KindSetAccessor:
			// visitFunctionLike handles the computed-name (if any) itself.
			b.visitFunctionLike(member, classScope)
		case ast.KindPropertyDeclaration:
			// Properties don't go through visitFunctionLike — walk the
			// computed key here.
			if memberName := member.Name(); memberName != nil && memberName.Kind == ast.KindComputedPropertyName {
				if cpn := memberName.AsComputedPropertyName(); cpn != nil && cpn.Expression != nil {
					b.visitExpression(cpn.Expression, classScope)
				}
			}
			if init := member.Initializer(); init != nil {
				b.visitExpression(init, classScope)
			}
		case ast.KindClassStaticBlockDeclaration:
			sb := member.AsClassStaticBlockDeclaration()
			if sb != nil && sb.Body != nil && sb.Body.Kind == ast.KindBlock {
				// Static block is its own function-like variable scope for var purposes.
				staticScope := b.push(scopeFunction, member, classScope)
				block := sb.Body.AsBlock()
				if block != nil && block.Statements != nil {
					b.hoistStatements(block.Statements.Nodes, staticScope)
					for _, stmt := range block.Statements.Nodes {
						b.visitStatement(stmt, staticScope)
					}
				}
			}
		}
	}
}

func (b *builder) visitTypeDecl(node *ast.Node, outer *scope) {
	typeScope := b.push(scopeType, node, outer)
	b.addTypeParameters(node, typeScope)
	// Recurse into type body / heritage / members to discover FunctionType
	// and similar type-level scopes.
	node.ForEachChild(func(child *ast.Node) bool {
		b.visitExpression(child, typeScope)
		return false
	})
}

func (b *builder) visitEnum(node *ast.Node, outer *scope) {
	enumScope := b.push(scopeBlock, node, outer)
	ed := node.AsEnumDeclaration()
	if ed == nil || ed.Members == nil {
		return
	}
	// Enum members are bindings in the enum's inner scope. An outer
	// declaration with the same name is reported as shadowed by the member.
	for _, m := range ed.Members.Nodes {
		if m == nil {
			continue
		}
		em := m.AsEnumMember()
		if em == nil || em.Name() == nil {
			continue
		}
		if em.Name().Kind == ast.KindIdentifier {
			enumScope.add(&variable{
				name:           em.Name().Text(),
				id:             em.Name(),
				defNode:        m,
				parent:         m.Parent,
				kind:           defVariable,
				isValueBinding: true,
			})
		}
		if em.Initializer != nil {
			b.visitExpression(em.Initializer, enumScope)
		}
	}
}

func (b *builder) visitModuleDecl(node *ast.Node, outer *scope) {
	md := node.AsModuleDeclaration()
	if md == nil {
		return
	}
	// `declare global { ... }` — treat as continuation of the enclosing global
	// scope. Any bindings inside are conceptually global and the scopes under
	// it are marked so shadowing checks are skipped.
	if ast.IsGlobalScopeAugmentation(node) {
		augScope := b.push(scopeModule, node, outer)
		augScope.globalAugmentation = true
		if md.Body != nil {
			b.walkModuleBlock(md.Body, augScope)
		}
		return
	}
	moduleScope := b.push(scopeModule, node, outer)
	// Inherit the global-augmentation flag from the parent chain.
	if outer != nil && outer.globalAugmentation {
		moduleScope.globalAugmentation = true
	}
	if md.Body == nil {
		return
	}
	if md.Body.Kind == ast.KindModuleDeclaration {
		// Nested namespace chain: `namespace A.B.C { }` — unwrap.
		b.visitModuleDecl(md.Body, moduleScope)
		return
	}
	b.walkModuleBlock(md.Body, moduleScope)
}

func (b *builder) walkModuleBlock(body *ast.Node, s *scope) {
	if body == nil {
		return
	}
	if body.Kind == ast.KindModuleBlock {
		mb := body.AsModuleBlock()
		if mb != nil && mb.Statements != nil {
			b.hoistStatements(mb.Statements.Nodes, s)
			for _, stmt := range mb.Statements.Nodes {
				b.visitStatement(stmt, s)
			}
		}
	}
}

func (b *builder) visitBlock(block *ast.Node, outer *scope) {
	parent := block.Parent
	// If this block is the body of a function-like, it has already been processed by visitFunctionLike.
	if parent != nil {
		switch parent.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
			ast.KindArrowFunction, ast.KindMethodDeclaration,
			ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindClassStaticBlockDeclaration:
			return
		case ast.KindCatchClause:
			// Handled in visitCatch.
			return
		}
	}
	s := b.push(scopeBlock, block, outer)
	blk := block.AsBlock()
	if blk == nil || blk.Statements == nil {
		return
	}
	// Block-scoped declarations (let/const/class/function-in-strict/type/...) live at this block.
	for _, stmt := range blk.Statements.Nodes {
		b.hoistStatement(stmt, s, true)
	}
	// `var` hoists up to the enclosing variable scope.
	varHost := s.variableScope()
	if varHost == nil {
		varHost = s
	}
	for _, stmt := range blk.Statements.Nodes {
		b.hoistVarOnly(stmt, varHost)
	}
	// Recurse.
	for _, stmt := range blk.Statements.Nodes {
		b.visitStatement(stmt, s)
	}
}

// forInitScope handles a for/for-in/for-of initializer. A let/const
// initializer introduces a new block scope wrapping the loop body; otherwise
// the initializer stays in `outer`. Returns the scope the loop body lives in.
func (b *builder) forInitScope(forNode *ast.Node, initializer *ast.Node, outer *scope) *scope {
	if initializer != nil && initializer.Kind == ast.KindVariableDeclarationList && !utils.IsVarKeyword(initializer) {
		forScope := b.push(scopeBlock, forNode, outer)
		b.collectVariableDeclarations(initializer, forScope, false)
		b.visitVarDeclList(initializer, forScope)
		return forScope
	}
	if initializer != nil {
		if initializer.Kind == ast.KindVariableDeclarationList {
			b.visitVarDeclList(initializer, outer)
		} else {
			b.visitExpression(initializer, outer)
		}
	}
	return outer
}

func (b *builder) visitForStatement(stmt *ast.Node, outer *scope) {
	fs := stmt.AsForStatement()
	if fs == nil {
		return
	}
	body := b.forInitScope(stmt, fs.Initializer, outer)
	if fs.Condition != nil {
		b.visitExpression(fs.Condition, body)
	}
	if fs.Incrementor != nil {
		b.visitExpression(fs.Incrementor, body)
	}
	if fs.Statement != nil {
		b.visitStatement(fs.Statement, body)
	}
}

func (b *builder) visitForInOrOf(stmt *ast.Node, outer *scope) {
	fs := stmt.AsForInOrOfStatement()
	if fs == nil {
		return
	}
	body := b.forInitScope(stmt, fs.Initializer, outer)
	if fs.Expression != nil {
		b.visitExpression(fs.Expression, outer)
	}
	if fs.Statement != nil {
		b.visitStatement(fs.Statement, body)
	}
}

func (b *builder) visitCatch(node *ast.Node, outer *scope) {
	cc := node.AsCatchClause()
	if cc == nil {
		return
	}
	catchScope := b.push(scopeCatch, node, outer)
	if cc.VariableDeclaration != nil {
		vd := cc.VariableDeclaration.AsVariableDeclaration()
		if vd != nil && vd.Name() != nil {
			utils.CollectBindingNames(vd.Name(), func(id *ast.Node, name string) {
				catchScope.add(&variable{
					name:           name,
					id:             id,
					defNode:        cc.VariableDeclaration,
					parent:         node,
					kind:           defCatch,
					isValueBinding: true,
				})
			})
		}
	}
	if cc.Block != nil && cc.Block.Kind == ast.KindBlock {
		// The catch block is a BlockStatement whose direct parent is this CatchClause.
		// We treat it as nested: create an extra block scope for its body.
		bodyScope := b.push(scopeBlock, cc.Block, catchScope)
		blk := cc.Block.AsBlock()
		if blk != nil && blk.Statements != nil {
			for _, stmt := range blk.Statements.Nodes {
				b.hoistStatement(stmt, bodyScope, true)
			}
			varHost := bodyScope.variableScope()
			if varHost == nil {
				varHost = bodyScope
			}
			for _, stmt := range blk.Statements.Nodes {
				b.hoistVarOnly(stmt, varHost)
			}
			for _, stmt := range blk.Statements.Nodes {
				b.visitStatement(stmt, bodyScope)
			}
		}
	}
}

func (b *builder) visitSwitchCases(sw *ast.SwitchStatement, outer *scope) {
	if sw.CaseBlock == nil {
		return
	}
	cb := sw.CaseBlock.AsCaseBlock()
	if cb == nil || cb.Clauses == nil {
		return
	}
	// Switch introduces a block scope for its clauses.
	switchScope := b.push(scopeBlock, sw.AsNode(), outer)
	// First: hoist block-level declarations across all clauses.
	for _, clause := range cb.Clauses.Nodes {
		if clause == nil {
			continue
		}
		c := clause.AsCaseOrDefaultClause()
		if c == nil || c.Statements == nil {
			continue
		}
		for _, s := range c.Statements.Nodes {
			b.hoistStatement(s, switchScope, true)
		}
	}
	// Vars hoist to enclosing var scope.
	varHost := switchScope.variableScope()
	if varHost == nil {
		varHost = switchScope
	}
	for _, clause := range cb.Clauses.Nodes {
		if clause == nil {
			continue
		}
		c := clause.AsCaseOrDefaultClause()
		if c == nil || c.Statements == nil {
			continue
		}
		for _, s := range c.Statements.Nodes {
			b.hoistVarOnly(s, varHost)
		}
	}
	// Recurse into statements.
	for _, clause := range cb.Clauses.Nodes {
		if clause == nil {
			continue
		}
		c := clause.AsCaseOrDefaultClause()
		if c == nil {
			continue
		}
		if c.Expression != nil {
			b.visitExpression(c.Expression, switchScope)
		}
		if c.Statements != nil {
			for _, s := range c.Statements.Nodes {
				b.visitStatement(s, switchScope)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Rule entry
// ---------------------------------------------------------------------------

var NoShadowRule = rule.Rule{
	Name: "no-shadow",
	Run:  runWithDefaults(defaultOptions()),
}

// RunTSESLint exposes the rule body with typescript-eslint's defaults so the
// `@typescript-eslint/no-shadow` wrapper can reuse the implementation. The
// underlying closure is built once at package init — `parseOptionsWith`
// copies the captured `allow` map per invocation, so this is safe.
var runTSESLint = runWithDefaults(defaultOptionsTSESLint())

func RunTSESLint(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
	return runTSESLint(ctx, rawOptions)
}

func runWithDefaults(defaults options) func(rule.RuleContext, any) rule.RuleListeners {
	return func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptionsWith(rawOptions, defaults)
		if ctx.SourceFile == nil {
			return rule.RuleListeners{}
		}

		b := &builder{sourceFile: ctx.SourceFile}
		global := b.buildProgram(ctx.SourceFile)

		filename := ""
		if ctx.SourceFile.AsNode() != nil {
			filename = ctx.SourceFile.FileName()
		}
		isDeclFile := strings.HasSuffix(filename, ".d.ts") ||
			strings.HasSuffix(filename, ".d.cts") ||
			strings.HasSuffix(filename, ".d.mts")

		// Pre-compute the set of default-library globals. We seed with a
		// hard-coded ECMAScript list (so the rule works without a
		// TypeChecker, and so it still flags `var Object = 0` at module
		// scope where TypeChecker would resolve `Object` to the local
		// binding), then union in whatever the default library exposes.
		builtinGlobals := map[string]bool{}
		if opts.builtinGlobals {
			for name := range ecmaScriptGlobals {
				builtinGlobals[name] = true
			}
			if ctx.TypeChecker != nil && ctx.Program != nil {
				for _, sym := range ctx.TypeChecker.GetSymbolsInScope(ctx.SourceFile.AsNode(), ast.SymbolFlagsValue) {
					if sym == nil || sym.Name == "" {
						continue
					}
					if utils.IsSymbolFromDefaultLibrary(ctx.Program, sym) {
						builtinGlobals[sym.Name] = true
					}
				}
			}
		}
		b.builtinGlobals = builtinGlobals

		// Walk scopes top-down and check each variable. The global scope is
		// included so that `builtinGlobals: true` can flag a module-level
		// declaration shadowing an ECMAScript global (ESLint's module scope
		// sits between the file and the global scope; we collapse them and
		// compensate by letting checkVariable consult the globals table).
		for _, s := range b.allScopes {
			if s.globalAugmentation {
				continue
			}
			for _, v := range s.vars {
				if opts.allow[v.name] {
					continue
				}
				if v.name == "this" {
					continue
				}
				if isDuplicatedClassNameInClassScope(v) {
					continue
				}
				if isDeclFile && v.declareModifier {
					continue
				}
				b.checkVariable(ctx, s, v, opts, global)
			}
		}

		return rule.RuleListeners{}
	}
}

// isDuplicatedClassNameInClassScope suppresses the inner class-name binding
// that ESLint-scope adds for ClassDeclarations.
func isDuplicatedClassNameInClassScope(v *variable) bool {
	return v.kind == defClassInnerName && v.defNode != nil && v.defNode.Kind == ast.KindClassDeclaration
}

// checkVariable tests whether `v` shadows a variable in some outer scope.
func (b *builder) checkVariable(ctx rule.RuleContext, s *scope, v *variable, opts options, global *scope) {
	shadowed := findShadowed(v, s.parent)
	shadowedGlobal := shadowed == nil && opts.builtinGlobals && b.builtinGlobals[v.name]
	if shadowed == nil && !shadowedGlobal {
		return
	}

	// Ignore function-name-initializer exceptions:
	// var a = function a() {};  /  var A = class A {};
	if shadowed != nil && isFunctionNameInitializerException(v, shadowed) {
		return
	}

	// ignoreOnInitialization: shadow is inside the initializer of the outer binding,
	// and the inner variable's own variableScope is a FunctionExpression / ArrowFunction
	// whose enclosing call is inside that initializer.
	if opts.ignoreOnInitialization && shadowed != nil && isInInitPatternCall(v, shadowed) {
		return
	}

	// hoist modes: in `functions` / `never` / `types` / `functions-and-types`,
	// shadow reports are suppressed when the outer declaration appears *after*
	// the inner declaration (TDZ-like). `all` always reports.
	if shadowed != nil && opts.hoist != hoistAll && isInTdz(v, shadowed, opts.hoist) {
		return
	}

	// TS: ignoreTypeValueShadow
	if shadowed != nil && opts.ignoreTypeValueShadow && isTypeValueShadow(v, shadowed) {
		return
	}

	// TS: ignoreFunctionTypeParameterNameValueShadow
	if opts.ignoreFunctionTypeParameterNameValueShadow && isFunctionTypeParameterShadow(v) {
		return
	}

	// TS: type parameter of a static method shadowing the enclosing class's
	// type parameter is a no-op at runtime and ESLint ignores it.
	if isGenericOfStaticMethod(v) {
		return
	}

	// External declaration merging: type-only import + module augmentation.
	if shadowed != nil && isExternalDeclarationMerging(v, shadowed) {
		return
	}

	if shadowedGlobal && shadowed == nil {
		ctx.ReportNode(v.id, rule.RuleMessage{
			Id:          "noShadowGlobal",
			Description: fmt.Sprintf("'%s' is already a global variable.", v.name),
		})
		return
	}
	if shadowed != nil && shadowed.id != nil {
		line, column := getLineColumn(b.sourceFile, shadowed.id)
		ctx.ReportNode(v.id, rule.RuleMessage{
			Id: "noShadow",
			Description: fmt.Sprintf(
				"'%s' is already declared in the upper scope on line %d column %d.",
				v.name, line, column,
			),
		})
		return
	}
	// Builtin-global match without an identifier (from default library).
	ctx.ReportNode(v.id, rule.RuleMessage{
		Id:          "noShadowGlobal",
		Description: fmt.Sprintf("'%s' is already a global variable.", v.name),
	})
}

// findShadowed walks outward from `start` and returns the first outer
// variable with the same name as `v`, or nil if none is found. Builtin
// globals are handled separately by checkVariable.
func findShadowed(v *variable, start *scope) *variable {
	for cur := start; cur != nil; cur = cur.parent {
		if matches, ok := cur.byName[v.name]; ok {
			return matches[0]
		}
	}
	return nil
}

// isTypeValueShadow mirrors the ESLint/typescript-eslint logic.
// ESLint's check treats the shadowed binding as a "type import" if ANY
// specifier in the same ImportDeclaration is type-only — this is a quirk
// that bubbles the `type` marker across all specifiers of one declaration.
func isTypeValueShadow(v *variable, shadowed *variable) bool {
	isInnerValue := v.isValueBinding

	isTypeImport := shadowed.kind == defImport && importHasAnyTypeOnlySpecifier(shadowed.defNode)
	isShadowedValue := shadowed.isValueBinding && !isTypeImport

	return isInnerValue != isShadowedValue
}

// importHasAnyTypeOnlySpecifier returns true when the ImportDeclaration
// carries either a top-level `type` modifier or a named specifier with
// `import { type X }` syntax.
func importHasAnyTypeOnlySpecifier(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindImportDeclaration {
		return false
	}
	importDecl := node.AsImportDeclaration()
	if importDecl == nil || importDecl.ImportClause == nil {
		return false
	}
	if importDecl.ImportClause.IsTypeOnly() {
		return true
	}
	clause := importDecl.ImportClause.AsImportClause()
	if clause == nil || clause.NamedBindings == nil {
		return false
	}
	if clause.NamedBindings.Kind != ast.KindNamedImports {
		return false
	}
	named := clause.NamedBindings.AsNamedImports()
	if named == nil || named.Elements == nil {
		return false
	}
	for _, elem := range named.Elements.Nodes {
		if elem == nil {
			continue
		}
		spec := elem.AsImportSpecifier()
		if spec != nil && spec.IsTypeOnly {
			return true
		}
	}
	return false
}

// isGenericOfStaticMethod returns true when `v` is a type parameter of a
// method declaration carrying the `static` modifier. ESLint ignores these
// shadows since the static method runs against a class-independent `this`.
func isGenericOfStaticMethod(v *variable) bool {
	if v.kind != defTypeParameter {
		return false
	}
	tp := v.defNode
	if tp == nil {
		return false
	}
	// Walk up the first couple of parents — tsgo may or may not wrap type
	// parameters in a TypeParameterList node depending on the form — and
	// stop at the enclosing method/function node.
	for cur := tp.Parent; cur != nil; cur = cur.Parent {
		switch cur.Kind {
		case ast.KindMethodDeclaration:
			return ast.HasStaticModifier(cur)
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
			ast.KindArrowFunction, ast.KindConstructor,
			ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindClassDeclaration, ast.KindClassExpression,
			ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
			return false
		}
	}
	return false
}

// isFunctionTypeParameterShadow returns true when `v` is a parameter of a
// TS function type / construct signature — its binding lives in a type-level
// position and ESLint ignores these shadows by default.
func isFunctionTypeParameterShadow(v *variable) bool {
	if v.kind != defParameter {
		return false
	}
	p := v.defNode
	if p == nil {
		return false
	}
	parent := p.Parent
	if parent == nil {
		return false
	}
	if ast.IsFunctionTypeNode(parent) || ast.IsConstructorTypeNode(parent) ||
		ast.IsCallSignatureDeclaration(parent) || ast.IsConstructSignatureDeclaration(parent) ||
		ast.IsMethodSignatureDeclaration(parent) {
		return true
	}
	// Bodyless function-like declarations (e.g. `declare function f()`,
	// method signatures on `declare class`, overload signatures) also count.
	if ast.IsFunctionLikeDeclaration(parent) && parent.Body() == nil {
		return true
	}
	return false
}

// isExternalDeclarationMerging covers the `import type Foo from 'bar'` +
// `declare module 'bar' { interface Foo {} }` case.
func isExternalDeclarationMerging(v *variable, shadowed *variable) bool {
	if shadowed.kind != defImport || !shadowed.isTypeOnlyImport {
		return false
	}
	if shadowed.defNode == nil || shadowed.defNode.Kind != ast.KindImportDeclaration {
		return false
	}
	importDecl := shadowed.defNode.AsImportDeclaration()
	if importDecl == nil || importDecl.ModuleSpecifier == nil || !ast.IsStringLiteral(importDecl.ModuleSpecifier) {
		return false
	}
	importSrc := importDecl.ModuleSpecifier.Text()
	mod := ast.FindAncestor(v.id, func(n *ast.Node) bool {
		return n.Kind == ast.KindModuleDeclaration
	})
	if mod == nil {
		return false
	}
	md := mod.AsModuleDeclaration()
	if md == nil || md.Name() == nil {
		return false
	}
	return md.Name().Text() == importSrc
}

// isInTdz tests whether the inner variable appears *before* the outer
// declaration and should therefore be suppressed under the given hoist mode.
func isInTdz(inner *variable, outer *variable, mode hoistMode) bool {
	if inner.id == nil || outer.id == nil {
		return false
	}
	if inner.id.End() >= outer.id.Pos() {
		return false
	}
	switch mode {
	case hoistAll:
		return false
	case hoistTypes:
		// Suppress only for outer type declarations.
		if outer.kind == defType {
			return false
		}
		return true
	case hoistFunctionsAndTypes:
		if outer.kind == defFunctionName || outer.kind == defType {
			return false
		}
		return true
	case hoistFunctions:
		if outer.kind == defFunctionName {
			return false
		}
		return true
	case hoistNever:
		return true
	}
	return false
}

// isFunctionNameInitializerException implements the `var a = function a() {}`
// / `var A = class A {}` / default-destructuring variants that ESLint ignores.
//
// Mirrors ESLint's `isOnInitializer`: requires (a) the inner is a Function-
// Expression name or ClassExpression inner-name, (b) the inner identifier
// sits inside the outer binding's declarator/parameter range, and (c) the
// inner's enclosing scope IS the scope owning the outer binding. Together,
// (b)+(c) handle arbitrary call/decorator wrappers (`wrap(function x() {})`)
// without an AST-walk whitelist, while still rejecting unrelated siblings
// (`const a = 1; const b = function a() {}` — different declarator ranges,
// so (b) fails and we report).
func isFunctionNameInitializerException(inner *variable, outer *variable) bool {
	if outer.defNode == nil || inner.defNode == nil {
		return false
	}
	if inner.kind != defFnExprName && (inner.kind != defClassInnerName || inner.defNode.Kind != ast.KindClassExpression) {
		return false
	}
	if inner.scope == nil || inner.scope.parent == nil || outer.scope == nil {
		return false
	}
	startPos, endPos, ok := outerInitializerLexicalRange(outer.defNode)
	if !ok {
		return false
	}
	expr := inner.defNode
	if startPos > expr.Pos() || expr.End() > endPos {
		return false
	}
	return inner.scope.parent == outer.scope
}

// outerInitializerLexicalRange returns the range ESLint's scope manager would
// expose as the binding's `Definition.parent.range`: the enclosing
// VariableDeclaration for var/let/const + destructuring elements, and the
// enclosing function-like node for parameters.
func outerInitializerLexicalRange(defNode *ast.Node) (int, int, bool) {
	for cur := defNode; cur != nil; cur = cur.Parent {
		switch cur.Kind {
		case ast.KindVariableDeclaration:
			return cur.Pos(), cur.End(), true
		case ast.KindParameter:
			for p := cur.Parent; p != nil; p = p.Parent {
				if ast.IsFunctionLike(p) {
					return p.Pos(), p.End(), true
				}
			}
			return cur.Pos(), cur.End(), true
		}
	}
	return 0, 0, false
}

// isInInitPatternCall handles the `ignoreOnInitialization` option.
// The inner variable's variableScope block must be a function expression or
// arrow whose enclosing call lies inside the outer variable's initializer,
// AND the function's outer variable scope must BE the scope that owns the
// shadowed variable (matching ESLint's `getOuterScope === shadowedVariable.scope`
// check, which prevents suppressing shadows inside nested closures).
func isInInitPatternCall(inner *variable, outer *variable) bool {
	if inner.scope == nil {
		return false
	}
	vs := inner.scope.variableScope()
	if vs == nil || vs.block == nil {
		return false
	}
	if vs.block.Kind != ast.KindFunctionExpression && vs.block.Kind != ast.KindArrowFunction {
		return false
	}
	// The function's immediate outer variable scope must be the same variable
	// scope that owns `outer`. ESLint additionally skips a
	// `function-expression-name` scope between the two; we do the same by
	// peeling it off.
	outerOfFn := vs.parent
	for outerOfFn != nil && outerOfFn.kind == scopeFunctionExprName {
		outerOfFn = outerOfFn.parent
	}
	outerOfFnVS := outerOfFn
	if outerOfFnVS != nil && outerOfFnVS.kind != scopeFunction && outerOfFnVS.kind != scopeModule && outerOfFnVS.kind != scopeGlobal {
		outerOfFnVS = outerOfFnVS.variableScope()
	}
	if outer.scope == nil {
		return false
	}
	outerScopeVS := outer.scope
	if outerScopeVS.kind != scopeFunction && outerScopeVS.kind != scopeModule && outerScopeVS.kind != scopeGlobal {
		outerScopeVS = outerScopeVS.variableScope()
	}
	if outerOfFnVS != outerScopeVS {
		return false
	}
	fn := vs.block
	call := ast.FindAncestor(fn, func(n *ast.Node) bool {
		return n.Kind == ast.KindCallExpression
	})
	if call == nil {
		return false
	}
	location := call.End()
	// Walk ancestors of the outer declaration's identifier.
	node := outer.id
	for node != nil {
		parent := node.Parent
		if parent == nil {
			break
		}
		switch parent.Kind {
		case ast.KindVariableDeclaration:
			vd := parent.AsVariableDeclaration()
			if vd != nil && vd.Initializer != nil && vd.Initializer.Pos() <= location && location <= vd.Initializer.End() {
				return true
			}
			// for-in / for-of expression RHS.
			if parent.Parent != nil && parent.Parent.Parent != nil {
				forStmt := parent.Parent.Parent
				if forStmt.Kind == ast.KindForInStatement || forStmt.Kind == ast.KindForOfStatement {
					fs := forStmt.AsForInOrOfStatement()
					if fs != nil && fs.Expression != nil && fs.Expression.Pos() <= location && location <= fs.Expression.End() {
						return true
					}
				}
			}
			return false
		case ast.KindBindingElement:
			be := parent.AsBindingElement()
			if be != nil && be.Initializer != nil && be.Initializer.Pos() <= location && location <= be.Initializer.End() {
				return true
			}
		case ast.KindParameter:
			init := parent.Initializer()
			if init != nil && init.Pos() <= location && location <= init.End() {
				return true
			}
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
			ast.KindClassDeclaration, ast.KindClassExpression,
			ast.KindArrowFunction, ast.KindCatchClause,
			ast.KindImportDeclaration, ast.KindExportDeclaration,
			ast.KindMethodDeclaration, ast.KindConstructor,
			ast.KindGetAccessor, ast.KindSetAccessor:
			return false
		}
		node = parent
	}
	return false
}

// getLineColumn returns 1-based line and column for the identifier's start position.
func getLineColumn(sf *ast.SourceFile, n *ast.Node) (int, int) {
	if sf == nil || n == nil {
		return 0, 0
	}
	pos := scanner.GetTokenPosOfNode(n, sf, false)
	line, col := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, pos)
	return line + 1, int(col) + 1
}
