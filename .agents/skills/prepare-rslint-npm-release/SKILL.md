---
name: prepare-rslint-npm-release
description: Prepare a stable release PR for the unified rslint npm packages by updating package versions and rule/TypeScript release metadata. Use for release preparation or version bump PRs, not for publishing npm packages, canary releases, VS Code Marketplace releases, or the independent tsgo/Rust release line.
---

# Prepare an Rslint npm Release PR

Prepare a reviewable version and metadata PR, following the shape of
[PR #2080](https://github.com/web-infra-dev/rslint/pull/2080). Use the current
`pnpm sync:version-info` command introduced in #2078; #2080 used the older
`sync:rule-releases` command and data filename.

## Establish the release state

Follow the repository's [AGENTS.md](../../../AGENTS.md) for branches and local
verification. Read these current sources before choosing commands or file scope:

- [scripts/version.mjs](../../../scripts/version.mjs) and the root
  [package.json](../../../package.json): version calculation and package selection.
- [scripts/sync-version-info/](../../../scripts/sync-version-info/): rule history
  and the pinned TypeScript commit/release mapping.
- [Sync release information](../../../CONTRIBUTING.md#sync-release-information):
  the manual preparation sequence.

Resolve the intended stable version or `major`, `minor`, or `patch` bump from the
request. Ask for the target if it cannot be inferred. The version script bumps
from the highest base version in its selected manifests, so inspect those
versions before predicting the result.

Fetch the main branch and release tags. Determine the previous stable `vX.Y.Z`
tag, excluding prerelease tags, and check for an existing branch or PR for the
requested release. Start new preparation from current `origin/main` on a branch
such as `chore/release-<version>`; resume an existing release branch instead of
bumping again. If the target is already released, report that state rather than
silently choosing another version.

## Update versions and metadata

Run commands from the repository root. For an agreed patch release:

```bash
pnpm run version patch
pnpm sync:version-info
```

Use the requested bump type in place of `patch`. If the intended version bump is
already present, skip the first command and review or refresh the metadata.
`version` is not idempotent: running it again advances the version again. A failed
metadata lookup is a reason to retry the sync after resolving the failure, not
to repeat the bump.

The version script owns the unified package selection: the root manifest,
`packages/*/package.json` except `packages/tsgo/package.json`, and
`npm/rslint/*/package.json`. This includes private workspace packages and the
VS Code extension manifest; updating that manifest is part of version alignment,
not a Marketplace release. Derive the current selection from the script rather
than hardcoding #2080's package count. Preserve `workspace:` dependency ranges.
The independent tsgo packages and Rust crate versions are outside this workflow.

The sync writes `website/releases.json` for the new stable version. It reads the
indexed `typescript-go` gitlink and queries upstream TypeScript tags. Record the
compiler already pinned for this release; preparing a release does not imply a
compiler upgrade. Stage any separately authorized submodule revision change
before syncing. The sync can read the pin without initializing the submodule.

An existing release tag makes the default sync skip that published version. An
unexpected skip during new release preparation is not successful generation:
check the package version and fetched tags. Use incremental sync for normal
preparation; `full` rebuilds historical rule records and belongs to a separate
history repair request.

## Review the generated diff

Check these invariants before committing:

- Every selected manifest has the intended version. Ordinary version changes
  affect only its `version` field; investigate unrelated dependency, script,
  formatting, or lockfile changes.
- The new release appears once in `website/releases.json`. Earlier records and
  compiler bindings are unchanged, and no `main` entry or compiler backfill is
  introduced.
- New rule IDs are unique and sorted, refer to registered rules, and were absent
  from the previous stable release and earlier recorded releases. Cross-check
  their implementations and the relevant core/plugin `all.go` aggregators;
  a helper or fixture directory is not a rule. A release with no new rules is
  valid.
- `typescript.commit` equals the indexed gitlink. `typescript.releaseVersion` is
  an exact upstream tag match, including peeled annotated tags, or `null` when
  no tag matches. Do not substitute the latest TypeScript release or a version
  reported by upstream source files.
- A second `pnpm sync:version-info` run leaves the generated file unchanged.

For version and metadata-only changes, follow #2080's focused verification:
run `pnpm run check-spell` with every changed text path explicitly, run
`pnpm run format:check`, and run `git diff --check`. Apply AGENTS.md's affected
package rules if source or build behavior also changes. The generic build/test
suggestions printed by `version.mjs` are not a request for full local suites.

## Deliver the preparation PR

Commit the version manifests and release metadata together, with a message and
PR title such as `chore: release npm <version>`. Push the release branch and open
a PR targeting `main`, or update the existing release PR. Include the version
transition, actual manifest count, new rule count, TypeScript binding, and the
verification commands/results. Keep temporary verification artifacts outside
the commit.

This workflow ends at the preparation PR. Actual publication is a separate
manual [release.yml](../../../.github/workflows/release.yml) workflow; merging a
version PR does not publish npm packages. Its npm-only selection is
`to_release=npm`. Draft release notes, when requested, use the separate
[create-draft-release-notes skill](../create-draft-release-notes/SKILL.md).
