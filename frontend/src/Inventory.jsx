import { useMemo, useState } from 'react'
import { api } from './api'
import { Icon, ItemIcon } from './Icon.jsx'
import { ItemPicker } from './Items.jsx'
import { CATS, fmt } from './labels.js'
import { COINS, COIN_NAMES, COIN_ORDER, priceGP } from './labels.js'

const w1 = (n) => +(n ?? 0).toFixed(1)
const COIN_GP = { pp: 10, gp: 1, ep: 0.5, sp: 0.1, cp: 0.01 }

// Inventory — инвентарь героя: нагрузка, кошелёк, предметы по категориям, передача другому герою.
export default function Inventory({ h, d, rs, lib, chars, save, act, guard, reload }) {
  const inv = h.inventory ?? []
  const [q, setQ] = useState('')
  const [sort, setSort] = useState('cat')
  const [xfer, setXfer] = useState(null) // {i, to, qty}
  const [delta, setDelta] = useState(10)
  const [nw, setNw] = useState({ name: '', qty: 1, weight: 0, cat: 'gear', price: '' })
  const setInv = (i, patch) => save({ inventory: inv.map((x, j) => (j === i ? { ...x, ...patch } : x)) })
  const drop = (i) => save({ inventory: inv.filter((_, j) => j !== i) })
  const mates = chars.filter((c) => c.kind !== 'monster' && c.id !== h.id && !c.dead)

  const str = d.carry > 0 ? d.carry / 15 : 0
  const load = d.load ?? 0
  const level = !d.carry ? 0 : load > d.carry ? 3 : load > 10 * str ? 2 : load > 5 * str ? 1 : 0
  const LEVELS = ['Налегке', 'Нагружен (−10 фт скорости)', 'Сильно нагружен (помеха, −20 фт)', 'Перегруз: не сдвинуться']
  const pct = d.carry ? Math.min(100, (load / d.carry) * 100) : 0

  const purse = h.purse ?? {}
  const coinGP = COIN_ORDER.reduce((s, k) => s + (purse[k] ?? 0) * COIN_GP[k], 0)
  const coinCount = COIN_ORDER.reduce((s, k) => s + (purse[k] ?? 0), 0)
  const itemsGP = inv.reduce((s, x) => s + priceGP(x.price) * Math.max(1, x.qty), 0)

  const rows = useMemo(() => {
    const needle = q.trim().toLowerCase()
    const list = inv.map((x, i) => ({ x, i })).filter(({ x }) => !needle || x.name.toLowerCase().includes(needle))
    if (sort === 'weight') return [{ cat: '', list: list.sort((a, b) => b.x.weight * Math.max(1, b.x.qty) - a.x.weight * Math.max(1, a.x.qty)) }]
    if (sort === 'name') return [{ cat: '', list: list.sort((a, b) => a.x.name.localeCompare(b.x.name, 'ru')) }]
    const g = new Map()
    for (const r of list) {
      const k = r.x.equipped ? '__eq' : r.x.cat || '__other'
      g.set(k, [...(g.get(k) ?? []), r])
    }
    const order = ['__eq', ...Object.keys(CATS), '__other']
    return order.filter((k) => g.has(k)).map((k) => ({ cat: k, list: g.get(k) }))
  }, [inv, q, sort])
  const catName = (k) => (k === '__eq' ? 'Надето' : k === '__other' ? 'Прочее' : k ? CATS[k] : '')
  const gw = (list) => w1(list.reduce((s, { x }) => s + x.weight * Math.max(1, x.qty), 0))

  const doTransfer = () => guard(async () => {
    await api.TransferItem(h.id, xfer.to, xfer.i, xfer.qty); setXfer(null); await reload()
  })
  const addNew = () => {
    const it = { name: nw.name.trim(), qty: Math.max(1, nw.qty), weight: Math.max(0, nw.weight), cat: nw.cat, price: nw.price.trim(), desc: '' }
    save({ inventory: [...inv, it] }); setNw({ ...nw, name: '', qty: 1, weight: 0, price: '' })
  }

  return (
    <div className="inv">
      <h3><Icon n="pack" size={22} /> Инвентарь</h3>
      <div className="load">
        <div className="loadbar" role="meter" aria-valuemin={0} aria-valuemax={d.carry || 0} aria-valuenow={w1(load)} aria-label="Нагрузка">
          <i className={'l' + level} style={{ width: pct + '%' }} />
          {str > 0 && <><b style={{ left: (5 * str / d.carry) * 100 + '%' }} /><b style={{ left: (10 * str / d.carry) * 100 + '%' }} /></>}
        </div>
        <div className="row tight"><Icon n="weight" size={16} /> <b>{w1(load)}{d.carry > 0 ? ` / ${d.carry}` : ''} фнт</b>
          <small className={level > 1 ? 'bad' : ''}>{d.carry > 0 ? LEVELS[level] : ''}</small>
          <small className="push">≈ {w1(itemsGP + coinGP)} зм всего</small></div>
      </div>

      <div className="purse">
        <div className="row tight"><Icon n="purse" size={20} /><b>Кошелёк</b>
          <small>{coinCount} монет · {w1(coinCount / 50)} фнт · ≈ {w1(coinGP)} зм</small>
          <label className="inl push">шаг <input type="number" min="1" value={delta} onChange={(e) => setDelta(Math.max(1, +e.target.value || 1))} style={{ width: 70 }} aria-label="Шаг изменения монет" /></label></div>
        <div className="coins">{COIN_ORDER.map((k) => (
          <div className={'coin ' + k} key={k} title={COIN_NAMES[k]}>
            <span>{COINS[k]}</span>
            <input type="number" min="0" aria-label={COIN_NAMES[k] + ' монеты'} defaultValue={purse[k] ?? 0} key={k + (purse[k] ?? 0)}
              onBlur={(e) => { const n = parseInt(e.target.value, 10); if (!Number.isNaN(n) && n !== (purse[k] ?? 0)) act(api.SetCoins(h.id, k, n)); else e.target.value = purse[k] ?? 0 }} />
            <span className="cb"><button aria-label={`Потратить ${COINS[k]}`} disabled={(purse[k] ?? 0) < 1} onClick={() => act(api.AdjustCoins(h.id, k, -Math.min(delta, purse[k] ?? 0)))}>−</button>
              <button aria-label={`Добавить ${COINS[k]}`} onClick={() => act(api.AdjustCoins(h.id, k, delta))}>+</button></span>
          </div>))}</div>
        <table className="coinref" aria-label="Расшифровка монет">
          <thead><tr><th>Сокр.</th><th>Монета</th><th>Стоимость</th></tr></thead>
          <tbody>{[['pp', '10 ЗМ'], ['gp', '1 ЗМ — основная'], ['ep', '½ ЗМ (5 СМ)'], ['sp', '1/10 ЗМ (10 ММ)'], ['cp', '1/100 ЗМ']].map(([k, v]) =>
            <tr key={k}><td><b>{COINS[k]}</b></td><td>{COIN_NAMES[k]}</td><td>{v}</td></tr>)}</tbody>
        </table>
        <p className="hint">50 монет любого вида весят 1 фунт.</p>
      </div>

      <div className="row tight">
        <input type="search" placeholder="Поиск по инвентарю" aria-label="Поиск по инвентарю" value={q} onChange={(e) => setQ(e.target.value)} />
        <select aria-label="Сортировка" value={sort} onChange={(e) => setSort(e.target.value)}>
          <option value="cat">По категориям</option><option value="weight">Самые тяжёлые сверху</option><option value="name">По алфавиту</option></select>
      </div>
      {inv.length === 0 && <p className="hint">Пока пусто. Добавьте предмет из библиотеки или создайте свой ниже.</p>}
      {rows.map((g) => (
        <div className="igroup" key={g.cat || 'all'}>
          {g.cat && <h4>{catName(g.cat)} <small>{gw(g.list)} фнт</small></h4>}
          {g.list.map(({ x, i }) => {
            const info = [x.desc, x.weapon && `⚔ ${x.weapon.dice}`, x.acBonus && `${rs.acLabel} ${fmt(x.acBonus)}`, x.abilityBonus && `${rs.abilities.find((a) => a.id === x.ability)?.name} ${fmt(x.abilityBonus)}`, x.armorBase > 0 && `доспех ${x.armorBase}${x.armorType === 'heavy' ? '' : x.armorType === 'medium' ? ' + Лов (макс. 2)' : ' + Лов'}`, x.price].filter(Boolean).join(' · ')
            const tot = x.weight * Math.max(1, x.qty)
            return (
              <div className={'irow it' + (x.equipped ? ' on' : '')} key={i}>
                <ItemIcon item={x} size={26} />
                <span className="ln"><b>{x.name}</b>{info && <small>{info}</small>}</span>
                <span className="wt" title="Вес одной штуки / всего">{x.weight ? <>{w1(x.weight)}{x.qty > 1 ? <> · <b>{w1(tot)}</b></> : null} <small>фнт</small></> : <small>—</small>}</span>
                <span className="qty"><button aria-label={`Меньше: ${x.name}`} onClick={() => (x.qty > 1 ? setInv(i, { qty: x.qty - 1 }) : drop(i))}>−</button>{x.qty}<button aria-label={`Больше: ${x.name}`} onClick={() => setInv(i, { qty: x.qty + 1 })}>+</button></span>
                <label className="eq"><input type="checkbox" checked={!!x.equipped} onChange={(e) => setInv(i, { equipped: e.target.checked })} /> надето</label>
                {mates.length > 0 && <button className="ghost small" title="Передать другому герою" aria-label={`Передать: ${x.name}`} onClick={() => setXfer(xfer?.i === i ? null : { i, to: mates[0].id, qty: 1 })}>⇄</button>}
                <button aria-label={`Убрать ${x.name}`} onClick={() => drop(i)}>✕</button>
                {xfer?.i === i && <div className="xfer row tight">Передать
                  <input type="number" min="1" max={Math.max(1, x.qty)} value={xfer.qty} aria-label="Сколько передать" onChange={(e) => setXfer({ ...xfer, qty: Math.min(Math.max(1, x.qty), Math.max(1, +e.target.value || 1)) })} style={{ width: 64 }} />
                  <select aria-label="Кому" value={xfer.to} onChange={(e) => setXfer({ ...xfer, to: e.target.value })}>{mates.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}</select>
                  <button className="primary small" onClick={doTransfer}>Передать</button></div>}
              </div>)
          })}
        </div>))}

      <details className="cf"><summary>＋ Добавить из каталога</summary>
        {lib.length > 0 ? <ItemPicker lib={lib} onGive={(id, qty) => act(api.GiveItem(h.id, id, qty))} /> : <p className="hint">Каталог пуст — загрузите предметы в «Библиотеке».</p>}</details>
      <details className="cf"><summary>✎ Быстро записать свой предмет</summary>
        <div className="row tight">
          <input placeholder="Название" aria-label="Название предмета" value={nw.name} onChange={(e) => setNw({ ...nw, name: e.target.value })} />
          <select aria-label="Категория" value={nw.cat} onChange={(e) => setNw({ ...nw, cat: e.target.value })}>{Object.entries(CATS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}</select>
          <label className="inl">шт <input type="number" min="1" value={nw.qty} onChange={(e) => setNw({ ...nw, qty: +e.target.value || 1 })} style={{ width: 60 }} /></label>
          <label className="inl">вес, фнт <input type="number" min="0" step="0.1" value={nw.weight} onChange={(e) => setNw({ ...nw, weight: Math.max(0, +e.target.value || 0) })} style={{ width: 70 }} /></label>
          <input placeholder="Цена (5 зм)" aria-label="Цена" value={nw.price} onChange={(e) => setNw({ ...nw, price: e.target.value })} style={{ width: 100 }} />
          <button className="ghost" disabled={!nw.name.trim()} onClick={addNew}>Добавить</button>
        </div></details>
    </div>
  )
}
