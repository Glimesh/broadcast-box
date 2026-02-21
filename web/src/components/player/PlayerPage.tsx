import { useCallback, useContext, useEffect, useRef, useState } from "react";
import Player from "./Player";
import { useNavigate } from "react-router-dom";
import { CinemaModeContext } from "../../providers/CinemaModeProvider";
import ModalTextInput from "../shared/ModalTextInput";
import Button from "../shared/Button";
import AvailableStreams from "../selection/AvailableStreams";
import { LocaleContext } from "../../providers/LocaleProvider";
import ChatPanel from "./components/ChatPanel";
import { ChatAdapter } from "../../hooks/useChatSession";
import { StreamMOTD } from "./components/StreamMOTD";
import { StreamStatus } from "../../providers/StatusProvider";

const PlayerPage = () => {
  const navigate = useNavigate();
  const { locale } = useContext(LocaleContext);
  const { cinemaMode, toggleCinemaMode } = useContext(CinemaModeContext);
  const [streamKeys, setStreamKeys] = useState<string[]>([ window.location.pathname.substring(1) ]);
  const [isModalOpen, setIsModelOpen] = useState<boolean>(false);
  const [isChatOpen, setIsChatOpen] = useState<boolean>(() => localStorage.getItem("chat-open") !== "false");
  const [chatAdapters, setChatAdapters] = useState<Record<string, ChatAdapter | undefined>>({});
  const [streamStatuses, setStreamStatuses] = useState<Record<string, StreamStatus | undefined>>({});
  const [singlePlayerHeightPx, setSinglePlayerHeightPx] = useState<number | undefined>(undefined);
  const [isDisplayNameModalOpen, setIsDisplayNameModalOpen] = useState<boolean>(false);
  const [chatDisplayName, setChatDisplayName] = useState<string>(() => localStorage.getItem("chatDisplayName") ?? "");
  const singlePlayerColumnRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    localStorage.setItem("chat-open", String(isChatOpen));
  }, [isChatOpen]);

  useEffect(() => {
    if (streamKeys.length !== 1) {
      setSinglePlayerHeightPx(undefined);
      return;
    }

    const playerColumnElement = singlePlayerColumnRef.current;
    if (!playerColumnElement) {
      return;
    }

    const updateHeight = () => {
      setSinglePlayerHeightPx(playerColumnElement.getBoundingClientRect().height);
    };

    updateHeight();
    const animationFrame = window.requestAnimationFrame(updateHeight);
    const resizeObserver = new ResizeObserver(updateHeight);
    resizeObserver.observe(playerColumnElement);

    return () => {
      window.cancelAnimationFrame(animationFrame);
      resizeObserver.disconnect();
    };
  }, [isChatOpen, streamKeys]);

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
  }, []);

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

  return (
    <div>
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

      <div className={`flex flex-col w-full items-center ${!cinemaMode && "mx-auto px-2 py-2 container gap-2"}`} >
        {streamKeys.length === 1 ? (
          <div className="flex w-full flex-col gap-2 2xl:flex-row 2xl:items-start">
            <div ref={singlePlayerColumnRef} className="min-w-0 flex-1 flex flex-col gap-1">
              <StreamMOTD
                isOnline={streamStatuses[streamKeys[0]]?.isOnline ?? false}
                motd={streamStatuses[streamKeys[0]]?.motd ?? ""}
                className="px-4"
              />
              <Player
                key={`${streamKeys[0]}_player`}
                streamKey={streamKeys[0]}
                cinemaMode={cinemaMode}
                isChatOpen={isChatOpen}
                onToggleChat={() => setIsChatOpen((prev) => !prev)}
                onChatAdapterChange={setStreamChatAdapter}
                onStreamStatusChange={setStreamStatus}
                onCloseStream={() => navigate("/")}
              />
            </div>

            <ChatPanel
              streamKey={streamKeys[0]}
              variant="sidebar"
              isOpen={isChatOpen}
              adapter={chatAdapters[streamKeys[0]]}
              fixedHeightPx={singlePlayerHeightPx}
              displayName={chatDisplayName}
              onChangeDisplayNameRequested={() => setIsDisplayNameModalOpen(true)}
            />
          </div>
        ) : (
          <div className="grid w-full grid-cols-2 gap-2">
            {streamKeys.map((streamKey) => (
              <div key={`${streamKey}_player_card`} className="min-w-0 flex flex-col gap-1">
                <StreamMOTD
                  isOnline={streamStatuses[streamKey]?.isOnline ?? false}
                  motd={streamStatuses[streamKey]?.motd ?? ""}
                  className="px-4"
                />
                <Player
                  key={`${streamKey}_player`}
                  streamKey={streamKey}
                  cinemaMode={cinemaMode}
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
                  displayName={chatDisplayName}
                  onChangeDisplayNameRequested={() => setIsDisplayNameModalOpen(true)}
                />
              </div>
            ))}
          </div>
        )}

        {/*Footer menu*/}
        <div className="flex flex-row gap-2">
          <Button
            title={cinemaMode ? locale.player_page.cinema_mode_disable : locale.player_page.cinema_mode_enable}
            onClick={toggleCinemaMode}
            iconRight="CodeBracketSquare"
          />

          {/*Show modal to add stream keys with*/}
          <Button
            title={locale.player_page.modal_add_stream_title}
            color="Accept"
            onClick={() => setIsModelOpen((prev) => !prev)}
            iconRight="SquaresPlus"
          />
        </div>
      </div>
    </div>
  );
};

export default PlayerPage;
