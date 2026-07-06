export { defineConfig, globalIgnores } from './config/define-config.js';
export type {
  RslintConfig,
  RslintConfigEntry,
  ESLintPlugin,
} from './config/define-config.js';
export {
  ts,
  js,
  reactPlugin,
  reactHooksPlugin,
  importPlugin,
  promisePlugin,
  jestPlugin,
  unicornPlugin,
  jsxA11yPlugin,
} from './config/presets/index.js';

// The ESLint v10-aligned programmatic Node.js API (issue #1106). This `Rslint`
// class and the `runCLI` wrapper are the only linting surfaces exported from
// the package root; alongside them the root exports just the config-authoring
// helpers (`defineConfig` / `globalIgnores`) and the plugin presets. The
// low-level engine (the `lint` convenience, `RSLintService`, and the Node
// backend) lives on internal subpaths — `@rslint/core/internal` and
// `@rslint/core/service` — not on the public root. (The browser/web-worker
// backend lives in `@rslint/wasm`.)
export { Rslint } from './api/rslint.js';
export type {
  RslintOptions,
  LintResult,
  LintMessage,
  LintSuggestion,
  LintMessageFix,
} from './api/rslint.js';
export { runCLI } from './cli/cli.js';
export type { RunCLIOptions } from './cli/cli.js';
