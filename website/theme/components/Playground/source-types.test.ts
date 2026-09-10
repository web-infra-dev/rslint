import { describe, expect, test } from 'rstack/test';
import {
  addSourceTypeResolutions,
  cancelUnusedSourceTypeLoads,
  findSourceTypePackages,
  loadSourceTypes,
  sourceTypeKey,
} from './source-types';

describe('Playground source dependency types', () => {
  test('detects supported static imports and exports', () => {
    for (const source of [
      `import { expect, test } from '@rstest/core';`,
      `import { expect, test } from"@rstest/core";`,
      `import { expect, test } from   '@rstest/core';`,
      `import { expect, test } from\t"@rstest/core";`,
      `import { expect, test } from\n'@rstest/core';`,
      `export { expect } from   "@rstest/core";`,
    ]) {
      expect(findSourceTypePackages(source)).toEqual(['@rstest/core']);
    }
    expect(findSourceTypePackages(`import('@rstest/core');`)).toEqual([]);
    expect(findSourceTypePackages(`require('@rstest/core');`)).toEqual([]);
    expect(
      findSourceTypePackages(`import { test } from 'rstack/test';`),
    ).toEqual([]);
  });

  test('adds declaration paths without mutating the tsconfig', () => {
    const environment = {
      key: '@rstest/core',
      declarations: [
        {
          specifier: '@rstest/core',
          content: 'export declare const test: unknown;',
          monacoPath: 'file:///node_modules/@rstest/core/index.d.ts',
          lintPath: '/source-types/rstest-core.d.ts',
        },
      ],
    };
    const tsconfig = {
      compilerOptions: {
        strict: true,
        paths: { existing: ['./existing.d.ts'] },
      },
    };

    expect(addSourceTypeResolutions(tsconfig, environment)).toEqual({
      compilerOptions: {
        strict: true,
        paths: {
          existing: ['./existing.d.ts'],
          '@rstest/core': ['/source-types/rstest-core.d.ts'],
        },
      },
    });
    expect(tsconfig).toEqual({
      compilerOptions: {
        strict: true,
        paths: { existing: ['./existing.d.ts'] },
      },
    });
  });

  test('cancels a declaration request that is no longer needed', async () => {
    const originalFetch = globalThis.fetch;
    globalThis.fetch = (_input, init) =>
      new Promise((_resolve, reject) => {
        init?.signal?.addEventListener('abort', () => {
          reject(new DOMException('Aborted', 'AbortError'));
        });
      });

    try {
      const request = loadSourceTypes(['@rstest/core']);
      cancelUnusedSourceTypeLoads([]);
      await expect(request).rejects.toThrow('Aborted');
    } finally {
      globalThis.fetch = originalFetch;
    }
  });

  test('resolves each latest package once and reuses it in the session', async () => {
    const originalFetch = globalThis.fetch;
    const requests: string[] = [];
    globalThis.fetch = async (input) => {
      const url = String(input);
      requests.push(url);
      if (url === 'https://esm.sh/@rstest/core') {
        return new Response('', {
          headers: {
            'X-TypeScript-Types':
              'https://esm.sh/@rstest/core@1.2.3/dist/index.d.ts',
          },
        });
      }
      return new Response('export declare const test: unknown;');
    };

    try {
      const specifiers = ['@rstest/core'];
      const [first, second] = await Promise.all([
        loadSourceTypes(specifiers),
        loadSourceTypes(specifiers),
      ]);
      expect(first.key).toBe(sourceTypeKey(specifiers));
      expect(first.declarations[0].content).toContain('declare const test');
      expect(second.declarations[0]).toBe(first.declarations[0]);
      expect(requests).toEqual([
        'https://esm.sh/@rstest/core',
        'https://esm.sh/@rstest/core@1.2.3/dist/index.d.ts',
      ]);

      await loadSourceTypes(specifiers);
      expect(requests).toHaveLength(2);
    } finally {
      globalThis.fetch = originalFetch;
    }
  });
});
