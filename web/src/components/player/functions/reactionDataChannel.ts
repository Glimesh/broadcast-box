const DATA_CHANNEL_LABEL = "bb-data-v1";

export type ReactionStatus = "connecting" | "connected" | "error" | "disconnected";

export interface ReactionAdapter {
	connect(streamKey: string): Promise<void>;
	subscribe(
		onReaction: () => void,
		onStatus: (status: ReactionStatus) => void,
		onError: (error: string) => void,
	): () => void;
	send(): Promise<void>;
}

export class ReactionDataChannelAdapter implements ReactionAdapter {
	private isConnectedToStream = false;
	private channel: RTCDataChannel | null = null;
	private onReaction: (() => void) | null = null;
	private onStatus: ((status: ReactionStatus) => void) | null = null;
	private onError: ((error: string) => void) | null = null;

	private setDisconnected() {
		this.onStatus?.("disconnected");
	}

	private setError(errorMessage: string) {
		this.onStatus?.("error");
		this.onError?.(errorMessage);
	}

	private handleOpen = () => {
		this.onStatus?.("connected");
	};

	private handleClose = () => {
		this.setDisconnected();
	};

	private handleError = () => {
		this.setError("Reaction data channel error");
	};

	private handleMessage = (event: MessageEvent) => {
		if (typeof event.data !== "string") {
			return;
		}

		try {
			const payload = JSON.parse(event.data) as { type?: string };
			if (payload.type === "reaction") {
				this.onReaction?.();
			}
		} catch {
			return;
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

		this.channel.removeEventListener("open", this.handleOpen);
		this.channel.removeEventListener("close", this.handleClose);
		this.channel.removeEventListener("error", this.handleError);
		this.channel.removeEventListener("message", this.handleMessage);
		this.channel = null;
		this.isConnectedToStream = false;
	}

	async connect(streamKey: string): Promise<void> {
		this.isConnectedToStream = streamKey.length > 0;
	}

	subscribe(
		onReaction: () => void,
		onStatus: (status: ReactionStatus) => void,
		onError: (error: string) => void,
	): () => void {
		this.onReaction = onReaction;
		this.onStatus = onStatus;
		this.onError = onError;

		if (this.channel && this.channel.readyState !== "closed") {
			this.onStatus(this.channel.readyState === "open" ? "connected" : "connecting");
		}

		return () => {
			if (this.onReaction === onReaction) {
				this.onReaction = null;
			}

			if (this.onStatus === onStatus) {
				this.onStatus = null;
			}

			if (this.onError === onError) {
				this.onError = null;
			}
		};
	}

	async send(): Promise<void> {
		if (!this.channel || this.channel.readyState !== "open") {
			throw new Error("Reaction data channel is not open");
		}

		if (!this.isConnectedToStream) {
			throw new Error("Reaction data channel is not connected to stream");
		}

		this.channel.send(JSON.stringify({ type: "reaction" }));
	}
}

export { DATA_CHANNEL_LABEL };
