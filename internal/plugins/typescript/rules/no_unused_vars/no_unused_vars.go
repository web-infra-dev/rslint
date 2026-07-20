package no_unused_vars

import (
	"github.com/web-infra-dev/rslint/internal/rule"
	corerule "github.com/web-infra-dev/rslint/internal/rules/no_unused_vars"
)

// NoUnusedVarsRule extends the shared no-unused-vars implementation with
// TypeScript declarations and @typescript-eslint's type-only usage semantics.
var NoUnusedVarsRule = rule.CreateRule(corerule.NewTypeScriptRule())
