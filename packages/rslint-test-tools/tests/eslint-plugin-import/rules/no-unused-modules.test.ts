// Upstream: eslint-plugin-import v2.32.0 tests and documentation.
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-unused-modules.js
//
// Upstream keeps one module-level usage index across all of its test runs and
// updates it from each linted file, so many cases depend on the text an earlier
// case linted into another file. rslint builds the index once per request from
// one source view; each group passes that earlier file text as an API overlay
// instead. Upstream asserts messages only; positions are TSESTree's.
import fs from 'node:fs';
import path from 'node:path';
import { lint } from '@rslint/core/internal';

const root = fs.mkdtempSync(path.join(process.cwd(), '.no-unused-modules-'));
const archive = fs.readFileSync(
  path.resolve(
    import.meta.dirname,
    '../../../../../internal/plugins/import/rules/no_unused_modules/testdata/no_unused_modules.txtar',
  ),
  'utf8',
);
const chunks = archive.split(/^-- (.+) --\r?$/m);
if (chunks.length < 3) throw new Error('Missing upstream fixtures');
for (let i = 1; i < chunks.length; i += 2) {
  const target = path.join(root, chunks[i]);
  fs.mkdirSync(path.dirname(target), { recursive: true });
  fs.writeFileSync(target, chunks[i + 1].replace(/^\r?\n/, ''));
}
afterAll(() => fs.rmSync(root, { recursive: true, force: true }));

const file = (relative: string) => path.join(root, relative);
const fixture = (relative: string) =>
  file(path.join('no-unused-modules', relative));

type Expected = [message: string, line: number, column: number];
interface Case {
  filename: string;
  code: string;
  options?: Record<string, unknown>;
  settings?: Record<string, unknown>;
  errors?: Expected[];
}

const unused = (name: string, line: number, column: number): Expected => [
  `exported declaration '${name}' not used within other modules`,
  line,
  column,
];
const noExports = (line: number, column: number): Expected => [
  'No exports found',
  line,
  column,
];

const missingExportsOptions = { missingExports: true };
const unusedExportsOptions = {
  unusedExports: true,
  src: [fixture('**/*.js')],
  ignoreExports: [fixture('*ignored*.js')],
};
const unusedExportsTypescriptOptions = {
  unusedExports: true,
  src: [fixture('typescript')],
};
const unusedExportsTypescriptIgnoreUnusedTypesOptions = {
  unusedExports: true,
  ignoreUnusedTypeExports: true,
  src: [fixture('typescript')],
};
const unusedExportsJsxOptions = { unusedExports: true, src: [fixture('jsx')] };
// config/typescript.js and config/react.js
const typescriptExtensions = ['.ts', '.cts', '.mts', '.tsx'];
const typescriptSettings = {
  'import/extensions': [...typescriptExtensions, '.js', '.jsx', '.mjs', '.cjs'],
  'import/external-module-folders': ['node_modules', 'node_modules/@types'],
  'import/parsers': { '@typescript-eslint/parser': typescriptExtensions },
  'import/resolver': {
    node: {
      extensions: [...typescriptExtensions, '.js', '.jsx', '.mjs', '.cjs'],
    },
  },
};
const jsxSettings = { 'import/extensions': ['.js', '.jsx'] };

async function run(testCase: Case, overlays: Record<string, string>) {
  const result = await lint({
    config: [
      {
        plugins: ['import'],
        languageOptions: {
          parserOptions: { project: false, projectService: false },
        },
        settings: testCase.settings ?? {},
        rules: {
          'import/no-unused-modules': testCase.options
            ? ['error', testCase.options]
            : 'error',
        },
      },
    ],
    configDirectory: root,
    workingDirectory: root,
    files: [testCase.filename],
    fileContents: { ...overlays, [testCase.filename]: testCase.code },
  });
  assert.strictEqual(result.fileCount, 1, testCase.filename);
  const actual = result.diagnostics.map((diagnostic) => {
    assert.strictEqual(diagnostic.ruleName, 'import/no-unused-modules');
    return [
      diagnostic.message,
      diagnostic.range.start.line,
      diagnostic.range.start.column,
    ];
  });
  assert.deepStrictEqual(
    actual,
    testCase.errors ?? [],
    `${testCase.filename}\n${testCase.code}`,
  );
}

function group(
  name: string,
  cases: Case[],
  overlays: Record<string, string> = {},
) {
  test(name, async () => {
    for (const testCase of cases) await run(testCase, overlays);
  });
}

const file0 = fixture('file-0.js');
// file-0.js as the "tests for exports" invalid case lints it. Unlike the file
// on disk, it no longer imports `q` from file-q.js.
const file0Code = `
        import eslint from 'eslint'
        import fileA from './file-a'
        import { b } from './file-b'
        import { c1, c2 } from './file-c'
        import { d } from './file-d'
        import { e } from './file-e'
        import { e2 } from './file-e'
        import { h2 } from './file-h'
        import * as l from './file-l'
        export * from './file-n'
        export { default, o0, o3 } from './file-o'
        export { p } from './file-p'
        import s from './file-s'
      `;
const js = (name: string, code: string, errors?: Expected[]): Case => ({
  filename: fixture(name),
  code,
  options: unusedExportsOptions,
  errors,
});

describe('import/no-unused-modules', () => {
  group('tests for missing exports', [
    ...[
      'export default () => 1',
      'export const a = 1',
      'const a = 1; export { a }',
      'function a() { return true }; export { a }',
      'const a = 1; const b = 2; export { a, b }',
      'const a = 1; export default a',
      'export class Foo {}',
      'export const [foobar] = [];',
      'export const [foobar] = foobarFactory();',
      `
        export default function NewComponent () {
          return 'I am new component'
        }
      `,
    ].map((code) => ({
      filename: file('missing.js'),
      code,
      options: missingExportsOptions,
    })),
    {
      filename: file('missing.js'),
      code: 'export default function noOptions() {}',
    },
    {
      filename: file('missing.js'),
      code: 'const a = 1',
      options: missingExportsOptions,
      errors: [noExports(1, 1)],
    },
    {
      filename: file('missing.js'),
      code: '/* const a = 1 */',
      options: missingExportsOptions,
      errors: [noExports(1, 18)],
    },
  ]);

  group('tests for exports', [
    js('file-a.js', 'import { o2 } from "./file-o";export default () => 12'),
    js('file-b.js', 'export const b = 2'),
    js(
      'file-c.js',
      'const c1 = 3; function c2() { return 3 }; export { c1, c2 }',
    ),
    js('file-d.js', 'export function d() { return 4 }'),
    js('file-q.js', 'export class q { q0() {} }'),
    js('file-e.js', 'const e0 = 5; export { e0 as e }'),
    js(
      'file-l.js',
      'const l0 = 5; const l = 10; export { l0 as l1, l }; export default () => {}',
    ),
    js(
      'file-o.js',
      'const o0 = 0; const o1 = 1; export { o0, o1 as o2 }; export default () => {}',
    ),
    js(
      'file-o.js',
      `
        export const [o0, o2] = createLoadingAndErrorSelectors(
          AUTH_USER
        );
      `,
    ),
    js('file-0.js', file0Code, [
      unused('default', 12, 18),
      unused('o0', 12, 27),
      unused('o3', 12, 31),
      unused('p', 13, 18),
    ]),
    js(
      'file-n.js',
      `const n0 = 'n0'; const n1 = 42; export { n0, n1 }; export default () => {}`,
      [unused('default', 1, 52)],
    ),
  ]);

  group(
    'test for unused exports',
    [
      js('file-f.js', 'export default () => 13', [unused('default', 1, 1)]),
      js('file-g.js', 'export const g = 2', [unused('g', 1, 1)]),
      js(
        'file-h.js',
        'const h1 = 3; function h2() { return 3 }; const h3 = true; export { h1, h2, h3 }',
        [unused('h1', 1, 69)],
      ),
      js(
        'file-i.js',
        'const i1 = 3; function i2() { return 3 }; export { i1, i2 }',
        [unused('i1', 1, 52), unused('i2', 1, 56)],
      ),
      js('file-j.js', 'export function j() { return 4 }', [unused('j', 1, 1)]),
      js('file-q.js', 'export class q { q0() {} }', [unused('q', 1, 1)]),
      js('file-k.js', 'const k0 = 5; export { k0 as k }', [unused('k', 1, 24)]),
    ],
    { [file0]: file0Code },
  );

  group('dynamic imports', [
    js(
      'exports-for-dynamic-js.js',
      `
            export const a = 10
            export const b = 20
            export const c = 30
            const d = 40
            export default d
            `,
    ),
    js(
      'exports-for-dynamic-js-2.js',
      `
        export const a = 10
        export const b = 20
        export const c = 30
        const d = 40
        export default d
        `,
      [
        unused('a', 2, 9),
        unused('b', 3, 9),
        unused('c', 4, 9),
        unused('default', 6, 9),
      ],
    ),
    {
      filename: fixture('typescript/exports-for-dynamic-ts.ts'),
      options: unusedExportsTypescriptOptions,
      settings: typescriptSettings,
      code: `
            export const ts_a = 10
            export const ts_b = 20
            export const ts_c = 30
            const ts_d = 40
            export default ts_d
            `,
    },
    {
      filename: file('unused-modules-reexport-crash/src/index.tsx'),
      settings: typescriptSettings,
      options: { unusedExports: true, ignoreExports: ['**/magic/**'] },
      code: `
        import App from './App';
      `,
    },
  ]);

  group('test for export from', [
    js('file-s.js', `export { default } from './file-o'`),
    js(
      'file-j.js',
      `export { k } from '${fixture('file-k.js').replace(/\\/g, '/')}'`,
      [unused('k', 1, 10)],
    ),
  ]);
  group(
    'export from usage',
    [js('file-k.js', 'const k0 = 5; export { k0 as k }')],
    {
      [fixture('file-j.js')]:
        `export { k } from '${fixture('file-k.js').replace(/\\/g, '/')}'`,
    },
  );

  group('test for ignored files', [
    js('file-ignored-a.js', 'export default () => 14'),
    js('file-ignored-b.js', 'export const b = 2'),
    js(
      'file-ignored-c.js',
      'const c1 = 3; function c2() { return 3 }; export { c1, c2 }',
    ),
    js('file-ignored-d.js', 'export function d() { return 4 }'),
    js('file-ignored-e.js', 'const f = 5; export { f as e }'),
    js(
      'file-ignored-l.js',
      'const l0 = 5; const l = 10; export { l0 as l1, l }; export default () => {}',
    ),
  ]);

  const importerGroup = (name: string, importer: string, cases: Case[]) =>
    group(name, [js('file-0.js', importer), ...cases], { [file0]: importer });
  const fileF = fixture('file-f.js');
  const fileG = fixture('file-g.js');
  const fileM = fixture('file-m.js');
  const fileMCode =
    'const m0 = 5; const m = 10; export { m0 as m1, m }; export default () => {}';
  const allMUnused = [
    unused('m1', 1, 38),
    unused('m', 1, 48),
    unused('default', 1, 53),
  ];

  importerGroup(
    'add named import for file with default export',
    `import { f } from '${fileF.replace(/\\/g, '/')}'`,
    [js('file-f.js', 'export default () => 15', [unused('default', 1, 1)])],
  );
  importerGroup(
    'add default import for file with default export',
    `import f from '${fileF.replace(/\\/g, '/')}'`,
    [js('file-f.js', 'export default () => 16')],
  );
  importerGroup(
    'add default import for file with named export',
    `import g from '${fileG.replace(/\\/g, '/')}';import {h} from '${fixture('file-gg.js').replace(/\\/g, '/')}'`,
    [js('file-g.js', 'export const g = 2', [unused('g', 1, 1)])],
  );
  importerGroup(
    'add named import for file with named export',
    `import { g } from '${fileG.replace(/\\/g, '/')}'; import eslint from 'eslint'`,
    [js('file-g.js', 'export const g = 2')],
  );
  importerGroup(
    'add different named import for file with named export',
    `import { c } from '${fixture('file-b.js').replace(/\\/g, '/')}'`,
    [js('file-b.js', 'export const b = 2', [unused('b', 1, 1)])],
  );
  importerGroup(
    'add renamed named import for file with named export',
    `import { g as g1 } from '${fileG.replace(/\\/g, '/')}'; import eslint from 'eslint'`,
    [js('file-g.js', 'export const g = 2')],
  );
  importerGroup(
    'add different renamed named import for file with named export',
    `import { g1 as g } from '${fileG.replace(/\\/g, '/')}'`,
    [js('file-g.js', 'export const g = 2', [unused('g', 1, 1)])],
  );
  importerGroup(
    'remove default import for file with default export',
    `import { a1, a2 } from '${fixture('file-a.js').replace(/\\/g, '/')}'`,
    [js('file-a.js', 'export default () => 17', [unused('default', 1, 1)])],
  );
  group('add namespace import for file with unused exports', [
    js('file-m.js', fileMCode, allMUnused),
  ]);
  importerGroup(
    'namespace import',
    `import * as m from '${fileM.replace(/\\/g, '/')}'; import unknown from 'unknown-module'`,
    [js('file-m.js', fileMCode)],
  );
  importerGroup(
    'remove all exports',
    `/* import * as m from '${fileM.replace(/\\/g, '/')}' */`,
    [js('file-m.js', fileMCode, allMUnused)],
  );
  importerGroup(
    'export star',
    `export * from '${fileM.replace(/\\/g, '/')}';`,
    [js('file-m.js', fileMCode, [unused('default', 1, 53)])],
  );
  group(
    'named re-export',
    [
      js('file-0.js', `export { m1, m} from '${fileM.replace(/\\/g, '/')}';`, [
        unused('m1', 1, 10),
        unused('m', 1, 14),
      ]),
      js('file-m.js', fileMCode, [unused('default', 1, 53)]),
    ],
    { [file0]: `export { m1, m} from '${fileM.replace(/\\/g, '/')}';` },
  );
  // The upstream TODO valid case for this re-export is commented out upstream.
  group(
    'default re-export',
    [
      js(
        'file-0.js',
        `export { default, m1 } from '${fileM.replace(/\\/g, '/')}';`,
        [unused('default', 1, 10), unused('m1', 1, 19)],
      ),
      js('file-m.js', fileMCode, [unused('m', 1, 48)]),
    ],
    { [file0]: `export { default, m1 } from '${fileM.replace(/\\/g, '/')}';` },
  );

  group('import and export in the same file both count as usage', [
    js('import-export-1.js', `export const a = 5;export const b = 't1'`),
  ]);

  group('renameDefault', [
    js(
      'renameDefault/components.js',
      'export { default as Component } from "./Component"',
    ),
    js('renameDefault/Component.js', 'export default function Component() {}'),
    js(
      'renameDefault-2/components.js',
      'export { default as ComponentA } from "./ComponentA";export { default as ComponentB } from "./ComponentB";',
    ),
    js(
      'renameDefault-2/ComponentA.js',
      'export default function ComponentA() {};',
    ),
  ]);

  // Upstream writes file-added-*.js to disk around these runs.
  const added = (index: string) => fixture(`file-added-${index}.js`);
  group(
    'add import in newly created file',
    [
      js(
        'file-added-0.js',
        `import * as m from '${fileM.replace(/\\/g, '/')}'`,
      ),
      js('file-m.js', fileMCode),
    ],
    { [added('0')]: `import * as m from '${fileM.replace(/\\/g, '/')}'` },
  );
  group('add export for newly created file', [
    js('file-added-0.js', 'export default () => {2}', [
      unused('default', 1, 1),
    ]),
  ]);
  group(
    'default import of newly created file',
    [
      js('file-0.js', `import def from '${added('0').replace(/\\/g, '/')}'`),
      js('file-added-0.js', 'export default () => {}'),
    ],
    { [file0]: `import def from '${added('0').replace(/\\/g, '/')}'` },
  );
  // export * only considers named imports. default imports still need to be reported
  group(
    'export star of newly created file',
    [
      js('file-0.js', `export * from '${added('0').replace(/\\/g, '/')}'`),
      js('file-r.js', `export * from 'external-compiled-library'`),
      js('file-added-0.js', `export const z = 'z';export default () => {}`, [
        unused('default', 1, 22),
      ]),
      js('file-added-0.js', 'export const a = 2'),
    ],
    { [file0]: `export * from '${added('0').replace(/\\/g, '/')}'` },
  );
  // remove export *. all exports need to be reported
  group(
    'remove export star',
    [
      js('file-0.js', `export { a } from '${added('0').replace(/\\/g, '/')}'`, [
        unused('a', 1, 10),
      ]),
      js('file-added-0.js', `export const z = 'z';export default () => {}`, [
        unused('z', 1, 1),
        unused('default', 1, 22),
      ]),
    ],
    {
      [file0]: `export { a } from '${added('0').replace(/\\/g, '/')}'`,
      [added('0')]: 'export const a = 2',
    },
  );
  group(
    'export star of second newly created file',
    [
      js('file-0.js', `export * from '${added('1').replace(/\\/g, '/')}'`),
      js('file-added-1.js', `export const z = 'z';export default () => {}`, [
        unused('default', 1, 22),
      ]),
    ],
    { [file0]: `export * from '${added('1').replace(/\\/g, '/')}'` },
  );
  group(
    'default import between newly created files',
    [
      js(
        'file-added-1.js',
        `import added from '${added('2').replace(/\\/g, '/')}'`,
      ),
      js('file-added-2.js', 'export default () => {}'),
    ],
    { [added('1')]: `import added from '${added('2').replace(/\\/g, '/')}'` },
  );
  group(
    'named import between newly created files',
    [
      js(
        'file-added-1.js',
        `import { added } from '${added('3').replace(/\\/g, '/')}'`,
      ),
      js('file-added-3.js', 'export const added = () => {}'),
    ],
    {
      [added('1')]: `import { added } from '${added('3').replace(/\\/g, '/')}'`,
    },
  );
  group('test behavior for destructured exports', [
    js(
      'file-destructured-2.js',
      `import { destructured } from '${fixture('file-destructured-1.js').replace(/\\/g, '/')}'`,
    ),
    js('file-destructured-1.js', 'export const { destructured } = {};'),
    js('file-destructured-1.js', 'export const { destructured2 } = {};', [
      unused('destructured2', 1, 1),
    ]),
  ]);
  group(
    'namespace import of a file with a double extension',
    [
      js(
        'file-added-1.js',
        `import * as added from '${fixture('file-added-4.js.js').replace(/\\/g, '/')}'`,
      ),
      js(
        'file-added-4.js.js',
        'export const added = () => {}; export default () => {}',
      ),
    ],
    {
      [added('1')]:
        `import * as added from '${fixture('file-added-4.js.js').replace(/\\/g, '/')}'`,
    },
  );

  group('do not report missing export for ignored file', [
    {
      filename: fixture('file-ignored-a.js'),
      code: 'export const test = true',
      options: {
        src: [fixture('**/*.js')],
        ignoreExports: [fixture('*ignored*.js')],
        missingExports: true,
      },
    },
  ]);
  group('lint file not available in `src`', [
    {
      filename: file('jsx/named.jsx'),
      code: `export const jsxFoo = 'foo'; export const jsxBar = 'bar'`,
      options: unusedExportsOptions,
    },
  ]);
  group('do not report unused export for files mentioned in package.json', [
    js('bin.js', 'export const bin = "bin"'),
    js('binObject/index.js', 'export const binObject = "binObject"'),
    js('browser.js', 'export const browser = "browser"'),
    js(
      'browserObject/index.js',
      'export const browserObject = "browserObject"',
    ),
    js('main/index.js', 'export const main = "main"'),
    js('privatePkg/index.js', 'export const privatePkg = "privatePkg"', [
      unused('privatePkg', 1, 1),
    ]),
  ]);
  group('avoid errors if re-export all from umd compiled library', [
    js(
      'main/index.js',
      `export * from '${fixture('bin.js').replace(/\\/g, '/')}'`,
    ),
  ]);

  const ts = (
    name: string,
    code: string,
    errors?: Expected[],
    options = unusedExportsTypescriptOptions,
  ): Case => ({
    filename: fixture(`typescript/${name}`),
    code,
    options,
    settings: typescriptSettings,
    errors,
  });
  // Upstream repeats this block for every TypeScript parser.
  group('TypeScript', [
    ts(
      'file-ts-a.ts',
      `
          import {b} from './file-ts-b';
          import {c} from './file-ts-c';
          import {d} from './file-ts-d';
          import {e} from './file-ts-e';

          const a = b + 1 + e.f;
          const a2: c = {};
          const a3: d = {};
          `,
    ),
    ts('file-ts-b.ts', 'export const b = 2;'),
    ts('file-ts-c.ts', 'export interface c {};'),
    ts('file-ts-d.ts', 'export type d = {};'),
    ts('file-ts-e.ts', 'export enum e { f };'),
    ts(
      'file-ts-a-import-type.ts',
      `
          import type {b} from './file-ts-b-used-as-type';
          import type {c} from './file-ts-c-used-as-type';
          import type {d} from './file-ts-d-used-as-type';
          import type {e} from './file-ts-e-used-as-type';

          const a: typeof b = 2;
          const a2: c = {};
          const a3: d = {};
          const a4: typeof e = undefined;
          `,
    ),
    ts('file-ts-b-used-as-type.ts', 'export const b = 2;'),
    ts('file-ts-c-used-as-type.ts', 'export interface c {};'),
    ts('file-ts-d-used-as-type.ts', 'export type d = {};'),
    ts('file-ts-e-used-as-type.ts', 'export enum e { f };'),
    // Should also be valid when the exporting files are linted before the importing ones
    ts('file-ts-g.ts', 'export interface g {}'),
    ts('file-ts-f.ts', `import {g} from './file-ts-g';`),
    ts(
      'file-ts-g-used-as-type.ts',
      'export interface g {}; /* used-as-type */',
    ),
    ts('file-ts-f-import-type.ts', `import type {g} from './file-ts-g';`),
    ts('file-ts-b-unused.ts', 'export const b = 2;', [unused('b', 1, 1)]),
    ts('file-ts-c-unused.ts', 'export interface c {};', [unused('c', 1, 1)]),
    ts('file-ts-d-unused.ts', 'export type d = {};', [unused('d', 1, 1)]),
    ts('file-ts-e-unused.ts', 'export enum e { f };', [unused('e', 1, 1)]),
  ]);

  const ignoreTypes = unusedExportsTypescriptIgnoreUnusedTypesOptions;
  group('ignoreUnusedTypeExports', [
    // unused vars should not report
    ts('file-ts-c-unused.ts', 'export interface c {};', undefined, ignoreTypes),
    ts('file-ts-d-unused.ts', 'export type d = {};', undefined, ignoreTypes),
    ts('file-ts-e-unused.ts', 'export enum e { f };', undefined, ignoreTypes),
    // used vars should not report
    ts(
      'file-ts-c-used-as-type.ts',
      'export interface c {};',
      undefined,
      ignoreTypes,
    ),
    ts(
      'file-ts-d-used-as-type.ts',
      'export type d = {};',
      undefined,
      ignoreTypes,
    ),
    ts(
      'file-ts-e-used-as-type.ts',
      'export enum e { f };',
      undefined,
      ignoreTypes,
    ),
  ]);

  const jsxA = fixture('jsx/file-jsx-a.jsx');
  group(
    'correctly work with JSX only files',
    [
      {
        filename: jsxA,
        code: 'import a from "file-jsx-a";',
        options: unusedExportsJsxOptions,
        settings: jsxSettings,
      },
      {
        filename: fixture('jsx/file-jsx-b.jsx'),
        code: 'export const b = 2;',
        options: unusedExportsJsxOptions,
        settings: jsxSettings,
        errors: [unused('b', 1, 1)],
      },
    ],
    { [jsxA]: 'import a from "file-jsx-a";' },
  );

  // ignore flow types: Flow annotations need babel-eslint, which rslint does
  // not support; the five upstream cases are kept as Go skips.

  group('support (nested) destructuring assignment', [
    js('destructuring-a.js', 'import {a, b} from "./destructuring-b";'),
    js(
      'destructuring-b.js',
      'const obj = {a: 1, dummy: {b: 2}}; export const {a, dummy: {b}} = obj;',
    ),
  ]);

  group('support ES2022 Arbitrary module namespace identifier names', [
    js(
      'arbitrary-module-namespace-identifier-name-b.js',
      'import { "foo" as foo } from "./arbitrary-module-namespace-identifier-name-a"',
    ),
    js(
      'arbitrary-module-namespace-identifier-name-a.js',
      'const foo = 333;\nexport { foo as "foo" }',
    ),
    js(
      'arbitrary-module-namespace-identifier-name-c.js',
      'const foo = 333\nexport { foo as "foo" }',
      [unused('foo', 2, 10)],
    ),
  ]);

  group('parser ignores prefixes like BOM and hashbang', [
    js('prefix-child.js', 'export const foo = 1;\n'),
    js(
      'prefix-parent-bom.js',
      `\uFEFF#!/usr/bin/env node\nimport {foo} from './prefix-child.js';\n`,
    ),
    js(
      'prefix-parent-hashbang.js',
      `#!/usr/bin/env node\nimport {foo} from './prefix-child.js';\n`,
    ),
    js(
      'prefix-parent-bomhashbang.js',
      `\uFEFF#!/usr/bin/env node\nimport {foo} from './prefix-child.js';\n`,
    ),
    js('prefix-parent.js', `import {foo} from './prefix-child.js';\n`),
  ]);

  group('supports flat eslint', [
    js('file-a.js', 'import { o2 } from "./file-o"; export default () => 12'),
    js('file-f.js', 'export default () => 13', [unused('default', 1, 1)]),
  ]);

  // Upstream's "with eslint 9+" case checks the error ESLint 9 raises without
  // an .eslintrc file. rslint enumerates `src` itself and needs no legacy
  // configuration, so it has no equivalent.
});
