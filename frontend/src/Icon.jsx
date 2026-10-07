import { ICONS } from './icons.js'

// Icon – иконка из набора game-icons.net (CC BY 3.0). Неизвестное имя не рисуется.
export function Icon({ n, size = 18, title, className = '' }) {
  const d = ICONS[n]
  if (!d) return null
  return <svg className={'ic ' + className} width={size} height={size} viewBox="0 0 512 512" role={title ? 'img' : undefined} aria-label={title} aria-hidden={title ? undefined : true} dangerouslySetInnerHTML={{ __html: (title ? `<title>${title}</title>` : '') + d }} />
}

const RULES = [
  [/кинжал|нож|стилет/i, 'dagger'], [/топор|секира/i, 'axe'], [/арбалет/i, 'crossbow'], [/лук\b|лук,|стрел/i, 'bow'], [/копь|пик|глеф/i, 'spear'],
  [/булав|моргенштерн/i, 'mace'], [/молот|дубин/i, 'hammer'], [/цеп/i, 'flail'], [/алебард/i, 'halberd'], [/трезуб/i, 'trident'], [/праща/i, 'sling'],
  [/мушкет|ружь|винтов/i, 'musket'], [/пистолет/i, 'pistol'], [/посох|шест/i, 'staff'], [/жезл/i, 'rod'], [/палочк/i, 'wand'],
  [/меч|сабл|рапир|скимитар|палаш|клинок/i, 'sword'],
  [/щит/i, 'shield'], [/шлем/i, 'helm'], [/сапог|ботин/i, 'boots'], [/плащ|накидк/i, 'cloak'], [/перчат|рукавиц/i, 'gauntlet'],
  [/доспех|кольчуг|латы|кираса|кожан|стёган|бриган|чешуйч|полулат|кольчат/i, 'armor'],
  [/верёвк|канат/i, 'rope'], [/фонар|лампа/i, 'lantern'], [/факел/i, 'torch'], [/палатк/i, 'tent'], [/отмычк/i, 'lockpick'], [/ключ/i, 'key'], [/замок/i, 'lock'],
  [/зель|эликсир|настойк/i, 'potion'], [/свиток/i, 'scroll'], [/кольц|перстен/i, 'ring'], [/книг|гримуар|том\b/i, 'book'], [/перо/i, 'quill'],
  [/паёк|еда|хлеб|сыр|мяс/i, 'food'], [/вино|эль|пиво|фляг|бурдюк/i, 'drink'], [/бинт|повязк|аптеч|лекар/i, 'bandage'],
  [/самоцвет|алмаз|рубин|изумруд|сапфир|жемчуг|опал|топаз|аметист/i, 'gem'], [/монет/i, 'coins'],
]
const CAT = { weapon: 'sword', armor: 'armor', shield: 'shield', ammo: 'ammo', gear: 'gear', pack: 'pack', tool: 'tool', instrument: 'instrument', game: 'game', mount: 'mount', vehicle: 'vehicle', potion: 'potion', scroll: 'scroll', ring: 'ring', rod: 'rod', staff: 'staff', wand: 'wand', magic: 'magic' }

// itemIcon подбирает иконку по названию, затем по категории.
export function itemIcon(it) {
  const name = it?.name ?? ''
  for (const [re, ic] of RULES) if (re.test(name) && ICONS[ic]) return ic
  if (it?.weapon) return 'sword'
  if (it?.armorBase) return it.armorType === 'shield' ? 'shield' : 'armor'
  return CAT[it?.cat] ?? 'gear'
}

export const ItemIcon = ({ item, size = 20 }) => <Icon n={itemIcon(item)} size={size} />
export const CondIcon = ({ name, size = 14 }) => <Icon n={name} size={size} />
