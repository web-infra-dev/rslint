import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-adjacent-inline-elements', {} as never, {
  valid: [
    { code: '<div />;' },
    { code: '<div><div></div><div></div></div>;' },
    { code: '<div><p></p><div></div></div>;' },
    { code: '<div><p></p><a></a></div>;' },
    { code: '<div><a></a>&nbsp;<a></a></div>;' },
    { code: '<div><a></a>&nbsp;some text &nbsp; <a></a></div>;' },
    { code: '<div><a></a>&nbsp;some text <a></a></div>;' },
    { code: '<div><a></a> <a></a></div>;' },
    { code: '<div><ul><li><a></a></li><li><a></a></li></ul></div>;' },
    { code: '<div><a></a> some text <a></a></div>;' },
    { code: 'React.createElement("div", null, "some text");' },
    {
      code: 'React.createElement("div", undefined, [React.createElement("a"), " some text ", React.createElement("a")]);',
    },
    {
      code: 'React.createElement("div", undefined, [React.createElement("a"), " ", React.createElement("a")]);',
    },
    { code: 'React.createElement(a, b);' },
  ],
  invalid: [
    {
      code: '<div><a></a><a></a></div>;',
      errors: [{ messageId: 'inlineElement' }],
    },
    {
      code: '<div><a></a><span></span></div>;',
      errors: [{ messageId: 'inlineElement' }],
    },
    {
      code: 'React.createElement("div", undefined, [React.createElement("a"), React.createElement("span")]);',
      errors: [{ messageId: 'inlineElement' }],
    },
  ],
});
