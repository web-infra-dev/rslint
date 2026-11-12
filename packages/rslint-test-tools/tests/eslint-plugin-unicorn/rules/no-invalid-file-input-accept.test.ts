// Upstream: eslint-plugin-unicorn v79.0.0 test/no-invalid-file-input-accept.js and rule docs.
import { RuleTester, type TestCase } from '@rslint/rule-tester';
import { test } from 'rstack/test';

const name = 'no-invalid-file-input-accept';
const invalidAcceptPrefix = 'Invalid file input `accept` value: ';
const staticAcceptMessage = 'The file input `accept` value must be static.';
const missingAcceptMessage =
  invalidAcceptPrefix + 'The `accept` attribute must have a value.';
const emptyAcceptMessage =
  invalidAcceptPrefix + 'Empty entries are not allowed.';
const wildcardAcceptMessage =
  invalidAcceptPrefix +
  'Wildcard MIME types must be `audio/*`, `image/*`, or `video/*`.';

const normalizeMessage = (value: string) =>
  `Prefer \`${value}\` as the file input \`accept\` value.`;

function acceptCase(
  code: string,
  column: number,
  endColumn: number,
  id: string,
  message: string,
  output?: string,
): TestCase {
  return {
    code,
    output,
    errors: [
      {
        messageId: `${name}/${id}`,
        message,
        line: 1,
        column,
        endLine: 1,
        endColumn,
      },
    ],
  };
}

new RuleTester({ plugin: 'unicorn', filename: 'input.jsx' }).run(name, {
  valid: [
    // jsx.snapshot valid cases.
    '<input type="file" accept="image/png" />',
    '<input type="file" accept=".png" />',
    '<input type="file" accept="image/png, .png" />',
    '<input type="file" accept=".tar.gz" />',
    '<input type="file" accept="audio/*" />',
    '<input type="file" accept="video/*" />',
    '<input type="file" accept="image/*" />',
    '<input type="file" accept="audio/*, video/*" />',
    '<input type="FILE" accept="image/png" />',
    '<input type={"file"} accept="image/png" />',
    'const type = "file"; <input type={type} accept="image/png" />',
    'const accept = "image/png"; <input type="file" accept={accept} />',
    '<input accept="image/png" />',
    '<input type="text" accept="image/png" />',
    '<input type={type} accept="image/png" />',
    '<Input type="file" accept="image/jpg" />',
    '<input type="file" Accept="image/jpg" />',
    '<input Type="file" accept="image/jpg" />',
    '<Foo.input type="file" accept="image/jpg" />',
    '<input type="file" xlink:accept="image/jpg" />',
    // Documentation example. Its other correct JSX examples are above.
    '<input type="file" accept="image/jpeg" />',
  ],
  invalid: [
    // jsx.snapshot invalid cases, including the incorrect JSX examples in the docs.
    acceptCase(
      'const modes = new Set(["foo"]); modes.clear(); <input type="file" accept={modes.size ? "bad" : "image/jpeg"} />',
      74,
      109,
      'static',
      staticAcceptMessage,
    ),
    acceptCase(
      'const modes = new Set(["foo"]); modes.clear(); <input type="file" accept={(modes.size && "image/jpeg") || value} />',
      74,
      113,
      'static',
      staticAcceptMessage,
    ),
    acceptCase(
      'const object = {value: true}; Object.defineProperty(object, "value", {get() { return false; }}); <input type="file" accept={object.value ? "image/jpeg" : value} />',
      124,
      161,
      'static',
      staticAcceptMessage,
    ),
    acceptCase(
      '<input type="file" accept />',
      20,
      26,
      'invalid',
      missingAcceptMessage,
    ),
    acceptCase(
      '<input type="file" accept={allowedTypes} />',
      27,
      41,
      'static',
      staticAcceptMessage,
    ),
    acceptCase(
      '<input type="file" accept={`image/${type}`} />',
      27,
      44,
      'static',
      staticAcceptMessage,
    ),
    acceptCase(
      '<input type="file" accept="image/jpg" />',
      27,
      38,
      'normalize',
      normalizeMessage('image/jpeg'),
      '<input type="file" accept="image/jpeg" />',
    ),
    acceptCase(
      '<input type="file" accept="image/svg" />',
      27,
      38,
      'normalize',
      normalizeMessage('image/svg+xml'),
      '<input type="file" accept="image/svg+xml" />',
    ),
    acceptCase(
      '<input type="file" accept="png" />',
      27,
      32,
      'normalize',
      normalizeMessage('.png'),
      '<input type="file" accept=".png" />',
    ),
    acceptCase(
      '<input type="file" accept=".PNG" />',
      27,
      33,
      'normalize',
      normalizeMessage('.png'),
      '<input type="file" accept=".png" />',
    ),
    acceptCase(
      '<input type="file" accept=".png, .PNG" />',
      27,
      39,
      'normalize',
      normalizeMessage('.png'),
      '<input type="file" accept=".png" />',
    ),
    acceptCase(
      '<input type="file" accept="" />',
      27,
      29,
      'invalid',
      emptyAcceptMessage,
    ),
    acceptCase(
      '<input type="file" accept="IMAGE/*" />',
      27,
      36,
      'normalize',
      normalizeMessage('image/*'),
      '<input type="file" accept="image/*" />',
    ),
    acceptCase(
      '<input type="file" accept="TEXT/*" />',
      27,
      35,
      'invalid',
      wildcardAcceptMessage,
    ),
    acceptCase(
      '<input type="file" accept="*/plain" />',
      27,
      36,
      'invalid',
      wildcardAcceptMessage,
    ),
    acceptCase(
      '<input type="file" accept="not-real-mime-type/*" />',
      27,
      49,
      'invalid',
      wildcardAcceptMessage,
    ),
    acceptCase(
      '<input type="file" accept="image/png," />',
      27,
      39,
      'invalid',
      emptyAcceptMessage,
    ),
    acceptCase(
      '<input type="file" accept="image/png,, .png" />',
      27,
      45,
      'invalid',
      emptyAcceptMessage,
    ),
    acceptCase(
      '<input type="file" accept="IMAGE/PNG,.PNG, image/png" />',
      27,
      54,
      'normalize',
      normalizeMessage('image/png, .png'),
      '<input type="file" accept="image/png, .png" />',
    ),
    acceptCase(
      '<input type="file" accept="image/png,.png" />',
      27,
      43,
      'normalize',
      normalizeMessage('image/png, .png'),
      '<input type="file" accept="image/png, .png" />',
    ),
    acceptCase(
      '<input type="file" accept="image/png; charset=utf-8" />',
      27,
      53,
      'normalize',
      normalizeMessage('image/png'),
      '<input type="file" accept="image/png" />',
    ),
    acceptCase(
      '<input type="file" accept={"IMAGE/PNG"} />',
      27,
      40,
      'normalize',
      normalizeMessage('image/png'),
    ),
    acceptCase(
      '<input type="file" accept=<div /> />',
      27,
      34,
      'static',
      staticAcceptMessage,
    ),

    // jsx cases with explicit autofix expectations.
    acceptCase(
      '<input type="file" accept="image/x-icon" />',
      27,
      41,
      'normalize',
      normalizeMessage('image/vnd.microsoft.icon'),
      '<input type="file" accept="image/vnd.microsoft.icon" />',
    ),
    acceptCase(
      '<input type="file" accept="application/x-rar-compressed" />',
      27,
      57,
      'normalize',
      normalizeMessage('application/vnd.rar'),
      '<input type="file" accept="application/vnd.rar" />',
    ),
    acceptCase(
      '<input type="file" accept="application/x-zip-compressed" />',
      27,
      57,
      'normalize',
      normalizeMessage('application/zip'),
      '<input type="file" accept="application/zip" />',
    ),
    acceptCase(
      '<input type="file" accept="image/x-icon, image/vnd.microsoft.icon, application/x-zip-compressed" />',
      27,
      97,
      'normalize',
      normalizeMessage('image/vnd.microsoft.icon, application/zip'),
      '<input type="file" accept="image/vnd.microsoft.icon, application/zip" />',
    ),
    acceptCase(
      '<input type="file" accept="image/jpg; charset=utf-8" />',
      27,
      53,
      'normalize',
      normalizeMessage('image/jpeg'),
      '<input type="file" accept="image/jpeg" />',
    ),
    acceptCase(
      '<input type="file" accept="image/svg; charset=utf-8" />',
      27,
      53,
      'normalize',
      normalizeMessage('image/svg+xml'),
      '<input type="file" accept="image/svg+xml" />',
    ),
  ],
});

// The Go upstream suite retains the individual HTML fixtures.
test.skip('html.snapshot: requires html-eslint HTML nodes', () => {});
test.skip('htmlTemplates.snapshot: requires html-eslint HTML nodes', () => {});
test.skip('htmlPlugin: requires html-eslint HTML nodes', () => {});
test.skip('html: requires html-eslint HTML nodes', () => {});
test.skip('HTML documentation: requires html-eslint HTML nodes', () => {});
