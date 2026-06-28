# Output Formats

Use the `--format` flag to control how diagnostics are rendered.

## default

Human-readable terminal output with colored code snippets and diagnostic highlighting.

```bash
rslint .
```

```
src/index.ts:5:7
  error  @typescript-eslint/no-unused-vars  'foo' is declared but its value is never read.

Found 1 error and 0 warnings (linted 12 files in 42ms using 8 threads)
```

With `--type-check`, type errors are also included (see [Type Checking](/guide/type-checking) for details):

```bash
rslint --type-check .
```

```
src/index.ts:5:7
  error  @typescript-eslint/no-unused-vars  'foo' is declared but its value is never read.

src/utils.ts:3:7
  error  TypeScript(TS2322)  Type 'string' is not assignable to type 'number'.

Found 1 lint error, 1 type error and 0 warnings (linted 12 files in 85ms using 8 threads)
```

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
