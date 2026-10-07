import { raceArt, monsterArt } from './art.js'

// Art рисует готовую SVG-разметку из art.js. Разметка создаётся только из наших шаблонов, пользовательский текст в неё не попадает.
export function Art({ html, className = 'art', label }) {
  return <span className={className} role="img" aria-label={label} dangerouslySetInnerHTML={{ __html: html }} />
}

// artFor выбирает картинку: загруженный портрет, иначе рисунок существа или расы.
export function artFor(c) {
  if (c.kind === 'monster') return monsterArt(c.stat?.id ?? '', c.stat?.kind, (c.name ?? '').replace(/ \d+$/, ''))
  return raceArt(c.ruleset, c.race, c.subrace)
}

export function Avatar({ hero, size = 40 }) {
  const s = { width: size, height: size }
  if (hero.portrait) return <img className="avatar" style={s} src={hero.portrait} alt="" />
  return <span className="avatar" style={s} aria-hidden="true" dangerouslySetInnerHTML={{ __html: artFor(hero) }} />
}

// pickImage читает файл и сжимает его через canvas до квадрата max×max (обрезка по центру), чтобы сохранение не раздувалось.
export function pickImage(file, max = 256, quality = 0.82) {
  return new Promise((resolve, reject) => {
    if (!file || !file.type.startsWith('image/')) return reject(new Error('Нужна картинка'))
    const url = URL.createObjectURL(file)
    const img = new Image()
    img.onload = () => {
      const side = Math.min(img.naturalWidth, img.naturalHeight)
      const cv = document.createElement('canvas')
      cv.width = cv.height = Math.min(max, side)
      cv.getContext('2d').drawImage(img, (img.naturalWidth - side) / 2, (img.naturalHeight - side) / 2, side, side, 0, 0, cv.width, cv.height)
      URL.revokeObjectURL(url)
      resolve(cv.toDataURL('image/jpeg', quality))
    }
    img.onerror = () => { URL.revokeObjectURL(url); reject(new Error('Не удалось прочитать картинку')) }
    img.src = url
  })
}

// pickMap – то же для карты: без обрезки, длинная сторона не больше max.
export function pickMap(file, max = 3000, quality = 0.88) {
  return new Promise((resolve, reject) => {
    if (!file || !file.type.startsWith('image/')) return reject(new Error('Нужна картинка'))
    const url = URL.createObjectURL(file)
    const img = new Image()
    img.onload = () => {
      const k = Math.min(1, max / Math.max(img.naturalWidth, img.naturalHeight))
      const cv = document.createElement('canvas')
      cv.width = Math.round(img.naturalWidth * k)
      cv.height = Math.round(img.naturalHeight * k)
      cv.getContext('2d').drawImage(img, 0, 0, cv.width, cv.height)
      URL.revokeObjectURL(url)
      resolve(cv.toDataURL('image/jpeg', quality))
    }
    img.onerror = () => { URL.revokeObjectURL(url); reject(new Error('Не удалось прочитать картинку')) }
    img.src = url
  })
}
