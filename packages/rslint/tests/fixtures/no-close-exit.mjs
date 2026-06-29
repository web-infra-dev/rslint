// Regression fixture for the unref() fix: a script that lints with the Rslint
// class and DELIBERATELY never calls close(). The resident Go `--api` child and
// its stdio pipes are unref'd, so once this script's work is done the Node
// process must exit on its own — if unref regresses, the open child/pipes keep
// the event loop alive and this process hangs (the test spawns it with a
// timeout and asserts a clean exit).
import { Rslint } from '@rslint/core';

const rslint = new Rslint({
  overrideConfigFile: true,
  overrideConfig: [
    {
      files: ['**/*.ts'],
      plugins: ['@typescript-eslint'],
      rules: { '@typescript-eslint/no-explicit-any': 'error' },
    },
  ],
});
const results = await rslint.lintText('const x: any = 1;\n', {
  filePath: 'a.ts',
});
// Sanity-check the lint actually ran (so a 0-message config bug can't make the
// "no hang" assertion pass for the wrong reason); a mismatch exits non-zero.
if (results.length !== 1 || results[0].messages.length !== 1) {
  console.error(
    'unexpected lint result: ' + JSON.stringify(results?.[0]?.messages ?? null),
  );
  process.exit(2);
}
// Intentionally NO rslint.close(): unref() must let the process exit anyway.
