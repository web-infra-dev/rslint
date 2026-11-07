# Report unused disable directives

## Status and objective

Implementation was authorized after branch setup and design confirmation.
Work is on `feat/report-unused-disable-directives`, based on main
`c711b7328a6c8892c70def785d5dd3e846b38d95`. Functional verification has passed; performance benchmarks and the final PR
remain pending the user's workload instructions. Changes are not committed yet.

Implement unused disable directive reporting in a new PR, using
[#1312](https://github.com/web-infra-dev/rslint/pull/1312) as requirements and test
material rather than simply rebasing its implementation. Go owns each file's
directive state for both native rules and Node plugin rules. Filter diagnostics
as they arrive, then report unused directives after all relevant file work ends.

Remove the old proposal's native-only and once-per-run configuration limitations.

## Test migration result

The pinned ESLint 10.14.0 import now executes 89 engine tests and 486
source-level tests with their original assertions. It produces 88 engine cases
(528 variants) and 691 distinct source states. Provenance, exclusions, regeneration
commands, and the adjacent-fix scheduling difference are documented in
`internal/rule/testdata/eslint-directives/README.md`.

Final migration checks passed:

- All three generators; repeated generation produced identical fixture hashes.
- `go test -p 1 ./internal/rule ./internal/linter -run 'Disable|Directive'`.
- Scoped formatting and `git diff --check`.

An additional consumer check covering config resolver, API, LSP, and CLI was
stopped during dependency compilation to respect the user's time limit. Those
consumers have not been reverified after the latest unknown-rule validation
change. Earlier integration results do not cover that latest change. Benchmarks
and the final PR remain pending; this records completion of the test migration.

## Scope

- Add `--report-unused-disable-directives` and
  `--report-unused-disable-directives-severity <off|warn|error>`.
- Support `linterOptions.reportUnusedDisableDirectives` in each file's effective
  configuration. Explicit CLI options override configuration.
- Cover native rules, third-party Node plugin rules, and mixed execution.
- Cover CLI, JS API, LSP, and automatic fixing.
- Support equivalent `eslint-` and `rslint-` prefixes for block disable, enable,
  disable-line, and disable-next-line directives.
- Report unused directives and supply fixes consistent with the selected ESLint
  semantic baseline.

Do not add unused arbitrary inline rule configuration detection or rewrite the
plugin executor as part of this task.

## Research baseline

The initial investigation read these upstream revisions:

| Project         | Revision                                   | Relevant implementation                                                                                                                                                                                                                                                             |
| --------------- | ------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ESLint          | `bdd4b1da21dca81d50933edef0a015216059d877` | [Directive processing](https://github.com/eslint/eslint/blob/bdd4b1da21dca81d50933edef0a015216059d877/lib/linter/apply-disable-directives.js), [tests](https://github.com/eslint/eslint/blob/bdd4b1da21dca81d50933edef0a015216059d877/tests/lib/linter/apply-disable-directives.js) |
| oxlint          | `5038ba653982ae807ef51bad4d2b2ae87ca57f76` | [Directive state](https://github.com/oxc-project/oxc/blob/5038ba653982ae807ef51bad4d2b2ae87ca57f76/crates/oxc_linter/src/disable_directives.rs), [runner](https://github.com/oxc-project/oxc/blob/5038ba653982ae807ef51bad4d2b2ae87ca57f76/crates/oxc_linter/src/lint_runner.rs)    |
| oxlint-tsgolint | `6b043c5f3dd19e3a162b7f0edf587a9ae83bfe81` | [Go diagnostic reporting](https://github.com/oxc-project/tsgolint/blob/6b043c5f3dd19e3a162b7f0edf587a9ae83bfe81/internal/linter/linter.go), [headless output](https://github.com/oxc-project/tsgolint/blob/6b043c5f3dd19e3a162b7f0edf587a9ae83bfe81/cmd/tsgolint/headless.go)       |

ESLint collects diagnostics before applying directives. It processes block and
line directives separately, retains suppression metadata, and tracks individual
rule entries within their parent comments. A matching directive is not
automatically considered used: attribution follows directive order and existing
suppression state.

oxlint centralizes directive state in Rust. Native rules, returned JS plugin
diagnostics, and returned Go diagnostics all use that state. Its tsgolint
integration sends rule diagnostics to Rust rather than independently reporting
unused directives in Go. This ownership model is useful, but oxlint's interval
matching and rule-name matching are not an ESLint-equivalent semantic baseline.

Before implementation, pin the ESLint version used for conformance and record
any intentional differences. Reuse existing project capabilities; this plan
does not authorize installing dependencies or benchmark tools.

## Design

### One directive state per file execution

Go owns one directive state for a file, source version, and execution round.

1. Resolve the file's effective configuration and prepare its directive state.
2. Native rules check that state when reporting diagnostics. Record actual
   suppression before constructing fixes or suggestions, preserving the current
   ability to skip those builders.
3. Node plugins return diagnostics without applying disable filtering. After
   position conversion, Go filters them through the same directive state.
4. Once every relevant native task and plugin batch finishes, generate unused
   directive diagnostics exactly once.

Node currently constructs diagnostics and requested fixes before filtering
directives. Moving filtering to Go therefore primarily adds transport and
temporary storage for suppressed diagnostics; it does not inherently introduce
new fix computation. Computation performed by plugin code before reporting is
outside the host's control.

Do not maintain two directive parsers and merge separate used-directive sets
across the protocol.

### Directive representation and semantics

Retain the original comment location, directive kind, parent comment identity,
individual rule names and ranges, wildcard status, explanation text boundaries,
usage state, and source version identity.

- Follow ESLint's block-versus-line precedence and attribution for overlapping
  or repeated directives. Do not mark every matching directive as used.
- Track multiple rule entries independently so a partially unused comment can
  be reported and fixed precisely.
- Determine unused enable directives from whether they undo a corresponding
  disable state, rather than whether they suppress a diagnostic.
- Distinguish configured-off rules from enabled rules skipped by execution
  policy. An incomplete observation must not be treated as zero usage. Match
  ESLint's treatment of filtered rules and wildcard directives where applicable.
- Preserve and test established rslint rule aliases. Do not implement arbitrary
  plugin-prefix stripping.
- Finalize eligible files even if they produce no diagnostics or execute no
  rules. Directive discovery cannot depend solely on the first diagnostic.
- Keep the normal no-directive path inexpensive and retain lazy parsing where
  compatible with finalization.

Write and test an explicit configuration behavior table before implementation:
defaults, severity normalization, boolean compatibility, per-file overrides,
CLI precedence, and interactions between the two CLI flags. ESLint configuration
booleans and historical API/CLI booleans must not be assumed equivalent.

### Completion, lifetime, and protocol

- Integrate with the shared pipeline's completion mechanism. A file may have
  multiple plugin batches; the first returned batch is not file completion.
- On cancellation, plugin failure, or incomplete execution, conservatively omit
  unused diagnostics and deletion fixes for that file while preserving execution
  error reporting.
- Rebuild directive state after each source-changing fix round.
- Reject stale LSP results; old requests must not mutate a newer source version.
- Do not retain a Program, checker, or full AST merely to await plugin results.
  Keep only directive data and necessary source references, and release them
  after completion.
- Preserve the existing position-conversion boundary. Filtering and displayed
  diagnostics must agree on source identity and starting position, including BOM,
  Unicode, and CRLF handling.
- Audit every plugin entry point, not only the CLI path.
- Treat the change from filtered to unfiltered plugin results as a protocol
  semantic change even if field shapes remain identical. Update the appropriate
  compatibility marker and reject incompatible peers explicitly.

Because the editor consumes plugin wire types, open a matching change or issue
in `rstackjs/rstack-editor` during the same PR cycle and link both directions.
Coordinate any required editor adaptation with release preparation.

### Diagnostics and fixes

Unused directives are linter diagnostics, not ordinary rule diagnostics. Do not
feed them back through the directive that caused their creation.

- Remove or safely replace an entirely unused comment without joining adjacent
  tokens.
- Remove only unused rule entries from partially used comments, preserving
  valid separators and explanations.
- Implement enable reporting and fixing according to the selected ESLint
  baseline.
- Never apply fixes using ranges from another source version.
- Use existing conflict resolution and fix-round machinery.
- Keep CLI fixes, API fix results, and LSP code actions consistent.

Preserve equivalent behavior for `eslint-` and `rslint-` prefixes. oxlint treats
removal of `eslint-` comments as suggestions because another ESLint invocation
may still need them; adopting that distinction would be a separate product
decision, not an implicit part of this task.

## Functional acceptance criteria

| Area                 | Required coverage                                                                                |
| -------------------- | ------------------------------------------------------------------------------------------------ |
| Directive forms      | Both prefixes, all four kinds, explanations, multiple rule entries                               |
| Usage attribution    | Repeated and overlapping directives, block plus line matches, partial usage                      |
| Mixed execution      | Native-only usage, plugin-only usage, different entries used by each engine, wildcard disable    |
| Configuration        | Per-file overrides, multiple config directories, CLI precedence, off/warn/error                  |
| Execution boundaries | No diagnostics, no rules, policy-skipped rules, failure, cancellation, multiple plugin batches   |
| Fixing               | Whole-comment and partial removal, conflicts, multiple rounds, changed source                    |
| Editor               | Unsaved changes, rapid edits, stale result rejection, diagnostic clearing, Quick Fix and Fix All |
| Locations            | Chinese text, emoji, BOM, LF/CRLF, multiline comments                                            |
| Compatibility        | Mismatched protocol peers fail clearly instead of silently losing diagnostics                    |

Extend existing Go directive, linter, configuration, API, and LSP suites and
existing Node plugin/API integration suites. Add focused ESLint conformance
cases comparing messages, positions, severity, and fixed output.

Select verification scope from the actual diff and affected callers. Rebuild
with `pnpm --filter @rslint/core build:bin` before JS integration checks exercise
changed Go code. Follow repository test-layout rules and scoped verification;
this plan does not request whole-repository test suites. Update user-facing
documentation and affected architecture contracts.

<!-- cspell:ignore GOTMPDIR -->

## Implementation verification (2026-11-07)

The implementation is ready for benchmark instructions, not final delivery.
The branch is `feat/report-unused-disable-directives`; the main baseline remains
`c711b7328a6c8892c70def785d5dd3e846b38d95`. Changes are uncommitted. No performance
comparison has run, so there is no candidate commit SHA or performance claim.

Environment: Linux x64, Go 1.27.3, Node 26.5.0, pnpm 10.34.4. Go checks used
`GOTMPDIR=$HOME/.cache/rslint-unused-directives/tmp` after the default
`/tmp` build hit a quota limit. Package tests were serialized with `-p 1` to
avoid simultaneous large test-binary links exhausting memory.

Builds passed:

```sh
pnpm --filter @rslint/core build:bin
cargo build --release -p rslint-native --locked --offline
# Place the locally built addon using the existing host artifact layout.
cp target/release/librslint_native.so crates/rslint-native/rslint.linux-x64-gnu.node
node scripts/place-host-build.mjs node
pnpm --filter @rslint/core build:js
```

The workspace initially had an empty generated `packages/rslint/rule-schemas.json`
and an outdated native addon without `MemoryArena`. The existing
`go run ./tools/dump_rule_schemas` generator regenerated 823 schemas. The addon
was rebuilt using cached dependencies with Cargo locked and offline. Neither
repair changed dependency manifests or lockfiles.

Scoped Go checks passed across the nine affected packages. The main check was:

```sh
go test -p 1 ./internal/rule ./internal/config ./internal/config/lint ./internal/linter ./cmd/rslint ./internal/api/server ./internal/lsp ./internal/output -run 'Test(Unused|DisableManager|ParseRuleNames|MatchDirective|IsBlockDisabled|GetConfigForFile|ConfigDecode|ConfigJSONRoundTrip|ResolverUsesBound|Pipeline|PluginSource|PrepareLintPlan|ParseLintFlagsUnused|HandleLint_EslintPlugin|LSPGeneration|LSPUnused|MergePluginDiagnostics_DropsStaleGeneration|ComputeFixAllContent|HandleFixAllCodeActionUsesEditorBOM|QuickFixCancellation|HandleDidChange_Rapid|RapidChanges|RenderUnnamed)'
go test -p 1 ./internal/config/discovery -run 'Test(ValidateConfigLoadBatchRejectsProtocolViolations|ConfigDiscoveryPropagatesActivationFailures|ConfigDiscoveryForwardsFreshModuleLoadMode)'
```

After the source-alias projection correction, these affected checks also passed:

```sh
go test -p 1 ./internal/linter -run 'Test(Pipeline|UnusedDirective)'
go test -p 1 ./cmd/rslint ./internal/api/server ./internal/lsp -run 'Test(ParseLintFlagsUnused|HandleLint_EslintPlugin|HandleLint_ExplicitExternalConfig|LSPUnused|LSPGeneration|UnusedDirective|MergePluginDiagnostics_DropsStaleGeneration|ComputeFixAllContent|HandleFixAllCodeActionUsesEditorBOM|QuickFixCancellation)'
```

The final JS checks used matching rebuilt Go, JS, and native artifacts:

```sh
# 226 tests passed.
CI=true pnpm --dir packages/rslint exec rs test run tests/config-module-host.test.ts tests/native-config-activation.test.ts tests/node-service.test.ts tests/engine.test.ts tests/engine-warmup.test.ts tests/api.test.mjs tests/config-loader.test.ts tests/service.test.ts tests/args.test.ts
# 173 tests passed.
CI=true pnpm --dir packages/rslint-test-tools exec rs test run tests/cli/disable-comments.test.ts tests/eslint-plugin/linter/diagnostic-loc-and-directives.test.ts tests/eslint-plugin/regression-coverage.test.ts
pnpm exec tsgo -p packages/rslint/tsconfig.build.json --noEmit
pnpm exec tsgo -p packages/rslint/tsconfig.worker.json --noEmit
pnpm run format:check
```

Additional checks: the serialized config schema accepted all eight supported
severity values and rejected four invalid values using the installed Ajv;
15 conformance fixtures compare messages, positions, severity, fix ranges, and
fixed output against the pinned ESLint directive engine. Their generator reuses
the installed ESLint 10.9.0 parser/config helpers and loads the pinned 10.14.0
directive engine. The existing cached cspell 10.3.3 executable checked explicit
changed-file paths with `lint --no-progress --show-context` and found no issues.

`pnpm exec tsgo -p website/tsconfig.json --noEmit` remains blocked by ten existing
errors: two for the missing `tsconfigRootDir` overview mapping and eight for
playground `node` typing. The same errors were reproduced on main after rebuilding
its JS declarations. The task branch and all changes were restored afterward.
No unrelated website fixes or whole-repository test suites were run.

Tooling note: an initial `pnpx cspell --version` probe unexpectedly fetched
cspell 10.4.0 despite an offline environment setting. This should have required
permission first. Subsequent spell checks used the pre-existing cached 10.3.3
executable; repository dependency manifests and lockfiles are unchanged.

## Performance benchmark plan

Benchmarking is a required delivery step after implementation, not an optional
follow-up. The implementer must run the benchmarks and investigate regressions.

### Baseline and reproducibility

Compare the final PR implementation against its corresponding main baseline.
Record both exact SHAs. Use identical build profiles, Node/Go versions, source
inputs, configuration, rule selection, and command arguments. Each revision must
use its own matching JS, Go, and native runtime artifacts.

If a final rebase changes relevant execution paths, update the baseline and
rerun affected comparisons. Do not include build time in lint measurements.

### Workloads

| Workload                                          | Purpose                                                                |
| ------------------------------------------------- | ---------------------------------------------------------------------- |
| Native rules, no directives, reporting off        | Detect overhead on the common path                                     |
| Native rules with directives                      | Verify early suppression and skipped fix construction remain effective |
| JS plugins with many suppressed diagnostics       | Measure added transport and temporary memory                           |
| Large project with native and plugin rules        | Exercise scheduling, Program lifetime, and peak memory                 |
| Controlled input with many overlapping directives | Detect matching/finalization complexity problems                       |
| Reporting enabled, including fixes                | Measure new functionality and fix-round costs                          |

Separate two comparisons:

1. Equivalent functionality: reporting off on both revisions, with matching
   diagnostics, exit codes, and fixed output. This measures architectural
   regressions.
2. Feature cost: reporting on versus off within the new implementation. Do not
   present the old version's missing feature as an equivalent workload.

### Measurement procedure

- Use both real repositories and controlled fixtures.
- Warm up independently, then alternate baseline/candidate runs with at least
  ten valid pairs per principal scenario.
- Record wall-time medians, variability, and paired differences.
- Measure peak Node-plus-Go process-tree RSS separately from timing. Explain
  sampling limitations and shared-page accounting.
- Record returned plugin diagnostic counts and transport bytes where useful
  to explain suppression-heavy results.
- Restore identical inputs before every fixing run.
- Keep profiling runs separate from benchmark timing.
- Retain raw results, commands, environment details, and reasons for exclusions.
- Prefer existing benchmark tooling. Ask before installing a new tool.

### Performance acceptance

A repeatable wall-time regression above 5% or peak RSS increase above 10% is a
proposed investigation threshold, not an automatic statistical conclusion or
an allowance for unexplained regressions. Consider absolute milliseconds for
short workloads and measurement variability for all results.

Investigate material regressions before delivery, especially for plugin-heavy
suppression workloads. Do not conceal a slow scenario behind improvements in
unrelated scenarios. Document residual tradeoffs and any unresolved decision.

## Selected conformance and configuration behavior

The implementation targets ESLint 10.14.0 at
`bdd4b1da21dca81d50933edef0a015216059d877`. Both directive prefixes use the
same attribution and fixes. Reporting defaults to off to preserve existing
rslint behavior (ESLint flat config defaults to warn).

| Input                                                        | Effective severity                                                   |
| ------------------------------------------------------------ | -------------------------------------------------------------------- |
| No setting                                                   | off                                                                  |
| Config `false`, `0`, or `"off"`                              | off                                                                  |
| Config `true`, `1`, or `"warn"`                              | warn                                                                 |
| Config `2` or `"error"`                                      | error                                                                |
| Later matching config entry                                  | Overrides the earlier explicit setting                               |
| `--report-unused-disable-directives`                         | error, overriding config                                             |
| `--report-unused-disable-directives=false`                   | off, overriding config                                               |
| `--report-unused-disable-directives-severity off/warn/error` | Overrides config and the boolean flag, independent of argument order |
| Invalid config severity or CLI severity                      | Configuration/argument error                                         |

The JS API uses the per-file config setting. No historical API boolean is
introduced. Enabled rules skipped by execution policy exempt their named
entries and wildcard directives from unused reporting; configured-off rules do
not. Failure or cancellation prevents unused reports for incomplete files.

## Implementation sequence

1. Pin semantic expectations, document configuration behavior, and audit entry
   points, completion boundaries, and source lifetimes.
2. Update the Go directive model and validate matching, attribution, and fixes
   against existing tests and ESLint conformance cases.
3. Integrate unfiltered Node diagnostics, protocol compatibility, and file-level
   completion.
4. Finish CLI/config/API/LSP integration, failure handling, and fix lifecycles.
5. Complete scoped verification, documentation, and editor coordination.
6. Run benchmarks, investigate regressions, and attach reproducible results.

## Editor coordination

Counterpart issue: https://github.com/rstackjs/rstack-editor/issues/130.
Add the implementation PR link to that issue before delivery.

## Deliverables

- A new PR with implementation and regression coverage.
- Documented behavior, defaults, and migration implications.
- Updated architecture documentation for changed contracts and flows.
- Linked editor counterpart PR or issue, with required adaptation coordinated.
- Actual verification commands and results, including coverage gaps.
- A reproducible benchmark report with baseline/candidate SHAs, environment,
  workloads, commands, timing and memory results, and limitations.

Do not claim completion based only on functional tests or an unsupported claim
that performance is unchanged.
