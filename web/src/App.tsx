import React from 'react'
import { Routes, Route } from 'react-router-dom'

import RootWrapper from './components/rootWrapper/rootWrapper'
import Frontpage from "./components/selection/frontpage";
import BrowserBroadcaster from "./components/broadcast/Broadcast";
import PlayerPage from "./components/player/PlayerPage";

function App() {
  return (
    <Routes>
      <Route path='/' element={<RootWrapper />}>
        <Route index element={<Frontpage/>} />
        <Route path='/publish/*' element={<BrowserBroadcaster />} />
        <Route path='/*' element={<PlayerPage/>} 
      />
      </Route>
    </Routes>
  )
}

export default App