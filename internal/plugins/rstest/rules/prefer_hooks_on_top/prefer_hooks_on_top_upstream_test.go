package prefer_hooks_on_top_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_hooks_on_top"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestPreferHooksOnTopUpstream migrates the full valid/invalid suite the
// reference ESLint plugins keep for this rule, as preserved in rslint's Jest
// port. Position assertions cover line/column for every invalid case.
// Rstest-specific lock-in cases live in prefer_hooks_on_top_extras_test.go.
func TestPreferHooksOnTopUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_hooks_on_top.PreferHooksOnTopRule,
		[]rule_tester.ValidTestCase{
			{Code: `describe('foo', () => {
  beforeEach(() => {});
  someSetupFn();
  afterEach(() => {});

  test('bar', () => {
    someFn();
  });
})`},
			{Code: `describe('foo', () => {
  someSetupFn();
  beforeEach(() => {});
  afterEach(() => {});

  test('bar', () => {
    someFn();
  });
})`},
			{Code: `describe.skip('foo', () => {
  beforeEach(() => {});
  beforeAll(() => {});

  test('bar', () => {
    someFn();
  });
});

describe('foo', () => {
  beforeEach(() => {});

  test('bar', () => {
    someFn();
  });
});`},
			{Code: `describe('foo', () => {
  beforeEach(() => {});
  test('bar', () => {
    someFn();
  });

  describe('inner_foo', () => {
    beforeEach(() => {});
    test('inner bar', () => {
      someFn();
    });
  });
})`},
			// The reference plugin needs a modifier exemption here so that
			// building a fixture-extended test API does not close the scope.
			// The Rstest parser reaches the same result without one: a factory
			// call that is never invoked as a registration parses as no test
			// API at all. prefer_hooks_on_top_extras_test.go locks in both
			// halves of that distinction.
			{Code: `import { test as baseTest } from "@rstest/core";

const test = baseTest.extend({});

beforeEach(() => {});
afterEach(() => {});`},
			{Code: `import { it as baseIt } from "@rstest/core";

const it = baseIt.extend({});

beforeEach(() => {});
afterEach(() => {});`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `describe('foo', () => {
  beforeEach(() => {});
  test('bar', () => {
    someFn();
  });

  beforeAll(() => {});
  test('bar', () => {
    someFn();
  });
})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noHookOnTop", Line: 7, Column: 3},
				},
			},
			{
				Code: "describe('foo', () => {\n  beforeEach(() => {});\n  test.each``('bar', () => {\n    someFn();\n  });\n\n  beforeAll(() => {});\n  test.only('bar', () => {\n    someFn();\n  });\n})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noHookOnTop", Line: 7, Column: 3},
				},
			},
			{
				Code: "describe('foo', () => {\n  beforeEach(() => {});\n  test.only.each``('bar', () => {\n    someFn();\n  });\n\n  beforeAll(() => {});\n  test.only('bar', () => {\n    someFn();\n  });\n})",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noHookOnTop", Line: 7, Column: 3},
				},
			},
			{
				Code: `describe.skip('foo', () => {
  beforeEach(() => {});
  test('bar', () => {
    someFn();
  });

  beforeAll(() => {});
  test('bar', () => {
    someFn();
  });
});
describe('foo', () => {
  beforeEach(() => {});
  beforeEach(() => {});
  beforeAll(() => {});

  test('bar', () => {
    someFn();
  });
});

describe('foo', () => {
  test('bar', () => {
    someFn();
  });

  beforeEach(() => {});
  beforeEach(() => {});
  beforeAll(() => {});
})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noHookOnTop", Line: 7, Column: 3},
					{MessageId: "noHookOnTop", Line: 27, Column: 3},
					{MessageId: "noHookOnTop", Line: 28, Column: 3},
					{MessageId: "noHookOnTop", Line: 29, Column: 3},
				},
			},
			{
				Code: `describe('foo', () => {
  beforeAll(() => {});
  test('bar', () => {
    someFn();
  });

  describe('inner_foo', () => {
    beforeEach(() => {});
    test('inner bar', () => {
      someFn();
    });

    test('inner bar', () => {
      someFn();
    });

    beforeAll(() => {});
    afterAll(() => {});
    test('inner bar', () => {
      someFn();
    });
  });
})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noHookOnTop", Line: 17, Column: 5},
					{MessageId: "noHookOnTop", Line: 18, Column: 5},
				},
			},
		},
	)
}
