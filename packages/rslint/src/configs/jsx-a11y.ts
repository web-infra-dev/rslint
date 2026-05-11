import type { RslintConfigEntry } from '../define-config.js';

// Aligned with official eslint-plugin-jsx-a11y@6.x recommended.
// Rules commented out with "not implemented" are in the official preset but not yet available.
const recommended: RslintConfigEntry = {
  files: ['**/*.jsx', '**/*.tsx'],
  plugins: ['jsx-a11y'],
  rules: {
    'jsx-a11y/alt-text': 'error',
    // 'jsx-a11y/anchor-ambiguous-text': 'off', // not implemented (off in upstream)
    'jsx-a11y/anchor-has-content': 'error',
    'jsx-a11y/anchor-is-valid': 'error',
    // 'jsx-a11y/aria-activedescendant-has-tabindex': 'error', // not implemented
    // 'jsx-a11y/aria-props': 'error', // not implemented
    // 'jsx-a11y/aria-proptypes': 'error', // not implemented
    // 'jsx-a11y/aria-role': 'error', // not implemented
    'jsx-a11y/aria-unsupported-elements': 'error',
    'jsx-a11y/autocomplete-valid': 'error',
    // 'jsx-a11y/click-events-have-key-events': 'error', // not implemented
    // 'jsx-a11y/control-has-associated-label': 'off', // not implemented (off in upstream)
    'jsx-a11y/heading-has-content': 'error',
    'jsx-a11y/html-has-lang': 'error',
    'jsx-a11y/iframe-has-title': 'error',
    'jsx-a11y/img-redundant-alt': 'error',
    // 'jsx-a11y/interactive-supports-focus': 'error', // not implemented
    // 'jsx-a11y/label-has-associated-control': 'error', // not implemented
    // 'jsx-a11y/label-has-for': 'off', // not implemented (off in upstream)
    'jsx-a11y/media-has-caption': 'error',
    // 'jsx-a11y/mouse-events-have-key-events': 'error', // not implemented
    'jsx-a11y/no-access-key': 'error',
    'jsx-a11y/no-autofocus': 'error',
    'jsx-a11y/no-distracting-elements': 'error',
    // 'jsx-a11y/no-interactive-element-to-noninteractive-role': 'error', // not implemented
    // 'jsx-a11y/no-noninteractive-element-interactions': 'error', // not implemented
    // 'jsx-a11y/no-noninteractive-element-to-interactive-role': 'error', // not implemented
    'jsx-a11y/no-noninteractive-tabindex': [
      'error',
      {
        tags: [],
        roles: ['tabpanel'],
        allowExpressionValues: true,
      },
    ],
    'jsx-a11y/no-redundant-roles': 'error',
    // 'jsx-a11y/no-static-element-interactions': 'error', // not implemented
    // 'jsx-a11y/role-has-required-aria-props': 'error', // not implemented
    // 'jsx-a11y/role-supports-aria-props': 'error', // not implemented
    'jsx-a11y/scope': 'error',
    'jsx-a11y/tabindex-no-positive': 'error',
  },
};

export { recommended };
