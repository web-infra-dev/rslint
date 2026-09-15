import type { RslintConfigEntry } from '../define-config.js';

// Aligned with eslint-plugin-vue@10's `flat/essential` preset.
// Rules commented out with "not implemented" are in the official preset but not
// yet available.
//
// Unlike every other preset here, this one has to name `files`: `.vue` is a
// supported extension but deliberately not part of rslint's discovery
// baseline, so a component is linted only because a config selects it.
const essential: RslintConfigEntry = {
  files: ['**/*.vue'],
  plugins: ['vue'],
  rules: {
    // 'vue/no-arrow-functions-in-watch': 'error', // not implemented
    // 'vue/no-async-in-computed-properties': 'error', // not implemented
    // 'vue/no-child-content': 'error', // not implemented
    // 'vue/no-computed-properties-in-data': 'error', // not implemented
    // 'vue/no-deprecated-data-object-declaration': 'error', // not implemented
    // 'vue/no-dupe-keys': 'error', // not implemented
    // 'vue/no-dupe-v-else-if': 'error', // not implemented
    'vue/no-duplicate-attributes': 'error',
    'vue/no-export-in-script-setup': 'error',
    // 'vue/no-expose-after-await': 'error', // not implemented
    // 'vue/no-lifecycle-after-await': 'error', // not implemented
    // 'vue/no-mutating-props': 'error', // not implemented
    // 'vue/no-parsing-error': 'error', // not implemented
    // 'vue/no-reserved-keys': 'error', // not implemented
    // 'vue/no-side-effects-in-computed-properties': 'error', // not implemented
    // 'vue/no-template-key': 'error', // not implemented
    // 'vue/no-textarea-mustache': 'error', // not implemented
    // 'vue/no-unused-components': 'error', // not implemented
    // 'vue/no-unused-vars': 'error', // not implemented
    // 'vue/no-use-v-if-with-v-for': 'error', // not implemented
    // 'vue/no-useless-template-attributes': 'error', // not implemented
    // 'vue/no-v-text-v-html-on-component': 'error', // not implemented
    // 'vue/no-watch-after-await': 'error', // not implemented
    // 'vue/require-component-is': 'error', // not implemented
    // 'vue/require-prop-type-constructor': 'error', // not implemented
    // 'vue/require-render-return': 'error', // not implemented
    // 'vue/require-v-for-key': 'error', // not implemented
    // 'vue/require-valid-default-prop': 'error', // not implemented
    // 'vue/return-in-computed-property': 'error', // not implemented
    // 'vue/use-v-on-exact': 'error', // not implemented
    // 'vue/valid-template-root': 'error', // not implemented
    // 'vue/valid-v-bind': 'error', // not implemented
    // 'vue/valid-v-for': 'error', // not implemented
    // 'vue/valid-v-if': 'error', // not implemented
    // 'vue/valid-v-model': 'error', // not implemented
    // 'vue/valid-v-slot': 'error', // not implemented
  },
};

export { essential };
