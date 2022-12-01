import React from 'react'

function AdminPanel (props) {
  const [streamKey, setStreamKey] = React.useState('')

  const onStreamKeyChange = e => {
    setStreamKey(e.target.value)
  }
  const onSaveConfigurationClick = () => {
    fetch('http://localhost:8080/api/configure', {
      method: 'POST',
      body: JSON.stringify({
        streamKey
      })
    }).then(() => {
      props.onConfigurationSuccess()
    })
  }

  return (
    <div class='w-full max-w-xs'>
      <form class='bg-white shadow-md rounded px-8 pt-6 pb-8 mb-4'>
        <div class='mb-4'>
          <label class='block text-gray-700 text-sm font-bold mb-2' for='password'>
            Admin Password
          </label>
          <input class='shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline' id='password' type='text' placeholder='Admin Password' />
        </div>
        <div class='mb-4'>
          <label class='block text-gray-700 text-sm font-bold mb-2' for='streamKey'>
            Stream Key
          </label>
          <input class='shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline' id='streamKey' type='text' placeholder='Stream Key' onChange={onStreamKeyChange} />
        </div>
        <div class='flex items-center justify-between'>
          <button class='bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline' type='button' onClick={onSaveConfigurationClick}>
            Save
          </button>
        </div>
      </form>
    </div>
  )
}

export default AdminPanel
