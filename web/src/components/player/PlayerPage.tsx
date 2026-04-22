import { useCallback, useContext, useEffect, useState } from "react";
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
  const [streamKeys, setStreamKeys] = useState<string[]>([window.location.pathname.substring(1)]);
  const [isModalOpen, setIsModelOpen] = useState<boolean>(false);
  const [isChatOpen, setIsChatOpen] = useState<boolean>(() => localStorage.getItem("chat-open") !== "false");
  const [chatAdapters, setChatAdapters] = useState<Record<string, ChatAdapter | undefined>>({});
  const [streamStatuses, setStreamStatuses] = useState<Record<string, StreamStatus | undefined>>({});
  const [isDisplayNameModalOpen, setIsDisplayNameModalOpen] = useState<boolean>(false);
  const [chatDisplayName, setChatDisplayName] = useState<string>(() => localStorage.getItem("chatDisplayName") ?? "");

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

  const isSingleStream = streamKeys.length === 1;

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
        <div className={isSingleStream ? "w-full" : "grid w-full grid-cols-2 gap-2"}>
          {streamKeys.map((streamKey, index) => {
            const isPrimarySingleStream = isSingleStream && index === 0;

            return (
              <div key={`${streamKey}_player_card`} className="min-w-0 flex flex-col gap-1">
                <div className={isPrimarySingleStream ? "relative flex flex-col gap-4 w-full" : "flex flex-col gap-1"}>
                  <div className={isPrimarySingleStream ? `min-w-0 transition-[margin] duration-200 ${isChatOpen ? (cinemaMode ? "lg:mr-80" : "lg:mr-84") : ""}` : "min-w-0"}>
                    <Player
                      key={`${streamKey}_player`}
                      streamKey={streamKey}
                      cinemaMode={cinemaMode}
                      isChatOpen={isChatOpen}
                      onToggleChat={() => setIsChatOpen((prev) => !prev)}
                      onChatAdapterChange={setStreamChatAdapter}
                      onStreamStatusChange={setStreamStatus}
                      onCloseStream={isPrimarySingleStream ? () => navigate("/") : () => removeStream(streamKey)}
                    />
                  </div>

                  <ChatPanel
                    streamKey={streamKey}
                    variant={isPrimarySingleStream ? "sidebar" : "below"}
                    isOpen={isChatOpen}
                    adapter={chatAdapters[streamKey]}
                    displayName={chatDisplayName}
                    onChangeDisplayNameRequested={() => setIsDisplayNameModalOpen(true)}
                  />
                </div>
                <StreamMOTD
                  isOnline={streamStatuses[streamKey]?.isOnline ?? false}
                  motd={streamStatuses[streamKey]?.motd ?? ""}
                  className={isPrimarySingleStream ? "px-1" : "px-4"}
                />

              </div>
            );
          })}
        </div>

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
