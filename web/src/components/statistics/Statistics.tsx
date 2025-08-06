import React, {useContext, useEffect} from "react";
import {StatusContext} from "../../providers/StatusProvider";
import {useNavigate} from "react-router-dom";

const Statistics = () => {
  const {streamStatus, refreshStatus} = useContext(StatusContext);
  const navigate = useNavigate();

  useEffect(() => {
    refreshStatus();
  }, []);
  
  return (
    <div className="p-6 min-h-screen">
      <h2 className="text-4xl font-semibold mb-4">📊 Statistics</h2>

      {!streamStatus || streamStatus?.length === 0 && (
        <p className="text-center text-gray-500 mt-10 text-3xl">No statistics currently available</p>
      )}

      <div className="space-y-6">
        {streamStatus?.map((status, i) => (
          <div key={i} className="border border-gray-300 rounded-lg p-4 shadow-sm " >
            <div className="text-lg font-medium text-indigo-400 m-0 flex flex-row justify-between content-center">
              <div
                className="px-4 py-2 rounded-lg ">
                Stream Key: {status.streamKey}
              </div>
              <button
                onClick={() => navigate(`/${status.streamKey}`)}
                className="bg-blue-500 hover:bg-blue-700 px-4 py-2 rounded-lg text-white">
                Watch stream
              </button>
            </div>

            {/* VideoStreams */}
            <div className="mb-4 mt-4">
              <h3 className="text-md font-semibold mb-2">🎥 Video Streams</h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
                {status.videoStreams.map((stream, index) => (
                  <div
                    key={index}
                    className="rounded-md p-3 border border-indigo-100"
                  >
                    <div><strong>RID:</strong> {stream.rid}</div>
                    <div><strong>Packets Received:</strong> {stream.packetsReceived}</div>
                    <div><strong>Last Key Frame:</strong> {stream.lastKeyFrameSeen}</div>
                  </div>
                ))}
              </div>
            </div>

            {/* WhepStreams */}
            <div>
              <h3 className="text-md font-semibold mb-2">🧬 WHEP Sessions</h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
                {status.whepSessions.map((session, index) => (
                  <div
                    key={index}
                    className="rounded-md p-3 border border-indigo-100"
                  >
                    <div><strong>ID:</strong> {session.id}</div>
                    <div><strong>Layer:</strong> {session.currentLayer}</div>
                    <div><strong>Timestamp:</strong> {session.timestamp}</div>
                    <div><strong>Packets Written:</strong> {session.packetsWritten}</div>
                    <div><strong>Seq Num:</strong> {session.sequenceNumber}</div>
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
