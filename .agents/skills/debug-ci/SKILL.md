---
name: debug-ci
description: Reproduce Linux CI failures locally using Docker when the same tests pass on the host, especially Go platform differences and VS Code extension tests requiring xvfb. Read the workflow and setup actions for the revision being tested; Docker does not reproduce Windows or macOS runners.
---

# Debug CI

Reproduce the selected failing Linux checks in an isolated checkout with the relevant CI toolchain. Follow [AGENTS.md](../../../AGENTS.md) for the task branch, verification scope, test organization and result reuse. Keep the existing task branch and progress record when continuing a diagnosis.

## Identify the failure and select the scope

```bash
gh pr checks <PR_NUMBER>
gh api repos/web-infra-dev/rslint/actions/runs/<RUN_ID>/jobs \
  --jq '.jobs[] | select(.conclusion == "failure") | {name, id}'
gh api repos/web-infra-dev/rslint/actions/jobs/<JOB_ID>/logs
```

Record the failing commit, job, runner architecture and failed test or check. Read `.github/workflows/ci.yml` and its referenced setup actions at the failing revision to reproduce the original failure. To validate the current fix, use the intended task revision plus its current changes and the CI configuration applicable to that target; record any environment differences from the failed run.

Trace the affected packages and consumers, then state the selected commands before running them. A failed CI job does not authorize all of that job's tests locally. If the user or reviewer explicitly requests a whole job or full suite, read its commands from the workflow at the selected revision rather than maintaining a second full-job script here. Windows and macOS failures need the corresponding platform; report that gap instead of claiming Docker parity.

## Isolate the reproduction

Use a disposable independent clone outside the working checkout. This keeps its `.git` directory usable inside Docker and keeps Linux dependencies, Go binaries and VS Code downloads out of the user's current workspace.

```bash
git clone --no-hardlinks --no-checkout <repository-url-or-local-path> <repro-dir>
```

Before checkout, ensure the exact tested commit is present. A normal clone may omit PR merge/fork refs or unpublished local commits; fetch the corresponding CI/PR ref, or fetch from a local checkout that holds the commit. Verify the resolved SHA matches the recorded target: a current PR ref may have advanced since the failed run.

```bash
git -C <repro-dir> checkout --detach <tested-commit>
```

Read-only reproduction may stay detached. Before applying a patch or editing in the clone, select a task branch at the tested commit and verify `git branch --show-current` there, following AGENTS.md.

When validating an uncommitted fix, apply the intended staged and unstaged changes and copy any required untracked inputs into this clone before testing. A clone of `HEAD` alone does not contain the fix. Review the isolated diff against the intended input; exclude unrelated edits, host `node_modules`, generated binaries and caches.

After applying the selected patch, initialize the compiler submodule when the selected checks need it. Verify its revision and include any intended changes inside the submodule as well:

```bash
git -C <repro-dir> submodule sync -- typescript-go
git -C <repro-dir> submodule update --init --depth 1
```

Ensure the chosen base ref and enough history for the lint merge-base exist in the isolated clone. A local clone's `origin` points to the source checkout, so verify the base commit rather than assuming its `origin/main` matches the source checkout's remote-tracking ref.

Reuse the same isolated clone, image and container while their relevant inputs remain unchanged. Install dependencies or rebuild outputs only when the selected execution path needs them and they are missing or stale. Do not delete the current workspace's `node_modules` or `.vscode-test` directories to obtain a clean Linux environment.

## Match the Linux environment

Reuse a matching image if available. Otherwise prepare a temporary Docker build context outside the repository, using the selected revision's sources below. Install only the tools needed by the selected checks.

| Input                                              | Source                                                                                                                          |
| -------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| Linux base and architecture                        | Failing job's runner and setup logs; select the matching Docker platform                                                        |
| Go version                                         | Selected job's `go-version` or matrix in `.github/workflows/ci.yml`                                                             |
| Node version and pnpm setup                        | Selected job and `.github/actions/setup-node/action.yml`; pnpm version also follows root `package.json`                         |
| golangci-lint version                              | `golangci-lint-action` version input in the selected workflow                                                                   |
| Rust toolchain, when the native parser is required | `.github/actions/setup-rust/action.yml` and the selected job's build prerequisites                                              |
| xvfb and GUI packages                              | Selected workflow's install step; add missing Electron runtime libraries when the bare image lacks runner-provided dependencies |

Pass the versions read from those sources as Docker build arguments, with no fallback version constants in the Dockerfile. For example, declare `ARG GO_VERSION`, `ARG NODE_VERSION` and `ARG GOLANGCI_LINT_VERSION`, use those arguments in the corresponding installations, and pass their resolved values when building. Keep the base image and installation architecture consistent with `--platform`; for Go downloads, Docker's `TARGETARCH` can select the archive architecture.

```bash
docker build --platform <ci-linux-platform> \
  --build-arg GO_VERSION=<selected-go-version> \
  --build-arg NODE_VERSION=<selected-node-version> \
  --build-arg GOLANGCI_LINT_VERSION=<selected-lint-version> \
  -t rslint-ci-test <temporary-build-context>
```

Omit unneeded tools and arguments, or add the selected Rust toolchain when the execution path builds the native parser. Do not copy fixed versions or the entire CI job into this skill. Record any runner services, image packages or architecture details that cannot be matched. On Apple Silicon, matching an x64 Linux runner may require emulation and increase runtime; do not treat an ARM run as equivalent evidence.

## Prepare and run the selected checks

Mount only the isolated clone. Keep the container alive during the diagnosis so its tool caches can be reused:

```bash
docker run --rm -it --platform <ci-linux-platform> \
  --mount "type=bind,src=<absolute-repro-dir>,dst=/workspace" \
  --workdir /workspace \
  rslint-ci-test bash
```

Run commands inside `/workspace`. For JS-based checks, use `pnpm install --frozen-lockfile` when dependencies need preparation. Read the selected workspace's scripts and the applicable CI build prerequisites before building: compiled JS artifacts, the native parser and generated rule schemas may be required by the selected integration chain. Build those dependencies with their existing workspace commands; do not default to a root build for every diagnosis. Before JS integration tests exercise changed Go code, run `pnpm --filter @rslint/core build:bin`.

The following are scoped examples, not a checklist. Replace the package or file with the selected target and include affected consumers when needed.

| Check                       | Existing scoped command                                                                        |
| --------------------------- | ---------------------------------------------------------------------------------------------- |
| Go rule package             | `go test -count=1 ./internal/rules/max_params`                                                 |
| Rstest integration file     | `pnpm --dir packages/rslint-test-tools exec rs test run tests/eslint/rules/max-params.test.ts` |
| VS Code extension workspace | `xvfb-run -a pnpm --filter rslint test`                                                        |
| Go lint                     | `golangci-lint run --new-from-merge-base=origin/main ./internal/rules/max_params`              |
| Go formatting check         | `golangci-lint fmt --diff ./internal/rules/max_params`                                         |
| JS/TS/docs formatting check | `pnpm exec rs fmt --check <changed-files>`                                                     |

The VS Code package uses its own `__tests__/runTest.ts` and Mocha runner, not Rstest. Its existing `test` script compiles and runs all extension suites and currently exposes no test-file or suite selector. The workspace command above is the smallest supported entry point; report that scope rather than inventing a filter or sending extension tests to `rs test`.

Run Go lint once with the branch-diff filter and selected package directories; do not follow it with the root `lint:go` script. Substitute the actual base ref when it is not `origin/main`. Preserve the formatter's existing behavior: the scoped `golangci-lint fmt --diff` command corresponds to root `pnpm run format:go --diff` without adding formatter options such as `--enable gofmt`.

## Record the result

Keep the tested revision and patch, container tool versions and architecture, selected commands, outcomes and remaining platform gaps in the task's existing progress record. Reuse passing results while their relevant inputs remain unchanged; rebuilding an image or reaching the commit step is not itself a reason to repeat unrelated checks.

Bring only intentional fixes back from the isolated clone and review them on the existing task branch. Dispose of task-owned reproduction resources when they are no longer needed. No host dependency restoration should be necessary because all Linux writes stayed in the isolated clone; `pnpm install` is not a way to restore a Go binary overwritten by a Linux build.
