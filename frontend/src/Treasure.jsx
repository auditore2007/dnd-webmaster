import { useState } from 'react'
import { api } from './api'
import { Icon, ItemIcon } from './Icon.jsx'
import { COINS, CRS, COIN_ORDER } from './labels.js'

// Treasure — генератор добычи. В бою (enc) считает за всех существ боя; в библиотеке — за выбранные CR.
export default function Treasure({ chars, guard, reload, enc = false, notify }) {
  const [crs, setCrs] = useState([])
  const [pick, setPick] = useState('1')
  const [hoard, setHoard] = useState(false)
  const [t, setT] = useState(null)
  const [who, setWho] = useState('')
  const [split, setSplit] = useState(null)
  const [given, setGiven] = useState({})
  const heroes = chars.filter((c) => c.kind !== 'monster' && !c.dead)
  const target = heroes.some((h) => h.id === who) ? who : heroes[0]?.id ?? ''
  const sel = split ?? heroes.map((h) => h.id)
  const roll = () => guard(async () => { setT(enc ? await api.EncounterTreasure(hoard) : await api.RollTreasure(crs, hoard)); setGiven({}) })
  const give = (it, i) => guard(async () => { await api.GiveLoot(target, it, Math.max(1, it.qty)); setGiven((g) => ({ ...g, [i]: target })); await reload(); notify?.(`${heroes.find((h) => h.id === target)?.name} получает ${it.name}`) })
  const giveAll = () => guard(async () => {
    const g = {}
    for (const [i, it] of t.items.entries()) if (!given[i]) { await api.GiveLoot(target, it, Math.max(1, it.qty)); g[i] = target }
    setGiven((x) => ({ ...x, ...g })); await reload()
  })
  const splitCoins = () => guard(async () => { await api.SplitCoins(sel, t.coins); await reload(); setT((x) => ({ ...x, coins: {} })); notify?.('Монеты поделены поровну') })
  const coinText = t && COIN_ORDER.filter((k) => t.coins?.[k]).map((k) => `${t.coins[k]} ${COINS[k]}`).join(' · ')
  return (
    <section className="treas">
      <h3><Icon n="treasure" size={22} /> Сокровища{enc ? ' за бой' : ''}</h3>
      {!enc && <>
        <p className="hint">Добавьте уровни опасности (CR) существ, с которых падает добыча. «Клад» — сокровища логова, обычно за главного врага. Таблицы упрощены по Руководству мастера.</p>
        <div className="row tight">
          <select aria-label="CR" value={pick} onChange={(e) => setPick(e.target.value)}>{CRS.map((c) => <option key={c} value={c}>CR {c}</option>)}</select>
          <button className="ghost" onClick={() => setCrs((x) => [...x, pick])}>＋ Добавить</button>
          {crs.length > 0 && <button className="ghost" onClick={() => setCrs([])}>Очистить</button>}
        </div>
        <div className="chips">{crs.map((c, i) => <button key={i} className="chip" title="Убрать" onClick={() => setCrs((x) => x.filter((_, j) => j !== i))}>CR {c} ✕</button>)}</div>
      </>}
      <div className="row tight">
        <label className="inl"><input type="checkbox" checked={hoard} onChange={(e) => setHoard(e.target.checked)} /> Клад (логово) вместо личной добычи</label>
        <button className="primary" disabled={!enc && crs.length === 0} onClick={roll}><Icon n="dice" /> Бросить добычу</button>
      </div>
      {t && <div className="loot box-in">
        <p className="hint">{t.text}</p>
        {coinText ? <div className="row tight"><Icon n="coins" size={22} /><b>{coinText}</b></div> : <small>Монет нет.</small>}
        {coinText && heroes.length > 0 && <details className="cf"><summary>Разделить монеты поровну</summary>
          <div className="chips">{heroes.map((h) => <button key={h.id} className="chip" aria-pressed={sel.includes(h.id)} onClick={() => setSplit(sel.includes(h.id) ? sel.filter((x) => x !== h.id) : [...sel, h.id])}>{h.name}</button>)}</div>
          <button className="primary" disabled={!sel.length} onClick={splitCoins}>Поделить между {sel.length}</button></details>}
        {t.items?.length > 0 && <>
          <div className="row tight"><h4>Предметы и камни</h4>
            <label className="inl push">Кому <select value={target} onChange={(e) => setWho(e.target.value)}>{heroes.map((h) => <option key={h.id} value={h.id}>{h.name}</option>)}</select></label>
            <button className="ghost" disabled={!target} onClick={giveAll}>Отдать всё</button></div>
          <ul className="lootl">{t.items.map((it, i) => <li key={i} className={given[i] ? 'given' : ''}>
            <ItemIcon item={it} size={26} />
            <span className="ln"><b>{it.name}{it.qty > 1 ? ` ×${it.qty}` : ''}</b><small>{[it.rarity, it.price, it.desc].filter(Boolean).join(' · ')}</small></span>
            {given[i] ? <small>✓ {heroes.find((h) => h.id === given[i])?.name}</small> : <button className="ghost small" disabled={!target} onClick={() => give(it, i)}>Взять</button>}
          </li>)}</ul></>}
      </div>}
    </section>
  )
}
