package jsx_uses_react

import (
	"github.com/web-infra-dev/rslint/internal/rule"
)

// JsxUsesReactRule is a no-op in rslint.
//
// In ESLint, this rule prevents React from being marked as unused when JSX is
// used, since ESLint's no-unused-vars rule does not understand that JSX
// implicitly references React. In rslint, the TypeScript type checker already
// tracks JSX factory usage and correctly marks React as used, so this rule is
// unnecessary. It exists only for configuration compatibility.
var JsxUsesReactRule = rule.Rule{
	Name: "react/jsx-uses-react",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{}
	},
}
