import { describe, test, expect } from '@rstest/core';
import {
  ruleIdFromSchemaPath,
  ruleIdToTypeName,
  injectIntoDts,
} from '../scripts/generate-rule-option-types.mjs';

describe('ruleIdFromSchemaPath', () => {
  test('resolves a core rule schema to its bare rule ID', () => {
    expect(
      ruleIdFromSchemaPath('internal/rules/eqeqeq/eqeqeq.schema.json'),
    ).toBe('eqeqeq');
    expect(
      ruleIdFromSchemaPath('internal/rules/no_console/no-console.schema.json'),
    ).toBe('no-console');
  });

  test('resolves a ported-plugin rule schema to its prefixed rule ID', () => {
    expect(
      ruleIdFromSchemaPath(
        'internal/plugins/typescript/rules/no_unused_vars/no-unused-vars.schema.json',
      ),
    ).toBe('@typescript-eslint/no-unused-vars');
    expect(
      ruleIdFromSchemaPath(
        'internal/plugins/react_hooks/rules/exhaustive_deps/exhaustive-deps.schema.json',
      ),
    ).toBe('react-hooks/exhaustive-deps');
    expect(
      ruleIdFromSchemaPath(
        'internal/plugins/jsx_a11y/rules/alt_text/alt-text.schema.json',
      ),
    ).toBe('jsx-a11y/alt-text');
  });

  test('throws for an unrecognized plugin directory', () => {
    expect(() =>
      ruleIdFromSchemaPath(
        'internal/plugins/mystery/rules/foo/foo.schema.json',
      ),
    ).toThrow(/unknown plugin directory/);
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
