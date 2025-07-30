# RSLint Build & Test Automation - Concurrent Execution

## Overview

The `automate-build-test.js` script now supports concurrent execution, allowing multiple Claude instances to work on different tests in parallel. This significantly reduces the total time needed to fix all failing tests.

## Usage

### Sequential Mode (Default)

```bash
node automate-build-test.js
```

### Concurrent Mode

```bash
# Use default 4 workers
node automate-build-test.js --concurrent

# Use custom number of workers
node automate-build-test.js --concurrent --workers=8
```

### Help

```bash
node automate-build-test.js --help
```

## How It Works

### Work Queue System

- Creates a temporary work queue in the system temp directory
- Each test is added as a work item with atomic claiming mechanism
- Workers claim work items one at a time to avoid conflicts

### File Locking via Hooks

- Uses Claude Code hooks (PreToolUse/PostToolUse) to prevent file conflicts
- When a worker edits a file, it acquires a lock
- Other workers wait if they need to edit the same file
- Locks are automatically released after edits complete

### Progress Tracking

- Main process monitors overall progress every 10 seconds
- Shows completed, failed, and in-progress counts
- Workers log their individual progress

### Architecture

```
Main Process
├── Builds the project
├── Creates work queue with all tests
├── Configures Claude Code hooks
├── Spawns N worker processes
├── Monitors progress
└── Reports final results

Worker Process (×N)
├── Claims work from queue
├── Runs test
├── If failed, sends to Claude CLI for fixing
├── Updates work status
└── Repeats until no work remains
```

## Benefits

1. **Speed**: Tests are processed in parallel, reducing total time
2. **Isolation**: Each worker has its own Claude instance
3. **Reliability**: Atomic work claiming prevents duplicate work
4. **Safety**: File locking prevents concurrent edit conflicts
5. **Visibility**: Real-time progress tracking across all workers

## Considerations

- Each worker spawns its own Claude CLI instance
- File locks prevent conflicts but may cause brief waits
- Workers automatically exit when no work remains
- Failed tests are retried up to MAX_FIX_ATTEMPTS times

## Example Output

```
╔═══════════════════════════════════════════════════════════╗
║          RSLint Automated Build & Test Runner             ║
╚═══════════════════════════════════════════════════════════╝

[10:15:23] → Script started (PID: 12345, Node: v20.10.0)
[10:15:23] → Max fix attempts per test: 500
[10:15:23] → Running in concurrent mode with 4 workers

=== BUILD PHASE ===
[10:15:23] → Starting build process...
[10:15:45] ✓ Build successful

=== TEST PHASE ===
[10:15:45] → Added 48 tests to work queue
[10:15:45] → Starting worker 1: worker_0_a1b2c3d4
[10:15:45] → Starting worker 2: worker_1_e5f6g7h8
[10:15:45] → Starting worker 3: worker_2_i9j0k1l2
[10:15:45] → Starting worker 4: worker_3_m3n4o5p6
[10:15:55] ◆ Progress: 4/48 (8% success) - 3 passed, 1 failed, 4 in progress
[10:16:05] ◆ Progress: 12/48 (75% success) - 9 passed, 3 failed, 4 in progress
...
```
