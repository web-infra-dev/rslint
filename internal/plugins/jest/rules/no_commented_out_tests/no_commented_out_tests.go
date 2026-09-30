package no_commented_out_tests

import (
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/no_commented_out_tests"
)

var jestTest = shared.Root{Kind: shared.RootTest}
var jestDescribe = shared.Root{Kind: shared.RootDescribe}

var NoCommentedOutTestsRule = shared.NewRule(shared.Config{
	Name: "jest/no-commented-out-tests",
	Profile: shared.Profile{
		// `x`/`f` prefixed aliases are the legacy Jest globals
		// (`xit`, `xtest`, `fit`, `xdescribe`, `fdescribe`).
		Roots: map[string]shared.Root{
			"test":      jestTest,
			"it":        jestTest,
			"xtest":     jestTest,
			"xit":       jestTest,
			"fit":       jestTest,
			"describe":  jestDescribe,
			"xdescribe": jestDescribe,
			"fdescribe": jestDescribe,
		},
		Members: map[string]shared.Member{
			"only":       {Kind: shared.MemberModifier},
			"skip":       {Kind: shared.MemberModifier},
			"todo":       {Kind: shared.MemberModifier},
			"concurrent": {Kind: shared.MemberModifier},
			"failing":    {Kind: shared.MemberModifier, TestOnly: true},
			"each":       {Kind: shared.MemberParameterizedFactory},
		},
		// eslint-plugin-jest flags `test.<anything>(` so a modifier added in
		// a future Jest release is still caught.
		AcceptUnknownRootMember: true,
	},
})
