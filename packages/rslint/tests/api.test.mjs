import { lint, RSLintService } from '@rslint/core';
import test from 'node:test';
import path from 'node:path';

test('lint api', async t => {
  let cwd = path.resolve(import.meta.dirname, '../fixtures');
  await t.test('virtual file support', async t => {
    let tsconfig = path.resolve(
      import.meta.dirname,
      '../fixtures/tsconfig.virtual.json',
    );
    let virtual_entry = path.resolve(cwd, 'src/virtual.ts');
    // Use virtual file contents instead of reading from disk
    const diags = await lint({
      tsconfig,
      cwd,
      fileContents: {
        [virtual_entry]: `
                    let a:any = 10;
                    a.b =10;
                `,
      },
    });

    t.assert.snapshot(diags);
  });
  await test('diag snapshot', async t => {
    let tsconfig = path.resolve(
      import.meta.dirname,
      '../fixtures/tsconfig.json',
    );
    const diags = await lint({ tsconfig, workingDirectory: cwd });
    t.assert.snapshot(diags);
  });
});
