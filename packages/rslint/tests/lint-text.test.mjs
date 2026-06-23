import { lintText, RSLintService, NodeRslintService } from '@rslint/core';
import { describe, test, expect } from '@rstest/core';
import path from 'node:path';

describe('lintText api', async (t) => {
  let cwd = path.resolve(import.meta.dirname, '../fixtures');
  let config = path.resolve(
    import.meta.dirname,
    '../fixtures/rslint.virtual.json',
  );
  let virtual_entry = path.resolve(cwd, 'src/virtual.ts');
  let ruleOptions = {
    '@typescript-eslint/no-unsafe-member-access': 'error',
  };
  let code = `let a:any = 10;
 a.b =10;`;

  test('lints in-memory content with fixture config', async (t) => {
    const diags = await lintText(code, {
      config,
      workingDirectory: cwd,
      ruleOptions,
      filePath: virtual_entry,
    });

    expect(Array.isArray(diags.diagnostics)).toBe(true);
    expect(diags.diagnostics.length).toBeGreaterThanOrEqual(1);
  });

  test('uses default filePath', async (t) => {
    const diags = await lintText('const value = 1;', {
      config,
      workingDirectory: cwd,
    });

    expect(Array.isArray(diags.diagnostics)).toBe(true);
  });

  test('returns diagnostics array for clean code', async (t) => {
    const diags = await lintText('const value = 1;\n', {
      config,
      workingDirectory: cwd,
      ruleOptions,
      filePath: virtual_entry,
    });

    expect(Array.isArray(diags.diagnostics)).toBe(true);
  });

  test('service lintText matches standalone helper', async (t) => {
    const options = {
      config,
      workingDirectory: cwd,
      ruleOptions,
      filePath: virtual_entry,
    };
    const standalone = await lintText(code, options);
    const service = new RSLintService(
      new NodeRslintService({
        workingDirectory: cwd,
      }),
    );

    try {
      const diags = await service.lintText(code, options);

      expect(Array.isArray(diags.diagnostics)).toBe(true);
      expect(diags.diagnostics.length).toBe(standalone.diagnostics.length);
    } finally {
      await service.close();
    }
  });
});
