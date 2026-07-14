import { describe, test, expect } from '@rstest/core';
import {
  collectRuleSchemas,
  ruleIdToTypeName,
  injectIntoDts,
} from '../scripts/generate-rule-option-types.mjs';

describe('collectRuleSchemas', () => {
  // Exercises the real `go run ./cmd/gen-rule-types` boundary (Go is
  // already a required toolchain for this package — `build:bin` compiles
  // the native binary the same way) rather than mocking it, since the
  // whole point of going through Go's rule registry is that it's the
  // single source of truth for rule IDs/prefixes and declared schemas.
  test('includes rules with a custom schema and the shared EmptyArraySchema, omits not-yet-migrated rules', () => {
    const rules = collectRuleSchemas();
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

  test('splices named properties before the marker and type declarations after the interface', () => {
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
    expect(result).toContain(
      '    "eqeqeq"?: RuleEntry<EqeqeqOptions>;\n    /** @__RULE_OPTIONS__ */',
    );
    // The pre-existing marker line (and everything after it) must survive untouched.
    expect(result).toContain(
      '    /** @__RULE_OPTIONS__ */\r\n    [key: string]: RuleEntry<any[]> | undefined;\r\n}\r\n',
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
        '    /** @__RULE_OPTIONS__ */',
    );
  });

  test('is a no-op on the record properties when nothing was generated', () => {
    const result = injectIntoDts(pristine, {
      typeDeclarations: [],
      recordProperties: [],
    });
    expect(result).toBe(pristine);
  });

  test('throws when the marker is missing', () => {
    expect(() =>
      injectIntoDts('interface RulesRecord {}', {
        typeDeclarations: [],
        recordProperties: [],
      }),
    ).toThrow(/couldn't find the/);
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
