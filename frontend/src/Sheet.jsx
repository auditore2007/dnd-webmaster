import { useEffect, useState } from 'react'
import { api } from './api'
import { Avatar, pickImage } from './Portrait.jsx'
import Inventory from './Inventory.jsx'
import { Icon, CondIcon } from './Icon.jsx'
import Abilities from './Abilities.jsx'
import { SpellFilters, filterSpells, lvlName } from './Spells.jsx'
import { HpMeter } from './Meter.jsx'

const fmt = (n) => ((n ?? 0) > 0 ? '+' : '') + (n ?? 0)
const pips = (n) => '●'.repeat(n) + '○'.repeat(Math.max(0, 3 - n))

export default function Sheet({ hero: h, rs, lib, chars, mode, dc, roll, guard, put, remove, reload }) {
  const d = h.derived ?? {}
  const mon = h.kind === 'monster'
  const race = rs.races.find((r) => r.id === h.race)
  const sub = race?.subs?.find((s) => s.id === h.subrace)
  const cls = rs.classes?.find((c) => c.id === h.class)
  const [armed, setArmed] = useState(false)
  const [w, setW] = useState({ name: '', dice: '1d8', ability: rs.abilities[0].id, type: rs.damageTypes?.[0]?.id ?? '' })
  const [sp, setSp] = useState({ name: '', level: 1, cost: 1, note: '' })
  const [xp, setXp] = useState(50)
  const save = (patch) => guard(async () => put(await api.UpdateCharacter({ ...h, ...patch })))
  const act = (p) => guard(async () => put(await p))
  const hp = (n) => act(n < 0 ? api.ApplyDamage(h.id, -n) : api.ApplyHeal(h.id, n))
  // числа сохраняются при потере фокуса; неверный ввод откатывается
  const num = (cur, f) => (e) => {
    const n = parseInt(e.target.value, 10)
    if (Number.isNaN(n)) e.target.value = cur
    else if (n !== cur) f(n)
  }
  const hasSaves = d.saves && Object.keys(d.saves).length > 0
  const inv = h.inventory ?? []
  const setInv = (i, patch) => save({ inventory: inv.map((x, j) => (j === i ? { ...x, ...patch } : x)) })
  const over = d.carry > 0 && d.load > d.carry
  const features = d.features ?? []

  return (
    <article className="sheet">
      <div className="head">
        <div className="portrait"><Avatar hero={h} size={96} />
          {!mon && <div className="pbtn">
            <label className="ghost small" title="Загрузить свою картинку">Фото<input type="file" accept="image/*" hidden onChange={(e) => { const f = e.target.files[0]; e.target.value = ''; if (f) guard(async () => save({ portrait: await pickImage(f) })) }} /></label>
            {h.portrait && <button className="ghost small" onClick={() => save({ portrait: '' })}>Сбросить</button>}</div>}
        </div>
        <div className="hn"><input className="name" key={h.name} defaultValue={h.name} aria-label="Имя героя" onBlur={(e) => e.target.value.trim() && e.target.value !== h.name && save({ name: e.target.value })} />
      <p className="sub">{mon ? `Существо · CR ${h.stat.cr}` : `${race?.name ?? ''}${sub ? ` · ${sub.name}` : ''}${cls ? ` · ${cls.name}` : ''}`}</p>
          {cls?.subclasses?.length > 0 && <label className="subcls">Подкласс
            <select value={h.subclass ?? ''} onChange={(e) => save({ subclass: e.target.value })}><option value="">– не выбран –</option>
              {cls.subclasses.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}</select></label>}
        </div>
      </div>

      <div className="vitals">
        {!mon && <div className="lvl"><span>Уровень</span>
          <button aria-label="Понизить уровень" disabled={h.level <= 1} onClick={() => save({ level: h.level - 1 })}>−</button>
          <b key={h.level} className="pop">{h.level}</b>
          <button aria-label="Повысить уровень" disabled={h.level >= 20} onClick={() => save({ level: h.level + 1 })}>+</button></div>}
        <div className="hp">
          <HpMeter hp={h.hp} max={d.maxHp} temp={h.tempHp} />
          {[-5, -1, 1, 5].map((n) => <button key={n} onClick={() => hp(n)}>{fmt(n)}</button>)}
        </div>
        {h.dead && <p className="bad"><b>Погиб</b> <button className="ghost" onClick={() => act(api.Revive(h.id))}>Воскресить</button></p>}
        {!h.dead && h.hp <= 0 && (rs.deathSaves && !mon
          ? <div className="death">{h.stable ? <b>Стабилизирован</b> : <><span>Успехи {pips(h.deathOk)}</span><span className="bad">Провалы {pips(h.deathFail)}</span>
            <button className="ghost" onClick={() => act(api.DeathSave(h.id))}>Спасбросок от смерти</button></>}</div>
          : <p className="bad"><b>Без сознания</b></p>)}
        <dl className="stats">
          <div><dt>{rs.acLabel}</dt><dd>{d.ac}</dd></div>
          {d.prof > 0 && <div><dt>Мастерство</dt><dd>{fmt(d.prof)}</dd></div>}
          {mon && <div><dt>Атака</dt><dd>{fmt(d.atkBonus)}</dd></div>}
          {d.attacks > 1 && <div><dt>Атак за ход</dt><dd>{d.attacks}</dd></div>}
          {d.spellDc > 0 && <div><dt>СЛ закл.</dt><dd>{d.spellDc}</dd></div>}
          {d.spellDc > 0 && <div><dt>Атака закл.</dt><dd>{fmt(d.spellAtk)}</dd></div>}
          <div><dt>Инициатива</dt><dd>{fmt(d.initiative)}</dd></div>
          {d.speed > 0 && <div><dt>Скорость</dt><dd>{d.speed}</dd></div>}
          <div><dt>Бонус к {rs.acLabel}</dt><dd><input type="number" key={h.id + h.acBonus} defaultValue={h.acBonus} onBlur={num(h.acBonus, (n) => save({ acBonus: n }))} aria-label="Бонус к защите" /></dd></div>
          <div><dt>Врем. HP</dt><dd><input type="number" min="0" key={'t' + h.tempHp} defaultValue={h.tempHp} onBlur={num(h.tempHp, (n) => save({ tempHp: Math.max(0, n) }))} aria-label="Временные HP" /></dd></div>
        </dl>
        {d.effects?.length > 0 && <div className="chips">{d.effects.map((x) => <span key={x} className="chip fx">{x}</span>)}</div>}
        {!mon && <div className="mt"><span>Опыт</span>
          <div className="meter xp" style={{ '--p': Math.min(100, Math.round((h.xp / Math.max(1, d.nextXp)) * 100)) + '%' }}><b>{h.xp} / {d.nextXp > 1e9 ? '–' : d.nextXp}</b></div>
          <input type="number" min="1" value={xp} onChange={(e) => setXp(parseInt(e.target.value, 10) || 0)} aria-label="Сколько опыта" style={{ width: 80 }} />
          <button onClick={() => act(api.AddXP(h.id, xp))}>+XP</button></div>}
        {h.points > 0 && <p className="hint">Очки характеристик: <b>{h.points}</b> – тратятся кнопками «+» ниже.</p>}
        {!mon && <div className="row tight">{rs.rests.map((r) => <button key={r.id} className="ghost" onClick={() => act(api.Rest(h.id, r.id))}>{r.name}</button>)}</div>}
      </div>

      <div className="abilities">
        {rs.abilities.map((a) => (
          <div className="ab" key={a.id}>
            <small>{a.name}</small>
            <b key={d.mods?.[a.id]} className="pop">{fmt(d.mods?.[a.id])}</b>
            <input type="number" min="0" max="30" aria-label={a.name} key={a.id + h.abilities[a.id]} defaultValue={h.abilities[a.id]}
              onBlur={num(h.abilities[a.id], (n) => save({ abilities: { ...h.abilities, [a.id]: Math.max(0, Math.min(30, n)) } }))} />
            <div>
              <button onClick={() => roll(() => api.Check(h.id, a.id, 'check', mode, dc))}>Проверка</button>
              {hasSaves && <button onClick={() => roll(() => api.Check(h.id, a.id, 'save', mode, dc))}>Спас {fmt(d.saves[a.id])}</button>}
              {h.points > 0 && <button aria-label={`Повысить: ${a.name}`} onClick={() => save({ abilities: { ...h.abilities, [a.id]: h.abilities[a.id] + 1 } })}>+</button>}
            </div>
          </div>
        ))}
      </div>

      {!mon && (race || cls) && <Lore race={race} sub={sub} cls={cls} />}

      {(d.traits?.length > 0 || features.length > 0) && <><h3>Особенности</h3>
        <div className="chips">
          {d.traits?.filter((t) => !features.some((f) => f.name === t)).map((t) => <span key={t} className="chip">{t}</span>)}
          {features.filter((f) => f.kind !== 'ability').map((f) => f.kind === 'toggle'
            ? <button key={f.key + f.name} className="chip" aria-pressed={f.on} title={f.desc} onClick={() => act(api.UseFeature(h.id, f.key))}>{f.name}{f.on ? ' · активно' : ''}</button>
            : f.kind === 'once'
              ? <button key={f.key + f.name} className="chip" disabled={f.on} title={f.desc} onClick={() => act(api.UseFeature(h.id, f.key))}>{f.name}{f.on ? ' · использовано' : ' – применить'}</button>
              : <span key={f.key + f.name} className="chip" title={f.desc}>{f.name}{f.on ? ' · использовано' : ''}</span>)}
        </div></>}
      {!mon && <Abilities h={h} chars={chars} call={act} />}
      {mon && (h.stat.vuln?.length > 0 || h.stat.resist?.length > 0 || h.stat.immune?.length > 0) && <div className="chips dmgx">
        {[['immune', 'Иммунитет', 'ok'], ['resist', 'Сопротивление', ''], ['vuln', 'Уязвимость', 'bad']].flatMap(([k, n, cl]) => (h.stat[k] ?? []).map((t) =>
          <span key={k + t} className={'chip ' + cl}>{n}: {rs.damageTypes?.find((x) => x.id === t)?.name ?? t}</span>))}</div>}
      {cls && <ClassFeatures cls={cls} h={h} />}
      {rs.skills?.length > 0 && !mon && <Skills rs={rs} h={h} d={d} save={save} roll={roll} mode={mode} dc={dc} />}

      <h3>Состояния</h3>
      <div className="chips">{rs.conditions.map((c) => {
        const on = h.conditions?.includes(c)
        return <button key={c} className="chip" aria-pressed={!!on}
          onClick={() => save({ conditions: on ? h.conditions.filter((x) => x !== c) : [...(h.conditions ?? []), c] })}><CondIcon name={c} /> {c}</button>
      })}</div>
      {(h.effects?.length > 0 || h.conc) && <>
        <h3><Icon n="effect" size={20} /> Эффекты</h3>
        <div className="chips">
          {h.conc && <span className="chip conc"><Icon n="conc" size={14} /> Концентрация: «{h.conc}»
            <button aria-label="Прервать концентрацию" onClick={() => act(api.DropConcentration(h.id).then(() => api.Characters().then((cs) => cs.find((c) => c.id === h.id))))}>✕</button></span>}
          {(h.effects ?? []).map((e, k) => <span className="chip fx" key={e.id ?? k} title={[e.cond && 'Состояние: ' + e.cond, e.save && `повторный спасбросок ${e.save.toUpperCase()} СЛ ${e.dc}`].filter(Boolean).join(' · ')}>
            <Icon n={e.conc ? 'conc' : 'effect'} size={13} /> {e.name}{e.rounds ? ` · ${e.rounds} х.` : ''}
            <button aria-label={`Снять эффект ${e.name}`} onClick={() => act(api.RemoveEffect(h.id, k))}>✕</button></span>)}
        </div></>}

      {d.slots?.length > 0 && <><h3>Ячейки заклинаний</h3>
        {d.slots.map((s) => <div className="slots" key={s.level}><span>{s.level} ур.</span>
          {Array.from({ length: s.max }, (_, i) => {
            const used = h.slotsUsed[s.level - 1]
            const arr = [...h.slotsUsed]
            return <button key={i} className="pip" aria-pressed={i < used} aria-label={`Ячейка ${s.level} уровня ${i + 1}`}
              onClick={() => { arr[s.level - 1] = i < used ? i : i + 1; save({ slotsUsed: arr }) }} />
          })}</div>)}</>}

      {!mon && <><h3>{rs.id === 'dnd5e' ? 'Заклинания' : 'Навыки'}</h3>
        {(h.spells ?? []).map((s, i) => <Spell key={(s.ref || s.name) + ':' + i} s={s} i={i} h={h} d={d} rs={rs} act={act} save={save} />)}
        {rs.spells?.length > 0 && <SpellPicker rs={rs} h={h} save={save} act={act} guard={guard} />}
        <div className="f sp">
          <input placeholder="Название" aria-label="Название" value={sp.name} onChange={(e) => setSp({ ...sp, name: e.target.value })} />
          {rs.id === 'dnd5e'
            ? <input type="number" min="0" max="9" aria-label="Уровень (0 – заговор)" title="Уровень (0 – заговор)" value={sp.level} onChange={(e) => setSp({ ...sp, level: +e.target.value })} />
            : <input type="number" min="0" aria-label="Мана" title="Стоимость маны" value={sp.cost} onChange={(e) => setSp({ ...sp, cost: +e.target.value })} />}
          <input placeholder="Описание" aria-label="Описание" value={sp.note} onChange={(e) => setSp({ ...sp, note: e.target.value })} />
          <button className="ghost" disabled={!sp.name.trim()} onClick={() => { save({ spells: [...(h.spells ?? []), sp] }); setSp({ ...sp, name: '', note: '' }) }}>Добавить</button>
        </div></>}

      <Inventory h={h} d={d} rs={rs} lib={lib} chars={chars} save={save} act={act} guard={guard} reload={reload} />

      <h3>Оружие</h3>
      {(h.weapons ?? []).map((x, i) => (
        <div className="irow" key={i}><b>{x.name}</b><small>{x.dice} · {rs.abilities.find((a) => a.id === x.ability)?.name} · {rs.damageTypes?.find((t) => t.id === x.type)?.name}{x.finesse ? ' · фехтовальное' : ''}{x.ranged ? ' · дальнобойное' : ''}</small>
          <button aria-label={`Удалить ${x.name}`} onClick={() => save({ weapons: h.weapons.filter((_, j) => j !== i) })}>✕</button></div>
      ))}
      <div className="f">
        <input placeholder="Название" aria-label="Название оружия" value={w.name} onChange={(e) => setW({ ...w, name: e.target.value })} />
        <input placeholder="Урон, напр. 1d8+1" aria-label="Урон" value={w.dice} onChange={(e) => setW({ ...w, dice: e.target.value })} />
        <select aria-label="Характеристика" value={w.ability} onChange={(e) => setW({ ...w, ability: e.target.value })}>{rs.abilities.map((a) => <option key={a.id} value={a.id}>{a.name}</option>)}</select>
        <select aria-label="Тип урона" value={w.type} onChange={(e) => setW({ ...w, type: e.target.value })}>{rs.damageTypes?.map((t) => <option key={t.id} value={t.id}>{t.name}</option>)}</select>
        <button className="ghost" disabled={!w.name.trim()} onClick={() => { save({ weapons: [...(h.weapons ?? []), w] }); setW({ ...w, name: '' }) }}>Добавить</button>
      </div>

      <h3>Заметки</h3>
      <textarea key={'n' + (h.notes ?? '')} defaultValue={h.notes} aria-label="Заметки" onBlur={(e) => e.target.value !== h.notes && save({ notes: e.target.value })} />
      <div className="row"><button className="ghost danger" onClick={() => (armed ? remove(h.id) : setArmed(true))} onBlur={() => setArmed(false)}>{armed ? 'Точно удалить?' : 'Удалить'}</button></div>
    </article>
  )
}

function Spell({ s, i, h, d, rs, act, save }) {
  const levels = (d.slots ?? []).map((x) => x.level).filter((l) => l >= s.level)
  const [slot, setSlot] = useState(0)
  const dnd = rs.id === 'dnd5e'
  return (
    <div className="irow"><span><b>{s.name}</b> <small>{lvlName(s.level)}{s.note ? ` · ${s.note}` : ''}</small></span>
      {dnd && s.level > 0 && levels.length > 1 && <select aria-label="Ячейка" value={slot || s.level} onChange={(e) => setSlot(+e.target.value)}>{levels.map((l) => <option key={l} value={l}>{l} ур.</option>)}</select>}
      <button onClick={() => act(api.Cast(h.id, i, dnd ? (slot || s.level) : 0))}>Сотворить</button>
      <button aria-label={`Удалить ${s.name}`} onClick={() => save({ spells: h.spells.filter((_, j) => j !== i) })}>✕</button></div>
  )
}

function ClassFeatures({ cls, h }) {
  const sub = cls.subclasses?.find((x) => x.id === h.subclass)
  const list = [...(cls.features ?? []), ...(sub?.features ?? [])].filter((f) => f.level <= h.level).sort((a, b) => a.level - b.level)
  const next = [...(cls.features ?? []), ...(sub?.features ?? [])].filter((f) => f.level > h.level).sort((a, b) => a.level - b.level)[0]
  return (
    <details className="cf">
      <summary>Умения класса: {cls.name}{sub ? ` · ${sub.name}` : ''} <small>({list.length})</small></summary>
      {list.map((f, i) => <p key={i}><b>{f.level} ур. {f.name}.</b> {f.desc}</p>)}
      {next && <p className="hint">Дальше: {next.level} ур. – {next.name}</p>}
    </details>
  )
}

function Skills({ rs, h, d, save, roll, mode, dc }) {
  const has = (id) => (h.skills ?? []).includes(id)
  const exp = (id) => (h.expert ?? []).includes(id)
  // клик: нет владения → владение → экспертиза → нет
  const cycle = (id) => {
    if (!has(id) && !exp(id)) save({ skills: [...(h.skills ?? []), id] })
    else if (!exp(id)) save({ expert: [...(h.expert ?? []), id] })
    else save({ skills: (h.skills ?? []).filter((x) => x !== id), expert: (h.expert ?? []).filter((x) => x !== id) })
  }
  return (
    <>
      <h3>Навыки <small>клик по метке – владение / экспертиза</small></h3>
      <div className="skills">
        {rs.skills.map((s) => (
          <div className="sk" key={s.id}>
            <button className="mark" aria-label={`${s.name}: ${exp(s.id) ? 'экспертиза' : has(s.id) ? 'владение' : 'нет владения'}`} onClick={() => cycle(s.id)}>{exp(s.id) ? '◆' : has(s.id) ? '●' : '○'}</button>
            <span>{s.name} <small>{rs.abilities.find((a) => a.id === s.ability)?.name.slice(0, 3)}</small></span>
            <button onClick={() => roll(() => api.Check(h.id, s.id, 'skill', mode, dc))}>{fmt(d.skills?.[s.id])}</button>
          </div>))}
      </div>
    </>
  )
}

// Мистический рыцарь и мистический ловкач берут заклинания из списка волшебника.
const spellClass = (h) => ((h.class === 'fighter' && h.subclass === 'eldritchknight') || (h.class === 'rogue' && h.subclass === 'arcanetrickster') ? 'wizard' : h.class ?? '')

function SpellPicker({ rs, h, save, act, guard }) {
  const [f, setF] = useState({ q: '', level: '', cls: spellClass(h), auto: false })
  const [mine, setMine] = useState([])
  useEffect(() => { guard(async () => setMine(await api.SpellLibrary())) }, [])
  const have = new Set((h.spells ?? []).map((s) => s.ref))
  const list = filterSpells(rs.spells, f).filter((s) => !have.has(s.id))
  const [id, setId] = useState('')
  const cur = list.find((s) => s.id === id) ?? list[0]
  const [tpl, setTpl] = useState('')
  const tcur = mine.find((t) => t.id === tpl) ?? mine[0]
  return (
    <details className="cf">
      <summary>Каталог заклинаний <small>({list.length})</small></summary>
      <SpellFilters f={f} setF={setF} />
      {list.length > 0 ? <div className="row tight">
        <select aria-label="Заклинание из каталога" value={cur?.id} onChange={(e) => setId(e.target.value)}>
          {list.slice(0, 400).map((s) => <option key={s.id} value={s.id}>{lvlName(s.level)} · {s.name}{s.mode ? ' ⚡' : ''}</option>)}</select>
        <button className="ghost" onClick={() => cur && save({ spells: [...(h.spells ?? []), { name: cur.name, level: cur.level, cost: 0, note: cur.desc, ref: cur.id }] })}>Выучить</button>
      </div> : <p className="hint">Ничего не найдено.</p>}
      {list.length > 400 && <p className="hint">Показаны первые 400 – уточните поиск.</p>}
      {cur && <p className="hint">{cur.school}. {cur.desc}{cur.mode ? ' Бросок и урон считаются автоматически в бою (⚡).' : ' Тратит ячейку, эффект применяет мастер.'}</p>}
      {mine.length > 0 && <div className="row tight"><b>Свои:</b>
        <select aria-label="Своё заклинание" value={tcur?.id} onChange={(e) => setTpl(e.target.value)}>{mine.map((t) => <option key={t.id} value={t.id}>{lvlName(t.level)} · {t.name}{t.auto?.mode ? ' ⚡' : ''}</option>)}</select>
        <button className="ghost" onClick={() => tcur && act(api.GiveSpellTemplate(h.id, tcur.id))}>Выдать герою</button></div>}
    </details>
  )
}

function Lore({ race, sub, cls }) {
  return (
    <details className="cf">
      <summary>Происхождение: {race?.name}{sub ? ` · ${sub.name}` : ''}{cls ? ` · ${cls.name}` : ''}</summary>
      {race && <><p><b>{race.name}.</b> {race.desc}</p>{sub?.desc && <p><b>{sub.name}.</b> {sub.desc}</p>}{race.lore && <p className="hint">{race.lore}</p>}</>}
      {cls && <><p><b>{cls.name}.</b> {cls.desc}</p>{cls.lore && <p className="hint">{cls.lore}</p>}</>}
    </details>
  )
}
