// TestSortKeysUpstream migrates the full valid/invalid suite from upstream
// eslint/tests/lib/rules/sort-keys.js 1:1. Position assertions cover
// line/column/endLine/endColumn for every invalid case, and the exact
// message text is asserted throughout. rslint-specific lock-in cases live in
// the sort_keys_extras_test.go file.
package sort_keys

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestSortKeysUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&SortKeysRule,
		[]rule_tester.ValidTestCase{
			// ---- default (asc) ----
			{Code: "var obj = {'':1, [``]:2}"},
			{Code: "var obj = {[``]:1, '':2}"},
			{Code: "var obj = {'':1, a:2}"},
			{Code: "var obj = {[``]:1, a:2}"},
			{Code: "var obj = {_:2, a:1, b:3} // default"},
			{Code: "var obj = {a:1, b:3, c:2}"},
			{Code: "var obj = {a:2, b:3, b_:1}"},
			{Code: "var obj = {C:3, b_:1, c:2}"},
			{Code: "var obj = {$:1, A:3, _:2, a:4}"},
			{Code: "var obj = {1:1, '11':2, 2:4, A:3}"},
			{Code: "var obj = {'#':1, 'Z':2, À:3, è:4}"},
			{Code: "var obj = { [/(?<zero>0)/]: 1, '/(?<zero>0)/': 2 }"},
			// ---- ignore non-simple computed properties. ----
			{Code: "var obj = {a:1, b:3, [a + b]: -1, c:2}"},
			{Code: "var obj = {'':1, [f()]:2, a:3}"},
			{Code: "var obj = {a:1, [b++]:2, '':3}", Options: []any{"desc"}},
			// ---- ignore properties separated by spread properties ----
			{Code: "var obj = {a:1, ...z, b:1}"},
			{Code: "var obj = {b:1, ...z, a:1}"},
			{Code: "var obj = {...a, b:1, ...c, d:1}"},
			{Code: "var obj = {...a, b:1, ...d, ...c, e:2, z:5}"},
			{Code: "var obj = {b:1, ...c, ...d, e:2}"},
			{Code: "var obj = {a:1, ...z, '':2}"},
			{Code: "var obj = {'':1, ...z, 'a':2}", Options: []any{"desc"}},
			// ---- not ignore properties not separated by spread properties ----
			{Code: "var obj = {...z, a:1, b:1}"},
			{Code: "var obj = {...z, ...c, a:1, b:1}"},
			{Code: "var obj = {a:1, b:1, ...z}"},
			{Code: "var obj = {...z, ...x, a:1, ...c, ...d, f:5, e:4}", Options: []any{"desc"}},
			// ---- works when spread occurs somewhere other than an object literal ----
			{Code: "function fn(...args) { return [...args].length; }"},
			{Code: "function g() {}; function f(...args) { return g(...args); }"},
			// ---- ignore destructuring patterns. ----
			{Code: "let {a, b} = {}"},
			// ---- nested ----
			{Code: "var obj = {a:1, b:{x:1, y:1}, c:1}"},
			// ---- asc ----
			{Code: "var obj = {_:2, a:1, b:3} // asc", Options: []any{"asc"}},
			{Code: "var obj = {a:1, b:3, c:2}", Options: []any{"asc"}},
			{Code: "var obj = {a:2, b:3, b_:1}", Options: []any{"asc"}},
			{Code: "var obj = {C:3, b_:1, c:2}", Options: []any{"asc"}},
			{Code: "var obj = {$:1, A:3, _:2, a:4}", Options: []any{"asc"}},
			{Code: "var obj = {1:1, '11':2, 2:4, A:3}", Options: []any{"asc"}},
			{Code: "var obj = {'#':1, 'Z':2, À:3, è:4}", Options: []any{"asc"}},
			// ---- asc, minKeys should ignore unsorted keys when number of keys is less than minKeys ----
			{Code: "var obj = {a:1, c:2, b:3}", Options: []any{"asc", map[string]any{"minKeys": 4}}},
			// ---- asc, insensitive ----
			{Code: "var obj = {_:2, a:1, b:3} // asc, insensitive", Options: []any{"asc", map[string]any{"caseSensitive": false}}},
			{Code: "var obj = {a:1, b:3, c:2}", Options: []any{"asc", map[string]any{"caseSensitive": false}}},
			{Code: "var obj = {a:2, b:3, b_:1}", Options: []any{"asc", map[string]any{"caseSensitive": false}}},
			{Code: "var obj = {b_:1, C:3, c:2}", Options: []any{"asc", map[string]any{"caseSensitive": false}}},
			{Code: "var obj = {b_:1, c:3, C:2}", Options: []any{"asc", map[string]any{"caseSensitive": false}}},
			{Code: "var obj = {$:1, _:2, A:3, a:4}", Options: []any{"asc", map[string]any{"caseSensitive": false}}},
			{Code: "var obj = {1:1, '11':2, 2:4, A:3}", Options: []any{"asc", map[string]any{"caseSensitive": false}}},
			{Code: "var obj = {'#':1, 'Z':2, À:3, è:4}", Options: []any{"asc", map[string]any{"caseSensitive": false}}},
			// ---- asc, insensitive, minKeys should ignore unsorted keys when number of keys is less than minKeys ----
			{Code: "var obj = {$:1, A:3, _:2, a:4}", Options: []any{"asc", map[string]any{"caseSensitive": false, "minKeys": 5}}},
			// ---- asc, natural ----
			{Code: "var obj = {_:2, a:1, b:3} // asc, natural", Options: []any{"asc", map[string]any{"natural": true}}},
			{Code: "var obj = {a:1, b:3, c:2}", Options: []any{"asc", map[string]any{"natural": true}}},
			{Code: "var obj = {a:2, b:3, b_:1}", Options: []any{"asc", map[string]any{"natural": true}}},
			{Code: "var obj = {C:3, b_:1, c:2}", Options: []any{"asc", map[string]any{"natural": true}}},
			{Code: "var obj = {$:1, _:2, A:3, a:4}", Options: []any{"asc", map[string]any{"natural": true}}},
			{Code: "var obj = {1:1, 2:4, '11':2, A:3}", Options: []any{"asc", map[string]any{"natural": true}}},
			{Code: "var obj = {'#':1, 'Z':2, À:3, è:4}", Options: []any{"asc", map[string]any{"natural": true}}},
			// ---- asc, natural, minKeys should ignore unsorted keys when number of keys is less than minKeys ----
			{Code: "var obj = {b_:1, a:2, b:3}", Options: []any{"asc", map[string]any{"natural": true, "minKeys": 4}}},
			// ---- asc, natural, insensitive ----
			{Code: "var obj = {_:2, a:1, b:3} // asc, natural, insensitive", Options: []any{"asc", map[string]any{"natural": true, "caseSensitive": false}}},
			{Code: "var obj = {a:1, b:3, c:2}", Options: []any{"asc", map[string]any{"natural": true, "caseSensitive": false}}},
			{Code: "var obj = {a:2, b:3, b_:1}", Options: []any{"asc", map[string]any{"natural": true, "caseSensitive": false}}},
			{Code: "var obj = {b_:1, C:3, c:2}", Options: []any{"asc", map[string]any{"natural": true, "caseSensitive": false}}},
			{Code: "var obj = {b_:1, c:3, C:2}", Options: []any{"asc", map[string]any{"natural": true, "caseSensitive": false}}},
			{Code: "var obj = {$:1, _:2, A:3, a:4}", Options: []any{"asc", map[string]any{"natural": true, "caseSensitive": false}}},
			{Code: "var obj = {1:1, 2:4, '11':2, A:3}", Options: []any{"asc", map[string]any{"natural": true, "caseSensitive": false}}},
			{Code: "var obj = {'#':1, 'Z':2, À:3, è:4}", Options: []any{"asc", map[string]any{"natural": true, "caseSensitive": false}}},
			// ---- asc, natural, insensitive, minKeys should ignore unsorted keys when number of keys is less than minKeys ----
			{Code: "var obj = {a:1, _:2, b:3}", Options: []any{"asc", map[string]any{"natural": true, "caseSensitive": false, "minKeys": 4}}},
			// ---- desc ----
			{Code: "var obj = {b:3, a:1, _:2} // desc", Options: []any{"desc"}},
			{Code: "var obj = {c:2, b:3, a:1}", Options: []any{"desc"}},
			{Code: "var obj = {b_:1, b:3, a:2}", Options: []any{"desc"}},
			{Code: "var obj = {c:2, b_:1, C:3}", Options: []any{"desc"}},
			{Code: "var obj = {a:4, _:2, A:3, $:1}", Options: []any{"desc"}},
			{Code: "var obj = {A:3, 2:4, '11':2, 1:1}", Options: []any{"desc"}},
			{Code: "var obj = {è:4, À:3, 'Z':2, '#':1}", Options: []any{"desc"}},
			// ---- desc, minKeys should ignore unsorted keys when number of keys is less than minKeys ----
			{Code: "var obj = {a:1, c:2, b:3}", Options: []any{"desc", map[string]any{"minKeys": 4}}},
			// ---- desc, insensitive ----
			{Code: "var obj = {b:3, a:1, _:2} // desc, insensitive", Options: []any{"desc", map[string]any{"caseSensitive": false}}},
			{Code: "var obj = {c:2, b:3, a:1}", Options: []any{"desc", map[string]any{"caseSensitive": false}}},
			{Code: "var obj = {b_:1, b:3, a:2}", Options: []any{"desc", map[string]any{"caseSensitive": false}}},
			{Code: "var obj = {c:2, C:3, b_:1}", Options: []any{"desc", map[string]any{"caseSensitive": false}}},
			{Code: "var obj = {C:2, c:3, b_:1}", Options: []any{"desc", map[string]any{"caseSensitive": false}}},
			{Code: "var obj = {a:4, A:3, _:2, $:1}", Options: []any{"desc", map[string]any{"caseSensitive": false}}},
			{Code: "var obj = {A:3, 2:4, '11':2, 1:1}", Options: []any{"desc", map[string]any{"caseSensitive": false}}},
			{Code: "var obj = {è:4, À:3, 'Z':2, '#':1}", Options: []any{"desc", map[string]any{"caseSensitive": false}}},
			// ---- desc, insensitive, minKeys should ignore unsorted keys when number of keys is less than minKeys ----
			{Code: "var obj = {$:1, _:2, A:3, a:4}", Options: []any{"desc", map[string]any{"caseSensitive": false, "minKeys": 5}}},
			// ---- desc, natural ----
			{Code: "var obj = {b:3, a:1, _:2} // desc, natural", Options: []any{"desc", map[string]any{"natural": true}}},
			{Code: "var obj = {c:2, b:3, a:1}", Options: []any{"desc", map[string]any{"natural": true}}},
			{Code: "var obj = {b_:1, b:3, a:2}", Options: []any{"desc", map[string]any{"natural": true}}},
			{Code: "var obj = {c:2, b_:1, C:3}", Options: []any{"desc", map[string]any{"natural": true}}},
			{Code: "var obj = {a:4, A:3, _:2, $:1}", Options: []any{"desc", map[string]any{"natural": true}}},
			{Code: "var obj = {A:3, '11':2, 2:4, 1:1}", Options: []any{"desc", map[string]any{"natural": true}}},
			{Code: "var obj = {è:4, À:3, 'Z':2, '#':1}", Options: []any{"desc", map[string]any{"natural": true}}},
			// ---- desc, natural, minKeys should ignore unsorted keys when number of keys is less than minKeys ----
			{Code: "var obj = {b_:1, a:2, b:3}", Options: []any{"desc", map[string]any{"natural": true, "minKeys": 4}}},
			// ---- desc, natural, insensitive ----
			{Code: "var obj = {b:3, a:1, _:2} // desc, natural, insensitive", Options: []any{"desc", map[string]any{"natural": true, "caseSensitive": false}}},
			{Code: "var obj = {c:2, b:3, a:1}", Options: []any{"desc", map[string]any{"natural": true, "caseSensitive": false}}},
			{Code: "var obj = {b_:1, b:3, a:2}", Options: []any{"desc", map[string]any{"natural": true, "caseSensitive": false}}},
			{Code: "var obj = {c:2, C:3, b_:1}", Options: []any{"desc", map[string]any{"natural": true, "caseSensitive": false}}},
			{Code: "var obj = {C:2, c:3, b_:1}", Options: []any{"desc", map[string]any{"natural": true, "caseSensitive": false}}},
			{Code: "var obj = {a:4, A:3, _:2, $:1}", Options: []any{"desc", map[string]any{"natural": true, "caseSensitive": false}}},
			{Code: "var obj = {A:3, '11':2, 2:4, 1:1}", Options: []any{"desc", map[string]any{"natural": true, "caseSensitive": false}}},
			{Code: "var obj = {è:4, À:3, 'Z':2, '#':1}", Options: []any{"desc", map[string]any{"natural": true, "caseSensitive": false}}},
			// ---- desc, natural, insensitive, minKeys should ignore unsorted keys when number of keys is less than minKeys ----
			{Code: "var obj = {a:1, _:2, b:3}", Options: []any{"desc", map[string]any{"natural": true, "caseSensitive": false, "minKeys": 4}}},
			// ---- allowLineSeparatedGroups option ----
			{Code: "\n                var obj = {\n                    e: 1,\n                    f: 2,\n                    g: 3,\n\n                    a: 4,\n                    b: 5,\n                    c: 6\n                }\n            ", Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}}},
			{Code: "\n                var obj = {\n                    b: 1,\n\n                    // comment\n                    a: 2,\n                    c: 3\n                }\n            ", Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}}},
			{Code: "\n                var obj = {\n                    b: 1\n\n                    ,\n\n                    // comment\n                    a: 2,\n                    c: 3\n                }\n            ", Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}}},
			{Code: "\n                var obj = {\n                    c: 1,\n                    d: 2,\n\n                    b() {\n                    },\n                    e: 4\n                }\n            ", Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}}},
			{Code: "\n                var obj = {\n                    c: 1,\n                    d: 2,\n                    // comment\n\n                    // comment\n                    b() {\n                    },\n                    e: 4\n                }\n            ", Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}}},
			{Code: "\n                var obj = {\n                  b,\n\n                  [a+b]: 1,\n                  a\n                }\n            ", Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}}},
			{Code: "\n                var obj = {\n                    c: 1,\n                    d: 2,\n\n                    a() {\n\n                    },\n\n                    // abce\n                    f: 3,\n\n                    /*\n\n                    */\n                    [a+b]: 1,\n                    cc: 1,\n                    e: 2\n                }\n            ", Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}}},
			{Code: "\n                var obj = {\n                    b: \"/*\",\n\n                    a: \"*/\",\n                }\n            ", Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}}},
			{Code: "\n                var obj = {\n                    b,\n                    /*\n                    */ //\n\n                    a\n                }\n            ", Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}}},
			{Code: "\n                var obj = {\n                    b,\n\n                    /*\n                    */ //\n                    a\n                }\n            ", Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}}},
			{Code: "\n                var obj = {\n                    b: 1\n\n                    ,a: 2\n                };\n            ", Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}}},
			{Code: "\n                var obj = {\n                    b: 1\n                // comment before comma\n\n                ,\n                a: 2\n                };\n            ", Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}}},
			{Code: "\n                var obj = {\n                  b,\n\n                  a,\n                  ...z,\n                  c\n                }\n            ", Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}}},
			{Code: "\n                var obj = {\n                  b,\n\n                  [foo()]: [\n\n                  ],\n                  a\n                }\n            ", Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}}},
			// ---- ignoreComputedKeys ----
			{Code: "var obj = { ['b']: 1, a: 2 }", Options: []any{"asc", map[string]any{"ignoreComputedKeys": true}}},
			{Code: "var obj = { a: 1, [c]: 2, b: 3 }", Options: []any{"asc", map[string]any{"ignoreComputedKeys": true}}},
			{Code: "var obj = { c: 1, ['b']: 2, a: 3 }", Options: []any{"asc", map[string]any{"ignoreComputedKeys": true}}}},
		[]rule_tester.InvalidTestCase{
			// ---- default (asc) ----
			{
				Code: "var obj = {a:1, '':2} // default",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. '' should be before 'a'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: "var obj = {a:1, [``]:2} // default",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. '' should be before 'a'.", Line: 1, Column: 18, EndLine: 1, EndColumn: 20},
				},
			},
			{
				Code: "var obj = {a:1, _:2, b:3} // default",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. '_' should be before 'a'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: "var obj = {a:1, c:2, b:3}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'b' should be before 'c'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code: "var obj = {b_:1, a:2, b:3}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b_'.", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code: "var obj = {b_:1, c:2, C:3}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'C' should be before 'c'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code: "var obj = {$:1, _:2, A:3, a:4}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'A' should be before '_'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code: "var obj = {1:1, 2:4, A:3, '11':2}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. '11' should be before 'A'.", Line: 1, Column: 27, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code: "var obj = {'#':1, À:3, 'Z':2, è:4}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'Z' should be before 'À'.", Line: 1, Column: 24, EndLine: 1, EndColumn: 27},
				},
			},
			{
				Code: "var obj = { null: 1, [/(?<zero>0)/]: 2 }",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. '/(?<zero>0)/' should be before 'null'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 35},
				},
			},
			// ---- not ignore properties not separated by spread properties ----
			{
				Code: "var obj = {...z, c:1, b:1}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'b' should be before 'c'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code: "var obj = {...z, ...c, d:4, b:1, ...y, ...f, e:2, a:1}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'b' should be before 'd'.", Line: 1, Column: 29, EndLine: 1, EndColumn: 30},
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'e'.", Line: 1, Column: 51, EndLine: 1, EndColumn: 52},
				},
			},
			{
				Code: "var obj = {c:1, b:1, ...a}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'b' should be before 'c'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code: "var obj = {...z, ...a, c:1, b:1}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'b' should be before 'c'.", Line: 1, Column: 29, EndLine: 1, EndColumn: 30},
				},
			},
			{
				Code: "var obj = {...z, b:1, a:1, ...d, ...c}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "var obj = {...z, a:2, b:0, ...x, ...c}",
				Options: []any{"desc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in descending order. 'b' should be before 'a'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "var obj = {...z, a:2, b:0, ...x}",
				Options: []any{"desc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in descending order. 'b' should be before 'a'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "var obj = {...z, '':1, a:2}",
				Options: []any{"desc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in descending order. 'a' should be before ''.", Line: 1, Column: 24, EndLine: 1, EndColumn: 25},
				},
			},
			// ---- ignore non-simple computed properties, but their position shouldn't affect other comparisons. ----
			{
				Code: "var obj = {a:1, [b+c]:2, '':3}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. '' should be before 'a'.", Line: 1, Column: 26, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:    "var obj = {'':1, [b+c]:2, a:3}",
				Options: []any{"desc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in descending order. 'a' should be before ''.", Line: 1, Column: 27, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:    "var obj = {b:1, [f()]:2, '':3, a:4}",
				Options: []any{"desc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in descending order. 'a' should be before ''.", Line: 1, Column: 32, EndLine: 1, EndColumn: 33},
				},
			},
			// ---- not ignore simple computed properties. ----
			{
				Code: "var obj = {a:1, b:3, [a]: -1, c:2}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			// ---- nested ----
			{
				Code: "var obj = {a:1, c:{y:1, x:1}, b:1}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'x' should be before 'y'.", Line: 1, Column: 25, EndLine: 1, EndColumn: 26},
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'b' should be before 'c'.", Line: 1, Column: 31, EndLine: 1, EndColumn: 32},
				},
			},
			// ---- asc ----
			{
				Code:    "var obj = {a:1, _:2, b:3} // asc",
				Options: []any{"asc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. '_' should be before 'a'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    "var obj = {a:1, c:2, b:3}",
				Options: []any{"asc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'b' should be before 'c'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "var obj = {b_:1, a:2, b:3}",
				Options: []any{"asc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b_'.", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "var obj = {b_:1, c:2, C:3}",
				Options: []any{"asc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'C' should be before 'c'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "var obj = {$:1, _:2, A:3, a:4}",
				Options: []any{"asc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'A' should be before '_'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "var obj = {1:1, 2:4, A:3, '11':2}",
				Options: []any{"asc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. '11' should be before 'A'.", Line: 1, Column: 27, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "var obj = {'#':1, À:3, 'Z':2, è:4}",
				Options: []any{"asc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'Z' should be before 'À'.", Line: 1, Column: 24, EndLine: 1, EndColumn: 27},
				},
			},
			// ---- asc, minKeys should error when number of keys is greater than or equal to minKeys ----
			{
				Code:    "var obj = {a:1, _:2, b:3}",
				Options: []any{"asc", map[string]any{"minKeys": 3}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. '_' should be before 'a'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			// ---- asc, insensitive ----
			{
				Code:    "var obj = {a:1, _:2, b:3} // asc, insensitive",
				Options: []any{"asc", map[string]any{"caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in insensitive ascending order. '_' should be before 'a'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    "var obj = {a:1, c:2, b:3}",
				Options: []any{"asc", map[string]any{"caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in insensitive ascending order. 'b' should be before 'c'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "var obj = {b_:1, a:2, b:3}",
				Options: []any{"asc", map[string]any{"caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in insensitive ascending order. 'a' should be before 'b_'.", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "var obj = {$:1, A:3, _:2, a:4}",
				Options: []any{"asc", map[string]any{"caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in insensitive ascending order. '_' should be before 'A'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "var obj = {1:1, 2:4, A:3, '11':2}",
				Options: []any{"asc", map[string]any{"caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in insensitive ascending order. '11' should be before 'A'.", Line: 1, Column: 27, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "var obj = {'#':1, À:3, 'Z':2, è:4}",
				Options: []any{"asc", map[string]any{"caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in insensitive ascending order. 'Z' should be before 'À'.", Line: 1, Column: 24, EndLine: 1, EndColumn: 27},
				},
			},
			// ---- asc, insensitive, minKeys should error when number of keys is greater than or equal to minKeys ----
			{
				Code:    "var obj = {a:1, _:2, b:3}",
				Options: []any{"asc", map[string]any{"caseSensitive": false, "minKeys": 3}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in insensitive ascending order. '_' should be before 'a'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			// ---- asc, natural ----
			{
				Code:    "var obj = {a:1, _:2, b:3} // asc, natural",
				Options: []any{"asc", map[string]any{"natural": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural ascending order. '_' should be before 'a'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    "var obj = {a:1, c:2, b:3}",
				Options: []any{"asc", map[string]any{"natural": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural ascending order. 'b' should be before 'c'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "var obj = {b_:1, a:2, b:3}",
				Options: []any{"asc", map[string]any{"natural": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural ascending order. 'a' should be before 'b_'.", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "var obj = {b_:1, c:2, C:3}",
				Options: []any{"asc", map[string]any{"natural": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural ascending order. 'C' should be before 'c'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "var obj = {$:1, A:3, _:2, a:4}",
				Options: []any{"asc", map[string]any{"natural": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural ascending order. '_' should be before 'A'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "var obj = {1:1, 2:4, A:3, '11':2}",
				Options: []any{"asc", map[string]any{"natural": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural ascending order. '11' should be before 'A'.", Line: 1, Column: 27, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "var obj = {'#':1, À:3, 'Z':2, è:4}",
				Options: []any{"asc", map[string]any{"natural": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural ascending order. 'Z' should be before 'À'.", Line: 1, Column: 24, EndLine: 1, EndColumn: 27},
				},
			},
			// ---- asc, natural, minKeys should error when number of keys is greater than or equal to minKeys ----
			{
				Code:    "var obj = {a:1, _:2, b:3}",
				Options: []any{"asc", map[string]any{"natural": true, "minKeys": 2}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural ascending order. '_' should be before 'a'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			// ---- asc, natural, insensitive ----
			{
				Code:    "var obj = {a:1, _:2, b:3} // asc, natural, insensitive",
				Options: []any{"asc", map[string]any{"natural": true, "caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural insensitive ascending order. '_' should be before 'a'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    "var obj = {a:1, c:2, b:3}",
				Options: []any{"asc", map[string]any{"natural": true, "caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural insensitive ascending order. 'b' should be before 'c'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "var obj = {b_:1, a:2, b:3}",
				Options: []any{"asc", map[string]any{"natural": true, "caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural insensitive ascending order. 'a' should be before 'b_'.", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "var obj = {$:1, A:3, _:2, a:4}",
				Options: []any{"asc", map[string]any{"natural": true, "caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural insensitive ascending order. '_' should be before 'A'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "var obj = {1:1, '11':2, 2:4, A:3}",
				Options: []any{"asc", map[string]any{"natural": true, "caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural insensitive ascending order. '2' should be before '11'.", Line: 1, Column: 25, EndLine: 1, EndColumn: 26},
				},
			},
			{
				Code:    "var obj = {'#':1, À:3, 'Z':2, è:4}",
				Options: []any{"asc", map[string]any{"natural": true, "caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural insensitive ascending order. 'Z' should be before 'À'.", Line: 1, Column: 24, EndLine: 1, EndColumn: 27},
				},
			},
			// ---- asc, natural, insensitive, minKeys should error when number of keys is greater than or equal to minKeys ----
			{
				Code:    "var obj = {a:1, _:2, b:3}",
				Options: []any{"asc", map[string]any{"natural": true, "caseSensitive": false, "minKeys": 3}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural insensitive ascending order. '_' should be before 'a'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			// ---- desc ----
			{
				Code:    "var obj = {'':1, a:'2'} // desc",
				Options: []any{"desc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in descending order. 'a' should be before ''.", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "var obj = {[``]:1, a:'2'} // desc",
				Options: []any{"desc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in descending order. 'a' should be before ''.", Line: 1, Column: 20, EndLine: 1, EndColumn: 21},
				},
			},
			{
				Code:    "var obj = {a:1, _:2, b:3} // desc",
				Options: []any{"desc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in descending order. 'b' should be before '_'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "var obj = {a:1, c:2, b:3}",
				Options: []any{"desc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in descending order. 'c' should be before 'a'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    "var obj = {b_:1, a:2, b:3}",
				Options: []any{"desc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in descending order. 'b' should be before 'a'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "var obj = {b_:1, c:2, C:3}",
				Options: []any{"desc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in descending order. 'c' should be before 'b_'.", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "var obj = {$:1, _:2, A:3, a:4}",
				Options: []any{"desc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in descending order. '_' should be before '$'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
					{MessageId: "sortKeys", Message: "Expected object keys to be in descending order. 'a' should be before 'A'.", Line: 1, Column: 27, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:    "var obj = {1:1, 2:4, A:3, '11':2}",
				Options: []any{"desc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in descending order. '2' should be before '1'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
					{MessageId: "sortKeys", Message: "Expected object keys to be in descending order. 'A' should be before '2'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "var obj = {'#':1, À:3, 'Z':2, è:4}",
				Options: []any{"desc"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in descending order. 'À' should be before '#'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
					{MessageId: "sortKeys", Message: "Expected object keys to be in descending order. 'è' should be before 'Z'.", Line: 1, Column: 31, EndLine: 1, EndColumn: 32},
				},
			},
			// ---- desc, minKeys should error when number of keys is greater than or equal to minKeys ----
			{
				Code:    "var obj = {a:1, _:2, b:3}",
				Options: []any{"desc", map[string]any{"minKeys": 3}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in descending order. 'b' should be before '_'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			// ---- desc, insensitive ----
			{
				Code:    "var obj = {a:1, _:2, b:3} // desc, insensitive",
				Options: []any{"desc", map[string]any{"caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in insensitive descending order. 'b' should be before '_'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "var obj = {a:1, c:2, b:3}",
				Options: []any{"desc", map[string]any{"caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in insensitive descending order. 'c' should be before 'a'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    "var obj = {b_:1, a:2, b:3}",
				Options: []any{"desc", map[string]any{"caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in insensitive descending order. 'b' should be before 'a'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "var obj = {b_:1, c:2, C:3}",
				Options: []any{"desc", map[string]any{"caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in insensitive descending order. 'c' should be before 'b_'.", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "var obj = {$:1, _:2, A:3, a:4}",
				Options: []any{"desc", map[string]any{"caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in insensitive descending order. '_' should be before '$'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
					{MessageId: "sortKeys", Message: "Expected object keys to be in insensitive descending order. 'A' should be before '_'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "var obj = {1:1, 2:4, A:3, '11':2}",
				Options: []any{"desc", map[string]any{"caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in insensitive descending order. '2' should be before '1'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
					{MessageId: "sortKeys", Message: "Expected object keys to be in insensitive descending order. 'A' should be before '2'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "var obj = {'#':1, À:3, 'Z':2, è:4}",
				Options: []any{"desc", map[string]any{"caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in insensitive descending order. 'À' should be before '#'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
					{MessageId: "sortKeys", Message: "Expected object keys to be in insensitive descending order. 'è' should be before 'Z'.", Line: 1, Column: 31, EndLine: 1, EndColumn: 32},
				},
			},
			// ---- desc, insensitive should error when number of keys is greater than or equal to minKeys ----
			{
				Code:    "var obj = {a:1, _:2, b:3}",
				Options: []any{"desc", map[string]any{"caseSensitive": false, "minKeys": 2}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in insensitive descending order. 'b' should be before '_'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			// ---- desc, natural ----
			{
				Code:    "var obj = {a:1, _:2, b:3} // desc, natural",
				Options: []any{"desc", map[string]any{"natural": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural descending order. 'b' should be before '_'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "var obj = {a:1, c:2, b:3}",
				Options: []any{"desc", map[string]any{"natural": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural descending order. 'c' should be before 'a'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    "var obj = {b_:1, a:2, b:3}",
				Options: []any{"desc", map[string]any{"natural": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural descending order. 'b' should be before 'a'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "var obj = {b_:1, c:2, C:3}",
				Options: []any{"desc", map[string]any{"natural": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural descending order. 'c' should be before 'b_'.", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "var obj = {$:1, _:2, A:3, a:4}",
				Options: []any{"desc", map[string]any{"natural": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural descending order. '_' should be before '$'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural descending order. 'A' should be before '_'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural descending order. 'a' should be before 'A'.", Line: 1, Column: 27, EndLine: 1, EndColumn: 28},
				},
			},
			{
				Code:    "var obj = {1:1, 2:4, A:3, '11':2}",
				Options: []any{"desc", map[string]any{"natural": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural descending order. '2' should be before '1'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural descending order. 'A' should be before '2'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "var obj = {'#':1, À:3, 'Z':2, è:4}",
				Options: []any{"desc", map[string]any{"natural": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural descending order. 'À' should be before '#'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural descending order. 'è' should be before 'Z'.", Line: 1, Column: 31, EndLine: 1, EndColumn: 32},
				},
			},
			// ---- desc, natural should error when number of keys is greater than or equal to minKeys ----
			{
				Code:    "var obj = {a:1, _:2, b:3}",
				Options: []any{"desc", map[string]any{"natural": true, "minKeys": 3}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural descending order. 'b' should be before '_'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			// ---- desc, natural, insensitive ----
			{
				Code:    "var obj = {a:1, _:2, b:3} // desc, natural, insensitive",
				Options: []any{"desc", map[string]any{"natural": true, "caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural insensitive descending order. 'b' should be before '_'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "var obj = {a:1, c:2, b:3}",
				Options: []any{"desc", map[string]any{"natural": true, "caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural insensitive descending order. 'c' should be before 'a'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
				},
			},
			{
				Code:    "var obj = {b_:1, a:2, b:3}",
				Options: []any{"desc", map[string]any{"natural": true, "caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural insensitive descending order. 'b' should be before 'a'.", Line: 1, Column: 23, EndLine: 1, EndColumn: 24},
				},
			},
			{
				Code:    "var obj = {b_:1, c:2, C:3}",
				Options: []any{"desc", map[string]any{"natural": true, "caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural insensitive descending order. 'c' should be before 'b_'.", Line: 1, Column: 18, EndLine: 1, EndColumn: 19},
				},
			},
			{
				Code:    "var obj = {$:1, _:2, A:3, a:4}",
				Options: []any{"desc", map[string]any{"natural": true, "caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural insensitive descending order. '_' should be before '$'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural insensitive descending order. 'A' should be before '_'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			{
				Code:    "var obj = {1:1, 2:4, '11':2, A:3}",
				Options: []any{"desc", map[string]any{"natural": true, "caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural insensitive descending order. '2' should be before '1'.", Line: 1, Column: 17, EndLine: 1, EndColumn: 18},
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural insensitive descending order. '11' should be before '2'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26},
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural insensitive descending order. 'A' should be before '11'.", Line: 1, Column: 30, EndLine: 1, EndColumn: 31},
				},
			},
			{
				Code:    "var obj = {'#':1, À:3, 'Z':2, è:4}",
				Options: []any{"desc", map[string]any{"natural": true, "caseSensitive": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural insensitive descending order. 'À' should be before '#'.", Line: 1, Column: 19, EndLine: 1, EndColumn: 20},
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural insensitive descending order. 'è' should be before 'Z'.", Line: 1, Column: 31, EndLine: 1, EndColumn: 32},
				},
			},
			// ---- desc, natural, insensitive should error when number of keys is greater than or equal to minKeys ----
			{
				Code:    "var obj = {a:1, _:2, b:3}",
				Options: []any{"desc", map[string]any{"natural": true, "caseSensitive": false, "minKeys": 2}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in natural insensitive descending order. 'b' should be before '_'.", Line: 1, Column: 22, EndLine: 1, EndColumn: 23},
				},
			},
			// ---- When allowLineSeparatedGroups option is false ----
			{
				Code:    "\n                var obj = {\n                    b: 1,\n                    c: 2,\n                    a: 3\n                }\n            ",
				Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'c'.", Line: 5, Column: 21, EndLine: 5, EndColumn: 22},
				},
			},
			{
				Code:    "\n                let obj = {\n                    b\n\n                    ,a\n                }\n            ",
				Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": false}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 5, Column: 22, EndLine: 5, EndColumn: 23},
				},
			},
			{
				Code: "\n                let obj = {\n                    b\n\n                    ,a\n                }\n            ",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 5, Column: 22, EndLine: 5, EndColumn: 23},
				},
			},
			// ---- When allowLineSeparatedGroups option is true ----
			{
				Code:    "\n                 var obj = {\n                    b: 1,\n                    c () {\n\n                    },\n                    a: 3\n                  }\n             ",
				Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'c'.", Line: 7, Column: 21, EndLine: 7, EndColumn: 22},
				},
			},
			{
				Code:    "\n                 var obj = {\n                    a: 1,\n                    b: 2,\n\n                    z () {\n\n                    },\n                    y: 3\n                  }\n             ",
				Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'y' should be before 'z'.", Line: 9, Column: 21, EndLine: 9, EndColumn: 22},
				},
			},
			{
				Code:    "\n                 var obj = {\n                    b: 1,\n                    c () {\n                    },\n                    // comment\n                    a: 3\n                  }\n             ",
				Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'c'.", Line: 7, Column: 21, EndLine: 7, EndColumn: 22},
				},
			},
			{
				Code:    "\n                var obj = {\n                  b,\n                  [a+b]: 1,\n                  a // sort-keys: 'a' should be before 'b'\n                }\n            ",
				Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 5, Column: 19, EndLine: 5, EndColumn: 20},
				},
			},
			{
				Code:    "\n                var obj = {\n                    c: 1,\n                    d: 2,\n                    // comment\n                    // comment\n                    b() {\n                    },\n                    e: 4\n                }\n            ",
				Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'b' should be before 'd'.", Line: 7, Column: 21, EndLine: 7, EndColumn: 22},
				},
			},
			{
				Code:    "\n                var obj = {\n                    c: 1,\n                    d: 2,\n\n                    z() {\n\n                    },\n                    f: 3,\n                    /*\n\n\n                    */\n                    [a+b]: 1,\n                    b: 1,\n                    e: 2\n                }\n            ",
				Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'f' should be before 'z'.", Line: 9, Column: 21, EndLine: 9, EndColumn: 22},
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'b' should be before 'f'.", Line: 15, Column: 21, EndLine: 15, EndColumn: 22},
				},
			},
			{
				Code:    "\n                var obj = {\n                    b: \"/*\",\n                    a: \"*/\",\n                }\n            ",
				Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 4, Column: 21, EndLine: 4, EndColumn: 22},
				},
			},
			{
				Code:    "\n                var obj = {\n                    b: 1\n                    // comment before comma\n                    , a: 2\n                };\n            ",
				Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 5, Column: 23, EndLine: 5, EndColumn: 24},
				},
			},
			{
				Code:    "\n                let obj = {\n                  b,\n                  [foo()]: [\n                  // ↓ this blank is inside a property and therefore should not count\n\n                  ],\n                  a\n                }\n            ",
				Options: []any{"asc", map[string]any{"allowLineSeparatedGroups": true}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 8, Column: 19, EndLine: 8, EndColumn: 20},
				},
			},
			{
				Code:    "var obj = { d: 1, ['c']: 2, b: 3, a: 4 }",
				Options: []any{"asc", map[string]any{"ignoreComputedKeys": true, "minKeys": 4}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "sortKeys", Message: "Expected object keys to be in ascending order. 'a' should be before 'b'.", Line: 1, Column: 35, EndLine: 1, EndColumn: 36},
				},
			}},
	)
}
