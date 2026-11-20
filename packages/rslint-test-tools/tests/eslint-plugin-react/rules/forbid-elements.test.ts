import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

function expectForbiddenElementError1(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'forbiddenElement',
    message: '<button> is forbidden',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectForbiddenElementError2(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'forbiddenElement',
    message: '<input> is forbidden',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('forbid-elements', {
  valid: [
    // ---- Upstream valid cases ----
    { code: '<button />' },
    { code: '<button />', options: [{ forbid: [] }] },
    { code: '<Button />', options: [{ forbid: ['button'] }] },
    { code: '<Button />', options: [{ forbid: [{ element: 'button' }] }] },
    { code: 'React.createElement(button)', options: [{ forbid: ['button'] }] },
    { code: 'createElement("button")', options: [{ forbid: ['button'] }] },
    {
      code: 'NotReact.createElement("button")',
      options: [{ forbid: ['button'] }],
    },
    {
      code: 'React.createElement("_thing")',
      options: [{ forbid: ['_thing'] }],
    },
    { code: 'React.createElement("Modal")', options: [{ forbid: ['Modal'] }] },
    {
      code: 'React.createElement("dotted.component")',
      options: [{ forbid: ['dotted.component'] }],
    },
    {
      code: 'React.createElement(function() {})',
      options: [{ forbid: ['button'] }],
    },
    { code: 'React.createElement({})', options: [{ forbid: ['button'] }] },
    { code: 'React.createElement(1)', options: [{ forbid: ['button'] }] },
    { code: 'React.createElement()' },
  ],
  invalid: [
    // ---- Upstream invalid cases ----
    {
      code: '<button />',
      options: [{ forbid: ['button'] }],
      errors: [expectForbiddenElementError1([1, 2, 1, 8])],
    },
    {
      code: '[<Modal />, <button />]',
      options: [{ forbid: ['button', 'Modal'] }],
      errors: [
        {
          messageId: 'forbiddenElement',
          message: '<Modal> is forbidden',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 8,
        },
        expectForbiddenElementError1([1, 14, 1, 20]),
      ],
    },
    {
      code: '<dotted.component />',
      options: [{ forbid: ['dotted.component'] }],
      errors: [
        {
          messageId: 'forbiddenElement',
          message: '<dotted.component> is forbidden',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<dotted.Component />',
      options: [
        {
          forbid: [{ element: 'dotted.Component', message: "that ain't cool" }],
        },
      ],
      errors: [
        {
          messageId: 'forbiddenElement_message',
          message: "<dotted.Component> is forbidden, that ain't cool",
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: '<button />',
      options: [
        { forbid: [{ element: 'button', message: 'use <Button> instead' }] },
      ],
      errors: [
        {
          messageId: 'forbiddenElement_message',
          message: '<button> is forbidden, use <Button> instead',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: '<button><input /></button>',
      options: [{ forbid: [{ element: 'button' }, { element: 'input' }] }],
      errors: [
        expectForbiddenElementError1([1, 2, 1, 8]),
        expectForbiddenElementError2([1, 10, 1, 15]),
      ],
    },
    {
      code: '<button><input /></button>',
      options: [{ forbid: [{ element: 'button' }, 'input'] }],
      errors: [
        expectForbiddenElementError1([1, 2, 1, 8]),
        expectForbiddenElementError2([1, 10, 1, 15]),
      ],
    },
    {
      code: '<button><input /></button>',
      options: [{ forbid: ['input', { element: 'button' }] }],
      errors: [
        expectForbiddenElementError1([1, 2, 1, 8]),
        expectForbiddenElementError2([1, 10, 1, 15]),
      ],
    },
    {
      code: '<button />',
      options: [
        {
          forbid: [
            { element: 'button', message: 'use <Button> instead' },
            { element: 'button', message: 'use <Button2> instead' },
          ],
        },
      ],
      errors: [
        {
          messageId: 'forbiddenElement_message',
          message: '<button> is forbidden, use <Button2> instead',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
    {
      code: 'React.createElement("button", {}, child)',
      options: [{ forbid: ['button'] }],
      errors: [expectForbiddenElementError1([1, 21, 1, 29])],
    },
    {
      code: '[React.createElement(Modal), React.createElement("button")]',
      options: [{ forbid: ['button', 'Modal'] }],
      errors: [
        {
          messageId: 'forbiddenElement',
          message: '<Modal> is forbidden',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 27,
        },
        expectForbiddenElementError1([1, 50, 1, 58]),
      ],
    },
    {
      code: 'React.createElement(dotted.Component)',
      options: [
        {
          forbid: [{ element: 'dotted.Component', message: "that ain't cool" }],
        },
      ],
      errors: [
        {
          messageId: 'forbiddenElement_message',
          message: "<dotted.Component> is forbidden, that ain't cool",
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: 'React.createElement(dotted.component)',
      options: [{ forbid: ['dotted.component'] }],
      errors: [
        {
          messageId: 'forbiddenElement',
          message: '<dotted.component> is forbidden',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      code: 'React.createElement(_comp)',
      options: [{ forbid: ['_comp'] }],
      errors: [
        {
          messageId: 'forbiddenElement',
          message: '<_comp> is forbidden',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'React.createElement("button")',
      options: [
        { forbid: [{ element: 'button', message: 'use <Button> instead' }] },
      ],
      errors: [
        {
          messageId: 'forbiddenElement_message',
          message: '<button> is forbidden, use <Button> instead',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      code: 'React.createElement("button", {}, React.createElement("input"))',
      options: [{ forbid: [{ element: 'button' }, { element: 'input' }] }],
      errors: [
        expectForbiddenElementError1([1, 21, 1, 29]),
        expectForbiddenElementError2([1, 55, 1, 62]),
      ],
    },
  ],
});
