import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import { NextIntlClientProvider } from 'next-intl';
import { getMessages } from 'next-intl/server';

import './globals.css';
import { Providers } from '@/components/providers';
import { Toaster } from '@/components/ui/toaster';

const inter = Inter({ 
  subsets: ['latin', 'vietnamese'],
  variable: '--font-sans',
});

export const metadata: Metadata = {
  title: {
    default: 'Captcha Platform',
    template: '%s | Captcha Platform',
  },
  description: 'Nền tảng giải và huấn luyện captcha sử dụng AI',
  keywords: ['captcha', 'ocr', 'ai', 'machine learning', 'solver'],
  authors: [{ name: 'Captcha Platform Team' }],
  viewport: 'width=device-width, initial-scale=1',
  robots: 'index, follow',
};

export default async function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const messages = await getMessages();
  const locale = messages.locale as string || 'vi';

  return (
    <html lang={locale} suppressHydrationWarning>
      <body className={`${inter.variable} font-sans antialiased`}>
        <NextIntlClientProvider messages={messages}>
          <Providers>
            {children}
            <Toaster />
          </Providers>
        </NextIntlClientProvider>
      </body>
    </html>
  );
}