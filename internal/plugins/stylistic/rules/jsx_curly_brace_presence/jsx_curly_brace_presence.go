// Package jsx_curly_brace_presence ports `@stylistic/jsx-curly-brace-presence`
// to rslint. The rule is a fork of `react/jsx-curly-brace-presence`; the only
// behavioral delta is in the unnecessary-curly check: @stylistic leaves an
// attribute string literal whose value contains a quote character untouched
// (its guard is `isJSX(parent) || !containsQuoteCharacters(value)`), whereas
// eslint-plugin-react reports there regardless. The full implementation is
// shared via the react rule's BuildRule; this package selects the @stylistic
// variant with stylisticQuotes=true.
package jsx_curly_brace_presence

import (
	reactRule "github.com/web-infra-dev/rslint/internal/plugins/react/rules/jsx_curly_brace_presence"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// JsxCurlyBracePresenceRule is the @stylistic/eslint-plugin variant of
// jsx-curly-brace-presence.
var JsxCurlyBracePresenceRule rule.Rule = reactRule.BuildRule("@stylistic/jsx-curly-brace-presence", true)
