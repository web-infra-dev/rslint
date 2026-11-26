import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';

function expectInvalidCombinationError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'invalidCombination',
    message:
      'An iframe element defines a sandbox attribute with both allow-scripts and allow-same-origin which is invalid',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectAttributeMissingError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'attributeMissing',
    message: 'An iframe element is missing a sandbox attribute',
    line,
    column,
    endLine,
    endColumn,
  };
}

function expectInvalidValueError(
  range: [line: number, column: number, endLine: number, endColumn: number],
) {
  const [line, column, endLine, endColumn] = range;
  return {
    messageId: 'invalidValue',
    message:
      'An iframe element defines a sandbox attribute with invalid value "__unknown__"',
    line,
    column,
    endLine,
    endColumn,
  };
}

const ruleTester = new RuleTester(fixtureConfig);

ruleTester.run('iframe-missing-sandbox', {
  valid: [
    {
      code: `React.createElement("iframe", /** @type {any} */ ({ sandbox: "allow-forms" }));`,
      filename: 'files/iframe-sandbox.jsx',
    },
    {
      code: `const sandbox = "sandbox"; React.createElement("iframe", { [(sandbox)]: "allow-forms" });`,
    },
    {
      code: `React.createElement("iframe", { get [(sandbox)]() {} });`,
    },
    {
      code: `React.createElement("iframe", { sandbox: ("unknown" as string) });`,
    },
    {
      code: `import { createElement } from "react"; createElement("div");`,
    },
    { code: `<div sandbox="__unknown__" />;` },
    { code: `<iframe sandbox="" />;` },
    { code: `<iframe sandbox={""} />` },
    { code: `React.createElement("iframe", { sandbox: "" });` },
    { code: `<iframe src="foo.htm" sandbox></iframe>` },
    {
      code: `React.createElement("iframe", { src: "foo.htm", sandbox: true })`,
    },
    { code: `<iframe src="foo.htm" sandbox sandbox></iframe>` },
    { code: `<iframe sandbox="allow-forms"></iframe>` },
    { code: `<iframe sandbox="allow-modals"></iframe>` },
    { code: `<iframe sandbox="allow-orientation-lock"></iframe>` },
    { code: `<iframe sandbox="allow-pointer-lock"></iframe>` },
    { code: `<iframe sandbox="allow-popups"></iframe>` },
    { code: `<iframe sandbox="allow-popups-to-escape-sandbox"></iframe>` },
    { code: `<iframe sandbox="allow-presentation"></iframe>` },
    { code: `<iframe sandbox="allow-same-origin"></iframe>` },
    { code: `<iframe sandbox="allow-scripts"></iframe>` },
    { code: `<iframe sandbox="allow-top-navigation"></iframe>` },
    {
      code: `<iframe sandbox="allow-top-navigation-by-user-activation"></iframe>`,
    },
    { code: `<iframe sandbox="allow-forms allow-modals"></iframe>` },
    {
      code: `<iframe sandbox="allow-popups allow-popups-to-escape-sandbox allow-pointer-lock allow-same-origin allow-top-navigation"></iframe>`,
    },
    { code: `React.createElement("iframe", { sandbox: "allow-forms" })` },
    { code: `React.createElement("iframe", { sandbox: "allow-modals" })` },
    {
      code: `React.createElement("iframe", { sandbox: "allow-orientation-lock" })`,
    },
    {
      code: `React.createElement("iframe", { sandbox: "allow-pointer-lock" })`,
    },
    { code: `React.createElement("iframe", { sandbox: "allow-popups" })` },
    {
      code: `React.createElement("iframe", { sandbox: "allow-popups-to-escape-sandbox" })`,
    },
    {
      code: `React.createElement("iframe", { sandbox: "allow-presentation" })`,
    },
    {
      code: `React.createElement("iframe", { sandbox: "allow-same-origin" })`,
    },
    {
      code: `React.createElement("iframe", { sandbox: "allow-scripts" })`,
    },
    {
      code: `React.createElement("iframe", { sandbox: "allow-top-navigation" })`,
    },
    {
      code: `React.createElement("iframe", { sandbox: "allow-top-navigation-by-user-activation" })`,
    },
    {
      code: `React.createElement("iframe", { sandbox: "allow-forms allow-modals" })`,
    },
    {
      code: `React.createElement("iframe", { sandbox: "allow-popups allow-popups-to-escape-sandbox allow-pointer-lock allow-same-origin allow-top-navigation" })`,
    },
  ],
  invalid: [
    {
      code: `React.createElement("iframe", { sandbox: /** @type {string} */ ("allow-scripts allow-same-origin") });`,
      filename: 'files/iframe-sandbox.jsx',
      errors: [expectInvalidCombinationError([1, 1, 1, 102])],
    },
    {
      code: `React.createElement(/** @satisfies {string} */ ("iframe"));`,
      filename: 'files/iframe-sandbox.jsx',
      errors: [expectAttributeMissingError([1, 1, 1, 59])],
    },
    {
      code: `const sandbox = "sandbox"; React.createElement("iframe", { [/** @type {string} */ (sandbox)]: "unknown" });`,
      filename: 'files/iframe-sandbox.jsx',
      errors: [
        {
          messageId: 'invalidValue',
          message:
            'An iframe element defines a sandbox attribute with invalid value "unknown"',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 107,
        },
      ],
    },
    {
      code: `React.createElement("iframe", ({ sandbox: "allow-forms" } as any));`,
      errors: [expectAttributeMissingError([1, 1, 1, 67])],
    },
    {
      code: `<iframe></iframe>;`,
      errors: [expectAttributeMissingError([1, 1, 1, 9])],
    },
    {
      code: `<iframe/>;`,
      errors: [expectAttributeMissingError([1, 1, 1, 10])],
    },
    {
      code: `React.createElement("iframe");`,
      errors: [expectAttributeMissingError([1, 1, 1, 30])],
    },
    {
      code: `React.createElement("iframe", {});`,
      errors: [expectAttributeMissingError([1, 1, 1, 34])],
    },
    {
      code: `React.createElement("iframe", null);`,
      errors: [expectAttributeMissingError([1, 1, 1, 36])],
    },
    {
      code: `<iframe sandbox="__unknown__"></iframe>`,
      errors: [expectInvalidValueError([1, 1, 1, 31])],
    },
    {
      code: `React.createElement("iframe", { sandbox: "__unknown__" })`,
      errors: [expectInvalidValueError([1, 1, 1, 58])],
    },
    {
      code: `<iframe sandbox="allow-popups __unknown__"/>`,
      errors: [expectInvalidValueError([1, 1, 1, 45])],
    },
    {
      code: `<iframe sandbox="__unknown__ allow-popups"/>`,
      errors: [expectInvalidValueError([1, 1, 1, 45])],
    },
    {
      code: `<iframe sandbox=" allow-forms __unknown__ allow-popups __unknown__  "/>`,
      errors: [
        expectInvalidValueError([1, 1, 1, 72]),
        expectInvalidValueError([1, 1, 1, 72]),
      ],
    },
    {
      code: `<iframe sandbox="allow-scripts allow-same-origin"></iframe>;`,
      errors: [expectInvalidCombinationError([1, 1, 1, 51])],
    },
    {
      code: `<iframe sandbox="allow-same-origin allow-scripts"/>;`,
      errors: [expectInvalidCombinationError([1, 1, 1, 52])],
    },
  ],
});
