import * as path from 'node:path';
import { pluginSass } from '@rsbuild/plugin-sass';
import { defineConfig } from '@rspress/core';
import { pluginLlms } from '@rspress/plugin-llms';
import { pluginGoogleAnalytics } from 'rsbuild-plugin-google-analytics';
import { pluginOpenGraph } from 'rsbuild-plugin-open-graph';
import { pluginFontOpenSans } from 'rspress-plugin-font-open-sans';
import pluginSitemap from 'rspress-plugin-sitemap';
import { pluginPreloadRule } from './preload-plugin';

const siteUrl = 'https://rslint.rs';
const description = 'The Rspack-based testing framework';

export default defineConfig({
  root: path.join(__dirname, 'docs'),
  title: 'Rslint',
  icon: 'https://assets.rspack.rs/rslint/rslint-logo.svg',
  logo: {
    light: 'https://assets.rspack.rs/rslint/rslint-logo.svg',
    dark: 'https://assets.rspack.rs/rslint/rslint-logo.svg',
  },
  logoText: 'Rslint',
  search: {
    codeBlocks: true,
  },
  lang: 'en',
  locales: [
    {
      lang: 'en',
      label: 'English',
      description,
    },
  ],
  route: {
    cleanUrls: true,
    // exclude document fragments from routes
    exclude: ['**/zh/shared/**', '**/en/shared/**', './theme'],
  },
  themeConfig: {
    socialLinks: [
      {
        icon: 'github',
        mode: 'link',
        content: 'https://github.com/web-infra-dev/rslint',
      },
      {
        icon: 'x',
        mode: 'link',
        content: 'https://twitter.com/rspack_dev',
      },
      {
        icon: 'discord',
        mode: 'link',
        content: 'https://discord.gg/XsaKEEk4mW',
      },
    ],
  },
  plugins: [
    pluginPreloadRule(),
    pluginFontOpenSans(),
    pluginSitemap({
      domain: siteUrl,
    }),
    pluginLlms(),
  ],
  builderConfig: {
    plugins: [
      pluginSass(),
      pluginGoogleAnalytics({
        // cspell:disable-next-line
        id: 'G-9WKFF5YJXQ',
      }),
      pluginOpenGraph({
        title: 'Rslint',
        type: 'website',
        url: siteUrl,
        image: 'https://assets.rspack.rs/rslint/rslint-logo.svg',
        description,
        twitter: {
          site: '@rspack_dev',
          card: 'summary_large_image',
        },
      }),
    ],
  },
});
