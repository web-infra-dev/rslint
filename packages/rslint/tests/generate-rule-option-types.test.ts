import { describe, test, expect, beforeAll, afterAll } from '@rstest/core';
import { execFileSync } from 'node:child_process';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {
  collectRuleSchemas,
  ruleIdToTypeName,
  injectIntoDts,
} from '../../../scripts/generate-rule-option-types.mjs';

const REPO_ROOT = path.resolve(__dirname, '../../..');

describe('collectRuleSchemas', () => {
  // Exercises the real `go run ./cmd/dump-rule-schemas` boundary (requires a
  // Go toolchain) rather than mocking it, since the whole point of going
  // through Go's rule registry is that it's the single source of truth for
  // rule IDs/prefixes and declared schemas. collectRuleSchemas() itself is
  // just a JSON file read, so this dumps into a temp file to exercise it.
  let schemasPath: string;

  beforeAll(() => {
    schemasPath = path.join(
      fs.mkdtempSync(path.join(os.tmpdir(), 'rule-schemas-')),
      'rule-schemas.json',
    );
    const output = execFileSync('go', ['run', './cmd/dump-rule-schemas'], {
      cwd: REPO_ROOT,
      encoding: 'utf-8',
    });
    fs.writeFileSync(schemasPath, output);
  });

  afterAll(() => {
    fs.rmSync(path.dirname(schemasPath), { recursive: true, force: true });
  });

  test('includes rules with a custom schema and the shared EmptyArraySchema, omits not-yet-migrated rules', async () => {
    const rules = await collectRuleSchemas(schemasPath);
    const byName = new Map(rules.map((r) => [r.name, r]));

    expect(byName.has('eqeqeq')).toBe(true);
    expect(byName.has('no-console')).toBe(true);

    // no-debugger has no on-disk *.schema.json — it references the shared
    // rule.EmptyArraySchema directly in Go — which is exactly the case a
    // filesystem scan alone can't see.
    expect(byName.get('no-debugger')?.schema).toEqual({
      type: 'array',
      maxItems: 0,
    });

    // A rule with no declared Schema at all must be omitted, not present
    // with a null/empty schema.
    expect(byName.has('no-var')).toBe(false);
  });

  test('throws a clear, tagged error when the schemas dump is missing', async () => {
    await expect(
      collectRuleSchemas(
        path.join(REPO_ROOT, 'packages/rslint/does-not-exist.json'),
      ),
    ).rejects.toMatchObject({ code: 'RULE_SCHEMAS_NOT_FOUND' });
  });
});

describe('ruleIdToTypeName', () => {
  test('converts a bare rule ID to PascalCase', () => {
    expect(ruleIdToTypeName('no-console')).toBe('NoConsole');
    expect(ruleIdToTypeName('eqeqeq')).toBe('Eqeqeq');
  });

  test('converts a prefixed rule ID to PascalCase, dropping separators', () => {
    expect(ruleIdToTypeName('@typescript-eslint/no-unused-vars')).toBe(
      'TypescriptEslintNoUnusedVars',
    );
    expect(ruleIdToTypeName('react-hooks/exhaustive-deps')).toBe(
      'ReactHooksExhaustiveDeps',
    );
  });
});

describe('injectIntoDts', () => {
  const pristine =
    '/**\r\n' +
    ' * Doc comment for RulesRecord.\r\n' +
    ' */\r\n' +
    'declare interface RulesRecord {\r\n' +
    '    /** @__RULE_OPTIONS__ */\r\n' +
    '    [key: string]: RuleEntry<any[]> | undefined;\r\n' +
    '}\r\n';

  test('splices named properties in place of the marker and type declarations after the interface', () => {
    const result = injectIntoDts(pristine, {
      typeDeclarations: ['export type EqeqeqOptions = [];'],
      recordProperties: ['"eqeqeq"?: RuleEntry<EqeqeqOptions>;'],
    });

    expect(result).toContain('export type EqeqeqOptions = [];');
    // The generated type must land AFTER the interface's closing brace, not
    // wedged between RulesRecord's own doc comment and its declaration.
    expect(result.indexOf('export type EqeqeqOptions')).toBeGreaterThan(
      result.indexOf('}'),
    );
    expect(result).toContain(
      'Doc comment for RulesRecord.\r\n */\r\ndeclare interface RulesRecord {',
    );
    // The marker comment itself must not survive into the output.
    expect(result).not.toContain('@__RULE_OPTIONS__');
    expect(result).toContain(
      '    "eqeqeq"?: RuleEntry<EqeqeqOptions>;\n    [key: string]: RuleEntry<any[]> | undefined;\r\n}\r\n',
    );
  });

  test("keeps every injected property at the marker line's own indentation", () => {
    const result = injectIntoDts(pristine, {
      typeDeclarations: [],
      recordProperties: [
        '"eqeqeq"?: RuleEntry<EqeqeqOptions>;',
        '"no-console"?: RuleEntry<NoConsoleOptions>;',
      ],
    });

    expect(result).toContain(
      '    "eqeqeq"?: RuleEntry<EqeqeqOptions>;\n' +
        '    "no-console"?: RuleEntry<NoConsoleOptions>;\n' +
        '    [key: string]: RuleEntry<any[]> | undefined;',
    );
  });

  test('strips the marker comment even when nothing was generated, leaving the fallback signature intact', () => {
    const result = injectIntoDts(pristine, {
      typeDeclarations: [],
      recordProperties: [],
    });
    expect(result).toBe(
      '/**\r\n' +
        ' * Doc comment for RulesRecord.\r\n' +
        ' */\r\n' +
        'declare interface RulesRecord {\r\n' +
        '    [key: string]: RuleEntry<any[]> | undefined;\r\n' +
        '}\r\n',
    );
  });

  test('returns dts as-is when the marker is missing', () => {
    const input = 'interface RulesRecord {}';
    expect(
      injectIntoDts(input, {
        typeDeclarations: [],
        recordProperties: [],
      }),
    ).toBe(input);
  });

  test('throws when RulesRecord is missing', () => {
    expect(() =>
      injectIntoDts('/** @__RULE_OPTIONS__ */', {
        typeDeclarations: [],
        recordProperties: [],
      }),
    ).toThrow(/couldn't find `interface RulesRecord`/);
  });
});
