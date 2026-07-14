// scripts/generate-rule-option-types.mjs
//
// Scans internal/rules/*/*.schema.json and internal/plugins/*/rules/**/*.schema.json
// for native rules' options JSON Schemas, compiles each into a TypeScript
// type via json-schema-to-typescript, and injects the result into the built
// `dist/index.d.ts` at the `@__RULE_OPTIONS__` marker inside `RulesRecord`
// (see src/config/define-config.ts). Rules that haven't declared a schema
// yet (internal/rule.Rule.Schema == nil) simply have no *.schema.json file
// and keep falling back to RulesRecord's untyped index signature.
//
// Run after `rslib build` (dist/index.d.ts must already exist) as part of
// `pnpm build` — see the `generate:rule-types` script in package.json.
import { compile } from 'json-schema-to-typescript';
import { glob } from 'tinyglobby';
import fs from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const PACKAGE_ROOT = path.resolve(fileURLToPath(import.meta.url), '../..');
const REPO_ROOT = path.resolve(PACKAGE_ROOT, '../..');
const DIST_INDEX_DTS = path.join(PACKAGE_ROOT, 'dist/index.d.ts');
const MARKER = '/** @__RULE_OPTIONS__ */';

// Ported-plugin directory name (internal/plugins/<dir>) -> rule-ID prefix.
// Mirrors NATIVE_PLUGINS in src/config/define-config.ts.
const PLUGIN_PREFIXES = {
  typescript: '@typescript-eslint',
  import: 'import',
  react: 'react',
  react_hooks: 'react-hooks',
  jest: 'jest',
  jsx_a11y: 'jsx-a11y',
  promise: 'promise',
  unicorn: 'unicorn',
};

const SCHEMA_GLOBS = [
  'internal/rules/*/*.schema.json',
  'internal/plugins/*/rules/**/*.schema.json',
];

/**
 * Resolves a `<rule-name>.schema.json` file's path (relative to the repo
 * root, forward-slash separated) to its full rule ID, per the plugin-prefix
 * convention documented in .agents/skills/port-rule/references/PORT_RULE.md.
 */
export function ruleIdFromSchemaPath(relativePath) {
  const ruleName = path.basename(relativePath, '.schema.json');
  const pluginMatch = relativePath.match(
    /^internal\/plugins\/([^/]+)\/rules\//,
  );
  if (!pluginMatch) {
    return ruleName;
  }
  const prefix = PLUGIN_PREFIXES[pluginMatch[1]];
  if (!prefix) {
    throw new Error(
      `generate-rule-option-types: unknown plugin directory for schema ` +
        `${relativePath}: ${pluginMatch[1]}`,
    );
  }
  return `${prefix}/${ruleName}`;
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

async function findSchemaFiles() {
  const files = await glob(SCHEMA_GLOBS, { cwd: REPO_ROOT });
  return files.sort();
}

async function compileRuleOptionTypes() {
  const files = await findSchemaFiles();

  const typeDeclarations = [];
  const recordProperties = [];

  for (const relativePath of files) {
    const ruleId = ruleIdFromSchemaPath(relativePath);
    const typeName = `${ruleIdToTypeName(ruleId)}Options`;
    const schema = JSON.parse(
      await fs.readFile(path.join(REPO_ROOT, relativePath), 'utf-8'),
    );
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
 * named properties land just before `RulesRecord`'s `@__RULE_OPTIONS__`-
 * marked index signature (so they take precedence over the fallback
 * `any[]` shape), and the type declarations land right after the
 * interface's closing brace — not just above it, which would wedge them
 * between the interface's own doc comment and its declaration.
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
  // with the fallback index signature already in the bundled output. The
  // marker line itself (from markerLineStart onward) is left untouched and
  // kept as-is — splicing in new lines *before* it rather than doing a
  // MARKER substring `.replace()` avoids double-counting that line's own
  // pre-existing indentation.
  const markerLineStart = dts.lastIndexOf('\n', markerIndex) + 1;
  const indent = dts.slice(markerLineStart, markerIndex);
  const propertiesBlock = recordProperties
    .map((property) => `${indent}${property}\n`)
    .join('');

  const withProperties =
    dts.slice(0, markerLineStart) +
    propertiesBlock +
    dts.slice(markerLineStart);

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
