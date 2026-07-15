// Failure-path counterpart to no-close-plugin-exit.mjs. The config contains a
// loadable community plugin (so its worker host initializes) but a missing
// parserOptions.project file (so Go rejects the lint request). Rslint must shut
// the host down in finally before this caught rejection returns control here.
import { Rslint } from '@rslint/core';
import path from 'node:path';

const fixtureDirectory = path.resolve(
  import.meta.dirname,
  '../eslint-plugin/fixtures',
);
const rslint = new Rslint({
  cwd: fixtureDirectory,
  overrideConfigFile: 'local-invalid.config.mjs',
});
try {
  await rslint.lintText('const value = 1;\n', { filePath: 'probe.ts' });
  console.error('expected lintText to reject missing parserOptions.project');
  process.exit(2);
} catch (error) {
  if (!String(error).includes('tsconfig file')) {
    console.error('unexpected lint error: ' + String(error));
    process.exit(3);
  }
}
// Intentionally no rslint.close().
