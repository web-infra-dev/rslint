package vars_on_top

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestVarsOnTopUpstream migrates the valid and invalid cases from
// eslint/tests/lib/rules/vars-on-top.js. rslint-specific edge shapes live in
// vars_on_top_extras_test.go.
func TestVarsOnTopUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &VarsOnTopRule,
		[]rule_tester.ValidTestCase{
			{Code: "var first = 0; function foo() { first = 2; }"},
			{Code: "function foo() {}"},
			{Code: "function foo() { var first; if (true) { first = true; } }"},
			{Code: "function foo() { var first; var second = 1; var third; if (true) { third = true; } }"},
			{Code: "function foo() { var i; for (i = 0; i < 10; i++) { alert(i); } }"},
			{Code: "function foo() { var outer; function inner() { var inner = 1; var outer = inner; } outer = 1; }"},
			{Code: "function foo() { var first; // Hello\n var second = 1; first = second; }"},
			{Code: "function foo() { var first; /* Hello */ var second = 1; first = second; }"},
			{Code: "function foo() { var first; function bar() { var first; first = 5; } first = 2; }"},
			{Code: "function foo() { var first; var second = 1; function bar() { var third; third = 5; } first = second; }"},
			{Code: "function foo() { var first; var bar = function() { var third; third = 5; }; first = 5; }"},
			{Code: "function foo() { var first; first.onclick(function() { var third; third = 5; }); first = 5; }"},
			{Code: "function foo() { var i = 0; for (let j = 0; j < 10; j++) { alert(j); } i = i + 1; }", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6}},
			{Code: "'use strict'; var x; f();"},
			{Code: "'use strict'; 'directive'; var x; var y; f();"},
			{Code: "function f() { 'use strict'; var x; f(); }"},
			{Code: "function f() { 'use strict'; 'directive'; var x; var y; f(); }"},
			{Code: "import React from 'react'; var y; function f() { 'use strict'; var x; var y; f(); }", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6, SourceType: "module"}},
			{Code: "'use strict'; import React from 'react'; var y;", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6, SourceType: "module"}},
			{Code: "import React from 'react'; 'use strict'; var y;", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6, SourceType: "module"}},
			{Code: "import * as foo from 'mod.js'; 'use strict'; var y;", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6, SourceType: "module"}},
			{Code: "import { square, diag } from 'lib'; 'use strict'; var y;", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6, SourceType: "module"}},
			{Code: "import { default as foo } from 'lib'; 'use strict'; var y;", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6, SourceType: "module"}},
			{Code: "import 'src/mylib'; 'use strict'; var y;", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6, SourceType: "module"}},
			{Code: "import theDefault, { named1, named2 } from 'src/mylib'; 'use strict'; var y;", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6, SourceType: "module"}},
			{Code: "export var x; var y; var z;", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6, SourceType: "module"}},
			{Code: "var x; export var y; var z;", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6, SourceType: "module"}},
			{Code: "var x; var y; export var z;", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6, SourceType: "module"}},
			{Code: "class C { static { var x; } }", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: "class C { static { var x; foo(); } }", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: "class C { static { var x; var y; } }", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: "class C { static { var x; var y; foo(); } }", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: "class C { static { let x; var y; } }", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
			{Code: "class C { static { foo(); let x; } }", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "var first = 0; function foo() { first = 2; second = 2; } var second = 0;", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "function foo() { var first; first = 1; var second = 1; }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "function foo() { var first; if (true) { var second = true; } }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "function foo() { for (var i = 0; i < 10; i++) { alert(i); } }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "function foo() { var first = 10; var i; for (i = 0; i < first; i++) { var second = i; } }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "function foo() { var first = 10; var i; switch (first) { case 10: var hello = 1; break; } }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "function foo() { var first = 10; try { var hello = 1; } catch (e) {} }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "function foo() { var first = 10; try { asdf; } catch (e) { var hello = 1; } }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "function foo() { var first = 10; while (first) { var hello = 1; } }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "function foo() { var first = 10; do { var hello = 1; } while (first == 10); }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "function foo() { var first = [1,2,3]; for (var item in first) { item++; } }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "function foo() { var first = [1,2,3]; var item; for (item in first) { var hello = item; } }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "var foo = () => { var first = [1,2,3]; var item; for (item in first) { var hello = item; } }", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "'use strict'; 0; var x; f();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "'use strict'; var x; 'directive'; var y; f();", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "function f() { 'use strict'; 0; var x; f(); }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "function f() { 'use strict'; var x; 'directive'; var y; f(); }", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "export function f() {}\nvar x;", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6, SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "var x; export function f() {}\nvar y;", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6, SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "import {foo} from 'foo'; export {foo}; var test = 1;", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6, SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "export {foo} from 'foo';\nvar test = 1;", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6, SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "export * from 'foo';\nvar test = 1;", LanguageOptions: rule.LanguageOptions{ECMAVersion: 6, SourceType: "module"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "class C { static { foo(); var x; } }", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description, Line: 1, Column: 27}}},
			{Code: "class C { static { 'use strict'; var x; } }", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "class C {\n    static {\n        var x;\n        foo();\n        var y;\n    }\n}", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description, Line: 5}}},
			{Code: "class C { static { if (foo) { var x; } } }", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
			{Code: "class C { static { if (foo) var x; } }", LanguageOptions: rule.LanguageOptions{ECMAVersion: 2022}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "top", Message: varsOnTopMessage.Description}}},
		},
	)
}
