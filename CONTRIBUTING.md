# Rslint contribution guide

Thank you for your interest in contributing to Rslint! Before you start your contribution, please take a moment to read the following guidelines.

## Setup the environment

Install [Node.js](https://nodejs.org/) and [Go](https://go.dev/) first.

## Build locally

Build the project:

```bash
# init typescript-go submodule
git submodule update --init
pnpm install
pnpm build
```

Test the setup:

```bash
pnpm -r lint
```

## Debugging VSCode Extension

To Debug the VSCode Extension:

1. **Setup launch configuration**

```bash
cp .vscode/launch.template.json .vscode/launch.json
```

2. **Start debugging**

- Open the Command Palette (`Cmd+Shift+P`)
- Run `Debug: Start Debugging` or press `F5`
- Alternatively, go to the `Run and Debug` sidebar and select `Run Extension`
