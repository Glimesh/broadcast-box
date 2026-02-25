import { useCallback, useEffect, useState } from "react";

export type ChatStatus = "connecting" | "connected" | "error" | "disconnected";

export interface Message {
	id: string;
	ts: number;
	text: string;
	displayName: string;
}

export interface ChatAdapter {
	connect(streamKey: string): Promise<void>;
	subscribe(
		onMessage: (message: Message) => void,
		onStatus: (status: ChatStatus) => void,
		onError: (error: string) => void,
	): () => void;
	send(text: string, displayName: string): Promise<void>;
}

const MAX_MESSAGES = 1000;

const appendUniqueCapped = (current: Message[], message: Message): Message[] => {
	if (current.some((existing) => existing.id === message.id)) {
		return current;
	}

	const next = [...current, message];
	if (next.length <= MAX_MESSAGES) {
		return next;
	}

	return next.slice(next.length - MAX_MESSAGES);
};

export const useChatSession = (streamKey: string, adapter?: ChatAdapter, connectionErrorMessage?: string) => {
	const [messages, setMessages] = useState<Message[]>([]);
	const [status, setStatus] = useState<ChatStatus>(adapter ? "connecting" : "disconnected");
	const [error, setError] = useState<string | null>(null);

	const [session, setSession] = useState({ adapter, streamKey });
	if (session.adapter !== adapter || session.streamKey !== streamKey) {
		setSession({ adapter, streamKey });
		setMessages([]);
		setError(null);
		setStatus(adapter ? "connecting" : "disconnected");
	}

	useEffect(() => {
		if (!adapter) {
			return;
		}

		let unsubscribe = () => {};
		let stopped = false;

		unsubscribe = adapter.subscribe(
			(message) => {
				if (stopped) {
					return;
				}

				setMessages((current) => appendUniqueCapped(current, message));
			},
			(nextStatus) => {
				if (!stopped) {
					setStatus(nextStatus);
				}
			},
			(nextError) => {
				if (!stopped) {
					setError(nextError);
					setStatus("error");
				}
			},
		);

		adapter.connect(streamKey).catch((connectError) => {
			if (!stopped) {
				const message =
					connectError instanceof Error
						? connectError.message
						: connectionErrorMessage ?? "Failed to connect chat";
				setError(message);
				setStatus("error");
			}
		});

		return () => {
			stopped = true;
			unsubscribe();
		};
	}, [adapter, streamKey, connectionErrorMessage]);

	const sendMessage = useCallback(
		async (text: string, displayName: string, notConnectedErrorMessage?: string) => {
			if (!adapter) {
				throw new Error(notConnectedErrorMessage ?? "Chat is not connected");
			}

			setError(null);
			await adapter.send(text, displayName);
		},
		[adapter],
	);

	return { messages, status, error, sendMessage };
};
