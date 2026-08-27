import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const valid = (code: string, filename: string = 'src/virtual.tsx') => ({
  code,
  filename,
});

const invalid = (
  code: string,
  errors: number,
  filename: string = 'src/virtual.tsx',
) => ({ code, filename, errors });

ruleTester.run('no-exports-in-scripts', null as never, {
  valid: [
    // ---- No shebang: the file is a module, exports are allowed ----
    valid('export const foo = 1;'),
    valid('export default foo;'),
    valid('export * from "./foo.js";'),
    valid('const foo = 1;\nexport {foo};'),
    valid('export {};'),

    // ---- TypeScript: type-only exports without a shebang are valid ----
    valid('export type Foo = string;', 'file.ts'),
    valid('export interface Foo {}', 'file.ts'),

    // ---- Shebang present, no export statements ----
    valid(
      '#!/usr/bin/env node\n' +
        "import process from 'node:process';\n" +
        '\n' +
        'console.log(process.argv);',
    ),
    valid(
      '#!/usr/bin/env node\n' + 'const foo = 1;\n' + 'module.exports = foo;',
    ),

    // ---- Pseudo-shebang in a comment does not count ----
    valid('// #!/usr/bin/env node\nexport const foo = 1;'),

    // ---- Pseudo-shebang in a string does not count ----
    valid("console.log('#!/usr/bin/env node');\nexport const foo = 1;"),
  ],
  invalid: [
    invalid('#!/usr/bin/env node\nexport const foo = 1;', 1),
    invalid('#!/usr/bin/env node\nexport default foo;', 1),
    invalid("#!/usr/bin/env node\nexport * from './foo.js';", 1),
    invalid("#!/usr/bin/env node\nexport * as foo from './foo.js';", 1),
    invalid('#!/usr/bin/env node\nconst foo = 1;\nexport {foo};', 1),
    invalid("#!/usr/bin/env node\nexport {foo} from './foo.js';", 1),
    invalid('#!/usr/bin/env node\nexport {};', 1),
    invalid(
      '#!/usr/bin/env node\nexport const foo = 1;\nexport const bar = 2;',
      2,
    ),
    // ---- TypeScript: type-only exports under a shebang ----
    invalid('#!/usr/bin/env node\nexport type Foo = string;', 1, 'file.ts'),
    invalid('#!/usr/bin/env node\nexport interface Foo {}', 1, 'file.ts'),
  ],
});
