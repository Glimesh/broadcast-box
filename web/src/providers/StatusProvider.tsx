import React, { useEffect, useMemo, useRef, useState } from "react";

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

	videoTracks: VideoTrack[];
	audioTracks: AudioTrack[];

	sessions: WhepSession[];
}

interface VideoTrack {
	rid: string;
	packetsReceived: number;
	lastKeyframe: string;
}

interface AudioTrack {
	rid: string;
	packetsReceived: number;
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
	streamStatus: StatusResult[] | undefined
	refreshStatus: () => void
}

export const StatusContext = React.createContext<StatusProviderContextProps>({
	streamStatus: undefined,
	refreshStatus: () => { }
});

export function StatusProvider(props: StatusProviderProps) {
	const [isStatusActive, setIsStatusActive] = useState<boolean>(false)
	const [streamStatus, setStreamStatus] = useState<StatusResult[] | undefined>(undefined)
	const intervalCountRef = useRef<number>(5000);
	const intervalRef = useRef<number>(0)

	const fetchStatusResultHandler = (result: StatusResult[]) => {
		setStreamStatus(_ => result);
	}
	const fetchStatusErrorHandler = (error: FetchError) => {
		if (error.status === 503) {
			setIsStatusActive(() => false)
			setStreamStatus(() => undefined);
		}
	}

	useEffect(() => {
		fetchStatus(
			(result) => {
				setStreamStatus(_ => result)
				setIsStatusActive(_ => true)
			},
			(error) => {
				if (error.status === 503) {
					setIsStatusActive(() => false)
					setStreamStatus(() => undefined);
				}

				console.error("StatusProviderError", error.status, error.message)
			})
			.catch((err) => console.error("StatusProviderError", err))

		return () => {
			clearInterval(intervalRef.current)
		}
	}, []);

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
		streamStatus: streamStatus,
		refreshStatus: async () => {
			await fetchStatus(
				fetchStatusResultHandler,
				fetchStatusErrorHandler)
		}
	}), [streamStatus]);

	return (
		<StatusContext.Provider value={state}>
			{props.children}
		</StatusContext.Provider>
	);
}
