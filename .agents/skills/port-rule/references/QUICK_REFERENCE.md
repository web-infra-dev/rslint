# Rslint Quick Reference Card

A quick reference for common commands, file locations, and checklists when porting ESLint rules.

> **Note**: This is a reference document for [PORT_RULE.md](./PORT_RULE.md). See that document for the complete rule porting workflow.

---

## Commands Cheatsheet

| Task                     | Command                                                                                          |
| ------------------------ | ------------------------------------------------------------------------------------------------ |
| Create branch            | `git checkout -b feat/port-rule-<name>-$(date +%Y%m%d)`                                          |
| Go unit test             | `go test -count=1 ./internal/rules/<rule_name>`                                                  |
| Go related tests         | `go test -count=1 <changed package dirs and direct consumer package dirs>`                       |
| Build binary             | `cd packages/rslint && pnpm run build:bin`                                                       |
| JS unit test             | `cd packages/rslint-test-tools && pnpm exec rs test --testTimeout=10000 <rule-name>`             |
| Type check               | `pnpm typecheck`                                                                                 |
| Lint check               | `pnpm lint`                                                                                      |
| Format check             | `pnpm format:check`                                                                              |
| Format fix               | `pnpm format`                                                                                    |
| Spell check              | `pnpm -w run check-spell`                                                                        |
| Go lint changed packages | `golangci-lint run --new-from-rev=origin/main --timeout=10m <dirs containing changed .go files>` |
| Go format fix            | `pnpm format:go`                                                                                 |

---

## File Locations

| File Type         | Core Rules                                                                | Plugin Rules                                                     |
| ----------------- | ------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| Go implementation | `internal/rules/<name>/`                                                  | `internal/plugins/<plugin>/rules/<name>/`                        |
| Go tests          | `<name>_upstream_test.go` + `<name>_extras_test.go` in the rule directory | Same two-file split in the plugin rule directory                 |
| Documentation     | `internal/rules/<name>/<name>.md`                                         | `internal/plugins/<plugin>/rules/<name>/<name>.md`               |
| JS tests          | `packages/rslint-test-tools/tests/eslint/rules/<name>.test.ts`            | `packages/rslint-test-tools/tests/<plugin>/rules/<name>.test.ts` |

---

## Framework Performance Defaults

| Need                                           | Use                                                                     | Avoid                                                                        |
| ---------------------------------------------- | ----------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| Autofixes or suggestions                       | Matching `ReportNodeWithDeferred*` or `ReportRangeWithDeferred*` method | Constructing edit-only text, ranges, slices, or suggestions before reporting |
| References to one declared symbol              | `ctx.Refs.References(decl.Symbol())`                                    | Full-file AST walk with one `GetSymbolAtLocation` call per identifier        |
| Identifier → symbol, including globals/`.d.ts` | `ctx.Refs.Resolve(node)`                                                | A hand-rolled "try `ctx.Refs`, fall back to the checker" wrapper             |
| Every comment in the file                      | `ctx.Comments.All()`                                                    | Calling `ForEachComment` on `ctx.SourceFile.AsNode()`                        |
| Resolve a module/source target                 | `ctx.Program().ResolveModule(...)`                                      | A resolver helper in `internal/utils` or a raw compiler Program              |
| Enumerate generic module references            | `ctx.Program().ModuleGraph().References(...)`                           | A second graph/runtime stored in `RuleContext`                               |

Deferred edit builders may return nil and run synchronously only when their
artifact category is requested and the diagnostic is not suppressed. Detection,
message, and report range must remain independent of edit demand.

---

## Rule Catalog Inclusion

Core rules are listed in `internal/rules/all.go`'s `coreRules()`; each plugin's `all.go` exports `GetAllRules() []rule.Rule`. `rules.All()` combines those sources into the shared catalog — **do not edit `internal/config` for a new rule**.

| Rule Type                                                                            | File to edit                         | Final catalog key                                    |
| ------------------------------------------------------------------------------------ | ------------------------------------ | ---------------------------------------------------- |
| ESLint Core                                                                          | `internal/rules/all.go`              | `"no-debugger"`                                      |
| `@typescript-eslint`                                                                 | `internal/plugins/typescript/all.go` | `"@typescript-eslint/no-explicit-any"`               |
| Other plugins (react, jest, import, jsx-a11y, promise, react-hooks, rstest, unicorn) | `internal/plugins/<plugin>/all.go`   | `"<plugin>/<rule>"` (e.g. `"import/no-self-import"`) |

**How to add a rule**: add the import path, then append the rule var to `coreRules()` for a core rule or the plugin's `GetAllRules()` for a plugin rule:

```go
import "github.com/web-infra-dev/rslint/internal/.../my_rule"

// Core rule: internal/rules/all.go
func coreRules() []rule.Rule {
    return []rule.Rule{
        // …existing entries…
        my_rule.MyRuleRule,
    }
}

// Plugin rule: internal/plugins/<plugin>/all.go
func GetAllRules() []rule.Rule {
    return []rule.Rule{
        // …existing entries…
        my_rule.MyRuleRule,
    }
}
```

The catalog key comes from `rule.Name`. Core rules use `rule.Rule{Name: "…"}` (bare). `@typescript-eslint` rules use `rule.CreateRule(rule.Rule{Name: "…"})` which auto-prefixes `@typescript-eslint/`; **never** use `rule.CreateRule` outside `@typescript-eslint/` — it silently produces the wrong key.

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

    // Unified source/module facade (only when naming Program module types)
    "github.com/web-infra-dev/rslint/internal/program"

    // AST and type system (from typescript-go submodule)
    "github.com/microsoft/TypeScript/tsc/shim/ast"
    "github.com/microsoft/TypeScript/tsc/shim/checker"
    "github.com/microsoft/TypeScript/tsc/shim/core"

    // Utility functions
    "github.com/web-infra-dev/rslint/internal/utils"

    // JavaScript semantics — trim/blank/case/number, never strings.TrimSpace
    // or strings.ToLower
    "github.com/web-infra-dev/rslint/internal/utils/ecmascript"

    // A general category (\p{Lu}, \p{L}, \p{M}) on the edition of Unicode Node
    // reads — never the standard library's unicode package
    "github.com/web-infra-dev/rslint/internal/utils/unicode17"

    // A JavaScript RegExp — never the standard library's regexp, which is RE2
    esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"

    // Globs the way a plugin reads them (minimatch 3.x); minimatch 10 is NOT ported
    "github.com/web-infra-dev/rslint/internal/utils/minimatch3"

    // "Is this a glob or a plain path?" (is-glob 4.0.3)
    "github.com/web-infra-dev/rslint/internal/utils/isglob"

    // Test framework
    "github.com/web-infra-dev/rslint/internal/rule_tester"

    // Fixtures (test files only)
    "github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
)
```

---

## Checklist Before Submission

- [ ] Go tests pass for the rule package and any changed-related Go package directories (see Phase 4 Step 2 in `PORT_RULE.md`)
- [ ] Build binary (`cd packages/rslint && pnpm run build:bin`)
- [ ] JS snapshots generated (`pnpm exec rs test <name> -u`)
- [ ] JS tests pass (`cd packages/rslint-test-tools && pnpm exec rs test <name>`)
- [ ] Go/JS test coverage is semantically aligned (`JS ⊆ Go upstream`; extras remain Go-only)
- [ ] Autofixes/suggestions use deferred report builders and have `Test<Rule>EditDemand` in the existing extras test file
- [ ] ESLint `variable.references` usage maps to `ctx.Refs` with a binder symbol
- [ ] Whole-file comment scans use `ctx.Comments.All()`
- [ ] Cross-file source/module queries use `ctx.Program()` without backend-kind branches
- [ ] Regexps go through `esregexp`, globs through `minimatch3`/`isglob`, and JS string/number semantics through `ecmascript` — no `strings.TrimSpace`, `strings.ToLower`/`ToUpper`, stdlib `regexp`, or `doublestar` on a value that came from JavaScript
- [ ] A character question goes to `ecmascript` (case, whitespace), `esregexp` (`/i` comparison), `unicode17` (a general category) or tsgo's `scanner` (identifier) — never to the standard library's `unicode`
- [ ] Grep the change for `"regexp"` and account for every hit: a stdlib pattern is allowed only when it is written here, RE2 and JavaScript read it the same way, and no user input reaches it — otherwise it takes `esregexp`
- [ ] If the upstream rule reads globs with anything but minimatch 3 or is-glob — `minimatch@10` included — it was reported to the user rather than silently ported onto `minimatch3`/`doublestar` or hand-rolled
- [ ] Type check passes (`pnpm typecheck`)
- [ ] Lint check passes (`pnpm lint`)
- [ ] Spell check passes (`pnpm -w run check-spell`)
- [ ] Format check passes (`pnpm format:check`)
- [ ] Changed-package Go lint passes (packages containing changed `.go` files under `cmd/` and `internal/`; see Phase 4 Step 7 in `PORT_RULE.md`)
- [ ] Rule included in the catalog (in `internal/rules/all.go` for core, or the plugin's `all.go` otherwise)
- [ ] Test file registered (`packages/rslint-test-tools/rstack.config.mts`)
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
