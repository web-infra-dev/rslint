import { lint, RSLintService } from '@rslint/core';
import test from 'node:test';
import path from 'node:path';
import fs from 'node:fs';

// Create a test rslint.json configuration
const testRslintConfig = {
  "language": "javascript",
  "files": [],
  "languageOptions": {
    "parserOptions": {
      "projectService": false,
      "project": ["./tsconfig.json"]
    }
  },
  "rules": {
    "@typescript-eslint/no-unsafe-member-access": "error"
  },
  "plugins": ["@typescript-eslint"]
};

test('lint api', async t => {
  let cwd = path.resolve(import.meta.dirname, '../fixtures');
  
  await t.test('new API with rslint config', async t => {
    let configPath = path.resolve(cwd, 'test_rslint.json');
    fs.writeFileSync(configPath, JSON.stringify([testRslintConfig], null, 2));
    
    try {
      const diags = await lint({
        config: 'test_rslint.json',
        workingDirectory: cwd,
      });

      t.assert.equal(diags.errorCount, 2, 'Should find 2 errors');
      t.assert.equal(diags.fileCount, 1, 'Should lint 1 file');
      t.assert.ok(diags.diagnostics.length > 0, 'Should have diagnostics');
    } finally {
      fs.unlinkSync(configPath);
    }
  });

  await t.test('virtual file support', async t => {
    let configPath = path.resolve(cwd, 'test_rslint_virtual.json');
    fs.writeFileSync(configPath, JSON.stringify([{
      ...testRslintConfig,
      "languageOptions": {
        "parserOptions": {
          "projectService": false,
          "project": ["./tsconfig.virtual.json"]
        }
      }
    }], null, 2));
    
    let virtual_entry = path.resolve(cwd, 'src/virtual.ts');
    
    try {
      // Use virtual file contents instead of reading from disk
      const diags = await lint({
        config: 'test_rslint_virtual.json',
        workingDirectory: cwd,
        fileContents: {
          [virtual_entry]: `
                    let a:any = 10;
                    a.b =10;
                `,
        },
      });

      t.assert.ok(diags.errorCount > 0, 'Should find errors in virtual file');
    } finally {
      fs.unlinkSync(configPath);
    }
  });

  await t.test('legacy tsconfig support (backward compatibility)', async t => {
    let consoleWarns = [];
    const originalWarn = console.warn;
    console.warn = (msg) => consoleWarns.push(msg);
    
    try {
      let tsconfig = path.resolve(
        import.meta.dirname,
        '../fixtures/tsconfig.json',
      );
      const diags = await lint({ tsconfig, workingDirectory: cwd });
      
      t.assert.ok(diags.errorCount >= 0, 'Legacy API should work');
      t.assert.ok(consoleWarns.some(msg => msg.includes('deprecated')), 'Should show deprecation warning');
    } finally {
      console.warn = originalWarn;
    }
  });

  await test('default config', async t => {
    // Create a default rslint.json
    let defaultConfigPath = path.resolve(cwd, 'rslint.json');
    fs.writeFileSync(defaultConfigPath, JSON.stringify([testRslintConfig], null, 2));
    
    try {
      const diags = await lint({ workingDirectory: cwd });
      t.assert.equal(diags.errorCount, 2, 'Should use default config');
    } finally {
      fs.unlinkSync(defaultConfigPath);
    }
  });
});
