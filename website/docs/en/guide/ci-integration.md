# CI Integration

## GitHub Actions

Use `--format github` to get inline annotations on pull request diffs:

```yaml
name: Lint
on: [push, pull_request]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 22
      - run: npm ci
      - run: npx rslint --format github .
```

Rslint auto-detects the `GITHUB_ACTIONS` environment variable and enables colored output in CI.

## Other CI Environments

```bash
# Fail on any error
npx rslint .

# Lint with type checking
npx rslint --type-check .

# Fail on errors or if warnings exceed threshold
npx rslint --max-warnings 10 .

# Report errors only (cleaner CI logs)
npx rslint --quiet .
```
