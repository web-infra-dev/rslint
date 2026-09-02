package prefer_to_be_falsy

import (
	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var PreferToBeFalsyRule rule.Rule = rstestUtils.MakePreferBooleanMatcherRule(rstestUtils.PreferBooleanMatcherConfig{
	RuleName:           "rstest/prefer-to-be-falsy",
	MessageId:          "preferToBeFalsy",
	Description:        "Prefer using toBeFalsy()",
	ExpectedLiteral:    ast.KindFalseKeyword,
	ReplacementMatcher: "toBeFalsy",
})
