import React, {useContext, useEffect, useRef, useState} from 'react'
import {useLocation} from 'react-router-dom'
import {useNavigate} from 'react-router-dom'
import PlayerHeader from '../playerHeader/PlayerHeader';
import {StatusContext} from "../../providers/StatusProvider";
import {UsersIcon} from "@heroicons/react/20/solid";

const mediaOptions = {
	audio: true,
	video: {
		width: { ideal: 1920 },
		height: { ideal: 1080 },
	},
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
	const location = useLocation()
	const navigate = useNavigate();
	const streamKey = location.pathname.split('/').pop()
	const { streamStatus } = useContext(StatusContext);
	const [mediaAccessError, setMediaAccessError] = useState<ErrorMessageEnum | null>(null)
	const [publishSuccess, setPublishSuccess] = useState(false)
	const [useDisplayMedia, setUseDisplayMedia] = useState<"Screen" | "Webcam" | "None">("None");
	const [peerConnectionDisconnected, setPeerConnectionDisconnected] = useState(false)
	const [currentViewersCount, setCurrentViewersCount] = useState<number>(0)
	const [hasPacketLoss, setHasPacketLoss] = useState<boolean>(false)
	const [hasSignal, setHasSignal] = useState<boolean>(false);
	const [connectFailed, setConnectFailed] = useState<boolean>(false)

	const peerConnectionRef = useRef<RTCPeerConnection | null>(null);
	const videoRef = useRef<HTMLVideoElement>(null)
	const hasSignalRef = useRef<boolean>(false);
	const badSignalCountRef = useRef<number>(10);

	const apiPath = import.meta.env.VITE_API_PATH;

	const endStream = () => {
		navigate('/')
	}
	
	useEffect(() => {
		peerConnectionRef.current = new RTCPeerConnection();
		
		return () => peerConnectionRef.current?.close()
	}, [])

	useEffect(() => {
		if(!streamKey || !streamStatus){
			return;
		}

		const sessions = streamStatus.filter((session) => session.streamKey === streamKey);

		if(sessions.length !== 0){
			setCurrentViewersCount(() => 
				sessions.length !== 0 
					? sessions[0].whepSessions.length
					: 0)
		}
	}, [streamStatus]);

	useEffect(() => {
		if (useDisplayMedia === "None" || !peerConnectionRef.current) {
			return;
		}

		let stream: MediaStream | undefined = undefined;

		if (!navigator.mediaDevices) {
			setMediaAccessError(() => ErrorMessageEnum.NoMediaDevices);
			setUseDisplayMedia(() => "None")
			return
		}

		const isScreenShare = useDisplayMedia === "Screen"
		const mediaPromise = isScreenShare ?
			navigator.mediaDevices.getDisplayMedia(mediaOptions) :
			navigator.mediaDevices.getUserMedia(mediaOptions)

		mediaPromise.then(mediaStream => {
			if (peerConnectionRef.current!.connectionState === "closed") {
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
						peerConnectionRef.current!.addTransceiver(mediaStreamTrack, {
							direction: 'sendonly'
						})
					} else {
						peerConnectionRef.current!.addTransceiver(mediaStreamTrack, {
							direction: 'sendonly',
							sendEncodings: isScreenShare ? [] : [
								{
									rid: 'high',
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

			peerConnectionRef.current!.oniceconnectionstatechange = () => {
				if (peerConnectionRef.current!.iceConnectionState === 'connected' || peerConnectionRef.current!.iceConnectionState === 'completed') {
					setPublishSuccess(() => true)
					setMediaAccessError(() => null)
					setPeerConnectionDisconnected(() => false)
				} else if (peerConnectionRef.current!.iceConnectionState === 'disconnected' || peerConnectionRef.current!.iceConnectionState === 'failed') {
					setPublishSuccess(() => false)
					setPeerConnectionDisconnected(() => true)
				}
			}

			peerConnectionRef
				.current!
				.createOffer()
				.then(offer => {
					peerConnectionRef.current!.setLocalDescription(offer)
						.catch((err) => console.error("SetLocalDescription", err));

					fetch(`${apiPath}/whip`, {
						method: 'POST',
						body: offer.sdp,
						headers: {
							Authorization: `Bearer ${streamKey}`,
							'Content-Type': 'application/sdp'
						}
					}).then(r => {
						setConnectFailed(r.status !== 201)
						if (connectFailed) {
							throw new DOMException("WHIP endpoint did not return 201");
						}

						return r.text()
					})
					.then(answer => {
						peerConnectionRef.current!.setRemoteDescription({
							sdp: answer,
							type: 'answer'
						})
						.catch((err) => console.error("SetRemoveDescription", err))
					})
				})
		}, (reason: ErrorMessageEnum) => {
			setMediaAccessError(() => reason)
			setUseDisplayMedia("None");
		})

		return () => {
			peerConnectionRef.current?.close()
			if (stream) {
				stream
					.getTracks()
					.forEach((streamTrack: MediaStreamTrack) => streamTrack.stop())
			}
		}
	}, [videoRef, useDisplayMedia, location.pathname])

	useEffect(() => {
		hasSignalRef.current = hasSignal;

		const intervalHandler = () => {
			let senderHasPacketLoss = false;
			peerConnectionRef.current?.getSenders().forEach(sender => {
				if (sender) {
					sender.getStats()
						.then(stats => {
							stats.forEach(report => {
									if (report.type === "outbound-rtp") {
										senderHasPacketLoss = report.totalPacketSendDelay > 10;
									}
									if (report.type === "candidate-pair") {
										const signalIsValid = report.availableIncomingBitrate !== undefined;
										badSignalCountRef.current = signalIsValid ? 0 : badSignalCountRef.current + 1;

										if (badSignalCountRef.current > 2) {
											setHasSignal(() => false);
										} else if (badSignalCountRef.current === 0 && !hasSignalRef.current) {
											setHasSignal(() => true);
										}
									}
								}
							)
						})
				}
			})

			setHasPacketLoss(() => senderHasPacketLoss);
		}

		const interval = setInterval(intervalHandler, hasSignal ? 15_000 : 2_500)

		return () => {
			clearInterval(interval);
		}
	}, [hasSignal]);

	return (
		<div className='container mx-auto'>
			{mediaAccessError != null && <PlayerHeader headerType={"Error"}> {getMediaErrorMessage(mediaAccessError)} </PlayerHeader>}
			{peerConnectionDisconnected && <PlayerHeader headerType={"Error"}> WebRTC has disconnected or failed to connect at all 😭 </PlayerHeader>}
			{connectFailed && <PlayerHeader headerType={"Error"}> Failed to start Broadcast Box session 👮 </PlayerHeader>}
			{hasPacketLoss && <PlayerHeader headerType={"Warning"}> WebRTC is experiencing packet loss</PlayerHeader>}
			{publishSuccess && <PlayerHeader headerType={"Success"}> Live: Currently streaming to <a href={window.location.href.replace('publish/', '')} target="_blank" rel="noreferrer" className="hover:underline">{window.location.href.replace('publish/', '')}</a> </PlayerHeader>}

			<video
				ref={videoRef}
				autoPlay
				muted
				controls
				playsInline
				className='w-full h-full'
			/>
			
			<div className={"justify-items-end"} >
				<div className={"flex flex-row items-center"}>
					<UsersIcon className={"size-4"}/>
					{currentViewersCount}
				</div>
			</div>
			<div className="flex flex-row gap-2">
				<button
					onClick={() => setUseDisplayMedia("Screen")}
					className="appearance-none border w-full mt-5 py-2 px-3 leading-tight focus:outline-hidden focus:shadow-outline bg-blue-900 hover:bg-blue-800 border-gray-700 text-white rounded-sm shadow-md placeholder-gray-200">
					Publish Screen/Window/Tab
				</button>
				<button
					onClick={() => setUseDisplayMedia("Webcam")}
					className="appearance-none border w-full mt-5 py-2 px-3 leading-tight focus:outline-hidden focus:shadow-outline bg-blue-900 hover:bg-blue-800 border-gray-700 text-white rounded-sm shadow-md placeholder-gray-200">
					Publish Webcam
				</button>
			</div>

			{publishSuccess && (
				<div>
					<button
						onClick={endStream}
						className="appearance-none border w-full mt-5 py-2 px-3 leading-tight focus:outline-hidden focus:shadow-outline bg-red-900 hover:bg-red-800 border-gray-700 text-white rounded-sm shadow-md placeholder-gray-200">
						End stream
					</button>
				</div>
			)}
		</div>
	)
}

export default BrowserBroadcaster