import React, { useEffect, useState } from "react";

const ADMIN_TOKEN = "adminToken";

const StatusPage = () => {
  const [response, setResponse] = useState<StatusResult[]>()

  const refreshStatus = () => {
    fetch(`/api/admin/status`, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${localStorage.getItem(ADMIN_TOKEN)}`,
      },
    })
      .then((result) => {
        if (result.status >= 400 && result.status < 500) {
          localStorage.removeItem(ADMIN_TOKEN)
          return;
        }

        return result.json();
      })
      .then((result) => {
        setResponse(() => result)
      });
  };

  useEffect(() => {
    refreshStatus()
  }, [])

  return (
    <div className="p-6 w-full max-w-6xl mx-auto">
      <h1 className="text-3xl font-bold text-gray-800 mb-6">Stream Status Overview</h1>

      <div className="overflow-x-auto">
        <table className="min-w-full bg-amber-800 border border-gray-200 rounded-lg shadow">
          <thead className="bg-blue-900 text-white">
            <tr>
              <th className="px-4 py-2 text-left">Stream Key</th>
              <th className="px-4 py-2 text-left">Is Public</th>
              <th className="px-4 py-2 text-left">Video Tracks</th>
              <th className="px-4 py-2 text-left">Audio Tracks</th>
              <th className="px-4 py-2 text-left">Sessions</th>
              <th className="px-4 py-2 text-left">Total Packets</th>
            </tr>
          </thead>
          <tbody>
            {response?.map((status, index) => {
              const totalVideoPackets = status.videoTracks.reduce(
                (sum, track) => sum + track.packetsReceived,
                0
              );
              const totalAudioPackets = status.audioTracks.reduce(
                (sum, track) => sum + track.packetsReceived,
                0
              );
              const totalPackets = totalVideoPackets + totalAudioPackets;

              return (
                <tr key={index} className="border-t border-gray-200 hover:bg-gray-50">
                  <td className="px-4 py-2 font-medium text-blue-700">{status.streamKey}</td>
                  <td className="px-4 py-2 font-medium ">{status.isPublic ? "Yes" : "No"}</td>
                  <td className="px-4 py-2">{status.videoTracks.length}</td>
                  <td className="px-4 py-2">{status.audioTracks.length}</td>
                  <td className="px-4 py-2">{status.sessions.length}</td>
                  <td className="px-4 py-2">{totalPackets}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
export default StatusPage;

interface WhepSession {
  id: string;

  audioLayerCurrent: string;
  audioTimestamp: string;
  audioPacketsWritten: number;
  audioSequenceNumber: number;

  videoLayerCurrent: string;
  videoTimestamp: string;
  videoPacketsWritten: number;
  videoSequenceNumber: number;

  sequenceNumber: number;
  timestamp: number;
}

interface StatusResult {
  streamKey: string;
  isPublic: boolean;

  videoTracks: VideoTrack[];
  audioTracks: AudioTrack[];

  sessions: WhepSession[];
}

interface VideoTrack {
  rid: string;
  packetsReceived: number;
  lastKeyframe: string;
}

interface AudioTrack {
  rid: string;
  packetsReceived: number;
}
