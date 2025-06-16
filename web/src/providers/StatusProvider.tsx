import React, {useEffect, useMemo, useState} from "react";

interface WhepSession {
	currentLayer: string;
	id: string;
}

interface StatusResult {
	streamKey: string;
	whepSessions: WhepSession[];
	videoStreams: VideoStream[];
}

interface VideoStream {
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
		if (!result.ok) {
			throw new FetchError('Unknown error when calling status', result.status);
		}

		if (result.status === 503) {
			throw new FetchError('Status API disabled', result.status);
		}

		return result.json()
	})
		.then((result: StatusResult[]) => onSuccess?.(result))
		.catch((err: FetchError) => onError?.(err));

interface StatusProviderContextProps {
	streamStatus: StatusResult[]
}

export const StatusContext = React.createContext<StatusProviderContextProps>({
	streamStatus: []
});
export function StatusProvider(props: StatusProviderProps) {
	const [streamStatus, setStreamStatus] = useState<StatusResult[]>([])

	useEffect(() => {
		const intervalHandler = async () => {
			await fetchStatus(
				(result) => setStreamStatus(_ => result),
				(errorMessage) => console.error("StatusProviderError", errorMessage.status, errorMessage.message))
		}

		const interval = setInterval(intervalHandler, 5000)

		return () => clearInterval(interval)
	}, []);

	const state = useMemo<StatusProviderContextProps>(() => ({
		streamStatus: streamStatus
	}), [streamStatus]);

	return (
		<StatusContext.Provider value={state}>
			{props.children}
		</StatusContext.Provider>
	);
}
