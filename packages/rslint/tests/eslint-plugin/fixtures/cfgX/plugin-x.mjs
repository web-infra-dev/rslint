// Fake plugin X — only knows the rule `no-foo`, reports any
// Identifier named `foo`. Paired with `../cfgY/plugin-y.mjs` (a
// completely different module that only knows `no-bar`) to verify
// that worker-side per-config LoadedPlugins maps stay disjoint when
// the two configs use BOTH different prefixes AND different plugin
// instances — i.e. there is no shared rule registry.

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
