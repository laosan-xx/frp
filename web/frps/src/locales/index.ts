import { createI18n } from 'vue-i18n'
import en from './en'
import zhCN from './zh-CN'

const LOCALE_KEY = 'frps-locale'

function getDefaultLocale(): string {
  const saved = localStorage.getItem(LOCALE_KEY)
  if (saved && ['en', 'zh-CN'].includes(saved)) {
    return saved
  }
  // Detect browser language
  const lang = navigator.language
  if (lang.startsWith('zh')) {
    return 'zh-CN'
  }
  return 'en'
}

const i18n = createI18n({
  legacy: false,
  locale: getDefaultLocale(),
  fallbackLocale: 'en',
  messages: {
    en,
    'zh-CN': zhCN,
  },
})

export function setLocale(locale: string) {
  ;(i18n.global.locale as unknown as { value: string }).value = locale
  localStorage.setItem(LOCALE_KEY, locale)
}

export function getLocale(): string {
  return (i18n.global.locale as unknown as { value: string }).value
}

export default i18n
