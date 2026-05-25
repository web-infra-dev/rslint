package no_shadow

import (
	"github.com/web-infra-dev/rslint/internal/rule"
	core "github.com/web-infra-dev/rslint/internal/rules/no_shadow"
)

// NoShadowRule is the typescript-eslint wrapper around the core no-shadow
// implementation. The rule shares the entire scope/shadow-detection pipeline
// with the ESLint core rule (`internal/rules/no_shadow`); the only behavioral
// difference is the default `hoist` option, which is `functions-and-types`
// here versus `functions` in core ESLint.
var NoShadowRule = rule.CreateRule(rule.Rule{
	Name: "no-shadow",
	Run:  core.RunTSESLint,
})
