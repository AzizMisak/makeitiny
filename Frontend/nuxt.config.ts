import tailwindcss from '@tailwindcss/vite'

export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  css: ['~/assets/css/main.css'],
  runtimeConfig: {
    public: {
      apiBase: 'http://localhost:8080',
      routerBase: 'http://localhost:8081'
    }
  },
  vite: {
    plugins: [tailwindcss()]
  },
  app: {
    head: {
      htmlAttrs: { lang: 'tr' },
      title: 'makeiTiny — uzun linkleri küçült',
      meta: [
        {
          name: 'description',
          content:
            'makeiTiny ile uzun bağlantıları kısa, hatırlanabilir adreslere dönüştür. Tıklamaları izle, linkleri şifrele, istediğin an düzenle.'
        }
      ],
      link: [
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
        {
          rel: 'stylesheet',
          href: 'https://fonts.googleapis.com/css2?family=Bricolage+Grotesque:opsz,wght@12..96,500;12..96,700;12..96,800&family=Instrument+Sans:wght@400;500;600&family=JetBrains+Mono:wght@400;500&display=swap'
        }
      ]
    }
  }
})
