import { memo, useState } from 'react'
import { api } from './api'
import Die3D from './Die3D.jsx'
import { useCountUp } from './motion.js'
import { attackSummary, describe, diceHint, dimmed, toneOf } from './rollText.js'
import { setSoundOn, soundOn } from './sound.js'
import { Icon } from './Icon.jsx'

export { describe } from './rollText.js'

const MODES = [['', 'Обычный'], ['adv', 'Преимущество'], ['dis', 'Помеха']]
const QUICK = [4, 6, 8, 10, 12, 20, 100]
const MAX_TRAY_DICE = 10
const SPARKS = 14

// trayView сводит результат(ы) к тому, что видно в лотке: кубики, большое число, подпись и тон.
function trayView(results) {
  const atk = results.filter((r) => r.kind === 'attack')
  if (atk.length > 1) {
    const s = attackSummary(atk)
    const dice = atk.flatMap((r) => { const dm = dimmed({ ...r, sides: 20 }); return (r.rolls ?? []).map((v, i) => ({ v, sides: 20, dim: dm[i] })) })
    return { dice, big: s.damage, label: `урон · попаданий ${s.hits} из ${s.total}`, tone: s.crit ? 'crit' : s.hits ? 'ok' : 'bad' }
  }
  const r = results[0]
  const d = describe(r)
  const sides = r.kind === 'attack' ? 20 : r.sides || 20
  const dm = dimmed({ ...r, sides })
  const dice = (r.rolls ?? []).map((v, i) => ({ v, sides, dim: dm[i] }))
  return { dice, big: d.big, label: d.bigLabel || (sides === 20 ? 'итог' : 'сумма'), tone: d.tone }
}

function Total({ value, label, tone }) {
  const n = useCountUp(value, { fromZero: true })
  return <div className={'tsum ' + tone}><b>{n}</b><span>{label}</span></div>
}

function Tray({ tray }) {
  if (!tray) return <div className="tray empty" aria-hidden="true"><Icon n="dice" size={34} /><small>Брось кубики – они покатятся здесь</small></div>
  if (tray.rolling) {
    const n = Math.min(tray.count ?? 1, 8)
    return <div className="tray" aria-live="polite">{Array.from({ length: n }, (_, i) => <Die3D key={i} index={i} sides={tray.sides ?? 20} rolling />)}</div>
  }
  const v = trayView(tray.results)
  const shown = v.dice.slice(0, MAX_TRAY_DICE)
  const fx = v.tone === 'crit' ? ' crit' : v.dice.some((d) => !d.dim && d.sides === 20 && d.v === 1) ? ' fumble' : ''
  return (
    <div className={'tray landed' + fx} aria-live="polite" key={tray.id}>
      {fx === ' crit' && <div className="sparks" aria-hidden="true">{Array.from({ length: SPARKS }, (_, i) => <i key={i} style={{ '--a': `${(360 / SPARKS) * i}deg` }} />)}</div>}
      {shown.map((d, i) => <Die3D key={i} index={i} sides={d.sides} value={d.v} dim={d.dim} tone={toneOf(d.sides, d.v)} />)}
      {v.dice.length > MAX_TRAY_DICE && <small>+{v.dice.length - MAX_TRAY_DICE}</small>}
      <Total value={v.big} label={v.label} tone={v.tone} />
    </div>
  )
}

const RollItem = memo(function RollItem({ r, fresh }) {
  const d = describe(r)
  return (
    <li className={'roll ' + d.tone + (fresh ? ' new' : '')}>
      <div className="rn"><b>{d.big}</b><small>{d.bigLabel}</small></div>
      <div className="rb">
        <strong>{d.title}</strong>
        <span className="rf">{d.formula}</span>
        {d.outcome && <em>{d.outcome}</em>}
      </div>
    </li>
  )
})

export default function Dice({ mode, setMode, dc, setDc, rolls, throwDice, tray, clear }) {
  const [expr, setExpr] = useState('1d20')
  const [sound, setSound] = useState(soundOn)
  const go = (e = expr) => throwDice(() => api.Roll(e, mode), { ...diceHint(e), ...(mode && /^\s*1?d20\b/i.test(e) ? { count: 2 } : {}) })
  const toggleSound = () => { setSoundOn(!sound); setSound(!sound) }
  return (
    <aside className="box dice">
      <div className="dhead">
        <h2>Кубики</h2>
        <button className="ghost small icon" aria-pressed={sound} onClick={toggleSound} title={sound ? 'Звук кубиков включён' : 'Звук кубиков выключен'} aria-label="Звук кубиков">{sound ? '🔊' : '🔇'}</button>
      </div>
      <Tray tray={tray} />
      <div className="seg" role="group" aria-label="Режим броска d20">
        {MODES.map(([k, v]) => <button key={k} aria-pressed={mode === k} onClick={() => setMode(k)}>{v}</button>)}
      </div>
      <label className="inline">Сложность (СЛ) <input type="number" value={dc || ''} min="0" max="40" placeholder="–" onChange={(e) => setDc(Math.max(0, parseInt(e.target.value, 10) || 0))} /></label>
      <form className="row tight" onSubmit={(e) => { e.preventDefault(); go() }}>
        <input value={expr} onChange={(e) => setExpr(e.target.value)} aria-label="Выражение кубиков" placeholder="2d6+3" />
        <button className="primary">Бросить</button>
      </form>
      <div className="quick">{QUICK.map((d) => <button key={d} className={'qd q' + d} onClick={() => { setExpr('1d' + d); go('1d' + d) }}>d{d}</button>)}</div>
      <div className="logbar"><h3>Журнал бросков <small>{rolls.length}</small></h3>
        <button className="ghost small" disabled={!rolls.length} onClick={clear}>Очистить</button></div>
      {rolls.length === 0 && <p className="hint">Здесь появятся проверки, спасброски, атаки и свободные броски – с расшифровкой, что выпало и чем всё кончилось.</p>}
      <ul className="rolls">{rolls.map((r, i) => <RollItem key={r.id ?? i} r={r} fresh={i === 0} />)}</ul>
    </aside>
  )
}
