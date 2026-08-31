# Rslint VS Code Extension

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
`Rslint Language Server(LSP)` output channel, where each physical line includes
its workspace URI and core identity.

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
