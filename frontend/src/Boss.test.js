import { expect, test } from 'vitest'
import { legendaryActs, legendaryLeft } from './Boss.jsx'

test('легендарные действия: из бестиария или атаки каждым оружием', () => {
  const dragon = { stat: { legendary: 3, legActs: [{ key: 'tail', name: 'Хвост', cost: 1, weapon: 2 }] }, weapons: [] }
  expect(legendaryActs(dragon).map((a) => a.key)).toEqual(['tail'])
  const custom = { stat: { legendary: 2 }, weapons: [{ name: 'Меч' }, { name: 'Лук' }] }
  expect(legendaryActs(custom)).toEqual([{ key: 'w0', name: 'Атака: Меч', cost: 1, weapon: 0 }, { key: 'w1', name: 'Атака: Лук', cost: 1, weapon: 1 }])
  expect(legendaryActs({ stat: { legendary: 0 }, weapons: [{ name: 'x' }] })).toEqual([])
})

test('остаток очков легендарных действий', () => {
  expect(legendaryLeft({ stat: { legendary: 3 }, counters: { leg: 2 } })).toBe(1)
  expect(legendaryLeft({ stat: { legendary: 3 } })).toBe(3)
  expect(legendaryLeft({ stat: { legendary: 1 }, counters: { leg: 5 } })).toBe(0)
})
