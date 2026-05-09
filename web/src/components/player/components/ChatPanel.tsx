import {
	FormEvent,
	memo,
	useCallback,
	useContext,
	useEffect,
	useRef,
	useState,
} from "react";
import {
	ChatBubbleLeftRightIcon,
	PencilSquareIcon,
	PaperAirplaneIcon,
} from "@heroicons/react/24/outline";
import {
	ChatAdapter,
	ChatStatus,
	Message,
	useChatSession,
} from "../../../hooks/useChatSession";
import { LocaleContext } from "../../../providers/LocaleProvider";

const noop = () => {};

type ChatVariant = "sidebar" | "compact-below" | "below";

interface ChatPanelProps {
	streamKey: string;
	variant: ChatVariant;
	isOpen: boolean;
	adapter?: ChatAdapter;
	displayName?: string;
	onChangeDisplayNameRequested?: () => void;
}

const getNameColor = (displayName: string) => {
	let hash = 0;
	for (let i = 0; i < displayName.length; i += 1) {
		hash = displayName.charCodeAt(i) + ((hash << 5) - hash);
	}

	return `hsl(${Math.abs(hash) % 360}, 70%, 60%)`;
};

const ChatMessage = memo(function ChatMessage(props: { message: Message }) {
	const { message } = props;
	const timestamp = new Date(message.ts).toLocaleTimeString([], {
		hour: "2-digit",
		minute: "2-digit",
	});

	return (
		<div className="bg-gray-900/40 p-2">
			<div className="flex items-center gap-2 text-xs">
				<span
					className="max-w-56 truncate font-bold"
					style={{ color: getNameColor(message.displayName) }}
				>
					{message.displayName}
				</span>
				<span className="text-gray-400">{timestamp}</span>
			</div>
			<p className="mt-1 break-words text-sm text-gray-100">{message.text}</p>
		</div>
	);
});

interface ChatComposerProps {
	status: ChatStatus;
	isSending: boolean;
	onNameRequested(): void;
	onSend(text: string): Promise<boolean>;
	locale: {
		placeholder_input: string;
		button_change_display_name_title: string;
		button_send_title: string;
	};
}

const ChatComposer = memo(function ChatComposer(props: ChatComposerProps) {
	const { status, isSending, onNameRequested, onSend, locale } = props;
	const [text, setText] = useState("");
	const canSend =
		text.trim().length > 0 && !isSending && status === "connected";

	const submit = async (event: FormEvent<HTMLFormElement>) => {
		event.preventDefault();

		if (!text.trim()) {
			return;
		}

		const sent = await onSend(text);
		if (sent) {
			setText("");
		}
	};

	return (
		<form
			onSubmit={submit}
			className="border-t border-gray-700 bg-gray-900/70 p-3"
		>
			<div className="flex items-center gap-2">
				<input
					type="text"
					value={text}
					maxLength={2000}
					onChange={(event) => setText(event.target.value)}
					placeholder={locale.placeholder_input}
					className="h-9 flex-1 rounded-md border border-gray-700 bg-gray-800 px-3 text-sm text-gray-100 placeholder:text-gray-400 focus:outline-hidden"
				/>

				<button
					type="button"
					onClick={onNameRequested}
					className="inline-flex h-9 w-9 items-center justify-center rounded-md border border-gray-700 bg-gray-800 text-gray-100 hover:bg-gray-700"
					title={locale.button_change_display_name_title}
				>
					<PencilSquareIcon className="h-5 w-5" />
				</button>

				<button
					type="submit"
					disabled={!canSend}
					className="inline-flex h-9 w-9 items-center justify-center rounded-md bg-blue-600 text-white disabled:cursor-not-allowed disabled:bg-gray-700 disabled:text-gray-400"
					title={locale.button_send_title}
				>
					<PaperAirplaneIcon className="h-5 w-5" />
				</button>
			</div>

			{text.length > 1800 && (
				<div className="mt-1 text-right text-xs text-gray-400">
					{text.length}/2000
				</div>
			)}
		</form>
	);
});

const statusColorClass = (status: ChatStatus) => {
	if (status === "connected") {
		return "bg-green-500";
	}

	if (status === "connecting") {
		return "animate-pulse bg-yellow-400";
	}

	return "bg-red-500";
};

const getLocalizedStatus = (status: ChatStatus, locale: { status_connecting: string; status_connected: string; status_error: string; status_disconnected: string }) => {
	switch (status) {
		case "connecting":
			return locale.status_connecting;
		case "connected":
			return locale.status_connected;
		case "error":
			return locale.status_error;
		case "disconnected":
			return locale.status_disconnected;
		default:
			return status;
	}
};

const ChatPanel = (props: ChatPanelProps) => {
	const { streamKey, variant, isOpen, adapter, displayName, onChangeDisplayNameRequested } = props;
	const { locale } = useContext(LocaleContext);
	const { messages, status, error, sendMessage } = useChatSession(
		streamKey,
		adapter,
		locale.chat.error_failed_to_connect,
	);

	const [isSending, setIsSending] = useState(false);
	const [sendError, setSendError] = useState<string | null>(null);

	const messageListRef = useRef<HTMLDivElement>(null);
	const shouldStickToBottomRef = useRef(true);
	const firstBatchRef = useRef(true);

	useEffect(() => {
		if (!isOpen) {
			return;
		}

		const node = messageListRef.current;
		if (!node) {
			return;
		}

		if (firstBatchRef.current || shouldStickToBottomRef.current) {
			node.scrollTop = node.scrollHeight;
			firstBatchRef.current = false;
		}
	}, [isOpen, messages]);

	useEffect(() => {
		firstBatchRef.current = true;
		shouldStickToBottomRef.current = true;
	}, [streamKey]);

	const onMessageListScroll = () => {
		const node = messageListRef.current;
		if (!node) {
			return;
		}

		const distanceToBottom =
			node.scrollHeight - node.scrollTop - node.clientHeight;
		shouldStickToBottomRef.current = distanceToBottom <= 100;
	};

	const onSend = useCallback(
		async (text: string) => {
			if (!displayName?.trim()) {
				onChangeDisplayNameRequested?.();
				return false;
			}

			setIsSending(true);
			setSendError(null);

			try {
				await sendMessage(text.trim(), displayName!.trim(), locale.chat.error_not_connected);
				return true;
			} catch (nextError) {
				const message =
					nextError instanceof Error
						? nextError.message
						: locale.chat.error_failed_to_send;
				setSendError(message);
				return false;
			} finally {
				setIsSending(false);
			}
		},
		[displayName, sendMessage, onChangeDisplayNameRequested, locale.chat.error_failed_to_send, locale.chat.error_not_connected],
	);

	const base =
		"flex flex-col overflow-hidden rounded-md border border-gray-700 bg-slate-900 text-gray-100 transition-[height,max-height,width,opacity,transform,border-color] duration-200 ease-out";
	const belowHeightClass = variant === "compact-below" ? "h-80" : "h-96";
	const panelClassName = variant === "sidebar"
		? `${base} min-h-0 shrink-0 ${
			isOpen
				? "absolute top-0 right-0 h-full w-80 opacity-100"
				: "absolute top-0 right-0 h-full w-0 max-h-none translate-x-2 translate-y-0 opacity-0 pointer-events-none border-transparent"
		}`
		: `${base} ${isOpen ? `${belowHeightClass} translate-y-0 opacity-100` : "h-0 translate-y-1 border-transparent opacity-0 pointer-events-none"}`;

	return (
		<div
			className={panelClassName}
		>
			<div className="flex items-center justify-between border-b border-gray-700 bg-gray-900/70 px-3 py-2">
				<div className="flex items-center gap-2 text-sm font-semibold">
					<ChatBubbleLeftRightIcon className="h-4 w-4" />
					<span>{locale.chat.title}</span>
				</div>

				<div className="flex items-center gap-2 text-xs text-gray-300">
					<span
						className={`inline-flex h-2.5 w-2.5 rounded-full ${statusColorClass(status)}`}
					/>
					<span className="capitalize">{getLocalizedStatus(status, locale.chat)}</span>
				</div>
			</div>

			<div
				ref={messageListRef}
				onScroll={onMessageListScroll}
				style={{ colorScheme: "dark" }}
				className="min-h-0 flex-1 overflow-y-auto px-3 py-2"
			>
				{error && (
					<div className="mb-2 rounded-md border border-red-400 bg-red-950/40 px-2 py-1 text-xs text-red-200">
						{error}
					</div>
				)}
				{sendError && (
					<div className="mb-2 rounded-md border border-red-400 bg-red-950/40 px-2 py-1 text-xs text-red-200">
						{sendError}
					</div>
				)}

				{!error && messages.length === 0 && status === "connected" && (
					<div className="text-xs text-gray-400">{locale.chat.no_messages_yet}</div>
				)}

				<div className="space-y-0">
					{messages.map((message) => (
						<ChatMessage key={message.id} message={message} />
					))}
				</div>
			</div>

			<ChatComposer
				status={status}
				isSending={isSending}
				onNameRequested={onChangeDisplayNameRequested ?? noop}
				onSend={onSend}
				locale={locale.chat}
			/>
		</div>
	);
};

export default ChatPanel;
