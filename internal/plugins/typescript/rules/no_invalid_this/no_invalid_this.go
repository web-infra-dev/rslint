package no_invalid_this

import (
	_ "embed"

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
// via `isDefaultThisBinding`, JSDoc `@this`, computed-key deferral,
// method-decorator handling, `capIsConstructor` at every uppercase-name
// branch) — are the exact same algorithm ESLint core's own `no-invalid-this`
// uses: core natively recognizes both the `this` parameter and `accessor`
// fields as of the version this port targets, which is what let this rule
// become a thin wrapper around internal/rules/no_invalid_this.BuildListeners
// rather than a second copy of the walker. See that package's BuildListeners
// doc comment for the field-frame policy where the rules genuinely differ.
//
// https://typescript-eslint.io/rules/no-invalid-this
var NoInvalidThisRule = rule.CreateRule(rule.Rule{
	Name:   "no-invalid-this",
	Schema: rule.NewSchema(schemaJSON),
	Run:    run,
})

func run(ctx rule.RuleContext, options []any) rule.RuleListeners {
	opts := no_invalid_this_core.ParseOptions(options)
	if ctx.LanguageOptions.SourceType == "" {
		// The upstream TypeScript RuleTester defaults to module semantics.
		// Production rslint contexts already carry an effective source type.
		ctx.LanguageOptions.SourceType = "module"
	}
	engineOptions := no_invalid_this_core.TypeScriptEngineOptions(ctx, opts)
	engineOptions.FieldFrameScopedToValue = false
	return no_invalid_this_core.BuildListeners(ctx, engineOptions)
}
