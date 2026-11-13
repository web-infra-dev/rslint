import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import type { RuleTesterConfig } from '@rslint/rule-tester';

export const config: RuleTesterConfig = {
  filename: 'src/virtual.ts',
  fixtureFiles: {
    'tsconfig.json': readFileSync(
      join(import.meta.dirname, 'tsconfig.virtual.json'),
      'utf8',
    ),
  },
  languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
};
