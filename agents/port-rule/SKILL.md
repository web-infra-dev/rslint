---
name: port-rule
description: Port ESLint rules to rslint (a Go-based linter). Use when users want to port/migrate/implement an ESLint rule, add a new lint rule, or mention "port rule" or specific ESLint rule names like "no-unused-vars", "no-console", etc. This skill guides through the complete workflow from setup to PR submission.
---

# Port Rule

Port ESLint rules to rslint with 1:1 behavior parity.

## Prerequisites

### Step 1: Get Rule Name

If user didn't provide the rule name, ask for it. Accept formats:

- ESLint core: `no-console`, `no-unused-vars`
- TypeScript-ESLint: `@typescript-eslint/no-explicit-any`
- Other plugins: `import/no-duplicates`, `react/jsx-uses-react`

### Step 2: Get Rule Documentation URL

If user already provided a documentation URL, skip this step.

Otherwise, ask user to choose:

1. **Auto search** (Recommended) - Run the search script to find documentation
2. **Provide URL manually** - User provides the URL directly

**For auto search**, run:

```bash
node agents/port-rule/scripts/search_rule.mjs <rule-name>
```

The script searches:

- ESLint core rules → https://eslint.org/docs/latest/rules/
- TypeScript-ESLint rules → https://typescript-eslint.io/rules/
- Other plugins → GitHub repositories

**After search**:

- If found: Show results and ask user to confirm the URLs are correct
- If not found: Ask user to provide the documentation URL manually

## Workflow

After prerequisites are complete, follow the phases in [PORT_RULE.md](references/PORT_RULE.md):

1. **Phase 0: Branch Setup** - Create feature branch from main
2. **Phase 1: Preparation** - Collect test cases and identify edge cases
3. **Phase 2: Implementation** - Write Go rule, tests, and documentation
4. **Phase 3: Integration** - Add JS tests and register rule
5. **Phase 4: Verification** - Build binary and run all tests
6. **Phase 5: Submission** - Commit and create PR

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
