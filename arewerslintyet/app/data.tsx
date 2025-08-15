import 'server-only';
import { kv } from '@vercel/kv';
import { revalidateTag, unstable_cache } from 'next/cache';
import { Linter, getLinter } from './bundler';

const kvPrefix = 'rslint-';
const linterTag = 'rslint';

export function revalidateAll() {
  revalidateTag(linterTag);
}

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

export const getDevelopmentLintResults = unstable_cache(
  async () => {
    try {
      const [failing, passing] = await Promise.all([
        kv.get(`${kvPrefix}failing-lint-tests`),
        kv.get(`${kvPrefix}passing-lint-tests`),
      ]);

      if (failing === null && passing === null) {
        return null;
      }

      return { passing, failing };
    } catch (error) {
      // Mock data for development when KV is not available
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
    }
  },
  [kvPrefix, 'lint-results-new'],
  {
    tags: [linterTag],
    revalidate: 600,
  },
);

export const getProductionLintResults = unstable_cache(
  async () => {
    try {
      const [failing, passing] = await Promise.all([
        kv.get(`${kvPrefix}failing-lint-tests-production`),
        kv.get(`${kvPrefix}passing-lint-tests-production`),
      ]);

      if (failing === null && passing === null) {
        return null;
      }

      return { passing, failing };
    } catch (error) {
      // Mock data for development when KV is not available
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
    }
  },
  [kvPrefix, 'lint-results-new-production'],
  {
    tags: [linterTag],
    revalidate: 600,
  },
);

export const getRuleExamplesResults = unstable_cache(
  async () => {
    try {
      const data: { [ruleName: string]: /* isPassing */ boolean } =
        await kv.get(`${kvPrefix}rule-examples-data`);
      return data;
    } catch (error) {
      // Mock data for development
      return {
        'basic-linting': true,
        'typescript-support': true,
        'custom-rules': false,
        'performance-test': true,
        'error-handling': false,
        'complex-patterns': true,
      };
    }
  },
  [kvPrefix, 'rule-examples-results'],
  {
    tags: [linterTag],
    revalidate: 600,
  },
);

export const getDevelopmentLintRuns = unstable_cache(
  async () => {
    try {
      const [graphData] = await Promise.all([
        kv.lrange(`${kvPrefix}lint-runs`, 0, -1).then(processGraphData),
      ]);

      const mostRecent = graphData[graphData.length - 1];
      return { graphData, mostRecent };
    } catch (error) {
      // Mock data for development
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
    }
  },
  [kvPrefix, 'lint-runs-new'],
  {
    tags: [linterTag],
    revalidate: 600,
  },
);

export const getProductionLintRuns = unstable_cache(
  async () => {
    try {
      const [graphData] = await Promise.all([
        kv
          .lrange(`${kvPrefix}lint-runs-production`, 0, -1)
          .then(processGraphData),
      ]);

      const mostRecent = graphData[graphData.length - 1];
      return { graphData, mostRecent };
    } catch (error) {
      // Mock data for development
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
    }
  },
  [kvPrefix, 'lint-runs-new-production'],
  {
    tags: [linterTag],
    revalidate: 600,
  },
);
