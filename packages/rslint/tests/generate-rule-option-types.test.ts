import { describe, test, expect } from '@rstest/core';
import {
  ruleIdToTypeName,
  injectIntoDts,
} from '../../../scripts/generate-rule-option-types.mjs';

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
