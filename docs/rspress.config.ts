import * as path from 'node:path';
import { defineConfig } from 'rspress/config';

export default defineConfig({
  root: path.join(__dirname, 'docs'),
  title: 'Harlequin Toolkit',
  description: 'A comprehensive toolkit for building on the Permaweb',
  icon: '/rspress-icon.png',
  logo: {
    light: '/rspress-light-logo.png',
    dark: '/rspress-dark-logo.png',
  },
  base: '/',
  outDir: 'dist',
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
        text: 'Guide',
        link: '/guide/',
      },
      {
        text: 'CLI',
        link: '/cli/',
      },
      {
        text: 'SDK',
        link: '/sdk/',
      },
      {
        text: 'App',
        link: '/app/',
      },
    ],
    sidebar: {
      '/guide/': [
        {
          text: 'Getting Started',
          items: [
            {
              text: 'Introduction',
              link: '/guide/',
            },
            {
              text: 'Installation',
              link: '/guide/installation',
            },
          ],
        },
      ],
      '/cli/': [
        {
          text: 'CLI Documentation',
          items: [
            {
              text: 'Overview',
              link: '/cli/',
            },
            {
              text: 'Commands',
              link: '/cli/commands',
            },
          ],
        },
      ],
      '/sdk/': [
        {
          text: 'SDK Documentation',
          items: [
            {
              text: 'Overview',
              link: '/sdk/',
            },
            {
              text: 'API Reference',
              link: '/sdk/api',
            },
          ],
        },
      ],
      '/app/': [
        {
          text: 'App Documentation',
          items: [
            {
              text: 'Overview',
              link: '/app/',
            },
            {
              text: 'Components',
              link: '/app/components',
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
