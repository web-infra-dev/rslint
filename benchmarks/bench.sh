#!/usr/bin/env -S bash -euxo pipefail

pushd vscode
hyperfine --warmup 1 --ignore-failure \
  --command-name 'eslint' 'NODE_OPTIONS="--max-old-space-size=16384" TSGOLINT_BENCHMARK_PROJECT=vscode ./node_modules/eslint/bin/eslint.js --no-inline-config --config ./eslint.config.mjs ./src' \
  --command-name 'tsgolint' 'cd src && ../../../tsgolint --tsconfig ./tsconfig.json'
popd

pushd typescript
hyperfine --warmup 1 --ignore-failure \
  --command-name 'eslint' 'NODE_OPTIONS="--max-old-space-size=16384" TSGOLINT_BENCHMARK_PROJECT=typescript ./node_modules/eslint/bin/eslint.js --no-inline-config --config ./eslint.config.mjs ./src' \
  --command-name 'tsgolint' 'cd src && ../../../tsgolint --tsconfig ./tsconfig-eslint.json'
popd

pushd typeorm
hyperfine --warmup 1 --ignore-failure \
  --command-name 'eslint' 'NODE_OPTIONS="--max-old-space-size=16384" TSGOLINT_BENCHMARK_PROJECT=typeorm ./node_modules/eslint/bin/eslint.js --no-inline-config --config ./eslint.config.mjs .' \
  --command-name 'tsgolint' '../../tsgolint --tsconfig ./tsconfig.json'
popd

pushd vuejs
hyperfine --warmup 1 --ignore-failure \
  --command-name 'eslint' 'NODE_OPTIONS="--max-old-space-size=16384" TSGOLINT_BENCHMARK_PROJECT=vuejs ./node_modules/eslint/bin/eslint.js --no-inline-config --config ./eslint.config.mjs .' \
  --command-name 'tsgolint' '../../tsgolint --tsconfig ./tsconfig.json'
popd
