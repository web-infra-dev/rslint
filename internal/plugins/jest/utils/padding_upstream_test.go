package utils_test

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/padding_around_after_all_blocks"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/padding_around_after_each_blocks"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/padding_around_all"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/padding_around_before_all_blocks"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/padding_around_before_each_blocks"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/padding_around_describe_blocks"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/padding_around_expect_groups"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/padding_around_test_blocks"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func paddingErrors(count int) []rule_tester.InvalidTestCaseError {
	errors := make([]rule_tester.InvalidTestCaseError, count)
	for i := range errors {
		errors[i].MessageId = "missingPadding"
	}
	return errors
}

func hookFixture(name string) (string, string) {
	code := fmt.Sprintf(`
const someText = 'abc';
%[1]s(() => {
});
describe('someText', () => {
  const something = 'abc';
  // A comment
  %[1]s(() => {
    // stuff
  });
  %[1]s(() => {
    // other stuff
  });
});

describe('someText', () => {
  const something = 'abc';
  %[1]s(() => {
    // stuff
  });
});
`, name)
	output := fmt.Sprintf(`
const someText = 'abc';

%[1]s(() => {
});

describe('someText', () => {
  const something = 'abc';

  // A comment
  %[1]s(() => {
    // stuff
  });

  %[1]s(() => {
    // other stuff
  });
});

describe('someText', () => {
  const something = 'abc';

  %[1]s(() => {
    // stuff
  });
});
`, name)
	return code, output
}

func runPaddingFixture(t *testing.T, lintRule *rule.Rule, code, output string, errorCount int) {
	t.Helper()
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		lintRule,
		[]rule_tester.ValidTestCase{{Code: output}},
		[]rule_tester.InvalidTestCase{{
			Code: code, Output: []string{output}, Errors: paddingErrors(errorCount),
		}},
	)
}

func TestPaddingAroundHookBlocksUpstream(t *testing.T) {
	tests := []struct {
		name string
		api  string
		rule *rule.Rule
	}{
		{name: "after-all", api: "afterAll", rule: &padding_around_after_all_blocks.PaddingAroundAfterAllBlocksRule},
		{name: "after-each", api: "afterEach", rule: &padding_around_after_each_blocks.PaddingAroundAfterEachBlocksRule},
		{name: "before-all", api: "beforeAll", rule: &padding_around_before_all_blocks.PaddingAroundBeforeAllBlocksRule},
		{name: "before-each", api: "beforeEach", rule: &padding_around_before_each_blocks.PaddingAroundBeforeEachBlocksRule},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			code, output := hookFixture(test.api)
			runPaddingFixture(t, test.rule, code, output, 5)
		})
	}
}

func TestPaddingAroundDescribeBlocksUpstream(t *testing.T) {
	code := `
foo();
bar();

const someText = 'abc';
const someObject = { one: 1, two: 2 };
// A comment before describe
describe('someText', () => {
  describe('some condition', () => {});
  describe('some other condition', () => {});
});
xdescribe('someObject', () => {
  describe('some condition', () => {
    const anotherThing = 500;
    describe('yet another condition', () => {});
  });
});fdescribe('weird', () => {});
describe.skip('skip me', () => {});
const BOOP = 'boop';
describe
  .skip('skip me too', () => {});
`
	output := `
foo();
bar();

const someText = 'abc';
const someObject = { one: 1, two: 2 };

// A comment before describe
describe('someText', () => {
  describe('some condition', () => {});

  describe('some other condition', () => {});
});

xdescribe('someObject', () => {
  describe('some condition', () => {
    const anotherThing = 500;

    describe('yet another condition', () => {});
  });
});

fdescribe('weird', () => {});

describe.skip('skip me', () => {});

const BOOP = 'boop';

describe
  .skip('skip me too', () => {});
`
	runPaddingFixture(t, &padding_around_describe_blocks.PaddingAroundDescribeBlocksRule, code, output, 8)
}

func TestPaddingAroundExpectGroupsUpstream(t *testing.T) {
	code := `
test('groups assertions', async () => {
  let value = 123;
  expect(value).toEqual(123);
  expect(123).toEqual(value); // Line comment
  value = 456;
  expect(value).toEqual(456);

  const load = () => Promise.resolve('ready');
  await expect(load()).resolves.toEqual('ready');
  expect(value).toEqual(456);

  await load();
  await expect(load()).resolves.toEqual('ready');
});
`
	output := `
test('groups assertions', async () => {
  let value = 123;

  expect(value).toEqual(123);
  expect(123).toEqual(value); // Line comment

  value = 456;

  expect(value).toEqual(456);

  const load = () => Promise.resolve('ready');

  await expect(load()).resolves.toEqual('ready');
  expect(value).toEqual(456);

  await load();

  await expect(load()).resolves.toEqual('ready');
});
`
	runPaddingFixture(t, &padding_around_expect_groups.PaddingAroundExpectGroupsRule, code, output, 5)
}

func TestPaddingAroundTestBlocksUpstream(t *testing.T) {
	code := `
const value = 'ready';
it('works', () => {});
fit('focuses', () => {});
test('also works', () => {});
describe('nested', () => {
  const nested = true;
  test.skip('skips', () => {});
  it.skip('also skips', () => {});
});xtest('disabled', () => {});
test
  .skip('multiline', () => {});
xit('disabled alias', () => {});
`
	output := `
const value = 'ready';

it('works', () => {});

fit('focuses', () => {});

test('also works', () => {});

describe('nested', () => {
  const nested = true;

  test.skip('skips', () => {});

  it.skip('also skips', () => {});
});

xtest('disabled', () => {});

test
  .skip('multiline', () => {});

xit('disabled alias', () => {});
`
	runPaddingFixture(t, &padding_around_test_blocks.PaddingAroundTestBlocksRule, code, output, 9)
}

func TestPaddingAroundAllUpstream(t *testing.T) {
	code := `const value = 'ready'
;afterEach(() => {})
test('switch values', () => {
  switch (value) {
  case 'ready':
    expect(value).toBe('ready');
    break;
  default:
    console.log(value);
  }
});`
	output := `const value = 'ready'

;afterEach(() => {})

test('switch values', () => {
  switch (value) {
  case 'ready':
    expect(value).toBe('ready');

    break;
  default:
    console.log(value);
  }
});`
	runPaddingFixture(t, &padding_around_all.PaddingAroundAllRule, code, output, 3)
}
