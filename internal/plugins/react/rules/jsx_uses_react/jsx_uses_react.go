package jsx_uses_react

import (
	"github.com/web-infra-dev/rslint/internal/rule"
)

// JsxUsesReactRule is a no-op in rslint.
//
// In ESLint, this rule (eslint-plugin-react) marks the JSX pragma as used
// whenever JSX appears, so no-unused-vars does not flag it. Native
// no-unused-vars records implicit factory references for classic JSX itself,
// using file pragmas, then explicit React settings, then compiler options.
// It resolves each reference in the JSX node's lexical scope and supports
// custom factories and fragment factories. Automatic JSX does not consume
// the classic factory. This rule exists for configuration compatibility;
// enabling it alone does not select the JSX runtime.
var JsxUsesReactRule = rule.Rule{
	Name:   "react/jsx-uses-react",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{}
	},
}
