import type { RslintConfigEntry } from '../define-config.js';

// Aligned with official @stylistic/eslint-plugin@5.x recommended.
// Upstream exposes `recommended` as `customize()` invoked with default
// options. Option values for not-yet-implemented rules are omitted —
// see upstream's customize.ts for the full configuration each rule receives.
// Rules commented out with "not implemented" are in the official preset but
// not yet available.
const recommended: RslintConfigEntry = {
  plugins: ['@stylistic'],
  rules: {
    '@stylistic/array-bracket-spacing': ['error', 'never'],
    // '@stylistic/arrow-parens': 'error', // not implemented
    // '@stylistic/arrow-spacing': 'error', // not implemented
    // '@stylistic/block-spacing': 'error', // not implemented
    // '@stylistic/brace-style': 'error', // not implemented
    // '@stylistic/comma-dangle': 'error', // not implemented
    // '@stylistic/comma-spacing': 'error', // not implemented
    // '@stylistic/comma-style': 'error', // not implemented
    // '@stylistic/computed-property-spacing': 'error', // not implemented
    // '@stylistic/dot-location': 'error', // not implemented
    // '@stylistic/eol-last': 'error', // not implemented
    // '@stylistic/generator-star-spacing': 'error', // not implemented
    // '@stylistic/indent': 'error', // not implemented
    // '@stylistic/indent-binary-ops': 'error', // not implemented
    // '@stylistic/jsx-closing-bracket-location': 'error', // not implemented
    // '@stylistic/jsx-closing-tag-location': 'error', // not implemented
    // '@stylistic/jsx-curly-brace-presence': 'error', // not implemented
    // '@stylistic/jsx-curly-newline': 'error', // not implemented
    // '@stylistic/jsx-curly-spacing': 'error', // not implemented
    // '@stylistic/jsx-equals-spacing': 'error', // not implemented
    // '@stylistic/jsx-first-prop-new-line': 'error', // not implemented
    // '@stylistic/jsx-function-call-newline': 'error', // not implemented
    // '@stylistic/jsx-indent-props': 'error', // not implemented
    // '@stylistic/jsx-max-props-per-line': 'error', // not implemented
    // '@stylistic/jsx-one-expression-per-line': 'error', // not implemented
    // '@stylistic/jsx-quotes': 'error', // not implemented
    // '@stylistic/jsx-tag-spacing': 'error', // not implemented
    // '@stylistic/jsx-wrap-multilines': 'error', // not implemented
    // '@stylistic/key-spacing': 'error', // not implemented
    // '@stylistic/keyword-spacing': 'error', // not implemented
    // '@stylistic/lines-between-class-members': 'error', // not implemented
    // '@stylistic/max-statements-per-line': 'error', // not implemented
    // '@stylistic/member-delimiter-style': 'error', // not implemented
    // '@stylistic/multiline-ternary': 'error', // not implemented
    // '@stylistic/new-parens': 'error', // not implemented
    // '@stylistic/no-extra-parens': 'error', // not implemented
    // '@stylistic/no-floating-decimal': 'error', // not implemented
    // '@stylistic/no-mixed-operators': 'error', // not implemented
    // '@stylistic/no-mixed-spaces-and-tabs': 'error', // not implemented
    // '@stylistic/no-multi-spaces': 'error', // not implemented
    // '@stylistic/no-multiple-empty-lines': 'error', // not implemented
    // '@stylistic/no-tabs': 'error', // not implemented
    // '@stylistic/no-trailing-spaces': 'error', // not implemented
    // '@stylistic/no-whitespace-before-property': 'error', // not implemented
    // '@stylistic/object-curly-spacing': 'error', // not implemented
    // '@stylistic/operator-linebreak': 'error', // not implemented
    // '@stylistic/padded-blocks': 'error', // not implemented
    // '@stylistic/quote-props': 'error', // not implemented
    // '@stylistic/quotes': 'error', // not implemented
    // '@stylistic/rest-spread-spacing': 'error', // not implemented
    // '@stylistic/semi': 'error', // not implemented
    // '@stylistic/semi-spacing': 'error', // not implemented
    // '@stylistic/space-before-blocks': 'error', // not implemented
    // '@stylistic/space-before-function-paren': 'error', // not implemented
    // '@stylistic/space-in-parens': 'error', // not implemented
    // '@stylistic/space-infix-ops': 'error', // not implemented
    // '@stylistic/space-unary-ops': 'error', // not implemented
    // '@stylistic/spaced-comment': 'error', // not implemented
    // '@stylistic/template-curly-spacing': 'error', // not implemented
    // '@stylistic/template-tag-spacing': 'error', // not implemented
    // '@stylistic/type-annotation-spacing': 'error', // not implemented
    // '@stylistic/type-generic-spacing': 'error', // not implemented
    // '@stylistic/type-named-tuple-spacing': 'error', // not implemented
    // '@stylistic/wrap-iife': 'error', // not implemented
    // '@stylistic/yield-star-spacing': 'error', // not implemented
  },
};

export { recommended };
