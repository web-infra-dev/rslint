package no_magic_numbers

import (
	_ "embed"

	"github.com/web-infra-dev/rslint/internal/rule"
	core "github.com/web-infra-dev/rslint/internal/rules/no_magic_numbers"
)

//go:embed no_magic_numbers.schema.json
var schemaJSON []byte

// NoMagicNumbersRule is the typescript-eslint wrapper around the core
// no-magic-numbers implementation (internal/rules/no_magic_numbers). The core
// rule's schema already covers the extension's options (ignoreEnums,
// ignoreNumericLiteralTypes, ignoreReadonlyClassProperties, ignoreTypeIndexes),
// so the two share one implementation, selected here through core.RunTSESLint
// because the extension settles those TypeScript-specific positions on its own
// before the core checks ever run.
var NoMagicNumbersRule = rule.CreateRule(rule.Rule{
	Name:   "no-magic-numbers",
	Schema: rule.NewSchema(schemaJSON),
	Run:    core.RunTSESLint,
})
