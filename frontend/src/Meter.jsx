import { usePulse } from './motion.js'

const LOW = 30
const MID = 60

/**
 * HpMeter – полоска здоровья: цвет зависит от остатка, урон даёт вспышку и встряску, лечение – свечение.
 * @param {{hp:number, max:number, temp?:number, label?:string}} props
 */
export function HpMeter({ hp, max, temp = 0, label }) {
  const pct = Math.max(0, Math.min(100, Math.round((Math.max(0, hp) / Math.max(1, max)) * 100)))
  const pulse = usePulse(hp)
  const lvl = pct <= LOW ? ' low' : pct <= MID ? ' mid' : ''
  return (
    <div className={'meter hpm' + lvl + (pulse ? ' ' + pulse : '')} style={{ '--p': pct + '%' }} role="meter" aria-valuenow={hp} aria-valuemin={0} aria-valuemax={max} aria-label={label ?? 'Здоровье'}>
      <b>{hp} / {max}{temp > 0 ? ` (+${temp})` : ''}</b>
    </div>
  )
}
