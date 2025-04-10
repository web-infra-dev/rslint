# Benchmarks

> [!NOTE]
> Oxlint, Biome and `deno lint` are not considered in this benchmark because they [do not support typed linting](https://www.joshuakgoldberg.com/blog/why-typed-linting-needs-typescript-today/).

## Results

> On AMD Ryzen 7 5800H (8 cores, 16 threads)

| Repository                                                      | ESLint + typescript-eslint | tsgolint | Speedup |
| --------------------------------------------------------------- | -------------------------- | -------- | ------- |
| [microsoft/vscode](https://github.com/microsoft/vscode)         | 166.1s                     | 4.39s    | **38x** |
| [microsoft/typescript](https://github.com/microsoft/typescript) | 47.1s                      | 1.95s    | **24x** |
| [typeorm/typeorm](https://github.com/typeorm/typeorm)           | 27.1s                      | 0.76s    | **35x** |
| [vuejs/core](https://github.com/vuejs/core)                     | 20.5s                      | 0.89s    | **23x** |

<details>

<summary>Detailed report</summary>

| microsoft/vscode |
| ---------------- |

```plaintext
Benchmark 1: eslint
  Time (mean ± σ):     166.106 s ±  1.831 s    [User: 209.714 s, System: 10.169 s]
  Range (min … max):   162.952 s … 168.447 s    10 runs

  Warning: Ignoring non-zero exit code.

Benchmark 2: tsgolint
  Time (mean ± σ):      4.391 s ±  0.108 s    [User: 58.501 s, System: 4.573 s]
  Range (min … max):    4.268 s …  4.589 s    10 runs

Summary
  tsgolint ran
   37.83 ± 1.02 times faster than eslint
```

| microsoft/typescript |
| -------------------- |

```plaintext
Benchmark 1: eslint
  Time (mean ± σ):     47.146 s ±  0.976 s    [User: 69.958 s, System: 4.263 s]
  Range (min … max):   45.922 s … 48.800 s    10 runs

  Warning: Ignoring non-zero exit code.

Benchmark 2: tsgolint
  Time (mean ± σ):      1.947 s ±  0.019 s    [User: 17.527 s, System: 1.323 s]
  Range (min … max):    1.916 s …  1.975 s    10 runs

Summary
  tsgolint ran
   24.21 ± 0.55 times faster than eslint
```

| typeorm/typeorm |
| --------------- |

```plaintext
Benchmark 1: eslint
  Time (mean ± σ):     27.109 s ±  0.232 s    [User: 42.639 s, System: 2.311 s]
  Range (min … max):   26.809 s … 27.527 s    10 runs

  Warning: Ignoring non-zero exit code.

Benchmark 2: tsgolint
  Time (mean ± σ):     767.8 ms ±  23.5 ms    [User: 9276.8 ms, System: 947.1 ms]
  Range (min … max):   743.8 ms … 804.4 ms    10 runs

Summary
  tsgolint ran
   35.31 ± 1.12 times faster than eslint
```

| vuejs/core |
| ---------- |

```plaintext
Benchmark 1: eslint
  Time (mean ± σ):     20.503 s ±  0.417 s    [User: 35.541 s, System: 2.026 s]
  Range (min … max):   19.955 s … 21.143 s    10 runs

  Warning: Ignoring non-zero exit code.

Benchmark 2: tsgolint
  Time (mean ± σ):     895.2 ms ±  21.9 ms    [User: 10928.3 ms, System: 918.2 ms]
  Range (min … max):   866.0 ms … 924.3 ms    10 runs

Summary
  tsgolint ran
   22.90 ± 0.73 times faster than eslint
```

</details>

## How to run benchmarks

Prerequisites:

- Built `tsgolint` binary. See [README.md](../README.md) for how to build it.
- Node.js & Corepack
- [`hyperfine`](https://github.com/sharkdp/hyperfine)

### Running in Docker/Podman

```shell
docker build --file ./Containerfile --progress plain ..

# or

podman build --file ./Containerfile --progress plain ..
```

### Running locally

1. Clone the repositories
   ```bash
   ./clone-projects.sh
   ```
2. Install deps and setup ESLint configs
   ```bash
   ./setup.sh
   ```
3. Run benchmarks
   ```bash
   ./bench.sh
   ```
