import { useContext } from "react";
import { LocaleContext } from "../../../providers/LocaleProvider";

interface StreamMOTDProps {
  isOnline: boolean;
  motd: string;
  className?: string;
  onEdit?: () => void;
}

export const StreamMOTD = (props: StreamMOTDProps) => {
  const { isOnline, motd, className = '', onEdit } = props;
  const { locale } = useContext(LocaleContext);

  return (
    <div className={`motd-strip ${isOnline ? "" : "motd-strip--offline"} ${className}`}>
      <span className="motd-strip__text">
        {isOnline ? (motd || "no title yet ♡") : locale.player.stream_status_offline}
      </span>

      {onEdit && (
        <button
          type="button"
          className="motd-strip__edit"
          title={locale.player_page.edit_stream_title}
          aria-label={locale.player_page.edit_stream_title}
          onClick={onEdit}
        >
          edit ♡
        </button>
      )}
    </div>
  );
};
