import React, {useState} from 'react'
import {useNavigate} from 'react-router-dom'
import AvailableStreams from "./availableStreams.jsx";

function Selection() {
  const [streamKey, setStreamKey] = useState('')
  const navigate = useNavigate()

  const onStreamKeyChange = e => {
    setStreamKey(e.target.value)
  }
  const onWatchStreamClick = () => {
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
        <p>Broadcast Box is a tool that allows you to efficiently stream high-quality video in real time, using the
          latest in video codecs and WebRTC technology.</p>

        <div className='my-4'>
          <label className='block text-sm font-bold mb-2' htmlFor='streamKey'>
            Stream Key
          </label>

          <input
            className='appearance-none border w-full py-2 px-3 leading-tight focus:outline-hidden focus:shadow-outline bg-gray-700 border-gray-700 text-white rounded-sm shadow-md placeholder-gray-200'
            id='streamKey' type='text' placeholder='Stream Key' onChange={onStreamKeyChange} autoFocus/>
        </div>

        <div className='flex justify-center'>
          <button
            className='py-2 px-4 bg-blue-500 text-white font-semibold rounded-lg shadow-md hover:bg-blue-700 focus:outline-hidden focus:ring-2 focus:ring-blue-400 focus:ring-opacity-75'
            type='submit' onClick={onWatchStreamClick}>
            Watch Stream
          </button>

          <button
            className='ml-10 py-2 px-4 bg-blue-500 text-white font-semibold rounded-lg shadow-md hover:bg-blue-700 focus:outline-hidden focus:ring-2 focus:ring-blue-400 focus:ring-opacity-75'
            type='button' onClick={onPublishStreamClick}>
            Publish Stream
          </button>
          
        </div>
        
        <AvailableStreams/>
      </form>
    </div>
  )
}

export default Selection