import React, { createRef, useEffect, useState } from 'react';
import { parseLinkHeader } from '@web3-storage/parse-link-header';
import { useLocation } from 'react-router-dom';
const Player = (props) => {
    const apiPath = import.meta.env.VITE_API_PATH;
    const { cinemaMode, setPeerConnectionDisconnected } = props;
    const location = useLocation();
    const videoRef = createRef();
    const [videoLayers, setVideoLayers] = useState([]);
    const [mediaSrcObject, setMediaSrcObject] = useState(null);
    const [layerEndpoint, setLayerEndpoint] = useState('');
    console.log("Sut");
    useEffect(() => {
        if (videoRef.current) {
            videoRef.current.srcObject = mediaSrcObject;
        }
    }, [mediaSrcObject, videoRef]);
    useEffect(() => {
        const peerConnection = new RTCPeerConnection();
        peerConnection.ontrack = function (event) {
            setMediaSrcObject(event.streams[0]);
        };
        peerConnection.oniceconnectionstatechange = () => {
            if (peerConnection.iceConnectionState === 'connected' || peerConnection.iceConnectionState === 'completed') {
                setPeerConnectionDisconnected(false);
            }
            else if (peerConnection.iceConnectionState === 'disconnected' || peerConnection.iceConnectionState === 'failed') {
                setPeerConnectionDisconnected(true);
            }
        };
        peerConnection.addTransceiver('audio', { direction: 'recvonly' });
        peerConnection.addTransceiver('video', { direction: 'recvonly' });
        peerConnection
            .createOffer()
            .then(offer => {
            offer["sdp"] = offer["sdp"].replace("useinbandfec=1", "useinbandfec=1;stereo=1");
            peerConnection.setLocalDescription(offer)
                .catch((err) => console.error(err));
            fetch(`${apiPath}/whep`, {
                method: 'POST',
                body: offer.sdp,
                headers: {
                    Authorization: `Bearer ${location.pathname.split('/').pop()}`,
                    'Content-Type': 'application/sdp'
                }
            }).then(r => {
                const parsedLinkHeader = parseLinkHeader(r.headers.get('Link'));
                if (parsedLinkHeader === null || parsedLinkHeader === undefined) {
                    throw new DOMException("Missing link header");
                }
                setLayerEndpoint(`${window.location.protocol}//${parsedLinkHeader['urn:ietf:params:whep:ext:core:layer'].url}`);
                const evtSource = new EventSource(`${window.location.protocol}//${parsedLinkHeader['urn:ietf:params:whep:ext:core:server-sent-events'].url}`);
                evtSource.onerror = _ => evtSource.close();
                evtSource.addEventListener("layers", event => {
                    const parsed = JSON.parse(event.data);
                    setVideoLayers(parsed['1']['layers'].map((layer) => layer.encodingId));
                });
                return r.text();
            }).then(answer => {
                peerConnection.setRemoteDescription({
                    sdp: answer,
                    type: 'answer'
                })
                    .catch((err) => console.error(err));
            });
        });
        return function cleanup() {
            peerConnection.close();
        };
    }, [location.pathname, setPeerConnectionDisconnected]);
    const onLayerChange = (event) => {
        fetch(layerEndpoint, {
            method: 'POST',
            body: JSON.stringify({ mediaId: '1', encodingId: event.target.value }),
            headers: {
                'Content-Type': 'application/json'
            }
        }).catch((err) => console.error(err));
    };
    return (React.createElement(React.Fragment, null,
        React.createElement("video", { ref: videoRef, autoPlay: true, muted: true, controls: true, playsInline: true, className: `bg-black w-full ${cinemaMode && "h-full"}`, style: cinemaMode ? {
                maxHeight: '100vh',
                maxWidth: '100vw'
            } : {} }),
        videoLayers.length >= 2 &&
            React.createElement("select", { defaultValue: "disabled", onChange: onLayerChange, className: "appearance-none border w-full py-2 px-3 leading-tight focus:outline-hidden focus:shadow-outline bg-gray-700 border-gray-700 text-white rounded-sm shadow-md placeholder-gray-200" },
                React.createElement("option", { value: "disabled", disabled: true }, "Choose Quality Level"),
                videoLayers.map(layer => React.createElement("option", { key: layer, value: layer }, layer)))));
};
export default Player;
