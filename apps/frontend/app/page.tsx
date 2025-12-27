'use client'

import Link from 'next/link'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { ArrowRight, Sparkles, Zap, ShieldCheck } from 'lucide-react'

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
  const router = useRouter()
  const [isChecking, setIsChecking] = useState(true)

  // Check if user is already logged in and redirect
  useEffect(() => {
    const accessToken = localStorage.getItem('access_token')
    if (accessToken) {
      router.replace('/app/projects')
    } else {
      setIsChecking(false)
    }
  }, [router])

  // Show loading while checking auth
  if (isChecking) {
    return (
      <div className="app-shell min-h-screen flex items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent" />
      </div>
    )
  }

  return (
    <div className="app-shell text-slate-900">
      <div className="relative z-10 mx-auto flex min-h-screen w-full max-w-6xl flex-col px-6 pb-20 pt-10 md:px-10">
        <header className="flex items-center justify-between gap-3 rounded-full border border-white/50 bg-white/80 px-4 py-3 shadow-xl backdrop-blur-md sm:gap-6 sm:px-6 sm:py-4">
          <Link href="/" className="text-xs font-semibold uppercase tracking-[0.35em] text-blue-600 sm:text-sm">
            Landly
          </Link>
          <nav className="hidden items-center gap-4 text-xs font-medium text-slate-600 sm:gap-6 sm:text-sm md:flex">
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
          <div className="flex items-center gap-2 sm:gap-3">
            <Link
              href="/auth/login"
              className="inline-flex h-9 items-center justify-center rounded-full border border-white/60 bg-white/70 px-3 text-xs font-semibold text-slate-700 shadow-sm transition hover:bg-white sm:h-10 sm:px-5 sm:text-sm"
            >
              Войти
            </Link>
            <Link
              href="/auth/signup"
              className="inline-flex h-9 items-center justify-center rounded-full bg-blue-600 px-4 text-xs font-semibold text-white shadow-lg transition hover:bg-blue-700 sm:h-10 sm:px-6 sm:text-sm"
            >
              <span className="hidden sm:inline">Создать аккаунт</span>
              <span className="sm:hidden">Регистр.</span>
            </Link>
          </div>
        </header>

        <main className="mt-14 flex flex-1 flex-col gap-16">
          <section className="glass-panel overflow-hidden px-4 py-8 sm:px-8 sm:py-12 md:px-12" id="hero">
            <div className="flex flex-col gap-6 sm:gap-8 md:flex-row md:items-center md:justify-between">
              <div className="max-w-xl space-y-4 sm:space-y-6">
                <span className="inline-flex items-center gap-1.5 rounded-full border border-blue-500/30 bg-blue-500/10 px-3 py-1.5 text-[10px] font-semibold uppercase tracking-wide text-blue-600 sm:gap-2 sm:px-4 sm:py-2 sm:text-xs">
                  <Sparkles className="h-3.5 w-3.5 sm:h-4 sm:w-4" /> AI-first генерация
                </span>
                <h1 className="text-2xl font-bold leading-tight tracking-tight text-slate-900 sm:text-3xl md:text-4xl lg:text-5xl">
                  Создавайте лендинги как в чате. Вдохновляйтесь, правьте и публикуйте за минуты.
                </h1>
                <p className="text-sm text-slate-600 sm:text-base md:text-lg">
                  Landly — это workspace, где генерация и предпросмотр работают синхронно. Пишите промпты как в GPT, получайте готовые блоки и мгновенно выкладывайте их на публичный URL.
                </p>
                <div className="flex flex-col gap-2 sm:flex-row sm:gap-3">
                  <Link
                    href="/auth/signup"
                    className="inline-flex h-10 items-center justify-center rounded-full bg-blue-600 px-5 text-xs font-semibold text-white shadow-lg transition hover:bg-blue-700 sm:h-11 sm:px-6 sm:text-sm"
                  >
                    Начать бесплатно
                    <ArrowRight className="ml-1.5 h-3.5 w-3.5 sm:ml-2 sm:h-4 sm:w-4" />
                  </Link>
                  <Link
                    href="/auth/login"
                    className="inline-flex h-10 items-center justify-center rounded-full border border-slate-200 bg-white/70 px-5 text-xs font-semibold text-slate-700 shadow-sm transition hover:bg-white sm:h-11 sm:px-6 sm:text-sm"
                  >
                    <span className="hidden sm:inline">Посмотреть workspace</span>
                    <span className="sm:hidden">Workspace</span>
                  </Link>
                </div>
              </div>
              <div className="relative mt-6 flex flex-1 justify-center sm:mt-10 md:mt-0">
                <div className="surface-card w-full max-w-md overflow-hidden border-white/60 p-4 text-left shadow-2xl sm:p-6">
                  <div className="mb-3 flex items-center justify-between text-[10px] font-semibold uppercase tracking-wide text-slate-500 sm:mb-4 sm:text-xs">
                    <span>AI-чат</span>
                    <span>Предпросмотр</span>
                  </div>
                  <div className="grid gap-3 text-xs text-slate-600 sm:gap-4 sm:text-sm">
                    <div className="rounded-2xl bg-blue-600/10 px-3 py-2.5 text-blue-700 sm:px-4 sm:py-3">
                      Пользователь: «Сделай лендинг для нового AI-продукта, который автоматизирует рассылки»
                    </div>
                    <div className="rounded-2xl border border-white/70 bg-white/80 px-3 py-2.5 shadow-sm sm:px-4 sm:py-3">
                      Landly AI: «Готово! Посмотрите hero-блок с CTA и секциями цен. Всё можно поправить новым промптом.»
                    </div>
                    <div className="rounded-2xl border border-white/70 bg-white/90 px-3 py-2.5 shadow-sm sm:px-4 sm:py-3">
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

          <section id="workflow" className="grid gap-4 sm:gap-6 md:grid-cols-3">
            {benefits.map((benefit) => (
              <div key={benefit.title} className="surface-card border-white/40 p-4 sm:p-6">
                <div className="flex h-9 w-9 items-center justify-center rounded-full bg-blue-500/10 text-blue-600 sm:h-10 sm:w-10">
                  {benefit.icon}
                </div>
                <h3 className="mt-3 text-base font-semibold text-slate-900 sm:mt-4 sm:text-lg">{benefit.title}</h3>
                <p className="mt-2 text-xs text-slate-600 sm:text-sm">{benefit.description}</p>
              </div>
            ))}
          </section>

          <section id="security" className="glass-panel px-4 py-8 sm:px-8 sm:py-12 md:px-12">
            <div className="grid gap-6 sm:gap-10 md:grid-cols-2 md:items-center">
              <div className="space-y-3 sm:space-y-4">
                <span className="inline-flex items-center rounded-full border border-emerald-400/40 bg-emerald-500/10 px-3 py-1 text-[10px] font-semibold uppercase tracking-wide text-emerald-600 sm:px-4 sm:py-1.5 sm:text-xs">
                  CDN ready
                </span>
                <h2 className="text-xl font-bold text-slate-900 sm:text-2xl md:text-3xl lg:text-4xl">Публикация — это экспорт статики, а не сложный deploy.</h2>
                <p className="text-xs text-slate-600 sm:text-sm">
                  Каждая публикация — это статический HTML + CSS, загруженный в S3-совместимое хранилище. Никаких рантаймов, никакой ручной сборки. Предпросмотр и продакшн используют один и тот же шаблон.
                </p>
              </div>
              <div className="surface-card border-white/40 p-4 text-xs text-slate-600 sm:p-6 sm:text-sm">
                <ul className="space-y-3 sm:space-y-4">
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

          <section className="glass-panel flex flex-col items-center gap-4 px-4 py-8 text-center sm:gap-6 sm:px-10 sm:py-12 md:px-16">
            <p className="text-[10px] font-semibold uppercase tracking-[0.3em] text-blue-600 sm:text-xs">Landly workspace</p>
            <h2 className="text-xl font-bold text-slate-900 sm:text-2xl md:text-3xl lg:text-4xl">Попробуйте workflow, похожий на Cursor и ChatGPT.</h2>
            <p className="max-w-2xl text-xs text-slate-600 sm:text-sm">
              Войдите и увидите последний проект сразу — чат, предпросмотр и публикация объединены в одном окне. Никаких форм и бесконечных настроек.
            </p>
            <div className="flex flex-col gap-2 sm:flex-row sm:gap-3">
              <Link
                href="/auth/login"
                className="inline-flex h-10 items-center justify-center rounded-full bg-blue-600 px-5 text-xs font-semibold text-white shadow-lg transition hover:bg-blue-700 sm:h-11 sm:px-6 sm:text-sm"
              >
                Войти
              </Link>
              <Link
                href="/auth/signup"
                className="inline-flex h-10 items-center justify-center rounded-full border border-slate-200 bg-white/80 px-5 text-xs font-semibold text-slate-700 shadow-sm transition hover:bg-white sm:h-11 sm:px-6 sm:text-sm"
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
