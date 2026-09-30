package unbound_method

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const serviceDeclaration = "class Service { method() {} }\nconst service = new Service();\n"

// TestUnboundMethodExtras locks in the exemption's edge shapes. Every expected
// result was taken from eslint-plugin-jest v29.16.6 running on the same code.
func TestUnboundMethodExtras(t *testing.T) {
	rule_tester.RunRuleTester(upstreamRoot(), "tsconfig.json", t, &UnboundMethodRule,
		[]rule_tester.ValidTestCase{
			// A matcher that does not invoke the subject exempts every argument in
			// the expect chain, through parentheses, nested calls, and modifiers.
			{Code: serviceDeclaration + "expect((service.method)).toBe(1);"},
			{Code: serviceDeclaration + "expect(service.method).resolves.toBe(1);"},
			{Code: serviceDeclaration + "expect(service.method).rejects.toBe(1);"},
			{Code: serviceDeclaration + "expect(value).toBe(service.method);"},
			{Code: serviceDeclaration + "expect(wrap(service.method)).toBe(1);"},
			{Code: serviceDeclaration + "expect.soft(service.method).toBe(1);"},
			{Code: serviceDeclaration + "expect(service['method']).toBe(1);"},
			{Code: serviceDeclaration + "expect(service.method).not.toBe(1);"},
			{Code: serviceDeclaration + "expect(service.method)[\"toBe\"](1);"},
			{Code: serviceDeclaration + "expect(service.method, 'message').toBe(1);"},
			{Code: serviceDeclaration + "expect(value).toHaveBeenCalledWith(expect.any(service.method));"},
			{Code: serviceDeclaration + "expect(value)?.toBe(service.method);"},
			{Code: serviceDeclaration + "wrap(expect(value)?.toBe(service.method));"},
			{Code: serviceDeclaration + "it('works', () => { expect(service.method).toHaveBeenCalled(); });"},
			{Code: serviceDeclaration + "import { expect } from '@jest/globals';\nexpect(service.method).toBe(1);"},
			// `jest.mocked` is matched by name: string and template keys, a computed
			// identifier, optional calls, and even a local `jest` binding.
			{Code: serviceDeclaration + "jest.mocked(service.method, true);"},
			{Code: serviceDeclaration + "jest.mocked(value, service.method);"},
			{Code: serviceDeclaration + "jest['mocked'](service.method);"},
			{Code: serviceDeclaration + "jest[`mocked`](service.method);"},
			{Code: serviceDeclaration + "'jest'.mocked(service.method);"},
			{Code: serviceDeclaration + "declare const mocked: string;\njest[mocked](service.method);"},
			{Code: serviceDeclaration + "(jest).mocked(service.method);"},
			{Code: serviceDeclaration + "(jest.mocked)(service.method);"},
			{Code: serviceDeclaration + "jest?.mocked(service.method);"},
			{Code: serviceDeclaration + "jest.mocked?.(service.method);"},
			{Code: serviceDeclaration + "const jest = { mocked(fn: unknown) {} };\njest.mocked(service.method);"},
			// `jest.mocked` resolved through the Jest call parser.
			{Code: serviceDeclaration + "import { jest as j } from '@jest/globals';\nj.mocked(service.method);"},
			{Code: serviceDeclaration + "jest.mocked(value).mockReturnValue(service.method);"},
			// Upstream reports this: its call parser stops at the ChainExpression
			// the parentheses end. The shared Jest call parser reads through them.
			{Code: serviceDeclaration + "(jest?.mocked)(service.method);"},
			// Options are forwarded to the base rule.
			{Code: serviceDeclaration + "class Static { static method() {} }\nexpect(Static.method).toThrow();", Options: map[string]any{"ignoreStatic": true}},
		},
		[]rule_tester.InvalidTestCase{
			// Upstream checks the direct ESTree parent, so an optional chain
			// (ChainExpression) or a TypeScript wrapper hides the call.
			{Code: serviceDeclaration + "expect(service?.method).toBe(1);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 8, EndLine: 3, EndColumn: 23}}},
			{Code: serviceDeclaration + "expect((service?.method)).toBe(1);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 9, EndLine: 3, EndColumn: 24}}},
			{Code: serviceDeclaration + "expect(service.method as any).toBe(1);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 8, EndLine: 3, EndColumn: 22}}},
			{Code: serviceDeclaration + "expect(service.method!).toBe(1);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 8, EndLine: 3, EndColumn: 22}}},
			{Code: serviceDeclaration + "expect(<any>service.method).toBe(1);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 13, EndLine: 3, EndColumn: 27}}},
			{Code: serviceDeclaration + "expect(service.method satisfies unknown).toBe(1);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 8, EndLine: 3, EndColumn: 22}}},
			// Upstream follows the outermost call chain, which may leave the expect
			// call; a dynamic import is not a CallExpression.
			{Code: serviceDeclaration + "wrap(expect(service.method).toBe(1));", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 13, EndLine: 3, EndColumn: 27}}},
			{Code: serviceDeclaration + "expect(import(wrap(service.method))).toBe(1);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 20, EndLine: 3, EndColumn: 34}}},
			{Code: serviceDeclaration + "import(service.method);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 8, EndLine: 3, EndColumn: 22}}},
			// Accessors that match neither the name check nor the parsed `mocked` member.
			{Code: serviceDeclaration + "import { jest as j } from '@jest/globals';\nj['mocked'](service.method);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 4, Column: 13, EndLine: 4, EndColumn: 27}}},
			{Code: serviceDeclaration + "import { jest as j } from '@jest/globals';\nj[`mocked`](service.method);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 4, Column: 13, EndLine: 4, EndColumn: 27}}},
			{Code: serviceDeclaration + "jest[`mo\\x63ked`](service.method);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 19, EndLine: 3, EndColumn: 33}}},
			{Code: serviceDeclaration + "jest.mocked!(service.method);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 14, EndLine: 3, EndColumn: 28}}},
			// Throwing matchers, broken expect chains, and other calls are left to the base rule.
			{Code: serviceDeclaration + "expect(value).toThrow(service.method);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 23, EndLine: 3, EndColumn: 37}}},
			{Code: serviceDeclaration + "expect(service.method).resolves.toThrow();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 8, EndLine: 3, EndColumn: 22}}},
			{Code: serviceDeclaration + "expect(service.method).toHaveBeenCalledTimes;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 1, EndLine: 3, EndColumn: 45}, {MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 8, EndLine: 3, EndColumn: 22}}},
			{Code: serviceDeclaration + "expect(service.method);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 8, EndLine: 3, EndColumn: 22}}},
			{Code: serviceDeclaration + "it('works', service.method);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 13, EndLine: 3, EndColumn: 27}}},
			{Code: serviceDeclaration + "describe('suite', service.method);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 19, EndLine: 3, EndColumn: 33}}},
			{Code: serviceDeclaration + "jest.spyOn(service.method);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 12, EndLine: 3, EndColumn: 26}}},
			{Code: serviceDeclaration + "new Wrapper(service.method);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 13, EndLine: 3, EndColumn: 27}}},
			{Code: serviceDeclaration + "expect.extend({ method: service.method });", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 25, EndLine: 3, EndColumn: 39}}},
			{Code: serviceDeclaration + "wrap?.(service.method);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 8, EndLine: 3, EndColumn: 22}}},
			{Code: serviceDeclaration + "const expect = (value: unknown) => ({ toBe(expected: unknown) {} });\nexpect(service.method).toBe(1);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 4, Column: 8, EndLine: 4, EndColumn: 22}}},
			// Destructuring is never exempted.
			{Code: serviceDeclaration + "const { method } = service;\nexpect(method).toBe(1);", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 9, EndLine: 3, EndColumn: 15}}},
			{Code: serviceDeclaration + "class Static { static method() {} }\nexpect(Static.method).toThrow();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 4, Column: 8, EndLine: 4, EndColumn: 21}}},
			// Throwing matchers through optional and computed accessors.
			{Code: serviceDeclaration + "expect(service.method).toThrow?.();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 8, EndLine: 3, EndColumn: 22}}},
			{Code: serviceDeclaration + "expect(service.method)?.toThrow();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 8, EndLine: 3, EndColumn: 22}}},
			{Code: serviceDeclaration + "expect(service.method)[`toThrow`]();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "unboundWithoutThisAnnotation", Line: 3, Column: 8, EndLine: 3, EndColumn: 22}}},
		})
}
