package jsx_uses_react

import (
	"github.com/web-infra-dev/rslint/internal/rule"
)

// JsxUsesReactRule is a no-op in rslint.
//
// In ESLint, this rule (eslint-plugin-react) marks the JSX pragma as used
// whenever JSX appears, so no-unused-vars does not flag it. rslint handles the
// default pragma inside no-unused-vars itself: markJsxFactoryUsed marks the JSX
// factory import (default "React", or the tsconfig `jsxFactory`) as used
// whenever a file contains JSX, in any jsx runtime — matching
// @typescript-eslint/parser's `jsxPragma` baseline. So this rule is redundant
// for the default pragma and exists only for configuration compatibility.
//
// Known gap: a custom pragma declared via an `@jsx X` comment or
// `settings.react.pragma` is not marked as used (rslint has no cross-rule
// "mark as used" channel yet), so such an import may still be reported as
// unused even with this rule enabled.
var JsxUsesReactRule = rule.Rule{
	Name: "react/jsx-uses-react",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{}
	},
}
