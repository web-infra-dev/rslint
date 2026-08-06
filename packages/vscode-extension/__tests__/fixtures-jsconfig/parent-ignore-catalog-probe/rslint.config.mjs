import fs from 'node:fs';

const loadMarker = new URL('./config-loads.txt', import.meta.url);
fs.appendFileSync(loadMarker, 'x');

export default [{ files: ['**/*.ts'], rules: { 'no-console': 'error' } }];
