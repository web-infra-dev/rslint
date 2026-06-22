/**
 * Conformance: @stylistic/eslint-plugin (jsx-a) mounted in rslint via `plugins`
 * must report identically to ESLint v10. Representative triggers from the
 * upstream suite; each verified to reproduce ESLint v10 byte-for-byte.
 *
 * Four rules here keep only CLEAN (valid) cases — their invalid triggers report
 * at JSX-node positions rslint and ESLint v10 compute differently (start loc, or
 * a null endLine/endColumn for a bare-Position report), so the diagnostic
 * locations diverge. Excluded rather than faked green:
 * jsx-child-element-spacing, jsx-closing-bracket-location, jsx-curly-spacing,
 * jsx-equals-spacing. Their valid cases (confirming no false positives) match
 * and are kept.
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'exp-jsx-props-style',
    code: '<App foo bar />',
    options: [{ singleLine: { maxItems: 1 } }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'exp-jsx-props-style',
    code: '<App\n  foo\n  bar\n/>',
    options: [{ multiLine: { minItems: 3 } }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'exp-jsx-props-style',
    code: '<App\n  foo bar\n  baz\n/>',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'exp-jsx-props-style',
    code: '<App\n  foo bar baz\n/>',
    options: [{ multiLine: { maxItemsPerLine: 2 } }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-closing-tag-location',
    code: '\n        <App>\n          foo\n          </App>\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-closing-tag-location',
    code: '\n        <App>\n          foo</App>\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-closing-tag-location',
    code: '\n        const x = () => {\n          return <App>\n              foo</App>\n        }\n      ',
    options: ['line-aligned'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-closing-tag-location',
    code: '\n        const x = <App>\n              foo\n                  </App>\n      ',
    options: ['line-aligned'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-curly-brace-presence',
    code: '<App prop={`foo`} />',
    options: [{ props: 'never' }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-curly-brace-presence',
    code: '<App>{`foo`}</App>',
    options: [{ children: 'never' }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-curly-brace-presence',
    code: "<MyComponent>{'foo'}</MyComponent>",
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-curly-brace-presence',
    code: "<MyComponent prop='bar'>foo</MyComponent>",
    options: [{ props: 'always' }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-curly-newline',
    code: '<div>{foo\n}</div>',
    options: [{ singleline: 'consistent', multiline: 'require' }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-curly-newline',
    code: '<div>{\nfoo}</div>',
    options: [{ singleline: 'consistent', multiline: 'require' }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-curly-newline',
    code: '\n        <div>\n          { foo \n}\n        </div>\n      ',
    options: ['consistent'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-curly-newline',
    code: '\n        <div>\n          {\n            foo &&\n            foo.bar\n          }\n        </div>\n      ',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-first-prop-new-line',
    code: '\n        <Foo propOne="one" propTwo="two" />\n      ',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-first-prop-new-line',
    code: '\n        <Foo\n          propOne="one"\n          propTwo="two"\n        />\n      ',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-first-prop-new-line',
    code: '\n        <Foo prop={{\n        }} />\n      ',
    options: ['multiline'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-first-prop-new-line',
    code: '\n      <Foo\nbar />\n      ',
    options: ['multiprop'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-function-call-newline',
    code: 'fn(<div\n        />)',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-function-call-newline',
    code: 'fn(<div />)',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-function-call-newline',
    code: 'fn(\n<div />,<div />,\n<div />)',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-function-call-newline',
    code: 'fn(<div />, <span>\n</span>)',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-indent',
    code: '\n        <App>\n          <Foo />\n        </App>\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-indent',
    code: '\n        <>\n          <Foo />\n        </>\n      ',
  },
];

const CLEAN_CASES: DiffCase[] = [
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'exp-jsx-props-style',
    code: '<App foo bar baz />',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'exp-jsx-props-style',
    code: '<App\n  foo bar\n  baz qux\n/>',
    options: [{ multiLine: { maxItemsPerLine: 2 } }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-child-element-spacing',
    code: '\n        <App>\n          foo<a>bar</a>baz\n        </App>\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-child-element-spacing',
    code: "\n        <App>\n          foo{' '}\n          <a>bar</a>\n          {' '}baz\n        </App>\n      ",
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-closing-bracket-location',
    code: '\n        <App foo />\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-closing-bracket-location',
    code: '\n        <App\n          foo\n          />\n      ',
    options: ['props-aligned'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-closing-tag-location',
    code: '\n        <App>foo</App>\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-closing-tag-location',
    code: '\n        const foo = <App>\n              bar\n        </App>\n      ',
    options: ['line-aligned'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-curly-brace-presence',
    code: '<App {...props}>foo</App>',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-curly-brace-presence',
    code: '<App>{`Hello ${word} World`}</App>',
    options: [{ children: 'never' }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-curly-newline',
    code: '<div>{foo}</div>',
    options: ['consistent'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-curly-newline',
    code: '<div foo={bar} />',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-curly-spacing',
    code: '<App foo={bar} />;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-curly-spacing',
    code: '<App foo={ bar } />;',
    options: [{ when: 'always' }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-equals-spacing',
    code: '<App foo={e => bar(e)} />',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-equals-spacing',
    code: '<App foo = "bar" />',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-first-prop-new-line',
    code: '<Foo prop="bar" />',
    options: ['never'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-first-prop-new-line',
    code: '\n        <Foo\n          propOne="one"\n          propTwo="two"\n        />\n      ',
    options: ['multiline'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-function-call-newline',
    code: 'fn(<div />)',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-function-call-newline',
    code: 'fn(\n<div />\n)',
    options: ['always'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-indent',
    code: '\n        <App>\n          <Foo />\n        </App>\n      ',
    options: [2],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-indent',
    code: '\n\t\t\t\t<App>\n\t\t\t\t\t<Foo />\n\t\t\t\t</App>\n\t\t\t',
    options: ['tab'],
  },
];

runConformanceSuite('@stylistic/eslint-plugin', CASES, CLEAN_CASES);
