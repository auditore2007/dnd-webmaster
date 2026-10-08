import { useEffect, useRef, useState } from 'react'
import { api } from './api'
import { pickMap } from './Portrait.jsx'
import { Icon } from './Icon.jsx'
import { MapTools } from './MapTools.jsx'
import { PlaceCard, PlaceMarker, PlaceTool } from './Places.jsx'

// partyLevel – средний уровень живых героев (для силы врагов и наград на картах).
const partyLevel = (chars) => {
  const hs = (chars ?? []).filter((c) => c.kind !== 'monster' && !c.dead)
  return hs.length ? Math.max(1, Math.round(hs.reduce((s, c) => s + c.level, 0) / hs.length)) : 1
}

// MapsView – галерея карт мастера: добавляй сколько угодно своих карт, у каждой свои метки.
export default function MapsView({ guard, rev = 0, chars = [], cat = [], reload, notify }) {
  const [panel, setPanel] = useState('')
  const [level, setLevel] = useState(() => partyLevel(chars))
  const [maps, setMaps] = useState(null)
  const [cur, setCur] = useState('')
  const [img, setImg] = useState('')
  const [name, setName] = useState('')
  const [renaming, setRenaming] = useState(false)
  const [armed, setArmed] = useState(false)
  const load = (pick) => guard(async () => { const m = await api.Maps(); setMaps(m); setCur((c) => pick ?? (m.some((x) => x.id === c) ? c : m[0]?.id ?? '')) })
  useEffect(() => { load() }, [rev])
  useEffect(() => { setImg(''); setArmed(false); setRenaming(false); if (cur) guard(async () => setImg(await api.GetMapImage(cur))) }, [cur])
  const map = maps?.find((m) => m.id === cur)
  // trail – цепочка карт от мира до текущей локации
  const trail = []
  for (let m = map; m && trail.length < 10; m = maps.find((x) => x.id === m.parent)) trail.unshift(m)
  const parentMap = map?.parent ? maps.find((m) => m.id === map.parent) : null
  const parentPlace = parentMap?.places?.find((p) => p.id === map.parentPlace)
  const add = (file) => guard(async () => {
    const data = await pickMap(file)
    const m = await api.AddMap(name.trim() || file.name.replace(/\.[^.]+$/, '') || 'Карта', data)
    setName(''); await load(m.id)
  })
  const onMap = (m) => setMaps((ms) => ms.map((x) => (x.id === m.id ? m : x)))
  const created = (m) => { notify?.(`Карта «${m.name}» готова: мест ${m.places?.length ?? 0}`); load(m.id) }
  if (!maps) return <section><h2>Карты</h2></section>
  return (
    <section className="map">
      <h2>Карты</h2>
      <div className="chips mtabs">
        {maps.filter((m) => !m.parent).map((m) => <button key={m.id} className="chip" aria-pressed={m.id === trail[0]?.id} onClick={() => setCur(m.id)}>{m.name}</button>)}
      </div>
      {trail.length > 1 && <nav className="row tight crumbs" aria-label="Путь к локации">
        {trail.map((m, i) => <span key={m.id}>{i > 0 && ' › '}{i < trail.length - 1
          ? <button className="ghost small" onClick={() => setCur(m.id)}>{m.name}</button> : <b>{m.name}</b>}</span>)}
      </nav>}
      <div className="row tight">
        <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Название новой карты" aria-label="Название новой карты" />
        <label className="primary small" title="Загрузить картинку карты со своего компьютера">＋ Добавить карту
          <input type="file" accept="image/*" hidden onChange={(e) => { const f = e.target.files[0]; e.target.value = ''; if (f) add(f) }} /></label>
        {map && !renaming && <button className="ghost" onClick={() => { setName(map.name); setRenaming(true) }}>Переименовать</button>}
        {map && renaming && <button className="ghost" onClick={() => guard(async () => { await api.RenameMap(map.id, name); setRenaming(false); setName(''); await load(map.id) })}>Сохранить название</button>}
        {map && <button className="ghost danger" onClick={() => (armed ? guard(async () => { await api.DeleteMap(map.id); setArmed(false); await load('') }) : setArmed(true))} onBlur={() => setArmed(false)}>{armed ? 'Точно удалить?' : 'Удалить карту'}</button>}
      </div>
      <div className="row tight mapgen">
        {[['world', '🌍 Создать мир'], ['dungeon', '🏰 Создать подземелье']].map(([k, v]) =>
          <button key={k} className="ghost" aria-pressed={panel === k} onClick={() => setPanel(panel === k ? '' : k)}>{v}</button>)}
      </div>
      <MapTools cat={cat} level={level} setLevel={setLevel} guard={guard} onDone={created} panel={panel} setPanel={setPanel} />
      {maps.length === 0
        ? <p className="empty">Карт пока нет. Создайте мир или подземелье прямо здесь – или загрузите свою картинку и разметьте её вручную.</p>
        : map && img ? <Viewer key={map.id + ':' + rev} map={map} img={img} guard={guard} reload={() => load(map.id)}
          onMap={onMap} cat={cat} level={level} reloadChars={reload} notify={notify}
          parentMap={parentMap} parentPlace={parentPlace} onOpenMap={(m) => load(m.id)} /> : <p className="hint">Загрузка карты…</p>}
    </section>
  )
}

const FOGMAX = 1024

// INSIDE – места, у которых на карте локации есть своя карта (здание изнутри, подземелье, замок в городе).
const INSIDE = ['tavern', 'temple', 'shop', 'smithy', 'dungeon', 'castle']

function Viewer({ map, img, guard, reload, onMap, cat, level, reloadChars, notify, parentMap, parentPlace, onOpenMap }) {
  const [sel, setSel] = useState('')
  const [about, setAbout] = useState(false)
  const [placeKind, setPlaceKind] = useState('tavern')
  const [placeName, setPlaceName] = useState('')
  const [showPlaces, setShowPlaces] = useState(true)
  const [labels, setLabels] = useState(() => (map.places?.length ?? 0) <= 40) // на густой карте подписи мешают
  const [busy, setBusy] = useState('')
  const places = map.places ?? []
  const selected = places.find((p) => p.id === sel)
  const box = useRef(null)
  const im = useRef(null)
  const fogc = useRef(null)
  const [v, setV] = useState({ x: 0, y: 0, k: 1 })
  const [text, setText] = useState('')
  const [tool, setTool] = useState('pan') // pan | pin | reveal | hide | ruler
  const [dim, setDim] = useState(null)
  const [brush, setBrush] = useState(60)
  const [player, setPlayer] = useState(false)
  const [fog, setFog] = useState(map.hasFog)
  const [ruler, setRuler] = useState(null)
  const [gridOn, setGridOn] = useState(map.grid > 0)
  const [cell, setCell] = useState(map.grid || 70)
  const [feet, setFeet] = useState(map.feet || 5)
  const pts = useRef(new Map())
  const drag = useRef({ start: null, last: 0, paint: false })
  const dirty = useRef(false)

  const fit = () => {
    const b = box.current, i = im.current
    if (!b || !i?.naturalWidth) return
    const k = Math.min(b.clientWidth / i.naturalWidth, b.clientHeight / i.naturalHeight)
    setV({ k, x: (b.clientWidth - i.naturalWidth * k) / 2, y: (b.clientHeight - i.naturalHeight * k) / 2 })
  }
  const zoom = (f, cx, cy) => setV((s) => {
    const b = box.current.getBoundingClientRect(), px = cx ?? b.width / 2, py = cy ?? b.height / 2
    const k = Math.min(6, Math.max(0.05, s.k * f)), r = k / s.k
    return { k, x: px - (px - s.x) * r, y: py - (py - s.y) * r }
  })
  const rect = () => box.current.getBoundingClientRect()
  // колесо мыши – масштаб. Обработчик не пассивный, иначе браузер заодно прокрутит страницу.
  useEffect(() => {
    const el = box.current
    if (!el) return
    const h = (e) => { e.preventDefault(); const r = el.getBoundingClientRect(); zoom(e.deltaY < 0 ? 1.15 : 1 / 1.15, e.clientX - r.left, e.clientY - r.top) }
    el.addEventListener('wheel', h, { passive: false })
    return () => el.removeEventListener('wheel', h)
  }, [])
  const toImg = (e) => { const r = rect(); return [(e.clientX - r.left - v.x) / v.k, (e.clientY - r.top - v.y) / v.k] }

  // ----- туман -----
  const fscale = dim ? Math.min(1, FOGMAX / Math.max(dim.w, dim.h)) : 1
  const fogSize = dim ? { w: Math.round(dim.w * fscale), h: Math.round(dim.h * fscale) } : null
  const fogFill = (reveal) => {
    const c = fogc.current; if (!c) return
    const g = c.getContext('2d'); g.globalCompositeOperation = 'source-over'
    g.clearRect(0, 0, c.width, c.height)
    if (!reveal) { g.fillStyle = '#000'; g.fillRect(0, 0, c.width, c.height) }
  }
  const saveFog = () => { dirty.current = false; guard(async () => { await api.SetFog(map.id, fogc.current.toDataURL('image/png')); await reload() }) }
  useEffect(() => {
    if (!fog || !fogSize) return
    let dead = false
    guard(async () => {
      const data = await api.GetFog(map.id)
      if (dead || !fogc.current) return
      if (!data) { fogFill(false); return }
      const im2 = new Image()
      im2.onload = () => { if (dead || !fogc.current) return; fogFill(true); fogc.current.getContext('2d').drawImage(im2, 0, 0, fogc.current.width, fogc.current.height) }
      im2.src = data
    })
    return () => { dead = true }
  }, [fog, dim])
  const stamp = (x, y, reveal) => {
    const c = fogc.current; if (!c) return
    const g = c.getContext('2d'); g.globalCompositeOperation = reveal ? 'destination-out' : 'source-over'
    g.fillStyle = '#000'; g.beginPath(); g.arc(x * fscale, y * fscale, brush * fscale, 0, Math.PI * 2); g.fill()
    dirty.current = true
  }
  const enableFog = () => { setFog(true); dirty.current = true }
  useEffect(() => { if (fog && !map.hasFog && dirty.current && fogc.current && fogSize) { fogFill(false); saveFog() } }, [fog, dim])
  const fogAll = (reveal) => { fogFill(reveal); saveFog() }
  const dropFog = () => guard(async () => { await api.SetFog(map.id, ''); setFog(false); await reload() })

  // ----- сетка -----
  const saveGrid = (on, size = cell, ft = feet) => guard(async () => { await api.SetMapGrid(map.id, on ? Math.round(size) : 0, ft); await reload() })
  const dist = ruler ? Math.hypot(ruler[2] - ruler[0], ruler[3] - ruler[1]) : 0
  const feetOf = map.grid > 0 ? (dist / map.grid) * map.feet : null

  const paint = (e) => { const [x, y] = toImg(e); stamp(x, y, tool === 'reveal') }
  const down = (e) => {
    if (e.target.closest('.pin')) return
    e.currentTarget.setPointerCapture(e.pointerId); pts.current.set(e.pointerId, [e.clientX, e.clientY])
    drag.current = { start: [e.clientX, e.clientY], last: 0, paint: false }
    if (pts.current.size === 1) {
      if ((tool === 'reveal' || tool === 'hide') && fog) { drag.current.paint = true; paint(e) }
      if (tool === 'ruler') { const [x, y] = toImg(e); setRuler([x, y, x, y]) }
    }
  }
  const move = (e) => {
    const o = pts.current.get(e.pointerId); if (!o) return
    const n = [e.clientX, e.clientY]
    if (pts.current.size === 1) {
      if (drag.current.paint) paint(e)
      else if (tool === 'ruler') { const [x, y] = toImg(e); setRuler((r) => r && [r[0], r[1], x, y]) }
      else setV((s) => ({ ...s, x: s.x + n[0] - o[0], y: s.y + n[1] - o[1] }))
    }
    pts.current.set(e.pointerId, n)
    if (pts.current.size === 2) {
      const [a, b] = [...pts.current.values()], d = Math.hypot(a[0] - b[0], a[1] - b[1]), r = rect()
      if (drag.current.last) zoom(d / drag.current.last, (a[0] + b[0]) / 2 - r.left, (a[1] + b[1]) / 2 - r.top)
      drag.current.last = d
    }
  }
  // select: выбрать место; pan – прокрутить карту к нему (переход из задания)
  // canOpen – у места есть своя карта: на карте мира у всех, кроме комнат; на карте локации – у зданий,
  // подземелий и замка (но не у построек внутри самого замка); на боевых картах (с сеткой) – ни у кого.
  const canOpen = (p) => p.kind !== 'room' && !(map.grid > 0) &&
    (!map.parent || (INSIDE.includes(p.kind) && !(p.kind === 'castle' && parentPlace?.kind === 'castle')))
  const openPlace = (p) => guard(async () => {
    setBusy(p.map ? `Открываем «${p.name}»…` : `Рисуем карту «${p.name}»…`)
    try { onOpenMap(await api.OpenPlace(map.id, p.id, level)) } finally { setBusy('') }
  })
  // pick – клик по месту: открыть его карту; с Shift (или если своей карты не бывает) – показать карточку.
  const pick = (id, e) => {
    const p = places.find((x) => x.id === id)
    if (p && canOpen(p) && !e?.shiftKey) openPlace(p)
    else select(id)
  }
  const select = (id, pan) => {
    setSel(id)
    const p = places.find((x) => x.id === id)
    // справа открыта карточка места – цель ставим в центр видимой левой части
    if (pan && p && box.current) setV((s) => ({ ...s, x: box.current.clientWidth * 0.32 - p.x * s.k, y: box.current.clientHeight / 2 - p.y * s.k }))
  }
  const up = (e) => {
    const s = drag.current.start
    if (tool === 'place' && pts.current.size === 1 && s && Math.hypot(e.clientX - s[0], e.clientY - s[1]) < 5) {
      const [x, y] = toImg(e)
      guard(async () => { const m = await api.AddPlace(map.id, x, y, placeKind, placeName.trim()); onMap(m); setPlaceName(''); setSel(m.places.at(-1)?.id ?? '') })
    }
    if (tool === 'pin' && pts.current.size === 1 && s && Math.hypot(e.clientX - s[0], e.clientY - s[1]) < 5) {
      const [x, y] = toImg(e)
      guard(async () => { await api.AddPin(map.id, x, y, text.trim() || 'Метка'); await reload() })
    }
    if (drag.current.paint && dirty.current) saveFog()
    pts.current.delete(e.pointerId); drag.current.last = 0; drag.current.paint = false
  }
  const cursor = tool === 'pan' ? 'grab' : tool === 'pin' || tool === 'ruler' || tool === 'place' ? 'crosshair' : 'cell'
  const T = (k, ic, label) => <button className="ghost" aria-pressed={tool === k} onClick={() => setTool(k)} title={label}><Icon n={ic} /> {label}</button>
  const fogged = fog && fogSize
  return (
    <>
      <div className="row tight tools">
        {T('pan', 'map', 'Двигать')}{T('place', 'pin', 'Места')}{T('pin', 'pin', 'Метки')}{T('ruler', 'ruler', 'Линейка')}
        {fog && <>{T('reveal', 'fog', 'Открыть')}{T('hide', 'fog', 'Закрыть')}</>}
        <button className="ghost" onClick={() => zoom(1.25)} aria-label="Приблизить">+</button>
        <button className="ghost" onClick={() => zoom(0.8)} aria-label="Отдалить">−</button>
        <button className="ghost" onClick={fit}>По размеру окна</button>
        <small>{map.pins.length ? `меток: ${map.pins.length}` : 'меток нет'} · мест: {places.length}</small>
      </div>
      <div className="row tight">
        {busy ? <span className="busy"><span className="spinner">🎲</span> {busy}</span> : <>
          {parentPlace && <button className="ghost" aria-pressed={about} onClick={() => { setSel(''); setAbout(!about) }}>📜 О месте: {parentPlace.name}</button>}
          <label className="eq"><input type="checkbox" checked={showPlaces} onChange={(e) => setShowPlaces(e.target.checked)} /> показывать места</label>
          <label className="eq"><input type="checkbox" checked={labels} onChange={(e) => setLabels(e.target.checked)} /> подписи городов</label></>}
      </div>
      {tool === 'place' && <PlaceTool kind={placeKind} setKind={setPlaceKind} name={placeName} setName={setPlaceName} />}
      {tool === 'pin' && <div className="row tight"><input value={text} onChange={(e) => setText(e.target.value)} placeholder="Подпись метки, затем клик по карте" aria-label="Подпись метки" /></div>}
      {(tool === 'reveal' || tool === 'hide') && <div className="row tight"><label className="inl">Кисть <input type="range" min="10" max="300" value={brush} onChange={(e) => setBrush(+e.target.value)} /> {brush}</label>
        <button className="ghost" onClick={() => fogAll(true)}>Открыть всё</button><button className="ghost" onClick={() => fogAll(false)}>Закрыть всё</button></div>}
      <details className="cf" open={false}>
        <summary><Icon n="fog" /> Туман войны · <Icon n="grid" /> сетка · <Icon n="ruler" /> масштаб</summary>
        <div className="row tight">
          {!fog ? <button className="ghost" onClick={enableFog}><Icon n="fog" /> Включить туман (вся карта скрыта)</button>
            : <><button className="ghost" aria-pressed={player} onClick={() => setPlayer(!player)} title="Как видят игроки: скрытое полностью чёрное">{player ? 'Вид игроков' : 'Вид мастера'}</button>
              <button className="ghost danger" onClick={dropFog}>Убрать туман</button></>}
        </div>
        <div className="row tight">
          <label className="inl"><input type="checkbox" checked={gridOn} onChange={(e) => { setGridOn(e.target.checked); saveGrid(e.target.checked) }} /> Сетка</label>
          <label className="inl">клетка, px <input type="number" min="8" max="2000" value={cell} onChange={(e) => setCell(+e.target.value)} onBlur={() => gridOn && saveGrid(true)} style={{ width: 80 }} /></label>
          <label className="inl">футов в клетке <input type="number" min="1" max="1000" value={feet} onChange={(e) => setFeet(+e.target.value)} onBlur={() => gridOn && saveGrid(true)} style={{ width: 70 }} /></label>
        </div>
        <p className="hint">Линейка: выберите «Линейка» и протяните по карте. Чтобы подогнать сетку под картинку, измерьте одну клетку на карте линейкой и нажмите «Принять за клетку».</p>
      </details>
      {ruler && dist > 2 && <div className="rulerinfo"><Icon n="ruler" /> {feetOf != null ? <b>{Math.round(feetOf)} фт ({(feetOf / map.feet).toFixed(1)} клеток)</b> : <b>{Math.round(dist)} px</b>}
        {feetOf == null && <small> – включите сетку, чтобы считать в футах</small>}
        <button className="ghost small" onClick={() => { setCell(Math.round(dist)); setGridOn(true); saveGrid(true, dist) }}>Принять за клетку</button>
        <button className="ghost small" onClick={() => setRuler(null)}>Скрыть</button></div>}
      <div className="mapwrap">
      {selected && <PlaceCard key={selected.id} map={map} place={selected} cat={cat} level={level} guard={guard} onMap={onMap} onSelect={select} reload={reloadChars} notify={notify}
        onOpen={canOpen(selected) ? () => openPlace(selected) : null} />}
      {!selected && about && parentMap && parentPlace && <PlaceCard key={'about' + parentPlace.id} map={parentMap} place={parentPlace} cat={cat} level={level} guard={guard}
        onMap={() => reload()} onSelect={() => setAbout(false)} reload={reloadChars} notify={notify} />}
      <div className="mapbox" ref={box} style={{ cursor }} onPointerDown={down} onPointerMove={move} onPointerUp={up} onPointerCancel={up}>
        <div className="mapl" style={{ transform: `translate(${v.x}px,${v.y}px) scale(${v.k})`, '--ik': 1 / v.k }}>
          <img ref={im} src={img} alt={map.name} draggable="false" onLoad={(e) => { setDim({ w: e.target.naturalWidth, h: e.target.naturalHeight }); fit() }} />
          {map.grid > 0 && dim && <div className="gridl" style={{ width: dim.w, height: dim.h, backgroundSize: `${map.grid}px ${map.grid}px`, '--gw': `${Math.max(1, 1 / v.k)}px` }} />}
          {fogged && <canvas ref={fogc} className="fogc" data-testid="fog" width={fogSize.w} height={fogSize.h} style={{ width: dim.w, height: dim.h, opacity: player ? 1 : 0.55 }} />}
          {ruler && <svg className="rulersvg" width={dim?.w} height={dim?.h}><line x1={ruler[0]} y1={ruler[1]} x2={ruler[2]} y2={ruler[3]} stroke="#ffd34d" strokeWidth={3 / v.k} strokeLinecap="round" />
            <circle cx={ruler[0]} cy={ruler[1]} r={5 / v.k} fill="#ffd34d" /><circle cx={ruler[2]} cy={ruler[3]} r={5 / v.k} fill="#ffd34d" /></svg>}
          {map.pins.map((p, i) => <div className="pin" key={`${i}:${p.x}:${p.y}`} style={{ left: p.x, top: p.y }}><span>{p.text}</span>
            <button aria-label={`Удалить метку ${p.text}`} onClick={() => guard(async () => { await api.RemovePin(map.id, i); await reload() })}>✕</button></div>)}
          {showPlaces && places.map((p) => <PlaceMarker key={p.id} p={p} selected={p.id === sel} onSelect={pick} labels={labels} opens={canOpen(p)} />)}
        </div>
      </div>
      </div>
    </>
  )
}
