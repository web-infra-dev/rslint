// A minimal self-contained ESLint plugin, mounted via the config's
// `eslintPlugins`. Kept dependency-free so the fixture needs no node_modules:
//   - no-null            reports every `null` literal (suggestion only)
//   - prefer-array-some  reports `.filter` member access, with an auto fix
//                        (filter -> some) so it participates in source.fixAll
export default {
  meta: { name: 'local-plugin', version: '1.0.0' },
  rules: {
    'no-null': {
      meta: {
        type: 'suggestion',
        hasSuggestions: true,
        schema: [],
        messages: {
          error: 'Do not use `null`; prefer `undefined`.',
          replaceWithUndefined: 'Replace `null` with `undefined`.',
        },
      },
      create(context) {
        return {
          Literal(node) {
            if (node.raw !== 'null') return;
            context.report({
              node,
              messageId: 'error',
              suggest: [
                {
                  messageId: 'replaceWithUndefined',
                  fix: (fixer) => fixer.replaceText(node, 'undefined'),
                },
              ],
            });
          },
        };
      },
    },
    'prefer-array-some': {
      meta: {
        type: 'suggestion',
        fixable: 'code',
        schema: [],
        messages: { preferSome: 'Prefer `.some(…)` over `.filter(…)`.' },
      },
      create(context) {
        return {
          MemberExpression(node) {
            if (
              node.property &&
              node.property.type === 'Identifier' &&
              node.property.name === 'filter'
            ) {
              context.report({
                node: node.property,
                messageId: 'preferSome',
                fix: (fixer) => fixer.replaceText(node.property, 'some'),
              });
            }
          },
        };
      },
    },
  },
};
