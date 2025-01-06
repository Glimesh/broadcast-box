import React from 'react'
import ReactDOM from 'react-dom/client'
import './index.css'
import App from './App'
import { CinemaModeProvider } from './components/player'
import { BrowserRouter } from 'react-router-dom'

const root = ReactDOM.createRoot(document.getElementById('root'))
root.render(
  <React.StrictMode>
    <BrowserRouter basename={process.env.PUBLIC_URL}>
      <CinemaModeProvider>
        <App />
      </CinemaModeProvider>
    </BrowserRouter>
  </React.StrictMode>
)
