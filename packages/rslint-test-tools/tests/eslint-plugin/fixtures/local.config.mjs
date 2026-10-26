import localPlugin from './local-plugin.mjs';

export default [
  {
    plugins: {
      local: localPlugin,
    },
    rules: {
      'local/program-listener': 'error',
    },
  },
];
