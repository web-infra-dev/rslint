// TestBlockSpacingUpstream migrates the full valid/invalid suite from upstream
// packages/eslint-plugin/rules/block-spacing/block-spacing.test.ts 1:1.
// Position assertions cover line/column for every invalid case. rslint-specific
// lock-in cases live in block_spacing_extras_test.go.
package block_spacing_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/block_spacing"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func optsAlways() []interface{} { return []interface{}{"always"} }
func optsNever() []interface{}  { return []interface{}{"never"} }

func TestBlockSpacingUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&block_spacing.BlockSpacingRule,
		[]rule_tester.ValidTestCase{
			// ---- default/always ----
			{Code: `{ foo(); }`, Options: optsAlways()},
			{Code: `{ foo(); }`},
			{Code: "{ foo();\n}"},
			{Code: "{\nfoo(); }"},
			{Code: "{\r\nfoo();\r\n}"},
			{Code: `if (a) { foo(); }`},
			{Code: `if (a) {} else { foo(); }`},
			{Code: `switch (a) {}`},
			{Code: `switch (a) { case 0: foo(); }`},
			{Code: `while (a) { foo(); }`},
			{Code: `do { foo(); } while (a);`},
			{Code: `for (;;) { foo(); }`},
			{Code: `for (var a in b) { foo(); }`},
			{Code: `for (var a of b) { foo(); }`},
			{Code: `try { foo(); } catch (e) { foo(); }`},
			{Code: `function foo() { bar(); }`},
			{Code: `(function() { bar(); });`},
			{Code: `(() => { bar(); });`},
			{Code: `if (a) { /* comment */ foo(); /* comment */ }`},
			{Code: "if (a) { //comment\n foo(); }"},
			{Code: `class C { static {} }`},
			{Code: `class C { static { foo; } }`},
			{Code: `class C { static { /* comment */foo;/* comment */ } }`},

			// ---- never ----
			{Code: `{foo();}`, Options: optsNever()},
			{Code: "{foo();\n}", Options: optsNever()},
			{Code: "{\nfoo();}", Options: optsNever()},
			{Code: "{\r\nfoo();\r\n}", Options: optsNever()},
			{Code: `if (a) {foo();}`, Options: optsNever()},
			{Code: `if (a) {} else {foo();}`, Options: optsNever()},
			{Code: `switch (a) {}`, Options: optsNever()},
			{Code: `switch (a) {case 0: foo();}`, Options: optsNever()},
			{Code: `while (a) {foo();}`, Options: optsNever()},
			{Code: `do {foo();} while (a);`, Options: optsNever()},
			{Code: `for (;;) {foo();}`, Options: optsNever()},
			{Code: `for (var a in b) {foo();}`, Options: optsNever()},
			{Code: `for (var a of b) {foo();}`, Options: optsNever()},
			{Code: `try {foo();} catch (e) {foo();}`, Options: optsNever()},
			{Code: `function foo() {bar();}`, Options: optsNever()},
			{Code: `(function() {bar();});`, Options: optsNever()},
			{Code: `(() => {bar();});`, Options: optsNever()},
			{Code: `if (a) {/* comment */ foo(); /* comment */}`, Options: optsNever()},
			{Code: "if (a) { //comment\n foo();}", Options: optsNever()},
			{Code: `class C { static { } }`, Options: optsNever()},
			{Code: `class C { static {foo;} }`, Options: optsNever()},
			{Code: `class C { static {/* comment */ foo; /* comment */} }`, Options: optsNever()},
			{Code: "class C { static { // line comment is allowed\n foo;\n} }", Options: optsNever()},
			{Code: "class C { static {\nfoo;\n} }", Options: optsNever()},
			{Code: "class C { static { \n foo; \n } }", Options: optsNever()},
		},
		[]rule_tester.InvalidTestCase{
			// ---- default/always ----
			{
				Code:    `{foo();}`,
				Output:  []string{`{ foo(); }`},
				Options: optsAlways(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 1},
					{MessageId: "missing", Line: 1, Column: 8},
				},
			},
			{
				Code:   `{foo();}`,
				Output: []string{`{ foo(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 1},
					{MessageId: "missing", Line: 1, Column: 8},
				},
			},
			{
				Code:   `{ foo();}`,
				Output: []string{`{ foo(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 9},
				},
			},
			{
				Code:   `{foo(); }`,
				Output: []string{`{ foo(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 1},
				},
			},
			{
				Code:   "{\nfoo();}",
				Output: []string{"{\nfoo(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 2, Column: 7},
				},
			},
			{
				Code:   "{foo();\n}",
				Output: []string{"{ foo();\n}"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 1},
				},
			},
			{
				Code:   `if (a) {foo();}`,
				Output: []string{`if (a) { foo(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 8},
					{MessageId: "missing", Line: 1, Column: 15},
				},
			},
			{
				Code:   `if (a) {} else {foo();}`,
				Output: []string{`if (a) {} else { foo(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 16},
					{MessageId: "missing", Line: 1, Column: 23},
				},
			},
			{
				Code:   `switch (a) {case 0: foo();}`,
				Output: []string{`switch (a) { case 0: foo(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 12},
					{MessageId: "missing", Line: 1, Column: 27},
				},
			},
			{
				Code:   `while (a) {foo();}`,
				Output: []string{`while (a) { foo(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 11},
					{MessageId: "missing", Line: 1, Column: 18},
				},
			},
			{
				Code:   `do {foo();} while (a);`,
				Output: []string{`do { foo(); } while (a);`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 4},
					{MessageId: "missing", Line: 1, Column: 11},
				},
			},
			{
				Code:   `for (;;) {foo();}`,
				Output: []string{`for (;;) { foo(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 10},
					{MessageId: "missing", Line: 1, Column: 17},
				},
			},
			{
				Code:   `for (var a in b) {foo();}`,
				Output: []string{`for (var a in b) { foo(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 18},
					{MessageId: "missing", Line: 1, Column: 25},
				},
			},
			{
				Code:   `for (var a of b) {foo();}`,
				Output: []string{`for (var a of b) { foo(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 18},
					{MessageId: "missing", Line: 1, Column: 25},
				},
			},
			{
				Code:   `try {foo();} catch (e) {foo();} finally {foo();}`,
				Output: []string{`try { foo(); } catch (e) { foo(); } finally { foo(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
					{MessageId: "missing", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "missing", Line: 1, Column: 24, EndLine: 1, EndColumn: 25},
					{MessageId: "missing", Line: 1, Column: 31, EndLine: 1, EndColumn: 32},
					{MessageId: "missing", Line: 1, Column: 41, EndLine: 1, EndColumn: 42},
					{MessageId: "missing", Line: 1, Column: 48, EndLine: 1, EndColumn: 49},
				},
			},
			{
				Code:   `function foo() {bar();}`,
				Output: []string{`function foo() { bar(); }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 16},
					{MessageId: "missing", Line: 1, Column: 23},
				},
			},
			{
				Code:   `(function() {bar();});`,
				Output: []string{`(function() { bar(); });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 13},
					{MessageId: "missing", Line: 1, Column: 20},
				},
			},
			{
				Code:   `(() => {bar();});`,
				Output: []string{`(() => { bar(); });`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 8},
					{MessageId: "missing", Line: 1, Column: 15},
				},
			},
			{
				Code:   `if (a) {/* comment */ foo(); /* comment */}`,
				Output: []string{`if (a) { /* comment */ foo(); /* comment */ }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 8},
					{MessageId: "missing", Line: 1, Column: 43},
				},
			},
			{
				Code:   "if (a) {//comment\n foo(); }",
				Output: []string{"if (a) { //comment\n foo(); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 8, EndLine: 1, EndColumn: 9},
				},
			},

			// ---- class static blocks (always) ----
			{
				Code:   `class C { static {foo; } }`,
				Output: []string{`class C { static { foo; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:   `class C { static { foo;} }`,
				Output: []string{`class C { static { foo; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 24, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:   `class C { static {foo;} }`,
				Output: []string{`class C { static { foo; } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
					{MessageId: "missing", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:   `class C { static {/* comment */} }`,
				Output: []string{`class C { static { /* comment */ } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
					{MessageId: "missing", Line: 1, Column: 32, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code:   `class C { static {/* comment 1 */ foo; /* comment 2 */} }`,
				Output: []string{`class C { static { /* comment 1 */ foo; /* comment 2 */ } }`},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
					{MessageId: "missing", Line: 1, Column: 55, EndLine: 1, EndColumn: 56},
				},
			},
			{
				Code:   "class C {\n static {foo()\nbar()} }",
				Output: []string{"class C {\n static { foo()\nbar() } }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 2, Column: 9, EndLine: 2, EndColumn: 10},
					{MessageId: "missing", Line: 3, Column: 6, EndLine: 3, EndColumn: 7},
				},
			},

			// ---- never ----
			{
				Code:    `{ foo(); }`,
				Output:  []string{`{foo();}`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
					{MessageId: "extra", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
				},
			},
			{
				Code:    `{ foo();}`,
				Output:  []string{`{foo();}`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code:    `{foo(); }`,
				Output:  []string{`{foo();}`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 8, EndLine: 1, EndColumn: 9},
				},
			},
			{
				Code:    "{\nfoo(); }",
				Output:  []string{"{\nfoo();}"},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 2, Column: 7, EndLine: 2, EndColumn: 8},
				},
			},
			{
				Code:    "{ foo();\n}",
				Output:  []string{"{foo();\n}"},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 2, EndLine: 1, EndColumn: 3},
				},
			},
			{
				Code:    `if (a) { foo(); }`,
				Output:  []string{`if (a) {foo();}`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
					{MessageId: "extra", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:    `if (a) {} else { foo(); }`,
				Output:  []string{`if (a) {} else {foo();}`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
					{MessageId: "extra", Line: 1, Column: 24, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    `switch (a) { case 0: foo(); }`,
				Output:  []string{`switch (a) {case 0: foo();}`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
					{MessageId: "extra", Line: 1, Column: 28, EndLine: 1, EndColumn: 29},
				},
			},
			{
				Code:    `while (a) { foo(); }`,
				Output:  []string{`while (a) {foo();}`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
					{MessageId: "extra", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    `do { foo(); } while (a);`,
				Output:  []string{`do {foo();} while (a);`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 5, EndLine: 1, EndColumn: 6},
					{MessageId: "extra", Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
				},
			},
			{
				Code:    `for (;;) { foo(); }`,
				Output:  []string{`for (;;) {foo();}`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 11, EndLine: 1, EndColumn: 12},
					{MessageId: "extra", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    `for (var a in b) { foo(); }`,
				Output:  []string{`for (var a in b) {foo();}`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
					{MessageId: "extra", Line: 1, Column: 26, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:    `for (var a of b) { foo(); }`,
				Output:  []string{`for (var a of b) {foo();}`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
					{MessageId: "extra", Line: 1, Column: 26, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code:    `try { foo(); } catch (e) { foo(); } finally { foo(); }`,
				Output:  []string{`try {foo();} catch (e) {foo();} finally {foo();}`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 6, EndLine: 1, EndColumn: 7},
					{MessageId: "extra", Line: 1, Column: 13, EndLine: 1, EndColumn: 14},
					{MessageId: "extra", Line: 1, Column: 27, EndLine: 1, EndColumn: 28},
					{MessageId: "extra", Line: 1, Column: 34, EndLine: 1, EndColumn: 35},
					{MessageId: "extra", Line: 1, Column: 46, EndLine: 1, EndColumn: 47},
					{MessageId: "extra", Line: 1, Column: 53, EndLine: 1, EndColumn: 54},
				},
			},
			{
				Code:    `function foo() { bar(); }`,
				Output:  []string{`function foo() {bar();}`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
					{MessageId: "extra", Line: 1, Column: 24, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    `(function() { bar(); });`,
				Output:  []string{`(function() {bar();});`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 14, EndLine: 1, EndColumn: 15},
					{MessageId: "extra", Line: 1, Column: 21, EndLine: 1, EndColumn: 22},
				},
			},
			{
				Code:    `(() => { bar(); });`,
				Output:  []string{`(() => {bar();});`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
					{MessageId: "extra", Line: 1, Column: 16, EndLine: 1, EndColumn: 17},
				},
			},
			{
				Code:    `if (a) { /* comment */ foo(); /* comment */ }`,
				Output:  []string{`if (a) {/* comment */ foo(); /* comment */}`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 9, EndLine: 1, EndColumn: 10},
					{MessageId: "extra", Line: 1, Column: 44, EndLine: 1, EndColumn: 45},
				},
			},
			{
				Code:    `(() => {   bar();});`,
				Output:  []string{`(() => {bar();});`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 9, EndLine: 1, EndColumn: 12},
				},
			},
			{
				Code:    `(() => {bar();   });`,
				Output:  []string{`(() => {bar();});`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 15, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    `(() => {   bar();   });`,
				Output:  []string{`(() => {bar();});`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 9, EndLine: 1, EndColumn: 12},
					{MessageId: "extra", Line: 1, Column: 18, EndLine: 1, EndColumn: 21},
				},
			},

			// ---- class static blocks (never) ----
			{
				Code:    `class C { static { foo;} }`,
				Output:  []string{`class C { static {foo;} }`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code:    `class C { static {foo; } }`,
				Output:  []string{`class C { static {foo;} }`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    `class C { static { foo; } }`,
				Output:  []string{`class C { static {foo;} }`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
					{MessageId: "extra", Line: 1, Column: 24, EndLine: 1, EndColumn: 25},
				},
			},
			{
				Code:    `class C { static { /* comment */ } }`,
				Output:  []string{`class C { static {/* comment */} }`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
					{MessageId: "extra", Line: 1, Column: 33, EndLine: 1, EndColumn: 34},
				},
			},
			{
				Code:    `class C { static { /* comment 1 */ foo; /* comment 2 */ } }`,
				Output:  []string{`class C { static {/* comment 1 */ foo; /* comment 2 */} }`},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
					{MessageId: "extra", Line: 1, Column: 56, EndLine: 1, EndColumn: 57},
				},
			},
			{
				Code:    "class C { static\n{   foo()\nbar()  } }",
				Output:  []string{"class C { static\n{foo()\nbar()} }"},
				Options: optsNever(),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "extra", Line: 2, Column: 2, EndLine: 2, EndColumn: 5},
					{MessageId: "extra", Line: 3, Column: 6, EndLine: 3, EndColumn: 8},
				},
			},
		},
	)
}
