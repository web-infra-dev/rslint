// Fake plugin Y — only knows the rule `no-bar`, reports any
// Identifier named `bar`. Paired with `../cfgX/plugin-x.mjs`; the
// two plugins share no module identity, no prefix, no rule name.

export default {
  meta: { name: 'plugin-y', version: '1.0.0' },
  rules: {
    'no-bar': {
      meta: {
        type: 'problem',
        schema: [],
        messages: {
          found: 'plugin-y: identifier `bar` is banned',
        },
      },
      create(ctx) {
        return {
          Identifier(node) {
            if (node.name === 'bar') {
              ctx.report({ node, messageId: 'found' });
            }
          },
        };
      },
    },
  },
};
