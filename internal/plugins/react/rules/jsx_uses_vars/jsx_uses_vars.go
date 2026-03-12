package jsx_uses_vars

import (
	"github.com/web-infra-dev/rslint/internal/rule"
)

// JsxUsesVarsRule is a no-op in rslint.
//
// In ESLint, this rule prevents variables used in JSX from being marked as
// unused by the no-unused-vars rule, since ESLint does not natively understand
// JSX component references. In rslint, the TypeScript type checker already
// tracks all JSX references and correctly marks variables as used, so this
// rule is unnecessary. It exists only for configuration compatibility.
var JsxUsesVarsRule = rule.Rule{
	Name: "react/jsx-uses-vars",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{}
	},
}
