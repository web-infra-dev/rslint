import path from 'node:path';

import upstreamGroups from '../../../../../internal/plugins/unicorn/rules/name_replacements/testdata/name_replacements_v77.json';

import { RuleTester } from '../rule-tester';

type UnknownRecord = Record<string, any>;

function normalizeRegExpOptions(value: any): {
  value?: any;
  unsupported: boolean;
} {
  if (Array.isArray(value)) {
    const result = [];
    for (const item of value) {
      const normalized = normalizeRegExpOptions(item);
      if (normalized.unsupported) {
        return normalized;
      }
      result.push(normalized.value);
    }
    return { value: result, unsupported: false };
  }

  if (value && typeof value === 'object') {
    if (typeof value.__regexp === 'string') {
      // Native configuration can carry string patterns, but not a JavaScript
      // RegExp's flags. The flagged upstream case stays in the Go fixture as
      // an explicit skip.
      if (String(value.flags).includes('i')) {
        return { unsupported: true };
      }
      return { value: value.__regexp, unsupported: false };
    }

    const result: UnknownRecord = {};
    for (const [key, item] of Object.entries(value)) {
      const normalized = normalizeRegExpOptions(item);
      if (normalized.unsupported) {
        return normalized;
      }
      result[key] = normalized.value;
    }
    return { value: result, unsupported: false };
  }

  return { value, unsupported: false };
}

function normalizeMessage(value: any): any {
  if (value && typeof value === 'object' && value.__regexp) {
    return new RegExp(value.__regexp, value.flags);
  }
  return value;
}

function supportedFilename(filename?: string): boolean {
  if (!filename) {
    return true;
  }
  if (path.basename(filename).startsWith('.')) {
    return false;
  }
  return [
    '.js',
    '.jsx',
    '.mjs',
    '.cjs',
    '.ts',
    '.tsx',
    '.mts',
    '.cts',
  ].includes(path.extname(filename));
}

function normalizeCase(
  raw: any,
  globals: Record<string, unknown> | undefined,
): any | undefined {
  const testCase: UnknownRecord =
    typeof raw === 'string' ? { code: raw } : structuredClone(raw);
  if (testCase.language || !supportedFilename(testCase.filename)) {
    return;
  }

  const normalizedOptions = normalizeRegExpOptions(testCase.options);
  if (normalizedOptions.unsupported) {
    return;
  }
  testCase.options = normalizedOptions.value;
  if (globals) {
    testCase.languageOptions = { globals };
  }
  if (Array.isArray(testCase.errors)) {
    testCase.errors = testCase.errors.map((error: any) => {
      if (error && typeof error === 'object') {
        error.message = normalizeMessage(error.message);
      }
      return error;
    });
  }
  return testCase;
}

const valid: any[] = [];
const invalid: any[] = [];

for (const group of upstreamGroups) {
  // Unicorn v77 also exercises Vue, CSS, HTML, JSON, YAML, TOML, and Markdown.
  // Those cases remain in the shared upstream fixture and are explicit Go
  // skips because native rslint currently parses JavaScript and TypeScript.
  if (group.kind !== 'javascript' && group.kind !== 'typescript') {
    continue;
  }
  const globals = group.cases.testerOptions?.languageOptions?.globals;
  for (const raw of group.cases.valid) {
    const testCase = normalizeCase(raw, globals);
    if (testCase) {
      valid.push(testCase);
    }
  }
  for (const raw of group.cases.invalid) {
    const testCase = normalizeCase(raw, globals);
    if (testCase) {
      invalid.push(testCase);
    }
  }
}

new RuleTester().run('name-replacements', {} as never, { valid, invalid });
