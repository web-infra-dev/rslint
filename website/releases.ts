import releases from '../releases.json' with { type: 'json' };

export default releases;

// Prerelease compiler bindings share the release history, but rule badges
// continue to identify the first stable release that supports each rule.
export const ruleVersionById = new Map(
  releases
    .filter(({ version }) => /^\d+\.\d+\.\d+$/.test(version))
    .flatMap(({ version, rules }) =>
      rules.map((rule) => [rule, version] as const),
    ),
);
