'use client'

import { LandingSchema, Block } from '@/lib/types'

interface LandingPreviewProps {
  schema: LandingSchema
}

export function LandingPreview({ schema }: LandingPreviewProps) {
  const page = schema.pages[0] // Показываем главную страницу
  const theme = schema.theme
  const payment = schema.payment

  return (
    <div className="w-full bg-white border rounded-lg overflow-hidden shadow-sm">
      {/* Preview Header */}
      <div className="bg-gray-100 px-4 py-2 border-b flex items-center gap-2">
        <div className="flex gap-1">
          <div className="w-3 h-3 rounded-full bg-red-400" />
          <div className="w-3 h-3 rounded-full bg-yellow-400" />
          <div className="w-3 h-3 rounded-full bg-green-400" />
        </div>
        <div className="text-sm text-gray-600 ml-2">Preview: {page.title}</div>
      </div>

      {/* Preview Content */}
      <div className="overflow-y-auto max-h-[800px]">
        {page.blocks
          .sort((a, b) => a.order - b.order)
          .map((block, idx) => (
            <BlockPreview key={idx} block={block} theme={theme} payment={payment} />
          ))}
      </div>
    </div>
  )
}

function BlockPreview({ block, theme, payment }: { block: Block; theme?: any; payment?: any }) {
  const primaryColor = theme?.palette?.primary || '#3B82F6'

  switch (block.type) {
    case 'hero':
      return (
        <section className="py-20 px-6 text-center bg-gradient-to-b from-blue-50 to-white">
          <h1 className="text-5xl font-bold mb-4" style={{ color: primaryColor }}>
            {block.props.headline}
          </h1>
          <p className="text-xl text-gray-600 mb-8">{block.props.subheadline}</p>
          <button
            className="px-8 py-3 rounded-lg text-white font-semibold"
            style={{ backgroundColor: primaryColor }}
          >
            {block.props.ctaText || 'Get Started'}
          </button>
        </section>
      )

    case 'features':
      return (
        <section className="py-16 px-6 bg-white">
          <h2 className="text-3xl font-bold text-center mb-12">{block.props.title}</h2>
          <div className="grid md:grid-cols-3 gap-8 max-w-6xl mx-auto">
            {block.props.items?.map((item: any, idx: number) => (
              <div key={idx} className="text-center">
                <div className="text-4xl mb-4">{item.icon}</div>
                <h3 className="text-xl font-semibold mb-2">{item.title}</h3>
                <p className="text-gray-600">{item.description}</p>
              </div>
            ))}
          </div>
        </section>
      )

    case 'pricing':
      return (
        <section className="py-16 px-6 bg-gray-50">
          <h2 className="text-3xl font-bold text-center mb-12">{block.props.title}</h2>
          <div className="grid md:grid-cols-3 gap-8 max-w-6xl mx-auto">
            {block.props.plans?.map((plan: any, idx: number) => (
              <div
                key={idx}
                className={`bg-white p-8 rounded-lg shadow ${plan.featured ? 'border-2' : 'border'}`}
                style={plan.featured ? { borderColor: primaryColor } : {}}
              >
                <h3 className="text-2xl font-bold mb-2">{plan.name}</h3>
                <div className="mb-6">
                  <span className="text-4xl font-bold">{plan.price}</span>
                  <span className="text-gray-600"> {plan.currency}/{plan.period}</span>
                </div>
                <ul className="space-y-2 mb-6">
                  {plan.features?.map((feature: string, fidx: number) => (
                    <li key={fidx} className="flex items-start">
                      <span className="text-green-500 mr-2">✓</span>
                      <span>{feature}</span>
                    </li>
                  ))}
                </ul>
                {payment && (
                  <a
                    href={payment.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="block w-full text-center px-6 py-3 rounded-lg text-white font-semibold"
                    style={{ backgroundColor: primaryColor }}
                  >
                    {payment.buttonText || 'Оплатить'}
                  </a>
                )}
              </div>
            ))}
          </div>
        </section>
      )

    case 'testimonials':
      return (
        <section className="py-16 px-6 bg-white">
          <h2 className="text-3xl font-bold text-center mb-12">{block.props.title}</h2>
          <div className="grid md:grid-cols-2 gap-8 max-w-4xl mx-auto">
            {block.props.items?.map((item: any, idx: number) => (
              <div key={idx} className="bg-gray-50 p-6 rounded-lg">
                <p className="text-gray-700 mb-4">"{item.text}"</p>
                <div>
                  <div className="font-semibold">{item.author}</div>
                  <div className="text-sm text-gray-600">{item.role}</div>
                </div>
              </div>
            ))}
          </div>
        </section>
      )

    case 'faq':
      return (
        <section className="py-16 px-6 bg-gray-50">
          <h2 className="text-3xl font-bold text-center mb-12">{block.props.title}</h2>
          <div className="max-w-3xl mx-auto space-y-4">
            {block.props.items?.map((item: any, idx: number) => (
              <div key={idx} className="bg-white p-6 rounded-lg">
                <h3 className="font-semibold mb-2">{item.question}</h3>
                <p className="text-gray-600">{item.answer}</p>
              </div>
            ))}
          </div>
        </section>
      )

    case 'cta':
      return (
        <section className="py-20 px-6 text-center" style={{ backgroundColor: primaryColor }}>
          <h2 className="text-3xl font-bold text-white mb-4">{block.props.title}</h2>
          <p className="text-white/90 mb-8">{block.props.description}</p>
          <button className="px-8 py-3 bg-white rounded-lg font-semibold" style={{ color: primaryColor }}>
            {block.props.buttonText || 'Get Started'}
          </button>
        </section>
      )

    default:
      return (
        <section className="py-8 px-6 bg-gray-100">
          <div className="text-center text-gray-500">Block: {block.type}</div>
        </section>
      )
  }
}

