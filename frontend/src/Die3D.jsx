import { useEffect, useId, useMemo, useState } from 'react'
import { cubeTurn } from './rollText.js'

// Кубики для лотка. d6 – настоящий CSS-3D куб с точками; остальные – гранёные многогранники (SVG) с объёмной заливкой,
// которые кувыркаются в 3D, пока идёт бросок, и «приземляются» с отскоком.

const PIPS = { 1: [4], 2: [0, 8], 3: [0, 4, 8], 4: [0, 2, 6, 8], 5: [0, 2, 4, 6, 8], 6: [0, 2, 3, 5, 6, 8] }
const FACES = [['front', 1], ['back', 6], ['right', 3], ['left', 4], ['top', 2], ['bottom', 5]]

const pt = (r, deg, cx = 50, cy = 50) => [cx + r * Math.cos((deg * Math.PI) / 180), cy + r * Math.sin((deg * Math.PI) / 180)]
const poly = (pts) => pts.map((p) => p.map((v) => +v.toFixed(1)).join(',')).join(' ')
const ring = (n, r, start, cy = 50) => Array.from({ length: n }, (_, i) => pt(r, start + (360 / n) * i, 50, cy))

// Форма каждого кубика: контур, «передняя» грань и рёбра от неё к контуру. y – где писать число.
function shape(sides) {
  switch (sides) {
    case 4: {
      const o = [[50, 5], [95, 88], [5, 88]]
      const c = [50, 62]
      return { outer: o, face: null, edges: o.map((p) => [c, p]), y: 77, size: 30 }
    }
    case 8: {
      const o = [[50, 3], [95, 50], [50, 97], [5, 50]]
      const f = [[50, 14], [84, 66], [16, 66]]
      return { outer: o, face: f, edges: [[f[0], o[0]], [f[1], o[1]], [f[1], o[2]], [f[2], o[2]], [f[2], o[3]]], y: 59, size: 30 }
    }
    case 10:
    case 100: {
      const o = [[50, 3], [95, 42], [80, 80], [50, 97], [20, 80], [5, 42]]
      const f = [[50, 14], [76, 50], [50, 82], [24, 50]]
      return { outer: o, face: f, edges: [[f[0], o[0]], [f[1], o[1]], [f[1], o[2]], [f[2], o[3]], [f[3], o[4]], [f[3], o[5]]], y: 60, size: sides === 100 ? 24 : 32 }
    }
    case 12: {
      const o = ring(10, 47, -90)
      const f = ring(5, 25, -90, 52)
      return { outer: o, face: f, edges: f.map((p, i) => [p, o[i * 2]]), y: 63, size: 31 }
    }
    case 20: {
      const o = [[50, 3], [93, 27], [93, 73], [50, 97], [7, 73], [7, 27]]
      const f = [[50, 21], [79, 68], [21, 68]]
      const e = [[f[0], o[0]], [f[0], o[1]], [f[0], o[5]], [f[1], o[1]], [f[1], o[2]], [f[1], o[3]], [f[2], o[3]], [f[2], o[4]], [f[2], o[5]]]
      return { outer: o, face: f, edges: e, y: 62, size: 31 }
    }
    default:
      return { outer: null, face: null, edges: [], y: 62, size: 32 }
  }
}

function Cube({ value, rolling }) {
  // стартовый наклон случаен: из него кубик «докатывается» до нужной грани (переход по transform в CSS)
  const [tilt] = useState(() => [Math.random() * 720 - 360, Math.random() * 720 - 360])
  const [settled, setSettled] = useState(false)
  useEffect(() => {
    if (rolling || value == null) { setSettled(false); return }
    // два кадра: сначала браузер рисует наклон, потом получает итоговый поворот и анимирует переход
    let r2 = 0
    const r1 = requestAnimationFrame(() => { r2 = requestAnimationFrame(() => setSettled(true)) })
    return () => { cancelAnimationFrame(r1); cancelAnimationFrame(r2) }
  }, [rolling, value])
  const [x, y] = cubeTurn(value)
  const style = settled ? { transform: `rotateX(${x + 720}deg) rotateY(${y + 360}deg)` } : { transform: `rotateX(${tilt[0]}deg) rotateY(${tilt[1]}deg)` }
  return (
    <div className={'cube' + (rolling ? ' spinning' : '')} style={style}>
      {FACES.map(([side, n]) => (
        <div key={side} className={'cf6 ' + side}>
          {Array.from({ length: 9 }, (_, i) => <i key={i} className={PIPS[n].includes(i) ? 'pip6' : ''} />)}
        </div>
      ))}
    </div>
  )
}

function Poly({ sides, shown }) {
  const id = useId()
  const s = useMemo(() => shape(sides), [sides])
  const text = sides === 100 && typeof shown === 'number' ? String(shown).padStart(2, '0') : shown
  return (
    <svg viewBox="0 0 100 100" className="poly">
      <defs>
        <radialGradient id={id + 'g'} cx="35%" cy="28%" r="80%">
          <stop offset="0" className="g0" /><stop offset=".55" className="g1" /><stop offset="1" className="g2" />
        </radialGradient>
      </defs>
      {s.outer ? <polygon className="body" points={poly(s.outer)} fill={`url(#${id}g)`} /> : <circle className="body" cx="50" cy="50" r="46" fill={`url(#${id}g)`} />}
      {s.face && <polygon className="face" points={poly(s.face)} />}
      {s.edges.map(([a, b], i) => <line key={i} className="edge" x1={a[0]} y1={a[1]} x2={b[0]} y2={b[1]} />)}
      <text x="50" y={s.y} textAnchor="middle" className="num" style={{ fontSize: s.size }}>{text}</text>
    </svg>
  )
}

/**
 * Die3D – один кубик. Пока rolling, цифры мелькают и кубик кувыркается; потом встаёт на выпавшее значение.
 * @param {{sides:number, value?:number, rolling?:boolean, dim?:boolean, tone?:string, index?:number}} props
 */
export default function Die3D({ sides = 20, value, rolling = false, dim = false, tone = '', index = 0 }) {
  const [flick, setFlick] = useState(value ?? '?')
  useEffect(() => {
    if (!rolling) { setFlick(value ?? '?'); return }
    const t = setInterval(() => setFlick(1 + Math.floor(Math.random() * (sides || 20))), 75)
    return () => clearInterval(t)
  }, [rolling, value, sides])
  const cls = ['die3d', 'd' + sides, rolling ? 'rolling' : 'landed', dim && 'dim', tone].filter(Boolean).join(' ')
  return (
    <div className={cls} style={{ '--i': index }} role="img" aria-label={rolling ? 'кубик катится' : `d${sides}: ${value}`}>
      <div className="dthrow">
        <div className="dspin">{sides === 6 ? <Cube value={value} rolling={rolling} /> : <Poly sides={sides} shown={flick} />}</div>
      </div>
      <span className="dshadow" />
      {!rolling && sides !== 6 && sides !== 20 && <small className="dlabel">d{sides}</small>}
    </div>
  )
}
