import localPlugin from './local-plugin.mjs';

export default [
  {
    eslintPlugins: {
      local: localPlugin,
    },
  },
];
