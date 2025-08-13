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

### rslint.binPath

- **Type:** `"local"` | `"built-in"` | `"custom"`
- **Default:** `"local"`

Choose which Rslint binary to use:

- `local`: Use workspace node_modules Rslint binary
- `built-in`: Use extension's built-in Rslint binary
- `custom`: Use a custom path to Rslint binary

### rslint.customBinPath

- **Type:** `string`
- **Default:** `undefined`

Custom path to Rslint executable. Only used when `rslint.binPath` is set to `custom`. Requires reloading VS Code to take effect.

### rslint.trace.server

- **Type:** `"off"` | `"messages"` | `"verbose"`
- **Default:** `"off"`

Traces the communication between VS Code and the language server.

## 💬 Community

Join our community:

- [GitHub](https://github.com/web-infra-dev/rslint) - Report bugs and request features
- [Discord](https://discord.gg/uPSudkun2b) - Chat with the team and community
