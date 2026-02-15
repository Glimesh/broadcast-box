/* eslint-disable no-unused-vars */
import { parseLinkHeader } from "@web3-storage/parse-link-header";
import { StreamStatus } from "../../../providers/StatusProvider";
import { RefObject } from "react";

export interface CurrentLayersMessage {
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

enum SetupPeerConnectionError {
	INVALID_WHEP_RESPONSE
}
enum SetupPeerConnectionStateChange {
	ONLINE,
	OFFLINE
}
export interface SetupPeerConnectionProps {
	streamKey: string,
	videoRef: RefObject<HTMLVideoElement | null>,
	layerEndpointRef: RefObject<string>,

	onError: (error: SetupPeerConnectionError) => void,
	onStreamStatus: (status: StreamStatus) => void,
	onLayerStatus: (layers: CurrentLayersMessage) => void,
	onAudioLayerChange: (layers: []) => void,
	onVideoLayerChange: (layers: []) => void,
	onStateChange: (state: SetupPeerConnectionStateChange) => void,
	onStreamRestart: () => void,
}

const stopVideoTrack = (videoElement: HTMLVideoElement | null) => {
	const currentStream = videoElement?.srcObject;
	if (currentStream instanceof MediaStream) {
		currentStream.getTracks().forEach(track => track.stop());
	}
}
const clearVideoElement = (videoElement: HTMLVideoElement | null) => {
	if(videoElement){
		videoElement.muted = true
		videoElement.srcObject = null
	}
}

export async function PeerConnectionSetup(props: SetupPeerConnectionProps): Promise<RTCPeerConnection> {
	const {
		streamKey,
		videoRef,
		layerEndpointRef,
		onStreamRestart,
		onStreamStatus,
		onLayerStatus,
		onAudioLayerChange,
		onVideoLayerChange,
		onStateChange,
		onError } = props

	if (videoRef.current === null){
		throw new Error("PeerConnection.VideoRef is null")
	}

	stopVideoTrack(videoRef.current)
	clearVideoElement(videoRef.current)

	// Create peerconnection
	const peerConnection = await createPeerConnection()

	// Config
	peerConnection.addTransceiver('audio', { direction: 'recvonly' })
	peerConnection.addTransceiver('video', { direction: 'recvonly' })

	// Setup events
	const remoteStream = new MediaStream();
	peerConnection.ontrack = (event: RTCTrackEvent) => {
		remoteStream.addTrack(event.track);
		if (videoRef.current) {
			videoRef.current!.srcObject = remoteStream;
		} else {
			console.log("PeerConnection.onTrack", "Could not find VideoRef")
		}

		event.track.onended = () => remoteStream.removeTrack(event.track)
	}

	// Begin negotiation
	const offer = await peerConnection.createOffer({ iceRestart: true })
	offer["sdp"] = offer["sdp"]!.replace("useinbandfec=1", "useinbandfec=1;stereo=1")

	await peerConnection
		.setLocalDescription(offer)
		.catch((err) => console.error("PeerConnection.SetLocalDescription", err));

	await waitForIceGatheringComplete(peerConnection)

	const whepResponse = await fetch(`/api/whep`, {
		method: 'POST',
		headers: {
			Authorization: `Bearer ${streamKey}`,
			'Content-Type': 'application/sdp'
		},
		body: offer.sdp,
	})

	if (!whepResponse.ok) {
		console.log("PeerConnection.WhepResponse.Error", SetupPeerConnectionError.INVALID_WHEP_RESPONSE)
		onError(SetupPeerConnectionError.INVALID_WHEP_RESPONSE)
	}

	const parsedLinkHeader = parseLinkHeader(whepResponse.headers.get('Link'))

	if (parsedLinkHeader === null || parsedLinkHeader === undefined) {
		throw new DOMException("Missing link header");
	}

	layerEndpointRef.current = `${parsedLinkHeader['urn:ietf:params:whep:ext:core:layer'].url}`
	const evtSource = new EventSource(`${parsedLinkHeader['urn:ietf:params:whep:ext:core:server-sent-events'].url}`)

	evtSource.onerror = (ev: Event) => {
		console.error("PeerConnection.EventSource", ev)
		evtSource.close();
		onStateChange(SetupPeerConnectionStateChange.OFFLINE)
	}

	// Receive current status of the whep stream
	evtSource.addEventListener("streamStart", () => {
		console.log("PeerConnection.EventSource", "Reset Stream", streamKey)

		evtSource.close()
		peerConnection.close()

		onStreamRestart()
	})

	// Receive current status of the whep stream
	evtSource.addEventListener("status", (event: MessageEvent) => {
		onStreamStatus(JSON.parse(event.data) as StreamStatus)
	})

	// Receive current current layers of this whep stream
	evtSource.addEventListener("currentLayers", (event: MessageEvent) => {
		onLayerStatus(JSON.parse(event.data) as CurrentLayersMessage)
	})

	// Receive layers
	evtSource.addEventListener("layers", event => {
		const parsed = JSON.parse(event.data)
		onVideoLayerChange(parsed['1']['layers'].map((layer: any) => layer.encodingId))
		onAudioLayerChange(parsed['2']['layers'].map((layer: any) => layer.encodingId))
	})

	const answer = await whepResponse.text()
	await peerConnection.setRemoteDescription({
		sdp: answer,
		type: 'answer'
	}).catch((err) => console.error("PeerConnection.RemoteDescription", err))

	return peerConnection;
}

async function createPeerConnection(): Promise<RTCPeerConnection> {
	return await fetch(`/api/ice-servers`, {
		method: 'GET',
	}).then(r => r.json())
		.then((result) => {
			return new RTCPeerConnection({
				iceServers: result
			});
		}).catch(() => {
			console.error("Error calling Ice-Servers endpoint. Ignoring STUN/TURN configuration")
			return new RTCPeerConnection();
		})
}

export function waitForIceGatheringComplete(peerConnection: RTCPeerConnection) {
	return new Promise(resolve => {
		if (peerConnection.iceGatheringState === 'complete') {
			resolve(true);
		} else {
			const checkState = () => {
				if (peerConnection.iceGatheringState === 'complete') {
					peerConnection.removeEventListener('icegatheringstatechange', checkState);
					resolve(true);
				}
			};
			peerConnection.addEventListener('icegatheringstatechange', checkState);
		}
	});
}
