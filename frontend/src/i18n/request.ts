import { getRequestConfig } from 'next-intl/server';
import { cookies, headers } from 'next/headers';

// Supported locales
export const locales = ['vi', 'en'] as const;
export type Locale = (typeof locales)[number];
export const defaultLocale: Locale = 'vi';

// Get locale from cookie or header
async function getLocale(): Promise<Locale> {
  // Try to get from cookie first
  const cookieStore = await cookies();
  const localeCookie = cookieStore.get('NEXT_LOCALE');
  if (localeCookie && locales.includes(localeCookie.value as Locale)) {
    return localeCookie.value as Locale;
  }

  // Try to get from Accept-Language header
  const headerStore = await headers();
  const acceptLanguage = headerStore.get('Accept-Language');
  if (acceptLanguage) {
    const preferredLocale = acceptLanguage
      .split(',')
      .map(lang => lang.split(';')[0].trim().substring(0, 2))
      .find(lang => locales.includes(lang as Locale));
    if (preferredLocale) {
      return preferredLocale as Locale;
    }
  }

  return defaultLocale;
}

export default getRequestConfig(async () => {
  const locale = await getLocale();

  return {
    locale,
    messages: (await import(`./locales/${locale}.json`)).default,
  };
});