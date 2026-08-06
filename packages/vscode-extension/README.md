# Rslint VS Code Extension

The official VS Code extension for [Rslint](https://github.com/web-infra-dev/rslint), a high-performance JavaScript and TypeScript linter written in Go.

## 📦 Installation

- **VS Code**: Install the extension from the [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=rstack.rslint)
- **Cursor/Trae**: Install the extension from the [Open VSX Registry](https://open-vsx.org/extension/rstack/rslint).

## ⚙️ Configuration

The extension uses the `@rslint/core` installed by your project; it does not
bundle a separate linter. Install it in the package or monorepo root where you
want it to apply:

```bash
pnpm add -D @rslint/core
```

Resolution is document-local, so nested packages may use different core
versions. Within a workspace folder, documents that resolve the same physical
installation share one language-server and worker pool. A monorepo with one
root installation therefore uses one runtime. A Yarn PnP boundary is
authoritative: declare `@rslint/core` in that dependency graph instead of
falling through to an unrelated ancestor `node_modules` tree. Each PnP runtime
starts from that boundary and does not evaluate configs inside a nested foreign
PnP graph.

### rslint.enable

- **Type:** `boolean`
- **Default:** `true`

Enable/disable Rslint.

### rslint.runtime.path

- **Type:** `string`
- **Default:** `""`

Optional path to a complete `@rslint/core` package directory. Relative paths
are resolved from the VS Code workspace folder. Leave it empty to use normal
node_modules or Yarn PnP resolution for each document. The package must expose
the `./editor-runtime` subpath supplied by current `@rslint/core` releases. This
setting selects the core implementation only: project config and plugin imports
still execute in each document's node_modules or Yarn PnP domain, so separate
PnP dependency graphs do not leak into one another. The configured directory is
watched even before it becomes a valid package, and all active generation files
are watched inside or outside the workspace. The generation includes core's
executable `dist`/`bin` payload, so a shared chunk or worker-only rebuild is not
mistaken for the already-running version. Completing or replacing a local build
therefore migrates open documents only after the new runtime finishes its
initial config commit.

### rslint.trace.server

- **Type:** `"off"` | `"messages"` | `"verbose"`
- **Default:** `"off"`

Traces the communication between VS Code and the language server.

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
