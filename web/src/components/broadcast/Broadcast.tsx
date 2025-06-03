import React, {useEffect, useRef, useState} from 'react'
import {useLocation} from 'react-router-dom'
import ErrorHeader from '../error-header/errorHeader'

const mediaOptions = {
	audio: true,
	video: true
}

enum ErrorMessageEnum {
	NoMediaDevices,
	NotAllowedError,
	NotFoundError
}

function getMediaErrorMessage(value: ErrorMessageEnum): string {
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
	const videoRef = useRef<HTMLVideoElement>(null)
	const location = useLocation()
	const [mediaAccessError, setMediaAccessError] = useState<ErrorMessageEnum | null>(null)
	const [publishSuccess, setPublishSuccess] = useState(false)
	const [useDisplayMedia, setUseDisplayMedia] = useState(false)
	const [peerConnectionDisconnected, setPeerConnectionDisconnected] = useState(false)

	const apiPath = import.meta.env.VITE_API_PATH;

	useEffect(() => {
		const peerConnection = new RTCPeerConnection()
		let stream: MediaStream | undefined = undefined;

		if (!navigator.mediaDevices) {
			setMediaAccessError(ErrorMessageEnum.NoMediaDevices);
			return
		}

		const mediaPromise = useDisplayMedia ?
			navigator.mediaDevices.getDisplayMedia(mediaOptions) :
			navigator.mediaDevices.getUserMedia(mediaOptions)

		mediaPromise.then(mediaStream => {
			if (peerConnection.connectionState === "closed") {
				mediaStream
					.getTracks()
					.forEach(mediaStreamTrack => mediaStreamTrack.stop())

				return;
			}

			stream = mediaStream
			videoRef.current!.srcObject = mediaStream

			mediaStream
				.getTracks()
				.forEach(mediaStreamTrack => {
					if (mediaStreamTrack.kind === 'audio') {
						peerConnection.addTransceiver(mediaStreamTrack, {
							direction: 'sendonly'
						})
					} else {
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
						})
					}
				})

			peerConnection.oniceconnectionstatechange = () => {
				if (peerConnection.iceConnectionState === 'connected' || peerConnection.iceConnectionState === 'completed') {
					setPublishSuccess(true)
					setPeerConnectionDisconnected(false)
				} else if (peerConnection.iceConnectionState === 'disconnected' || peerConnection.iceConnectionState === 'failed') {
					setPublishSuccess(false)
					setPeerConnectionDisconnected(true)
				}
			}

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
							.catch((err) => console.error(err))
						})
				})
		}, setMediaAccessError)

		return function cleanup() {
			peerConnection.close()
			if (stream) {
				stream
					.getTracks()
					.forEach((streamTrack: MediaStreamTrack) => streamTrack.stop())
			}
		}
	}, [videoRef, useDisplayMedia, location.pathname])

	return (
		<div className='container mx-auto'>
			{mediaAccessError != null && <ErrorHeader> {getMediaErrorMessage(mediaAccessError)} </ErrorHeader>}
			{peerConnectionDisconnected && <ErrorHeader> WebRTC has disconnected or failed to connect at all 😭 </ErrorHeader>}
			{publishSuccess && <PublishSuccess/>}

			<video
				ref={videoRef}
				autoPlay
				muted
				controls
				playsInline
				className='w-full h-full'
			/>

			<button
				onClick={() => setUseDisplayMedia(!useDisplayMedia)}
				className="appearance-none border w-full mt-5 py-2 px-3 leading-tight focus:outline-hidden focus:shadow-outline bg-gray-700 border-gray-700 text-white rounded-sm shadow-md placeholder-gray-200">
				{!useDisplayMedia && <> Publish Screen/Window/Tab instead </>}
				{useDisplayMedia && <> Publish Webcam instead </>}
			</button>
		</div>
	)
}

function PublishSuccess() {
	const subscribeUrl = window.location.href.replace('publish/', '')

	return (
		<p className={'bg-green-800 text-white text-lg text-center p-5 rounded-t-lg whitespace-pre-wrap'}>
			Live: Currently streaming to <a href={subscribeUrl} target="_blank" rel="noreferrer" className="hover:underline">{subscribeUrl}</a>
		</p>
	)
}

export default BrowserBroadcaster