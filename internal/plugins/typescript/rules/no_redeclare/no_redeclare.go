package no_redeclare

import (
	"github.com/web-infra-dev/rslint/internal/rule"
	coreNoRedeclare "github.com/web-infra-dev/rslint/internal/rules/no_redeclare"
)

var NoRedeclareRule = rule.CreateRule(rule.Rule{
	Name: "no-redeclare",
	Run:  coreNoRedeclare.RunTSESLint,
})
