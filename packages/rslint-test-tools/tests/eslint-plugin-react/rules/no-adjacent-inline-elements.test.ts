import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
import { diagnostic } from '../expectations';

function expectInlineElementError(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic(
    'inlineElement',
    'Child elements which render as inline HTML elements should be separated by a space or wrapped in block level elements.',
    range,
  );
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('no-adjacent-inline-elements', {
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
      errors: [expectInlineElementError([1, 1, 1, 26])],
    },
    {
      code: '<div><a></a><span></span></div>;',
      errors: [expectInlineElementError([1, 1, 1, 32])],
    },
    {
      code: 'React.createElement("div", undefined, [React.createElement("a"), React.createElement("span")]);',
      errors: [expectInlineElementError([1, 1, 1, 95])],
    },
  ],
});
