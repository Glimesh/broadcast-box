import React from 'react'
import {parseLinkHeader} from '@web3-storage/parse-link-header'
import { useLocation } from 'react-router-dom'

function Player(props) {
  const videoRef = React.createRef()
  const location = useLocation()
  const [videoLayers, setVideoLayers] = React.useState([]);
  const [mediaSrcObject, setMediaSrcObject] = React.useState(null);
  const [layerEndpoint, setLayerEndpoint] = React.useState('');

  const onLayerChange = event => {
    fetch(layerEndpoint, {
      method: 'POST',
      body: JSON.stringify({mediaId: '1', encodingId: event.target.value}),
      headers: {
        'Content-Type': 'application/json'
      }
    })
  }

  React.useEffect(() => {
    if (videoRef.current) {
      videoRef.current.srcObject = mediaSrcObject
    }
  }, [mediaSrcObject, videoRef])

  React.useEffect(() => {
    const peerConnection = new RTCPeerConnection() // eslint-disable-line

    peerConnection.ontrack = function (event) {
      setMediaSrcObject(event.streams[0])
    }

    peerConnection.addTransceiver('audio', {direction: 'recvonly'})
    peerConnection.addTransceiver('video', {direction: 'recvonly'})

    peerConnection.createOffer().then(offer => {
      peerConnection.setLocalDescription(offer)

      fetch(`${process.env.REACT_APP_API_PATH}/whep`, {
        method: 'POST',
        body: offer.sdp,
        headers: {
          Authorization: `Bearer ${location.pathname.substring(1)}`,
          'Content-Type': 'application/sdp'
        }
      }).then(r => {
        const parsedLinkHeader = parseLinkHeader(r.headers.get('Link'))
        setLayerEndpoint(`${window.location.protocol}//${parsedLinkHeader['urn:ietf:params:whep:ext:core:layer'].url}`)

        const evtSource = new EventSource(`${window.location.protocol}//${parsedLinkHeader['urn:ietf:params:whep:ext:core:server-sent-events'].url}`)
        evtSource.onerror = err => evtSource.close();

        evtSource.addEventListener("layers", event => {
          const parsed = JSON.parse(event.data)
          setVideoLayers(parsed['1']['layers'].map(l => l.encodingId))
        })


        return r.text()
      }).then(answer => {
        peerConnection.setRemoteDescription({
          sdp: answer,
          type: 'answer'
        })
      })
    })

    return function cleanup() {
      peerConnection.close()
    }
  }, [location.pathname])

  return (
    <>
      <video
        ref={videoRef}
        autoPlay
        muted
        controls
        playsInline
        className='mx-auto h-full'
      />

      {videoLayers.length >= 2 &&
        <select defaultValue="disabled" onChange={onLayerChange} className="appearance-none border w-full py-2 px-3 leading-tight focus:outline-none focus:shadow-outline bg-gray-700 border-gray-700 text-white rounded shadow-md placeholder-gray-200">
          <option value="disabled" disabled={true}>Choose Quality Level</option>
          {videoLayers.map(layer => {
            return <option key={layer} value={layer}>{layer}</option>
          })}
        </select>
      }
    </>
  )
}

export default Player
