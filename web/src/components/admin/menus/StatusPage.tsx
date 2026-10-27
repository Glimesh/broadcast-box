import { useCallback, useContext, useEffect, useState } from "react";
import { LocaleContext } from "../../../providers/LocaleProvider";
import {
  clearAdminToken,
  getAdminAuthorizationHeader,
  isInvalidAdminSessionResponse,
} from "../adminAuth";

const StatusPage = () => {
  const { locale } = useContext(LocaleContext)
  const [response, setResponse] = useState<StatusResult[]>()
  const statusRefreshIntervalMs = 5000

  const refreshStatus = useCallback(() => {
    fetch(`/api/admin/status`, {
      method: "GET",
      headers: {
        Authorization: getAdminAuthorizationHeader(),
      },
    })
      .then((result) => {
        if (isInvalidAdminSessionResponse(result.status)) {
          clearAdminToken()
          return;
        }

        return result.json();
      })
      .then((result) => {
        setResponse(() => result)
      });
  }, []);

  useEffect(() => {
    refreshStatus()
    const interval = setInterval(refreshStatus, statusRefreshIntervalMs);

    return () => clearInterval(interval);
  }, [refreshStatus, statusRefreshIntervalMs])

  return (
    <div className="flex w-full flex-col gap-4">
      <h1 className="panel-heading text-base">{locale.admin_page_status_page.title}</h1>

      <div className="admin-table-wrap">
        <table className="data-table">
          <thead>
            <tr>
              <th>{locale.admin_page_status_page.table_header_stream_key}</th>
              <th>{locale.admin_page_status_page.table_header_is_public}</th>
              <th>{locale.admin_page_status_page.table_header_video_tracks}</th>
              <th>{locale.admin_page_status_page.table_header_audio_tracks}</th>
              <th>{locale.admin_page_status_page.table_header_sessions}</th>
              <th>{locale.admin_page_status_page.table_header_total_packets}</th>
            </tr>
          </thead>
          <tbody>
            {response?.sort().map((status, index) => {
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
                  <td>{status.streamKey}</td>
                  <td>{status.isPublic ? locale.admin_page_status_page.yes : locale.admin_page_status_page.no}</td>
                  <td>{status.videoTracks.length}</td>
                  <td>{status.audioTracks.length}</td>
                  <td>{status.sessions.length}</td>
                  <td>{totalPackets}</td>
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
