import React, { useEffect, useRef, useState } from 'react';
import { useLocation } from 'react-router-dom';
import ErrorHeader from '../error-header/errorHeader';
const mediaOptions = {
    audio: true,
    video: true
};
var ErrorMessageEnum;
(function (ErrorMessageEnum) {
    ErrorMessageEnum[ErrorMessageEnum["NoMediaDevices"] = 0] = "NoMediaDevices";
    ErrorMessageEnum[ErrorMessageEnum["NotAllowedError"] = 1] = "NotAllowedError";
    ErrorMessageEnum[ErrorMessageEnum["NotFoundError"] = 2] = "NotFoundError";
})(ErrorMessageEnum || (ErrorMessageEnum = {}));
function getMediaErrorMessage(value) {
    switch (value) {
        case ErrorMessageEnum.NoMediaDevices:
            return `MediaDevices API was not found. Publishing in Broadcast Box requires HTTPS 👮`;
        case ErrorMessageEnum.NotFoundError:
            return `Seems like you don't have camera 😭 Or you just blocked access to it...\nCheck camera settings, browser permissions and system permissions.`;
        case ErrorMessageEnum.NotAllowedError:
            return `You can't publish stream using your camera, because you have blocked access to it 😞`;
        default:
            return "Could not access your media device";
    }
}
function BrowserBroadcaster() {
    const videoRef = useRef(null);
    const location = useLocation();
    const [mediaAccessError, setMediaAccessError] = useState(null);
    const [publishSuccess, setPublishSuccess] = useState(false);
    const [useDisplayMedia, setUseDisplayMedia] = useState(false);
    const [peerConnectionDisconnected, setPeerConnectionDisconnected] = useState(false);
    const apiPath = import.meta.env.VITE_API_PATH;
    useEffect(() => {
        const peerConnection = new RTCPeerConnection();
        let stream = undefined;
        if (!navigator.mediaDevices) {
            setMediaAccessError(ErrorMessageEnum.NoMediaDevices);
            return;
        }
        const mediaPromise = useDisplayMedia ?
            navigator.mediaDevices.getDisplayMedia(mediaOptions) :
            navigator.mediaDevices.getUserMedia(mediaOptions);
        mediaPromise.then(mediaStream => {
            if (peerConnection.connectionState === "closed") {
                mediaStream
                    .getTracks()
                    .forEach(mediaStreamTrack => mediaStreamTrack.stop());
                return;
            }
            stream = mediaStream;
            videoRef.current.srcObject = mediaStream;
            mediaStream
                .getTracks()
                .forEach(mediaStreamTrack => {
                if (mediaStreamTrack.kind === 'audio') {
                    peerConnection.addTransceiver(mediaStreamTrack, {
                        direction: 'sendonly'
                    });
                }
                else {
                    peerConnection.addTransceiver(mediaStreamTrack, {
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
                    });
                }
            });
            peerConnection.oniceconnectionstatechange = () => {
                if (peerConnection.iceConnectionState === 'connected' || peerConnection.iceConnectionState === 'completed') {
                    setPublishSuccess(true);
                    setPeerConnectionDisconnected(false);
                }
                else if (peerConnection.iceConnectionState === 'disconnected' || peerConnection.iceConnectionState === 'failed') {
                    setPublishSuccess(false);
                    setPeerConnectionDisconnected(true);
                }
            };
            peerConnection
                .createOffer()
                .then(offer => {
                peerConnection.setLocalDescription(offer)
                    .catch((err) => console.error(err));
                fetch(`${apiPath}/whip`, {
                    method: 'POST',
                    body: offer.sdp,
                    headers: {
                        Authorization: `Bearer ${location.pathname.split('/').pop()}`,
                        'Content-Type': 'application/sdp'
                    }
                }).then(r => r.text())
                    .then(answer => {
                    peerConnection.setRemoteDescription({
                        sdp: answer,
                        type: 'answer'
                    })
                        .catch((err) => console.error(err));
                });
            });
        }, setMediaAccessError);
        return function cleanup() {
            peerConnection.close();
            if (stream) {
                stream
                    .getTracks()
                    .forEach((streamTrack) => streamTrack.stop());
            }
        };
    }, [videoRef, useDisplayMedia, location.pathname]);
    return (React.createElement("div", { className: 'container mx-auto' },
        mediaAccessError != null && React.createElement(ErrorHeader, null,
            " ",
            getMediaErrorMessage(mediaAccessError),
            " "),
        peerConnectionDisconnected && React.createElement(ErrorHeader, null, " WebRTC has disconnected or failed to connect at all \uD83D\uDE2D "),
        publishSuccess && React.createElement(PublishSuccess, null),
        React.createElement("video", { ref: videoRef, autoPlay: true, muted: true, controls: true, playsInline: true, className: 'w-full h-full' }),
        React.createElement("button", { onClick: () => setUseDisplayMedia(!useDisplayMedia), className: "appearance-none border w-full mt-5 py-2 px-3 leading-tight focus:outline-hidden focus:shadow-outline bg-gray-700 border-gray-700 text-white rounded-sm shadow-md placeholder-gray-200" },
            !useDisplayMedia && React.createElement(React.Fragment, null, " Publish Screen/Window/Tab instead "),
            useDisplayMedia && React.createElement(React.Fragment, null, " Publish Webcam instead "))));
}
function PublishSuccess() {
    const subscribeUrl = window.location.href.replace('publish/', '');
    return (React.createElement("p", { className: 'bg-green-800 text-white text-lg text-center p-5 rounded-t-lg whitespace-pre-wrap' },
        "Live: Currently streaming to ",
        React.createElement("a", { href: subscribeUrl, target: "_blank", rel: "noreferrer", className: "hover:underline" }, subscribeUrl)));
}
export default BrowserBroadcaster;
