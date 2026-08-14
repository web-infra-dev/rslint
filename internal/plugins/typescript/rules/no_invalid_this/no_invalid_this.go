package no_invalid_this

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	no_invalid_this_core "github.com/web-infra-dev/rslint/internal/rules/no_invalid_this"
)

//go:embed no_invalid_this.schema.json
var schemaJSON []byte

// NoInvalidThisRule mirrors @typescript-eslint/no-invalid-this, which wraps
// ESLint core's no-invalid-this with two TypeScript-specific recognitions:
//   - A function whose signature declares a `this` parameter
//     (`function foo(this: T)`) has its `this` validity short-circuited to
//     `true`, matching the upstream `thisIsValidStack` push.
//   - Class field initializers (regular and `accessor`-modified) are
//     short-circuited to `true`, since their implicit-function context
//     binds `this` to the class instance.
//
// Both recognitions — along with every other validity decision (parent-walk
// via `isDefaultThisBinding`, JSDoc `@this`, computed-key deferral, decorator
// handling, `capIsConstructor` at every uppercase-name branch) — are the
// exact same algorithm ESLint core's own `no-invalid-this` uses: core
// natively recognizes both the `this` parameter and `accessor` fields as of
// the version this port targets, which is what let this rule become a thin
// wrapper around internal/rules/no_invalid_this.BuildListeners rather than a
// second copy of the walker. See that package's BuildListeners doc comment
// for the two policy points where the rules genuinely differ (strict-mode
// gating and top-level validity).
//
// https://typescript-eslint.io/rules/no-invalid-this
var NoInvalidThisRule = rule.CreateRule(rule.Rule{
	Name:   "no-invalid-this",
	Schema: rule.NewSchema(schemaJSON),
	Run:    run,
})

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	opts := no_invalid_this_core.ParseOptions(options)
	return no_invalid_this_core.BuildListeners(ctx, no_invalid_this_core.EngineOptions{
		CapIsConstructor: opts.CapIsConstructor,
		// typescript-eslint's wrapper defaults to `parserOptions.sourceType:
		// 'module'` (its RuleTester's own default), which makes top-level
		// `this` always invalid and every function frame always strict.
		// rslint does not expose `parserOptions.sourceType` as an override
		// independent of file content, so this port adopts the same
		// always-module/always-strict default rather than deriving it from
		// `ast.IsExternalModule` the way the bare `no-invalid-this` rule
		// does — a framework-layer consequence of rslint not surfacing
		// parser options, applied uniformly across rules.
		TopLevelValid: false,
		IsStrict:      func(*ast.Node, *ast.SourceFile) bool { return true },
	})
}
