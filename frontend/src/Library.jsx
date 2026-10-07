import { useState } from 'react'
import { MonstersTab } from './Monsters.jsx'
import { ItemsTab } from './Items.jsx'
import { SpellsTab } from './Spells.jsx'
import Treasure from './Treasure.jsx'

const TABS = [['monsters', '🐉 Существа'], ['items', '🎒 Предметы'], ['spells', '📖 Заклинания'], ['treasure', '💰 Сокровища']]

// Library – справочники мастера: существа, предметы, заклинания; везде можно создавать своё.
export default function Library({ cat, lib, chars, guard, reload, notify }) {
  const [tab, setTab] = useState('monsters')
  return (
    <section>
      <h2>Библиотека</h2>
      <div className="seg tabs" role="tablist">{TABS.map(([k, v]) => <button key={k} role="tab" aria-selected={tab === k} aria-pressed={tab === k} onClick={() => setTab(k)}>{v}</button>)}</div>
      {tab === 'monsters' && <MonstersTab guard={guard} reload={reload} />}
      {tab === 'items' && <ItemsTab cat={cat} lib={lib} guard={guard} reload={reload} />}
      {tab === 'spells' && <SpellsTab cat={cat} guard={guard} />}
      {tab === 'treasure' && <Treasure chars={chars} guard={guard} reload={reload} notify={notify} />}
    </section>
  )
}
