package no_duplicate_hooks_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_duplicate_hooks"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoDuplicateHooksRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_duplicate_hooks.NoDuplicateHooksRule,
		[]rule_tester.ValidTestCase{
			{Code: `describe("foo", () => {
  beforeEach(() => {})
  test("bar", () => {
    someFn();
  })
})`},
			{Code: `beforeEach(() => {})
test("bar", () => {
  someFn();
})`},
			{Code: `describe("foo", () => {
  beforeAll(() => {}),
  beforeEach(() => {})
  afterEach(() => {})
  afterAll(() => {})

  test("bar", () => {
    someFn();
  })
})`},
			{Code: `describe.skip("foo", () => {
  beforeEach(() => {}),
  beforeAll(() => {}),
  test("bar", () => {
    someFn();
  })
})
describe("foo", () => {
  beforeEach(() => {}),
  beforeAll(() => {}),
  test("bar", () => {
    someFn();
  })
})`},
			{Code: `describe("foo", () => {
  beforeEach(() => {}),
  test("bar", () => {
    someFn();
  })
  describe("inner_foo", () => {
    beforeEach(() => {})
    test("inner bar", () => {
      someFn();
    })
  })
})`},
			{Code: `describe.each(['hello'])('%s', () => {
  beforeEach(() => {});

  it('is fine', () => {});
});`},
			{Code: `describe('something', () => {
  describe.each(['hello'])('%s', () => {
    beforeEach(() => {});

    it('is fine', () => {});
  });

  describe.each(['world'])('%s', () => {
    beforeEach(() => {});

    it('is fine', () => {});
  });
});`},
			{Code: "describe.each``('%s', () => {\n  beforeEach(() => {});\n\n  it('is fine', () => {});\n});"},
			{Code: "describe('something', () => {\n  describe.each``('%s', () => {\n    beforeEach(() => {});\n\n    it('is fine', () => {});\n  });\n\n  describe.each``('%s', () => {\n    beforeEach(() => {});\n\n    it('is fine', () => {});\n  });\n});"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `describe("foo", () => {
  beforeEach(() => {}),
  beforeEach(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicateHook", Line: 3, Column: 3},
				},
			},
			{
				Code: `describe.skip("foo", () => {
  beforeEach(() => {}),
  beforeAll(() => {}),
  beforeAll(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicateHook", Line: 4, Column: 3},
				},
			},
			{
				Code: `describe.skip("foo", () => {
  afterEach(() => {}),
  afterEach(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicateHook", Line: 3, Column: 3},
				},
			},
			{
				Code: `import { afterEach } from '@jest/globals';

describe.skip("foo", () => {
  afterEach(() => {}),
  afterEach(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicateHook", Line: 5, Column: 3},
				},
			},
			{
				Code: `import { afterEach, afterEach as somethingElse } from '@jest/globals';

describe.skip("foo", () => {
  afterEach(() => {}),
  somethingElse(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicateHook", Line: 5, Column: 3},
				},
			},
			{
				Code: `describe.skip("foo", () => {
  afterAll(() => {}),
  afterAll(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicateHook", Line: 3, Column: 3},
				},
			},
			{
				Code: `afterAll(() => {}),
afterAll(() => {}),
test("bar", () => {
  someFn();
})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicateHook", Line: 2, Column: 1},
				},
			},
			{
				Code: `describe("foo", () => {
  beforeEach(() => {}),
  beforeEach(() => {}),
  beforeEach(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicateHook", Line: 3, Column: 3},
					{MessageId: "noDuplicateHook", Line: 4, Column: 3},
				},
			},
			{
				Code: `describe.skip("foo", () => {
  afterAll(() => {}),
  afterAll(() => {}),
  beforeAll(() => {}),
  beforeAll(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicateHook", Line: 3, Column: 3},
					{MessageId: "noDuplicateHook", Line: 5, Column: 3},
				},
			},
			{
				Code: `describe.skip("foo", () => {
  beforeEach(() => {}),
  beforeAll(() => {}),
  test("bar", () => {
    someFn();
  })
})
describe("foo", () => {
  beforeEach(() => {}),
  beforeEach(() => {}),
  beforeAll(() => {}),
  test("bar", () => {
    someFn();
  })
})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicateHook", Line: 10, Column: 3},
				},
			},
			{
				Code: `describe("foo", () => {
  beforeAll(() => {}),
  test("bar", () => {
    someFn();
  })
  describe("inner_foo", () => {
    beforeEach(() => {})
    beforeEach(() => {})
    test("inner bar", () => {
      someFn();
    })
  })
})`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicateHook", Line: 8, Column: 5},
				},
			},
			{
				Code: `describe.each(['hello'])('%s', () => {
  beforeEach(() => {});
  beforeEach(() => {});

  it('is not fine', () => {});
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicateHook", Line: 3, Column: 3},
				},
			},
			{
				Code: `describe('something', () => {
  describe.each(['hello'])('%s', () => {
    beforeEach(() => {});

    it('is fine', () => {});
  });

  describe.each(['world'])('%s', () => {
    beforeEach(() => {});
    beforeEach(() => {});

    it('is not fine', () => {});
  });
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicateHook", Line: 10, Column: 5},
				},
			},
			{
				Code: `describe('something', () => {
  describe.each(['hello'])('%s', () => {
    beforeEach(() => {});

    it('is fine', () => {});
  });

  describe.each(['world'])('%s', () => {
    describe('some more', () => {
      beforeEach(() => {});
      beforeEach(() => {});

      it('is not fine', () => {});
    });
  });
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicateHook", Line: 11, Column: 7},
				},
			},
			{
				Code: "describe.each``('%s', () => {\n  beforeEach(() => {});\n  beforeEach(() => {});\n\n  it('is fine', () => {});\n});",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicateHook", Line: 3, Column: 3},
				},
			},
			{
				Code: "describe('something', () => {\n  describe.each``('%s', () => {\n    beforeEach(() => {});\n\n    it('is fine', () => {});\n  });\n\n  describe.each``('%s', () => {\n    beforeEach(() => {});\n    beforeEach(() => {});\n\n    it('is not fine', () => {});\n  });\n});",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "noDuplicateHook", Line: 10, Column: 5},
				},
			},
		},
	)
}
