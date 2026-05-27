import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const MISSING = 'Missing line break around JSX';

const ALWAYS = ['always'];

ruleTester.run('jsx-function-call-newline', null as never, {
  valid: [
    { code: `fn(<div />)` },
    { code: `fn(<div />, <div />)` },
    { code: `fn(<div />,\n<div />)` },
    { code: `fn(\n<div />, <div />)` },
    { code: `fn(\n<div />, <div />\n)` },
    { code: `fn(\n<div />\n)`, options: ALWAYS },
    { code: `fn(<div />, \n<div \n style={{ color: 'red' }}\n />\n)` },
    { code: `fn(<div />, <div />, <div />)` },
    { code: `fn(<div />, <div />\n, <div />)` },
    { code: `fn(\n<div />\n,\n<div />\n,\n<div />\n)` },
    { code: `fn(\n<div />\n,\n<div />\n,\n<div />\n)`, options: ALWAYS },
    { code: `fn(\n<div />\n,\n<div ></div>)` },
    { code: `fn((<div style={{}} />), <div />, <div />)` },
    { code: `new OBJ((<div style={{}} />), <div />, <div />)` },
    { code: `new OBJ(<div />, <div />, <div />)` },
    { code: `new OBJ(<div />, <div />\n, <div />)` },
    { code: `new OBJ(\n<div />\n,\n<div />\n,\n<div />\n)` },
    { code: `new OBJ(\n<div />\n,\n<div />\n,\n<div />\n)`, options: ALWAYS },
    { code: `new OBJ(\n<div />\n,\n<div ></div>)` },
  ],

  invalid: [
    {
      code: `fn(<div\n        />)`,
      errors: [{ messageId: 'missingLineBreak', message: MISSING }],
    },
    {
      code: `new OBJ(<div\n        />)`,
      errors: [{ messageId: 'missingLineBreak', message: MISSING }],
    },
    {
      code: `fn(<div />)`,
      options: ALWAYS,
      errors: [{ messageId: 'missingLineBreak', message: MISSING }],
    },
    {
      code: `fn(\n<div />,<div />,\n<div />)`,
      options: ALWAYS,
      errors: [
        { messageId: 'missingLineBreak', message: MISSING },
        { messageId: 'missingLineBreak', message: MISSING },
      ],
    },
    {
      code: `new OBJ(\n<div />,<div />,\n<div />)`,
      options: ALWAYS,
      errors: [
        { messageId: 'missingLineBreak', message: MISSING },
        { messageId: 'missingLineBreak', message: MISSING },
      ],
    },
    {
      code: `fn((\n<div />),<div />,\n<div />)`,
      options: ALWAYS,
      errors: [
        { messageId: 'missingLineBreak', message: MISSING },
        { messageId: 'missingLineBreak', message: MISSING },
        { messageId: 'missingLineBreak', message: MISSING },
      ],
    },
    {
      code: `fn(<div />, <span>\n</span>)`,
      errors: [{ messageId: 'missingLineBreak', message: MISSING }],
    },
    {
      code: `fn(<div \n />, <span>\n</span>)`,
      errors: [
        { messageId: 'missingLineBreak', message: MISSING },
        { messageId: 'missingLineBreak', message: MISSING },
      ],
    },
  ],
});
