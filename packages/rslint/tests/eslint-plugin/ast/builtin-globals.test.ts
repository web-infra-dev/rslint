/**
 * Drift guard for the edition-aware ECMAScript globals used by the community
 * plugin worker. The maintained table is checked against the Go source so the
 * two execution engines cannot silently disagree about `ecmaVersion`.
 */
import fs from 'node:fs';
import path from 'node:path';
import { describe, expect, test } from 'rstack/test';
import { fileURLToPath } from 'node:url';
import {
  ECMASCRIPT_GLOBALS,
  LATEST_ECMASCRIPT_VERSION,
} from '../../../src/eslint-plugin/ast/builtin-globals.js';
import {
  makeScopeManagerFactory,
  seedEcmaGlobals,
  seedGlobals,
} from '../../../src/eslint-plugin/ast/scope-factory.js';
import { parse as nativeParse } from '../../../src/eslint-plugin/native/load-binding.js';
import {
  buildLineStartOffsets,
  normalizeAst,
} from '../../../src/eslint-plugin/ast/normalize-ast.js';
import type { ESTreeNode } from '../../../src/eslint-plugin/linter/context.js';

interface ProbeVariable {
  name: string;
  references: unknown[];
  writeable?: boolean;
  eslintImplicitGlobalSetting?: string;
  isTypeVariable?: boolean;
  isValueVariable?: boolean;
}

interface ProbeScopeManager {
  globalScope: {
    set: Map<string, ProbeVariable>;
    through: Array<{
      identifier?: { name?: string };
      isTypeReference?: boolean;
      isValueReference?: boolean;
    }>;
  };
}

function analyzeTypeScriptSource(
  source: string,
  ecmaVersion: number | 'latest',
  globals?: Record<string, 'readonly' | 'writable' | 'off'>,
): ProbeScopeManager {
  const parsed = nativeParse('probe.ts', source, 'script', false);
  const ast = JSON.parse(parsed.program) as ESTreeNode;
  normalizeAst(ast, buildLineStartOffsets(source), source);
  const scopeManager = makeScopeManagerFactory(ast, {
    filePath: 'probe.ts',
    sourceType: 'script',
    ecmaVersion,
  })() as ProbeScopeManager;
  seedEcmaGlobals(scopeManager, ecmaVersion);
  seedGlobals(scopeManager, globals);
  return scopeManager;
}

function analyzeTypeScriptValueGlobal(
  name: string,
  ecmaVersion: number | 'latest',
  access?: 'readonly' | 'writable' | 'off',
): { variable: ProbeVariable | undefined; throughNames: string[] } {
  const position = { line: 1, column: 0 };
  const identifier = {
    type: 'Identifier',
    name,
    range: [0, name.length],
    loc: { start: position, end: { line: 1, column: name.length } },
  };
  const ast = {
    type: 'Program',
    sourceType: 'module',
    body: [
      {
        type: 'ExpressionStatement',
        expression: identifier,
        range: [0, name.length],
        loc: identifier.loc,
      },
    ],
    range: [0, name.length],
    loc: identifier.loc,
  };
  const scopeManager = makeScopeManagerFactory(ast, {
    filePath: 'probe.ts',
    sourceType: 'module',
    ecmaVersion,
  })() as {
    globalScope: {
      set: Map<string, ProbeVariable>;
      through: Array<{ identifier?: { name?: string } }>;
    };
  };
  seedEcmaGlobals(scopeManager, ecmaVersion);
  if (access !== undefined) seedGlobals(scopeManager, { [name]: access });
  return {
    variable: scopeManager.globalScope.set.get(name),
    throughNames: scopeManager.globalScope.through
      .map((reference) => reference.identifier?.name)
      .filter((candidate): candidate is string => candidate != null),
  };
}

function seededGlobalNames(ecmaVersion: number | 'latest'): string[] {
  const variables: Array<{
    name: string;
    defs: unknown[];
    identifiers: unknown[];
    references: Array<{
      identifier?: { name?: string };
      resolved?: unknown;
    }>;
    writeable: boolean;
    eslintImplicitGlobalSetting?: 'readonly' | 'writable';
    scope: unknown;
  }> = [];
  const set = new Map<string, (typeof variables)[number]>();
  const scopeManager = { globalScope: { variables, set, through: [] } };
  seedEcmaGlobals(scopeManager, ecmaVersion);
  return variables.map((variable) => variable.name);
}

describe('edition-aware ECMAScript global data', () => {
  test('matches internal/rule/globals.go', () => {
    const testDirectory = path.dirname(fileURLToPath(import.meta.url));
    const sourcePath = path.resolve(
      testDirectory,
      '../../../../../internal/rule/globals.go',
    );
    const source = fs.readFileSync(sourcePath, 'utf8').replaceAll('\r\n', '\n');
    const latest = source.match(/const LatestECMAScriptVersion = (\d+)/);
    expect(latest).not.toBeNull();
    expect(LATEST_ECMASCRIPT_VERSION).toBe(Number(latest?.[1]));

    const mapDeclaration = 'var ecmaScriptGlobalIntroducedIn = map[string]int{';
    const mapStart = source.indexOf(mapDeclaration);
    const mapEnd = source.indexOf('\n}\n\n// Globals', mapStart);
    expect(mapStart).toBeGreaterThanOrEqual(0);
    expect(mapEnd).toBeGreaterThan(mapStart);

    const entryLines = source
      .slice(mapStart + mapDeclaration.length, mapEnd)
      .split('\n')
      .filter((line) => {
        const trimmedLine = line.trim();
        return trimmedLine.length > 0 && !trimmedLine.startsWith('//');
      });
    const entries = entryLines.map((line) => {
      const match = line.match(/^\s*("(?:[^"\\]|\\.)*"):\s+(\d+),\s*$/);
      expect(match).not.toBeNull();
      return [JSON.parse(match?.[1] ?? ''), Number(match?.[2])];
    });

    expect(ECMASCRIPT_GLOBALS).toEqual(entries);
  });

  test.each([2025, 16] as const)(
    'ES%s excludes globals introduced in ES2026',
    (ecmaVersion) => {
      expect(seededGlobalNames(ecmaVersion)).not.toEqual(
        expect.arrayContaining([
          'AsyncDisposableStack',
          'DisposableStack',
          'SuppressedError',
          'Temporal',
        ]),
      );
    },
  );

  test.each([2026, 17, 'latest'] as const)(
    'ES%s includes globals introduced in ES2026',
    (ecmaVersion) => {
      expect(seededGlobalNames(ecmaVersion)).toEqual(
        expect.arrayContaining([
          'AsyncDisposableStack',
          'DisposableStack',
          'SuppressedError',
          'Temporal',
        ]),
      );
    },
  );

  test('derives the latest edition alias from the configured latest version', () => {
    const latestAlias = LATEST_ECMASCRIPT_VERSION - 2009;
    expect(seededGlobalNames(latestAlias)).toEqual(seededGlobalNames('latest'));
  });

  test('gates real TypeScript analyzer value globals by ecmaVersion', () => {
    const beforeIntroduction = analyzeTypeScriptValueGlobal('Temporal', 2025);
    expect(beforeIntroduction.variable?.isTypeVariable).toBe(true);
    expect(beforeIntroduction.variable?.isValueVariable).toBe(false);
    expect(
      beforeIntroduction.variable?.eslintImplicitGlobalSetting,
    ).toBeUndefined();
    expect(beforeIntroduction.variable?.references).toHaveLength(0);
    expect(beforeIntroduction.throughNames).toContain('Temporal');

    for (const ecmaVersion of [2026, 'latest'] as const) {
      const selected = analyzeTypeScriptValueGlobal('Temporal', ecmaVersion);
      expect(selected.variable?.isValueVariable).toBe(true);
      expect(selected.variable?.references).toHaveLength(1);
      expect(selected.throughNames).not.toContain('Temporal');
    }
  });

  test('reconciles TypeScript lib values for ecmaVersion 3', () => {
    const json = analyzeTypeScriptValueGlobal('JSON', 3);
    expect(json.variable?.isTypeVariable).toBe(true);
    expect(json.variable?.isValueVariable).toBe(false);
    expect(json.throughNames).toContain('JSON');

    const parseInt = analyzeTypeScriptValueGlobal('parseInt', 3);
    expect(parseInt.variable).toBeDefined();
    expect(parseInt.variable?.references).toHaveLength(1);
    expect(parseInt.throughNames).not.toContain('parseInt');
  });

  test('keeps TypeScript type bindings while off removes value resolution', () => {
    for (const name of ['Array', 'Temporal']) {
      const selected = analyzeTypeScriptValueGlobal(name, 2026, 'off');
      expect(selected.variable?.isTypeVariable).toBe(true);
      expect(selected.variable?.isValueVariable).toBe(false);
      expect(selected.variable?.references).toHaveLength(0);
      expect(selected.throughNames).toContain(name);
    }
  });

  test('lets explicit globals override a TypeScript lib global mode', () => {
    const selected = analyzeTypeScriptValueGlobal('Array', 2025, 'writable');
    expect(selected.variable?.isValueVariable).toBe(true);
    expect(selected.variable?.writeable).toBe(true);
    expect(selected.variable?.eslintImplicitGlobalSetting).toBe('writable');
    expect(selected.variable?.references).toHaveLength(1);
    expect(selected.throughNames).not.toContain('Array');
  });

  test.each([
    'interface Temporal {}\ntype UsesTemporal = Temporal;\nTemporal;',
    'type Temporal = string;\ntype UsesTemporal = Temporal;\nTemporal;',
  ])(
    'does not let a type-only declaration restore an excluded lib value: %s',
    (source) => {
      for (const globals of [undefined, { Temporal: 'off' as const }]) {
        const scopeManager = analyzeTypeScriptSource(source, 2025, globals);
        const variable = scopeManager.globalScope.set.get('Temporal');
        expect(variable?.isTypeVariable).toBe(true);
        expect(variable?.isValueVariable).toBe(false);
        expect(variable?.references).toHaveLength(1);
        expect(scopeManager.globalScope.through).toEqual(
          expect.arrayContaining([
            expect.objectContaining({
              identifier: expect.objectContaining({ name: 'Temporal' }),
              isValueReference: true,
            }),
          ]),
        );
      }
    },
  );

  test.each([
    'interface parseInt {}\ntype UsesParseInt = parseInt;\nparseInt("1");',
    'type parseInt = string;\ntype UsesParseInt = parseInt;\nparseInt("1");',
  ])(
    'adds an ES value alongside a same-named type declaration: %s',
    (source) => {
      const scopeManager = analyzeTypeScriptSource(source, 3);
      const variable = scopeManager.globalScope.set.get('parseInt');
      expect(variable?.isTypeVariable).toBe(true);
      expect(variable?.isValueVariable).toBe(true);
      expect(variable?.writeable).toBe(false);
      expect(variable?.eslintImplicitGlobalSetting).toBe('readonly');
      expect(variable?.references).toHaveLength(2);
      expect(
        scopeManager.globalScope.through.some(
          (reference) => reference.identifier?.name === 'parseInt',
        ),
      ).toBe(false);
    },
  );

  test.each(['readonly', 'writable', 'off'] as const)(
    'merges a configured %s value with a same-named type declaration',
    (access) => {
      const scopeManager = analyzeTypeScriptSource(
        'interface window {}\ntype UsesWindow = window;\nwindow;',
        2025,
        { window: access },
      );
      const variable = scopeManager.globalScope.set.get('window');
      expect(variable?.isTypeVariable).toBe(true);
      expect(variable?.isValueVariable).toBe(access !== 'off');
      expect(variable?.eslintImplicitGlobalSetting).toBe(
        access === 'off' ? undefined : access,
      );
      expect(variable?.writeable).toBe(
        access === 'off' ? undefined : access === 'writable',
      );
      expect(variable?.references).toHaveLength(access === 'off' ? 1 : 2);
      expect(
        scopeManager.globalScope.through.some(
          (reference) => reference.identifier?.name === 'window',
        ),
      ).toBe(access === 'off');
    },
  );

  test.each(['readonly', 'writable', 'off'] as const)(
    'applies configured %s mode without removing a source value',
    (access) => {
      const scopeManager = analyzeTypeScriptSource(
        'var Temporal = 1; Temporal = 2;',
        2025,
        { Temporal: access },
      );
      const variable = scopeManager.globalScope.set.get('Temporal');
      expect(variable?.isValueVariable).toBe(true);
      expect(variable?.eslintImplicitGlobalSetting).toBe(
        access === 'off' ? undefined : access,
      );
      expect(variable?.writeable).toBe(
        access === 'off' ? undefined : access === 'writable',
      );
      expect(variable?.references.length).toBeGreaterThan(0);
      expect(
        scopeManager.globalScope.through.some(
          (reference) => reference.identifier?.name === 'Temporal',
        ),
      ).toBe(false);
    },
  );

  test.each([
    'var Temporal = 1; Temporal;',
    'class Temporal {} Temporal;',
    'namespace Temporal { export const marker = 1; } Temporal;',
  ])('never removes a source-defined TypeScript value: %s', (source) => {
    const scopeManager = analyzeTypeScriptSource(source, 2025, {
      Temporal: 'off',
    });
    const variable = scopeManager.globalScope.set.get('Temporal');
    expect(variable?.isValueVariable).toBe(true);
    expect(variable?.references.length).toBeGreaterThan(0);
    expect(
      scopeManager.globalScope.through.some(
        (reference) => reference.identifier?.name === 'Temporal',
      ),
    ).toBe(false);
  });
});
