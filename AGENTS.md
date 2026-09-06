# Working on rslint

Repository skills follow the branch, local verification, and test-layout rules below. Explicit user instructions take precedence.

## Branches

- Before editing code or documentation, inspect `git status --short --branch`. Use a dedicated task branch; create it before editing if currently on the default branch. Read-only investigation needs no new branch.
- Name new branches `<type>/<short-kebab-case-description>`, using `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `ci`, or `perf` according to the request. Honor an explicit user branch name. After creating or switching, verify `git branch --show-current` against the intended name before editing.
- Continue the same task on its existing branch, including after a session resumes. Preserve unrelated work when starting a separate task; do not restart from main merely because a workflow lists branch setup.

## Local verification

- Before running checks, identify affected packages from the branch diff plus staged, unstaged and untracked changes, trace changed APIs and their callers, and briefly state the selected scope. Include affected consumers for shared behavior and cross-language changes. Use the existing commands below; examples are in `CONTRIBUTING.md#verify-a-change`.
- Go: `go test <affected-package-dirs>`. JS/TS: select the workspace and test files with pnpm filters; for Rstest verification use `CI=true pnpm --dir <workspace> exec rs test run <test-files>` so missing snapshots fail. Rust: `cargo test -p <affected-crate>`. Expand scope only on concrete dependency impact; appending paths to root scripts that already contain full-tree paths does not narrow them.
- Full-repository or whole-plugin tests require an explicit user or reviewer request for that scope. This includes `go test ./...`, `go test ./internal/...`, `go test ./internal/plugins/<plugin>/...`, `pnpm run test:go`, root `pnpm test`, and `cargo test --workspace`. A CI command list is not a request to run it locally.
- Shared code, many changed files, or uncertain impact do not authorize full suites: trace dependencies and test the affected packages. Passing results remain valid until relevant code/configuration changes; do not rerun merely because the commit step started.
- For local Go lint, use `golangci-lint run --new-from-merge-base=<base-branch> <affected-package-dirs>`, normally against `origin/main`. Keep the branch-diff filter; unfiltered repository-wide lint requires an explicit request.
- Documentation/configuration-only edits need no language tests unless they change executable examples, generated content, builds, or runtime behavior. Format/fix only changed files with the repository tooling.
- Before JS integration tests exercise changed Go code, rebuild with `pnpm --filter @rslint/core build:bin`.
- When delivering changes, report the branch and actual verification commands/results, including any relevant coverage gap.

## Test organization

- Extend the appropriate existing suite for a regression; do not create a separate test file for each fix.
- Go lint-rule tests keep upstream cases in `<rule>_upstream_test.go` and rslint-added regressions, edge cases, and branch coverage in `<rule>_extras_test.go`. Do not mix them. Large extras suites may split by area as `<rule>_extras_<area>_test.go`.
- That split applies to Go rule tests. JS rule integration tests stay in one `<rule>.test.ts` and mirror upstream behavior; do not duplicate Go extras there. Detailed porting requirements live in `.agents/skills/port-rule/references/PORT_RULE.md`.
- Keep small inputs inline. Put multi-file text fixtures under the owning package's `testdata/`; reuse `internal/testutil/txtarfs` for related portable text trees. Construct symlinks, permissions, concurrency, and other OS behavior directly in Go tests.
- Keep package-specific helpers beside their tests; shared Go test infrastructure belongs in `internal/testutil`. Fixture helpers must fail on missing or empty selections.

## JavaScript compatibility

- Rules and their supporting helpers preserve JavaScript/upstream semantics. Use `internal/utils/ecmascript` for JS string/number operations, `internal/utils/unicode17` for Unicode categories, and tsgo's `scanner` for identifiers. Go's case, whitespace, and Unicode helpers are not equivalent substitutes.
- Use `internal/utils/ecmascript/regexp` (`esregexp`) for patterns from rule options, config, or linted source. Go's `regexp` is allowed only for repository-authored patterns with equivalent RE2/JS behavior and no user-controlled pattern construction.
- Match the upstream glob package and version. Only `minimatch3` and `isglob` are ported; report other requirements instead of substituting a matcher or introducing a new port.

## Commits

- Use Conventional Commits. Before committing, ensure `pnpm run check-spell <changed-text-files>` and `pnpm run format:check` have passed after the final relevant edit; reuse valid results. Pass spell-check paths explicitly so changes under hidden directories such as `.agents/` are covered. These checks do not authorize broader tests or repository-wide automatic fixes.
- Preserve existing public CLI behavior unless the requested change requires otherwise; update affected user documentation when behavior changes.

## Task-specific references

- Use pnpm for JS/TS tooling. Build/check scripts live in the relevant `package.json`; setup is in `CONTRIBUTING.md`. Initialize dependencies and submodules only when needed.
- For module-boundary, entrypoint, or runtime-flow changes, read the relevant sections of `architecture.md`. Update those sections when their contracts or flows change.
- For a new rule, use `.agents/skills/port-rule/SKILL.md`. Existing-rule fixes follow the branch, test-layout and verification rules above; a rule name alone does not request a new port. Load only reference sections needed by the current task, not the entire porting guide or architecture document.
- For website UI, reuse shadcn/ui components from `@components/ui/*`, `lucide-react` icons, and existing layout utilities. Add scoped CSS only where those cannot express the required visual.
