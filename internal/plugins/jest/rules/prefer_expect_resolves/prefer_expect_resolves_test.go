package prefer_expect_resolves_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_expect_resolves"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferExpectResolvesRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_expect_resolves.PreferExpectResolvesRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect.hasAssertions()`},
			{Code: `
      it('passes', async () => {
        await expect(someValue()).resolves.toBe(true);
      });
    `},
			{Code: `
      it('is true', async () => {
        const myPromise = Promise.resolve(true);

        await expect(myPromise).resolves.toBe(true);
      });
    `},
			{Code: `
      it('errors', async () => {
        await expect(Promise.reject(new Error('oh noes!'))).rejects.toThrowError(
          'oh noes!',
        );
      });
    `},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `
        it('passes', async () => {
          expect(await someValue(),).toBe(true);
        });
      `,
				Output: []string{`
        it('passes', async () => {
          await expect(someValue(),).resolves.toBe(true);
        });
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectResolves", Line: 3, Column: 18, EndColumn: 35},
				},
			},
			{
				Code: `
        it('is true', async () => {
          const myPromise = Promise.resolve(true);

          expect(await myPromise).toBe(true);
        });
      `,
				Output: []string{`
        it('is true', async () => {
          const myPromise = Promise.resolve(true);

          await expect(myPromise).resolves.toBe(true);
        });
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectResolves", Line: 5, Column: 18, EndColumn: 33},
				},
			},
			{
				Code: `
        import { expect as pleaseExpect } from '@jest/globals';

        it('is true', async () => {
          const myPromise = Promise.resolve(true);

          pleaseExpect(await myPromise).toBe(true);
        });
      `,
				Output: []string{`
        import { expect as pleaseExpect } from '@jest/globals';

        it('is true', async () => {
          const myPromise = Promise.resolve(true);

          await pleaseExpect(myPromise).resolves.toBe(true);
        });
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectResolves", Line: 7, Column: 24, EndColumn: 39},
				},
			},
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
			{
				Code: `
        it('prefers moving await before expect rejects', async () => {
          const myPromise = Promise.reject(new Error('oh noes!'));

          expect(await myPromise).rejects.toThrow('oh noes!');
        });
      `,
				Output: []string{`
        it('prefers moving await before expect rejects', async () => {
          const myPromise = Promise.reject(new Error('oh noes!'));

          await expect(myPromise).rejects.toThrow('oh noes!');
        });
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectResolves", Line: 5, Column: 18, EndColumn: 33},
				},
			},
			{
				Code: `
        it('does not duplicate resolves modifier', async () => {
          const myPromise = Promise.resolve(true);

          expect(await myPromise).resolves.toBe(true);
        });
      `,
				Output: []string{`
        it('does not duplicate resolves modifier', async () => {
          const myPromise = Promise.resolve(true);

          await expect(myPromise).resolves.toBe(true);
        });
      `},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "expectResolves", Line: 5, Column: 18, EndColumn: 33},
				},
			},
		},
	)
}
