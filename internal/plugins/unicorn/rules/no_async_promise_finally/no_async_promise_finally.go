package no_async_promise_finally

import "github.com/web-infra-dev/rslint/internal/rule"

// NoAsyncPromiseFinallyRule is intentionally inert while the test-first port is red.
var NoAsyncPromiseFinallyRule = rule.Rule{
	Name:   "unicorn/no-async-promise-finally",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{}
	},
}
