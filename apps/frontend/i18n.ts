import { getRequestConfig } from 'next-intl/server'
import { cookies, headers } from 'next/headers'

export const locales = ['en', 'ru'] as const
export type Locale = (typeof locales)[number]
export const defaultLocale: Locale = 'ru'

export default getRequestConfig(async () => {
  // Try to get locale from cookie first, then Accept-Language header
  const cookieStore = await cookies()
  const headerStore = await headers()
  
  let locale: Locale = defaultLocale
  
  // Check cookie
  const cookieLocale = cookieStore.get('locale')?.value as Locale | undefined
  if (cookieLocale && locales.includes(cookieLocale)) {
    locale = cookieLocale
  } else {
    // Check Accept-Language header
    const acceptLanguage = headerStore.get('accept-language')
    if (acceptLanguage) {
      const preferredLocale = acceptLanguage.split(',')[0]?.split('-')[0] as Locale
      if (preferredLocale && locales.includes(preferredLocale)) {
        locale = preferredLocale
      }
    }
  }

  return {
    locale,
    messages: (await import(`./messages/${locale}.json`)).default,
  }
})

