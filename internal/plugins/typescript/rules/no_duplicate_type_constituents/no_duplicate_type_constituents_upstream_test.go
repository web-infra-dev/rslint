package no_duplicate_type_constituents

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

// All 34 valid and 48 invalid cases from typescript-eslint v8.72.2.
func TestNoDuplicateTypeConstituentsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDuplicateTypeConstituentsRule, []rule_tester.ValidTestCase{
		{Code: "type T = 1 | 2;"},
		{Code: "type T = 1 | '1';"},
		{Code: "type T = true & boolean;"},
		{Code: "type T = null | undefined;"},
		{Code: "type T = any | unknown;"},
		{Code: "type T = { a: string } | { b: string };"},
		{Code: "type T = { a: string; b: number } | { b: number; a: string };"},
		{Code: "type T = { a: string | number };"},
		{Code: "type T = Set<string> | Set<number>;"},
		{Code: "type T = Class<string> | Class<number>;"},
		{Code: "type T = string[] | number[];"},
		{Code: "type T = string[][] | string[];"},
		{Code: "type T = [1, 2, 3] | [1, 2, 4];"},
		{Code: "type T = [1, 2, 3] | [1, 2, 3, 4];"},
		{Code: "type T = 'A' | string[];"},
		{Code: "type T = (() => string) | (() => void);"},
		{Code: "type T = () => string | void;"},
		{Code: "type T = () => null | undefined;"},
		{Code: "type T = (arg: string | number) => void;"},
		{Code: "type T = A | A;"},
		{Code: "\ntype A = 'A';\ntype B = 'B';\ntype T = A | B;\n      "},
		{Code: "\ntype A = 'A';\ntype B = 'B';\nconst a: A | B = 'A';\n      "},
		{Code: "\ntype A = 'A';\ntype B = 'B';\ntype T = A | /* comment */ B;\n      "},
		{Code: "\ntype A = 'A';\ntype B = 'B';\ntype T = 'A' | 'B';\n      "},
		{Code: "\ntype A = 'A';\ntype B = 'B';\ntype C = 'C';\ntype T = A | B | C;\n      "},
		{Code: "type T = readonly string[] | string[];"},
		{Code: "\ntype A = 'A';\ntype B = 'B';\ntype C = 'C';\ntype D = 'D';\ntype T = (A | B) | (C | D);\n      "},
		{Code: "\ntype A = 'A';\ntype B = 'B';\ntype T = (A | B) | (A & B);\n      "},
		{Code: "\ntype A = 'A';\ntype B = 'B';\ntype T = Record<string, A | B>;\n      "},
		{Code: "type T = A | A;", Options: map[string]any{"ignoreUnions": true}},
		{Code: "type T = A & A;", Options: map[string]any{"ignoreIntersections": true}},
		{Code: "type T = Class<string> | Class<string>;"},
		{Code: "type T = A | A | string;"},
		{Code: "(a: string | undefined) => {};"},
	}, []rule_tester.InvalidTestCase{
		{Code: "type T = 1 | 1;", Output: []string{"type T = 1  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with 1.", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
		}},
		{Code: "type T = true & true;", Output: []string{"type T = true  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Intersection type constituent is duplicated with true.", Line: 1, Column: 17, EndLine: 1, EndColumn: 21},
		}},
		{Code: "type T = null | null;", Output: []string{"type T = null  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with null.", Line: 1, Column: 17, EndLine: 1, EndColumn: 21},
		}},
		{Code: "type T = any | any;", Output: []string{"type T = any  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with any.", Line: 1, Column: 16, EndLine: 1, EndColumn: 19},
		}},
		{Code: "type T = { a: string | string };", Output: []string{"type T = { a: string   };"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with string.", Line: 1, Column: 24, EndLine: 1, EndColumn: 30},
		}},
		{Code: "type T = { a: string } | { a: string };", Output: []string{"type T = { a: string }  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with { a: string }.", Line: 1, Column: 26, EndLine: 1, EndColumn: 39},
		}},
		{Code: "type T = { a: string; b: number } | { a: string; b: number };", Output: []string{"type T = { a: string; b: number }  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with { a: string; b: number }.", Line: 1, Column: 37, EndLine: 1, EndColumn: 61},
		}},
		{Code: "type T = Set<string> | Set<string>;", Output: []string{"type T = Set<string>  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with Set<string>.", Line: 1, Column: 24, EndLine: 1, EndColumn: 35},
		}},
		{Code: "\ntype IsArray<T> = T extends any[] ? true : false;\ntype ActuallyDuplicated = IsArray<number> | IsArray<string>;\n      ", Output: []string{"\ntype IsArray<T> = T extends any[] ? true : false;\ntype ActuallyDuplicated = IsArray<number>  ;\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with IsArray<number>.", Line: 3, Column: 45, EndLine: 3, EndColumn: 60},
		}},
		{Code: "type T = string[] | string[];", Output: []string{"type T = string[]  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with string[].", Line: 1, Column: 21, EndLine: 1, EndColumn: 29},
		}},
		{Code: "type T = string[][] | string[][];", Output: []string{"type T = string[][]  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with string[][].", Line: 1, Column: 23, EndLine: 1, EndColumn: 33},
		}},
		{Code: "type T = [1, 2, 3] | [1, 2, 3];", Output: []string{"type T = [1, 2, 3]  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with [1, 2, 3].", Line: 1, Column: 22, EndLine: 1, EndColumn: 31},
		}},
		{Code: "type T = () => string | string;", Output: []string{"type T = () => string  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with string.", Line: 1, Column: 25, EndLine: 1, EndColumn: 31},
		}},
		{Code: "type T = () => null | null;", Output: []string{"type T = () => null  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with null.", Line: 1, Column: 23, EndLine: 1, EndColumn: 27},
		}},
		{Code: "type T = (arg: string | string) => void;", Output: []string{"type T = (arg: string  ) => void;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with string.", Line: 1, Column: 25, EndLine: 1, EndColumn: 31},
		}},
		{Code: "type T = 'A' | 'A';", Output: []string{"type T = 'A'  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with 'A'.", Line: 1, Column: 16, EndLine: 1, EndColumn: 19},
		}},
		{Code: "\ntype A = 'A';\ntype T = A | A;\n      ", Output: []string{"\ntype A = 'A';\ntype T = A  ;\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with A.", Line: 3, Column: 14, EndLine: 3, EndColumn: 15},
		}},
		{Code: "\ntype A = 'A';\nconst a: A | A = 'A';\n      ", Output: []string{"\ntype A = 'A';\nconst a: A   = 'A';\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with A.", Line: 3, Column: 14, EndLine: 3, EndColumn: 15},
		}},
		{Code: "\ntype A = 'A';\ntype T = A | /* comment */ A;\n      ", Output: []string{"\ntype A = 'A';\ntype T = A  /* comment */ ;\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with A.", Line: 3, Column: 28, EndLine: 3, EndColumn: 29},
		}},
		{Code: "\ntype A1 = 'A';\ntype A2 = 'A';\ntype A3 = 'A';\ntype T = A1 | A2 | A3;\n      ", Output: []string{"\ntype A1 = 'A';\ntype A2 = 'A';\ntype A3 = 'A';\ntype T = A1    ;\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with A1.", Line: 5, Column: 15, EndLine: 5, EndColumn: 17},
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with A1.", Line: 5, Column: 20, EndLine: 5, EndColumn: 22},
		}},
		{Code: "\ntype A = 'A';\ntype B = 'B';\ntype T = A | B | A;\n      ", Output: []string{"\ntype A = 'A';\ntype B = 'B';\ntype T = A | B  ;\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with A.", Line: 4, Column: 18, EndLine: 4, EndColumn: 19},
		}},
		{Code: "\ntype A = 'A';\ntype B = 'B';\ntype T = A | B | A | B;\n      ", Output: []string{"\ntype A = 'A';\ntype B = 'B';\ntype T = A | B    ;\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with A.", Line: 4, Column: 18, EndLine: 4, EndColumn: 19},
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with B.", Line: 4, Column: 22, EndLine: 4, EndColumn: 23},
		}},
		{Code: "\ntype A = 'A';\ntype B = 'B';\ntype T = A | B | A | A;\n      ", Output: []string{"\ntype A = 'A';\ntype B = 'B';\ntype T = A | B    ;\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with A.", Line: 4, Column: 18, EndLine: 4, EndColumn: 19},
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with A.", Line: 4, Column: 22, EndLine: 4, EndColumn: 23},
		}},
		{Code: "\ntype A = 'A';\ntype B = 'B';\ntype C = 'C';\ntype T = A | B | A | C;\n      ", Output: []string{"\ntype A = 'A';\ntype B = 'B';\ntype C = 'C';\ntype T = A | B   | C;\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with A.", Line: 5, Column: 18, EndLine: 5, EndColumn: 19},
		}},
		{Code: "\ntype A = 'A';\ntype B = 'B';\ntype T = (A | B) | (A | B);\n      ", Output: []string{"\ntype A = 'A';\ntype B = 'B';\ntype T = (A | B)  ;\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with A | B.", Line: 4, Column: 21, EndLine: 4, EndColumn: 27},
		}},
		{Code: "\ntype A = 'A';\ntype T = A | (A | A);\n      ", Output: []string{"\ntype A = 'A';\ntype T = A  ;\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with A.", Line: 3, Column: 15, EndLine: 3, EndColumn: 21},
		}},
		{Code: "\ntype A = 'A';\ntype B = 'B';\ntype C = 'C';\ntype D = 'D';\ntype F = (A | B) | (A | B) | ((C | D) & (A | B)) | (A | B);\n      ", Output: []string{"\ntype A = 'A';\ntype B = 'B';\ntype C = 'C';\ntype D = 'D';\ntype F = (A | B)   | ((C | D) & (A | B))  ;\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with A | B.", Line: 6, Column: 21, EndLine: 6, EndColumn: 27},
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with A | B.", Line: 6, Column: 53, EndLine: 6, EndColumn: 59},
		}},
		{Code: "\ntype A = 'A';\ntype B = 'B';\ntype C = (A | B) | A | B | (A | B);\n      ", Output: []string{"\ntype A = 'A';\ntype B = 'B';\ntype C = (A | B)      ;\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with A.", Line: 4, Column: 20, EndLine: 4, EndColumn: 21},
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with B.", Line: 4, Column: 24, EndLine: 4, EndColumn: 25},
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with A | B.", Line: 4, Column: 29, EndLine: 4, EndColumn: 35},
		}},
		{Code: "type A = (number | string) | number | string;", Output: []string{"type A = (number | string)    ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with number.", Line: 1, Column: 30, EndLine: 1, EndColumn: 36},
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with string.", Line: 1, Column: 39, EndLine: 1, EndColumn: 45},
		}},
		{Code: "type A = (number | (string | null)) | (string | (null | number));", Output: []string{"type A = (number | (string | null))  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with number | (string | null).", Line: 1, Column: 40, EndLine: 1, EndColumn: 65},
		}},
		{Code: "type A = (number & string) & number & string;", Output: []string{"type A = (number & string)    ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Intersection type constituent is duplicated with number.", Line: 1, Column: 30, EndLine: 1, EndColumn: 36},
			{MessageId: "duplicate", Message: "Intersection type constituent is duplicated with string.", Line: 1, Column: 39, EndLine: 1, EndColumn: 45},
		}},
		{Code: "type A = number & string & (number & string);", Output: []string{"type A = number & string & (  string);", "type A = number & string    ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Intersection type constituent is duplicated with number.", Line: 1, Column: 29, EndLine: 1, EndColumn: 35},
			{MessageId: "duplicate", Message: "Intersection type constituent is duplicated with string.", Line: 1, Column: 38, EndLine: 1, EndColumn: 44},
		}},
		{Code: "\ntype A = 'A';\ntype T = Record<string, A | A>;\n      ", Output: []string{"\ntype A = 'A';\ntype T = Record<string, A  >;\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with A.", Line: 3, Column: 29, EndLine: 3, EndColumn: 30},
		}},
		{Code: "type T = A | A | string | string;", Output: []string{"type T = A | A | string  ;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "duplicate", Message: "Union type constituent is duplicated with string.", Line: 1, Column: 27, EndLine: 1, EndColumn: 33},
		}},
		{Code: "(a?: string | undefined) => {};", Output: []string{"(a?: string  ) => {};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 1, Column: 15, EndLine: 1, EndColumn: 24},
		}},
		{Code: "\ntype T = undefined;\n(arg?: T | string) => {};\n      ", Output: []string{"\ntype T = undefined;\n(arg?:   string) => {};\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 3, Column: 8, EndLine: 3, EndColumn: 9},
		}},
		{Code: "\ninterface F {\n  (a?: string | undefined): void;\n}\n      ", Output: []string{"\ninterface F {\n  (a?: string  ): void;\n}\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 3, Column: 17, EndLine: 3, EndColumn: 26},
		}},
		{Code: "type fn = new (a?: string | undefined) => void;", Output: []string{"type fn = new (a?: string  ) => void;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 1, Column: 29, EndLine: 1, EndColumn: 38},
		}},
		{Code: "function f(a?: string | undefined) {}", Output: []string{"function f(a?: string  ) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 1, Column: 25, EndLine: 1, EndColumn: 34},
		}},
		{Code: "f = function (a?: string | undefined) {};", Output: []string{"f = function (a?: string  ) {};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 1, Column: 28, EndLine: 1, EndColumn: 37},
		}},
		{Code: "declare function f(a?: string | undefined): void;", Output: []string{"declare function f(a?: string  ): void;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 1, Column: 33, EndLine: 1, EndColumn: 42},
		}},
		{Code: "\ndeclare class bb {\n  f(a?: string | undefined): void;\n}\n      ", Output: []string{"\ndeclare class bb {\n  f(a?: string  ): void;\n}\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 3, Column: 18, EndLine: 3, EndColumn: 27},
		}},
		{Code: "\ninterface ee {\n  f(a?: string | undefined): void;\n}\n      ", Output: []string{"\ninterface ee {\n  f(a?: string  ): void;\n}\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 3, Column: 18, EndLine: 3, EndColumn: 27},
		}},
		{Code: "\ninterface ee {\n  new (a?: string | undefined): void;\n}\n      ", Output: []string{"\ninterface ee {\n  new (a?: string  ): void;\n}\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 3, Column: 21, EndLine: 3, EndColumn: 30},
		}},
		{Code: "type fn = (a?: string | undefined) => void;", Output: []string{"type fn = (a?: string  ) => void;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 1, Column: 25, EndLine: 1, EndColumn: 34},
		}},
		{Code: "type fn = (a?: string | (undefined | number)) => void;", Output: []string{"type fn = (a?: string | (  number)) => void;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 1, Column: 26, EndLine: 1, EndColumn: 35},
		}},
		{Code: "type fn = (a?: (undefined | number) | string) => void;", Output: []string{"type fn = (a?: (  number) | string) => void;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 1, Column: 17, EndLine: 1, EndColumn: 26},
		}},
		{Code: "\nabstract class cc {\n  abstract f(a?: string | undefined): void;\n}\n      ", Output: []string{"\nabstract class cc {\n  abstract f(a?: string  ): void;\n}\n      "}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "unnecessary", Message: "Explicit undefined is unnecessary on an optional parameter.", Line: 3, Column: 27, EndLine: 3, EndColumn: 36},
		}},
	})
}

// Preserve the original upstream inputs, including their older whitespace variants.
func TestNoDuplicateTypeConstituentsLegacy(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoDuplicateTypeConstituentsRule, []rule_tester.ValidTestCase{
		{
			Code: "type T = 1 | 2;",
		},
		{
			Code: "type T = 1 | '1';",
		},
		{
			Code: "type T = true & boolean;",
		},
		{
			Code: "type T = null | undefined;",
		},
		{
			Code: "type T = any | unknown;",
		},
		{
			Code: "type T = { a: string } | { b: string };",
		},
		{
			Code: "type T = { a: string; b: number } | { b: number; a: string };",
		},
		{
			Code: "type T = { a: string | number };",
		},
		{
			Code: "type T = Set<string> | Set<number>;",
		},
		{
			Code: "type T = Class<string> | Class<number>;",
		},
		{
			Code: "type T = string[] | number[];",
		},
		{
			Code: "type T = string[][] | string[];",
		},
		{
			Code: "type T = [1, 2, 3] | [1, 2, 4];",
		},
		{
			Code: "type T = [1, 2, 3] | [1, 2, 3, 4];",
		},
		{
			Code: "type T = 'A' | string[];",
		},
		{
			Code: "type T = (() => string) | (() => void);",
		},
		{
			Code: "type T = () => string | void;",
		},
		{
			Code: "type T = () => null | undefined;",
		},
		{
			Code: "type T = (arg: string | number) => void;",
		},
		{
			Code: "type T = A | A;",
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type T = A | B;
      `,
		},
		{
			Code: `
type A = 'A';
type B = 'B';
const a: A | B = 'A';
      `,
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type T = A | /* comment */ B;
      `,
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type T = 'A' | 'B';
      `,
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type C = 'C';
type T = A | B | C;
      `,
		},
		{
			Code: "type T = readonly string[] | string[];",
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type C = 'C';
type D = 'D';
type T = (A | B) | (C | D);
      `,
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type T = (A | B) | (A & B);
      `,
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type T = Record<string, A | B>;
      `,
		},
		{
			Code:    "type T = A | A;",
			Options: map[string]interface{}{"ignoreUnions": true},
		},
		{
			Code:    "type T = A & A;",
			Options: map[string]interface{}{"ignoreIntersections": true},
		},
		{
			Code: "type T = Class<string> | Class<string>;",
		},
		{
			Code: "type T = A | A | string;",
		},
		{
			Code: "(a: string | undefined) => {};",
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   "type T = 1 | 1;",
			Output: []string{"type T = 1  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = true & true;",
			Output: []string{"type T = true  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = null | null;",
			Output: []string{"type T = null  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = any | any;",
			Output: []string{"type T = any  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = { a: string | string };",
			Output: []string{"type T = { a: string   };"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = { a: string } | { a: string };",
			Output: []string{"type T = { a: string }  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = { a: string; b: number } | { a: string; b: number };",
			Output: []string{"type T = { a: string; b: number }  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = Set<string> | Set<string>;",
			Output: []string{"type T = Set<string>  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type IsArray<T> = T extends any[] ? true : false;
type ActuallyDuplicated = IsArray<number> | IsArray<string>;
      `,
			Output: []string{`
type IsArray<T> = T extends any[] ? true : false;
type ActuallyDuplicated = IsArray<number>  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = string[] | string[];",
			Output: []string{"type T = string[]  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = string[][] | string[][];",
			Output: []string{"type T = string[][]  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = [1, 2, 3] | [1, 2, 3];",
			Output: []string{"type T = [1, 2, 3]  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = () => string | string;",
			Output: []string{"type T = () => string  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = () => null | null;",
			Output: []string{"type T = () => null  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = (arg: string | string) => void;",
			Output: []string{"type T = (arg: string  ) => void;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = 'A' | 'A';",
			Output: []string{"type T = 'A'  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type T = A | A;
      `,
			Output: []string{`
type A = 'A';
type T = A  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
const a: A | A = 'A';
      `,
			Output: []string{`
type A = 'A';
const a: A   = 'A';
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type T = A | /* comment */ A;
      `,
			Output: []string{`
type A = 'A';
type T = A  /* comment */ ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A1 = 'A';
type A2 = 'A';
type A3 = 'A';
type T = A1 | A2 | A3;
      `,
			Output: []string{`
type A1 = 'A';
type A2 = 'A';
type A3 = 'A';
type T = A1    ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type T = A | B | A;
      `,
			Output: []string{`
type A = 'A';
type B = 'B';
type T = A | B  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type T = A | B | A | B;
      `,
			Output: []string{`
type A = 'A';
type B = 'B';
type T = A | B    ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type T = A | B | A | A;
      `,
			Output: []string{`
type A = 'A';
type B = 'B';
type T = A | B    ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type C = 'C';
type T = A | B | A | C;
      `,
			Output: []string{`
type A = 'A';
type B = 'B';
type C = 'C';
type T = A | B   | C;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type T = (A | B) | (A | B);
      `,
			Output: []string{`
type A = 'A';
type B = 'B';
type T = (A | B)  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type T = A | (A | A);
      `,
			Output: []string{`
type A = 'A';
type T = A  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type C = 'C';
type D = 'D';
type F = (A | B) | (A | B) | ((C | D) & (A | B)) | (A | B);
      `,
			Output: []string{`
type A = 'A';
type B = 'B';
type C = 'C';
type D = 'D';
type F = (A | B)   | ((C | D) & (A | B))  ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type B = 'B';
type C = (A | B) | A | B | (A | B);
      `,
			Output: []string{`
type A = 'A';
type B = 'B';
type C = (A | B)      ;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
				{
					MessageId: "duplicate",
				},
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type A = (number | string) | number | string;",
			Output: []string{"type A = (number | string)    ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type A = (number | (string | null)) | (string | (null | number));",
			Output: []string{"type A = (number | (string | null))  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type A = (number & string) & number & string;",
			Output: []string{"type A = (number & string)    ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: "type A = number & string & (number & string);",
			Output: []string{"type A = number & string & (  string);",
				"type A = number & string    ;",
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code: `
type A = 'A';
type T = Record<string, A | A>;
      `,
			Output: []string{`
type A = 'A';
type T = Record<string, A  >;
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "type T = A | A | string | string;",
			Output: []string{"type T = A | A | string  ;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "duplicate",
				},
			},
		},
		{
			Code:   "(a?: string | undefined) => {};",
			Output: []string{"(a?: string  ) => {};"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code: `
        type T = undefined;
        (arg?: T | string) => {};
      `,
			Output: []string{`
        type T = undefined;
        (arg?:   string) => {};
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code: `
        interface F {
          (a?: string | undefined): void;
        }
      `,
			Output: []string{`
        interface F {
          (a?: string  ): void;
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code:   "type fn = new (a?: string | undefined) => void;",
			Output: []string{"type fn = new (a?: string  ) => void;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code:   "function f(a?: string | undefined) {}",
			Output: []string{"function f(a?: string  ) {}"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code:   "f = function (a?: string | undefined) {};",
			Output: []string{"f = function (a?: string  ) {};"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code:   "declare function f(a?: string | undefined): void;",
			Output: []string{"declare function f(a?: string  ): void;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code: `
        declare class bb {
          f(a?: string | undefined): void;
        }
      `,
			Output: []string{`
        declare class bb {
          f(a?: string  ): void;
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code: `
        interface ee {
          f(a?: string | undefined): void;
        }
      `,
			Output: []string{`
        interface ee {
          f(a?: string  ): void;
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code: `
        interface ee {
          new (a?: string | undefined): void;
        }
      `,
			Output: []string{`
        interface ee {
          new (a?: string  ): void;
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code:   "type fn = (a?: string | undefined) => void;",
			Output: []string{"type fn = (a?: string  ) => void;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code:   "type fn = (a?: string | (undefined | number)) => void;",
			Output: []string{"type fn = (a?: string | (  number)) => void;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code:   "type fn = (a?: (undefined | number) | string) => void;",
			Output: []string{"type fn = (a?: (  number) | string) => void;"},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
		{
			Code: `
        abstract class cc {
          abstract f(a?: string | undefined): void;
        }
      `,
			Output: []string{`
        abstract class cc {
          abstract f(a?: string  ): void;
        }
      `,
			},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unnecessary",
				},
			},
		},
	})
}
