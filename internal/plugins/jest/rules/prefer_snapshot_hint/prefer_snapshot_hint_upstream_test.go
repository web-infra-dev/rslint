// Complete jest upstream suite; additions live in prefer_snapshot_hint_extras_test.go.
package prefer_snapshot_hint

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferSnapshotHintUpstream0(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferSnapshotHintRule, []rule_tester.ValidTestCase{
		{Code: "expect(something).toStrictEqual(somethingElse);", Options: []any{"always"}},
		{Code: "a().toEqual('b')", Options: []any{"always"}},
		{Code: "expect(a);", Options: []any{"always"}},
		{Code: "expect(1).toMatchSnapshot({}, \"my snapshot\");", Options: []any{"always"}},
		{Code: "expect(1).toMatchSnapshot(\"my snapshot\");", Options: []any{"always"}},
		{Code: "expect(1).toMatchSnapshot(`my snapshot`);", Options: []any{"always"}},
		{Code: "const x = {};\nexpect(1).toMatchSnapshot(x, \"my snapshot\");", Options: []any{"always"}},
		{Code: "expect(1).toThrowErrorMatchingSnapshot(\"my snapshot\");", Options: []any{"always"}},
		{Code: "expect(1).toMatchInlineSnapshot();", Options: []any{"always"}},
		{Code: "expect(1).toThrowErrorMatchingInlineSnapshot();", Options: []any{"always"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "expect(1).toMatchSnapshot();", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 11},
		}},
		{Code: "expect(1).toMatchSnapshot({});", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 11},
		}},
		{Code: "const x = \"we can't know if this is a string or not\";\nexpect(1).toMatchSnapshot(x);", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 11},
		}},
		{Code: "expect(1).toThrowErrorMatchingSnapshot();", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 1, Column: 11},
		}},
		{Code: "it('is true', () => {\n  expect(1).toMatchSnapshot();\n});", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 13},
		}},
		{Code: "it('is true', () => {\n  expect(1).toMatchSnapshot();\n  expect(2).toMatchSnapshot();\n});", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 13},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 3, Column: 13},
		}},
		{Code: "it('is true', () => {\n  expect(1).toMatchSnapshot();\n  expect(2).toThrowErrorMatchingSnapshot(\"my error\");\n});", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 13},
		}},
		{Code: "const expectSnapshot = value => {\n  expect(value).toMatchSnapshot();\n};", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 17},
		}},
		{Code: "const expectSnapshot = value => {\n  expect(value).toThrowErrorMatchingSnapshot();\n};", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 17},
		}},
		{Code: "it('is true', () => {\n  { expect(1).toMatchSnapshot(); }\n});", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 15},
		}},
		{Code: "const x = \"snapshot\";\nexpect(1).toMatchSnapshot(`my ${x}`);", Options: []any{"always"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 11},
		}},
	})
}
func TestPreferSnapshotHintUpstream1(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferSnapshotHintRule, []rule_tester.ValidTestCase{
		{Code: "expect(something).toStrictEqual(somethingElse);", Options: []any{"multi"}},
		{Code: "a().toEqual('b')", Options: []any{"multi"}},
		{Code: "expect(a);", Options: []any{"multi"}},
		{Code: "expect(1).toMatchSnapshot({}, \"my snapshot\");", Options: []any{"multi"}},
		{Code: "expect(1).toThrowErrorMatchingSnapshot(\"my snapshot\");", Options: []any{"multi"}},
		{Code: "expect(1).toMatchSnapshot({});", Options: []any{"multi"}},
		{Code: "expect(1).toThrowErrorMatchingSnapshot();", Options: []any{"multi"}},
		{Code: "it('is true', () => {\n  expect(1).toMatchSnapshot();\n});", Options: []any{"multi"}},
		{Code: "it('is true', () => {\n  expect(1).toMatchSnapshot(undefined, 'my first snapshot');\n});", Options: []any{"multi"}},
		{Code: "describe('my tests', () => {\n  it('is true', () => {\n    expect(1).toMatchSnapshot('this is a hint, all by itself');\n  });\n\n  it('is false', () => {\n    expect(2).toMatchSnapshot('this is a hint');\n    expect(2).toMatchSnapshot('and so is this');\n  });\n});", Options: []any{"multi"}},
		{Code: "it('is true', () => {\n  expect(1).toMatchSnapshot();\n});\n\nit('is false', () => {\n  expect(2).toMatchSnapshot('this is a hint');\n  expect(2).toMatchSnapshot('and so is this');\n});", Options: []any{"multi"}},
		{Code: "it('is true', () => {\n  expect(1).toMatchSnapshot();\n});\n\nit('is false', () => {\n  expect(2).toThrowErrorMatchingSnapshot();\n});", Options: []any{"multi"}},
		{Code: "it('is true', () => {\n  expect(1).toStrictEqual(1);\n  expect(1).toStrictEqual(2);\n  expect(1).toMatchSnapshot();\n});\n\nit('is false', () => {\n  expect(1).toStrictEqual(1);\n  expect(1).toStrictEqual(2);\n  expect(2).toThrowErrorMatchingSnapshot();\n});", Options: []any{"multi"}},
		{Code: "it('is true', () => {\n  expect(1).toMatchInlineSnapshot();\n});\n\nit('is false', () => {\n  expect(1).toMatchInlineSnapshot();\n  expect(1).toMatchInlineSnapshot();\n  expect(1).toThrowErrorMatchingInlineSnapshot();\n});", Options: []any{"multi"}},
		{Code: "it('is true', () => {\n  expect(1).toMatchSnapshot();\n});\n\nit('is false', () => {\n  expect(1).toMatchSnapshot();\n});", Options: []any{"multi"}},
		{Code: "import { it as itIs } from '@jest/globals';\n\nit('is true', () => {\n  expect(1).toMatchSnapshot();\n});\n\nitIs('false', () => {\n  expect(1).toMatchSnapshot();\n});", Options: []any{"multi"}},
		{Code: "const myReusableTestBody = (value, snapshotHint) => {\n  const innerFn = anotherValue => {\n    expect(anotherValue).toMatchSnapshot();\n\n    expect(value).toBe(1);\n  };\n\n  expect(value).toBe(1);\n};\n\nit('my test', () => {\n  expect(1).toMatchSnapshot();\n});", Options: []any{"multi"}},
		{Code: "const myReusableTestBody = (value, snapshotHint) => {\n  const innerFn = anotherValue => {\n    expect(value).toBe(1);\n  };\n\n  expect(value).toBe(1);\n  expect(anotherValue).toMatchSnapshot();\n};\n\nit('my test', () => {\n  expect(1).toMatchSnapshot();\n});", Options: []any{"multi"}},
		{Code: "const myReusableTestBody = (value, snapshotHint) => {\n  const innerFn = anotherValue => {\n    expect(anotherValue).toMatchSnapshot();\n\n    expect(value).toBe(1);\n  };\n\n  expect(value).toBe(1);\n};\n\nexpect(1).toMatchSnapshot();", Options: []any{"multi"}},
	}, []rule_tester.InvalidTestCase{
		{Code: "it('is true', () => {\n  expect(1).toMatchSnapshot();\n  expect(2).toMatchSnapshot();\n});", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 13},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 3, Column: 13},
		}},
		{Code: "it('is true', () => {\n  expect(1).toMatchSnapshot();\n  expect(2).toThrowErrorMatchingSnapshot();\n});", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 13},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 3, Column: 13},
		}},
		{Code: "it('is true', () => {\n  expect(1).toThrowErrorMatchingSnapshot();\n  expect(2).toMatchSnapshot();\n});", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 13},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 3, Column: 13},
		}},
		{Code: "it('is true', () => {\n  expect(1).toMatchSnapshot({});\n  expect(2).toMatchSnapshot({});\n});", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 13},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 3, Column: 13},
		}},
		{Code: "it('is true', () => {\n  expect(1).toMatchSnapshot({});\n  {\n    expect(2).toMatchSnapshot({});\n  }\n});", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 13},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 4, Column: 15},
		}},
		{Code: "it('is true', () => {\n  { expect(1).toMatchSnapshot(); }\n  { expect(2).toMatchSnapshot(); }\n});", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 15},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 3, Column: 15},
		}},
		{Code: "it('is true', () => {\n  expect(1).toMatchSnapshot();\n  expect(2).toMatchSnapshot(undefined, 'my second snapshot');\n});", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 13},
		}},
		{Code: "it('is true', () => {\n  expect(1).toMatchSnapshot({});\n  expect(2).toMatchSnapshot(undefined, 'my second snapshot');\n});", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 13},
		}},
		{Code: "it('is true', () => {\n  expect(1).toMatchSnapshot({}, 'my first snapshot');\n  expect(2).toMatchSnapshot(undefined);\n});", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 3, Column: 13},
		}},
		{Code: "it('is true', () => {\n  expect(1).toMatchSnapshot({}, 'my first snapshot');\n  expect(2).toMatchSnapshot(undefined);\n  expect(2).toMatchSnapshot();\n});", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 3, Column: 13},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 4, Column: 13},
		}},
		{Code: "it('is true', () => {\n  expect(2).toMatchSnapshot();\n  expect(1).toMatchSnapshot({}, 'my second snapshot');\n  expect(2).toMatchSnapshot();\n});", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 13},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 4, Column: 13},
		}},
		{Code: "it('is true', () => {\n  expect(2).toMatchSnapshot(undefined);\n  expect(2).toMatchSnapshot();\n  expect(1).toMatchSnapshot(null, 'my third snapshot');\n});", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 13},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 3, Column: 13},
		}},
		{Code: "describe('my tests', () => {\n  it('is true', () => {\n    expect(1).toMatchSnapshot();\n  });\n\n  it('is false', () => {\n    expect(2).toMatchSnapshot();\n    expect(2).toMatchSnapshot();\n  });\n});", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 7, Column: 15},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 8, Column: 15},
		}},
		{Code: "describe('my tests', () => {\n  it('is true', () => {\n    expect(1).toMatchSnapshot();\n  });\n\n  it('is false', () => {\n    expect(2).toMatchSnapshot();\n    expect(2).toMatchSnapshot('hello world');\n  });\n});", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 7, Column: 15},
		}},
		{Code: "describe('my tests', () => {\n  describe('more tests', () => {\n    it('is true', () => {\n      expect(1).toMatchSnapshot();\n    });\n  });\n\n  it('is false', () => {\n    expect(2).toMatchSnapshot();\n    expect(2).toMatchSnapshot('hello world');\n  });\n});", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 9, Column: 15},
		}},
		{Code: "describe('my tests', () => {\n  it('is true', () => {\n    expect(1).toMatchSnapshot();\n  });\n\n  describe('more tests', () => {\n    it('is false', () => {\n      expect(2).toMatchSnapshot();\n      expect(2).toMatchSnapshot('hello world');\n    });\n  });\n});", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 8, Column: 17},
		}},
		{Code: "import { describe as context, it as itIs } from '@jest/globals';\n\ndescribe('my tests', () => {\n  it('is true', () => {\n    expect(1).toMatchSnapshot();\n  });\n\n  context('more tests', () => {\n    itIs('false', () => {\n      expect(2).toMatchSnapshot();\n      expect(2).toMatchSnapshot('hello world');\n    });\n  });\n});", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 10, Column: 17},
		}},
		{Code: "const myReusableTestBody = (value, snapshotHint) => {\n  expect(value).toMatchSnapshot();\n\n  const innerFn = anotherValue => {\n    expect(anotherValue).toMatchSnapshot();\n  };\n\n  expect(value).toBe(1);\n  expect(value + 1).toMatchSnapshot(null);\n  expect(value + 2).toThrowErrorMatchingSnapshot(snapshotHint);\n};", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 17},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 5, Column: 26},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 9, Column: 21},
		}},
		{Code: "const myReusableTestBody = (value, snapshotHint) => {\n  expect(value).toMatchSnapshot();\n\n  const innerFn = anotherValue => {\n    expect(anotherValue).toMatchSnapshot();\n\n    expect(value).toBe(1);\n    expect(value + 1).toMatchSnapshot(null);\n    expect(value + 2).toMatchSnapshot(null, snapshotHint);\n  };\n};", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 2, Column: 17},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 5, Column: 26},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 8, Column: 23},
		}},
		{Code: "const myReusableTestBody = (value, snapshotHint) => {\n  const innerFn = anotherValue => {\n    expect(anotherValue).toMatchSnapshot();\n\n    expect(value).toBe(1);\n    expect(value + 1).toMatchSnapshot(null);\n    expect(value + 2).toMatchSnapshot(null, snapshotHint);\n  };\n\n  expect(value).toThrowErrorMatchingSnapshot();\n};", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 3, Column: 26},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 6, Column: 23},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 10, Column: 17},
		}},
		{Code: "const myReusableTestBody = (value, snapshotHint) => {\n  const innerFn = anotherValue => {\n    expect(anotherValue).toMatchSnapshot();\n\n    expect(value).toBe(1);\n  };\n\n  expect(value).toMatchSnapshot();\n};\n\nit('my test', () => {\n  expect(1).toMatchSnapshot();\n});", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 3, Column: 26},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 8, Column: 17},
		}},
		{Code: "const myReusableTestBody = value => {\n  expect(value).toMatchSnapshot();\n};\n\nexpect(1).toMatchSnapshot();\nexpect(1).toThrowErrorMatchingSnapshot();", Options: []any{"multi"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 5, Column: 11},
			{MessageId: "missingHint", Message: "You should provide a hint for this snapshot", Line: 6, Column: 11},
		}},
	})
}
