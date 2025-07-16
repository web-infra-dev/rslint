import { lint } from '@rslint/core';
import test from 'node:test';
import path from 'node:path';

test('lint api',  () => {
    test('diag snapshot', async (t) => {
        let tsconfig = path.resolve(import.meta.dirname, '../fixtures/tsconfig.json');
        
        const diags = await lint(tsconfig);
        
        t.assert.snapshot(diags)
    })
})