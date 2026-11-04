---
description: 'Install the Rstack VS Code extension for Rslint diagnostics, code actions, and automatic fixes.'
---

# VSCode Extension

Editor integration for Rslint lives in the unified [Rstack extension](https://github.com/rstackjs/rstack-editor), which also provides Rstest testing and `rs fmt` formatting.

Install `rstack.rstack` from the [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=rstack.rstack) or, for Cursor, Trae, VSCodium, and other VS Code forks, from [Open VSX](https://open-vsx.org/extension/rstack/rstack).

The extension ships no tool binaries. It resolves `@rslint/core` from your project, so the editor runs the same version as your CLI and picks up your `rslint.config.ts` automatically. Rslint diagnostics are provided by the project's language server: real-time diagnostics as you type, code actions for auto-fixable rules, auto-fix on save, and multi-workspace support.

:::warning The standalone `rstack.rslint` extension is retired

`rstack.rslint` shipped its final release and no longer receives updates. Uninstall or disable it so only one copy of Rslint runs.

:::

## Settings

| Setting                      | Default | Description                                                                                        |
| ---------------------------- | ------- | -------------------------------------------------------------------------------------------------- |
| `rstack.rslint.enable`       | `true`  | Enable or disable the Rslint integration.                                                          |
| `rstack.rslint.corePath`     | —       | Path to an `@rslint/core` package directory or its `package.json`; overrides automatic resolution. |
| `rstack.rslint.trace.server` | `off`   | LSP trace level: `off`, `messages`, or `verbose`.                                                  |

## Auto-fix on Save

To automatically fix lint issues when you save a file, add the following to your VS Code settings (`.vscode/settings.json`):

```json
{
  "editor.codeActionsOnSave": {
    "source.fixAll.rslint": "explicit"
  }
}
```

| Value        | Behavior                                                   |
| ------------ | ---------------------------------------------------------- |
| `"explicit"` | Fix on manual save (Ctrl+S / Cmd+S) only — **recommended** |
| `"always"`   | Fix on every save, including auto-save                     |
| `"never"`    | Disable auto-fix on save                                   |
