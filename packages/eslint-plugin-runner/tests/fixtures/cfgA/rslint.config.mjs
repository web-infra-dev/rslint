// Fixture cfgA: the local plugin loaded under prefix `pkgA`.
// Paired with cfgB to test that the worker dispatches a file with
// configKey=cfgA only against pkgA's plugin set.
import localPlugin from '../local-plugin.mjs';

export default [
  {
    eslintPlugins: {
      pkgA: localPlugin,
    },
  },
];
