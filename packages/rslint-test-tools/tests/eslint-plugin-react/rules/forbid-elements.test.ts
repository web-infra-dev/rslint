import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
import { diagnostic } from '../expectations';

function expectForbiddenElementError1(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic('forbiddenElement', '<button> is forbidden', range);
}

function expectForbiddenElementError2(range: Parameters<typeof diagnostic>[2]) {
  return diagnostic('forbiddenElement', '<input> is forbidden', range);
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
        diagnostic('forbiddenElement', '<Modal> is forbidden', [1, 3, 1, 8]),
        expectForbiddenElementError1([1, 14, 1, 20]),
      ],
    },
    {
      code: '<dotted.component />',
      options: [{ forbid: ['dotted.component'] }],
      errors: [
        diagnostic(
          'forbiddenElement',
          '<dotted.component> is forbidden',
          [1, 2, 1, 18],
        ),
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
        diagnostic(
          'forbiddenElement_message',
          "<dotted.Component> is forbidden, that ain't cool",
          [1, 2, 1, 18],
        ),
      ],
    },
    {
      code: '<button />',
      options: [
        { forbid: [{ element: 'button', message: 'use <Button> instead' }] },
      ],
      errors: [
        diagnostic(
          'forbiddenElement_message',
          '<button> is forbidden, use <Button> instead',
          [1, 2, 1, 8],
        ),
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
        diagnostic(
          'forbiddenElement_message',
          '<button> is forbidden, use <Button2> instead',
          [1, 2, 1, 8],
        ),
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
        diagnostic('forbiddenElement', '<Modal> is forbidden', [1, 22, 1, 27]),
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
        diagnostic(
          'forbiddenElement_message',
          "<dotted.Component> is forbidden, that ain't cool",
          [1, 21, 1, 37],
        ),
      ],
    },
    {
      code: 'React.createElement(dotted.component)',
      options: [{ forbid: ['dotted.component'] }],
      errors: [
        diagnostic(
          'forbiddenElement',
          '<dotted.component> is forbidden',
          [1, 21, 1, 37],
        ),
      ],
    },
    {
      code: 'React.createElement(_comp)',
      options: [{ forbid: ['_comp'] }],
      errors: [
        diagnostic('forbiddenElement', '<_comp> is forbidden', [1, 21, 1, 26]),
      ],
    },
    {
      code: 'React.createElement("button")',
      options: [
        { forbid: [{ element: 'button', message: 'use <Button> instead' }] },
      ],
      errors: [
        diagnostic(
          'forbiddenElement_message',
          '<button> is forbidden, use <Button> instead',
          [1, 21, 1, 29],
        ),
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
