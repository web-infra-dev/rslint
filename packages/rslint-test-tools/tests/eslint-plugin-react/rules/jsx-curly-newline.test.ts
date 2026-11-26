import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, upstreamModuleOptions } from '../config';

function expectUnexpectedBeforeError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unexpectedBefore',
    message: "Unexpected newline before '}'.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectExpectedBeforeError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'expectedBefore',
    message: "Expected newline before '}'.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectUnexpectedAfterError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'unexpectedAfter',
    message: "Unexpected newline after '{'.",
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectExpectedAfterError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'expectedAfter',
    message: "Expected newline after '{'.",
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);
const multilineRequire = [{ singleline: 'consistent', multiline: 'require' }];

ruleTester.run('jsx-curly-newline', {
  valid: [
    { code: '<div>{foo}</div>', options: ['consistent'] },
    { code: '<div>{\nfoo\n}</div>', options: ['consistent'] },
    { code: '<div>{ foo &&\nfoo.bar }</div>', options: ['never'] },
    { code: '<div>{\nfoo\n}</div>', options: multilineRequire },
    { code: '<div foo={bar} />', options: multilineRequire },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream valid 1 (7.37.5)',
      code: `
        <div>
          {
            foo
          }
        </div>
      `,
      options: ['consistent'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 2 (7.37.5)',
      code: `
        <div>
          { foo &&
            foo.bar }
        </div>
      `,
      options: ['consistent'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 3 (7.37.5)',
      code: `
        <div>
          {
            foo &&
            foo.bar
          }
        </div>
      `,
      options: ['consistent'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 4 (7.37.5)',
      code: `
        <div foo={
          bar
        } />
      `,
      options: ['consistent'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 5 (7.37.5)',
      code: `<div>{foo}</div>`,
      options: [{ singleline: 'consistent', multiline: 'require' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 7 (7.37.5)',
      code: `
        <div>
          {
            foo &&
            foo.bar
          }
        </div>
      `,
      options: [{ singleline: 'consistent', multiline: 'require' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 8 (7.37.5)',
      code: `
        <div>
          {
            foo
          }
        </div>
      `,
      options: [{ singleline: 'consistent', multiline: 'require' }],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 9 (7.37.5)',
      code: `<div>{foo}</div>`,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 10 (7.37.5)',
      code: `<div foo={bar} />`,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream valid 11 (7.37.5)',
      code: `
        <div>
          { foo &&
            foo.bar }
        </div>
      `,
      options: ['never'],
      languageOptions: upstreamModuleOptions,
    },
  ],
  invalid: [
    {
      code: '<div>{ foo\n}</div>',
      options: ['consistent'],
      output: '<div>{ foo}</div>',
      errors: [expectUnexpectedBeforeError([2, 1, 2, 2])],
    },
    {
      code: '<div>{\nfoo}</div>',
      options: multilineRequire,
      output: '<div>{\nfoo\n}</div>',
      errors: [expectExpectedBeforeError([2, 4, 2, 5])],
    },
    {
      code: '<div>{\nfoo\n}</div>',
      options: ['never'],
      output: '<div>{foo}</div>',
      errors: [
        expectUnexpectedAfterError([1, 6, 1, 7]),
        expectUnexpectedBeforeError([3, 1, 3, 2]),
      ],
    },
    {
      code: '<div>{ foo &&\nbar }</div>',
      options: multilineRequire,
      output: '<div>{\n foo &&\nbar \n}</div>',
      errors: [
        expectExpectedAfterError([1, 6, 1, 7]),
        expectExpectedBeforeError([2, 5, 2, 6]),
      ],
    },

    // Additional input/configuration combinations from pinned upstream.
    {
      name: 'Upstream invalid 0 (7.37.5)',
      code: `
        <div>
          { foo\x20
}
        </div>
      `,
      options: ['consistent'],
      errors: [expectUnexpectedBeforeError([4, 1, 4, 2])],
      output: `
        <div>
          { foo}
        </div>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 1 (7.37.5)',
      code: `
        <div>
          { foo &&
            foo.bar\x20
}
        </div>
      `,
      options: ['consistent'],
      errors: [expectUnexpectedBeforeError([5, 1, 5, 2])],
      output: `
        <div>
          { foo &&
            foo.bar}
        </div>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 2 (7.37.5)',
      code: `
        <div>
          { foo &&
            bar
          }
        </div>
      `,
      options: ['consistent'],
      errors: [expectUnexpectedBeforeError([5, 11, 5, 12])],
      output: `
        <div>
          { foo &&
            bar}
        </div>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 3 (7.37.5)',
      code: `<div>{foo
}</div>`,
      options: [{ singleline: 'consistent', multiline: 'require' }],
      errors: [expectUnexpectedBeforeError([2, 1, 2, 2])],
      output: `<div>{foo}</div>`,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 5 (7.37.5)',
      code: `
        <div>
          { foo &&
            bar }
        </div>
      `,
      options: [{ singleline: 'consistent', multiline: 'require' }],
      errors: [
        expectExpectedAfterError([3, 11, 3, 12]),
        expectExpectedBeforeError([4, 17, 4, 18]),
      ],
      output: `
        <div>
          {
 foo &&
            bar\x20
}
        </div>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 6 (7.37.5)',
      code: `
        <div style={foo &&
          foo.bar
        } />
      `,
      options: [{ singleline: 'consistent', multiline: 'require' }],
      errors: [expectExpectedAfterError([2, 20, 2, 21])],
      output: `
        <div style={
foo &&
          foo.bar
        } />
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 7 (7.37.5)',
      code: `
        <div>
          {
foo
}
        </div>
      `,
      options: ['never'],
      errors: [
        expectUnexpectedAfterError([3, 11, 3, 12]),
        expectUnexpectedBeforeError([5, 1, 5, 2]),
      ],
      output: `
        <div>
          {foo}
        </div>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 8 (7.37.5)',
      code: `
        <div>
          {
            foo &&
            foo.bar
          }
        </div>
      `,
      options: ['never'],
      errors: [
        expectUnexpectedAfterError([3, 11, 3, 12]),
        expectUnexpectedBeforeError([6, 11, 6, 12]),
      ],
      output: `
        <div>
          {foo &&
            foo.bar}
        </div>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 9 (7.37.5)',
      code: `
        <div>
          { foo &&
            foo.bar
          }
        </div>
      `,
      options: ['never'],
      errors: [expectUnexpectedBeforeError([5, 11, 5, 12])],
      output: `
        <div>
          { foo &&
            foo.bar}
        </div>
      `,
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 10 (7.37.5)',
      code: `
        <div>
          { /* not fixed due to comment */
            foo }
        </div>
      `,
      options: ['never'],
      errors: [expectUnexpectedAfterError([3, 11, 3, 12])],
      languageOptions: upstreamModuleOptions,
    },
    {
      name: 'Upstream invalid 11 (7.37.5)',
      code: `
        <div>
          { foo
            /* not fixed due to comment */}
        </div>
      `,
      options: ['never'],
      errors: [expectUnexpectedBeforeError([4, 43, 4, 44])],
      languageOptions: upstreamModuleOptions,
    },
  ],
});
