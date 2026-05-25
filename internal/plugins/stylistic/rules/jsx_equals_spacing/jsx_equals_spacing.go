// Package jsx_equals_spacing ports `@stylistic/jsx-equals-spacing` to rslint.
// The @stylistic rule is a byte-identical fork of `react/jsx-equals-spacing`
// (no behavioral delta), so the full implementation is shared via the react
// rule's BuildRule; only the registered name differs.
package jsx_equals_spacing

import (
	reactRule "github.com/web-infra-dev/rslint/internal/plugins/react/rules/jsx_equals_spacing"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// JsxEqualsSpacingRule is the @stylistic/eslint-plugin variant of
// jsx-equals-spacing.
var JsxEqualsSpacingRule rule.Rule = reactRule.BuildRule("@stylistic/jsx-equals-spacing")
