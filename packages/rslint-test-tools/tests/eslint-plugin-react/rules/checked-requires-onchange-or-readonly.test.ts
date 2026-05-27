import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('checked-requires-onchange-or-readonly', {} as never, {
  valid: [
    { code: `<input type="checkbox" />` },
    { code: `<input type="checkbox" onChange={noop} />` },
    { code: `<input type="checkbox" readOnly />` },
    { code: `<input type="checkbox" checked onChange={noop} />` },
    { code: `<input type="checkbox" checked={true} onChange={noop} />` },
    { code: `<input type="checkbox" checked={false} onChange={noop} />` },
    { code: `<input type="checkbox" checked readOnly />` },
    { code: `<input type="checkbox" checked={true} readOnly />` },
    { code: `<input type="checkbox" checked={false} readOnly />` },
    { code: `<input type="checkbox" defaultChecked />` },
    { code: `React.createElement('input')` },
    { code: `React.createElement('input', { checked: true, onChange: noop })` },
    {
      code: `React.createElement('input', { checked: false, onChange: noop })`,
    },
    { code: `React.createElement('input', { checked: true, readOnly: true })` },
    {
      code: `React.createElement('input', { checked: true, onChange: noop, readOnly: true })`,
    },
    {
      code: `React.createElement('input', { checked: foo, onChange: noop, readOnly: true })`,
    },
    {
      code: `<input type="checkbox" checked />`,
      options: [{ ignoreMissingProperties: true }],
    },
    {
      code: `<input type="checkbox" checked={true} />`,
      options: [{ ignoreMissingProperties: true }],
    },
    {
      code: `<input type="checkbox" onChange={noop} checked defaultChecked />`,
      options: [{ ignoreExclusiveCheckedAttribute: true }],
    },
    {
      code: `<input type="checkbox" onChange={noop} checked={true} defaultChecked />`,
      options: [{ ignoreExclusiveCheckedAttribute: true }],
    },
    {
      code: `<input type="checkbox" onChange={noop} checked defaultChecked />`,
      options: [
        {
          ignoreMissingProperties: true,
          ignoreExclusiveCheckedAttribute: true,
        },
      ],
    },
    { code: `<span/>` },
    { code: `React.createElement('span')` },
    { code: `(()=>{})()` },
  ],
  invalid: [
    {
      code: `<input type="radio" checked />`,
      errors: [{ messageId: 'missingProperty' }],
    },
    {
      code: `<input type="radio" checked={true} />`,
      errors: [{ messageId: 'missingProperty' }],
    },
    {
      code: `<input type="checkbox" checked />`,
      errors: [{ messageId: 'missingProperty' }],
    },
    {
      code: `<input type="checkbox" checked={true} />`,
      errors: [{ messageId: 'missingProperty' }],
    },
    {
      code: `<input type="checkbox" checked={condition ? true : false} />`,
      errors: [{ messageId: 'missingProperty' }],
    },
    {
      code: `<input type="checkbox" checked defaultChecked />`,
      errors: [
        { messageId: 'exclusiveCheckedAttribute' },
        { messageId: 'missingProperty' },
      ],
    },
    {
      code: `React.createElement("input", { checked: false })`,
      errors: [{ messageId: 'missingProperty' }],
    },
    {
      code: `React.createElement("input", { checked: true, defaultChecked: true })`,
      errors: [
        { messageId: 'exclusiveCheckedAttribute' },
        { messageId: 'missingProperty' },
      ],
    },
    {
      code: `<input type="checkbox" checked defaultChecked />`,
      options: [{ ignoreMissingProperties: true }],
      errors: [{ messageId: 'exclusiveCheckedAttribute' }],
    },
    {
      code: `<input type="checkbox" checked defaultChecked />`,
      options: [{ ignoreExclusiveCheckedAttribute: true }],
      errors: [{ messageId: 'missingProperty' }],
    },
    {
      code: `<input type="checkbox" checked defaultChecked />`,
      options: [
        {
          ignoreMissingProperties: false,
          ignoreExclusiveCheckedAttribute: false,
        },
      ],
      errors: [
        { messageId: 'exclusiveCheckedAttribute' },
        { messageId: 'missingProperty' },
      ],
    },
  ],
});
