'use client'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

interface SimpleLandingPreviewProps {
  schema: any
}

export function SimpleLandingPreview({ schema }: SimpleLandingPreviewProps) {
  if (!schema) {
    return null
  }

  const page = schema.pages?.[0]
  if (!page) {
    return null
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>🎨 Предпросмотр лендинга</CardTitle>
        <CardDescription>
          Результат генерации
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          {/* Hero блок */}
          {page.blocks.find((block: any) => block.type === 'hero') && (
            <div className="bg-gradient-to-r from-blue-500 to-purple-600 text-white p-8 rounded-lg">
              <h1 className="text-3xl font-bold mb-4">
                {page.blocks.find((block: any) => block.type === 'hero')?.props?.headline}
              </h1>
              <p className="text-xl mb-6">
                {page.blocks.find((block: any) => block.type === 'hero')?.props?.subheadline}
              </p>
              <button className="bg-white text-blue-600 px-6 py-3 rounded-lg font-semibold hover:bg-gray-100 transition-colors">
                {page.blocks.find((block: any) => block.type === 'hero')?.props?.ctaText}
              </button>
            </div>
          )}

          {/* Features блок */}
          {page.blocks.find((block: any) => block.type === 'features') && (
            <div>
              <h2 className="text-2xl font-bold mb-6 text-center">
                {page.blocks.find((block: any) => block.type === 'features')?.props?.title}
              </h2>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                {page.blocks.find((block: any) => block.type === 'features')?.props?.items?.map((item: any, index: number) => (
                  <div key={index} className="text-center p-6 border rounded-lg">
                    <div className="text-4xl mb-4">{item.icon}</div>
                    <h3 className="text-xl font-semibold mb-2">{item.title}</h3>
                    <p className="text-gray-600">{item.description}</p>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Pricing блок */}
          {page.blocks.find((block: any) => block.type === 'pricing') && (
            <div>
              <h2 className="text-2xl font-bold mb-6 text-center">
                {page.blocks.find((block: any) => block.type === 'pricing')?.props?.title}
              </h2>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                {page.blocks.find((block: any) => block.type === 'pricing')?.props?.plans?.map((plan: any, index: number) => (
                  <div key={index} className={`p-6 border rounded-lg ${plan.featured ? 'border-blue-500 bg-blue-50' : ''}`}>
                    {plan.featured && (
                      <div className="bg-blue-500 text-white text-center py-1 px-3 rounded-full text-sm font-semibold mb-4">
                        Популярный
                      </div>
                    )}
                    <h3 className="text-xl font-semibold mb-2">{plan.name}</h3>
                    <div className="text-3xl font-bold mb-4">
                      {plan.price} {plan.currency}/{plan.period}
                    </div>
                    <ul className="space-y-2 mb-6">
                      {plan.features.map((feature: string, featureIndex: number) => (
                        <li key={featureIndex} className="flex items-center">
                          <span className="text-green-500 mr-2">✓</span>
                          {feature}
                        </li>
                      ))}
                    </ul>
                    <button className={`w-full py-2 px-4 rounded-lg font-semibold ${
                      plan.featured 
                        ? 'bg-blue-500 text-white hover:bg-blue-600' 
                        : 'bg-gray-200 text-gray-800 hover:bg-gray-300'
                    } transition-colors`}>
                      Выбрать план
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Testimonials блок */}
          {page.blocks.find((block: any) => block.type === 'testimonials') && (
            <div>
              <h2 className="text-2xl font-bold mb-6 text-center">
                {page.blocks.find((block: any) => block.type === 'testimonials')?.props?.title}
              </h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {page.blocks.find((block: any) => block.type === 'testimonials')?.props?.items?.map((testimonial: any, index: number) => (
                  <div key={index} className="p-6 border rounded-lg bg-gray-50">
                    <div className="flex items-center mb-4">
                      {[...Array(testimonial.rating)].map((_, i) => (
                        <span key={i} className="text-yellow-400">★</span>
                      ))}
                    </div>
                    <p className="text-gray-700 mb-4">"{testimonial.text}"</p>
                    <div>
                      <p className="font-semibold">{testimonial.author}</p>
                      <p className="text-gray-600 text-sm">{testimonial.role}</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* FAQ блок */}
          {page.blocks.find((block: any) => block.type === 'faq') && (
            <div>
              <h2 className="text-2xl font-bold mb-6 text-center">
                {page.blocks.find((block: any) => block.type === 'faq')?.props?.title}
              </h2>
              <div className="space-y-4">
                {page.blocks.find((block: any) => block.type === 'faq')?.props?.items?.map((faq: any, index: number) => (
                  <div key={index} className="border rounded-lg p-4">
                    <h3 className="font-semibold mb-2">{faq.question}</h3>
                    <p className="text-gray-600">{faq.answer}</p>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* CTA блок */}
          {page.blocks.find((block: any) => block.type === 'cta') && (
            <div className="bg-gray-900 text-white p-8 rounded-lg text-center">
              <h2 className="text-2xl font-bold mb-4">
                {page.blocks.find((block: any) => block.type === 'cta')?.props?.title}
              </h2>
              <p className="text-xl mb-6">
                {page.blocks.find((block: any) => block.type === 'cta')?.props?.description}
              </p>
              <button className="bg-white text-gray-900 px-8 py-3 rounded-lg font-semibold hover:bg-gray-100 transition-colors">
                {page.blocks.find((block: any) => block.type === 'cta')?.props?.buttonText}
              </button>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
