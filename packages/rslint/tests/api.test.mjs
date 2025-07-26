import { lint, RSLintService } from '@rslint/core';
import test from 'node:test';
import assert from 'node:assert';
import path from 'node:path';
import fs from 'node:fs';

// Helper function for snapshot testing compatibility
function assertSnapshot(t, actual, snapshotFile) {
  if (typeof t.assert.snapshot === 'function') {
    // Use native snapshot if available (Node.js 24+)
    t.assert.snapshot(actual);
  } else {
    // Fallback for older Node.js versions
    const snapshotPath = path.resolve(import.meta.dirname, snapshotFile);

    if (fs.existsSync(snapshotPath)) {
      const expected = JSON.parse(fs.readFileSync(snapshotPath, 'utf8'));
      assert.deepStrictEqual(actual, expected);
    } else {
      // Create snapshot file if it doesn't exist
      fs.writeFileSync(snapshotPath, JSON.stringify(actual, null, 2));
      console.log(`Created snapshot file: ${snapshotPath}`);
    }
  }
}

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

    assertSnapshot(t, diags, 'api.test.mjs.snapshot.virtual');
  });
  await test('diag snapshot', async t => {
    let tsconfig = path.resolve(
      import.meta.dirname,
      '../fixtures/tsconfig.json',
    );
    const diags = await lint({ tsconfig, workingDirectory: cwd });
    assertSnapshot(t, diags, 'api.test.mjs.snapshot.diag');
  });
});
