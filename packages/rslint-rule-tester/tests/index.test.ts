import { beforeEach, describe, expect, rs, test } from 'rstack/test' with {
  rstest: 'importActual',
};
import * as actualTest from 'rstack/test' with { rstest: 'importActual' };
import { lint } from '@rslint/core/internal';
import { RuleTester } from '../src/index.ts';

const runner = rs.hoisted(() => ({
  test: rs.fn<(name: string, callback: () => Promise<void>) => void>(),
  skip: rs.fn(),
}));

rs.mock('rstack/test', () => ({
  ...actualTest,
  describe: (_name: string, callback: () => void) => {
    callback();
  },
  beforeAll: rs.fn(),
  afterAll: rs.fn(),
  test: Object.assign(runner.test, { skip: runner.skip }),
}));

rs.mock('@rslint/core/internal', () => ({ lint: rs.fn() }));

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
});
