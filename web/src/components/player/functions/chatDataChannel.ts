import {
	ChatAdapter,
	ChatStatus,
	Message,
} from "../../../hooks/useChatSession";

const DATA_CHANNEL_LABEL = "bb-chat-v1";

type ChatInbound = {
	type: "chat.send";
	clientMsgId: string;
	text: string;
	displayName: string;
};

type ChatHistoryEvent = {
	type: "message";
	message: Message;
};

type ChatOutbound =
	| { type: "chat.connected" }
	| { type: "chat.history"; events: ChatHistoryEvent[] }
	| { type: "chat.message"; eventId: number; message: Message }
	| { type: "chat.ack"; clientMsgId: string }
	| { type: "chat.error"; error: string; clientMsgId?: string };

interface PendingMessage {
	resolve(): void;
	reject(error: Error): void;
	timeout: ReturnType<typeof setTimeout>;
}

const createMessageId = () => {
	if (
		typeof crypto !== "undefined" &&
		typeof crypto.randomUUID === "function"
	) {
		return crypto.randomUUID();
	}

	return `${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;
};

export class ChatDataChannelAdapter implements ChatAdapter {
	private streamKey = "";
	private channel: RTCDataChannel | null = null;
	private pending = new Map<string, PendingMessage>();

	private onMessage: ((message: Message) => void) | null = null;
	private onStatus: ((status: ChatStatus) => void) | null = null;
	private onError: ((error: string) => void) | null = null;

	private handleOpen = () => {
		this.onStatus?.("connecting");
	};

	private handleClose = () => {
		this.onStatus?.("disconnected");
		this.rejectPending("Chat disconnected");
	};

	private handleError = () => {
		this.onStatus?.("disconnected");
		this.onError?.("Chat data channel error");
		this.rejectPending("Chat data channel error");
	};

	private handleMessage = (event: MessageEvent) => {
		let payload: ChatOutbound;

		try {
			payload = JSON.parse(event.data) as ChatOutbound;
		} catch {
			this.onError?.("Invalid chat payload");
			return;
		}

		switch (payload.type) {
			case "chat.connected":
				this.onStatus?.("connected");
				return;
			case "chat.history":
				payload.events.forEach((entry) => {
					if (entry.type === "message") {
						this.onMessage?.(entry.message);
					}
				});
				return;
			case "chat.message":
				this.onMessage?.(payload.message);
				return;
			case "chat.ack": {
				const pendingItem = this.pending.get(payload.clientMsgId);
				if (pendingItem) {
					clearTimeout(pendingItem.timeout);
					pendingItem.resolve();
					this.pending.delete(payload.clientMsgId);
				}
				return;
			}
			case "chat.error": {
				this.onError?.(payload.error);
				this.onStatus?.("error");

				if (payload.clientMsgId) {
					const pendingItem = this.pending.get(payload.clientMsgId);
					if (pendingItem) {
						clearTimeout(pendingItem.timeout);
						pendingItem.reject(new Error(payload.error));
						this.pending.delete(payload.clientMsgId);
					}
				}

				return;
			}
			default:
				this.onError?.("Unsupported chat payload");
		}
	};

	attachChannel(channel: RTCDataChannel) {
		this.detachChannel();

		this.channel = channel;
		this.channel.addEventListener("open", this.handleOpen);
		this.channel.addEventListener("close", this.handleClose);
		this.channel.addEventListener("error", this.handleError);
		this.channel.addEventListener("message", this.handleMessage);

		this.onStatus?.("connecting");
	}

	detachChannel() {
		if (!this.channel) {
			return;
		}

		this.rejectPending("Chat disconnected");

		this.channel.removeEventListener("open", this.handleOpen);
		this.channel.removeEventListener("close", this.handleClose);
		this.channel.removeEventListener("error", this.handleError);
		this.channel.removeEventListener("message", this.handleMessage);
		this.channel = null;
	}

	private rejectPending(reason: string) {
		this.pending.forEach((item) => {
			clearTimeout(item.timeout);
			item.reject(new Error(reason));
		});
		this.pending.clear();
	}

	async connect(streamKey: string): Promise<void> {
		this.streamKey = streamKey;
	}

	subscribe(
		onMessage: (message: Message) => void,
		onStatus: (status: ChatStatus) => void,
		onError: (error: string) => void,
	): () => void {
		this.onMessage = onMessage;
		this.onStatus = onStatus;
		this.onError = onError;

		if (this.channel && this.channel.readyState !== "closed") {
			this.onStatus("connecting");
		}

		return () => {
			if (this.onMessage === onMessage) {
				this.onMessage = null;
			}

			if (this.onStatus === onStatus) {
				this.onStatus = null;
			}

			if (this.onError === onError) {
				this.onError = null;
			}
		};
	}

	async send(text: string, displayName: string): Promise<void> {
		if (!this.channel || this.channel.readyState !== "open") {
			throw new Error("Chat data channel is not open");
		}

		if (!this.streamKey) {
			throw new Error("Chat is not connected to stream");
		}

		const clientMsgId = createMessageId();
		const payload: ChatInbound = {
			type: "chat.send",
			clientMsgId,
			text,
			displayName,
		};

		const rawPayload = JSON.stringify(payload);

		return new Promise<void>((resolve, reject) => {
			const timeout = setTimeout(() => {
				this.pending.delete(clientMsgId);
				reject(new Error("Chat message timed out"));
			}, 10_000);

			this.pending.set(clientMsgId, { resolve, reject, timeout });

			try {
				this.channel?.send(rawPayload);
			} catch (error) {
				clearTimeout(timeout);
				this.pending.delete(clientMsgId);
				reject(
					error instanceof Error
						? error
						: new Error("Failed to send chat message"),
				);
			}
		});
	}
}

export { DATA_CHANNEL_LABEL };
