#!/usr/bin/env node
const os = require('node:os');

function main() {
  const binPath = require.resolve(`@rslint/${os.platform()}-${os.arch()}/bin`);

  try {
    require('child_process').execFileSync(binPath, process.argv.slice(2), {
      stdio: 'inherit',
    });
  } catch (error) {
    // Preserve the exit code from the child process
    if (error.status != null) {
      process.exit(error.status);
    } else {
      console.error(`Failed to execute ${binPath}: ${error.message}`);
      process.exit(1);
    }
  }
}

main();
