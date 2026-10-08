import { describe, expect, it } from 'vitest'
import { renderToString } from 'react-dom/server'
import { GroupPanel, TripPanel } from './Maps.jsx'

const render = (el) => renderToString(el).replace(/<!-- -->/g, '')

const chars = [
  { id: 'a', name: 'Лира', kind: 'hero' },
  { id: 'b', name: 'Торг', kind: 'hero' },
  { id: 'c', name: 'Эльм', kind: 'hero' },
  { id: 'g', name: 'Гоблин', kind: 'monster' },
]
const groups = [
  { id: 'g1', name: 'Отряд', x: 10, y: 10, members: ['a', 'b'], color: '#e0b040' },
  { id: 'g2', name: 'Разведка', x: 50, y: 50, members: ['c'], color: '#4aa3df' },
]
const noop = () => {}

describe('группы на карте', () => {
  it('показывает группы, состав и действия', () => {
    const html = render(<GroupPanel groups={groups} group={groups[0]} chars={chars} making={null} setMaking={noop} select={noop} save={noop} remove={noop} />)
    expect(html).toContain('Отряд')
    expect(html).toContain('Разведка')
    expect(html).toContain('Торг')
    expect(html).toContain('присоединить к')
    expect(html).not.toContain('Гоблин')
  })
  it('новая группа: выбор героев без существ', () => {
    const html = render(<GroupPanel groups={[]} chars={chars} making={{ name: '', members: ['a'] }} setMaking={noop} select={noop} save={noop} remove={noop} />)
    expect(html).toContain('Новая группа')
    expect(html).toContain('Щёлкните по карте')
    expect(html).not.toContain('Гоблин')
  })
  it('маршрут и встречи в пути', () => {
    const trip = { miles: 120.4, days: 5.25, legs: [{ terrain: 'дорога', miles: 80 }, { terrain: 'лес', miles: 40.4 }], path: [], dest: [1, 1], done: true,
      encounters: [{ day: 2, terrain: 'лес', text: 'Из чащи выходят на тропу', foes: [{ id: 'wolf', name: 'Волк', cr: '1/4', count: 3 }] }] }
    const html = render(<TripPanel trip={trip} group={groups[0]} pace="normal" setPace={noop} busy="" go={noop} moveParty={noop} close={noop} toTable={noop} />)
    expect(html).toContain('120 миль')
    expect(html).toContain('День 2')
    expect(html).toContain('Волк ×3')
    expect(renderToString(<TripPanel trip={null} group={groups[0]} />)).toBe('')
  })
})
