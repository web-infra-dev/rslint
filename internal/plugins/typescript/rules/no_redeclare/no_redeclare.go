package no_redeclare

import (
	_ "embed"

	"github.com/web-infra-dev/rslint/internal/rule"
	coreNoRedeclare "github.com/web-infra-dev/rslint/internal/rules/no_redeclare"
)

//go:embed no_redeclare.schema.json
var schemaJSON []byte

var NoRedeclareRule = rule.CreateRule(rule.Rule{
	Name:   "no-redeclare",
	Schema: rule.NewSchema(schemaJSON),
	Run:    coreNoRedeclare.RunTSESLint,
})
