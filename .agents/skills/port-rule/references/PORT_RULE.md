# Rule Porting Contracts

This reference supplements [SKILL.md](../SKILL.md). Read the section needed for the current work. Commands and repository locations have one home in [QUICK_REFERENCE.md](QUICK_REFERENCE.md); branch, local verification and test-layout rules remain in [AGENTS.md](../../../../AGENTS.md).

## Upstream contract

- Use the requested upstream version; otherwise select the latest released tag. Read source, tests and documentation at that tag. Discovery links on the default branch may describe unreleased behavior.
- Determine whether the requested rule belongs to ESLint core, typescript-eslint or another plugin. Honor an explicitly requested legacy rule; otherwise resolve deprecation/replacement before choosing its catalog key. Do not register both a core and TypeScript alias for one port.
- Preserve the rule's purpose, accepted options, defaults, common diagnostics, ranges, fixes and suggestions. Use the reuse policy below for narrow edge differences; implementation mistakes are not intended differences.
- A public difference belongs in an implementation comment, the rule documentation's `Differences from upstream` section and a regression test. State the actual compatibility scope when delivering; do not claim exact parity with known differences.
- Check existing configuration and test-harness support before declaring an ESLint concept unsupported. For example, Go tests can pass `LanguageOptions` (including `sourceType`) and `Globals`. Preserve unsupported upstream cases as explained Go skips; a JS wrapper may not implement `skip`.

## Reuse and compatibility

Start from the rule's actual dependency calls and enabled options. Use the [capability lookup](QUICK_REFERENCE.md#api-and-contract-lookup) to inspect existing rslint helpers, Program services, tsgo shims and installed Go packages before implementing parsing, matching, traversal or resolution yourself. Read a definition and relevant caller to establish its contract; a different package name does not establish incompatibility.

Prefer direct reuse, then a small adaptation or extension of an existing helper. If two consumers need the same operation, expose or extract the common capability at its owning boundary rather than copy it into the rule. Keep configuration policy and backend details out of rule code. Introduce a new utility only for a demonstrated gap; a complete upstream dependency port needs a benefit beyond the existence of that import.

Compare behavior that can affect this rule, using pinned upstream cases and a few discriminating inputs. Do not turn this into a survey of every API or a full library-conformance project. Default and realistic configurations must retain their diagnostics and safe edits. Bounded, uncommon differences in an existing tool may be accepted under the repository's reuse policy without a new permission round: record the triggering input, upstream result, rslint result and practical scope in `Differences from upstream`, and cover the chosen behavior in a regression. Retain affected upstream cases and explain their differing expectation instead of silently omitting them.

A difference is not narrow merely because it occurs in an optional setting. Missing whole options, common false positives or negatives, unsafe fixes, or a change to the rule's purpose need resolution or explicit user direction. Existing user authorization for a specific tradeoff remains valid; do not ask again. Report a concrete unresolved capability and its effect when clarification is actually needed.

## Coverage and assertions

Use the tests as the coverage record. Preserve upstream groups/source references and explain non-obvious regressions beside the cases; fixed header wording, category tags and case-count quotas are not required.

| Suite                     | Required coverage                                                                                                                                                                          |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `<rule>_upstream_test.go` | Every valid/invalid case from the pinned upstream tests and documentation, including fixture-driven cases; explained skips for unsupported framework behavior.                             |
| `<rule>_extras_test.go`   | Reachable semantic decisions missing upstream, relevant tsgo/ESTree shape differences, supplied regressions and realistic inputs. Split large extras by area when it improves readability. |
| One `<rule>.test.ts`      | The upstream semantic set through the selected JS wrapper and compiled binary. Do not duplicate Go extras here.                                                                            |

Go and JS case counts may differ when one suite expands fixtures. Check semantic coverage, not numeric equality. Protocol/serialization regressions belong in their owning suites.

### Diagnostic and option assertions

- Use `rule_tester.RunRuleTester` with `fixtures.GetRootDir()`. Extend existing suites for fixes; keep small inputs inline and package fixtures under `testdata/`.
- Every invalid case asserts the expected message ID and start position. Across the suite, assert each exact message variant and complete start/end range for every reporting shape, including multiline and non-ASCII input where positions can differ. Columns follow UTF-16, not byte offsets.
- Those fields are checked only when provided: a passing Go test without `Message`, `EndLine` or `EndColumn` does not establish them.
- Use JSON-shaped options, not typed Go structs. The tester normalizes a single object map or positional `[]any` and validates the schema before calling the rule. Do not duplicate an object case solely to test bare versus array-wrapped input.
- Cover each option's accepted values and behavior-changing combinations, including omitted options versus explicit runtime defaults. `[{}]` is valid only for an object option allowing an empty object; a primitive option needs its actual default value.
- A focused reproduction uses `go test <package> -run '<test-or-subtest>'`, followed by the affected package's verification. `RunRuleTester` rejects `Only: true`; `Skip: true` requires an explained upstream limitation.

### Fixes, suggestions and edit demand

Assert fixed text in `InvalidTestCase.Output`, and suggestion message IDs plus applied text in `InvalidTestCaseError.Suggestions`. Cases that must not offer an edit also need assertions.

For a rule with fixes or suggestions, keep an edit-demand test in its extras suite. Run representative diagnostics with `EditDemandNone`, `EditDemandAutofix`, `EditDemandSuggestion` and `EditDemandAll`: count, message and range stay identical; artifacts appear only for the requested category and match the all-edits result. Ordinary `RunRuleTester` requests all edits, so output assertions alone do not check this boundary.

See `internal/plugins/typescript/rules/no_restricted_types/no_restricted_types_extras_test.go` for a combined fix/suggestion example, and `internal/rule_tester/rule_tester.go` for the current case fields.

## AST and language semantics

Select edge cases from the upstream operations, not from a fixed checklist of unrelated syntax. Existing upstream cases can already cover a decision; extras fill the uncovered behavior and representation gaps.

| Operation the rule uses               | Adaptation to check                                                                                                                                                                                                          |
| ------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Inspects an expression receiver/child | Use `utils.ESTreeRuntimeExpression` for parentheses and JS JSDoc cast wrappers; preserve authored TS wrappers. For a direct call callee, use `utils.ESTreeCallCallee` on the raw callee to retain optional-chain boundaries. |
| Matches member access                 | Distinguish identifier, computed, private and optional access. Dotted JSX tags and TypeScript heritage/type names can use similar tsgo nodes but have different ESTree roles.                                                |
| Reads literal values or source text   | Literal kinds differ; `.Text()` may be normalized. Choose decoded value versus original token text according to upstream.                                                                                                    |
| Handles assignment or sequence        | These share `BinaryExpression` in tsgo; branch on `OperatorToken.Kind`.                                                                                                                                                      |
| Walks scopes or nested containers     | Test the boundaries the rule uses: shadowing, arrows, class/static bodies, body-absent declarations and same-kind nesting where relevant.                                                                                    |
| Produces edits                        | Preserve comments and side effects, guard token fusion, and distinguish trimmed node text from raw positions.                                                                                                                |

Use [member and call expressions](AST_PATTERNS.md#member-and-call-expressions) for those node shapes and helper contracts; a static-name helper can accept computed keys or reject private names differently from upstream. Other operations are indexed in [API and contract lookup](QUICK_REFERENCE.md#api-and-contract-lookup).

Follow AGENTS.md's JavaScript helper requirements. Use `ecmascript` for JS string/number semantics, `unicode17` for Unicode categories and tsgo's `scanner` for identifiers. User-controlled regexps require `esregexp`. For globs and ignore patterns, evaluate existing matching capabilities under [reuse and compatibility](#reuse-and-compatibility); upstream package/version identifies the comparison reference, not a mandatory implementation dependency.

## Options and schema

`Rule.Run` already receives normalized `options []any`. Guard each positional access and parse that representation directly; do not call `NormalizeOptions` again inside the rule.

Every rule declares `Schema`:

- No options: reuse `rule.EmptyArraySchema`.
- Options: place the pinned upstream schema in `<rule_name>.schema.json`, embed it and use `rule.NewSchema`.
- For an upstream array of positional schemas, wrap it as `{"type":"array","items":<upstream array>,"minItems":0,"maxItems":<length>}`. A full schema object is copied as-is. The schema dialect is Draft 4.

CLI, API and LSP configuration paths validate options. Schema defaults populate existing option objects but do not create omitted outer objects or positional elements; keep runtime defaults in the parser. Verify defaults against the upstream implementation as well as its schema.

## Framework boundaries

Reuse the existing framework. An existing helper and the new rule are two consumers: expose or extract their common operation when needed, keeping rule-specific policy local.

| Need                                               | Existing boundary                                                                                                                                                                                           |
| -------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Resolve identifiers or enumerate symbol references | `ctx.Refs`, using binder symbols. Choose `ResolveInFile` when upstream excludes ambient/lib/cross-file declarations; `Resolve` has a checker fallback.                                                      |
| Configured globals                                 | `ctx.Globals.Access` is the final effective value; `Override` is authored-only, and `ConfiguredAccess` excludes inline comments. Match upstream's configuration provenance; see `internal/rule/globals.go`. |
| Type information                                   | Type-aware plugin rules declare `RequiresTypeInfo`. Core rules must still run on JS inputs; guard optional checker access and use the existing fallback semantics.                                          |
| Whole-file comments                                | `ctx.Comments.All()`.                                                                                                                                                                                       |
| Source/module services                             | `ctx.Program()`; cross-file infrastructure belongs in `internal/program`, without rule-side backend branching.                                                                                              |
| Reusable plugin semantics                          | The owning `<plugin>util` package; general AST/type operations belong in `internal/utils`.                                                                                                                  |
| Edit-only work                                     | A matching `ReportNodeWithDeferred*` or `ReportRangeWithDeferred*` builder. Diagnostic identity remains eager. The builder may return nil; do not pass a nil builder.                                       |

Do not add another per-identifier checker walk, scope index, comment scanner or module resolver where these APIs already provide the needed semantics. Read the relevant [AST_PATTERNS](AST_PATTERNS.md) or [UTILS_REFERENCE](UTILS_REFERENCE.md) section before using them. If caching is needed, keys must include the inputs/configuration that determine the result.

## Integration and documentation

Use the exact locations in [QUICK_REFERENCE](QUICK_REFERENCE.md#rule-files-and-registration).

- Core rules use `rule.Rule{Name: "<rule-name>"}`. Only TypeScript plugin rules use `rule.CreateRule`, which adds `@typescript-eslint/`. Other plugins include their prefix in `rule.Rule.Name` directly.
- Add a core rule to `coreRules()`; add an existing plugin rule to its `GetAllRules()`. The first native rule of a new plugin also needs its explicit aggregation in `internal/rules/all.go`; directories are not auto-discovered. Read the relevant architecture boundary for that case.
- Register the single JS test file in `packages/rslint-test-tools/rstack.config.mts`'s `include`. For a new plugin suite, reuse its owning wrapper/configuration conventions; use JS/TS configuration, not a new legacy `rslint.json`.
- Inspect the selected wrapper's types and assertions before copying cases. The core wrapper uses object-shaped options for object rules; other wrappers can accept positional arrays. Prefixing, skip support and snapshot support differ. In particular, the jsx-a11y wrapper does not provide the core wrapper's snapshots, message-ID or position assertions.
- Unsupported cases cannot be silently discarded. Use a supported skip mechanism or retain an explained case/comment and report the coverage gap when the JS wrapper cannot express it.
- Keep documentation focused on the rule's behavior, options and correct/incorrect examples. Include official docs when available and a source link pinned to the exact tag. Record requested/established public differences there; reusable AST/API discoveries belong in their reference, not product-facing rule explanations.
- If the plugin is already enabled in repo-root `rslint.config.ts`, add the new rule at `warn` before verification. Otherwise do not enable a plugin or change the preset as a side effect of porting.

Build prerequisites are listed once in the command reference. A fresh binary does not imply fresh core JS, generated option types or wrapper output.

## Differential validation

For non-trivial semantics, compare against the pinned upstream implementation on relevant real source files before claiming alignment. Reuse installed versions only after checking them; otherwise prepare exact versions in a scratch directory, without changing repository dependencies.

1. Resolve one explicit file set. Enable only the target rule, with matching options, parser settings and language mode. Type-aware rules require the same project/tsconfig context.
2. Confirm both tools actually parsed/linted those files and enabled the target rule. Include a known match so two empty outputs cannot masquerade as alignment. Parser/configuration warnings, ignored files and missing coverage evidence are gaps; do not conditionally skip a coverage assertion when its required metadata is absent.
3. Compare normalized file paths, rule names, message IDs/text, severity, complete ranges and applicable fixes/suggestions. Record the tool versions, file/diagnostic counts and fields actually compared.
4. Classify differences: input/configuration mismatch, documented requested difference, or implementation bug. Fix unexplained rule differences, add a Go regression and rerun the affected comparison.

The CLI's `--format jsonline` includes paths, rule names, messages, severity and ranges, but omits message IDs and edits. Use `lint` from `@rslint/core/internal` to compare those fields, as the JS wrappers do. A CLI-only comparison does not establish fix/suggestion or message-ID parity.

Snapshot generation is not verification. For wrappers using snapshots, review generated expectations against the pinned reference, then run the selected file with `CI=true`. A wrapper that checks only counts/messages does not verify diagnostic positions just because its test passes.

## Delivery and troubleshooting

Follow AGENTS.md's requested delivery scope and commit checks. For Go lint, preserve the branch-diff filter and select packages containing changed Go files; file arguments from different directories are not a substitute for package selection. Choose JS checks from the affected workspace's actual scripts rather than copying an unconditional root checklist.

Format only changed, supported files. Explicit paths still obey formatter exclusions, including rule Markdown; skip empty selections. Inspect spell failures before adding intentional technical words to `scripts/dictionary.txt`. Do not add Markdown cspell directives or suppress lint simply to obtain a green result.

| Symptom                                           | Check before expanding scope                                                                   |
| ------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| JS cannot find a new rule or sees old Go behavior | Catalog entry, rebuilt binary, selected test registration and wrapper prefix.                  |
| New options absent from generated JS types        | Schema dump before core JS build; the JS build does not refresh the dump itself.               |
| Missing module/dist                               | The selected workspace's actual package path and required build output.                        |
| Passing comparison with no diagnostics            | Intended files, parser/config warnings, rule enablement and a known positive input.            |
| Position mismatch                                 | UTF-16 columns, source trivia and whether upstream reports the expression or its container.    |
| TypeChecker is nil                                | Whether the rule declares type information and the test includes the required program/project. |

Fix recoverable failures within the task and reuse checks whose inputs remain valid. Record each actual result; do not let a later successful command hide a failed prerequisite.
