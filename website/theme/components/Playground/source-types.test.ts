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
          lintPath: '/rstest-core.d.ts',
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
          '@rstest/core': ['/rstest-core.d.ts'],
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

  test('loads supported transitive declarations and caches the graph', async () => {
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
      if (url.includes('@rstest/core@1.2.3')) {
        return new Response(`
          import type { assert } from 'https://esm.sh/@types/chai@5.2.3/index.d.ts';
          import type { RsbuildPlugin } from 'https://esm.sh/@rsbuild/core@2.2.5/dist/index.d.ts';
          import type { Writable } from 'node:stream';
          export declare const expect: typeof assert;
        `);
      }
      if (url.includes('@types/chai@5.2.3')) {
        return new Response(`
          import deepEqual = require("https://esm.sh/@types/deep-eql@4.0.2/index.d.ts");
          import { AssertionError } from "https://esm.sh/assertion-error@2.0.1/index.d.ts";
          export { deepEqual, AssertionError };
        `);
      }
      if (url.includes('@types/deep-eql@4.0.2')) {
        return new Response('declare function deepEqual(): boolean;');
      }
      if (url.includes('assertion-error@2.0.1')) {
        return new Response('export declare class AssertionError {}');
      }
      return new Response('', { status: 404 });
    };

    try {
      const specifiers = ['@rstest/core'];
      const [first, second] = await Promise.all([
        loadSourceTypes(specifiers),
        loadSourceTypes(specifiers),
      ]);
      expect(first.key).toBe(sourceTypeKey(specifiers));
      expect(second.declarations[0]).toBe(first.declarations[0]);
      expect(first.declarations).toHaveLength(4);

      const declarationsByPath = Object.fromEntries(
        first.declarations.map(({ content, lintPath }) => [lintPath, content]),
      );
      expect(declarationsByPath['/rstest-core.d.ts']).toContain(
        `from './rstest-chai.d.ts'`,
      );
      expect(declarationsByPath['/rstest-core.d.ts']).toContain(
        `from 'https://esm.sh/@rsbuild/core@2.2.5/dist/index.d.ts'`,
      );
      expect(declarationsByPath['/rstest-core.d.ts']).toContain(
        `from 'node:stream'`,
      );
      expect(declarationsByPath['/rstest-chai.d.ts']).toContain(
        `require("./rstest-deep-eql.d.ts")`,
      );
      expect(declarationsByPath['/rstest-chai.d.ts']).toContain(
        `from "./rstest-assertion-error.d.ts"`,
      );

      expect(requests.sort()).toEqual(
        [
          'https://esm.sh/@rstest/core',
          'https://esm.sh/@rstest/core@1.2.3/dist/index.d.ts',
          'https://esm.sh/@types/chai@5.2.3/index.d.ts',
          'https://esm.sh/@types/deep-eql@4.0.2/index.d.ts',
          'https://esm.sh/assertion-error@2.0.1/index.d.ts',
        ].sort(),
      );

      await loadSourceTypes(specifiers);
      expect(requests).toHaveLength(5);
    } finally {
      globalThis.fetch = originalFetch;
    }
  });
});
