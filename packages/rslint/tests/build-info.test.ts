import { afterEach, describe, expect, rstest, test } from 'rstack/test';
import { buildInfo } from '../src/build-info.js';
import { runCLI } from '../src/cli/cli.js';
import { parseArgs } from '../src/utils/args.js';

rstest.mock('../src/internal/resolve-binary.js', () => ({
  resolveRslintBinary() {
    throw new Error('Version queries must not resolve a native binary');
  },
}));

describe('build information', () => {
  const originalExitCode = process.exitCode;
  afterEach(() => {
    rstest.restoreAllMocks();
    process.exitCode = originalExitCode;
  });

  test('exports immutable metadata', () => {
    expect(Object.isFrozen(buildInfo)).toBe(true);
    expect(Object.isFrozen(buildInfo.typescript)).toBe(true);
    expect(buildInfo.typescript.commit).toMatch(/^[0-9a-f]{40}$/);
  });

  test('JSON query needs neither a native installation nor a valid project config', async () => {
    const output = rstest
      .spyOn(process.stdout, 'write')
      .mockImplementation(() => true);
    await runCLI({
      argv: [
        'node',
        'rslint',
        '--version',
        '--json',
        '--config',
        '/missing/config.ts',
      ],
    });
    expect(output).toHaveBeenCalledExactlyOnceWith(
      `${JSON.stringify(buildInfo)}\n`,
    );
    expect(process.exitCode).toBe(0);
  });

  test('short version flag prints the compiler hash without initializing a project', async () => {
    const output = rstest
      .spyOn(process.stdout, 'write')
      .mockImplementation(() => true);
    await runCLI({ argv: ['node', 'rslint', '-v', '--init'] });
    expect(output).toHaveBeenCalledTimes(1);
    const text = String(output.mock.calls[0][0]);
    expect(text).toContain(`rslint ${buildInfo.version}\n`);
    expect(text).toContain(buildInfo.typescript.commit);
    expect(process.exitCode).toBe(0);
  });

  test.each([
    ['--', '--version'],
    ['--config=--version'],
    ['--rule=--version'],
    ['--version=false'],
  ])(
    'does not mistake positional arguments or option values for a version query: %j',
    (...args) => {
      expect(parseArgs(args).version).toBe(false);
    },
  );
});
