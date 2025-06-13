import React, {useEffect, useRef, useState} from 'react'
import {parseLinkHeader} from '@web3-storage/parse-link-header'
import {ArrowsPointingOutIcon, Square2StackIcon} from "@heroicons/react/16/solid";
import VolumeComponent from "./components/VolumeComponent";
import PlayPauseComponent from "./components/PlayPauseComponent";
import QualitySelectorComponent from "./components/QualitySelectorComponent";

interface PlayerProps {
	streamKey: string;
	cinemaMode: boolean;
	onCloseStream?: () => void;
}

const Player = (props: PlayerProps) => {
	const apiPath = import.meta.env.VITE_API_PATH;
	const {streamKey, cinemaMode} = props;

	const videoRef = useRef<HTMLVideoElement>(null);
	const [videoLayers, setVideoLayers] = useState([]);
	const [peerConnection, setPeerConnection] = useState<RTCPeerConnection>();
	const [hasSignal, setHasSignal] = useState<boolean>(false);
	const [hasPacketLoss, setHasPacketLoss] = useState<boolean>(false)
	
	const layerEndpointRef = useRef<string>('');
	const hasSignalRef = useRef<boolean>(false);
	const peerRef = useRef(peerConnection);
	const badSignalCountRef = useRef<number>(10);

	useEffect(() => {
		hasSignalRef.current = hasSignal;

		const intervalHandler = () => {
			let receiversHasPacketLoss = false;
			peerRef.current?.getReceivers().forEach(receiver => {
				if (receiver) {
					receiver.getStats()
						.then(stats => {
							stats.forEach(report => {
									if (report.type === "inbound-rtp") {
										const lossRate = report.packetsLost / (report.packetsLost + report.packetsReceived);
										receiversHasPacketLoss = receiversHasPacketLoss ? true : lossRate > 5;
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
			
			setHasPacketLoss(() => receiversHasPacketLoss);
		}

		const interval = setInterval(intervalHandler, hasSignal ? 15_000 : 2_500)

		return () => {
			clearInterval(interval);
		}
	}, [hasSignal]);

	useEffect(() => {
		if (!peerConnection && !!videoRef.current) {
			setPeerConnection(() => new RTCPeerConnection());
		}
	}, [videoRef])

	useEffect(() => {
		if (!peerConnection) {
			return;
		}

		peerRef.current = peerConnection;
		peerConnection.ontrack = (event: RTCTrackEvent) => {
			if (videoRef.current) {
				videoRef.current.srcObject = event.streams[0];
			}
		}

		peerConnection.addTransceiver('audio', {direction: 'recvonly'})
		peerConnection.addTransceiver('video', {direction: 'recvonly'})

		peerConnection
			.createOffer()
			.then(offer => {
				offer["sdp"] = offer["sdp"]!.replace("useinbandfec=1", "useinbandfec=1;stereo=1")

				peerConnection.setLocalDescription(offer)
					.catch((err) => console.error("SetLocalDescription", err));

				fetch(`${apiPath}/whep`, {
					method: 'POST',
					body: offer.sdp,
					headers: {
						Authorization: `Bearer ${streamKey}`,
						'Content-Type': 'application/sdp'
					}
				}).then(r => {
					const parsedLinkHeader = parseLinkHeader(r.headers.get('Link'))

					if (parsedLinkHeader === null || parsedLinkHeader === undefined) {
						throw new DOMException("Missing link header");
					}

					layerEndpointRef.current = `${window.location.protocol}//${parsedLinkHeader['urn:ietf:params:whep:ext:core:layer'].url}`

					const evtSource = new EventSource(`${window.location.protocol}//${parsedLinkHeader['urn:ietf:params:whep:ext:core:server-sent-events'].url}`)
					evtSource.onerror = _ => evtSource.close();

					evtSource.addEventListener("layers", event => {
						const parsed = JSON.parse(event.data)
						setVideoLayers(() => parsed['1']['layers'].map((layer: any) => layer.encodingId))
					})

					return r.text()
				}).then(answer => {
					peerConnection.setRemoteDescription({
						sdp: answer,
						type: 'answer'
					}).catch((err) => console.error("RemoteDescription", err))
				}).catch((err) => {
					console.error("PeerConnectionError", err)
				})
			})

		return function cleanup() {
			peerConnection.close()
		}
	}, [peerConnection])

	return (
		<div
			className="inline-block w-full relative"
			style={cinemaMode ? {
				maxHeight: '100vh',
				maxWidth: '100vw'
			} : {}}>
			<div
				className={`
					absolute
					rounded-md
					w-full
					h-full
					z-10
					${!hasSignal && "bg-gray-800"}
					${hasSignal && `
						transition-opacity
						duration-500
						opacity-0
						hover:opacity-100
					`}
				`}
			>

				{/*Opaque background*/}
				<div
					onDoubleClick={() => videoRef.current?.requestFullscreen()}
					className="absolute w-full bg-gray-950 opacity-40 h-full"/>

				{/*Buttons */}
				{videoRef.current !== null && (
					<div className="absolute h-full w-full flex place-items-end">
						<div className="bg-blue-950 w-full flex flex-row gap-2 h-1/14 rounded-b-md p-1 max-h-8 min-h-8">

							<PlayPauseComponent videoRef={videoRef}/>

							<VolumeComponent
								isMuted={videoRef.current?.muted ?? false}
								onVolumeChanged={(newValue) => videoRef.current!.volume = newValue}
								onStateChanged={(newState) => videoRef.current!.muted = newState}
							/>

							<div className="w-full"></div>

							<QualitySelectorComponent layers={videoLayers} layerEndpoint={layerEndpointRef.current} hasPacketLoss={hasPacketLoss}/>
							<Square2StackIcon onClick={() => videoRef.current?.requestPictureInPicture()}/>
							<ArrowsPointingOutIcon onClick={() => videoRef.current?.requestFullscreen()}/>

						</div>
					</div>)}

				{!!props.onCloseStream && (
					<button
						onClick={props.onCloseStream}
						className="absolute top-2 right-2 p-2 rounded-full bg-red-400 hover:bg-red-500 pointer-events-auto">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							className="h-6 w-6 text-gray-700"
							viewBox="0 0 24 24"
							fill="black"
						>
							<path
								fillRule="evenodd"
								d="M6.225 6.225a.75.75 0 011.06 0L12 10.94l4.715-4.715a.75.75 0 111.06 1.06L13.06 12l4.715 4.715a.75.75 0 11-1.06 1.06L12 13.06l-4.715 4.715a.75.75 0 11-1.06-1.06L10.94 12 6.225 7.285a.75.75 0 010-1.06z"
								clipRule="evenodd"
							/>
						</svg>
					</button>
				)}

				{videoLayers.length === 0 && !hasSignal && (
					<h2
						className="absolute w-full top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 font-light leading-tight text-4xl text-center">
						{props.streamKey} is not currently streaming
					</h2>
				)}
				{
					videoLayers.length > 0 && !hasSignal && (
						<h2
							className="absolute animate-pulse w-full top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 font-light leading-tight text-4xl text-center">
							Loading video
						</h2>)
				}

			</div>

			<video
				ref={videoRef}
				autoPlay
				muted
				playsInline
				className="bg-transparent rounded-md w-full h-full relative"
			/>
		</div>
	)
}

export default Player