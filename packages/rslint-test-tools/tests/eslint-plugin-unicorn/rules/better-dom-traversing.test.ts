// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/better-dom-traversing.js
import { RuleTester, type ValidTestCase } from '../rule-tester';

const messages = {
  'first-child': 'Prefer `.firstChild` over `.childNodes[0]`.',
  'first-element-child': 'Prefer `.firstElementChild` over `.children[0]`.',
  'query-selector':
    'Consider `.querySelector()` over positional child traversal.',
  closest: 'Consider `.closest()` over chaining `.parentElement`.',
  'merge-query-selector': 'Consider merging chained `.querySelector()` calls.',
};

const notDomNodeTypes = [
  '[]',
  '[element]',
  '[...elements]',
  '() => {}',
  'class Node {}',
  'function() {}',
  '0',
  '1',
  '0.1',
  '""',
  '"string"',
  '/regex/',
  'null',
  '0n',
  '1n',
  'true',
  'false',
  '{}',
  '`templateLiteral`',
  'undefined',
];

const js = (code: string): ValidTestCase => ({
  code,
  filename: 'src/virtual.js',
});

new RuleTester().run('better-dom-traversing', {} as never, {
  valid: [
    js(''),
    // Writes to collection indexes.
    js('element.childNodes[0] = node;'),
    js('element.children[0] = node;'),
    js('element.childNodes[0] += node;'),
    js('element.childNodes[0] ||= node;'),
    js('element.childNodes[0]++;'),
    js('delete element.childNodes[0];'),
    js('for (element.childNodes[0] of nodes) {}'),
    js('for (element.childNodes[0] in nodes) {}'),
    js('[element.childNodes[0]] = nodes;'),
    js('({node: element.childNodes[0]} = object);'),
    js('element.children[1] = node;'),
    // Preferred APIs.
    js('element.firstChild;'),
    js('element.firstElementChild;'),
    js('element.querySelector("li");'),
    js('element.closest("form");'),
    // Traversal on a bare identifier.
    js('children[0];'),
    js('parentElement.parentElement;'),
    // Non-literal and unsupported indexes.
    js('element.children[index];'),
    js('element.children["0"];'),
    js('element.children[-1];'),
    js('element.children[1.5];'),
    js('element.childNodes[index];'),
    js('element.childNodes[1];'),
    // Computed collection names.
    js('element["children"][0];'),
    js('element[children][0];'),
    js('element["childNodes"][0];'),
    js('element[childNodes][0];'),
    js('element["parentElement"].parentElement;'),
    js('element.parentElement["parentElement"];'),
    // Optional chaining.
    js('element?.children[0];'),
    js('element.children?.[0];'),
    js('element?.parentElement.parentElement;'),
    js('element.parentElement?.parentElement;'),
    js('element.querySelector?.("a").querySelector("b");'),
    js('element.querySelector("a")?.querySelector("b");'),
    js('element.querySelector?.("a").querySelector("b").querySelector("c");'),
    js('element.querySelector("a")?.querySelector("b").querySelector("c");'),
    js('element.querySelector("a").querySelector("b")?.querySelector("c");'),
    js('element.querySelector("a").querySelector("b")?.foo;'),
    js(
      'element.querySelector("a").querySelector("b").foo?.querySelector("c");',
    ),
    js('(element?.querySelector("a")).querySelector("b").querySelector("c");'),
    // Other methods and argument counts.
    js('element.querySelector();'),
    js('element.querySelector("a", root).querySelector("b");'),
    js('element.querySelector("a").querySelector();'),
    js('element.querySelector("a").querySelector("b", root);'),
    js('element.querySelectorAll("a").querySelector("b");'),
    js('element.querySelector("a").querySelectorAll("b");'),
    // Non-static selectors.
    js('element.querySelector(selector).querySelector("b");'),
    js('element.querySelector("a").querySelector(selector);'),
    js('element.querySelector(`${selector}`).querySelector("b");'),
    js('element.querySelector("a").querySelector(`${selector}`);'),
    js('element.querySelector(tag`a`).querySelector("b");'),
    // Component props.children.
    {
      code: 'class Component { render() { return <div>{this.props.children[0]}</div>; } }',
      filename: 'src/virtual.jsx',
    },
    {
      code: 'const Component = props => <div>{props.children[0]}</div>;',
      filename: 'src/virtual.jsx',
    },
    js('const child = props.children[1];'),
    js('const child = component.props.children[0];'),
    ...notDomNodeTypes.flatMap((receiver) =>
      [
        '.children[0]',
        '.childNodes[0]',
        '.parentElement.parentElement',
        '.querySelector("a").querySelector("b")',
      ].map((access) => js('(' + receiver + ')' + access + ';')),
    ),
    // Remaining documentation examples.
    js("element.querySelector(':scope selector');"),
    js("element.closest('selector');"),
    js("element.querySelector(':scope a b');"),
  ],
  // The Go suite checks ranges, suggestion output, and the absence of autofixes.
  invalid: (
    [
      ['element.childNodes[0];', 'first-child'],
      ['(element.childNodes)[0];', 'first-child'],
      ['(element.children)[0];', 'first-element-child'],
      ['element.children[0];', 'first-element-child'],
      ['element.children[1];', 'query-selector'],
      ['element.children[10];', 'query-selector'],
      ['element.children[1].children[2];', 'query-selector'],
      ['element.parentElement.parentElement;', 'closest'],
      ['element.parentElement.parentElement.parentElement;', 'closest'],
      [
        'element.querySelector("a").querySelector("b");',
        'merge-query-selector',
      ],
      [
        'document.querySelector("a").querySelector("b");',
        'merge-query-selector',
      ],
      [
        'document.body.querySelector("a").querySelector("b");',
        'merge-query-selector',
      ],
      [
        "element.querySelector('a').querySelector('b');",
        'merge-query-selector',
      ],
      [
        'element.querySelector(`a`).querySelector(`b`);',
        'merge-query-selector',
      ],
      [
        'element.querySelector("a").querySelector(`b`);',
        'merge-query-selector',
      ],
      [
        'element.querySelector("a > b").querySelector(".c");',
        'merge-query-selector',
      ],
      [
        'element.querySelector(".a, .b").querySelector(".c");',
        'merge-query-selector',
      ],
      [
        'element.querySelector(".a").querySelector(".b, .c");',
        'merge-query-selector',
      ],
      [
        'element.querySelector(":scope a").querySelector("b");',
        'merge-query-selector',
      ],
      [
        'element.querySelector("a").querySelector("b").querySelector("c");',
        'merge-query-selector',
      ],
      [
        'element.querySelector("a").querySelector("b").querySelector(selector);',
        'merge-query-selector',
      ],
      [
        'element.querySelector(selector).querySelector("b").querySelector("c");',
        'merge-query-selector',
      ],
      [
        '(getElement()).querySelector("a").querySelector("b");',
        'merge-query-selector',
      ],
      [
        '(foo || bar).querySelector("a").querySelector("b");',
        'merge-query-selector',
      ],
      ['element.childNodes[/* comment */ 0];', 'first-child'],
      ['element.children[/* comment */ 0];', 'first-element-child'],
      [
        'element.querySelector(/* comment */ "a").querySelector("b");',
        'merge-query-selector',
      ],
      [
        'element.querySelector("a").querySelector(/* comment */ "b");',
        'merge-query-selector',
      ],
      ['const item = element\n\t.children[0];', 'first-element-child'],
      [
        'const item = element\n\t.querySelector("a")\n\t.querySelector("b");',
        'merge-query-selector',
      ],
      ['element.childNodes[0].textContent = text;', 'first-child'],
      [
        'window.document.querySelector("a").querySelector("b");',
        'merge-query-selector',
      ],
      [
        'globalThis.document.querySelector("a").querySelector("b");',
        'merge-query-selector',
      ],
      [
        'window.frame.querySelector("a").querySelector("b");',
        'merge-query-selector',
      ],
      [
        'window[document].querySelector("a").querySelector("b");',
        'merge-query-selector',
      ],
      // Object-literal receiver regressions.
      [
        '({element}.element.querySelector("a")).querySelector("b");',
        'merge-query-selector',
      ],
      [
        'const result = ({element}.element.querySelector("a")).querySelector("b");',
        'merge-query-selector',
      ],
      [
        'const read = () => ({element}.element.querySelector("a")).querySelector("b");',
        'merge-query-selector',
      ],
      [
        '({element}.element).querySelector("a").querySelector("b");',
        'merge-query-selector',
      ],
      // Additional documentation examples.
      ['element.children[2];', 'query-selector'],
    ] as const
  ).map(([code, messageId]) => ({
    ...js(code),
    errors: [{ messageId, message: messages[messageId] }],
  })),
});
