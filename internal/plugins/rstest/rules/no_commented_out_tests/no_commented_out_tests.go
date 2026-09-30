package no_commented_out_tests

import (
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_commented_out_tests"
)

var NoCommentedOutTestsRule = shared.NewRule(shared.Config{
	Name: "rstest/no-commented-out-tests",
	Profile: shared.Profile{
		Roots: map[string]shared.Root{
			"test":     {Kind: shared.RootTest, Extendable: true},
			"it":       {Kind: shared.RootTest, Extendable: true},
			"describe": {Kind: shared.RootDescribe},
		},
		Members: map[string]shared.Member{
			"only":       {Kind: shared.MemberModifier},
			"skip":       {Kind: shared.MemberModifier},
			"todo":       {Kind: shared.MemberModifier},
			"concurrent": {Kind: shared.MemberModifier},
			"sequential": {Kind: shared.MemberModifier},
			"fails":      {Kind: shared.MemberModifier, TestOnly: true},
			"runIf":      {Kind: shared.MemberConditionalFactory},
			"skipIf":     {Kind: shared.MemberConditionalFactory},
			"each":       {Kind: shared.MemberParameterizedFactory},
			"for":        {Kind: shared.MemberParameterizedFactory},
			"extend":     {Kind: shared.MemberExtendFactory, TestOnly: true},
		},
	},
})
