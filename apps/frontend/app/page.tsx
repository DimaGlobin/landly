import Link from 'next/link'
import { ArrowRight, Sparkles, Zap, ShieldCheck } from 'lucide-react'

import { Button } from '@/components/ui/button'

const benefits = [
  {
    icon: <Sparkles className="h-5 w-5" />,
    title: 'AI-комбинация контента и дизайна',
    description: 'Генерируем hero, преимущества, тарифы и CTA из одного промпта. Можно прорабатывать в диалоге, как в чате.',
  },
  {
    icon: <Zap className="h-5 w-5" />,
    title: 'Публикация в один клик',
    description: 'Статика уходит в CDN, аналитика подключается автоматически. Никаких деплоев и ручных настроек.',
  },
  {
    icon: <ShieldCheck className="h-5 w-5" />,
    title: 'Командный workspace',
    description: 'Проекты сортируются по свежести, чат и предпросмотр синхронизированы. Коллеги подключаются за секунды.',
  },
]

export default function HomePage() {
  return (
    <div className="app-shell text-slate-900">
      <div className="relative z-10 mx-auto flex min-h-screen w-full max-w-6xl flex-col px-6 pb-20 pt-10 md:px-10">
        <header className="flex items-center justify-between gap-6 rounded-full border border-white/50 bg-white/80 px-6 py-4 shadow-xl backdrop-blur-md">
          <Link href="/" className="text-sm font-semibold uppercase tracking-[0.35em] text-blue-600">
            Landly
          </Link>
          <nav className="hidden items-center gap-6 text-sm font-medium text-slate-600 md:flex">
            <a className="transition hover:text-slate-900" href="#workflow">
              Как работает
            </a>
            <a className="transition hover:text-slate-900" href="#features">
              Возможности
            </a>
            <a className="transition hover:text-slate-900" href="#security">
              Статический хостинг
            </a>
          </nav>
          <div className="flex items-center gap-3">
            <Link
              href="/auth/login"
              className="inline-flex h-10 items-center justify-center rounded-full border border-white/60 bg-white/70 px-5 text-sm font-semibold text-slate-700 shadow-sm transition hover:bg-white"
            >
              Войти
            </Link>
            <Link
              href="/auth/signup"
              className="inline-flex h-10 items-center justify-center rounded-full bg-blue-600 px-6 text-sm font-semibold text-white shadow-lg transition hover:bg-blue-700"
            >
              Создать аккаунт
            </Link>
          </div>
        </header>

        <main className="mt-14 flex flex-1 flex-col gap-16">
          <section className="glass-panel overflow-hidden px-8 py-12 md:px-12" id="hero">
            <div className="flex flex-col gap-8 md:flex-row md:items-center md:justify-between">
              <div className="max-w-xl space-y-6">
                <span className="inline-flex items-center gap-2 rounded-full border border-blue-500/30 bg-blue-500/10 px-4 py-2 text-xs font-semibold uppercase tracking-wide text-blue-600">
                  <Sparkles className="h-4 w-4" /> AI-first генерация
                </span>
                <h1 className="text-4xl font-bold leading-tight tracking-tight text-slate-900 md:text-5xl">
                  Создавайте лендинги как в чате. Вдохновляйтесь, правьте и публикуйте за минуты.
                </h1>
                <p className="text-lg text-slate-600">
                  Landly — это workspace, где генерация и предпросмотр работают синхронно. Пишите промпты как в GPT, получайте готовые блоки и мгновенно выкладывайте их на публичный URL.
                </p>
                <div className="flex flex-col gap-3 sm:flex-row">
                  <Link
                    href="/auth/signup"
                    className="inline-flex h-11 items-center justify-center rounded-full bg-blue-600 px-6 text-sm font-semibold text-white shadow-lg transition hover:bg-blue-700"
                  >
                    Начать бесплатно
                    <ArrowRight className="ml-2 h-4 w-4" />
                  </Link>
                  <Link
                    href="/auth/login"
                    className="inline-flex h-11 items-center justify-center rounded-full border border-slate-200 bg-white/70 px-6 text-sm font-semibold text-slate-700 shadow-sm transition hover:bg-white"
                  >
                    Посмотреть workspace
                  </Link>
                </div>
              </div>
              <div className="relative mt-10 flex flex-1 justify-center md:mt-0">
                <div className="surface-card w-full max-w-md overflow-hidden border-white/60 p-6 text-left shadow-2xl">
                  <div className="mb-4 flex items-center justify-between text-xs font-semibold uppercase tracking-wide text-slate-500">
                    <span>AI-чат</span>
                    <span>Предпросмотр</span>
                  </div>
                  <div className="grid gap-4 text-sm text-slate-600">
                    <div className="rounded-2xl bg-blue-600/10 px-4 py-3 text-blue-700">
                      Пользователь: «Сделай лендинг для нового AI-продукта, который автоматизирует рассылки»
                    </div>
                    <div className="rounded-2xl border border-white/70 bg-white/80 px-4 py-3 shadow-sm">
                      Landly AI: «Готово! Посмотрите hero-блок с CTA и секциями цен. Всё можно поправить новым промптом.»
                    </div>
                    <div className="rounded-2xl border border-white/70 bg-white/90 px-4 py-3 shadow-sm">
                      ─ Hero. Заголовок, субтитр, кнопки
                      <br />─ Features. 3 ключевые выгоды
                      <br />─ Pricing. Тарифы с CTA
                      <br />─ FAQ. Автоматически сгенерированные ответы
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </section>

          <section id="workflow" className="grid gap-6 md:grid-cols-3">
            {benefits.map((benefit) => (
              <div key={benefit.title} className="surface-card border-white/40 p-6">
                <div className="flex h-10 w-10 items-center justify-center rounded-full bg-blue-500/10 text-blue-600">
                  {benefit.icon}
                </div>
                <h3 className="mt-4 text-lg font-semibold text-slate-900">{benefit.title}</h3>
                <p className="mt-2 text-sm text-slate-600">{benefit.description}</p>
              </div>
            ))}
          </section>

          <section id="security" className="glass-panel px-8 py-12 md:px-12">
            <div className="grid gap-10 md:grid-cols-2 md:items-center">
              <div className="space-y-4">
                <span className="inline-flex items-center rounded-full border border-emerald-400/40 bg-emerald-500/10 px-4 py-1.5 text-xs font-semibold uppercase tracking-wide text-emerald-600">
                  CDN ready
                </span>
                <h2 className="text-3xl font-bold text-slate-900 md:text-4xl">Публикация — это экспорт статики, а не сложный deploy.</h2>
                <p className="text-sm text-slate-600">
                  Каждая публикация — это статический HTML + CSS, загруженный в S3-совместимое хранилище. Никаких рантаймов, никакой ручной сборки. Предпросмотр и продакшн используют один и тот же шаблон.
                </p>
              </div>
              <div className="surface-card border-white/40 p-6 text-sm text-slate-600">
                <ul className="space-y-4">
                  <li>
                    <span className="font-semibold text-slate-900">MinIO / S3.</span> Хостинг через CDN, мгновенные ссылки, готовы к интеграции с вашим доменом.
                  </li>
                  <li>
                    <span className="font-semibold text-slate-900">Analytics.</span> Отслеживаем просмотры и CTA события через встроенный JS без трекеров третьих лиц.
                  </li>
                  <li>
                    <span className="font-semibold text-slate-900">Редактирование в чат-истории.</span> Каждая итерация хранится на сервере, можно вернуться и откатиться.
                  </li>
                </ul>
              </div>
            </div>
          </section>

          <section className="glass-panel flex flex-col items-center gap-6 px-10 py-12 text-center md:px-16">
            <p className="text-xs font-semibold uppercase tracking-[0.3em] text-blue-600">Landly workspace</p>
            <h2 className="text-3xl font-bold text-slate-900 md:text-4xl">Попробуйте workflow, похожий на Cursor и ChatGPT.</h2>
            <p className="max-w-2xl text-sm text-slate-600">
              Войдите и увидите последний проект сразу — чат, предпросмотр и публикация объединены в одном окне. Никаких форм и бесконечных настроек.
            </p>
            <div className="flex flex-col gap-3 sm:flex-row">
              <Link
                href="/auth/login"
                className="inline-flex h-11 items-center justify-center rounded-full bg-blue-600 px-6 text-sm font-semibold text-white shadow-lg transition hover:bg-blue-700"
              >
                Войти
              </Link>
              <Link
                href="/auth/signup"
                className="inline-flex h-11 items-center justify-center rounded-full border border-slate-200 bg-white/80 px-6 text-sm font-semibold text-slate-700 shadow-sm transition hover:bg-white"
              >
                Создать аккаунт
              </Link>
            </div>
          </section>
        </main>

        <footer className="mt-16 border-t border-white/40 py-6 text-center text-xs text-slate-500">
          © {new Date().getFullYear()} Landly. Сделано с заботой о продуктах, которые хочется запускать быстро.
        </footer>
      </div>
    </div>
  )
}

