import { defineConfig, type RsbuildPlugin } from '@rslib/core';
import { compile } from 'json-schema-to-typescript';
import fs from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const PACKAGE_ROOT = path.dirname(fileURLToPath(import.meta.url));
const SCHEMAS_PATH = path.join(PACKAGE_ROOT, 'rule-schemas.json');
const DIST_INDEX_DTS = path.join(PACKAGE_ROOT, 'dist/index.d.ts');

interface RuleSchemaEntry {
  name: string;
  // JSON Schema (draft-07), as compiled by json-schema-to-typescript.
  schema: any;
}

/**
 * Reads the `{name, schema}[]` dump written by `tools/dump_rule_schemas` —
 * including rules that only reference the shared `rule.EmptyArraySchema`
 * (no options, no on-disk `.schema.json` file), which a filesystem scan of
 * `*.schema.json` alone can't see. Returns null if the dump hasn't been
 * produced — it isn't generated automatically, so a missing one just means
 * typed rule options get skipped for this build.
 */
async function collectRuleSchemas(): Promise<RuleSchemaEntry[] | null> {
  try {
    return JSON.parse(await fs.readFile(SCHEMAS_PATH, 'utf-8'));
  } catch (err) {
    if ((err as NodeJS.ErrnoException).code === 'ENOENT') return null;
    throw err;
  }
}

/**
 * Converts a rule ID into a unique PascalCase TypeScript identifier, e.g.
 * `no-console` -> `NoConsole`, `@typescript-eslint/no-unused-vars` ->
 * `TypescriptEslintNoUnusedVars`. Rule IDs are unique, so the identifiers
 * derived from them are too.
 */
function ruleIdToTypeName(ruleId: string): string {
  return ruleId
    .split(/[^a-zA-Z0-9]+/)
    .filter(Boolean)
    .map((word) => word[0].toUpperCase() + word.slice(1))
    .join('');
}

/**
 * True for the shared `rule.EmptyArraySchema` (internal/rule/schema.go) that
 * no-options rules like `no-debugger` reference directly. Special-cased so
 * it maps straight to `RuleEntry<[]>` instead of round-tripping through
 * json-schema-to-typescript for a named `FooOptions = []` type nothing else
 * needs.
 */
function isEmptyArraySchema(schema: any): boolean {
  return (
    schema !== null &&
    typeof schema === 'object' &&
    Object.keys(schema).length === 2 &&
    schema.type === 'array' &&
    schema.maxItems === 0
  );
}

function getRuleDocUrl(ruleId: string): string {
  if (ruleId.includes('/')) {
    const parts = ruleId.split('/');
    const scope = parts[0];
    const name = parts.slice(1).join('/');
    const urlName = scope.replace(/^@/, '');
    return `https://rslint.rs/rules/${urlName}/${name}`;
  } else {
    return `https://rslint.rs/rules/eslint/${ruleId}`;
  }
}

async function compileRuleOptionTypes(rules: RuleSchemaEntry[]) {
  const typeDeclarations: string[] = [];
  const recordProperties: string[] = [];

  for (const { name: ruleId, schema } of rules) {
    const url = getRuleDocUrl(ruleId);
    const comment = `/**\n * @see ${url}\n */\n`;

    if (isEmptyArraySchema(schema)) {
      recordProperties.push(
        `${comment}${JSON.stringify(ruleId)}?: RuleEntry<[]>;`,
      );
      continue;
    }

    const typeName = `${ruleIdToTypeName(ruleId)}Options`;
    const ts = await compile(schema, typeName, {
      bannerComment: '',
      additionalProperties: false,
      style: { semi: true },
    });
    typeDeclarations.push(ts.trim());
    recordProperties.push(
      `${comment}${JSON.stringify(ruleId)}?: RuleEntry<${typeName}>;`,
    );
  }

  return { typeDeclarations, recordProperties };
}

/**
 * Splices generated rule-option types into a built `dist/index.d.ts`: the
 * named properties land right inside `RulesRecord`'s opening brace, ahead of
 * its fallback index signature (so they take precedence over it), and the
 * type declarations land right after the interface's closing brace — not
 * just above it, which would wedge them between the interface's own doc
 * comment and its declaration.
 */
function injectIntoDts(
  dts: string,
  {
    typeDeclarations,
    recordProperties,
  }: { typeDeclarations: string[]; recordProperties: string[] },
): string {
  const interfaceIndex = dts.indexOf('interface RulesRecord');
  if (interfaceIndex === -1) {
    throw new Error(
      "generate-rule-option-types: couldn't find `interface RulesRecord` " +
        'in dist/index.d.ts',
    );
  }
  // Match the existing index signature line's indentation so injected
  // properties line up with it.
  const braceOpen = dts.indexOf('{', interfaceIndex);
  const bodyStart = dts.indexOf('\n', braceOpen) + 1;
  const indent = dts.slice(bodyStart).match(/^[ \t]*/)![0];
  const propertiesBlock = recordProperties
    .map(
      (property) =>
        property
          .split('\n')
          .map((line) => (line ? `${indent}${line}` : ''))
          .join('\n') + '\n',
    )
    .join('');

  const withProperties =
    dts.slice(0, bodyStart) + propertiesBlock + dts.slice(bodyStart);

  if (!typeDeclarations.length) {
    return withProperties;
  }

  // Find the interface's closing brace by depth-counting braces from its
  // opening one (the body is a flat index signature, so this never nests,
  // but counting depth keeps it correct if that ever changes).
  const reopenedBrace = withProperties.indexOf('{', interfaceIndex);
  let depth = 0;
  let braceClose = -1;
  for (let i = reopenedBrace; i < withProperties.length; i++) {
    if (withProperties[i] === '{') depth++;
    else if (withProperties[i] === '}') {
      depth--;
      if (depth === 0) {
        braceClose = i;
        break;
      }
    }
  }
  let insertAt = braceClose + 1;
  if (withProperties[insertAt] === '\r') insertAt++;
  if (withProperties[insertAt] === '\n') insertAt++;

  return (
    withProperties.slice(0, insertAt) +
    '\n' +
    typeDeclarations.join('\n\n') +
    '\n' +
    withProperties.slice(insertAt)
  );
}

/**
 * Splices typed rule options into `dist/index.d.ts` once the whole
 * (multi-`lib`-entry) build finishes — `onAfterBuild` fires once for the
 * entire rslib build, not per entry, so by the time it runs `librarySurface`'s
 * `dist/index.d.ts` already exists.
 *
 * Reads every registered native rule's options JSON Schema from
 * `rule-schemas.json` — a `{name, schema}[]` dump produced by
 * `tools/dump_rule_schemas` (which walks internal/config.GlobalRuleRegistry,
 * the single source of truth for rule IDs, prefixes, and declared schemas).
 * That dump isn't produced automatically, so a missing one just skips with a
 * warning rather than failing the build. Rules that haven't declared a schema
 * yet (internal/rule.Rule.Schema == nil) are omitted by the Go side and keep
 * falling back to `RulesRecord`'s untyped index signature
 * (see packages/rslint/src/config/define-config.ts).
 */
const generateRuleOptionTypesPlugin = (): RsbuildPlugin => ({
  name: 'generate-rule-option-types',
  setup(api) {
    api.onAfterBuild(async () => {
      const rules = await collectRuleSchemas();
      if (!rules) {
        api.logger.warn(
          'generate-rule-option-types: skipped — rule schemas dump not ' +
            `found at ${SCHEMAS_PATH} (run \`go run ./tools/dump_rule_schemas > ` +
            'packages/rslint/rule-schemas.json` first)',
        );
        return;
      }

      const generated = await compileRuleOptionTypes(rules);
      const dts = await fs.readFile(DIST_INDEX_DTS, 'utf-8');
      await fs.writeFile(DIST_INDEX_DTS, injectIntoDts(dts, generated));
      api.logger.log(
        `generate-rule-option-types: injected ${generated.recordProperties.length} typed rule(s) into dist/index.d.ts`,
      );
    });
  },
});

/**
 * Single rslib build for all of `@rslint/core`'s JS: the public library surface
 * plus the internal `eslint-plugin` worker runtime. Replaces the former split
 * (tsgo `build:js` + rslib `build:worker`) — one `build:js` now emits both.
 *
 * Two groups of `lib` blocks:
 *
 * 1. Library surface → `dist/` (`tsconfig.lib.json`, which inherits root's
 *    exclude of `src/eslint-plugin/**`). A dts build is a TS project, so it must
 *    not share its `tsBuildInfoFile` with the tsgo `typecheck` over the same
 *    `src` — the two tools' incremental formats clash. Hence a tsconfig per
 *    consumer: `tsconfig.lib.json` (here), `tsconfig.worker.json` (below), and
 *    `tsconfig.build.json` (typecheck). `autoExternal` externalizes `dependencies`
 *    (`picomatch`) + `peerDependencies` (`jiti`); `tinyglobby` is a devDep so it
 *    bundles in. But `tinyglobby`'s `fdir` loads `picomatch` via `createRequire`,
 *    which rspack can't statically follow — so `picomatch` can't be bundled away
 *    and stays a runtime dep. One `lib` block with all entries: the surface
 *    modules share a graph, so shared chunks between entries are fine here.
 *
 * 2. eslint-plugin worker → `dist/eslint-plugin/` (`tsconfig.worker.json`,
 *    which includes `src/eslint-plugin/**`). Each entry is its own `lib` block
 *    so Rspack inlines each output's full module graph with NO shared chunks —
 *    crucial for the worker (`new Worker(...)` spawns a fresh V8 isolate that
 *    pays a filesystem-open + parse cost per extra chunk; modules can't be
 *    reused across isolates). The ESLint-compat libs (`@typescript-eslint/
 *    scope-manager`, `eslint-scope`, `esquery`) are devDependencies imported
 *    statically so they bundle in; consumers need none at runtime. The native
 *    parser loader (`src/eslint-plugin/native/load-binding.ts`) bundles in too,
 *    but the platform `.node` it loads stays external: the loader selects the
 *    `@rslint/native-<tuple>` package at runtime via `createRequire`, which
 *    rspack can't statically follow (so the binary is never inlined — intended).
 */
const librarySurface = {
  format: 'esm' as const,
  bundle: true,
  autoExternal: true,
  output: {
    target: 'node' as const,
    distPath: { root: './dist' },
  },
  source: {
    tsconfigPath: './tsconfig.lib.json',
    entry: {
      index: './src/index.ts',
      service: './src/service/service.ts',
      internal: './src/internal/node.ts',
      'config-loader': './src/config/config-loader.ts',
      cli: './src/cli/cli.ts',
    },
  },
  dts: { bundle: true },
  plugins: [generateRuleOptionTypesPlugin()],
};

const workerBase = {
  format: 'esm' as const,
  bundle: true,
  autoExternal: true,
  output: {
    target: 'node' as const,
    distPath: { root: './dist/eslint-plugin' },
  },
  source: {
    tsconfigPath: './tsconfig.worker.json',
  },
};

export default defineConfig({
  lib: [
    librarySurface,
    {
      ...workerBase,
      source: {
        ...workerBase.source,
        entry: { index: './src/eslint-plugin/index.ts' },
      },
      // Bundle dts only on the main entry — the others re-export from `index`
      // or are tiny standalone modules; per-entry dts would duplicate types.
      dts: { bundle: true },
    },
    {
      ...workerBase,
      source: {
        ...workerBase.source,
        entry: { 'lint-worker': './src/eslint-plugin/lint-worker.ts' },
      },
      dts: false,
    },
    {
      ...workerBase,
      source: {
        ...workerBase.source,
        entry: { types: './src/eslint-plugin/types.ts' },
      },
      dts: { bundle: true },
    },
  ],
});
