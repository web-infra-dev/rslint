# Benchmarks

> [!NOTE]
> Oxlint, Biome and `deno lint` are not considered in this benchmark because they [do not support typed linting](https://www.joshuakgoldberg.com/blog/why-typed-linting-needs-typescript-today/).

## Results

> On AMD Ryzen 7 5800H (8 cores, 16 threads)

| Repository                                                      | ESLint + typescript-eslint | tsgolint | Speedup |
| --------------------------------------------------------------- | -------------------------- | -------- | ------- |
| [microsoft/vscode](https://github.com/microsoft/vscode)         | 156.9s                     | 4.52s    | **35x** |
| [microsoft/typescript](https://github.com/microsoft/typescript) | 43.9s                      | 1.95s    | **23x** |
| [typeorm/typeorm](https://github.com/typeorm/typeorm)           | 25.9s                      | 0.74s    | **34x** |
| [vuejs/core](https://github.com/vuejs/core)                     | 19.8s                      | 0.94s    | **21x** |

## How to run benchmarks

Prerequisites:

- Node.js & Corepack
- [`hyperfine`](https://github.com/sharkdp/hyperfine)

### 1. Clone the repositories

```bash
./clone-projects.sh
```

### 2. Install deps and setup ESLint configs

```bash
./setup.sh
```

### 3. Run benchmarks

```bash
./bench.sh
```
