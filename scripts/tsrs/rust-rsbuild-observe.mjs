// The guard is also used by timed runs. Native request/result tracing is only
// enabled for correctness proof and is never enabled during measurement.
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import childProcess from 'node:child_process';
import { syncBuiltinESMExports } from 'node:module';
import { pathToFileURL } from 'node:url';

const addon = process.env.RSLINT_RUST_GUARD;
const observation = process.env.RSLINT_RUST_OBSERVATION;
if (addon) {
  const expected = fs.realpathSync(addon);
  let sequence = 0;
  const append = (row) => {
    if (observation) fs.appendFileSync(observation, `${JSON.stringify(row)}\n`);
  };
  for (const name of [
    'spawn',
    'spawnSync',
    'exec',
    'execSync',
    'execFile',
    'execFileSync',
    'fork',
  ]) {
    childProcess[name] = function (...args) {
      append({
        event: 'forbiddenChildProcess',
        method: name,
        argv: args.slice(0, 2),
      });
      throw new Error(
        `Rust benchmark candidate must not spawn child processes: ${name}`,
      );
    };
  }
  syncBuiltinESMExports();
  const original = process.dlopen;
  process.dlopen = function (module, filename, ...args) {
    const actual = fs.realpathSync(filename);
    assert.equal(actual, expected, 'candidate loaded an unpinned native addon');
    const result = original.call(this, module, filename, ...args);
    append({ event: 'nativeAddon', filename: actual });
    const exports = module.exports;
    assert.equal(
      typeof exports.lintProjectsTsrsSync,
      'function',
      'native configured-lint export missing',
    );
    // N-API properties can be non-configurable. Proxy a fresh object so the
    // original property's descriptor cannot prevent correctness observation.
    if (observation) {
      assert.equal(
        typeof exports.planTsrsSync,
        'function',
        'native project planner missing',
      );
      const wrapped = Object.create(Object.getPrototypeOf(exports));
      for (const name of Object.getOwnPropertyNames(exports)) {
        const value = exports[name];
        if (name !== 'lintProjectsTsrsSync') {
          Object.defineProperty(
            wrapped,
            name,
            Object.getOwnPropertyDescriptor(exports, name),
          );
          continue;
        }
        Object.defineProperty(wrapped, name, {
          enumerable: true,
          value(request) {
            const call = ++sequence;
            const parsed = JSON.parse(request);
            append({ event: 'nativeRequest', call, request: parsed });
            try {
              // Resolve the very policies and overlays sent by the normal CLI
              // or public API. The baseline/expected manifest is never supplied
              // to this native planner and cannot select candidate projects.
              const planRequest = {
                schemaVersion: parsed.schemaVersion,
                cwd: parsed.cwd,
                files: parsed.files.map(({ path: filename, policy }) => ({
                  path: filename,
                  policy,
                })),
                sources: [
                  ...(parsed.sources ?? []),
                  ...parsed.files
                    .filter((file) => Object.hasOwn(file, 'text'))
                    .map(({ path: filename, text }) => ({
                      path: filename,
                      text,
                    })),
                ],
              };
              const plan = JSON.parse(
                exports.planTsrsSync(JSON.stringify(planRequest)),
              );
              append({
                event: 'nativeProjectPlan',
                call,
                request: planRequest,
                plan,
              });
              const response = value.call(exports, request);
              append({
                event: 'nativeResponse',
                call,
                response: JSON.parse(response),
              });
              return response;
            } catch (error) {
              append({ event: 'nativeError', call, error: String(error) });
              throw error;
            }
          },
        });
      }
      module.exports = wrapped;
    }
    return result;
  };
}

if (
  process.argv[1] &&
  import.meta.url === pathToFileURL(path.resolve(process.argv[1])).href
) {
  const [mode, config, destination] = process.argv.slice(2);
  assert.equal(mode, '--api');
  assert(
    config && destination,
    'usage: rust-rsbuild-observe.mjs --api CONFIG OUTPUT',
  );
  const { Rslint } = await import('@rslint/core');
  const result = { schema_version: 1 };
  for (const fix of [false, true]) {
    const linter = new Rslint({
      cwd: process.cwd(),
      overrideConfigFile: path.resolve(config),
      fix,
    });
    try {
      const files = await linter.lintFiles(['.']);
      for (const entry of files)
        entry.filePath = path
          .relative(process.cwd(), entry.filePath)
          .split(path.sep)
          .join('/');
      files.sort((a, b) =>
        Buffer.compare(Buffer.from(a.filePath), Buffer.from(b.filePath)),
      );
      result[fix ? 'fixed' : 'lint'] = files;
    } finally {
      await linter.close();
    }
  }
  fs.writeFileSync(destination, `${JSON.stringify(result)}\n`);
}
