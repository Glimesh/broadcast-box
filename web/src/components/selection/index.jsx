import React, {useEffect, useState} from 'react'
import { useNavigate } from 'react-router-dom'

function Selection() {
  const apiPath = import.meta.env.VITE_API_PATH;
  const [streamKey, setStreamKey] = useState('')
  const navigate = useNavigate()
  
  const [streams, setStreams] = useState([]);
  useEffect(() => {
    updateStreams();
    
    const interval = setInterval(() => {
      updateStreams()
    }, 5000);

    return () => clearInterval(interval); 
  }, []);
  
  const isActiveSession = (videoStreams) => {
    if(videoStreams === undefined || videoStreams.length === 0){
      return false;
    }
    
    return videoStreams.filter(stream => (new Date() - new Date(stream.lastKeyFrameSeen)) > 5_000).length > 0;
  }
  const updateStreams = () => {
    fetch(`${apiPath}/status`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    })
      .then(result => result.json())
      .then(result => {
        if(result){
          
          setStreams(result.map((e) => ({
            key: e.streamKey,
            isActive: isActiveSession(e.videoStreams)
          })
        ));
        }
      });
  }
  const onStreamKeyChange = e => {
    setStreamKey(e.target.value)
  }
  const onWatchStreamClick = (key) => {
    if(key !== ''){
      navigate(`/${key}`);
    }
    
    if (streamKey !== '') {
      navigate(`/${streamKey}`)
    }
  }
  const onPublishStreamClick = () => {
    if (streamKey !== '') {
      navigate(`/publish/${streamKey}`)
    }
  }
  
  return (
    <div className='space-y-4 mx-auto max-w-2xl pt-20 md:pt-24'>
      <form className='rounded-md bg-gray-800 shadow-md p-8'>
        <h2 className="font-light leading-tight text-4xl mt-0 mb-2">Welcome to Broadcast Box</h2>
        <p>Broadcast Box is a tool that allows you to efficiently stream high-quality video in real time, using the latest in video codecs and WebRTC technology.</p>

        <div className='my-4'>
          <label className='block text-sm font-bold mb-2' htmlFor='streamKey'>
            Stream Key
          </label>
          
          <input className='appearance-none border w-full py-2 px-3 leading-tight focus:outline-hidden focus:shadow-outline bg-gray-700 border-gray-700 text-white rounded-sm shadow-md placeholder-gray-200' id='streamKey' type='text' placeholder='Stream Key' onChange={onStreamKeyChange} autoFocus />
        </div>
        
        <div className='flex justify-center'>
          <button className='py-2 px-4 bg-blue-500 text-white font-semibold rounded-lg shadow-md hover:bg-blue-700 focus:outline-hidden focus:ring-2 focus:ring-blue-400 focus:ring-opacity-75' type='submit' onClick={onWatchStreamClick}>
            Watch Stream
          </button>

          <button className='ml-10 py-2 px-4 bg-blue-500 text-white font-semibold rounded-lg shadow-md hover:bg-blue-700 focus:outline-hidden focus:ring-2 focus:ring-blue-400 focus:ring-opacity-75' type='button' onClick={onPublishStreamClick}>
            Publish Stream
          </button>
        </div>
        
        <h2 className="font-light leading-tight text-4xl mb-2 mt-6">Current Streams</h2>
        <p className='flex justify-center mt-6'>{streams.length === 0 && "No streams currently available"}</p>
        <p>{streams.length !== 0 && "Click a stream to join it"}</p>
        <div className="m-6"/>
        
        <div className='flex flex-col'>
        {streams.map((e, i) => (
          <button
            key={i+'_'+e.key}
            className={`mt-2 py-2 px-4 ${e.isActive ? 'bg-blue-500' : 'bg-orange-500'} text-white font-semibold rounded-lg shadow-md hover:bg-blue-700 focus:outline-hidden focus:ring-2 focus:ring-blue-400 focus:ring-opacity-75`}
            onClick={() => onPublishStreamClick(e.key)}>
            {e.key}
          </button>))
        }
        </div>
      </form>
    </div>
  )
}

export default Selection