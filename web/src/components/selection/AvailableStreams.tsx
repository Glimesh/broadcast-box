import React, { useContext, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { StatusContext, StatusResult } from "../../providers/StatusProvider";
import Button from "../shared/Button";

interface StreamEntry {
  streamKey: string;
  motd: string;
}

const AvailableStreams = () => {
  const navigate = useNavigate();

  const { activeStreamsStatus: streamStatus, refreshStatus, subscribe, unsubscribe } = useContext(StatusContext)
  const [streams, setStreams] = useState<StreamEntry[] | undefined>(undefined);

  const sortByStreamKey = (a: StatusResult, b: StatusResult) => a.streamKey.localeCompare(b.streamKey)

  useEffect(() => {
    subscribe()
    refreshStatus()

    return () => unsubscribe()
  }, []);

  useEffect(() => {
    setStreams(() =>
      streamStatus?.filter((resultEntry) => resultEntry.videoTracks.length > 0)
        .sort(sortByStreamKey)
        .map((resultEntry: StatusResult) => ({
          streamKey: resultEntry.streamKey,
          videoStreams: resultEntry.videoTracks,
          motd: resultEntry.motd
        })));
  }, [streamStatus])

  const onWatchStreamClick = (key: string) => {
    if (key !== '') {
      navigate(`/${key}`);
    }
  }

  if (streams === undefined) {
    return <></>;
  }

  return (
    <div className="flex flex-col">
      <h2 className="font-light leading-tight text-4xl mb-2 mt-6">Current Streams</h2>
      {streams.length === 0 && <p className='flex justify-center mt-6'>No streams currently available</p>}
      {streams.length !== 0 && <p>Click a stream to join it</p>}

      <div className="m-2" />

      <div className='flex flex-col gap-2'>
        {streams.map((e, i) => (
          <Button
            title={e.streamKey}
            subTitle={e.motd}
            stretch
            center
            key={i + '_' + e.streamKey}
            onClick={() => onWatchStreamClick(e.streamKey)}
          />
        ))
        }
      </div>
    </div>
  )
}

export default AvailableStreams
