import fs from 'node:fs';
import path from 'node:path';
import type { RuleTesterConfig } from '@rslint/rule-tester';

// Preserve the existing projects and declaration files in each isolated root.
const fixtureRoot = path.join(import.meta.dirname, 'fixtures');
const fixtureFiles: Record<string, string> = {};
function collect(directory: string): void {
  for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
    const filename = path.join(directory, entry.name);
    if (entry.isDirectory()) collect(filename);
    else if (/\.(?:json|ts|tsx)$/.test(entry.name)) {
      fixtureFiles[path.relative(fixtureRoot, filename).replaceAll('\\', '/')] =
        fs.readFileSync(filename, 'utf8');
    }
  }
}
collect(fixtureRoot);

export const config: RuleTesterConfig = {
  plugin: '@typescript-eslint',
  filename: 'virtual.ts',
  fixtureFiles,
  languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
};
