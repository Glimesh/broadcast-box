import React, { useContext } from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'

import Header from './components/header'
import Selection from './components/selection'
import PlayerPage, { CinemaModeContext } from './components/player'
import Publish from './components/publish'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path='/' element={<Header />}>
          <Route index element={<Selection />} />
          <Route path='/publish/*' element={<Publish />} />
          <Route path='/*' element={<PlayerPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
