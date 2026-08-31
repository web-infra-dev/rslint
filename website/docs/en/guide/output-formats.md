# Output Formats

Use the `--format` flag to control how diagnostics are rendered.

## default

Human-readable terminal output with colored code snippets and diagnostic highlighting.

```bash
rslint .
```

```
start   Linting...

src/index.ts:5:7
  error  @typescript-eslint/no-unused-vars  'foo' is declared but its value is never read.

error   Lint failed with 1 error in 42ms (12 files, 1 rule, 8 threads)
```

In color-enabled terminals, `start` is cyan and bold, `success` is green, and `error` is red. The complete parenthesized execution details are rendered dim.

With `--type-check`, type errors are also included (see [Type Checking](/guide/type-checking) for details):

```bash
rslint --type-check .
```

```
start   Linting and type checking...

src/index.ts:5:7
  error  @typescript-eslint/no-unused-vars  'foo' is declared but its value is never read.

src/utils.ts:3:7
  error  TypeScript(TS2322)  Type 'string' is not assignable to type 'number'.

error   Lint and type check failed with 1 lint error and 1 TypeScript error in 85ms (14 files, 1 rule, 8 threads)
```

Combined mode reports one canonical file count: the deduplicated union of lint targets and compiler roots. This remains accurate when CLI arguments and rslint ignore patterns restrict the lint phase while type-check follows each tsconfig's program-wide scope.

The machine-readable formats below emit diagnostics only on stdout. They do not include default-format lifecycle/status text, thread counts, or fix counts. With `--timing`, the timing table remains on stderr.

## jsonline

One diagnostic per line as compact JSON. Suitable for programmatic consumption.

```bash
rslint --format jsonline .
```

## github

GitHub Actions workflow command format. Creates annotations directly on pull request diffs.

```bash
rslint --format github .
```

## gitlab

[GitLab Code Quality report](https://docs.gitlab.com/ci/testing/code_quality/) format. A single JSON array, suitable for the `codequality` report artifact that GitLab CI uses to annotate merge requests.

```bash
rslint --format gitlab . > gl-code-quality-report.json
```

```json
[
  {
    "description": "'foo' is declared but its value is never read.",
    "check_name": "@typescript-eslint/no-unused-vars",
    "fingerprint": "27e4b8b16cb47e2d6e6d4b8b6f6c6b6f",
    "severity": "major",
    "location": {
      "path": "src/index.ts",
      "lines": { "begin": 5, "end": 5 },
      "positions": {
        "begin": { "line": 5, "column": 7 },
        "end": { "line": 5, "column": 10 }
      }
    }
  }
]
```

Error diagnostics map to `major` severity and warnings map to `minor`. To wire this into a pipeline, add the report as a `codequality` artifact in `.gitlab-ci.yml`:

```yaml
lint:
  script:
    - rslint --format gitlab . > gl-code-quality-report.json
  artifacts:
    reports:
      codequality: gl-code-quality-report.json
```
