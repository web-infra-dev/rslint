// Fake ESLint plugin for the single-config vscode e2e smoke test.
// Reports every `Identifier` named `forbidden`. Self-contained (no
// npm dependency) so the fixture works in any CI environment.

export default {
  meta: { name: 'fixture-plugin', version: '1.0.0' },
  rules: {
    'no-forbidden': {
      meta: {
        type: 'problem',
        schema: [],
        messages: {
          found: '`forbidden` is banned by the fixture plugin',
        },
      },
      create(ctx) {
        return {
          Identifier(node) {
            if (node.name === 'forbidden') {
              ctx.report({ node, messageId: 'found' });
            }
          },
        };
      },
    },
    // Used by suite-eslint-plugin's U10 fixAll test. Rewrites every
    // `BANNED` identifier to `ALLOWED`. Kept separate from no-forbidden
    // so the U9 didChange test (which counts no-forbidden diagnostics)
    // isn't affected.
    'rename-banned': {
      meta: {
        type: 'problem',
        fixable: 'code',
        schema: [],
        messages: {
          banned: '`BANNED` must be renamed to `ALLOWED`',
        },
      },
      create(ctx) {
        return {
          Identifier(node) {
            if (node.name === 'BANNED') {
              ctx.report({
                node,
                messageId: 'banned',
                fix(fixer) {
                  return fixer.replaceText(node, 'ALLOWED');
                },
              });
            }
          },
        };
      },
    },
  },
};
