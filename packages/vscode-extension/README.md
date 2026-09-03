# Rslint VS Code Extension

> [!IMPORTANT]
>
> **This extension is retired. Migrate to the [Rstack](https://github.com/rstackjs/rstack-editor) extension.**
>
> New editor features land in the unified [Rstack](https://marketplace.visualstudio.com/items?itemName=rstack.rstack) extension (`rstack.rstack`), which covers testing, linting, and formatting in one install. It is also available on the [Open VSX Registry](https://open-vsx.org/extension/rstack/rstack) for Cursor, Trae, VSCodium, and other VS Code forks. This standalone extension stays published and keeps working while the transition is underway, but receives no new features.
>
> To switch: install `rstack.rstack`, disable or uninstall `rstack.rslint` so only one copy of Rslint runs, then re-enter your settings under the `rstack.rslint.*` keys. Settings are not migrated automatically. Legacy `rslint.binPath` and `rslint.customBinPath` have no equivalent; use `rstack.rslint.corePath` if you need an override. See the [migration notes](https://github.com/rstackjs/rstack-editor/blob/main/packages/vscode/README.md#coming-from-the-standalone-extensions).

The official VS Code extension for [Rslint](https://github.com/web-infra-dev/rslint), a high-performance JavaScript and TypeScript linter written in Go.

## 📦 Installation

- **VS Code**: Install the extension from the [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=rstack.rslint)
- **Cursor/Trae**: Install the extension from the [Open VSX Registry](https://open-vsx.org/extension/rstack/rslint).

## ⚙️ Configuration

The extension can be configured through VS Code settings:

### rslint.enable

- **Type:** `boolean`
- **Default:** `true`

Enable/disable Rslint.

### rslint.corePath

- **Type:** `string`
- **Default:** `""`

The extension uses the nearest `@rslint/core` installation for each open file.
This follows normal `node_modules` ancestry, so a monorepo can use a root
installation, nested versions, or both. Runtimes backed by the same physical
installation are shared within a workspace folder.

Yarn Plug'n'Play is not resolved or executed. Use a regular local installation
or set `rslint.corePath` to an exact package directory.

Set `rslint.corePath` to an `@rslint/core` package directory to override
automatic resolution. Relative paths are resolved from the workspace folder.
The extension does not ship a fallback core, so the project must install one.
If an installation is replaced in place without changing its package directory,
reload the VS Code window so its Node-loaded code and binary refresh together.

### rslint.trace.server

- **Type:** `"off"` | `"messages"` | `"verbose"`
- **Default:** `"off"`

Traces the communication between VS Code and the language server.
The level applies to every Rslint runtime in the current VS Code window,
including runtimes backed by different workspace folders or physical
`@rslint/core` installations. Changing the value takes effect immediately and
does not restart the language servers. All runtimes write to the shared
`Rslint Language Server(LSP)` output channel.

Keep tracing off unless it is needed for diagnosis. In particular, `verbose`
output can include source text and file paths.

## 🔧 Auto-fix on Save

To automatically fix lint issues when saving, add the following to your VS Code settings (`.vscode/settings.json`):

```json
{
  "editor.codeActionsOnSave": {
    "source.fixAll.rslint": "explicit"
  }
}
```

- `"explicit"` — Fix on manual save only (Ctrl+S / Cmd+S) — **recommended**
- `"always"` — Fix on every save, including auto-save
- `"never"` — Disable auto-fix on save

## 💬 Community

Join our community:

- [GitHub](https://github.com/web-infra-dev/rslint) - Report bugs and request features
- [Discord](https://discord.gg/uPSudkun2b) - Chat with the team and community
