'use client'

import { CSSProperties, useMemo } from 'react'

import { LandingSchema, Block } from '@/lib/types'

interface LandingPreviewProps {
  schema: LandingSchema
}

type Palette = {
  primary: string
  secondary: string
  accent: string
  background: string
  text: string
}

const DEFAULT_PALETTE: Palette = {
  primary: '#2563EB',
  secondary: '#7C3AED',
  accent: '#F97316',
  background: '#FFFFFF',
  text: '#1F2937',
}

export function LandingPreview({ schema }: LandingPreviewProps) {
  const page = schema.pages?.[0]

  const palette = useMemo<Palette>(() => {
    const themePalette = schema.theme?.palette ?? {}
    return {
      primary: themePalette.primary || DEFAULT_PALETTE.primary,
      secondary: themePalette.secondary || DEFAULT_PALETTE.secondary,
      accent: themePalette.accent || DEFAULT_PALETTE.accent,
      background: themePalette.background || DEFAULT_PALETTE.background,
      text: themePalette.text || DEFAULT_PALETTE.text,
    }
  }, [schema.theme])

  const sortedBlocks = useMemo(() => {
    return [...(page?.blocks ?? [])].sort((a, b) => a.order - b.order)
  }, [page?.blocks])

  const cssVariables = useMemo(() => {
    return {
      '--landing-primary': palette.primary,
      '--landing-secondary': palette.secondary,
      '--landing-accent': palette.accent,
      '--landing-background': palette.background,
      '--landing-surface': '#ffffff',
      '--landing-text': palette.text,
    } as CSSProperties
  }, [palette])

  return (
    <div className="w-full overflow-hidden rounded-lg border bg-white shadow-sm">
      <div className="flex min-w-0 items-center gap-2 border-b bg-gray-100 px-4 py-2">
        <div className="flex flex-none gap-1">
          <div className="h-3 w-3 rounded-full bg-red-400" />
          <div className="h-3 w-3 rounded-full bg-yellow-400" />
          <div className="h-3 w-3 rounded-full bg-green-400" />
        </div>
        <div className="ml-2 flex-1 truncate text-sm text-gray-600" title={page?.title || 'Landing preview'}>
          Preview: {page?.title || 'Landing preview'}
        </div>
      </div>

      <div className="max-h-[800px] overflow-y-auto">
        {page ? (
          <main className="landing" style={cssVariables}>
            {sortedBlocks.length > 0 ? (
              sortedBlocks.map((block) => (
                <BlockRenderer key={`${block.type}-${block.order}`} block={block} schema={schema} />
              ))
            ) : (
              <div className="landing-empty-state">Нет блоков для отображения</div>
            )}
          </main>
        ) : (
          <div className="landing-empty-state">Схема пока не содержит страниц</div>
        )}
      </div>
    </div>
  )
}

function BlockRenderer({ block, schema }: { block: Block; schema: LandingSchema }) {
  switch (block.type) {
    case 'hero':
      return <HeroSection props={block.props} />
    case 'features':
      return <FeaturesSection props={block.props} />
    case 'pricing':
      return <PricingSection props={block.props} payment={schema.payment} />
    case 'testimonials':
      return <TestimonialsSection props={block.props} />
    case 'faq':
      return <FAQSection props={block.props} />
    case 'cta':
      return <CTASection props={block.props} />
    default:
      return (
        <section className="landing-section" data-block={block.type}>
          <div className="landing-container">
            <div className="landing-empty-state">Блок «{block.type}» пока не поддерживается</div>
          </div>
        </section>
      )
  }
}

function HeroSection({ props }: { props: Record<string, any> }) {
  const headline = props?.headline ?? 'Заголовок лендинга'
  const subheadline = props?.subheadline ?? ''
  const ctaText = props?.ctaText ?? ''
  const ctaUrl = props?.ctaUrl ?? '#'
  const secondaryText = props?.secondaryCtaText ?? 'Подробнее'
  const secondaryUrl = props?.secondaryCtaUrl ?? '#'
  const eyebrow = props?.eyebrow ?? 'Инновационная платформа'
  const brand = props?.brand ?? 'Landly'
  const navItems: string[] = Array.isArray(props?.navItems)
    ? props.navItems.filter((item: unknown): item is string => typeof item === 'string' && item.trim().length > 0)
    : []
  const navActionText = props?.navActionText ?? 'Войти'
  const navActionUrl = props?.navActionUrl ?? '#'
  const heroImage = typeof props?.image === 'string' ? props.image : ''
  const imageAlt = props?.imageAlt || headline

  const items = navItems.length > 0 ? navItems : ['Возможности', 'Цены', 'Отзывы', 'Контакты']

  return (
    <section className="landing-section landing-section--hero" data-block="hero">
      <div className="landing-hero-overlay" />
      <div className="landing-container">
        <div className="landing-topbar">
          <span className="landing-brand">{brand}</span>
          <nav className="landing-nav">
            {items.map((item) => (
              <a key={item} href="#">
                {item}
              </a>
            ))}
          </nav>
          {navActionText && (
            <a className="landing-nav-action" href={navActionUrl} target="_blank" rel="noopener noreferrer">
              {navActionText}
            </a>
          )}
        </div>

        <div className="landing-hero-grid">
          <div className="landing-hero-content">
            {eyebrow && <span className="landing-eyebrow">{eyebrow}</span>}
            <h1>{headline}</h1>
            {subheadline && <p>{subheadline}</p>}
            {(ctaText || secondaryText) && (
              <div className="landing-actions landing-actions--hero">
                {ctaText && (
                  <a className="landing-button landing-button--primary" data-track="cta_click" href={ctaUrl}>
                    {ctaText}
                  </a>
                )}
                {secondaryText && (
                  <a className="landing-button landing-button--ghost" data-track="cta_secondary" href={secondaryUrl}>
                    {secondaryText}
                  </a>
                )}
              </div>
            )}
          </div>
          {heroImage && (
            <div className="landing-hero-media">
              <div className="landing-hero-media-card">
                <img src={heroImage} alt={imageAlt} />
              </div>
            </div>
          )}
        </div>
      </div>
    </section>
  )
}

function FeaturesSection({ props }: { props: Record<string, any> }) {
  const title = props?.title ?? 'Наши преимущества'
  const items: Array<Record<string, any>> = props?.items ?? []

  return (
    <section className="landing-section landing-section--features" data-block="features">
      <div className="landing-container">
        <div className="landing-section-header">
          <h2 className="landing-section-title">{title}</h2>
        </div>
        <div className="landing-features__grid">
          {items.length > 0 ? (
            items.map((item, index) => (
              <div key={index} className="landing-card landing-feature-card">
                {item.icon && (
                  <div className="landing-feature-icon">
                    <span>{item.icon}</span>
                  </div>
                )}
                <h3>{item.title}</h3>
                {item.description && <p>{item.description}</p>}
              </div>
            ))
          ) : (
            <div className="landing-card landing-feature-card">
              <p className="landing-empty-state">Добавьте преимущества, чтобы показать их здесь</p>
            </div>
          )}
        </div>
      </div>
    </section>
  )
}

function PricingSection({ props, payment }: { props: Record<string, any>; payment?: LandingSchema['payment'] }) {
  const title = props?.title ?? 'Тарифы'
  const plans: Array<Record<string, any>> = props?.plans ?? []

  return (
    <section className="landing-section landing-section--pricing" data-block="pricing">
      <div className="landing-container">
        <div className="landing-section-header">
          <h2 className="landing-section-title">{title}</h2>
        </div>
        <div className="landing-pricing__grid">
          {plans.length > 0 ? (
            plans.map((plan, index) => {
              const buttonText = plan.buttonText || payment?.buttonText || 'Выбрать тариф'
              const buttonHref = plan.url || payment?.url || ''

              return (
                <div
                  key={index}
                  className={`pricing-card ${plan.featured ? 'pricing-card--featured' : ''}`.trim()}
                  data-featured={plan.featured ? 'true' : 'false'}
                >
                  <div className="pricing-name">{plan.name}</div>
                  <div className="pricing-price">
                    <span className="pricing-price__value">{plan.price}</span>
                    <span className="pricing-price__period">
                      {plan.currency}
                      {plan.period ? ` / ${plan.period}` : ''}
                    </span>
                  </div>
                  <ul className="pricing-features">
                    {(plan.features ?? []).map((feature: string, featureIndex: number) => (
                      <li key={featureIndex} className="pricing-feature">
                        <span className="pricing-feature-icon">✓</span>
                        <span>{feature}</span>
                      </li>
                    ))}
                  </ul>
                  <div className="pricing-action">
                    {buttonHref ? (
                      <a
                        className="landing-button landing-button--secondary"
                        data-track="pay_click"
                        href={buttonHref}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        {buttonText}
                      </a>
                    ) : (
                      <button type="button" className="landing-button landing-button--secondary" data-track="pay_click">
                        {buttonText}
                      </button>
                    )}
                  </div>
                </div>
              )
            })
          ) : (
            <div className="landing-card">
              <div className="landing-empty-state">Добавьте тарифы в описании проекта</div>
            </div>
          )}
        </div>
      </div>
    </section>
  )
}

function TestimonialsSection({ props }: { props: Record<string, any> }) {
  const title = props?.title ?? 'Отзывы клиентов'
  const items: Array<Record<string, any>> = props?.items ?? []

  return (
    <section className="landing-section landing-section--testimonials" data-block="testimonials">
      <div className="landing-container">
        <div className="landing-section-header">
          <h2 className="landing-section-title">{title}</h2>
        </div>
        <div className="landing-testimonials__grid">
          {items.length > 0 ? (
            items.map((item, index) => (
              <div key={index} className="landing-card landing-testimonial-card">
                {item.text && <p className="landing-testimonial-quote">“{item.text}”</p>}
                <div className="landing-testimonial-author">
                  <strong>{item.author}</strong>
                  {item.role && <span>{item.role}</span>}
                  {item.rating && <span className="landing-testimonial-rating">⭐ {item.rating}</span>}
                </div>
              </div>
            ))
          ) : (
            <div className="landing-card landing-testimonial-card">
              <p className="landing-empty-state">Добавьте отзывы, чтобы повысить доверие</p>
            </div>
          )}
        </div>
      </div>
    </section>
  )
}

function FAQSection({ props }: { props: Record<string, any> }) {
  const title = props?.title ?? 'Частые вопросы'
  const items: Array<Record<string, any>> = props?.items ?? []

  return (
    <section className="landing-section landing-section--faq" data-block="faq">
      <div className="landing-container">
        <div className="landing-section-header">
          <h2 className="landing-section-title">{title}</h2>
        </div>
        <div className="landing-faq__list">
          {items.length > 0 ? (
            items.map((item, index) => (
              <div key={index} className="faq-item">
                <div className="faq-question">{item.question}</div>
                {item.answer && <div className="faq-answer">{item.answer}</div>}
              </div>
            ))
          ) : (
            <div className="faq-item">
              <div className="landing-empty-state">Добавьте вопросы и ответы, которые волнуют клиентов</div>
            </div>
          )}
        </div>
      </div>
    </section>
  )
}

function CTASection({ props }: { props: Record<string, any> }) {
  const title = props?.title ?? 'Готовы начать?'
  const description = props?.description ?? ''
  const buttonText = props?.buttonText ?? 'Связаться с нами'
  const buttonUrl = props?.buttonUrl ?? '#'
  const secondaryText = props?.secondaryButtonText ?? ''
  const secondaryUrl = props?.secondaryButtonUrl ?? '#'

  return (
    <section className="landing-section landing-section--cta" data-block="cta">
      <div className="landing-container">
        <div className="landing-section-header">
          <h2 className="landing-section-title">{title}</h2>
          {description && <p>{description}</p>}
        </div>
        <div className="landing-actions landing-actions--center">
          <a className="landing-button landing-button--primary" href={buttonUrl} data-track="cta_click">
            {buttonText}
          </a>
          {secondaryText && (
            <a className="landing-button landing-button--ghost" href={secondaryUrl} data-track="cta_secondary">
              {secondaryText}
            </a>
          )}
        </div>
      </div>
    </section>
  )
}

