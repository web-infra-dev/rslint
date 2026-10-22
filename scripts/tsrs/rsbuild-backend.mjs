// Benchmark-only package routing. No installed package or lint input is edited.
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { registerHooks, syncBuiltinESMExports } from 'node:module';
import { pathToFileURL } from 'node:url';
import childProcess from 'node:child_process';

const core = process.env.RSLINT_BENCH_CORE;
assert(core, 'RSLINT_BENCH_CORE must select the complete built core package');
const root = path.resolve(core);
const pkg = JSON.parse(
  fs.readFileSync(path.join(root, 'package.json'), 'utf8'),
);
assert.equal(pkg.name, '@rslint/core');
registerHooks({
  resolve(specifier, context, next) {
    if (specifier === '@rslint/core' || specifier.startsWith('@rslint/core/')) {
      const key =
        specifier === '@rslint/core'
          ? '.'
          : `.${specifier.slice('@rslint/core'.length)}`;
      const entry = pkg.exports[key];
      const target = typeof entry === 'string' ? entry : entry?.default;
      assert(target, `missing core export: ${specifier}`);
      const file = path.resolve(root, target);
      assert(fs.statSync(file).isFile(), `unbuilt core export: ${file}`);
      if (process.env.RSLINT_BENCH_RESOLUTIONS) {
        fs.appendFileSync(
          process.env.RSLINT_BENCH_RESOLUTIONS,
          `${JSON.stringify({ specifier, file })}\n`,
        );
      }
      return next(pathToFileURL(file).href, context);
    }
    return next(specifier, context);
  },
});

// Correctness-only Go overlay binaries have the same production source plus a
// file/rule trace. Timed invocations never set this override or write traces.
if (process.env.RSLINT_BENCH_BINARY) {
  const expected = fs.realpathSync(process.env.RSLINT_BENCH_ORIGINAL_BINARY);
  const replacement = fs.realpathSync(process.env.RSLINT_BENCH_BINARY);
  const originalSpawn = childProcess.spawn;
  childProcess.spawn = function (command, ...args) {
    let actual;
    try {
      actual = fs.realpathSync(command);
    } catch {
      /* unrelated executable */
    }
    return originalSpawn.call(
      this,
      actual === expected ? replacement : command,
      ...args,
    );
  };
  syncBuiltinESMExports();
}
