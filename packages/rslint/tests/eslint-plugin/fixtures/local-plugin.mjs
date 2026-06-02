// A self-contained local plugin used to exercise the runner's full
// plugin-load → lint → options → fix/suggest pipeline WITHOUT an
// external dependency.
//
// This stands in for a real ESLint plugin in the e2e / plugin-flow
// fixtures. The REAL-plugin compatibility e2e (running unmodified
// community plugins end-to-end) lives on the eslint-plugin-compat
// branch; this package only needs a fixture that round-trips the same
// runner code paths: plugin loading, per-config prefixing, eager
// option-default materialization, autofixes, and suggestions.
//
// Rules:
//   no-null            — reports `null` literals. Ships
//                        `defaultOptions:[{checkStrictEquality:false}]`
//                        and eagerly destructures `context.options[0]`,
//                        so it only passes when the runner materializes
//                        that default. Under the default it skips the
//                        `x === null` / `x !== null` strict-equality
//                        forms, and offers a suggestion rewriting
//                        `null` → `undefined`.
//   prefer-array-some  — autofix rule: rewrites a `.filter` member name
//                        to `some`.

export default {
  meta: { name: 'local-plugin', version: '1.0.0' },
  rules: {
    'no-null': {
      meta: {
        type: 'suggestion',
        hasSuggestions: true,
        schema: [
          {
            type: 'object',
            properties: { checkStrictEquality: { type: 'boolean' } },
            additionalProperties: false,
          },
        ],
        defaultOptions: [{ checkStrictEquality: false }],
        messages: {
          error: 'Do not use `null`; prefer `undefined`.',
          replaceWithUndefined: 'Replace `null` with `undefined`.',
        },
      },
      create(context) {
        // Eager destructure — only valid if the runner has materialized
        // `options[0]` from `defaultOptions` for the no-options config.
        const { checkStrictEquality } = context.options[0];
        return {
          Literal(node) {
            if (node.raw !== 'null') return;
            const parent = node.parent;
            if (
              !checkStrictEquality &&
              parent &&
              parent.type === 'BinaryExpression' &&
              (parent.operator === '===' || parent.operator === '!==')
            ) {
              return;
            }
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
        messages: {
          preferSome: 'Prefer `.some(…)` over `.filter(…).length`.',
        },
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
