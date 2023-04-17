import React from 'react'
import { useLocation } from 'react-router-dom'

function Player(props) {
  const videoRef = React.createRef()
  const location = useLocation()

  React.useEffect(() => {
    const peerConnection = new RTCPeerConnection() // eslint-disable-line
    let stream = null

    navigator.mediaDevices.getUserMedia({
      audio: true,
      video: true
    })
    .then(s => {
      if (peerConnection.connectionState === "closed") {
        s.getTracks().forEach(t => t.stop())
        return;
      }

      stream = s
      videoRef.current.srcObject = s

      s.getTracks().forEach(t => {
        peerConnection.addTransceiver(t, {direction: 'sendonly'})
      })

      peerConnection.createOffer().then(offer => {
        peerConnection.setLocalDescription(offer)

        fetch(`${process.env.REACT_APP_API_PATH}/whip`, {
          method: 'POST',
          body: offer.sdp,
          headers: {
            Authorization: `Bearer ${location.pathname.substring(1).replace('publish/', '')}`,
            'Content-Type': 'application/sdp'
          }
        }).then(r => {
          return r.text()
        }).then(answer => {
          peerConnection.setRemoteDescription({
            sdp: answer,
            type: 'answer'
          })
        })
      })
    })

    return function cleanup() {
      peerConnection.close()
      if (stream !== null) {
        stream.getTracks().forEach(t => t.stop())
      }
    }
  }, [videoRef, location.pathname])

  return (
    <video
      ref={videoRef}
      autoPlay
      muted
      controls
      playsInline
      className='mx-auto h-full'
    />
  )
}

export default Player
