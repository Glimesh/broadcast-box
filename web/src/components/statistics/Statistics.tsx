import React, { useContext, useEffect } from "react";
import { StatusContext, StatusResult, WhepSession } from "../../providers/StatusProvider";
import { useNavigate } from "react-router-dom";
import Button from "../shared/Button";
import { LocaleContext } from "../../providers/LocaleProvider";

const Statistics = () => {
  const { activeStreamsStatus: streamStatus, refreshStatus, subscribe, unsubscribe } = useContext(StatusContext);
  const navigate = useNavigate();
  const { locale } = useContext(LocaleContext)

  const sortByStreamKey = (a: StatusResult, b: StatusResult) => a.streamKey.localeCompare(b.streamKey)
  const sortByUuid = (a: WhepSession, b: WhepSession) => a.id.localeCompare(b.id)
  const isStreamActive = (stream: StatusResult) => stream.videoTracks.length !== 0 || stream.audioTracks.length !== 0

  useEffect(() => {
    subscribe()
    refreshStatus();

    return () => unsubscribe();
	// eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div className="p-6 min-h-screen">
      <h2 className="text-4xl font-semibold mb-4">📊 {locale.statistics.title}</h2>

      {!streamStatus || streamStatus?.length === 0 && (
        <p className="text-center text-gray-500 mt-10 text-3xl">{locale.statistics.no_statistics_available}</p>
      )}

      <div className="space-y-6">
        {streamStatus?.sort(sortByStreamKey)
          .map((status, i) => (
            <div key={i} className="border border-gray-300 rounded-lg p-4 shadow-sm">
              <div className="text-lg font-medium text-indigo-400 m-0 flex flex-row justify-between content-center">
                <div className="px-4 py-2 rounded-lg text-2xl">
                  Stream Key: {status.streamKey}
                </div>
                <Button
                  title={locale.statistics.button_watch_stream}
                  onClick={() => navigate(`/${status.streamKey}`)}
                  isDisabled={isStreamActive(status)}
                />
              </div>

              <div className="flex justify-start">
                {/* VideoTracks */}
                <div className="mb-4 mt-4 mr-4">
                  <h3 className="text-2xl font-semibold mb-4">🎥 {locale.statistics.video_tracks}</h3>
                  <div className="gap-4">
                    {status.videoTracks.length === 0 && (
                      <p className="text-center text-gray-500 text-3xl">{locale.statistics.video_track_not_available}</p>
                    )}
                    {status.videoTracks.map((stream, index) => (
                      <div key={index} className="rounded-md p-3 min-h-25 border border-indigo-100" >
                        <div><strong>{locale.statistics.rid}:</strong> {stream.rid}</div>
                        <div><strong>{locale.statistics.packets_received}:</strong> {stream.packetsReceived}</div>
                        <div><strong>{locale.statistics.packets_dropped}:</strong> {stream.packetsDropped}</div>
                        <div><strong>{locale.statistics.last_key_frame}:</strong> {new Date(stream.lastKeyframe).toISOString()}</div>
                      </div>
                    ))}
                  </div>
                </div>

                {/* AudioTracks */}
                <div className="mb-4 mt-4">
                  <h3 className="text-2xl font-semibold mb-4">🎥 {locale.statistics.audio_tracks}</h3>
                  <div className="gap-4">
                    {status.audioTracks.length === 0 && (
                      <p className="text-center text-gray-500 text-3xl">{locale.statistics.audio_track_not_available}</p>
                    )}
                    {status.audioTracks.map((stream, index) => (
                      <div key={index} className="rounded-md p-3 min-h-25 border border-indigo-100">
                        <div><strong>{locale.statistics.rid}:</strong> {stream.rid}</div>
                        <div><strong>{locale.statistics.packets_received}:</strong> {stream.packetsReceived}</div>
                        <div><strong>{locale.statistics.packets_dropped}:</strong> {stream.packetsDropped}</div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              {/* WhepSessions */}
              <div>
                <h3 className="text-2xl font-semibold mb-4">🧬 {locale.statistics.whep_sessions}</h3>
                <div className="grid grid-cols-1 gap-4">
                  {status.sessions.length === 0 && (
                    <p className="text-center text-gray-500 text-3xl">{locale.statistics.no_sessions}</p>
                  )}
                  {status.sessions.sort(sortByUuid)
                    .map((session, index) => (
                      <div key={index} className="rounded-md p-3 border border-indigo-100">
                        <div><strong>ID:</strong> {session.id}</div>
                        <div className="mb-2" />

                        <div className="flex flex-row">
                          <div>
                            <div className="text-xl"><strong>{locale.statistics.audio}</strong> </div>
                            <div><strong>{locale.statistics.layer}:</strong> {session.audioLayerCurrent}</div>
                            <div><strong>{locale.statistics.packets_written}:</strong> {session.audioPacketsWritten}</div>
                            <div><strong>{locale.statistics.timestamp}:</strong> {session.audioTimestamp}</div>
                            <div><strong>{locale.statistics.sequence_number}:</strong> {session.audioSequenceNumber}</div>
                          </div>

                          <div className="mr-8" />

                          <div>
                            <div className="text-xl"><strong>{locale.statistics.video}</strong> </div>
                            <div><strong>{locale.statistics.layer}:</strong> {session.videoLayerCurrent}</div>
                            <div><strong>{locale.statistics.timestamp}:</strong> {session.videoTimestamp}</div>
                            <div><strong>{locale.statistics.packets_written}:</strong> {session.videoPacketsWritten}</div>
                            <div><strong>{locale.statistics.sequence_number}:</strong> {session.videoSequenceNumber}</div>
                          </div>
                        </div>
                      </div>
                    ))}
                </div>
              </div>
            </div>
          ))}
      </div>
    </div>
  );
};

export default Statistics;
