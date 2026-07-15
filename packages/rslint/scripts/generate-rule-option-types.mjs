// scripts/generate-rule-option-types.mjs
//
// Dumps every registered native rule's options JSON Schema from the host
// `rslint` binary's hidden `--dump-rule-schemas` flag (see
// cmd/rslint/dump_rule_schemas.go — it walks internal/config.GlobalRuleRegistry,
// the single source of truth for rule IDs, prefixes, and declared schemas),
// compiles each schema into a TypeScript type via json-schema-to-typescript,
// and injects the result into the built `dist/index.d.ts` at the
// `@__RULE_OPTIONS__` marker inside `RulesRecord` (see
// src/config/define-config.ts). Rules that haven't declared a schema yet
// (internal/rule.Rule.Schema == nil) are omitted by the Go side and keep
// falling back to RulesRecord's untyped index signature.
//
// Requires the host binary already placed by `pnpm build:bin` (see
// scripts/place-host-build.mjs) — this script does not compile Go itself, so
// it errors out with a pointer to that command if the binary is missing.
//
// Run after `rslib build` (dist/index.d.ts must already exist) as part of
// `pnpm build` — see the `build:types` script in package.json.
import { compile } from 'json-schema-to-typescript';
import { execFileSync } from 'node:child_process';
import fs from 'node:fs/promises';
import { existsSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { hostTuple } from '../../../scripts/place-host-build.mjs';

const PACKAGE_ROOT = path.resolve(fileURLToPath(import.meta.url), '../..');
const REPO_ROOT = path.resolve(PACKAGE_ROOT, '../..');
const DIST_INDEX_DTS = path.join(PACKAGE_ROOT, 'dist/index.d.ts');
const MARKER = '/** @__RULE_OPTIONS__ */';

/**
 * Resolves the host `rslint` binary placed by `place-host-build.mjs bin`
 * (`pnpm build:bin`) under `npm/rslint/<tuple>/`. Throws rather than falling
 * back to `go run` — the binary is expected to already exist by the time
 * `build:types` runs in the `build` script chain
 * (`build:bin && build:js && build:types`), and CI jobs that only
 * download a prebuilt binary (no Go toolchain, e.g. test-vscode-windows)
 * rely on that same placement.
 */
function resolveHostBinary() {
  const ext = process.platform === 'win32' ? '.exe' : '';
  const bin = path.join(REPO_ROOT, 'npm', 'rslint', hostTuple(), `rslint${ext}`);
  if (!existsSync(bin)) {
    throw new Error(
      `generate-rule-option-types: host rslint binary not found at ${bin} — ` +
        'run `pnpm build:bin` first',
    );
  }
  return bin;
}

/**
 * Runs `<rslint> --dump-rule-schemas`, which registers every native rule and
 * returns `{name, schema}` for each one that declares an options JSON
 * Schema — including a rule that only references the shared
 * `rule.EmptyArraySchema` (no options, no on-disk `.schema.json` file),
 * which a filesystem scan of `*.schema.json` alone can't see.
 *
 * @returns {{ name: string, schema: import('json-schema').JSONSchema4 }[]}
 */
export function collectRuleSchemas() {
  const output = execFileSync(resolveHostBinary(), ['--dump-rule-schemas'], {
    encoding: 'utf-8',
    maxBuffer: 64 * 1024 * 1024,
  });
  return JSON.parse(output);
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

async function compileRuleOptionTypes() {
  const rules = collectRuleSchemas();

  const typeDeclarations = [];
  const recordProperties = [];

  for (const { name: ruleId, schema } of rules) {
    if (isEmptyArraySchema(schema)) {
      recordProperties.push(`${JSON.stringify(ruleId)}?: RuleEntry<[]>;`);
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
      `${JSON.stringify(ruleId)}?: RuleEntry<${typeName}>;`,
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
    throw new Error(
      `generate-rule-option-types: couldn't find the ${MARKER} marker in ` +
        'dist/index.d.ts — has RulesRecord moved or lost its marker comment ' +
        'in src/config/define-config.ts?',
    );
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
    .map((property) => `${indent}${property}\n`)
    .join('');

  const withProperties =
    dts.slice(0, markerLineStart) +
    propertiesBlock +
    dts.slice(markerLineEnd);

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

async function main() {
  const generated = await compileRuleOptionTypes();
  const dts = await fs.readFile(DIST_INDEX_DTS, 'utf-8');
  await fs.writeFile(DIST_INDEX_DTS, injectIntoDts(dts, generated));
  console.log(
    `generate-rule-option-types: injected ${generated.recordProperties.length} ` +
      'typed rule(s) into dist/index.d.ts',
  );
}

if (fileURLToPath(import.meta.url) === process.argv[1]) {
  main().catch((err) => {
    console.error(err);
    process.exitCode = 1;
  });
}
