<!-- cspell:ignore Reselects -->

# Project selection behavior and coverage

The selector receives a target, a search boundary and a caller-owned filesystem
generation. It discovers configured ownership and returns a complete Program.
It does not match lint configs, resolve rules, choose gap capabilities or own
editor lifetime. Config matching and project policy belong to `internal/config`;
CLI/API loading and the LSP store supply their own Program lifetimes.

JavaScript evaluates and transports configuration. Go consumes the existing
`ResolvedFileConfig` for matched parser options. Ordinary explicit projects keep
the owner's original declaration list. Effective service requests select their
own complete Program or an authoritative gap. Configuration merging does not
require every invocation to use the same project construction strategy.

## Configuration decisions

`S` is `projectService`, `P` is `project`, and `R` is `tsconfigRootDir`.
The comparisons refer to typescript-eslint 8.70.0 without environment overrides.
They describe individual behaviors, not complete parser or tsserver compatibility.

| ID  | Input / interaction                                      | Rslint behavior                                                                                                 | Upstream comparison                                                                                                  |
| --- | -------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| P1  | Effective S=true, no effective project paths             | Discover configured ownership for the selected target                                                           | Boolean configured-project selection is compared with upstream                                                       |
| P2  | Effective S=true with P string, nonempty array or []     | Report a conflict after matching and merging                                                                    | Aligned; an empty array still conflicts                                                                              |
| P3  | Ordinary P strings/arrays across entries                 | Collect owner declarations in their original order, including entries that do not match the target              | Difference: upstream uses the final matching P value                                                                 |
| P4  | Unmatched S/R/false/null                                 | Do not change service, root or clear state                                                                      | Aligned for these fields; ordinary P paths still follow P3                                                           |
| P5  | Final matching P=false/null                              | Disable the target's explicit/default binding; enabled service may still run                                    | Effective clear is supported; the owner declaration list remains available for other targets and program-wide checks |
| P6  | P:a followed by P=[]                                     | Keep a in the declaration list                                                                                  | Difference: [] does not clear earlier rslint declarations                                                            |
| P7  | P=[] without any declared paths                          | Suppress the implicit default; effective S=true still conflicts                                                 | These outcomes agree in this case                                                                                    |
| P8  | Final matching S=false or runtime null                   | Disable service and implicit default binding; retain existing owner P declarations, including unmatched entries | Service reset agrees; owner-wide P is a rslint difference                                                            |
| P9  | Later undefined for S/P/R                                | Preserve the inherited field                                                                                    | Aligned field merging                                                                                                |
| P10 | P:a, false, then P:b                                     | Final clear is canceled; ordinary selection again sees [a,b]                                                    | Difference: false does not delete declaration history                                                                |
| P11 | P/S omitted                                              | Use owner/tsconfig.json when present                                                                            | Difference: upstream has no typed project by default                                                                 |
| P12 | Service finds no configured owner                        | Source-only gap linting; skip typed rules                                                                       | Difference: upstream rejects unowned service files by default                                                        |
| P13 | Explicit/service targets share a tsconfig                | Preserve distinct construction modes and reference contexts                                                     | Explicit reference behavior also depends on the parser/editor host                                                   |
| P14 | Selected config parsing or Program construction fails    | Report the failure                                                                                              | A normal ownership miss is distinct from failure upstream too                                                        |
| P15 | Explicit list has direct-root and import-only candidates | Prefer direct roots, then import membership, with declaration order within each priority                        | Existing rslint selection contract                                                                                   |

A missing ordinary P declaration can fail a load even when its entry does not
match a lint target. Adding S/R does not switch ordinary P between two merge
modes. To force a target to use gap linting, disable both service and explicit
binding for it. A final false/null clear affects binding, not the owner's
program-wide declaration inventory.

Root lists support early selection or rejection. Import-only membership can
require constructing a complete candidate Program. Every successful selection
also validates the target source in that Program; references require checking
source redirects. Direct root membership alone is not final proof. Selected
Programs retain all roots and dependencies. Ordinary targets cannot borrow
service Programs, and service misses cannot borrow an ordinary Program.

## CLI and API scopes

| Scope        | Entry point                                     | Explicit-project construction                                                                      |
| ------------ | ----------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| ActiveOwners | Plain CLI lint of the whole cwd                 | Eager construction for active ordinary owners, followed by binding within applicable root contexts |
| Targeted     | Focused file/subdirectory CLI lint and API lint | Root metadata followed by import membership selects needed projects                                |
| AllDeclared  | --type-check and --type-check-only              | Build every declaration in each applicable root context and check complete Programs                |

Service discovery remains target-driven. Scoped R also restricts ordinary target
binding: two contexts under one owner cannot compete by global owner order.
Shared explicit configurations are constructed once per filesystem generation;
service uses a distinct construction mode.

AllDeclared retains rslint's program-wide explicit-project contract:

- With actual targets, every target contributes its effective R context for raw
  explicit declarations, including service and clear targets. Each context builds
  the whole declaration list. Without an explicit R, each declaration keeps its
  authored base; the service default root must not replace these bases.
- Without raw explicit paths, only ordinary targets that allow the implicit
  default can request owner/tsconfig.json. S/clear alone cannot request it.
- An owner without selected targets retains its original declaration/default
  lookup. An empty service-only scope therefore builds nothing in plain lint,
  while program-wide type checking can still check the owner default.

These are rslint scope rules, not ESLint behavior. Pure P type-check-only keeps
its path without lint-target discovery. New service/root/clear options can need
target discovery; type-check-only still never executes lint rules.

## Root and path origins

| ID  | Input / interaction                                                         | Rslint behavior                                                                                       | Upstream comparison                                                                                       |
| --- | --------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| R1  | Omitted R with a loaded config                                              | Governing config file's directory, resolved in Go                                                     | Matches the documented default; upstream stack inference can instead fall back to process cwd             |
| R2  | Nested config or an explicitly selected custom config filename              | That selected config's directory                                                                      | Rslint has no dependence on standard config filenames or preset calls                                     |
| R3  | Imported config/preset, helper, object/rules spread or JSON copy            | Default remains the governing config directory                                                        | Difference: upstream call-stack candidates can differ or become ambiguous                                 |
| R4  | Inline-only API config                                                      | API cwd, through its config-directory context                                                         | Upstream without candidates uses process cwd                                                              |
| R5  | Loaded config plus inline override or entry basePath                        | Default remains the loaded config directory                                                           | Rslint ownership contract; basePath scopes an entry without changing its owner                            |
| R6  | Explicit absolute R                                                         | Override the service boundary and every relative raw P declaration base, preserving declaration order | Explicit root anchors project paths upstream too; raw declaration collection differs                      |
| R7  | Later R null / undefined                                                    | Restore the service owner default and each P authored base / preserve the inherited value             | Reset/merge agrees; inferred defaults and raw declaration origins can differ                              |
| R8  | Relative or empty R in an unmatched or subsequently overridden entry        | Validate only the final matching root, so no path error from that entry                               | Aligned                                                                                                   |
| R9  | Relative or empty R is the final matching value                             | Report an absolute-path error, including when service is false                                        | Aligned                                                                                                   |
| R10 | Number, object, array or boolean R                                          | Reject the shape when loading configuration                                                           | Difference: upstream can ignore these invalid types in unmatched or overridden entries                    |
| R11 | Boundary is not a target ancestor                                           | Continue searching the target's ancestors                                                             | Aligned; the boundary is not a containment fence                                                          |
| R12 | References or extends leave the boundary                                    | Follow them                                                                                           | Aligned                                                                                                   |
| R13 | Project literal/glob with a base containing glob characters                 | Keep the authored pattern separate from the literal directory                                         | Preserves path meaning instead of turning a directory such as `pkg[1]` into a pattern                     |
| R14 | Relative P without explicit R                                               | Use each raw P declaration's authored config/basePath origin                                          | Difference: upstream uses its inferred root or process cwd                                                |
| R15 | Trailing separator/dot segments; a foreign operating system's root spelling | Normalize a host-absolute root before matching; reject a final non-host-absolute value                | Measured on macOS: child/, child/unused/../ retain the boundary; Windows drive/UNC spellings are rejected |

R does not move the implicit owner/tsconfig.json fallback when P is omitted. This
fallback is specific to rslint; upstream does not provide it.

Presets are ordinary reusable values and do not set S/P/R. Reading, mutating,
importing or serializing them does not compute a root, change their identity or
attach provenance. Config-module discovery is a separate contract: js/mjs/ts/mts
are discovered automatically; cjs/cts loading requires explicit config selection.

## Public types and runtime values

| Option | Rslint public TypeScript type           | Runtime behavior and upstream comparison                                                                                                                                      |
| ------ | --------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| S      | `boolean`                               | `false` is public upstream too. Runtime `null` disables service, but is not in either public type. Upstream additionally exposes service object options, which rslint rejects |
| P      | `string`, `string[]`, `false` or `null` | `false` and `null` are public upstream resets. Upstream also supports `true`; rslint rejects a matching effective `true` and directs users to S                               |
| R      | `string`                                | Runtime `null` resets the root but is not in either public type. A final string must be absolute; invalid value shapes fail config loading in rslint                          |

Runtime null acceptance applies to JavaScript/JSON configuration. It does not
claim that JavaScript checked with `checkJs` and `strictNullChecks` accepts those
values. Migration keeps a null root reset in JavaScript output rather than
emitting an invalid TypeScript config or dropping the reset.

Upstream environment switches are not emulated. In the recorded 8.70.0 probe,
`TYPESCRIPT_ESLINT_PROJECT_SERVICE=true` with `S=null` and explicit P failed while
reading `allowDefaultProject` from null, before project/service conflict checking.
That exception is outside the aligned boolean behavior above.

## Permanent regression index

This index maps the behavior tables to permanent regressions. The redesigned
integrations passed 342 JavaScript tests across five files against the rebuilt
native binary, plus the affected Go config, loader, selector, CLI, API and LSP
suites. Upstream oracle evidence supports the comparison columns; the integration
tests separately validate rslint orchestration.

| Behavior                                                                               | Permanent evidence                                                                                                                                                                                                                                                                                               |
| -------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Nested configs, imports, references/order/cycles, JS, aliases and boundaries           | `selector_test.go`: `TestUpstreamProjectSelection`; 71 scenarios / 94 selections in `testdata/upstream.txtar`                                                                                                                                                                                                    |
| Host failures, lexical identity, loaded projects and compiler mode isolation           | `selector_test.go`: `TestSelectionPropagatesHostFailure`, `TestSelectionKeepsLexicalConfigIdentity`, `TestSelectionReusesLoadedProjects`, `TestServiceProgramKeepsParsedConfigModesIsolated`                                                                                                                     |
| Construction scope, root contexts and mode isolation                                   | `internal/program/loader/loader_integration_test.go`: `TestBuildProjectsRootContextsShareExecution`, `TestBuildProjectsRootContextsDeduplicateSharedPath`, `TestBuildProjectsKeepsProgramModesSeparate`, `TestBuildProjectsAllDeclaredKeepsRootForServiceAndClear`, `TestBuildProjectsAllDeclaredWithoutTargets` |
| Shared file config before/after binding and owner aliases                              | `internal/config/lint/resolver_test.go`: `TestResolverTargetConfigSurvivesSourceBinding`, `TestResolverLiteralOwnerWinsCanonicalAlias`, `TestResolverRuleOverridePreservesProjectOptions`                                                                                                                        |
| Pure explicit-project broad/focused loading and type-check-only without lint discovery | `cmd/rslint/cmd_test.go`: `TestHandleLintCommandPreservesProjectConstructionScope`, `TestTypeCheckOnlySkipsLintConfigResolution`                                                                                                                                                                                 |
| A single candidate skips root ranking but still requires its actual SourceFile         | `internal/program/loader/loader_integration_test.go`: `TestLoadProgramsSingleCandidateSkipsRootRanking`                                                                                                                                                                                                          |
| Owner P order; unmatched S/R/clear versus raw missing P                                | `packages/rslint/tests/rslint.test.mjs`: `unmatched $option preserves owner project declarations (declared=$declared)`                                                                                                                                                                                           |
| P arrays/[], false/null, restoration and matching service false                        | Same file: `explicit lintText preserves declaration order with $name`                                                                                                                                                                                                                                            |
| S:false with an unmatched P declaration versus implicit default                        | Same file: `projectService false preserves scoped owner declarations but disables the implicit default (declared=%s)`                                                                                                                                                                                            |
| Public overlapping explicit/service targets without duplicate lint execution           | Same file: `projectService lintFiles keeps overlapping %s programs from duplicating targets`                                                                                                                                                                                                                     |
| JS files/references/imports/globs with allowJs:false                                   | Same file: `projectService lintFiles handles JavaScript ownership through $name`                                                                                                                                                                                                                                 |
| Config owner/cwd, imports, presets and root resets                                     | Same file: `projectService config directory default: %s`, `projectService config owner and cwd: %s`                                                                                                                                                                                                              |
| AND selectors/basePath/ignores and API isolation/reloads                               | Same file: `projectService AND files, basePath and ignores with %s root policy`, `projectService root policies isolate concurrent API instances and fresh reloads`                                                                                                                                               |
| Service/clear targets contribute R to complete explicit type checking                  | `packages/rslint/tests/lint-targets.test.mjs`: `program-wide explicit projects use actual root contexts from $selection targets in $mode`                                                                                                                                                                        |
| Scoped R cannot steal targets through overlapping Programs; broad/focused/type-check   | Same file: `scoped root contexts keep overlapping projects separate in $name`                                                                                                                                                                                                                                    |
| Clear/default behavior and inactive-owner lint/type-check differences                  | Same file: `project scope preserves $name`                                                                                                                                                                                                                                                                       |
| Complete service peer context without implicit owner-root pollution                    | Same file: `projectService discovers the selected file project and preserves full type context (%j)`                                                                                                                                                                                                             |
| CLI roots, final root validation and null-root migration                               | Same file: `projectService CLI root uses %s`, `projectService validates final root $value after $mode configuration`, `projectService migration preserves a runtime null boundary in $configName`                                                                                                                |
| Root trailing separators/dot segments cannot escape into an owning parent project      | Same file: `projectService normalizes the CLI root boundary with %s`; non-Windows invalid roots also use the final-root-validation table                                                                                                                                                                         |
| Frozen LSP root and normal/speculative consistency                                     | `internal/lsp/lint_project_selection_test.go`: `TestProjectServiceLSPFrozenRootDirectory`, `TestProjectServiceLSPGenerationParity`                                                                                                                                                                               |
| Editor gap diagnostics/fixes and return to typed context                               | Same file: `TestProjectServiceLSPGapKeepsDiagnosticsAndFixes`, `TestProjectServiceLSPTypedGapTyped`                                                                                                                                                                                                              |
| Watched custom-project change/deletion/recreation                                      | `internal/lsp/config_watch_test.go`: `TestHandleDidChangeWatchedFilesReevaluatesCustomProject`; `internal/lsp/lint_program_store_test.go`: `TestLintProgramStoreWatchedChangeRefreshesCustomProjectDiagnostics`                                                                                                  |

Actual ESLint loading and malformed-value/root-inference probes also support the
comparison columns; temporary probes are not permanent regressions. Compound
files/basePath/ignore cases use the shared matcher and do not each have an
independent upstream oracle. Recorded runtime probes used macOS and Node 22.
Old Node with jiti, Windows and Electron were not independently replayed.

Public scope tests assert user-visible results. The owning Go tests separately
assert construction counts, reference-mode compiler behavior, candidate isolation
and frozen identities. This does not cover every LSP lifecycle combination.
Service object options, extraFileExtensions, project:true, upstream environment
switches and full tsserver project garbage collection remain unsupported.
