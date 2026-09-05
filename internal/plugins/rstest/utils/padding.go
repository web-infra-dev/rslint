package utils

import (
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/test_framework/padding"
)

type PaddingRuleKind uint8

const (
	PaddingAfterAll PaddingRuleKind = iota
	PaddingAfterEach
	PaddingBeforeAll
	PaddingBeforeEach
	PaddingDescribe
	PaddingExpect
	PaddingTest
	PaddingAll
)

var rstestPaddingNames = padding.StatementNames{
	"afterAll":   padding.StatementAfterAll,
	"afterEach":  padding.StatementAfterEach,
	"beforeAll":  padding.StatementBeforeAll,
	"beforeEach": padding.StatementBeforeEach,
	"describe":   padding.StatementDescribe,
	"expect":     padding.StatementExpect,
	"it":         padding.StatementTest,
	"test":       padding.StatementTest,
}

var rstestPaddingMessage = rule.RuleMessage{
	Id:          "missingPadding",
	Description: "Expected blank line before this statement.",
}

func MakePaddingRule(name string, kind PaddingRuleKind) rule.Rule {
	configs := paddingConfigs(kind)
	priority := 0
	if kind == PaddingAll {
		priority = 100
	}
	return padding.NewRule(padding.Definition{
		Name: name, Family: "rstest", Priority: priority,
		Message: rstestPaddingMessage, Names: rstestPaddingNames, Configs: configs,
	})
}

func paddingConfigs(kind PaddingRuleKind) []padding.Config {
	switch kind {
	case PaddingAfterAll:
		return around(padding.StatementAfterAll)
	case PaddingAfterEach:
		return around(padding.StatementAfterEach)
	case PaddingBeforeAll:
		return around(padding.StatementBeforeAll)
	case PaddingBeforeEach:
		return around(padding.StatementBeforeEach)
	case PaddingDescribe:
		return around(padding.StatementDescribe)
	case PaddingExpect:
		return append(around(padding.StatementExpect), padding.Config{
			Padding: padding.PaddingAny, Previous: padding.Types(padding.StatementExpect), Next: padding.Types(padding.StatementExpect),
		})
	case PaddingTest:
		return around(padding.StatementTest)
	case PaddingAll:
		var configs []padding.Config
		for _, item := range []PaddingRuleKind{
			PaddingAfterAll, PaddingAfterEach, PaddingBeforeAll, PaddingBeforeEach,
			PaddingDescribe, PaddingExpect, PaddingTest,
		} {
			configs = append(configs, paddingConfigs(item)...)
		}
		return configs
	default:
		return nil
	}
}

func around(statementType padding.StatementType) []padding.Config {
	return []padding.Config{
		{Padding: padding.PaddingAlways, Previous: padding.Types(padding.StatementAny), Next: padding.Types(statementType)},
		{Padding: padding.PaddingAlways, Previous: padding.Types(statementType), Next: padding.Types(padding.StatementAny)},
	}
}
