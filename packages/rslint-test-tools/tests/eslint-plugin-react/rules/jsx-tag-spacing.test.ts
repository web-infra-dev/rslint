// Upstream: https://github.com/jsx-eslint/eslint-plugin-react/blob/v7.37.5/tests/lib/rules/jsx-tag-spacing.js
// Retain the upstream parser cases below. tsgo cannot parse `< /Tag>`, so
// those inputs are retained with explicit skip reasons, matching the Go suite.
// Intentional differences for unsafe fixes and proportional spacing are noted below.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

// generate options object that disables checks other than the tested one

function closingSlashOptions(option: string) {
  return [
    {
      closingSlash: option,
      beforeSelfClosing: 'allow',
      afterOpening: 'allow',
      beforeClosing: 'allow',
    },
  ];
}

function beforeSelfClosingOptions(option: string) {
  return [
    {
      closingSlash: 'allow',
      beforeSelfClosing: option,
      afterOpening: 'allow',
      beforeClosing: 'allow',
    },
  ];
}

function afterOpeningOptions(option: string) {
  return [
    {
      closingSlash: 'allow',
      beforeSelfClosing: 'allow',
      afterOpening: option,
      beforeClosing: 'allow',
    },
  ];
}

function beforeClosingOptions(option: string) {
  return [
    {
      closingSlash: 'allow',
      beforeSelfClosing: 'allow',
      afterOpening: 'allow',
      beforeClosing: option,
    },
  ];
}

// -----------------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------------

function expectBeforeCloseNeedSpaceError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'beforeCloseNeedSpace',
    message: 'Whitespace is required before closing bracket',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectBeforeSelfCloseNeedSpaceError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'beforeSelfCloseNeedSpace',
    message: 'A space is required before closing bracket',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectBeforeSelfCloseNoSpaceError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'beforeSelfCloseNoSpace',
    message: 'A space is forbidden before closing bracket',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectAfterOpenNoSpaceError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'afterOpenNoSpace',
    message: 'A space is forbidden after opening bracket',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectAfterOpenNeedSpaceError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'afterOpenNeedSpace',
    message: 'A space is required after opening bracket',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectBeforeCloseNoSpaceError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'beforeCloseNoSpace',
    message: 'A space is forbidden before closing bracket',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);
ruleTester.run('jsx-tag-spacing', {
  valid: [
    {
      code: '<App />',
    },
    {
      code: '<App />',
      options: beforeSelfClosingOptions('always'),
    },
    {
      code: '<App foo />',
      options: beforeSelfClosingOptions('always'),
    },
    {
      code: '<App foo={bar} />',
      options: beforeSelfClosingOptions('always'),
    },
    {
      code: '<App {...props} />',
      options: beforeSelfClosingOptions('always'),
    },
    {
      code: '<App></App>',
      options: beforeSelfClosingOptions('always'),
    },
    {
      code: `
        <App
          foo={bar}
        />
      `,
      options: beforeSelfClosingOptions('always'),
    },
    {
      code: '<App/>',
      options: beforeSelfClosingOptions('never'),
    },
    {
      code: '<App />',
      options: beforeSelfClosingOptions('proportional-always'),
    },
    {
      code: '<App foo />',
      options: beforeSelfClosingOptions('proportional-always'),
    },
    {
      code: `
        <App
          foo={bar}
          blat
        >
          hello
        </App>
      `,
      options: beforeClosingOptions('proportional-always'),
    },
    {
      code: `
        <App
          foo={bar}
        />
      `,
      options: beforeSelfClosingOptions('proportional-always'),
    },
    {
      code: '<App foo/>',
      options: beforeSelfClosingOptions('never'),
    },
    {
      code: '<App foo={bar}/>',
      options: beforeSelfClosingOptions('never'),
    },
    {
      code: '<App {...props}/>',
      options: beforeSelfClosingOptions('never'),
    },
    {
      code: '<App></App>',
      options: beforeSelfClosingOptions('never'),
    },
    {
      code: `
        <App
          foo={bar}
        />
      `,
      options: beforeSelfClosingOptions('never'),
    },
    {
      code: '<App/>;',
      options: closingSlashOptions('never'),
    },
    {
      code: '<App />;',
      options: closingSlashOptions('never'),
    },
    {
      code: '<div className="bar"></div>;',
      options: closingSlashOptions('never'),
    },
    {
      code: '<div className="bar"></ div>;',
      options: closingSlashOptions('never'),
    },

    {
      code: '<p/ >',
      options: closingSlashOptions('always'),
    },
    {
      code: '<App/>',
      options: afterOpeningOptions('never'),
    },
    {
      code: '<App></App>',
      options: afterOpeningOptions('never'),
    },
    {
      code: '< App></ App>',
      options: afterOpeningOptions('always'),
    },
    {
      code: '< App/>',
      options: afterOpeningOptions('always'),
    },
    {
      code: `
        <
        App/>
      `,
      options: afterOpeningOptions('allow-multiline'),
    },
    {
      code: '<App />',
      options: beforeClosingOptions('never'),
    },
    {
      code: '<App></App>',
      options: beforeClosingOptions('never'),
    },
    {
      code: `
        <App
        foo="bar"
        >
        </App>
      `,
      options: beforeClosingOptions('never'),
    },
    {
      code: `
        <App
           foo="bar"
        >
        </App>
      `,
      options: beforeClosingOptions('never'),
    },
    {
      code: '<App ></App >',
      options: beforeClosingOptions('always'),
    },
    {
      code: `
        <App
        foo="bar"
        >
        </App >
      `,
      options: beforeClosingOptions('always'),
    },
    {
      code: `
        <App
            foo="bar"
        >
        </App >
      `,
      options: beforeClosingOptions('always'),
    },
    {
      code: '<App/>',
      options: [
        {
          closingSlash: 'never',
          beforeSelfClosing: 'never',
          afterOpening: 'never',
          beforeClosing: 'never',
        },
      ],
    },
    {
      code: '< App / >',
      options: [
        {
          closingSlash: 'always',
          beforeSelfClosing: 'always',
          afterOpening: 'always',
          beforeClosing: 'always',
        },
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 22 (7.37.5)',
      skip: "The native TypeScript parser rejects this JSX form (JSX element 'App' has no corresponding closing tag.); upstream uses a different parser.",
      code: `<App prop="foo">< /App>`,
      options: [
        {
          closingSlash: 'always',
          beforeSelfClosing: 'allow',
          afterOpening: 'allow',
          beforeClosing: 'allow',
        },
      ],
      languageOptions: upstreamModuleOptions,
    },
  ],

  invalid: [
    {
      // Upstream accepts this, but single-line opening tags require spacing.
      code: `
        <App foo={bar}>
          hello
        </App>
      `,
      output: `
        <App foo={bar} >
          hello
        </App>
      `,
      errors: [expectBeforeCloseNeedSpaceError([2, 23, 2, 23])],
      options: beforeClosingOptions('proportional-always'),
    },
    {
      code: '<App/>',
      output: '<App />',
      options: beforeSelfClosingOptions('always'),
      errors: [expectBeforeSelfCloseNeedSpaceError([1, 5, 1, 5])],
    },
    {
      code: '<App foo/>',
      output: '<App foo />',
      options: beforeSelfClosingOptions('always'),
      errors: [expectBeforeSelfCloseNeedSpaceError([1, 9, 1, 9])],
    },
    {
      code: '<App foo={bar}/>',
      output: '<App foo={bar} />',
      options: beforeSelfClosingOptions('always'),
      errors: [expectBeforeSelfCloseNeedSpaceError([1, 15, 1, 15])],
    },
    {
      code: '<App {...props}/>',
      output: '<App {...props} />',
      options: beforeSelfClosingOptions('always'),
      errors: [expectBeforeSelfCloseNeedSpaceError([1, 16, 1, 16])],
    },
    {
      code: '<App />',
      output: '<App/>',
      options: beforeSelfClosingOptions('never'),
      errors: [expectBeforeSelfCloseNoSpaceError([1, 6, 1, 6])],
    },
    {
      code: '<App foo />',
      output: '<App foo/>',
      options: beforeSelfClosingOptions('never'),
      errors: [expectBeforeSelfCloseNoSpaceError([1, 10, 1, 10])],
    },
    {
      code: '<App foo={bar} />',
      output: '<App foo={bar}/>',
      options: beforeSelfClosingOptions('never'),
      errors: [expectBeforeSelfCloseNoSpaceError([1, 16, 1, 16])],
    },
    {
      code: '<App/>',
      output: '<App />',
      options: beforeSelfClosingOptions('proportional-always'),
      errors: [expectBeforeSelfCloseNeedSpaceError([1, 5, 1, 5])],
    },
    {
      code: '<App foo/>',
      output: '<App foo />',
      options: beforeSelfClosingOptions('proportional-always'),
      errors: [expectBeforeSelfCloseNeedSpaceError([1, 9, 1, 9])],
    },
    {
      code: `
        <App
          foo={bar}/>`,
      output: `
        <App
          foo={bar}
/>`,
      options: beforeSelfClosingOptions('proportional-always'),
      errors: [
        {
          messageId: 'beforeSelfCloseNeedNewline',
          message: 'A newline is required before closing bracket',
          line: 3,
          column: 20,
          endLine: 3,
          endColumn: 20,
        },
      ],
    },
    {
      code: `
        <App
          foo={bar} />`,
      output: `
        <App
          foo={bar}${' '}
/>`,
      options: beforeSelfClosingOptions('proportional-always'),
      errors: [
        {
          messageId: 'beforeSelfCloseNeedNewline',
          message: 'A newline is required before closing bracket',
          line: 3,
          column: 20,
          endLine: 3,
          endColumn: 20,
        },
      ],
    },
    {
      code: `
        <App
          foo={bar}
          blat >
          hello
        </App>
      `,
      output: `
        <App
          foo={bar}
          blat${' '}
>
          hello
        </App>
      `,
      options: beforeClosingOptions('proportional-always'),
      errors: [
        {
          messageId: 'beforeCloseNeedNewline',
          message: 'A newline is required before closing bracket',
          line: 4,
          column: 15,
          endLine: 4,
          endColumn: 15,
        },
      ],
    },
    {
      code: `
        <App
          foo={bar}>
          hello
        </App>
      `,
      output: `
        <App
          foo={bar}
>
          hello
        </App>
      `,
      options: beforeClosingOptions('proportional-always'),
      errors: [
        {
          messageId: 'beforeCloseNeedNewline',
          message: 'A newline is required before closing bracket',
          line: 3,
          column: 20,
          endLine: 3,
          endColumn: 20,
        },
      ],
    },
    {
      code: '<App {...props} />',
      output: '<App {...props}/>',
      options: beforeSelfClosingOptions('never'),
      errors: [expectBeforeSelfCloseNoSpaceError([1, 17, 1, 17])],
    },
    {
      code: '<App/ >;',
      output: '<App/>;',
      errors: [
        {
          messageId: 'selfCloseSlashNoSpace',
          message: 'Whitespace is forbidden between `/` and `>`; write `/>`',
          line: 1,
          column: 5,
          endLine: 1,
          endColumn: 8,
        },
      ],
      options: closingSlashOptions('never'),
    },
    {
      code: `
        <App_selfCloseSlashNoSpace/
        >
      `,
      output: `
        <App_selfCloseSlashNoSpace/>
      `,
      errors: [
        {
          messageId: 'selfCloseSlashNoSpace',
          message: 'Whitespace is forbidden between `/` and `>`; write `/>`',
          line: 2,
          column: 35,
          endLine: 3,
          endColumn: 10,
        },
      ],
      options: closingSlashOptions('never'),
    },

    {
      code: '<App prop="foo"></App>',
      // Upstream inserts `< /App>`; rslint omits this unparseable fix.

      errors: [
        {
          messageId: 'closeSlashNeedSpace',
          message: 'Whitespace is required between `<` and `/`; write `< /`',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 19,
        },
      ],
      options: closingSlashOptions('always'),
    },
    {
      code: '<p/>',
      output: '<p/ >',
      errors: [
        {
          messageId: 'selfCloseSlashNeedSpace',
          message: 'Whitespace is required between `/` and `>`; write `/ >`',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 5,
        },
      ],
      options: closingSlashOptions('always'),
    },
    {
      code: '< App/>',
      output: '<App/>',
      errors: [expectAfterOpenNoSpaceError([1, 1, 1, 3])],
      options: afterOpeningOptions('never'),
    },
    {
      code: '< App></App>',
      output: '<App></App>',
      errors: [expectAfterOpenNoSpaceError([1, 1, 1, 3])],
      options: afterOpeningOptions('never'),
    },
    {
      code: '<App></ App>',
      output: '<App></App>',
      errors: [expectAfterOpenNoSpaceError([1, 7, 1, 9])],
      options: afterOpeningOptions('never'),
    },
    {
      code: '< App></ App>',
      output: '<App></App>',
      errors: [
        expectAfterOpenNoSpaceError([1, 1, 1, 3]),
        expectAfterOpenNoSpaceError([1, 8, 1, 10]),
      ],
      options: afterOpeningOptions('never'),
    },
    {
      code: `
        <
        App1/>
      `,
      output: `
        <App1/>
      `,
      errors: [expectAfterOpenNoSpaceError([2, 9, 3, 9])],
      options: afterOpeningOptions('never'),
    },
    {
      code: '<App2></ App2>',
      output: '< App2></ App2>',
      errors: [expectAfterOpenNeedSpaceError([1, 1, 1, 2])],
      options: afterOpeningOptions('always'),
    },
    {
      code: '< App3></App3>',
      output: '< App3></ App3>',
      errors: [expectAfterOpenNeedSpaceError([1, 9, 1, 10])],
      options: afterOpeningOptions('always'),
    },
    {
      code: '<App4></App4>',
      output: '< App4></ App4>',
      errors: [
        expectAfterOpenNeedSpaceError([1, 1, 1, 2]),
        expectAfterOpenNeedSpaceError([1, 8, 1, 9]),
      ],
      options: afterOpeningOptions('always'),
    },
    {
      code: '<App5/>',
      output: '< App5/>',
      errors: [expectAfterOpenNeedSpaceError([1, 1, 1, 2])],
      options: afterOpeningOptions('always'),
    },
    {
      code: '< App6/>',
      output: '<App6/>',
      errors: [expectAfterOpenNoSpaceError([1, 1, 1, 3])],
      options: afterOpeningOptions('allow-multiline'),
    },
    {
      code: '<App7 ></App7>',
      output: '<App7></App7>',
      errors: [expectBeforeCloseNoSpaceError([1, 6, 1, 7])],
      options: beforeClosingOptions('never'),
    },
    {
      code: '<App8></App8 >',
      output: '<App8></App8>',
      errors: [expectBeforeCloseNoSpaceError([1, 13, 1, 14])],
      options: beforeClosingOptions('never'),
    },
    {
      code: `
        <App9
        foo="bar"
        >
        </App9 >
      `,
      output: `
        <App9
        foo="bar"
        >
        </App9>
      `,
      errors: [expectBeforeCloseNoSpaceError([5, 15, 5, 16])],
      options: beforeClosingOptions('never'),
    },
    {
      code: '<App10></App10 >',
      output: '<App10 ></App10 >',
      errors: [expectBeforeCloseNeedSpaceError([1, 7, 1, 7])],
      options: beforeClosingOptions('always'),
    },
    {
      code: '<App11 ></App11>',
      output: '<App11 ></App11 >',
      errors: [expectBeforeCloseNeedSpaceError([1, 16, 1, 16])],
      options: beforeClosingOptions('always'),
    },
    {
      code: `
        <App12
        foo="bar"
        >
        </App12>
      `,
      output: `
        <App12
        foo="bar"
        >
        </App12 >
      `,
      errors: [expectBeforeCloseNeedSpaceError([5, 16, 5, 16])],
      options: beforeClosingOptions('always'),
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 16 (7.37.5)',
      skip: "The native TypeScript parser rejects this JSX form (JSX element 'div' has no corresponding closing tag.); upstream uses a different parser.",
      // Preserve the historical fixture without inventing native diagnostics.
      create: () => {
        throw new Error(
          JSON.stringify({
            code: '<div className="bar">< /div>;',
            options: [
              {
                closingSlash: 'never',
                beforeSelfClosing: 'allow',
                afterOpening: 'allow',
                beforeClosing: 'allow',
              },
            ],
            settings: {},
            parserOptions: {
              ecmaVersion: 2018,
              sourceType: 'module',
              ecmaFeatures: {
                jsx: true,
              },
            },
            features: ['no-ts'],
            upstreamErrors: [
              {
                messageId: 'closeSlashNoSpace',
              },
            ],
            upstreamOutput: '<div className="bar"></div>;',
          }),
        );
      },
    },
    {
      name: 'Upstream invalid 17 (7.37.5)',
      skip: "The native TypeScript parser rejects this JSX form (JSX element 'div' has no corresponding closing tag.); upstream uses a different parser.",
      // Preserve the historical fixture without inventing native diagnostics.
      create: () => {
        throw new Error(
          JSON.stringify({
            code: '\n        <div className="bar"><\n        /div>;\n      ',
            options: [
              {
                closingSlash: 'never',
                beforeSelfClosing: 'allow',
                afterOpening: 'allow',
                beforeClosing: 'allow',
              },
            ],
            settings: {},
            parserOptions: {
              ecmaVersion: 2018,
              sourceType: 'module',
              ecmaFeatures: {
                jsx: true,
              },
            },
            features: ['no-ts'],
            upstreamErrors: [
              {
                messageId: 'closeSlashNoSpace',
              },
            ],
            upstreamOutput: '\n        <div className="bar"></div>;\n      ',
          }),
        );
      },
    },
  ],
});
