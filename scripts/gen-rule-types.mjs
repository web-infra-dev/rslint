#!/usr/bin/env node
import { execFileSync } from 'node:child_process';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const repoRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  '..',
);

if (
  process.env.SKIP_GEN_RULE_TYPES === '1' ||
  process.env.SKIP_GEN_RULE_TYPES === 'true'
) {
  console.log(
    '[rslint] SKIP_GEN_RULE_TYPES is set. Skipping rule type generation.',
  );
  process.exit(0);
}

try {
  execFileSync('go', ['run', './tools/gen-rule-types/main.go'], {
    cwd: repoRoot,
    stdio: 'inherit',
  });
} catch (err) {
  console.error('[rslint] Failed to run gen-rule-types:', err);
  process.exit(1);
}
