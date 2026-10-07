import { createRoot } from 'react-dom/client'
// только кириллица и латиница: остальные наборы (греческий, вьетнамский…) интерфейсу не нужны и раздували exe
import '@fontsource/alegreya/cyrillic-700.css'
import '@fontsource/alegreya/latin-700.css'
import '@fontsource/alegreya/cyrillic-800.css'
import '@fontsource/alegreya/latin-800.css'
import '@fontsource/onest/cyrillic-400.css'
import '@fontsource/onest/latin-400.css'
import '@fontsource/onest/cyrillic-600.css'
import '@fontsource/onest/latin-600.css'
import './style.css'
import App from './App.jsx'

createRoot(document.getElementById('root')).render(<App />)
