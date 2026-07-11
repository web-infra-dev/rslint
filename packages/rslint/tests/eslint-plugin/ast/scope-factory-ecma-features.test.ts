/**
 * Regression tests pinning that user-supplied
 * `languageOptions.parserOptions.ecmaFeatures.*` flags actually reach
 * the scope analyzer — pre-fix the runner hard-coded `impliedStrict: true`
 * and silently dropped `globalReturn`, so:
 *
 *   - sourceType:'script' files were always analysed strict, producing
 *     spurious shadow / restricted-name diagnostics on legitimate
 *     non-strict code.
 *   - Node CJS scripts (top-level `return`) had no Function wrapper, so
 *     `variableScope.type` for top-level bindings was 'global' instead
 *     of 'function' — broke `no-unused-vars` and any rule that walks
 *     the scope chain.
 *
 * The tests probe the resulting `ScopeManager` directly via a stub
 * plugin that reads `ctx.sourceCode.scopeManager`.
 */
import { describe, test, expect } from '@rstest/core';

import { lintFile } from '../../../src/eslint-plugin/linter/ecma-language-plugin.js';
import type { LoadedPlugins } from '../../../src/eslint-plugin/plugin/plugin-loader.js';
import type { RuleContext } from '../../../src/eslint-plugin/linter/context.js';
import {
  seedEcmaGlobals,
  seedGlobals,
} from '../../../src/eslint-plugin/ast/scope-factory.js';

interface Probe {
  globalScopeType?: string;
  variableScopeTypeForTopVar?: string;
  childScopeTypes?: string[];
}

function makeProbePlugin(observed: Probe): LoadedPlugins {
  return {
    plugins: [],
    rules: new Map<string, unknown>([
      [
        'stub/scope-probe',
        {
          meta: { name: 'scope-probe' },
          create(ctx: RuleContext) {
            return {
              Program(node: unknown) {
                const sm = ctx.sourceCode.scopeManager as {
                  globalScope: {
                    type: string;
                    childScopes: Array<{ type: string }>;
                    variables: Array<{ name: string }>;
                  };
                  acquire: (n: unknown) => {
                    type: string;
                    variableScope?: { type: string };
                  } | null;
                };
                observed.globalScopeType = sm.globalScope.type;
                observed.childScopeTypes = sm.globalScope.childScopes.map(
                  (c) => c.type,
                );
                const programScope = sm.acquire(node);
                if (programScope?.variableScope) {
                  observed.variableScopeTypeForTopVar =
                    programScope.variableScope.type;
                }
              },
            };
          },
        },
      ],
    ]),
  };
}

function baseReq(text: string) {
  return {
    filePath: 'probe.js',
    text,
    rules: { 'stub/scope-probe': { options: [] } },
    collectFixes: false,
    suggestionsMode: 'off' as const,
  };
}

describe('scope-factory: globalReturn forwards to eslint-scope nodejsScope', () => {
  test('sourceType:script + globalReturn:true → Program is wrapped in a Function scope', () => {
    const probe: Probe = {};
    const loaded = makeProbePlugin(probe);

    const result = lintFile(
      {
        ...baseReq('var topVar = 1;\nreturn topVar;\n'),
        languageOptions: {
          sourceType: 'script',
          parserOptions: {
            ecmaFeatures: { globalReturn: true },
          },
        },
      },
      loaded,
    );

    // Pre-fix: globalReturn dropped on the floor, eslint-scope sees no
    // `nodejsScope: true`, top-level `return` either errors out or
    // resolves to module scope. Post-fix: eslint-scope wraps Program
    // in a Function scope, child of globalScope.
    expect(result.parseError).toBeUndefined();
    expect(probe.globalScopeType).toBe('global');
    // The wrapper Function scope is the immediate child of global.
    expect(probe.childScopeTypes).toContain('function');
  });

  test('sourceType:script + globalReturn:false (default) → NO Function wrapper', () => {
    // Counterpart: confirm we didn't accidentally always enable nodejsScope.
    // top-level `return` is invalid here, so use code without it.
    const probe: Probe = {};
    const loaded = makeProbePlugin(probe);

    lintFile(
      {
        ...baseReq('var topVar = 1;\n'),
        languageOptions: {
          sourceType: 'script',
          // ecmaFeatures omitted entirely
        },
      },
      loaded,
    );

    expect(probe.globalScopeType).toBe('global');
    // No function wrapper: every child of globalScope should be a
    // non-function scope (block / catch / class etc., but typically
    // empty for `var topVar = 1;`).
    expect(probe.childScopeTypes ?? []).not.toContain('function');
  });
});

describe('scope-factory: impliedStrict forwards to scope analyzers', () => {
  // The behaviour difference between impliedStrict true/false is subtle
  // in pure scope analysis (most strict-mode constraints are parser-
  // level). The robust proof here is: the parameter REACHES the
  // analyzer. We pin this by checking that strict-only reserved
  // identifiers (e.g. `package`) are accepted as parameter binding
  // names without a parse error in the non-strict path.

  test('sourceType:script + impliedStrict:false (default) → `package` usable as parameter', () => {
    const probe: Probe = {};
    const loaded = makeProbePlugin(probe);

    const result = lintFile(
      {
        ...baseReq('function go(package) { return package; }\n'),
        languageOptions: {
          sourceType: 'script',
          // impliedStrict omitted — defaults to v10's false
        },
      },
      loaded,
    );

    // Without strict, `package` is a valid identifier — no parser /
    // scope error.
    expect(result.parseError).toBeUndefined();
  });

  test('user impliedStrict:true is accepted (does not crash analyzer)', () => {
    const probe: Probe = {};
    const loaded = makeProbePlugin(probe);

    const result = lintFile(
      {
        ...baseReq('"use strict";\nfunction foo() { return 1; }\n'),
        languageOptions: {
          sourceType: 'script',
          parserOptions: {
            ecmaFeatures: { impliedStrict: true },
          },
        },
      },
      loaded,
    );

    // With explicit strict + impliedStrict, analyzer must still
    // process the file. Pin no crash and that scope analysis
    // produced a valid global scope.
    expect(result.parseError).toBeUndefined();
    expect(probe.globalScopeType).toBe('global');
  });
});

// ── P2 #12 — seedGlobals 'off' restores references ────────────────

describe("scope-factory: globals: { name: 'off' } restores refs to gs.through", () => {
  test("'Array': 'off' leaves Array references in globalScope.through (unresolved)", () => {
    // Pre-fix: `seedEcmaGlobals` moved Array refs onto the synthetic
    // Variable; then `seedGlobals('off')` deleted the Variable but
    // left its `.references[]` array intact. Refs were stranded —
    // not on any Variable accessible from the scope, AND no longer
    // in `gs.through`. ReferenceTracker / `iterateGlobalReferences`
    // -style plugin walks found zero Array references, breaking
    // every rule that depended on them.

    interface Probe {
      hasArrayVar?: boolean;
      throughNames?: string[];
    }
    const observed: Probe = {};

    const loaded: LoadedPlugins = {
      plugins: [],
      rules: new Map<string, unknown>([
        [
          'stub/probe',
          {
            meta: { name: 'probe' },
            create(ctx: RuleContext) {
              return {
                Program() {
                  const sm = ctx.sourceCode.scopeManager as {
                    globalScope: {
                      variables: Array<{ name: string }>;
                      through: Array<{ identifier?: { name?: string } }>;
                    };
                  };
                  observed.hasArrayVar = sm.globalScope.variables.some(
                    (v) => v.name === 'Array',
                  );
                  observed.throughNames = sm.globalScope.through
                    .map((r) => r.identifier?.name)
                    .filter((n): n is string => typeof n === 'string');
                },
              };
            },
          },
        ],
      ]),
    };

    lintFile(
      {
        filePath: 'probe.js',
        text: 'const a = Array.from([]);\nconst b = Array.of(1);\n',
        rules: { 'stub/probe': { options: [] } },
        collectFixes: false,
        suggestionsMode: 'off' as const,
        languageOptions: {
          sourceType: 'module',
          globals: { Array: 'off' },
        },
      },
      loaded,
    );

    // Variable was removed.
    expect(observed.hasArrayVar).toBe(false);
    // BUT — and this is the post-fix contract — the references that
    // were initially resolved to the now-deleted Array Variable are
    // now back in `gs.through`. ReferenceTracker style walks see them.
    expect(observed.throughNames).toEqual(
      expect.arrayContaining(['Array']) as never,
    );
  });
});

describe('scope-factory: global access aliases', () => {
  interface GlobalProbeVariable {
    name: string;
    defs?: unknown[];
    writeable: boolean;
    eslintImplicitGlobalSetting?: 'readonly' | 'writable';
    references: Array<{ resolved?: unknown }>;
  }

  function makeScopeManager() {
    const variables: GlobalProbeVariable[] = [];
    const set = new Map<string, GlobalProbeVariable>();
    return {
      variables,
      set,
      scopeManager: { globalScope: { variables, set, through: [] } },
    };
  }

  test('normalizes every legal writable and readonly alias', () => {
    const { variables, set, scopeManager } = makeScopeManager();

    seedGlobals(scopeManager, {
      writableBoolean: true,
      writableString: 'true',
      writable: 'writable',
      writeable: 'writeable',
      readonlyBoolean: false,
      readonlyString: 'false',
      readonly: 'readonly',
      readable: 'readable',
      nullReadonly: null,
      offDisabled: 'off',
    });

    const expectedModes = {
      writableBoolean: 'writable',
      writableString: 'writable',
      writable: 'writable',
      writeable: 'writable',
      readonlyBoolean: 'readonly',
      readonlyString: 'readonly',
      readonly: 'readonly',
      readable: 'readonly',
      nullReadonly: 'readonly',
    } as const;
    for (const [name, mode] of Object.entries(expectedModes)) {
      const variable = set.get(name);
      expect(variable?.writeable).toBe(mode === 'writable');
      expect(variable?.eslintImplicitGlobalSetting).toBe(mode);
    }

    expect(set.has('offDisabled')).toBe(false);
    expect(variables).toHaveLength(Object.keys(expectedModes).length);
    for (const variable of variables) {
      expect(['readonly', 'writable']).toContain(
        variable.eslintImplicitGlobalSetting,
      );
    }
  });

  test('null keeps an existing global readonly while off removes it', () => {
    const { set, scopeManager } = makeScopeManager();
    seedEcmaGlobals(scopeManager);
    expect(set.has('Array')).toBe(true);
    expect(set.has('Object')).toBe(true);

    seedGlobals(scopeManager, { Array: null, Object: 'off' });

    expect(set.get('Array')?.writeable).toBe(false);
    expect(set.get('Array')?.eslintImplicitGlobalSetting).toBe('readonly');
    expect(set.has('Object')).toBe(false);
  });

  test('configured globals never remove or rewrite source declarations', () => {
    const { variables, set, scopeManager } = makeScopeManager();
    const lexical = {
      name: 'Array',
      defs: [{}],
      writeable: false,
      references: [],
    } satisfies GlobalProbeVariable;
    variables.push(lexical);
    set.set('Array', lexical);

    seedGlobals(scopeManager, { Array: 'writable' });
    expect(set.get('Array')).toBe(lexical);
    expect(lexical.writeable).toBe(false);
    expect(lexical).not.toHaveProperty('eslintImplicitGlobalSetting');

    seedGlobals(scopeManager, { Array: 'off' });
    expect(set.get('Array')).toBe(lexical);
    expect(variables).toContain(lexical);
  });
});
