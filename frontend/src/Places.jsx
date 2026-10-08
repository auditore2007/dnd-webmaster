import { useState } from 'react'
import { api } from './api'

// Места на карте: значки, карточка места (жители, задания, враги, добыча), правка и население.

export const KIND_ICON = {
  capital: '👑', city: '🏙️', town: '🏘️', village: '🏡', tavern: '🍺', temple: '⛪', shop: '🛒', smithy: '⚒️', castle: '🏰', port: '⚓',
  dungeon: '🗝️', cave: '🕳️', ruins: '🏛️', lair: '🐉', camp: '⛺', tower: '🗼', shrine: '⛩️', mine: '⛏️', room: '🚪', forest: '🌲',
  mountain: '⛰️', swamp: '🐸', lake: '💧', other: '📍',
}
export const KIND_NAME = {
  capital: 'Столица', city: 'Город', town: 'Городок', village: 'Деревня', tavern: 'Таверна', temple: 'Храм', shop: 'Лавка', smithy: 'Кузница',
  castle: 'Замок', port: 'Порт', dungeon: 'Подземелье', cave: 'Пещера', ruins: 'Руины', lair: 'Логово', camp: 'Лагерь', tower: 'Башня',
  shrine: 'Святилище', mine: 'Шахта', room: 'Комната', forest: 'Лес', mountain: 'Горы', swamp: 'Болото', lake: 'Озеро', other: 'Место',
}
const BIG = ['capital', 'city']
const ATTITUDE = { 'дружелюбен': 'ok', 'враждебен': 'bad' }

/** PlaceMarker – значок места на карте; подпись у крупных поселений и у выбранного. */
export function PlaceMarker({ p, selected, onSelect, labels, opens }) {
  const label = selected || (labels && BIG.includes(p.kind))
  const hint = opens ? (p.map ? ' – клик: открыть карту места, Shift+клик: сведения' : ' – клик: нарисовать карту места, Shift+клик: сведения') : ''
  return (
    <button className={'place k-' + p.kind + (selected ? ' sel' : '') + (p.foes?.length ? ' danger' : '') + (p.map ? ' has-map' : '')} style={{ left: p.x, top: p.y }}
      title={`${p.name} · ${KIND_NAME[p.kind] ?? p.kind}${hint}`} onPointerDown={(e) => e.stopPropagation()} onClick={(e) => { e.stopPropagation(); onSelect(p.id, e) }}>
      <span className="pi">{KIND_ICON[p.kind] ?? '📍'}</span>{label && <span className="pl">{p.name}</span>}
    </button>
  )
}

/** PlaceCard – всё о месте: правка, жители, задания (с переходом к цели), враги (на стол), добыча. */
export function PlaceCard({ map, place: p, cat, level, guard, onMap, onSelect, reload, notify, onOpen }) {
  const raceName = (id) => cat[0]?.races.find((r) => r.id === id)?.name ?? ''
  const [edit, setEdit] = useState(false)
  const [f, setF] = useState(p)
  const byId = (id) => map.places.find((x) => x.id === id)
  const save = () => guard(async () => { onMap(await api.UpdatePlace(map.id, { ...p, name: f.name, kind: f.kind, note: f.note, race: f.race })); setEdit(false) })
  const toTable = () => guard(async () => {
    for (const foe of p.foes) await api.AddMonster(foe.id, foe.count)
    await reload()
    notify(`На стол: ${p.foes.map((x) => `${x.name} ×${x.count}`).join(', ')}`)
  })
  return (
    <aside className="pcard" key={p.id}>
      <div className="pch"><span className="pi big">{KIND_ICON[p.kind]}</span>
        <div><b>{p.name}</b><small>{KIND_NAME[p.kind]}{p.race ? ` · ${raceName(p.race)}` : ''}</small></div>
        <button className="ghost small push" aria-label="Закрыть" onClick={() => onSelect('')}>✕</button></div>
      {edit ? <div className="editor">
        <label>Название<input value={f.name} onChange={(e) => setF({ ...f, name: e.target.value })} /></label>
        <label>Вид<select value={f.kind} onChange={(e) => setF({ ...f, kind: e.target.value })}>{Object.entries(KIND_NAME).map(([k, v]) => <option key={k} value={k}>{KIND_ICON[k]} {v}</option>)}</select></label>
        <label>Народ<select value={f.race} onChange={(e) => setF({ ...f, race: e.target.value })}><option value="">смешанный</option>{(cat[0]?.races ?? []).map((r) => <option key={r.id} value={r.id}>{r.name}</option>)}</select></label>
        <label>Заметка<textarea value={f.note} onChange={(e) => setF({ ...f, note: e.target.value })} /></label>
        <div className="row tight"><button className="primary small" onClick={save}>Сохранить</button><button className="ghost small" onClick={() => setEdit(false)}>Отмена</button></div>
      </div> : p.note && <p className="pnote">{p.note}</p>}

      {p.npcs?.length > 0 && <><h4>👥 Жители</h4>
        {p.npcs.map((n, i) => <details className="npc" key={i}>
          <summary><b>{n.name}</b> <small>{n.role} · {raceName(n.race)}</small> <span className={'chip ' + (ATTITUDE[n.attitude] ?? '')}>{n.attitude}</span></summary>
          <p><i>Черта:</i> {n.trait}. <i>Хочет:</i> {n.want}.</p>
          <p className="secret"><i>Секрет:</i> {n.secret}.</p>
        </details>)}</>}

      {p.quests?.length > 0 && <><h4>📜 Задания</h4>
        {p.quests.map((q, i) => <div className="quest" key={i}>
          <b>{q.title}</b>
          <p>{q.text}</p>
          <small>{q.giver && `Даёт: ${q.giver} · `}Награда: {q.reward}{q.xp ? ` · ${q.xp} опыта` : ''}</small>
          {q.target && byId(q.target) && <button className="ghost small" onClick={() => onSelect(q.target, true)}>📍 Показать на карте</button>}
        </div>)}</>}

      {p.foes?.length > 0 && <><h4>⚔️ Враги</h4>
        <div className="chips">{p.foes.map((x, i) => <span className="chip bad" key={i}>{x.name} ×{x.count} <small>CR {x.cr}</small></span>)}</div>
        <button className="primary small" onClick={toTable}>Выставить на стол</button></>}
      {p.loot && <p className="hint">💰 Добыча: {p.loot}</p>}

      <div className="row tight">
        {onOpen && <button className="primary small" onClick={onOpen}>🗺 {p.map ? 'Карта места' : 'Нарисовать карту места'}</button>}
        {!edit && <button className="ghost small" onClick={() => { setF(p); setEdit(true) }}>✎ Править</button>}
        <button className="ghost small" title="Новые жители, задания или враги" onClick={() => guard(async () => onMap(await api.PopulatePlace(map.id, p.id, level)))}>🎲 Населить заново</button>
        <button className="ghost small danger" onClick={() => guard(async () => { onMap(await api.RemovePlace(map.id, p.id)); onSelect('') })}>Удалить</button>
      </div>
    </aside>
  )
}

/** PlaceTool – что ставить при клике по карте в режиме «Места». */
export function PlaceTool({ kind, setKind, name, setName }) {
  return (
    <div className="row tight">
      <select aria-label="Вид места" value={kind} onChange={(e) => setKind(e.target.value)}>{Object.entries(KIND_NAME).map(([k, v]) => <option key={k} value={k}>{KIND_ICON[k]} {v}</option>)}</select>
      <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Название (необязательно), затем клик по карте" aria-label="Название места" />
    </div>
  )
}
