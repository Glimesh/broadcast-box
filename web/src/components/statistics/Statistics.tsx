import { useContext, useEffect, useState } from "react";
import { StatusContext, StatusResult, WhepSession } from "../../providers/StatusProvider";
import { useNavigate } from "react-router-dom";
import Button from "../shared/Button";
import { LocaleContext } from "../../providers/LocaleProvider";

const Statistics = () => {
  const { activeStreamsStatus: streamStatus, refreshStatus, subscribe, unsubscribe } = useContext(StatusContext);
  const navigate = useNavigate();
  const { locale } = useContext(LocaleContext);
  const [nowMs, setNowMs] = useState<number>(() => Date.now());

  const sortByStreamKey = (a: StatusResult, b: StatusResult) => a.streamKey.localeCompare(b.streamKey);
  const sortByUuid = (a: WhepSession, b: WhepSession) => a.id.localeCompare(b.id);
  const bytesToMbps = (bytesPerSecond: number) => ((bytesPerSecond * 8) / 1_000_000).toFixed(2);
  const runtime = (startTime: Date): number => nowMs - startTime.getTime();
  const formatRuntime = (ms: number): string => {
    const hours = Math.floor(ms / (1000 * 60 * 60));
    const minutes = Math.floor((ms / (1000 * 60)) % 60);
    const seconds = Math.floor((ms / 1000) % 60);

    return `${hours}h ${minutes}m ${seconds}s`;
  };

  useEffect(() => {
    subscribe();
    refreshStatus();

    return () => unsubscribe();
  }, [refreshStatus, subscribe, unsubscribe]);

  useEffect(() => {
    const interval = setInterval(() => {
      setNowMs(Date.now());
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  return (
    <div className="page-container animate-slide-up">
      <div className="section-head mt-0">
        <div className="section-head__left">
          <span className="section-kicker">node telemetry</span>
          <h2>{locale.statistics.title}</h2>
        </div>
        <div className="section-head__right">auto refresh ♡</div>
      </div>

      {!streamStatus || streamStatus?.length === 0 && (
        <div className="empty-state">
          <h3>nothing live right now</h3>
          <p>{locale.statistics.no_statistics_available}</p>
        </div>
      )}

      <div className="space-y-6">
        {streamStatus?.sort(sortByStreamKey)
          .map((status, i) => (
            <div key={i} className="window">
              <div className="window-bar">
                <span className="window-dots" aria-hidden="true">
                  <span className="window-dot" />
                  <span className="window-dot" />
                  <span className="window-dot" />
                </span>
                <span>stream ♡ {status.streamKey}</span>
                <span className="window-bar__spacer" />
                <span className="window-bar__code">since {new Date(status.streamStart).toLocaleString()}</span>
              </div>

              <div className="window-body">
                <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="status-pill status-pill--live">STREAM KEY: {status.streamKey}</span>
                    <span className="status-pill">RUNTIME {formatRuntime(runtime(new Date(status.streamStart)))}</span>
                  </div>
                  <Button
                    title={locale.statistics.button_watch_stream}
                    onClick={() => navigate(`/${status.streamKey}`)}
                  />
                </div>

                <div className="mt-5 flex flex-col gap-4 lg:flex-row">
                  <div className="min-w-0 flex-1">
                    <h3 className="panel-heading">{locale.statistics.video_tracks}</h3>
                    <div className="flex flex-col gap-3">
                      {status.videoTracks.length === 0 && (
                        <p className="muted text-sm">{locale.statistics.video_track_not_available}</p>
                      )}
                      {status.videoTracks.map((stream, index) => (
                        <div key={index} className="pixel-panel text-sm">
                          <div><strong className="muted">{locale.statistics.rid}:</strong> {stream.rid}</div>
                          <div><strong className="muted">{locale.statistics.video_bitrate}:</strong> {bytesToMbps(stream.bitrate)} Mbps</div>
                          <div><strong className="muted">{locale.statistics.video_bitrate_total_out}:</strong> {bytesToMbps(status.sessions.reduce((prev, curr) => prev + curr.videoBitrate, 0))} Mbps</div>
                          <div><strong className="muted">{locale.statistics.packets_received}:</strong> {stream.packetsReceived}</div>
                          <div><strong className="muted">{locale.statistics.packets_dropped}:</strong> {stream.packetsDropped}</div>
                          <div><strong className="muted">{locale.statistics.last_key_frame}:</strong> {new Date(stream.lastKeyframe).toISOString()}</div>
                        </div>
                      ))}
                    </div>
                  </div>

                  <div className="min-w-0 flex-1">
                    <h3 className="panel-heading">{locale.statistics.audio_tracks}</h3>
                    <div className="flex flex-col gap-3">
                      {status.audioTracks.length === 0 && (
                        <p className="muted text-sm">{locale.statistics.audio_track_not_available}</p>
                      )}
                      {status.audioTracks.map((stream, index) => (
                        <div key={index} className="pixel-panel text-sm">
                          <div><strong className="muted">{locale.statistics.rid}:</strong> {stream.rid}</div>
                          <div><strong className="muted">{locale.statistics.packets_received}:</strong> {stream.packetsReceived}</div>
                          <div><strong className="muted">{locale.statistics.packets_dropped}:</strong> {stream.packetsDropped}</div>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>

                <div className="mt-5">
                  <h3 className="panel-heading">{locale.statistics.whep_sessions}</h3>
                  <div className="flex flex-col gap-3">
                    {status.sessions.length === 0 && (
                      <p className="muted text-sm">{locale.statistics.no_sessions}</p>
                    )}
                    {status.sessions.sort(sortByUuid)
                      .map((session, index) => (
                        <div key={index} className="pixel-panel">
                          <div className="mb-2 font-mono text-sm muted">ID: {session.id}</div>

                          <div className="flex flex-col gap-4 sm:flex-row sm:gap-8">
                            <div className="text-xs">
                              <div className="panel-heading mb-1 text-[0.58rem]">{locale.statistics.audio}</div>
                              <div><strong className="muted">{locale.statistics.layer}:</strong> {session.audioLayerCurrent}</div>
                              <div><strong className="muted">{locale.statistics.packets_written}:</strong> {session.audioPacketsWritten}</div>
                              <div><strong className="muted">{locale.statistics.timestamp}:</strong> {session.audioTimestamp}</div>
                              <div><strong className="muted">{locale.statistics.sequence_number}:</strong> {session.audioSequenceNumber}</div>
                            </div>

                            <div className="text-xs">
                              <div className="panel-heading mb-1 text-[0.58rem]">{locale.statistics.video}</div>
                              <div><strong className="muted">{locale.statistics.layer}:</strong> {session.videoLayerCurrent}</div>
                              <div><strong className="muted">{locale.statistics.timestamp}:</strong> {session.videoTimestamp}</div>
                              <div><strong className="muted">{locale.statistics.video_bitrate}:</strong> {bytesToMbps(session.videoBitrate)} Mbps</div>
                              <div><strong className="muted">{locale.statistics.packets_written}:</strong> {session.videoPacketsWritten}</div>
                              <div><strong className="muted">{locale.statistics.sequence_number}:</strong> {session.videoSequenceNumber}</div>
                            </div>
                          </div>
                        </div>
                      ))}
                  </div>
                </div>
              </div>
            </div>
          ))}
      </div>
    </div>
  );
};

export default Statistics;
