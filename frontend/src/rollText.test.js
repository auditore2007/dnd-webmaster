import { describe as suite, expect, test } from 'vitest'
import { attackSummary, cubeTurn, describe, diceHint, dimmed, keptD20, toneOf } from './rollText.js'
import { priceGP } from './labels.js'

suite('describe', () => {
  test('проверка с преимуществом показывает оба кубика и взятый', () => {
    const d = describe({ kind: 'check', label: 'Лира: Сила', rolls: [4, 17], kept: 17, bonus: 3, total: 20, sides: 20, mode: 'adv', dc: 15, success: true })
    expect(d.title).toBe('Лира: Сила')
    expect(d.formula).toBe('d20 [4, 17] – преимущество, берём 17 + 3 = 20')
    expect(d.outcome).toBe('Успех – сложность 15')
    expect(d.tone).toBe('ok')
  })

  test('натуральная 20 подсвечивается как крит', () => {
    const d = describe({ kind: 'save', label: 'Лира: спасбросок – Ловкость', rolls: [20], kept: 20, bonus: -1, total: 19, sides: 20 })
    expect(d.tone).toBe('crit')
    expect(d.formula).toContain('− 1')
  })

  test('натуральная 1 – провал', () => {
    expect(describe({ kind: 'check', label: 'x', rolls: [1], kept: 1, total: 1, sides: 20 }).tone).toBe('bad')
  })

  test('атака: попадание показывает урон', () => {
    const d = describe({ kind: 'attack', label: 'A → B', rolls: [15], total: 19, target: 13, hit: true, crit: false, damage: 7, damageText: '1d8+3' })
    expect(d.big).toBe(7)
    expect(d.bigLabel).toBe('урон')
    expect(d.outcome).toBe('Попадание. Урон 7 (1d8+3)')
  })

  test('атака: промах', () => {
    expect(describe({ kind: 'attack', label: 'A', rolls: [2], total: 5, target: 13, hit: false }).outcome).toBe('Промах')
  })

  test('свободный бросок без подписи не падает', () => {
    expect(describe({ kind: 'roll', rolls: [3, 4], total: 7, sides: 6 }).formula).toBe('[3 + 4] = 7')
  })
})

suite('diceHint', () => {
  test.each([['2d6+3', 2, 6], ['d20', 1, 20], ['1d100', 1, 100], ['мусор', 1, 20], ['0d8', 1, 8]])('%s', (e, count, sides) => {
    expect(diceHint(e)).toEqual({ count, sides })
  })
})

suite('keptD20', () => {
  test('преимущество – больший, помеха – меньший', () => {
    expect(keptD20([3, 18], 'adv')).toBe(18)
    expect(keptD20([3, 18], 'dis')).toBe(3)
    expect(keptD20([12], 'adv')).toBe(12)
    expect(keptD20([], '')).toBe(0)
  })
})

suite('dimmed', () => {
  test('отброшенный d20 тускнеет', () => {
    expect(dimmed({ sides: 20, rolls: [3, 18], kept: 18 })).toEqual([true, false])
    expect(dimmed({ sides: 20, rolls: [7, 7], kept: 7 })).toEqual([false, true])
    expect(dimmed({ sides: 6, rolls: [1, 2], kept: 3 })).toEqual([false, false])
  })
})

test('toneOf', () => {
  expect(toneOf(20, 20)).toBe('crit')
  expect(toneOf(20, 1)).toBe('bad')
  expect(toneOf(6, 1)).toBe('')
})

test('cubeTurn даёт поворот для каждой грани', () => {
  const seen = new Set([1, 2, 3, 4, 5, 6].map((v) => cubeTurn(v).join()))
  expect(seen.size).toBe(6)
  expect(cubeTurn(1)).toEqual([0, 0])
})

test('attackSummary', () => {
  expect(attackSummary([{ hit: true, damage: 5 }, { hit: false }, { hit: true, damage: 3, crit: true }])).toEqual({ hits: 2, total: 3, damage: 8, crit: true })
})

test('priceGP', () => {
  expect(priceGP('15 зм')).toBe(15)
  expect(priceGP('1 000 зм')).toBe(1000)
  expect(priceGP('5 см')).toBeCloseTo(0.5)
  expect(priceGP('')).toBe(0)
})
