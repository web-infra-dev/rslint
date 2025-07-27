#!/usr/bin/env python3
# Read the test file and extract invalid test cases systematically
with open('packages/rslint-test-tools/tests/typescript-eslint/rules/consistent-indexed-object-style.test.ts', 'r') as f:
    content = f.read()

# Find invalid array start
invalid_start = content.find('invalid: [')
lines = content[invalid_start:].split('\n')

test_cases = []
current_case = []
depth = 0
in_test = False

for line in lines:
    if line.strip().startswith('{') and ('code:' in line or (not in_test and '{' in line)):
        if current_case and in_test:
            test_cases.append('\n'.join(current_case))
        current_case = [line]
        in_test = True
        depth = 1
    elif in_test:
        current_case.append(line)
        depth += line.count('{') - line.count('}')
        if depth == 0 and line.strip().endswith('},'):
            test_cases.append('\n'.join(current_case))
            current_case = []
            in_test = False
        elif line.strip() == '},' and depth <= 0:
            test_cases.append('\n'.join(current_case))
            current_case = []
            in_test = False

# Process the test cases to find the failing ones
print('Total invalid test cases found:', len(test_cases))

# Extract specific cases
target_indices = [19, 20, 23, 28]  # 0-based indices for test cases 20, 21, 24, 29

for i in target_indices:
    if i < len(test_cases):
        print(f'\n=== INVALID TEST CASE {i+1} ===')
        case = test_cases[i]
        print(case)
        print('=== END OF TEST CASE ===\n')