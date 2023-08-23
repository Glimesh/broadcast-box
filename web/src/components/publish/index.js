import React from 'react'
import { useLocation } from 'react-router-dom'

function Player(props) {
  const videoRef = React.useRef(null)
  const location = useLocation()
  const [mediaAccessError, setMediaAccessError] = React.useState(null);

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
        if (t.kind === 'audio') {
          peerConnection.addTransceiver(t, {direction: 'sendonly'})
        } else {
          peerConnection.addTransceiver(t, {
            direction: 'sendonly',
            sendEncodings: [
              {
                rid: 'high'
              },
              {
                rid: 'med',
                scaleResolutionDownBy: 2.0
              },
              {
                rid: 'low',
                scaleResolutionDownBy: 4.0
              }
            ]
          })
        }
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
    }, setMediaAccessError)

    return function cleanup() {
      peerConnection.close()
      if (stream !== null) {
        stream.getTracks().forEach(t => t.stop())
      }
    }
  }, [videoRef, location.pathname])

  return (
    <div className='container mx-auto'>
      {mediaAccessError != null && <MediaAccessError>{mediaAccessError}</MediaAccessError>}
      <video
        ref={videoRef}
        autoPlay
        muted
        controls
        playsInline
        className='w-full h-full'
      />
    </div>
  )
}

const mediaErrorMessages = {
  NotAllowedError: `You can't publish stream using your camera, because you have blocked access to it 😞`,
  NotFoundError: `Seems like you don't have camera 😭 Or you just blocked access to it...\n` +
    `Check camera settings, browser permissions and system permissions.`,
}

function MediaAccessError({ children: error }) {
  return (
    <p className={'bg-red-700 text-white text-lg ' +
      'text-center p-5 rounded-t-lg whitespace-pre-wrap'
    }>
      {mediaErrorMessages[error.name] ?? 'Could not access your media device:\n' + error}
    </p>
  )
}

export default Player
