import React, { useContext, useEffect } from "react";
import { StatusContext, StatusResult, WhepSession } from "../../providers/StatusProvider";
import { useNavigate } from "react-router-dom";

const Statistics = () => {
  const { activeStreamsStatus: streamStatus, refreshStatus, subscribe, unsubscribe } = useContext(StatusContext);
  const navigate = useNavigate();

  const sortByStreamKey = (a: StatusResult, b: StatusResult) => a.streamKey.localeCompare(b.streamKey)
  const sortByUuid = (a: WhepSession, b: WhepSession) => a.id.localeCompare(b.id)
  const isStreamActive = (stream: StatusResult) => stream.videoTracks.length !== 0 || stream.audioTracks.length !== 0

  useEffect(() => {
    subscribe()
    refreshStatus();

    return () => unsubscribe();
  }, []);

  return (
    <div className="p-6 min-h-screen">
      <h2 className="text-4xl font-semibold mb-4">📊 Statistics</h2>

      {!streamStatus || streamStatus?.length === 0 && (
        <p className="text-center text-gray-500 mt-10 text-3xl">No statistics currently available</p>
      )}

      <div className="space-y-6">
        {streamStatus?.sort(sortByStreamKey)
          .map((status, i) => (
            <div key={i} className="border border-gray-300 rounded-lg p-4 shadow-sm">
              <div className="text-lg font-medium text-indigo-400 m-0 flex flex-row justify-between content-center">
                <div className="px-4 py-2 rounded-lg text-2xl">
                  Stream Key: {status.streamKey}
                </div>
                <button
                  disabled={!isStreamActive(status)}
                  onClick={() => navigate(`/${status.streamKey}`)}
                  className={`${isStreamActive(status) ? "bg-blue-500 hover:bg-blue-700" : "bg-gray-700"} px-4 py-2 rounded-lg text-white`}>
                  Watch stream
                </button>
              </div>

              <div className="flex justify-start">
                {/* VideoTracks */}
                <div className="mb-4 mt-4 mr-4">
                  <h3 className="text-2xl font-semibold mb-4">🎥 Video Tracks</h3>
                  <div className="gap-4">
                    {status.videoTracks.length === 0 && (
                      <p className="text-center text-gray-500 text-3xl">No video tracks</p>
                    )}
                    {status.videoTracks.map((stream, index) => (
                      <div key={index} className="rounded-md p-3 min-h-25 border border-indigo-100" >
                        <div><strong>RID:</strong> {stream.rid}</div>
                        <div><strong>Packets Received:</strong> {stream.packetsReceived}</div>
                        <div><strong>Last Key Frame:</strong> {stream.lastKeyframe}</div>
                      </div>
                    ))}
                  </div>
                </div>

                {/* AudioTracks */}
                <div className="mb-4 mt-4">
                  <h3 className="text-2xl font-semibold mb-4">🎥 Audio Tracks</h3>
                  <div className="gap-4">
                    {status.audioTracks.length === 0 && (
                      <p className="text-center text-gray-500 text-3xl">No audio tracks</p>
                    )}
                    {status.audioTracks.map((stream, index) => (
                      <div key={index} className="rounded-md p-3 min-h-25 border border-indigo-100">
                        <div><strong>RID:</strong> {stream.rid}</div>
                        <div><strong>Packets Received:</strong> {stream.packetsReceived}</div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              {/* WhepSessions */}
              <div>
                <h3 className="text-2xl font-semibold mb-4">🧬 WHEP Sessions</h3>
                <div className="grid grid-cols-1 gap-4">
                  {status.sessions.length === 0 && (
                    <p className="text-center text-gray-500 text-3xl">No current sessions</p>
                  )}
                  {status.sessions.sort(sortByUuid)
                    .map((session, index) => (
                      <div key={index} className="rounded-md p-3 border border-indigo-100">
                        <div><strong>ID:</strong> {session.id}</div>
                        <div className="mb-2" />

                        <div className="flex flex-row">
                          <div>
                            <div className="text-xl"><strong>Audio</strong> </div>
                            <div><strong>Layer:</strong> {session.audioLayerCurrent}</div>
                            <div><strong>Packets Written:</strong> {session.audioPacketsWritten}</div>
                            <div><strong>Timestamp:</strong> {session.audioTimestamp}</div>
                            <div><strong>Seq Num:</strong> {session.audioSequenceNumber}</div>
                          </div>

                          <div className="mr-8" />

                          <div>
                            <div className="text-xl"><strong>Video</strong> </div>
                            <div><strong>Layer:</strong> {session.videoLayerCurrent}</div>
                            <div><strong>Timestamp:</strong> {session.videoTimestamp}</div>
                            <div><strong>Packets Written:</strong> {session.videoPacketsWritten}</div>
                            <div><strong>Seq Num:</strong> {session.videoSequenceNumber}</div>
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
