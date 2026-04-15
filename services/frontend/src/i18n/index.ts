import { createI18n } from 'vue-i18n'
import en from './en.json'
import ja from './ja.json'

const i18n = createI18n({
  legacy: false,
  locale: localStorage.getItem('locale') || 'en',
  fallbackLocale: 'en',
  messages: { en, ja },
})

export default i18n
