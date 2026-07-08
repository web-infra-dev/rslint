/**
 * Pins that `languageOptions.parserOptions.jsxPragma` / `jsxFragmentName`
 * reach the (TS) scope analyzer. These are top-level `parserOptions` keys —
 * siblings of `ecmaFeatures`, not nested under it — per
 * `@typescript-eslint/types`' `ParserOptions`.
 *
 * `jsxPragma` only affects the TS scope-manager (`.tsx`/`.ts`), so the probe
 * uses a `.tsx` file.
 */
import { describe, test, expect } from '@rstest/core';

import { lintFile } from '../../../src/eslint-plugin/linter/ecma-language-plugin.js';
import type { LoadedPlugins } from '../../../src/eslint-plugin/plugin/plugin-loader.js';
import type { RuleContext } from '../../../src/eslint-plugin/linter/context.js';

interface ScopeLike {
  variables: Array<{ name: string; references: unknown[] }>;
  childScopes: ScopeLike[];
}

function findVariable(
  scope: ScopeLike,
  name: string,
): { name: string; references: unknown[] } | undefined {
  for (const v of scope.variables) {
    if (v.name === name) return v;
  }
  for (const child of scope.childScopes) {
    const found = findVariable(child, name);
    if (found) return found;
  }
  return undefined;
}

function makeProbePlugin(observed: {
  referencedNames: string[];
}): LoadedPlugins {
  return {
    plugins: [],
    rules: new Map<string, unknown>([
      [
        'stub/jsx-pragma-probe',
        {
          meta: { name: 'jsx-pragma-probe' },
          create(ctx: RuleContext) {
            return {
              'Program:exit'() {
                const sm = ctx.sourceCode.scopeManager as {
                  globalScope: ScopeLike;
                };
                for (const name of ['h', 'Fragment', 'React']) {
                  const v = findVariable(sm.globalScope, name);
                  if (v && v.references.length > 0) {
                    observed.referencedNames.push(name);
                  }
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
    filePath: 'probe.tsx',
    text,
    rules: { 'stub/jsx-pragma-probe': { options: [] } },
    collectFixes: false,
    suggestionsMode: 'off' as const,
  };
}

describe('scope-factory: jsxPragma / jsxFragmentName forward from languageOptions.parserOptions', () => {
  test('jsxPragma:"h" marks the `h` import as referenced by <div />, not "React"', () => {
    const observed = { referencedNames: [] as string[] };
    const loaded = makeProbePlugin(observed);

    const result = lintFile(
      {
        ...baseReq(`import { h } from 'preact';\nconst el = <div>hi</div>;\n`),
        languageOptions: {
          parserOptions: {
            ecmaFeatures: { jsx: true },
            jsxPragma: 'h',
          },
        },
      },
      loaded,
    );

    expect(result.parseError).toBeUndefined();
    expect(observed.referencedNames).toContain('h');
    expect(observed.referencedNames).not.toContain('React');
  });

  test('jsxFragmentName:"Fragment" marks the `Fragment` import as referenced by <>...</>', () => {
    const observed = { referencedNames: [] as string[] };
    const loaded = makeProbePlugin(observed);

    const result = lintFile(
      {
        ...baseReq(`import { Fragment } from 'preact';\nconst el = <>hi</>;\n`),
        languageOptions: {
          parserOptions: {
            ecmaFeatures: { jsx: true },
            jsxFragmentName: 'Fragment',
          },
        },
      },
      loaded,
    );

    expect(result.parseError).toBeUndefined();
    expect(observed.referencedNames).toContain('Fragment');
  });

  test('omitted jsxPragma still defaults to "React" (back-compat)', () => {
    const observed = { referencedNames: [] as string[] };
    const loaded = makeProbePlugin(observed);

    const result = lintFile(
      {
        ...baseReq(`import React from 'react';\nconst el = <div>hi</div>;\n`),
        languageOptions: {
          parserOptions: {
            ecmaFeatures: { jsx: true },
          },
        },
      },
      loaded,
    );

    expect(result.parseError).toBeUndefined();
    expect(observed.referencedNames).toContain('React');
  });
});
