import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

function expectMissingPropertyError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'missingProperty',
    message: '`checked` should be used with either `onChange` or `readOnly`.',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectExclusiveCheckedAttributeError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'exclusiveCheckedAttribute',
    message: 'Use either `checked` or `defaultChecked`, but not both.',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('checked-requires-onchange-or-readonly', {
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
      errors: [expectMissingPropertyError([1, 1, 1, 31])],
    },
    {
      code: `<input type="radio" checked={true} />`,
      errors: [expectMissingPropertyError([1, 1, 1, 38])],
    },
    {
      code: `<input type="checkbox" checked />`,
      errors: [expectMissingPropertyError([1, 1, 1, 34])],
    },
    {
      code: `<input type="checkbox" checked={true} />`,
      errors: [expectMissingPropertyError([1, 1, 1, 41])],
    },
    {
      code: `<input type="checkbox" checked={condition ? true : false} />`,
      errors: [expectMissingPropertyError([1, 1, 1, 61])],
    },
    {
      code: `<input type="checkbox" checked defaultChecked />`,
      errors: [
        expectExclusiveCheckedAttributeError([1, 1, 1, 49]),
        expectMissingPropertyError([1, 1, 1, 49]),
      ],
    },
    {
      code: `React.createElement("input", { checked: false })`,
      errors: [expectMissingPropertyError([1, 1, 1, 49])],
    },
    {
      code: `React.createElement("input", { checked: true, defaultChecked: true })`,
      errors: [
        expectExclusiveCheckedAttributeError([1, 1, 1, 70]),
        expectMissingPropertyError([1, 1, 1, 70]),
      ],
    },
    {
      code: `<input type="checkbox" checked defaultChecked />`,
      options: [{ ignoreMissingProperties: true }],
      errors: [expectExclusiveCheckedAttributeError([1, 1, 1, 49])],
    },
    {
      code: `<input type="checkbox" checked defaultChecked />`,
      options: [{ ignoreExclusiveCheckedAttribute: true }],
      errors: [expectMissingPropertyError([1, 1, 1, 49])],
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
        expectExclusiveCheckedAttributeError([1, 1, 1, 49]),
        expectMissingPropertyError([1, 1, 1, 49]),
      ],
    },
  ],
});
