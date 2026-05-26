// Package jsx_first_prop_new_line ports `@stylistic/jsx-first-prop-new-line` to
// rslint. The @stylistic rule is a byte-identical fork of
// `react/jsx-first-prop-new-line` (no behavioral delta), so the full
// implementation is shared via the react rule's BuildRule; only the registered
// name differs.
package jsx_first_prop_new_line

import (
	reactRule "github.com/web-infra-dev/rslint/internal/plugins/react/rules/jsx_first_prop_new_line"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// JsxFirstPropNewLineRule is the @stylistic/eslint-plugin variant of
// jsx-first-prop-new-line.
var JsxFirstPropNewLineRule rule.Rule = reactRule.BuildRule("@stylistic/jsx-first-prop-new-line")
