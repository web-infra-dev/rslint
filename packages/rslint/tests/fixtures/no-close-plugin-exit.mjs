// The per-call community-plugin host must be shut down before lintText returns.
// This script intentionally leaves the Rslint instance open: the Go child is
// unref'd, so a leaked plugin worker is the only thing that can keep Node alive.
import { Rslint } from '@rslint/core';
import path from 'node:path';

const fixtureDirectory = path.resolve(
  import.meta.dirname,
  '../eslint-plugin/fixtures',
);
const rslint = new Rslint({
  cwd: fixtureDirectory,
  overrideConfigFile: 'local.config.mjs',
});
const [result] = await rslint.lintText('const value = 1;\n', {
  filePath: 'probe.ts',
});
if (
  result?.messages.length !== 1 ||
  result.messages[0].ruleId !== 'local/program-listener'
) {
  console.error('unexpected plugin result: ' + JSON.stringify(result ?? null));
  process.exit(2);
}
// Intentionally no rslint.close().
