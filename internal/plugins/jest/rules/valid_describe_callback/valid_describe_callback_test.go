package valid_describe_callback_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/valid_describe_callback"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestValidDescribeCallbackRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&valid_describe_callback.ValidDescribeCallbackRule,
		[]rule_tester.ValidTestCase{
			{Code: "describe.each([1, 2, 3])('%s', (a, b) => {});"},
			{Code: "describe('foo', function() {})"},
			{Code: "describe('foo', () => {})"},
			{Code: "describe(`foo`, () => {})"},
			{Code: "xdescribe('foo', () => {})"},
			{Code: "fdescribe('foo', () => {})"},
			{Code: "describe.only('foo', () => {})"},
			{Code: "describe.skip('foo', () => {})"},
			{Code: `describe('foo', () => {
			    it('bar', () => {
			        return Promise.resolve(42).then(value => {
			            expect(value).toBe(42)
			        })
			    })
			})
			`},
			{Code: `describe('foo', () => {
			    it('bar', async () => {
			        expect(await Promise.resolve(42)).toBe(42)
			    })
			})
			`},
			{Code: "if (hasOwnProperty(obj, key)) {}"},
			{Code: "describe.each`\n                    foo  | foe\n                    ${'1'} | ${'2'}\n                `('$something', ({ foo, foe }) => {});"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "describe.each()()",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nameAndCallback", Line: 1, Column: 1},
				},
			},
			{
				Code: "describe['each']()()",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nameAndCallback", Line: 1, Column: 1},
				},
			},
			{
				Code: "describe.each(() => {})()",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nameAndCallback", Line: 1, Column: 1},
				},
			},
			{
				Code: "describe.each(() => {})('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nameAndCallback", Line: 1, Column: 25},
				},
			},
			{
				Code: "describe.each()(() => {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nameAndCallback", Line: 1, Column: 17},
				},
			},
			{
				Code: "describe['each']()(() => {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nameAndCallback", Line: 1, Column: 20},
				},
			},
			{
				Code: "describe.each('foo')(() => {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nameAndCallback", Line: 1, Column: 22},
				},
			},
			{
				Code: "describe.only.each('foo')(() => {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nameAndCallback", Line: 1, Column: 27},
				},
			},
			{
				Code: "describe(() => {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nameAndCallback", Line: 1, Column: 10},
				},
			},
			{
				Code: "describe('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nameAndCallback", Line: 1, Column: 10},
				},
			},
			{
				Code: "describe('foo', 'foo2')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "secondArgumentMustBeFunction", Line: 1, Column: 17},
				},
			},
			{
				Code: "describe()",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nameAndCallback", Line: 1, Column: 1},
				},
			},
			{
				Code: "describe('foo', async () => {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAsyncDescribeCallback", Line: 1, Column: 17},
				},
			},
			{
				Code: "describe('foo', async function () {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAsyncDescribeCallback", Line: 1, Column: 17},
				},
			},
			{
				Code: "xdescribe('foo', async function () {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAsyncDescribeCallback", Line: 1, Column: 18},
				},
			},
			{
				Code: "fdescribe('foo', async function () {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAsyncDescribeCallback", Line: 1, Column: 18},
				},
			},
			{
				Code: `
					import { fdescribe } from '@jest/globals';
					fdescribe('foo', async function () {})
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAsyncDescribeCallback", Line: 3, Column: 23},
				},
			},
			{
				Code: "describe.only('foo', async function () {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAsyncDescribeCallback", Line: 1, Column: 22},
				},
			},
			{
				Code: "describe.skip('foo', async function () {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAsyncDescribeCallback", Line: 1, Column: 22},
				},
			},
			{
				Code: `describe('sample case', () => {
					it('works', () => {
						expect(true).toEqual(true);
					});
					describe('async', async () => {
						await new Promise(setImmediate);
						it('breaks', () => {
							throw new Error('Fail');
						});
					});
				});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAsyncDescribeCallback", Line: 5, Column: 24},
				},
			},
			{
				Code: `describe('foo', function () {
			    return Promise.resolve().then(() => {
			        it('breaks', () => {
			            throw new Error('Fail')
			        })
			    })
			})`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unexpectedReturnInDescribe"}},
			},
			{
				Code: `describe('foo', () => {
			    return Promise.resolve().then(() => {
			        it('breaks', () => {
			            throw new Error('Fail')
			        })
			    })
			    describe('nested', () => {
			        return Promise.resolve().then(() => {
			            it('breaks', () => {
			                throw new Error('Fail')
			            })
			        })
			    })
			})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedReturnInDescribe"},
					{MessageId: "unexpectedReturnInDescribe"},
				},
			},
			{
				Code: `describe('foo', async () => {
			    await something()
			    it('does something')
			    describe('nested', () => {
			        return Promise.resolve().then(() => {
			            it('breaks', () => {
			                throw new Error('Fail')
			            })
			        })
			    })
			})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAsyncDescribeCallback"},
					{MessageId: "unexpectedReturnInDescribe"},
				},
			},
			{
				Code: `describe('foo', () => test('bar', () => {}))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedReturnInDescribe", Line: 1, Column: 23},
				},
			},
			{
				Code: "describe('foo', done => {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedDescribeArgument", Line: 1, Column: 17},
				},
			},
			{
				Code: "describe('foo', function (done) {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedDescribeArgument", Line: 1, Column: 27},
				},
			},
			{
				Code: "describe('foo', function (one, two, three) {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedDescribeArgument", Line: 1, Column: 27},
				},
			},
			{
				Code: "describe('foo', async function (done) {})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noAsyncDescribeCallback", Line: 1, Column: 17},
					{MessageId: "unexpectedDescribeArgument", Line: 1, Column: 33},
				},
			},
		},
	)
}
