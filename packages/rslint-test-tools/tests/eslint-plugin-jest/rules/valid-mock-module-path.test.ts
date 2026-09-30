// Upstream: https://github.com/jest-community/eslint-plugin-jest/blob/v29.16.6/src/rules/__tests__/valid-mock-module-path.test.ts
import fs from 'node:fs';
import path from 'node:path';
import { RuleTester } from '../rule-tester';

// Package specifiers resolve from the linted file, so the fixture tree provides
// upstream's module fixtures and the packages its cases name.
const root = fs.mkdtempSync(
  path.join(process.cwd(), '.valid-mock-module-path-'),
);
const archive = fs.readFileSync(
  path.resolve(
    import.meta.dirname,
    '../../../../../internal/plugins/jest/rules/valid_mock_module_path/testdata/fixtures.txtar',
  ),
  'utf8',
);
const chunks = archive.split(/^-- (.+) --\r?$/m);
if (chunks.length < 3)
  throw new Error('Missing valid-mock-module-path fixtures');
for (let i = 1; i < chunks.length; i += 2) {
  const file = path.join(root, chunks[i]);
  fs.mkdirSync(path.dirname(file), { recursive: true });
  fs.writeFileSync(file, chunks[i + 1].replace(/^\r?\n/, ''));
}
afterAll(() => fs.rmSync(root, { recursive: true, force: true }));

const filename = path.join(
  root,
  'src/rules/__tests__/valid-mock-module-path.test.ts',
);
const message = (moduleName: string) =>
  `Module path ${moduleName} does not exist or is not exported`;

const ruleTester = new RuleTester();

// Not ported: upstream's final `describe` block mocks `path.resolve` to throw
// an unexpected error code and asserts the rule rethrows it, which has no
// counterpart outside Node's resolver.
ruleTester.run('valid-mock-module-path', null as never, {
  valid: [
    { filename, code: 'jest.mock("./fixtures/module")' },
    { filename, code: 'jest.mock("./fixtures/module", () => {})' },
    { filename, code: 'jest.mock()' },
    { filename, code: 'jest.doMock("./fixtures/module", () => {})' },
    { filename, code: 'describe("foo", () => {});' },
    { filename, code: 'jest.doMock("./fixtures/module")' },
    { filename, code: 'jest.mock("./fixtures/module/foo.ts")' },
    { filename, code: 'jest.doMock("./fixtures/module/foo.ts")' },
    { filename, code: 'jest.mock("./fixtures/module/foo.js")' },
    { filename, code: 'jest.doMock("./fixtures/module/foo.js")' },
    { filename, code: 'jest.mock("eslint")' },
    { filename, code: 'jest.doMock("eslint")' },
    { filename, code: 'jest.mock("child_process")' },
    { filename, code: 'jest.mock(() => {})' },
    {
      filename,
      code: 'const a = "../module/does/not/exist";\njest.mock(a);',
    },
    { filename, code: 'jest.mock("./fixtures/module/jsx/foo")' },
    { filename, code: 'jest.mock("./fixtures/module/tsx/foo")' },
    {
      filename,
      code: 'jest.mock("./fixtures/module/tsx/foo")',
      options: [{ moduleFileExtensions: ['.jsx'] }],
    },
    {
      filename,
      code: 'jest.mock("./fixtures/module/bar")',
      options: [{ moduleFileExtensions: ['.json'] }],
    },
    {
      filename,
      code: 'jest.mock("./fixtures/module/bar")',
      options: [{ moduleFileExtensions: ['.css'] }],
    },
    {
      filename,
      code: 'jest.mock("./fixtures/module/tsx/foo", undefined, { virtual: false })',
    },
    {
      filename,
      code: 'jest.doMock("./fixtures/module/tsx/foo", undefined, { virtual: false })',
    },
    {
      filename,
      code: 'jest.doMock("./fixtures/module/tsx/foo", undefined, { ...{} })',
    },
    {
      filename,
      code: 'jest.doMock("./fixtures/module/tsx/foo", undefined, { virtual: [] })',
    },
    {
      filename,
      code: 'jest.mock("../module/does/not/exist", undefined, { virtual: true })',
    },
    {
      filename,
      code: 'jest.doMock("../module/does/not/exist", undefined, { virtual: true })',
    },
    {
      filename,
      code: 'jest.doMock("../module/does/not/exist", undefined, { "virtual": true })',
    },
    {
      filename,
      code: 'jest.doMock("../module/does/not/exist", undefined, { ["virtual"]: true })',
    },
    {
      filename,
      code: 'jest.doMock("../module/does/not/exist", undefined, { [`virtual`]: true })',
    },
    {
      // jest only cares if the value is truthy...
      filename,
      code: 'jest.doMock("../module/does/not/exist", undefined, { virtual: 1 })',
    },
    {
      // jest only cares if the value is truthy...
      filename,
      code: 'jest.doMock("../module/does/not/exist", undefined, { virtual: "yes" })',
    },
  ],
  invalid: [
    {
      filename,
      code: "jest.mock('../module/does/not/exist')",
      errors: [{ message: message("'../module/does/not/exist'") }],
    },
    {
      filename,
      code: 'jest.mock("../file/does/not/exist.ts")',
      errors: [{ message: message('"../file/does/not/exist.ts"') }],
    },
    {
      filename,
      code: 'jest.mock("./fixtures/module/foo.jsx")',
      options: [{ moduleFileExtensions: ['.tsx'] }],
      errors: [{ message: message('"./fixtures/module/foo.jsx"') }],
    },
    {
      filename,
      code: 'jest.mock("./fixtures/module/foo.jsx")',
      options: [{ moduleFileExtensions: undefined }],
      errors: [{ message: message('"./fixtures/module/foo.jsx"') }],
    },
    {
      filename,
      code: 'jest.mock("@doesnotexist/module")',
      errors: [{ message: message('"@doesnotexist/module"') }],
    },
    // the imported file does not exist, but since it's not in `exports`
    // a ERR_PACKAGE_PATH_NOT_EXPORTED error will be thrown instead
    {
      filename,
      code: 'jest.mock("jest-util/build/isInteractive")',
      errors: [{ message: message('"jest-util/build/isInteractive"') }],
    },
    // the imported file does exist, but since it's not in `exports`
    // a ERR_PACKAGE_PATH_NOT_EXPORTED error will be thrown instead
    {
      filename,
      code: 'jest.mock("jackspeak/dist/commonjs/parse-args.js")',
      errors: [{ message: message('"jackspeak/dist/commonjs/parse-args.js"') }],
    },
    {
      filename,
      code: "jest.mock('../module/does/not/exist', undefined, {})",
      errors: [{ message: message("'../module/does/not/exist'") }],
    },
    {
      filename,
      code: "jest.mock('../module/does/not/exist', undefined, { assumeExists: true })",
      errors: [{ message: message("'../module/does/not/exist'") }],
    },
    {
      filename,
      code: "jest.mock('../module/does/not/exist', undefined, { virtual: false })",
      errors: [{ message: message("'../module/does/not/exist'") }],
    },
    {
      filename,
      code: "jest.doMock('../module/does/not/exist', undefined, { virtual: false })",
      errors: [{ message: message("'../module/does/not/exist'") }],
    },
    {
      filename,
      code: "jest.doMock('../module/does/not/exist', undefined, { virtual: 0 })",
      errors: [{ message: message("'../module/does/not/exist'") }],
    },
    {
      filename,
      code: "const virtual = false;\n\njest.doMock('../module/does/not/exist', undefined, { virtual })",
      errors: [{ message: message("'../module/does/not/exist'") }],
    },
    {
      filename,
      code: "const virtual = true;\n\njest.doMock('../module/does/not/exist', undefined, { virtual })",
      errors: [{ message: message("'../module/does/not/exist'") }],
    },
    {
      filename,
      code: "const prop = 'virtual';\n\njest.doMock('../module/does/not/exist', undefined, { [prop]: true })",
      errors: [{ message: message("'../module/does/not/exist'") }],
    },
    {
      // we don't attempt to resolve the result of object spreads
      filename,
      code: "jest.doMock('../module/does/not/exist', undefined, { ...{ virtual: true } })",
      errors: [{ message: message("'../module/does/not/exist'") }],
    },
  ],
});
