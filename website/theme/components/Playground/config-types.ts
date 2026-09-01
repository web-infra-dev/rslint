import type * as monaco from 'monaco-editor';
import { javascriptDefaults } from 'monaco-editor/languages/features/typescript/register';

const CORE_URL = 'https://esm.sh/@rslint/core@';
const CORE_TYPES_PATH = 'file:///node_modules/@rslint/core/index.d.ts';
const TYPE_IMPORT = /(?:import\(|from\s+)["'](\.{1,2}\/[^"']+\.d\.ts)["']/g;

let installedVersion: string | undefined;
let installedLibraries: monaco.IDisposable[] = [];
let requestId = 0;

function typeUrl(version: string) {
  return `${CORE_URL}${encodeURIComponent(version)}/dist/index.d.ts`;
}

async function loadTypeLibraries(
  url: string,
  path: string,
  libraries: Array<{ content: string; path: string }>,
) {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`unable to load @rslint/core types: ${response.status}`);
  }

  const content = await response.text();
  libraries.push({ content, path });

  const imports = [...content.matchAll(TYPE_IMPORT)].map((match) => match[1]);
  await Promise.all(
    imports.map((specifier) => {
      const dependencyUrl = new URL(specifier, url).href;
      const dependencyPath = new URL(specifier, path).href;
      return loadTypeLibraries(dependencyUrl, dependencyPath, libraries);
    }),
  );
}

/** Install exact-version @rslint/core declarations for rslint.config.js. */
export async function installRslintCoreTypes(version: string) {
  const currentRequestId = ++requestId;
  if (installedVersion === version) return;

  const libraries: Array<{ content: string; path: string }> = [];
  await loadTypeLibraries(typeUrl(version), CORE_TYPES_PATH, libraries);
  if (currentRequestId !== requestId) return;

  for (const library of installedLibraries) library.dispose();
  installedLibraries = libraries.map(({ content, path }) =>
    javascriptDefaults.addExtraLib(content, path),
  );
  installedVersion = version;
}
