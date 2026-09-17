// Ported from eslint-plugin-unicorn v75.0.0 tests and documentation; see LICENSE.
import {
  RuleTester,
  type ValidTestCase,
  type InvalidTestCase,
} from '../rule-tester';

new RuleTester().run('prefer-dom-node-append', {} as never, {
  valid: [
    ...[
      'parent.append(child);',
      'new parent.appendChild(child);',
      'appendChild(child);',
      "parent['appendChild'](child);",
      'parent[appendChild](child);',
      'parent.foo(child);',
      'parent.appendChild(one, two);',
      'parent.appendChild();',
      'parent.appendChild(...argumentsArray)',
      'parent.appendChild?.(child)',
      '([]).appendChild(foo)',
      '([element]).appendChild(foo)',
      '([...elements]).appendChild(foo)',
      '(() => {}).appendChild(foo)',
      '(class Node {}).appendChild(foo)',
      '(function() {}).appendChild(foo)',
      '(0).appendChild(foo)',
      '(1).appendChild(foo)',
      '(0.1).appendChild(foo)',
      '("").appendChild(foo)',
      '("string").appendChild(foo)',
      '(/regex/).appendChild(foo)',
      '(null).appendChild(foo)',
      '(0n).appendChild(foo)',
      '(1n).appendChild(foo)',
      '(true).appendChild(foo)',
      '(false).appendChild(foo)',
      '({}).appendChild(foo)',
      '(`templateLiteral`).appendChild(foo)',
      '(undefined).appendChild(foo)',
      'foo.appendChild([])',
      'foo.appendChild([element])',
      'foo.appendChild([...elements])',
      'foo.appendChild(() => {})',
      'foo.appendChild(class Node {})',
      'foo.appendChild(function() {})',
      'foo.appendChild(0)',
      'foo.appendChild(1)',
      'foo.appendChild(0.1)',
      'foo.appendChild("")',
      'foo.appendChild("string")',
      'foo.appendChild(/regex/)',
      'foo.appendChild(null)',
      'foo.appendChild(0n)',
      'foo.appendChild(1n)',
      'foo.appendChild(true)',
      'foo.appendChild(false)',
      'foo.appendChild({})',
      'foo.appendChild(`templateLiteral`)',
      'foo.appendChild(undefined)',
      '// ✅\nelement.append(child);\n',
      '// ✅\n// append() can handle multiple nodes in one call\nparent.append(child1, child2, child3);\n',
      "// ✅\n// append() can mix nodes and strings\ncontainer.append(divElement, 'Some text content');\n",
    ].map((code): ValidTestCase => ({
      code,
      ...{
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
  ],
  invalid: [
    ...[
      'node.appendChild(child);',
      'document.body.appendChild(child);',
      'node.appendChild(foo)',
      'function foo() {\n\tnode.appendChild(bar);\n}',
      'const foo = node.appendChild(child);',
      'console.log(node.appendChild(child));',
      'node.appendChild(child) || "foo";',
      'node.appendChild(child) + 0;',
      '+node.appendChild(child);',
      'node.appendChild(child) ? "foo" : "bar";',
      'if (node.appendChild(child)) {}',
      'const foo = [node.appendChild(child)]',
      'const foo = { bar: node.appendChild(child) }',
      'function foo() { return node.appendChild(child); }',
      'const foo = () => { return node.appendChild(child); }',
      'foo(bar = node.appendChild(child))',
      'node?.appendChild(child);',
      '() => node?.appendChild(child)',
      '// ❌\nelement.appendChild(child);\n\n',
      '// ❌\n// appendChild only works with Node objects\ncontainer.appendChild(divElement);\n\n',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message: 'Prefer `Element#append()` over `Node#appendChild()`.',
            messageId: 'prefer-dom-node-append',
          },
        ],
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
    ...['node.appendChild(child).appendChild(grandchild);'].map(
      (code): InvalidTestCase => ({
        code,
        ...{
          errors: [
            {
              message: 'Prefer `Element#append()` over `Node#appendChild()`.',
              messageId: 'prefer-dom-node-append',
            },
            {
              message: 'Prefer `Element#append()` over `Node#appendChild()`.',
              messageId: 'prefer-dom-node-append',
            },
          ],
          filename: 'src/virtual.js',
          languageOptions: {
            globals: {
              global: 'readonly',
              self: 'readonly',
              window: 'readonly',
            },
            sourceType: 'module',
          },
        },
      }),
    ),
    ...[
      '// ❌\n// Multiple nodes require chaining\nparent.appendChild(child1);\nparent.appendChild(child2);\nparent.appendChild(child3);\n\n',
    ].map((code): InvalidTestCase => ({
      code,
      ...{
        errors: [
          {
            message: 'Prefer `Element#append()` over `Node#appendChild()`.',
            messageId: 'prefer-dom-node-append',
          },
          {
            message: 'Prefer `Element#append()` over `Node#appendChild()`.',
            messageId: 'prefer-dom-node-append',
          },
          {
            message: 'Prefer `Element#append()` over `Node#appendChild()`.',
            messageId: 'prefer-dom-node-append',
          },
        ],
        filename: 'src/virtual.js',
        languageOptions: {
          globals: { global: 'readonly', self: 'readonly', window: 'readonly' },
          sourceType: 'module',
        },
      },
    })),
  ],
});
