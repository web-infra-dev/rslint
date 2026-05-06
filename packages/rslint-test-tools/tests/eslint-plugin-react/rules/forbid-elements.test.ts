import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('forbid-elements', {} as never, {
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
      errors: [{ messageId: 'forbiddenElement' }],
    },
    {
      code: '[<Modal />, <button />]',
      options: [{ forbid: ['button', 'Modal'] }],
      errors: [
        { messageId: 'forbiddenElement' },
        { messageId: 'forbiddenElement' },
      ],
    },
    {
      code: '<dotted.component />',
      options: [{ forbid: ['dotted.component'] }],
      errors: [{ messageId: 'forbiddenElement' }],
    },
    {
      code: '<dotted.Component />',
      options: [
        {
          forbid: [{ element: 'dotted.Component', message: "that ain't cool" }],
        },
      ],
      errors: [{ messageId: 'forbiddenElement_message' }],
    },
    {
      code: '<button />',
      options: [
        { forbid: [{ element: 'button', message: 'use <Button> instead' }] },
      ],
      errors: [{ messageId: 'forbiddenElement_message' }],
    },
    {
      code: '<button><input /></button>',
      options: [{ forbid: [{ element: 'button' }, { element: 'input' }] }],
      errors: [
        { messageId: 'forbiddenElement' },
        { messageId: 'forbiddenElement' },
      ],
    },
    {
      code: '<button><input /></button>',
      options: [{ forbid: [{ element: 'button' }, 'input'] }],
      errors: [
        { messageId: 'forbiddenElement' },
        { messageId: 'forbiddenElement' },
      ],
    },
    {
      code: '<button><input /></button>',
      options: [{ forbid: ['input', { element: 'button' }] }],
      errors: [
        { messageId: 'forbiddenElement' },
        { messageId: 'forbiddenElement' },
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
      errors: [{ messageId: 'forbiddenElement_message' }],
    },
    {
      code: 'React.createElement("button", {}, child)',
      options: [{ forbid: ['button'] }],
      errors: [{ messageId: 'forbiddenElement' }],
    },
    {
      code: '[React.createElement(Modal), React.createElement("button")]',
      options: [{ forbid: ['button', 'Modal'] }],
      errors: [
        { messageId: 'forbiddenElement' },
        { messageId: 'forbiddenElement' },
      ],
    },
    {
      code: 'React.createElement(dotted.Component)',
      options: [
        {
          forbid: [{ element: 'dotted.Component', message: "that ain't cool" }],
        },
      ],
      errors: [{ messageId: 'forbiddenElement_message' }],
    },
    {
      code: 'React.createElement(dotted.component)',
      options: [{ forbid: ['dotted.component'] }],
      errors: [{ messageId: 'forbiddenElement' }],
    },
    {
      code: 'React.createElement(_comp)',
      options: [{ forbid: ['_comp'] }],
      errors: [{ messageId: 'forbiddenElement' }],
    },
    {
      code: 'React.createElement("button")',
      options: [
        { forbid: [{ element: 'button', message: 'use <Button> instead' }] },
      ],
      errors: [{ messageId: 'forbiddenElement_message' }],
    },
    {
      code: 'React.createElement("button", {}, React.createElement("input"))',
      options: [{ forbid: [{ element: 'button' }, { element: 'input' }] }],
      errors: [
        { messageId: 'forbiddenElement' },
        { messageId: 'forbiddenElement' },
      ],
    },
  ],
});
