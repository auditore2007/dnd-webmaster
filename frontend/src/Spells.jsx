import { Icon } from './Icon.jsx'
import { useEffect, useMemo, useState } from 'react'
import { api } from './api'
import { DMG, MODES } from './labels.js'

const CLASS_NAMES = { bard: 'бард', cleric: 'жрец', druid: 'друид', paladin: 'паладин', ranger: 'следопыт', sorcerer: 'чародей', warlock: 'колдун', wizard: 'волшебник', artificer: 'изобретатель' }
export const lvlName = (l) => (l ? `${l} ур.` : 'заговор')

// filterSpells – общий поиск по каталогу: текст, уровень, класс, только автоматические.
export function filterSpells(list, { q = '', level = '', cls = '', auto = false }) {
  const s = q.trim().toLowerCase()
  return list.filter((x) => (level === '' || x.level === +level) && (!cls || x.classes.includes(cls)) && (!auto || x.mode) && (!s || x.name.toLowerCase().includes(s)))
}

export function SpellFilters({ f, setF, classes = true }) {
  return (
    <div className="row tight">
      <input placeholder="Поиск заклинания" aria-label="Поиск заклинания" value={f.q} onChange={(e) => setF({ ...f, q: e.target.value })} />
      <select aria-label="Уровень заклинания" value={f.level} onChange={(e) => setF({ ...f, level: e.target.value })}>
        <option value="">Все уровни</option><option value="0">Заговоры</option>{[1, 2, 3, 4, 5, 6, 7, 8, 9].map((l) => <option key={l} value={l}>{l} уровень</option>)}</select>
      {classes && <select aria-label="Класс" value={f.cls} onChange={(e) => setF({ ...f, cls: e.target.value })}>
        <option value="">Все классы</option>{Object.entries(CLASS_NAMES).map(([k, v]) => <option key={k} value={k}>{v}</option>)}</select>}
      <label className="eq"><input type="checkbox" checked={f.auto} onChange={(e) => setF({ ...f, auto: e.target.checked })} /> ⚡ с автоматикой</label>
    </div>
  )
}

const blank = { name: '', level: 1, school: '', desc: '', mode: '', dmg: '2d6', type: 'fire', save: 'dex', half: true, area: false, up: '1d6', scale: false }

// SpellEditor – создание и правка собственного заклинания, в том числе с автоматикой урона.
function SpellEditor({ guard, initial, abilities, onDone }) {
  const [f, setF] = useState(initial ? { id: initial.id, name: initial.name, level: initial.level, school: initial.school ?? '', desc: initial.desc ?? '', mode: initial.auto?.mode ?? '', dmg: initial.auto?.dmg ?? '2d6', type: initial.auto?.type || 'fire', save: initial.auto?.save || 'dex', half: initial.auto?.half ?? true, area: initial.auto?.area ?? false, up: initial.auto?.up ?? '', scale: initial.auto?.scale ?? false } : blank)
  const set = (k) => (e) => setF({ ...f, [k]: e.target.type === 'checkbox' ? e.target.checked : e.target.type === 'number' ? parseInt(e.target.value, 10) || 0 : e.target.value })
  const save = () => guard(async () => {
    await api.SaveSpellTemplate({ id: f.id ?? '', name: f.name, level: f.level, school: f.school, desc: f.desc,
      auto: f.mode ? { mode: f.mode, dmg: f.dmg, type: f.mode === 'heal' ? '' : f.type, save: f.mode === 'save' ? f.save : '', half: f.mode === 'save' && f.half, area: f.mode === 'save' || f.mode === 'heal' ? f.area : false, up: f.up, scale: f.scale && f.level === 0 } : null })
    onDone()
  })
  return (
    <div className="editor">
      <h3>{f.id ? 'Правка заклинания' : 'Новое заклинание'}</h3>
      <div className="grid2">
        <label>Название<input value={f.name} onChange={set('name')} /></label>
        <label>Уровень (0 – заговор)<input type="number" min="0" max="9" value={f.level} onChange={set('level')} /></label>
        <label>Школа<input value={f.school} onChange={set('school')} placeholder="Воплощение" /></label>
        <label>Как считается в бою<select value={f.mode} onChange={set('mode')}>{Object.entries(MODES).map(([k, v]) => <option key={k} value={k}>{v}</option>)}</select></label>
      </div>
      {f.mode && <div className="row tight">
        <label className="eq">{f.mode === 'heal' ? 'Лечение' : 'Урон'} <input value={f.dmg} onChange={set('dmg')} placeholder="8d6" style={{ width: 100 }} /></label>
        {f.mode !== 'heal' && <select aria-label="Тип урона" value={f.type} onChange={set('type')}>{Object.entries(DMG).map(([k, v]) => <option key={k} value={k}>{v}</option>)}</select>}
        {f.mode === 'save' && <select aria-label="Спасбросок" value={f.save} onChange={set('save')}>{abilities.map((a) => <option key={a.id} value={a.id}>спасбросок: {a.name}</option>)}</select>}
        {f.mode === 'save' && <label className="eq"><input type="checkbox" checked={f.half} onChange={set('half')} /> половина при успехе</label>}
        {(f.mode === 'save' || f.mode === 'heal') && <label className="eq"><input type="checkbox" checked={f.area} onChange={set('area')} /> несколько целей</label>}
        <label className="eq">+кубики за ячейку выше <input value={f.up} onChange={set('up')} placeholder="1d6" style={{ width: 80 }} /></label>
        {f.level === 0 && <label className="eq"><input type="checkbox" checked={f.scale} onChange={set('scale')} /> растёт с 5/11/17 уровня</label>}
      </div>}
      <label>Описание<textarea value={f.desc} onChange={set('desc')} /></label>
      <div className="row"><button className="primary" disabled={!f.name.trim()} onClick={save}>Сохранить</button><button className="ghost" onClick={() => onDone()}>Отмена</button></div>
    </div>
  )
}

// SpellsTab – вкладка «Заклинания» библиотеки.
export function SpellsTab({ cat, guard }) {
  const rs = cat[0]
  const [f, setF] = useState({ q: '', level: '', cls: '', auto: false })
  const [limit, setLimit] = useState(60)
  const [mine, setMine] = useState([])
  const [edit, setEdit] = useState(null)
  const [open, setOpen] = useState('')
  const load = () => guard(async () => setMine(await api.SpellLibrary()))
  useEffect(() => { load() }, [])
  const list = useMemo(() => filterSpells(rs?.spells ?? [], f), [rs, f])
  if (!rs) return null
  if (edit) return <SpellEditor guard={guard} initial={edit === 'new' ? null : edit} abilities={rs.abilities} onDone={() => { setEdit(null); load() }} />
  return (
    <div>
      <p className="hint">В каталоге {rs.spells.length} заклинаний D&D 5e (PHB, SRD, Xanathar, Tasha – по памяти, сверяйте с книгой). Значок ⚡ – урон или лечение считаются автоматически; остальные тратят ячейку, а эффект применяет мастер. Выучить заклинание можно на листе героя.</p>
      <SpellFilters f={f} setF={(v) => { setF(v); setLimit(60) }} />
      <small>Найдено: {list.length}</small>
      {list.slice(0, limit).map((s) => (
        <div className="irow col" key={s.id}>
          <button className="link" onClick={() => setOpen(open === s.id ? '' : s.id)}><Icon n={s.school} size={16} /> <b>{s.name}</b>{s.mode ? ' ⚡' : ''}{s.conc ? ' ◎' : ''} <small>{lvlName(s.level)} · {s.school} · {s.classes.map((c) => CLASS_NAMES[c] ?? c).join(', ')}</small></button>
          {open === s.id && <p className="hint">{s.desc}</p>}
        </div>))}
      {list.length > limit && <button className="ghost" onClick={() => setLimit(limit + 100)}>Показать ещё ({list.length - limit})</button>}

      <h3>Мои заклинания <button className="primary" onClick={() => setEdit('new')}>＋ Создать</button></h3>
      {mine.length === 0 && <p className="hint">Свои заклинания можно создать с готовой автоматикой урона и лечения и выдавать любому герою.</p>}
      {mine.map((s) => <div className="irow" key={s.id}><span><b>{s.name}</b> <small>{lvlName(s.level)}{s.school ? ` · ${s.school}` : ''}{s.auto?.mode ? ` · ⚡ ${s.auto.dmg}` : ''}</small></span>
        <button onClick={() => setEdit(s)}>Править</button>
        <button aria-label={`Удалить ${s.name}`} onClick={() => guard(async () => { await api.DeleteSpellTemplate(s.id); load() })}>✕</button></div>)}
    </div>
  )
}
