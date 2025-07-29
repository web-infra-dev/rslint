import { lint } from '@rslint/core';
import { describe, test, expect } from '@rstest/core';
import path from 'node:path';

describe('lint api', async t => {
  let cwd = path.resolve(import.meta.dirname, '../fixtures');
  test('virtual file support', async t => {
    let config = path.resolve(
      import.meta.dirname,
      '../fixtures/rslint.virtual.json',
    );
    let virtual_entry = path.resolve(cwd, 'src/virtual.ts');
    // Use virtual file contents instead of reading from disk
    const diags = await lint({
      config,
      workingDirectory: cwd,
      fileContents: {
        [virtual_entry]: `
                    let a:any = 10;
                    a.b =10;
                `,
      },
    });

    expect(diags).toMatchSnapshot();
  });
  test('diag snapshot', async t => {
    let config = path.resolve(import.meta.dirname, '../fixtures/rslint.json');
    const diags = await lint({ config, workingDirectory: cwd });
    expect(diags).toMatchSnapshot();
  });
});
