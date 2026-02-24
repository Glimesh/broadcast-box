import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";

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
	onSuccess?: (statusResults: StatusResult[]) => void,
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
	const [currentStreamStatus, setCurrentStreamStatusState] = useState<StreamStatus | undefined>(undefined)
	const intervalRef = useRef<number | undefined>(undefined)
	const subscribers = useRef<number>(0)

	const fetchStatusResultHandler = useCallback((result: StatusResult[]) => {
		setStreamStatus(() => result);
	}, [])

	const fetchStatusErrorHandler = useCallback((error: FetchError) => {
		if (error.status === 503) {
			setStreamStatus(() => undefined);
		}
	}, [])

	const refreshStatus = useCallback(async () => {
		await fetchStatus(fetchStatusResultHandler, fetchStatusErrorHandler)
	}, [fetchStatusErrorHandler, fetchStatusResultHandler])

	const refreshStatusContext = useCallback(() => {
		void refreshStatus()
	}, [refreshStatus])

	const stopFetching = useCallback(() => {
		clearInterval(intervalRef.current)
		intervalRef.current = undefined
	}, [])

	const startFetching = useCallback(() => {
		if (!intervalRef.current) {
			intervalRef.current = setInterval(() => {
				void refreshStatus()
			}, STATUS_POLL_INTERVAL_MS)
		}
	}, [refreshStatus])

	const subscribe = useCallback(() => {
		subscribers.current++;

		if (subscribers.current >= 1) {
			startFetching()
		}
	}, [startFetching])

	const unsubscribe = useCallback(() => {
		subscribers.current--;

		if (subscribers.current === 0) {
			stopFetching()
		}
	}, [stopFetching])

	const setCurrentStreamStatus = useCallback((value: StreamStatus) => {
		setCurrentStreamStatusState(() => value)
	}, [])

	useEffect(() => stopFetching, [stopFetching])

	const state = useMemo<StatusProviderContextProps>(() => ({
		activeStreamsStatus: streamStatus,
		currentStreamStatus: currentStreamStatus,
		refreshStatus: refreshStatusContext,
		subscribe: subscribe,
		unsubscribe: unsubscribe,
		setCurrentStreamStatus: setCurrentStreamStatus
	}), [currentStreamStatus, refreshStatusContext, setCurrentStreamStatus, streamStatus, subscribe, unsubscribe]);

	return (
		<StatusContext.Provider value={state}>
			{props.children}
		</StatusContext.Provider>
	);
}
