---
name: debug-ci
description: Reproduce and debug CI failures locally using Docker. Use when a CI job fails but tests pass locally, especially for Linux-specific issues (VSCode extension tests with xvfb, Go tests on Linux). Builds a container matching the CI environment defined in .github/workflows/ci.yml.
---

# Debug CI

Reproduce CI test failures locally using a Docker container that mirrors the CI environment.

## When to Use

- CI job fails but the same tests pass locally on macOS/Windows
- VSCode extension tests fail (require xvfb + Linux GUI libraries)
- Platform-specific issues (Linux vs macOS vs Windows)

## CI Jobs Overview

Reference: `.github/workflows/ci.yml`

| CI Job                | Runner                   | Key Tools               | Docker Reproducible? |
| --------------------- | ------------------------ | ----------------------- | -------------------- |
| **Test Go**           | Ubuntu / Windows / macOS | Go, Node                | Yes (Linux)          |
| **Lint&Check**        | Ubuntu                   | Go, Node, golangci-lint | Yes                  |
| **Test npm packages** | Ubuntu / Windows         | Go, Node, xvfb          | Yes (Linux)          |
| **Test WASM**         | Ubuntu                   | Go, Node                | Yes                  |
| **Test Rust**         | macOS                    | Rust, Go, Node          | No (macOS only)      |
| **Build Website**     | Ubuntu                   | Node                    | Yes                  |

## Workflow

### Step 1: Identify the Failing Job

```bash
gh pr checks <PR_NUMBER>
```

Fetch detailed failure logs:

```bash
# Find failed job IDs
gh api repos/web-infra-dev/rslint/actions/runs/<RUN_ID>/jobs \
  --jq '.jobs[] | select(.conclusion == "failure") | {name, id}'

# Get logs for a specific job
gh api repos/web-infra-dev/rslint/actions/jobs/<JOB_ID>/logs
```

### Step 2: Build Docker Image

Check current tool versions from CI config before building:

| Tool            | Version Source                                                                                  |
| --------------- | ----------------------------------------------------------------------------------------------- |
| Go              | `.github/workflows/ci.yml` → `go-version` matrix (currently `1.25.0`)                           |
| Node            | `.github/actions/setup-node/action.yml` → `node-version` (currently `24`)                       |
| xvfb + GUI deps | `.github/workflows/ci.yml` → `test-node` job → "Install xvfb and dependencies" step             |
| golangci-lint   | `.github/workflows/ci.yml` → `lint` job → `golangci-lint-action` `version` (currently `v2.4.0`) |

Write a Dockerfile. The apt packages for xvfb **must match** the CI step in `test-node`:

```
# From CI: sudo apt install -y libasound2 libgbm1 libgtk-3-0 libnss3 xvfb
```

Additional GUI dependencies (libxss1, libatk-bridge2.0-0, etc.) are needed for VSCode Electron to run headlessly — they are implicit in the CI runner image but not in bare ubuntu:22.04.

Complete Dockerfile:

```dockerfile
FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive
ENV GOMAXPROCS=8

# Base tools
RUN apt-get update && apt-get install -y \
    curl wget git build-essential \
    && rm -rf /var/lib/apt/lists/*

# xvfb and GUI deps (from ci.yml test-node job + Electron implicit deps)
RUN apt-get update && apt-get install -y \
    libasound2 libgbm1 libgtk-3-0 libnss3 xvfb \
    libxss1 libatk-bridge2.0-0 libdrm2 libxcomposite1 libxdamage1 libxrandr2 \
    libpango-1.0-0 libcairo2 libcups2 libdbus-1-3 libexpat1 libfontconfig1 \
    libgcc1 libglib2.0-0 libnspr4 libpangocairo-1.0-0 libstdc++6 libx11-6 \
    libx11-xcb1 libxcb1 libxcursor1 libxfixes3 libxi6 libxrender1 libxtst6 \
    ca-certificates fonts-liberation lsb-release xdg-utils \
    && rm -rf /var/lib/apt/lists/*

# Go (version from ci.yml go-version matrix)
RUN curl -fsSL https://go.dev/dl/go1.25.0.linux-amd64.tar.gz | tar -C /usr/local -xz
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/root/go"
ENV PATH="${GOPATH}/bin:${PATH}"

# Node (version from .github/actions/setup-node/action.yml)
RUN curl -fsSL https://deb.nodesource.com/setup_24.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# pnpm (from .github/actions/setup-node/action.yml: corepack enable)
RUN corepack enable

# golangci-lint (version from ci.yml lint job → golangci-lint-action)
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
    | sh -s -- -b /usr/local/bin v2.4.0

WORKDIR /workspace
```

Build:

```bash
docker build -t rslint-ci-test <dockerfile-dir>
```

### Step 3: Write Test Script

Choose the test script based on which CI job failed:

**Test npm packages (Linux)** — the most common scenario requiring Docker reproduction:

```bash
#!/bin/bash
set -e
cd /workspace

rm -rf node_modules packages/*/node_modules
find . -name '.vscode-test' -type d -exec rm -rf {} + 2>/dev/null || true

pnpm install --frozen-lockfile    # .github/actions/setup-node
pnpm format:check                 # ci.yml test-node "Format" (Linux only)
pnpm run build                    # ci.yml test-node "Build" step
pnpm run lint --format github     # ci.yml test-node "Dogfooding" (Linux only)
pnpm typecheck                    # ci.yml test-node "TypeCheck" (Linux only)
xvfb-run -a pnpm run test         # ci.yml test-node "Test on Linux" step
```

**Lint&Check**:

```bash
#!/bin/bash
set -e
cd /workspace
rm -rf node_modules packages/*/node_modules
pnpm install --frozen-lockfile
golangci-lint run --timeout=5m ./cmd/... ./internal/...   # ci.yml lint "golangci-lint"
npm run lint:go                                           # ci.yml lint "go vet"
npm run format:go                                         # ci.yml lint "go fmt"
pnpm check-spell                                         # ci.yml lint "Check Spell"
```

**Test Go (Linux)**:

```bash
#!/bin/bash
set -e
cd /workspace
go test -parallel 8 ./internal/...   # ci.yml test-go "Unit Test"
```

**Test WASM**:

```bash
#!/bin/bash
set -e
cd /workspace
rm -rf node_modules packages/*/node_modules
pnpm install --frozen-lockfile
pnpm --filter '@rslint/core' build:js
pnpm --filter '@rslint/wasm' build   # ci.yml test-wasm "Build"
```

**Build Website**:

```bash
#!/bin/bash
set -e
cd /workspace
rm -rf node_modules packages/*/node_modules
pnpm install --frozen-lockfile
pnpm run build:website               # ci.yml website "Build"
```

### Step 4: Run in Docker

```bash
docker run --rm \
  -v <project-root>:/workspace \
  -v <test-script>:/run-test.sh:ro \
  rslint-ci-test bash /run-test.sh
```

### Step 5: Restore Host Environment

Docker overwrites host node_modules and Go binaries with Linux versions. Always restore after testing:

```bash
pnpm install
```

## Caveats

- **typescript-go submodule ~1.2GB**: Do not `cp` / `rsync` the project inside the container — use a bind mount directly
- **Apple Silicon**: Docker runs via x86_64 emulation, Go compilation will be 5-10x slower
- **node_modules conflict**: The container installs Linux-native dependencies, overwriting macOS/Windows host dependencies. Always run `pnpm install` after testing to restore
- **Windows CI**: Docker cannot simulate a Windows runner. Windows-specific failures require a Windows machine to debug
- **Test Rust**: CI runs on macOS, not suitable for Linux Docker reproduction
