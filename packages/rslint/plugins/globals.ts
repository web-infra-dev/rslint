import type { RsbuildPlugin, Rspack } from 'rstack/lib';
import fs from 'node:fs';
import fsPromises from 'node:fs/promises';
import { createRequire } from 'node:module';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const EMIT_PLUGIN_NAME = 'emit-globals-assets';
const BUNDLE_TYPES_PLUGIN_NAME = 'bundle-globals-types';
const ASSET_DIRECTORY = 'globals';
const require = createRequire(import.meta.url);
const globalsJsonPath = require.resolve('globals/globals.json');
const packageRoot = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  '..',
);
const distIndexDts = path.join(packageRoot, 'dist/index.d.ts');
const distGlobalsDts = path.join(packageRoot, 'dist/globals/index.d.ts');
const globalsPackageRoot = path.dirname(
  require.resolve('globals/package.json'),
);
const globalsDts = path.join(globalsPackageRoot, 'index.d.ts');
const globalsPackageJson = path.join(globalsPackageRoot, 'package.json');

const IMPORT_MARKER = "import globalsCatalog = require('globals');";
const ALIAS_MARKER =
  'declare type UpstreamGlobalsCatalog = typeof globalsCatalog;';
const UPSTREAM_EXPORT = 'declare const globals: Globals;\n\nexport = globals;';
const INTERNAL_TYPE_REFERENCE =
  "declare type UpstreamGlobalsCatalog = import('./globals/index.js').Globals;";

type GlobalsCatalog = Record<string, unknown>;

function isPlainRecord(value: unknown): value is Record<string, unknown> {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    return false;
  }
  const prototype = Object.getPrototypeOf(value);
  return prototype === Object.prototype || prototype === null;
}

function loadCatalog(): GlobalsCatalog {
  const parsed: unknown = JSON.parse(fs.readFileSync(globalsJsonPath, 'utf8'));
  if (!isPlainRecord(parsed)) {
    throw new TypeError(
      `${EMIT_PLUGIN_NAME}: globals.json must contain an object`,
    );
  }

  // This is the only validation the emitter needs: top-level keys become file
  // names. The contents stay exactly as provided by the pinned dependency.
  for (const name of Object.keys(parsed)) {
    if (!/^[A-Za-z0-9-]+$/.test(name)) {
      throw new TypeError(
        `${EMIT_PLUGIN_NAME}: unsafe global-set asset name ${JSON.stringify(name)}`,
      );
    }
  }

  return parsed as GlobalsCatalog;
}

/** Mechanically splits the upstream catalog into one JSON asset per set. */
export function emitGlobalsAssetsPlugin(): Rspack.Plugin {
  return {
    apply(compiler: Rspack.Compiler) {
      compiler.hooks.thisCompilation.tap(
        EMIT_PLUGIN_NAME,
        (compilation: Rspack.Compilation) => {
          compilation.fileDependencies.add(globalsJsonPath);
          compilation.hooks.processAssets.tap(
            {
              name: EMIT_PLUGIN_NAME,
              stage:
                compiler.webpack.Compilation.PROCESS_ASSETS_STAGE_ADDITIONAL,
            },
            () => {
              const catalog = loadCatalog();
              for (const [name, globalSet] of Object.entries(catalog)) {
                compilation.emitAsset(
                  `${ASSET_DIRECTORY}/${name}.json`,
                  new compiler.webpack.sources.RawSource(
                    `${JSON.stringify(globalSet)}\n`,
                  ),
                );
              }
            },
          );
        },
      );
    },
  };
}

function replaceOnce(
  source: string,
  search: string,
  replacement: string,
): string {
  const first = source.indexOf(search);
  if (first === -1 || source.indexOf(search, first + search.length) !== -1) {
    throw new Error(
      `${BUNDLE_TYPES_PLUGIN_NAME}: expected exactly one ${JSON.stringify(search)} marker`,
    );
  }
  return (
    source.slice(0, first) + replacement + source.slice(first + search.length)
  );
}

/** Bundles the upstream declarations into a private package-local module. */
export function bundleGlobalsTypesPlugin(): RsbuildPlugin {
  return {
    name: BUNDLE_TYPES_PLUGIN_NAME,
    setup(api) {
      api.onAfterBuild(async () => {
        const [distDts, upstreamDts, packageJsonText] = await Promise.all([
          fsPromises.readFile(distIndexDts, 'utf8'),
          fsPromises.readFile(globalsDts, 'utf8'),
          fsPromises.readFile(globalsPackageJson, 'utf8'),
        ]);
        const packageJson: unknown = JSON.parse(packageJsonText);
        const version =
          packageJson !== null &&
          typeof packageJson === 'object' &&
          'version' in packageJson &&
          typeof packageJson.version === 'string'
            ? packageJson.version
            : undefined;
        if (version === undefined) {
          throw new TypeError(
            `${BUNDLE_TYPES_PLUGIN_NAME}: globals has no valid version`,
          );
        }

        const normalizedUpstream = upstreamDts.replaceAll('\r\n', '\n').trim();
        if (!normalizedUpstream.endsWith(UPSTREAM_EXPORT)) {
          throw new Error(
            `${BUNDLE_TYPES_PLUGIN_NAME}: unsupported globals declaration-file shape`,
          );
        }
        const declarations = normalizedUpstream
          .slice(0, -UPSTREAM_EXPORT.length)
          .trimEnd();
        const bundledDeclarations = [
          `// Types bundled from globals ${version}; see THIRD-PARTY-NOTICES.md.`,
          declarations,
          '',
          'export type { Globals };',
          '',
        ].join('\n');

        let output = replaceOnce(
          distDts.replaceAll('\r\n', '\n'),
          `${IMPORT_MARKER}\n\n`,
          '',
        );
        output = replaceOnce(output, ALIAS_MARKER, INTERNAL_TYPE_REFERENCE);
        await fsPromises.mkdir(path.dirname(distGlobalsDts), {
          recursive: true,
        });
        await Promise.all([
          fsPromises.writeFile(distIndexDts, output),
          fsPromises.writeFile(distGlobalsDts, bundledDeclarations),
        ]);
      });
    },
  };
}
