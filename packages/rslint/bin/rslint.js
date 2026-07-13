#!/usr/bin/env node
import nodeModule from 'node:module';

// Enable on-disk code caching for modules loaded by Node.js.
// Available in Node.js >= 22.8.0.
const { enableCompileCache } = nodeModule;
if (enableCompileCache) {
  try {
    enableCompileCache();
  } catch {
    // Ignore cache setup errors; the CLI should still run normally.
  }
}

async function main() {
  const { runCLI } = await import('../dist/cli.js');
  await runCLI();
}

main().catch((err) => {
  process.stderr.write(`rslint: ${err}\n`);
  process.exitCode = 1;
});
