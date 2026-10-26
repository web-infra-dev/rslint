import localPlugin from './local-plugin.mjs';

export default [
  {
    plugins: { local: localPlugin },
    // Keep the discovery payload structurally valid so Go can activate the
    // community-plugin host before program creation fails on this project.
    languageOptions: {
      parserOptions: { project: ['./missing-tsconfig.json'] },
    },
    rules: { 'local/program-listener': 'error' },
  },
];
