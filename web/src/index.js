import React from 'react'
import ReactDOM from 'react-dom/client'
import './index.css'
import App from './App'
import { CinemaModeProvider } from './components/player'
import { BrowserRouter as Router } from 'react-router-dom'

const root = ReactDOM.createRoot(document.getElementById('root'))
root.render(
  <React.StrictMode>
    <Router>
      <CinemaModeProvider>
        <App />
      </CinemaModeProvider>
    </Router>
  </React.StrictMode>
)
