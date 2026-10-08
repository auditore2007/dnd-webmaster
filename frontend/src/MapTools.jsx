import { useEffect, useState } from 'react'
import { api } from './api'
import { pickMap } from './Portrait.jsx'

// Создание карт прямо в приложении (мир, подземелье), импорт из Azgaar и One Page Dungeon, настройки ИИ.

const readText = (file) => new Promise((resolve, reject) => {
  const r = new FileReader()
  r.onload = () => resolve(String(r.result))
  r.onerror = () => reject(new Error('Не удалось прочитать файл'))
  r.readAsText(file)
})

const SIZES = [['small', 'Маленькая'], ['medium', 'Средняя'], ['large', 'Большая']]

// TEMPLATES – шаблоны рельефа генератора (как в Azgaar's Fantasy Map Generator).
const TEMPLATES = [['', 'Случайный'], ['continents', 'Материки'], ['oldWorld', 'Старый Свет'], ['pangea', 'Пангея'],
  ['highIsland', 'Высокий остров'], ['lowIsland', 'Низкий остров'], ['archipelago', 'Архипелаг'], ['mediterranean', 'Внутреннее море'],
  ['peninsula', 'Полуостров'], ['isthmus', 'Перешеек'], ['shattered', 'Осколки'], ['fractious', 'Раздробленные земли'],
  ['volcano', 'Вулкан'], ['atoll', 'Атолл'], ['taklamakan', 'Пустынное нагорье']]

/** MapTools – вкладки «Создать мир», «Подземелье», «Импорт», «ИИ». onDone(map) – новая карта готова. */
export function MapTools({ cat, level, setLevel, guard, onDone, panel, setPanel }) {
  if (!panel) return null
  const close = () => setPanel('')
  return (
    <div className="editor maptools">
      {panel === 'world' && <WorldForm cat={cat} level={level} guard={guard} onDone={onDone} close={close} />}
      {panel === 'dungeon' && <DungeonForm level={level} guard={guard} onDone={onDone} close={close} />}
      {panel === 'import' && <ImportForm level={level} guard={guard} onDone={onDone} close={close} />}
      {panel === 'ai' && <AIForm guard={guard} close={close} />}
      <label className="eq">Уровень группы <input type="number" min="1" max="20" value={level} onChange={(e) => setLevel(Math.max(1, Math.min(20, parseInt(e.target.value, 10) || 1)))} style={{ width: 60 }} />
        <small>– от него зависят враги и награды</small></label>
    </div>
  )
}

function Busy({ busy, children }) {
  return busy ? <span className="busy"><span className="spinner">🎲</span> {busy}</span> : children
}

function WorldForm({ cat, level, guard, onDone, close }) {
  const races = cat[0]?.races ?? []
  const [name, setName] = useState('')
  const [size, setSize] = useState('medium')
  const [template, setTemplate] = useState('')
  const [seed, setSeed] = useState('')
  const [picked, setPicked] = useState(() => races.map((r) => r.id))
  const [busy, setBusy] = useState('')
  const toggle = (id) => setPicked((p) => (p.includes(id) ? p.filter((x) => x !== id) : [...p, id]))
  const go = () => guard(async () => {
    setBusy('Рисуем горы, реки и расселяем народы…')
    try { onDone(await api.GenerateWorld(name, parseInt(seed, 10) || 0, picked, level, size, template)); close() } finally { setBusy('') }
  })
  return (
    <>
      <h3>🌍 Создать карту мира</h3>
      <p className="hint">Карта в стиле старинного атласа: рельеф по шаблонам Azgaar, климат и биомы, реки и озёра, государства народов с границами, городами, замками и дорогами. Каждая выбранная раса получает столицу и поселения в своей местности: дварфы – в горах, эльфы – в лесах, ящеролюды – в болотах. В диких землях появятся логова, руины и пещеры с врагами под уровень группы, в поселениях – жители и задания.</p>
      <div className="grid2">
        <label>Название<input value={name} onChange={(e) => setName(e.target.value)} placeholder="Новый мир" /></label>
        <label>Размер<select value={size} onChange={(e) => setSize(e.target.value)}>{SIZES.map(([k, v]) => <option key={k} value={k}>{v}</option>)}</select></label>
        <label>Рельеф<select value={template} onChange={(e) => setTemplate(e.target.value)}>{TEMPLATES.map(([k, v]) => <option key={k} value={k}>{v}</option>)}</select></label>
        <label>Зерно (одно и то же число – та же карта)<input value={seed} onChange={(e) => setSeed(e.target.value.replace(/\D/g, ''))} placeholder="случайно" /></label>
      </div>
      <div className="row tight"><b>Народы мира</b> <small>{picked.length} из {races.length}</small>
        <button className="ghost small" onClick={() => setPicked(races.map((r) => r.id))}>Все</button>
        <button className="ghost small" onClick={() => setPicked([])}>Никого</button></div>
      <div className="chips">{races.map((r) => <button key={r.id} className="chip" aria-pressed={picked.includes(r.id)} onClick={() => toggle(r.id)}>{r.name}</button>)}</div>
      <div className="row"><Busy busy={busy}><button className="primary" disabled={!picked.length} onClick={go}>🌍 Создать мир</button></Busy>
        <button className="ghost" onClick={close}>Закрыть</button></div>
    </>
  )
}

function DungeonForm({ level, guard, onDone, close }) {
  const [name, setName] = useState('')
  const [rooms, setRooms] = useState(10)
  const [busy, setBusy] = useState('')
  const go = () => guard(async () => {
    setBusy('Копаем коридоры…')
    try { onDone(await api.GenerateDungeon(name, 0, rooms, level)); close() } finally { setBusy('') }
  })
  return (
    <>
      <h3>🏰 Создать подземелье</h3>
      <p className="hint">Комнаты, коридоры с петлями и дверями, колонны в больших залах. В комнатах – ловушки (со сложностью и уроном), тайники и враги под уровень группы; треть комнат пустует.</p>
      <div className="grid2">
        <label>Название<input value={name} onChange={(e) => setName(e.target.value)} placeholder="Подземелье" /></label>
        <label>Комнат<input type="number" min="4" max="40" value={rooms} onChange={(e) => setRooms(Math.max(4, Math.min(40, parseInt(e.target.value, 10) || 4)))} /></label>
      </div>
      <div className="row"><Busy busy={busy}><button className="primary" onClick={go}>🏰 Создать</button></Busy><button className="ghost" onClick={close}>Закрыть</button></div>
    </>
  )
}

function ImportForm({ level, guard, onDone, close }) {
  const [json, setJson] = useState(null)
  const [image, setImage] = useState(null)
  const [busy, setBusy] = useState('')
  const azgaar = () => guard(async () => {
    setBusy('Читаем карту Azgaar…')
    try { onDone(await api.ImportAzgaar('', await readText(json), await pickMap(image), level)); close() } finally { setBusy('') }
  })
  const opd = (file) => guard(async () => {
    setBusy('Рисуем подземелье…')
    try { onDone(await api.ImportOnePageDungeon('', await readText(file), level)); close() } finally { setBusy('') }
  })
  return (
    <>
      <h3>📥 Импорт карты с разметкой</h3>
      <p className="hint">Карты из бесплатных генераторов приходят вместе с данными о том, что где стоит, – метки встают точно, без угадывания.</p>
      <div className="impbox">
        <b>Azgaar's Fantasy Map Generator</b> <a href="https://azgaar.github.io/Fantasy-Map-Generator/" target="_blank" rel="noreferrer">открыть</a>
        <p className="hint">В Azgaar: «Export → JSON → Minimal» и «Export → PNG». Выберите оба файла. Города станут поселениями, метки – тавернами, подземельями и логовами; каждой культуре достанется своя раса.</p>
        <div className="row tight">
          <label className="ghost small">{json ? '✓ ' + json.name : 'JSON-файл'}<input type="file" accept=".json,application/json" hidden onChange={(e) => setJson(e.target.files[0] ?? null)} /></label>
          <label className="ghost small">{image ? '✓ ' + image.name : 'Картинка карты'}<input type="file" accept="image/*" hidden onChange={(e) => setImage(e.target.files[0] ?? null)} /></label>
          <Busy busy={busy}><button className="primary small" disabled={!json || !image} onClick={azgaar}>Импортировать</button></Busy>
        </div>
      </div>
      <div className="impbox">
        <b>One Page Dungeon</b> <a href="https://watabou.itch.io/one-page-dungeon" target="_blank" rel="noreferrer">открыть</a>
        <p className="hint">В генераторе нажмите J (экспорт JSON) и выберите файл: подземелье нарисуется заново, заметки станут комнатами с врагами.</p>
        <label className="ghost small">JSON-файл подземелья<input type="file" accept=".json,application/json" hidden onChange={(e) => { const f = e.target.files[0]; e.target.value = ''; if (f) opd(f) }} /></label>
      </div>
      <div className="row"><button className="ghost" onClick={close}>Закрыть</button></div>
    </>
  )
}

function AIForm({ guard, close }) {
  const [view, setView] = useState(null)
  const [key, setKey] = useState('')
  const [model, setModel] = useState('')
  useEffect(() => { guard(async () => { const v = await api.AISettings(); setView(v); setModel(v.model) }) }, [guard])
  const save = (clear = false) => guard(async () => { setView(await api.SetAISettings(key, model, clear)); setKey('') })
  if (!view) return null
  return (
    <>
      <h3>🤖 Распознавание карт – Google Gemini</h3>
      <p className="hint">Нейросеть смотрит на картинку, находит города, таверны, логова и руины, а приложение ставит метки и населяет их. Бесплатный ключ: <a href="https://aistudio.google.com/apikey" target="_blank" rel="noreferrer">aistudio.google.com/apikey</a>. Нужен интернет; в некоторых странах Gemini недоступен без VPN. На бесплатном тарифе Google может использовать картинки для улучшения моделей.</p>
      <p>{view.hasKey ? <>Ключ сохранён <b>{view.keyHint}</b></> : <span className="bad">Ключ не задан</span>}</p>
      <div className="grid2">
        <label>Ключ API<input type="password" value={key} onChange={(e) => setKey(e.target.value)} placeholder={view.hasKey ? 'оставьте пустым, чтобы не менять' : 'AIza…'} autoComplete="off" /></label>
        <label>Модель<input value={model} onChange={(e) => setModel(e.target.value)} placeholder="gemini-3.8-flash" /></label>
      </div>
      <div className="row"><button className="primary" onClick={() => save(false)}>Сохранить</button>
        {view.hasKey && <button className="ghost danger" onClick={() => save(true)}>Удалить ключ</button>}
        <button className="ghost" onClick={close}>Закрыть</button></div>
    </>
  )
}
