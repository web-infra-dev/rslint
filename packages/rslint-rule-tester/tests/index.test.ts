import path from 'node:path';
import { beforeEach, describe, expect, rs, test } from 'rstack/test' with {
  rstest: 'importActual',
};
import * as actualTest from 'rstack/test' with { rstest: 'importActual' };
import { lint } from '@rslint/core/internal';
import { RuleTester } from '../src/index.ts';

const runner = rs.hoisted(() => ({
  test: rs.fn<(name: string, callback: () => Promise<void>) => void>(),
  skip: rs.fn(),
  beforeAll: rs.fn<(callback: () => Promise<void>) => void>(),
  afterAll: rs.fn<(callback: () => Promise<void>) => void>(),
}));

rs.mock('rstack/test', () => ({
  ...actualTest,
  describe: (_name: string, callback: () => void) => {
    callback();
  },
  beforeAll: runner.beforeAll,
  afterAll: runner.afterAll,
  test: Object.assign(runner.test, { skip: runner.skip }),
}));

rs.mock('@rslint/core/internal', () => ({ lint: rs.fn() }));
rs.mock('../src/fixture.ts', () => ({
  createFixtureRoot: rs
    .fn<() => Promise<string>>()
    .mockResolvedValue(path.resolve('/fixture')),
  removeFixtureRoot: rs.fn<() => Promise<void>>().mockResolvedValue(undefined),
}));

describe('RuleTester.run', () => {
  const tester = new RuleTester({ plugin: 'node' });

  beforeEach(() => {
    rs.clearAllMocks();
  });

  test('runs core rules with the default constructor config', async () => {
    rs.mocked(lint).mockResolvedValue({
      diagnostics: [],
      errorCount: 0,
      warningCount: 0,
      fixableErrorCount: 0,
      fixableWarningCount: 0,
      fileCount: 1,
      ruleCount: 1,
    });
    new RuleTester().run('constructor-super', {
      valid: ['class A {}'],
      invalid: [],
    });
    await runner.test.mock.calls[0][1]();
    expect(lint).toHaveBeenCalledWith(
      expect.objectContaining({
        config: [
          expect.objectContaining({
            rules: { 'constructor-super': ['error'] },
          }),
        ],
      }),
    );
    expect(rs.mocked(lint).mock.calls[0][0].config?.[0]).toHaveProperty(
      'plugins',
      [],
    );
  });

  test.each(['valid', 'invalid'] as const)(
    'rejects a blank skip reason before registering a %s case',
    (kind) => {
      expect(() => {
        tester.run('no-process-exit', {
          valid: [],
          invalid: [],
          [kind]: [{ code: 'foo()', skip: '  ', errors: [] }],
        });
      }).toThrow("'skip' must be a non-empty reason");
      expect(runner.test).not.toHaveBeenCalled();
      expect(runner.skip).not.toHaveBeenCalled();
    },
  );

  test('registers a skipped case with a stated reason', () => {
    tester.run('no-process-exit', {
      valid: [{ code: 'foo()', skip: 'tsgo rejects this syntax' }],
      invalid: [],
    });
    expect(runner.skip).toHaveBeenCalledTimes(1);
    expect(runner.test).not.toHaveBeenCalled();
  });

  test('registers a runnable case with an undefined skip', () => {
    tester.run('no-process-exit', {
      valid: [{ code: 'foo()', skip: undefined }],
      invalid: [],
    });
    expect(runner.test).toHaveBeenCalledTimes(1);
    expect(runner.skip).not.toHaveBeenCalled();
  });
  test('creates a valid case only after fixture setup', async () => {
    const create = rs.fn((root: string) => ({ code: JSON.stringify(root) }));
    rs.mocked(lint).mockResolvedValue({
      diagnostics: [],
      errorCount: 0,
      warningCount: 0,
      fixableErrorCount: 0,
      fixableWarningCount: 0,
      fileCount: 1,
      ruleCount: 1,
    });
    tester.run('no-process-exit', {
      valid: [{ name: 'fixture case', create }],
      invalid: [],
    });
    expect(create).not.toHaveBeenCalled();
    await runner.beforeAll.mock.calls[0][0]();
    await runner.test.mock.calls[0][1]();
    expect(create).toHaveBeenCalledTimes(1);
    expect(create).toHaveBeenCalledWith(path.resolve('/fixture'));
    expect(lint).toHaveBeenCalledWith(
      expect.objectContaining({
        fileContents: {
          [path.resolve('/fixture/input.js')]: JSON.stringify(
            path.resolve('/fixture'),
          ),
        },
      }),
    );
  });

  test('uses one created case for diagnostics and the autofix pass', async () => {
    const create = rs.fn((root: string) => {
      const code = JSON.stringify(root);
      return {
        filename: 'src/input.js',
        code,
        settings: { 'import/resolver': { node: { paths: ['./lib'] } } },
        errors: [
          {
            messageId: 'expected',
            message: 'Expected diagnostic',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: code.length + 1,
          },
        ],
        output: '"fixed"',
      };
    });
    const root = path.resolve('/fixture');
    rs.mocked(lint)
      .mockResolvedValueOnce({
        diagnostics: [
          {
            ruleName: 'node/no-process-exit',
            messageId: 'expected',
            message: 'Expected diagnostic',
            severity: 'error',
            filePath: 'src/input.js',
            range: {
              start: { line: 1, column: 1 },
              end: { line: 1, column: JSON.stringify(root).length + 1 },
            },
            fixes: [
              {
                startPos: 0,
                endPos: JSON.stringify(root).length,
                text: '"fixed"',
              },
            ],
          },
        ],
        errorCount: 1,
        warningCount: 0,
        fixableErrorCount: 1,
        fixableWarningCount: 0,
        fileCount: 1,
        ruleCount: 1,
      })
      .mockResolvedValueOnce({
        diagnostics: [],
        output: { 'src/input.js': '"fixed"' },
        errorCount: 0,
        warningCount: 0,
        fixableErrorCount: 0,
        fixableWarningCount: 0,
        fileCount: 1,
        ruleCount: 1,
      });
    tester.run('no-process-exit', { valid: [], invalid: [{ create }] });
    await runner.beforeAll.mock.calls[0][0]();
    await runner.test.mock.calls[0][1]();
    expect(create).toHaveBeenCalledTimes(1);
    expect(lint).toHaveBeenCalledTimes(2);
    const [initial] = rs.mocked(lint).mock.calls[0];
    const [fixed] = rs.mocked(lint).mock.calls[1];
    expect(fixed).toEqual({ ...initial, fix: true });
    expect(initial.config?.[0]).toHaveProperty('settings', {
      'import/resolver': { node: { paths: ['./lib'] } },
    });
  });

  test('never creates an explicitly skipped case', async () => {
    const create = rs.fn(() => ({ code: 'x()' }));
    const reason = 'This input requires a case-insensitive filesystem.';
    tester.run('no-process-exit', {
      valid: [{ name: 'casing', skip: reason, create }],
      invalid: [],
    });
    await runner.beforeAll.mock.calls[0][0]();
    expect(runner.skip).toHaveBeenCalledTimes(1);
    expect(runner.skip.mock.calls[0][0]).toContain(reason);
    expect(create).not.toHaveBeenCalled();
    expect(lint).not.toHaveBeenCalled();
  });

  test('validates created errors before evaluating rule options', async () => {
    const options = rs.fn(() => []);
    tester.run('no-process-exit', {
      valid: [],
      invalid: [
        {
          create: () => ({ code: 'x()', errors: [], options }),
        },
      ],
    });
    await expect(runner.test.mock.calls[0][1]()).rejects.toThrow(
      'Invalid test case must have at least one error',
    );
    expect(options).not.toHaveBeenCalled();
    expect(lint).not.toHaveBeenCalled();
  });

  test.each(['name', 'skip', 'create'])(
    'rejects %s returned inside create',
    async (key) => {
      const entry = {
        create: () => ({ code: 'x()', [key]: 'late metadata' }),
      };
      tester.run('no-process-exit', { valid: [entry], invalid: [] });
      await expect(runner.test.mock.calls[0][1]()).rejects.toThrow(
        `${key} belongs outside create(root)`,
      );
      expect(lint).not.toHaveBeenCalled();
    },
  );
});
