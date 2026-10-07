import { useState } from 'react'
import { api } from './api'
import { Icon } from './Icon.jsx'

const isMon = (c) => c?.kind === 'monster'

// Ресурсы и способности класса/подкласса. call(promise) выполняет запрос и обновляет экран.
export default function Abilities({ h, chars, call, inFight }) {
  const d = h.derived ?? {}
  const list = (d.features ?? []).filter((f) => f.kind === 'ability')
  const pools = d.pools ?? []
  const [tg, setTg] = useState({})
  if (!list.length && !pools.length) return null
  const allies = chars.filter((c) => !isMon(c) && !c.dead)
  const foes = chars.filter((c) => isMon(c) && !c.dead && c.hp > 0)
  const opts = (f) => (f.target === 'foe' ? foes : allies.slice().sort((a, b) => (a.id === h.id ? -1 : b.id === h.id ? 1 : 0)))
  const sel = (f) => tg[f.key] ?? opts(f)[0]?.id ?? ''
  const lacks = (f) => (f.pool && f.cost > 0 && f.left < f.cost) || (f.target === 'foe' && !opts(f).length)

  return (
    <div className="abil">
      <h3>Способности и ресурсы</h3>
      {pools.length > 0 && <div className="chips">
        {pools.map((p) => <span key={p.key} className={'chip pool' + (p.left === 0 ? ' empty' : '')}
          title={`Восстановление: ${p.rest === 'short' ? 'короткий отдых' : p.rest === 'long' ? 'долгий отдых' : 'конец боя'}`}>
          <Icon n="effect" size={13} /> {p.name}: <b>{p.left}/{p.max}</b></span>)}
      </div>}
      {list.map((f) => (
        <div className="ab" key={f.key}>
          <button className="chip" aria-pressed={!!f.on} disabled={lacks(f)} title={f.desc}
            onClick={() => call(api.UseAbility(h.id, f.key, f.target ? sel(f) : ''))}>
            {f.name}{f.cost > 0 ? ` · −${f.cost}` : ''}{f.on ? ' · готово' : ''}</button>
          {f.target && <select aria-label={`Цель: ${f.name}`} value={sel(f)} onChange={(e) => setTg({ ...tg, [f.key]: e.target.value })}>
            {opts(f).map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}</select>}
          <small>{f.desc}{f.pool && f.max > 0 && !f.cost ? ` (запас ${f.left}/${f.max})` : ''}</small>
        </div>))}
      {inFight && <p className="hint">Способности не тратят действие боя; дополнительные атаки добавляются в текущий ход.</p>}
    </div>
  )
}
