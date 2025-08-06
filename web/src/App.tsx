import React from 'react'
import { Routes, Route } from 'react-router-dom'

import BrowserBroadcaster from "./components/broadcast/Broadcast";
import PlayerPage from "./components/player/PlayerPage";
import RootWrapper from "./components/rootWrapper/RootWrapper";
import Frontpage from "./components/selection/Frontpage";
import Statistics from "./components/statistics/Statistics";

function App() {
  return (
    <Routes>
      <Route path='/' element={<RootWrapper />}>
        <Route index element={<Frontpage/>} />
        <Route path='/statistics' element={<Statistics/>} />
        <Route path='/publish/*' element={<BrowserBroadcaster />} />
        <Route path='/*' element={<PlayerPage/>} />
      </Route>
    </Routes>
  )
}

export default App