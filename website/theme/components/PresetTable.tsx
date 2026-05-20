import React from 'react';
import { CodeBlockRuntime, Link } from '@rspress/core/theme';
import { expandPluginPresets } from '../plugin-registry';

const PRESET_PLUGINS = expandPluginPresets();
console.log('PRESET_PLUGINS::', PRESET_PLUGINS);

/**
 * Renders the table of every recommended preset shipped by `@rslint/core`,
 * each row linking to the rules explorer filtered by that preset.
 */
export const PresetTable: React.FC = () => (
  <table>
    <thead>
      <tr>
        <th>Preset</th>
        <th>Description</th>
        <th />
      </tr>
    </thead>
    <tbody>
      {PRESET_PLUGINS.map((p) => (
        <tr key={p.name}>
          <td>
            <code>{p.name}</code>
          </td>
          <td>{p.description}</td>
          <td>
            <Link href={`/rules/?preset=${p.name}`}>View rules →</Link>
          </td>
        </tr>
      ))}
    </tbody>
  </table>
);

/**
 * Renders an `import { … } from '@rslint/core'` snippet that pulls in
 * `defineConfig` and every preset export listed in {@link PLUGIN_REGISTRY}.
 */
export const PresetImportSnippet: React.FC = () => {
  const importNames = [...new Set(PRESET_PLUGINS.map((p) => p.importName))];
  const names = ['defineConfig', ...importNames];
  const code = `import {\n${names.map((n) => `  ${n},`).join('\n')}\n} from '@rslint/core';\n`;
  return <CodeBlockRuntime lang="ts" code={code} />;
};

export default PresetTable;
