---
name: port-rule
description: Port a new ESLint core or plugin rule to rslint, including explicitly requested batches. Use for rules not implemented yet, not for fixing, reviewing, or optimizing an existing rule.
---

# Port Rule

Implement the requested rule with upstream diagnostic, option, fix and suggestion behavior. Follow [AGENTS.md](../../../AGENTS.md) for branches, verification scope, test organization and JavaScript semantics.

## Start with the source

1. Inspect the task branch and whether the rule already exists. Continue the same task on its branch. For a new task, fetch the intended base (normally `origin/main`), create a branch following AGENTS.md, and verify its name before editing.
2. Use the supplied upstream version and URLs. Otherwise use `node .agents/skills/port-rule/scripts/search_rule.mjs <rule-name>` for discovery, then pin source, tests and docs to the latest released tag. A discovery URL on `main` is not a version pin. Resolve ambiguous origins before choosing a different rule. An upstream dependency without a matching Go port is a capability question, not an automatic blocker.
3. Read the pinned source and tests. Identify the rule's inputs, decisions and outputs, then check existing rslint, tsgo and installed Go capabilities for the operations actually used. Prefer direct reuse or a small adaptation to copying helpers or porting whole dependencies; follow [reuse and compatibility](references/PORT_RULE.md#reuse-and-compatibility) when behavior differs. Implement and verify in small increments. The tests can serve as the coverage record; a separate exhaustive plan and repeated checklists are unnecessary.

Keep one short progress record when work spans sessions or multiple rules: branch/base, upstream version, current work, valid check results and remaining gaps. Reuse it after resuming.

## Preserve coverage

- Migrate every upstream test and documentation example into Go upstream tests. Preserve grouping and origin so omissions can be reviewed. Keep unsupported framework cases as explained skips and report the gap.
- Add Go extras for reachable decisions not covered upstream and for the AST shapes the rule actually inspects. Preserve JavaScript behavior across parentheses, optional access, computed/private keys, JSX and TypeScript forms where applicable. Combine cases when they exercise the same behavior; do not add duplicate tests to satisfy a count or comment-format quota.
- Include supplied regressions and relevant real-code shapes. Search upstream issues when the source/tests leave a semantic question or missing realistic case; an issue-count quota is not a prerequisite for every port.
- Assert message IDs, exact message variants, diagnostic ranges, options/defaults and any edits. Use the selected RuleTester's actual capabilities; passing a diagnostic-count check does not verify positions or fixes.
- Keep the single JS file as the upstream mirror. Go extras stay in Go. For non-trivial rule semantics, also compare against the pinned reference on relevant real code and record actual file/rule coverage.

The detailed [coverage and assertions](references/PORT_RULE.md#coverage-and-assertions) contract explains options, diagnostic positions, skips and edit demand. Read it when writing or reviewing tests, not as a second planning checklist.

## Find the next detail

Use the following references for the current operation. Open a named path or symbol directly; use the location tables before searching whole trees. Read the relevant source definition to resolve a semantic question, not to rediscover every mapped build/catalog path. In an existing task, retain confirmed paths and check results in its progress record.

| Need                                                    | Reference                                                                                                       |
| ------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| Commands, build dependencies and exact repository paths | [QUICK_REFERENCE](references/QUICK_REFERENCE.md)                                                                |
| Existing tools, dependency reuse and narrow differences | [Reuse and compatibility](references/PORT_RULE.md#reuse-and-compatibility)                                      |
| Origin, deprecation or unsupported framework behavior   | [Upstream contract](references/PORT_RULE.md#upstream-contract)                                                  |
| Options parsing and schema                              | [Options and schema](references/PORT_RULE.md#options-and-schema)                                                |
| AST adaptation and JavaScript operations                | [AST and language semantics](references/PORT_RULE.md#ast-and-language-semantics), then the relevant API section |
| Symbols, comments, modules or deferred edits            | [Framework boundaries](references/PORT_RULE.md#framework-boundaries)                                            |
| Catalog, JS wrapper and rule documentation              | [Integration and documentation](references/PORT_RULE.md#integration-and-documentation)                          |
| Compare real diagnostics with upstream                  | [Differential validation](references/PORT_RULE.md#differential-validation)                                      |

For API lookup, start with [API and contract lookup](references/QUICK_REFERENCE.md#api-and-contract-lookup), then the linked section or named source symbol. If it does not cover the operation, search that reference or owning package before expanding. Do not enumerate all reference headings or load neighboring suites as a default preparation step. Batch independent reads, but bound their combined output to avoid truncation and repeated reads.

## Verify and deliver

Select checks from the diff and actual consumers using the command reference. Complete required artifact builds before JS tests, review generated snapshots against upstream, and verify with `CI=true`. Do not treat a successful final shell command as evidence that earlier dependent steps passed; stop at a failed prerequisite and record each check's result.

Before declaring alignment, account for upstream coverage, relevant branch/AST cases, options and diagnostic/edit behavior. Fix unexplained differences instead of recording implementation accidents as expected output.

Finish at the requested scope. For commits or publication, follow AGENTS.md and the repository PR template; local work does not imply a commit, push or PR. Report the branch, actual checks and unresolved coverage. Load [delivery and troubleshooting](references/PORT_RULE.md#delivery-and-troubleshooting) only for those steps.
