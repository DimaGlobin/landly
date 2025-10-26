'use client'

import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'

export default function HomePage() {
  const router = useRouter()

  return (
    <div className="min-h-screen">
      {/* Hero */}
      <section className="py-20 px-6 text-center bg-gradient-to-b from-blue-50 to-white">
        <div className="max-w-4xl mx-auto">
          <h1 className="text-6xl font-bold mb-6 bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
            Создавайте лендинги с помощью AI
          </h1>
          <p className="text-xl text-gray-600 mb-8">
            Опишите ваш продукт — получите готовый профессиональный лендинг за минуты
          </p>
          <div className="flex gap-4 justify-center">
            <Button size="lg" onClick={() => router.push('/app/projects')}>
              Начать бесплатно
            </Button>
            <Button size="lg" variant="outline" onClick={() => router.push('/auth/login')}>
              Войти
            </Button>
          </div>
        </div>
      </section>

      {/* Features */}
      <section className="py-16 px-6 bg-white">
        <div className="max-w-6xl mx-auto">
          <h2 className="text-3xl font-bold text-center mb-12">Как это работает</h2>
          <div className="grid md:grid-cols-3 gap-8">
            <div className="text-center">
              <div className="text-5xl mb-4">🤖</div>
              <h3 className="text-xl font-semibold mb-2">1. Опишите продукт</h3>
              <p className="text-gray-600">
                Заполните AI-бриф: ниша, аудитория, преимущества
              </p>
            </div>
            <div className="text-center">
              <div className="text-5xl mb-4">✨</div>
              <h3 className="text-xl font-semibold mb-2">2. AI генерирует лендинг</h3>
              <p className="text-gray-600">
                Получите готовый дизайн с блоками Hero, Features, Pricing и др.
              </p>
            </div>
            <div className="text-center">
              <div className="text-5xl mb-4">🚀</div>
              <h3 className="text-xl font-semibold mb-2">3. Публикуйте</h3>
              <p className="text-gray-600">
                Одна кнопка — и ваш лендинг live в CDN с аналитикой
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="py-20 px-6 text-center bg-blue-600">
        <div className="max-w-3xl mx-auto">
          <h2 className="text-4xl font-bold text-white mb-6">
            Готовы создать свой первый лендинг?
          </h2>
          <Button size="lg" variant="outline" className="bg-white text-blue-600 hover:bg-gray-100" onClick={() => router.push('/auth/signup')}>
            Регистрация — бесплатно
          </Button>
        </div>
      </section>

      {/* Footer */}
      <footer className="py-8 px-6 bg-gray-100 text-center text-gray-600">
        <p>&copy; 2025 Landly. Все права защищены.</p>
      </footer>
    </div>
  )
}

