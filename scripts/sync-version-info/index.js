#!/usr/bin/env node

const { syncFullHistory, syncCurrentVersion } = require('./release');
const { getTypeScriptBinding } = require('./version');

function printUsage() {
  console.log('\nUsage:');
  console.log(
    '  pnpm sync:version-info       Sync main and the current stable release',
  );
  console.log(
    '  pnpm sync:version-info full  Rebuild rule history, preserving TypeScript bindings',
  );
}

function main() {
  const args = process.argv.slice(2);
  if (args.length > 1 || (args[0] && args[0] !== 'full')) {
    throw new Error('The only supported argument is "full"');
  }

  if (args[0] === 'full') {
    syncFullHistory();
  } else {
    syncCurrentVersion(getTypeScriptBinding);
  }
  console.log('Generated website/releases.json.');
}

try {
  main();
} catch (error) {
  console.error(`Version info sync failed: ${error.message}`);
  process.exitCode = 1;
} finally {
  printUsage();
}
