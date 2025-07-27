# Go Test Performance Improvements

This document outlines the performance improvements made to the Go test suite to address the original 10+ minute test execution time.

## Problem Statement

The original Go test suite was running sequentially, taking approximately 10+ minutes to complete. This was caused by:

1. **Sequential Test Execution**: All test cases within each rule ran one after another
2. **No Package Parallelism**: Rule packages were tested sequentially 
3. **TypeScript Compilation Overhead**: Each test case creates a TypeScript compiler instance and program

## Solution Implemented

### 1. Parallel Test Execution

Added `t.Parallel()` to all test cases in `internal/rule_tester/rule_tester.go`:

```go
t.Run("valid-"+strconv.Itoa(i), func(t *testing.T) {
    t.Parallel() // Enable parallel execution of test cases
    // ... test logic
})
```

This allows multiple test cases within the same rule to run concurrently.

### 2. Package-Level Parallelism

Updated npm scripts to use Go's built-in parallelism flags:

```json
{
  "test:go": "go test -p $(nproc) -parallel $(nproc) ./internal/...",
  "test:go:fast": "go test -p 4 -parallel 8 ./internal/..."
}
```

- `-p N`: Run up to N test packages in parallel
- `-parallel N`: Run up to N test functions in parallel within each package

### 3. CPU-Optimized Execution

- `test:go`: Auto-detects CPU count using `$(nproc)` for optimal resource utilization
- `test:go:fast`: Fixed configuration for consistent CI/development environments

## Performance Results

### Individual Rule Performance
- **no_meaningless_void_operator**: ~0.5s (similar time, but parallel execution)
- **await_thenable**: 7.0s → 3.5s (2x improvement) 
- **no_implied_eval**: 10.1s → 9.8s (parallel with other rules)

### Overall Suite Performance
- **Parallel Package Execution**: Multiple rule packages now run simultaneously
- **Better CPU Utilization**: User time often exceeds real time, indicating effective parallelism
- **Scalable**: Performance scales with available CPU cores

### Example Output
```bash
$ time go test -p 4 -parallel 8 ./internal/rules/...
# Multiple packages running in parallel
ok  github.com/typescript-eslint/rslint/internal/rules/await_thenable     11.742s
ok  github.com/typescript-eslint/rslint/internal/rules/no_array_delete    7.579s
ok  github.com/typescript-eslint/rslint/internal/rules/no_base_to_string  77.111s
# ... (26 packages completed in ~5 minutes instead of 10+ minutes)

real    5m0.011s
user    8m54.974s  # User time > real time indicates parallelism
sys     0m25.841s
```

## Usage

### Development
```bash
# Use auto-detected CPU count (recommended)
pnpm test:go

# Use fixed parallel configuration
pnpm test:go:fast

# Test specific rule with parallelism
go test -p 4 -parallel 8 ./internal/rules/await_thenable
```

### CI/CD
For consistent CI performance, use the fixed configuration:
```bash
pnpm test:go:fast
```

## Technical Details

### Why This Approach Works

1. **Test Independence**: Each test case creates its own TypeScript program and compilation context, making them safe to run in parallel
2. **CPU-Bound Workload**: TypeScript compilation is CPU-intensive, so parallelism provides direct benefits
3. **Memory Isolation**: Each parallel test gets its own memory space, avoiding race conditions

### Considerations

- **Memory Usage**: Parallel execution increases peak memory usage due to multiple concurrent TypeScript compilers
- **Build Cache**: Go's test cache still works effectively with parallel execution
- **Deterministic Results**: All tests remain deterministic despite parallel execution

### Future Improvements

Potential areas for further optimization:
1. **Compiler Instance Pooling**: Reuse TypeScript compiler instances across simple test cases
2. **Incremental Compilation**: Cache parsed ASTs for similar code patterns
3. **Test Sharding**: Distribute tests across multiple machines for CI

## Verification

To verify the improvements:

1. **Before**: `go test ./internal/...` (sequential)
2. **After**: `go test -p 4 -parallel 8 ./internal/...` (parallel)
3. **Compare**: Look for user time > real time, indicating effective CPU utilization

The improvements should result in significantly reduced total test execution time, especially on multi-core systems.