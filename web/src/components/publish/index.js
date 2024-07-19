import React from 'react'
import linkHeader from 'http-link-header'
import { useLocation } from 'react-router-dom'

let mediaOptions = {
  audio: true,
  video: true
}

function Player(props) {
  const videoRef = React.useRef(null)
  const location = useLocation()
  const [mediaAccessError, setMediaAccessError] = React.useState(null);
  const [publishSuccess, setPublishSuccess] = React.useState(false);
  const [useDisplayMedia, setUseDisplayMedia] = React.useState(false);

  React.useEffect(() => {
    const peerConnection = new RTCPeerConnection() // eslint-disable-line
    let stream = null

    const mediaPromise = useDisplayMedia ?
      navigator.mediaDevices.getDisplayMedia(mediaOptions) :
      navigator.mediaDevices.getUserMedia(mediaOptions)

    mediaPromise.then(s => {
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
                rid: 'l',
                scaleResolutionDownBy: 4,
              },
              {
                rid: 'm',
                scaleResolutionDownBy: 2,
              },
              {
                rid: 'h'
              },
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
        return new Promise(resolve => {
          r.text().then(answer => {
            const parsedLinkHeader = new linkHeader(r.headers.get('Link'))

            let iceServers = parsedLinkHeader.refs
              .filter(l => l.rel === 'ice-server')
              .map(i => {
                i.urls = i.uri
                return i
              })

            if (iceServers.length !== 0) {
              peerConnection.setConfiguration({
                iceServers,
              })
              peerConnection.createOffer().then(offer => {
                peerConnection.setLocalDescription(offer)
                resolve(answer)
              })
            } else {
              resolve(answer)
            }
          })
        })
        }).then(answer => {
          peerConnection.setRemoteDescription({
            sdp: answer,
            type: 'answer'
          })
          setPublishSuccess(true)
        })
      })
    }, setMediaAccessError)

    return function cleanup() {
      peerConnection.close()
      if (stream !== null) {
        stream.getTracks().forEach(t => t.stop())
      }
    }
  }, [videoRef, useDisplayMedia, location.pathname])

  return (
    <div className='container mx-auto'>
      {mediaAccessError != null && <MediaAccessError>{mediaAccessError}</MediaAccessError>}
      {publishSuccess === true && <PublishSuccess />}
      <video
        ref={videoRef}
        autoPlay
        muted
        controls
        playsInline
        className='w-full h-full'
      />

      <button
        onClick={() => { setUseDisplayMedia(!useDisplayMedia)}}
        className="appearance-none border w-full mt-5 py-2 px-3 leading-tight focus:outline-none focus:shadow-outline bg-gray-700 border-gray-700 text-white rounded shadow-md placeholder-gray-200">
          {!useDisplayMedia && <> Publish Screen/Window/Tab instead </>}
          {useDisplayMedia && <> Publish Webcam instead </>}
      </button>
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

function PublishSuccess() {
  const subscribeUrl = window.location.href.replace('publish/', '')

  return (
    <p className={'bg-green-800 text-white text-lg ' +
      'text-center p-5 rounded-t-lg whitespace-pre-wrap'
    }>
      Live: Currently streaming to <a href={subscribeUrl} target="_blank" rel="noreferrer" className="hover:underline">{subscribeUrl}</a>
    </p>
  )
}

export default Player
