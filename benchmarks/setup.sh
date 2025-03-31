#!/usr/bin/env -S bash -euxo pipefail

for proj in {vscode,typescript,typeorm}; do
  pushd $proj
  cp ../eslint.config.mjs ./eslint.config.mjs
  npm install --ignore-scripts -D eslint@latest typescript-eslint@latest
  popd
done


