import tailwindcss from '@tailwindcss/vite'

const themeInitScript = `(function(){try{var m=document.cookie.match(/(?:^|; )mit_theme=([^;]+)/);var t=m?decodeURIComponent(m[1]):localStorage.getItem('mit_theme');if(t==='dark'||(!t&&window.matchMedia('(prefers-color-scheme: dark)').matches)){document.documentElement.classList.add('dark');}}catch(e){}})();`

export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  css: ['~/assets/css/main.css'],
  runtimeConfig: {
    public: {
      apiBase: 'http://localhost/api',
      routerBase: 'http://localhost/r'
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
      script: [
        {
          innerHTML: themeInitScript,
          tagPosition: 'head'
        }
      ],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
        {
          rel: 'stylesheet',
          href: 'https://fonts.googleapis.com/css2?family=Anybody:wdth,wght@50..150,300..900&family=Hanken+Grotesk:wght@400..800&family=JetBrains+Mono:wght@400..700&display=swap'
        }
      ]
    }
  }
})
