package no_did_update_set_state

import (
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoDidUpdateSetStateRule is a thin wrapper around upstream's
// `makeNoMethodSetStateRule('componentDidUpdate')` factory.
//
// `componentDidUpdate` is in upstream's `methodNoopsAsOf` map (>= 16.3.0),
// so the rule becomes a no-op when the user explicitly pins React in
// [16.3.0, 999.999.999). No `UNSAFE_` alias is checked — upstream calls the
// factory without a `shouldCheckUnsafeCb`.
var NoDidUpdateSetStateRule rule.Rule = reactutil.MakeNoMethodSetStateRule(reactutil.NoMethodSetStateConfig{
	RuleName:     "react/no-did-update-set-state",
	MethodName:   "componentDidUpdate",
	ShouldBeNoop: reactutil.MethodNoopAtReactVersion(16, 3, 0),
})
