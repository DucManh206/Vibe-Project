import { useTranslations } from 'next-intl';
import Link from 'next/link';
import { ArrowRight, Shield, Zap, Brain, Code } from 'lucide-react';

import { Button } from '@/components/ui/button';

export default function HomePage() {
  const t = useTranslations();

  return (
    <div className="min-h-screen bg-gradient-to-b from-background to-muted">
      {/* Header */}
      <header className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container flex h-16 items-center justify-between">
          <div className="flex items-center gap-2">
            <Shield className="h-8 w-8 text-primary" />
            <span className="text-xl font-bold">{t('common.appName')}</span>
          </div>
          <nav className="flex items-center gap-4">
            <Link href="/login">
              <Button variant="ghost">{t('auth.login')}</Button>
            </Link>
            <Link href="/register">
              <Button>{t('auth.register')}</Button>
            </Link>
          </nav>
        </div>
      </header>

      {/* Hero Section */}
      <section className="container py-24 md:py-32">
        <div className="mx-auto max-w-3xl text-center">
          <h1 className="text-4xl font-bold tracking-tight sm:text-5xl md:text-6xl">
            {t('landing.hero.title')}{' '}
            <span className="text-primary">AI</span>
          </h1>
          <p className="mt-6 text-lg text-muted-foreground">
            {t('landing.hero.subtitle')}
          </p>
          <div className="mt-10 flex flex-col gap-4 sm:flex-row sm:justify-center">
            <Link href="/register">
              <Button size="lg" className="w-full sm:w-auto">
                {t('landing.hero.getStarted')}
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </Link>
            <Link href="/docs">
              <Button size="lg" variant="outline" className="w-full sm:w-auto">
                {t('landing.hero.viewDocs')}
              </Button>
            </Link>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="container py-24">
        <div className="mx-auto max-w-2xl text-center">
          <h2 className="text-3xl font-bold tracking-tight">
            {t('landing.features.title')}
          </h2>
          <p className="mt-4 text-muted-foreground">
            {t('landing.features.subtitle')}
          </p>
        </div>

        <div className="mt-16 grid gap-8 md:grid-cols-2 lg:grid-cols-3">
          {/* Feature 1 */}
          <div className="rounded-lg border bg-card p-8">
            <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
              <Zap className="h-6 w-6 text-primary" />
            </div>
            <h3 className="mt-4 text-xl font-semibold">{t('landing.features.speed.title')}</h3>
            <p className="mt-2 text-muted-foreground">
              {t('landing.features.speed.description')}
            </p>
          </div>

          {/* Feature 2 */}
          <div className="rounded-lg border bg-card p-8">
            <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
              <Brain className="h-6 w-6 text-primary" />
            </div>
            <h3 className="mt-4 text-xl font-semibold">{t('landing.features.ai.title')}</h3>
            <p className="mt-2 text-muted-foreground">
              {t('landing.features.ai.description')}
            </p>
          </div>

          {/* Feature 3 */}
          <div className="rounded-lg border bg-card p-8">
            <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
              <Code className="h-6 w-6 text-primary" />
            </div>
            <h3 className="mt-4 text-xl font-semibold">{t('landing.features.api.title')}</h3>
            <p className="mt-2 text-muted-foreground">
              {t('landing.features.api.description')}
            </p>
          </div>

          {/* Feature 4 */}
          <div className="rounded-lg border bg-card p-8">
            <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
              <Shield className="h-6 w-6 text-primary" />
            </div>
            <h3 className="mt-4 text-xl font-semibold">{t('landing.features.security.title')}</h3>
            <p className="mt-2 text-muted-foreground">
              {t('landing.features.security.description')}
            </p>
          </div>

          {/* Feature 5 */}
          <div className="rounded-lg border bg-card p-8">
            <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
              <svg className="h-6 w-6 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            </div>
            <h3 className="mt-4 text-xl font-semibold">{t('landing.features.stats.title')}</h3>
            <p className="mt-2 text-muted-foreground">
              {t('landing.features.stats.description')}
            </p>
          </div>

          {/* Feature 6 */}
          <div className="rounded-lg border bg-card p-8">
            <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
              <svg className="h-6 w-6 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
            </div>
            <h3 className="mt-4 text-xl font-semibold">{t('landing.features.training.title')}</h3>
            <p className="mt-2 text-muted-foreground">
              {t('landing.features.training.description')}
            </p>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="border-t bg-muted/50">
        <div className="container py-24">
          <div className="mx-auto max-w-2xl text-center">
            <h2 className="text-3xl font-bold tracking-tight">
              {t('landing.cta.title')}
            </h2>
            <p className="mt-4 text-muted-foreground">
              {t('landing.cta.subtitle')}
            </p>
            <div className="mt-8">
              <Link href="/register">
                <Button size="lg">
                  {t('landing.cta.createAccount')}
                  <ArrowRight className="ml-2 h-4 w-4" />
                </Button>
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t bg-background">
        <div className="container py-12">
          <div className="flex flex-col items-center justify-between gap-4 md:flex-row">
            <div className="flex items-center gap-2">
              <Shield className="h-6 w-6 text-primary" />
              <span className="font-semibold">{t('common.appName')}</span>
            </div>
            <p className="text-sm text-muted-foreground">
              {t('footer.copyright')}
            </p>
            <div className="flex gap-4">
              <Link href="/privacy" className="text-sm text-muted-foreground hover:underline">
                {t('footer.privacy')}
              </Link>
              <Link href="/terms" className="text-sm text-muted-foreground hover:underline">
                {t('footer.terms')}
              </Link>
              <Link href="/contact" className="text-sm text-muted-foreground hover:underline">
                {t('footer.contact')}
              </Link>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}