package no_use_before_define

import (
	"cmp"
	_ "embed"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

//go:embed no_use_before_define.schema.json
var schemaJSON []byte

// https://typescript-eslint.io/rules/no-use-before-define
//
// Scope semantics come from the shared scope model in
// `internal/utils/scope`, the same one the ESLint core `no-use-before-define`
// rule uses. The two rules deliberately do NOT share a body: typescript-eslint
// extends an older core rule and its decision logic differs in ways that are
// observable, not incidental —
//
//   - it has no class-definition-evaluation handling at all, so `class C
//     extends C {}` and `class C { [C](){} }` are not reported;
//   - a class field initializer and a class static block are plain separate
//     variable scopes, with none of core's "static initializers run during
//     class definition" folding;
//   - `ignoreTypeReferences` covers every type position, not just a bare type
//     annotation;
//   - a reference from a function/constructor type or call/construct/method
//     signature is never reported;
//   - the options are consulted as an ordered chain in which a function
//     declaration short-circuits the rest.

// options mirrors the typescript-eslint rule schema.
// See https://typescript-eslint.io/rules/no-use-before-define/#options
type options struct {
	functions            bool
	classes              bool
	variables            bool
	enums                bool
	typedefs             bool
	ignoreTypeReferences bool
	allowNamedExports    bool
}

func parseOptions(rawOptions []any) options {
	opts := options{
		functions:            true,
		classes:              true,
		variables:            true,
		enums:                true,
		typedefs:             true,
		ignoreTypeReferences: true,
		allowNamedExports:    false,
	}
	if len(rawOptions) == 0 {
		return opts
	}

	// Handle "nofunc" string option.
	if str, ok := rawOptions[0].(string); ok {
		if str == "nofunc" {
			opts.functions = false
		}
		return opts
	}

	optsMap, _ := rawOptions[0].(map[string]interface{})
	readBool := func(key string, target *bool) {
		if v, ok := optsMap[key].(bool); ok {
			*target = v
		}
	}
	readBool("functions", &opts.functions)
	readBool("classes", &opts.classes)
	readBool("variables", &opts.variables)
	readBool("enums", &opts.enums)
	readBool("typedefs", &opts.typedefs)
	readBool("ignoreTypeReferences", &opts.ignoreTypeReferences)
	readBool("allowNamedExports", &opts.allowNamedExports)

	return opts
}

var NoUseBeforeDefineRule = rule.CreateRule(rule.Rule{
	Name:   "no-use-before-define",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		if ctx.SourceFile == nil {
			return rule.RuleListeners{}
		}

		manager := scope.Build(ctx.SourceFile, scope.Options{CollectReferences: true})

		// Upstream walks the scope tree depth-first; ordering the flat
		// reference list by source position gives the same sequence for every
		// input the rule can report on, without depending on scope-creation
		// order.
		references := slices.Clone(manager.References)
		slices.SortStableFunc(references, func(a, b *scope.Reference) int {
			return cmp.Compare(a.Identifier.Pos(), b.Identifier.Pos())
		})

		for _, ref := range references {
			check(ctx, opts, ref)
		}

		return rule.RuleListeners{}
	},
})

func check(ctx rule.RuleContext, opts options, ref *scope.Reference) {
	declaration := ref.Resolved()
	definitionIdentifier := ref.ResolvedIdentifier()

	// A named export gets its own option and a plain positional check — none
	// of the option chain below applies to it.
	if isNamedExport(ref.Identifier) {
		if opts.allowNamedExports {
			return
		}
		if declaration == nil || definitionIdentifier == nil ||
			!isDefinedBeforeUse(declaration, definitionIdentifier, ref) {
			report(ctx, ref.Identifier)
		}
		return
	}

	// Definitions without identifiers — string-literal enum members, for
	// example — still participate in resolution. Upstream skips the binding
	// only when none of its merged definitions supplies an identifier.
	if declaration == nil || definitionIdentifier == nil {
		return
	}
	if isDefinedBeforeUse(declaration, definitionIdentifier, ref) {
		return
	}
	if !isForbidden(opts, declaration, ref) {
		return
	}
	if isClassRefInClassDecorator(declaration, ref) {
		return
	}
	// A function type's parameter list has no runtime evaluation order, so a
	// reference from inside one is never "before" anything.
	if isFunctionTypeScope(ref.From) {
		return
	}

	report(ctx, ref.Identifier)
}

// isDefinedBeforeUse reports whether the declaration is already in place when
// the reference runs: it ends at or before the reference, and the reference is
// not a value read from inside the declaration's own initializer.
func isDefinedBeforeUse(declaration *scope.Variable, definitionIdentifier *ast.Node, ref *scope.Reference) bool {
	if definitionIdentifier.End() > ref.Identifier.End() {
		return false
	}
	// A value read from inside the declaration's own initializer runs before
	// the binding is set up, even though it sits after the name.
	return !ref.IsValueReference() || !isInInitializer(declaration, definitionIdentifier, ref)
}

// isInInitializer applies upstream's `variable.scope !== reference.from`
// precondition before the positional walk: a reference that crosses a scope
// boundary is never treated as part of the declaration's initialization, even
// when it sits inside the initializer's source range.
func isInInitializer(declaration *scope.Variable, definitionIdentifier *ast.Node, ref *scope.Reference) bool {
	if declaration.Scope != ref.From {
		return false
	}
	return scope.IsInsideOwnInitializer(definitionIdentifier, ref.Identifier.End())
}

// isForbidden decides whether a use-before-define should be reported, based on
// the options and what kind of declaration was referenced.
//
// The chain is ordered and short-circuits, matching upstream: a function
// declaration is decided by `functions` alone and never falls through to the
// later arms. For classes, variables, and enums the option only suppresses a
// reference from a different variable scope; a same-scope reference is a
// temporal dead zone error and is always reported.
func isForbidden(opts options, declaration *scope.Variable, ref *scope.Reference) bool {
	if opts.ignoreTypeReferences && isTypeReference(ref) {
		return false
	}

	switch {
	case isFunctionNameDef(declaration):
		return opts.functions
	case isClassNameDef(declaration) && isFromOuterVariableScope(declaration, ref):
		return opts.classes
	case declaration.Kind == scope.DefVariable && isFromOuterVariableScope(declaration, ref):
		return opts.variables
	case declaration.Kind == scope.DefEnumName && isFromOuterVariableScope(declaration, ref):
		return opts.enums
	case declaration.Kind == scope.DefType:
		return opts.typedefs
	}

	return true
}

func isFunctionNameDef(v *scope.Variable) bool {
	return v.Kind == scope.DefFunctionName || v.Kind == scope.DefFnExprName
}

func isClassNameDef(v *scope.Variable) bool {
	return v.Kind == scope.DefClassName || v.Kind == scope.DefClassInnerName
}

// isFromOuterVariableScope reports whether the reference is evaluated in a
// different variable scope than the one holding the declaration. Unlike the
// ESLint core rule, class field initializers and static blocks count as
// ordinary separate scopes here — upstream does no static-initializer folding.
func isFromOuterVariableScope(declaration *scope.Variable, ref *scope.Reference) bool {
	if declaration.Scope == nil || ref.From == nil {
		return false
	}
	return declaration.Scope.VariableScope() != ref.From.VariableScope()
}

// isNamedExport reports whether the identifier is the local name of an export
// specifier — the `a` in `export { a }` or `export { a as b }`. The scope model
// only ever creates a reference for that half of a specifier.
func isNamedExport(node *ast.Node) bool {
	return node.Parent != nil && node.Parent.Kind == ast.KindExportSpecifier
}

// isFunctionTypeScope reports whether a scope is one of scope-manager's
// `functionType` scopes: a function/constructor type or call/construct/method
// signature.
func isFunctionTypeScope(s *scope.Scope) bool {
	return s != nil && s.Kind == scope.KindFunctionType
}

// isTypeReference mirrors upstream's effective predicate. Most type-space
// information comes directly from the shared Reference; a type query is the
// one syntax form upstream additionally recognizes from the reference's AST.
func isTypeReference(ref *scope.Reference) bool {
	return ref != nil && (ref.IsTypeReference() || ast.IsPartOfTypeQuery(ref.Identifier))
}

// isClassRefInClassDecorator reports whether a reference to a class binding
// appears in one of that class's own decorators. Decorators are applied after
// the class is defined, so the binding is already initialized.
func isClassRefInClassDecorator(declaration *scope.Variable, ref *scope.Reference) bool {
	if !isClassNameDef(declaration) || declaration.DefNode == nil {
		return false
	}
	for _, decorator := range declaration.DefNode.Decorators() {
		if decorator == nil {
			continue
		}
		if ref.Identifier.Pos() >= decorator.Pos() && ref.Identifier.End() <= decorator.End() {
			return true
		}
	}
	return false
}

func report(ctx rule.RuleContext, node *ast.Node) {
	ctx.ReportNode(node, rule.RuleMessage{
		Id:          "noUseBeforeDefine",
		Description: "'" + node.Text() + "' was used before it was defined.",
	})
}
