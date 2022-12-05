import React from 'react'
import { useNavigate } from 'react-router-dom'

function Selection (props) {
  const [streamKey, setStreamKey] = React.useState('')
  const navigate = useNavigate()

  const onStreamKeyChange = e => {
    setStreamKey(e.target.value)
  }
  const onWatchStreamClick = () => {
    navigate(streamKey)
  }

  return (
    <div className='w-full max-w-xs'>
      <form className='bg-white shadow-md rounded px-8 pt-6 pb-8 mb-4'>
        <div className='mb-4'>
          <label className='block text-gray-700 text-sm font-bold mb-2' htmlFor='streamKey'>
            Stream Key
          </label>
          <input className='shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline' id='streamKey' type='text' placeholder='Stream Key' onChange={onStreamKeyChange} />
        </div>
        <div className='flex items-center justify-between'>
          <button className='bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline' type='button' onClick={onWatchStreamClick}>
            Watch Stream
          </button>
        </div>
      </form>
    </div>
  )
}

export default Selection
