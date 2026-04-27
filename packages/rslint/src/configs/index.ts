import type { RslintConfigEntry } from '../define-config.js';
import { recommended as tsRecommended } from './typescript.js';
import { recommended as jsRecommended } from './javascript.js';
import { recommended as reactRecommended } from './react.js';
import { recommended as importRecommended } from './import.js';
import { recommended as promiseRecommended } from './promise.js';
import { recommended as jestRecommended } from './jest.js';
import { recommended as unicornRecommended } from './unicorn.js';

interface PluginExport {
  configs: { recommended: RslintConfigEntry };
}

export const ts: PluginExport = {
  configs: { recommended: tsRecommended },
};

export const js: PluginExport = {
  configs: { recommended: jsRecommended },
};

export const reactPlugin: PluginExport = {
  configs: { recommended: reactRecommended },
};

export const importPlugin: PluginExport = {
  configs: { recommended: importRecommended },
};

export const promisePlugin: PluginExport = {
  configs: { recommended: promiseRecommended },
};

export const jestPlugin: PluginExport = {
  configs: { recommended: jestRecommended },
};

export const unicornPlugin: PluginExport = {
  configs: { recommended: unicornRecommended },
};
