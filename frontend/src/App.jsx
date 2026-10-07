import { useCallback, useEffect, useState } from 'react'
import { api } from './api'
import Sheet from './Sheet.jsx'
import Dice from './Dice.jsx'
import { Avatar, Art } from './Portrait.jsx'
import { raceArt } from './art.js'
import Encounter from './Encounter.jsx'
import Library from './Library.jsx'
import MapsView from './Maps.jsx'
import { GMScreen, LogView } from './Panels.jsx'
import { Icon } from './Icon.jsx'

const sleep = (ms) => new Promise((r) => setTimeout(r, ms))
const reduced = typeof matchMedia === 'function' && matchMedia('(prefers-reduced-motion: reduce)').matches
let seq = 0

const VIEWS = [['battle', 'battle', 'Бой'], ['gm', 'gm', 'Экран мастера'], ['library', 'library', 'Библиотека'], ['map', 'map', 'Карта'], ['log', 'log', 'Журнал']]

export default function App() {
  const [cat, setCat] = useState([])
  const [chars, setChars] = useState([])
  const [lib, setLib] = useState([])
  const [sel, setSel] = useState(null)
  const [view, setView] = useState('sheet')
  const [mode, setMode] = useState('')
  const [dc, setDc] = useState(0)
  const [rolls, setRolls] = useState([])
  const [tray, setTray] = useState(null)
  const [err, setErr] = useState('')
  const [info, setInfo] = useState('')
  const [creating, setCreating] = useState(false)

  // guard: любая ошибка бэкенда показывается пользователю, а не теряется в консоли
  const guard = useCallback(async (f) => {
    try { setErr(''); return await f() } catch (e) { setErr(String(e?.message ?? e)) }
  }, [])
  const reload = useCallback(() => guard(async () => { setChars(await api.Characters()); setLib(await api.ItemLibrary()) }), [guard])
  useEffect(() => {
    guard(async () => { setCat(await api.Catalog()); await reload(); const w = await api.Warning(); if (w) setErr(w) })
  }, [])
  useEffect(() => { if (!info) return; const t = setTimeout(() => setInfo(''), 4000); return () => clearTimeout(t) }, [info])

  const hero = chars.find((c) => c.id === sel) ?? chars.find((c) => c.kind !== 'monster') ?? chars[0]
  const rs = hero && cat.find((r) => r.id === hero.ruleset)
  const put = (v) => { if (!v) return; setChars((cs) => cs.map((c) => (c.id === v.id ? v : c))); if (v.msg) setInfo(v.msg) }
  const addRolls = (rs) => setRolls((x) => [...rs.map((r) => ({ ...r, id: ++seq })).reverse(), ...x].slice(0, 100))
  const roll = (f, hint) => guard(async () => {
    setTray({ rolling: true, ...(hint ?? { sides: 20, count: 1 }) })
    try {
      const [r] = await Promise.all([f(), sleep(reduced ? 0 : 700)])
      setRolls((x) => [{ ...r, id: ++seq }, ...x].slice(0, 100)); setTray({ rolling: false, r })
    } catch (e) { setTray(null); throw e }
  })
  const open = (id) => { setSel(id); setView('sheet') }

  const create = (a) => guard(async () => {
    const v = await api.CreateCharacter(...a)
    setChars((cs) => [...cs, v]); open(v.id); setCreating(false)
  })
  const remove = (id) => guard(async () => { await api.DeleteCharacter(id); setSel(null); await reload() })
  const save = () => guard(async () => { const s = await api.QuickSave(); setInfo('Сохранено: ' + s.name) })
  useEffect(() => {
    const h = (e) => { if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') { e.preventDefault(); save() } }
    window.addEventListener('keydown', h); return () => window.removeEventListener('keydown', h)
  })
  const undo = () => guard(async () => { await api.Undo(); await reload(); setInfo('Последнее действие отменено') })

  const li = (c) => (
    <li key={c.id}><button aria-current={view === 'sheet' && hero?.id === c.id} onClick={() => open(c.id)} className={c.dead ? 'dead' : ''}>
      <Avatar hero={c} size={34} /><span className="ln"><span>{c.name}</span><small>{c.dead ? 'погиб' : `${c.kind === 'monster' ? 'CR ' + c.stat.cr : 'ур. ' + c.level} · ${c.hp}/${c.derived?.maxHp}`}</small></span>
    </button></li>
  )
  const heroes = chars.filter((c) => c.kind !== 'monster')
  const monsters = chars.filter((c) => c.kind === 'monster')

  return (
    <div className="app">
      <nav className="box rail">
        <h1>Книга героев</h1>
        <button className="primary" onClick={() => setCreating(true)}><Icon n="hero" /> Новый герой</button>
        <button className="savebtn" onClick={save} title="Быстрое сохранение (Ctrl+S). Хранятся пять последних; загрузить — в Журнале."><Icon n="save" size={20} /> Сохранить <kbd>Ctrl+S</kbd></button>
        <div className="navs">
          {VIEWS.map(([k, ic, v]) => <button key={k} className="ghost" aria-pressed={view === k && !creating} onClick={() => { setCreating(false); setView(view === k ? 'sheet' : k) }}><Icon n={ic} /> {v}</button>)}
          <button className="ghost" onClick={undo}><Icon n="undo" /> Отменить</button>
        </div>
        <ul>{heroes.map(li)}</ul>
        {monsters.length > 0 && <><h4>Существа</h4><ul>{monsters.map(li)}</ul></>}
        <small className="credits">Иконки: <a href="https://game-icons.net" target="_blank" rel="noreferrer">game-icons.net</a> (CC BY 3.0)</small>
      </nav>
      <main className="box">
        {creating && <Create cat={cat} onCreate={create} onCancel={() => setCreating(false)} />}
        {!creating && view === 'battle' && <Encounter chars={chars} cat={cat} mode={mode} guard={guard} reload={reload} addRolls={addRolls} notify={setInfo} />}
        {!creating && view === 'gm' && <GMScreen chars={chars} cat={cat} guard={guard} reload={reload} put={put} open={open} />}
        {!creating && view === 'library' && <Library cat={cat} lib={lib} chars={chars} guard={guard} reload={reload} notify={setInfo} />}
        {!creating && view === 'map' && <MapsView guard={guard} />}
        {!creating && view === 'log' && <LogView guard={guard} reload={reload} />}
        {!creating && view === 'sheet' && (hero && rs
          ? <Sheet key={hero.id} hero={hero} rs={rs} lib={lib} chars={chars} reload={reload} mode={mode} dc={dc} roll={roll} guard={guard} put={put} remove={remove} />
          : <p className="empty">Создайте первого героя.</p>)}
      </main>
      <Dice mode={mode} setMode={setMode} dc={dc} setDc={setDc} rolls={rolls} roll={roll} tray={tray} clear={() => { setRolls([]); setTray(null) }} />
      {err && <div className="toast bad-toast" role="alert" onClick={() => setErr('')}>{err}</div>}
      {!err && info && <div className="toast" role="status" onClick={() => setInfo('')}>{info}</div>}
    </div>
  )
}

function Create({ cat, onCreate, onCancel }) {
  const rs = cat[0]
  const [name, setName] = useState('')
  const [race, setRace] = useState('')
  const [sub, setSub] = useState('')
  const [cls, setCls] = useState('')
  if (!rs) return null
  const r = rs.races.find((x) => x.id === race) ?? rs.races[0]
  const sr = r.subs?.find((x) => x.id === sub) ?? r.subs?.[0]
  const s = sr?.id ?? ''
  const cl = rs.classes?.find((x) => x.id === cls) ?? rs.classes?.[0]
  const bonus = Object.entries(r.bonus ?? {}).filter(([, v]) => v).map(([k, v]) => `${rs.abilities.find((a) => a.id === k)?.name ?? k} +${v}`).join(', ')
  return (
    <section className="create">
      <h2>Новый герой</h2>
      <Art className="art big" html={raceArt(rs.id, r.id, s)} label={r.name} />
      <label>Имя<input value={name} onChange={(e) => setName(e.target.value)} placeholder="Новый герой" autoFocus /></label>
      <label>Раса<select value={r.id} onChange={(e) => { setRace(e.target.value); setSub('') }}>
        {rs.races.map((x) => <option key={x.id} value={x.id}>{x.name}</option>)}</select></label>
      {r.subs?.length > 0 && <label>Подраса<select value={s} onChange={(e) => setSub(e.target.value)}>
        {r.subs.map((x) => <option key={x.id} value={x.id}>{x.name}</option>)}</select></label>}
      {cl && <label>Класс<select value={cl.id} onChange={(e) => setCls(e.target.value)}>
        {rs.classes.map((x) => <option key={x.id} value={x.id}>{x.name} (d{x.hitDie})</option>)}</select></label>}
      {bonus && <p className="hint">Бонусы расы: {bonus}</p>}
      {r.traits?.length > 0 && <div className="chips">{r.traits.map((t) => <span key={t} className="chip">{t}</span>)}</div>}
      <div className="lore">
        <h3>{r.name}</h3>{r.desc && <p>{r.desc}</p>}
        {r.lore && <details><summary>История народа</summary><p>{r.lore}</p></details>}
        {sr?.desc && <p><b>{sr.name}.</b> {sr.desc}</p>}
      </div>
      {cl && <div className="lore">
        <h3><Icon n={cl.id} size={22} /> {cl.name}</h3>{cl.desc && <p>{cl.desc}</p>}
        {cl.lore && <details><summary>История класса</summary><p>{cl.lore}</p></details>}
      </div>}
      <div className="row">
        <button className="primary" onClick={() => onCreate([rs.id, name, r.id, s, cl?.id ?? ''])}>Создать</button>
        <button className="ghost" onClick={onCancel}>Отмена</button>
      </div>
    </section>
  )
}
