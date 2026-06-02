// Fixture cfgY: only plugin-y loaded under prefix `py`. See cfgX for
// the disjoint-plugin-sets dispatch test that exercises both.
import pluginY from './plugin-y.mjs';

export default [
  {
    eslintPlugins: {
      py: pluginY,
    },
  },
];
