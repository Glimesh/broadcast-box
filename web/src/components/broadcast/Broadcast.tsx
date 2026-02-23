import React, { useContext, useEffect, useRef, useState } from 'react'
import { useLocation } from 'react-router-dom'
import { useNavigate } from 'react-router-dom'
import PlayerHeader from '../playerHeader/PlayerHeader';
import { parseLinkHeader } from '@web3-storage/parse-link-header';
import Button from '../shared/Button';
import { ErrorMessageEnum, getMediaErrorMessage } from './errorMessage';
import ProfileSettings from './ProfileSettings';
import Player from '../player/Player';
import { LocaleContext } from '../../providers/LocaleProvider';
import toBase64Utf8 from '../../utilities/base64';

const mediaOptions = {
	audio: true,
	video: {
		width: { ideal: 1920 },
		height: { ideal: 1080 },
	},
}

function BrowserBroadcaster() {
	const location = useLocation()
	const { locale } = useContext(LocaleContext)
	const navigate = useNavigate();
	const streamKey = decodeURIComponent(location.pathname.split('/').pop() ?? "")
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
	const localMediaStreamRef = useRef<MediaStream | null>(null)
	const eventSourceRef = useRef<EventSource | null>(null)
	const videoRef = useRef<HTMLVideoElement>(null)
	const hasSignalRef = useRef<boolean>(false);
	const badSignalCountRef = useRef<number>(10);

	const endStream = () => navigate('/')

	const stopLocalMediaStream = (localMediaStream: MediaStream | null) => {
		if (!localMediaStream) {
			return
		}

		localMediaStream
			.getTracks()
			.forEach((streamTrack: MediaStreamTrack) => streamTrack.stop())
	}

	const getSenderByKind = (peerConnection: RTCPeerConnection, kind: "audio" | "video") => {
		return peerConnection.getTransceivers().find(transceiver => transceiver.sender.track?.kind === kind)?.sender ??
			peerConnection.getTransceivers().find(transceiver => transceiver.receiver.track.kind === kind)?.sender ??
			null
	}

	useEffect(() => {
		return () => {
			eventSourceRef.current?.close()
			stopLocalMediaStream(localMediaStreamRef.current)
			localMediaStreamRef.current = null
			peerConnectionRef.current?.close()
			peerConnectionRef.current = null
		}
	}, [])

	useEffect(() => {
		if (useDisplayMedia === "None") {
			return;
		}

		if (!navigator.mediaDevices) {
			setMediaAccessError(() => ErrorMessageEnum.NoMediaDevices);
			setUseDisplayMedia(() => "None")
			return
		}

		let cancelled = false

		const mediaPromise = useDisplayMedia == "Screen" ?
			navigator.mediaDevices.getDisplayMedia(mediaOptions) :
			navigator.mediaDevices.getUserMedia(mediaOptions)

		mediaPromise.then(async mediaStream => {
			const nextLocalMediaStream = mediaStream

			if (cancelled) {
				stopLocalMediaStream(nextLocalMediaStream)
				return;
			}

			const videoTrack = mediaStream.getVideoTracks()[0] ?? null
			const audioTrack = mediaStream.getAudioTracks()[0] ?? null

			const existingPeerConnection = peerConnectionRef.current
			if (existingPeerConnection) {
				const videoSender = getSenderByKind(existingPeerConnection, "video")
				const audioSender = getSenderByKind(existingPeerConnection, "audio")

				await Promise.all([
					videoSender?.replaceTrack(videoTrack) ?? Promise.resolve(),
					audioSender?.replaceTrack(audioTrack) ?? Promise.resolve(),
				])

				if (
					cancelled ||
					peerConnectionRef.current !== existingPeerConnection
				) {
					stopLocalMediaStream(nextLocalMediaStream)
					return;
				}

				videoRef.current!.srcObject = mediaStream
				const previousLocalMediaStream = localMediaStreamRef.current
				localMediaStreamRef.current = nextLocalMediaStream
				stopLocalMediaStream(previousLocalMediaStream)
				return
			}

			const peerConnection = new RTCPeerConnection();
			peerConnectionRef.current = peerConnection

			if (
				cancelled ||
				peerConnectionRef.current !== peerConnection
			) {
				if (peerConnectionRef.current === peerConnection) {
					peerConnectionRef.current = null
				}
				peerConnection.close()
				stopLocalMediaStream(nextLocalMediaStream)
				return
			}

			videoRef.current!.srcObject = mediaStream
			const previousLocalMediaStream = localMediaStreamRef.current
			localMediaStreamRef.current = nextLocalMediaStream
			stopLocalMediaStream(previousLocalMediaStream)

			peerConnection.addTransceiver(audioTrack ? audioTrack : "audio", { direction: 'sendonly' })

			const isFirefox = navigator.userAgent.toLowerCase().includes('firefox')
			const encodingPrefix = "Web"
			peerConnection.addTransceiver(videoTrack ? videoTrack : "video", {
				direction: 'sendonly',
				sendEncodings: isFirefox ? undefined : [
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
				],
			})

			peerConnection.oniceconnectionstatechange = () => {
				if (peerConnection.iceConnectionState === 'connected' || peerConnection.iceConnectionState === 'completed') {
					setPublishSuccess(() => true)
					setMediaAccessError(() => null)
					setPeerConnectionDisconnected(() => false)
				} else if (peerConnection.iceConnectionState === 'disconnected' || peerConnection.iceConnectionState === 'failed') {
					setPublishSuccess(() => false)
					setPeerConnectionDisconnected(() => true)
				}
			}

			peerConnection
				.createOffer()
				.then(offer => {
					peerConnection.setLocalDescription(offer)
						.catch((err) => console.error("SetLocalDescription", err));

					fetch(`/api/whip`, {
						method: 'POST',
						body: offer.sdp,
						headers: {
							Authorization: `Bearer ${toBase64Utf8(streamKey)}`,
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

						eventSourceRef.current?.close()
						const evtSource = new EventSource(`${parsedLinkHeader['urn:ietf:params:whep:ext:core:server-sent-events'].url}`)
						eventSourceRef.current = evtSource

						evtSource.onerror = () => evtSource.close();

						// Receive current status of the stream
						// evtSource.addEventListener("status", (event: MessageEvent) => setCurrentStreamStatus(JSON.parse(event.data)))

						return r.text()
					}).then(answer => {
						peerConnection.setRemoteDescription({
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
			cancelled = true
		}
	// eslint-disable-next-line react-hooks/exhaustive-deps
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
			{mediaAccessError != null && <PlayerHeader headerType={"Error"}>{getMediaErrorMessage(locale, mediaAccessError)}</PlayerHeader>}
			{peerConnectionDisconnected && <PlayerHeader headerType={"Error"}>{locale.player_header.connection_disconnected}</PlayerHeader>}
			{connectFailed && <PlayerHeader headerType={"Error"}>{locale.player_header.connection_failed}</PlayerHeader>}
			{hasPacketLoss && <PlayerHeader headerType={"Warning"}>{locale.player_header.connection_has_packetloss}</PlayerHeader>}
			{publishSuccess && <PlayerHeader headerType={"Success"}>{locale.player_header.connection_established} <a href={window.location.href.replace('publish/', '')} target="_blank" rel="noreferrer" className="hover:underline">{window.location.href.replace('publish/', '')}</a></PlayerHeader>}

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
						title={locale.player_header.publish_screen}
						onClick={() => setUseDisplayMedia("Screen")}
					/>
					<Button
						title={locale.player_header.publish_webcam}
						onClick={() => setUseDisplayMedia("Webcam")}
					/>
				</div>
			)}

			{/* Conclude browser stream */}
			{publishSuccess && (
				<Button
					title={locale.player_header.button_end_stream}
					onClick={endStream}
				/>
			)}
		</div>
	)
}

export default BrowserBroadcaster
