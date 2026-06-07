// Fixture cfgB: the SAME local plugin module loaded under a DIFFERENT
// prefix `pkgB`. Tasks with configKey=cfgB must resolve `pkgB/*` rules.
// A worker that confuses the two configs would either return
// "rule not found" for pkgB/no-null or — worse — silently route to
// pkgA's plugin.
import localPlugin from '../local-plugin.mjs';

export default [
  {
    plugins: {
      pkgB: localPlugin,
    },
  },
];
