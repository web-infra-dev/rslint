// Correctness probe only. Run Node's tests with --test-isolation=none.
// A direct Rust lint call must not invoke another compiler or lint process.
import childProcess from 'node:child_process';
import { syncBuiltinESMExports } from 'node:module';

for (const name of [
  'spawn',
  'spawnSync',
  'exec',
  'execSync',
  'execFile',
  'execFileSync',
  'fork',
]) {
  childProcess[name] = () => {
    throw new Error(`Direct Rust runtime attempted child_process.${name}`);
  };
}
syncBuiltinESMExports();
