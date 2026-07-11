import localPlugin from './local-plugin.mjs';

export default [
  {
    plugins: { local: localPlugin },
    // normalizeConfig preserves nested parser options; Go's typed config
    // decoder rejects this project value only after the host has initialized.
    languageOptions: { parserOptions: { project: 42 } },
    rules: { 'local/program-listener': 'error' },
  },
];
