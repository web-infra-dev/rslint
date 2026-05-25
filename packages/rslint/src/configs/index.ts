import { base as tsBase, recommended as tsRecommended } from './typescript.js';
import { recommended as jsRecommended } from './javascript.js';
import { recommended as reactRecommended } from './react.js';
import { recommended as reactHooksRecommended } from './react-hooks.js';
import { recommended as importRecommended } from './import.js';
import { recommended as promiseRecommended } from './promise.js';
import { recommended as jestRecommended, style as jestStyle } from './jest.js';
import { recommended as unicornRecommended } from './unicorn.js';
import { recommended as jsxA11yRecommended } from './jsx-a11y.js';
import { recommended as stylisticRecommended } from './stylistic.js';

export const ts = {
  configs: { base: tsBase, recommended: tsRecommended },
};

export const js = {
  configs: { recommended: jsRecommended },
};

export const reactPlugin = {
  configs: { recommended: reactRecommended },
};

export const reactHooksPlugin = {
  configs: { recommended: reactHooksRecommended },
};

export const importPlugin = {
  configs: { recommended: importRecommended },
};

export const promisePlugin = {
  configs: { recommended: promiseRecommended },
};

export const jestPlugin = {
  configs: { recommended: jestRecommended, style: jestStyle },
};

export const unicornPlugin = {
  configs: { recommended: unicornRecommended },
};

export const jsxA11yPlugin = {
  configs: { recommended: jsxA11yRecommended },
};

export const stylisticPlugin = {
  configs: { recommended: stylisticRecommended },
};
