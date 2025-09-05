import React, { useEffect, useRef, useState } from 'react'
import { useLocation } from 'react-router-dom'
import { useNavigate } from 'react-router-dom'
import PlayerHeader from '../playerHeader/PlayerHeader';
import { parseLinkHeader } from '@web3-storage/parse-link-header';
import Button from '../shared/Button';
import { ErrorMessageEnum, getMediaErrorMessage } from './errorMessage';
import ProfileSettings from './ProfileSettings';
import Player from '../player/Player';

const mediaOptions = {
	audio: true,
	video: {
		width: { ideal: 1920 },
		height: { ideal: 1080 },
	},
}

function BrowserBroadcaster() {
	const location = useLocation()
	const navigate = useNavigate();
	const streamKey = location.pathname.split('/').pop()
	const [mediaAccessError, setMediaAccessError] = useState<ErrorMessageEnum | null>(null)
	const [publishSuccess, setPublishSuccess] = useState(false)
	const [useDisplayMedia, setUseDisplayMedia] = useState<"Screen" | "Webcam" | "None">("None");
	const [peerConnectionDisconnected, setPeerConnectionDisconnected] = useState(false)
	const [hasPacketLoss, setHasPacketLoss] = useState<boolean>(false)
	const [hasSignal, setHasSignal] = useState<boolean>(false);
	const [connectFailed, setConnectFailed] = useState<boolean>(false);
	const [profileStateIsActive, setProfileStateIsActive] = useState<boolean>(false)
	const [profileStreamKey, setProfileStreamKey] = useState<string>("")

	const peerConnectionRef = useRef<RTCPeerConnection | null>(null);
	const videoRef = useRef<HTMLVideoElement>(null)
	const hasSignalRef = useRef<boolean>(false);
	const badSignalCountRef = useRef<number>(10);

	const endStream = () => navigate('/')

	useEffect(() => {
		// Fetch ICE-Servers
		fetch(`/api/ice-servers`, {
			method: 'GET',
		}).then(r => r.json())
			.then((result) => {
				peerConnectionRef.current = new RTCPeerConnection({
					iceServers: result,
					iceTransportPolicy: result != undefined ? "relay" : undefined
				});
			})

		return () => peerConnectionRef.current?.close()
	}, [])

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

		const mediaPromise = useDisplayMedia == "Screen" ?
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

			const encodingPrefix = "Web"
			mediaStream
				.getTracks()
				.forEach(mediaStreamTrack => {
					if (mediaStreamTrack.kind === 'audio') {
						peerConnectionRef.current!.addTransceiver(mediaStreamTrack, {
							direction: 'sendonly',
						})
					} else {
						peerConnectionRef.current!.addTransceiver(mediaStreamTrack, {
							direction: 'sendonly',
							sendEncodings: [
								{
									rid: encodingPrefix + 'High',
								},
								{
									rid: encodingPrefix + 'Mid',
									scaleResolutionDownBy: 2.0
								},
								{
									rid: encodingPrefix + 'Low',
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

					fetch(`/api/whip`, {
						method: 'POST',
						body: offer.sdp,
						headers: {
							Authorization: `Bearer ${streamKey}`,
							'Content-Type': 'application/sdp'
						}
					}).then(r => {

						if (r.status !== 201) {
							setConnectFailed(() => true)
							console.error("WHIP Endpoint did not return 201")
						}
						const parsedLinkHeader = parseLinkHeader(r.headers.get('Link'))

						if (parsedLinkHeader === null || parsedLinkHeader === undefined) {
							throw new DOMException("Missing link header");
						}

						const evtSource = new EventSource(`${parsedLinkHeader['urn:ietf:params:whep:ext:core:server-sent-events'].url}`)

						evtSource.onerror = _ => evtSource.close();

						// Receive current status of the stream
						// evtSource.addEventListener("status", (event: MessageEvent) => setCurrentStreamStatus(JSON.parse(event.data)))

						return r.text()
					}).then(answer => {
						peerConnectionRef.current!.setRemoteDescription({
							sdp: answer,
							type: 'answer'
						}).catch((err) => console.error("SetRemoteDescription", err))
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
		<div className='flex flex-col container mx-auto gap-2'>
			{mediaAccessError != null && <PlayerHeader headerType={"Error"}>{getMediaErrorMessage(mediaAccessError)}</PlayerHeader>}
			{peerConnectionDisconnected && <PlayerHeader headerType={"Error"}>WebRTC has disconnected or failed to connect at all</PlayerHeader>}
			{connectFailed && <PlayerHeader headerType={"Error"}>Failed to start Broadcast Box session</PlayerHeader>}
			{hasPacketLoss && <PlayerHeader headerType={"Warning"}>WebRTC is experiencing packet loss</PlayerHeader>}
			{publishSuccess && <PlayerHeader headerType={"Success"}>Live: Currently streaming to <a href={window.location.href.replace('publish/', '')} target="_blank" rel="noreferrer" className="hover:underline">{window.location.href.replace('publish/', '')}</a></PlayerHeader>}

			{/* Browser video feed */}
			{profileStateIsActive ? (
				<Player
					streamKey={profileStreamKey}
					cinemaMode={false} />
			) : (
				<video
					ref={videoRef}
					autoPlay
					muted
					controls
					playsInline
					className='w-full h-full aspect-video'
				/>
			)}

			{/* TODO: Add this view instead of only relying on the Player */}
			{/* Current stream status */}
			{/* <StreamStatus currentViewerCount={0} /> */}

			{/* Profile settings */}
			<ProfileSettings stateHasChanged={(isActive, streamKey) => {
				setProfileStateIsActive(() => isActive)
				setProfileStreamKey(() => streamKey)
			}
			} />

			{/* Buttons */}
			{!profileStateIsActive && (
				<div className="flex flex-row gap-2">
					<Button
						color='Accept'
						title="Publish Screen/Window/Tab"
						onClick={() => setUseDisplayMedia("Screen")}
					/>
					<Button
						title="Publish Webcam"
						onClick={() => setUseDisplayMedia("Webcam")}
					/>
				</div>
			)}

			{/* Conclude browser stream */}
			{publishSuccess && (
				<Button
					title="End stream"
					onClick={endStream}
				/>
			)}
		</div>
	)
}

export default BrowserBroadcaster
