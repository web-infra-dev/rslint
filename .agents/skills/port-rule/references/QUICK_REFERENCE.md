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
| Spell check   | `pnpm -w run check-spell`                                                         |
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

Each rule lives in a per-group `all.go` that exports a `GetAllRules() []rule.Rule` slice. Append your rule there; `config.go` iterates each slice automatically — **do not edit `config.go`**.

| Rule Type                                                                    | File to edit                         | Final registered key                                 |
| ---------------------------------------------------------------------------- | ------------------------------------ | ---------------------------------------------------- |
| ESLint Core                                                                  | `internal/rules/all.go`              | `"no-debugger"`                                      |
| `@typescript-eslint`                                                         | `internal/plugins/typescript/all.go` | `"@typescript-eslint/no-explicit-any"`               |
| Other plugins (react, jest, import, jsx-a11y, promise, react-hooks, unicorn) | `internal/plugins/<plugin>/all.go`   | `"<plugin>/<rule>"` (e.g. `"import/no-self-import"`) |

**How to add a rule**: in the relevant `all.go`, add the import path and append the rule var to the `GetAllRules()` return slice:

```go
import "github.com/web-infra-dev/rslint/internal/.../my_rule"

func GetAllRules() []rule.Rule {
    return []rule.Rule{
        // …existing entries…
        my_rule.MyRuleRule,
    }
}
```

The registration key comes from `rule.Name`. Core rules use `rule.Rule{Name: "…"}` (bare). `@typescript-eslint` rules use `rule.CreateRule(rule.Rule{Name: "…"})` which auto-prefixes `@typescript-eslint/`; **never** use `rule.CreateRule` outside `@typescript-eslint/` — it silently mis-registers the key.

---

## Naming Conventions

| Item                          | Convention                     | Example                                                                       |
| ----------------------------- | ------------------------------ | ----------------------------------------------------------------------------- |
| Go directory name             | snake_case                     | `no_empty_interface/`                                                         |
| Go file name                  | snake_case                     | `no_empty_interface.go`                                                       |
| Go variable name (Rule)       | PascalCase + Rule suffix       | `NoEmptyInterfaceRule`                                                        |
| Rule name (ESLint Core)       | kebab-case                     | `"no-debugger"`                                                               |
| Rule name (typescript-eslint) | kebab-case (auto-prefixed)     | `"no-explicit-any"` → `"@typescript-eslint/no-explicit-any"`                  |
| Rule name (import)            | kebab-case (manually prefixed) | `"import/no-self-import"`                                                     |
| JS test file name             | kebab-case                     | `no-empty-interface.test.ts`                                                  |
| MessageId                     | camelCase                      | `"unexpectedAny"`, `"missingSuper"` (JS rule-tester auto-converts kebab-case) |

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
- [ ] JS snapshots generated (`npx rstest run <name> -u`)
- [ ] JS tests pass (`cd packages/rslint-test-tools && npx rstest run <name>`)
- [ ] Go/JS test coverage aligned (same invalid cases, including comments/multi-line/nested)
- [ ] Type check passes (`pnpm typecheck`)
- [ ] Lint check passes (`pnpm lint`)
- [ ] Spell check passes (`pnpm -w run check-spell`)
- [ ] Format check passes (`pnpm format:check`)
- [ ] Go lint check passes (`pnpm lint:go`)
- [ ] Rule registered (in the appropriate `all.go`: `internal/rules/all.go` for core, `internal/plugins/<plugin>/all.go` otherwise)
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
