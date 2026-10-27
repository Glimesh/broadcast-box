import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { CSSProperties, MouseEvent } from 'react';
import PlayPauseComponent from "./components/PlayPauseComponent";
import VideoLayerSelectorComponent from "./components/VideoLayerSelectorComponent";
import AudioLayerSelectorComponent from "./components/AudioLayerSelectorComponent";
import CurrentViewersComponent from "./components/CurrentViewersComponent";
import { StreamStatus } from '../../providers/StatusProvider';
import { CurrentLayersMessage, PeerConnectionDataChannels, PeerConnectionSetup, SetupPeerConnectionFailureType, SetupPeerConnectionProps, SetupPeerConnectionStateChange } from './functions/peerconnection';
import { ChatAdapter } from '../../hooks/useChatSession';
import { ArrowsPointingOutIcon, Square2StackIcon, HeartIcon, XMarkIcon } from '@heroicons/react/20/solid';
import { ChatBubbleLeftRightIcon } from '@heroicons/react/24/outline';
import VolumeComponent from './components/VolumeComponent';
import { StatusMessageComponent } from './components/StatusMessageComponent';
import { useReconnectController } from '../../hooks/useReconnectController';

interface PlayerProps {
	streamKey: string;
	cinemaMode: boolean;
	fillContainer?: boolean;
	isChatOpen?: boolean;
	onToggleChat?(): void;
	onChatAdapterChange?(streamKey: string, adapter: ChatAdapter | undefined): void;
	onReactionSenderChange?(streamKey: string, sender: ReactionSender | undefined): void;
	onStreamStatusChange?(streamKey: string, status: StreamStatus): void;
	onCloseStream?(): void;
}

export type ReactionSender = () => void;

interface FullscreenElement extends HTMLElement {
	webkitRequestFullscreen?: () => void | Promise<void>;
	msRequestFullscreen?: () => void | Promise<void>;
	webkitEnterFullscreen?: () => void;
}

const Player = (props: PlayerProps) => {
	const {
		cinemaMode,
		fillContainer = false,
		isChatOpen,
		onToggleChat,
		onChatAdapterChange,
		onReactionSenderChange,
		onStreamStatusChange,
		onCloseStream,
	} = props
	const streamKey = decodeURIComponent(props.streamKey).replace(/ /g, '_')

	const [currentStreamStatus, setCurrentStreamStatus] = useState<StreamStatus>({
		streamKey: streamKey,
		motd: "",
		viewers: 0,
		isOnline: false,
	})

	const [currentLayersStatus, setCurrentLayersStatus] = useState<CurrentLayersMessage | undefined>()
	const [audioLayers, setAudioLayers] = useState<string[]>([]);
	const [videoLayers, setVideoLayers] = useState<string[]>([]);
	const [videoElement, setVideoElement] = useState<HTMLVideoElement | null>(null)
	const [layerEndpoint, setLayerEndpoint] = useState<string>('')
	const [streamState, setStreamState] = useState<"Loading" | "Playing" | "Offline" | "Error">("Loading");
	const [videoOverlayVisible, setVideoOverlayVisible] = useState<boolean>(false)
	const [isVideoMuted, setIsVideoMuted] = useState<boolean>(true)
	const [videoVolume, setVideoVolume] = useState<number>(50)
	const [reactionAnimations, setReactionAnimations] = useState<{ id: number; x: number }[]>([])
	const [dataChannels, setDataChannels] = useState<PeerConnectionDataChannels | undefined>()

	const clickDelay = 250;
	const videoRef = useRef<HTMLVideoElement>(null);
	const fatalSetupErrorRef = useRef(false)
	const layerEndpointRef = useRef<string>('');
	const videoOverlayVisibleTimeoutRef = useRef<number | undefined>(undefined);
	const reactionAnimationIdRef = useRef(0);
	const lastClickTimeRef = useRef(0);
	const clickTimeoutRef = useRef<number | undefined>(undefined);
	const streamVideoPlayerId = streamKey + "_videoPlayer";
	const {
		scheduleReconnect,
		reset: resetReconnect,
		cancel: cancelReconnect,
	} = useReconnectController({
		baseDelayMs: 500,
		maxDelayMs: 8_000,
		maxAttempts: 8,
	})
	const setVideoRef = useCallback((element: HTMLVideoElement | null) => {
		videoRef.current = element
		setVideoElement(element)
	}, [])
	const isPictureInPictureSupported = videoElement !== null
		&& document.pictureInPictureEnabled
		&& typeof videoElement.requestPictureInPicture === 'function'

	const peerConnectionConfig = useMemo<SetupPeerConnectionProps>(() => ({
		streamKey: streamKey,
		videoRef: videoRef,
		layerEndpointRef: layerEndpointRef,
		onStateChange: (state) => console.log("PeerConnection.onStateChange", state),
		onStreamRestart: () => console.log("PeerConnection.onStreamRestart: Missing setup"),
		onAudioLayerChange: (layers) => setAudioLayers(layers),
		onVideoLayerChange: (layers) => setVideoLayers(layers),
		onLayerEndpointChange: (endpoint) => setLayerEndpoint(endpoint),
		onLayerStatus: (status) => setCurrentLayersStatus(status),
		onStreamStatus: (status) => {
			setCurrentStreamStatus(status)
			onStreamStatusChange?.(streamKey, status)

			if (!status.isOnline) {
				setStreamState("Offline")
				return
			}

			const videoElement = videoRef.current
			if (videoElement !== null && !videoElement.paused && videoElement.readyState >= HTMLMediaElement.HAVE_CURRENT_DATA) {
				setStreamState("Playing")
				return
			}

			setStreamState("Loading")
		},
		onError: (_, failureType) => {
			if (failureType === SetupPeerConnectionFailureType.FATAL) {
				fatalSetupErrorRef.current = true
				setStreamState("Error")
			}
		},
		onDataChannelsChange: (channels, active) => {
			setDataChannels((current) => active ? channels : current === channels ? undefined : current)
		},
	}), [onStreamStatusChange, streamKey])

	const addReactionAnimation = useCallback(() => {
		const id = reactionAnimationIdRef.current + 1;
		reactionAnimationIdRef.current = id;
		const x = Math.round((Math.random() - 0.5) * 32);

		setReactionAnimations((current) => [...current.slice(-7), { id, x }]);
	}, []);

	const handleEnterFullscreen = () => {
		const videoElement = videoRef.current as FullscreenElement | null;
		if (!videoElement) {
			return;
		}

		try {
			if (videoElement.requestFullscreen) {
				void videoElement.requestFullscreen().catch((err) => {
					console.error("VideoPlayer_RequestFullscreen", err);
				});
			} else if (videoElement.webkitRequestFullscreen) {
				void videoElement.webkitRequestFullscreen();
			} else if (videoElement.msRequestFullscreen) {
				void videoElement.msRequestFullscreen();
			} else if (videoElement.webkitEnterFullscreen) {
				videoElement.webkitEnterFullscreen();
			}
		} catch (err) {
			console.error("VideoPlayer_RequestFullscreen", err)
		}
	};


	const resetTimer = useCallback((isVisible: boolean) => {
		setVideoOverlayVisible(isVisible);

		if (videoOverlayVisibleTimeoutRef) {
			clearTimeout(videoOverlayVisibleTimeoutRef.current)
		}

		videoOverlayVisibleTimeoutRef.current = setTimeout(() => {
			setVideoOverlayVisible(false)
		}, 2500)
	}, [])

	const handleVideoPlayerClick = () => {
		lastClickTimeRef.current = Date.now();

			clickTimeoutRef.current = setTimeout(() => {
				const timeSinceLastClick = Date.now() - lastClickTimeRef.current;
				if (timeSinceLastClick >= clickDelay && (timeSinceLastClick - clickDelay) < 5000) {
					if (videoRef.current?.paused) {
						void videoRef.current.play()
					} else {
						videoRef.current?.pause()
					}
				}
			}, clickDelay);
	};

	const handleVideoPlayerDoubleClick = () => {
		clearTimeout(clickTimeoutRef.current);
		lastClickTimeRef.current = 0;
		handleEnterFullscreen();
	};

	const stopOverlayClickPropagation = (event: MouseEvent<HTMLElement>) => {
		event.preventDefault();
		event.stopPropagation();
	};

	useEffect(() => {
		const player = document.getElementById(streamVideoPlayerId)
		const handleMouseUp = () => resetTimer(true)
		const handleMouseMove = () => resetTimer(true)
		const handleMouseEnter = () => resetTimer(true)
		const handleMouseLeave = () => resetTimer(false)
		player?.addEventListener('mouseup', handleMouseUp)
		player?.addEventListener('mousemove', handleMouseMove)
		player?.addEventListener('mouseenter', handleMouseEnter)
		player?.addEventListener('mouseleave', handleMouseLeave)

		let currentPeerConnection: RTCPeerConnection | null = null
		let disposed = false
		let setupInProgress = false
		const beforeUnloadHandler = () => currentPeerConnection?.close()
		window.addEventListener("beforeunload", beforeUnloadHandler)
		resetReconnect()

		const setupPeerConnection = () => {
			if (setupInProgress || disposed) {
				return
			}

			setupInProgress = true
			fatalSetupErrorRef.current = false
			setStreamState("Loading")

			const setupProps: SetupPeerConnectionProps = {
				...peerConnectionConfig,
				onStateChange: (state) => {
					if (state === SetupPeerConnectionStateChange.ONLINE) {
						resetReconnect()
						return
					}

					scheduleReconnect(setupPeerConnection)
				},
				onStreamRestart: () => {
					resetReconnect()
					scheduleReconnect(setupPeerConnection, { immediate: true })
				},
			}

			currentPeerConnection?.close()
			currentPeerConnection = null

			PeerConnectionSetup(setupProps)
				.then((peerConnection) => {
					setupInProgress = false
					if (disposed) {
						peerConnection.close()
						return
					}
					currentPeerConnection = peerConnection
				})
				.catch((err) => {
					setupInProgress = false
					console.log("PeerConnectionConfig.Error", err)

					if (disposed || fatalSetupErrorRef.current) {
						if (fatalSetupErrorRef.current) {
							cancelReconnect()
						}
						return
					}

					scheduleReconnect(setupPeerConnection)
				})
		}

		setupPeerConnection()

		return () => {
			disposed = true
			cancelReconnect()
			player?.removeEventListener('mouseup', handleMouseUp)
			player?.removeEventListener('mouseenter', handleMouseEnter)
			player?.removeEventListener('mouseleave', handleMouseLeave)
			player?.removeEventListener('mousemove', handleMouseMove)

			window.removeEventListener("beforeunload", beforeUnloadHandler)
			currentPeerConnection?.close()
			clearTimeout(videoOverlayVisibleTimeoutRef.current)
		}
	}, [cancelReconnect, peerConnectionConfig, resetReconnect, resetTimer, scheduleReconnect, streamVideoPlayerId])

	useEffect(() => {
		const adapter = dataChannels?.chatAdapter;
		if (!adapter) {
			return;
		}

		onChatAdapterChange?.(streamKey, adapter);

		return () => {
			onChatAdapterChange?.(streamKey, undefined);
		};
	}, [dataChannels, onChatAdapterChange, streamKey]);

	useEffect(() => {
		const channel = dataChannels?.reactionDataChannel;
		if (!channel) {
			return;
		}

		const sendReaction = () => {
			if (channel.readyState !== "open") {
				return;
			}

			try {
				channel.send(JSON.stringify({ type: "reaction" }));
				addReactionAnimation();
			} catch (error) {
				console.log("ReactionDataChannel.Send.Error", error);
			}
		};
		const updateSender = () => {
			onReactionSenderChange?.(streamKey, channel.readyState === "open" ? sendReaction : undefined);
		};
		const handleMessage = (event: MessageEvent) => {
			if (typeof event.data !== "string") {
				return;
			}

			try {
				const payload = JSON.parse(event.data) as { type?: string };
				if (payload.type === "reaction") {
					addReactionAnimation();
				}
			} catch {
				return;
			}
		};

		channel.addEventListener("open", updateSender);
		channel.addEventListener("close", updateSender);
		channel.addEventListener("error", updateSender);
		channel.addEventListener("message", handleMessage);
		updateSender();

		return () => {
			channel.removeEventListener("open", updateSender);
			channel.removeEventListener("close", updateSender);
			channel.removeEventListener("error", updateSender);
			channel.removeEventListener("message", handleMessage);
			onReactionSenderChange?.(streamKey, undefined);
		};
	}, [addReactionAnimation, dataChannels, onReactionSenderChange, streamKey]);

	return (
		<div className={`w-full flex items-end ${fillContainer ? "h-full" : ""}`}>
			<div
				key={`${streamVideoPlayerId}`}
				id={streamVideoPlayerId}
				className={`inline-block w-full relative z-0 rounded-md ${fillContainer ? "h-full" : "aspect-video"}`}
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
						{videoElement !== null && (
						<div className="absolute bottom-0 h-8 w-full flex place-items-end z-30">
							<div
								onClick={stopOverlayClickPropagation}
								onDoubleClick={stopOverlayClickPropagation}
								className="player-drag-handle bg-blue-950 w-full flex flex-row gap-2 h-1/14 p-1 max-h-8 min-h-8 rounded-md cursor-move">

								<span className="player-drag-cancel flex h-full items-center cursor-pointer">
									<PlayPauseComponent videoRef={videoRef} />
								</span>

								<span className="player-drag-cancel flex h-full items-center cursor-pointer">
									<VolumeComponent
										isMuted={isVideoMuted}
										volume={videoVolume}
										onVolumeChanged={(newValue) => {
											if (videoRef.current) {
												videoRef.current.volume = newValue / 100
											}
										}}
										onStateChanged={(newState) => {
											if (videoRef.current) {
												videoRef.current.muted = newState
											}
										}}
									/>
								</span>

								<div className="w-full pointer-events-none"></div>

								<span className="player-drag-cancel flex h-full items-center cursor-default">
									<CurrentViewersComponent currentViewersCount={currentStreamStatus?.viewers ?? 0} />
								</span>
								<span className="player-drag-cancel flex h-full items-center cursor-pointer">
									<VideoLayerSelectorComponent layers={videoLayers} layerEndpoint={layerEndpoint} hasPacketLoss={false} currentLayer={currentLayersStatus?.videoLayerCurrent ?? ""} />
								</span>

								{audioLayers.length > 1 && (
									<span className="player-drag-cancel flex h-full items-center cursor-pointer">
										<AudioLayerSelectorComponent layers={audioLayers} layerEndpoint={layerEndpoint} hasPacketLoss={false} currentLayer={currentLayersStatus?.videoLayerCurrent ?? ""} />
									</span>
								)}

								{isPictureInPictureSupported && (
									<Square2StackIcon className="player-drag-cancel cursor-pointer" onClick={() => videoElement.requestPictureInPicture()} />
								)}
								<ArrowsPointingOutIcon className="player-drag-cancel cursor-pointer" onClick={handleEnterFullscreen} />
							</div>
						</div>)
					}

					{/* Status messages */}
					<StatusMessageComponent
						streamKey={streamKey}
						state={streamState}
					/>

					<div
						onDoubleClick={stopOverlayClickPropagation}
						className="absolute top-2 right-2 flex flex-row gap-2 pointer-events-auto z-60">
					{!!onToggleChat && (
						<button
							onClick={(e) => {
								e.preventDefault();
								e.stopPropagation();
								onToggleChat();
							}}
							className={`p-2 rounded-full border ${isChatOpen ? 'bg-blue-600 border-blue-500 text-white' : 'bg-black/60 border-gray-700 text-gray-200 hover:bg-gray-800'}`}
						>
							<ChatBubbleLeftRightIcon className="h-5 w-5" />
						</button>
					)}

						{!!onCloseStream && (
							<button
								onClick={(e) => {
									e.preventDefault();
									e.stopPropagation();
									onCloseStream();
								}}
								className="p-2 rounded-full bg-red-400 hover:bg-red-500">
								<XMarkIcon className="h-6 w-6 text-black" />
							</button>
						)}
					</div>

				</div>

					<video
						key={`${streamVideoPlayerId}_video`}
						ref={setVideoRef}
					autoPlay
					muted={isVideoMuted}
					playsInline
					className="rounded-md w-full h-full relative bg-gray-950"
					onPlaying={() => {
						setStreamState("Playing")
						resetReconnect()
					}}
					onLoadStart={() => setStreamState("Loading")}
					onVolumeChange={(event) => {
						setIsVideoMuted(event.currentTarget.muted)
						setVideoVolume(Math.round(event.currentTarget.volume * 100))
					}}
						onLoadedData={(event) => {
							console.log("VideoPlayer.onLoadedMetadata", event)
							event.currentTarget.volume = videoVolume / 100
							event.currentTarget.play()
						}}
					onError={(error) => console.log("VideoPlayer.Error", error)}
					onErrorCapture={(error) => console.log("VideoPlayer.ErrorCapture", error)}
					onEnded={() => setStreamState("Offline")}
				/>

					<div className="pointer-events-none absolute right-5 bottom-10 z-70 h-20 w-14 overflow-visible">
						{reactionAnimations.map((reaction) => (
							<HeartIcon
								key={reaction.id}
								className="animate-reaction-float absolute right-0 bottom-0 h-7 w-7 text-rose-500 drop-shadow"
								style={{ "--reaction-x": `${reaction.x}px` } as CSSProperties}
								onAnimationEnd={() => setReactionAnimations((current) => current.filter((item) => item.id !== reaction.id))}
							/>
						))}
					</div>

				</div>
			</div>
		</div>
	)
}

export default Player
