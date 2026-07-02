/**
 * Conformance: @stylistic/eslint-plugin (jsx-b) mounted in rslint via `plugins`
 * must report identically to ESLint v10. Representative triggers from the
 * upstream suite; each verified to reproduce ESLint v10 byte-for-byte.
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-max-props-per-line',
    code: '\n        <App foo bar baz />;\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-max-props-per-line',
    code: '\n        <App foo bar baz />;\n      ',
    options: [{ maximum: 2 }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-max-props-per-line',
    code: '\n        <App {...this.props} bar />;\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-max-props-per-line',
    code: '\n        <App\n          foo bar\n          baz\n        />\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-newline',
    code: '\n        <div>\n          <Button>{data.label}</Button>\n          <List />\n        </div>\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-newline',
    code: '\n        <div>\n          <Button>{data.label}</Button>\n          {showSomething === true && <Something />}\n        </div>\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-pascal-case',
    code: '<Test_component />',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-pascal-case',
    code: '<TEST_COMPONENT />',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-pascal-case',
    code: '<TEST_COMPONENT_ />',
    options: [{ allowAllCaps: true }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-indent-props',
    code: '\n        <App\n          foo\n        />\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-indent-props',
    code: '\n        <App\n            foo\n        />\n      ',
    options: [2],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-one-expression-per-line',
    code: '\n        <App>{"foo"}</App>\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-one-expression-per-line',
    code: '\n        <App>foo</App>\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-one-expression-per-line',
    code: '<App\n  foo\n>bar\n</App>',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-props-no-multi-spaces',
    code: '\n        <App  foo />\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-props-no-multi-spaces',
    code: '\n        <App foo="with  spaces   "   bar />\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-props-no-multi-spaces',
    code: '<Foo.Bar  baz="quux" />',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-props-no-multi-spaces',
    code: '\n        <button\n          title=\'Some button\'\n\n          type="button"\n        />\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-quotes',
    code: "<foo bar='baz' />",
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-quotes',
    code: '<foo bar="baz" />',
    options: ['prefer-single'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-quotes',
    code: '<foo bar="&quot;" />',
    options: ['prefer-single'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-quotes',
    code: "<foo bar='&#39;' />",
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-self-closing-comp',
    code: 'var contentContainer = <div className="content"></div>;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-self-closing-comp',
    code: 'var HelloJohn = <Hello name="John"></Hello>;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-self-closing-comp',
    code: 'var HelloJohn = <Hello name="John">\n</Hello>;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-self-closing-comp',
    code: 'var contentContainer = <div className="content"></div>;',
    options: [{ html: true }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-sort-props',
    code: '<App b a />;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-sort-props',
    code: '<App aB a />;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-sort-props',
    code: '<App fistName="John" tel={5555555} name="John Smith" lastName="Smith" Number="2" />;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-tag-spacing',
    code: '<App7 ></App7>',
    options: [
      {
        closingSlash: 'allow',
        beforeSelfClosing: 'allow',
        afterOpening: 'allow',
        beforeClosing: 'never',
      },
    ],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-wrap-multilines',
    code: '\n  var Hello = createReactClass({\n    render: function() {\n      return <div>\n        <p>Hello {this.props.name}</p>\n      </div>;\n    }\n  });\n',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-wrap-multilines',
    code: '\n  var Hello = createReactClass({\n    render: function() {\n      return <>\n        <p>Hello {this.props.name}</p>\n      </>;\n    }\n  });\n',
  },
];

const CLEAN_CASES: DiffCase[] = [
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-indent-props',
    code: '\n        <App\n          foo\n        />\n      ',
    options: [2],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-indent-props',
    code: '\n        <App aaa\n             b\n             cc\n        />\n      ',
    options: ['first'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-max-props-per-line',
    code: '<App foo bar />',
    options: [{ maximum: 2 }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-max-props-per-line',
    code: '\n        <App\n          foo\n          bar\n        />\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-newline',
    code: '\n        <>\n          <Button>{data.label}</Button>\n          Test\n          <span>Should be in new line</span>\n        </>\n      ',
    options: [{ prevent: true }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-newline',
    code: "\n        <Button popoverOpen='#settings-popover' style={{ width: 'fit-content' }}>\n          {/* should work inside a component */}\n          {/* and it should work when using multiple comments */}\n          <Icon f7='gear' />\n        </Button>\n      ",
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-one-expression-per-line',
    code: '<App />',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-one-expression-per-line',
    code: '\n        <App>\n          <Foo />\n          <Bar />\n        </App>\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-pascal-case',
    code: '<TestComponent />',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-pascal-case',
    code: '<YMCA />',
    options: [{ allowAllCaps: true }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-props-no-multi-spaces',
    code: '\n        <App foo bar />\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-props-no-multi-spaces',
    code: '\n        <App foo="with  spaces   " bar />\n      ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-quotes',
    code: '<foo bar="baz" />',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-quotes',
    code: '<foo bar="\'" />',
    options: ['prefer-single'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-self-closing-comp',
    code: 'var HelloJohn = <Hello name="John" />;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-self-closing-comp',
    code: 'var HelloJohn = <Hello name="John"></Hello>;',
    options: [{ component: false }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-sort-props',
    code: '<App a b c />;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-sort-props',
    code: '<App a z onBar onFoo />;',
    options: [{ callbacksLast: true }],
  },
  { pkg: '@stylistic/eslint-plugin', rule: 'jsx-tag-spacing', code: '<App />' },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-tag-spacing',
    code: '<App/>',
    options: [
      {
        closingSlash: 'allow',
        beforeSelfClosing: 'never',
        afterOpening: 'allow',
        beforeClosing: 'allow',
      },
    ],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'jsx-wrap-multilines',
    code: '\n  var Hello = createReactClass({\n    render: function() {\n      return (<div>\n        <p>Hello {this.props.name}</p>\n      </div>);\n    }\n  });\n',
  },
];

runConformanceSuite('@stylistic/eslint-plugin', CASES, CLEAN_CASES);
