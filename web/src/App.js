import React from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'

import Header from './components/header'
import Selection from './components/selection'
import Player from './components/player'

function App() {
  return (
    <div>
      <BrowserRouter>
        <Routes>
          <Route path='/' element={<Header />}>
            <Route index element={<Selection />} />
            <Route path='*' element={<Player />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </div>
  )
}

export default App
