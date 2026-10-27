import { useContext } from "react";
import { VideoCameraIcon, VideoCameraSlashIcon } from "@heroicons/react/16/solid";
import { LocaleContext } from "../../../providers/LocaleProvider";

interface StatusMessageComponentProps {
  streamKey: string;
  state: "Loading" | "Playing" | "Offline" | "Error";
}

export const StatusMessageComponent = (props: StatusMessageComponentProps) => {
  const { locale } = useContext(LocaleContext);
  const { streamKey, state } = props;

  if (state === "Playing") {
    return null;
  }

  const icon = state === "Error" || state === "Offline"
    ? <VideoCameraSlashIcon className="h-10 w-10" />
    : <VideoCameraIcon className="h-10 w-10 animate-pulse" />;

  const message = state === "Error"
    ? locale.player.message_error
    : state === "Offline"
      ? locale.player.message_is_not_online
      : locale.player.message_loading_video;

  const stateClass = state === "Error"
    ? "video-status--error"
    : state === "Offline"
      ? "video-status--offline"
      : "";

  return (
    <div className={`video-status ${stateClass}`}>
      <div className="video-status__inner">
        <div className="video-status__icon">{icon}</div>
        <div className="video-status__key">{streamKey}</div>
        <div className="video-status__message">{message}</div>
      </div>
    </div>
  );
};
