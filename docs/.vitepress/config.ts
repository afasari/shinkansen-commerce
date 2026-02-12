import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Shinkansen Commerce',
  description: 'High-performance e-commerce platform for Japanese market with microservices architecture',
  lang: 'en',
  base: '/shinkansen-commerce/',
  ignoreDeadLinks: true,
  
  head: [
    ['meta', { name: 'theme-color', content: '#3c8772' }],
    ['meta', { name: 'description', content: 'High-performance e-commerce platform for Japanese market with microservices architecture' }],
    ['meta', { name: 'keywords', content: 'e-commerce, microservices, gRPC, Go, Rust, Python, Kubernetes, DevOps' }],
    ['meta', { name: 'author', content: 'afasari' }],
    ['link', { rel: 'icon', href: '/favicon.ico' }],
  ],

  themeConfig: {
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Architecture', link: '/architecture/overview' },
      { text: 'API', link: '/api/overview' },
      { text: 'Development', link: '/development/setup' },
      { text: 'Runbooks', link: '/runbooks/deployment' },
    ],

    sidebar: {
      '/': [
        { text: 'Quick Start', link: '/quickstart' },
        { text: 'Introduction', link: '/introduction' },
        { text: 'Tech Stack', link: '/tech-stack' },
      ],
    },

    search: {
      provider: 'local',
    },

    editLink: {
      pattern: 'https://github.com/afasari/shinkansen-commerce/edit/main/docs/:path',
    },

    lastUpdated: {
      text: 'Last updated',
      formatOptions: {
        dateStyle: 'full',
        timeStyle: 'short'
      }
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/afasari/shinkansen-commerce' }
    ],

    footer: {
      message: 'Released under MIT License.',
      copyright: 'Copyright © 2024-present Shinkansen Commerce'
    }
  }
})
