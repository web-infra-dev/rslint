package no_unnecessary_type_conversion

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/rules/fixtures"
)

func TestNoUnnecessaryTypeConversion(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessaryTypeConversionRule, []rule_tester.ValidTestCase{
		{Code: `
			const value = "test";
			console.log(value);
		`},
		{Code: `
			const num = 42;
			console.log(num);
		`},
		{Code: `
			const bool = true;
			console.log(bool);
		`},
		{Code: `
			// Different types - should be valid
			const str = "test";
			const result = Number(str);
		`},
	}, []rule_tester.InvalidTestCase{
		// For now, leave empty until the rule logic is fully implemented
	})
}