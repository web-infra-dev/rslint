---
name: port-rule
description: Port ESLint rules to rslint (a Go-based linter). Use when users want to port/migrate/implement an ESLint rule, add a new lint rule, or mention "port rule" or specific ESLint rule names like "no-unused-vars", "no-console", etc. Supports both single-rule and batch-rule porting. This skill guides through the complete workflow from setup to PR submission.
---

# Port Rule

Port ESLint rules to rslint with 1:1 behavior parity. Supports porting a single rule or multiple rules in batch.

## Prerequisites

### Step 1: Get Rule Names

Accept one or more rule names from the user. Supported formats:

- ESLint core: `no-console`, `no-unused-vars`
- TypeScript-ESLint: `@typescript-eslint/no-explicit-any`
- Other plugins: `import/no-duplicates`, `react/jsx-uses-react`

Users may provide rules as:

- A single rule name
- Multiple names separated by commas, spaces, or newlines
- A list provided one-by-one in conversation

### Step 2: Get Rule Documentation URLs

For each rule, obtain the documentation URL.

If user already provided documentation URLs, skip to the next rule.

Otherwise, ask user to choose:

1. **Auto search** (Recommended) - Run the search script to find documentation
2. **Provide URL manually** - User provides the URL directly

**For auto search**, run for each rule:

```bash
node agents/port-rule/scripts/search_rule.mjs <rule-name>
```

Multiple rules can be searched in parallel.

The script searches:

- ESLint core rules → https://eslint.org/docs/latest/rules/
- TypeScript-ESLint rules → https://typescript-eslint.io/rules/
- Other plugins → GitHub repositories

**After search**:

- If found: Show results and ask user to confirm the URLs are correct
- If not found: Ask user to provide the documentation URL manually

## Planning

**Before starting any implementation**, you MUST output a structured plan, then proceed to execute immediately without waiting for user confirmation.

**IMPORTANT — Task Tracking**: After outputting the text plan, you MUST use the `TaskCreate` tool to create a persistent task for each phase/step. This ensures progress is tracked even if the conversation context is compressed. Specifically:

- Create one task per phase (for single rule) or one task per phase per rule (for batch).
- Use `TaskUpdate` to mark tasks as `in_progress` when starting and `completed` when done.
- If you are unsure about current progress (e.g., after context compression or conversation resume), call `TaskList` first to check which tasks remain.

### Single Rule

For a single rule, output a brief checklist:

```
## Plan

- [ ] Phase 0: Branch setup
- [ ] Phase 1: Preparation — collect test cases for `<rule-name>`
- [ ] Phase 2: Implementation — Go rule + tests + docs
- [ ] Phase 3: Integration — JS tests + register
- [ ] Phase 4: Verification — build + test
- [ ] Phase 5: Commit & PR
```

Then create corresponding tasks via `TaskCreate`:

- "Phase 0: Branch setup"
- "Phase 1: Preparation — collect test cases for `<rule-name>`"
- "Phase 2: Implementation — Go rule + tests + docs for `<rule-name>`"
- "Phase 3: Integration — JS tests + register for `<rule-name>`"
- "Phase 4: Verification — build + test for `<rule-name>`"
- "Phase 5: Commit & PR for `<rule-name>`"

### Batch Rules

For multiple rules, output a structured plan with a summary table and per-rule breakdown:

```
## Plan

### Rules to port (N rules)
| # | Rule | Doc URL | Status |
|---|------|---------|--------|
| 1 | <rule-name-1> | <url> | pending |
| 2 | <rule-name-2> | <url> | pending |
| ... | ... | ... | ... |

### Progress
- [ ] Phase 0: Branch setup
- Rule 1: `<rule-name-1>`
  - [ ] Phase 1: Preparation
  - [ ] Phase 2: Implementation
  - [ ] Phase 3: Integration
  - [ ] Phase 4: Verification
  - [ ] Commit
- Rule 2: `<rule-name-2>`
  - [ ] Phase 1: Preparation
  - [ ] Phase 2: Implementation
  - [ ] Phase 3: Integration
  - [ ] Phase 4: Verification
  - [ ] Commit
- ...
- [ ] Phase 5: Create PR (summarize all rules)
```

Then create corresponding tasks via `TaskCreate` — one for Phase 0, then for each rule create tasks for Phase 1–4 + Commit, and finally one for Phase 5.

Output the plan, then proceed to Phase 0 immediately.

## Workflow

Determine the mode based on the number of rules:

- **1 rule** → Single Rule Mode
- **2+ rules** → Batch Mode

### Single Rule Mode

Follow the phases in [PORT_RULE.md](references/PORT_RULE.md) sequentially:

1. **Phase 0: Branch Setup** - Create feature branch from main
2. **Phase 1: Preparation** - Collect test cases and identify edge cases
3. **Phase 2: Implementation** - Write Go rule, tests, and documentation
4. **Phase 3: Integration** - Add JS tests and register rule
5. **Phase 4: Verification** - Build binary and run all tests
6. **Phase 5: Submission** - Commit and create PR

For each phase: mark its task as `in_progress` (via `TaskUpdate`) before starting, and `completed` after finishing. Update the text checklist as well.

### Batch Mode

Follow the batch workflow in [PORT_RULE.md](references/PORT_RULE.md):

1. **Phase 0: Branch Setup** - Create a single feature branch for the batch (once)
2. **For each rule**, execute in order:
   - **Phase 1: Preparation** - Collect test cases for this rule
   - **Phase 2: Implementation** - Write Go rule, tests, and documentation
   - **Phase 3: Integration** - Add JS tests and register rule
   - **Phase 4: Verification** - Build binary and run all tests
   - **Commit** - Create an independent commit: `feat: port rule <rule-name>`
   - **Report** - Briefly report the result before moving to the next rule
3. **Phase 5: Create PR** - One PR summarizing all ported rules (once), see [Phase 5 Details](#phase-5-commit--pr-details) below

**Progress tracking**: After completing each rule, update the checklist (mark as `[x]`) and the corresponding tasks (via `TaskUpdate` to `completed`), then report the status. After a failure, mark the checklist as `[!]` with a reason.

**Failure handling**: If a rule fails at any phase, stop and ask the user:

- **(a) Skip** this rule and continue with the next one
- **(b) Attempt to fix** the issue
- **(c) Abort** the entire batch

Already-committed rules are not affected by later failures.

### Phase 5: Commit & PR Details

**Commit constraints**:

- Commit message: `feat: port rule <rule-name>`
- Do NOT include AI-related information in commit messages (no `Co-Authored-By: Claude` or similar)
- Only stage files related to the current rule(s). `pnpm format:go` may reformat unrelated files — discard them with `git checkout -- <file>` before committing.
- If the rule's plugin is already in `rslint.json` `plugins`, add the rule with `"warn"` severity. Otherwise, do NOT modify `rslint.json`.

**PR title format**:

- Single rule: `feat: port rule <rule-name>`
- Batch (single plugin): `feat: port N <plugin-name> rules`
- Batch (multiple plugins): `feat: port N rules from <plugin-1>, <plugin-2>`

**PR body template** (batch single plugin):

```
## Summary

Port N <plugin-name> rules to rslint.

### Rules ported
| Rule | Description | Doc |
|------|-------------|-----|
| `<rule-1>` | [brief description] | [link](<url>) |
| `<rule-2>` | [brief description] | [link](<url>) |

## Checklist

- [x] Tests updated (or not required).
- [x] Documentation updated (or not required).
```

**PR body template** (single rule):

```
## Summary

Port the `<rule-name>` rule from ESLint to rslint.

[Brief description of what the rule does]

## Related Links

- ESLint rule: <link_to_eslint_doc>
- Source code: <link_to_source_code>

## Checklist

- [x] Tests updated (or not required).
- [x] Documentation updated (or not required).
```

Do NOT include AI-related information in PR title or body. If any rules were skipped during batch execution, note them in the PR body.

### Completion Constraint

The workflow is complete ONLY when all tasks created during Planning are marked as `completed` (or explicitly skipped due to failure). Do NOT stop or wait for user instructions while there are still pending tasks. If the conversation context was compressed or the session was resumed, call `TaskList` first to check remaining work before continuing.

## Quick Reference

**Directory Structure**:

- Core rules: `internal/rules/<rule_name_snake_case>/`
- Plugin rules: `internal/plugins/<plugin_name>/rules/<rule_name_snake_case>/`

**Key Commands**:

```bash
# Build binary (REQUIRED before JS tests)
cd packages/rslint && pnpm run build:bin

# Run Go tests
go test -count=1 ./internal/rules/<rule_name>

# Run JS tests
cd packages/rslint-test-tools && npx rstest run --testTimeout=10000 <rule-name>

# Lint and format checks (REQUIRED before commit)
pnpm format:check && pnpm lint:go

# Auto-fix formatting issues
pnpm format && pnpm format:go
```

## References

- [PORT_RULE.md](references/PORT_RULE.md) - Complete porting guide with code templates
- [AST_PATTERNS.md](../AST_PATTERNS.md) - AST traversal patterns
- [UTILS_REFERENCE.md](../UTILS_REFERENCE.md) - Utility functions
- [QUICK_REFERENCE.md](../QUICK_REFERENCE.md) - Commands cheatsheet
