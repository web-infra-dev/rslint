import { readFileSync } from 'node:fs';
import path from 'node:path';
import type { GlobalsConfig, RslintConfigEntry } from '../define-config.js';
import { globals } from '../globals/index.js';

// Aligned with eslint-plugin-n@18.3.0, using rslint's native node/* names.
// https://github.com/eslint-community/eslint-plugin-n/tree/v18.3.0/lib/configs
const recommendedRules: RslintConfigEntry['rules'] = {
  'node/hashbang': 'error',
  'node/no-deprecated-api': 'error',
  'node/no-exports-assign': 'error',
  'node/no-extraneous-import': 'error',
  'node/no-extraneous-require': 'error',
  'node/no-missing-import': 'error',
  'node/no-missing-require': 'error',
  'node/no-process-exit': 'error',
  'node/no-unpublished-import': 'error',
  'node/no-unpublished-require': 'error',
  'node/no-unsupported-features/es-builtins': 'error',
  'node/no-unsupported-features/node-builtins': 'error',
  // 'node/process-exit-as-throw': 'error', // not implemented
};

const recommendedModule: RslintConfigEntry = {
  plugins: ['node'],
  languageOptions: {
    sourceType: 'module',
    // Preserve lazy globals loading when users only import @rslint/core.
    get globals(): GlobalsConfig {
      return {
        ...globals.node,
        ...globals.es2021,
        __dirname: 'off',
        __filename: 'off',
        exports: 'off',
        module: 'off',
        require: 'off',
      };
    },
  },
  rules: {
    ...recommendedRules,
    'node/no-unsupported-features/es-syntax': [
      'error',
      { ignores: ['modules'] },
    ],
  },
};

const recommendedScript: RslintConfigEntry = {
  plugins: ['node'],
  languageOptions: {
    // Flat config represents Node's function scope/globalReturn with commonjs.
    sourceType: 'commonjs',
    get globals(): GlobalsConfig {
      return {
        ...globals.node,
        ...globals.es2021,
        __dirname: 'readonly',
        __filename: 'readonly',
        exports: 'writable',
        module: 'readonly',
        require: 'readonly',
      };
    },
  },
  rules: {
    ...recommendedRules,
    'node/no-unsupported-features/es-syntax': ['error', { ignores: [] }],
  },
};

function usesModules(): boolean {
  // Like upstream, use the nearest valid package.json from the process cwd.
  // Delay this lookup until recommended is selected, rather than root import.
  let directory = process.cwd();
  for (;;) {
    try {
      const pkg: unknown = JSON.parse(
        readFileSync(path.join(directory, 'package.json'), 'utf8'),
      );
      if (pkg !== null && typeof pkg === 'object' && !Array.isArray(pkg)) {
        return 'type' in pkg && pkg.type === 'module';
      }
    } catch {
      // Upstream skips unreadable or malformed package files.
    }
    const parent = path.dirname(directory);
    if (parent === directory) return false;
    directory = parent;
  }
}

function createRecommended(): RslintConfigEntry[] {
  // Translate upstream recommended's legacy overrides into flat entries.
  return [
    usesModules() ? recommendedModule : recommendedScript,
    { ...recommendedScript, files: ['**/*.cjs'] },
    { ...recommendedModule, files: ['**/*.mjs'] },
  ];
}

export { createRecommended, recommendedModule, recommendedScript };
