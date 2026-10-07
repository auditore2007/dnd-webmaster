import { ItemIcon } from './Icon.jsx'
import { useMemo, useState } from 'react'
import { api } from './api'
import { ARMOR, CATS, DMG, RARITY, fmt } from './labels.js'

export const itemInfo = (x, abil = []) => [x.price, x.rarity, x.weapon && `⚔ ${x.weapon.dice}${x.weapon.bonus ? ` ${fmt(x.weapon.bonus)}` : ''}`, x.acBonus && `КД ${fmt(x.acBonus)}`,
  x.armorBase > 0 && `доспех ${x.armorBase}${x.armorType ? ` (${ARMOR[x.armorType]})` : ''}`, x.abilityBonus && `${abil.find((a) => a.id === x.ability)?.name ?? x.ability} ${fmt(x.abilityBonus)}`,
  x.weight && `${x.weight} фнт`, x.desc].filter(Boolean).join(' · ')

const catOrder = Object.keys(CATS)

// filterItems – общий поиск по библиотеке предметов.
export function filterItems(lib, q, cat, rar) {
  const s = q.trim().toLowerCase()
  return lib.filter((x) => (!cat || x.cat === cat) && (!rar || x.rarity === rar) && (!s || x.name.toLowerCase().includes(s) || (x.desc ?? '').toLowerCase().includes(s)))
    .sort((a, b) => (catOrder.indexOf(a.cat) - catOrder.indexOf(b.cat)) || a.name.localeCompare(b.name, 'ru'))
}

export function ItemFilters({ q, setQ, cat, setCat, rar, setRar, lib }) {
  const cats = catOrder.filter((c) => lib.some((x) => x.cat === c))
  return (
    <div className="row tight">
      <input placeholder="Поиск предмета" aria-label="Поиск предмета" value={q} onChange={(e) => setQ(e.target.value)} />
      <select aria-label="Категория" value={cat} onChange={(e) => setCat(e.target.value)}><option value="">Все категории</option>
        {cats.map((c) => <option key={c} value={c}>{CATS[c]}</option>)}{lib.some((x) => !x.cat) && <option value="__none">Без категории</option>}</select>
      <select aria-label="Редкость" value={rar} onChange={(e) => setRar(e.target.value)}><option value="">Любая редкость</option>{RARITY.map((r) => <option key={r} value={r}>{r}</option>)}</select>
    </div>
  )
}

// ItemPicker – выдача предмета герою: поиск + категория + выбор.
export function ItemPicker({ lib, onGive }) {
  const [q, setQ] = useState('')
  const [cat, setCat] = useState('')
  const [rar, setRar] = useState('')
  const [qty, setQty] = useState(1)
  const [id, setId] = useState('')
  const list = useMemo(() => filterItems(lib, q, cat === '__none' ? '' : cat, rar).filter((x) => cat !== '__none' || !x.cat), [lib, q, cat, rar])
  const cur = list.find((x) => x.id === id) ?? list[0]
  return (
    <div className="picker">
      <ItemFilters q={q} setQ={setQ} cat={cat} setCat={setCat} rar={rar} setRar={setRar} lib={lib} />
      {list.length > 0 ? <div className="row tight">
        <select aria-label="Предмет из библиотеки" value={cur?.id} onChange={(e) => setId(e.target.value)}>
          {list.slice(0, 300).map((x) => <option key={x.id} value={x.id}>{x.name}{x.price ? ` – ${x.price}` : ''}</option>)}</select>
        <input type="number" min="1" value={qty} onChange={(e) => setQty(parseInt(e.target.value, 10) || 1)} aria-label="Количество" style={{ width: 70 }} />
        <button className="ghost" onClick={() => cur && onGive(cur.id, qty)}>Добавить</button>
      </div> : <p className="hint">Ничего не найдено.</p>}
      {cur && <p className="hint">{itemInfo(cur)}</p>}
      {list.length > 300 && <p className="hint">Показаны первые 300 из {list.length} – уточните поиск.</p>}
    </div>
  )
}

const blank = { name: '', cat: 'gear', price: '', rarity: '', weight: 0, desc: '', acBonus: 0, ability: '', abilityBonus: 0, wdice: '', wability: 'str', wtype: 'slashing', wbonus: 0, armorBase: 0, armorType: 'light', finesse: false, ranged: false }

// ItemsTab – вкладка «Предметы»: весь каталог, поиск, создание своих.
export function ItemsTab({ cat: catalog, lib, guard, reload }) {
  const abil = [...new Map(catalog.flatMap((r) => r.abilities).map((a) => [a.id, a])).values()]
  const [q, setQ] = useState('')
  const [cat, setCat] = useState('')
  const [rar, setRar] = useState('')
  const [limit, setLimit] = useState(60)
  const [f, setF] = useState(blank)
  const [msg, setMsg] = useState('')
  const set = (k) => (e) => setF({ ...f, [k]: e.target.type === 'number' ? parseFloat(e.target.value) || 0 : e.target.type === 'checkbox' ? e.target.checked : e.target.value })
  const list = useMemo(() => filterItems(lib, q, cat === '__none' ? '' : cat, rar).filter((x) => cat !== '__none' || !x.cat), [lib, q, cat, rar])
  const add = () => guard(async () => {
    await api.SaveLibraryItem({
      name: f.name, cat: f.cat, price: f.price, rarity: f.rarity, weight: f.weight, desc: f.desc, acBonus: f.acBonus,
      ability: f.abilityBonus ? f.ability || abil[0].id : '', abilityBonus: f.abilityBonus,
      weapon: f.wdice ? { name: f.name, dice: f.wdice, ability: f.wability, type: f.wtype, finesse: f.finesse, ranged: f.ranged, bonus: f.wbonus } : null,
      armorBase: f.armorBase, armorType: f.armorBase ? f.armorType : '',
    })
    setF(blank); await reload(); setMsg('Предмет создан и добавлен в библиотеку')
  })
  return (
    <div>
      <p className="hint">В библиотеке {lib.length} предметов: оружие, доспехи, снаряжение, инструменты, транспорт и магические предметы (SRD / PHB / DMG, по памяти – сверяйте цены и свойства с книгой). Выдавайте их героям на листе. Магические «+1/+2/+3» оружие, доспехи и щиты, кольца и плащи защиты действуют автоматически, пока надеты.</p>
      <ItemFilters q={q} setQ={(v) => { setQ(v); setLimit(60) }} cat={cat} setCat={(v) => { setCat(v); setLimit(60) }} rar={rar} setRar={(v) => { setRar(v); setLimit(60) }} lib={lib} />
      <small>Найдено: {list.length}</small>
      {list.slice(0, limit).map((x) => (
        <div className="irow" key={x.id}><span className="ln2"><ItemIcon item={x} size={22} /> <b>{x.name}</b> <small>{CATS[x.cat] ? `${CATS[x.cat]} · ` : ''}{itemInfo(x, abil)}</small></span>
          <button aria-label={`Удалить ${x.name}`} onClick={() => guard(async () => { await api.DeleteLibraryItem(x.id); await reload() })}>✕</button></div>))}
      {list.length > limit && <button className="ghost" onClick={() => setLimit(limit + 100)}>Показать ещё ({list.length - limit})</button>}
      <div className="row tight"><button className="ghost" onClick={() => guard(async () => { const n = await api.AddSRDItems(); await reload(); setMsg(n ? `Возвращено предметов каталога: ${n}` : 'Все предметы каталога уже в библиотеке') })}>Вернуть удалённые предметы каталога</button>{msg && <small>{msg}</small>}</div>

      <details className="cf" open>
        <summary>＋ Создать свой предмет</summary>
        <div className="editor">
          <div className="grid2">
            <label>Название<input value={f.name} onChange={set('name')} /></label>
            <label>Категория<select value={f.cat} onChange={set('cat')}>{catOrder.map((c) => <option key={c} value={c}>{CATS[c]}</option>)}</select></label>
            <label>Цена<input value={f.price} onChange={set('price')} placeholder="50 зм" /></label>
            <label>Редкость<select value={f.rarity} onChange={set('rarity')}><option value="">–</option>{RARITY.map((r) => <option key={r} value={r}>{r}</option>)}</select></label>
            <label>Вес<input type="number" step="0.1" min="0" value={f.weight || ''} onChange={set('weight')} /></label>
            <label>Бонус к КД<input type="number" value={f.acBonus || ''} onChange={set('acBonus')} /></label>
            <label>Бонус к характеристике<select value={f.ability} onChange={set('ability')}><option value="">–</option>{abil.map((a) => <option key={a.id} value={a.id}>{a.name}</option>)}</select></label>
            <label>Размер бонуса<input type="number" value={f.abilityBonus || ''} onChange={set('abilityBonus')} /></label>
          </div>
          <label>Описание<textarea value={f.desc} onChange={set('desc')} /></label>
          <h4>Оружие <small>(заполните «Урон», если предмет – оружие)</small></h4>
          <div className="row tight">
            <input placeholder="Урон, напр. 1d8" aria-label="Урон" value={f.wdice} onChange={set('wdice')} style={{ width: 130 }} />
            <select aria-label="Тип урона" value={f.wtype} onChange={set('wtype')}>{Object.entries(DMG).map(([k, v]) => <option key={k} value={k}>{v}</option>)}</select>
            <select aria-label="Характеристика оружия" value={f.wability} onChange={set('wability')}>{abil.map((a) => <option key={a.id} value={a.id}>{a.name}</option>)}</select>
            <input type="number" aria-label="Магический бонус" title="Магический бонус к атаке и урону" placeholder="+N" value={f.wbonus || ''} onChange={set('wbonus')} style={{ width: 70 }} />
            <label className="eq"><input type="checkbox" checked={f.finesse} onChange={set('finesse')} /> фехтовальное</label>
            <label className="eq"><input type="checkbox" checked={f.ranged} onChange={set('ranged')} /> дальнее</label>
          </div>
          <h4>Доспех <small>(0 – не доспех)</small></h4>
          <div className="row tight">
            <input type="number" min="0" max="30" aria-label="Базовый КД доспеха" placeholder="КД" value={f.armorBase || ''} onChange={set('armorBase')} style={{ width: 90 }} />
            <select aria-label="Тип доспеха" value={f.armorType} onChange={set('armorType')}>{Object.entries(ARMOR).map(([k, v]) => <option key={k} value={k}>{v}</option>)}</select>
          </div>
          <div className="row"><button className="primary" disabled={!f.name.trim()} onClick={add}>Создать предмет</button></div>
        </div>
      </details>
    </div>
  )
}
