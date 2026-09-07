import data from './releases.json' with { type: 'json' };

interface Release {
  version: string;
  rules: string[];
  typescript?: {
    commit: string;
    releaseVersion: string | null;
  };
}

const releases: Release[] = data;

export default releases;

export const ruleVersionById = new Map(
  releases.flatMap(({ version, rules }) =>
    rules.map((rule) => [rule, version] as const),
  ),
);
