import React from 'react'

function Player (props) {
  const videoRef = React.createRef()

  React.useEffect(() => {
    const peerConnection = new RTCPeerConnection() // eslint-disable-line

    peerConnection.ontrack = function (event) {
      videoRef.current.srcObject = event.streams[0]
    }

    peerConnection.addTransceiver('audio')
    peerConnection.addTransceiver('video')

    peerConnection.createOffer().then(offer => {
      peerConnection.setLocalDescription(offer)

      fetch('http://localhost:8080/api/whep', {
        method: 'POST',
        body: offer.sdp
      }).then(r => {
        return r.text()
      }).then(answer => {
        peerConnection.setRemoteDescription({
          sdp: answer,
          type: 'answer'
        })
      })
    })

    return function cleanup () {
      peerConnection.close()
    }
  }, [videoRef])

  return (
    <div className='w-full max-w-xs'>
      <video
        ref={videoRef}
        autoPlay
        controls
      />
    </div>
  )
}

export default Player
