import type { RslintConfigEntry } from '../define-config.js';

// Aligned with official eslint-plugin-unicorn@64.x recommended.
// Rules commented out with "not implemented" are in the official preset but not yet available.
// The official preset also injects `languageOptions.globals` (Array, Promise, Map, …);
// rslint's config entry doesn't expose a `globals` field, so that part is omitted.
const recommended: RslintConfigEntry = {
  plugins: ['unicorn'],
  rules: {
    // Core ESLint rules disabled by upstream once their unicorn equivalents take over.
    // Keep them enabled (i.e. don't override) until the unicorn replacements are implemented,
    // otherwise users would lose all coverage for these patterns.
    // 'no-negated-condition': 'off',
    // 'no-nested-ternary': 'off',

    // 'unicorn/better-regex': 'off', // not implemented
    // 'unicorn/catch-error-name': 'error', // not implemented
    // 'unicorn/consistent-assert': 'error', // not implemented
    // 'unicorn/consistent-date-clone': 'error', // not implemented
    // 'unicorn/consistent-destructuring': 'off', // not implemented
    // 'unicorn/consistent-empty-array-spread': 'error', // not implemented
    // 'unicorn/consistent-existence-index-check': 'error', // not implemented
    // 'unicorn/consistent-function-scoping': 'error', // not implemented
    // 'unicorn/consistent-template-literal-escape': 'error', // not implemented
    // 'unicorn/custom-error-definition': 'off', // not implemented
    // 'unicorn/empty-brace-spaces': 'error', // not implemented
    // 'unicorn/error-message': 'error', // not implemented
    // 'unicorn/escape-case': 'error', // not implemented
    // 'unicorn/expiring-todo-comments': 'error', // not implemented
    // 'unicorn/explicit-length-check': 'error', // not implemented
    'unicorn/filename-case': 'error',
    // 'unicorn/import-style': 'error', // not implemented
    // 'unicorn/isolated-functions': 'error', // not implemented
    // 'unicorn/new-for-builtins': 'error', // not implemented
    // 'unicorn/no-abusive-eslint-disable': 'error', // not implemented
    // 'unicorn/no-accessor-recursion': 'error', // not implemented
    // 'unicorn/no-anonymous-default-export': 'error', // not implemented
    // 'unicorn/no-array-callback-reference': 'error', // not implemented
    // 'unicorn/no-array-for-each': 'error', // not implemented
    // 'unicorn/no-array-method-this-argument': 'error', // not implemented
    // 'unicorn/no-array-reduce': 'error', // not implemented
    // 'unicorn/no-array-reverse': 'error', // not implemented
    // 'unicorn/no-array-sort': 'error', // not implemented
    // 'unicorn/no-await-expression-member': 'error', // not implemented
    // 'unicorn/no-await-in-promise-methods': 'error', // not implemented
    // 'unicorn/no-console-spaces': 'error', // not implemented
    // 'unicorn/no-document-cookie': 'error', // not implemented
    // 'unicorn/no-empty-file': 'error', // not implemented
    // 'unicorn/no-for-loop': 'error', // not implemented
    // 'unicorn/no-hex-escape': 'error', // not implemented
    // 'unicorn/no-immediate-mutation': 'error', // not implemented
    // 'unicorn/no-instanceof-builtins': 'error', // not implemented
    // 'unicorn/no-invalid-fetch-options': 'error', // not implemented
    // 'unicorn/no-invalid-remove-event-listener': 'error', // not implemented
    // 'unicorn/no-keyword-prefix': 'off', // not implemented
    // 'unicorn/no-lonely-if': 'error', // not implemented
    // 'unicorn/no-magic-array-flat-depth': 'error', // not implemented
    // 'unicorn/no-named-default': 'error', // not implemented
    // 'unicorn/no-negated-condition': 'error', // not implemented
    // 'unicorn/no-negation-in-equality-check': 'error', // not implemented
    // 'unicorn/no-nested-ternary': 'error', // not implemented
    // 'unicorn/no-new-array': 'error', // not implemented
    // 'unicorn/no-new-buffer': 'error', // not implemented
    // 'unicorn/no-null': 'error', // not implemented
    // 'unicorn/no-object-as-default-parameter': 'error', // not implemented
    // 'unicorn/no-process-exit': 'error', // not implemented
    // 'unicorn/no-single-promise-in-promise-methods': 'error', // not implemented
    'unicorn/no-static-only-class': 'error',
    // 'unicorn/no-thenable': 'error', // not implemented
    // 'unicorn/no-this-assignment': 'error', // not implemented
    // 'unicorn/no-typeof-undefined': 'error', // not implemented
    // 'unicorn/no-unnecessary-array-flat-depth': 'error', // not implemented
    // 'unicorn/no-unnecessary-array-splice-count': 'error', // not implemented
    // 'unicorn/no-unnecessary-await': 'error', // not implemented
    // 'unicorn/no-unnecessary-polyfills': 'error', // not implemented
    // 'unicorn/no-unnecessary-slice-end': 'error', // not implemented
    // 'unicorn/no-unreadable-array-destructuring': 'error', // not implemented
    // 'unicorn/no-unreadable-iife': 'error', // not implemented
    // 'unicorn/no-unused-properties': 'off', // not implemented
    // 'unicorn/no-useless-collection-argument': 'error', // not implemented
    // 'unicorn/no-useless-error-capture-stack-trace': 'error', // not implemented
    // 'unicorn/no-useless-fallback-in-spread': 'error', // not implemented
    // 'unicorn/no-useless-iterator-to-array': 'error', // not implemented
    // 'unicorn/no-useless-length-check': 'error', // not implemented
    // 'unicorn/no-useless-promise-resolve-reject': 'error', // not implemented
    // 'unicorn/no-useless-spread': 'error', // not implemented
    // 'unicorn/no-useless-switch-case': 'error', // not implemented
    // 'unicorn/no-useless-undefined': 'error', // not implemented
    // 'unicorn/no-zero-fractions': 'error', // not implemented
    // 'unicorn/number-literal-case': 'error', // not implemented
    // 'unicorn/numeric-separators-style': 'error', // not implemented
    // 'unicorn/prefer-add-event-listener': 'error', // not implemented
    // 'unicorn/prefer-array-find': 'error', // not implemented
    // 'unicorn/prefer-array-flat': 'error', // not implemented
    // 'unicorn/prefer-array-flat-map': 'error', // not implemented
    // 'unicorn/prefer-array-index-of': 'error', // not implemented
    // 'unicorn/prefer-array-some': 'error', // not implemented
    // 'unicorn/prefer-at': 'error', // not implemented
    // 'unicorn/prefer-bigint-literals': 'error', // not implemented
    // 'unicorn/prefer-blob-reading-methods': 'error', // not implemented
    // 'unicorn/prefer-class-fields': 'error', // not implemented
    // 'unicorn/prefer-classlist-toggle': 'error', // not implemented
    // 'unicorn/prefer-code-point': 'error', // not implemented
    // 'unicorn/prefer-date-now': 'error', // not implemented
    // 'unicorn/prefer-default-parameters': 'error', // not implemented
    // 'unicorn/prefer-dom-node-append': 'error', // not implemented
    // 'unicorn/prefer-dom-node-dataset': 'error', // not implemented
    // 'unicorn/prefer-dom-node-remove': 'error', // not implemented
    // 'unicorn/prefer-dom-node-text-content': 'error', // not implemented
    // 'unicorn/prefer-event-target': 'error', // not implemented
    // 'unicorn/prefer-export-from': 'error', // not implemented
    // 'unicorn/prefer-global-this': 'error', // not implemented
    // 'unicorn/prefer-import-meta-properties': 'off', // not implemented
    // 'unicorn/prefer-includes': 'error', // not implemented
    // 'unicorn/prefer-json-parse-buffer': 'off', // not implemented
    // 'unicorn/prefer-keyboard-event-key': 'error', // not implemented
    // 'unicorn/prefer-logical-operator-over-ternary': 'error', // not implemented
    // 'unicorn/prefer-math-min-max': 'error', // not implemented
    // 'unicorn/prefer-math-trunc': 'error', // not implemented
    // 'unicorn/prefer-modern-dom-apis': 'error', // not implemented
    // 'unicorn/prefer-modern-math-apis': 'error', // not implemented
    // 'unicorn/prefer-module': 'error', // not implemented
    // 'unicorn/prefer-native-coercion-functions': 'error', // not implemented
    // 'unicorn/prefer-negative-index': 'error', // not implemented
    // 'unicorn/prefer-node-protocol': 'error', // not implemented
    // 'unicorn/prefer-number-properties': 'error', // not implemented
    // 'unicorn/prefer-object-from-entries': 'error', // not implemented
    // 'unicorn/prefer-optional-catch-binding': 'error', // not implemented
    // 'unicorn/prefer-prototype-methods': 'error', // not implemented
    // 'unicorn/prefer-query-selector': 'error', // not implemented
    // 'unicorn/prefer-reflect-apply': 'error', // not implemented
    // 'unicorn/prefer-regexp-test': 'error', // not implemented
    // 'unicorn/prefer-response-static-json': 'error', // not implemented
    // 'unicorn/prefer-set-has': 'error', // not implemented
    // 'unicorn/prefer-set-size': 'error', // not implemented
    // 'unicorn/prefer-simple-condition-first': 'error', // not implemented
    // 'unicorn/prefer-single-call': 'error', // not implemented
    // 'unicorn/prefer-spread': 'error', // not implemented
    // 'unicorn/prefer-string-raw': 'error', // not implemented
    // 'unicorn/prefer-string-replace-all': 'error', // not implemented
    // 'unicorn/prefer-string-slice': 'error', // not implemented
    // 'unicorn/prefer-string-starts-ends-with': 'error', // not implemented
    // 'unicorn/prefer-string-trim-start-end': 'error', // not implemented
    // 'unicorn/prefer-structured-clone': 'error', // not implemented
    // 'unicorn/prefer-switch': 'error', // not implemented
    // 'unicorn/prefer-ternary': 'error', // not implemented
    // 'unicorn/prefer-top-level-await': 'error', // not implemented
    // 'unicorn/prefer-type-error': 'error', // not implemented
    // 'unicorn/prevent-abbreviations': 'error', // not implemented
    // 'unicorn/relative-url-style': 'error', // not implemented
    // 'unicorn/require-array-join-separator': 'error', // not implemented
    // 'unicorn/require-module-attributes': 'error', // not implemented
    // 'unicorn/require-module-specifiers': 'error', // not implemented
    // 'unicorn/require-number-to-fixed-digits-argument': 'error', // not implemented
    // 'unicorn/require-post-message-target-origin': 'off', // not implemented
    // 'unicorn/string-content': 'off', // not implemented
    // 'unicorn/switch-case-braces': 'error', // not implemented
    // 'unicorn/switch-case-break-position': 'error', // not implemented
    // 'unicorn/template-indent': 'error', // not implemented
    // 'unicorn/text-encoding-identifier-case': 'error', // not implemented
    // 'unicorn/throw-new-error': 'error', // not implemented
  },
};

export { recommended };
