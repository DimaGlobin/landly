import { ApiError, ApiErrorCode } from '../api'

describe('ApiError', () => {
  describe('constructor', () => {
    it('should create error with code and message', () => {
      const error = new ApiError('UNAUTHORIZED', 'Authorization required')
      
      expect(error.code).toBe('UNAUTHORIZED')
      expect(error.originalMessage).toBe('Authorization required')
      expect(error.message).toBe('Authorization required')
      expect(error.name).toBe('ApiError')
    })

    it('should use code as message if message not provided', () => {
      const error = new ApiError('INTERNAL_ERROR')
      
      expect(error.code).toBe('INTERNAL_ERROR')
      expect(error.message).toBe('INTERNAL_ERROR')
    })
  })

  describe('isAuthError', () => {
    const authCodes: ApiErrorCode[] = ['UNAUTHORIZED', 'INVALID_TOKEN', 'TOKEN_EXPIRED', 'INVALID_CREDENTIALS']
    const nonAuthCodes: ApiErrorCode[] = ['NOT_FOUND', 'INTERNAL_ERROR', 'VALIDATION_ERROR', 'NETWORK_ERROR']

    authCodes.forEach((code) => {
      it(`should return true for ${code}`, () => {
        const error = new ApiError(code)
        expect(error.isAuthError()).toBe(true)
      })
    })

    nonAuthCodes.forEach((code) => {
      it(`should return false for ${code}`, () => {
        const error = new ApiError(code)
        expect(error.isAuthError()).toBe(false)
      })
    })
  })

  describe('shouldRedirectToLogin', () => {
    const redirectCodes: ApiErrorCode[] = ['UNAUTHORIZED', 'INVALID_TOKEN', 'TOKEN_EXPIRED']
    const noRedirectCodes: ApiErrorCode[] = ['INVALID_CREDENTIALS', 'NOT_FOUND', 'INTERNAL_ERROR']

    redirectCodes.forEach((code) => {
      it(`should return true for ${code}`, () => {
        const error = new ApiError(code)
        expect(error.shouldRedirectToLogin()).toBe(true)
      })
    })

    noRedirectCodes.forEach((code) => {
      it(`should return false for ${code}`, () => {
        const error = new ApiError(code)
        expect(error.shouldRedirectToLogin()).toBe(false)
      })
    })
  })
})

describe('API error codes', () => {
  const allCodes: ApiErrorCode[] = [
    'UNAUTHORIZED',
    'INVALID_TOKEN',
    'TOKEN_EXPIRED',
    'INVALID_CREDENTIALS',
    'USER_ALREADY_EXISTS',
    'USER_NOT_FOUND',
    'FORBIDDEN',
    'NOT_FOUND',
    'PROJECT_NOT_FOUND',
    'ALREADY_EXISTS',
    'VALIDATION_ERROR',
    'INVALID_INPUT',
    'BAD_REQUEST',
    'GENERATION_FAILED',
    'RENDER_FAILED',
    'PUBLISH_FAILED',
    'INTERNAL_ERROR',
    'NETWORK_ERROR',
    'UNKNOWN',
  ]

  it('should have all expected error codes', () => {
    // This test ensures type safety - if a code is removed from the type,
    // this array will have a type error
    expect(allCodes.length).toBe(19)
  })

  allCodes.forEach((code) => {
    it(`should be able to create error with code ${code}`, () => {
      const error = new ApiError(code, `Message for ${code}`)
      expect(error.code).toBe(code)
    })
  })
})

