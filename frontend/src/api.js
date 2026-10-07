// Методы Go-структуры App; привязки wailsjs генерирует сам Wails при dev/build.
// Пустой срез в Go приходит как null – для списков заменяем на [].
const LISTS = new Set(['Characters', 'ItemLibrary', 'Maps', 'SpellLibrary', 'Bestiary', 'Log', 'Snapshots', 'Catalog', 'AddMonster'])
export const api = new Proxy({}, {
  get: (_, k) => async (...a) => {
    const r = await window.go.main.App[k](...a)
    return LISTS.has(k) ? r ?? [] : r
  },
})
