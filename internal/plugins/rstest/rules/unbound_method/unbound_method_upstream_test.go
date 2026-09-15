package unbound_method

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const serviceDeclaration = "export {};\nclass Service { method() {} }\nconst service = new Service();\n"

func unboundError(line, column, width int) []rule_tester.InvalidTestCaseError {
	return []rule_tester.InvalidTestCaseError{{
		MessageId: "unboundWithoutThisAnnotation",
		Message:   "Avoid referencing unbound methods which may cause unintentional scoping of `this`.\nIf your function does not access `this`, you can annotate it with `this: void`, or consider using an arrow function instead.",
		Line:      line, Column: column, EndLine: line, EndColumn: column + width,
	}}
}

func TestUnboundMethodUpstream(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	for _, code := range []string{
		"expect(service.method).toHaveBeenCalledTimes(1);",
		"expect(service.method).not.toHaveBeenCalled();",
		"expect(service.method).toStrictEqual(somethingElse);",
		"rs.mocked(service.method).mockImplementation(() => {});",
		"rstest.mocked(service.method).mockImplementation(() => {});",
		"expect(() => expect(service.method).toHaveBeenCalled()).not.toThrow();",
		"const parameter = rs.mocked(service.method).mock.calls[0][0];",
		"const calls = rs.mocked(service.method).mock.calls;",
		"const lastCall = rs.mocked(service.method).mock.calls[0];",
		"const mockedMethod = rs.mocked(service.method); const parameter = mockedMethod.mock.calls[0][0];",
		"rs.mocked(service.method).mock;",
		"const mockProp = rs.mocked(service.method).mock;",
		"const result = rs.mocked(service.method, true);",
		"rs.mocked(service.method, { shallow: true });",
	} {
		valid = append(valid, rule_tester.ValidTestCase{Code: serviceDeclaration + code})
	}
	valid = append(valid, rule_tester.ValidTestCase{Code: "expect(() => Promise.resolve().then(console.log)).not.toThrow();"})
	invalid := []rule_tester.InvalidTestCase{
		{Code: serviceDeclaration + "expect(service.method);", Errors: unboundError(4, 8, 14)},
		{Code: serviceDeclaration + "expect(service.method).toHaveBeenCalledTimes;", Errors: unboundError(4, 8, 14)},
		{Code: serviceDeclaration + "const method = service.method; rs.mocked(method);", Errors: unboundError(4, 16, 14)},
		{Code: serviceDeclaration + "const method = service.method;", Errors: unboundError(4, 16, 14)},
		{Code: serviceDeclaration + "Promise.resolve().then(service.method);", Errors: unboundError(4, 24, 14)},
		{Code: serviceDeclaration + "expect(() => { Promise.resolve().then(service.method); }).not.toThrow();", Errors: unboundError(4, 39, 14)},
	}
	for _, matcher := range []string{"toThrow", "toThrowError", "toThrowErrorMatchingSnapshot", "toThrowErrorMatchingInlineSnapshot"} {
		for _, modifier := range []string{"", "not."} {
			valid = append(valid, rule_tester.ValidTestCase{Code: "expect(console.log)." + modifier + matcher + "();"})
			invalid = append(invalid, rule_tester.InvalidTestCase{
				Code:   serviceDeclaration + "expect(service.method)." + modifier + matcher + "();",
				Errors: unboundError(4, 8, 14),
			})
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &UnboundMethodRule, valid, invalid)
}
