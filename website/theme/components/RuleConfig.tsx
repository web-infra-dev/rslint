import React from 'react';
import { CodeBlockRuntime } from '@rspress/core/theme';
import { PLUGIN_REGISTRY } from '../plugin-registry';

const GROUP_CONFIG: Record<string, { importName: string; preset: string }> =
  Object.fromEntries(
    PLUGIN_REGISTRY.filter((p) => p.presetName).map((p) => [
      p.group,
      { importName: p.importName, preset: p.presetName! },
    ]),
  );

/**
 * Displays a complete rslint configuration snippet that users can copy
 * for a specific rule. Includes the correct import, preset, and rule override.
 */
export const RuleConfig: React.FC<{ name: string; group: string }> = ({
  name,
  group,
}) => {
  const config = GROUP_CONFIG[group];

  const code = config
    ? `import { defineConfig, ${config.importName} } from '@rslint/core';

export default defineConfig([
  ${config.preset},
  {
    rules: {
      '${name}': 'error',
    },
  },
]);`
    : `import { defineConfig } from '@rslint/core';

export default defineConfig([
  {
    rules: {
      '${name}': 'error',
    },
  },
]);`;

  return <CodeBlockRuntime lang="ts" code={code} title="rslint.config.ts" />;
};

export default RuleConfig;
