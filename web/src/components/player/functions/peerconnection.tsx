/* eslint-disable no-unused-vars */
import { parseLinkHeader } from "@web3-storage/parse-link-header";
import { StreamStatus } from "../../../providers/StatusProvider";
import toBase64Utf8 from "../../../utilities/base64";
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
	onStreamRestart:() => void,
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
		
	const currentStream = videoRef.current?.srcObject;
  if (currentStream instanceof MediaStream) {
    currentStream.getTracks().forEach(track => track.stop());
  }
	videoRef.current!.srcObject = null;
	await new Promise(resolve => setTimeout(resolve, 150));
	
	// Create peerconnection
	const peerConnection = await createPeerConnection()

	// Config
	peerConnection.addTransceiver('audio', { direction: 'recvonly' })
	peerConnection.addTransceiver('video', { direction: 'recvonly' })

	// Setup events
	peerConnection.ontrack = (event: RTCTrackEvent) => {
		if (videoRef.current) {
			videoRef.current!.srcObject = event.streams[0];
		}else{
			console.log("Could not find VideoRef")
		}
	}

	// Begin negotiation
	const offer = await peerConnection.createOffer()
	offer["sdp"] = offer["sdp"]!.replace("useinbandfec=1", "useinbandfec=1;stereo=1")

	await peerConnection
	.setLocalDescription(offer)
	.catch((err) => console.error("PeerConnection.SetLocalDescription", err));

	// await waitForIceGatheringComplete(peerConnection)

	const whepResponse = await fetch(`/api/whep`, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/sdp'
		},
		body: toBase64Utf8(JSON.stringify({
			streamKey: streamKey,
			offer: offer.sdp
		})),
	})

	if(!whepResponse.ok){
		console.log("WhepSession.Response.Error closed")
		onError(SetupPeerConnectionError.INVALID_WHEP_RESPONSE)
	}

	const parsedLinkHeader = parseLinkHeader(whepResponse.headers.get('Link'))

	if (parsedLinkHeader === null || parsedLinkHeader === undefined) {
		throw new DOMException("Missing link header");
	}

	layerEndpointRef.current = `${parsedLinkHeader['urn:ietf:params:whep:ext:core:layer'].url}`
	const evtSource = new EventSource(`${parsedLinkHeader['urn:ietf:params:whep:ext:core:server-sent-events'].url}`)

	evtSource.onerror = (ev: Event) => {
		evtSource.close();
		console.log("PeerConnection.Eventsource Offline")
		console.error("PeerConnection.Eventsource", ev)
		onStateChange(SetupPeerConnectionStateChange.OFFLINE)
	}

	// Receive current status of the whep stream
	evtSource.addEventListener("streamStart", () => {
		console.log("PeerConnection.EventSource", "Reset Stream")
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
	}).catch((err) => console.error("RemoteDescription", err))

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

