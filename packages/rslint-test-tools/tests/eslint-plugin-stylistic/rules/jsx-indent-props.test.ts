import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

// Mirrors upstream packages/eslint-plugin/rules/jsx-indent-props/
// jsx-indent-props.test.ts (Layer 1). This JS suite verifies binary
// registration + wire protocol + ESLint-compatible diagnostics; edge-shape and
// branch lock-in cases live in the Go jsx_indent_props_extras_test.go.
ruleTester.run('jsx-indent-props', null as never, {
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

    // ---- ignoreTernaryOperator: false (default) — props inside the ternary
    // side already have the bump applied. ----
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
      errors: [{ messageId: 'wrongIndent' }],
    },

    // ---- 2-space indent — child at 12 must be 10. ----
    {
      code: `\n        <App\n            foo\n        />\n      `,
      output: `\n        <App\n          foo\n        />\n      `,
      options: [2],
      errors: [{ messageId: 'wrongIndent' }],
    },

    // ---- Conditional with JSX in both branches — bump applies. ----
    {
      code: `\n        const test = true\n          ? <span\n            attr="value"\n            />\n          : <span\n            attr="otherValue"\n            />\n      `,
      output: `\n        const test = true\n          ? <span\n              attr="value"\n            />\n          : <span\n              attr="otherValue"\n            />\n      `,
      options: [2],
      errors: [{ messageId: 'wrongIndent' }, { messageId: 'wrongIndent' }],
    },

    // ---- Alternate wrapped in its own parens — no ternary bump. ----
    {
      code: `\n        const test = true\n          ? <span attr="value" />\n          : (\n            <span\n                attr="otherValue"\n            />\n          )\n      `,
      output: `\n        const test = true\n          ? <span attr="value" />\n          : (\n            <span\n              attr="otherValue"\n            />\n          )\n      `,
      options: [2],
      errors: [{ messageId: 'wrongIndent' }],
    },

    // ---- Ternary alternate, single prop. ----
    {
      code: `\n        {test.isLoading\n          ? <Value/>\n          : <OtherValue\n            some={aaa}/>\n        }\n      `,
      output: `\n        {test.isLoading\n          ? <Value/>\n          : <OtherValue\n              some={aaa}/>\n        }\n      `,
      options: [2],
      errors: [{ messageId: 'wrongIndent' }],
    },

    // ---- Ternary alternate, two props bumped. ----
    {
      code: `\n        {test.isLoading\n          ? <Value/>\n          : <OtherValue\n            some={aaa}\n            other={bbb}/>\n        }\n      `,
      output: `\n        {test.isLoading\n          ? <Value/>\n          : <OtherValue\n              some={aaa}\n              other={bbb}/>\n        }\n      `,
      options: [2],
      errors: [{ messageId: 'wrongIndent' }, { messageId: 'wrongIndent' }],
    },

    // ---- Ternary consequent inside JSXExpressionContainer. ----
    {
      code: `\n        {this.props.test\n          ? <span\n            className="value"\n            some={{aaa}}\n            />\n          : null}\n      `,
      output: `\n        {this.props.test\n          ? <span\n              className="value"\n              some={{aaa}}\n            />\n          : null}\n      `,
      options: [2],
      errors: [{ messageId: 'wrongIndent' }, { messageId: 'wrongIndent' }],
    },

    // ---- Tab option, prop has no leading tab. ----
    {
      code: `\n        <App1\n            foo\n        />\n      `,
      output: `\n        <App1\n\tfoo\n        />\n      `,
      options: ['tab'],
      errors: [{ messageId: 'wrongIndent' }],
    },

    // ---- Tab option, too many tabs. ----
    {
      code: `\n\t\t\t\t<App\n\t\t\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t`,
      output: `\n\t\t\t\t<App\n\t\t\t\t\tfoo\n\t\t\t\t/>\n\t\t\t`,
      options: ['tab'],
      errors: [{ messageId: 'wrongIndent' }],
    },

    // ---- 'first' alignment failures. ----
    {
      code: `\n        <App a\n          b\n        />\n      `,
      output: `\n        <App a\n             b\n        />\n      `,
      options: ['first'],
      errors: [{ messageId: 'wrongIndent' }],
    },
    {
      code: `\n        <App  a\n           b\n        />\n      `,
      output: `\n        <App  a\n              b\n        />\n      `,
      options: ['first'],
      errors: [{ messageId: 'wrongIndent' }],
    },
    {
      code: `\n        <App\n              a\n           b\n        />\n      `,
      output: `\n        <App\n              a\n              b\n        />\n      `,
      options: ['first'],
      errors: [{ messageId: 'wrongIndent' }],
    },
    {
      code: `\n        <App\n          a\n         b\n           c\n        />\n      `,
      output: `\n        <App\n          a\n          b\n          c\n        />\n      `,
      options: ['first'],
      errors: [{ messageId: 'wrongIndent' }, { messageId: 'wrongIndent' }],
    },

    // ---- Return JSX off by one indent unit, both ignoreTernaryOperator
    // values (return JSX is not in a ternary so behaviour matches). ----
    {
      code: `\n        const F = () => {\n          const foo = true\n            ? <div id="id">test</div>\n            : false;\n\n          return <div\n              id="id"\n          >\n            test\n          </div>\n        }\n      `,
      output: `\n        const F = () => {\n          const foo = true\n            ? <div id="id">test</div>\n            : false;\n\n          return <div\n            id="id"\n          >\n            test\n          </div>\n        }\n      `,
      options: [{ indentMode: 2, ignoreTernaryOperator: false }],
      errors: [{ messageId: 'wrongIndent' }],
    },
    {
      code: `\n        const F = () => {\n          const foo = true\n            ? <div id="id">test</div>\n            : false;\n\n          return <div\n              id="id"\n          >\n            test\n          </div>\n        }\n      `,
      output: `\n        const F = () => {\n          const foo = true\n            ? <div id="id">test</div>\n            : false;\n\n          return <div\n            id="id"\n          >\n            test\n          </div>\n        }\n      `,
      options: [{ indentMode: 2, ignoreTernaryOperator: true }],
      errors: [{ messageId: 'wrongIndent' }],
    },

    // ---- Tab indent + return JSX off by one. ----
    {
      code: `\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id="id">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\t\tid="id"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n`,
      output: `\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id="id">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\tid="id"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n`,
      options: [{ indentMode: 'tab', ignoreTernaryOperator: false }],
      errors: [{ messageId: 'wrongIndent' }],
    },
    {
      code: `\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id="id">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\t\tid="id"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n`,
      output: `\n\t\t\t\tconst F = () => {\n\t\t\t\t\tconst foo = true\n\t\t\t\t\t\t? <div id="id">test</div>\n\t\t\t\t\t\t: false;\n\n\t\t\t\t\treturn <div\n\t\t\t\t\t\tid="id"\n\t\t\t\t\t>\n\t\t\t\t\t\ttest\n\t\t\t\t\t</div>\n\t\t\t\t}\n`,
      options: [{ indentMode: 'tab', ignoreTernaryOperator: true }],
      errors: [{ messageId: 'wrongIndent' }],
    },
  ],
});
