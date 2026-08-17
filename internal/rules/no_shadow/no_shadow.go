package no_shadow

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

//go:embed no_shadow.schema.json
var schemaJSON []byte

// https://eslint.org/docs/latest/rules/no-shadow
//
// Scope semantics come from the shared scope model in
// `internal/utils/scope`, which reconstructs ESLint's scope tree by walking
// the AST (rslint has no eslint-scope equivalent). This covers the common
// cases exercised by the ESLint test suite. The rule reports shadowing against
// declarations visible within the file plus, when `builtinGlobals` is on, the
// effective ESLint global scope from `ctx.Globals`: ECMAScript builtins and
// globals declared via config `languageOptions.globals` or `/* global */`
// comments. The typescript-eslint variant also includes scope-manager's
// default TypeScript type globals. Concepts rslint does not expose (for
// example `parserOptions.globalReturn`) remain unmodeled.

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

type ruleVariant struct {
	defaults                            options
	includeDefaultTypeScriptTypeGlobals bool
	typeImportUsesOwnSpecifier          bool
	reportEnumShadow                    bool
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

func parseOptionsWith(rawOptions []any, opts options) options {
	// Always copy the allow map: the caller's `opts` may be a long-lived
	// defaults instance shared across rule invocations (e.g. the closure
	// captured by `runWithVariant`). Mutating in-place would leak state
	// from one source file's lint run to the next.
	src := opts.allow
	opts.allow = make(map[string]bool, len(src)+4)
	for k, v := range src {
		opts.allow[k] = v
	}
	if len(rawOptions) == 0 {
		return opts
	}
	optsMap, _ := rawOptions[0].(map[string]interface{})
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
// Rule entry
// ---------------------------------------------------------------------------

var NoShadowRule = rule.Rule{
	Name:   "no-shadow",
	Schema: rule.NewSchema(schemaJSON),
	Run: runWithVariant(ruleVariant{
		defaults: defaultOptions(),
	}),
}

// RunTSESLint exposes the rule body with typescript-eslint's defaults so the
// `@typescript-eslint/no-shadow` wrapper can reuse the implementation. The
// underlying closure is built once at package init — `parseOptionsWith`
// copies the captured `allow` map per invocation, so this is safe.
var runTSESLint = runWithVariant(ruleVariant{
	defaults:                            defaultOptionsTSESLint(),
	includeDefaultTypeScriptTypeGlobals: true,
	typeImportUsesOwnSpecifier:          true,
	reportEnumShadow:                    true,
})

func RunTSESLint(ctx rule.RuleContext, options []any) rule.RuleListeners {
	return runTSESLint(ctx, options)
}

// checker carries the per-run configuration that `checkVariable` needs.
type checker struct {
	sourceFile                          *ast.SourceFile
	builtinGlobals                      map[string]bool
	includeDefaultTypeScriptTypeGlobals bool
	typeImportUsesOwnSpecifier          bool
	reportEnumShadow                    bool
}

func runWithVariant(variant ruleVariant) func(rule.RuleContext, []any) rule.RuleListeners {
	return func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptionsWith(rawOptions, variant.defaults)
		if ctx.SourceFile == nil {
			return rule.RuleListeners{}
		}

		manager := scope.Build(ctx.SourceFile, scope.Options{})

		c := &checker{
			sourceFile:                          ctx.SourceFile,
			includeDefaultTypeScriptTypeGlobals: variant.includeDefaultTypeScriptTypeGlobals,
			typeImportUsesOwnSpecifier:          variant.typeImportUsesOwnSpecifier,
			reportEnumShadow:                    variant.reportEnumShadow,
		}

		filename := ""
		if ctx.SourceFile.AsNode() != nil {
			filename = ctx.SourceFile.FileName()
		}
		isDeclFile := strings.HasSuffix(filename, ".d.ts") ||
			strings.HasSuffix(filename, ".d.cts") ||
			strings.HasSuffix(filename, ".d.mts")

		// Build the configured and language-default globals from ESLint's
		// effective global scope. The TypeScript variant checks scope-manager's
		// default type globals separately; host-library value symbols such as
		// window/top/console are intentionally not inferred from a TypeChecker.
		builtinGlobals := map[string]bool{}
		if opts.builtinGlobals {
			ctx.Globals.ApplyTo(builtinGlobals)
		}
		c.builtinGlobals = builtinGlobals

		// Walk scopes top-down and check each variable. The global scope is
		// included so that `builtinGlobals: true` can flag a module-level
		// declaration shadowing an ECMAScript global (ESLint's module scope
		// sits between the file and the global scope; we collapse them and
		// compensate by letting checkVariable consult the globals table).
		for _, s := range manager.Scopes {
			if s.GlobalAugmentation {
				continue
			}
			for _, v := range s.Vars {
				if v.Anonymous {
					continue
				}
				if opts.allow[v.Name] {
					continue
				}
				if v.Name == "this" {
					continue
				}
				if isDuplicatedClassNameInClassScope(v) {
					continue
				}
				if isDeclFile && v.DeclareModifier {
					continue
				}
				c.checkVariable(ctx, s, v, opts)
			}
		}

		return rule.RuleListeners{}
	}
}

// isDuplicatedClassNameInClassScope suppresses the inner class-name binding
// that ESLint-scope adds for ClassDeclarations.
func isDuplicatedClassNameInClassScope(v *scope.Variable) bool {
	return v.Kind == scope.DefClassInnerName && v.DefNode != nil && v.DefNode.Kind == ast.KindClassDeclaration
}

// checkVariable tests whether `v` shadows a variable in some outer scope.
func (c *checker) checkVariable(ctx rule.RuleContext, s *scope.Scope, v *scope.Variable, opts options) {
	shadowed := findShadowed(v, s.Parent)
	// An outer binding with no identifier — a string-literal enum member —
	// answers the name lookup but has no declaration site to name in the
	// report, so nothing is reported against it.
	if shadowed != nil && shadowed.Anonymous {
		return
	}
	shadowedDefaultTypeScriptGlobal := shadowed == nil && opts.builtinGlobals &&
		c.includeDefaultTypeScriptTypeGlobals && rule.IsDefaultTypeScriptTypeGlobal(v.Name)
	shadowedGlobal := shadowed == nil && opts.builtinGlobals &&
		(c.builtinGlobals[v.Name] || shadowedDefaultTypeScriptGlobal)
	if shadowed == nil && !shadowedGlobal {
		return
	}
	if shadowedGlobal {
		merged, first := mergeImplicitGlobalShadow(s, v)
		if !first {
			return
		}
		v = merged
	}

	// Ignore function-name-initializer exceptions:
	// var a = function a() {};  /  var A = class A {};
	if shadowed != nil && isFunctionNameInitializerException(v, shadowed) {
		return
	}

	// ignoreOnInitialization: shadow is inside the initializer of the outer binding,
	// and the inner variable's own variable scope is a FunctionExpression / ArrowFunction
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
	if opts.ignoreTypeValueShadow && isTypeValueShadow(v, shadowed, c.typeImportUsesOwnSpecifier) {
		return
	}

	// TS: ignoreFunctionTypeParameterNameValueShadow
	if opts.ignoreFunctionTypeParameterNameValueShadow &&
		isFunctionTypeParameterNameValueShadow(v, shadowed, shadowedDefaultTypeScriptGlobal) {
		return
	}

	// TS: a static method's type parameter shadowing its enclosing class's
	// type parameter is a special exception. Other outer bindings still count.
	if isGenericOfAStaticMethodShadow(v, shadowed) {
		return
	}

	// External declaration merging: type-only import + module augmentation.
	if shadowed != nil && isExternalDeclarationMerging(v, shadowed) {
		return
	}

	if shadowedGlobal && shadowed == nil {
		ctx.ReportNode(v.ID, rule.RuleMessage{
			Id:          "noShadowGlobal",
			Description: fmt.Sprintf("'%s' is already a global variable.", v.Name),
		})
		return
	}
	if shadowed != nil && shadowed.ID != nil {
		line, column := getLineColumn(c.sourceFile, shadowed.ID)
		if c.reportEnumShadow && hasDefinitionKind(shadowed, scope.DefEnumName) {
			ctx.ReportNode(v.ID, rule.RuleMessage{
				Id: "noEnumShadow",
				Description: fmt.Sprintf(
					"Enum members are added to the enum scope, so references to '%s' in enum member initializers resolve to this member instead of the declaration in the upper scope on line %d column %d.",
					v.Name, line, column,
				),
			})
			return
		}
		ctx.ReportNode(v.ID, rule.RuleMessage{
			Id: "noShadow",
			Description: fmt.Sprintf(
				"'%s' is already declared in the upper scope on line %d column %d.",
				v.Name, line, column,
			),
		})
		return
	}
	// Builtin-global match without an identifier (from default library).
	ctx.ReportNode(v.ID, rule.RuleMessage{
		Id:          "noShadowGlobal",
		Description: fmt.Sprintf("'%s' is already a global variable.", v.Name),
	})
}

// hasDefinitionKind mirrors scope-manager's merged Variable definitions. The
// shared scope model stores each declaration separately, so inspect all
// same-name declarations when upstream asks whether any definition is an enum.
func hasDefinitionKind(v *scope.Variable, kind scope.DefKind) bool {
	if v == nil || v.Scope == nil {
		return false
	}
	for _, definition := range v.Scope.Declarations(v.Name) {
		if definition.Kind == kind {
			return true
		}
	}
	return false
}

// mergeImplicitGlobalShadow models scope-manager's one-variable-per-name
// representation. TypeScript declaration merging can add several definitions
// to that variable; no-shadow still checks it once, reports its first
// identifier, and treats it as a value if any definition is value-capable.
func mergeImplicitGlobalShadow(s *scope.Scope, v *scope.Variable) (*scope.Variable, bool) {
	definitions := s.Declarations(v.Name)
	if len(definitions) == 0 || definitions[0] != v {
		return v, false
	}
	if len(definitions) == 1 {
		return v, true
	}

	merged := *v
	for _, definition := range definitions {
		if isValueVariable(definition) {
			merged.IsValueBinding = true
			break
		}
	}
	return &merged, true
}

// findShadowed walks outward from `start` and returns the first outer
// variable with the same name as `v`, or nil if none is found. Builtin
// globals are handled separately by checkVariable.
func findShadowed(v *scope.Variable, start *scope.Scope) *scope.Variable {
	for cur := start; cur != nil; cur = cur.Parent {
		if matches := cur.Declarations(v.Name); len(matches) > 0 {
			return matches[0]
		}
	}
	return nil
}

// isTypeValueShadow mirrors the typescript-eslint logic. A nil shadowed
// variable represents an ESLint implicit global: scope-manager variables with
// no definition are treated as values for this option, even when the global
// itself is type-only. Import bindings are also value variables in
// scope-manager, including `import type` bindings.
// ESLint core treats the shadowed binding as a type import when any specifier
// in its ImportDeclaration is type-only. The typescript-eslint extension only
// checks the current binding's own specifier; the variant flag preserves both
// upstream behaviors.
func isTypeValueShadow(v *scope.Variable, shadowed *scope.Variable, typeImportUsesOwnSpecifier bool) bool {
	isInnerValue := isValueVariable(v)
	if shadowed == nil {
		return !isInnerValue
	}

	isTypeImport := shadowed.Kind == scope.DefImport &&
		((typeImportUsesOwnSpecifier && shadowed.IsTypeOnlyImport) ||
			(!typeImportUsesOwnSpecifier && importHasAnyTypeOnlySpecifier(shadowed.DefNode)))
	isShadowedValue := isValueVariable(shadowed) && !isTypeImport

	return isInnerValue != isShadowedValue
}

// isValueVariable translates the scope model's declaration metadata to
// scope-manager's isValueVariable flag. Scope-manager considers every import
// binding value-capable here, including `import type` bindings.
func isValueVariable(v *scope.Variable) bool {
	return v != nil && (v.IsValueBinding || v.Kind == scope.DefImport)
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

// isGenericOfStaticMethod reports whether `v` is a type parameter declared by
// a static method. The caller still has to verify that the shadowed binding is
// the enclosing class's type parameter.
func isGenericOfStaticMethod(v *scope.Variable) bool {
	if v.Kind != scope.DefTypeParameter {
		return false
	}
	tp := v.DefNode
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

func isGenericOfClass(v *scope.Variable) bool {
	if v == nil || v.Kind != scope.DefTypeParameter || v.DefNode == nil {
		return false
	}
	for cur := v.DefNode.Parent; cur != nil; cur = cur.Parent {
		switch cur.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			return true
		case ast.KindMethodDeclaration, ast.KindFunctionDeclaration,
			ast.KindFunctionExpression, ast.KindArrowFunction:
			return false
		}
	}
	return false
}

func isGenericOfAStaticMethodShadow(v *scope.Variable, shadowed *scope.Variable) bool {
	return shadowed != nil && isGenericOfStaticMethod(v) && isGenericOfClass(shadowed)
}

// isFunctionTypeParameterShadow returns true when `v` is a parameter of a
// TS function type / construct signature. Whether the option ignores it also
// depends on the shadowed variable being value-capable.
func isFunctionTypeParameterShadow(v *scope.Variable) bool {
	if v.Kind != scope.DefParameter {
		return false
	}
	p := v.DefNode
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

func isFunctionTypeParameterNameValueShadow(v *scope.Variable, shadowed *scope.Variable, shadowedDefaultTypeScriptGlobal bool) bool {
	if !isFunctionTypeParameterShadow(v) {
		return false
	}
	if shadowed == nil {
		// TypeScript default globals have an isValueVariable field set to
		// false. Other ESLint implicit globals have no such field and are
		// treated as values by the upstream rule.
		return !shadowedDefaultTypeScriptGlobal
	}
	return isValueVariable(shadowed)
}

// isExternalDeclarationMerging covers the `import type Foo from 'bar'` +
// `declare module 'bar' { interface Foo {} }` case.
func isExternalDeclarationMerging(v *scope.Variable, shadowed *scope.Variable) bool {
	if shadowed.Kind != scope.DefImport || !shadowed.IsTypeOnlyImport {
		return false
	}
	if shadowed.DefNode == nil || shadowed.DefNode.Kind != ast.KindImportDeclaration {
		return false
	}
	importDecl := shadowed.DefNode.AsImportDeclaration()
	if importDecl == nil || importDecl.ModuleSpecifier == nil || !ast.IsStringLiteral(importDecl.ModuleSpecifier) {
		return false
	}
	importSrc := importDecl.ModuleSpecifier.Text()
	mod := ast.FindAncestor(v.ID, func(n *ast.Node) bool {
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
func isInTdz(inner *scope.Variable, outer *scope.Variable, mode hoistMode) bool {
	if inner.ID == nil || outer.ID == nil {
		return false
	}
	if inner.ID.End() >= outer.ID.Pos() {
		return false
	}
	switch mode {
	case hoistAll:
		return false
	case hoistTypes:
		// Suppress only for outer type declarations.
		if outer.Kind == scope.DefType {
			return false
		}
		return true
	case hoistFunctionsAndTypes:
		if outer.Kind == scope.DefFunctionName || outer.Kind == scope.DefType {
			return false
		}
		return true
	case hoistFunctions:
		if outer.Kind == scope.DefFunctionName {
			return false
		}
		return true
	case hoistNever:
		return true
	}
	return false
}

// isFunctionNameInitializerException implements the direct initializer cases
// that ESLint ignores: `var a = function a() {}`, `var A = class A {}`, and
// their logical/conditional/default-value variants. Calls and TypeScript
// assertion expressions are not transparent, so `wrap(function a() {})` and
// `function a() {} as unknown` still report.
func isFunctionNameInitializerException(inner *scope.Variable, outer *scope.Variable) bool {
	if outer == nil || outer.ID == nil || inner == nil || inner.DefNode == nil {
		return false
	}
	if inner.Kind != scope.DefFnExprName && (inner.Kind != scope.DefClassInnerName || inner.DefNode.Kind != ast.KindClassExpression) {
		return false
	}
	initializer := bindingInitializer(outer.ID)
	if initializer == nil {
		return false
	}
	nodeToCheck := inner.DefNode
	if initializer.Pos() > nodeToCheck.Pos() || nodeToCheck.End() > initializer.End() {
		return false
	}
	return initializer == unwrapInitializerExpression(nodeToCheck)
}

// bindingInitializer returns the default/initializer attached to this exact
// binding. Stopping at the nearest binding element is important for sibling
// destructuring entries: the initializer of `b` is not the initializer of `a`.
func bindingInitializer(identifier *ast.Node) *ast.Node {
	for cur := identifier.Parent; cur != nil; cur = cur.Parent {
		switch cur.Kind {
		case ast.KindBindingElement:
			binding := cur.AsBindingElement()
			if binding == nil {
				return nil
			}
			return binding.Initializer
		case ast.KindVariableDeclaration:
			declaration := cur.AsVariableDeclaration()
			if declaration == nil {
				return nil
			}
			return declaration.Initializer
		case ast.KindParameter:
			parameter := cur.AsParameterDeclaration()
			if parameter == nil {
				return nil
			}
			return parameter.Initializer
		}
	}
	return nil
}

// unwrapInitializerExpression follows only the expression shapes that ESTree
// treats as transparently capable of evaluating to the child: parentheses,
// logical operands, and the two result branches of a conditional expression.
func unwrapInitializerExpression(node *ast.Node) *ast.Node {
	current := node
	for current != nil && current.Parent != nil {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			parenthesized := parent.AsParenthesizedExpression()
			if parenthesized == nil || parenthesized.Expression != current {
				return current
			}
			current = parent
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary == nil || binary.OperatorToken == nil ||
				(binary.Left != current && binary.Right != current) {
				return current
			}
			switch binary.OperatorToken.Kind {
			case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
				current = parent
			default:
				return current
			}
		case ast.KindConditionalExpression:
			conditional := parent.AsConditionalExpression()
			if conditional == nil || (conditional.WhenTrue != current && conditional.WhenFalse != current) {
				return current
			}
			current = parent
		default:
			return current
		}
	}
	return current
}

// isInInitPatternCall handles the `ignoreOnInitialization` option.
// The inner variable's variable-scope block must be a function expression or
// arrow whose enclosing call lies inside the outer variable's initializer,
// AND the function's outer variable scope must BE the scope that owns the
// shadowed variable (matching ESLint's `getOuterScope === shadowedVariable.scope`
// check, which prevents suppressing shadows inside nested closures).
func isInInitPatternCall(inner *scope.Variable, outer *scope.Variable) bool {
	if inner.Scope == nil {
		return false
	}
	vs := inner.Scope.VariableScope()
	if vs == nil || vs.Block == nil {
		return false
	}
	if vs.Block.Kind != ast.KindFunctionExpression && vs.Block.Kind != ast.KindArrowFunction {
		return false
	}
	// The function's immediate outer variable scope must be the same variable
	// scope that owns `outer`. ESLint additionally skips a
	// `function-expression-name` scope between the two; we do the same by
	// peeling it off.
	outerOfFn := vs.Parent
	for outerOfFn != nil && outerOfFn.Kind == scope.KindFunctionExprName {
		outerOfFn = outerOfFn.Parent
	}
	outerOfFnVS := outerOfFn
	if outerOfFnVS != nil {
		outerOfFnVS = outerOfFnVS.VariableScope()
	}
	if outer.Scope == nil {
		return false
	}
	outerScopeVS := outer.Scope.VariableScope()
	if outerOfFnVS != outerScopeVS {
		return false
	}
	fn := vs.Block
	call := ast.FindAncestor(fn, func(n *ast.Node) bool {
		return n.Kind == ast.KindCallExpression
	})
	if call == nil {
		return false
	}
	location := call.End()
	// Walk ancestors of the outer declaration's identifier.
	node := outer.ID
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
