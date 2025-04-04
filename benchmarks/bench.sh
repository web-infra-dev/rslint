#!/usr/bin/env -S bash -euxo pipefail

export NODE_OPTIONS="--max-old-space-size=16384" 

pushd vscode
export TSGOLINT_BENCHMARK_PROJECT=vscode
hyperfine --warmup 1 --ignore-failure \
  --command-name 'eslint' 'node ./node_modules/eslint/bin/eslint.js --no-inline-config --config ./eslint.config.mjs ./src' \
  --command-name 'tsgolint' '../../tsgolint --tsconfig ./src/tsconfig.json'
popd

pushd typescript
export TSGOLINT_BENCHMARK_PROJECT=typescript
hyperfine --warmup 1 --ignore-failure \
  --command-name 'eslint' 'node ./node_modules/eslint/bin/eslint.js --no-inline-config --config ./eslint.config.mjs ./src' \
  --command-name 'tsgolint' '../../tsgolint --tsconfig ./src/tsconfig-eslint.json'
popd

pushd typeorm
export TSGOLINT_BENCHMARK_PROJECT=typeorm
hyperfine --warmup 1 --ignore-failure \
  --command-name 'eslint' 'node ./node_modules/eslint/bin/eslint.js --no-inline-config --config ./eslint.config.mjs .' \
  --command-name 'tsgolint' '../../tsgolint --tsconfig ./tsconfig.json'
popd

pushd vuejs
export TSGOLINT_BENCHMARK_PROJECT=vuejs
hyperfine --warmup 1 --ignore-failure \
  --command-name 'eslint' 'node ./node_modules/eslint/bin/eslint.js --no-inline-config --config ./eslint.config.mjs .' \
  --command-name 'tsgolint' '../../tsgolint --tsconfig ./tsconfig.json'
popd
