package no_magic_numbers

import (
	_ "embed"

	"github.com/web-infra-dev/rslint/internal/rule"
	core "github.com/web-infra-dev/rslint/internal/rules/no_magic_numbers"
)

//go:embed no_magic_numbers.schema.json
var schemaJSON []byte

// NoMagicNumbersRule is the typescript-eslint wrapper around the core
// no-magic-numbers implementation. As of the ESLint version the core rule
// (internal/rules/no_magic_numbers) was ported from, its schema and behavior
// already fully subsume the typescript-eslint extension's options
// (ignoreEnums, ignoreNumericLiteralTypes, ignoreReadonlyClassProperties,
// ignoreTypeIndexes), so the two rules share one implementation.
var NoMagicNumbersRule = rule.CreateRule(rule.Rule{
	Name:   "no-magic-numbers",
	Schema: rule.NewSchema(schemaJSON),
	Run:    core.RunTSESLint,
})
