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
      <div className={`flex flex-col items-center ${!cinemaMode && 'mx-auto px-2 py-2 w-full'}`}>
        <div className={`w-full ${cinemaMode ? 'flex flex-col' : 'grid grid-cols-1 lg:grid-cols-3 gap-2'}`}>
          <div className={cinemaMode ? 'w-full' : 'lg:col-span-2'}>
            <div className="bg-white overflow-hidden">
              <Player 
                cinemaMode={cinemaMode} 
                peerConnectionDisconnected={peerConnectionDisconnected} 
                setPeerConnectionDisconnected={setPeerConnectionDisconnected}
                onToggleCinemaMode={toggleCinemaMode}
              />
            </div>
          </div>
          
          {!cinemaMode && (
            <div className="lg:h-[calc(100vh-160px)] max-h-[500px]">
              <Chat roomName={streamKey} />
            </div>
          )}
        </div>
      </div>
    </>
  )
}

function Player({ cinemaMode, peerConnectionDisconnected, setPeerConnectionDisconnected, onToggleCinemaMode }) {
  const videoRef = React.createRef()
  const location = useLocation()
  const [videoLayers, setVideoLayers] = React.useState([]);
  const [mediaSrcObject, setMediaSrcObject] = React.useState(null);
  const [layerEndpoint, setLayerEndpoint] = React.useState('');
  const [isLoading, setIsLoading] = React.useState(true);

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
      if (mediaSrcObject) {
        setIsLoading(false);
      }
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
      }).catch(err => {
        console.error('Error connecting to stream:', err);
        setIsLoading(false);
      })
    })

    return function cleanup() {
      peerConnection.close()
    }
  }, [location.pathname, setPeerConnectionDisconnected])

  return (
    <div className="relative">
      {isLoading && (
        <div className="absolute inset-0 flex items-center justify-center bg-black/5 z-10">
          <div className="flex flex-col items-center">
            <svg className="animate-spin h-10 w-10 text-[var(--color-primary)] mb-3" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <p className="text-[var(--color-text-secondary)]">Connecting to stream...</p>
          </div>
        </div>
      )}
      
      <div className="relative">
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
        <button 
          onClick={onToggleCinemaMode}
          className="absolute top-2 right-2 p-1 bg-black/60 text-white z-10 hover:bg-black/80"
          title={cinemaMode ? "Exit Full Screen" : "Full Screen Mode"}
        >
          {cinemaMode ? (
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" className="w-5 h-5">
              <path fillRule="evenodd" d="M3.22 3.22a.75.75 0 011.06 0L12 10.94l7.72-7.72a.75.75 0 111.06 1.06L13.06 12l7.72 7.72a.75.75 0 11-1.06 1.06L12 13.06l-7.72 7.72a.75.75 0 01-1.06-1.06L10.94 12 3.22 4.28a.75.75 0 010-1.06z" />
            </svg>
          ) : (
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" className="w-5 h-5">
              <path fillRule="evenodd" d="M15 3.75a.75.75 0 0 1 .75-.75h4.5a.75.75 0 0 1 .75.75v4.5a.75.75 0 0 1-1.5 0V5.56l-3.97 3.97a.75.75 0 1 1-1.06-1.06l3.97-3.97h-2.69a.75.75 0 0 1-.75-.75Zm-12 0A.75.75 0 0 1 3.75 3h4.5a.75.75 0 0 1 0 1.5H5.56l3.97 3.97a.75.75 0 0 1-1.06 1.06L4.5 5.56v2.69a.75.75 0 0 1-1.5 0v-4.5Zm11.47 11.78a.75.75 0 1 1 1.06 1.06l-3.97 3.97h2.69a.75.75 0 0 1 0 1.5h-4.5a.75.75 0 0 1-.75-.75v-4.5a.75.75 0 0 1 1.5 0v2.69l3.97-3.97Zm-4.94-1.06a.75.75 0 0 1 0 1.06L5.56 19.5h2.69a.75.75 0 0 1 0 1.5h-4.5a.75.75 0 0 1-.75-.75v-4.5a.75.75 0 0 1 1.5 0v2.69l3.97-3.97a.75.75 0 0 1 1.06 0Z" clipRule="evenodd" />
            </svg>
          )}
        </button>
      </div>

      {videoLayers.length >= 2 && (
        <div className="p-4 bg-white border-t border-[var(--color-border)]">
          <div className="flex items-center">
            <span className="text-sm font-medium mr-3 text-[var(--color-text-secondary)]">Quality:</span>
            <select 
              defaultValue="disabled" 
              onChange={onLayerChange} 
              className="bg-white border border-[var(--color-border)] py-2 px-3 text-sm
                focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)] focus:border-transparent
                shadow-sm transition-all duration-200"
            >
              <option value="disabled" disabled={true}>Choose Quality Level</option>
              {videoLayers.map(layer => (
                <option key={layer} value={layer}>{layer}</option>
              ))}
            </select>
          </div>
        </div>
      )}
    </div>
  )
}

export default PlayerPage
