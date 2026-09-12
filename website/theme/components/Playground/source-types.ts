export interface SourceTypeDeclaration {
  specifier?: string;
  content: string;
  monacoPath: string;
  lintPath: string;
}

export interface SourceTypeEnvironment {
  key: string;
  declarations: SourceTypeDeclaration[];
}

interface SourceTypePackage {
  specifier: string;
  packageUrl: string;
  monacoPath: string;
  lintPath: string;
  importPattern: RegExp;
}

interface TransitiveTypeDependency {
  urlPattern: RegExp;
  virtualPath: string;
}

// Add another descriptor here when the Playground supports types for another
// source dependency. Packages are loaded only when their pattern matches.
const SOURCE_TYPE_PACKAGES: SourceTypePackage[] = [
  {
    specifier: '@rstest/core',
    packageUrl: 'https://esm.sh/@rstest/core',
    monacoPath: 'file:///node_modules/@rstest/core/index.d.ts',
    lintPath: '/rstest-core.d.ts',
    importPattern: /\bfrom\s*(['"])@rstest\/core\1/,
  },
];

const TRANSITIVE_TYPE_DEPENDENCIES: TransitiveTypeDependency[] = [
  {
    urlPattern: new RegExp(
      '^https://esm\\.sh/@types/chai@[^/]+/index\\.d\\.ts$',
    ),
    virtualPath: 'rstest-chai.d.ts',
  },
  {
    urlPattern: new RegExp(
      '^https://esm\\.sh/@types/deep-eql@[^/]+/index\\.d\\.ts$',
    ),
    virtualPath: 'rstest-deep-eql.d.ts',
  },
  {
    urlPattern: new RegExp(
      '^https://esm\\.sh/assertion-error@[^/]+/index\\.d\\.ts$',
    ),
    virtualPath: 'rstest-assertion-error.d.ts',
  },
];

const TYPE_IMPORT =
  /(?:from\s+|import\(\s*|require\(\s*)(['"])([^'"]+\.d\.ts)\1/g;

const cachedDeclarations = new Map<string, SourceTypeDeclaration[]>();
const pendingDeclarations = new Map<
  string,
  { controller: AbortController; promise: Promise<SourceTypeDeclaration[]> }
>();

export function findSourceTypePackages(source: string): string[] {
  return SOURCE_TYPE_PACKAGES.filter(({ importPattern }) =>
    importPattern.test(source),
  ).map(({ specifier }) => specifier);
}

export function sourceTypeKey(specifiers: string[]): string {
  return specifiers.join('\0');
}

function relativeSpecifier(fromPath: string, toPath: string): string {
  const from = fromPath.split('/').slice(0, -1);
  const to = toPath.split('/');
  while (from.length > 0 && to.length > 0 && from[0] === to[0]) {
    from.shift();
    to.shift();
  }
  const relative = `${'../'.repeat(from.length)}${to.join('/')}`;
  return relative.startsWith('../') ? relative : `./${relative}`;
}

async function fetchDeclarationGraph(
  dependency: SourceTypePackage,
  entryUrl: URL,
  signal: AbortSignal,
): Promise<SourceTypeDeclaration[]> {
  const declarations: SourceTypeDeclaration[] = [];
  const loaded = new Map<string, Promise<void>>();
  const monacoRoot = new URL('.', dependency.monacoPath);
  const lintRoot = new URL('.', `file://${dependency.lintPath}`);

  function load(
    url: URL,
    monacoPath: string,
    lintPath: string,
    specifier?: string,
  ): Promise<void> {
    const existing = loaded.get(url.href);
    if (existing) return existing;

    const request = (async () => {
      const response = await fetch(url, { signal });
      if (!response.ok) {
        throw new Error(
          `unable to load ${dependency.specifier} declarations: ${response.status}`,
        );
      }

      let content = await response.text();
      const imports = [...content.matchAll(TYPE_IMPORT)].map(
        (match) => match[2],
      );
      await Promise.all(
        imports.map(async (imported) => {
          const importedUrl = new URL(imported, url);
          let importedMonacoPath: string;
          let importedLintPath: string;

          if (imported.startsWith('./') || imported.startsWith('../')) {
            importedMonacoPath = new URL(imported, monacoPath).href;
            importedLintPath = new URL(imported, `file://${lintPath}`).pathname;
          } else {
            const transitive = TRANSITIVE_TYPE_DEPENDENCIES.find(
              ({ urlPattern }) => urlPattern.test(importedUrl.href),
            );
            if (!transitive) return;
            importedMonacoPath = new URL(transitive.virtualPath, monacoRoot)
              .href;
            importedLintPath = new URL(transitive.virtualPath, lintRoot)
              .pathname;
          }

          content = content.replaceAll(
            imported,
            relativeSpecifier(lintPath, importedLintPath),
          );
          await load(importedUrl, importedMonacoPath, importedLintPath);
        }),
      );
      declarations.push({ specifier, content, monacoPath, lintPath });
    })();
    loaded.set(url.href, request);
    return request;
  }

  await load(
    entryUrl,
    dependency.monacoPath,
    dependency.lintPath,
    dependency.specifier,
  );
  return declarations;
}

async function fetchLatestDeclarations(
  dependency: SourceTypePackage,
  signal: AbortSignal,
): Promise<SourceTypeDeclaration[]> {
  const packageResponse = await fetch(dependency.packageUrl, {
    cache: 'no-cache',
    signal,
  });
  if (!packageResponse.ok) {
    throw new Error(
      `unable to resolve the latest ${dependency.specifier} version: ${packageResponse.status}`,
    );
  }

  const typesUrl = packageResponse.headers.get('X-TypeScript-Types');
  if (!typesUrl) {
    throw new Error(
      `esm.sh did not provide ${dependency.specifier} declarations.`,
    );
  }
  return fetchDeclarationGraph(
    dependency,
    new URL(typesUrl, dependency.packageUrl),
    signal,
  );
}

function loadDeclaration(
  dependency: SourceTypePackage,
): Promise<SourceTypeDeclaration[]> {
  const cached = cachedDeclarations.get(dependency.specifier);
  if (cached) return Promise.resolve(cached);

  const pending = pendingDeclarations.get(dependency.specifier);
  if (pending) return pending.promise;

  const controller = new AbortController();
  const promise = fetchLatestDeclarations(dependency, controller.signal)
    .then((declarations) => {
      cachedDeclarations.set(dependency.specifier, declarations);
      return declarations;
    })
    .finally(() => {
      if (pendingDeclarations.get(dependency.specifier)?.promise === promise) {
        pendingDeclarations.delete(dependency.specifier);
      }
    });
  pendingDeclarations.set(dependency.specifier, { controller, promise });
  return promise;
}

/** Load declarations for the supported static imports present in source. */
export async function loadSourceTypes(
  specifiers: string[],
): Promise<SourceTypeEnvironment> {
  const dependencies = SOURCE_TYPE_PACKAGES.filter(({ specifier }) =>
    specifiers.includes(specifier),
  );
  return {
    key: sourceTypeKey(specifiers),
    declarations: (await Promise.all(dependencies.map(loadDeclaration))).flat(),
  };
}

/** Cancel in-flight downloads for imports that are no longer present. */
export function cancelUnusedSourceTypeLoads(specifiers: string[]) {
  const retained = new Set(specifiers);
  for (const [specifier, pending] of pendingDeclarations) {
    if (!retained.has(specifier)) {
      pending.controller.abort();
      pendingDeclarations.delete(specifier);
    }
  }
}

/** Add virtual declaration mappings without mutating the editor's tsconfig. */
export function addSourceTypeResolutions(
  tsconfig: Record<string, unknown>,
  environment: SourceTypeEnvironment,
): Record<string, unknown> {
  const compilerOptions =
    tsconfig.compilerOptions && typeof tsconfig.compilerOptions === 'object'
      ? (tsconfig.compilerOptions as Record<string, unknown>)
      : {};
  const paths =
    compilerOptions.paths && typeof compilerOptions.paths === 'object'
      ? (compilerOptions.paths as Record<string, unknown>)
      : {};

  return {
    ...tsconfig,
    compilerOptions: {
      ...compilerOptions,
      paths: {
        ...paths,
        ...Object.fromEntries(
          environment.declarations
            .filter(
              (
                declaration,
              ): declaration is SourceTypeDeclaration & {
                specifier: string;
              } => declaration.specifier !== undefined,
            )
            .map(({ specifier, lintPath }) => [specifier, [lintPath]]),
        ),
      },
    },
  };
}
