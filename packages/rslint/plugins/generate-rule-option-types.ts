// Adds typed options for the frozen rsbuild rule set to dist/index.d.ts.
// The committed JSON preserves schema metadata from the original rslint
// baseline. Its provenance records the source commit and content hashes;
// normal builds need neither the old schema dumper nor a compiler service.
import type { RsbuildPlugin } from 'rstack/lib';
import { compile } from 'json-schema-to-typescript';
import fs from 'node:fs/promises';
import { createHash } from 'node:crypto';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const PACKAGE_ROOT = path.join(fileURLToPath(import.meta.url), '..', '..');
const SCHEMAS_PATH = path.join(PACKAGE_ROOT, 'rust-rule-schemas.json');
const PROVENANCE_PATH = path.join(
  PACKAGE_ROOT,
  'rust-rule-schemas.provenance.json',
);
const DIST_INDEX_DTS = path.join(PACKAGE_ROOT, 'dist/index.d.ts');

interface RuleSchemaEntry {
  name: string;
  // JSON Schema (draft-04), as compiled by json-schema-to-typescript.
  schema: any;
}

/** Read and verify the committed scoped schemas; a missing input fails builds. */
export async function collectRuleSchemas(
  schemasPath = SCHEMAS_PATH,
  provenancePath = PROVENANCE_PATH,
): Promise<RuleSchemaEntry[]> {
  const [bytes, provenanceText] = await Promise.all([
    fs.readFile(schemasPath),
    fs.readFile(provenancePath, 'utf8'),
  ]);
  const provenance = JSON.parse(provenanceText);
  const rules = JSON.parse(bytes.toString('utf8')) as RuleSchemaEntry[];
  const digest = createHash('sha256').update(bytes).digest('hex');
  if (
    !Array.isArray(rules) ||
    rules.length === 0 ||
    rules.length !== provenance.schemas?.count ||
    digest !== provenance.schemas?.sha256 ||
    bytes.length !== provenance.schemas?.bytes ||
    rules.some(
      (rule) =>
        !rule || typeof rule.name !== 'string' || !rule.name || !rule.schema,
    ) ||
    new Set(rules.map((rule) => rule.name)).size !== rules.length
  ) {
    throw new Error(
      'Committed Rust rule option schemas do not match their provenance',
    );
  }
  return rules;
}

/**
 * Converts a rule ID into a readable PascalCase TypeScript identifier, e.g.
 * `no-console` -> `NoConsole`, `@typescript-eslint/no-unused-vars` ->
 * `TypescriptEslintNoUnusedVars`. Punctuation can collapse to the same name;
 * compilation below disambiguates against all emitted declarations.
 */
function ruleIdToTypeName(ruleId: string): string {
  return ruleId
    .split(/[^a-zA-Z0-9]+/)
    .filter(Boolean)
    .map((word) => word[0].toUpperCase() + word.slice(1))
    .join('');
}

/**
 * True for the baseline's shared empty-array schema used by no-options rules
 * such as `no-debugger`. Special-cased so
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

/** Escapes a TypeScript identifier for use inside a RegExp. */
function escapeForRegExp(identifier: string): string {
  return identifier.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

/** Every top-level type/interface a compiled schema declares, in order. */
function declaredTypeNames(ts: string): string[] {
  return [...ts.matchAll(/^export (?:type|interface) ([A-Za-z0-9_$]+)/gm)].map(
    (match) => match[1],
  );
}

/**
 * Returns the name json-schema-to-typescript actually declared for the type it
 * was asked to call `requestedName`.
 *
 * It normalizes the name it is handed rather than emitting it verbatim: asked
 * for `JsxA11yAltTextOptions` it declares `JsxA11YAltTextOptions`. Referencing
 * the requested name would then point at a type no declaration defines, and
 * TypeScript silently widens that rule's options to `any` — so `RulesRecord`
 * must reference what was emitted, not what was asked for. Case is the only
 * thing that differs, which is why the match ignores it.
 */
function findDeclaredTypeName(
  ts: string,
  requestedName: string,
): string | null {
  return (
    declaredTypeNames(ts).find(
      (name) => name.toLowerCase() === requestedName.toLowerCase(),
    ) ?? null
  );
}

/**
 * Renames the helper types json-schema-to-typescript emits alongside a rule's
 * main type so they can't collide with another rule's.
 *
 * Helpers are named after the schema's `$defs` keys, which are chosen per rule
 * and repeat across rules — `no-restricted-imports` and `return-await` both
 * define `NameList`, and several jest rules define `Matcher`. Each rule
 * compiles independently, so every one of them emits a top-level
 * `export type NameList`, and the concatenated `.d.ts` then declares the same
 * name several times. Prefixing each helper with the rule's own type name
 * keeps them distinct; helpers are only referenced from within their own
 * rule's block, so renaming them there is complete.
 *
 * The main type is left untouched — it is the name `RulesRecord` references.
 */
function namespaceHelperTypes(ts: string, mainTypeName: string): string {
  let namespaced = ts;
  for (const name of new Set(declaredTypeNames(ts))) {
    if (name === mainTypeName) {
      continue;
    }
    namespaced = namespaced.replace(
      new RegExp(`\\b${escapeForRegExp(name)}\\b`, 'g'),
      `${mainTypeName}${name}`,
    );
  }
  return namespaced;
}

export async function compileRuleOptionTypes(rules: RuleSchemaEntry[]) {
  const typeDeclarations: string[] = [];
  const recordProperties: string[] = [];
  const usedTypeNames = new Set<string>();

  for (const { name: ruleId, schema } of [...rules].sort((a, b) =>
    a.name < b.name ? -1 : a.name > b.name ? 1 : 0,
  )) {
    const url = getRuleDocUrl(ruleId);
    const comment = `/**\n * @see ${url}\n */\n`;

    if (isEmptyArraySchema(schema)) {
      recordProperties.push(
        `${comment}${JSON.stringify(ruleId)}?: RuleEntry<[]>;`,
      );
      continue;
    }

    const baseName = `${ruleIdToTypeName(ruleId)}Options`;
    for (let suffix = 1; ; suffix++) {
      const requestedName = suffix === 1 ? baseName : `${baseName}${suffix}`;
      const ts = await compile(schema, requestedName, {
        bannerComment: '',
        style: { semi: true },
      });
      const typeName = findDeclaredTypeName(ts, requestedName);
      if (typeName === null) {
        throw new Error(
          `generate-rule-option-types: json-schema-to-typescript declared no type ` +
            `named like ${requestedName} for rule ${ruleId}; ` +
            `it declared ${declaredTypeNames(ts).join(', ') || '<nothing>'}`,
        );
      }
      const declaration = namespaceHelperTypes(ts.trim(), typeName);
      const names = declaredTypeNames(declaration);
      if (names.some((name) => usedTypeNames.has(name))) continue;
      names.forEach((name) => usedTypeNames.add(name));
      typeDeclarations.push(declaration);
      recordProperties.push(
        `${comment}${JSON.stringify(ruleId)}?: RuleEntry<${typeName}>;`,
      );
      break;
    }
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

export function generateRuleOptionTypesPlugin(): RsbuildPlugin {
  return {
    name: 'generate-rule-option-types',
    setup(api) {
      api.onAfterBuild(async () => {
        const rules = await collectRuleSchemas();

        const generated = await compileRuleOptionTypes(rules);
        const dts = await fs.readFile(DIST_INDEX_DTS, 'utf-8');
        await fs.writeFile(DIST_INDEX_DTS, injectIntoDts(dts, generated));
        api.logger.log(
          `generate-rule-option-types: injected ${generated.recordProperties.length} typed rule(s) into dist/index.d.ts`,
        );
      });
    },
  };
}
