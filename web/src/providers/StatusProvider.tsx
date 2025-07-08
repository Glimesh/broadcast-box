import React, {useEffect, useMemo, useRef, useState} from "react";

interface WhepSession {
	id: string;
	currentLayer: string;
	sequenceNumber: number;
	timestamp: number;
	packetsWritten: number;
}

interface StatusResult {
	streamKey: string;
	whepSessions: WhepSession[];
	videoStreams: VideoStream[];
}

interface VideoStream {
	rid: string;
	packetsReceived: number;
	lastKeyFrameSeen: string;
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

const apiPath = import.meta.env.VITE_API_PATH;
const fetchStatus = (
	onSuccess?: (statusResults: StatusResult[]) => void,
	onError?: (error: FetchError) => void
) =>
	fetch(`${apiPath}/status`, {
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
	
	const fetchStatusResultHandler = (result: StatusResult[]) => {
		setStreamStatus(_ => result);
	}
	const fetchStatusErrorHandler = (error: FetchError) => {
		console.error("StatusProviderError", error.status, error.message)

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

		const interval = setInterval(intervalHandler, intervalCountRef.current)
		return () => clearInterval(interval)
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