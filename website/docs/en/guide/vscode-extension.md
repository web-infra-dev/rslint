# VSCode Extension

The [Rstack](https://github.com/rstackjs/rstack-editor) VS Code extension provides Rslint integration alongside testing and formatting in one install.

## Installation

Install `rstack.rstack` from the [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=rstack.rstack):

- Open the _Extensions_ view (<kbd>Ctrl</kbd>/<kbd title="Cmd">⌘</kbd> + <kbd>Shift</kbd> + <kbd>X</kbd>) and search for `rstack.rstack`.
- Open the _Quick Open_ panel (<kbd>Ctrl</kbd>/<kbd title="Cmd">⌘</kbd> + <kbd>P</kbd>), enter `ext install rstack.rstack`, and press Enter.
- Run `code --install-extension rstack.rstack` in the terminal.

For Cursor, Trae, VSCodium, and other VS Code forks, install it from the [Open VSX Registry](https://open-vsx.org/extension/rstack/rstack).

:::tip Coming from the standalone Rslint extension

The standalone `rstack.rslint` extension is retired and receives no new features. It stays published and keeps working while the transition is underway, but we recommend switching now:

1. Install `rstack.rstack` as described above.
2. Disable or uninstall `rstack.rslint` so only one copy of Rslint runs.
3. Re-enter your settings under the `rstack.rslint.*` keys and re-bind any keybindings to the new `rstack.rslint.*` command ids. Settings are not migrated automatically. Legacy `rslint.binPath` and `rslint.customBinPath` have no equivalent; use `rstack.rslint.corePath` if you need an override. See the [migration notes](https://github.com/rstackjs/rstack-editor/blob/main/packages/vscode/README.md#coming-from-the-standalone-extensions) for details.

:::

## Features

Rstack provides:

- Real-time diagnostics as you type
- Code actions for auto-fixable rules
- Auto-fix on save via `source.fixAll.rslint`
- Multi-workspace support

The extension resolves `@rslint/core` from your project and automatically detects your `rslint.config.ts`.

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

## Settings

| Setting                      | Default | Description                                      |
| ---------------------------- | ------- | ------------------------------------------------ |
| `rstack.rslint.enable`       | `true`  | Enable or disable the Rslint integration         |
| `rstack.rslint.corePath`     | —       | Path to an `@rslint/core` package directory      |
| `rstack.rslint.trace.server` | `off`   | LSP trace level: `off`, `messages`, or `verbose` |
