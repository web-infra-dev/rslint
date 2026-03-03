# Rslint website

## Setup

Install the dependencies:

```bash
pnpm install
```

Build the `@rslint/wasm` package (required for the Playground; generates `rslint.wasm.gz`):

```bash
pnpm --filter '@rslint/wasm' build
```

> **Note:** This step requires Go to be installed. The wasm build compiles the rslint CLI to WebAssembly.

## Get started

Start the dev server:

```bash
pnpm run dev
```

Build the website for production:

```bash
pnpm run build
```

Preview the production build locally:

```bash
pnpm run preview
```
