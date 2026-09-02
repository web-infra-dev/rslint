package prefer_to_be_truthy

import (
	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var PreferToBeTruthyRule rule.Rule = rstestUtils.MakePreferBooleanMatcherRule(rstestUtils.PreferBooleanMatcherConfig{
	RuleName:           "rstest/prefer-to-be-truthy",
	MessageId:          "preferToBeTruthy",
	Description:        "Prefer using `toBeTruthy` to test value is `true`",
	ExpectedLiteral:    ast.KindTrueKeyword,
	ReplacementMatcher: "toBeTruthy",
})
