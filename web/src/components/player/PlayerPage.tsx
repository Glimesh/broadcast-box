import { useCallback, useContext, useEffect, useState } from "react";
import Player from "./Player";
import { Link, useNavigate } from "react-router-dom";
import { CinemaModeContext } from "../../providers/CinemaModeProvider";
import ModalTextInput from "../shared/ModalTextInput";
import Button from "../shared/Button";
import AvailableStreams from "../selection/AvailableStreams";
import { LocaleContext } from "../../providers/LocaleProvider";
import ChatPanel from "./components/ChatPanel";
import { ChatAdapter } from "../../hooks/useChatSession";
import { StreamMOTD } from "./components/StreamMOTD";
import { StreamStatus } from "../../providers/StatusProvider";
import { Account, AccountContext } from "../../providers/AccountProvider";
import toBase64Utf8 from "../../utilities/base64";

const PlayerPage = () => {
  const navigate = useNavigate();
  const { locale } = useContext(LocaleContext);
  const { account } = useContext(AccountContext);
  const { cinemaMode, toggleCinemaMode } = useContext(CinemaModeContext);
  const [streamKeys, setStreamKeys] = useState<string[]>([window.location.pathname.substring(1)]);
  const [isModalOpen, setIsModelOpen] = useState<boolean>(false);
  const [isChatOpen, setIsChatOpen] = useState<boolean>(() => localStorage.getItem("chat-open") !== "false");
  const [chatAdapters, setChatAdapters] = useState<Record<string, ChatAdapter | undefined>>({});
  const [streamStatuses, setStreamStatuses] = useState<Record<string, StreamStatus | undefined>>({});
  const [isDisplayNameModalOpen, setIsDisplayNameModalOpen] = useState<boolean>(false);
  const [chatDisplayName, setChatDisplayName] = useState<string>(() => localStorage.getItem("chatDisplayName") ?? "");
  const [isPublic, setIsPublic] = useState<boolean>(true);
  const [streamTitle, setStreamTitle] = useState<string>("");
  const [isTitleModalOpen, setIsTitleModalOpen] = useState<boolean>(false);
  const [titleError, setTitleError] = useState<string>("");
  const [channelAccount, setChannelAccount] = useState<Account | null>(null);
  const primaryStreamKey = streamKeys[0] ?? "";
  const isAccountOwner = !!channelAccount && account?.username === channelAccount.username;
  const profileAuthToken = isAccountOwner && account?.streamToken ? account.streamToken : primaryStreamKey;
  const effectiveChatDisplayName = chatDisplayName || account?.displayName || "";

  useEffect(() => {
    localStorage.setItem("chat-open", String(isChatOpen));
  }, [isChatOpen]);

  const addStream = (streamKey: string) => {
    if (streamKeys.some((key: string) => key.toLowerCase() === streamKey.toLowerCase())) {
      return;
    }
    setStreamKeys((prev) => [...prev, streamKey]);
    setIsModelOpen((prev) => !prev);
  };

  const setStreamChatAdapter = useCallback((streamKey: string, adapter: ChatAdapter | undefined) => {
    setChatAdapters((current) => {
      if (current[streamKey] === adapter) {
        return current;
      }

      return {
        ...current,
        [streamKey]: adapter,
      };
    });
  }, []);

  const setStreamStatus = useCallback((streamKey: string, status: StreamStatus | undefined) => {
    setStreamStatuses((current) => {
      if (current[streamKey] === status) {
        return current;
      }

      return {
        ...current,
        [streamKey]: status,
      };
    });

    if (streamKey === primaryStreamKey && status?.motd) {
      setStreamTitle(status.motd);
    }
  }, [primaryStreamKey]);

  const removeStream = (streamKey: string) => {
    setStreamKeys((prev) => prev.filter((key) => key !== streamKey));
    setChatAdapters((current) => {
      const next = { ...current };
      delete next[streamKey];
      return next;
    });
    setStreamStatuses((current) => {
      const next = { ...current };
      delete next[streamKey];
      return next;
    });
  };

  const saveDisplayName = useCallback((value: string) => {
    const trimmedValue = value.trim();
    if (!trimmedValue) {
      return;
    }
    setChatDisplayName(trimmedValue);
    localStorage.setItem("chatDisplayName", trimmedValue);
    setIsDisplayNameModalOpen(false);
  }, []);

  useEffect(() => {
    if (!primaryStreamKey) return;
    fetch(`/api/whip/profile`, {
      method: "GET",
      headers: { Authorization: `Bearer ${toBase64Utf8(profileAuthToken)}` },
    })
      .then((r) => (r.ok ? r.json() : null))
      .then((profile) => {
        if (!profile) return;
        if (typeof profile.isPublic === "boolean") {
          setIsPublic(profile.isPublic);
        }
        if (typeof profile.motd === "string") {
          setStreamTitle(profile.motd);
        }
      })
      .catch(() => {});
  }, [primaryStreamKey, profileAuthToken]);

  useEffect(() => {
    let cancelled = false;

    if (!primaryStreamKey) {
      return;
    }

    fetch(`/api/accounts/${encodeURIComponent(primaryStreamKey)}`)
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
  }, [primaryStreamKey]);

  const toggleVisibility = useCallback(() => {
    if (!primaryStreamKey) return;
    const newIsPublic = !isPublic;
    fetch(`/api/whip/profile`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${toBase64Utf8(profileAuthToken)}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ motd: "", isPublic: newIsPublic }),
    })
      .then((r) => {
        if (r.ok) {
          setIsPublic(newIsPublic);
        }
      })
      .catch(() => {});
  }, [primaryStreamKey, profileAuthToken, isPublic]);

  const saveStreamTitle = useCallback((value: string) => {
    const trimmedValue = value.trim();
    if (!trimmedValue) {
      setTitleError(locale.player_page.error_edit_stream_empty);
      return;
    }

    fetch(`/api/whip/profile`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${toBase64Utf8(profileAuthToken)}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ motd: trimmedValue, isPublic }),
    })
      .then((result) => {
        if (!result.ok) {
          setTitleError(locale.player_page.error_edit_stream_title);
          return;
        }

        setStreamTitle(trimmedValue);
        setTitleError("");
        setIsTitleModalOpen(false);
      })
      .catch(() => {
        setTitleError(locale.player_page.error_edit_stream_title);
      });
  }, [
    isPublic,
    locale.player_page.error_edit_stream_empty,
    locale.player_page.error_edit_stream_title,
    profileAuthToken,
  ]);

  const isSingleStream = streamKeys.length === 1;
  const primaryStatus = streamStatuses[primaryStreamKey];

  const copyLink = () => {
    const url = window.location.href;
    if (navigator.clipboard) {
      void navigator.clipboard.writeText(url);
    }
  };

  const streamInfoBar = (
    <div className="channel-bar">
      <div className="channel-identity">
        {channelAccount?.avatarUrl ? (
          <span className="avatar-pixel channel-avatar">
            <img src={channelAccount.avatarUrl} alt={channelAccount.displayName} />
          </span>
        ) : (
          <span className="avatar-pixel channel-avatar">
            {(channelAccount?.displayName ?? primaryStreamKey).charAt(0).toUpperCase()}
          </span>
        )}
        <div className="channel-copy">
          <div className="channel-title">
            <h2>
              {channelAccount?.displayName ?? ((primaryStatus?.motd && primaryStatus.motd.trim()) ? primaryStatus.motd : primaryStreamKey)}
            </h2>
            <span className={`status-pill ${primaryStatus?.isOnline ? "status-pill--live" : ""}`}>
              {primaryStatus?.isOnline ? "live ♡" : "offline"}
            </span>
          </div>
          <div className="channel-meta">
            {channelAccount ? (
              <Link to={`/u/${channelAccount.username}`} className="pixel-link">@{channelAccount.username}</Link>
            ) : (
              <span>{primaryStreamKey}</span>
            )}
            <span>·</span>
            <span>♡ {primaryStatus?.viewers ?? 0} watching</span>
          </div>
        </div>
      </div>

      <div className="channel-actions">
        <Button title={isChatOpen ? 'close chat ♡' : 'open chat ♡'} onClick={() => setIsChatOpen((prev) => !prev)} />
        <Button title='copy link ♡' iconRight='Copy' onClick={copyLink} />
        <Button
          title={cinemaMode ? locale.player_page.cinema_mode_disable : locale.player_page.cinema_mode_enable}
          onClick={toggleCinemaMode}
          iconRight="CodeBracketSquare"
        />
        {isAccountOwner && (
          <Button
            title={isPublic ? locale.player_page.visibility_private : locale.player_page.visibility_public}
            onClick={toggleVisibility}
            iconRight="Eye"
          />
        )}
        <Button
          title={locale.player_page.modal_add_stream_title}
          color="Accept"
          onClick={() => setIsModelOpen((prev) => !prev)}
          iconRight="SquaresPlus"
        />
      </div>
    </div>
  );

  return (
    <div className="page-container animate-slide-up">
      {isModalOpen && (
        <ModalTextInput<string>
          title={locale.player_page.modal_add_stream_title}
          message={locale.player_page.modal_add_stream_message}
          placeholder={locale.player_page.modal_add_stream_placeholder}
          isOpen={isModalOpen}
          canCloseOnBackgroundClick={true}
          onClose={() => setIsModelOpen(false)}
          onAccept={(result: string) => addStream(result)}
        >
          <AvailableStreams
            showHeader={false}
            onClickOverride={(streamKey) => addStream(streamKey)}
          />
        </ModalTextInput>
      )}

      {isDisplayNameModalOpen && (
        <ModalTextInput<string>
          title={locale.chat.modal_display_name_title}
          message={locale.chat.modal_display_name_message}
          placeholder={locale.chat.modal_display_name_placeholder}
          isOpen={isDisplayNameModalOpen}
          canCloseOnBackgroundClick
          onClose={() => setIsDisplayNameModalOpen(false)}
          onAccept={saveDisplayName}
        />
      )}

      {isTitleModalOpen && (
        <ModalTextInput<string>
          title={locale.player_page.modal_edit_stream_title}
          message={locale.player_page.modal_edit_stream_message}
          placeholder={locale.player_page.modal_edit_stream_placeholder}
          initialValue={streamTitle || primaryStatus?.motd || ""}
          errorTitle={locale.player_page.modal_edit_stream_title}
          errorMessage={titleError}
          isOpen={isTitleModalOpen}
          canCloseOnBackgroundClick
          onClose={() => {
            setIsTitleModalOpen(false);
            setTitleError("");
          }}
          onAccept={saveStreamTitle}
        />
      )}

      {isSingleStream ? (
        <div className={`player-layout ${cinemaMode ? "player-layout--cinema" : ""}`}>
          <div className="player-main">
            <StreamMOTD
              isOnline={primaryStatus?.isOnline ?? false}
              motd={streamTitle || primaryStatus?.motd || ""}
              onEdit={isAccountOwner ? () => {
                setTitleError("");
                setIsTitleModalOpen(true);
              } : undefined}
            />

            <Player
              key={`${primaryStreamKey}_player`}
              streamKey={primaryStreamKey}
              cinemaMode={cinemaMode}
              isOnline={primaryStatus?.isOnline ?? false}
              isChatOpen={isChatOpen}
              onToggleChat={() => setIsChatOpen((prev) => !prev)}
              onChatAdapterChange={setStreamChatAdapter}
              onStreamStatusChange={setStreamStatus}
              onCloseStream={() => navigate("/")}
            />

            {streamInfoBar}

            {cinemaMode && isChatOpen && (
              <ChatPanel
                streamKey={primaryStreamKey}
                variant="below"
                isOpen={isChatOpen}
                adapter={chatAdapters[primaryStreamKey]}
                displayName={effectiveChatDisplayName}
                onChangeDisplayNameRequested={() => setIsDisplayNameModalOpen(true)}
                onClose={() => setIsChatOpen(false)}
              />
            )}
          </div>

          {!cinemaMode && isChatOpen && (
            <aside className="player-aside">
              <ChatPanel
                streamKey={primaryStreamKey}
                variant="docked"
                isOpen={isChatOpen}
                adapter={chatAdapters[primaryStreamKey]}
                displayName={effectiveChatDisplayName}
                onChangeDisplayNameRequested={() => setIsDisplayNameModalOpen(true)}
                onClose={() => setIsChatOpen(false)}
              />
            </aside>
          )}
        </div>
      ) : (
        <>
          <div className="grid w-full grid-cols-1 md:grid-cols-2 gap-4">
            {streamKeys.map((streamKey) => (
              <div key={`${streamKey}_player_card`} className="min-w-0 flex flex-col gap-3">
                <StreamMOTD
                  isOnline={streamStatuses[streamKey]?.isOnline ?? false}
                  motd={streamStatuses[streamKey]?.motd ?? ""}
                  className='px-1'
                />

                <div className="flex flex-col gap-2">
                  <Player
                    key={`${streamKey}_player`}
                    streamKey={streamKey}
                    cinemaMode={cinemaMode}
                    isOnline={streamStatuses[streamKey]?.isOnline ?? false}
                    isChatOpen={isChatOpen}
                    onToggleChat={() => setIsChatOpen((prev) => !prev)}
                    onChatAdapterChange={setStreamChatAdapter}
                    onStreamStatusChange={setStreamStatus}
                    onCloseStream={() => removeStream(streamKey)}
                  />
                  <ChatPanel
                    streamKey={streamKey}
                    variant="below"
                    isOpen={isChatOpen}
                    adapter={chatAdapters[streamKey]}
                    displayName={effectiveChatDisplayName}
                    onChangeDisplayNameRequested={() => setIsDisplayNameModalOpen(true)}
                    onClose={() => setIsChatOpen(false)}
                  />
                </div>
              </div>
            ))}
          </div>

          <div className="broadcast-actions mt-4">
            <Button
              title={cinemaMode ? locale.player_page.cinema_mode_disable : locale.player_page.cinema_mode_enable}
              onClick={toggleCinemaMode}
              iconRight="CodeBracketSquare"
            />
            <Button
              title={locale.player_page.modal_add_stream_title}
              color="Accept"
              onClick={() => setIsModelOpen((prev) => !prev)}
              iconRight="SquaresPlus"
            />
          </div>
        </>
      )}

      <section className="mt-10">
        <div className="section-head">
          <div className="section-head__left">
            <span className="section-kicker">02 // keep exploring</span>
            <h2>more channels to visit</h2>
          </div>
        </div>
        <AvailableStreams showHeader={false} />
      </section>
    </div>
  );
};

export default PlayerPage;
