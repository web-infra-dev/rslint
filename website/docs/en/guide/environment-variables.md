# Environment Variables

Rslint respects the following environment variables:

| Variable         | Description                                                          |
| ---------------- | -------------------------------------------------------------------- |
| `NO_COLOR`       | Disable colored output                                               |
| `FORCE_COLOR`    | Force colored output                                                 |
| `GITHUB_ACTIONS` | Automatically detected — enables colored output in GitHub Actions CI |

CLI flags `--no-color` and `--force-color` take precedence over environment variables.
