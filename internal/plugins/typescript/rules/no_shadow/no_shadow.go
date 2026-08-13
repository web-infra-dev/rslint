package no_shadow

import (
	_ "embed"

	"github.com/web-infra-dev/rslint/internal/rule"
	core "github.com/web-infra-dev/rslint/internal/rules/no_shadow"
)

//go:embed no_shadow.schema.json
var schemaJSON []byte

// NoShadowRule is the typescript-eslint wrapper around the core no-shadow
// implementation. The rule shares the scope/shadow-detection pipeline with
// the ESLint core rule (`internal/rules/no_shadow`) while selecting
// typescript-eslint's defaults and its default TypeScript type globals.
var NoShadowRule = rule.CreateRule(rule.Rule{
	Name:   "no-shadow",
	Schema: rule.NewSchema(schemaJSON),
	Run:    core.RunTSESLint,
})
