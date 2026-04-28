import React from 'react';
import { PLUGIN_REGISTRY } from '../plugin-registry';

const PREFIXED_PLUGINS = PLUGIN_REGISTRY.filter((p) => p.prefix !== '');

/**
 * Renders the table of plugin name → rule key prefix used in the
 * `plugins` configuration field. Core ESLint rules (no prefix) are
 * intentionally excluded.
 */
export const PluginPrefixTable: React.FC = () => (
  <table>
    <thead>
      <tr>
        <th>Plugin</th>
        <th>Rules Prefix</th>
      </tr>
    </thead>
    <tbody>
      {PREFIXED_PLUGINS.map((p) => (
        <tr key={p.prefix}>
          <td>
            <code>{p.prefix}</code>
          </td>
          <td>
            <code>{p.prefix}/*</code>
          </td>
        </tr>
      ))}
    </tbody>
  </table>
);

export default PluginPrefixTable;
