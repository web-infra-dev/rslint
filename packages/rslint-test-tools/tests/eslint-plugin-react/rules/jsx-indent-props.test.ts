import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
import { diagnostic } from '../expectations';

function expectWrongIndentError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'wrongIndent',
    'Expected indentation of 14 space characters but found 12.',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('jsx-indent-props', {
  valid: [
    // ---- Default 4-space indent ----
    { code: `\n        <App foo\n        />\n      ` },

    // ---- 2-space indent ----
    {
      code: `\n        <App\n          foo\n        />\n      `,
      options: [2],
    },

    // ---- Complex array literal with multiple JSX elements ----
    {
      code: `\n        const Test = () => ([\n          (x\n            ? <div key="1" />\n            : <div key="2" />),\n          <div\n            key="3"\n            align="left"\n          />,\n          <div\n            key="4"\n            align="left"\n          />,\n        ]);\n      `,
      options: [2],
    },

    // ---- 0-indent ----
    {
      code: `\n        <App\n        foo\n        />\n      `,
      options: [0],
    },

    // ---- Negative indent ----
    {
      code: `\n          <App\n        foo\n          />\n      `,
      options: [-2],
    },

    // ---- Tab indent ----
    {
      code: `\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t`,
      options: ['tab'],
    },

    // ---- 'first' option, no props ----
    {
      code: `\n        <App/>\n      `,
      options: ['first'],
    },

    // ---- 'first' option, props aligned with first prop's column ----
    {
      code: `\n        <App aaa\n             b\n             cc\n        />\n      `,
      options: ['first'],
    },
    {
      code: `\n        <App   aaa\n               b\n               cc\n        />\n      `,
      options: ['first'],
    },
    {
      code: `\n        const test = <App aaa\n                          b\n                          cc\n                     />\n      `,
      options: ['first'],
    },
    {
      code: `\n        <App aaa x\n             b y\n             cc\n        />\n      `,
      options: ['first'],
    },
    {
      code: `\n        const test = <App aaa x\n                          b y\n                          cc\n                     />\n      `,
      options: ['first'],
    },
    {
      code: `\n        <App aaa\n             b\n        >\n            <Child c\n                   d/>\n        </App>\n      `,
      options: ['first'],
    },
    {
      code: `\n        <Fragment>\n          <App aaa\n               b\n               cc\n          />\n          <OtherApp a\n                    bbb\n                    c\n          />\n        </Fragment>\n      `,
      options: ['first'],
    },
    {
      code: `\n        <App\n          a\n          b\n        />\n      `,
      options: ['first'],
    },

    // ---- ignoreTernaryOperator: false (default) — props inside the
    // ternary side already have the bump applied. ----
    {
      code: `\n        {this.props.ignoreTernaryOperatorFalse\n          ? <span\n              className="value"\n              some={{aaa}}\n            />\n          : null}\n      `,
      options: [{ indentMode: 2, ignoreTernaryOperator: false }],
    },

    // ---- Function returning JSX in conditional, both flag values. ----
    {
      code: `\n        const F = () => {\n          const foo = true\n            ? <div id="id">test</div>\n            : false;\n\n          return <div\n            id="id"\n          >\n            test\n          </div>\n        }\n      `,
      options: [{ indentMode: 2, ignoreTernaryOperator: false }],
    },
    {
      code: `\n        const F = () => {\n          const foo = true\n            ? <div id="id">test</div>\n            : false;\n\n          return <div\n            id="id"\n          >\n            test\n          </div>\n        }\n      `,
      options: [{ indentMode: 2, ignoreTernaryOperator: true }],
    },

    // ---- Tab indent with conditional + return JSX ----
    {
      code: `\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id="id">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\tid="id"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n`,
      options: [{ indentMode: 'tab', ignoreTernaryOperator: false }],
    },
    {
      code: `\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id="id">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\tid="id"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n`,
      options: [{ indentMode: 'tab', ignoreTernaryOperator: true }],
    },

    // ---- ignoreTernaryOperator:true — no bump applied. ----
    {
      code: `\n        {this.props.ignoreTernaryOperatorTrue\n          ? <span\n            className="value"\n            some={{aaa}}\n            />\n          : null}\n      `,
      options: [{ indentMode: 2, ignoreTernaryOperator: true }],
    },

    // ---- Realistic anchor element, indentMode-only object form ----
    {
      code: `\n        <a\n          role={'button'}\n          className={\`navbar-burger \${open ? 'is-active' : ''}\`}\n          href={'#'}\n          aria-label={'menu'}\n          aria-expanded={false}\n          onClick={openMenu}>\n          <span aria-hidden={'true'}/>\n          <span aria-hidden={'true'}/>\n          <span aria-hidden={'true'}/>\n        </a>\n      `,
      options: [{ indentMode: 2 }],
    },

    // ---- Same realistic element, 'first' alignment ----
    {
      code: `\n        <a role={'button'}\n           className={\`navbar-burger \${open ? 'is-active' : ''}\`}\n           href={'#'}\n           aria-label={'menu'}\n           aria-expanded={false}\n           onClick={openMenu}>\n          <span aria-hidden={'true'}/>\n          <span aria-hidden={'true'}/>\n          <span aria-hidden={'true'}/>\n        </a>\n      `,
      options: ['first'],
    },
  ],
  invalid: [
    // ---- Default 4-space indent — child at 10 must be 12. ----
    {
      code: `\n        <App\n          foo\n        />\n      `,
      output: `\n        <App\n            foo\n        />\n      `,
      errors: [
        diagnostic(
          'wrongIndent',
          'Expected indentation of 12 space characters but found 10.',
          [3, 11, 3, 14],
        ),
      ],
    },

    // ---- 2-space indent — child at 12 must be 10. ----
    {
      code: `\n        <App\n            foo\n        />\n      `,
      output: `\n        <App\n          foo\n        />\n      `,
      options: [2],
      errors: [
        diagnostic(
          'wrongIndent',
          'Expected indentation of 10 space characters but found 12.',
          [3, 13, 3, 16],
        ),
      ],
    },

    // ---- Conditional with JSX in both branches — bump applies. ----
    {
      code: `\n        const test = true\n          ? <span\n            attr="value"\n            />\n          : <span\n            attr="otherValue"\n            />\n      `,
      output: `\n        const test = true\n          ? <span\n              attr="value"\n            />\n          : <span\n              attr="otherValue"\n            />\n      `,
      options: [2],
      errors: [
        expectWrongIndentError([4, 13, 4, 25]),
        expectWrongIndentError([7, 13, 7, 30]),
      ],
    },

    // ---- Alternate wrapped in its own parens — no ternary bump. ----
    {
      code: `\n        const test = true\n          ? <span attr="value" />\n          : (\n            <span\n                attr="otherValue"\n            />\n          )\n      `,
      output: `\n        const test = true\n          ? <span attr="value" />\n          : (\n            <span\n              attr="otherValue"\n            />\n          )\n      `,
      options: [2],
      errors: [
        diagnostic(
          'wrongIndent',
          'Expected indentation of 14 space characters but found 16.',
          [6, 17, 6, 34],
        ),
      ],
    },

    // ---- Ternary alternate, single prop. ----
    {
      code: `\n        {test.isLoading\n          ? <Value/>\n          : <OtherValue\n            some={aaa}/>\n        }\n      `,
      output: `\n        {test.isLoading\n          ? <Value/>\n          : <OtherValue\n              some={aaa}/>\n        }\n      `,
      options: [2],
      errors: [expectWrongIndentError([5, 13, 5, 23])],
    },

    // ---- Ternary alternate, two props bumped. ----
    {
      code: `\n        {test.isLoading\n          ? <Value/>\n          : <OtherValue\n            some={aaa}\n            other={bbb}/>\n        }\n      `,
      output: `\n        {test.isLoading\n          ? <Value/>\n          : <OtherValue\n              some={aaa}\n              other={bbb}/>\n        }\n      `,
      options: [2],
      errors: [
        expectWrongIndentError([5, 13, 5, 23]),
        expectWrongIndentError([6, 13, 6, 24]),
      ],
    },

    // ---- Ternary consequent inside JSXExpressionContainer. ----
    {
      code: `\n        {this.props.test\n          ? <span\n            className="value"\n            some={{aaa}}\n            />\n          : null}\n      `,
      output: `\n        {this.props.test\n          ? <span\n              className="value"\n              some={{aaa}}\n            />\n          : null}\n      `,
      options: [2],
      errors: [
        expectWrongIndentError([4, 13, 4, 30]),
        expectWrongIndentError([5, 13, 5, 25]),
      ],
    },

    // ---- Tab option, prop has no leading tab. ----
    {
      code: `\n        <App1\n            foo\n        />\n      `,
      output: `\n        <App1\n\tfoo\n        />\n      `,
      options: ['tab'],
      errors: [
        diagnostic(
          'wrongIndent',
          'Expected indentation of 1 tab character but found 0.',
          [3, 13, 3, 16],
        ),
      ],
    },

    // ---- Tab option, too many tabs. ----
    {
      code: `\n\t\t\t\t<App\n\t\t\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t`,
      output: `\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t`,
      options: ['tab'],
      errors: [
        diagnostic(
          'wrongIndent',
          'Expected indentation of 5 tab characters but found 7.',
          [3, 8, 3, 11],
        ),
      ],
    },

    // ---- 'first' alignment failures. ----
    {
      code: `\n        <App a\n          b\n        />\n      `,
      output: `\n        <App a\n             b\n        />\n      `,
      options: ['first'],
      errors: [
        diagnostic(
          'wrongIndent',
          'Expected indentation of 13 space characters but found 10.',
          [3, 11, 3, 12],
        ),
      ],
    },
    {
      code: `\n        <App  a\n           b\n        />\n      `,
      output: `\n        <App  a\n              b\n        />\n      `,
      options: ['first'],
      errors: [
        diagnostic(
          'wrongIndent',
          'Expected indentation of 14 space characters but found 11.',
          [3, 12, 3, 13],
        ),
      ],
    },
    {
      code: `\n        <App\n              a\n           b\n        />\n      `,
      output: `\n        <App\n              a\n              b\n        />\n      `,
      options: ['first'],
      errors: [
        diagnostic(
          'wrongIndent',
          'Expected indentation of 14 space characters but found 11.',
          [4, 12, 4, 13],
        ),
      ],
    },
    {
      code: `\n        <App\n          a\n         b\n           c\n        />\n      `,
      output: `\n        <App\n          a\n          b\n          c\n        />\n      `,
      options: ['first'],
      errors: [
        diagnostic(
          'wrongIndent',
          'Expected indentation of 10 space characters but found 9.',
          [4, 10, 4, 11],
        ),
        diagnostic(
          'wrongIndent',
          'Expected indentation of 10 space characters but found 11.',
          [5, 12, 5, 13],
        ),
      ],
    },

    // ---- Return JSX off by one indent unit, both ignoreTernaryOperator
    // values (return JSX is not in a ternary so behaviour matches). ----
    {
      code: `\n        const F = () => {\n          const foo = true\n            ? <div id="id">test</div>\n            : false;\n\n          return <div\n              id="id"\n          >\n            test\n          </div>\n        }\n      `,
      output: `\n        const F = () => {\n          const foo = true\n            ? <div id="id">test</div>\n            : false;\n\n          return <div\n            id="id"\n          >\n            test\n          </div>\n        }\n      `,
      options: [{ indentMode: 2, ignoreTernaryOperator: false }],
      errors: [
        diagnostic(
          'wrongIndent',
          'Expected indentation of 12 space characters but found 14.',
          [8, 15, 8, 22],
        ),
      ],
    },
    {
      code: `\n        const F = () => {\n          const foo = true\n            ? <div id="id">test</div>\n            : false;\n\n          return <div\n              id="id"\n          >\n            test\n          </div>\n        }\n      `,
      output: `\n        const F = () => {\n          const foo = true\n            ? <div id="id">test</div>\n            : false;\n\n          return <div\n            id="id"\n          >\n            test\n          </div>\n        }\n      `,
      options: [{ indentMode: 2, ignoreTernaryOperator: true }],
      errors: [
        diagnostic(
          'wrongIndent',
          'Expected indentation of 12 space characters but found 14.',
          [8, 15, 8, 22],
        ),
      ],
    },

    // ---- Tab indent + return JSX off by one. ----
    {
      code: `\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id="id">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\t\tid="id"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n`,
      output: `\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id="id">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\tid="id"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n`,
      options: [{ indentMode: 'tab', ignoreTernaryOperator: false }],
      errors: [
        diagnostic(
          'wrongIndent',
          'Expected indentation of 6 tab characters but found 7.',
          [8, 8, 8, 15],
        ),
      ],
    },
    {
      code: `\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id="id">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\t\tid="id"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n`,
      output: `\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id="id">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\tid="id"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n`,
      options: [{ indentMode: 'tab', ignoreTernaryOperator: true }],
      errors: [
        diagnostic(
          'wrongIndent',
          'Expected indentation of 6 tab characters but found 7.',
          [8, 8, 8, 15],
        ),
      ],
    },
  ],
});
