import * as path from 'node:path';
import { defineConfig } from 'rspress/config';

export default defineConfig({
  root: path.join(__dirname, 'docs'),
  title: 'Harlequin Toolkit',
  description: 'A comprehensive toolkit for building on the Permaweb',
  icon: '/harlequin_mascot.png',
  logo: {
    light: '/harlequin_mascot.png',
    dark: '/harlequin_mascot_dark.png',
  },
  logoText: 'Harlequin Toolkit',
  base: '.',
  outDir: path.join(__dirname, 'dist'),
  globalStyles: path.join(__dirname, './styles/globals.css'),
  builderConfig: {
    output: {
      assetPrefix: '/public/',
    },
  },
  themeConfig: {
    socialLinks: [
      {
        icon: 'github',
        mode: 'link',
        content: 'https://github.com/the-permaweb-harlequin/harlequin-toolkit',
      },
    ],
    nav: [
      {
        text: 'CLI',
        link: '/cli/',
      },
    ],
    sidebar: {
      '/cli/': [
        {
          text: 'CLI Documentation',
          items: [
            {
              text: 'Overview',
              link: '/cli/',
            },
            {
              text: 'Installation',
              link: '/cli/installation',
            },
            {
              text: 'Commands',
              link: '/cli/commands',
              items: ['/cli/commands/build'],
            },
          ],
        },
      ],
    },
    footer: {
      message: 'Built with RSPress',
    },
  },
});
