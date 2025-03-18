import React from 'react'
import { useNavigate } from 'react-router-dom'

function Selection(props) {
  const [streamKey, setStreamKey] = React.useState('')
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
    <div className='space-y-8 mx-auto max-w-3xl pt-16 md:pt-20 px-2'>
      <div className='rounded-lg bg-white shadow-[var(--shadow-md)] p-8'>
        <div className="text-center mb-8">
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="var(--color-primary)" className="w-16 h-16 mx-auto mb-4">
            <path d="M4.5 4.5a3 3 0 0 0-3 3v9a3 3 0 0 0 3 3h8.25a3 3 0 0 0 3-3v-9a3 3 0 0 0-3-3H4.5ZM19.94 18.75l-2.69-2.69V7.94l2.69-2.69c.944-.945 2.56-.276 2.56 1.06v11.38c0 1.336-1.616 2.005-2.56 1.06Z" />
          </svg>
          <h1 className="text-4xl font-bold text-[var(--color-text-primary)] mb-3">Welcome to Broadcast Box</h1>
          <p className="text-[var(--color-text-secondary)] text-lg max-w-2xl mx-auto">
            Stream high-quality video in real time using the latest WebRTC technology
          </p>
        </div>

        <div className='bg-[var(--color-background)] p-6 rounded-lg border border-[var(--color-border)] mb-6'>
          <div className='mb-4'>
            <label className='block text-sm font-medium text-[var(--color-text-primary)] mb-2' htmlFor='streamKey'>
              Enter Stream Key
            </label>
            <input 
              className='w-full py-3 px-4 bg-white border border-[var(--color-border)] rounded-md shadow-sm
                focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)] focus:border-transparent
                transition-all duration-200 text-[var(--color-text-primary)]' 
              id='streamKey' 
              type='text' 
              placeholder='Your unique stream key...' 
              onChange={onStreamKeyChange} 
              autoFocus 
            />
          </div>
          <div className='flex flex-col sm:flex-row gap-4'>
            <button 
              className='flex-1 py-3 px-6 bg-[var(--color-primary)] hover:bg-[var(--color-primary-hover)]
                text-white font-medium rounded-md shadow-sm transition-colors duration-200
                focus:outline-none focus:ring-2 focus:ring-[var(--color-primary)] focus:ring-opacity-50
                flex items-center justify-center gap-2'
              type='button' 
              onClick={onWatchStreamClick}
              disabled={!streamKey}
            >
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" className="w-5 h-5">
                <path fillRule="evenodd" d="M4.5 5.653c0-1.427 1.529-2.33 2.779-1.643l11.54 6.347c1.295.712 1.295 2.573 0 3.286L7.28 19.99c-1.25.687-2.779-.217-2.779-1.643V5.653Z" clipRule="evenodd" />
              </svg>
              Watch Stream
            </button>

            <button 
              className='flex-1 py-3 px-6 bg-[var(--color-secondary)] hover:bg-[var(--color-secondary)]/90
                text-white font-medium rounded-md shadow-sm transition-colors duration-200
                focus:outline-none focus:ring-2 focus:ring-[var(--color-secondary)] focus:ring-opacity-50
                flex items-center justify-center gap-2'
              type='button' 
              onClick={onPublishStreamClick}
              disabled={!streamKey}
            >
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" className="w-5 h-5">
                <path d="M15.75 8.25a.75.75 0 0 1 .75.75c0 1.12-.492 2.126-1.27 2.812a.75.75 0 1 1-.992-1.124A2.243 2.243 0 0 0 15 9a.75.75 0 0 1 .75-.75Z" />
                <path fillRule="evenodd" d="M12 2.25c-5.385 0-9.75 4.365-9.75 9.75s4.365 9.75 9.75 9.75 9.75-4.365 9.75-9.75S17.385 2.25 12 2.25ZM4.575 15.6a8.25 8.25 0 0 0 9.348 4.425 1.966 1.966 0 0 0-1.84-1.275.983.983 0 0 1-.97-.822l-.073-.437c-.094-.565.25-1.11.8-1.267l.99-.282c.427-.123.783-.418.982-.816l.036-.073a1.453 1.453 0 0 1 2.328-.377L16.5 15.9c.612.54.94 1.28.94 2.058v.073a8.25 8.25 0 0 0 1.5-15.332 8.25 8.25 0 0 0-14.365 12.901Z" clipRule="evenodd" />
              </svg>
              Publish Stream
            </button>
          </div>
        </div>
        
        {/* Feature highlights */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mt-8">
          <div className="flex flex-col items-start">
            <div className="p-3 bg-[var(--color-primary)]/10 rounded-full mb-3">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="var(--color-primary)" className="w-6 h-6">
                <path d="M4.5 4.5a3 3 0 0 0-3 3v9a3 3 0 0 0 3 3h8.25a3 3 0 0 0 3-3v-9a3 3 0 0 0-3-3H4.5ZM19.94 18.75l-2.69-2.69V7.94l2.69-2.69c.944-.945 2.56-.276 2.56 1.06v11.38c0 1.336-1.616 2.005-2.56 1.06Z" />
              </svg>
            </div>
            <h3 className="text-lg font-medium mb-2">Real-time Streaming</h3>
            <p className="text-[var(--color-text-secondary)]">Stream high-quality video with sub-second latency using WebRTC technology.</p>
          </div>
          
          <div className="flex flex-col items-start">
            <div className="p-3 bg-[var(--color-secondary)]/10 rounded-full mb-3">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="var(--color-secondary)" className="w-6 h-6">
                <path d="M4.5 6.375a4.125 4.125 0 1 1 8.25 0 4.125 4.125 0 0 1-8.25 0ZM14.25 8.625a3.375 3.375 0 1 1 6.75 0 3.375 3.375 0 0 1-6.75 0ZM1.5 19.125a7.125 7.125 0 0 1 14.25 0v.003l-.001.119a.75.75 0 0 1-.363.63 13.067 13.067 0 0 1-6.761 1.873c-2.472 0-4.786-.684-6.76-1.873a.75.75 0 0 1-.364-.63l-.001-.122ZM17.25 19.128l-.001.144a2.25 2.25 0 0 1-.233.96 10.088 10.088 0 0 0 5.06-1.01.75.75 0 0 0 .42-.643 4.875 4.875 0 0 0-6.957-4.611 8.586 8.586 0 0 1 1.71 5.157v.003Z" />
              </svg>
            </div>
            <h3 className="text-lg font-medium mb-2">Built-in Chat</h3>
            <p className="text-[var(--color-text-secondary)]">Interactive chat room for each stream with user status indicators.</p>
          </div>
          
          <div className="flex flex-col items-start">
            <div className="p-3 bg-[var(--color-accent)]/10 rounded-full mb-3">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="var(--color-accent)" className="w-6 h-6">
                <path fillRule="evenodd" d="M14.615 1.595a.75.75 0 0 1 .359.852L12.982 9.75h7.268a.75.75 0 0 1 .548 1.262l-10.5 11.25a.75.75 0 0 1-1.272-.71l1.992-7.302H3.75a.75.75 0 0 1-.548-1.262l10.5-11.25a.75.75 0 0 1 .913-.143Z" clipRule="evenodd" />
              </svg>
            </div>
            <h3 className="text-lg font-medium mb-2">Easy Setup</h3>
            <p className="text-[var(--color-text-secondary)]">No public IP or port forwarding required. Just generate a stream key and go.</p>
          </div>
          
          <div className="flex flex-col items-start">
            <div className="p-3 bg-purple-100 rounded-full mb-3">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="#8b5cf6" className="w-6 h-6">
                <path fillRule="evenodd" d="M12 6.75a5.25 5.25 0 0 1 6.775-5.025.75.75 0 0 1 .313 1.248l-3.32 3.319c.063.475.276.934.641 1.299.365.365.824.578 1.3.64l3.318-3.319a.75.75 0 0 1 1.248.313 5.25 5.25 0 0 1-5.472 6.756c-1.018-.086-1.87.1-2.309.634L7.344 21.3A3.298 3.298 0 1 1 2.7 16.657l8.684-7.151c.533-.44.72-1.291.634-2.309A5.342 5.342 0 0 1 12 6.75ZM4.117 19.125a.75.75 0 0 1 .75-.75h.008a.75.75 0 0 1 .75.75v.008a.75.75 0 0 1-.75.75h-.008a.75.75 0 0 1-.75-.75v-.008Z" clipRule="evenodd" />
              </svg>
            </div>
            <h3 className="text-lg font-medium mb-2">Advanced Codecs</h3>
            <p className="text-[var(--color-text-secondary)]">Support for the latest video codecs for high quality at low bandwidth.</p>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Selection
