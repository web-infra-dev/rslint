# Benchmarks

> [!NOTE]
> Oxlint, Biome and `deno lint` are not considered in this benchmark because they [do not support typed linting](https://www.joshuakgoldberg.com/blog/why-typed-linting-needs-typescript-today/).

## Results

> On AMD Ryzen 7 5800H (8 cores, 16 threads)

| Repository                                                      | ESLint + typescript-eslint | tsgolint | Speedup |
| --------------------------------------------------------------- | -------------------------- | -------- | ------- |
| [microsoft/vscode](https://github.com/microsoft/vscode)         | 165.7s                     | 4.48s    | **37x** |
| [microsoft/typescript](https://github.com/microsoft/typescript) | 43.5s                      | 2.14s    | **22x** |
| [typeorm/typeorm](https://github.com/typeorm/typeorm)           | 27.2s                      | 0.79s    | **34x** |
| [vuejs/core](https://github.com/vuejs/core)                     | 20.5s                      | 0.91s    | **22x** |

<details>

<summary>Detailed report</summary>

| microsoft/vscode |
| ---------------- |

```plaintext
Benchmark 1: eslint
  Time (mean ± σ):     165.673 s ±  1.079 s    [User: 208.923 s, System: 10.336 s]
  Range (min … max):   163.702 s … 167.609 s    10 runs

  Warning: Ignoring non-zero exit code.

Benchmark 2: tsgolint
  Time (mean ± σ):      4.483 s ±  0.104 s    [User: 59.480 s, System: 4.569 s]
  Range (min … max):    4.298 s …  4.602 s    10 runs

Summary
  tsgolint ran
   36.96 ± 0.89 times faster than eslint
```

| microsoft/typescript |
| -------------------- |

```plaintext
Benchmark 1: eslint
  Time (mean ± σ):     46.470 s ±  0.158 s    [User: 68.835 s, System: 3.706 s]
  Range (min … max):   46.311 s … 46.759 s    10 runs

  Warning: Ignoring non-zero exit code.

Benchmark 2: tsgolint
  Time (mean ± σ):      2.143 s ±  0.021 s    [User: 17.554 s, System: 1.349 s]
  Range (min … max):    2.118 s …  2.176 s    10 runs

Summary
  tsgolint ran
   21.69 ± 0.22 times faster than eslint
```

| typeorm/typeorm |
| --------------- |

```plaintext
Benchmark 1: eslint
  Time (mean ± σ):     27.155 s ±  0.444 s    [User: 42.132 s, System: 2.504 s]
  Range (min … max):   26.623 s … 27.810 s    10 runs

  Warning: Ignoring non-zero exit code.

Benchmark 2: tsgolint
  Time (mean ± σ):     794.6 ms ±  19.4 ms    [User: 9460.3 ms, System: 956.8 ms]
  Range (min … max):   761.5 ms … 823.7 ms    10 runs

Summary
  tsgolint ran
   34.18 ± 1.00 times faster than eslint
```

| vuejs/core |
| ---------- |

```plaintext
Benchmark 1: eslint
  Time (mean ± σ):     20.512 s ±  0.269 s    [User: 36.099 s, System: 1.909 s]
  Range (min … max):   20.062 s … 21.025 s    10 runs

  Warning: Ignoring non-zero exit code.

Benchmark 2: tsgolint
  Time (mean ± σ):     912.3 ms ±  24.0 ms    [User: 10872.0 ms, System: 929.4 ms]
  Range (min … max):   883.3 ms … 955.7 ms    10 runs

Summary
  tsgolint ran
   22.48 ± 0.66 times faster than eslint
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
