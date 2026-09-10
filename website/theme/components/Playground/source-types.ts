export interface SourceTypeDeclaration {
  specifier: string;
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

// Add another descriptor here when the Playground supports types for another
// source dependency. Packages are loaded only when their pattern matches.
const SOURCE_TYPE_PACKAGES: SourceTypePackage[] = [
  {
    specifier: '@rstest/core',
    packageUrl: 'https://esm.sh/@rstest/core',
    monacoPath: 'file:///node_modules/@rstest/core/index.d.ts',
    lintPath: '/source-types/rstest-core.d.ts',
    importPattern: /\bfrom\s*(['"])@rstest\/core\1/,
  },
];

const cachedDeclarations = new Map<string, SourceTypeDeclaration>();
const pendingDeclarations = new Map<
  string,
  { controller: AbortController; promise: Promise<SourceTypeDeclaration> }
>();

export function findSourceTypePackages(source: string): string[] {
  return SOURCE_TYPE_PACKAGES.filter(({ importPattern }) =>
    importPattern.test(source),
  ).map(({ specifier }) => specifier);
}

export function sourceTypeKey(specifiers: string[]): string {
  return specifiers.join('\0');
}

async function fetchLatestDeclaration(
  dependency: SourceTypePackage,
  signal: AbortSignal,
): Promise<SourceTypeDeclaration> {
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

  const typesResponse = await fetch(new URL(typesUrl, dependency.packageUrl), {
    signal,
  });
  if (!typesResponse.ok) {
    throw new Error(
      `unable to load ${dependency.specifier} declarations: ${typesResponse.status}`,
    );
  }

  return {
    specifier: dependency.specifier,
    content: await typesResponse.text(),
    monacoPath: dependency.monacoPath,
    lintPath: dependency.lintPath,
  };
}

function loadDeclaration(
  dependency: SourceTypePackage,
): Promise<SourceTypeDeclaration> {
  const cached = cachedDeclarations.get(dependency.specifier);
  if (cached) return Promise.resolve(cached);

  const pending = pendingDeclarations.get(dependency.specifier);
  if (pending) return pending.promise;

  const controller = new AbortController();
  const promise = fetchLatestDeclaration(dependency, controller.signal)
    .then((declaration) => {
      cachedDeclarations.set(dependency.specifier, declaration);
      return declaration;
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
    declarations: await Promise.all(dependencies.map(loadDeclaration)),
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
          environment.declarations.map(({ specifier, lintPath }) => [
            specifier,
            [lintPath],
          ]),
        ),
      },
    },
  };
}
