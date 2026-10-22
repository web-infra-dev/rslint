// Compare discovery and effective flat config with the frozen original workload.
// This does not assert compiler project selection or rule implementations.
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';
import { registerHooks } from 'node:module';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { execFileSync } from 'node:child_process';

const [projectArg, manifestArg, outputArg, baselineArg] = process.argv.slice(2);
assert(
  projectArg && manifestArg && outputArg,
  'usage: node scripts/tsrs/rust-admission-oracle.mjs PROJECT MANIFEST OUTPUT',
);
const project = fs.realpathSync(projectArg);
const manifestPath = path.resolve(manifestArg);
const output = path.resolve(outputArg);
const sourceRoot = fileURLToPath(
  new URL('../../packages/rslint/src/', import.meta.url),
);
const hooks = registerHooks({
  resolve(specifier, context, nextResolve) {
    if (
      specifier.endsWith('.js') &&
      specifier.startsWith('.') &&
      context.parentURL?.startsWith(pathToFileURL(sourceRoot).href)
    ) {
      const url = new URL(specifier.slice(0, -3) + '.ts', context.parentURL);
      if (fs.existsSync(url)) return nextResolve(url.href, context);
    }
    return nextResolve(specifier, context);
  },
});
const hash = (value) => crypto.createHash('sha256').update(value).digest('hex');
const manifestBytes = fs.readFileSync(manifestPath);
const manifest = JSON.parse(manifestBytes);
const report = {
  schemaVersion: 1,
  status: 'running',
  node: process.version,
  manifestSha256: hash(manifestBytes),
  admissionSourceSha256: hash(
    fs.readFileSync(path.join(sourceRoot, 'rust/admission.ts')),
  ),
  projectCommit: execFileSync('git', ['-C', project, 'rev-parse', 'HEAD'], {
    encoding: 'utf8',
  }).trim(),
  scope:
    'File admission, source hashes, exact enabled non-type rules/options/severity, and configured typed rules. Checker availability is taken from the frozen oracle only for comparison and is NOT independently validated here.',
  files: 0,
  matchedConfigurations: 0,
  typedRuleBindingValidated: false,
};
const baseline = baselineArg ? fs.realpathSync(baselineArg) : undefined;
const previousCwd = process.cwd();
try {
  assert.equal(
    report.projectCommit,
    manifest.project.commit,
    'project pin differs',
  );
  for (const [name, expectedHash] of manifest.project.inputs)
    assert.equal(
      hash(fs.readFileSync(path.join(project, name))),
      expectedHash,
      `project input changed: ${name}`,
    );
  const { admitRustFiles } = await import(
    pathToFileURL(path.join(sourceRoot, 'rust/admission.ts')).href
  );
  process.chdir(project);
  const configPath = path.resolve('node_modules/rstack/dist/rslintConfig.js');
  const { resolveRustConfig } = await import(
    pathToFileURL(path.join(sourceRoot, 'rust/config.ts')).href
  );
  const resolved = await resolveRustConfig({
    cwd: project,
    overrideConfigFile: configPath,
  });
  const config = resolved.config;
  report.rustConfigAdapterSha256 = hash(
    fs.readFileSync(path.join(sourceRoot, 'rust/config.ts')),
  );
  report.configAdapterSha256 = hash(fs.readFileSync(configPath));
  report.resolvedConfigSha256 = hash(JSON.stringify(config));
  const actual = admitRustFiles({
    ...resolved,
  });
  const expected = new Map(manifest.files.map((row) => [row[0], row]));
  const actualPaths = actual
    .map((file) => path.relative(project, file.path).split(path.sep).join('/'))
    .sort();
  assert.deepEqual(
    actualPaths,
    [...expected.keys()].sort(),
    'admitted files differ',
  );
  const byName = new Map(manifest.rules.map((rule) => [rule.name, rule]));
  const shapes = new Set();
  for (const file of actual) {
    const relative = path
      .relative(project, file.path)
      .split(path.sep)
      .join('/');
    const row = expected.get(relative);
    const source = fs.readFileSync(file.path);
    assert.equal(hash(source), row[1], `source hash changed: ${relative}`);
    assert.equal(
      source.byteLength,
      row[2],
      `source bytes changed: ${relative}`,
    );
    const configuration = manifest.configurations[row[3]];
    const environment = manifest.environments[configuration.environment];
    assert.deepEqual(
      file.config.settings,
      environment.Settings ?? {},
      `settings differ: ${relative}`,
    );
    assert.deepEqual(
      file.config.languageOptions.globals ?? {},
      environment.Globals ?? {},
      `globals differ: ${relative}`,
    );
    const nativeRules = configuration.rules.all
      .map((index) => {
        const rule = manifest.rules[index];
        return {
          name: rule.name,
          severity: rule.severityName === 'error' ? 2 : 1,
          options: rule.options ?? [],
        };
      })
      .sort((a, b) => a.name.localeCompare(b.name));
    const admittedRules = file.config.rules.filter((rule) => {
      const meta = byName.get(rule.name);
      assert(meta, `unknown configured rule: ${rule.name}`);
      return configuration.hasTypeChecker || !meta.requiresTypeInformation;
    });
    assert.deepEqual(
      admittedRules,
      nativeRules,
      `effective rule config differs: ${relative}`,
    );
    shapes.add(JSON.stringify(nativeRules));
  }
  report.files = actual.length;
  report.matchedConfigurations = shapes.size;
  if (baseline)
    report.fixtureCases = await fixtureOracle(admitRustFiles, baseline);
  report.status = 'passed';
} catch (error) {
  report.status = 'failed';
  report.error = String(error);
  report.stack = error?.stack;
  throw error;
} finally {
  process.chdir(previousCwd);
  hooks.deregister();
  fs.mkdirSync(path.dirname(output), { recursive: true });
  fs.writeFileSync(output, JSON.stringify(report, null, 2) + '\n');
  console.log(JSON.stringify(report));
}

async function fixtureOracle(admit, baseline) {
  const pin = 'e0a29f751d4def86d27931759098f27748fa3003';
  assert.equal(
    execFileSync('git', ['-C', baseline, 'rev-parse', 'HEAD'], {
      encoding: 'utf8',
    }).trim(),
    pin,
  );
  assert.equal(
    execFileSync('git', ['-C', baseline, 'diff', '--name-only', 'HEAD'], {
      encoding: 'utf8',
    }),
    '',
  );
  const { Rslint } = await import(
    pathToFileURL(path.join(baseline, 'packages/rslint/dist/index.js')).href
  );
  const source = 'debugger; true ? 1 : 0;';
  const cases = [
    {
      name: 'all-extensions',
      files: Object.fromEntries(
        [
          'a.ts',
          'a.tsx',
          'a.js',
          'a.jsx',
          'a.mjs',
          'a.cjs',
          'a.mts',
          'a.cts',
          '.hidden.ts',
          'upper.TS',
          'node_modules/a.ts',
        ].map((name) => [name, source]),
      ),
      config: [{ rules: { 'no-debugger': 'warn', 'no-ternary': 'error' } }],
    },
    {
      name: 'ignores-and-overrides',
      files: {
        'blocked/a.ts': source,
        'covered/a.ts': source,
        'covered/b.ts': source,
        'local.ts': source,
      },
      config: [
        {
          ignores: [
            'blocked/**',
            '!blocked/a.ts',
            'covered/**/*',
            '!covered/a.ts',
          ],
        },
        { rules: { 'no-debugger': 'error' } },
        {
          ignores: ['local.ts'],
          rules: { 'no-debugger': 'off', 'no-ternary': 'warn' },
        },
      ],
    },
    {
      name: 'gitignore-hierarchy',
      files: {
        '.gitignore': 'blocked/\n*.generated.ts\n',
        'blocked/.gitignore': '!a.ts\n',
        'blocked/a.ts': source,
        'src/.gitignore': '!keep.generated.ts\n',
        'src/keep.generated.ts': source,
        'src/drop.generated.ts': source,
        'src/plain.ts': source,
      },
      config: [{ rules: { 'no-debugger': 'error' } }],
    },
    {
      name: 'base-path-and-selectors',
      files: {
        'src/a.ts': source,
        'src/a.test.ts': source,
        'src/b.tsx': source,
        'other/a.ts': source,
      },
      config: [
        [{ rules: { 'no-debugger': 'warn' } }],
        {
          basePath: 'src',
          files: [['**/*.{ts,tsx}', '!**/*.test.ts']],
          rules: { 'no-debugger': 'error', 'no-ternary': 'warn' },
        },
      ],
    },
    {
      name: 'git-reopened-directory',
      files: {
        '.gitignore': 'blocked/\n',
        'blocked/a.ts': source,
        'blocked/b.ts': source,
      },
      config: [
        { ignores: ['!blocked', '!blocked/a.ts'] },
        { rules: { 'no-debugger': 'error' } },
      ],
    },
    {
      name: 'git-quoted-patterns',
      files: {
        '.gitignore':
          'literal-{a,b}.ts\na(b).ts\nquoted\\*.ts\n leading.ts\ntrailing\\ .ts  \n[!a].ts\n',
        'literal-a.ts': source,
        'literal-{a,b}.ts': source,
        'ab.ts': source,
        'a(b).ts': source,
        'quoted*.ts': source,
        'quotedX.ts': source,
        ' leading.ts': source,
        'leading.ts': source,
        'trailing .ts': source,
        'a.ts': source,
        'b.ts': source,
      },
      config: [{ rules: { 'no-debugger': 'error' } }],
    },
  ];
  const results = [];
  for (const item of cases) {
    const cwd = fs.realpathSync(
      fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-admission-oracle-')),
    );
    let linter;
    try {
      for (const [name, text] of Object.entries(item.files)) {
        const filename = path.join(cwd, name);
        fs.mkdirSync(path.dirname(filename), { recursive: true });
        fs.writeFileSync(filename, text);
      }
      const actual = admit({ cwd, config: item.config })
        .map((file) => ({
          path: path.relative(cwd, file.path),
          messages: file.config.rules
            .map((rule) => ({ ruleId: rule.name, severity: rule.severity }))
            .sort((a, b) => a.ruleId.localeCompare(b.ruleId)),
        }))
        .sort((a, b) => a.path.localeCompare(b.path));
      linter = new Rslint({
        cwd,
        overrideConfigFile: true,
        overrideConfig: item.config,
      });
      const expected = (await linter.lintFiles(['.']))
        .map((file) => ({
          path: path.relative(cwd, file.filePath),
          messages: file.messages
            .map((message) => ({
              ruleId: message.ruleId,
              severity: message.severity,
            }))
            .sort((a, b) => a.ruleId.localeCompare(b.ruleId)),
        }))
        .sort((a, b) => a.path.localeCompare(b.path));
      assert.deepEqual(actual, expected, item.name);
      results.push({ name: item.name, files: actual.length, status: 'passed' });
    } finally {
      await linter?.close();
      fs.rmSync(cwd, { recursive: true, force: true });
    }
  }
  return results;
}
