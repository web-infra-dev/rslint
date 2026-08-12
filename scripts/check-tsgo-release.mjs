#!/usr/bin/env node

import {
  TSGO_NPM_PACKAGES,
  readTsgoReleaseState,
  validateTsgoReleaseState,
} from './tsgo-release-utils.mjs';

try {
  const state = await readTsgoReleaseState();
  const version = validateTsgoReleaseState(state);
  console.log(
    `Validated ${TSGO_NPM_PACKAGES.length} npm packages and tsgo-client at ${version}`,
  );
} catch (error) {
  console.error(error.message);
  process.exit(1);
}
