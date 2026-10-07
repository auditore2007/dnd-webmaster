import { useEffect, useState } from 'react'
import { api } from './api'
import { MonsterPicker } from './Monsters.jsx'
import { Avatar } from './Portrait.jsx'
import { Icon, CondIcon } from './Icon.jsx'
import Treasure from './Treasure.jsx'
import Abilities from './Abilities.jsx'
import { HpMeter } from './Meter.jsx'
import { keptD20 } from './rollText.js'

const isMon = (c) => c?.kind === 'monster'

export default function Encounter({ chars, cat, mode, guard, reload, throwDice, notify, rev }) {
  const [enc, setEnc] = useState(null)
  const [ids, setIds] = useState([])
  const [target, setTarget] = useState('')
  const [wi, setWi] = useState(null)
  const [confirm, setConfirm] = useState(false)
  useEffect(() => { guard(async () => setEnc(await api.Encounter())) }, [guard, rev])

  const by = Object.fromEntries(chars.map((c) => [c.id, c]))
  const toggle = (id) => setIds((s) => (s.includes(id) ? s.filter((x) => x !== id) : [...s, id]))
  const run = (f) => guard(async () => { await f(); await reload() })
  const heroes = chars.filter((c) => !isMon(c))
  const mons = chars.filter(isMon)

  if (!enc) return (
    <section>
      <h2>Бой</h2>
      <p className="hint">Отметьте участников. Инициатива бросается автоматически. В бою можно будет добавить новых врагов, не прерывая его.</p>
      <div className="row tight">
        <button className="ghost" onClick={() => setIds(chars.filter((c) => !c.dead).map((c) => c.id))}>Все</button>
        <button className="ghost" onClick={() => setIds(heroes.filter((c) => !c.dead).map((c) => c.id))}>Все герои</button>
        <button className="ghost" onClick={() => setIds(mons.map((c) => c.id))}>Все существа</button>
        <button className="ghost" onClick={() => setIds([])}>Снять отметки</button>
      </div>
      {heroes.length > 0 && <h4>Герои</h4>}
      {heroes.map((c) => <Pick key={c.id} c={c} ids={ids} toggle={toggle} />)}
      {mons.length > 0 && <h4>Существа</h4>}
      {mons.map((c) => <Pick key={c.id} c={c} ids={ids} toggle={toggle} />)}
      <button className="primary" disabled={ids.length < 2} onClick={() => run(async () => setEnc(await api.StartEncounter(ids)))}>Начать бой ({ids.length})</button>
      <h3>Добавить врагов</h3>
      <MonsterPicker guard={guard} height={220} button="Добавить на стол"
        onAdd={(id, n) => guard(async () => { const added = await api.AddMonster(id, n); await reload(); setIds((s) => [...s, ...added.map((c) => c.id)]) })} />
    </section>
  )

  const cur = by[enc.order[enc.turn]]
  const rs = cat.find((r) => r.id === cur?.ruleset)
  const alive = (c) => c && !c.dead && c.hp > 0
  const mixed = enc.order.some((id) => isMon(by[id])) && enc.order.some((id) => by[id] && !isMon(by[id]))
  const sameSide = (id) => !mixed || isMon(by[id]) === isMon(cur)
  const foes = enc.order.filter((id) => id !== cur?.id && by[id] && !by[id].dead && !sameSide(id) && (alive(by[id]) || rs?.deathSaves))
  const foesAlive = foes.filter((id) => alive(by[id]))
  const allies = enc.order.filter((id) => by[id] && alive(by[id]) && (mixed ? isMon(by[id]) === isMon(cur) : true))
  const tgt = foes.includes(target) ? target : foes[0]
  const ws = cur?.derived?.weapons ?? []
  const wsel = wi === null ? (ws.length ? 0 : -1) : wi < ws.length ? wi : -1
  const wname = wsel >= 0 ? ws[wsel].name : 'Безоружный удар'
  const specials = (cur?.derived?.features ?? []).filter((f) => f.kind === 'special')
  const out = chars.filter((c) => !enc.order.includes(c.id) && !c.dead)
  const standing = enc.order.filter((id) => alive(by[id]))
  const entry = (res, name) => ({ kind: 'attack', sides: 20, mode, label: `${cur.name} → ${name} (${wname})`, rolls: res.rolls ?? [], kept: res.kept || keptD20(res.rolls, mode), total: res.total, target: res.target, hit: res.hit, crit: res.crit, damage: res.damage, damageText: res.damageText })
  // attack: бросок атаки с анимацией в лотке; all – все оставшиеся атаки хода по одной цели
  const attack = (all) => {
    const name = by[tgt]?.name ?? ''
    const count = (mode ? 2 : 1) * (all ? Math.max(1, enc.left) : 1)
    return throwDice(async () => {
      const v = all ? await api.AttackAll(tgt, wsel, mode) : await api.Attack(tgt, wsel, mode)
      setEnc(v.encounter)
      await reload()
      return (all ? v.results ?? [] : [v.result]).map((r) => entry(r, name))
    }, { sides: 20, count })
  }
  const end = (clear) => run(async () => { const n = await api.EndEncounter(clear); setEnc(null); setConfirm(false); notify(n ? `Бой окончен. Убрано существ: ${n}` : 'Бой окончен') })

  return (
    <section className="bt">
      <h2 className="round" key={enc.round}>Раунд <b>{enc.round}</b></h2>
      <div className="order">
        {enc.order.map((id, i) => { const c = by[id]; if (!c) return null
          return <div key={id} style={{ '--i': i }} className={'bi' + (i === enc.turn ? ' now' : '') + (!alive(c) ? ' down' : '') + (isMon(c) ? ' foe' : '')}>
            <div className="bh"><Avatar hero={c} size={38} /><b>{c.name}</b>
              <button className="ghost small push" title="Убрать из боя (убежал или не участвует)" aria-label={`Убрать из боя: ${c.name}`} onClick={() => run(async () => { const e = await api.RemoveFromEncounter(id); setEnc(e) ; if (!e) notify('В бою не осталось участников – бой завершён') })}>✕</button></div>
            <small>{isMon(c) ? 'враг' : 'герой'} · иниц. {enc.init[id]}{c.dead ? ' · погиб' : !alive(c) ? (isMon(c) ? ' · повержен' : c.stable ? ' · стабилен' : ` · умирает ✓${c.deathOk} ✗${c.deathFail}`) : ''}</small>
            <HpMeter hp={c.hp} max={c.derived?.maxHp ?? 0} temp={c.tempHp} label={`Здоровье: ${c.name}`} />
            {(c.conditions?.length > 0 || c.effects?.length > 0 || c.conc) && <div className="chips">
              {c.conditions?.map((x) => <span className="chip" key={x}><CondIcon name={x} /> {x}</span>)}
              {c.effects?.map((e, k) => <span className="chip fx" key={e.id ?? k} title={[e.cond && 'Состояние: ' + e.cond, e.save && `повторный спасбросок ${e.save.toUpperCase()} СЛ ${e.dc}`, e.conc && 'держится на концентрации'].filter(Boolean).join(' · ')}>
                <Icon n={e.conc ? 'conc' : 'effect'} size={13} /> {e.name}{e.rounds ? ` · ${e.rounds} х.` : ''}
                <button aria-label={`Снять эффект ${e.name} с ${c.name}`} onClick={() => run(async () => { await api.RemoveEffect(c.id, k) })}>✕</button></span>)}
              {c.conc && <span className="chip conc" title="Заклинатель держит концентрацию; урон требует спасброска Телосложения"><Icon n="conc" size={13} /> держит «{c.conc}»
                <button aria-label={`Прервать концентрацию ${c.name}`} onClick={() => run(async () => setEnc(await api.DropConcentration(c.id)))}>✕</button></span>}
            </div>}
            {alive(c) && <button className={'react' + (enc.react?.[id] ? ' spent' : '')} aria-pressed={!!enc.react?.[id]} title="Реакция раз в раунд (шанс на проход, щит, контрзаклинание). Сбрасывается в начале хода." onClick={() => run(async () => setEnc(await api.ToggleReaction(id)))}><Icon n="reaction" size={13} /> {enc.react?.[id] ? 'реакция потрачена' : 'реакция есть'}</button>}
          </div> })}
      </div>

      <details className="cf">
        <summary>➕ Подкрепление: добавить в бой</summary>
        <p className="hint">Новый участник бросает инициативу и встаёт в очередь, ход не сбивается.</p>
        <MonsterPicker guard={guard} height={220} button="В бой" onAdd={(id, n) => guard(async () => { setEnc(await api.SpawnInBattle(id, n)); await reload() })} />
        {out.length > 0 && <><h4>Уже на столе, но не в бою</h4>
          <div className="chips">{out.map((c) => <button key={c.id} className="chip" onClick={() => run(async () => setEnc(await api.AddToEncounter([c.id])))}>＋ {c.name}</button>)}</div></>}
      </details>

      {enc.won === 'heroes' && <div className="win box-in">🏆 Победа! Все враги повержены.
        <div className="row"><button className="primary" onClick={() => end(false)}>Завершить бой и убрать павших врагов</button></div></div>}
      {enc.won === 'monsters' && <div className="win lose box-in">☠️ Герои пали.<div className="row"><button className="ghost danger" onClick={() => end(false)}>Завершить бой</button></div></div>}
      {!enc.won && standing.length < 2 && <p className="win">Остался один: {by[standing[0]]?.name ?? 'никого'}</p>}
      {!enc.won && standing.length >= 2 && alive(cur) && <div className="f act">
        <h3>Ходит: {cur.name}{isMon(cur) ? ' (враг)' : ''}{mode === 'adv' ? ' · преимущество' : mode === 'dis' ? ' · помеха' : ''}</h3>
        <label>Цель<select value={tgt ?? ''} onChange={(e) => setTarget(e.target.value)}>{foes.map((id) => <option key={id} value={id}>{by[id].name}{alive(by[id]) ? '' : ' (повержен)'}</option>)}</select></label>
        <label>Оружие<select value={wsel} onChange={(e) => setWi(+e.target.value)}><option value={-1}>Безоружный удар (1d4)</option>
          {ws.map((w, i) => <option key={i} value={i}>{w.name} · {w.dice}</option>)}</select></label>
        <button className="primary" disabled={!tgt} onClick={() => attack(false)}>Атаковать{enc.left > 1 ? ` (${enc.left} ост.)` : ''}</button>
        {enc.left > 1 && <button className="primary" disabled={!tgt} title="Все оставшиеся атаки хода по выбранной цели" onClick={() => attack(true)}>{isMon(cur) && cur.stat?.multi?.length ? 'Мультиатака' : 'Все атаки'}</button>}
        {!isMon(cur) && <Abilities key={cur.id} h={cur} chars={chars} inFight call={(p) => run(async () => { await p; setEnc(await api.Encounter()) })} />}
        {cur.spells?.length > 0 && rs?.spells?.length > 0 && <SpellPanel key={cur.id} cur={cur} rs={rs} foes={foesAlive} allies={allies} by={by} tgt={tgt} mode={mode} run={run} setEnc={setEnc} />}
        {specials.map((f) => <button key={f.key} className="ghost spec" disabled={f.on || (f.mode === 'single' && !tgt)} title={f.desc + (f.mode === 'single' ? ' Бьёт выбранную цель.' : mixed ? ' Бьёт только врагов.' : '')}
          onClick={() => run(async () => setEnc(f.mode === 'single' ? await api.UseSpecialAt(f.key, tgt) : await api.UseSpecial(f.key)))}>{f.mode === 'single' ? '🎯' : '💥'} {f.name}{f.on ? ' (нет)' : ''}</button>)}
        <EffectForm key={cur.id} chars={enc.order.map((id) => by[id]).filter(Boolean)} rs={rs} cur={cur} tgt={tgt} run={run} />
      </div>}
      <details className="cf" open={!!enc.won}>
        <summary><Icon n="treasure" /> Добыча за бой</summary>
        <Treasure enc chars={chars} guard={guard} reload={reload} notify={notify} />
      </details>
      <div className="row">
        <button className="ghost" onClick={() => run(async () => { setWi(null); setEnc(await api.NextTurn()) })}>Следующий ход</button>
        {!confirm
          ? <button className="ghost danger" onClick={() => setConfirm(true)}>Закончить бой…</button>
          : <>
            <button className="ghost danger" onClick={() => end(false)} title="Погибшие и поверженные существа исчезнут со стола, живые останутся">Закончить, убрать павших</button>
            <button className="ghost danger" onClick={() => end(true)} title="Все существа из этого боя исчезнут со стола">Закончить, убрать всех врагов</button>
            <button className="ghost" onClick={() => setConfirm(false)}>Отмена</button></>}
      </div>
      <ul className="log">{enc.log.map((l, i) => <li key={enc.seq - i}>{l}</li>)}</ul>
    </section>
  )
}

function Pick({ c, ids, toggle }) {
  return <label className="pk"><input type="checkbox" checked={ids.includes(c.id)} onChange={() => toggle(c.id)} /> {c.name} <small>{c.kind === 'monster' ? `CR ${c.stat.cr}` : `ур. ${c.level}`} · {c.hp}/{c.derived?.maxHp} HP{c.dead ? ' · погиб' : ''}</small></label>
}

// SpellPanel: заклинание текущего участника. ⚡ – урон и лечение считаются автоматически; ◎ – концентрация; ✨ – накладывает эффект на выбранные цели.
function SpellPanel({ cur, rs, foes, allies, by, tgt, mode, run, setEnc }) {
  const [i, setI] = useState(0)
  const [slot, setSlot] = useState(0)
  const [picked, setPicked] = useState([])
  const [ally, setAlly] = useState('')
  const spells = cur.spells ?? []
  const sp = spells[i] ?? spells[0]
  if (!sp) return null
  const def = rs.spells.find((x) => x.id === sp.ref) ?? (sp.auto?.mode ? { mode: sp.auto.mode, area: sp.auto.area } : null)
  const slots = (cur.derived?.slots ?? []).filter((x) => x.level >= Math.max(1, sp.level))
  const use = sp.level === 0 ? 0 : (slot && slots.some((x) => x.level === slot) ? slot : slots[0]?.level ?? sp.level)
  const heal = def?.mode === 'heal'
  const fxOnly = !def?.mode && !!def?.targets
  const needTarget = (def?.mode && !heal) || (fxOnly && def.targets === 'foe')
  const pool = heal || (fxOnly && def.targets !== 'foe') ? allies : foes
  const multi = def?.area || fxOnly
  const healTargets = def?.area ? (picked.length ? picked : allies) : [allies.includes(ally) ? ally : cur.id]
  let targets
  if (heal) targets = healTargets
  else if (fxOnly) targets = def.targets === 'self' ? [cur.id] : picked.length ? picked : def.targets === 'ally' ? [cur.id] : [tgt]
  else targets = def?.area ? (picked.length ? picked : foes) : [tgt]
  const left = (lvl) => { const sv = (cur.derived?.slots ?? []).find((x) => x.level === lvl); return sv ? sv.max - (cur.slotsUsed?.[lvl - 1] ?? 0) : 0 }
  const mark = (s) => { const d = rs.spells.find((x) => x.id === s.ref); return (d?.mode || s.auto?.mode ? ' ⚡' : '') + (d?.targets ? ' ✨' : '') + (d?.conc ? ' ◎' : '') }
  return (
    <div className="spellp">
      <select aria-label="Заклинание" value={i} onChange={(e) => { setI(+e.target.value); setSlot(0); setPicked([]) }}>
        {spells.map((s, k) => <option key={k} value={k}>{s.level ? `${s.level} ур.` : 'заг.'} · {s.name}{mark(s)}</option>)}</select>
      {sp.level > 0 && slots.length > 0 && <select aria-label="Ячейка" value={use} onChange={(e) => setSlot(+e.target.value)}>
        {slots.map((x) => <option key={x.level} value={x.level}>ячейка {x.level} ур. ({left(x.level)})</option>)}</select>}
      {heal && !def.area && <select aria-label="Кого лечить" value={ally || cur.id} onChange={(e) => setAlly(e.target.value)}>{allies.map((id) => <option key={id} value={id}>{by[id].name}{id === cur.id ? ' (себя)' : ''}</option>)}</select>}
      {multi && def?.targets !== 'self' && pool.length > 0 && <div className="chips">{pool.map((id) => <button key={id} className="chip" aria-pressed={picked.includes(id)} onClick={() => setPicked((p) => (p.includes(id) ? p.filter((x) => x !== id) : [...p, id]))}>{by[id].name}</button>)}
        <small>{picked.length ? 'выбранные цели' : fxOnly ? (def.targets === 'ally' ? 'без выбора – на себя' : 'без выбора – на выбранную выше цель') : heal ? 'без выбора – вся команда' : 'без выбора – все враги'}</small></div>}
      {def?.conc && <small className="hint">◎ Концентрация{cur.conc ? `: «${cur.conc}» будет прервана` : ''}. Урон заставит сделать спасбросок Телосложения.</small>}
      <button className="primary" disabled={needTarget && !targets[0]} onClick={() => run(async () => { setEnc(await api.CastSpell(i, use, targets.filter(Boolean), mode)); setPicked([]) })}>Сотворить{def?.mode ? ' ⚡' : fxOnly ? ' ✨' : ''}</button>
    </div>
  )
}

// EffectForm: мастер вешает на участника свой эффект (горение, проклятие, чужое заклинание) на N ходов.
function EffectForm({ chars, rs, cur, tgt, run }) {
  const [who, setWho] = useState('')
  const [name, setName] = useState('')
  const [rounds, setRounds] = useState(10)
  const [cond, setCond] = useState('')
  const id = chars.some((c) => c.id === who) ? who : tgt ?? cur.id
  return (
    <details className="cf">
      <summary><Icon n="effect" /> Наложить эффект</summary>
      <div className="row tight">
        <select aria-label="На кого" value={id} onChange={(e) => setWho(e.target.value)}>{chars.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}</select>
        <input aria-label="Название эффекта" placeholder="Название (горит, проклят…)" value={name} onChange={(e) => setName(e.target.value)} />
        <label className="inl">ходов <input type="number" min="0" max="1000" value={rounds} onChange={(e) => setRounds(Math.max(0, Math.min(1000, parseInt(e.target.value, 10) || 0)))} style={{ width: 64 }} title="0 – пока не снимут вручную" /></label>
        <select aria-label="Состояние" value={cond} onChange={(e) => setCond(e.target.value)}><option value="">без состояния</option>{(rs?.conditions ?? []).map((c) => <option key={c}>{c}</option>)}</select>
        <button className="ghost" disabled={!name.trim()} onClick={() => run(async () => { await api.AddEffect(id, name, rounds, cond); setName('') })}>Наложить</button>
      </div>
    </details>
  )
}
