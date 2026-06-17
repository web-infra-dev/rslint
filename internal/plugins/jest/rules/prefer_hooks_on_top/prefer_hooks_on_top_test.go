package prefer_hooks_on_top_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_hooks_on_top"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferHooksOnTopRule(t *testing.T) {
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
