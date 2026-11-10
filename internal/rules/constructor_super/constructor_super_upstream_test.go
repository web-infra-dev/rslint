package constructor_super

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

// Complete ESLint v10.8.1 constructor-super suite; see the JS mirror for grouping.
func TestConstructorSuperUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &ConstructorSuperRule,
		[]rule_tester.ValidTestCase{
			{Code: "class A { }"},
			{Code: "class A { constructor() { } }"},
			{Code: "class A extends null { }"},
			{Code: "class A extends B { }"},
			{Code: "class A extends B { constructor() { super(); } }"},
			{Code: "class A extends B { constructor() { if (true) { super(); } else { super(); } } }"},
			{Code: "class A extends (class B {}) { constructor() { super(); } }"},
			{Code: "class A extends (B = C) { constructor() { super(); } }"},
			{Code: "class A extends (B &&= C) { constructor() { super(); } }"},
			{Code: "class A extends (B ||= C) { constructor() { super(); } }"},
			{Code: "class A extends (B ??= C) { constructor() { super(); } }"},
			{Code: "class A extends (B ||= 5) { constructor() { super(); } }"},
			{Code: "class A extends (B ??= 5) { constructor() { super(); } }"},
			{Code: "class A extends (B || C) { constructor() { super(); } }"},
			{Code: "class A extends (5 && B) { constructor() { super(); } }"},
			{Code: "class A extends (false && B) { constructor() { super(); } }"},
			{Code: "class A extends (B || 5) { constructor() { super(); } }"},
			{Code: "class A extends (B ?? 5) { constructor() { super(); } }"},
			{Code: "class A extends (a ? B : C) { constructor() { super(); } }"},
			{Code: "class A extends (B, C) { constructor() { super(); } }"},
			{Code: "class A { constructor() { class B extends C { constructor() { super(); } } } }"},
			{Code: "class A extends B { constructor() { super(); class C extends D { constructor() { super(); } } } }"},
			{Code: "class A extends B { constructor() { super(); class C { constructor() { } } } }"},
			{Code: "class A extends B { constructor() { a ? super() : super(); } }"},
			{Code: "class A extends B { constructor() { if (a) super(); else super(); } }"},
			{Code: "class A extends B { constructor() { switch (a) { case 0: super(); break; default: super(); } } }"},
			{Code: "class A extends B { constructor() { try {} finally { super(); } } }"},
			{Code: "class A extends B { constructor() { if (a) throw Error(); super(); } }"},
			{Code: "class A extends B { constructor() { if (true) return a; super(); } }"},
			{Code: "class A extends null { constructor() { return a; } }"},
			{Code: "class A { constructor() { return a; } }"},
			{Code: "class A extends B { constructor(a) { super(); for (const b of a) { this.a(); } } }"},
			{Code: "class A extends B { constructor(a) { super(); for (b in a) ( foo(b) ); } }"},
			{Code: "class Foo extends Object { constructor(method) { super(); this.method = method || function() {}; } }"},
			{Code: "class A extends Object {\n    constructor() {\n        super();\n        for (let i = 0; i < 0; i++);\n    }\n}"},
			{Code: "class A extends Object {\n    constructor() {\n        super();\n        for (; i < 0; i++);\n    }\n}"},
			{Code: "class A extends Object {\n    constructor() {\n        super();\n        for (let i = 0;; i++) {\n            if (foo) break;\n        }\n    }\n}"},
			{Code: "class A extends Object {\n    constructor() {\n        super();\n        for (let i = 0; i < 0;);\n    }\n}"},
			{Code: "class A extends Object {\n    constructor() {\n        super();\n        for (let i = 0;;) {\n            if (foo) break;\n        }\n    }\n}"},
			{Code: "\n            class A extends B {\n                constructor(props) {\n                    super(props);\n\n                    try {\n                        let arr = [];\n                        for (let a of arr) {\n                        }\n                    } catch (err) {\n                    }\n                }\n            }\n        "},
			{Code: "class A extends obj?.prop { constructor() { super(); } }"},
			{Code: "\n            class A extends Base {\n                constructor(list) {\n                    for (const a of list) {\n                        if (a.foo) {\n                            super(a);\n                            return;\n                        }\n                    }\n                    super();\n                }\n            }\n        "},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "class A extends null { constructor() { super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 40, EndLine: 1, EndColumn: 47},
			}},
			{Code: "class A extends null { constructor() { } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 24, EndLine: 1, EndColumn: 41},
			}},
			{Code: "class A extends 100 { constructor() { super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 39, EndLine: 1, EndColumn: 46},
			}},
			{Code: "class A extends 'test' { constructor() { super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 42, EndLine: 1, EndColumn: 49},
			}},
			{Code: "class A extends (B = 5) { constructor() { super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 43, EndLine: 1, EndColumn: 50},
			}},
			{Code: "class A extends (B && 5) { constructor() { super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 44, EndLine: 1, EndColumn: 51},
			}},
			{Code: "class A extends (B &&= 5) { constructor() { super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 45, EndLine: 1, EndColumn: 52},
			}},
			{Code: "class A extends (B += C) { constructor() { super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 44, EndLine: 1, EndColumn: 51},
			}},
			{Code: "class A extends (B -= C) { constructor() { super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 44, EndLine: 1, EndColumn: 51},
			}},
			{Code: "class A extends (B **= C) { constructor() { super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 45, EndLine: 1, EndColumn: 52},
			}},
			{Code: "class A extends (B |= C) { constructor() { super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 44, EndLine: 1, EndColumn: 51},
			}},
			{Code: "class A extends (B &= C) { constructor() { super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "badSuper", Message: "Unexpected 'super()' because 'super' is not a constructor.", Line: 1, Column: 44, EndLine: 1, EndColumn: 51},
			}},
			{Code: "class A extends B { constructor() { } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 38},
			}},
			{Code: "class A extends B { constructor() { for (var a of b) super.foo(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 68},
			}},
			{Code: "class A extends B { constructor() { for (var i = 1; i < 10; i++) super.foo(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 80},
			}},
			{Code: "class A extends B { constructor() { var c = class extends D { constructor() { super(); } } } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 93},
			}},
			{Code: "class A extends B { constructor() { var c = () => super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 61},
			}},
			{Code: "class A extends B { constructor() { class C extends D { constructor() { super(); } } } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 87},
			}},
			{Code: "class A extends B { constructor() { var C = class extends D { constructor() { super(); } } } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 93},
			}},
			{Code: "class A extends B { constructor() { super(); class C extends D { constructor() { } } } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 66, EndLine: 1, EndColumn: 83},
			}},
			{Code: "class A extends B { constructor() { super(); var C = class extends D { constructor() { } } } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 72, EndLine: 1, EndColumn: 89},
			}},
			{Code: "class A extends B { constructor() { if (a) super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 54},
			}},
			{Code: "class A extends B { constructor() { if (a); else super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 60},
			}},
			{Code: "class A extends B { constructor() { a && super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 52},
			}},
			{Code: "class A extends B { constructor() { switch (a) { case 0: super(); } } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 70},
			}},
			{Code: "class A extends B { constructor() { switch (a) { case 0: break; default: super(); } } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 86},
			}},
			{Code: "class A extends B { constructor() { try { super(); } catch (err) {} } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 70},
			}},
			{Code: "class A extends B { constructor() { try { a; } catch (err) { super(); } } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 74},
			}},
			{Code: "class A extends B { constructor() { if (a) return; super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 62},
			}},
			{Code: "class A extends B { constructor() { super(); super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 46, EndLine: 1, EndColumn: 53},
			}},
			{Code: "class A extends B { constructor() { super() || super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 48, EndLine: 1, EndColumn: 55},
			}},
			{Code: "class A extends B { constructor() { if (a) super(); super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 53, EndLine: 1, EndColumn: 60},
			}},
			{Code: "class A extends B { constructor() { switch (a) { case 0: super(); default: super(); } } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 76, EndLine: 1, EndColumn: 83},
			}},
			{Code: "class A extends B { constructor(a) { while (a) super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 1, Column: 21, EndLine: 1, EndColumn: 58},
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 1, Column: 48, EndLine: 1, EndColumn: 55},
			}},
			{Code: "class A extends B { constructor() { return; super(); } }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 1, Column: 21, EndLine: 1, EndColumn: 55},
			}},
			{Code: "class Foo extends Bar {\n                constructor() {\n                    for (a in b) for (c in d);\n                }\n            }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 2, Column: 17, EndLine: 4, EndColumn: 18},
			}},
			{Code: "class C extends D {\n\n                constructor() {\n                    do {\n                        something();\n                    } while (foo);\n                }\n\n            }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 3, Column: 17, EndLine: 7, EndColumn: 18},
			}},
			{Code: "class C extends D {\n\n                constructor() {\n                    for (let i = 1;;i++) {\n                        if (bar) {\n                            break;\n                        }\n                    }\n                }\n\n            }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingAll", Message: "Expected to call 'super()'.", Line: 3, Column: 17, EndLine: 9, EndColumn: 18},
			}},
			{Code: "class C extends D {\n\n                constructor() {\n                    do {\n                        super();\n                    } while (foo);\n                }\n\n            }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "duplicate", Message: "Unexpected duplicate 'super()'.", Line: 5, Column: 25, EndLine: 5, EndColumn: 32},
			}},
			{Code: "class C extends D {\n\n                constructor() {\n                    while (foo) {\n                        if (bar) {\n                            super();\n                            break;\n                        }\n                    }\n                }\n\n            }", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "missingSome", Message: "Lacked a call of 'super()' in some code paths.", Line: 3, Column: 17, EndLine: 10, EndColumn: 18},
			}},
		},
	)
}
