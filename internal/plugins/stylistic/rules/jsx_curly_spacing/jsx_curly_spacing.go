// Package jsx_curly_spacing ports `@stylistic/jsx-curly-spacing` to rslint. It
// enforces or disallows whitespace inside the curly braces of JSX attributes
// and expressions: `{ bar }` vs `{bar}`, with separate control for object
// literals (`{{ a: 1 }}`), multiline bodies, and attribute vs children
// positions.
//
// The @stylistic rule is a fork of `react/jsx-curly-spacing`; the two upstream
// rules are case-identical (their test suites match verbatim). The full
// implementation is shared via the react rule's BuildRule. This variant passes
// stylisticScope=true, which selects @stylistic's option-normalization for the
// one combination where the two forks diverge — a per-side empty `spacing: {}`
// alongside a top-level `spacing.objectLiterals` (react inherits the top-level
// value; @stylistic falls back to `when`). See react BuildRule / normalizeConfig
// for the mechanism. Only the registered name and that flag differ.
package jsx_curly_spacing

import (
	reactRule "github.com/web-infra-dev/rslint/internal/plugins/react/rules/jsx_curly_spacing"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// JsxCurlySpacingRule is the @stylistic/eslint-plugin variant of
// jsx-curly-spacing.
var JsxCurlySpacingRule rule.Rule = reactRule.BuildRule("@stylistic/jsx-curly-spacing", true)
