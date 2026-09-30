// Companion rules from eslint-plugin-unicorn v76.0.0 test/no-lonely-if.js.
const pragmaAttachmentRule = {
  meta: {},
  create(context) {
    const { sourceCode } = context;

    return {
      IfStatement(node) {
        if (
          node.test.type !== 'LogicalExpression' ||
          node.test.operator !== '&&'
        ) {
          return;
        }

        const [nodeStart] = sourceCode.getRange(node);
        const commentsBefore = sourceCode.getCommentsBefore(node);
        const attachedComment = commentsBefore.at(-1);

        if (!attachedComment) {
          context.report({
            node,
            message: 'Merged `if` is missing attached pragma comment.',
          });
          return;
        }

        const commentGap = sourceCode.text.slice(
          sourceCode.getRange(attachedComment)[1],
          nodeStart,
        );
        if (
          commentGap.trim() !== '' ||
          !attachedComment.value.includes('@keep-next')
        ) {
          context.report({
            node,
            message: 'Merged `if` is missing attached pragma comment.',
          });
        }
      },
    };
  },
};

const fakeFormattingRule = {
  meta: {
    fixable: 'code',
  },
  create: (context) => ({
    Program(node) {
      context.report({
        node,
        message: 'format',
        fix(fixer) {
          const source = context.sourceCode.text;
          const formatted = source.includes('value < 10 && value < 5')
            ? "function some(value) {\n  if (value < 10 && value < 5) {\n    console.log(\n      'this is a long string this is a long string this is a long string this is a long string',\n      value,\n    );\n  }\n  return 0;\n}"
            : "function some(value) {\n  if (value < 10) {\n    if (value < 5) {\n      console.log(\n        'this is a long string this is a long string this is a long string this is a long string',\n        value,\n      );\n    }\n  }\n  return 0;\n}";

          if (formatted === source) {
            return;
          }

          return fixer.replaceTextRange([0, source.length], formatted);
        },
      });
    },
  }),
};

export default [
  {
    plugins: ['unicorn'],
    languageOptions: { ecmaVersion: 'latest', sourceType: 'module' },
    rules: { 'unicorn/no-lonely-if': 'error' },
  },
  {
    plugins: {
      fake: {
        rules: {
          'pragma-attachment': pragmaAttachmentRule,
          format: fakeFormattingRule,
        },
      },
    },
  },
];
