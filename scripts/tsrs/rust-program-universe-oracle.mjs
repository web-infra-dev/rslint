// Compare actual compiler source graphs. Run this read-only correctness probe
// under the shared machine lock; the observer builds every selected Program.
// It is not a lint, behavior-compatibility or performance measurement.
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { registerHooks } from 'node:module';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { execFileSync } from 'node:child_process';
import { isDeepStrictEqual } from 'node:util';

const hash = (value) => crypto.createHash('sha256').update(value).digest('hex');
const fields = [
  'textSHA256',
  'sourceBytes',
  'isDefaultLibrary',
  'isExternalLibrary',
];

function sourceMap(sources) {
  const map = new Map();
  for (const source of sources) {
    assert.equal(typeof source.path, 'string');
    assert(!map.has(source.path), `duplicate compiler source: ${source.path}`);
    assert.match(source.textSHA256, /^[a-f\d]{64}$/);
    assert(Number.isSafeInteger(source.sourceBytes) && source.sourceBytes >= 0);
    assert.equal(typeof source.isDefaultLibrary, 'boolean');
    assert.equal(typeof source.isExternalLibrary, 'boolean');
    map.set(source.path, source);
  }
  return map;
}

function compareSources(expectedSources, actualSources) {
  const expected = sourceMap(expectedSources);
  const actual = sourceMap(actualSources);
  const differences = [];
  for (const name of [
    ...new Set([...expected.keys(), ...actual.keys()]),
  ].sort()) {
    const before = expected.get(name);
    const after = actual.get(name);
    if (!before || !after) {
      differences.push({
        kind: before ? 'missingSource' : 'extraSource',
        path: name,
        expected: before ?? null,
        actual: after ?? null,
        defaultLibraryOnly: (before ?? after).isDefaultLibrary,
      });
      continue;
    }
    const changes = Object.fromEntries(
      fields
        .filter((field) => before[field] !== after[field])
        .map((field) => [
          field,
          { expected: before[field], actual: after[field] },
        ]),
    );
    if (Object.keys(changes).length) {
      differences.push({
        kind: 'changedSource',
        path: name,
        changes,
        defaultLibraryOnly: before.isDefaultLibrary && after.isDefaultLibrary,
      });
    }
  }
  return differences;
}

function selfTest() {
  const source = {
    path: '/project/source.ts',
    textSHA256: hash('export {};'),
    sourceBytes: 10,
    isDefaultLibrary: false,
    isExternalLibrary: false,
  };
  const external = {
    ...source,
    path: '/store/package/index.d.ts',
    isExternalLibrary: true,
  };
  const library = {
    ...source,
    path: 'bundled:///libs/lib.es5.d.ts',
    isDefaultLibrary: true,
  };
  assert.deepEqual(
    compareSources([source, external, library], [library, external, source]),
    [],
  );
  assert.equal(compareSources([external], [])[0].defaultLibraryOnly, false);
  assert.equal(compareSources([library], [])[0].kind, 'missingSource');
  assert.equal(
    compareSources([library], [{ ...library, textSHA256: hash('changed') }])[0]
      .changes.textSHA256.actual,
    hash('changed'),
  );
  assert.equal(
    compareSources([source], [{ ...source, isExternalLibrary: true }])[0]
      .changes.isExternalLibrary.actual,
    true,
  );
  assert.throws(() => compareSources([source, source], []), /duplicate/);
  console.log(
    'passed: source-graph comparison retains missing dependencies, library changes and flags',
  );
}

async function main() {
  const [projectArg, observerArg, traceArg, outputArg] = process.argv.slice(2);
  assert(
    projectArg && observerArg && traceArg && outputArg,
    'usage: node scripts/tsrs/rust-program-universe-oracle.mjs PROJECT PROGRAM_UNIVERSE_EXECUTABLE EXTERNAL_PROJECT_TRACE OUTPUT',
  );
  const project = fs.realpathSync(projectArg);
  const observer = fs.realpathSync(observerArg);
  const output = path.resolve(outputArg);
  const observationOutput = output.replace(/\.json$/, '') + '.observation.json';
  const stderrOutput = output.replace(/\.json$/, '') + '.observer.stderr.log';
  for (const name of [output, observationOutput, stderrOutput])
    assert(!fs.existsSync(name), `refusing to overwrite evidence: ${name}`);
  fs.mkdirSync(path.dirname(output), { recursive: true });
  const traceBytes = fs.readFileSync(traceArg);
  const trace = JSON.parse(traceBytes);
  const manifestBytes = fs.readFileSync(
    new URL('../../docs/tsrs-migration/rsbuild-workload.json', import.meta.url),
  );
  const manifest = JSON.parse(manifestBytes);
  const sourceRoot = fileURLToPath(
    new URL('../../packages/rslint/src/', import.meta.url),
  );
  const previousCwd = process.cwd();
  // Preserve complete pnpm package/version/content keys while removing the
  // private store root. Both sides use the same transform, with no realpath,
  // basename-only matching, library filtering or dependency substitution.
  const portableString = (value) =>
    value
      .replaceAll(project, '<project>')
      .replace(/\/[^\n"<>]*?\/pnpm\/store\//g, '<pnpm-store>/')
      .replaceAll(os.homedir(), '<home>');
  const portable = (value) => JSON.parse(portableString(JSON.stringify(value)));
  const report = {
    schemaVersion: 1,
    status: 'running',
    scope:
      'Complete Program source membership, source text hashes/bytes and compiler library flags; no lint or performance claim.',
    node: process.version,
    manifestSHA256: hash(manifestBytes),
    traceSHA256: hash(traceBytes),
    observerSHA256: hash(fs.readFileSync(observer)),
    sourceSHA256: {},
    baselineCommit: trace.baselineCommit,
    baselineCompilerCommit: trace.compilerCommit,
    candidateCompilerCommit: '280b426b89cd4ecd6153cf48228826bd60a99c61',
    projectCommit: manifest.project.commit,
    singleThreaded: process.env.RSLINT_PROGRAM_SINGLE_THREADED !== '0',
    programSourceUniverseValidated: false,
    nonDefaultLibrarySourcesMatch: false,
    files: 0,
    plans: [],
  };
  for (const name of [
    'crates/rslint-core/examples/program-universe.rs',
    'crates/rslint-core/src/programs.rs',
    'crates/rslint-core/src/projects.rs',
    'crates/rslint-core/src/source.rs',
    'scripts/tsrs/rust-program-universe-oracle.mjs',
  ]) {
    report.sourceSHA256[name] = hash(
      fs.readFileSync(new URL(`../../${name}`, import.meta.url)),
    );
  }
  const hook = registerHooks({
    resolve(specifier, context, nextResolve) {
      if (
        specifier.startsWith('.') &&
        specifier.endsWith('.js') &&
        context.parentURL?.startsWith(pathToFileURL(sourceRoot).href)
      ) {
        const url = new URL(specifier.slice(0, -3) + '.ts', context.parentURL);
        if (fs.existsSync(url)) return nextResolve(url.href, context);
      }
      return nextResolve(specifier, context);
    },
  });
  try {
    assert.equal(
      execFileSync('git', ['-C', project, 'rev-parse', 'HEAD'], {
        encoding: 'utf8',
      }).trim(),
      manifest.project.commit,
    );
    assert.equal(trace.manifestSHA256, report.manifestSHA256);
    assert.equal(trace.baselineCommit, manifest.baseline.commit);
    assert.equal(trace.projectCommit, manifest.project.commit);
    for (const [name, expected] of manifest.project.inputs)
      assert.equal(
        hash(fs.readFileSync(path.join(project, name))),
        expected,
        name,
      );
    const { admitRustFiles } = await import(
      pathToFileURL(path.join(sourceRoot, 'rust/admission.ts')).href
    );
    process.chdir(project);
    // The production explicit-config loader resolves symlinks before assigning
    // the configuration directory. Keep this observer on that same host cwd.
    const configPath = fs.realpathSync(
      path.join(project, 'node_modules/rstack/dist/rslintConfig.js'),
    );
    const { default: config } = await import(pathToFileURL(configPath).href);
    const admitted = admitRustFiles({
      cwd: project,
      config,
      configDirectory: path.dirname(configPath),
      configArrayBase: project,
    });
    const expectedFiles = new Map(manifest.files.map((row) => [row[0], row]));
    const relative = (name) =>
      path.relative(project, name).split(path.sep).join('/');
    assert.deepEqual(
      admitted.map((file) => relative(file.path)).sort(),
      [...expectedFiles.keys()].sort(),
    );
    const files = admitted.map((file) => {
      const row = expectedFiles.get(relative(file.path));
      const bytes = fs.readFileSync(file.path);
      assert.equal(hash(bytes), row[1]);
      assert.equal(bytes.byteLength, row[2]);
      const policy = file.config.projectPolicy;
      assert(
        !policy.explicitProject,
        'frozen rsbuild unexpectedly requires explicit project expansion',
      );
      if (policy.serviceRootDirectory)
        return {
          path: file.path,
          policy: {
            kind: 'service',
            rootDirectory: policy.serviceRootDirectory,
          },
        };
      assert(
        policy.defaultProjectDisabled || policy.projectDisabled,
        'unhandled implicit default project policy',
      );
      return { path: file.path, policy: { kind: 'sourceOnly' } };
    });
    report.files = files.length;
    const stderr = fs.openSync(stderrOutput, 'wx');
    let actual;
    try {
      actual = JSON.parse(
        execFileSync(observer, [], {
          input: JSON.stringify({
            schemaVersion: 1,
            cwd: project,
            sourceOnlyCurrentDirectory: path.dirname(configPath),
            files,
            singleThreaded: report.singleThreaded,
          }),
          encoding: 'utf8',
          stdio: ['pipe', 'pipe', stderr],
          maxBuffer: 128 * 1024 * 1024,
          timeout: 15 * 60 * 1000,
        }),
      );
    } finally {
      fs.closeSync(stderr);
    }
    assert.equal(actual.backend, 'tsrs-rust');
    assert.equal(actual.compiler.revision, report.candidateCompilerCommit);
    assert.equal(actual.singleThreaded, report.singleThreaded);
    assert.equal(actual.fileProjects.length, files.length);
    const contents = actual.texts.map((text) => ({
      textSHA256: hash(text),
      sourceBytes: Buffer.byteLength(text),
    }));
    delete actual.texts;
    const universes = new Map();
    for (const universe of actual.universes) {
      assert(
        !universes.has(universe.projectIndex),
        'duplicate observed Program',
      );
      universe.sourceFiles = universe.sourceFiles.map(
        ({ textIndex, ...source }) => {
          assert(
            Number.isSafeInteger(textIndex) && contents[textIndex],
            'invalid compiler text reference',
          );
          assert.equal(
            source.sourceBytes,
            contents[textIndex].sourceBytes,
            source.path,
          );
          return { ...source, textSHA256: contents[textIndex].textSHA256 };
        },
      );
      universes.set(universe.projectIndex, universe);
    }
    const oracleByFile = new Map();
    const planIds = new Set();
    for (const plan of trace.plans) {
      assert(!planIds.has(plan.manifestPlanIndex), 'duplicate manifest plan');
      planIds.add(plan.manifestPlanIndex);
      assert(plan.admittedFiles.length > 0, 'empty oracle plan');
      for (const file of plan.admittedFiles) {
        assert(!oracleByFile.has(file), 'duplicate oracle admitted file');
        oracleByFile.set(file, plan);
      }
    }
    assert.deepEqual(
      [...oracleByFile.keys()].sort(),
      [...expectedFiles.keys()].sort(),
    );
    const byPlan = new Map();
    for (const [index, file] of files.entries()) {
      const name = relative(file.path);
      const expected = oracleByFile.get(name);
      const projectIndex = actual.fileProjects[index];
      const typed =
        manifest.configurations[expectedFiles.get(name)[3]].hasTypeChecker;
      assert.equal(projectIndex !== null, typed, `${name}: type capability`);
      assert(universes.has(projectIndex), `${name}: missing compiler Program`);
      const members = byPlan.get(expected.manifestPlanIndex) ?? {
        projectIndex,
        files: [],
      };
      assert.equal(
        members.projectIndex,
        projectIndex,
        `${name}: baseline plan split across candidate Programs`,
      );
      members.files.push(name);
      byPlan.set(expected.manifestPlanIndex, members);
    }
    assert.equal(byPlan.size, trace.plans.length);
    const observedProjectIndexes = new Set();
    for (const expectedPlan of trace.plans) {
      const membership = byPlan.get(expectedPlan.manifestPlanIndex);
      assert.deepEqual(
        [...membership.files].sort(),
        [...expectedPlan.admittedFiles].sort(),
      );
      const index = membership.projectIndex;
      observedProjectIndexes.add(index);
      const universe = portable(universes.get(index));
      const expected = portable(expectedPlan.project);
      const selected = index === null ? null : portable(actual.projects[index]);
      const metadataDifferences = [];
      for (const [name, candidate] of Object.entries({
        configFileName: selected?.configFileName ?? '',
        currentDirectory: universe.currentDirectory,
        compilerOptions: universe.compilerOptions,
        rootFiles: [
          ...(selected?.rootFiles ??
            membership.files.map((name) => `<project>/${name}`)),
        ].sort(),
      })) {
        const baseline =
          name === 'rootFiles' ? [...expected[name]].sort() : expected[name];
        if (!isDeepStrictEqual(baseline, candidate))
          metadataDifferences.push({
            field: name,
            expected: baseline,
            actual: candidate,
          });
      }
      const differences = compareSources(
        expected.sourceFiles,
        universe.sourceFiles,
      );
      report.plans.push({
        manifestPlanIndex: expectedPlan.manifestPlanIndex,
        candidateProjectIndex: index,
        configFileName: selected?.configFileName ?? '',
        admittedFiles: membership.files,
        baselineSourceCount: expected.sourceFiles.length,
        candidateSourceCount: universe.sourceFiles.length,
        exactSourceMatch: differences.length === 0,
        nonDefaultLibrarySourcesMatch: differences.every(
          (difference) => difference.defaultLibraryOnly,
        ),
        metadataDifferences,
        differences,
      });
    }
    assert.equal(
      observedProjectIndexes.size,
      universes.size,
      'uncompared candidate Program',
    );
    const observation = portable({
      ...actual,
      manifestSHA256: report.manifestSHA256,
      observerSHA256: report.observerSHA256,
    });
    const observationBytes = `${JSON.stringify(observation)}\n`;
    fs.writeFileSync(observationOutput, observationBytes, { flag: 'wx' });
    report.observation = {
      file: path.basename(observationOutput),
      sha256: hash(observationBytes),
    };
    report.programSourceUniverseValidated = report.plans.every(
      (plan) => plan.exactSourceMatch,
    );
    report.nonDefaultLibrarySourcesMatch = report.plans.every(
      (plan) => plan.nonDefaultLibrarySourcesMatch,
    );
    report.sourceDifferences = report.plans.reduce(
      (sum, plan) => sum + plan.differences.length,
      0,
    );
    report.metadataDifferences = report.plans.reduce(
      (sum, plan) => sum + plan.metadataDifferences.length,
      0,
    );
    report.defaultLibraryDifferencePolicy =
      'Retained as mismatches. Different compiler pins are context, not proof that a library difference is harmless.';
    report.status =
      report.programSourceUniverseValidated && report.metadataDifferences === 0
        ? 'passed'
        : 'different';
    if (report.status !== 'passed') process.exitCode = 1;
  } catch (error) {
    report.status = 'failed';
    report.error = portableString(String(error.stack ?? error));
    process.exitCode = 1;
  } finally {
    process.chdir(previousCwd);
    hook.deregister();
    fs.writeFileSync(output, `${JSON.stringify(report, null, 2)}\n`, {
      flag: 'wx',
    });
  }
  console.log(
    `${report.status}: ${report.files} files, ${report.plans.length} plans, ${report.sourceDifferences ?? 'unknown'} source differences; ${output}`,
  );
}

if (process.argv[2] === '--self-test') selfTest();
else await main();
