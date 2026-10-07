import { useEffect, useState } from 'react'
import { api } from './api'

const MODES = [['', 'Обычный'], ['adv', 'Преимущество'], ['dis', 'Помеха']]
const SHAPES = {
  4: '50,5 95,90 5,90', 6: '14,14 86,14 86,86 14,86', 8: '50,3 93,50 50,97 7,50', 10: '50,3 95,40 79,95 21,95 5,40',
  12: '50,3 91,27 91,73 50,97 9,73 9,27', 20: '50,3 93,27 93,73 50,97 7,73 7,27',
}

// Die — один кубик. Пока rolling, цифры мелькают; потом показывается настоящее значение.
function Die({ sides, value, rolling, dim, tone }) {
  const [shown, setShown] = useState(value ?? '?')
  useEffect(() => {
    if (!rolling) { setShown(value ?? '?'); return }
    const t = setInterval(() => setShown(1 + Math.floor(Math.random() * (sides || 20))), 70)
    return () => clearInterval(t)
  }, [rolling, value, sides])
  const round = sides === 100
  return (
    <svg className={'die' + (rolling ? ' tumble' : ' land') + (dim ? ' dim' : '') + (tone ? ' ' + tone : '')} viewBox="0 0 100 100" role="img"
      aria-label={rolling ? 'кубик катится' : `d${sides}: ${value}`}>
      {round ? <circle cx="50" cy="50" r="45" /> : <polygon points={SHAPES[sides] ?? SHAPES[20]} />}
      <text x="50" y="60" textAnchor="middle">{shown}</text>
      <text x="50" y="90" textAnchor="middle" className="sd">d{sides}</text>
    </svg>
  )
}

// describe превращает результат в понятные строки: что бросали, что выпало, чем кончилось.
export function describe(r) {
  const nat = r.sides === 20 ? r.kept : 0
  if (r.kind === 'attack') {
    return {
      title: r.label,
      formula: `d20 [${r.rolls.join(', ')}] → ${r.total} против КД ${r.target}`,
      outcome: r.hit ? `${r.crit ? 'Критическое попадание! ' : 'Попадание. '}Урон ${r.damage}${r.damageText ? ` (${r.damageText})` : ''}` : 'Промах',
      tone: r.hit ? (r.crit ? 'crit' : 'ok') : 'bad',
      big: r.hit ? r.damage : r.total,
      bigLabel: r.hit ? 'урон' : 'атака',
    }
  }
  const adv = r.mode === 'adv' ? 'преимущество' : r.mode === 'dis' ? 'помеха' : ''
  const sum = r.sides === 20 ? `d20 [${r.rolls.join(', ')}]${adv ? ` — ${adv}, берём ${r.kept}` : ''}` : `[${r.rolls.join(' + ')}]`
  const bonus = r.bonus ? ` ${r.bonus > 0 ? '+' : '−'} ${Math.abs(r.bonus)}` : ''
  let outcome = '', tone = ''
  if (r.dc > 0) { outcome = `${r.success ? 'Успех' : 'Провал'} — сложность ${r.dc}`; tone = r.success ? 'ok' : 'bad' }
  if (nat === 20) { outcome = ('Натуральная 20! ' + outcome).trim(); tone = 'crit' }
  if (nat === 1) { outcome = ('Натуральная 1. ' + outcome).trim(); tone = 'bad' }
  const kindName = { check: 'Проверка', save: 'Спасбросок', skill: 'Навык', roll: 'Бросок' }[r.kind] ?? ''
  const title = r.kind === 'roll' ? `Бросок ${r.label}` : r.label.includes(':') ? r.label : `${kindName} ${r.label}`
  return { title, formula: `${sum}${bonus} = ${r.total}`, outcome, tone, big: r.total, bigLabel: '' }
}

function Tray({ tray }) {
  if (!tray) return <div className="tray empty" aria-hidden="true"><small>Брось кубики — они покатятся здесь</small></div>
  if (tray.rolling) {
    const n = Math.min(tray.count ?? 1, 8)
    return <div className="tray" aria-live="polite">{Array.from({ length: n }, (_, i) => <Die key={i} sides={tray.sides ?? 20} rolling />)}</div>
  }
  const r = tray.r
  const d = describe(r)
  const keep = r.sides === 20 && r.rolls.length === 2
  const shown = r.rolls.slice(0, 10)
  return (
    <div className="tray" aria-live="polite">
      {shown.map((v, i) => <Die key={i} sides={r.sides || 20} value={v} dim={keep && v !== r.kept && !(i === 1 && r.rolls[0] === r.rolls[1])} tone={r.sides === 20 && v === 20 ? 'crit' : r.sides === 20 && v === 1 ? 'bad' : ''} />)}
      {r.rolls.length > 10 && <small>+{r.rolls.length - 10}</small>}
      <div className={'tsum ' + d.tone}><b>{d.big}</b><span>{d.bigLabel || (r.sides === 20 ? 'итог' : 'сумма')}</span></div>
    </div>
  )
}

export default function Dice({ mode, setMode, dc, setDc, rolls, roll, tray, clear }) {
  const [expr, setExpr] = useState('1d20')
  const hint = (e) => { const m = /^\s*(\d*)d(\d+)/i.exec(e); return m ? { count: parseInt(m[1] || '1', 10), sides: parseInt(m[2], 10) } : { count: 1, sides: 20 } }
  const go = (e = expr) => roll(() => api.Roll(e, mode), hint(e))
  return (
    <aside className="box dice">
      <h2>Кубики</h2>
      <Tray tray={tray} />
      <div className="seg" role="group" aria-label="Режим броска d20">
        {MODES.map(([k, v]) => <button key={k} aria-pressed={mode === k} onClick={() => setMode(k)}>{v}</button>)}
      </div>
      <label className="inline">Сложность (СЛ) <input type="number" value={dc || ''} min="0" placeholder="—" onChange={(e) => setDc(parseInt(e.target.value, 10) || 0)} /></label>
      <form className="row tight" onSubmit={(e) => { e.preventDefault(); go() }}>
        <input value={expr} onChange={(e) => setExpr(e.target.value)} aria-label="Выражение кубиков" placeholder="2d6+3" />
        <button className="primary">Бросить</button>
      </form>
      <div className="quick">{[4, 6, 8, 10, 12, 20, 100].map((d) => <button key={d} onClick={() => { setExpr('1d' + d); go('1d' + d) }}>d{d}</button>)}</div>
      <div className="logbar"><h3>Журнал бросков <small>{rolls.length}</small></h3>
        <button className="ghost small" disabled={!rolls.length} onClick={clear}>Очистить</button></div>
      {rolls.length === 0 && <p className="hint">Здесь появятся проверки, спасброски, атаки и свободные броски — с расшифровкой, что выпало и чем всё кончилось.</p>}
      <ul className="rolls">
        {rolls.map((r, i) => {
          const d = describe(r)
          return (
            <li key={r.id ?? i} className={'roll ' + d.tone + (i === 0 ? ' new' : '')}>
              <div className="rn"><b>{d.big}</b><small>{d.bigLabel}</small></div>
              <div className="rb">
                <strong>{d.title}</strong>
                <span className="rf">{d.formula}</span>
                {d.outcome && <em>{d.outcome}</em>}
              </div>
            </li>
          )
        })}
      </ul>
    </aside>
  )
}
