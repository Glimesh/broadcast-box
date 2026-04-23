import { useCallback, useContext, useEffect, useRef, useState } from 'react';
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
import { useReconnectController } from '../../hooks/useReconnectController';

const mediaOptions = {
	audio: true,
	video: {
		width: { ideal: 1920 },
		height: { ideal: 1080 },
	},
}

type MediaSource = "Screen" | "Webcam"

function BrowserBroadcaster() {
	const location = useLocation()
	const { locale } = useContext(LocaleContext)
	const navigate = useNavigate();
	const streamKey = decodeURIComponent(location.pathname.split('/').pop() ?? "")
	const [mediaAccessError, setMediaAccessError] = useState<ErrorMessageEnum | null>(null)
	const [publishSuccess, setPublishSuccess] = useState(false)
	const [useDisplayMedia, setUseDisplayMedia] = useState<MediaSource | "None">("None");
	const [mediaRequestCount, setMediaRequestCount] = useState(0)
	const [peerConnectionDisconnected, setPeerConnectionDisconnected] = useState(false)
	const [hasPacketLoss, setHasPacketLoss] = useState<boolean>(false)
	const [hasSignal, setHasSignal] = useState<boolean>(false);
	const [connectFailed, setConnectFailed] = useState<boolean>(false);
	const [profileStateIsActive, setProfileStateIsActive] = useState<boolean>(false)
	const [profileStreamKey, setProfileStreamKey] = useState<string>("")

	const peerConnectionRef = useRef<RTCPeerConnection | null>(null);
	const localMediaStreamRef = useRef<MediaStream | null>(null)
	const eventSourceRef = useRef<EventSource | null>(null)
	const whipResourceUrlRef = useRef<string | null>(null)
	const setupInProgressRef = useRef<boolean>(false)
	const videoRef = useRef<HTMLVideoElement>(null)
	const hasSignalRef = useRef<boolean>(false);
	const badSignalCountRef = useRef<number>(10);
	const shouldAutoReconnectRef = useRef<boolean>(false)
	const {
		scheduleReconnect,
		reset: resetReconnect,
		cancel: cancelReconnect,
	} = useReconnectController({
		baseDelayMs: 500,
		maxDelayMs: 8_000,
		maxAttempts: 8,
	})

	const endStream = () => navigate('/')
	const requestMedia = (source: MediaSource) => {
		if (!navigator.mediaDevices) {
			setMediaAccessError(() => ErrorMessageEnum.NoMediaDevices);
			return
		}

		setUseDisplayMedia(source)
		setMediaRequestCount(prev => prev + 1)
	}

	const stopLocalMediaStream = useCallback((localMediaStream: MediaStream | null) => {
		if (!localMediaStream) {
			return
		}

		localMediaStream
			.getTracks()
			.forEach((streamTrack: MediaStreamTrack) => streamTrack.stop())
	}, [])

	const closeEventSource = useCallback(() => {
		eventSourceRef.current?.close()
		eventSourceRef.current = null
	}, [])

	const deleteWhipSession = useCallback(async () => {
		const currentWhipResource = whipResourceUrlRef.current
		if (!currentWhipResource) {
			return
		}

		whipResourceUrlRef.current = null

		await fetch(currentWhipResource, {
			method: 'DELETE'
		}).catch((err) => {
			console.error("WHIP.DeleteSession", err)
		})
	}, [])

	const closePeerConnectionAndSession = useCallback(async () => {
		closeEventSource()
		peerConnectionRef.current?.close()
		peerConnectionRef.current = null
		await deleteWhipSession()
	}, [closeEventSource, deleteWhipSession])

	const isFatalWhipStatus = useCallback((statusCode: number) => {
		return statusCode === 400 || statusCode === 401 || statusCode === 403 || statusCode === 404
	}, [])

	const triggerReconnect = useCallback((setupPublisherSession: () => Promise<void>) => {
		if (!shouldAutoReconnectRef.current) {
			return
		}

		scheduleReconnect(() => {
			void setupPublisherSession()
		})
	}, [scheduleReconnect])

	useEffect(() => {
		return () => {
			cancelReconnect()
			shouldAutoReconnectRef.current = false
			void closePeerConnectionAndSession()
			stopLocalMediaStream(localMediaStreamRef.current)
			localMediaStreamRef.current = null
		}
	}, [cancelReconnect, closePeerConnectionAndSession, stopLocalMediaStream])

	useEffect(() => {
		if (useDisplayMedia === "None") {
			shouldAutoReconnectRef.current = false
			cancelReconnect()
			return;
		}

		let cancelled = false
		shouldAutoReconnectRef.current = true

		const setupPublisherSession = async () => {
			if (setupInProgressRef.current || cancelled) {
				return
			}

			const mediaStream = localMediaStreamRef.current
			if (!mediaStream) {
				return
			}

			setupInProgressRef.current = true
			setPeerConnectionDisconnected(() => false)
			setConnectFailed(() => false)

			const videoTrack = mediaStream.getVideoTracks()[0] ?? null
			const audioTrack = mediaStream.getAudioTracks()[0] ?? null

			await closePeerConnectionAndSession()

			const peerConnection = new RTCPeerConnection()
			peerConnectionRef.current = peerConnection

			peerConnection.addTransceiver(audioTrack ? audioTrack : "audio", { direction: 'sendonly' })

			const isFirefox = navigator.userAgent.toLowerCase().includes('firefox')
			const encodingPrefix = "Web"
			peerConnection.addTransceiver(videoTrack ? videoTrack : "video", {
				direction: 'sendonly',
				sendEncodings: isFirefox ? undefined : [
					{ rid: encodingPrefix + 'High' },
					{ rid: encodingPrefix + 'Mid', scaleResolutionDownBy: 2.0 },
					{ rid: encodingPrefix + 'Low', scaleResolutionDownBy: 4.0 },
				],
			})

			peerConnection.oniceconnectionstatechange = () => {
				if (peerConnection.iceConnectionState === 'connected' || peerConnection.iceConnectionState === 'completed') {
					setPublishSuccess(() => true)
					setMediaAccessError(() => null)
					setPeerConnectionDisconnected(() => false)
					resetReconnect()
					return
				}

				if (peerConnection.iceConnectionState === 'disconnected' || peerConnection.iceConnectionState === 'failed') {
					setPublishSuccess(() => false)
					setPeerConnectionDisconnected(() => true)
					triggerReconnect(setupPublisherSession)
				}
			}

			try {
				const offer = await peerConnection.createOffer()
				await peerConnection.setLocalDescription(offer)

				const response = await fetch(`/api/whip`, {
					method: 'POST',
					body: offer.sdp,
					headers: {
						Authorization: `Bearer ${toBase64Utf8(streamKey)}`,
						'Content-Type': 'application/sdp'
					}
				})

				if (response.status !== 201) {
					setConnectFailed(() => true)
					setPublishSuccess(() => false)
					if (isFatalWhipStatus(response.status)) {
						shouldAutoReconnectRef.current = false
						cancelReconnect()
						return
					}

					throw new DOMException("WHIP Endpoint did not return 201")
				}

				whipResourceUrlRef.current = response.headers.get('Location')

				const parsedLinkHeader = parseLinkHeader(response.headers.get('Link'))
				if (parsedLinkHeader === null || parsedLinkHeader === undefined) {
					throw new DOMException("Missing link header")
				}

				closeEventSource()
				const evtSource = new EventSource(`${parsedLinkHeader['urn:ietf:params:whep:ext:core:server-sent-events'].url}`)
				eventSourceRef.current = evtSource

				evtSource.onerror = () => {
					closeEventSource()
					setPublishSuccess(() => false)
					setPeerConnectionDisconnected(() => true)
					triggerReconnect(setupPublisherSession)
				}

				const answer = await response.text()
				await peerConnection.setRemoteDescription({
					sdp: answer,
					type: 'answer'
				})
			} catch (err) {
				console.error("Broadcast.SetupPublisherSession", err)
				setPublishSuccess(() => false)
				triggerReconnect(setupPublisherSession)
			} finally {
				setupInProgressRef.current = false
			}
		}

		const requestAndStartSession = async () => {
			const mediaPromise = useDisplayMedia == "Screen"
				? navigator.mediaDevices.getDisplayMedia(mediaOptions)
				: navigator.mediaDevices.getUserMedia(mediaOptions)

			try {
				const mediaStream = await mediaPromise
				if (cancelled) {
					stopLocalMediaStream(mediaStream)
					return
				}

				videoRef.current!.srcObject = mediaStream
				const previousLocalMediaStream = localMediaStreamRef.current
				localMediaStreamRef.current = mediaStream
				stopLocalMediaStream(previousLocalMediaStream)

				await setupPublisherSession()
			} catch (reason) {
				const mediaError = reason as { name?: string }
				if (mediaError.name === 'NotAllowedError') {
					setMediaAccessError(() => ErrorMessageEnum.NotAllowedError)
				} else if (mediaError.name === 'NotFoundError') {
					setMediaAccessError(() => ErrorMessageEnum.NotFoundError)
				} else {
					setMediaAccessError(() => ErrorMessageEnum.NoMediaDevices)
				}

				shouldAutoReconnectRef.current = false
				cancelReconnect()
				setUseDisplayMedia("None")
			}
		}

		void requestAndStartSession()

		return () => {
			cancelled = true
		}
		}, [cancelReconnect, closeEventSource, closePeerConnectionAndSession, isFatalWhipStatus, mediaRequestCount, resetReconnect, stopLocalMediaStream, streamKey, triggerReconnect, useDisplayMedia])

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
						onClick={() => requestMedia("Screen")}
					/>
					<Button
						title={locale.player_header.publish_webcam}
						onClick={() => requestMedia("Webcam")}
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
