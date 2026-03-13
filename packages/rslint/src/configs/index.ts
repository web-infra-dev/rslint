import { recommended as tsRecommended } from './typescript.js';
import { recommended as jsRecommended } from './javascript.js';
import { recommended as reactRecommended } from './react.js';
import { recommended as importRecommended } from './import.js';

const configs = {
  ts: { recommended: tsRecommended },
  js: { recommended: jsRecommended },
  react: { recommended: reactRecommended },
  import: { recommended: importRecommended },
};

export default configs;
