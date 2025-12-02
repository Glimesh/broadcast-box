import React, { useEffect, useRef, useState } from 'react'
import PlayPauseComponent from "./components/PlayPauseComponent";
import VideoLayerSelectorComponent from "./components/VideoLayerSelectorComponent";
import AudioLayerSelectorComponent from "./components/AudioLayerSelectorComponent";
import CurrentViewersComponent from "./components/CurrentViewersComponent";
import { StreamStatus } from '../../providers/StatusProvider';
import { CurrentLayersMessage, PeerConnectionSetup, SetupPeerConnectionProps } from './functions/peerconnection';
import { ArrowsPointingOutIcon, Square2StackIcon } from '@heroicons/react/20/solid';
import VolumeComponent from './components/VolumeComponent';
import { StatusMessageComponent } from './components/StatusMessageComponent';
import { StreamMOTD } from './components/StreamMOTD';

interface PlayerProps {
	streamKey: string;
	cinemaMode: boolean;
	onCloseStream?: () => void;
}

const Player = (props: PlayerProps) => {
	const { cinemaMode } = props;
	const streamKey = decodeURIComponent(props.streamKey).replace(' ', '_')

	const [currentStreamStatus, setCurrentStreamStatus] = useState<StreamStatus>({
		streamKey: streamKey,
		motd: "",
		viewers: 0,
		isOnline: false,
	})

	const [currentLayersStatus, setCurrentLayersStatus] = useState<CurrentLayersMessage | undefined>()
	const [audioLayers, setAudioLayers] = useState([]);
	const [videoLayers, setVideoLayers] = useState([]);
	const [streamState, setStreamState] = useState<"Loading" | "Playing" | "Offline" | "Error">("Loading");
	const [videoOverlayVisible, setVideoOverlayVisible] = useState<boolean>(false)

	const clickDelay = 250;
	const videoRef = useRef<HTMLVideoElement>(null);
	const layerEndpointRef = useRef<string>('');
	const videoOverlayVisibleTimeoutRef = useRef<number | undefined>(undefined);
	const lastClickTimeRef = useRef(0);
	const clickTimeoutRef = useRef<number | undefined>(undefined);
	const streamVideoPlayerId = streamKey + "_videoPlayer";

	const peerConnectionConfig: SetupPeerConnectionProps = {
		streamKey: streamKey,
		videoRef: videoRef,
		layerEndpointRef: layerEndpointRef,
		onStateChange: (state) => console.log("PeerConnectionState.Change", state),
		onStreamRestart: () => console.log("PeerConnection.onStreamRestart: Missing setup"),
		onAudioLayerChange: (layers) => setAudioLayers(layers),
		onVideoLayerChange: (layers) => setVideoLayers(layers),
		onLayerStatus: (status) => setCurrentLayersStatus(status),
		onStreamStatus: (status) => {
			if(!status.isOnline){
				setStreamState("Offline")
			}
			setCurrentStreamStatus(status)
		},
		onError: (error) => {
			console.log("PeerConnection.Error", error)
			setStreamState("Error")
		},
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
		const handleOverlayTimer = (isVisible: boolean) => resetTimer(isVisible)

		const player = document.getElementById(streamVideoPlayerId)
		player?.addEventListener('mouseup', () => handleOverlayTimer(true))
		player?.addEventListener('mousemove', () => handleOverlayTimer(true))
		player?.addEventListener('mouseenter', () => handleOverlayTimer(true))
		player?.addEventListener('mouseleave', () => handleOverlayTimer(false))

		peerConnectionConfig.onStreamRestart = () => PeerConnectionSetup(peerConnectionConfig)
		PeerConnectionSetup(peerConnectionConfig)
			.then((peerConnection) => {
				window.addEventListener("beforeunload", () => peerConnection.close())
			})

		return () => {
			player?.removeEventListener('mouseup', () => handleOverlayTimer)
			player?.removeEventListener('mouseenter', () => handleOverlayTimer)
			player?.removeEventListener('mouseleave', () => handleOverlayTimer)
			player?.removeEventListener('mousemove', () => handleOverlayTimer)

			clearTimeout(videoOverlayVisibleTimeoutRef.current)
		}
	}, [])

	return (
		<div
			id={streamVideoPlayerId}
			className="inline-block w-full relative z-0 aspect-video rounded-md mb-6"
			style={cinemaMode ? {
				maxHeight: '100vh',
				maxWidth: '100vw',
			} : {}}>

			<div className="absolute flex rounded-md w-full h-full">

				<div
					onClick={handleVideoPlayerClick}
					onDoubleClick={handleVideoPlayerDoubleClick}
					className={`
					absolute
					rounded-md
					w-full
					h-full
					z-20
					${streamState !== "Playing" && "bg-gray-950"}
					${streamState === "Playing" && `
						transition-opacity
						duration-500
						hover: ${videoOverlayVisible ? 'opacity-100' : 'opacity-0'}
						${!videoOverlayVisible ? 'cursor-none' : 'cursor-default'}
					`}
				`}
				>

					{/*Buttons */}
					{videoRef.current !== null && (
						<div className="absolute bottom-0 h-8 w-full flex place-items-end z-30">
							<div
								onClick={(e) => e.stopPropagation()}
								className="bg-blue-950 w-full flex flex-row gap-2 h-1/14 p-1 max-h-8 min-h-8 rounded-md">

								<PlayPauseComponent videoRef={videoRef} />

								<VolumeComponent
									isMuted={videoRef.current?.muted ?? false}
									isDisabled={audioLayers.length === 0}
									onVolumeChanged={(newValue) => videoRef.current!.volume = newValue}
									onStateChanged={(newState) => videoRef.current!.muted = newState}
								/>

								<div className="w-full pointer-events-none"></div>

								<CurrentViewersComponent currentViewersCount={currentStreamStatus?.viewers ?? 0} />
								<VideoLayerSelectorComponent layers={videoLayers} layerEndpoint={layerEndpointRef.current} hasPacketLoss={false} currentLayer={currentLayersStatus?.videoLayerCurrent ?? ""} />
								{audioLayers.length > 1 && (
									<AudioLayerSelectorComponent layers={audioLayers} layerEndpoint={layerEndpointRef.current} hasPacketLoss={false} currentLayer={currentLayersStatus?.videoLayerCurrent ?? ""} />
								)}
								<Square2StackIcon onClick={() => videoRef.current?.requestPictureInPicture()} />
								<ArrowsPointingOutIcon onClick={() => videoRef.current?.requestFullscreen()} />
							</div>
						</div>)
					}

					{!!props.onCloseStream && (
						<button
							onClick={props.onCloseStream}
							className="absolute top-2 right-2 p-2 rounded-full bg-red-400 hover:bg-red-500 pointer-events-auto">
							<svg
								xmlns="http://www.w3.org/2000/svg"
								className="h-6 w-6 text-gray-700"
								viewBox="0 0 24 24"
								fill="black">
								<path
									fillRule="evenodd"
									d="M6.225 6.225a.75.75 0 011.06 0L12 10.94l4.715-4.715a.75.75 0 111.06 1.06L13.06 12l4.715 4.715a.75.75 0 11-1.06 1.06L12 13.06l-4.715 4.715a.75.75 0 11-1.06-1.06L10.94 12 6.225 7.285a.75.75 0 010-1.06z"
									clipRule="evenodd"
								/>
							</svg>
						</button>
					)}

				</div>

				{/* Status messages */}
				<StatusMessageComponent
					streamKey={streamKey}
					state={streamState}
				/>

				<video
					ref={videoRef}
					autoPlay
					muted
					playsInline
					className="rounded-md w-full h-full relative bg-gray-950"
					onPlaying={() => setStreamState("Playing")}
					onLoadStart={() => setStreamState("Loading")}
					onEnded={() => setStreamState("Offline")}
				/>

			</div>

			{/* Stream MOTD*/}
			<StreamMOTD
				isOnline={currentStreamStatus.isOnline}
				motd={currentStreamStatus.motd}
			/>

		</div>
	)
}

export default Player
