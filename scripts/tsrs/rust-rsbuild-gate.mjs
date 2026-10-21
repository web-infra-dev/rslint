#!/usr/bin/env node
/**
 * Freeze the real rsbuild workload and reject an incomparable Rust candidate.
 *
 * capture --project DIR --baseline DIR --trace FILE --summary FILE --source FILE
 *         --record FILE --stdout FILE --stderr FILE [--manifest FILE]
 * check   --project DIR --addon FILE [--manifest FILE] [--output FILE]
 *
 * This is a correctness/capability preflight, never a performance measurement.
 * The negative probe retains every enabled rule name. It does not replace the
 * actual per-file configuration or establish diagnostic/fix equivalence.
 */
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import { createRequire } from 'node:module';
import { createInterface } from 'node:readline';
import { execFileSync, spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const SCRIPT = fileURLToPath(import.meta.url);
const ROOT = path.resolve(path.dirname(SCRIPT), '../..');
const PROJECT_PIN = 'e7d65898ce92d5f5d78afb0d93b4670b5e2a469d';
const BASELINE_PIN = 'e0a29f751d4def86d27931759098f27748fa3003';
const COMPILER_PIN = '280b426b89cd4ecd6153cf48228826bd60a99c61';
const DEFAULT_MANIFEST = path.join(
  ROOT,
  'docs/tsrs-migration/rsbuild-workload.json',
);
const EXPECTED = {
  plans: 33,
  file_passes: 2394,
  unique_files: 2394,
  rule_passes: 221014,
  type_aware_rule_passes: 31461,
  unique_rule_count: 129,
  typed_files: 1232,
};

function options(args) {
  const result = {};
  for (let index = 0; index < args.length; index += 2) {
    const key = args[index];
    assert.match(key, /^--[a-z-]+$/);
    assert.ok(
      args[index + 1] && !args[index + 1].startsWith('--'),
      `${key} needs a value`,
    );
    assert.ok(!(key.slice(2) in result), `duplicate ${key}`);
    result[key.slice(2)] = args[index + 1];
  }
  return result;
}
function required(args, name) {
  assert.ok(args[name], `--${name} is required`);
  return path.resolve(args[name]);
}
function sha(data) {
  return createHash('sha256').update(data).digest('hex');
}
function fileHash(file) {
  return sha(fs.readFileSync(file));
}
function json(file) {
  return JSON.parse(fs.readFileSync(file, 'utf8'));
}
function writeJSON(file, value) {
  fs.mkdirSync(path.dirname(file), { recursive: true });
  fs.writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`);
}
function canonical(value) {
  if (Array.isArray(value)) return `[${value.map(canonical).join(',')}]`;
  if (value !== null && typeof value === 'object') {
    return `{${Object.keys(value)
      .sort()
      .map((key) => `${JSON.stringify(key)}:${canonical(value[key])}`)
      .join(',')}}`;
  }
  return JSON.stringify(value);
}
function relative(project, filename) {
  const result = path.relative(project, filename).split(path.sep).join('/');
  assert.ok(
    result && !result.startsWith('../') && !path.isAbsolute(result),
    'input escapes project root',
  );
  return result;
}
function inside(project, name) {
  assert.ok(typeof name === 'string' && name && !name.includes('\\'));
  assert.ok(
    !path.posix.isAbsolute(name) && !name.split('/').includes('..'),
    'unsafe manifest path',
  );
  return path.join(project, ...name.split('/'));
}
function git(project, ...args) {
  return execFileSync('git', ['-C', project, ...args], {
    encoding: 'utf8',
    maxBuffer: 16 << 20,
  });
}
function verifyProject(project, pin = PROJECT_PIN) {
  assert.equal(
    git(project, 'rev-parse', 'HEAD').trim(),
    pin,
    'project commit differs',
  );
  assert.equal(
    git(project, 'status', '--porcelain=v1', '--untracked-files=no').trim(),
    '',
    'tracked project inputs are dirty',
  );
}
function intern(table, map, value) {
  const key = canonical(value);
  if (!map.has(key)) {
    map.set(key, table.length);
    table.push(value);
  }
  return map.get(key);
}
function portable(value) {
  const encoded = JSON.stringify(value);
  assert.ok(
    !/\/Users\/|\/home\/|[A-Za-z]:\\\\Users\\\\|Bearer\s|github_pat_|ghp_/.test(
      encoded,
    ),
    'metadata contains a private path or credential-like value',
  );
}

function runtimeAnchors(project) {
  const require = createRequire(path.join(project, 'package.json'));
  const result = [];
  for (const name of ['rstack', 'typescript']) {
    const metadata = require.resolve(`${name}/package.json`);
    const root = path.dirname(metadata);
    const version = json(metadata).version;
    const files =
      name === 'rstack'
        ? [
            'package.json',
            'bin/rs.js',
            'dist/rslintConfig.js',
            'dist/a.js',
            'dist/f.js',
          ]
        : ['package.json'];
    result.push({
      package: name,
      version,
      files: files.map((file) => [file, fileHash(path.join(root, file))]),
    });
  }
  return result;
}

function generatedSnapshot(project) {
  const roots = [];
  for (const parent of ['packages', 'scripts']) {
    const directory = path.join(project, parent);
    if (!fs.existsSync(directory)) continue;
    for (const entry of fs.readdirSync(directory).sort()) {
      for (const suffix of ['dist', 'compiled']) {
        const name = `${parent}/${entry}/${suffix}`;
        if (fs.existsSync(inside(project, name))) roots.push(name);
      }
    }
  }
  return roots.map((name) => {
    const root = inside(project, name);
    const rows = [];
    function walk(directory) {
      for (const entry of fs
        .readdirSync(directory, { withFileTypes: true })
        .sort((a, b) => a.name.localeCompare(b.name))) {
        const full = path.join(directory, entry.name);
        assert.ok(
          !entry.isSymbolicLink(),
          'generated source symlink needs explicit provenance',
        );
        if (entry.isDirectory()) walk(full);
        else if (entry.isFile())
          rows.push([
            relative(root, full),
            fs.statSync(full).size,
            fileHash(full),
          ]);
        else throw new Error('unsupported generated input kind');
      }
    }
    walk(root);
    return {
      path: name,
      files: rows.length,
      bytes: rows.reduce((sum, row) => sum + row[1], 0),
      sha256: sha(canonical(rows)),
    };
  });
}

function summarize(manifest) {
  for (const configuration of manifest.configurations) {
    assert.equal(typeof configuration.hasTypeChecker, 'boolean');
    assert.equal(typeof configuration.hasPreparedRules, 'boolean');
    assert.ok(manifest.environments[configuration.environment] !== undefined);
    for (const kind of ['all', 'native', 'plugins']) {
      assert.ok(Array.isArray(configuration.rules[kind]));
      for (const index of configuration.rules[kind]) {
        assert.ok(
          Number.isInteger(index) &&
            index >= 0 &&
            index < manifest.rules.length,
        );
      }
    }
  }
  const names = new Set();
  const files = new Set();
  let rules = 0;
  let typedRules = 0;
  let typedFiles = 0;
  const planFiles = manifest.plans.map(() => []);
  for (const [name, hash, bytes, group, plan, index] of manifest.files) {
    inside('/manifest-root', name);
    assert.match(hash, /^[a-f\d]{64}$/);
    assert.ok(Number.isInteger(bytes) && bytes >= 0);
    const config = manifest.configurations[group];
    assert.ok(config && manifest.plans[plan]);
    assert.equal(
      index,
      planFiles[plan].length,
      'noncontiguous admission index',
    );
    planFiles[plan].push(name);
    assert.ok(!files.has(name), 'duplicate input admission');
    files.add(name);
    typedFiles += Number(config.hasTypeChecker);
    for (const ruleIndex of config.rules.all) {
      const rule = manifest.rules[ruleIndex];
      assert.ok(rule && manifest.environments[rule.environment] !== undefined);
      names.add(rule.name);
      rules++;
      typedRules += Number(rule.requiresTypeInformation);
      assert.ok(
        !rule.requiresTypeInformation || config.hasTypeChecker,
        'typed rule without checker grant',
      );
    }
  }
  for (let index = 0; index < manifest.plans.length; index++) {
    assert.equal(
      planFiles[index].length,
      manifest.plans[index].files,
      'incomplete plan',
    );
  }
  return {
    plans: manifest.plans.length,
    file_passes: manifest.files.length,
    unique_files: files.size,
    rule_passes: rules,
    type_aware_rule_passes: typedRules,
    unique_rule_count: names.size,
    typed_files: typedFiles,
  };
}

async function capture(args) {
  const project = required(args, 'project');
  const baseline = required(args, 'baseline');
  const trace = required(args, 'trace');
  const summaryFile = required(args, 'summary');
  const sourceFile = required(args, 'source');
  const recordFile = required(args, 'record');
  const stdout = required(args, 'stdout');
  const stderr = required(args, 'stderr');
  const output = path.resolve(args.manifest ?? DEFAULT_MANIFEST);
  verifyProject(project);
  verifyProject(baseline, BASELINE_PIN);
  const summary = json(summaryFile);
  const source = json(sourceFile);
  const record = json(recordFile);
  assert.equal(summary.correctness_only, true);
  assert.equal(summary.exit_code, 0);
  assert.equal(record.exit_code, 0);
  assert.deepEqual(record.argv.slice(-3), ['lint', '--format', 'jsonline']);
  assert.equal(record.env.CI, 'true');
  assert.ok(path.isAbsolute(record.cwd), 'recorded source root is missing');
  assert.equal(fs.statSync(stdout).size, 0);
  assert.equal(fs.statSync(stderr).size, 0);
  assert.equal(source.sha, PROJECT_PIN);
  const tracked = new Set(
    git(project, 'ls-files', '-z').split('\0').filter(Boolean),
  );
  const environments = [];
  const environmentMap = new Map();
  const rules = [];
  const ruleMap = new Map();
  const configurations = [];
  const configurationMap = new Map();
  const plans = [];
  const planMap = new Map();
  const files = [];
  let active;
  const stream = fs.createReadStream(trace);
  const traceHash = createHash('sha256');
  stream.on('data', (chunk) => traceHash.update(chunk));
  const lines = createInterface({ input: stream, crlfDelay: Infinity });
  for await (const line of lines) {
    const row = JSON.parse(line);
    assert.equal(row.schemaVersion, 1);
    const key = `${row.pid}:${row.planSequence}`;
    if (row.event === 'planStart') {
      assert.equal(active, undefined, 'interleaved plans');
      assert.ok(!planMap.has(key), 'duplicate plan');
      const id = plans.length;
      plans.push({ files: row.fileCount });
      planMap.set(key, id);
      active = { key, id, seen: 0 };
    } else if (row.event === 'file') {
      assert.equal(active?.key, key, 'file outside a plan');
      assert.equal(row.fileIndex, active.seen++);
      const name = relative(record.cwd, row.filePath);
      assert.ok(
        tracked.has(name),
        'untracked admission requires a source mapping',
      );
      const full = inside(project, name);
      assert.ok(
        !fs.lstatSync(full).isSymbolicLink(),
        'admitted symlink requires a source mapping',
      );
      assert.equal(fileHash(full), row.sourceSHA256, `input changed: ${name}`);
      assert.equal(fs.statSync(full).size, row.sourceBytes);
      const groups = {};
      for (const kind of ['all', 'native', 'plugins']) {
        groups[kind] = row.rules[kind].map((rule) =>
          intern(rules, ruleMap, {
            ...rule,
            environment: intern(environments, environmentMap, rule.environment),
          }),
        );
      }
      const configuration = intern(configurations, configurationMap, {
        hasTypeChecker: row.hasTypeChecker,
        hasPreparedRules: row.hasPreparedRules,
        environment: intern(environments, environmentMap, row.environment),
        pluginLanguageOptions: row.pluginLanguageOptions,
        rules: groups,
      });
      files.push([
        name,
        row.sourceSHA256,
        row.sourceBytes,
        configuration,
        active.id,
        row.fileIndex,
      ]);
    } else if (row.event === 'planEnd') {
      assert.equal(active?.key, key);
      assert.equal(active.seen, row.fileCount);
      assert.equal(row.fileCount, plans[active.id].files);
      active = undefined;
    } else throw new Error('unknown trace event');
  }
  assert.equal(active, undefined, 'unfinished plan');
  assert.equal(
    traceHash.digest('hex'),
    summary.trace_sha256,
    'trace hash differs',
  );
  const inputs = Object.entries(source.files).map(([name, hash]) => {
    assert.equal(
      fileHash(inside(project, name)),
      hash,
      `config input changed: ${name}`,
    );
    return [name, hash];
  });
  const baselineCompiler = git(
    baseline,
    'ls-tree',
    BASELINE_PIN,
    'typescript-go',
  )
    .trim()
    .split(/\s+/)[2];
  const manifest = {
    schemaVersion: 1,
    purpose:
      'Frozen correctness-only rsbuild admission workload; not a performance result',
    project: {
      repository: 'https://github.com/web-infra-dev/rsbuild.git',
      commit: PROJECT_PIN,
      node: source.node,
      packageManager: json(path.join(project, 'package.json')).packageManager,
      lintScript: 'rs lint',
      inputs,
    },
    baseline: {
      repository: 'https://github.com/web-infra-dev/rslint.git',
      commit: BASELINE_PIN,
      compilerCommit: baselineCompiler,
      exitCode: 0,
      stdoutBytes: 0,
      stderrBytes: 0,
      observedDiagnostics: 0,
      fixParity: 'not_verified',
      traceSHA256: summary.trace_sha256,
      summarySHA256: fileHash(summaryFile),
      commandRecordSHA256: fileHash(recordFile),
      sourceRecordSHA256: fileHash(sourceFile),
      command: [
        'node',
        '--import',
        '<baseline-selector>',
        '<project>/node_modules/rstack/bin/rs.js',
        'lint',
        '--format',
        'jsonline',
      ],
    },
    counts: EXPECTED,
    fileColumns: [
      'projectRelativePath',
      'sha256',
      'sourceBytes',
      'configurationIndex',
      'planIndex',
      'fileIndex',
    ],
    tableContract:
      'Every file admission is retained, including silent files. Rule indices preserve all/native/plugins order; environment indices preserve exact metadata. Baseline severity 0 means error, 1 means warning, 2 means off.',
    environments,
    rules,
    configurations,
    plans,
    files,
    preparation: {
      capturedFor:
        'current correctness preflight; historical full dependency identity is not established',
      configurationRuntimeAnchors: runtimeAnchors(project),
      generatedOutputs: generatedSnapshot(project),
    },
    limitations: [
      'Admission trace is recorded before execution. Empty stdout/stderr and successful exit are separate baseline run evidence.',
      'No final Rust diagnostic/fix parity or repeated A/B has passed.',
      'Runtime anchors and generated output hashes do not establish the complete transitive dependency input closure.',
      'The trace does not record each plan’s resolved project or tsconfig identity; that equality check remains required.',
      'No historical Go+Rust or two-rule PoC timing is a pure-Rust result.',
    ],
  };
  assert.deepEqual(summarize(manifest), EXPECTED);
  for (const [key, value] of Object.entries(EXPECTED))
    assert.equal(summary[key], value);
  portable(manifest);
  const text = `${JSON.stringify(manifest, null, 2)}\n`;
  assert.ok(
    Buffer.byteLength(text) < 1_000_000,
    'workload manifest exceeds compact size limit',
  );
  fs.mkdirSync(path.dirname(output), { recursive: true });
  fs.writeFileSync(output, text);
  console.log(
    JSON.stringify({
      status: 'captured',
      counts: EXPECTED,
      configurationGroups: configurations.length,
      manifestBytes: Buffer.byteLength(text),
      traceSHA256: summary.trace_sha256,
    }),
  );
}

function sourceCapabilities(root = ROOT) {
  const directory = path.join(root, 'crates/rslint-core/src/rules');
  const source = fs.readFileSync(path.join(directory, 'mod.rs'), 'utf8');
  const expression = source.match(
    /pub\(crate\) fn supported\(name: &str\) -> bool \{([\s\S]*?)\n\}/,
  )?.[1];
  assert.ok(expression, 'unsupported Rust registry shape');
  const names = [];
  const sourceHashes = {
    'crates/rslint-core/src/rules/mod.rs': fileHash(
      path.join(directory, 'mod.rs'),
    ),
  };
  for (const part of expression.replace(/\s/g, '').split('||')) {
    const constant = part.match(/^name==([A-Z_]+)$/);
    const module = part.match(/^([a-z_]+)::NAMES\.contains\(&name\)$/);
    if (constant) {
      const name = source.match(
        new RegExp(`const ${constant[1]}: &str = "([^"]+)";`),
      )?.[1];
      assert.ok(name, 'unknown Rust rule constant');
      names.push(name);
    } else if (module) {
      const file = path.join(directory, `${module[1]}.rs`);
      const text = fs.readFileSync(file, 'utf8');
      const declaration = text.match(
        /const NAMES: &\[&str\] = &\[([\s\S]*?)\];/,
      )?.[1];
      assert.ok(declaration, 'unknown Rust rule list');
      const values = JSON.parse(`[${declaration.replace(/,\s*$/, '')}]`);
      assert.ok(values.every((name) => typeof name === 'string'));
      names.push(...values);
      sourceHashes[`crates/rslint-core/src/rules/${module[1]}.rs`] =
        fileHash(file);
    } else throw new Error('unsupported Rust registry expression');
  }
  assert.equal(
    new Set(names).size,
    names.length,
    'duplicate Rust registrations',
  );
  return {
    names: names.sort(),
    sourceSHA256: sourceHashes,
    caveat:
      'Source declarations only; this is not proof of complete addon capabilities or rule compatibility.',
  };
}

function probe(args) {
  const addon = required(args, 'addon');
  const require = createRequire(import.meta.url);
  const binding = require(addon);
  assert.equal(typeof binding.lintTsrsSync, 'function');
  const request = fs.readFileSync(0, 'utf8');
  try {
    const result = binding.lintTsrsSync(request);
    console.log(JSON.stringify({ returned: true, resultSHA256: sha(result) }));
  } catch (error) {
    console.log(
      JSON.stringify({
        returned: false,
        error: String(error.message ?? error),
      }),
    );
  }
}

function check(args) {
  const project = required(args, 'project');
  const addon = required(args, 'addon');
  const manifestFile = path.resolve(args.manifest ?? DEFAULT_MANIFEST);
  const manifest = json(manifestFile);
  assert.equal(manifest.schemaVersion, 1);
  assert.equal(manifest.project.commit, PROJECT_PIN);
  assert.equal(manifest.baseline.commit, BASELINE_PIN);
  assert.deepEqual(summarize(manifest), EXPECTED);
  assert.deepEqual(manifest.counts, EXPECTED);
  verifyProject(project);
  assert.equal(
    process.version,
    manifest.project.node,
    'Node version differs from captured baseline',
  );
  assert.equal(
    json(path.join(project, 'package.json')).packageManager,
    manifest.project.packageManager,
  );
  for (const [name, hash] of manifest.project.inputs)
    assert.equal(
      fileHash(inside(project, name)),
      hash,
      `config input changed: ${name}`,
    );
  for (const [name, hash, bytes] of manifest.files) {
    const full = inside(project, name);
    assert.ok(!fs.lstatSync(full).isSymbolicLink());
    assert.equal(fileHash(full), hash, `admitted input changed: ${name}`);
    assert.equal(fs.statSync(full).size, bytes);
  }
  assert.deepEqual(
    runtimeAnchors(project),
    manifest.preparation.configurationRuntimeAnchors,
    'configuration runtime changed',
  );
  assert.deepEqual(
    generatedSnapshot(project),
    manifest.preparation.generatedOutputs,
    'generated preparation outputs changed',
  );
  const capabilities = sourceCapabilities();
  const union = new Map();
  for (const config of manifest.configurations) {
    for (const index of config.rules.all) {
      const rule = manifest.rules[index];
      if (union.has(rule.name))
        assert.deepEqual(
          union.get(rule.name),
          rule,
          'per-file rule variants need a richer negative probe',
        );
      union.set(rule.name, rule);
    }
  }
  const enabled = [...union.keys()].sort();
  assert.equal(enabled.length, 129);
  const unsupported = enabled.filter(
    (name) => !capabilities.names.includes(name),
  );
  const implementedPairs = manifest.files.reduce(
    (sum, row) =>
      sum +
      manifest.configurations[row[3]].rules.all.filter((index) =>
        capabilities.names.includes(manifest.rules[index].name),
      ).length,
    0,
  );
  const blockers = [
    {
      kind: 'per_file_configuration',
      message:
        'The current direct API accepts one uniform rules array and one optional project; these five per-file configurations and projectService grants have no equivalent complete adapter.',
    },
    {
      kind: 'diagnostic_and_fix_parity',
      message:
        'Complete normalized diagnostics and fixes have not been compared against the frozen baseline.',
    },
    {
      kind: 'dependency_closure',
      message:
        'The complete installed dependency input closure is not frozen by this early gate.',
    },
  ];
  if (unsupported.length)
    blockers.unshift({
      kind: 'missing_rule_implementations',
      names: unsupported,
    });
  const addonHash = fileHash(addon);
  let negativeProbe = {
    status: 'not_run',
    reason:
      'No known missing rule remains; run a complete adapter/correctness gate before scheduling real lint.',
  };
  if (unsupported.length) {
    const rules = enabled.map((name) => {
      const rule = union.get(name);
      assert.ok(
        rule.severity === 0 || rule.severity === 1,
        'disabled rule in admitted work',
      );
      return {
        name,
        severity: rule.severity === 0 ? 2 : 1,
        options: rule.options ?? [],
      };
    });
    const request = {
      schemaVersion: 1,
      cwd: project,
      files: manifest.files.map(([name]) => ({ path: name })),
      rules,
    };
    const result = spawnSync(
      process.execPath,
      [SCRIPT, 'probe', '--addon', addon],
      {
        input: JSON.stringify(request),
        encoding: 'utf8',
        timeout: 10_000,
        maxBuffer: 1 << 20,
        env: { ...process.env, RAYON_NUM_THREADS: '2' },
      },
    );
    assert.equal(
      result.status,
      0,
      'negative capability probe failed to return normally',
    );
    assert.equal(
      result.stderr,
      '',
      'negative capability probe emitted unexpected stderr',
    );
    const reply = JSON.parse(result.stdout);
    assert.equal(
      reply.returned,
      false,
      'candidate unexpectedly accepted the union; do not infer workload parity',
    );
    assert.match(reply.error, /^rule has not been migrated to Rust: /);
    const rejectedRule = reply.error.slice(
      'rule has not been migrated to Rust: '.length,
    );
    assert.ok(
      unsupported.includes(rejectedRule),
      'addon/source capability mismatch',
    );
    negativeProbe = {
      status: 'rejected_as_expected',
      kind: 'full_enabled_name_union_only',
      rules: rules.length,
      listedFiles: request.files.length,
      rejectedRule,
      error: reply.error,
      actualProjectLintExecuted: false,
      explanation:
        'The actual workload already failed configuration admission. This uniform superset is only a negative capability probe; it preserves all 129 names and never substitutes for real per-file lint.',
    };
  }
  assert.equal(fileHash(addon), addonHash, 'addon changed during probe');
  const report = {
    schemaVersion: 1,
    status: 'incomplete',
    comparable: false,
    performance: null,
    projectCommit: PROJECT_PIN,
    baselineCommit: BASELINE_PIN,
    compilerCommit: COMPILER_PIN,
    manifestSHA256: fileHash(manifestFile),
    addonSHA256: addonHash,
    node: process.version,
    workload: EXPECTED,
    verifiedInputFiles: manifest.files.length,
    sourceDeclaredCapabilities: capabilities,
    sourceDeclaredEnabledRules: enabled.filter((name) =>
      capabilities.names.includes(name),
    ),
    sourceDeclaredCoveredRuleFilePairs: implementedPairs,
    missingEnabledRules: unsupported.length,
    negativeProbe,
    blockers,
    conclusion:
      'No valid pure-Rust rsbuild speedup or memory comparison can be reported. No formal A/B was started.',
  };
  portable(report);
  if (args.output) writeJSON(path.resolve(args.output), report);
  console.log(JSON.stringify(report, null, 2));
  process.exitCode = 1;
}

function selfTest() {
  assert.throws(() => inside('/project', '../outside'));
  assert.throws(() => inside('/project', '/outside'));
  assert.throws(() => portable({ secret: '/Users/private/path' }));
  assert.equal(canonical({ b: 2, a: 1 }), canonical({ a: 1, b: 2 }));
  const table = [];
  const indices = new Map();
  assert.equal(intern(table, indices, { b: 2, a: 1 }), 0);
  assert.equal(intern(table, indices, { a: 1, b: 2 }), 0);
  assert.ok(sourceCapabilities().names.length >= 17);
  console.log('Rust rsbuild gate self-test passed');
}

try {
  const [mode, ...rest] = process.argv.slice(2);
  const args = options(rest);
  if (mode === '--help' || mode === 'help') {
    console.log(`Usage:
  node scripts/tsrs/rust-rsbuild-gate.mjs capture --project DIR --baseline DIR
    --trace FILE --summary FILE --source FILE --record FILE --stdout FILE
    --stderr FILE [--manifest FILE]
  node scripts/tsrs/rust-rsbuild-gate.mjs check --project DIR --addon FILE
    [--manifest FILE] [--output FILE]
  node scripts/tsrs/rust-rsbuild-gate.mjs self-test

The default manifest is docs/tsrs-migration/rsbuild-workload.json.
The check verifies all recorded inputs and retains all 129 enabled rule names.
It is a negative correctness/capability gate, never a performance measurement.
Exit codes: 0 capture/self-test success, 1 incomplete migration, 2 invalid input.
Use the exact addon built from the source snapshot under review.`);
  } else if (mode === 'capture') await capture(args);
  else if (mode === 'check') check(args);
  else if (mode === 'probe') probe(args);
  else if (mode === 'self-test') selfTest();
  else
    throw new Error('select capture, check, or self-test; see script header');
} catch (error) {
  // Do not persist private source/cache paths from lower-level filesystem errors.
  let message = String(error.message ?? error).replaceAll(ROOT, '<candidate>');
  for (const value of process.argv.slice(3)) {
    if (!value.startsWith('--'))
      message = message.replaceAll(path.resolve(value), '<input>');
  }
  if (/\/Users\/|\/home\/|[A-Za-z]:\\Users\\/.test(message)) {
    message =
      'Preflight failed; lower-level private path details were redacted.';
  }
  console.error(
    JSON.stringify({
      status: 'invalid_preflight',
      comparable: false,
      performance: null,
      error: message,
    }),
  );
  process.exitCode = 2;
}
