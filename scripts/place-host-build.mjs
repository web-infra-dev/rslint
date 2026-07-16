#!/usr/bin/env node
/**
 * Place host-platform build outputs into the matching `@rslint/native-<tuple>`
 * platform package, so the core loader (`load-binding.ts`) and the bin launcher
 * (`bin/rslint.js`) resolve them in dev EXACTLY as in prod — no dev-only path.
 *
 *   node scripts/place-host-build.mjs bin [--debug]   # go build the rslint CLI
 *   node scripts/place-host-build.mjs node            # copy the napi .node
 *
 * In prod the same files arrive via `npm install` (one platform package, filtered
 * by os/cpu/libc) / CI's move-artifacts. Here we just fill the host one.
 */
import { execFileSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const repoRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  '..',
);

// musl detection on Linux (a glibc and a musl .node are not interchangeable).
function isMusl() {
  if (process.platform !== 'linux') return false;
  try {
    if (fs.readFileSync('/usr/bin/ldd', 'utf-8').includes('musl')) return true;
  } catch {}
  try {
    return execFileSync('ldd', ['--version'], { encoding: 'utf8' }).includes(
      'musl',
    );
  } catch {
    return false;
  }
}

function hostTuple() {
  const { platform, arch } = process;
  if (platform === 'darwin') return `darwin-${arch}`;
  if (platform === 'win32') return `win32-${arch}-msvc`;
  if (platform === 'linux') return `linux-${arch}-${isMusl() ? 'musl' : 'gnu'}`;
  throw new Error(`unsupported host platform ${platform}-${arch}`);
}

const what = process.argv[2];
const tuple = hostTuple();
const pkgDir = path.join(repoRoot, 'npm', 'rslint', tuple);
fs.mkdirSync(pkgDir, { recursive: true });

if (what === 'bin') {
  const debug = process.argv.includes('--debug');
  const ext = process.platform === 'win32' ? '.exe' : '';
  const out = path.join(pkgDir, `rslint${ext}`);
  const args = ['build', '-v'];
  if (debug) args.push('-gcflags=all=-N -l');
  args.push('-o', out, './cmd/rslint');
  execFileSync('go', args, {
    cwd: repoRoot,
    stdio: 'inherit',
    env: debug ? { ...process.env, GOEXPERIMENT: 'noregabi' } : process.env,
  });
} else if (what === 'node') {
  const file = `rslint.${tuple}.node`;
  const src = path.join(repoRoot, 'crates', 'rslint-native', file);
  if (!fs.existsSync(src)) {
    throw new Error(
      `napi build output not found: ${src}\n` +
        `run \`pnpm --filter @rslint/native build\` first`,
    );
  }
  fs.copyFileSync(src, path.join(pkgDir, file));
} else {
  throw new Error('usage: place-host-build.mjs <bin|node> [--debug]');
}
