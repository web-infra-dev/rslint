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
