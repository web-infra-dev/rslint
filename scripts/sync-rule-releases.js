#!/usr/bin/env node
// Retain the existing command while release metadata moves to releases.json.
import('./sync-releases.mjs')
  .then(({ main }) => main())
  .catch((error) => {
    console.error(error.message);
    process.exitCode = 1;
  });
