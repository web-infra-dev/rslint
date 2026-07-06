#!/usr/bin/env node

async function main() {
  const { runCLI } = await import('../dist/cli.js');
  await runCLI();
}

main().catch((err) => {
  process.stderr.write(`rslint: ${err}\n`);
  process.exitCode = 1;
});
