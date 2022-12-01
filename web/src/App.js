import React from 'react'

import AdminPanel from './components/admin-panel'
import Player from './components/player'

function App () {
  const [broadcastBoxStatus, setBroadcastBoxStatus] = React.useState('')
  const fetchStatus = () => {
    fetch('http://localhost:8080/api/status')
      .then(r => {
        return r.json()
      }).then(r => {
        setBroadcastBoxStatus(r.status)
      })
  }
  React.useEffect(fetchStatus, [])

  switch (broadcastBoxStatus) {
    case 'unconfigured':
      return <AdminPanel onConfigurationSuccess={fetchStatus} />
    case 'configured':
      return <Player />
    default:
      return <h1> Backend returned unexpected status </h1>
  }
}

export default App
