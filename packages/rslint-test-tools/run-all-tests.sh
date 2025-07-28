#!/bin/bash

echo "Running all tests to count passing ones..."

# Run tests and capture output
timeout 300s npm test > test_results.log 2>&1

# Count passing tests
passing_count=$(grep -E "^✔.*\([0-9]+.*ms\)$" test_results.log | wc -l | tr -d ' ')
total_files=$(ls tests/typescript-eslint/rules/*.test.ts | wc -l | tr -d ' ')

echo ""
echo "=== Test Results ==="
echo "Passing tests: $passing_count"
echo "Total test files: $total_files"
echo ""

# Show passing tests
echo "Passing tests:"
grep -E "^✔.*\([0-9]+.*ms\)$" test_results.log | sort

echo ""
echo "Failed tests (first 10):"
grep -E "^✖.*\([0-9]+.*ms\)$" test_results.log | head -10

# Clean up
rm -f test_results.log