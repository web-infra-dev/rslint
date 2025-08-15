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
    const [failing, passing] = await Promise.all([
      kv.get(`${kvPrefix}failing-lint-tests`),
      kv.get(`${kvPrefix}passing-lint-tests`),
    ]);

    if (failing === null && passing === null) {
      return null;
    }

    return { passing, failing };
  },
  [kvPrefix, 'lint-results-new'],
  {
    tags: [linterTag],
    revalidate: 600,
  },
);

export const getProductionLintResults = unstable_cache(
  async () => {
    const [failing, passing] = await Promise.all([
      kv.get(`${kvPrefix}failing-lint-tests-production`),
      kv.get(`${kvPrefix}passing-lint-tests-production`),
    ]);

    if (failing === null && passing === null) {
      return null;
    }

    return { passing, failing };
  },
  [kvPrefix, 'lint-results-new-production'],
  {
    tags: [linterTag],
    revalidate: 600,
  },
);

export const getRuleExamplesResults = unstable_cache(
  async () => {
    const data: { [ruleName: string]: /* isPassing */ boolean } = await kv.get(
      `${kvPrefix}rule-examples-data`,
    );
    return data;
  },
  [kvPrefix, 'rule-examples-results'],
  {
    tags: [linterTag],
    revalidate: 600,
  },
);

export const getDevelopmentLintRuns = unstable_cache(
  async () => {
    const [graphData] = await Promise.all([
      kv.lrange(`${kvPrefix}lint-runs`, 0, -1).then(processGraphData),
    ]);

    const mostRecent = graphData[graphData.length - 1];
    return { graphData, mostRecent };
  },
  [kvPrefix, 'lint-runs-new'],
  {
    tags: [linterTag],
    revalidate: 600,
  },
);

export const getProductionLintRuns = unstable_cache(
  async () => {
    const [graphData] = await Promise.all([
      kv
        .lrange(`${kvPrefix}lint-runs-production`, 0, -1)
        .then(processGraphData),
    ]);

    const mostRecent = graphData[graphData.length - 1];
    return { graphData, mostRecent };
  },
  [kvPrefix, 'lint-runs-new-production'],
  {
    tags: [linterTag],
    revalidate: 600,
  },
);
