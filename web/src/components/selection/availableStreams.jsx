import React, {useEffect, useState} from "react";
import {useNavigate} from "react-router-dom";

function AvailableStreams() {
  const apiPath = import.meta.env.VITE_API_PATH;
  const navigate = useNavigate();

  const [streams, setStreams] = useState(undefined);
  useEffect(() => {
    updateStreams();

    const interval = setInterval(() => {
      updateStreams()
    }, 5000);

    return () => clearInterval(interval);
  }, []);

  const isActiveSession = (videoStreams) => {
    if (videoStreams === undefined || videoStreams.length === 0) {
      return false;
    }

    return videoStreams.filter(stream => (new Date() - new Date(stream.lastKeyFrameSeen)) > 5_000).length === 0;
  }
  const updateStreams = () => {
    fetch(`${apiPath}/status`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    })
      .then(result => {
        if (result.status === 503) {
          throw new Error('Status API disabled');
        }

        return result.json()
      })
      .then(result => {
        if (result) {
          setStreams(result.map((e) => ({
              key: e.streamKey,
              isActive: isActiveSession(e.videoStreams)
            })
          ));
        }
      })
      .catch(() => {
        setStreams(undefined);
      });
  }
  const onWatchStreamClick = (key) => {
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
            key={i + '_' + e.key}
            className={`mt-2 py-2 px-4 ${e.isActive ? 'bg-blue-500' : 'bg-orange-500'} text-white font-semibold rounded-lg shadow-md hover:bg-blue-700 focus:outline-hidden focus:ring-2 focus:ring-blue-400 focus:ring-opacity-75`}
            onClick={() => onWatchStreamClick(e.key)}>
            {e.key}
          </button>
        ))
        }
      </div>
    </div>
  )
}

export default AvailableStreams