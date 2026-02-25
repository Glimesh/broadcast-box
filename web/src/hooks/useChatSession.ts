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

export const useChatSession = (streamKey: string, adapter?: ChatAdapter, connectionErrorMessage?: string) => {
	const [messages, setMessages] = useState<Message[]>([]);
	const [status, setStatus] = useState<ChatStatus>("disconnected");
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		setMessages([]);
		setError(null);

		if (!adapter) {
			setStatus("disconnected");
			return;
		}

		let unsubscribe = () => {};
		let stopped = false;

		setStatus("connecting");

		unsubscribe = adapter.subscribe(
			(message) => {
				if (stopped) {
					return;
				}

				setMessages((current) => {
					if (current.some((existing) => existing.id === message.id)) {
						return current;
					}

					const next = [...current, message];
					if (next.length <= MAX_MESSAGES) {
						return next;
					}

					return next.slice(next.length - MAX_MESSAGES);
				});
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
