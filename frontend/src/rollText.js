// Чистые функции для бросков: текст результата, подсказка для анимации, оставленный d20. Без React – их проверяют тесты.

const KIND_NAMES = { check: 'Проверка', save: 'Спасбросок', skill: 'Навык', roll: 'Бросок' }

/**
 * describe превращает результат в понятные строки: что бросали, что выпало, чем кончилось.
 * @param {object} r результат броска (RollView) или атаки (kind === 'attack')
 * @returns {{title: string, formula: string, outcome: string, tone: string, big: number, bigLabel: string}}
 */
export function describe(r) {
  const rolls = r.rolls ?? []
  const label = r.label ?? ''
  if (r.kind === 'attack') {
    return {
      title: label,
      formula: `d20 [${rolls.join(', ')}] → ${r.total} против КД ${r.target}`,
      outcome: r.hit ? `${r.crit ? 'Критическое попадание! ' : 'Попадание. '}Урон ${r.damage}${r.damageText ? ` (${r.damageText})` : ''}` : 'Промах',
      tone: r.hit ? (r.crit ? 'crit' : 'ok') : 'bad',
      big: r.hit ? r.damage : r.total,
      bigLabel: r.hit ? 'урон' : 'атака',
    }
  }
  const nat = r.sides === 20 ? r.kept : 0
  const adv = r.mode === 'adv' ? 'преимущество' : r.mode === 'dis' ? 'помеха' : ''
  const sum = r.sides === 20 ? `d20 [${rolls.join(', ')}]${adv ? ` – ${adv}, берём ${r.kept}` : ''}` : `[${rolls.join(' + ')}]`
  const bonus = r.bonus ? ` ${r.bonus > 0 ? '+' : '−'} ${Math.abs(r.bonus)}` : ''
  let outcome = ''
  let tone = ''
  if (r.dc > 0) { outcome = `${r.success ? 'Успех' : 'Провал'} – сложность ${r.dc}`; tone = r.success ? 'ok' : 'bad' }
  if (nat === 20) { outcome = ('Натуральная 20! ' + outcome).trim(); tone = 'crit' }
  if (nat === 1) { outcome = ('Натуральная 1. ' + outcome).trim(); tone = 'bad' }
  const title = r.kind === 'roll' ? `Бросок ${label}` : label.includes(':') ? label : `${KIND_NAMES[r.kind] ?? ''} ${label}`.trim()
  return { title, formula: `${sum}${bonus} = ${r.total}`, outcome, tone, big: r.total, bigLabel: '' }
}

/**
 * diceHint – сколько и каких кубиков показать, пока ждём ответ («2d6+3» → 2 × d6).
 * @param {string} expr
 * @returns {{count: number, sides: number}}
 */
export function diceHint(expr) {
  const m = /^\s*(\d*)d(\d+)/i.exec(expr ?? '')
  if (!m) return { count: 1, sides: 20 }
  return { count: Math.max(1, parseInt(m[1] || '1', 10)), sides: parseInt(m[2], 10) || 20 }
}

/**
 * keptD20 – какой из двух d20 засчитан при преимуществе или помехе.
 * @param {number[]} rolls
 * @param {string} mode '' | 'adv' | 'dis'
 */
export function keptD20(rolls, mode) {
  if (!rolls?.length) return 0
  if (rolls.length < 2) return rolls[0]
  return mode === 'dis' ? Math.min(...rolls) : mode === 'adv' ? Math.max(...rolls) : rolls[0]
}

/**
 * dimmed – какие кубики в лотке отброшены (второй d20 при преимуществе/помехе).
 * @returns {boolean[]}
 */
export function dimmed(r) {
  const rolls = r.rolls ?? []
  if (r.sides !== 20 || rolls.length !== 2) return rolls.map(() => false)
  const keep = rolls.indexOf(r.kept)
  return rolls.map((_, i) => i !== keep)
}

/** toneOf – подсветка одного кубика: натуральные 20 и 1 на d20. */
export const toneOf = (sides, v) => (sides === 20 && v === 20 ? 'crit' : sides === 20 && v === 1 ? 'bad' : '')

/**
 * cubeTurn – поворот кубика d6, при котором грань value смотрит на зрителя.
 * Грани: перед 1, право 3, верх 2, низ 5, лево 4, зад 6.
 */
export function cubeTurn(value) {
  return { 1: [0, 0], 2: [-90, 0], 3: [0, -90], 4: [0, 90], 5: [90, 0], 6: [0, 180] }[value] ?? [0, 0]
}

/** summary – сводка серии атак: сколько попало и общий урон. */
export function attackSummary(list) {
  const hits = list.filter((r) => r.hit)
  return { hits: hits.length, total: list.length, damage: hits.reduce((s, r) => s + (r.damage ?? 0), 0), crit: hits.some((r) => r.crit) }
}
