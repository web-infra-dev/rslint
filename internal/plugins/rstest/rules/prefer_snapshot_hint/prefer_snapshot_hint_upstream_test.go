// Complete vitest upstream suite; additions live in prefer_snapshot_hint_extras_test.go.
package prefer_snapshot_hint

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferSnapshotHintUpstream0(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferSnapshotHintRule, []rule_tester.ValidTestCase{
		{Code: "expect(something).toStrictEqual(somethingElse);", Options: []any{"multi"}},
		{Code: "a().toEqual('b')", Options: []any{"multi"}},
		{Code: "expect(a);", Options: []any{"multi"}},
		{Code: "expect(1).toMatchSnapshot({}, \"my snapshot\");", Options: []any{"multi"}},
		{Code: "expect(1).toThrowErrorMatchingSnapshot(\"my snapshot\");", Options: []any{"multi"}},
		{Code: "expect(1).toMatchSnapshot({});", Options: []any{"multi"}},
		{Code: "expect(1).toThrowErrorMatchingSnapshot();", Options: []any{"multi"}},
		{Code: "\n       it('is true', () => {\n      expect(1).toMatchSnapshot();\n       });\n     ", Options: []any{"multi"}},
		{Code: "\n       it('is true', () => {\n      expect(1).toMatchSnapshot(undefined, 'my first snapshot');\n       });\n     ", Options: []any{"multi"}},
		{Code: "\n       describe('my tests', () => {\n      it('is true', () => {\n        expect(1).toMatchSnapshot('this is a hint, all by itself');\n      });\n     \n      it('is false', () => {\n        expect(2).toMatchSnapshot('this is a hint');\n        expect(2).toMatchSnapshot('and so is this');\n      });\n       });\n     ", Options: []any{"multi"}},
		{Code: "\n       it('is true', () => {\n      expect(1).toMatchSnapshot();\n       });\n     \n       it('is false', () => {\n      expect(2).toMatchSnapshot('this is a hint');\n      expect(2).toMatchSnapshot('and so is this');\n       });\n     ", Options: []any{"multi"}},
		{Code: "\n       it('is true', () => {\n      expect(1).toMatchSnapshot();\n       });\n     \n       it('is false', () => {\n      expect(2).toThrowErrorMatchingSnapshot();\n       });\n     ", Options: []any{"multi"}},
		{Code: "\n       it('is true', () => {\n      expect(1).toStrictEqual(1);\n      expect(1).toStrictEqual(2);\n      expect(1).toMatchSnapshot();\n       });\n     \n       it('is false', () => {\n      expect(1).toStrictEqual(1);\n      expect(1).toStrictEqual(2);\n      expect(2).toThrowErrorMatchingSnapshot();\n       });\n     ", Options: []any{"multi"}},
		{Code: "\n       it('is true', () => {\n      expect(1).toMatchInlineSnapshot();\n       });\n     \n       it('is false', () => {\n      expect(1).toMatchInlineSnapshot();\n      expect(1).toMatchInlineSnapshot();\n      expect(1).toThrowErrorMatchingInlineSnapshot();\n       });\n     ", Options: []any{"multi"}},
		{Code: "\n       it('is true', () => {\n      expect(1).toMatchSnapshot();\n       });\n     \n       it('is false', () => {\n      expect(1).toMatchSnapshot();\n       });\n     ", Options: []any{"multi"}},
		{Code: "\n       const myReusableTestBody = (value, snapshotHint) => {\n      const innerFn = anotherValue => {\n        expect(anotherValue).toMatchSnapshot();\n     \n        expect(value).toBe(1);\n      };\n     \n      expect(value).toBe(1);\n       };\n     \n       it('my test', () => {\n      expect(1).toMatchSnapshot();\n       });\n     ", Options: []any{"multi"}},
		{Code: "\n       const myReusableTestBody = (value, snapshotHint) => {\n      const innerFn = anotherValue => {\n        expect(value).toBe(1);\n      };\n     \n      expect(value).toBe(1);\n      expect(anotherValue).toMatchSnapshot();\n       };\n     \n       it('my test', () => {\n      expect(1).toMatchSnapshot();\n       });\n     ", Options: []any{"multi"}},
		{Code: "\n       const myReusableTestBody = (value, snapshotHint) => {\n      const innerFn = anotherValue => {\n        expect(anotherValue).toMatchSnapshot();\n     \n        expect(value).toBe(1);\n      };\n     \n      expect(value).toBe(1);\n       };\n     \n       expect(1).toMatchSnapshot();\n     ", Options: []any{"multi"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "it('is true', () => {\n      expect(1).toMatchSnapshot();\n      expect(2).toMatchSnapshot();\n       });\n     ", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 17},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 3, Column: 17},
		}},
		{Code: "it('is true', () => {\n        expect(1).toMatchSnapshot();\n        expect(2).toThrowErrorMatchingSnapshot();\n      });\n       ", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 19},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 3, Column: 19},
		}},
		{Code: "it('is true', () => {\n        expect(1).toThrowErrorMatchingSnapshot();\n        expect(2).toMatchSnapshot();\n      });\n       ", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 19},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 3, Column: 19},
		}},
		{Code: "it('is true', () => {\n        expect(1).toMatchSnapshot({});\n        expect(2).toMatchSnapshot({});\n      });\n       ", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 19},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 3, Column: 19},
		}},
		{Code: "it('is true', () => {\n       expect(1).toMatchSnapshot({});\n       {\n      expect(2).toMatchSnapshot({});\n       }\n     });\n      ", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 18},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 4, Column: 17},
		}},
		{Code: "it('is true', () => {\n       { expect(1).toMatchSnapshot(); }\n       { expect(2).toMatchSnapshot(); }\n     });\n      ", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 20},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 3, Column: 20},
		}},
		{Code: "it('is true', () => {\n        expect(1).toMatchSnapshot();\n        expect(2).toMatchSnapshot(undefined, 'my second snapshot');\n      });\n       ", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 19},
		}},
		{Code: "it('is true', () => {\n        expect(1).toMatchSnapshot({});\n        expect(2).toMatchSnapshot(undefined, 'my second snapshot');\n      });", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 19},
		}},
		{Code: "it('is true', () => {\n        expect(1).toMatchSnapshot({}, 'my first snapshot');\n        expect(2).toMatchSnapshot(undefined);\n      });\n       ", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 3, Column: 19},
		}},
		{Code: "it('is true', () => {\n        expect(1).toMatchSnapshot({}, 'my first snapshot');\n        expect(2).toMatchSnapshot(undefined);\n        expect(2).toMatchSnapshot();\n      });\n       ", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 3, Column: 19},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 4, Column: 19},
		}},
		{Code: "it('is true', () => {\n        expect(2).toMatchSnapshot();\n        expect(1).toMatchSnapshot({}, 'my second snapshot');\n        expect(2).toMatchSnapshot();\n      });\n       ", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 19},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 4, Column: 19},
		}},
		{Code: "it('is true', () => {\n        expect(2).toMatchSnapshot(undefined);\n        expect(2).toMatchSnapshot();\n        expect(1).toMatchSnapshot(null, 'my third snapshot');\n      });\n       ", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 19},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 3, Column: 19},
		}},
		{Code: "describe('my tests', () => {\n        it('is true', () => {\n       expect(1).toMatchSnapshot();\n        });\n \n        it('is false', () => {\n       expect(2).toMatchSnapshot();\n       expect(2).toMatchSnapshot();\n        });\n      });\n       ", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 7, Column: 18},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 8, Column: 18},
		}},
		{Code: "describe('my tests', () => {\n        it('is true', () => {\n       expect(1).toMatchSnapshot();\n        });\n \n        it('is false', () => {\n       expect(2).toMatchSnapshot();\n       expect(2).toMatchSnapshot('hello world');\n        });\n      });\n       ", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 7, Column: 18},
		}},
		{Code: "describe('my tests', () => {\n        describe('more tests', () => {\n       it('is true', () => {\n         expect(1).toMatchSnapshot();\n       });\n        });\n \n        it('is false', () => {\n       expect(2).toMatchSnapshot();\n       expect(2).toMatchSnapshot('hello world');\n        });\n      });\n       ", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 9, Column: 18},
		}},
	})
}
