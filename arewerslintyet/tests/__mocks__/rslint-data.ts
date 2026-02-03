// Mock data for RSLint test results

export const mockPassingTests = `internal/rules/array_type/array_type.go
  ✓ should detect array type violations
  ✓ should suggest proper array syntax
  ✓ should handle nested arrays

internal/rules/no_unused_vars/no_unused_vars.go
  ✓ should detect unused variables
  ✓ should ignore used variables
  ✓ should handle function parameters

internal/rules/prefer_const/prefer_const.go
  ✓ should suggest const for immutable variables
  ✓ should allow let for mutable variables`;

export const mockFailingTests = `internal/rules/no_console/no_console.go
  ✗ should detect console statements
  ✗ should handle console methods

internal/rules/semicolon/semicolon.go
  ✗ should enforce semicolons
  ✗ should handle automatic semicolon insertion

internal/rules/indent/indent.go
  ✗ should enforce consistent indentation
  ✗ should handle mixed indentation`;

export const mockGraphData = [
  'abc1234\t2024-01-15T10:00:00Z\t85/100',
  'def5678\t2024-01-16T10:00:00Z\t87/100',
  'ghi9012\t2024-01-17T10:00:00Z\t89/100',
  'jkl3456\t2024-01-18T10:00:00Z\t91/100',
  'mno7890\t2024-01-19T10:00:00Z\t93/100',
];

export const mockExamplesData = {
  'basic-linting': true,
  'typescript-support': true,
  'custom-rules': false,
  'performance-test': true,
  'integration-test': false,
};

export const mockProcessedGraphData = [
  {
    gitHash: 'abc1234',
    date: Date.parse('2024-01-15T10:00:00Z'),
    total: 100,
    passing: 85,
    percent: 85.0,
  },
  {
    gitHash: 'def5678',
    date: Date.parse('2024-01-16T10:00:00Z'),
    total: 100,
    passing: 87,
    percent: 87.0,
  },
  {
    gitHash: 'ghi9012',
    date: Date.parse('2024-01-17T10:00:00Z'),
    total: 100,
    passing: 89,
    percent: 89.0,
  },
  {
    gitHash: 'jkl3456',
    date: Date.parse('2024-01-18T10:00:00Z'),
    total: 100,
    passing: 91,
    percent: 91.0,
  },
  {
    gitHash: 'mno7890',
    date: Date.parse('2024-01-19T10:00:00Z'),
    total: 100,
    passing: 93,
    percent: 93.0,
  },
];

export const mockTestResults = {
  passing: mockPassingTests,
  failing: mockFailingTests,
};

export const mockKVResponses = {
  'rslint-passing-tests': mockPassingTests,
  'rslint-failing-tests': mockFailingTests,
  'rslint-passing-tests-production': mockPassingTests,
  'rslint-failing-tests-production': mockFailingTests,
  'rslint-examples-data': mockExamplesData,
  'rslint-test-runs': mockGraphData,
  'rslint-test-runs-production': mockGraphData,
};

// Helper function to mock KV responses
export function mockKV() {
  const { kv } = require('@vercel/kv');

  kv.get.mockImplementation((key: string) => {
    return Promise.resolve(mockKVResponses[key] || null);
  });

  kv.lrange.mockImplementation((key: string) => {
    const data = mockKVResponses[key];
    return Promise.resolve(Array.isArray(data) ? data : []);
  });
}

// Reset mocks
export function resetKVMocks() {
  const { kv } = require('@vercel/kv');
  kv.get.mockReset();
  kv.lrange.mockReset();
}
