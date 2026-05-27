// Package jsx_indent_props ports `@stylistic/jsx-indent-props` to rslint. It
// enforces a consistent indentation style for the props of a JSX element:
// either N spaces / one tab relative to the element's own indent, or visual
// alignment with the column of the first prop (`'first'` mode).
//
// The @stylistic rule is a byte-identical fork of `react/jsx-indent-props` (no
// behavioral delta — their test suites match case-for-case), so the full
// implementation is shared via the react rule's BuildRule; only the registered
// name differs.
package jsx_indent_props

import (
	reactRule "github.com/web-infra-dev/rslint/internal/plugins/react/rules/jsx_indent_props"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// JsxIndentPropsRule is the @stylistic/eslint-plugin variant of
// jsx-indent-props.
var JsxIndentPropsRule rule.Rule = reactRule.BuildRule("@stylistic/jsx-indent-props")
