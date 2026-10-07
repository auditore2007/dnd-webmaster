// Портреты рас и существ. Всё рисуется кодом (SVG): никаких внешних картинок и лицензий.
// Каждый портрет – 100×100, лицо по центру, фон задаёт настроение. Функции возвращают разметку <svg>.

const E = (x, y, rx, ry, fill, x2 = '') => `<ellipse cx="${x}" cy="${y}" rx="${rx}" ry="${ry}" fill="${fill}" ${x2}/>`
const C = (x, y, r, fill, x2 = '') => `<circle cx="${x}" cy="${y}" r="${r}" fill="${fill}" ${x2}/>`
const P = (d, fill, x2 = '') => `<path d="${d}" fill="${fill}" ${x2}/>`
const S = (d, stroke, w = 2, x2 = '') => `<path d="${d}" fill="none" stroke="${stroke}" stroke-width="${w}" stroke-linecap="round" stroke-linejoin="round" ${x2}/>`
const both = (s) => s + `<g transform="translate(100 0) scale(-1 1)">${s}</g>`

// mix смешивает два цвета #rrggbb в пропорции t (0 – первый, 1 – второй).
function mix(a, b, t) {
  const n = (h) => [1, 3, 5].map((i) => parseInt(h.slice(i, i + 2), 16))
  const [x, y] = [n(a), n(b)]
  return '#' + x.map((v, i) => Math.round(v + (y[i] - v) * t).toString(16).padStart(2, '0')).join('')
}
const dark = (c, t = 0.3) => mix(c, '#000000', t)
const light = (c, t = 0.3) => mix(c, '#ffffff', t)

// ---------- части лица ----------
const eyes = (color = '#1d1a18', y = 44) => both(E(41.5, y, 3, 2.2, '#f4efe2') + C(42, y, 1.6, color) + C(42.4, y - 0.5, 0.5, '#fff'))
const glowEyes = (color, y = 44) => both(E(41.5, y, 3.2, 2, color) + E(41.5, y, 0.9, 1.9, '#140c08'))
const slitEyes = (color, y = 44) => both(E(41.5, y, 3.4, 2.4, color) + E(41.5, y, 0.8, 2.2, '#100804'))
const brows = (c, tilt = 1.5, y = 39) => both(S(`M35 ${y - tilt} L47 ${y + tilt}`, c, 2.4))
const nose = (skin) => S('M50 45 L48.2 52 L51.8 52', dark(skin, 0.35), 1.4)
const smile = () => S('M44.5 56 Q50 59.5 55.5 56', '#5a2a2a', 1.6)
const flat = () => S('M45 57 L55 57', '#4a2a2a', 1.6)
const grin = (fang = false) => P('M43 55.5 Q50 62 57 55.5 Q50 57.5 43 55.5Z', '#3a1414') + P('M44.5 56 L46.5 56.6 L45.5 59.5Z M55.5 56 L53.5 56.6 L54.5 59.5Z', '#f6f1e1', fang ? '' : 'opacity=".0"')
const teeth = () => P('M42 55 Q50 64 58 55Z', '#2b0f0f') + P('M44 55.4 l1.6 3.4 l1.6 -3 M49.2 56 l1.6 3.6 l1.6 -3.6 M54.4 55.4 l-1.6 3 l-1.6 -3', '#f4efe2')

// ---------- волосы, бороды, шлемы, рога ----------
const hairShort = (c) => P('M32 41 C30 23 42 18 50 18 C58 18 70 23 68 41 C65 32 59 29 50 29 C41 29 35 32 32 41Z', c)
const hairBack = (c, len = 80) => P(`M29 42 C26 22 40 14 50 14 C60 14 74 22 71 42 L73 ${len} L27 ${len}Z`, c)
const hairCurly = (c) => [[34, 28], [42, 22], [50, 20], [58, 22], [66, 28], [31, 37], [69, 37]].map(([x, y]) => C(x, y, 6.5, c)).join('')
const mohawk = (c) => P('M44 31 L46 9 L50 15 L54 8 L56 31Z', c)
const beard = (c, len = 74) => P(`M32 47 C32 62 40 ${len - 4} 50 ${len} C60 ${len - 4} 68 62 68 47 C63 55 58 58 50 58 C42 58 37 55 32 47Z`, c)
const braidBeard = (c) => beard(c, 86) + S('M42 76 L50 88 L58 76 M50 62 L50 86', dark(c, 0.3), 1.2) + both(C(40, 80, 2.4, '#caa24a'))
const stubble = () => P('M33 48 C34 60 41 66 50 67 C59 66 66 60 67 48 C62 56 58 58 50 58 C42 58 38 56 33 48Z', 'rgba(30,20,10,.28)')
const helm = (c, nasal = true) => P('M30 43 C28 21 40 13 50 13 C60 13 72 21 70 43 L65 43 C65 33 60 28 50 28 C40 28 35 33 35 43Z', c) + P('M30 36 L70 36 L70 40 L30 40Z', dark(c, 0.15)) + (nasal ? P('M47.5 28 L52.5 28 L52 50 L48 50Z', c) : '')
const horns = (c, big = 1) => both(P(`M37 27 C28 22 22 ${14 / big} 22 4 C28 ${12 / big} 34 14 41 22Z`, c) + S('M37 26 C30 21 25 14 24 6', light(c, 0.35), 1))
const curlHorns = (c) => both(P('M36 27 C22 28 14 20 18 8 C22 18 30 20 40 22Z', c) + P('M18 8 C20 4 24 5 25 8 C23 7 20 7 18 8Z', light(c, 0.5)))
const tusks = (c = '#f1ead6', len = 7) => both(P(`M42 55 L40 ${55 - len} L45 ${57}Z`, c))
const earsPointy = (skin, len = 17, y = 41) => both(P(`M35 ${y} L${35 - len} ${y - 11} L34 ${y + 12}Z`, skin) + P(`M33 ${y + 2} L${35 - len + 4} ${y - 8} L33 ${y + 9}Z`, dark(skin, 0.18)))
const earsGoblin = (skin) => both(P('M35 40 L10 30 L12 42 L35 54Z', skin) + P('M31 42 L16 35 L17 43 L31 50Z', dark(skin, 0.2)))
const earsRound = (c, inner, y = 30, x = 32, r = 7) => both(C(x, y, r, c) + C(x, y, r / 2, inner))
const earsTri = (c, inner, h = 23) => both(P(`M33 36 L29 ${h - 14} L45 27Z`, c) + P(`M35 32 L33 ${h - 6} L42 28Z`, inner))
const shoulders = (cloth, collar) => P('M8 100 C10 76 28 68 50 68 C72 68 90 76 92 100Z', cloth) + S('M37 70 L50 86 L63 70', collar ?? dark(cloth, 0.35), 2.2)
const scar = () => S('M58 33 L65 52', 'rgba(120,40,40,.7)', 1.6) + S('M60 40 l4 -1 M62 46 l4 -1', 'rgba(120,40,40,.6)', 1)
const scales = (c) => [[40, 33], [46, 31], [52, 31], [58, 33], [37, 52], [63, 52], [36, 44], [64, 44]].map(([x, y]) => S(`M${x - 2.4} ${y} q2.4 3 4.8 0`, c, 1)).join('')

// ---------- сборка портрета ----------
function bust(key, o) {
  const skin = o.skin
  const sh = o.shade ?? dark(skin, 0.22)
  const [b1, b2] = o.bg ?? ['#3d5266', '#1b2733']
  const hw = o.hw ?? 17.5, hh = o.hh ?? 20.5
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100" role="img" aria-label="${o.label ?? ''}"><defs>
<radialGradient id="${key}b" cx="50%" cy="32%" r="80%"><stop offset="0" stop-color="${b1}"/><stop offset="1" stop-color="${b2}"/></radialGradient>
<linearGradient id="${key}h" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="${light(skin, 0.1)}"/><stop offset="1" stop-color="${sh}"/></linearGradient></defs>
<rect width="100" height="100" fill="url(#${key}b)"/>${o.aura ?? ''}${o.back ?? ''}${o.ears ?? ''}${o.noBody ? '' : shoulders(o.cloth ?? '#55627a', o.collar)}${o.body ?? ''}
${o.noBody ? '' : P('M43 58 L43 72 Q50 77 57 72 L57 58Z', sh)}${o.head ?? E(50, 44.5, hw, hh, `url(#${key}h)`)}${o.face ?? ''}${o.front ?? ''}</svg>`
}

const human = (skin, face = '') => ({ skin, face: brows(dark(skin, 0.5)) + eyes() + nose(skin) + smile() + face })

// ---------- расы ----------
const dragonHead = (c, accent = '#e9d7a6') => ({
  skin: c, hw: 17, hh: 20,
  back: both(P('M33 30 C24 20 18 8 14 0 C26 6 36 14 42 24Z', accent) + P('M30 36 L18 34 L30 44Z M31 44 L19 46 L31 52Z', dark(c, 0.15))),
  head: E(50, 44, 17, 20, `url(#DRh)`) + E(50, 57, 12.5, 9.5, light(c, 0.12)) + P('M44 28 L50 20 L56 28Z', dark(c, 0.15)),
  face: scales(dark(c, 0.3)) + slitEyes('#f2c53d') + both(E(46, 55, 1.1, 1.6, dark(c, 0.55))) + S('M42 61 Q50 64 58 61', dark(c, 0.5), 1.4) + brows(dark(c, 0.45), 2),
})
function dragonborn(key, c, accent, extra = {}) {
  const o = dragonHead(c, accent)
  o.head = o.head.replace('DR', key)
  return { ...o, ...extra }
}

const RACE = {
  'dnd5e:dragonborn': (k) => bust(k, { ...dragonborn(k, '#b4472e', '#e9d7a6'), bg: ['#6a3a2a', '#26120d'], cloth: '#6b7280', label: 'Драконорождённый' }),
  'dnd5e:dwarf': (k) => bust(k, { ...human('#dba075'), bg: ['#5a4a3a', '#221a12'], cloth: '#6b4a2f', back: hairBack('#7a3a1c', 70), front: hairShort('#7a3a1c') + braidBeard('#8a421f') + helm('#7d8590', false) + P('M40 34 L60 34 L60 38 L40 38Z', '#caa24a'), label: 'Дварф' }),
  'dnd5e:elf': (k) => bust(k, { ...human('#f1d6b4'), bg: ['#35603f', '#101f15'], cloth: '#2f6a4c', ears: earsPointy('#f1d6b4', 19), back: hairBack('#e9e4d0', 84), front: hairShort('#e9e4d0') + P('M38 25 Q50 17 62 25 L60 29 Q50 22 40 29Z', '#caa24a'), label: 'Эльф' }),
  'dnd5e:gnome': (k) => bust(k, { ...human('#e4b088'), bg: ['#3c5a7a', '#14202f'], cloth: '#7a4a8a', ears: earsPointy('#e4b088', 9, 43), face: brows('#d8d0c0', 2) + eyes('#2a4a8a') + P('M46 45 q4 -1 8 0 q1 6 -4 8 q-5 -2 -4 -8Z', '#d9987a') + smile(), front: beard('#eeeae0', 68) + P('M30 31 C32 12 44 4 50 -2 C56 4 68 12 70 31 C62 24 38 24 30 31Z', '#c3453a') + P('M30 31 C38 25 62 25 70 31 L70 35 C60 29 40 29 30 35Z', '#8a2a24'), label: 'Гном' }),
  'dnd5e:halfelf': (k) => bust(k, { ...human('#ecc9a3'), bg: ['#4a6a80', '#16222c'], cloth: '#2e6f78', ears: earsPointy('#ecc9a3', 10), back: hairBack('#9a4a2a', 76), front: hairShort('#9a4a2a') + P('M30 40 C34 30 40 28 46 28 C40 34 36 40 35 52Z', '#9a4a2a'), label: 'Полуэльф' }),
  'dnd5e:halforc': (k) => bust(k, { ...human('#85a064', 'x'), bg: ['#4b5a3a', '#161b10'], cloth: '#5a4636', front: mohawk('#1c1a18') + tusks() + scar(), label: 'Полуорк' }),
  'dnd5e:halfling': (k) => bust(k, { ...human('#eec9a2'), hw: 18.5, hh: 19.5, bg: ['#6a7a3a', '#232b12'], cloth: '#6d8a3a', ears: earsPointy('#eec9a2', 8, 44), front: hairCurly('#6a3a1c') + both(S('M32 52 l-2 5', '#6a3a1c', 3)), face: undefined, label: 'Полурослик' }),
  'dnd5e:human': (k) => bust(k, { ...human('#e8c19a'), bg: ['#4a5f86', '#151c2e'], cloth: '#3f5a9a', collar: '#e6dcc0', back: '', front: hairShort('#5a3a22') + stubble(), label: 'Человек' }),
  'dnd5e:tiefling': (k) => bust(k, { ...human('#a8456a'), bg: ['#5a2a52', '#1a0c1c'], cloth: '#43275a', back: hairBack('#161018', 76) + curlHorns('#e8d5b0'), front: hairShort('#161018'), face: brows('#2a0f1a', 2.4) + glowEyes('#f4d03a') + nose('#a8456a') + smile(), label: 'Тифлинг' }),
}
// Полурослик: лицо задаётся отдельно, чтобы не перекрывать кудри
RACE['dnd5e:halfling'] = (k) => bust(k, { skin: '#eec9a2', hw: 18.5, hh: 19.5, bg: ['#6a7a3a', '#232b12'], cloth: '#6d8a3a', ears: earsPointy('#eec9a2', 8, 44), front: hairCurly('#6a3a1c') + both(S('M33 52 l-2 6', '#6a3a1c', 3)), face: brows('#4a2a14', 2.5, 40) + eyes() + nose('#eec9a2') + smile() + both(C(36, 52, 3, 'rgba(210,90,80,.3)')), label: 'Полурослик' })
RACE['dnd5e:halforc'] = (k) => bust(k, { skin: '#85a064', bg: ['#4b5a3a', '#161b10'], cloth: '#5a4636', hw: 19, face: brows('#1a2210', 3.5, 39) + slitEyes('#d8b030') + nose('#85a064') + flat(), front: mohawk('#1c1a18') + tusks() + scar(), label: 'Полуорк' })

const BEAST = {
  wolf: { c: '#8b9097', m: '#cfd2d6', i: '#4a4f58', ears: 'tri', bg: ['#4a5a6a', '#141a22'] },
  fox: { c: '#d9772c', m: '#f5ebd8', i: '#2a1a12', ears: 'tri', bg: ['#6a4a2a', '#22140a'] },
  bear: { c: '#7a5232', m: '#c9a77a', i: '#3a2416', ears: 'round', bg: ['#5a4a32', '#1c140c'] },
  cat: { c: '#d3a35a', m: '#f3e5c4', i: '#7a4a24', ears: 'tri', bg: ['#6a5a2a', '#221a0a'], stripes: true },
}

function beastFolk(k, sub) {
  if (sub === 'eagle') {
    return bust(k, { skin: '#f4f1e8', bg: ['#4a6a88', '#141f2c'], cloth: '#6a4a2e', hw: 17, hh: 20,
      back: both(P('M33 30 L16 24 L30 40Z M30 44 L12 46 L30 54Z', '#7a5a3a')),
      face: both(E(40, 41, 4, 3.6, '#f6c843') + C(40.4, 41, 1.8, '#1a1208')) + both(S('M33 36 L46 40', '#7a5a3a', 2.2)) + P('M43 47 L57 47 L50 66Z', '#f2b83a') + S('M50 47 L50 62', '#b5821c', 1.2),
      front: P('M31 34 C33 20 42 16 50 16 C58 16 67 20 69 34 C62 26 56 24 50 24 C44 24 38 26 31 34Z', '#6a4a2a'), label: 'Орлиный народ' })
  }
  if (sub === 'lizard') {
    const o = dragonHead('#4f9a58', '#d6c27a')
    o.head = o.head.replace('DR', k)
    return bust(k, { ...o, bg: ['#2f5a3a', '#0c1c10'], cloth: '#6a5230', label: 'Ящеролюди' })
  }
  const b = BEAST[sub]
  const ears = b.ears === 'round' ? earsRound(b.c, b.i, 30, 31, 7.5) : earsTri(b.c, b.i)
  return bust(k, { skin: b.c, bg: b.bg, cloth: '#5a4a3a', ears,
    head: E(50, 44.5, 18.5, 19, `url(#${k}h)`) + E(50, 56, 11, 8.5, b.m),
    face: both(E(41, 43, 3.4, 2.3, '#f2d36a') + E(41, 43, 0.9, 2.1, '#140c08')) + both(S('M34 38 L46 41', dark(b.c, 0.5), 2)) + E(50, 52, 3.6, 2.6, '#1a1210') + S('M50 54 L50 59 M44 60 Q50 63 56 60', '#2a1a14', 1.4)
      + (b.stripes ? both(S('M33 44 l5 1 M33 49 l5 -1', dark(b.c, 0.45), 1.4)) : '') + both(S('M38 55 l-9 -1.5 M38 57.5 l-9 2', 'rgba(255,255,255,.7)', 0.8)),
    label: 'Зверолюди' })
}

export function raceArt(ruleset, race, sub) {
  const k = `r-${race}-${sub || 'x'}`
  return (RACE[`${ruleset}:${race}`] ?? RACE['dnd5e:human'])(k, sub)
}

// ---------- существа ----------
const beastish = (k, b) => bust(k, { skin: b.c, bg: b.bg, noBody: true, ears: b.ears, back: b.back ?? '',
  head: E(50, 46, b.hw ?? 21, b.hh ?? 19, `url(#${k}h)`) + E(50, 58, b.mw ?? 12, b.mh ?? 9, b.m), face: b.face, front: b.front ?? '', label: b.label })
const furEyes = (c = '#f2c43a') => both(E(40.5, 43, 3.2, 2.2, c) + E(40.5, 43, 0.9, 2, '#120a06')) + both(S('M33 37.5 L46 41', '#1a100a', 2.2))
const MON = {
  commoner: (k) => bust(k, { ...human('#e4bd96'), bg: ['#6a6a50', '#1e1e14'], cloth: '#7a6a4a', front: hairShort('#6a5030') + stubble(), label: 'Обыватель' }),
  'giant-rat': (k) => beastish(k, { c: '#6f6258', m: '#a89888', bg: ['#4a4038', '#16120e'], label: 'Гигантская крыса', hw: 19, hh: 17, mw: 9, mh: 11,
    ears: both(C(30, 30, 10, '#6f6258') + C(30, 30, 6, '#d9a0a0')),
    face: both(E(40, 42, 2.6, 2.6, '#14100e') + C(40.6, 41.4, 0.8, '#fff')) + E(50, 54, 3, 2.4, '#d98a8a') + P('M46 60 h3.4 v6 h-3.4Z M50.6 60 h3.4 v6 h-3.4Z', '#f1ead0') + both(S('M40 56 l-16 -3 M40 58 l-16 1 M40 60 l-14 6', '#e8e0d0', 0.9)) }),
  kobold: (k) => bust(k, { skin: '#a8542e', hw: 15.5, hh: 18, bg: ['#5a3a2a', '#1c100a'], cloth: '#6a5a40', back: both(P('M37 28 L30 10 L43 24Z', '#e0cfa0')),
    head: E(50, 45, 15.5, 18, `url(#${k}h)`) + E(50, 56, 9.5, 8, '#c06a3a'),
    face: slitEyes('#f2a02a', 43) + both(E(46, 54, 1, 1.5, '#401a0c')) + S('M43 60 Q50 63 57 60', '#401a0c', 1.4) + P('M45 60 l1 2.4 l1 -2.2Z', '#fff'), label: 'Кобольд' }),
  bandit: (k) => bust(k, { ...human('#d9a982'), bg: ['#4a4038', '#14100c'], cloth: '#3f3a34',
    face: brows('#2a1a10', 2.5) + eyes('#22160c') + P('M32 48 C34 62 42 66 50 66 C58 66 66 62 68 48 C62 52 58 52 50 52 C42 52 38 52 32 48Z', '#6a1f1f'),
    front: P('M30 38 C30 18 42 13 50 13 C58 13 70 18 70 38 C64 30 58 27 50 27 C42 27 36 30 30 38Z', '#4a3a2e') + P('M30 38 L70 38 L70 33 C64 27 36 27 30 33Z', '#8a2a24'), label: 'Бандит' }),
  cultist: (k) => bust(k, { skin: '#2a1814', hw: 14.5, hh: 18, bg: ['#5a1a22', '#12060a'], cloth: '#4a1a24', collar: '#a83a3a',
    back: P('M20 100 C20 50 30 14 50 12 C70 14 80 50 80 100Z', '#3a1018'),
    head: E(50, 46, 14.5, 18, '#1e0e0c'), face: both(E(43, 44, 2.6, 1.6, '#e8302a')) + S('M44 56 Q50 58 56 56', '#6a2a2a', 1.2),
    front: P('M30 52 C28 20 40 14 50 14 C60 14 72 20 70 52 L64 52 C66 28 60 22 50 22 C40 22 34 28 36 52Z', '#3a1018') + S('M50 74 l0 12 M44 80 h12', '#d8b04a', 1.6), label: 'Культист' }),
  guard: (k) => bust(k, { ...human('#e0b690', 'x'), bg: ['#3a4a5c', '#12181f'], cloth: '#9a2f2f', collar: '#e6dcc0',
    face: brows('#3a2414', 2.5) + eyes() + nose('#e0b690') + flat(), front: helm('#8a9098') + stubble(), label: 'Стражник' }),
  goblin: (k) => bust(k, { skin: '#74a038', hw: 15.5, hh: 17.5, bg: ['#4a5a28', '#141a0a'], cloth: '#6a4a2a', ears: earsGoblin('#74a038'),
    face: brows('#1e2a0c', 3.2) + glowEyes('#f2cf2a', 43) + P('M46 45 q4 -1 8 0 q1 7 -4 9 q-5 -2 -4 -9Z', '#5d8a2c') + teeth(),
    front: P('M34 33 C38 24 62 24 66 33 C58 28 42 28 34 33Z', '#2a2012'), label: 'Гоблин' }),
  skeleton: (k) => bust(k, { skin: '#e6e0cc', shade: '#a89f84', hw: 16, hh: 19, bg: ['#3a3f4c', '#0e1016'], cloth: '#0000', collar: '#0000',
    body: P('M12 100 C14 78 30 70 50 70 C70 70 88 78 88 100Z', '#cfc8b0') + S('M30 100 C32 88 40 82 50 82 M70 100 C68 88 60 82 50 82 M50 72 L50 100 M36 92 H64', '#6a6450', 1.6),
    head: E(50, 43, 17, 19.5, `url(#${k}h)`) + P('M38 58 L38 66 L62 66 L62 58Z', '#d6cfb6'),
    face: both(E(42, 43, 4.2, 5, '#161214')) + P('M50 49 L47 55 L53 55Z', '#161214') + S('M40 62 H60 M44 58 V66 M50 58 V66 M56 58 V66', '#6a6450', 1.1), label: 'Скелет' }),
  zombie: (k) => bust(k, { ...human('#7f9a76', 'x'), bg: ['#3a4a3a', '#0e140e'], cloth: '#4a4238', collar: '#2a241c',
    face: brows('#2a3a2a', -0.5) + both(E(41.5, 44, 3, 2.4, '#e6e8c8')) + C(42, 44, 1.1, '#555') + E(58.5, 44, 3, 2.4, '#e6e8c8') + C(58, 45, 0.8, '#333') + S('M44 57 Q50 55 56 58', '#3a1a1a', 1.8) + S('M34 38 l7 4 m1 -2 l-4 6 M60 56 l8 -3 M62 53 v6 M65 52 v6', '#3a2a28', 1.2) + P('M52 36 q8 2 10 10 q-6 -1 -10 -10Z', 'rgba(110,30,30,.65)'),
    front: P('M34 31 C38 24 44 24 46 30 C42 28 38 29 34 36Z M58 28 C62 25 67 29 67 35 C64 31 60 30 58 28Z', '#33332a'), label: 'Зомби' }),
  wolf: (k) => beastish(k, { ...BEAST.wolf, bg: BEAST.wolf.bg, label: 'Волк', ears: earsTri('#8b9097', '#4a4f58'), c: '#8b9097', m: '#cfd2d6',
    face: furEyes() + E(50, 52, 4, 3, '#14100e') + S('M50 55 L50 60 M43 61 Q50 65 57 61', '#2a1a14', 1.4) + P('M44 61 l1.4 4 l1.4 -3.4 M56 61 l-1.4 4 l-1.4 -3.4', '#fff') }),
  gnoll: (k) => beastish(k, { c: '#b99458', m: '#2a2018', bg: ['#6a5a2a', '#1c1608'], label: 'Гнолл', hw: 20,
    ears: earsRound('#b99458', '#3a2a1a', 29, 30, 8.5), back: P('M42 12 L58 12 L66 40 L34 40Z', '#3a2c1a'),
    face: furEyes('#e8a02a') + E(50, 52, 4, 3, '#0e0a08') + P('M40 58 Q50 70 60 58 Q50 61 40 58Z', '#3a1010') + P('M42 59 l1.5 3.4 l1.5 -2.8 M55 59.4 l1.5 -3 l1.5 3.4', '#f4efe2') + [[32, 38], [36, 48], [64, 36], [66, 46], [44, 31], [56, 30]].map(([x, y]) => C(x, y, 1.7, '#3a2c1a')).join('') }),
  hobgoblin: (k) => bust(k, { skin: '#c4713a', hw: 17.5, bg: ['#6a3a22', '#1c0e08'], cloth: '#7a2f26', collar: '#caa24a',
    face: brows('#2a1408', 3) + glowEyes('#f0d040') + P('M45 47 h10 l-1 5 h-8Z', '#a85a2a') + both(P('M43 55 l-1.5 -5 l3.4 2.4Z', '#f1ead6')) + flat(),
    front: helm('#7d6a56') + P('M44 12 L56 12 L54 20 L46 20Z', '#a82a24'), label: 'Хобгоблин' }),
  orc: (k) => bust(k, { skin: '#6f8f4a', hw: 19.5, hh: 21, bg: ['#4a3a28', '#16100a'], cloth: '#4a3a2f', collar: '#2a2018',
    face: brows('#10160a', 4, 38.5) + slitEyes('#e2b82a') + nose('#6f8f4a') + flat() + both(S('M37 46 l5 3', '#b8372a', 1.8)),
    front: P('M36 30 C40 14 60 14 64 30 C56 25 44 25 36 30Z', '#1e1a16') + tusks('#f1ead6', 8) + both(C(31, 54, 2, '#caa24a', 'stroke="#7a5a1c" stroke-width=".5"')), label: 'Орк' }),
  bugbear: (k) => bust(k, { skin: '#9a7a46', hw: 19.5, hh: 21, bg: ['#5a4a2a', '#1a1408'], cloth: '#4a3a24', collar: '#2a2014',
    ears: earsRound('#9a7a46', '#4a3a24', 33, 31, 8), back: hairBack('#4a3220', 70),
    face: brows('#1e1408', 3.5, 38) + glowEyes('#e2c02a') + P('M44 46 h12 l1.5 8 q-7.5 3 -15 0Z', '#7a5a30') + both(E(46.5, 52, 1.4, 1.8, '#1a1008')) + P('M42 58 Q50 66 58 58Z', '#3a1010') + P('M44 58.4 l1.2 3 l1.2 -2.6 M53.6 58.4 l1.2 2.6 l1.2 -3', '#fff'), front: P('M32 36 C36 22 64 22 68 36 C60 30 40 30 32 36Z', '#4a3220'), label: 'Багбир' }),
  'brown-bear': (k) => beastish(k, { ...BEAST.bear, label: 'Бурый медведь', ears: earsRound('#7a5232', '#c9a77a', 28, 30, 8), c: '#7a5232', m: '#c9a77a', hw: 22, hh: 20, mw: 12, mh: 10,
    face: both(C(39.5, 42, 2.3, '#140c08') + C(40, 41.4, 0.6, '#fff')) + E(50, 53, 5.2, 3.8, '#14100e') + S('M50 56 L50 62 M42 63 Q50 67 58 63', '#2a1a14', 1.5) }),
  'giant-spider': (k) => bust(k, { skin: '#2a2230', noBody: true, bg: ['#3a2a44', '#0c0812'], aura: [20, 40, 60, 80].map((x) => S(`M${x} 0 L${50} 50`, '#c9c4d4', 0.5, 'opacity=".35"')).join(''),
    back: [[-10, 14], [-4, 30], [-2, 52], [4, 72]].map(([dx, y]) => both(S(`M42 54 Q${20 + dx} ${y - 10} ${8 + dx} ${y + 16}`, '#1d1824', 3.2) + S(`M42 54 Q${20 + dx} ${y - 10} ${8 + dx} ${y + 16}`, '#4a3a58', 1, 'opacity=".6"'))).join('') + E(50, 70, 17, 21, '#201a28') + S('M50 56 L50 88', '#12101a', 1.4) + P('M45 70 q5 -5 10 0 q-5 9 -10 0Z', '#b8342c'),
    head: E(50, 42, 16, 14, '#2c2434'),
    face: [[44, 37, 2.4], [56, 37, 2.4], [39, 41, 1.8], [61, 41, 1.8], [46, 42.5, 1.5], [54, 42.5, 1.5], [42, 33, 1.2], [58, 33, 1.2]].map(([x, y, r]) => C(x, y, r, '#c9302c') + C(x - 0.4, y - 0.4, r * 0.35, '#fff')).join('') + both(P('M44 48 L41 60 L47 51Z', '#d9d2c0')), label: 'Гигантский паук' }),
  'dire-wolf': (k) => beastish(k, { c: '#4a4d55', m: '#8d9199', bg: ['#3a2a2a', '#120a0a'], label: 'Лютый волк', hw: 22, hh: 20, mw: 13, mh: 10, ears: earsTri('#4a4d55', '#2a1a1a', 21),
    face: furEyes('#e0392f') + E(50, 52, 4.4, 3.2, '#0e0a08') + P('M40 60 Q50 72 60 60 Q50 63 40 60Z', '#400c0c') + P('M43 60.4 l1.6 5 l1.6 -4.4 M54.8 61 l1.6 4.4 l1.6 -5', '#fff') + S('M34 30 l6 12 M36 28 l6 12', 'rgba(150,40,40,.7)', 1.2) }),
  ghoul: (k) => bust(k, { skin: '#a2a8a0', hw: 15.5, hh: 19.5, bg: ['#323c3c', '#0a1010'], cloth: '#3a3430', collar: '#1a1612',
    face: both(E(42, 44, 4, 4.6, '#0c0a0a') + C(42, 44.4, 1.1, '#c8ffa0')) + S('M38 52 Q50 50 62 52 M38 36 l9 3', '#58605a', 1.5) + P('M41 57 Q50 69 59 57Z', '#2a0c0c') + P('M43 57.4 l1.3 5 l1.3 -4.6 M47.6 58.4 l1.3 5.4 l1.4 -5.2 M52.4 58.4 l1.4 5.2 l1.3 -5.4 M55.6 57.4 l1.4 4.6 l1.3 -5', '#e8e4d0'),
    front: P('M34 32 C38 22 48 22 52 28 L48 40 C44 32 38 32 34 40Z M54 26 C60 24 66 28 66 40 C62 34 58 32 56 36Z', '#4a4a44') , label: 'Вурдалак' }),
  ogre: (k) => bust(k, { skin: '#b6a26a', hw: 21, hh: 22, bg: ['#5a4a2a', '#1a1408'], cloth: '#6a4a2a', collar: '#2a1c0c', ears: earsRound('#b6a26a', '#8a7648', 44, 29, 6),
    face: both(S('M33 37 L46 41', '#2a1c0c', 3)) + both(E(41.5, 44, 2.6, 1.9, '#f0e8c8') + C(41.5, 44.4, 1.2, '#2a1c0c')) + E(50, 51, 5, 4, '#9a8650') + P('M39 58 Q50 54 61 58 L60 64 Q50 70 40 64Z', '#7a6636') + both(P('M42 58 l-1 -6 l4.4 4Z', '#f1ead6')) + C(36, 52, 1.5, '#8a7648') + C(63, 49, 1.8, '#8a7648'),
    front: P('M32 34 C34 20 44 18 50 22 C56 18 66 20 68 34 C60 27 40 27 32 34Z', '#3a2a14'), label: 'Огр' }),
  owlbear: (k) => beastish(k, { c: '#7a6048', m: '#d8c8a0', bg: ['#4a4030', '#16120a'], label: 'Совомедведь', hw: 22, hh: 21, mw: 10, mh: 8,
    ears: both(P('M34 34 L28 8 L46 28Z', '#5a4630') + P('M36 32 L32 14 L43 28Z', '#c8b088')),
    face: both(E(39.5, 43, 7.5, 7.5, '#f4e8c0') + E(39.5, 43, 5.5, 5.5, '#f0b829') + C(39.5, 43, 2.6, '#0a0806') + C(38.6, 42, 0.8, '#fff')) + P('M45 50 L55 50 L50 66Z', '#e5a92c') + S('M50 52 L50 63', '#9a6a14', 1)
      + [[34, 62], [66, 62], [28, 52], [72, 52]].map(([x, y]) => S(`M${x} ${y} l${x < 50 ? -6 : 6} 4`, '#d8c8a0', 1.2)).join(''), }),
  minotaur: (k) => beastish(k, { c: '#6a4a34', m: '#8a6a50', bg: ['#5a3a2a', '#1a0e08'], label: 'Минотавр', hw: 20, hh: 21, mw: 13, mh: 11,
    back: both(P('M34 32 C16 30 6 18 8 4 C14 14 22 18 38 22Z', '#e8dcc0') + P('M8 4 C10 0 14 1 15 5 C13 4 10 4 8 4Z', '#b8a47a')),
    ears: both(P('M32 38 L14 34 L16 44 L33 48Z', '#6a4a34') + P('M30 40 L19 38 L20 43 L30 45Z', '#c78a7a')),
    face: furEyes('#e2b02a') + both(S('M36 36 L46 41', '#120a06', 2.6)) + E(50, 56, 9, 6.5, '#9a7a60') + both(E(46, 56, 1.8, 2.4, '#14100e')) + S('M43 62 Q50 65 57 62', '#2a1a14', 1.4)
      + `<circle cx="50" cy="65" r="3.4" fill="none" stroke="#e0b94a" stroke-width="1.5"/>`, front: P('M38 20 C42 14 58 14 62 20 C58 24 42 24 38 20Z', '#3a2418') }),
  veteran: (k) => bust(k, { ...human('#d9a982'), bg: ['#47505c', '#12161c'], cloth: '#6a7480', collar: '#caa24a',
    face: brows('#6a6a68', 3) + eyes() + nose('#d9a982') + flat() + scar(), front: helm('#8a9098') + P('M33 48 C34 60 41 66 50 67 C59 66 66 60 67 48 C62 54 58 56 50 56 C42 56 38 54 33 48Z', '#8a8a86'), label: 'Ветеран' }),
  knight: (k) => bust(k, { skin: '#aab2bc', shade: '#6a727c', hw: 18.5, hh: 22, bg: ['#3a4a6a', '#0e1220'], cloth: '#8a929c', collar: '#caa24a',
    body: S('M30 80 L70 80 M50 70 L50 100', '#5a626c', 1.6) + both(C(24, 82, 6, '#9aa2ac', 'stroke="#5a626c" stroke-width="1"')),
    head: P('M31 46 C30 20 40 12 50 12 C60 12 70 20 69 46 L66 62 Q50 70 34 62Z', 'url(#' + k + 'h)'),
    face: P('M34 40 L66 40 L66 45 L34 45Z', '#10121a') + P('M48 45 L52 45 L52 62 L48 62Z', '#10121a') + both(S('M36 54 L44 54 M36 58 L44 58', '#10121a', 1.6)),
    front: P('M44 12 C42 2 58 -2 62 6 C56 4 52 8 54 13Z', '#b8342c'), label: 'Рыцарь' }),
  troll: (k) => bust(k, { skin: '#5f8a4c', hw: 19, hh: 22, bg: ['#2f4a2a', '#0a140a'], cloth: '#5a4a34', collar: '#2a2014', ears: earsPointy('#5f8a4c', 14, 40),
    face: both(S('M34 38 L46 42', '#142410', 3)) + glowEyes('#f2c82a', 44) + P('M47 44 C45 54 43 60 50 66 C57 60 55 54 53 44Z', '#4a7438') + C(40, 52, 1.4, '#4a6a30') + C(60, 48, 1.8, '#4a6a30') + C(36, 40, 1.2, '#4a6a30') + P('M41 60 Q50 70 59 60 Q50 63 41 60Z', '#2a0c0c') + P('M44 61 l1 4.4 l1.4 -3.6 M54 61.6 l-1 4 l-1.4 -3.4', '#e8e4d0'),
    front: P('M36 28 C40 14 48 16 50 24 C54 14 62 14 64 28 C58 22 42 22 36 28Z', '#3a6a2a') + S('M40 26 l-3 -8 M60 26 l3 -8', '#3a6a2a', 2), label: 'Тролль' }),
  'hill-giant': (k) => bust(k, { skin: '#cf8e68', hw: 21, hh: 23, bg: ['#6a5a3a', '#1c1608'], cloth: '#7a5a3a', collar: '#3a2a18',
    body: P('M8 100 C10 78 24 70 36 70 L50 80 L64 70 C76 70 90 78 92 100Z', '#a98a5c'),
    ears: earsRound('#cf8e68', '#a46c4c', 44, 28.5, 6),
    face: both(S('M33 37 L46 40.5', '#3a1c0c', 3.4)) + both(E(41.5, 44, 2.6, 1.9, '#f0e4c8') + C(41.5, 44.4, 1.2, '#2a1a0c')) + P('M45 46 h10 l1 7 h-12Z', '#b97a56') + flat(),
    front: P('M31 40 C28 16 44 14 50 14 C56 14 72 16 69 40 C64 28 36 28 31 40Z', '#4a2a14') + P('M30 46 C28 66 40 80 50 84 C60 80 72 66 70 46 C64 58 58 60 50 60 C42 60 36 58 30 46Z', '#4a2a14') + P('M41 60 Q50 54 59 60 Q50 64 41 60Z', '#cf8e68'), label: 'Холмовой великан' }),
  wyvern: (k) => bust(k, { skin: '#7d8a3c', noBody: true, bg: ['#4a4a2a', '#14140a'],
    back: both(P('M34 50 C14 44 4 24 2 6 C14 14 22 14 30 24 C22 26 20 34 28 38 C24 42 28 46 36 46Z', '#4a5224')),
    head: P('M30 40 C28 20 40 12 50 12 C60 12 72 20 70 40 C70 54 60 66 50 74 C40 66 30 54 30 40Z', 'url(#' + k + 'h)') + both(P('M38 20 L28 4 L44 16Z', '#d8cc9a')),
    face: slitEyes('#f0c030', 36) + both(E(45.5, 57, 1.3, 2, '#2a3010')) + S('M42 66 Q50 71 58 66', '#2a3010', 1.4) + P('M45 66 l1 3.4 l1.4 -3Z M55 66 l-1 3.4 l-1.4 -3Z', '#fff') + scales('#5a6a2a') + S('M50 14 L50 30', '#5a6a2a', 2), label: 'Виверна' }),
  'young-red-dragon': (k) => bust(k, { skin: '#b4382a', noBody: true, bg: ['#7a2a1a', '#1c0604'], aura: E(50, 100, 46, 26, '#f6a02a', 'opacity=".35"'),
    back: both(P('M30 40 C18 36 8 22 6 2 C20 8 34 14 40 26Z', '#8a2418') + P('M32 32 L22 26 M30 38 L16 34', '#e8b84a', 'stroke="#e8b84a" stroke-width="1.2"')),
    head: P('M29 42 C27 20 38 12 50 12 C62 12 73 20 71 42 C71 56 62 70 50 78 C38 70 29 56 29 42Z', 'url(#' + k + 'h)') + both(P('M36 20 C26 16 18 8 14 -2 C26 2 38 8 44 16Z', '#e8d6a4')) + P('M42 70 Q50 74 58 70 L58 80 Q50 90 42 80Z', '#f0983a'),
    face: slitEyes('#f4c82a', 36) + both(E(45, 58, 1.4, 2.2, '#2a0a06')) + S('M40 68 Q50 74 60 68', '#2a0a06', 1.6) + P('M43 68 l1.2 4.4 l1.6 -3.8Z M57 68 l-1.2 4.4 l-1.6 -3.8Z', '#fff') + scales('#8a2418') + both(S('M36 30 L44 36', '#2a0a06', 2.4)), label: 'Молодой красный дракон' }),
}

// =====================================================================
// Расширение: новые расы и портреты существ по типу
// =====================================================================
const hash = (s) => { let h = 2166136261; for (const ch of String(s)) { h ^= ch.charCodeAt(0); h = Math.imul(h, 16777619) } return h >>> 0 }
const pick = (arr, id, salt = '') => arr[hash(id + salt) % arr.length]
const relabel = (svg, label) => svg.replace(/aria-label="[^"]*"/, `aria-label="${label}"`)
const wings = (c, spread = 1) => both(P(`M32 54 C${12 / spread} 48 ${4 / spread} 26 ${6 / spread} 4 C20 14 30 26 36 40Z`, c) + S(`M31 48 C${20 / spread} 42 ${14 / spread} 30 ${12 / spread} 14`, dark(c, 0.25), 1))
const halo = (c = '#ffe08a') => E(50, 12, 17, 4.5, 'none', `stroke="${c}" stroke-width="2.2"`)

// ---------- новые расы ----------
for (const [id, src, name] of [['orc', 'orc', 'Орк'], ['goblin', 'goblin', 'Гоблин'], ['hobgoblin', 'hobgoblin', 'Хобгоблин'], ['bugbear', 'bugbear', 'Багбир'], ['kobold', 'kobold', 'Кобольд'], ['minotaur', 'minotaur', 'Минотавр']]) {
  RACE[`dnd5e:${id}`] = (k) => relabel(MON[src](k), name)
}
RACE['dnd5e:aarakocra'] = (k) => relabel(beastFolk(k, 'eagle'), 'Ааракокра')
RACE['dnd5e:tabaxi'] = (k) => relabel(beastFolk(k, 'cat'), 'Табакси')
RACE['dnd5e:lizardfolk'] = (k) => relabel(beastFolk(k, 'lizard'), 'Ящеролюд')

RACE['dnd5e:aasimar'] = (k, sub) => {
  const fallen = sub === 'fallen', scourge = sub === 'scourge'
  const skin = fallen ? '#b9a8b0' : '#f2dcc2'
  const glow = fallen ? '#d8453a' : scourge ? '#ffb347' : '#fff0a8'
  return bust(k, { skin, bg: fallen ? ['#3a2a44', '#0c0812'] : ['#6a7aa0', '#1a2038'], cloth: fallen ? '#2a2630' : '#e8e4d6', collar: '#caa24a',
    aura: C(50, 36, 46, glow, 'opacity=".14"'), back: wings(fallen ? '#26222c' : '#f6f2e4') + (scourge ? both(S('M30 20 L24 8 M36 16 L32 4', glow, 1.6)) : ''),
    face: brows(dark(skin, 0.5), 2) + glowEyes(glow) + nose(skin) + smile(), back2: '', front: hairShort(fallen ? '#e8e4ec' : '#d8b25a') + halo(glow), label: 'Аасимар' })
}
RACE['dnd5e:firbolg'] = (k) => bust(k, { skin: '#86a4b4', hw: 19.5, hh: 22, bg: ['#3a5a48', '#0e1c14'], cloth: '#6a5a3a', ears: earsPointy('#86a4b4', 14, 40), back: hairBack('#a9744a', 78),
  face: brows('#3a2a1c', 2.5) + eyes('#3a5a3a') + P('M46 44 q4 -1 8 0 q2 8 -4 10 q-6 -2 -4 -10Z', '#6f8e9e') + smile(), front: hairShort('#a9744a') + stubble(), label: 'Фирболг' })
const GEN = { air: { s: '#9ec3d8', h: '#f3f7fa', bg: ['#4a6a88', '#141f2c'], g: '#e8f6ff' }, earth: { s: '#8b6b4a', h: '#4a4036', bg: ['#5a4a32', '#1c140c'], g: '#d8c08a' }, fire: { s: '#c4503a', h: '#f2a23a', bg: ['#7a2a1a', '#240a06'], g: '#ffd25a' }, water: { s: '#4a8fc4', h: '#1f5f9a', bg: ['#1f4a7a', '#08182a'], g: '#bfe8ff' } }
RACE['dnd5e:genasi'] = (k, sub) => {
  const g = GEN[sub] ?? GEN.fire
  return bust(k, { skin: g.s, bg: g.bg, cloth: '#3f4a5a', aura: C(50, 40, 44, g.g, 'opacity=".12"') + (sub === 'earth' ? S('M30 30 l8 6 l-3 7 M66 36 l-6 7', '#2a2018', 1.2) : ''),
    back: hairBack(g.h, 78), face: brows(dark(g.s, 0.5), 2) + glowEyes(g.g) + nose(g.s) + smile(),
    front: hairShort(g.h) + (sub === 'fire' ? P('M38 22 L42 6 L46 20 L50 2 L54 20 L58 6 L62 22Z', '#f6b13a', 'opacity=".9"') : ''), label: 'Дженази' })
}
RACE['dnd5e:goliath'] = (k) => bust(k, { skin: '#b6b8b2', hw: 19.5, hh: 22, bg: ['#5a6670', '#161a1e'], cloth: '#6a5a4a', collar: '#2a2a2a',
  face: brows('#4a4c4e', 3.4) + eyes('#3a4a5a') + nose('#b6b8b2') + flat() + S('M36 30 l5 -6 l5 6 M54 30 l5 -6 l5 6 M44 24 h12', '#5a5e62', 1.6) + both(S('M33 50 l6 2 M34 55 l6 1', '#5a5e62', 1.4)), label: 'Голиаф' })
RACE['dnd5e:tortle'] = (k) => bust(k, { skin: '#7a9a4a', hw: 16.5, hh: 18, bg: ['#3a5a4a', '#0c1c14'], noBody: true,
  back: E(50, 62, 40, 36, '#6a5a30') + [[36, 52], [50, 48], [64, 52], [40, 66], [60, 66], [50, 78]].map(([x, y]) => P(`M${x - 8} ${y} l8 -6 l8 6 l-8 6Z`, '#8a7640', 'stroke="#4a3c1c" stroke-width="1"')).join(''),
  head: E(50, 52, 16.5, 18, `url(#${k}h)`) + E(50, 70, 18, 7, '#5f7f3c'),
  face: both(E(41.5, 47, 3.6, 3.2, '#f6e49a') + C(42, 47, 1.7, '#14100a')) + P('M42 58 Q50 54 58 58 L55 63 Q50 65 45 63Z', '#c9a85a') + S('M36 40 q4 -3 8 0 M56 40 q4 -3 8 0 M40 36 h20', '#5a7a30', 1.3), label: 'Тортл' })
RACE['dnd5e:triton'] = (k) => bust(k, { skin: '#5aa6b0', bg: ['#1f5a7a', '#08182a'], cloth: '#2f6f7a', collar: '#e8d8a0',
  ears: both(P('M33 40 L14 28 L19 42 L14 54 L33 50Z', '#2f7f94') + S('M30 40 L18 32 M30 46 L18 52', '#b8e8ee', 0.9)), back: hairBack('#1f5f7a', 80),
  face: brows('#1f4a58', 2) + eyes('#14343e') + nose('#5aa6b0') + smile() + both(S('M34 50 l4 1 M34 54 l4 0', '#2f7f94', 1.2)), front: hairShort('#1f5f7a') + P('M42 24 L50 14 L58 24Z', '#e8d8a0'), label: 'Тритон' })
RACE['dnd5e:kenku'] = (k) => bust(k, { skin: '#2a2a32', hw: 16, hh: 19, bg: ['#3a3a4c', '#0c0c14'], cloth: '#4a3a30', collar: '#8a7a5a',
  head: E(50, 43, 16, 19, `url(#${k}h)`) + P('M32 30 L26 20 L38 26Z M68 30 L74 20 L62 26Z M44 26 L50 12 L56 26Z', '#2a2a32'),
  face: both(E(40.5, 41, 4, 4, '#f2c82a') + C(41, 41, 1.8, '#0c0c10')) + P('M43 46 L57 46 L50 68Z', '#c9b890') + S('M50 47 L50 62', '#8a7a5a', 1.2) + both(S('M34 52 l5 2 M34 57 l5 0', '#4a4a56', 1.3)), label: 'Кенку' })
RACE['dnd5e:warforged'] = (k) => bust(k, { skin: '#8d949e', hw: 17, hh: 20, bg: ['#3a4452', '#0c1018'], cloth: '#5a626c', collar: '#caa24a', ears: both(P('M33 38 h-4 v10 h4Z', '#6a727c')),
  face: both(P('M36 41 h10 v4 h-10Z', '#7fd4ff') + P('M38 42 h6 v2 h-6Z', '#e6f8ff')) + S('M50 28 L50 62 M34 54 H66 M40 60 h20', '#555c66', 1.4) + both(C(36, 34, 1, '#565d68') + C(36, 62, 1, '#565d68')) + S('M44 58 h12', '#3a4048', 2.4),
  front: P('M34 30 C36 16 64 16 66 30 L62 30 C60 22 40 22 38 30Z', '#6a727c'), label: 'Воплощённый' })
RACE['dnd5e:yuanti'] = (k) => bust(k, { skin: '#6a9a5a', hw: 16.5, hh: 20, bg: ['#2f5a2a', '#0c1c08'], cloth: '#7a5a2a', collar: '#e8d070',
  back: both(P('M34 38 C16 32 8 14 12 0 C22 12 34 20 40 32Z', '#4a7a46') + S('M30 30 C22 22 18 12 18 6', '#c9d870', 1)),
  face: scales(dark('#6a9a5a', 0.3)) + slitEyes('#f0d030') + both(E(46.5, 52, 1, 1.5, '#1a2a10')) + S('M44 58 Q50 61 56 58 M50 59 L50 67 M50 67 l-2 3 M50 67 l2 3', '#b8342c', 1.3), label: 'Юань-ти' })
RACE['dnd5e:halfelf'] && 0

// ---------- существа по типу ----------
const WORDS = [['red', '#b4382a'], ['white', '#d7e6ee'], ['black', '#35353d'], ['green', '#4f8a3a'], ['blue', '#3a6ab4'], ['gold', '#d8a530'], ['silver', '#aab4c0'], ['brass', '#b89040'], ['bronze', '#a8703a'], ['copper', '#b8663a'], ['turtle', '#4a7a6a']]
const wordColor = (id, def) => (WORDS.find(([w]) => id.includes(w)) ?? [0, def])[1]
const has = (id, ...w) => w.some((x) => id.includes(x))
const bgOf = (c) => [mix(c, '#000000', 0.5), mix(c, '#000000', 0.86)]

const glowFace = (c, y = 44) => both(S('M34 38 L46 41', '#10080a', 2.2)) + glowEyes(c, y)
const fangs = (y = 58) => P(`M41 ${y} Q50 ${y + 11} 59 ${y} Q50 ${y + 3} 41 ${y}Z`, '#2a0c0c') + P(`M44 ${y + 0.4} l1.3 4.4 l1.4 -4 M54.6 ${y + 0.4} l1.4 4 l1.3 -4.4`, '#f4efe2')

const KIND = {
  beast(k, id, name) {
    if (has(id, 'snake', 'naga')) return snakeArt(k, name, has(id, 'poison') ? '#5a8a3a' : '#8a7a3a')
    if (has(id, 'hawk', 'raven', 'pteranodon', 'roc', 'eagle')) return birdArt(k, name, pick(['#4a3a2a', '#26262c', '#7a5a3a', '#5a6a7a'], id))
    if (has(id, 'crocodile', 'allosaurus', 'tyrannosaurus', 'triceratops', 'ankylosaurus', 'lizard')) return saurArt(k, name, id)
    if (has(id, 'wasp', 'crab', 'scorpion', 'stirge', 'spider', 'rat') && !has(id, 'rat')) return bugArt(k, name, id)
    if (has(id, 'shark')) return beastish(k, { c: '#7a8a96', m: '#d8e0e6', bg: ['#1f4a6a', '#08141f'], label: name, hw: 22, hh: 18, mw: 14, mh: 8, ears: '', face: both(C(36, 43, 2.2, '#10161c')) + P('M32 56 Q50 70 68 56 Q50 60 32 56Z', '#2a0c10') + Array.from({ length: 7 }, (_, i) => P(`M${36 + i * 4.6} 58 l1.6 4.4 l1.6 -4.4`, '#fff')).join('') + S('M50 14 L44 30 L56 30Z', '#7a8a96', 1.5) })
    const [c, m] = pick([['#8b7355', '#d8c8a0'], ['#6f6258', '#a89888'], ['#a8743a', '#f0e0c0'], ['#4a4d55', '#8d9199'], ['#c9a04a', '#f4e8c8'], ['#5a4a3a', '#b8a888']], id)
    const big = has(id, 'elephant', 'mammoth', 'rhinoceros', 'boar', 'bear')
    return beastish(k, { c, m, bg: bgOf(c), label: name, hw: big ? 22 : 20, hh: 19, mw: 12, mh: 9,
      ears: hash(id + 'e') % 2 ? earsTri(c, dark(c, 0.4)) : earsRound(c, dark(c, 0.4), 29, 31, 7.5),
      face: furEyes() + E(50, 52, 4, 3, '#14100e') + S('M50 55 L50 60 M43 61 Q50 65 57 61', '#2a1a14', 1.4) + (has(id, 'boar', 'rhinoceros', 'elephant', 'mammoth') ? tusks('#f1ead6', 9) : '') + (has(id, 'tiger', 'panther', 'lion', 'saber', 'cat') ? both(S('M33 44 l5 1 M33 49 l5 -1', dark(c, 0.5), 1.4)) : '') + (has(id, 'saber') ? both(P('M43 60 l-1.4 10 l3.4 -9Z', '#fff')) : '') })
  },
  humanoid(k, id, name) {
    if (has(id, 'wererat', 'werewolf', 'wereboar', 'weretiger')) return KIND.beast(k, has(id, 'tiger') ? 'tiger' : has(id, 'boar') ? 'giant-boar' : has(id, 'rat') ? 'wolf' : 'wolf', name)
    const hood = (c) => P('M24 100 C24 50 30 14 50 12 C70 14 76 50 76 100Z', c)
    const skin = pick(['#e4bd96', '#d9a982', '#c48a62', '#f0d0aa', '#a8714a'], id)
    if (has(id, 'assassin', 'spy', 'scout', 'thug')) return bust(k, { skin, bg: ['#3a3a44', '#101014'], cloth: '#2f3038', back: hood(has(id, 'scout') ? '#3a5a34' : '#26262c'), face: brows('#1a1410', 2) + eyes('#22160c') + nose(skin) + flat(), front: P('M30 40 C30 24 40 20 50 20 C60 20 70 24 70 40 C64 33 58 31 50 31 C42 31 36 33 30 40Z', has(id, 'scout') ? '#3a5a34' : '#26262c') + P('M33 50 C36 64 64 64 67 50 C62 54 56 54 50 54 C44 54 38 54 33 50Z', '#1d1d22'), label: name })
    if (has(id, 'mage', 'priest', 'acolyte', 'druid', 'cult', 'noble')) {
      const rob = has(id, 'mage') ? '#3a4a8a' : has(id, 'priest', 'acolyte') ? '#e6dcc0' : has(id, 'druid') ? '#4a6a34' : has(id, 'noble') ? '#7a2f5a' : '#4a1a24'
      return bust(k, { skin, bg: bgOf(rob), cloth: rob, collar: '#caa24a', face: brows('#4a3a2a', 2) + eyes() + nose(skin) + smile(),
        front: has(id, 'mage') ? P('M28 36 L50 -6 L72 36 C62 30 38 30 28 36Z', rob) + P('M28 36 h44 v4 h-44Z', '#caa24a') + beard('#d8d2c0', 68) : has(id, 'druid') ? hairBack('#8a6a3a', 76) + hairShort('#8a6a3a') + P('M36 24 l-6 -6 l10 2Z M64 24 l6 -6 l-10 2Z', '#6aa43c') : has(id, 'priest', 'acolyte') ? hairShort('#8a8a86') + S('M50 80 v14 M44 85 h12', '#caa24a', 2) : hairShort('#3a2a1c') + stubble(), label: name })
    }
    if (has(id, 'knight', 'veteran', 'guard', 'gladiator', 'captain', 'warrior')) return bust(k, { skin, bg: ['#47505c', '#12161c'], cloth: '#6a7480', collar: '#caa24a', face: brows('#4a4a48', 3) + eyes() + nose(skin) + flat() + (has(id, 'gladiator', 'captain') ? scar() : ''),
      front: has(id, 'warrior') ? mohawk('#3a2a1c') + S('M34 44 l8 3 M66 44 l-8 3', '#b8342c', 2) : helm(has(id, 'captain') ? '#6a4a2a' : '#8a9098', true) + stubble(), label: name })
    return bust(k, { skin, bg: ['#4a4a50', '#14141a'], cloth: pick(['#6a4a3a', '#3f5a6a', '#5a5a3a'], id), face: brows('#3a2a1c', 2.4) + eyes() + nose(skin) + smile(), front: hairShort(pick(['#3a2a1c', '#6a4a2a', '#2a2a2a'], id)) + stubble(), label: name })
  },
  undead(k, id, name) {
    if (has(id, 'ghost', 'banshee', 'wraith', 'specter', 'shadow', 'wisp')) {
      const c = has(id, 'shadow') ? '#26262e' : has(id, 'wisp') ? '#bfe8ff' : '#aee0e6'
      if (has(id, 'wisp')) return bust(k, { skin: c, noBody: true, bg: ['#1a2a3a', '#04080c'], aura: C(50, 44, 36, c, 'opacity=".25"'), head: C(50, 44, 15, c, 'opacity=".9"') + C(50, 44, 8, '#fff', 'opacity=".9"'), label: name })
      return bust(k, { skin: c, noBody: true, bg: ['#1a2a30', '#04080a'], aura: C(50, 44, 40, c, 'opacity=".10"'),
        back: P('M16 100 C14 56 28 16 50 14 C72 16 86 56 84 100 L72 90 L62 100 L50 88 L38 100 L28 90Z', c, 'opacity=".38"'),
        head: E(50, 44, 15, 19, c, 'opacity=".55"'), face: both(E(43, 42, 3.6, 4.6, '#06141a')) + both(C(43, 42, 1.2, '#c8ffff')) + E(50, 58, 5, 7, '#06141a', 'opacity=".85"'), label: name })
    }
    if (has(id, 'mummy')) return bust(k, { skin: '#cbbd92', bg: ['#4a4030', '#14100a'], cloth: '#bfae82', collar: '#8a7a54', face: S('M33 34 L67 40 M32 46 L68 38 M34 54 L66 48 M38 62 L62 58', '#8a7a54', 2.2) + both(E(42, 44, 3.4, 2.2, '#0c0806') + C(42, 44, 1.2, has(id, 'lord') ? '#ffd24a' : '#e8412a')), front: has(id, 'lord') ? P('M34 22 L40 8 L46 20 L50 4 L54 20 L60 8 L66 22Z', '#d8b04a') : '', label: name })
    if (has(id, 'vampire', 'ghast', 'ghoul')) return bust(k, { skin: '#d6d2cc', bg: ['#4a1a22', '#10060a'], cloth: '#2a1a24', collar: '#8a1f2a', ears: earsPointy('#d6d2cc', 9), face: brows('#1a1214', 2.4) + glowEyes('#e0302a') + nose('#d6d2cc') + P('M44 57 Q50 60 56 57 Q50 59 44 57Z', '#5a1a1a') + P('M45.5 57.6 l1 3.6 l1 -3 M54.5 57.6 l-1 3.6 l-1 -3', '#fff'), front: P('M30 40 C28 18 42 14 50 14 C58 14 72 18 70 40 C64 28 58 24 50 26 C42 24 36 28 30 40Z', '#14101a'), label: name })
    if (has(id, 'lich', 'skull', 'death-knight', 'flameskull')) {
      const kn = has(id, 'knight')
      return bust(k, { skin: kn ? '#5a6270' : '#e6e0cc', shade: '#8a846c', hw: 16, hh: 19.5, bg: kn ? ['#1a2a4a', '#04060e'] : ['#3a2a5a', '#0c0818'], cloth: kn ? '#3a4250' : '#2a1a3a', collar: '#8a5ad8',
        head: E(50, 43, 17, 19.5, `url(#${k}h)`), face: both(E(42, 43, 4.2, 5, '#0c0a14')) + both(C(42, 43.5, 1.5, kn ? '#6ab4ff' : has(id, 'flame') ? '#ff9a2a' : '#8affb0')) + P('M50 49 L47 55 L53 55Z', '#0c0a14') + S('M40 62 H60 M44 58 V66 M50 58 V66 M56 58 V66', '#6a6450', 1.1),
        front: kn ? helm('#5a6270', false) : (has(id, 'lich') ? P('M34 24 L38 10 L44 20 L50 6 L56 20 L62 10 L66 24Z', '#d8b04a') : '') + (has(id, 'flame') ? P('M34 22 C36 8 44 12 46 2 C50 10 56 8 58 0 C62 10 66 14 66 24Z', '#f6942a', 'opacity=".85"') : ''), label: name })
    }
    return KIND.undead(k, 'skull-' + id, name)
  },
  dragon(k, id, name) {
    const c = wordColor(id, '#4a8a6a')
    const o = dragonHead(c, mix(c, '#ffffff', 0.5)); o.head = o.head.replace('DR', k)
    const big = has(id, 'adult', 'ancient')
    return bust(k, { ...o, bg: bgOf(c), cloth: '#4a4f5a', aura: has(id, 'ancient') ? C(50, 40, 46, mix(c, '#ffffff', 0.4), 'opacity=".14"') : '', front: big ? both(P('M40 14 L36 2 L46 12Z', mix(c, '#ffffff', 0.4))) : '', label: name })
  },
  giant(k, id, name) {
    const skin = has(id, 'frost') ? '#b8d6e6' : has(id, 'fire') ? '#a8442a' : has(id, 'stone') ? '#9a9a96' : has(id, 'cloud') ? '#a8c6e0' : has(id, 'storm') ? '#7a86d4' : has(id, 'oni') ? '#4a6ab4' : '#b8946a'
    const hair = has(id, 'frost', 'cloud') ? '#f1f4f6' : has(id, 'fire') ? '#14100e' : '#4a3a28'
    return bust(k, { skin, hw: 21, hh: 23, bg: bgOf(skin), cloth: '#6a5a44', collar: '#2a2014', ears: earsRound(skin, dark(skin, 0.2), 44, 29, 6),
      face: both(S('M33 37 L46 40.5', '#201408', 3.2)) + (has(id, 'fire', 'storm') ? glowEyes(has(id, 'fire') ? '#ffb83a' : '#ffffff', 44) : both(E(41.5, 44, 2.6, 1.9, '#f0e4c8') + C(41.5, 44.4, 1.2, '#2a1a0c'))) + P('M45 46 h10 l1 7 h-12Z', dark(skin, 0.12)) + flat(),
      front: P('M31 40 C28 16 44 14 50 14 C56 14 72 16 69 40 C64 28 36 28 31 40Z', hair) + (has(id, 'frost', 'cloud', 'storm', 'fire', 'stone') ? P('M30 46 C28 66 40 80 50 84 C60 80 72 66 70 46 C64 58 58 60 50 60 C42 60 36 58 30 46Z', hair) : '') + (has(id, 'oni', 'ettin') ? horns('#e8dcc0') : ''), label: name })
  },
  fiend(k, id, name) {
    const c = has(id, 'imp', 'devil', 'erinyes', 'hound') ? '#a82a2a' : has(id, 'succubus') ? '#c46a8a' : pick(['#6a2a5a', '#3a2a4a', '#8a3a2a'], id)
    if (has(id, 'hell-hound', 'nightmare')) return beastish(k, { c: '#2a2224', m: '#4a3a3a', bg: ['#5a1a10', '#140404'], label: name, hw: 21, hh: 20, mw: 12, mh: 10, ears: earsTri('#2a2224', '#8a2a1a'), aura: '', face: furEyes('#ff6a1a') + E(50, 52, 4, 3, '#0c0808') + P('M42 60 Q50 70 58 60 Q50 63 42 60Z', '#ff7a2a') + P('M44 60.4 l1.4 4 l1.4 -3.6 M56 60.4 l-1.4 4 l-1.4 -3.6', '#fff') + S('M36 24 q-3 -8 2 -14 M64 24 q3 -8 -2 -14', '#ff9a2a', 2) })
    return bust(k, { skin: c, bg: ['#5a1a1a', '#140404'], cloth: '#2a1a1e', collar: '#8a2a2a', ears: earsPointy(c, 12), back: has(id, 'succubus', 'vrock', 'erinyes', 'balor', 'nalfeshnee', 'imp', 'horned') ? wings('#3a1a22') : '', face: brows('#14060a', 2.6) + glowEyes('#ffd23a') + nose(c) + fangs(57).slice(0, 0) + grin(true), front: (has(id, 'imp') ? '' : curlHorns('#d8c8a0')) + (has(id, 'pit', 'balor') ? aura0() : ''), label: name })
  },
  fey(k, id, name) {
    const skin = has(id, 'dryad', 'green-hag') ? '#7a9a5a' : '#f1d6b4'
    return bust(k, { skin, bg: ['#3a6a44', '#0e1c12'], cloth: '#4a7a4a', ears: earsPointy(skin, 16), back: has(id, 'sprite', 'pseudo') ? wings('#d8f0ff', 1.2) : '', aura: [[20, 20], [80, 30], [70, 80], [26, 70]].map(([x, y]) => C(x, y, 1.6, '#fff6b0', 'opacity=".8"')).join(''),
      face: brows(dark(skin, 0.5), 2) + eyes(has(id, 'hag') ? '#d8a02a' : '#2a6a3a') + nose(skin) + (has(id, 'hag') ? flat() + C(60, 52, 1.5, '#5a7a3a') : smile()), front: has(id, 'hag') ? hairBack('#2a2a30', 90) + hairShort('#2a2a30') : hairShort('#5a9a3a') + P('M36 24 l-4 -8 l9 4Z M64 24 l4 -8 l-9 4Z', '#8ac43c'), label: name })
  },
  construct(k, id, name) {
    const c = has(id, 'flesh') ? '#a8967a' : has(id, 'clay') ? '#b0764a' : has(id, 'stone') ? '#8a8a8a' : '#6a727c'
    return bust(k, { skin: c, hw: 18, hh: 21, bg: bgOf(c), cloth: dark(c, 0.3), collar: '#caa24a', ears: has(id, 'flesh') ? both(C(31, 44, 2.4, '#8a7a60') + C(31, 50, 2.4, '#8a7a60')) : '',
      face: has(id, 'armor', 'guardian') ? P('M34 40 L66 40 L66 46 L34 46Z', '#0c0e14') + both(P('M38 41 h6 v3 h-6Z', '#7fd4ff')) : glowFace(has(id, 'flesh') ? '#e6e8a0' : '#ff9a3a') + S('M50 28 L50 62 M34 54 H66', dark(c, 0.35), 1.4) + (has(id, 'flesh') ? S('M36 34 l12 4 M40 36 v4 M44 37 v4 M60 54 l6 -2 M62 52 v4', '#3a2a28', 1.2) : ''),
      front: has(id, 'armor', 'guardian') ? helm(c, false) : '', label: name })
  },
  ooze(k, id, name) {
    const c = has(id, 'black') ? '#2a2a34' : has(id, 'gelatinous') ? '#8ad0d8' : '#7ab04a'
    return bust(k, { skin: c, noBody: true, bg: bgOf(c), head: P('M16 90 C12 50 28 22 50 22 C72 22 88 50 84 90 C74 96 26 96 16 90Z', c, 'opacity=".85"') + P('M28 40 C32 30 40 28 48 30', 'rgba(255,255,255,.4)', 'stroke="none"') + [[30, 76, 3], [64, 70, 4], [52, 82, 2.6], [72, 82, 2]].map(([x, y, r]) => C(x, y, r, 'rgba(255,255,255,.25)')).join(''),
      face: has(id, 'gelatinous') ? `<rect x="22" y="30" width="56" height="56" rx="4" fill="${c}" opacity=".35" stroke="#e6fbff" stroke-width="1.2"/>` + P('M40 52 h8 v14 h-8Z', '#d8c8a0') + C(62, 64, 4, '#c8a8a8') : both(C(42, 52, 4.6, '#fff')) + both(C(42.6, 52.6, 2, '#10140a')) + S('M42 66 Q50 70 58 66', dark(c, 0.5), 1.6), label: name })
  },
  elemental(k, id, name) {
    if (has(id, 'gargoyle')) return bust(k, { skin: '#8d8f94', hw: 17, hh: 20, bg: ['#3a3f4a', '#0c0e14'], cloth: '#6a6c72', back: wings('#6a6c72'), ears: '', face: glowFace('#f2c82a') + nose('#8d8f94') + fangs(57).slice(0, 0) + grin(true), front: horns('#5a5c62'), label: name })
    const e = has(id, 'fire', 'efreeti', 'salamander') ? ['#d8502e', '#ffc84a', ['#7a2a1a', '#240a06']] : has(id, 'water') ? ['#3a8ac4', '#cdeeff', ['#1f4a7a', '#08182a']] : has(id, 'earth', 'xorn') ? ['#8a6a44', '#d8c08a', ['#5a4a32', '#1c140c']] : ['#9ec3d8', '#f3f7fa', ['#4a6a88', '#141f2c']]
    return bust(k, { skin: e[0], hw: 17, hh: 20, bg: e[2], noBody: true, aura: C(50, 44, 42, e[1], 'opacity=".14"'),
      back: [0, 1, 2, 3, 4].map((i) => P(`M${16 + i * 17} 100 C${6 + i * 17} 70 ${24 + i * 17} 56 ${16 + i * 17} 30 C${34 + i * 17} 56 ${30 + i * 17} 80 ${28 + i * 17} 100Z`, e[0], 'opacity=".55"')).join(''),
      head: E(50, 46, 17, 20, e[0]) + E(50, 46, 11, 14, e[1], 'opacity=".5"'), face: glowEyes(e[1], 44) + S('M42 58 Q50 62 58 58', dark(e[0], 0.5), 1.6), label: name })
  },
  monstrosity(k, id, name) {
    if (has(id, 'mimic')) return bust(k, { skin: '#8a5a2a', noBody: true, bg: ['#4a3a2a', '#14100a'], head: `<rect x="14" y="32" width="72" height="34" rx="6" fill="#8a5a2a" stroke="#4a2c0c" stroke-width="2"/>` + `<rect x="14" y="24" width="72" height="14" rx="6" fill="#a8703a" stroke="#4a2c0c" stroke-width="2"/>` + P('M18 56 Q50 74 82 56 L82 66 Q50 82 18 66Z', '#7a1a1a') + Array.from({ length: 8 }, (_, i) => P(`M${22 + i * 8} 58 l3 6 l3 -6`, '#fff')).join('') + C(50, 46, 4, '#e8c050'), face: both(E(36, 43, 3, 3, '#ffd23a')) + both(C(36, 43, 1, '#14100a')), label: name })
    if (has(id, 'worm', 'bulette', 'remorhaz', 'ankheg')) return bust(k, { skin: '#6a5a8a', noBody: true, bg: bgOf('#6a5a8a'), head: E(50, 50, 38, 38, '#6a5a8a') + E(50, 50, 28, 28, '#2a0c1c') + E(50, 50, 18, 18, '#14060e'), face: Array.from({ length: 16 }, (_, i) => { const a = (i / 16) * Math.PI * 2; return P(`M${50 + Math.cos(a) * 28} ${50 + Math.sin(a) * 28} L${50 + Math.cos(a) * 17} ${50 + Math.sin(a) * 17} L${50 + Math.cos(a + 0.2) * 28} ${50 + Math.sin(a + 0.2) * 28}Z`, '#f1ead6') }).join(''), label: name })
    if (has(id, 'medusa', 'lamia', 'hydra')) return bust(k, { skin: '#7a9a6a', bg: ['#2f4a2a', '#0a140a'], cloth: '#4a3a5a', face: brows('#1a2a14', 2) + slitEyes('#e8d030') + nose('#7a9a6a') + smile(), front: [-30, -15, 0, 15, 30].map((a) => S(`M${50 + a * 0.5} 30 Q${50 + a * 1.2} 10 ${50 + a * 1.6} ${14 + Math.abs(a) * 0.3}`, '#3a6a2a', 4)).join('') + [-30, 30].map((a) => C(50 + a * 1.6, 14 + Math.abs(a) * 0.3, 2.4, '#3a6a2a')).join(''), label: name })
    const c = pick(['#6a4a34', '#4a6a4a', '#7a4a6a', '#8a7a34', '#4a5a7a'], id)
    return bust(k, { skin: c, hw: 19, hh: 21, bg: bgOf(c), cloth: dark(c, 0.3), back: has(id, 'wyvern', 'chimera', 'manticore', 'harpy', 'roc') ? wings(dark(c, 0.1)) : '', ears: earsTri(c, dark(c, 0.4)),
      face: brows('#10080a', 3) + slitEyes(pick(['#f2c82a', '#e8402a', '#6aff9a'], id)) + nose(c) + grin(true), front: horns(pick(['#e8dcc0', '#3a2a24'], id)) + (hash(id) % 2 ? scales(dark(c, 0.3)) : ''), label: name })
  },
  aberration(k, id, name) {
    if (has(id, 'beholder')) return bust(k, { skin: '#8a5a7a', noBody: true, bg: ['#4a1a4a', '#0c040c'], back: [-34, -20, -8, 8, 20, 34].map((a) => S(`M50 40 Q${50 + a * 1.2} ${22 - Math.abs(a) * 0.2} ${50 + a * 1.5} ${12 + Math.abs(a) * 0.2}`, '#6a3a5a', 3) + C(50 + a * 1.5, 12 + Math.abs(a) * 0.2, 3, '#f2d03a') + C(50 + a * 1.5, 12 + Math.abs(a) * 0.2, 1.2, '#10060c')).join(''),
      head: C(50, 52, 34, 'url(#' + k + 'h)'), face: C(50, 48, 14, '#f4efe2') + C(50, 48, 9, '#3acb6a') + E(50, 48, 3, 8, '#06100a') + P('M20 66 Q50 90 80 66 Q50 74 20 66Z', '#2a0c1a') + Array.from({ length: 9 }, (_, i) => P(`M${26 + i * 6} 70 l2.4 5 l2.4 -5`, '#fff')).join(''), label: name })
    return bust(k, { skin: '#5a7a8a', noBody: true, bg: ['#1f3a4a', '#04101a'], back: [14, 30, 70, 86].map((x) => S(`M${x} 100 Q${x + (x < 50 ? 10 : -10)} 70 ${x < 50 ? x + 18 : x - 18} 50`, '#4a6a7a', 5)).join(''),
      head: E(50, 46, 24, 26, 'url(#' + k + 'h)'), face: [[50, 32, 6], [38, 44, 4.5], [62, 44, 4.5]].map(([x, y, r]) => C(x, y, r, '#f2d03a') + E(x, y, r * 0.3, r * 0.9, '#10060c')).join('') + S('M34 62 Q50 72 66 62', '#10202a', 2), label: name })
  },
  plant(k, id, name) { return bust(k, { skin: '#4a8a3a', noBody: true, bg: ['#1f4a24', '#06120a'], head: Array.from({ length: 14 }, (_, i) => P(`M50 ${54 + (i % 3) * 3} Q${10 + i * 6} ${10 + (i % 4) * 8} ${6 + i * 6.4} ${40 + (i % 5) * 7} Q${24 + i * 4} 60 50 ${60}Z`, pick(['#3a7a2a', '#4a8a3a', '#2a6a24', '#5a9a3a'], id, i), 'opacity=".9"')).join('') + E(50, 56, 16, 18, '#2a4a1c'), face: both(E(43, 52, 3, 2.4, '#e8d84a')) + P('M42 64 Q50 70 58 64Z', '#0c1a08'), label: name }) },
  celestial(k, id, name) { return bust(k, { skin: '#f2dcc2', bg: ['#7a8ab4', '#1a2038'], cloth: '#f2eede', collar: '#d8b04a', aura: C(50, 36, 46, '#fff4b0', 'opacity=".2"'), back: wings('#fbf8ee', 1.1), face: brows('#8a6a2a', 2) + glowEyes('#fff0a8') + nose('#f2dcc2') + smile(), front: hairShort('#e8c850') + halo(), label: name }) },
}
function aura0() { return '' }

function birdArt(k, name, c) {
  return bust(k, { skin: c, noBody: true, bg: bgOf('#4a6a88'), back: both(P('M33 30 L12 22 L30 40Z M30 44 L8 46 L30 54Z', dark(c, 0.15))), head: E(50, 46, 20, 22, `url(#${k}h)`),
    face: both(E(39, 40, 5, 4.6, '#f6c843') + C(39.4, 40, 2.2, '#10100c')) + P('M42 48 L58 48 L50 72Z', '#f2b83a') + S('M50 49 L50 66', '#b5821c', 1.2), front: P('M32 36 C34 20 42 16 50 16 C58 16 66 20 68 36 C62 28 56 26 50 26 C44 26 38 28 32 36Z', dark(c, 0.25)), label: name })
}
function snakeArt(k, name, c) {
  return bust(k, { skin: c, noBody: true, bg: bgOf(c), back: both(P('M34 42 C16 36 10 18 14 4 C24 14 34 22 42 34Z', dark(c, 0.1))), head: E(50, 46, 17, 23, `url(#${k}h)`),
    face: scales(dark(c, 0.3)) + slitEyes('#f2d030', 40) + both(E(46, 54, 0.9, 1.4, '#10140a')) + S('M44 64 Q50 67 56 64 M50 66 L50 78 M50 78 l-3 4 M50 78 l3 4', '#b8342c', 1.4), label: name })
}
function saurArt(k, name, id) {
  const c = has(id, 'crocodile') ? '#5a7a3a' : has(id, 'triceratops') ? '#8a7a54' : has(id, 'ankylosaurus') ? '#7a6a4a' : '#7a8a4a'
  return bust(k, { skin: c, noBody: true, bg: bgOf(c), head: P('M26 42 C24 22 38 14 50 14 C62 14 76 22 74 42 C78 56 70 74 50 78 C30 74 22 56 26 42Z', `url(#${k}h)`),
    face: slitEyes('#f2c82a', 34) + both(E(45, 52, 1.2, 1.8, '#10140a')) + S('M32 66 Q50 74 68 66', '#10140a', 1.6) + Array.from({ length: 8 }, (_, i) => P(`M${36 + i * 4} 67 l1.4 4 l1.4 -4`, '#fff')).join('') + scales(dark(c, 0.3)) + (has(id, 'triceratops') ? both(P('M36 28 L26 6 L42 24Z', '#f1ead6')) + P('M47 40 L50 26 L53 40Z', '#f1ead6') : '') + (has(id, 'tyranno') ? '' : ''), label: name })
}
function bugArt(k, name, id) {
  const c = has(id, 'wasp', 'stirge') ? '#c9a22a' : has(id, 'crab', 'scorpion') ? '#b8542e' : '#2a2230'
  return bust(k, { skin: c, noBody: true, bg: bgOf(c), back: both(S('M42 30 Q30 8 18 6 M38 28 Q24 14 12 16', dark(c, 0.3), 2)), head: E(50, 48, 22, 22, `url(#${k}h)`),
    face: both(E(38, 42, 9, 11, '#14100c') + C(35, 38, 2.4, 'rgba(255,255,255,.55)')) + both(P('M44 62 Q38 74 32 72 Q38 64 42 58Z', dark(c, 0.3))) + (has(id, 'wasp') ? S('M24 52 h52 M26 60 h48', '#14100c', 3) : ''), label: name })
}

export const monsterArt = (id, kind, name = '') => {
  const k = `m-${id}`
  const f = MON[id]
  if (f) return f(k)
  return (KIND[kind] ?? KIND.monstrosity)(k, id, name || id)
}
export const MONSTER_IDS = Object.keys(MON)
export const RACE_KEYS = Object.keys(RACE)
