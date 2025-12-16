import React, { useContext, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { StatusContext, StatusResult } from "../../providers/StatusProvider";
import Button from "../shared/Button";
import { LocaleContext } from '../../providers/LocaleProvider';

interface StreamEntry {
  streamKey: string;
  motd: string;
}

interface AvailableStreamsProps {
  showHeader?: boolean;
	// eslint-disable-next-line no-unused-vars
  onClickOverride?: (streamKey: string) => void;
}

const AvailableStreams = (props: AvailableStreamsProps) => {
  const navigate = useNavigate();
  const { locale } = useContext(LocaleContext)

  const { activeStreamsStatus: streamStatus, refreshStatus, subscribe, unsubscribe } = useContext(StatusContext)
  const [streams, setStreams] = useState<StreamEntry[] | undefined>(undefined);

  const sortByStreamKey = (a: StatusResult, b: StatusResult) => a.streamKey.localeCompare(b.streamKey)

  useEffect(() => {
    subscribe()
    refreshStatus()

    return () => unsubscribe()
    // eslint-disable-next-line react-hooks/exhaustive-deps
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
      {props.showHeader !== false && (
        <div>
          <div className="font-light leading-tight text-4xl mb-2">{locale.available_streams.title}</div>
          {streams.length !== 0 && <p>{locale.available_streams.stream_join_message}</p>}

          <div className="m-2" />
        </div>
      )}
      {streams.length === 0 && <p className='flex justify-center mb-2 mt-2'>{locale.available_streams.no_streams_message}</p>}

      <div className='flex flex-col gap-2'>
        {streams.map((e, i) => (
          <Button
            title={e.streamKey}
            subTitle={e.motd}
            stretch
            center
            key={i + '_' + e.streamKey}
            onClick={() => {
              if (props.onClickOverride !== undefined) {
                props.onClickOverride(e.streamKey)
              } else {
                onWatchStreamClick(e.streamKey)
              }
            }}
          />
        ))
        }
      </div>
    </div>
  )
}

export default AvailableStreams
