# VSCode Extension

Install the official extension from the [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=rstack.rslint). It provides:

- Real-time diagnostics as you type
- Code actions for auto-fixable rules
- Auto-fix on save via `source.fixAll.rslint`
- Multi-workspace support

The extension works out of the box — it uses the built-in rslint binary and automatically picks up your `rslint.config.ts`.

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

| Setting                | Default    | Description                                                 |
| ---------------------- | ---------- | ----------------------------------------------------------- |
| `rslint.enable`        | `true`     | Enable or disable the linter                                |
| `rslint.binPath`       | `built-in` | Binary source: `built-in`, `local` (workspace), or `custom` |
| `rslint.customBinPath` | —          | Path to a custom rslint binary (when `binPath` is `custom`) |
| `rslint.trace.server`  | `off`      | LSP trace level: `off`, `messages`, or `verbose`            |
