export const fmt = (n) => ((n ?? 0) > 0 ? '+' : '') + (n ?? 0)

export const DMG = { slashing: 'рубящий', piercing: 'колющий', bludgeoning: 'дробящий', fire: 'огонь', cold: 'холод', lightning: 'молния', thunder: 'звук', poison: 'яд', acid: 'кислота', necrotic: 'некротический', radiant: 'излучение', psychic: 'психический', force: 'силовое поле' }

export const KINDS = { beast: 'Звери', humanoid: 'Гуманоиды', undead: 'Нежить', dragon: 'Драконы', giant: 'Великаны', fiend: 'Исчадия', fey: 'Феи', construct: 'Конструкты', ooze: 'Слизи', elemental: 'Элементали', monstrosity: 'Чудовища', aberration: 'Аберрации', plant: 'Растения', celestial: 'Небожители' }

export const CATS = { weapon: 'Оружие', armor: 'Доспехи', shield: 'Щиты', ammo: 'Боеприпасы', gear: 'Снаряжение', pack: 'Наборы', tool: 'Инструменты', instrument: 'Музыкальные инструменты', game: 'Игры', mount: 'Скакуны и сёдла', vehicle: 'Транспорт', potion: 'Зелья', scroll: 'Свитки', ring: 'Кольца', rod: 'Жезлы', staff: 'Посохи', wand: 'Палочки', magic: 'Чудесные предметы' }

export const RARITY = ['обычный', 'необычный', 'редкий', 'очень редкий', 'легендарный', 'артефакт']

export const ARMOR = { light: 'лёгкий', medium: 'средний', heavy: 'тяжёлый' }

export const MODES = { '': 'описательное (эффект решает мастер)', attack: 'бросок атаки заклинанием', save: 'спасбросок цели', auto: 'без броска (как «Волшебная стрела»)', heal: 'лечение' }

export const crSort = (a) => { const [n, d] = String(a).split('/'); return d ? n / d : parseFloat(n) }

export const COINS = { pp: 'ПМ', gp: 'ЗМ', ep: 'ЭМ', sp: 'СМ', cp: 'ММ' }
export const COIN_NAMES = { pp: 'платиновые', gp: 'золотые', ep: 'электрумовые', sp: 'серебряные', cp: 'медные' }
export const COIN_ORDER = ['pp', 'gp', 'ep', 'sp', 'cp']
export const CRS = ['0', '1/8', '1/4', '1/2', ...Array.from({ length: 30 }, (_, i) => String(i + 1))]
// priceGP: «15 зм», «2 см», «5 мм», «1 000 зм» → золотых (пусто или непонятно → 0)
export function priceGP(p) {
  const m = /([\d\s.,]+)\s*(пм|зм|эм|см|мм)/i.exec(p ?? '')
  if (!m) return 0
  const n = parseFloat(m[1].replace(/\s/g, '').replace(',', '.')) || 0
  return n * { пм: 10, зм: 1, эм: 0.5, см: 0.1, мм: 0.01 }[m[2].toLowerCase()]
}
