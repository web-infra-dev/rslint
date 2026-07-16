// scripts/generate-rule-option-types.mjs
//
// Reads every registered native rule's options JSON Schema from
// packages/rslint/rule-schemas.json — a `{name, schema}[]` dump produced by
// tools/dump-rule-schemas (which walks internal/config.GlobalRuleRegistry, the
// single source of truth for rule IDs, prefixes, and declared schemas; see
// scripts/place-host-build.mjs's `bin` mode, which writes this file, and CI's
// per-workflow equivalents for jobs without a Go toolchain), compiles each
// schema into a TypeScript type via json-schema-to-typescript, and injects
// the result into a built `dist/index.d.ts` at the `@__RULE_OPTIONS__`
// marker inside `RulesRecord` (see packages/rslint/src/config/define-config.ts).
// Rules that haven't declared a schema yet (internal/rule.Rule.Schema == nil)
// are omitted by the Go side and keep falling back to RulesRecord's untyped
// index signature.
//
// Invoked automatically via the `generate-rule-option-types` rslib plugin's
// `onAfterBuild` hook (see packages/rslint/rslib.config.ts) as part of
// `pnpm --filter @rslint/core build:js` — no separate `build:types` script.
import { compile } from 'json-schema-to-typescript';
import fs from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const REPO_ROOT = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  '..',
);
const DEFAULT_SCHEMAS_PATH = path.join(
  REPO_ROOT,
  'packages/rslint/rule-schemas.json',
);
const DEFAULT_DIST_INDEX_DTS = path.join(
  REPO_ROOT,
  'packages/rslint/dist/index.d.ts',
);
const MARKER = '/** @__RULE_OPTIONS__ */';

/**
 * Reads the `{name, schema}[]` dump written by `tools/dump-rule-schemas` —
 * including rules that only reference the shared `rule.EmptyArraySchema`
 * (no options, no on-disk `.schema.json` file), which a filesystem scan of
 * `*.schema.json` alone can't see.
 *
 * @returns {Promise<{ name: string, schema: import('json-schema').JSONSchema4 }[]>}
 */
export async function collectRuleSchemas(
  schemasPath = DEFAULT_SCHEMAS_PATH,
) {
  let raw;
  try {
    raw = await fs.readFile(schemasPath, 'utf-8');
  } catch (err) {
    if (err.code === 'ENOENT') {
      const notFound = new Error(
        `generate-rule-option-types: rule schemas dump not found at ${schemasPath} — ` +
          'run `pnpm build:bin` first (or, in CI jobs without a Go toolchain, ' +
          'fetch the prebuilt rule-schemas artifact)',
      );
      notFound.code = 'RULE_SCHEMAS_NOT_FOUND';
      throw notFound;
    }
    throw err;
  }
  return JSON.parse(raw);
}

/**
 * Converts a rule ID into a unique PascalCase TypeScript identifier, e.g.
 * `no-console` -> `NoConsole`, `@typescript-eslint/no-unused-vars` ->
 * `TypescriptEslintNoUnusedVars`. Rule IDs are unique, so the identifiers
 * derived from them are too.
 */
export function ruleIdToTypeName(ruleId) {
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
function isEmptyArraySchema(schema) {
  return (
    schema !== null &&
    typeof schema === 'object' &&
    Object.keys(schema).length === 2 &&
    schema.type === 'array' &&
    schema.maxItems === 0
  );
}

const SCOPE_TO_URL_NAME = {
  '@typescript-eslint': 'typescript-eslint',
  react: 'react',
  'react-hooks': 'react-hooks',
  import: 'import',
  promise: 'promise',
  jest: 'jest',
  unicorn: 'unicorn',
  'jsx-a11y': 'jsx-a11y',
};

function getRuleDocUrl(ruleId) {
  if (ruleId.includes('/')) {
    const parts = ruleId.split('/');
    const scope = parts[0];
    const name = parts.slice(1).join('/');
    const urlName = SCOPE_TO_URL_NAME[scope] || scope.replace(/^@/, '');
    return `https://rslint.rs/rules/${urlName}/${name}`;
  } else {
    return `https://rslint.rs/rules/eslint/${ruleId}`;
  }
}

async function compileRuleOptionTypes(schemasPath) {
  const rules = await collectRuleSchemas(schemasPath);

  const typeDeclarations = [];
  const recordProperties = [];

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
 * named properties land where `RulesRecord`'s `@__RULE_OPTIONS__`-marked
 * index signature was (so they take precedence over the fallback `any[]`
 * shape), and the type declarations land right after the interface's
 * closing brace — not just above it, which would wedge them between the
 * interface's own doc comment and its declaration. The marker comment
 * itself is build-time-only wiring and is dropped, not shipped in the
 * published `.d.ts`.
 */
export function injectIntoDts(dts, { typeDeclarations, recordProperties }) {
  const markerIndex = dts.indexOf(MARKER);
  if (markerIndex === -1) {
    return dts;
  }
  const interfaceIndex = dts.lastIndexOf('interface RulesRecord', markerIndex);
  if (interfaceIndex === -1) {
    throw new Error(
      "generate-rule-option-types: couldn't find `interface RulesRecord` " +
        'before the marker in dist/index.d.ts',
    );
  }
  // Match the marker line's own indentation so injected properties line up
  // with the fallback index signature already in the bundled output, then
  // replace the whole marker line (comment included) with the injected
  // properties rather than splicing in front of it, so the marker never
  // survives into the output.
  const markerLineStart = dts.lastIndexOf('\n', markerIndex) + 1;
  const markerLineEnd = dts.indexOf('\n', markerIndex) + 1;
  const indent = dts.slice(markerLineStart, markerIndex);
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
    dts.slice(0, markerLineStart) + propertiesBlock + dts.slice(markerLineEnd);

  if (!typeDeclarations.length) {
    return withProperties;
  }

  // Find the interface's closing brace by depth-counting braces from its
  // opening one (the body is a flat index signature, so this never nests,
  // but counting depth keeps it correct if that ever changes).
  const braceOpen = withProperties.indexOf('{', interfaceIndex);
  let depth = 0;
  let braceClose = -1;
  for (let i = braceOpen; i < withProperties.length; i++) {
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
 * Reads `distIndexDtsPath`, splices in the generated rule-option types, and
 * writes it back. Called by the rslib `onAfterBuild` plugin
 * (packages/rslint/rslib.config.ts) once `dist/index.d.ts` exists, and by
 * this file's own CLI entrypoint for manual/debug runs.
 */
export async function generateRuleOptionTypes({
  distIndexDtsPath = DEFAULT_DIST_INDEX_DTS,
  schemasPath = DEFAULT_SCHEMAS_PATH,
} = {}) {
  const generated = await compileRuleOptionTypes(schemasPath);
  const dts = await fs.readFile(distIndexDtsPath, 'utf-8');
  await fs.writeFile(distIndexDtsPath, injectIntoDts(dts, generated));
  return generated.recordProperties.length;
}

async function main() {
  const count = await generateRuleOptionTypes();
  console.log(
    `generate-rule-option-types: injected ${count} typed rule(s) into dist/index.d.ts`,
  );
}

if (import.meta.main) {
  main().catch((err) => {
    console.error(err);
    process.exitCode = 1;
  });
}
