// Verify independent JS admission and Rust project selection, before linting.
// The external trace is an oracle input, never a source of candidate decisions.
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';
import path from 'node:path';
import { createRequire, registerHooks } from 'node:module';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { execFileSync } from 'node:child_process';

const [projectArg, addonArg, traceArg, outputArg] = process.argv.slice(2);
assert(
  projectArg && addonArg && traceArg && outputArg,
  'usage: node scripts/tsrs/rust-project-oracle.mjs PROJECT ADDON EXTERNAL_PROJECT_TRACE OUTPUT',
);
const project = fs.realpathSync(projectArg);
const addon = fs.realpathSync(addonArg);
const output = path.resolve(outputArg);
const trace = JSON.parse(fs.readFileSync(traceArg, 'utf8'));
const manifestBytes = fs.readFileSync(
  new URL('../../docs/tsrs-migration/rsbuild-workload.json', import.meta.url),
);
const manifest = JSON.parse(manifestBytes);
const hash = (value) => crypto.createHash('sha256').update(value).digest('hex');
const sourceRoot = fileURLToPath(
  new URL('../../packages/rslint/src/', import.meta.url),
);
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
const previousCwd = process.cwd();
const portable = (value) =>
  JSON.parse(JSON.stringify(value).replaceAll(project, '<project>'));
const report = {
  schemaVersion: 1,
  status: 'running',
  node: process.version,
  manifestSHA256: hash(manifestBytes),
  traceSHA256: hash(fs.readFileSync(traceArg)),
  addonSHA256: hash(fs.readFileSync(addon)),
  scope:
    'File admission and selected project configuration/root/options only; no lint or performance claim.',
  programSourceUniverseValidated: false,
  files: 0,
  typedFiles: 0,
  projects: [],
};
try {
  assert.equal(
    execFileSync('git', ['-C', project, 'rev-parse', 'HEAD'], {
      encoding: 'utf8',
    }).trim(),
    manifest.project.commit,
  );
  assert.equal(trace.manifestSHA256, report.manifestSHA256);
  assert.equal(trace.baselineCommit, manifest.baseline.commit);
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
  const configPath = path.join(
    project,
    'node_modules/rstack/dist/rslintConfig.js',
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
      'frozen rsbuild unexpectedly needs explicit project expansion',
    );
    if (policy.serviceRootDirectory)
      return {
        path: file.path,
        policy: { kind: 'service', rootDirectory: policy.serviceRootDirectory },
      };
    assert(
      policy.defaultProjectDisabled || policy.projectDisabled,
      'implicit default project policy is not covered by this oracle',
    );
    return { path: file.path, policy: { kind: 'sourceOnly' } };
  });
  const binding = createRequire(import.meta.url)(addon);
  const actual = JSON.parse(
    binding.planTsrsSync(
      JSON.stringify({ schemaVersion: 1, cwd: project, files }),
    ),
  );
  assert.equal(actual.backend, 'tsrs-rust');
  assert.equal(
    actual.compiler.revision,
    '280b426b89cd4ecd6153cf48228826bd60a99c61',
  );
  assert.equal(actual.fileProjects.length, files.length);
  const oracleByFile = new Map();
  for (const plan of trace.plans)
    for (const file of plan.admittedFiles) {
      assert(!oracleByFile.has(file), 'duplicate oracle file');
      oracleByFile.set(file, plan);
    }
  const validated = new Set();
  for (const [index, file] of admitted.entries()) {
    const name = relative(file.path);
    const expected = oracleByFile.get(name);
    assert(expected, `missing oracle plan for ${name}`);
    const typed =
      manifest.configurations[expectedFiles.get(name)[3]].hasTypeChecker;
    const projectIndex = actual.fileProjects[index];
    assert.equal(
      projectIndex !== null,
      typed,
      `checker capability differs: ${name}`,
    );
    if (!typed) continue;
    report.typedFiles++;
    const selected = actual.projects[projectIndex];
    const expectedProject = expected.project;
    assert.equal(
      portable(selected.configFileName),
      expectedProject.configFileName,
      name,
    );
    const key = `${projectIndex}/${expected.manifestPlanIndex}`;
    if (validated.has(key)) continue;
    validated.add(key);
    const observed = portable(selected);
    assert.equal(
      observed.currentDirectory,
      expectedProject.currentDirectory,
      name,
    );
    assert.deepEqual(
      [...observed.rootFiles].sort(),
      [...expectedProject.rootFiles].sort(),
      `${name}: roots`,
    );
    assert.deepEqual(
      observed.compilerOptions,
      expectedProject.compilerOptions,
      `${name}: options`,
    );
    assert.equal(observed.sourceReferences, true);
    report.projects.push({
      configFileName: observed.configFileName,
      roots: observed.rootFiles.length,
      rootsSHA256: hash(JSON.stringify([...observed.rootFiles].sort())),
      compilerOptions: observed.compilerOptions,
      manifestPlanIndex: expected.manifestPlanIndex,
    });
  }
  report.files = admitted.length;
  assert.equal(report.typedFiles, manifest.counts.typed_files);
  assert.equal(
    report.projects.length,
    trace.plans.filter((plan) => plan.project.configFileName !== '').length,
  );
  report.status = 'passed';
} catch (error) {
  report.status = 'failed';
  report.error = String(error.stack ?? error).replaceAll(project, '<project>');
  process.exitCode = 1;
} finally {
  process.chdir(previousCwd);
  hook.deregister();
  fs.mkdirSync(path.dirname(output), { recursive: true });
  fs.writeFileSync(output, `${JSON.stringify(report, null, 2)}\n`);
}
console.log(
  `${report.status}: ${report.files} files, ${report.typedFiles} typed, ${report.projects.length} project plans; ${output}`,
);
