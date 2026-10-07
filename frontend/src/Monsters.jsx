import { useEffect, useMemo, useState } from 'react'
import { api } from './api'
import { Art } from './Portrait.jsx'
import { monsterArt } from './art.js'
import { DMG, KINDS, fmt } from './labels.js'
import { Icon } from './Icon.jsx'

export const artOf = (m) => monsterArt(m.id, m.kind, m.name)

export function MonsterCard({ m, children }) {
  return (
    <div className="mcard">
      <Art className="art mid" html={artOf(m)} label={m.name} />
      <div>
        <b>{m.name}</b> <small><Icon n={m.kind} size={14} /> CR {m.cr} · {KINDS[m.kind] ?? m.kind}{m.custom ? ' · своё' : ''}</small>
        <p className="hint">КД {m.ac} · HP {m.hp} · атака {fmt(m.attack)} · скорость {m.speed}</p>
        <p className="hint">{m.weapons.map((w) => `${w.name} ${w.dice}`).join(' · ')}{m.multi?.length > 1 ? ` · мультиатака ×${m.multi.length}` : ''}</p>
        <div className="chips">
          {m.specials?.map((x) => <span className="chip fx" key={x.key} title={x.desc}>{x.mode === 'area' ? '💥' : '🎯'} {x.name}</span>)}
          {m.traits?.map((t) => <span className="chip" key={t}>{t}</span>)}
          {m.immune?.map((t) => <span className="chip ok" key={'i' + t}>иммунитет: {DMG[t] ?? t}</span>)}
          {m.resist?.map((t) => <span className="chip" key={'r' + t}>сопротивление: {DMG[t] ?? t}</span>)}
          {m.vuln?.map((t) => <span className="chip bad" key={'v' + t}>уязвимость: {DMG[t] ?? t}</span>)}
        </div>
        {m.desc && <p className="hint">{m.desc}</p>}
        {children}
      </div>
    </div>
  )
}

// useBestiary — полный список существ (встроенные и свои), уже упорядоченный по CR и HP.
export function useBestiary(guard, tick = 0) {
  const [list, setList] = useState([])
  useEffect(() => { guard(async () => setList(await api.Bestiary())) }, [guard, tick])
  return list
}

// MonsterPicker — список существ по возрастанию силы, поиск, фильтр по типу, карточка и кнопка добавления.
export function MonsterPicker({ guard, onAdd, button = 'Добавить', tick = 0, extra, height = 300 }) {
  const list = useBestiary(guard, tick)
  const [q, setQ] = useState('')
  const [kind, setKind] = useState('')
  const [id, setId] = useState('')
  const [n, setN] = useState(1)
  const shown = useMemo(() => list.filter((m) => (!kind || m.kind === kind) && m.name.toLowerCase().includes(q.trim().toLowerCase())), [list, q, kind])
  const m = list.find((x) => x.id === id) ?? shown[0]
  let last = ''
  return (
    <div className="bestiary">
      <div className="row tight">
        <input placeholder="Поиск существа" aria-label="Поиск существа" value={q} onChange={(e) => setQ(e.target.value)} />
        <select aria-label="Тип существа" value={kind} onChange={(e) => setKind(e.target.value)}>
          <option value="">Все типы</option>{Object.entries(KINDS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}</select>
        <small>{shown.length} из {list.length} · по возрастанию силы</small>
      </div>
      <div className="mlist" style={{ maxHeight: height }} role="listbox" aria-label="Существа">
        {shown.map((x) => {
          const head = x.cr !== last ? <div className="mhead" key={'h' + x.cr + x.id}>CR {x.cr}</div> : null
          last = x.cr
          return [head,
            <button key={x.id} role="option" aria-selected={m?.id === x.id} className={'mrow' + (m?.id === x.id ? ' sel' : '')} onClick={() => setId(x.id)}>
              <Art className="avatar" html={artOf(x)} label="" /><span className="mn">{x.name}{x.custom ? ' ★' : ''}</span>
              <small>КД {x.ac} · HP {x.hp}</small></button>]
        })}
        {shown.length === 0 && <p className="hint">Никого не найдено.</p>}
      </div>
      {m && <MonsterCard m={m}>
        <div className="row tight">
          <input type="number" min="1" max="20" value={n} onChange={(e) => setN(parseInt(e.target.value, 10) || 1)} aria-label="Количество" style={{ width: 70 }} />
          <button className="primary" onClick={() => onAdd(m.id, n, m)}>{button}</button>
          {extra?.(m)}
        </div>
      </MonsterCard>}
    </div>
  )
}

const TYPES = Object.keys(DMG)
const blank = () => ({ name: '', cr: '1', kind: 'beast', ac: 12, hp: 20, attack: 3, speed: 30, abilities: [10, 10, 10, 10, 10, 10], weapons: [{ name: 'Удар', dice: '1d6+1', type: 'slashing', ranged: false }], multi: '', traits: '', resist: [], vuln: [], immune: [], desc: '', specials: [] })
const newSpecial = () => ({ name: '', mode: 'single', attack: true, dmg: '2d6', type: 'piercing', xDmg: '', xType: 'poison', save: '', dc: 12, half: false, cond: '', rounds: 0, repeat: false, recharge: false, once: false })
const ABN = ['Сила', 'Ловкость', 'Телосложение', 'Интеллект', 'Мудрость', 'Харизма']

const fromMonster = (m) => ({ id: m.id.startsWith('custom-') ? m.id : undefined, name: m.id.startsWith('custom-') ? m.name : m.name + ' (копия)', cr: m.cr, kind: m.kind, ac: m.ac, hp: m.hp, attack: m.attack, speed: m.speed,
  abilities: [...m.abilities], weapons: m.weapons.map((w) => ({ name: w.name, dice: w.dice, type: w.type, ranged: !!w.ranged })), multi: (m.multi ?? []).map((i) => i + 1).join(', '),
  traits: (m.traits ?? []).join(', '), resist: [...(m.resist ?? [])], vuln: [...(m.vuln ?? [])], immune: [...(m.immune ?? [])], desc: m.desc ?? '', specials: (m.specials ?? []).map((x) => ({ ...newSpecial(), ...x })) })

// MonsterEditor — создание и правка собственного существа.
export function MonsterEditor({ guard, initial, onSaved, onCancel }) {
  const [f, setF] = useState(initial ? fromMonster(initial) : blank())
  const set = (k, v) => setF((s) => ({ ...s, [k]: v }))
  const num = (k) => (e) => set(k, parseInt(e.target.value, 10) || 0)
  const setW = (i, patch) => set('weapons', f.weapons.map((w, j) => (j === i ? { ...w, ...patch } : w)))
  const toggle = (k, t) => set(k, f[k].includes(t) ? f[k].filter((x) => x !== t) : [...f[k], t])
  const save = () => guard(async () => {
    const multi = f.multi.split(/[\s,;]+/).filter(Boolean).map((x) => parseInt(x, 10) - 1)
    if (multi.some((x) => Number.isNaN(x))) throw new Error('Мультиатака: номера атак через запятую, например 1, 1, 2')
    const saved = await api.SaveMonster({
      id: f.id ?? '', name: f.name, cr: f.cr, kind: f.kind, ac: f.ac, hp: f.hp, attack: f.attack, speed: f.speed, abilities: f.abilities,
      weapons: f.weapons.map((w) => ({ ...w, ability: w.ranged ? 'dex' : 'str', finesse: false, bonus: 0 })), weapon: { name: '', dice: '', ability: '', type: '', finesse: false, ranged: false, bonus: 0 },
      multi, traits: f.traits.split(',').map((x) => x.trim()).filter(Boolean), resist: f.resist, vuln: f.vuln, immune: f.immune, desc: f.desc, specials: f.specials, custom: true,
    })
    onSaved(saved)
  })
  return (
    <div className="editor">
      <h3>{f.id ? 'Правка существа' : 'Новое существо'}</h3>
      <div className="grid2">
        <label>Название<input value={f.name} onChange={(e) => set('name', e.target.value)} placeholder="Например, Пещерный змей" /></label>
        <label>Тип<select value={f.kind} onChange={(e) => set('kind', e.target.value)}>{Object.entries(KINDS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}</select></label>
        <label>Уровень опасности (CR)<input value={f.cr} onChange={(e) => set('cr', e.target.value)} placeholder="0, 1/8, 1/4, 1/2, 1 … 30" /></label>
        <label>Скорость<input type="number" value={f.speed} onChange={num('speed')} /></label>
        <label>Класс доспеха<input type="number" value={f.ac} onChange={num('ac')} /></label>
        <label>Здоровье (HP)<input type="number" value={f.hp} onChange={num('hp')} /></label>
        <label>Бонус атаки<input type="number" value={f.attack} onChange={num('attack')} /></label>
      </div>
      <div className="abrow">{ABN.map((n, i) => <label key={n}>{n.slice(0, 3)}<input type="number" min="1" max="30" value={f.abilities[i]} onChange={(e) => set('abilities', f.abilities.map((v, j) => (j === i ? parseInt(e.target.value, 10) || 1 : v)))} /></label>)}</div>
      <h4>Атаки</h4>
      {f.weapons.map((w, i) => (
        <div className="row tight" key={i}>
          <small>{i + 1}.</small>
          <input value={w.name} placeholder="Название" aria-label="Название атаки" onChange={(e) => setW(i, { name: e.target.value })} />
          <input value={w.dice} placeholder="Урон 2d6+3" aria-label="Урон" style={{ width: 110 }} onChange={(e) => setW(i, { dice: e.target.value })} />
          <select aria-label="Тип урона" value={w.type} onChange={(e) => setW(i, { type: e.target.value })}>{TYPES.map((t) => <option key={t} value={t}>{DMG[t]}</option>)}</select>
          <label className="eq"><input type="checkbox" checked={w.ranged} onChange={(e) => setW(i, { ranged: e.target.checked })} /> дальняя</label>
          <button className="ghost small" disabled={f.weapons.length < 2} aria-label="Убрать атаку" onClick={() => set('weapons', f.weapons.filter((_, j) => j !== i))}>✕</button>
        </div>))}
      <div className="row tight">
        <button className="ghost" disabled={f.weapons.length >= 8} onClick={() => set('weapons', [...f.weapons, { name: '', dice: '1d6', type: 'slashing', ranged: false }])}>＋ Атака</button>
        <label className="eq">Мультиатака — номера атак за ход: <input value={f.multi} onChange={(e) => set('multi', e.target.value)} placeholder="1, 1, 2" style={{ width: 110 }} /></label>
      </div>
      {[['immune', 'Иммунитет', 'ok'], ['resist', 'Сопротивление', ''], ['vuln', 'Уязвимость', 'bad']].map(([k, n, cl]) => (
        <div key={k}><small>{n}</small>
          <div className="chips">{TYPES.map((t) => <button key={t} className={'chip ' + cl} aria-pressed={f[k].includes(t)} onClick={() => toggle(k, t)}>{DMG[t]}</button>)}</div></div>))}
      <h4>Особые способности <small>(яд, паралич, дыхание…)</small></h4>
      {f.specials.map((x, i) => {
        const setS = (patch) => set('specials', f.specials.map((y, j) => (j === i ? { ...y, ...patch } : y)))
        return <div className="spedit" key={i}>
          <div className="row tight">
            <input value={x.name} placeholder="Название (Ядовитый укус, Огненное дыхание)" aria-label="Название способности" onChange={(e) => setS({ name: e.target.value })} />
            <select aria-label="Режим" value={x.mode} onChange={(e) => setS({ mode: e.target.value })}><option value="single">по одной цели</option><option value="area">по всем врагам (область)</option></select>
            <button className="ghost small" aria-label="Убрать способность" onClick={() => set('specials', f.specials.filter((_, j) => j !== i))}>✕</button></div>
          <div className="row tight">
            {x.mode === 'single' && <label className="eq"><input type="checkbox" checked={x.attack} onChange={(e) => setS({ attack: e.target.checked })} /> бросок атаки</label>}
            <input value={x.dmg} placeholder="Урон 2d6" aria-label="Урон способности" style={{ width: 90 }} onChange={(e) => setS({ dmg: e.target.value })} />
            <select aria-label="Тип урона способности" value={x.type} onChange={(e) => setS({ type: e.target.value })}>{TYPES.map((t) => <option key={t} value={t}>{DMG[t]}</option>)}</select>
            <input value={x.xDmg} placeholder="Доп. урон (яд 2d6)" aria-label="Дополнительный урон" style={{ width: 130 }} onChange={(e) => setS({ xDmg: e.target.value })} />
            <select aria-label="Тип доп. урона" value={x.xType} onChange={(e) => setS({ xType: e.target.value })}>{TYPES.map((t) => <option key={t} value={t}>{DMG[t]}</option>)}</select></div>
          <div className="row tight">
            <label className="eq">Спасбросок <select value={x.save} onChange={(e) => setS({ save: e.target.value })}><option value="">нет</option>{[['str', 'Сила'], ['dex', 'Ловкость'], ['con', 'Телосложение'], ['int', 'Интеллект'], ['wis', 'Мудрость'], ['cha', 'Харизма']].map(([k, v]) => <option key={k} value={k}>{v}</option>)}</select></label>
            <label className="eq">СЛ <input type="number" min="0" max="40" value={x.dc} onChange={(e) => setS({ dc: +e.target.value || 0 })} style={{ width: 60 }} /></label>
            <label className="eq"><input type="checkbox" checked={x.half} onChange={(e) => setS({ half: e.target.checked })} /> половина при успехе</label>
            <label className="eq">Состояние <select value={x.cond} onChange={(e) => setS({ cond: e.target.value })}><option value="">нет</option>{['Ослеплён', 'Очарован', 'Оглохший', 'Напуган', 'Схвачен', 'Недееспособен', 'Парализован', 'Окаменел', 'Отравлен', 'Сбит с ног', 'Опутан', 'Ошеломлён'].map((c) => <option key={c}>{c}</option>)}</select></label>
            <label className="eq">ходов <input type="number" min="0" max="1000" value={x.rounds} onChange={(e) => setS({ rounds: +e.target.value || 0 })} style={{ width: 60 }} /></label></div>
          <div className="row tight">
            <label className="eq"><input type="checkbox" checked={x.repeat} onChange={(e) => setS({ repeat: e.target.checked })} /> повтор спасброска в конце хода цели</label>
            <label className="eq"><input type="checkbox" checked={x.recharge} onChange={(e) => setS({ recharge: e.target.checked })} /> перезарядка 5–6</label>
            <label className="eq"><input type="checkbox" checked={x.once} onChange={(e) => setS({ once: e.target.checked })} /> раз за бой</label></div>
        </div>
      })}
      <button className="ghost" disabled={f.specials.length >= 6} onClick={() => set('specials', [...f.specials, newSpecial()])}>＋ Способность</button>
      <label>Особенности (через запятую)<input value={f.traits} onChange={(e) => set('traits', e.target.value)} placeholder="Полёт, Тёмное зрение" /></label>
      <label>Описание<textarea value={f.desc} onChange={(e) => set('desc', e.target.value)} /></label>
      <div className="row"><button className="primary" disabled={!f.name.trim()} onClick={save}>Сохранить</button><button className="ghost" onClick={onCancel}>Отмена</button></div>
    </div>
  )
}

// MonstersTab — вкладка «Существа» в библиотеке.
export function MonstersTab({ guard, reload }) {
  const [tick, setTick] = useState(0)
  const [edit, setEdit] = useState(null)
  const [msg, setMsg] = useState('')
  const [mine, setMine] = useState([])
  useEffect(() => { guard(async () => setMine((await api.Bestiary()).filter((m) => m.custom))) }, [guard, tick])
  const done = () => { setEdit(null); setTick((t) => t + 1) }
  if (edit) return <MonsterEditor guard={guard} initial={edit === 'new' ? null : edit} onSaved={done} onCancel={() => setEdit(null)} />
  return (
    <div>
      <p className="hint">Существа идут по возрастанию уровня опасности (CR), внутри одного CR — по здоровью. Выберите и добавьте на стол: имена нумеруются сами. Данные взяты из SRD и Monster Manual по памяти — сверяйте с книгой.</p>
      <MonsterPicker guard={guard} tick={tick} button="Добавить на стол" onAdd={(id, n, m) => guard(async () => { await api.AddMonster(id, n); await reload(); setMsg(`Добавлено: ${m.name} ×${n}`) })}
        extra={(m) => <button className="ghost" title="Создать своё существо на основе этого" onClick={() => setEdit(m)}>Копия для правки</button>} />
      {msg && <p className="hint ok">{msg}</p>}
      <div className="row"><button className="primary" onClick={() => setEdit('new')}>＋ Создать своё существо</button>
        <button className="ghost danger" onClick={() => guard(async () => { const n = await api.ClearMonsters(); await reload(); setMsg(n ? `Убрано со стола существ: ${n}` : 'На столе нет существ') })}>Убрать всех существ со стола</button></div>
      {mine.length > 0 && <><h3>Мои существа</h3>
        {mine.map((m) => <div className="irow" key={m.id}><span><b>{m.name}</b> <small>CR {m.cr} · КД {m.ac} · HP {m.hp}</small></span>
          <button onClick={() => setEdit(m)}>Править</button>
          <button aria-label={`Удалить ${m.name}`} onClick={() => guard(async () => { await api.DeleteMonster(m.id); setTick((t) => t + 1) })}>✕</button></div>)}</>}
    </div>
  )
}
