import React, {useContext, useEffect, useState} from "react";
import {useNavigate} from "react-router-dom";
import {StatusContext} from "../../providers/StatusProvider";

interface StatusResult {
  streamKey: string;
  videoStreams: VideoStream[];
}

interface VideoStream {
  lastKeyFrameSeen: string;
}

interface StreamEntry {
  streamKey: string;
}

const AvailableStreams = () => {
  const navigate = useNavigate();

  const {streamStatus, refreshStatus} = useContext(StatusContext)
  const [streams, setStreams] = useState<StreamEntry[] | undefined>(undefined);

  useEffect(() => {
    refreshStatus()
  }, []);

  useEffect(() => {
    setStreams(() =>
      streamStatus?.filter((resultEntry) => resultEntry.videoStreams.length > 0)
        .map((resultEntry: StatusResult) => ({
          streamKey: resultEntry.streamKey,
          videoStreams: resultEntry.videoStreams
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

      <div className="m-2"/>

      <div className='flex flex-col'>
        {streams.map((e, i) => (
          <button
            key={i + '_' + e.streamKey}
            className={`mt-2 py-2 px-4 bg-blue-500 text-white font-semibold rounded-lg shadow-md hover:bg-blue-700 focus:outline-hidden focus:ring-2 focus:ring-blue-400 focus:ring-opacity-75`}
            onClick={() => onWatchStreamClick(e.streamKey)}>
            {e.streamKey}
          </button>
        ))
        }
      </div>
    </div>
  )
}

export default AvailableStreams