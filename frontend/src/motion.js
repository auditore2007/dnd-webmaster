import { useEffect, useRef, useState } from 'react'

// Общие помощники анимации. Пользователь, попросивший систему уменьшить движение, получает итог сразу.

export const reducedMotion = () => typeof matchMedia === 'function' && matchMedia('(prefers-reduced-motion: reduce)').matches

export const sleep = (ms) => new Promise((r) => setTimeout(r, ms))

/** useCountUp – число плавно «набегает» до значения (итог броска, урон); fromZero – и при первом показе. */
export function useCountUp(value, { ms = 450, fromZero = false } = {}) {
  const start0 = fromZero && typeof value === 'number' && !reducedMotion() ? 0 : value
  const [shown, setShown] = useState(start0)
  const from = useRef(start0)
  useEffect(() => {
    if (typeof value !== 'number' || reducedMotion()) { setShown(value); from.current = value; return }
    const start = performance.now()
    const a = typeof from.current === 'number' ? from.current : 0
    let raf = 0
    const tick = (t) => {
      const k = Math.min(1, (t - start) / ms)
      const e = 1 - (1 - k) ** 3
      setShown(Math.round(a + (value - a) * e))
      if (k < 1) raf = requestAnimationFrame(tick)
      else from.current = value
    }
    raf = requestAnimationFrame(tick)
    return () => cancelAnimationFrame(raf)
  }, [value, ms])
  return shown
}

/**
 * usePulse – класс-вспышка при изменении числа: 'hit', когда оно упало, 'heal', когда выросло.
 * Первое значение вспышки не даёт.
 */
export function usePulse(value, ms = 650) {
  const prev = useRef(value)
  const [cls, setCls] = useState('')
  useEffect(() => {
    if (prev.current === value) return
    const next = value < prev.current ? 'hit' : 'heal'
    prev.current = value
    setCls(next)
    const t = setTimeout(() => setCls(''), ms)
    return () => clearTimeout(t)
  }, [value, ms])
  return cls
}
