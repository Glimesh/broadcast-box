import React, { useEffect, useMemo, useRef, useState } from "react";

export interface StreamStatus {
	streamKey: string;
	motd: string;
	viewers: number;
	isOnline: boolean;
}

export interface WhepSession {
	id: string;

	audioLayerCurrent: string;
	audioTimestamp: string;
	audioPacketsWritten: number;
	audioSequenceNumber: number;

	videoBitrate: number;
	videoLayerCurrent: string;
	videoTimestamp: string;
	videoPacketsWritten: number;
	videoSequenceNumber: number;

	sequenceNumber: number;
	timestamp: number;
}

export interface StatusResult {
	streamKey: string;
	motd: string;
	streamStart: Date;

	videoTracks: VideoTrack[];
	audioTracks: AudioTrack[];

	sessions: WhepSession[];
}

interface VideoTrack {
	rid: string;
	bitrate: number;
	packetsReceived: number;
	packetsDropped: number;
	lastKeyframe: string;
}

interface AudioTrack {
	rid: string;
	packetsReceived: number;
	packetsDropped: number;
}

interface StatusProviderProps {
	children: React.ReactNode;
}

const STATUS_POLL_INTERVAL_MS = 5000;

class FetchError extends Error {
	status: number;

	constructor(message: string, status: number) {
		super(message);
		this.status = status;
	}
}

const fetchStatus = (
	// eslint-disable-next-line no-unused-vars
	onSuccess?: (statusResults: StatusResult[]) => void,
	// eslint-disable-next-line no-unused-vars
	onError?: (error: FetchError) => void
) =>
	fetch(`/api/status`, {
		method: 'GET',
		headers: {
			'Content-Type': 'application/json'
		}
	}).then(result => {
		if (result.status === 503) {
			throw new FetchError('Status API disabled', result.status);
		}
		if (!result.ok) {
			throw new FetchError('Unknown error when calling status', result.status);
		}
		return result.json()
	})
		.then((result: StatusResult[]) => onSuccess?.(result))
		.catch((err: FetchError) => onError?.(err));

interface StatusProviderContextProps {
	activeStreamsStatus: StatusResult[] | undefined
	currentStreamStatus: StreamStatus | undefined,
	refreshStatus: () => void
	subscribe: () => void
	unsubscribe: () => void

	// eslint-disable-next-line no-unused-vars
	setCurrentStreamStatus: (status: StreamStatus) => void
}

export const StatusContext = React.createContext<StatusProviderContextProps>({
	activeStreamsStatus: undefined,
	currentStreamStatus: undefined,
	refreshStatus: () => { },
	subscribe: () => { },
	unsubscribe: () => { },
	setCurrentStreamStatus: () => { }
});

export function StatusProvider(props: StatusProviderProps) {
	const [streamStatus, setStreamStatus] = useState<StatusResult[] | undefined>(undefined)
	const [currentStreamStatus, setCurrentStreamStatus] = useState<StreamStatus | undefined>(undefined)
	const intervalRef = useRef<number | undefined>(undefined)
	const subscribers = useRef<number>(0)

	const fetchStatusResultHandler = (result: StatusResult[]) => {
		setStreamStatus(() => result);
	}

	const fetchStatusErrorHandler = (error: FetchError) => {
		if (error.status === 503) {
			setStreamStatus(() => undefined);
		}
	}

	const refreshStatus = async () => {
		await fetchStatus(fetchStatusResultHandler, fetchStatusErrorHandler)
	}

	const stopFetching = () => {
		clearInterval(intervalRef.current)
		intervalRef.current = undefined
	}

	const subscribe = () => {
		subscribers.current++;

		if (subscribers.current >= 1) {
			startFetching()
		}
	}

	const unsubscribe = () => {
		subscribers.current--;

		if (subscribers.current === 0) {
			stopFetching()
		}
	}

	const startFetching = () => {
		if (!intervalRef.current) {
			intervalRef.current = setInterval(() => {
				void refreshStatus()
			}, STATUS_POLL_INTERVAL_MS)
		}
	}

	useEffect(() => stopFetching, [])

	const state = useMemo<StatusProviderContextProps>(() => ({
		activeStreamsStatus: streamStatus,
		currentStreamStatus: currentStreamStatus,
		refreshStatus: () => {
			void refreshStatus()
		},
		subscribe: subscribe,
		unsubscribe: unsubscribe,
		setCurrentStreamStatus: (value: StreamStatus) => setCurrentStreamStatus(() => value)
	// eslint-disable-next-line react-hooks/exhaustive-deps
	}), [streamStatus, currentStreamStatus]);

	return (
		<StatusContext.Provider value={state}>
			{props.children}
		</StatusContext.Provider>
	);
}
