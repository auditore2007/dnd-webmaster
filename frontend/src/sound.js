// Звук кубиков, синтезированный WebAudio: короткие щелчки шума, без аудиофайлов. Включение запоминается в браузере.

const KEY = 'hb-sound'
let ctx = null
let noise = null

export function soundOn() {
  try { return localStorage.getItem(KEY) !== 'off' } catch { return true }
}

export function setSoundOn(on) {
  try { localStorage.setItem(KEY, on ? 'on' : 'off') } catch { /* хранилище недоступно – звук просто не запомнится */ }
}

function audio() {
  if (ctx) return ctx
  const AC = typeof window !== 'undefined' && (window.AudioContext || window.webkitAudioContext)
  if (!AC) return null
  try {
    ctx = new AC()
    noise = ctx.createBuffer(1, ctx.sampleRate * 0.05, ctx.sampleRate)
    const d = noise.getChannelData(0)
    for (let i = 0; i < d.length; i++) d[i] = (Math.random() * 2 - 1) * (1 - i / d.length) ** 3
  } catch {
    ctx = null
  }
  return ctx
}

// clack – один удар кубика о стол: отфильтрованный всплеск шума.
function clack(a, at, vol, freq) {
  const src = a.createBufferSource()
  src.buffer = noise
  const f = a.createBiquadFilter()
  f.type = 'bandpass'
  f.frequency.value = freq
  f.Q.value = 4
  const g = a.createGain()
  g.gain.setValueAtTime(vol, at)
  g.gain.exponentialRampToValueAtTime(0.001, at + 0.06)
  src.connect(f).connect(g).connect(a.destination)
  src.start(at)
}

/** playRoll – стук катящихся кубиков (чем больше кубиков, тем гуще). */
export function playRoll(count = 1, ms = 800) {
  if (!soundOn()) return
  const a = audio()
  if (!a) return
  if (a.state === 'suspended') a.resume().catch(() => {})
  const hits = Math.min(18, 6 + count * 2)
  for (let i = 0; i < hits; i++) {
    const t = a.currentTime + (ms / 1000) * (i / hits) ** 0.8 + Math.random() * 0.03
    clack(a, t, 0.18 * (1 - (i / hits) * 0.6), 1800 + Math.random() * 2600)
  }
}

/** playLand – финальный удар; крит звенит выше, провал – глухо. */
export function playLand(tone = '') {
  if (!soundOn()) return
  const a = audio()
  if (!a) return
  const t = a.currentTime
  clack(a, t, 0.3, tone === 'bad' ? 700 : 2200)
  if (tone !== 'crit') return
  for (const [i, hz] of [[0, 880], [1, 1320], [2, 1760]]) {
    const o = a.createOscillator()
    const g = a.createGain()
    o.frequency.value = hz
    g.gain.setValueAtTime(0.0001, t + i * 0.07)
    g.gain.exponentialRampToValueAtTime(0.08, t + i * 0.07 + 0.02)
    g.gain.exponentialRampToValueAtTime(0.0001, t + i * 0.07 + 0.5)
    o.connect(g).connect(a.destination)
    o.start(t + i * 0.07)
    o.stop(t + i * 0.07 + 0.55)
  }
}
