import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const withBrowserGlobals = (code: string) => ({
  code,
  filename: 'file.mjs',
  languageOptions: {
    // The upstream test harness enables browser and Node globals by default.
    globals: {
      document: 'readonly',
      global: 'readonly',
      self: 'readonly',
      window: 'readonly',
    } as const,
  },
});
const invalid = (code: string) => ({
  ...withBrowserGlobals(code),
  errors: [
    {
      messageId: 'no-document-cookie',
      message: 'Do not use `document.cookie` directly.',
    },
  ],
});

ruleTester.run('no-document-cookie', null as never, {
  valid: [
    withBrowserGlobals('document.cookie'),
    withBrowserGlobals('const foo = document.cookie'),
    withBrowserGlobals('foo = document.cookie'),
    withBrowserGlobals('foo = document?.cookie'),
    withBrowserGlobals('foo = document.cookie + ";foo=bar"'),
    withBrowserGlobals('delete document.cookie'),
    withBrowserGlobals('if (document.cookie.includes("foo")){}'),
    withBrowserGlobals('Object.assign(document, {cookie: "foo=bar"})'),
    withBrowserGlobals('document[CONSTANTS_COOKIE] = "foo=bar"'),
    withBrowserGlobals('document[cookie] = "foo=bar"'),
    withBrowserGlobals(
      'const CONSTANTS_COOKIE = "cookie";\ndocument[CONSTANTS_COOKIE] = "foo=bar";',
    ),
  ],
  invalid: [
    invalid('document.cookie = "foo=bar"'),
    invalid('document.cookie += ";foo=bar"'),
    invalid('document.cookie = document.cookie + ";foo=bar"'),
    invalid('document.cookie &&= true'),
    invalid('document["coo" + "kie"] = "foo=bar"'),
    invalid('foo = document.cookie = "foo=bar"'),
    invalid('var doc = document; doc.cookie = "foo=bar"'),
    invalid('let doc = document; doc.cookie = "foo=bar"'),
    invalid('const doc = globalThis.document; doc.cookie = "foo=bar"'),
    invalid('window.document.cookie = "foo=bar"'),
  ],
});
