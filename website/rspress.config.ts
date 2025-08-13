import * as path from 'node:path';
import { defineConfig } from '@rspress/core';

export default defineConfig({
  root: path.join(__dirname, 'docs'),
  title: 'Rslint',
  icon: 'https://assets.rspack.rs/rslint/rslint-banner.png',
  logo: {
    light: 'https://assets.rspack.rs/rslint/rslint-logo.svg',
    dark: 'https://assets.rspack.rs/rslint/rslint-logo.svg',
  },
  themeConfig: {
    socialLinks: [
      {
        icon: 'github',
        mode: 'link',
        content: 'https://github.com/web-infra-dev/rslint',
      },
    ],
  },
});
