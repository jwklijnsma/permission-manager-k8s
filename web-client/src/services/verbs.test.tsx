import {grantsRead, grantsWrite} from './verbs'

describe('grantsRead', () => {
  it('accepts the wildcard', () => {
    expect(grantsRead(['*'])).toBe(true)
  })

  it('accepts get/list/watch', () => {
    expect(grantsRead(['get'])).toBe(true)
    expect(grantsRead(['list', 'watch'])).toBe(true)
  })

  it('does not treat "read" as a Kubernetes verb (issue #186)', () => {
    expect(grantsRead(['read'])).toBe(false)
  })

  it('handles write-only and empty verb lists', () => {
    expect(grantsRead(['create'])).toBe(false)
    expect(grantsRead([])).toBe(false)
    expect(grantsRead(undefined)).toBe(false)
  })
})

describe('grantsWrite', () => {
  it('accepts the wildcard', () => {
    expect(grantsWrite(['*'])).toBe(true)
  })

  it('accepts any mutating verb (issue #84)', () => {
    expect(grantsWrite(['update'])).toBe(true)
    expect(grantsWrite(['patch'])).toBe(true)
    expect(grantsWrite(['delete'])).toBe(true)
    expect(grantsWrite(['create'])).toBe(true)
  })

  it('is false for read-only verbs', () => {
    expect(grantsWrite(['get', 'list', 'watch'])).toBe(false)
    expect(grantsWrite(null)).toBe(false)
  })
})
