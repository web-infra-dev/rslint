import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

function expectNoNamespaceError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'noNamespace',
    message:
      'React component ns:Panel must not be in a namespace, as React does not support them',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

function namespaceCallError(code: string) {
  const call = 'React.createElement("ns:B")';
  const start = code.lastIndexOf(call);
  if (start < 0)
    throw new Error('Missing namespace call in fixed source template');
  const before = code.slice(0, start);
  const line = before.split('\n').length;
  const column = start - before.lastIndexOf('\n');
  return {
    messageId: 'noNamespace',
    message: message('ns:B'),
    line,
    column,
    endLine: line,
    endColumn: column + call.length,
  };
}

const message = (name: string) =>
  `React component ${name} must not be in a namespace, as React does not support them`;

ruleTester.run('no-namespace', {
  valid: [
    {
      code: 'import { createElement } from "react"; function f() { type createElement = {}; createElement("ns:Panel"); }',
    },
    {
      code: 'const { createElement } = require?.("react"); createElement("ns:Panel");',
    },
    // ---- Upstream valid: JSX elements ----
    { code: '<testcomponent />' },
    { code: '<testComponent />' },
    { code: '<test_component />' },
    { code: '<TestComponent />' },
    { code: '<object.testcomponent />' },
    { code: '<object.testComponent />' },
    { code: '<object.test_component />' },
    { code: '<object.TestComponent />' },
    { code: '<Object.testcomponent />' },
    { code: '<Object.testComponent />' },
    { code: '<Object.test_component />' },
    { code: '<Object.TestComponent />' },

    // ---- Upstream valid: React.createElement ----
    { code: 'React.createElement("testcomponent")' },
    { code: 'React.createElement("testComponent")' },
    { code: 'React.createElement("test_component")' },
    { code: 'React.createElement("TestComponent")' },
    { code: 'React.createElement("object.testcomponent")' },
    { code: 'React.createElement("object.testComponent")' },
    { code: 'React.createElement("object.test_component")' },
    { code: 'React.createElement("object.TestComponent")' },
    { code: 'React.createElement("Object.testcomponent")' },
    { code: 'React.createElement("Object.testComponent")' },
    { code: 'React.createElement("Object.test_component")' },
    { code: 'React.createElement("Object.TestComponent")' },
    { code: 'React.createElement(null)' },
    { code: 'React.createElement(true)' },
    { code: 'React.createElement({})' },
  ],
  invalid: [
    {
      code: 'function child() { const createElement = React.createElement; } createElement("ns:Panel");',
      errors: [expectNoNamespaceError([1, 65, 1, 90])],
    },
    {
      code: 'namespace A.B.C { const createElement = React.createElement; } createElement("ns:Panel");',
      errors: [expectNoNamespaceError([1, 64, 1, 89])],
    },
    {
      code: 'class C { #createElement; f(React) { React.#createElement("ns:Panel"); } }',
      errors: [expectNoNamespaceError([1, 38, 1, 70])],
    },
    ...[
      '/* eslint-disable react/no-namespace */ React.createElement("ns:A"); /* eslint-enable react/no-namespace */ React.createElement("ns:B");',
      '/* eslint-disable */ React.createElement("ns:A"); /* eslint-enable react/no-namespace */ React.createElement("ns:B");',
      '/* eslint-disable react/no-namespace */ React.createElement("ns:A"); /* eslint-enable */ React.createElement("ns:B");',
      '/* eslint-disable react/no-namespace */\nReact.createElement("ns:A");\n/* eslint-enable react/no-namespace */\nReact.createElement("ns:B");',
    ].map((code) => ({ code, errors: [namespaceCallError(code)] })),
    {
      code: 'React.createElement("ns:A"); /* eslint-disable react/no-namespace */ React.createElement("ns:B");',
      errors: [
        {
          messageId: 'noNamespace',
          message:
            'React component ns:A must not be in a namespace, as React does not support them',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 28,
        },
      ],
    },
    // ---- Upstream invalid: lower-case namespace ----
    ...[
      'ns:testcomponent',
      'ns:testComponent',
      'ns:test_component',
      'ns:TestComponent',
    ].flatMap((name) => [
      {
        code: `<${name} />`,
        errors: [
          {
            messageId: 'noNamespace',
            message: message(name),
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: name.length + 5,
          },
        ],
      },
      {
        code: `React.createElement("${name}")`,
        errors: [
          {
            messageId: 'noNamespace',
            message: message(name),
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: name.length + 24,
          },
        ],
      },
    ]),
    // ---- Upstream invalid: upper-case namespace ----
    ...[
      'Ns:testcomponent',
      'Ns:testComponent',
      'Ns:test_component',
      'Ns:TestComponent',
    ].flatMap((name) => [
      {
        code: `<${name} />`,
        errors: [
          {
            messageId: 'noNamespace',
            message: message(name),
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: name.length + 5,
          },
        ],
      },
      {
        code: `React.createElement("${name}")`,
        errors: [
          {
            messageId: 'noNamespace',
            message: message(name),
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: name.length + 24,
          },
        ],
      },
    ]),
  ],
});
