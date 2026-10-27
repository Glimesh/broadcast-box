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
import { Account, AccountContext } from "../../../providers/AccountProvider";
import { LocaleContext } from "../../../providers/LocaleProvider";

const noop = () => {};

type ChatVariant = "sidebar" | "below" | "docked";

interface CustomEmote {
  name: string;
  url: string;
  animated: boolean;
}

const useCustomEmotes = (streamKey: string) => {
  const [emotes, setEmotes] = useState<CustomEmote[]>([]);

  useEffect(() => {
    let cancelled = false;

    fetch(`/api/emotes?streamKey=${encodeURIComponent(streamKey)}`)
      .then((result) => (result.ok ? result.json() : []))
      .then((result: CustomEmote[]) => {
        if (!cancelled && Array.isArray(result)) {
          setEmotes(result);
        }
      })
      .catch(() => {});

    return () => {
      cancelled = true;
    };
  }, [streamKey]);

  return emotes;
};

const renderMessageText = (text: string, emotes: CustomEmote[]) => {
  if (emotes.length === 0) {
    return text;
  }

  const emoteMap = new Map(emotes.map((emote) => [emote.name.toLowerCase(), emote]));
  const parts = text.split(/(:[a-zA-Z0-9_-]+:)/g);

  return parts.map((part, index) => {
    const match = /^:([a-zA-Z0-9_-]+):$/.exec(part);
    if (!match) {
      return part;
    }

    const emote = emoteMap.get(match[1].toLowerCase());
    if (!emote) {
      return part;
    }

    return (
      <img
        key={`${emote.name}_${index}`}
        src={emote.url}
        alt={`:${emote.name}:`}
        title={`:${emote.name}:`}
      />
    );
  });
};

interface ChatPanelProps {
  streamKey: string;
  variant: ChatVariant;
  isOpen: boolean;
  adapter?: ChatAdapter;
  displayName?: string;
  onChangeDisplayNameRequested?: () => void;
  onClose?: () => void;
}

const getNameColor = (displayName: string) => {
  let hash = 0;
  for (let i = 0; i < displayName.length; i += 1) {
    hash = displayName.charCodeAt(i) + ((hash << 5) - hash);
  }

  return `hsl(${Math.abs(hash) % 360}, 62%, 38%)`;
};

const ChatMessage = memo(function ChatMessage(props: {
  message: Message;
  emotes: CustomEmote[];
  canModerate: boolean;
  onDelete: (messageId: string) => void;
}) {
  const { message, emotes, canModerate, onDelete } = props;
  const timestamp = new Date(message.ts).toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
  });

  return (
    <div className="chat-message">
      <div className="chat-message__meta">
        <span
          className="chat-message__name"
          style={{ color: getNameColor(message.displayName) }}
        >
          {message.displayName}
        </span>
        <span className="chat-message__time">{timestamp}</span>
        {canModerate && (
          <button
            type="button"
            className="chat-message__delete"
            onClick={() => onDelete(message.id)}
            title="delete message"
          >
            ✕
          </button>
        )}
      </div>
      <p className="chat-message__text">{renderMessageText(message.text, emotes)}</p>
    </div>
  );
});

interface ChatComposerProps {
  status: ChatStatus;
  isSending: boolean;
  onNameRequested(): void;
  onSend(text: string): Promise<boolean>;
  customEmotes: CustomEmote[];
  locale: {
    placeholder_input: string;
    button_change_display_name_title: string;
    button_send_title: string;
  };
}

const EMOTES = ["😀", "😂", "😍", "🤔", "😎", "👍", "👎", "❤️", "🔥", "🎉", "😭", "😮", "🙏", "👏", "💯", "🚀"];

const ChatComposer = memo(function ChatComposer(props: ChatComposerProps) {
  const { status, isSending, onNameRequested, onSend, customEmotes, locale } = props;
  const [text, setText] = useState("");
  const [showEmotes, setShowEmotes] = useState(false);
  const canSend = text.trim().length > 0 && !isSending && status === "connected";

  const insertEmote = (emote: string) => {
    setText((prev) => prev + emote);
  };

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
    <form onSubmit={submit} className="chat-composer">
      {showEmotes && (
        <div className="chat-emotes">
          {customEmotes.length > 0 && (
            <>
              <div className="chat-emotes__label">Custom emotes</div>
              <div className="chat-emotes__grid">
                {customEmotes.map((emote) => (
                  <button
                    key={emote.name}
                    type="button"
                    onClick={() => insertEmote(`:${emote.name}:`)}
                    className="chat-emote-btn"
                    title={`:${emote.name}:`}
                  >
                    <img src={emote.url} alt={emote.name} />
                  </button>
                ))}
              </div>
            </>
          )}
          <div className="chat-emotes__label">Emoji</div>
          <div className="chat-emotes__grid chat-emotes__grid--emoji">
            {EMOTES.map((emote) => (
              <button
                key={emote}
                type="button"
                onClick={() => insertEmote(emote)}
                className="chat-emote-btn chat-emote-btn--emoji"
              >
                {emote}
              </button>
            ))}
          </div>
        </div>
      )}

      <div className="chat-compose-row">
        <button
          type="button"
          onClick={() => setShowEmotes((prev) => !prev)}
          className="chat-icon-btn"
          title="Emotes"
        >
          ☺
        </button>

        <input
          type="text"
          value={text}
          maxLength={2000}
          onChange={(event) => setText(event.target.value)}
          placeholder={locale.placeholder_input}
          className="chat-input"
        />

        <button
          type="button"
          onClick={onNameRequested}
          className="chat-icon-btn"
          title={locale.button_change_display_name_title}
        >
          <PencilSquareIcon className="h-4 w-4" />
        </button>

        <button
          type="submit"
          disabled={!canSend}
          className="chat-icon-btn chat-icon-btn--send"
          title={locale.button_send_title}
        >
          <PaperAirplaneIcon className="h-4 w-4" />
        </button>
      </div>

      {text.length > 1800 && (
        <div className="chat-count">{text.length}/2000</div>
      )}
    </form>
  );
});

const statusColorClass = (status: ChatStatus) => {
  if (status === "connected") {
    return "chat-dot--connected";
  }

  if (status === "connecting") {
    return "chat-dot--connecting";
  }

  return "chat-dot--error";
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
  const { streamKey, variant, isOpen, adapter, displayName, onChangeDisplayNameRequested, onClose } = props;
  const { locale } = useContext(LocaleContext);
  const { account } = useContext(AccountContext);
  const [channelAccount, setChannelAccount] = useState<Account | null>(null);
  const customEmotes = useCustomEmotes(streamKey);

  useEffect(() => {
    let cancelled = false;

    fetch(`/api/accounts/${encodeURIComponent(streamKey)}`)
      .then((result) => (result.ok ? result.json() : null))
      .then((result: Account | null) => {
        if (!cancelled) {
          setChannelAccount(result);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setChannelAccount(null);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [streamKey]);

  const canModerate =
    !!account &&
    (account.username === channelAccount?.username ||
      (channelAccount?.chatMods ?? []).includes(account.displayName));
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

    const distanceToBottom = node.scrollHeight - node.scrollTop - node.clientHeight;
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

  const variantClass = variant === "docked"
    ? "chat-panel--docked"
    : variant === "below"
      ? "chat-panel--below"
      : "";

  return (
    <div className={`chat-panel ${variantClass} ${!isOpen ? "chat-panel--closed" : ""}`}>
      <div className="chat-head">
        <div className="chat-head__title">
          <ChatBubbleLeftRightIcon className="h-4 w-4" />
          <span>{locale.chat.title}</span>
        </div>

        <div className="chat-head__status">
          <span className={`chat-dot ${statusColorClass(status)}`} />
          <span>{getLocalizedStatus(status, locale.chat)}</span>
          {onClose && (
            <button
              type="button"
              onClick={onClose}
              className="chat-close"
              title="close chat"
            >
              ✕
            </button>
          )}
        </div>
      </div>

      <div
        ref={messageListRef}
        onScroll={onMessageListScroll}
        className="chat-log"
      >
        {error && <div className="chat-notice">{error}</div>}
        {sendError && <div className="chat-notice">{sendError}</div>}

        {!error && messages.length === 0 && status === "connected" && (
          <div className="chat-empty">{locale.chat.no_messages_yet}</div>
        )}

        <div>
          {messages.map((message) => (
            <ChatMessage
              key={message.id}
              message={message}
              emotes={customEmotes}
              canModerate={canModerate}
              onDelete={(messageId) => {
                void fetch("/api/chat/delete", {
                  method: "POST",
                  headers: { "Content-Type": "application/json" },
                  body: JSON.stringify({ streamKey, messageId }),
                });
              }}
            />
          ))}
        </div>
      </div>

      <ChatComposer
        status={status}
        isSending={isSending}
        onNameRequested={onChangeDisplayNameRequested ?? noop}
        onSend={onSend}
        customEmotes={customEmotes}
        locale={locale.chat}
      />
    </div>
  );
};

export default ChatPanel;
