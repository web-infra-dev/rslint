import React from 'react';
import { CodeBlockRuntime } from '@rspress/core/theme';

/**
 * Mapping from plugin group to the import name and preset used in rslint config.
 * Core eslint rules use `js`, everything else maps to its plugin export.
 */
const GROUP_CONFIG: Record<string, { importName: string; preset: string }> = {
  eslint: { importName: 'js', preset: 'js.configs.recommended' },
  '@typescript-eslint': { importName: 'ts', preset: 'ts.configs.recommended' },
  'eslint-plugin-import': {
    importName: 'importPlugin',
    preset: 'importPlugin.configs.recommended',
  },
  react: {
    importName: 'reactPlugin',
    preset: 'reactPlugin.configs.recommended',
  },
};

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
