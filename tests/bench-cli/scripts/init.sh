#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd -- "${SCRIPT_DIR}/../../.." && pwd)"
TMP_DIR="${TMPDIR:-/tmp}/rslint-bench"
VSCODE_DIR="${TMP_DIR}/vscode"
VSCODE_REPO_URL="https://github.com/microsoft/vscode.git"
VSCODE_COMMIT="41dd792b5e652393e7787322889ed5fdc58bd75b"
CONFIG_TEMPLATE="${REPO_ROOT}/tests/bench-cli/fixtures/vscode-rslint.config.mjs"

echo "============================================"
echo "Initializing VS Code benchmark fixture"
echo "============================================"
echo ""

mkdir -p "${TMP_DIR}"

if [ -d "${VSCODE_DIR}" ]; then
  echo "VS Code directory already exists, skipping clone..."
elif [ ! -d "${VSCODE_DIR}/.git" ]; then
  echo "Cloning VS Code repository..."
  git clone --depth=1 "${VSCODE_REPO_URL}" "${VSCODE_DIR}"
else
  echo "VS Code repository already exists, reusing..."
fi

if [ -d "${VSCODE_DIR}/.git" ]; then
  CURRENT_COMMIT="$(git -C "${VSCODE_DIR}" rev-parse HEAD 2>/dev/null || true)"
else
  CURRENT_COMMIT=""
fi

if [ -d "${VSCODE_DIR}/.git" ] && [ "${CURRENT_COMMIT}" != "${VSCODE_COMMIT}" ]; then
  echo "Checking out VS Code commit ${VSCODE_COMMIT}..."
  git -C "${VSCODE_DIR}" fetch --depth=1 --no-tags origin "${VSCODE_COMMIT}"
  git -C "${VSCODE_DIR}" checkout --detach FETCH_HEAD
fi

# Place benchmark config at the default discovery location so CLI can lint
# without passing an explicit --config flag.
cp "${CONFIG_TEMPLATE}" "${VSCODE_DIR}/rslint.config.mjs"
rm -f "${VSCODE_DIR}/.rslint.bench.config.mjs"

echo ""
echo "VS Code benchmark fixture is ready:"
echo "  ${VSCODE_DIR}"
