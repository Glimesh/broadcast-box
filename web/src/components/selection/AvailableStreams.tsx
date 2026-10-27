import { useContext, useEffect, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { StatusContext, StatusResult } from "../../providers/StatusProvider";
import { LocaleContext } from '../../providers/LocaleProvider';

interface StreamEntry {
  streamKey: string;
  displayName: string;
  avatarUrl: string;
  motd: string;
  viewers: number;
}

interface AvailableStreamsProps {
  showHeader?: boolean;
  onClickOverride?: (streamKey: string) => void;
}

const AvailableStreams = (props: AvailableStreamsProps) => {
  const navigate = useNavigate();
  const { locale } = useContext(LocaleContext);

  const { activeStreamsStatus: streamStatus, refreshStatus, subscribe, unsubscribe } = useContext(StatusContext);

  useEffect(() => {
    subscribe();
    refreshStatus();

    return () => unsubscribe();
  }, [refreshStatus, subscribe, unsubscribe]);

  const streams = useMemo<StreamEntry[] | undefined>(() =>
    streamStatus?.filter((resultEntry) => resultEntry.videoTracks.length > 0 && resultEntry.isPublic)
      .sort((a: StatusResult, b: StatusResult) => a.streamKey.localeCompare(b.streamKey))
      .map((resultEntry: StatusResult) => ({
        streamKey: resultEntry.streamKey,
        displayName: resultEntry.displayName || resultEntry.streamKey,
        avatarUrl: resultEntry.avatarUrl || "",
        motd: resultEntry.motd,
        viewers: resultEntry.sessions?.length ?? 0,
      })), [streamStatus]);

  const onWatchStreamClick = (key: string) => {
    if (key !== '') {
      navigate(`/${key}`);
    }
  };

  if (streams === undefined) {
    return (
      <div className="stream-grid">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="stream-skeleton" />
        ))}
      </div>
    );
  }

  if (streams.length === 0) {
    return (
      <div className="empty-state">
        <div className="empty-state__icon">
          <svg className="h-9 w-9" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
        </div>
        <h3>nobody&apos;s live right now</h3>
        <p>{locale.available_streams.no_streams_message}</p>
        <span className="construction">check back soon ♡</span>
      </div>
    );
  }

  return (
    <div className="stream-grid">
      {streams.map((e, i) => (
        <button
          type="button"
          key={i + '_' + e.streamKey}
          onClick={() => {
            if (props.onClickOverride !== undefined) {
              props.onClickOverride(e.streamKey);
            } else {
              onWatchStreamClick(e.streamKey);
            }
          }}
          className="stream-card"
        >
          <div className="stream-thumb">
            {e.avatarUrl ? (
              <img
                src={e.avatarUrl}
                alt={e.displayName}
                className="stream-avatar"
              />
            ) : (
              <span className="stream-avatar stream-avatar--fallback">
                {e.displayName.charAt(0).toUpperCase()}
              </span>
            )}
            <span className="stream-live">live</span>
            <span className="stream-viewers">♡ {e.viewers}</span>
          </div>

          <div className="stream-body">
            <div className="stream-name">{e.displayName}</div>
            {e.motd && <div className="stream-motd">{e.motd}</div>}
            <div className="stream-cta">watch stream ♡</div>
          </div>
        </button>
      ))}
    </div>
  );
};

export default AvailableStreams;
