package prefer_expect_resolves_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_expect_resolves"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferExpectResolvesExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_expect_resolves.PreferExpectResolvesRule, nil, []rule_tester.InvalidTestCase{
		{
			Code: `
        it('keeps parens around awaited argument', async () => {
          const myPromise = Promise.resolve(true);

          expect(await (myPromise)).toBe(true);
        });
      `,
			Output: []string{`
        it('keeps parens around awaited argument', async () => {
          const myPromise = Promise.resolve(true);

          await expect((myPromise)).resolves.toBe(true);
        });
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectResolves", Line: 5, Column: 18, EndColumn: 35},
			},
		},
		{
			Code: `
        it('unwraps extra parens around await expression', async () => {
          const myPromise = Promise.resolve(true);

          expect((await myPromise)).toBe(true);
        });
      `,
			Output: []string{`
        it('unwraps extra parens around await expression', async () => {
          const myPromise = Promise.resolve(true);

          await expect((myPromise)).resolves.toBe(true);
        });
      `},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectResolves", Line: 5, Column: 19, EndColumn: 34},
			},
		},
	})
}
