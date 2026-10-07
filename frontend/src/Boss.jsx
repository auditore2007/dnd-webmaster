import { useEffect, useState } from 'react'
import { api } from './api'
import { keptD20 } from './rollText.js'

// Бой против боссов: настройки (автоход, мораль, логово), автоматический ход существ, легендарные действия.

const AUTO_DELAY = 900 // пауза перед ходом существа: видно, кто ходит, и кубики не сливаются
const isMon = (c) => c?.kind === 'monster'
const alive = (c) => c && !c.dead && c.hp > 0

/** legendaryActs – легендарные действия существа: из бестиария или по одной атаке каждым оружием (как на сервере). */
export function legendaryActs(c) {
  if (!(c?.stat?.legendary > 0)) return []
  if (c.stat.legActs?.length) return c.stat.legActs
  return (c.weapons ?? []).map((w, i) => ({ key: 'w' + i, name: 'Атака: ' + w.name, cost: 1, weapon: i }))
}

export const legendaryLeft = (c) => Math.max(0, (c?.stat?.legendary ?? 0) - (c?.counters?.leg ?? 0))

const attackEntry = (actor, target, r) => ({
  kind: 'attack', sides: 20, label: `${actor} → ${target}`, rolls: r.rolls ?? [], kept: r.kept || keptD20(r.rolls, ''),
  total: r.total, target: r.target, hit: r.hit, crit: r.crit, damage: r.damage, damageText: r.damageText,
})

/**
 * useAutoTurn – когда ходит существо и включён автоход, через паузу оно действует само (кубики катятся в лотке),
 * затем ход переходит дальше. Ошибка останавливает цепочку: новое состояние боя не приходит – нового хода нет.
 */
export function useAutoTurn({ enc, by, setEnc, reload, throwDice }) {
  const cur = enc && by[enc.order[enc.turn]]
  const standing = enc ? enc.order.filter((id) => alive(by[id])).length : 0
  const due = !!enc && enc.auto && !enc.won && isMon(cur) && alive(cur) && standing >= 2
  const multi = Math.max(1, cur?.stat?.multi?.length ?? 1)
  useEffect(() => {
    if (!due) return
    const t = setTimeout(() => {
      throwDice(async () => {
        const v = await api.MonsterTurn()
        setEnc(v.encounter)
        await reload()
        return (v.results ?? []).map((r, i) => attackEntry(v.actor, v.targets?.[i] ?? '', r))
      }, { sides: 20, count: multi })
    }, AUTO_DELAY)
    return () => clearTimeout(t)
  }, [due, enc?.seq, enc?.turn, enc?.round]) // eslint-disable-line react-hooks/exhaustive-deps
  return due
}

/** BattleOptions – переключатели боя. */
export function BattleOptions({ enc, by, setEnc, guard }) {
  const hasLair = enc.order.some((id) => by[id]?.stat?.lair?.length > 0)
  const set = (patch) => guard(async () => {
    const o = { auto: enc.auto, morale: enc.morale, lair: enc.lair, ...patch }
    setEnc(await api.SetEncounterOptions(o.auto, o.morale, o.lair))
  })
  return (
    <div className="bopts">
      <label className="eq" title="Существа сами выбирают цель и бьют; кнопку «Следующий ход» за них нажимать не нужно">
        <input type="checkbox" checked={!!enc.auto} onChange={(e) => set({ auto: e.target.checked })} /> 🎲 Враги ходят сами</label>
      <label className="eq" title="Тяжело раненные враги и те, чей вожак пал, проверяют мораль и могут сбежать или сдаться">
        <input type="checkbox" checked={!!enc.morale} onChange={(e) => set({ morale: e.target.checked })} /> 🏳 Мораль врагов</label>
      {hasLair && <label className="eq" title="В начале каждого раунда логово босса действует само">
        <input type="checkbox" checked={!!enc.lair} onChange={(e) => set({ lair: e.target.checked })} /> 🏰 Бой в логове</label>}
    </div>
  )
}

/** BossBadges – значки босса на карточке участника: легендарные действия, сопротивление, вторая фаза, бегство. */
export function BossBadges({ c }) {
  const s = c.stat
  if (!s) return null
  const res = (s.legRes ?? 0) - (c.counters?.legres ?? 0)
  return (
    <>
      {s.legendary > 0 && <span className="chip boss" title="Легендарные действия: тратятся в чужие ходы, восстанавливаются в начале хода босса">👑 лег. {legendaryLeft(c)}/{s.legendary}</span>}
      {s.legRes > 0 && <span className="chip boss" title="Легендарное сопротивление: проваленный спасбросок становится успешным">🛡 сопр. {res}/{s.legRes}</span>}
      {c.used?.phase && s.phase && <span className="chip phase" title="Вторая фаза босса">🔥 {s.phase.name}</span>}
    </>
  )
}

/** LegendaryPanel – легендарные действия боссов между чужими ходами. */
export function LegendaryPanel({ enc, by, setEnc, guard, reload, throwDice }) {
  const bosses = enc.order.map((id) => by[id]).filter((c) => c?.stat?.legendary > 0 && alive(c) && c.id !== enc.order[enc.turn])
  if (!bosses.length || enc.won) return null
  return (
    <div className="legp">
      <h4>👑 Легендарные действия <small>в чужой ход, восстанавливаются в начале хода босса{enc.auto ? ' · в автоходе босс тратит их сам' : ''}</small></h4>
      {bosses.map((b) => <BossActions key={b.id} b={b} enc={enc} by={by} setEnc={setEnc} guard={guard} reload={reload} throwDice={throwDice} />)}
    </div>
  )
}

function BossActions({ b, enc, by, setEnc, guard, reload, throwDice }) {
  const [tgt, setTgt] = useState('')
  const foes = enc.order.filter((id) => alive(by[id]) && !isMon(by[id]))
  const target = foes.includes(tgt) ? tgt : foes[0] ?? ''
  const left = legendaryLeft(b)
  const use = (a) => {
    const area = a.special?.mode === 'area'
    if (a.special) return guard(async () => { setEnc((await api.LegendaryAction(b.id, a.key, area ? '' : target)).encounter); await reload() })
    return throwDice(async () => {
      const v = await api.LegendaryAction(b.id, a.key, target)
      setEnc(v.encounter)
      await reload()
      return v.result ? [attackEntry(b.name, by[target]?.name ?? '', v.result)] : []
    }, { sides: 20, count: 1 })
  }
  return (
    <div className="row tight">
      <b>{b.name}</b><span className="pips" aria-label={`Осталось очков: ${left}`}>{Array.from({ length: b.stat.legendary }, (_, i) => <i key={i} className={i < left ? 'on' : ''} />)}</span>
      <select aria-label={`Цель для ${b.name}`} value={target} onChange={(e) => setTgt(e.target.value)}>{foes.map((id) => <option key={id} value={id}>{by[id].name}</option>)}</select>
      {legendaryActs(b).map((a) => <button key={a.key} className="ghost small" disabled={a.cost > left || (!target && a.special?.mode !== 'area')}
        title={a.special ? `${a.special.mode === 'area' ? 'по всем врагам' : 'по выбранной цели'}${a.special.save ? `, спасбросок ${a.special.save.toUpperCase()} СЛ ${a.special.dc}` : ''}` : 'атака по выбранной цели'}
        onClick={() => use(a)}>{a.special?.mode === 'area' ? '💥' : '⚔'} {a.name} <small>· {a.cost}</small></button>)}
    </div>
  )
}
