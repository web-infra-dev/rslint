# rslint TypeScript ESLint Rule Porting Gap Report

<!-- cspell:ignore LASTEXITCODE libasound libgbm libnss -->

Generated on: 2026-02-09

## Hard constraint

- DO NOT edit, update, sync, or otherwise alter the `typescript-go` git submodule at all.
- This applies to manual edits, `git submodule` commands that change refs, dependency updates, and generated changes.

## Scope and methodology

This report is intentionally scoped to `@typescript-eslint` rules only.

Upstream catalog source:

- `@typescript-eslint/eslint-plugin` from `typescript-eslint/typescript-eslint` tag `v8.48.1` (`packages/eslint-plugin/src/rules/*`)

rslint implemented rules source:

- `internal/config/config.go` (`GlobalRuleRegistry.Register(...)` calls)

## Rules excluded from this report (already in open PRs)

- `@typescript-eslint/member-ordering` - PR #468
- `@typescript-eslint/init-declarations` - PR #467
- `@typescript-eslint/explicit-member-accessibility` - PR #469
- `@typescript-eslint/class-methods-use-this` - PR #470
- `@typescript-eslint/explicit-function-return-type` - PR #465
- `@typescript-eslint/ban-tslint-comment` - PR #464

## Coverage summary (after exclusions)

- @typescript-eslint: implemented 79 / 132, missing 49

## Local validation commands from CI jobs

Source of truth:

- `.github/workflows/ci.yml`
- `.github/actions/setup-node/action.yml`
- `.github/actions/setup-go/action.yml`

Environment used in CI:

- Node.js `24`
- Go `1.25.0`
- Rust stable (for `test-rust` job)

Repository setup (run before job commands):

```bash
git submodule update --init --recursive
corepack enable
pnpm install --frozen-lockfile
```

Run commands that mirror `CI` workflow jobs:

1. `test-go` job

```bash
# non-Windows path from CI
go test -parallel 8 ./internal/...
```

Windows CI equivalent (batch logic):

```powershell
$packages = go list ./internal/... | Where-Object { $_ -ne "" }
$batchSize = 15
$batch = @()
foreach ($pkg in $packages) {
  $batch += $pkg
  if ($batch.Count -ge $batchSize) {
    go test -parallel 4 $batch
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    go clean -cache
    $batch = @()
  }
}
if ($batch.Count -gt 0) {
  go test -parallel 4 $batch
  if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}
```

2. `lint` job

```bash
# from golangci-lint action args
golangci-lint run --timeout=5m ./cmd/... ./internal/...
# CI command label is "go vet" but command actually used is:
npm run lint:go
# CI command label is "go fmt" and command actually used is:
npm run format:go
pnpm check-spell
```

3. `test-node` job

```bash
# Linux-only in CI
pnpm format:check
pnpm run build
pnpm run lint --format github
pnpm typecheck
# Linux test path
xvfb-run -a pnpm run test

# non-Linux test path
pnpm run test
```

Linux package prerequisites used by CI for tests:

```bash
sudo apt update
sudo apt install -y libasound2 libgbm1 libgtk-3-0 libnss3 xvfb
```

4. `test-wasm` job

```bash
pnpm --filter '@rslint/core' build:js
pnpm --filter '@rslint/wasm' build
```

5. `test-rust` job

```bash
cargo fmt --all -- --check
cargo clippy --all-targets --all-features -- -D warnings
pnpm --filter '@rslint/tsgo' build
cargo test --verbose
```

6. `website` job

```bash
pnpm run build:website
```

## How to register rule

1. Implement the rule under:

- `internal/plugins/typescript/rules/<rule_name>/<rule_name>.go`

2. Define a `rule.Rule` with:

- `Name`: canonical id (for example, `@typescript-eslint/no-explicit-any`)
- `Run`: callback returning listeners for AST node kinds

3. Register it in `internal/config/config.go`:

- Add a `GlobalRuleRegistry.Register("@typescript-eslint/<rule>", <RuleVar>)` entry in `registerAllTypeScriptEslintPluginRules`

4. Ensure config enabling works:

- `GetAllRulesForPlugin` in `internal/config/config.go`
- `RuleRegistry.GetEnabledRules` in `internal/config/rule_registry.go`

5. Add tests:

- Go tests alongside implementation (`*_test.go`)
- Optional upstream parity fixtures in `packages/rslint-test-tools/tests/typescript-eslint/...`

## Missing @typescript-eslint rules to port

### @typescript-eslint (49)

- `@typescript-eslint/explicit-module-boundary-types` - docs: https://typescript-eslint.io/rules/explicit-module-boundary-types | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/explicit-module-boundary-types.ts
- `@typescript-eslint/max-params` - docs: https://typescript-eslint.io/rules/max-params | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/max-params.ts
- `@typescript-eslint/method-signature-style` - docs: https://typescript-eslint.io/rules/method-signature-style | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/method-signature-style.ts
- `@typescript-eslint/naming-convention` - docs: https://typescript-eslint.io/rules/naming-convention | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/naming-convention.ts
- `@typescript-eslint/no-confusing-non-null-assertion` - docs: https://typescript-eslint.io/rules/no-confusing-non-null-assertion | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-confusing-non-null-assertion.ts
- `@typescript-eslint/no-deprecated` - docs: https://typescript-eslint.io/rules/no-deprecated | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-deprecated.ts
- `@typescript-eslint/no-dupe-class-members` - docs: https://typescript-eslint.io/rules/no-dupe-class-members | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-dupe-class-members.ts
- `@typescript-eslint/no-dynamic-delete` - docs: https://typescript-eslint.io/rules/no-dynamic-delete | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-dynamic-delete.ts
- `@typescript-eslint/no-empty-object-type` - docs: https://typescript-eslint.io/rules/no-empty-object-type | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-empty-object-type.ts
- `@typescript-eslint/no-import-type-side-effects` - docs: https://typescript-eslint.io/rules/no-import-type-side-effects | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-import-type-side-effects.ts
- `@typescript-eslint/no-invalid-this` - docs: https://typescript-eslint.io/rules/no-invalid-this | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-invalid-this.ts
- `@typescript-eslint/no-loop-func` - docs: https://typescript-eslint.io/rules/no-loop-func | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-loop-func.ts
- `@typescript-eslint/no-loss-of-precision` - docs: https://typescript-eslint.io/rules/no-loss-of-precision | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-loss-of-precision.ts
- `@typescript-eslint/no-magic-numbers` - docs: https://typescript-eslint.io/rules/no-magic-numbers | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-magic-numbers.ts
- `@typescript-eslint/no-redeclare` - docs: https://typescript-eslint.io/rules/no-redeclare | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-redeclare.ts
- `@typescript-eslint/no-restricted-imports` - docs: https://typescript-eslint.io/rules/no-restricted-imports | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-restricted-imports.ts
- `@typescript-eslint/no-restricted-types` - docs: https://typescript-eslint.io/rules/no-restricted-types | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-restricted-types.ts
- `@typescript-eslint/no-shadow` - docs: https://typescript-eslint.io/rules/no-shadow | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-shadow.ts
- `@typescript-eslint/no-type-alias` - docs: https://typescript-eslint.io/rules/no-type-alias | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-type-alias.ts
- `@typescript-eslint/no-unnecessary-condition` - docs: https://typescript-eslint.io/rules/no-unnecessary-condition | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-unnecessary-condition.ts
- `@typescript-eslint/no-unnecessary-parameter-property-assignment` - docs: https://typescript-eslint.io/rules/no-unnecessary-parameter-property-assignment | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-unnecessary-parameter-property-assignment.ts
- `@typescript-eslint/no-unnecessary-qualifier` - docs: https://typescript-eslint.io/rules/no-unnecessary-qualifier | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-unnecessary-qualifier.ts
- `@typescript-eslint/no-unnecessary-type-constraint` - docs: https://typescript-eslint.io/rules/no-unnecessary-type-constraint | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-unnecessary-type-constraint.ts
- `@typescript-eslint/no-unnecessary-type-conversion` - docs: https://typescript-eslint.io/rules/no-unnecessary-type-conversion | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-unnecessary-type-conversion.ts
- `@typescript-eslint/no-unnecessary-type-parameters` - docs: https://typescript-eslint.io/rules/no-unnecessary-type-parameters | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-unnecessary-type-parameters.ts
- `@typescript-eslint/no-unsafe-declaration-merging` - docs: https://typescript-eslint.io/rules/no-unsafe-declaration-merging | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-unsafe-declaration-merging.ts
- `@typescript-eslint/no-unsafe-function-type` - docs: https://typescript-eslint.io/rules/no-unsafe-function-type | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-unsafe-function-type.ts
- `@typescript-eslint/no-unused-expressions` - docs: https://typescript-eslint.io/rules/no-unused-expressions | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-unused-expressions.ts
- `@typescript-eslint/no-unused-private-class-members` - docs: https://typescript-eslint.io/rules/no-unused-private-class-members | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-unused-private-class-members.ts
- `@typescript-eslint/no-use-before-define` - docs: https://typescript-eslint.io/rules/no-use-before-define | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-use-before-define.ts
- `@typescript-eslint/no-useless-constructor` - docs: https://typescript-eslint.io/rules/no-useless-constructor | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-useless-constructor.ts
- `@typescript-eslint/no-wrapper-object-types` - docs: https://typescript-eslint.io/rules/no-wrapper-object-types | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/no-wrapper-object-types.ts
- `@typescript-eslint/parameter-properties` - docs: https://typescript-eslint.io/rules/parameter-properties | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/parameter-properties.ts
- `@typescript-eslint/prefer-destructuring` - docs: https://typescript-eslint.io/rules/prefer-destructuring | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/prefer-destructuring.ts
- `@typescript-eslint/prefer-enum-initializers` - docs: https://typescript-eslint.io/rules/prefer-enum-initializers | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/prefer-enum-initializers.ts
- `@typescript-eslint/prefer-find` - docs: https://typescript-eslint.io/rules/prefer-find | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/prefer-find.ts
- `@typescript-eslint/prefer-for-of` - docs: https://typescript-eslint.io/rules/prefer-for-of | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/prefer-for-of.ts
- `@typescript-eslint/prefer-function-type` - docs: https://typescript-eslint.io/rules/prefer-function-type | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/prefer-function-type.ts
- `@typescript-eslint/prefer-includes` - docs: https://typescript-eslint.io/rules/prefer-includes | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/prefer-includes.ts
- `@typescript-eslint/prefer-literal-enum-member` - docs: https://typescript-eslint.io/rules/prefer-literal-enum-member | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/prefer-literal-enum-member.ts
- `@typescript-eslint/prefer-namespace-keyword` - docs: https://typescript-eslint.io/rules/prefer-namespace-keyword | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/prefer-namespace-keyword.ts
- `@typescript-eslint/prefer-nullish-coalescing` - docs: https://typescript-eslint.io/rules/prefer-nullish-coalescing | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/prefer-nullish-coalescing.ts
- `@typescript-eslint/prefer-optional-chain` - docs: https://typescript-eslint.io/rules/prefer-optional-chain | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/prefer-optional-chain.ts
- `@typescript-eslint/prefer-regexp-exec` - docs: https://typescript-eslint.io/rules/prefer-regexp-exec | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/prefer-regexp-exec.ts
- `@typescript-eslint/prefer-string-starts-ends-with` - docs: https://typescript-eslint.io/rules/prefer-string-starts-ends-with | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/prefer-string-starts-ends-with.ts
- `@typescript-eslint/prefer-ts-expect-error` - docs: https://typescript-eslint.io/rules/prefer-ts-expect-error | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/prefer-ts-expect-error.ts
- `@typescript-eslint/sort-type-constituents` - docs: https://typescript-eslint.io/rules/sort-type-constituents | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/sort-type-constituents.ts
- `@typescript-eslint/strict-boolean-expressions` - docs: https://typescript-eslint.io/rules/strict-boolean-expressions | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/strict-boolean-expressions.ts
- `@typescript-eslint/typedef` - docs: https://typescript-eslint.io/rules/typedef | source: https://github.com/typescript-eslint/typescript-eslint/blob/v8.48.1/packages/eslint-plugin/src/rules/typedef.ts

## Notes

- This report only tracks `@typescript-eslint` porting gaps by request.
- Excluded rules are intentionally omitted from the missing list because they already have open PRs.
- This report intentionally preserves the submodule constraint: do not modify `typescript-go`.
