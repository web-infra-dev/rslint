# ESLint directive conformance

The baseline is ESLint 10.14.0, revision
`bdd4b1da21dca81d50933edef0a015216059d877`. Upstream tests and adapted
fixtures retain ESLint's MIT license in `LICENSE`.

## Coverage

- `original.json` preserves all 89 directive-engine test inputs and results.
  The importer executes the original assertions. `upstream.json` adapts 88
  source-representable cases into 528 severity/fix-demand variants. The remaining
  internal invalid-type exception is recorded explicitly and asserted during import.
- `source.json` captures 691 distinct source states from 486 passing upstream
  tests, including intermediate autofix rounds. Its origins ledger maps states
  back to upstream names and lines. Nine complete describe blocks are selected:
  rule filtering; block, line, and next-line directives; descriptions; unused
  reporting and autofix; suppressed messages; plugin directives; and disabling
  all reporting. Six description tests for globals, exports, and inline rule
  settings are excluded explicitly. Three tests never invoke the directive
  engine; their original assertions run and their names are recorded separately.
- Go checks suppression, unused attribution, message, severity, location, and
  exact fix range/text for both `eslint-` and `rslint-` prefixes. The engine
  corpus also varies severity and fix demand. Reports run in both forward and
  reverse order. The production fixer consumes the same oracle fixtures.
- `../unused-directives.json` contains 15 additional focused cases and remains
  separate from imported upstream tests.

ESLint rejects adjacent fix ranges in one pass; rslint already allows adjacent
nonempty replacements. The importer records an alternative batched result only
when it exactly matches a later ESLint output from the same upstream test.
The Go fixer must match either that independently observed result or the exact
ESLint single-pass output. Individual proposed fixes are always compared exactly.

These tests exercise directive behavior with upstream rule reports replayed as
inputs. They do not port ESLint's parser errors, unknown-rule error emission,
custom parser APIs, non-JavaScript language coordinates, or unsupported inline
configuration options. CLI, API, plugin, and LSP integration have separate tests.
This is not a claim that the entire ESLint linter suite was ported.

Plugin integration tests also cover compact usage summaries. Workers return a
per-file rule dictionary and numeric `[line, column, ruleIndex]` tuples before
structured cloning; a `null` dictionary entry represents wildcard suppression.
The host forwards the summary unchanged. Go validates its shape and indices
before resolving entries against the original comments. An explicitly empty
summary uses `{ rules: [], entries: [] }`; a missing or invalid summary preserves
the file's directives and suppresses their unused reports and fixes.

## Regeneration

Use an existing ESLint 10.14.0 installation and a source checkout at the revision
above. The scripts install nothing and require no workspace dependency changes.
`SOURCE` below is the checkout; `ESLINT` is the installed package directory.
The scripts verify the installed version and matching implementation files;
fixture manifests record source hashes and test inventories.

```sh
node internal/rule/testdata/generate-eslint-directives.cjs "$SOURCE" "$ESLINT"
node internal/rule/testdata/generate-eslint-source-directives.cjs "$SOURCE" "$ESLINT"
node internal/rule/testdata/generate-unused-directives.cjs "$ESLINT"
pnpm exec rs fmt internal/rule/testdata/eslint-directives/*.json internal/rule/testdata/unused-directives.json
go test -p 1 ./internal/rule ./internal/linter -run 'Disable|Directive'
```

Source provenance:
[engine tests](https://github.com/eslint/eslint/blob/bdd4b1da21dca81d50933edef0a015216059d877/tests/lib/linter/apply-disable-directives.js),
[linter tests](https://github.com/eslint/eslint/blob/bdd4b1da21dca81d50933edef0a015216059d877/tests/lib/linter/linter.js).
