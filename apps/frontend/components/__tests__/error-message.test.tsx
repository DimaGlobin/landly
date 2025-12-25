import { render, screen, fireEvent } from '@testing-library/react'
import { ApiError } from '@/lib/api'

// Mock next-intl
jest.mock('next-intl', () => ({
  useTranslations: () => (key: string) => {
    const translations: Record<string, string> = {
      UNAUTHORIZED: 'Требуется авторизация',
      INVALID_TOKEN: 'Сессия истекла, войдите снова',
      INVALID_CREDENTIALS: 'Неверный email или пароль',
      USER_ALREADY_EXISTS: 'Пользователь с таким email уже существует',
      PROJECT_NOT_FOUND: 'Проект не найден',
      GENERATION_FAILED: 'Не удалось сгенерировать лендинг',
      INTERNAL_ERROR: 'Что-то пошло не так. Попробуйте позже',
      NETWORK_ERROR: 'Нет соединения с сервером',
      UNKNOWN: 'Произошла неизвестная ошибка',
    }
    return translations[key] || key
  },
  NextIntlClientProvider: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}))

// Import after mocking
import { ErrorMessage } from '../error-message'

describe('ErrorMessage', () => {
  describe('rendering', () => {
    it('should not render when error is null', () => {
      const { container } = render(<ErrorMessage error={null} />)
      expect(container.firstChild).toBeNull()
    })

    it('should render translated message for ApiError', () => {
      const error = new ApiError('INVALID_CREDENTIALS', 'Invalid email or password')
      render(<ErrorMessage error={error} />)
      
      expect(screen.getByText('Неверный email или пароль')).toBeInTheDocument()
    })

    it('should render message for regular Error', () => {
      const error = new Error('Something went wrong')
      render(<ErrorMessage error={error} />)
      
      expect(screen.getByText('Something went wrong')).toBeInTheDocument()
    })

    it('should render string error', () => {
      render(<ErrorMessage error="Custom error message" />)
      
      expect(screen.getByText('Custom error message')).toBeInTheDocument()
    })
  })

  describe('variants', () => {
    it('should render inline variant by default', () => {
      const error = new ApiError('INTERNAL_ERROR')
      const { container } = render(<ErrorMessage error={error} />)
      
      expect(container.querySelector('.flex.items-center.gap-2')).toBeInTheDocument()
    })

    it('should render banner variant', () => {
      const error = new ApiError('INTERNAL_ERROR')
      const { container } = render(<ErrorMessage error={error} variant="banner" />)
      
      expect(container.querySelector('.justify-between')).toBeInTheDocument()
    })

    it('should render toast variant', () => {
      const error = new ApiError('INTERNAL_ERROR')
      const { container } = render(<ErrorMessage error={error} variant="toast" />)
      
      expect(container.querySelector('.shadow-lg')).toBeInTheDocument()
    })
  })

  describe('dismiss button', () => {
    it('should not show dismiss button when onDismiss not provided', () => {
      const error = new ApiError('INTERNAL_ERROR')
      render(<ErrorMessage error={error} variant="banner" />)
      
      expect(screen.queryByRole('button', { name: /dismiss/i })).not.toBeInTheDocument()
    })

    it('should show dismiss button when onDismiss provided', () => {
      const error = new ApiError('INTERNAL_ERROR')
      const onDismiss = jest.fn()
      render(<ErrorMessage error={error} variant="banner" onDismiss={onDismiss} />)
      
      expect(screen.getByRole('button', { name: /dismiss/i })).toBeInTheDocument()
    })

    it('should call onDismiss when button clicked', () => {
      const error = new ApiError('INTERNAL_ERROR')
      const onDismiss = jest.fn()
      render(<ErrorMessage error={error} variant="banner" onDismiss={onDismiss} />)
      
      fireEvent.click(screen.getByRole('button', { name: /dismiss/i }))
      
      expect(onDismiss).toHaveBeenCalledTimes(1)
    })
  })

  describe('error codes', () => {
    const testCases = [
      { code: 'UNAUTHORIZED' as const, expected: 'Требуется авторизация' },
      { code: 'INVALID_CREDENTIALS' as const, expected: 'Неверный email или пароль' },
      { code: 'USER_ALREADY_EXISTS' as const, expected: 'Пользователь с таким email уже существует' },
      { code: 'NETWORK_ERROR' as const, expected: 'Нет соединения с сервером' },
    ]

    testCases.forEach(({ code, expected }) => {
      it(`should display translated message for ${code}`, () => {
        const error = new ApiError(code)
        render(<ErrorMessage error={error} />)
        
        expect(screen.getByText(expected)).toBeInTheDocument()
      })
    })
  })

  describe('custom className', () => {
    it('should apply custom className', () => {
      const error = new ApiError('INTERNAL_ERROR')
      const { container } = render(
        <ErrorMessage error={error} className="custom-class" />
      )
      
      expect(container.querySelector('.custom-class')).toBeInTheDocument()
    })
  })
})
