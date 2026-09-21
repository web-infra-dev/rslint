// A Vue single file component is a supported extension outside rslint's
// discovery baseline, so unlike every other plugin's harness this config has to
// name `files` for the tests to select anything at all.
export default [
  {
    files: ['**/*.vue'],
    rules: {},
    plugins: ['vue'],
  },
];
