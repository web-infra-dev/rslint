// TestMaxExpectsUpstream migrates the full upstream eslint-plugin-jest
// v29.16.1 max-expects suite 1:1, plus the vitest chai-chain regression that
// rstest shares semantically. Position assertions cover every invalid case.
// rstest-specific lock-ins live in max_expects_extras_test.go.
package max_expects_test

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/max_expects"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

var (
	max1Option  = []any{map[string]any{"max": 1}}
	max2Option  = []any{map[string]any{"max": 2}}
	max3Option  = []any{map[string]any{"max": 3}}
	max5Option  = []any{map[string]any{"max": 5}}
	max10Option = []any{map[string]any{"max": 10}}
)

func exceededMaxError(count, maxAllowed, line, column int, endColumn ...int) rule_tester.InvalidTestCaseError {
	err := rule_tester.InvalidTestCaseError{
		MessageId: "exceededMaxAssertion",
		Message:   fmt.Sprintf("Too many assertion calls (%d) - maximum allowed is %d", count, maxAllowed),
		Line:      line,
		Column:    column,
	}
	if len(endColumn) == 1 {
		err.EndLine = line
		err.EndColumn = endColumn[0]
	}
	return err
}

func runMaxExpectsRuleTester(t *testing.T, valid []rule_tester.ValidTestCase, invalid []rule_tester.InvalidTestCase) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&max_expects.MaxExpectsRule,
		valid,
		invalid,
	)
}

func TestMaxExpectsUpstream(t *testing.T) {
	runMaxExpectsRuleTester(
		t,
		[]rule_tester.ValidTestCase{
			// ---- jest upstream valid ----
			{Code: `test('should pass')`},
			{Code: `test('should pass', () => {})`},
			{Code: `test.skip('should pass', () => {})`},
			{
				Code: `test('should pass', function () {
  expect(true).toBeDefined();
});`,
			},
			{
				Code: `test('should pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
			},
			{
				Code: `test('should pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  // expect(true).toBeDefined();
});`,
			},
			{
				Code: `it('should pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
			},
			{
				Code: `test('should pass', async () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
			},
			{
				Code: `test('should pass', async () => {
  expect.hasAssertions();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
			},
			{
				Code: `test('should pass', async () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toEqual(expect.any(Boolean));
});`,
			},
			{
				Code: `test('should pass', async () => {
  expect.hasAssertions();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toEqual(expect.any(Boolean));
});`,
			},
			{
				Code: `describe('test', () => {
  test('should pass', () => {
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
  });
});`,
			},
			{
				Code: `test.each(['should', 'pass'])('case', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
			},
			{
				Code: `test('should pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});
test('should pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
			},
			{
				Code: `test('should not pass', () => {
  const checkValue = (value) => {
    expect(value).toBeDefined();
    expect(value).toBeDefined();
  };

  checkValue(true);
});
test('should not pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
			},
			{
				Code: `test('should not pass', done => {
  emitter.on('event', value => {
    expect(value).toBeDefined();
    expect(value).toBeDefined();
    expect(value).toBeDefined();
    expect(value).toBeDefined();
    done();
  });
});
test('should not pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
			},
			{
				Code: `function myHelper() {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
}

test('should pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
			},
			{
				Code: `function myHelper1() {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
}

test('should pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});

function myHelper2() {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
}`,
			},
			{
				Code: `test('should pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});

function myHelper() {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
}`,
			},
			{
				Code: `const myHelper1 = () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
};

test('should pass', function() {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});

const myHelper2 = function() {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
};`,
			},
			{
				Code: `test('should pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
				Options: max10Option,
			},
			{
				Code: `describe('given decimal places', () => {
  it("test 1", fakeAsync(() => {
    expect(true).toBeTrue();
    expect(true).toBeTrue();
    expect(true).toBeTrue();
  }))

  it("test 2", fakeAsync(() => {
    expect(true).toBeTrue();
    expect(true).toBeTrue();
    expect(true).toBeTrue();
  }))
})`,
				Options: max5Option,
			},

			// ---- vitest reference parity: chai chain counts once ----
			{
				Code: `test('chai chain still counts once', () => {
  expect('hey').to.be.a('string').and.not.be.empty;
});`,
				Options: max1Option,
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- jest upstream invalid ----
			{
				Code: `test('should not pass', function () {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(6, 5, 7, 3, 29),
				},
			},
			{
				Code: `test('should not pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(6, 5, 7, 3, 29),
				},
			},
			{
				Code: `it('should not pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(6, 5, 7, 3, 29),
				},
			},
			{
				Code: `it('should not pass', async () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(6, 5, 7, 3, 29),
				},
			},
			{
				Code: `test('should not pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});
test('should not pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(6, 5, 7, 3, 29),
					exceededMaxError(6, 5, 15, 3, 29),
				},
			},
			{
				Code: `test('should not pass', () => {
  const checkValue = (value) => {
    expect(value).toBeDefined();
    expect(value).toBeDefined();
  };

  checkValue(true);
});
test('should not pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 4, 5, 32),
					exceededMaxError(2, 1, 11, 3, 29),
					exceededMaxError(3, 1, 12, 3, 29),
				},
			},
			{
				Code: `test('should not pass', () => {
  const checkValue = (value) => {
    expect(value).toBeDefined();
    expect(value).toBeDefined();
  };

  checkValue(true);
});
test('should not pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
				Options: max2Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(3, 2, 12, 3, 29),
				},
			},
			{
				Code: `test('should not pass', () => {
  const checkValue = (value) => {
    expect(value).toBeDefined();
    expect(value).toBeDefined();
  };

  expect(value).toBeDefined();
  checkValue(true);
});
test('should not pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
				Options: max2Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(3, 2, 13, 3, 29),
				},
			},
			{
				Code: `test('should not pass', done => {
  emitter.on('event', value => {
    expect(value).toBeDefined();
    expect(value).toBeDefined();

    done();
  });
});
test('should not pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 4, 5, 32),
					exceededMaxError(2, 1, 11, 3, 29),
					exceededMaxError(3, 1, 12, 3, 29),
				},
			},
			{
				Code: `test('should not pass', done => {
  emitter.on('event', value => {
    expect(value).toBeDefined();
    expect(value).toBeDefined();

    done();
  });
});
test('should not pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
				Options: max2Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(3, 2, 12, 3, 29),
				},
			},
			{
				Code: `describe('given decimal places', () => {
  it("test 1", fakeAsync(() => {
    expect(true).toBeTrue();
    expect(true).toBeTrue();
    expect(true).toBeTrue();
  }))

  it("test 2", fakeAsync(() => {
    expect(true).toBeTrue();
    expect(true).toBeTrue();
    expect(true).toBeTrue();
    expect(true).toBeTrue();
    expect(true).toBeTrue();
  }))
})`,
				Options: max3Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(4, 3, 12, 5, 28),
					exceededMaxError(5, 3, 13, 5, 28),
				},
			},
			{
				Code: `describe('test', () => {
  test('should not pass', () => {
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
    expect(true).toBeDefined();
  });
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(6, 5, 8, 5, 31),
				},
			},
			{
				Code: `test.each(['should', 'not', 'pass'])('case', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(6, 5, 7, 3, 29),
				},
			},
			{
				Code: `test('should not pass', () => {
  expect(true).toBeDefined();
  expect(true).toBeDefined();
});`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 3, 3, 29),
				},
			},
			{
				Code: `test('x', () => {
  expect(1).toBe(1);
  expect(1).toBe(1);
  expect(1).toBe(1);
  run((() => {
    expect(1).toBe(1);
    expect(1).toBe(1);
    expect(1).toBe(1);
  }));
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(6, 5, 8, 5, 22),
				},
			},

			// ---- vitest reference parity: chai chain counts once per factory ----
			{
				Code: `test('chai chains still count as assertions', () => {
  expect(true).toBeDefined();
  expect('hey').to.be.a('string');
});`,
				Options: max1Option,
				Errors: []rule_tester.InvalidTestCaseError{
					exceededMaxError(2, 1, 3, 3, 34),
				},
			},
		},
	)
}
