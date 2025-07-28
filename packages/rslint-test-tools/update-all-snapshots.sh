#!/bin/bash

echo "Updating all test snapshots..."

# Update snapshots for all test files
for test_file in tests/typescript-eslint/rules/*.test.ts; do
    if [ -f "$test_file" ]; then
        test_name=$(basename "$test_file" .test.ts)
        echo "Updating $test_name..."
        timeout 30s node --import=tsx/esm --test --test-timeout 30000 --test-update-snapshots --test-force-exit "$test_file" > /dev/null 2>&1
        if [ $? -eq 0 ]; then
            echo "✓ $test_name updated"
        else
            echo "✗ $test_name failed"
        fi
    fi
done

echo ""
echo "Running full test suite to count passing tests..."
passing_count=$(timeout 300s npm test 2>&1 | grep -E "^✔.*\([0-9]+.*ms\)$" | wc -l | tr -d ' ')
echo ""
echo "Total passing tests: $passing_count"