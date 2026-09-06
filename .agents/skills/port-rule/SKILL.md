---
name: port-rule
description: Port a new ESLint core or plugin rule to rslint, including explicitly requested batches of new rules. Use for adding or migrating a rule that is not implemented yet. Do not trigger merely because a rule is named, or for fixing, reviewing, or optimizing an existing rule.
---

# Port Rule

Implement a new rule with upstream diagnostic, option, fix, and suggestion parity. Follow [AGENTS.md](../../../AGENTS.md) for branch selection, verification scope, test layout, and JavaScript semantics. Existing-rule fixes follow those repository rules and load individual references only for the questions involved.

## Start from the task

1. Resolve the requested rule names and inspect whether they already exist. Reuse the current task branch when resuming.
2. Use documentation/source URLs supplied by the user. Otherwise run `node .agents/skills/port-rule/scripts/search_rule.mjs <rule-name>` to discover them. Resolve the source and tests to the same released upstream tag before porting; discovery URLs pointing at `main` are not a version pin. Ask only when the upstream package/version is ambiguous or unavailable.
3. Keep one brief progress record with the branch and base, upstream tag, current phase, completed checks, and remaining work. Use the available planning mechanism; this workflow does not require a particular task-tracking tool or a second printed checklist. After resuming, inspect the diff and existing progress before repeating work.

## Read references as needed

Find section headings with `rg -n '^#{1,3} ' .agents/skills/port-rule/references/<file>`. Read the sections needed for the current phase; do not load every reference at task startup.

| Work                                                             | Read                                                                                                                                                                               |
| ---------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Establish upstream coverage and unsupported framework boundaries | [PORT_RULE: Scope and Testing Philosophy](references/PORT_RULE.md#scope-rule-semantics-not-framework-parity), then [Phase 1](references/PORT_RULE.md#phase-1-preparation-critical) |
| Implement the rule, Go suites, schema and documentation          | [PORT_RULE: Phase 2](references/PORT_RULE.md#phase-2-implementation-go)                                                                                                            |
| Register the rule and its single JS upstream mirror              | [PORT_RULE: Phase 3](references/PORT_RULE.md#phase-3-integration-js)                                                                                                               |
| Verify semantic alignment and delivery checks                    | [PORT_RULE: Phase 4](references/PORT_RULE.md#phase-4-verification--build)                                                                                                          |
| AST, symbol, module or reporting APIs                            | Relevant section of [AST_PATTERNS](references/AST_PATTERNS.md)                                                                                                                     |
| Find a helper with the required JavaScript semantics             | Relevant section of [UTILS_REFERENCE](references/UTILS_REFERENCE.md)                                                                                                               |
| Look up one command or catalog location                          | [QUICK_REFERENCE](references/QUICK_REFERENCE.md)                                                                                                                                   |

## Implementation boundaries

- Keep all upstream cases in `<rule>_upstream_test.go`; keep rslint edge shapes, real-user regressions and branch coverage in `<rule>_extras_test.go` or its area splits. The coverage criteria and diagnostic assertions in PORT_RULE remain required.
- Search existing helpers by the operation you need. Reuse an equivalent helper; extract shared code only when the current implementation needs it and its consumers have matching semantics. Avoid speculative refactoring of neighboring rules.
- Use deferred report builders for edit-only work. Use `ctx.Refs` for symbols/references, `ctx.Comments.All()` for whole-file comments, and `ctx.Program()` for source/module services. Read the matching API section before introducing another implementation.
- Keep rule documentation focused on user-visible behavior. Unsupported framework concepts are recorded as explained skips in upstream tests; do not recreate framework features inside a rule.

## Verify and deliver

Use the existing scoped commands from [QUICK_REFERENCE](references/QUICK_REFERENCE.md#commands) and [PORT_RULE Phase 4](references/PORT_RULE.md#phase-4-verification--build). Inspect the branch diff plus staged, unstaged and untracked changes, trace affected consumers, and briefly state the package/test selection before running checks. For JS integration, select the exact registered test file and rebuild the binary after relevant Go changes.

Complete Phase 4's semantic contract review and differential validation where applicable. Reuse passing checks until their relevant inputs change. Keep commands/results and remaining coverage gaps in the existing progress record.

For a batch, keep one row per rule with its current phase and result. Repair ordinary implementation/test failures and continue independent work. Ask for a decision only when a missing requirement or unsupported capability changes the requested scope. Do not turn every recoverable failure into a skip/retry/abort question.

Finish at the requested delivery scope. For requested commits or publication, follow [Phase 5](references/PORT_RULE.md#phase-5-submission--pr), use Conventional Commits, preserve unrelated edits, and run the required pre-commit checks after the final relevant edit. A local commit does not imply a push or PR. When a PR is part of the task, describe the implemented rules, actual verification and any user-approved omissions.
