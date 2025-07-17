import { lint, RSLintService } from '@rslint/core';
import test from 'node:test';
import path from 'node:path';

test('lint api', () => {
    test('virtual file support', async (t) => {
        let tsconfig = path.resolve(import.meta.dirname, '../fixtures/tsconfig.virtual.json');
        // Use virtual file contents instead of reading from disk
        const diags = await lint({
            tsconfig,
            fileContents: {
                '/src/virtual.ts': `
                    let a:any = 10;
                    a.b =10;
                `
            }
        });

        t.assert.snapshot(diags);

    })
    test('diag snapshot', async (t) => {
        let tsconfig = path.resolve(import.meta.dirname, '../fixtures/tsconfig.json');
        const diags = await lint(tsconfig);
        t.assert.snapshot(diags)
    });


})