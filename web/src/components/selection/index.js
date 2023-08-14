import React from 'react'
import { useNavigate } from 'react-router-dom'

function Selection(props) {
  const [streamKey, setStreamKey] = React.useState('')
  const navigate = useNavigate()

  const onStreamKeyChange = e => {
    setStreamKey(e.target.value)
  }
  const onWatchStreamClick = () => {
    navigate(`/${streamKey}`)
  }

  const onPublishStreamClick = () => {
    navigate(`/publish/${streamKey}`)
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
          <input className='appearance-none border w-full py-2 px-3 leading-tight focus:outline-none focus:shadow-outline bg-gray-700 border-gray-700 text-white rounded shadow-md placeholder-gray-200' id='streamKey' type='text' placeholder='Stream Key' onChange={onStreamKeyChange} autoFocus />
        </div>
        <div className='flex'>
          <button className='py-2 px-4 bg-blue-500 text-white font-semibold rounded-lg shadow-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-400 focus:ring-opacity-75' type='button' onClick={onWatchStreamClick}>
            Watch Stream
          </button>

          <button className='ml-10 py-2 px-4 bg-blue-500 text-white font-semibold rounded-lg shadow-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-400 focus:ring-opacity-75' type='button' onClick={onPublishStreamClick}>
            Publish Stream
          </button>
        </div>
      </form>

      {/*
      <div className="rounded-md bg-gray-800 shadow-md p-8">
        <h2 className="font-light leading-tight text-2xl mb-2">Q: What is Broadcast Box?</h2>


        <p>A: Broadcast Box is a tool that lets you broadcast video to others in sub-second time. It is designed to be simple to use and easily modifiable, and is built using cutting-edge technology.</p>


        <h2 className="font-light leading-tight text-2xl mt-4 mb-2">Q: How does Broadcast Box work?</h2>

        <p>A: Broadcast Box uses WebRTC for broadcast and playback, which allows for fast and efficient video streaming. This is in contrast to RTMP and HLS, which can be slower and more complex to use.</p>


        <h2 className="font-light leading-tight text-2xl mt-4 mb-2">Q: Can I serve my video with Broadcast Box without a public IP or forwarding ports?</h2>

        <p>A: Yes, with Broadcast Box you can share your video without needing a public IP or forwarding ports. This makes it easy to share your video with others, even if you are running Broadcast Box on the same machine as your video source.</p>


        <h2 className="font-light leading-tight text-2xl mt-4 mb-2">Q: What are the benefits of using WebRTC with Broadcast Box?</h2>

        <p>A: There are several benefits to using WebRTC with Broadcast Box, including access to the latest in video codecs, the ability to upload multiple video streams in the same session, and the ability for your viewers to upload the same video at different quality levels. This can help to provide a high-quality viewing experience while keeping costs low for the server operator.</p>


        <h2 className="font-light leading-tight text-2xl mt-4 mb-2">Q: Can I use Broadcast Box to broadcast multiple camera angles or interactive video experiences?</h2>

        <p>A: Yes, with WebRTC you can upload multiple video streams in the same session, which means you can broadcast multiple camera angles or interactive video experiences in real time. This can help to provide a more immersive and engaging experience for your viewers.</p>
      </div>
  */}
    </div>
  )
}

export default Selection
