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

	videoTracks: VideoTrack[];
	audioTracks: AudioTrack[];

	sessions: WhepSession[];
}

interface VideoTrack {
	rid: string;
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
	const [isStatusActive, setIsStatusActive] = useState<boolean>(false)
	const [streamStatus, setStreamStatus] = useState<StatusResult[] | undefined>(undefined)
	const [currentStreamStatus, setCurrentStreamStatus] = useState<StreamStatus | undefined>(undefined)
	const intervalCountRef = useRef<number>(5000);
	const intervalRef = useRef<number | undefined>(0)
	const subscribers = useRef<number>(0)

	const fetchStatusResultHandler = (result: StatusResult[]) => {
		setStreamStatus(() => result);
	}
	const fetchStatusErrorHandler = (error: FetchError) => {
		if (error.status === 503) {
			setIsStatusActive(() => false)
			setStreamStatus(() => undefined);
		}
	}

	const subscribe = () => {
		subscribers.current++;

		if (subscribers.current >= 1) {
			startFetching()
		}
	}

	const unsubscribe = () => {
		subscribers.current--;

		if (subscribers.current == 0) {
			clearInterval(intervalRef.current)
			intervalRef.current = undefined
		}
	}

	const startFetching = () => {
		if (!intervalRef.current) {
			const intervalHandler = async () => {
				await fetchStatus(
					fetchStatusResultHandler,
					fetchStatusErrorHandler)
			}

			intervalRef.current = setInterval(intervalHandler, intervalCountRef.current)
		}
	}

	useEffect(() => {
		if (!isStatusActive) {
			return
		}

		const intervalHandler = async () => {
			await fetchStatus(
				fetchStatusResultHandler,
				fetchStatusErrorHandler)
		}

		intervalRef.current = setInterval(intervalHandler, intervalCountRef.current)
		return () => clearInterval(intervalRef.current)
	}, [isStatusActive]);

	const state = useMemo<StatusProviderContextProps>(() => ({
		activeStreamsStatus: streamStatus,
		currentStreamStatus: currentStreamStatus,
		refreshStatus: async () => {
			await fetchStatus(
				fetchStatusResultHandler,
				fetchStatusErrorHandler)
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
