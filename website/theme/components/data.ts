const kvPrefix = 'rslint-';

function processGraphData(rawGraphData: string[]) {
  return rawGraphData
    .map(string => {
      const [gitHash, dateStr, progress] = string.split(/[\t]/);
      // convert to a unix epoch timestamp
      const date = Date.parse(dateStr);
      const [passing, total] = progress.split(/\//).map(parseFloat);
      const percent = parseFloat(((passing / total) * 100).toFixed(1));

      return {
        gitHash: gitHash.slice(0, 7),
        date,
        total,
        passing,
        percent,
      };
    })
    .filter(({ date, percent }) => Number.isFinite(date) && percent > 0);
}

export const getDevelopmentLintResults = async () => {
  // Mock data for RSPress environment
  return {
    passing: `internal/rules/array_type/array_type.go
  ✓ should detect array type violations
  ✓ should suggest proper array syntax

internal/rules/no_unused_vars/no_unused_vars.go
  ✓ should detect unused variables
  ✓ should ignore used variables`,
    failing: `internal/rules/no_console/no_console.go
  ✗ should detect console statements
  ✗ should handle console methods

internal/rules/semicolon/semicolon.go
  ✗ should enforce semicolons`,
  };
};

export const getProductionLintResults = async () => {
  // Mock data for RSPress environment
  return {
    passing: `internal/rules/array_type/array_type.go
  ✓ should detect array type violations
  ✓ should suggest proper array syntax

internal/rules/no_unused_vars/no_unused_vars.go
  ✓ should detect unused variables
  ✓ should ignore used variables

internal/rules/prefer_const/prefer_const.go
  ✓ should suggest const over let
  ✓ should handle destructuring`,
    failing: `internal/rules/no_console/no_console.go
  ✗ should detect console statements
  ✗ should handle console methods

internal/rules/semicolon/semicolon.go
  ✗ should enforce semicolons

internal/rules/no_debugger/no_debugger.go  
  ✗ should detect debugger statements`,
  };
};

export const getRuleExamplesResults = async () => {
  // Mock data for RSPress environment
  return {
    'basic-linting': true,
    'typescript-support': true,
    'custom-rules': false,
    'performance-test': true,
    'error-handling': false,
    'complex-patterns': true,
  };
};

export const getDevelopmentLintRuns = async () => {
  // Mock data for RSPress environment
  const mockData = [
    'abc1234\t2024-01-15T10:00:00Z\t85/100',
    'def5678\t2024-01-16T10:00:00Z\t87/100',
    'ghi9012\t2024-01-17T10:00:00Z\t89/100',
    'jkl3456\t2024-01-18T10:00:00Z\t91/100',
    'mno7890\t2024-01-19T10:00:00Z\t93/100',
  ];
  const graphData = processGraphData(mockData);
  const mostRecent = graphData[graphData.length - 1];
  return { graphData, mostRecent };
};

export const getProductionLintRuns = async () => {
  // Mock data for RSPress environment
  const mockData = [
    'xyz1234\t2024-01-15T10:00:00Z\t88/100',
    'abc5678\t2024-01-16T10:00:00Z\t90/100',
    'def9012\t2024-01-17T10:00:00Z\t92/100',
    'ghi3456\t2024-01-18T10:00:00Z\t94/100',
    'jkl7890\t2024-01-19T10:00:00Z\t96/100',
  ];
  const graphData = processGraphData(mockData);
  const mostRecent = graphData[graphData.length - 1];
  return { graphData, mostRecent };
};
