# @rslint/eslint-plugin-runner

Internal ESLint-plugin compatibility runtime for [rslint](https://github.com/web-infra-dev/rslint).

This package is **not** intended for direct end-user consumption. It is embedded
by `@rslint/core` (CLI) and the VS Code extension to execute ESLint plugin rules
on a Node `worker_threads` pool, while the rslint Go binary drives the lint
pipeline.

## Why a separate package

`@rslint/core` is a thin CLI / API wrapper around the Go binary. The pieces that
must run inside Node — plugin loading, oxc-parser → ESTree normalization, scope
analysis, fixer / suggestion plumbing, IPC framing — live here so:

- `@rslint/core` keeps a small surface and can be loaded fast even when no
  ESLint plugin is configured.
- The same runtime is reusable by the VS Code extension's in-process
  `WorkerPool`, where Go is the LSP server and Node is the LSP client.

## Stability

The exports are an implementation detail of rslint's plugin compatibility
layer. They may change in any minor release without a deprecation cycle. If you
need to embed rslint in your own tool, use `@rslint/core` instead.

## License

MIT
