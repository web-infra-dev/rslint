// Fake ESLint plugin for the mixed-native+plugin / gap+plugin e2e.
// Reports every `Identifier` named `forbidden`. Self-contained (no npm
// dependency) so the fixture works in any CI environment.

export default {
  meta: { name: 'fixture-plugin-mixed', version: '1.0.0' },
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
  },
};
