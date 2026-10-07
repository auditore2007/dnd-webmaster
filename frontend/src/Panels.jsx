import { useEffect, useState } from 'react'
import { api } from './api'

import { fmt } from './labels.js'

export function GMScreen({ chars, cat, guard, reload, put, open }) {
  const [xp, setXp] = useState(100)
  const heroes = chars.filter((c) => c.kind !== 'monster')
  const hp = (c, n) => guard(async () => put(n < 0 ? await api.ApplyDamage(c.id, -n) : await api.ApplyHeal(c.id, n)))
  return (
    <section>
      <h2>Экран мастера</h2>
      <div className="row tight">
        {[['short', 'Короткий отдых всем'], ['long', 'Долгий отдых всем']].map(([k, v]) =>
          <button key={k} className="ghost" onClick={() => guard(async () => { await api.RestAll(k); await reload() })}>{v}</button>)}
        <input type="number" min="1" value={xp} onChange={(e) => setXp(parseInt(e.target.value, 10) || 1)} aria-label="Опыт каждому" style={{ width: 90 }} />
        <button className="ghost" onClick={() => guard(async () => { for (const c of heroes) if (!c.dead) await api.AddXP(c.id, xp); await reload() })}>Опыт живым героям</button>
        <button className="ghost danger" onClick={() => guard(async () => { await api.ClearMonsters(); await reload() })} title="Убирает со стола всех существ (когда боя нет)">Убрать всех существ</button>
      </div>
      <div className="gm">
        {chars.map((c) => {
          const rs = cat.find((r) => r.id === c.ruleset)
          const d = c.derived ?? {}
          return (
            <div key={c.id} className={'bi' + (c.dead ? ' down' : '')}>
              <div><button className="link" onClick={() => open(c.id)}><b>{c.name}</b></button> <small>{c.kind === 'monster' ? `CR ${c.stat.cr}` : `ур. ${c.level}`} · {rs?.acLabel} {d.ac}</small></div>
              <div className="hp"><div className="meter" style={{ '--p': Math.round((Math.max(0, c.hp) / Math.max(1, d.maxHp)) * 100) + '%' }}><b>{c.hp} / {d.maxHp}</b></div>
                {[-5, -1, 1, 5].map((n) => <button key={n} onClick={() => hp(c, n)}>{fmt(n)}</button>)}</div>
              <div className="chips">
                {c.dead && <span className="chip bad">погиб</span>}
                {!c.dead && c.hp <= 0 && rs?.deathSaves && c.kind !== 'monster' && <span className="chip bad">{c.stable ? 'стабилен' : `смерть: ✓${c.deathOk} ✗${c.deathFail}`}</span>}
                {c.conditions?.map((x) => <span className="chip" key={x}>{x}</span>)}
              </div>
            </div>
          )
        })}
      </div>
    </section>
  )
}

export function LogView({ guard, reload }) {
  const [log, setLog] = useState([])
  const [snaps, setSnaps] = useState([])
  const [name, setName] = useState('')
  const [path, setPath] = useState('')
  const load = () => guard(async () => { setLog(await api.Log()); setSnaps(await api.Snapshots()) })
  useEffect(() => { load() }, [])
  return (
    <section>
      <h2>Журнал и сохранения</h2>
      <h3>Сохранения игры</h3>
      <p className="hint">Игра сохраняется сама после каждого действия. Здесь — именованные копии, к которым можно вернуться.</p>
      <div className="row tight">
        <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Название (необязательно)" aria-label="Название сохранения" />
        <button className="primary" onClick={() => guard(async () => { await api.SaveSnapshot(name); setName(''); await load() })}>Сохранить игру</button>
      </div>
      {snaps.map((s) => <div className="irow" key={s.id}><span><b>{s.name}</b> <small>{s.time}</small></span>
        <button onClick={() => guard(async () => { await api.LoadSnapshot(s.id); await reload(); await load() })}>Загрузить</button>
        <button aria-label={`Удалить ${s.name}`} onClick={() => guard(async () => { await api.DeleteSnapshot(s.id); await load() })}>✕</button></div>)}
      <h3>Журнал сессии</h3>
      <div className="row tight">
        <button className="ghost" onClick={() => guard(async () => setPath(await api.ExportSession()))}>Сохранить в файл</button>
        <button className="ghost danger" onClick={() => guard(async () => { await api.ClearLog(); await load() })}>Очистить</button>
        <button className="ghost" onClick={load}>Обновить</button>
      </div>
      {path && <p className="hint">Сохранено: {path}</p>}
      <ul className="log">{[...log].reverse().map((l, i) => <li key={i}>{l}</li>)}</ul>
    </section>
  )
}
