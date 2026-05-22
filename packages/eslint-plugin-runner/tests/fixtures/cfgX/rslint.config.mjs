// Fixture cfgX: only plugin-x loaded under prefix `px`.
// Paired with cfgY (different prefix `py` AND a different plugin
// module). The worker's per-config LoadedPlugins map must keep them
// disjoint — a task with configKey=cfgX must not resolve any `py/*`
// rule, and vice versa.
import pluginX from './plugin-x.mjs';

export default [
  {
    eslintPlugins: {
      px: pluginX,
    },
  },
];
