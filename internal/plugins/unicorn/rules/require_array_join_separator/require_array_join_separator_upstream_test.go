// TestRequireArrayJoinSeparatorUpstream migrates the full valid/invalid suite
// from upstream test/require-array-join-separator.js 1:1. Position assertions
// cover line/column for every invalid case. rslint-specific lock-in cases live
// in require_array_join_separator_extras_test.go.
package require_array_join_separator_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	require_array_join_separator "github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/require_array_join_separator"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	messageID = "require-array-join-separator"
	message   = "Missing the separator argument."
)

func TestRequireArrayJoinSeparatorUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&require_array_join_separator.RequireArrayJoinSeparatorRule,
		[]rule_tester.ValidTestCase{
			{Code: `foo.join(",")`, FileName: "file.js"},
			{Code: `join()`, FileName: "file.js"},
			{Code: `foo.join(...[])`, FileName: "file.js"},
			{Code: `foo.join?.()`, FileName: "file.js"},
			{Code: `foo?.join?.()`, FileName: "file.js"},
			{Code: `foo[join]()`, FileName: "file.js"},
			{Code: `foo["join"]()`, FileName: "file.js"},
			{Code: `[].join.call(foo, ",")`, FileName: "file.js"},
			{Code: `[].join.call()`, FileName: "file.js"},
			{Code: `[].join.call(...[foo])`, FileName: "file.js"},
			{Code: `[].join?.call(foo)`, FileName: "file.js"},
			{Code: `[]?.join.call(foo)`, FileName: "file.js"},
			{Code: `[].join[call](foo)`, FileName: "file.js"},
			{Code: `[][join].call(foo)`, FileName: "file.js"},
			{Code: `[,].join.call(foo)`, FileName: "file.js"},
			{Code: `[].join.notCall(foo)`, FileName: "file.js"},
			{Code: `[].notJoin.call(foo)`, FileName: "file.js"},
			{Code: `Array.prototype.join.call(foo, "")`, FileName: "file.js"},
			{Code: `Array.prototype.join.call()`, FileName: "file.js"},
			{Code: `Array.prototype.join.call(...[foo])`, FileName: "file.js"},
			{Code: `Array.prototype.join?.call(foo)`, FileName: "file.js"},
			{Code: `Array.prototype?.join.call(foo)`, FileName: "file.js"},
			{Code: `Array?.prototype.join.call(foo)`, FileName: "file.js"},
			{Code: `Array.prototype.join[call](foo, "")`, FileName: "file.js"},
			{Code: `Array.prototype[join].call(foo)`, FileName: "file.js"},
			{Code: `Array[prototype].join.call(foo)`, FileName: "file.js"},
			{Code: `Array.prototype.join.notCall(foo)`, FileName: "file.js"},
			{Code: `Array.prototype.notJoin.call(foo)`, FileName: "file.js"},
			{Code: `Array.notPrototype.join.call(foo)`, FileName: "file.js"},
			{Code: `NotArray.prototype.join.call(foo)`, FileName: "file.js"},
			{Code: `path.join(__dirname, "./foo.js")`, FileName: "file.js"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     `foo.join()`,
				FileName: "file.js",
				Output:   []string{`foo.join(',')`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   message,
					Line:      1,
					Column:    9,
					EndLine:   1,
					EndColumn: 11,
				}},
			},
			{
				Code:     `[].join.call(foo)`,
				FileName: "file.js",
				Output:   []string{`[].join.call(foo, ',')`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   message,
					Line:      1,
					Column:    17,
					EndLine:   1,
					EndColumn: 18,
				}},
			},
			{
				Code:     `[].join.call(foo,)`,
				FileName: "file.js",
				Output:   []string{`[].join.call(foo, ',',)`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   message,
					Line:      1,
					Column:    18,
					EndLine:   1,
					EndColumn: 19,
				}},
			},
			{
				Code:     `[].join.call(foo , );`,
				FileName: "file.js",
				Output:   []string{`[].join.call(foo ,  ',',);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   message,
					Line:      1,
					Column:    19,
					EndLine:   1,
					EndColumn: 21,
				}},
			},
			{
				Code:     `Array.prototype.join.call(foo)`,
				FileName: "file.js",
				Output:   []string{`Array.prototype.join.call(foo, ',')`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   message,
					Line:      1,
					Column:    30,
					EndLine:   1,
					EndColumn: 31,
				}},
			},
			{
				Code:     `Array.prototype.join.call(foo, )`,
				FileName: "file.js",
				Output:   []string{`Array.prototype.join.call(foo,  ',',)`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   message,
					Line:      1,
					Column:    31,
					EndLine:   1,
					EndColumn: 33,
				}},
			},
			{
				Code: `(
	/**/
	[
		/**/
	]
		/**/
		.
		/**/
		join
		/**/
		.
		/**/
		call
		/**/
		(
			/**/
			(
				/**/
				foo
				/**/
			)
			/**/
			,
			/**/
		)/**/
)`,
				FileName: "file.js",
				Output: []string{`(
	/**/
	[
		/**/
	]
		/**/
		.
		/**/
		join
		/**/
		.
		/**/
		call
		/**/
		(
			/**/
			(
				/**/
				foo
				/**/
			)
			/**/
			,
			/**/
		 ',',)/**/
)`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   message,
					Line:      23,
					Column:    5,
					EndLine:   25,
					EndColumn: 4,
				}},
			},
			{
				Code:     `foo?.join()`,
				FileName: "file.js",
				Output:   []string{`foo?.join(',')`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: messageID,
					Message:   message,
					Line:      1,
					Column:    10,
					EndLine:   1,
					EndColumn: 12,
				}},
			},
		},
	)
}
