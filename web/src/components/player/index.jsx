import React, { useContext, useEffect, useMemo, useState } from 'react'
import { parseLinkHeader } from '@web3-storage/parse-link-header'
import { useLocation, useSearchParams } from 'react-router-dom'
import ErrorHeader from '../error-header'
import Chat from '../chat'

export const CinemaModeContext = React.createContext(null);

export function CinemaModeProvider({ children }) {
  const [searchParams] = useSearchParams();
  const cinemaModeInUrl = searchParams.get("cinemaMode") === "true"
  const [cinemaMode, setCinemaMode] = useState(() => cinemaModeInUrl || localStorage.getItem("cinema-mode") === "true")

  const state = useMemo(() => ({
    cinemaMode,
    setCinemaMode,
    toggleCinemaMode: () => setCinemaMode((prev) => !prev),
  }), [cinemaMode, setCinemaMode]);

  useEffect(() => localStorage.setItem("cinema-mode", cinemaMode), [cinemaMode]);
  return (
    <CinemaModeContext.Provider value={state}>
      {children}
    </CinemaModeContext.Provider>
  );
}

function PlayerPage() {
  const { cinemaMode, toggleCinemaMode } = useContext(CinemaModeContext);
  const [peerConnectionDisconnected, setPeerConnectionDisconnected] = React.useState(false)

  const location = useLocation()
  const streamKey = location.pathname.split('/').pop()
  
  return (
    <>
      {peerConnectionDisconnected && <ErrorHeader> WebRTC has disconnected or failed to connect at all 😭 </ErrorHeader>}
      <div className={`flex flex-col items-center ${!cinemaMode && 'mx-auto px-2 py-2 container'}`}>
        <div className={`w-full ${cinemaMode ? 'flex flex-col' : 'grid grid-cols-1 lg:grid-cols-3 gap-4'}`}>
          <div className={cinemaMode ? 'w-full' : 'lg:col-span-2'}>
            <Player 
              cinemaMode={cinemaMode} 
              peerConnectionDisconnected={peerConnectionDisconnected} 
              setPeerConnectionDisconnected={setPeerConnectionDisconnected} 
            />
            <button className='bg-blue-900 px-4 py-2 rounded-lg mt-6' onClick={toggleCinemaMode}>
              {cinemaMode ? "Disable cinema mode" : "Enable cinema mode"}
            </button>
          </div>
          
          {!cinemaMode && (
            <div className="h-[500px] lg:h-[600px]">
              <Chat roomName={streamKey} />
            </div>
          )}
        </div>
      </div>
    </>
  )
}

function Player({ cinemaMode, peerConnectionDisconnected, setPeerConnectionDisconnected }) {
  const videoRef = React.createRef()
  const location = useLocation()
  const [videoLayers, setVideoLayers] = React.useState([]);
  const [mediaSrcObject, setMediaSrcObject] = React.useState(null);
  const [layerEndpoint, setLayerEndpoint] = React.useState('');

  const onLayerChange = event => {
    fetch(layerEndpoint, {
      method: 'POST',
      body: JSON.stringify({ mediaId: '1', encodingId: event.target.value }),
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

    peerConnection.oniceconnectionstatechange = () => {
      if (peerConnection.iceConnectionState === 'connected' || peerConnection.iceConnectionState === 'completed') {
        setPeerConnectionDisconnected(false)
      } else if (peerConnection.iceConnectionState === 'disconnected' ||  peerConnection.iceConnectionState === 'failed') {
        setPeerConnectionDisconnected(true)
      }
    }

    peerConnection.addTransceiver('audio', { direction: 'recvonly' })
    peerConnection.addTransceiver('video', { direction: 'recvonly' })

    peerConnection.createOffer().then(offer => {
      offer["sdp"] = offer["sdp"].replace("useinbandfec=1", "useinbandfec=1;stereo=1")
      peerConnection.setLocalDescription(offer)

      // Get API path from environment variables, with fallback to deprecated REACT_APP_API_PATH
      const apiPath = import.meta.env.VITE_API_PATH ?? (() => {
        console.warn('[broadcast box] REACT_APP_API_PATH is deprecated, please use VITE_API_PATH instead');
        return import.meta.env.REACT_APP_API_PATH;
      })();

      // For API calls, always use dynamic URL construction for consistent behavior with WebSockets
      let fetchUrl;
      if (apiPath && (apiPath.startsWith('http://') || apiPath.startsWith('https://'))) {
        // Use the full URL from environment
        fetchUrl = `${apiPath}/whep`;
      } else if (apiPath) {
        // It's just a path, use with current host
        fetchUrl = `${window.location.protocol}//${window.location.host}${apiPath}/whep`;
      } else {
        // No API path, just use current host with /api prefix
        fetchUrl = `${window.location.protocol}//${window.location.host}/api/whep`;
      }
      
      console.log('Fetching from:', fetchUrl);
      
      fetch(fetchUrl, {
        method: 'POST',
        body: offer.sdp,
        headers: {
          Authorization: `Bearer ${location.pathname.split('/').pop()}`,
          'Content-Type': 'application/sdp'
        }
      }).then(r => {
        const linkHeader = r.headers.get('Link');
        if (!linkHeader) {
          console.error('WHEP server did not return Link header');
          throw new Error('WHEP server did not return Link header');
        }
        
        const parsedLinkHeader = parseLinkHeader(linkHeader);
        
        // Check for required endpoints
        if (!parsedLinkHeader['urn:ietf:params:whep:ext:core:layer'] || 
            !parsedLinkHeader['urn:ietf:params:whep:ext:core:server-sent-events']) {
          console.error('WHEP server returned incomplete Link header', parsedLinkHeader);
          throw new Error('WHEP server returned incomplete Link header');
        }
        
        // Set layer endpoint
        setLayerEndpoint(`${window.location.protocol}//${parsedLinkHeader['urn:ietf:params:whep:ext:core:layer'].url}`);

        // Create EventSource for layers
        const evtSource = new EventSource(`${window.location.protocol}//${parsedLinkHeader['urn:ietf:params:whep:ext:core:server-sent-events'].url}`);
        evtSource.onerror = err => {
          console.error('EventSource error:', err);
          evtSource.close();
        };

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
  }, [location.pathname, setPeerConnectionDisconnected])

  return (
    <>
      <video
        ref={videoRef}
        autoPlay
        muted
        controls
        playsInline
        className={`bg-black w-full ${cinemaMode && "h-full"}`}
        style={cinemaMode ? {
          maxHeight: '100vh',
          maxWidth: '100vw'
        } : {}}
      />

      {videoLayers.length >= 2 &&
        <select defaultValue="disabled" onChange={onLayerChange} className="appearance-none border w-full py-2 px-3 leading-tight focus:outline-hidden focus:shadow-outline bg-gray-700 border-gray-700 text-white rounded-sm shadow-md placeholder-gray-200">
          <option value="disabled" disabled={true}>Choose Quality Level</option>
          {videoLayers.map(layer => {
            return <option key={layer} value={layer}>{layer}</option>
          })}
        </select>
      }
    </>
  )
}

export default PlayerPage
