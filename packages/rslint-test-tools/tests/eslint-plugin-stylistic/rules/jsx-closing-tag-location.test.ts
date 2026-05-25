import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const MESSAGE_ON_OWN_LINE =
  'Closing tag of a multiline JSX expression must be on its own line.';
const MESSAGE_MATCH_INDENT =
  'Expected closing tag to match indentation of opening.';
const MESSAGE_ALIGN_WITH_OPENING =
  'Expected closing tag to be aligned with the line containing the opening tag';

ruleTester.run('jsx-closing-tag-location', null as never, {
  valid: [
    // default tag-aligned
    { code: `\n    <App>\n      foo\n    </App>\n  ` },
    { code: `\n    <App>foo</App>\n  ` },
    // fragment
    { code: `\n    <>\n      foo\n    </>\n  ` },
    { code: `\n    <>foo</>\n  ` },
    // line-aligned
    {
      code: `\n    const foo = () => {\n      return <App>\n   bar</App>\n    }\n  `,
      options: ['line-aligned'],
    },
    {
      code: `\n    const foo = () => {\n      return <App>\n          bar</App>\n    }\n  `,
    },
    {
      code: `\n    const foo = () => {\n      return <App>\n          bar\n      </App>\n    }\n  `,
      options: ['line-aligned'],
    },
    {
      code: `\n    const foo = <App>\n          bar\n    </App>\n  `,
      options: ['line-aligned'],
    },
    {
      code: `\n    const x = <App>\n          foo\n              </App>\n  `,
    },
    {
      code: `\n    const foo =\n      <App>\n          bar\n      </App>\n  `,
      options: ['line-aligned'],
    },
  ],

  invalid: [
    // default tag-aligned: closing tag alone on its line, wrong indent
    {
      code: `\n    <App>\n      foo\n      </App>\n  `,
      output: `\n    <App>\n      foo\n    </App>\n  `,
      errors: [{ messageId: 'matchIndent', message: MESSAGE_MATCH_INDENT }],
    },
    // default tag-aligned: closing tag shares line with content
    {
      code: `\n    <App>\n      foo</App>\n  `,
      output: `\n    <App>\n      foo\n    </App>\n  `,
      errors: [{ messageId: 'onOwnLine', message: MESSAGE_ON_OWN_LINE }],
    },
    // fragment, alone on its line, wrong indent
    {
      code: `\n    <>\n      foo\n      </>\n  `,
      output: `\n    <>\n      foo\n    </>\n  `,
      errors: [{ messageId: 'matchIndent', message: MESSAGE_MATCH_INDENT }],
    },
    // fragment, shares line with content
    {
      code: `\n    <>\n      foo</>\n  `,
      output: `\n    <>\n      foo\n    </>\n  `,
      errors: [{ messageId: 'onOwnLine', message: MESSAGE_ON_OWN_LINE }],
    },
    // line-aligned: closing tag shares line with content
    {
      code: `\n    const x = () => {\n      return <App>\n          foo</App>\n    }\n  `,
      output: `\n    const x = () => {\n      return <App>\n          foo\n      </App>\n    }\n  `,
      options: ['line-aligned'],
      errors: [{ messageId: 'onOwnLine', message: MESSAGE_ON_OWN_LINE }],
    },
    // line-aligned: closing tag alone on its line, over-indented
    {
      code: `\n    const x = <App>\n          foo\n              </App>\n  `,
      output: `\n    const x = <App>\n          foo\n    </App>\n  `,
      options: ['line-aligned'],
      errors: [
        { messageId: 'alignWithOpening', message: MESSAGE_ALIGN_WITH_OPENING },
      ],
    },
  ],
});
