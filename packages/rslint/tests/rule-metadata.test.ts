import { describe, expect, test } from 'rstack/test';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import ts from 'typescript';
import {
  collectRuleSchemas,
  compileRuleOptionTypes,
} from '../plugins/generate-rule-option-types.js';
import { createHash } from 'node:crypto';
import { fileURLToPath } from 'node:url';

function withFixture<T>(
  run: (root: string, write: (file: string, content: string) => void) => T,
): T {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-rule-metadata-'));
  const write = (file: string, content: string) => {
    const target = path.join(root, file);
    fs.mkdirSync(path.dirname(target), { recursive: true });
    fs.writeFileSync(target, content);
  };
  try {
    return run(root, write);
  } finally {
    fs.rmSync(root, { recursive: true, force: true });
  }
}

describe('rule metadata', () => {
  test('Rust registry and frozen baseline metadata cover the scoped schemas without registering unsupported rules', async () => {
    const root = fileURLToPath(new URL('../../../', import.meta.url));
    const provenance = JSON.parse(
      fs.readFileSync(
        path.join(root, 'packages/rslint/rust-rule-schemas.provenance.json'),
        'utf8',
      ),
    );
    const bytes = fs.readFileSync(path.join(root, provenance.workload.path));
    expect(createHash('sha256').update(bytes).digest('hex')).toBe(
      provenance.workload.sha256,
    );
    const manifest = JSON.parse(bytes.toString('utf8'));
    expect(manifest.project.commit).toBe(
      'e7d65898ce92d5f5d78afb0d93b4670b5e2a469d',
    );
    expect(manifest.counts).toMatchObject({
      unique_files: 2394,
      unique_rule_count: 129,
      plans: 33,
      rule_passes: 221014,
    });
    expect(provenance.baseline.commit).toBe(manifest.baseline.commit);
    for (const field of [
      'traceSHA256',
      'summarySHA256',
      'commandRecordSHA256',
      'sourceRecordSHA256',
    ]) {
      expect(provenance.baselineWorkloadProof[field]).toBe(
        manifest.baseline[field],
      );
      expect(provenance.baselineWorkloadProof[field]).toMatch(/^[a-f0-9]{64}$/);
    }
    // These tracked byte hashes bind the original oracle's metadata. Reading
    // or rebuilding the external Go checkout is not part of candidate tests.
    const catalogBytes = fs.readFileSync(
      path.join(root, provenance.baselineMetadataCatalog.path),
    );
    expect(createHash('sha256').update(catalogBytes).digest('hex')).toBe(
      provenance.baselineMetadataCatalog.sha256,
    );
    const catalog = JSON.parse(catalogBytes.toString('utf8'));
    expect(catalog.baseline.commit).toBe(provenance.baseline.commit);
    expect(catalog.baseline.tree).toBe(provenance.baseline.tree);
    const original = new Map(
      catalog.rules.map((rule: { name: string; options_schema: unknown }) => [
        rule.name,
        rule.options_schema,
      ]),
    );
    const schemas = await collectRuleSchemas();
    for (const rule of schemas)
      expect(rule.schema).toEqual(original.get(rule.name));

    // Fail closed on unknown registry expressions; do not treat metadata or
    // schema presence alone as proof of behavioral compatibility.
    const directory = path.join(root, 'crates/rslint-core/src/rules');
    const registry = fs.readFileSync(path.join(directory, 'mod.rs'), 'utf8');
    const body =
      /pub\(crate\) fn supported\(name: &str\) -> bool \{([\s\S]*?)\n\}/.exec(
        registry,
      )?.[1];
    expect(body).toBeDefined();
    const registered: string[] = [];
    for (const part of body!.replace(/\s/g, '').split('||')) {
      const module = /^([a-z_]+)::NAMES\.contains\(&name\)$/.exec(part)?.[1];
      if (module) {
        const source = fs.readFileSync(
          path.join(directory, `${module}.rs`),
          'utf8',
        );
        const names = /const NAMES: &\[&str\] = &\[([\s\S]*?)\];/.exec(
          source,
        )?.[1];
        expect(names, module).toBeDefined();
        registered.push(...JSON.parse(`[${names!.replace(/,\s*$/, '')}]`));
      } else {
        const constant = /^name==([A-Z_]+)$/.exec(part)?.[1];
        expect(constant, part).toBeDefined();
        const name = new RegExp(`const ${constant}: &str = "([^"]+)";`).exec(
          registry,
        )?.[1];
        expect(name, part).toBeDefined();
        registered.push(name!);
      }
    }
    expect(new Set(registered).size).toBe(registered.length);
    const typedNames = new Set(schemas.map((rule) => rule.name));
    expect(
      [...typedNames].filter((name) => !registered.includes(name)),
    ).toEqual([]);
    expect([...typedNames].sort()).toEqual(
      [
        ...new Set(manifest.rules.map((rule: { name: string }) => rule.name)),
      ].sort(),
    );
    const extras = registered.filter((name) => !typedNames.has(name)).sort();
    expect(extras).toEqual(provenance.retainedUntypedRules);
    expect(extras).toHaveLength(10);
    // Extra implemented rules are retained, but the 129-rule typed schema
    // export makes no additional compatibility claim about those ten rules.
    const configSource = fs.readFileSync(
      path.join(root, 'packages/rslint/src/config/define-config.ts'),
      'utf8',
    );
    expect(configSource).toContain(
      '[key: string]: RuleEntry<any[]> | undefined;',
    );
  });

  test('committed Rust option schemas cover the exact frozen workload and fail on missing or stale inputs', async () => {
    const rules = await collectRuleSchemas();
    const manifest = JSON.parse(
      fs.readFileSync(
        new URL(
          '../../../docs/tsrs-migration/rsbuild-workload.json',
          import.meta.url,
        ),
        'utf8',
      ),
    );
    expect(rules.map((rule) => rule.name).sort()).toEqual(
      [
        ...new Set(manifest.rules.map((rule: { name: string }) => rule.name)),
      ].sort(),
    );
    expect(rules).toHaveLength(129);
    const generated = await compileRuleOptionTypes(rules);
    expect(generated.recordProperties).toHaveLength(rules.length);

    const directory = fs.mkdtempSync(
      path.join(os.tmpdir(), 'rslint-rust-schemas-'),
    );
    const schemasPath = path.join(directory, 'schemas.json');
    const provenancePath = path.join(directory, 'provenance.json');
    try {
      await expect(
        collectRuleSchemas(schemasPath, provenancePath),
      ).rejects.toThrow();
      fs.writeFileSync(schemasPath, JSON.stringify(rules));
      fs.writeFileSync(
        provenancePath,
        JSON.stringify({
          schemas: {
            count: 129,
            sha256: 'wrong',
            bytes: fs.statSync(schemasPath).size,
          },
        }),
      );
      await expect(
        collectRuleSchemas(schemasPath, provenancePath),
      ).rejects.toThrow('do not match their provenance');
      const duplicates = Buffer.from(JSON.stringify([rules[0], rules[0]]));
      fs.writeFileSync(schemasPath, duplicates);
      fs.writeFileSync(
        provenancePath,
        JSON.stringify({
          schemas: {
            count: 2,
            sha256: createHash('sha256').update(duplicates).digest('hex'),
            bytes: duplicates.length,
          },
        }),
      );
      await expect(
        collectRuleSchemas(schemasPath, provenancePath),
      ).rejects.toThrow('do not match their provenance');
    } finally {
      fs.rmSync(directory, { recursive: true, force: true });
    }
  });

  test('generated option types preserve punctuation-distinct IDs without declaration collisions', async () => {
    const entries = [
      {
        name: 'node/prefer-global/url',
        schema: {
          type: 'array',
          items: [{ enum: ['nested'] }],
          additionalItems: false,
        },
      },
      {
        name: 'node/prefer-global-url',
        schema: {
          type: 'array',
          items: [{ type: 'boolean' }],
          additionalItems: false,
        },
      },
      {
        name: 'node/prefer_global_url',
        schema: {
          type: 'array',
          items: [{ type: 'number' }],
          additionalItems: false,
        },
      },
    ];
    const generated = await compileRuleOptionTypes(entries);
    expect(await compileRuleOptionTypes([...entries].reverse())).toEqual(
      generated,
    );
    withFixture((root, write) => {
      const source = `type RuleEntry<T extends unknown[]> = ['error', ...T];
interface RulesRecord { ${generated.recordProperties.join('\n')} }
${generated.typeDeclarations.join('\n')}
const valid: RulesRecord = {
  'node/prefer-global/url': ['error', 'nested'],
  'node/prefer-global-url': ['error', true],
  'node/prefer_global_url': ['error', 1],
};
const invalid: RulesRecord = {
  // @ts-expect-error -- the nested rule must retain its own options
  'node/prefer-global/url': ['error', true],
};
`;
      write('options.ts', source);
      const program = ts.createProgram([path.join(root, 'options.ts')], {
        strict: true,
        noEmit: true,
        types: [],
        target: ts.ScriptTarget.ES2022,
      });
      expect(
        ts
          .getPreEmitDiagnostics(program)
          .map((diagnostic) =>
            ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n'),
          ),
      ).toEqual([]);
    });
  });
});
