// Plugin Y — bans the identifier `bar`. Disjoint from plugin X
// (no shared module identity, no shared rule name, no shared
// prefix). The pair anchors the worker-side isolation test for
// multi-config plugin routing in a real vscode + LSP session.

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
