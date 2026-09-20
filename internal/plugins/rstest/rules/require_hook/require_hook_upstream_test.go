// TestRequireHookUpstream migrates the complete @vitest/eslint-plugin v1.6.27
// suite to Rstest API spellings, together with the eslint-plugin-jest v29.16.0
// cases the Vitest suite dropped. Rstest provenance, ts-go statement shapes and
// the utilities-object contract live in require_hook_extras_test.go.
package require_hook_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/require_hook"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func useHookError(line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "useHook",
		Message:   "This should be done within a hook",
		Line:      line,
		Column:    column,
	}
}

func allowedFunctionCalls(names ...string) []any {
	items := make([]any, 0, len(names))
	for _, name := range names {
		items = append(items, name)
	}
	return []any{map[string]any{"allowedFunctionCalls": items}}
}

func TestRequireHookUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(), "tsconfig.json", t,
		&require_hook.RequireHookRule,
		[]rule_tester.ValidTestCase{
			{Code: `describe()`},
			{Code: `describe("just a title")`},
			{Code: `describe.for([])('%s', (value) => {})`},
			{Code: `describe('a test', () =>
  test('something', () => {
    expect(true).toBe(true);
  }));`},
			{Code: `test('it', () => {
  //
});`},
			{Code: `import { myFn } from '../functions';

test('myFn', () => {
  expect(myFn()).toBe(1);
});`},
			{Code: `const { myFn } = require('../functions');

test('myFn', () => {
  expect(myFn()).toBe(1);
});`},
			{Code: `class MockLogger {
  log() {}
}

test('myFn', () => {
  expect(myFn()).toBe(1);
});`},
			{Code: `const { myFn } = require('../functions');

describe('myFn', () => {
  it('returns one', () => {
    expect(myFn()).toBe(1);
  });
});`},
			{Code: `describe('some tests', () => {
  it('is true', () => {
    expect(true).toBe(true);
  });
});`},
			{Code: `describe('some tests', () => {
  it('is true', () => {
    expect(true).toBe(true);
  });

  describe('more tests', () => {
    it('is false', () => {
      expect(true).toBe(false);
    });
  });
});`},
			{Code: `describe('some tests', () => {
  let consoleLogSpy;

  beforeEach(() => {
    consoleLogSpy = rs.spyOn(console, 'log');
  });

  it('prints a message', () => {
    printMessage('hello world');

    expect(consoleLogSpy).toHaveBeenCalledWith('hello world');
  });
});`},
			{Code: `let consoleErrorSpy = null;

beforeEach(() => {
  consoleErrorSpy = rs.spyOn(console, 'error');
});`},
			{Code: `let consoleErrorSpy = undefined;

beforeEach(() => {
  consoleErrorSpy = rs.spyOn(console, 'error');
});`},
			{Code: `describe('some tests', () => {
  beforeEach(() => {
    setup();
  });
});`},
			{Code: `beforeEach(() => {
  initializeCityDatabase();
});

afterEach(() => {
  clearCityDatabase();
});

test('city database has Vienna', () => {
  expect(isCity('Vienna')).toBeTruthy();
});

test('city database has San Juan', () => {
  expect(isCity('San Juan')).toBeTruthy();
});`},
			{Code: `describe('cities', () => {
  beforeEach(() => {
    initializeCityDatabase();
  });

  test('city database has Vienna', () => {
    expect(isCity('Vienna')).toBeTruthy();
  });

  test('city database has San Juan', () => {
    expect(isCity('San Juan')).toBeTruthy();
  });

  afterEach(() => {
    clearCityDatabase();
  });
});`},
			{
				Code: `enableAutoDestroy(afterEach);

describe('some tests', () => {
  it('is false', () => {
    expect(true).toBe(true);
  });
});`,
				Options: allowedFunctionCalls("enableAutoDestroy"),
			},
			{Code: `import { describe, test } from '@rstest/core';

interface Context { value: boolean }

const customTest = test.extend<Context>({ value: true });

describe.concurrent("extends", () => {
  customTest("should not trigger lint rules", ({ expect, value }) => {
    expect(value).toBe(true);
  });
});`},
			// eslint-plugin-jest's TypeScript edition: an ambient declaration is
			// not an executable statement.
			{Code: `import { myFn } from '../functions';

declare module 'eslint' {
  namespace ESLint {
    interface LintResult {
      fatalErrorCount: number;
    }
  }
}

test('myFn', () => {
  expect(myFn()).toBe(1);
});`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `setup();`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},
			{
				Code: `describe('some tests', () => {
  setup();
});`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(2, 3)},
			},
			{
				Code: `let { setup } = require('./test-utils');

describe('some tests', () => {
  setup();
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					useHookError(1, 1),
					useHookError(4, 3),
				},
			},
			{
				Code: `describe('some tests', () => {
  setup();

  it('is true', () => {
    expect(true).toBe(true);
  });

  describe('more tests', () => {
    setup();

    it('is false', () => {
      expect(true).toBe(false);
    });
  });
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					useHookError(2, 3),
					useHookError(9, 5),
				},
			},
			{
				Code: `let consoleErrorSpy = rs.spyOn(console, 'error');

describe('when loading cities from the api', () => {
  let consoleWarnSpy = rs.spyOn(console, 'warn');
});`,
				Errors: []rule_tester.InvalidTestCaseError{
					useHookError(1, 1),
					useHookError(4, 3),
				},
			},
			{
				Code: `let consoleErrorSpy = null;

describe('when loading cities from the api', () => {
  let consoleWarnSpy = rs.spyOn(console, 'warn');
});`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(4, 3)},
			},
			{
				Code:   `let value = 1`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},
			{
				Code:   `let consoleErrorSpy, consoleWarnSpy = rs.spyOn(console, 'error');`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},
			{
				Code:   `let consoleErrorSpy = rs.spyOn(console, 'error'), consoleWarnSpy;`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},
			{
				Code: `import { database, isCity } from '../database';
import { loadCities } from '../api';

rs.mock('../api');

const initializeCityDatabase = () => {
  database.addCity('Vienna');
  database.addCity('San Juan');
  database.addCity('Wellington');
};

const clearCityDatabase = () => {
  database.clear();
};

initializeCityDatabase();

test('that persists cities', () => {
  expect(database.cities.length).toHaveLength(3);
});

test('city database has Vienna', () => {
  expect(isCity('Vienna')).toBeTruthy();
});

test('city database has San Juan', () => {
  expect(isCity('San Juan')).toBeTruthy();
});

describe('when loading cities from the api', () => {
  let consoleWarnSpy = rs.spyOn(console, 'warn');

  loadCities.mockResolvedValue(['Wellington', 'London']);

  it('does not duplicate cities', async () => {
    await database.loadCities();

    expect(database.cities).toHaveLength(4);
  });

  it('logs any duplicates', async () => {
    await database.loadCities();

    expect(consoleWarnSpy).toHaveBeenCalledWith(
      'Ignored duplicate cities: Wellington',
    );
  });
});

clearCityDatabase();`,
				Errors: []rule_tester.InvalidTestCaseError{
					useHookError(16, 1),
					useHookError(31, 3),
					useHookError(33, 3),
					useHookError(50, 1),
				},
			},
			{
				Code: `enableAutoDestroy(afterEach);

describe('some tests', () => {
  it('is false', () => {
    expect(true).toBe(true);
  });
});`,
				Options: allowedFunctionCalls("someOtherName"),
				Errors:  []rule_tester.InvalidTestCaseError{useHookError(1, 1)},
			},
			// eslint-plugin-jest's TypeScript edition: the ambient declaration
			// does not stop the suite body from being checked.
			{
				Code: `import { setup } from '../test-utils';

declare module 'eslint' {
  namespace ESLint {
    interface LintResult {
      fatalErrorCount: number;
    }
  }
}

describe('some tests', () => {
  setup();
});`,
				Errors: []rule_tester.InvalidTestCaseError{useHookError(12, 3)},
			},
		},
	)
}
