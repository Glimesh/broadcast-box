import React, { useContext, useEffect, useState } from "react";
import { LocaleContext } from "../../../providers/LocaleProvider";
import toBase64Utf8 from "../../../utilities/base64";

const ADMIN_TOKEN = "adminToken";

const StatusPage = () => {
  const { locale } = useContext(LocaleContext)
  const [response, setResponse] = useState<StatusResult[]>()

  const refreshStatus = () => {
    fetch(`/api/admin/status`, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${toBase64Utf8(localStorage.getItem(ADMIN_TOKEN) ?? "")}`,
      },
    })
      .then((result) => {
        if (result.status > 400 && result.status < 500) {
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
      <h1 className="text-3xl font-bold mb-6">{locale.admin_page_status_page.title}</h1>

      <div className="overflow-x-auto">
        <table className="min-w-full rounded-lg shadow">
          <thead className="text-white">
            <tr>
              <th className="px-4 py-2 text-left">{locale.admin_page_status_page.table_header_stream_key}</th>
              <th className="px-4 py-2 text-left">{locale.admin_page_status_page.table_header_is_public}</th>
              <th className="px-4 py-2 text-left">{locale.admin_page_status_page.table_header_video_tracks}</th>
              <th className="px-4 py-2 text-left">{locale.admin_page_status_page.table_header_audio_tracks}</th>
              <th className="px-4 py-2 text-left">{locale.admin_page_status_page.table_header_sessions}</th>
              <th className="px-4 py-2 text-left">{locale.admin_page_status_page.table_header_total_packets}</th>
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
                <tr key={index} className="border-t">
                  <td className="px-4 py-2 font-medium ">{status.streamKey}</td>
                  <td className="px-4 py-2 font-medium ">{status.isPublic ? locale.admin_page_status_page.yes : locale.admin_page_status_page.no}</td>
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
