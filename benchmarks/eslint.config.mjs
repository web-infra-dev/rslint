import tseslint from 'typescript-eslint'
import path from 'node:path'
import process from 'node:process'
import { fileURLToPath } from 'node:url'

const { TSGOLINT_BENCHMARK_PROJECT } = process.env

let files = []
let project = ''

if (TSGOLINT_BENCHMARK_PROJECT === 'vscode') {
  files = ['src/**/*.ts']
  project = './src/tsconfig.json'
} else if (TSGOLINT_BENCHMARK_PROJECT === 'typescript') {
  files = ['src/**/*.ts']
  project = './src/tsconfig-eslint.json'
} else if (TSGOLINT_BENCHMARK_PROJECT === 'typeorm') {
  files = ['src/**/*.ts', 'sample/**/*.ts', 'test/**/*.ts']
  project = './tsconfig.json'
}


export default tseslint.config(
  {
    ignores: ['**/*.js'],
    linterOptions: {
      noInlineConfig: true,
      reportUnusedInlineConfigs: 'off',
      reportUnusedDisableDirectives: 'off',
    },
  },
  {
    files,
    languageOptions: {
      parser: tseslint.parser,
      parserOptions: {
        project,
        tsconfigRootDir: path.dirname(fileURLToPath(import.meta.url)),
      },
    },
    plugins: {
      '@typescript-eslint': tseslint.plugin,
    },
    rules: {
      '@typescript-eslint/await-thenable': 'error',
      '@typescript-eslint/no-array-delete': 'error',
      '@typescript-eslint/no-base-to-string': 'error',
      '@typescript-eslint/no-duplicate-type-constituents': 'error',
      '@typescript-eslint/no-floating-promises': 'error',
      '@typescript-eslint/no-for-in-array': 'error',
      '@typescript-eslint/no-implied-eval': 'error',
      '@typescript-eslint/no-misused-promises': 'error',
      '@typescript-eslint/no-redundant-type-constituents': 'error',
      '@typescript-eslint/no-unnecessary-type-assertion': 'error',
      '@typescript-eslint/no-unsafe-argument': 'error',
      '@typescript-eslint/no-unsafe-assignment': 'error',
      '@typescript-eslint/no-unsafe-call': 'error',
      '@typescript-eslint/no-unsafe-enum-comparison': 'error',
      '@typescript-eslint/no-unsafe-member-access': 'error',
      '@typescript-eslint/no-unsafe-return': 'error',
      '@typescript-eslint/no-unsafe-unary-minus': 'error',
      '@typescript-eslint/only-throw-error': 'error',
      '@typescript-eslint/prefer-promise-reject-errors': 'error',
      '@typescript-eslint/require-await': 'error',
      '@typescript-eslint/restrict-plus-operands': 'error',
      '@typescript-eslint/restrict-template-expressions': 'error',
      '@typescript-eslint/switch-exhaustiveness-check': 'error',
      '@typescript-eslint/unbound-method': 'error',
    },
  },
)
