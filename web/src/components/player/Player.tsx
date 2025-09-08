import React, { useContext, useEffect, useRef, useState } from 'react'
import { parseLinkHeader } from '@web3-storage/parse-link-header'
import { ArrowsPointingOutIcon, Square2StackIcon } from "@heroicons/react/16/solid";
import VolumeComponent from "./components/VolumeComponent";
import PlayPauseComponent from "./components/PlayPauseComponent";
import VideoLayerSelectorComponent from "./components/VideoLayerSelectorComponent";
import AudioLayerSelectorComponent from "./components/AudioLayerSelectorComponent";
import CurrentViewersComponent from "./components/CurrentViewersComponent";
import { HeaderContext } from '../../providers/HeaderProvider';
import { StatusContext } from '../../providers/StatusProvider';
import { LocaleContext } from '../../providers/LocaleProvider';

interface PlayerProps {
	streamKey: string;
	cinemaMode: boolean;
	onCloseStream?: () => void;
}

interface CurrentLayersMessage {
	id: string,
	audioLayerCurrent: string
	audioTimestamp: number
	audioPacketsWritten: number
	audioSequenceNumber: number

	videoLayerCurrent: string
	videoTimestamp: number
	videoPacketsWritten: number
	videoSequenceNumber: number
}

const Player = (props: PlayerProps) => {
	const { streamKey, cinemaMode } = props;
	const { setTitle } = useContext(HeaderContext)
	const { locale } = useContext(LocaleContext)
	const { currentStreamStatus, setCurrentStreamStatus } = useContext(StatusContext)

	const [currentLayersStatus, setCurrentLayersStatus] = useState<CurrentLayersMessage | undefined>()
	const [audioLayers, setAudioLayers] = useState([]);
	const [videoLayers, setVideoLayers] = useState([]);
	const [hasPreparedPeerConnectionRef, setHasPreparedPeerConnectionRef] = useState<boolean>(false);
	const [hasSignal, setHasSignal] = useState<boolean>(false);
	const [hasPacketLoss, setHasPacketLoss] = useState<boolean>(false)
	const [videoOverlayVisible, setVideoOverlayVisible] = useState<boolean>(false)

	const videoRef = useRef<HTMLVideoElement>(null);
	const layerEndpointRef = useRef<string>('');
	const hasSignalRef = useRef<boolean>(false);
	const peerConnectionRef = useRef<RTCPeerConnection | null>(null);
	const videoOverlayVisibleTimeoutRef = useRef<number | undefined>(undefined);
	const clickDelay = 250;
	const lastClickTimeRef = useRef(0);
	const clickTimeoutRef = useRef<number | undefined>(undefined);
	const streamVideoPlayerId = streamKey + "_videoPlayer";

	const setHasSignalHandler = (_: Event) => {
		setHasSignal(() => true);
	}
	const resetTimer = (isVisible: boolean) => {
		setVideoOverlayVisible(() => isVisible);

		if (videoOverlayVisibleTimeoutRef) {
			clearTimeout(videoOverlayVisibleTimeoutRef.current)
		}

		videoOverlayVisibleTimeoutRef.current = setTimeout(() => {
			setVideoOverlayVisible(() => false)
		}, 2500)
	}

	const handleVideoPlayerClick = () => {
		lastClickTimeRef.current = Date.now();

		clickTimeoutRef.current = setTimeout(() => {
			const timeSinceLastClick = Date.now() - lastClickTimeRef.current;
			if (timeSinceLastClick >= clickDelay && (timeSinceLastClick - clickDelay) < 5000) {
				videoRef.current?.paused
					? videoRef.current?.play()
					: videoRef.current?.pause();
			}
		}, clickDelay);
	};
	const handleVideoPlayerDoubleClick = () => {
		clearTimeout(clickTimeoutRef.current);
		lastClickTimeRef.current = 0;
		videoRef.current?.requestFullscreen()
			.catch(err => console.error("VideoPlayer_RequestFullscreen", err));
	};

	useEffect(() => {
		const handleWindowBeforeUnload = () => {
			peerConnectionRef.current?.close();
			peerConnectionRef.current = null;
		}

		document.title = streamKey + " - Broadcast Box"
		const handleOverlayTimer = (isVisible: boolean) => resetTimer(isVisible)
		const player = document.getElementById(streamVideoPlayerId)

		player?.addEventListener('mousemove', () => handleOverlayTimer(true))
		player?.addEventListener('mouseenter', () => handleOverlayTimer(true))
		player?.addEventListener('mouseleave', () => handleOverlayTimer(false))
		player?.addEventListener('mouseup', () => handleOverlayTimer(true))

		window.addEventListener("beforeunload", handleWindowBeforeUnload)

		fetch(`/api/ice-servers`, {
			method: 'GET',
		}).then(r => r.json())
			.then((result) => {
				peerConnectionRef.current = new RTCPeerConnection({
					iceServers: result
				});
				setHasPreparedPeerConnectionRef(() => true)
			}).catch(() => {
				console.error("Error calling Ice-Servers endpoint. Ignoring STUN/TURN configuration")
				peerConnectionRef.current = new RTCPeerConnection();

				setHasPreparedPeerConnectionRef(() => true)
			})

		return () => {
			peerConnectionRef.current?.close()
			peerConnectionRef.current = null

			videoRef.current?.removeEventListener("playing", setHasSignalHandler)

			player?.removeEventListener('mouseenter', () => handleOverlayTimer)
			player?.removeEventListener('mouseleave', () => handleOverlayTimer)
			player?.removeEventListener('mousemove', () => handleOverlayTimer)
			player?.removeEventListener('mouseup', () => handleOverlayTimer)

			window.removeEventListener("beforeunload", handleWindowBeforeUnload)

			clearTimeout(videoOverlayVisibleTimeoutRef.current)
		}
	}, [])

	useEffect(() => {
		setTitle(currentStreamStatus?.streamKey ?? "")
	}, [currentStreamStatus])

	useEffect(() => {
		hasSignalRef.current = hasSignal;

		const intervalHandler = () => {
			if (!peerConnectionRef.current) {
				return
			}

			let receiversHasPacketLoss = false;
			peerConnectionRef.current
				.getReceivers()
				.forEach(receiver => {
					if (receiver) {
						receiver.getStats()
							.then(stats => {
								stats.forEach(report => {
									if (report.type === "inbound-rtp") {
										const lossRate = report.packetsLost / (report.packetsLost + report.packetsReceived);
										receiversHasPacketLoss = receiversHasPacketLoss ? true : lossRate > 5;
									}
								}
								)
							})
					}
				})

			setHasPacketLoss(() => receiversHasPacketLoss);
		}

		const interval = setInterval(intervalHandler, hasSignal ? 15_000 : 2_500)

		return () => clearInterval(interval);
	}, [hasSignal]);

	useEffect(() => {
		if (!peerConnectionRef.current) {
			return;
		}
		peerConnectionRef.current.ontrack = (event: RTCTrackEvent) => {
			if (videoRef.current) {
				videoRef.current.srcObject = event.streams[0];
				videoRef.current.addEventListener("playing", setHasSignalHandler)
			}
		}

		peerConnectionRef.current.addTransceiver('audio', { direction: 'recvonly' })
		peerConnectionRef.current.addTransceiver('video', { direction: 'recvonly' })

		peerConnectionRef.current
			.createOffer()
			.then(async offer => {
				offer["sdp"] = offer["sdp"]!.replace("useinbandfec=1", "useinbandfec=1;stereo=1")

				await peerConnectionRef.current!
					.setLocalDescription(offer)
					.catch((err) => console.error("SetLocalDescription", err));

				const whepResponse = await fetch(`/api/whep`, {
					method: 'POST',
					body: offer.sdp,
					headers: {
						Authorization: `Bearer ${streamKey}`,
						'Content-Type': 'application/sdp'
					}
				})

				const parsedLinkHeader = parseLinkHeader(whepResponse.headers.get('Link'))

				if (parsedLinkHeader === null || parsedLinkHeader === undefined) {
					throw new DOMException("Missing link header");
				}

				layerEndpointRef.current = `${parsedLinkHeader['urn:ietf:params:whep:ext:core:layer'].url}`
				const evtSource = new EventSource(`${parsedLinkHeader['urn:ietf:params:whep:ext:core:server-sent-events'].url}`)

				evtSource.onerror = _ => evtSource.close();

				// Receive current status of the whip stream
				evtSource.addEventListener("status", (event: MessageEvent) => setCurrentStreamStatus(JSON.parse(event.data)))

				// Receive current current layers of this whep stream
				evtSource.addEventListener("currentLayers", (event: MessageEvent) => setCurrentLayersStatus(() => JSON.parse(event.data)))

				// Receive layers
				evtSource.addEventListener("layers", event => {
					const parsed = JSON.parse(event.data)
					setVideoLayers(() => parsed['1']['layers'].map((layer: any) => layer.encodingId))
					setAudioLayers(() => parsed['2']['layers'].map((layer: any) => layer.encodingId))
				})

				const answer = await whepResponse.text()
				await peerConnectionRef.current!.setRemoteDescription({
					sdp: answer,
					type: 'answer'
				}).catch((err) => console.error("RemoteDescription", err))
			})
	}, [peerConnectionRef, hasPreparedPeerConnectionRef])

	return (
		<div
			id={streamVideoPlayerId}
			className="inline-block w-full relative z-0 aspect-video"
			style={cinemaMode ? {
				maxHeight: '100vh',
				maxWidth: '100vw',
			} : {}}>
			<div
				onClick={handleVideoPlayerClick}
				onDoubleClick={handleVideoPlayerDoubleClick}
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
						hover: ${videoOverlayVisible ? 'opacity-100' : 'opacity-0'}
						${!videoOverlayVisible ? 'cursor-none' : 'cursor-default'}
					`}
				`}
			>

				{/*Opaque background*/}
				<div className={`absolute w-full bg-gray-950 ${!hasSignal ? 'opacity-40' : 'opacity-0'} h-full bg-red-100`} />

				{/*Buttons */}
				{videoRef.current !== null && (
					<div className="absolute bottom-0 h-8 w-full flex place-items-end z-20">
						<div
							onClick={(e) => e.stopPropagation()}
							className="bg-blue-950 w-full flex flex-row gap-2 h-1/14 p-1 max-h-8 min-h-8">

							<PlayPauseComponent videoRef={videoRef} />

							<VolumeComponent
								isMuted={videoRef.current?.muted ?? false}
								onVolumeChanged={(newValue) => videoRef.current!.volume = newValue}
								onStateChanged={(newState) => videoRef.current!.muted = newState}
							/>

							<div className="w-full"></div>

							<CurrentViewersComponent currentViewersCount={currentStreamStatus?.viewers ?? 0} />
							<VideoLayerSelectorComponent layers={videoLayers} layerEndpoint={layerEndpointRef.current} hasPacketLoss={hasPacketLoss} currentLayer={currentLayersStatus?.videoLayerCurrent ?? ""} />
							{audioLayers.length > 1 && (
								<AudioLayerSelectorComponent layers={audioLayers} layerEndpoint={layerEndpointRef.current} hasPacketLoss={hasPacketLoss} currentLayer={currentLayersStatus?.videoLayerCurrent ?? ""} />
							)}
							<Square2StackIcon onClick={() => videoRef.current?.requestPictureInPicture()} />
							<ArrowsPointingOutIcon onClick={() => videoRef.current?.requestFullscreen()} />

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
						{props.streamKey} {locale.player.message_is_not_online}
					</h2>
				)}
				{videoLayers.length > 0 && !hasSignal && (
					<h2
						className="absolute animate-pulse w-full top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 font-light leading-tight text-4xl text-center">
						{locale.player.message_loading_video}
					</h2>
				)}

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
