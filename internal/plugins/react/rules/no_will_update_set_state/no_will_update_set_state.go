package no_will_update_set_state

import (
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// NoWillUpdateSetStateRule is a thin wrapper around upstream's
// `makeNoMethodSetStateRule('componentWillUpdate', testReactVersion(>= 16.3.0))`
// factory.
//
// `componentWillUpdate` is NOT in upstream's `methodNoopsAsOf` map, so the
// rule stays active regardless of React version. The version check is used
// only to decide whether to also match the `UNSAFE_componentWillUpdate`
// alias introduced in React 16.3.
var NoWillUpdateSetStateRule rule.Rule = reactutil.MakeNoMethodSetStateRule(reactutil.NoMethodSetStateConfig{
	RuleName:          "react/no-will-update-set-state",
	MethodName:        "componentWillUpdate",
	ShouldCheckUnsafe: reactutil.CheckUnsafeAtReactVersion(16, 3, 0),
})
