import { lazy, Suspense, useCallback, useEffect, useRef, useState } from 'react'
import { api } from './api'
import Sheet from './Sheet.jsx'
import Dice from './Dice.jsx'
import { Avatar, Art } from './Portrait.jsx'
import { raceArt } from './art.js'
import { Icon } from './Icon.jsx'
import { reducedMotion, sleep } from './motion.js'
import { describe } from './rollText.js'
import { playLand, playRoll } from './sound.js'

// Редкие экраны грузятся по требованию: стартовый бандл меньше, первый экран открывается быстрее.
const Encounter = lazy(() => import('./Encounter.jsx'))
const Library = lazy(() => import('./Library.jsx'))
const MapsView = lazy(() => import('./Maps.jsx'))
const GMScreen = lazy(() => import('./Panels.jsx').then((m) => ({ default: m.GMScreen })))
const LogView = lazy(() => import('./Panels.jsx').then((m) => ({ default: m.LogView })))

const ROLL_MS = 900
const MAX_ROLLS = 100
const TOAST_MS = 4000
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
  // rev растёт после отмены и загрузки сохранения: экраны со своим состоянием (бой, карты) перечитывают его
  const [rev, setRev] = useState(0)
  const throwSeq = useRef(0)

  // guard: любая ошибка бэкенда показывается пользователю, а не теряется в консоли
  const guard = useCallback(async (f) => {
    try { setErr(''); return await f() } catch (e) { setErr(String(e?.message ?? e)) }
  }, [])
  const reload = useCallback(() => guard(async () => {
    const [cs, items] = await Promise.all([api.Characters(), api.ItemLibrary()])
    setChars(cs); setLib(items)
  }), [guard])
  useEffect(() => {
    guard(async () => { setCat(await api.Catalog()); await reload(); const w = await api.Warning(); if (w) setErr(w) })
  }, [guard, reload])
  useEffect(() => { if (!info) return; const t = setTimeout(() => setInfo(''), TOAST_MS); return () => clearTimeout(t) }, [info])

  const hero = chars.find((c) => c.id === sel) ?? chars.find((c) => c.kind !== 'monster') ?? chars[0]
  const rs = hero && cat.find((r) => r.id === hero.ruleset)
  const put = useCallback((v) => { if (!v) return; setChars((cs) => cs.map((c) => (c.id === v.id ? v : c))); if (v.msg) setInfo(v.msg) }, [])
  const addRolls = useCallback((list) => setRolls((x) => [...list.map((r) => ({ ...r, id: ++seq })).reverse(), ...x].slice(0, MAX_ROLLS)), [])

  // throwDice: бросок с анимацией. f возвращает результат или список (серия атак); hint – какие кубики катить, пока ждём.
  // Если пока кубики катились, начался новый бросок, лоток показывает только последний.
  const throwDice = useCallback((f, hint) => guard(async () => {
    const id = ++throwSeq.current
    const h = { sides: 20, count: 1, ...hint }
    setTray({ rolling: true, id, ...h })
    playRoll(h.count, ROLL_MS)
    try {
      const [r] = await Promise.all([f(), sleep(reducedMotion() ? 0 : ROLL_MS)])
      const list = (Array.isArray(r) ? r : [r]).filter(Boolean)
      if (!list.length) { if (throwSeq.current === id) setTray(null); return r }
      addRolls(list)
      if (throwSeq.current === id) {
        setTray({ rolling: false, id, results: list })
        playLand(list.some((x) => describe(x).tone === 'crit') ? 'crit' : describe(list[0]).tone)
      }
      return r
    } catch (e) {
      if (throwSeq.current === id) setTray(null)
      throw e
    }
  }), [guard, addRolls])
  const roll = useCallback((f) => throwDice(f, { sides: 20, count: mode ? 2 : 1 }), [throwDice, mode])
  const open = (id) => { setSel(id); setView('sheet'); setCreating(false) }

  const create = (a) => guard(async () => {
    const v = await api.CreateCharacter(...a)
    setChars((cs) => [...cs, v]); open(v.id)
  })
  const remove = (id) => guard(async () => { await api.DeleteCharacter(id); setSel(null); await reload() })
  const save = useCallback(() => guard(async () => { const s = await api.QuickSave(); setInfo('Сохранено: ' + s.name) }), [guard])
  const saveRef = useRef(save)
  saveRef.current = save
  useEffect(() => {
    const h = (e) => { if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') { e.preventDefault(); saveRef.current() } }
    window.addEventListener('keydown', h); return () => window.removeEventListener('keydown', h)
  }, [])
  const restored = useCallback(async () => { await reload(); setRev((n) => n + 1) }, [reload])
  const undo = () => guard(async () => { await api.Undo(); await restored(); setInfo('Последнее действие отменено') })

  const heroes = chars.filter((c) => c.kind !== 'monster')
  const monsters = chars.filter((c) => c.kind === 'monster')
  const li = (c, i) => (
    <li key={c.id} style={{ '--i': i }}><button aria-current={view === 'sheet' && !creating && hero?.id === c.id} onClick={() => open(c.id)} className={c.dead ? 'dead' : c.hp <= 0 ? 'down' : ''}>
      <Avatar hero={c} size={36} /><span className="ln"><span>{c.name}</span><small>{c.dead ? 'погиб' : `${c.kind === 'monster' ? 'CR ' + c.stat?.cr : 'ур. ' + c.level} · ${c.hp}/${c.derived?.maxHp}`}</small></span>
      <i className="mini" style={{ '--p': Math.round((Math.max(0, c.hp) / Math.max(1, c.derived?.maxHp ?? 1)) * 100) + '%' }} />
    </button></li>
  )
  const screen = creating ? 'create' : view
  const props = { chars, cat, guard, reload, notify: setInfo, rev }

  return (
    <div className="app">
      <nav className="box rail">
        <h1><span className="sigil" aria-hidden="true"><Icon n="book" size={26} /></span>Книга героев</h1>
        <button className="primary" onClick={() => setCreating(true)}><Icon n="hero" /> Новый герой</button>
        <button className="savebtn" onClick={save} title="Быстрое сохранение (Ctrl+S). Хранятся пять последних; загрузить – в Журнале."><Icon n="save" size={20} /> Сохранить <kbd>Ctrl+S</kbd></button>
        <div className="navs">
          {VIEWS.map(([k, ic, v]) => <button key={k} className="ghost nav" aria-pressed={view === k && !creating} onClick={() => { setCreating(false); setView(view === k ? 'sheet' : k) }}><Icon n={ic} /> {v}</button>)}
          <button className="ghost nav" onClick={undo}><Icon n="undo" /> Отменить</button>
        </div>
        {heroes.length > 0 && <h4>Герои <small>{heroes.length}</small></h4>}
        <ul>{heroes.map(li)}</ul>
        {monsters.length > 0 && <><h4>Существа <small>{monsters.length}</small></h4><ul>{monsters.map(li)}</ul></>}
        <small className="credits">Иконки: <a href="https://game-icons.net" target="_blank" rel="noreferrer">game-icons.net</a> (CC BY 3.0)</small>
      </nav>
      <main className="box">
        <div className="view" key={screen === 'sheet' ? 'sheet-' + (hero?.id ?? '') : screen}>
          <Suspense fallback={<div className="loading" aria-busy="true"><Icon n="dice" size={40} /></div>}>
            {screen === 'create' && <Create cat={cat} onCreate={create} onCancel={() => setCreating(false)} />}
            {screen === 'battle' && <Encounter {...props} mode={mode} throwDice={throwDice} />}
            {screen === 'gm' && <GMScreen {...props} put={put} open={open} />}
            {screen === 'library' && <Library {...props} lib={lib} />}
            {screen === 'map' && <MapsView guard={guard} rev={rev} chars={chars} cat={cat} reload={reload} notify={setInfo} />}
            {screen === 'log' && <LogView guard={guard} restored={restored} />}
            {screen === 'sheet' && (hero && rs
              ? <Sheet key={hero.id} hero={hero} rs={rs} lib={lib} chars={chars} reload={reload} mode={mode} dc={dc} roll={roll} guard={guard} put={put} remove={remove} />
              : <Welcome onCreate={() => setCreating(true)} />)}
          </Suspense>
        </div>
      </main>
      <Dice mode={mode} setMode={setMode} dc={dc} setDc={setDc} rolls={rolls} throwDice={throwDice} tray={tray} clear={() => { setRolls([]); setTray(null) }} />
      {err && <div className="toast bad-toast" role="alert" onClick={() => setErr('')}>{err}</div>}
      {!err && info && <div className="toast" role="status" key={info} onClick={() => setInfo('')}>{info}</div>}
    </div>
  )
}

function Welcome({ onCreate }) {
  return (
    <div className="welcome">
      <div className="wdice" aria-hidden="true"><Icon n="dice" size={72} /></div>
      <h2>Добро пожаловать за стол</h2>
      <p className="hint">Создайте первого героя – или откройте бестиарий в «Библиотеке» и выставьте врагов.</p>
      <button className="primary big" onClick={onCreate}><Icon n="hero" /> Создать героя</button>
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
      <div className="cgrid">
        <div className="cform">
          <Art className="art big" key={r.id + s} html={raceArt(rs.id, r.id, s)} label={r.name} />
          <label>Имя<input value={name} onChange={(e) => setName(e.target.value)} placeholder="Новый герой" autoFocus /></label>
          <label>Раса<select value={r.id} onChange={(e) => { setRace(e.target.value); setSub('') }}>
            {rs.races.map((x) => <option key={x.id} value={x.id}>{x.name}</option>)}</select></label>
          {r.subs?.length > 0 && <label>Подраса<select value={s} onChange={(e) => setSub(e.target.value)}>
            {r.subs.map((x) => <option key={x.id} value={x.id}>{x.name}</option>)}</select></label>}
          {cl && <label>Класс<select value={cl.id} onChange={(e) => setCls(e.target.value)}>
            {rs.classes.map((x) => <option key={x.id} value={x.id}>{x.name} (d{x.hitDie})</option>)}</select></label>}
          {bonus && <p className="hint">Бонусы расы: {bonus}</p>}
          {r.traits?.length > 0 && <div className="chips">{r.traits.map((t) => <span key={t} className="chip">{t}</span>)}</div>}
          <div className="row">
            <button className="primary" onClick={() => onCreate([rs.id, name, r.id, s, cl?.id ?? ''])}>Создать</button>
            <button className="ghost" onClick={onCancel}>Отмена</button>
          </div>
        </div>
        <div className="cinfo">
          <div className="lore" key={'r' + r.id + s}>
            <h3>{r.name}</h3>{r.desc && <p>{r.desc}</p>}
            {r.lore && <details><summary>История народа</summary><p>{r.lore}</p></details>}
            {sr?.desc && <p><b>{sr.name}.</b> {sr.desc}</p>}
          </div>
          {cl && <div className="lore" key={'c' + cl.id}>
            <h3><Icon n={cl.id} size={22} /> {cl.name}</h3>{cl.desc && <p>{cl.desc}</p>}
            {cl.lore && <details><summary>История класса</summary><p>{cl.lore}</p></details>}
          </div>}
        </div>
      </div>
    </section>
  )
}
