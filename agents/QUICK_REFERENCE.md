# Rslint Quick Reference Card

A quick reference for common commands, file locations, and checklists when porting ESLint rules.

> **Note**: This is a reference document for [PORT_RULE.md](./PORT_RULE.md). See that document for the complete rule porting workflow.

---

## Commands Cheatsheet

| Task          | Command                                                                           |
| ------------- | --------------------------------------------------------------------------------- |
| Create branch | `git checkout -b feat/port-rule-<name>-$(date +%Y%m%d)`                           |
| Go unit test  | `go test -count=1 ./internal/rules/<rule_name>`                                   |
| Go full test  | `go test -count=1 ./internal/rules/...`                                           |
| Build binary  | `cd packages/rslint && pnpm run build:bin`                                        |
| JS unit test  | `cd packages/rslint-test-tools && npx rstest run --testTimeout=10000 <rule-name>` |
| Type check    | `pnpm typecheck`                                                                  |
| Lint check    | `pnpm lint`                                                                       |
| Format check  | `pnpm format:check`                                                               |
| Format fix    | `pnpm format`                                                                     |
| Go lint check | `pnpm lint:go`                                                                    |
| Go format fix | `pnpm format:go`                                                                  |

---

## File Locations

| File Type         | Core Rules                                                     | Plugin Rules                                                     |
| ----------------- | -------------------------------------------------------------- | ---------------------------------------------------------------- |
| Go implementation | `internal/rules/<name>/`                                       | `internal/plugins/<plugin>/rules/<name>/`                        |
| Go tests          | `internal/rules/<name>/<name>_test.go`                         | `internal/plugins/<plugin>/rules/<name>/<name>_test.go`          |
| Documentation     | `internal/rules/<name>/<name>.md`                              | `internal/plugins/<plugin>/rules/<name>/<name>.md`               |
| JS tests          | `packages/rslint-test-tools/tests/eslint/rules/<name>.test.ts` | `packages/rslint-test-tools/tests/<plugin>/rules/<name>.test.ts` |

---

## Rule Registration

Location: `internal/config/config.go`

| Rule Type            | Registration Function                      | Name Format                            |
| -------------------- | ------------------------------------------ | -------------------------------------- |
| ESLint Core          | `registerAllCoreEslintRules()`             | `"no-debugger"`                        |
| @typescript-eslint   | `registerAllTypeScriptEslintPluginRules()` | `"@typescript-eslint/no-explicit-any"` |
| eslint-plugin-import | `registerAllEslintImportPluginRules()`     | `"import/no-self-import"`              |

**Registration Format**:

```go
GlobalRuleRegistry.Register("rule-name", package.RuleNameRule)
```

---

## Naming Conventions

| Item                          | Convention                     | Example                                                      |
| ----------------------------- | ------------------------------ | ------------------------------------------------------------ |
| Go directory name             | snake_case                     | `no_empty_interface/`                                        |
| Go file name                  | snake_case                     | `no_empty_interface.go`                                      |
| Go variable name (Rule)       | PascalCase + Rule suffix       | `NoEmptyInterfaceRule`                                       |
| Rule name (ESLint Core)       | kebab-case                     | `"no-debugger"`                                              |
| Rule name (typescript-eslint) | kebab-case (auto-prefixed)     | `"no-explicit-any"` → `"@typescript-eslint/no-explicit-any"` |
| Rule name (import)            | kebab-case (manually prefixed) | `"import/no-self-import"`                                    |
| JS test file name             | kebab-case                     | `no-empty-interface.test.ts`                                 |
| MessageId                     | camelCase                      | `"unexpectedAny"`, `"missingSuper"`                          |

---

## Go Module Imports

```go
import (
    // Core rule interface
    "github.com/web-infra-dev/rslint/internal/rule"

    // AST and type system (from typescript-go submodule)
    "github.com/microsoft/typescript-go/shim/ast"
    "github.com/microsoft/typescript-go/shim/checker"
    "github.com/microsoft/typescript-go/shim/core"

    // Utility functions
    "github.com/web-infra-dev/rslint/internal/utils"

    // Test framework
    "github.com/web-infra-dev/rslint/internal/rule_tester"

    // Fixtures (test files only)
    "github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
)
```

---

## Checklist Before Submission

- [ ] Go tests pass (`go test -count=1 ./internal/rules/<name>`)
- [ ] Build binary (`cd packages/rslint && pnpm run build:bin`)
- [ ] JS tests pass (`cd packages/rslint-test-tools && npx rstest run <name>`)
- [ ] Type check passes (`pnpm typecheck`)
- [ ] Lint check passes (`pnpm lint`)
- [ ] Format check passes (`pnpm format:check`)
- [ ] Go lint check passes (`pnpm lint:go`)
- [ ] Rule registered (`internal/config/config.go`)
- [ ] Test file registered (`packages/rslint-test-tools/rstest.config.mts`)
- [ ] Documentation created (`<rule_name>.md`)

**Quick Fix Commands** (run before committing if checks fail):

```bash
pnpm format      # Fix JS/TS formatting
pnpm format:go   # Fix Go formatting
```

---

## See Also

- [PORT_RULE.md](./PORT_RULE.md) - Main rule porting workflow
- [UTILS_REFERENCE.md](./UTILS_REFERENCE.md) - Utility functions reference
- [AST_PATTERNS.md](./AST_PATTERNS.md) - AST traversal patterns and examples
