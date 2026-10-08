---
description: 'Control Rslint terminal colors with NO_COLOR, FORCE_COLOR, and GITHUB_ACTIONS, including CLI flag precedence.'
---

# Environment Variables

Rslint respects the following environment variables:

| Variable         | Description                                                          |
| ---------------- | -------------------------------------------------------------------- |
| `NO_COLOR`       | Disable colored output                                               |
| `FORCE_COLOR`    | Force colored output                                                 |
| `GITHUB_ACTIONS` | Automatically detected — enables colored output in GitHub Actions CI |

CLI flags `--no-color` and `--force-color` take precedence over environment variables.
