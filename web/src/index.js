import React from 'react'
import ReactDOM from 'react-dom/client'
import './index.css'
import App from './App'
import { CinemaModeProvider } from './components/player'

const root = ReactDOM.createRoot(document.getElementById('root'))
root.render(
  <React.StrictMode>
    <CinemaModeProvider>
      <App />
    </CinemaModeProvider>
  </React.StrictMode>
)
