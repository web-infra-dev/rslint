// Compare positive and negative rule fixtures against the frozen external engine.
import assert from 'node:assert/strict';
import crypto from 'node:crypto';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { createRequire } from 'node:module';
import { execFileSync } from 'node:child_process';
import { pathToFileURL } from 'node:url';

const [baselineArg, addonArg, casesArg, outputArg] = process.argv.slice(2);
assert(
  baselineArg && addonArg && casesArg && outputArg,
  'usage: node scripts/tsrs/rust-rule-oracle.mjs EXTERNAL_BASELINE ADDON CASES_JSON OUTPUT',
);
const baseline = fs.realpathSync(baselineArg);
const addon = fs.realpathSync(addonArg);
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
const binding = createRequire(import.meta.url)(addon);
const cases = JSON.parse(fs.readFileSync(casesArg, 'utf8'));
assert(Array.isArray(cases) && cases.length > 0, 'empty rule selection');
const cwd = fs.realpathSync(
  fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-rule-oracle-')),
);
const report = {
  schemaVersion: 1,
  status: 'running',
  baselineCommit: pin,
  compilerCommit: '280b426b89cd4ecd6153cf48228826bd60a99c61',
  node: process.version,
  addonSha256: crypto
    .createHash('sha256')
    .update(fs.readFileSync(addon))
    .digest('hex'),
  casesSha256: crypto
    .createHash('sha256')
    .update(fs.readFileSync(casesArg))
    .digest('hex'),
  scope:
    'Selected positive and negative fixtures, diagnostic and automatic-fix parity; not complete rule-suite or project acceptance.',
  candidateEntryPoint:
    process.env.RSLINT_ORACLE_CONFIGURED === '1'
      ? 'lintProjectsTsrsSync'
      : 'lintTsrsSync',
  results: [],
  failures: [],
};
const extensions = [
  '.ts',
  '.tsx',
  '.cts',
  '.mts',
  '.js',
  '.jsx',
  '.cjs',
  '.mjs',
  '.d.ts',
  '.d.cts',
  '.d.mts',
];
function normalize(files, directory) {
  return files
    .map((file) => ({
      filePath: path
        .relative(directory, file.filePath)
        .split(path.sep)
        .join('/'),
      messages: file.messages.map(({ range, ...message }) => message),
      ...(file.output === undefined ? {} : { output: file.output }),
    }))
    .sort((a, b) => a.filePath.localeCompare(b.filePath));
}
try {
  const seen = new Set();
  for (const group of cases) {
    const groupKey = JSON.stringify([group.rule, group.options ?? []]);
    assert(
      typeof group.rule === 'string' && !seen.has(groupKey),
      'duplicate/missing rule',
    );
    seen.add(groupKey);
    assert(
      group.valid?.length > 0 && group.invalid?.length > 0,
      `rule needs positive and negative cases: ${group.rule}`,
    );
    for (const category of ['valid', 'invalid']) {
      for (const [index, entry] of group[category].entries()) {
        const text = typeof entry === 'string' ? entry : entry.text;
        const extension =
          typeof entry === 'string' ? '.ts' : (entry.extension ?? '.ts');
        assert(
          typeof text === 'string' && extensions.includes(extension),
          `unsupported fixture input: ${group.rule}/${category}/${index}/${extension}`,
        );
        const directory = path.join(cwd, `case-${report.results.length}`);
        fs.mkdirSync(directory);
        const file =
          typeof entry === 'object'
            ? (entry.fileName ?? `input${extension}`)
            : `input${extension}`;
        assert(
          typeof file === 'string' &&
            !path.isAbsolute(file) &&
            file.endsWith(extension) &&
            path.resolve(directory, file).startsWith(directory + path.sep),
          'invalid fixture filename',
        );
        fs.mkdirSync(path.dirname(path.join(directory, file)), {
          recursive: true,
        });
        fs.writeFileSync(path.join(directory, file), text);
        const fixture =
          typeof entry === 'object' && entry.fixtureSet !== undefined
            ? group.fixtureSets?.[entry.fixtureSet]
            : {};
        assert(
          fixture && typeof fixture === 'object' && !Array.isArray(fixture),
          'unknown or malformed fixture set',
        );
        const supportFiles = {
          ...(group.files ?? {}),
          ...(fixture.files ?? {}),
          ...(typeof entry === 'object' ? (entry.files ?? {}) : {}),
        };
        // The admitted input overlays a fixture's placeholder without admitting it twice.
        delete supportFiles[file];
        assert(
          supportFiles &&
            typeof supportFiles === 'object' &&
            !Array.isArray(supportFiles),
        );
        for (const [name, source] of Object.entries(supportFiles)) {
          assert(
            typeof source === 'string' &&
              !path.isAbsolute(name) &&
              name !== file &&
              name !== 'tsconfig.json',
          );
          const absolute = path.resolve(directory, name);
          assert(
            absolute.startsWith(directory + path.sep),
            'support file escapes fixture',
          );
          fs.mkdirSync(path.dirname(absolute), { recursive: true });
          fs.writeFileSync(absolute, source);
        }
        const overrideCompilerOptions = {
          ...(group.compilerOptions ?? {}),
          ...(fixture.compilerOptions ?? {}),
          ...(typeof entry === 'object' ? (entry.compilerOptions ?? {}) : {}),
        };
        assert(
          overrideCompilerOptions !== null &&
            typeof overrideCompilerOptions === 'object' &&
            !Array.isArray(overrideCompilerOptions),
          'compilerOptions must be an object',
        );
        const compilerOptions = {
          ...(group.exactCompilerOptions
            ? {}
            : {
                strict: true,
                allowJs: true,
                target: 'ES2022',
                module: 'ESNext',
                moduleResolution: 'Bundler',
                jsx: 'preserve',
              }),
          ...overrideCompilerOptions,
        };
        const projectFiles =
          typeof entry === 'object'
            ? (entry.projectFiles ?? fixture.projectFiles ?? [file])
            : (fixture.projectFiles ?? [file]);
        assert(
          Array.isArray(projectFiles) &&
            projectFiles.includes(file) &&
            new Set(projectFiles).size === projectFiles.length,
          'projectFiles must include the admitted input exactly once',
        );
        for (const name of projectFiles) {
          assert(
            typeof name === 'string' &&
              (name === file || Object.hasOwn(supportFiles, name)),
            'project root must be a declared fixture input or support file',
          );
        }
        fs.writeFileSync(
          path.join(directory, 'tsconfig.json'),
          JSON.stringify({
            compilerOptions,
            files: projectFiles,
          }),
        );
        const expectSyntaxErrors =
          typeof entry === 'object' && entry.expectSyntaxErrors === true;
        const options = group.options ?? [];
        const record = {
          rule: group.rule,
          category,
          index,
          source: text,
          fileName: file,
          supportFiles,
          projectFiles,
          extension,
          options,
          compilerOptions,
          expectSyntaxErrors,
          checks: [],
        };
        report.results.push(record);
        for (const fix of [false, true]) {
          const check = { fix };
          record.checks.push(check);
          try {
            const linter = new Rslint({
              cwd: directory,
              fix,
              overrideConfigFile: true,
              overrideConfig: [
                {
                  files: extensions.map((extension) => `**/*${extension}`),
                  plugins: [
                    ...new Set([
                      '@typescript-eslint',
                      ...(group.rule.includes('/')
                        ? [group.rule.slice(0, group.rule.lastIndexOf('/'))]
                        : []),
                    ]),
                  ],
                  languageOptions: {
                    parserOptions: {
                      project: path.join(directory, 'tsconfig.json'),
                    },
                  },
                  rules: { [group.rule]: ['error', ...options] },
                },
              ],
            });
            let expected;
            try {
              expected = normalize(await linter.lintFiles([file]), directory);
            } finally {
              await linter.close();
            }
            check.expected = expected;
            const result = JSON.parse(
              binding[report.candidateEntryPoint](
                JSON.stringify(
                  process.env.RSLINT_ORACLE_CONFIGURED === '1'
                    ? {
                        schemaVersion: 1,
                        cwd: directory,
                        files: [
                          {
                            path: file,
                            policy: {
                              kind: 'explicit',
                              configFiles: [
                                path.join(directory, 'tsconfig.json'),
                              ],
                            },
                            rules: [{ name: group.rule, options }],
                          },
                        ],
                        fix,
                      }
                    : {
                        schemaVersion: 1,
                        cwd: directory,
                        project: 'tsconfig.json',
                        files: [{ path: file }],
                        rules: [{ name: group.rule, options }],
                        fix,
                      },
                ),
              ),
            );
            assert.equal(result.backend, 'tsrs-rust');
            assert.equal(
              result.files.length,
              1,
              'silent admitted files must be retained',
            );
            check.actual = normalize(result.files, directory);
            assert.deepEqual(
              check.actual,
              expected,
              `${group.rule}/${category}/${index}/fix=${fix}`,
            );
            // A baseline automatic fix can introduce syntax errors. Validate the
            // final generation against the oracle, not the initial input flag.
            const finalSyntaxErrors = expected[0].messages.some((message) =>
              /^TypeScript\(TS\d+\)$/.test(message.ruleId),
            );
            assert.deepEqual(
              result.files[0].appliedRules,
              finalSyntaxErrors ? [] : [group.rule],
            );
            assert.equal(
              result.files[0].hasSyntaxErrors === true,
              finalSyntaxErrors,
            );
            if (!fix) {
              assert.equal(
                finalSyntaxErrors,
                expectSyntaxErrors,
                'fixture syntax expectation disagrees with original engine',
              );
              if (expectSyntaxErrors)
                assert(
                  expected[0].messages.length > 0 &&
                    expected[0].messages.every((message) =>
                      /^TypeScript\(TS\d+\)$/.test(message.ruleId),
                    ),
                  'syntax case must retain compiler diagnostics',
                );
              else if (category === 'valid')
                assert.deepEqual(
                  expected[0].messages,
                  [],
                  'valid case unexpectedly reports',
                );
              else
                assert(
                  expected[0].messages.some(
                    (message) => message.ruleId === group.rule,
                  ),
                  'invalid case does not exercise its rule',
                );
            }
            for (const [name, source] of Object.entries(supportFiles))
              assert.equal(
                fs.readFileSync(path.join(directory, name), 'utf8'),
                source,
                'lint changed support file on disk',
              );
            assert.equal(
              fs.readFileSync(path.join(directory, file), 'utf8'),
              text,
              'lint changed source on disk',
            );
          } catch (error) {
            check.error = (error.stack ?? String(error)).replaceAll(
              cwd,
              '<fixture>',
            );
            report.failures.push({
              rule: group.rule,
              options,
              category,
              index,
              fix,
              error: check.error,
            });
          }
        }
      }
    }
  }
  report.status = report.failures.length ? 'failed' : 'passed';
  if (report.failures.length) {
    report.error = `${report.failures.length} failed checks; all selected inputs were attempted`;
    process.exitCode = 1;
  }
} catch (error) {
  report.status = 'failed';
  report.error = (error.stack ?? String(error)).replaceAll(cwd, '<fixture>');
  process.exitCode = 1;
} finally {
  fs.mkdirSync(path.dirname(path.resolve(outputArg)), { recursive: true });
  fs.writeFileSync(outputArg, `${JSON.stringify(report, null, 2)}\n`);
  fs.rmSync(cwd, { recursive: true, force: true });
}
console.log(`${report.status}: ${report.results.length} cases; ${outputArg}`);
