// Plugin X — bans the identifier `foo`. Paired with plugin Y under
// packages/y/. Two disjoint plugins under disjoint prefixes verify
// that the LSP-side WorkerPool keeps each nested config's plugin set
// isolated when both configs are live in the same vscode session.

export default {
  meta: { name: 'plugin-x', version: '1.0.0' },
  rules: {
    'no-foo': {
      meta: {
        type: 'problem',
        schema: [],
        messages: {
          found: 'plugin-x: identifier `foo` is banned',
        },
      },
      create(ctx) {
        return {
          Identifier(node) {
            if (node.name === 'foo') {
              ctx.report({ node, messageId: 'found' });
            }
          },
        };
      },
    },
  },
};
