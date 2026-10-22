#!/usr/bin/env node
process.stderr.write(
  `${process.argv[2] || 'This entrypoint'} is not supported by the Rust+JS rsbuild migration. ` +
    'The migration currently builds the Node CLI/API and native Rust addon only.\n',
);
process.exitCode = 1;
