import path from 'node:path';
import { pluginNodePolyfill } from '@rsbuild/plugin-node-polyfill';
import { pluginSass } from '@rsbuild/plugin-sass';
import { pluginSitemap } from '@rspress/plugin-sitemap';
import { define } from 'rstack';
import { pluginGoogleAnalytics } from 'rsbuild-plugin-google-analytics';
import { pluginOpenGraph } from 'rsbuild-plugin-open-graph';
import { pluginFontOpenSans } from 'rspress-plugin-font-open-sans';
import { pluginRuleManifest } from './plugin-rule-manifest.ts';

const siteUrl = 'https://rslint.rs';
const logo = 'https://assets.rspack.rs/rslint/rslint-logo.svg';
const description = 'The high-performance TypeScript linter';

define.doc({
  title: 'Rslint',
  icon: logo,
  logo,
  logoText: 'Rslint',
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
  globalStyles: path.join(import.meta.dirname, 'styles/index.css'),
  llms: true,
  markdown: {
    link: {
      checkAnchors: true,
    },
  },
  themeConfig: {
    llmsUI: {
      placement: 'outline',
    },
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
    pluginRuleManifest(),
    pluginFontOpenSans(),
    pluginSitemap({ siteUrl }),
  ],
  builderConfig: {
    tools: {
      rspack(config) {
        config.ignoreWarnings = [
          {
            module: /(editorSimpleWorker|typescript)\.js/,
          },
        ];
      },
    },
    plugins: [
      pluginNodePolyfill({
        include: ['buffer'],
        globals: {
          Buffer: false,
          process: false,
        },
      }),
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
