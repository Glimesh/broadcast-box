import React, {createRef, useState} from 'react'
import {useNavigate} from 'react-router-dom'
import AvailableStreams from "./AvailableStreams";

const Frontpage = () => {
	const [streamType, setStreamType] = useState<'Watch' | 'Share'>('Watch');
	const streamKey = createRef<HTMLInputElement>()
	const navigate = useNavigate()

	const onStreamClick = () => {
		if(!streamKey.current || streamKey.current?.value === ''){
			return;
		}
		
		if(streamType === "Share"){
			navigate(`/publish/${streamKey.current.value}`)
		}

		if(streamType === "Watch"){
			navigate(`/${streamKey.current.value}`)
		}
	}

	return (
		<div className='space-y-4 mx-auto max-w-2xl pt-20 md:pt-24'>
			<div className='rounded-md bg-gray-800 shadow-md p-8'>
				<h2 className="font-light leading-tight text-4xl mt-0 mb-2">Welcome to Broadcast Box</h2>
				<p>Broadcast Box is a tool that allows you to efficiently stream high-quality video in real time, using the latest in video codecs and WebRTC technology.</p>

				<div className="flex rounded-md shadow-xs justify-center mt-6" role="group">

					<button
						type="button"
						onClick={() => setStreamType('Watch')}
						className={`flex items-center px-4 py-2 text-sm font-medium border border-gray-200 rounded-s-lg hover:text-blue-700 dark:border-gray-700 dark:text-white dark:hover:text-white dark:hover:bg-blue-700 dark:focus:ring-blue-500 dark:focus:text-white ${streamType === "Watch" ? "bg-blue-700" : ""}`}>
						<svg className="w-6 h-6 text-gray-800 dark:text-white" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="none" viewBox="0 0 24 24">
							<path stroke="currentColor" strokeLinecap="round" strokeWidth="2" d="M4.5 17H4a1 1 0 0 1-1-1 3 3 0 0 1 3-3h1m0-3.05A2.5 2.5 0 1 1 9 5.5M19.5 17h.5a1 1 0 0 0 1-1 3 3 0 0 0-3-3h-1m0-3.05a2.5 2.5 0 1 0-2-4.45m.5 13.5h-7a1 1 0 0 1-1-1 3 3 0 0 1 3-3h3a3 3 0 0 1 3 3 1 1 0 0 1-1 1Zm-1-9.5a2.5 2.5 0 1 1-5 0 2.5 2.5 0 0 1 5 0Z"/>
						</svg>
						I want to watch
					</button>
					<button
						type="button"
						onClick={() => setStreamType('Share')}
						className={`flex items-center px-4 py-2 text-sm font-medium border border-gray-200 rounded-e-lg hover:text-blue-700 dark:border-gray-700 dark:text-white dark:hover:text-white dark:hover:bg-blue-700 dark:focus:ring-blue-500 dark:focus:text-white ${streamType === "Share" ? "bg-blue-700" : ""}`}>
						<svg className="w-6 h-6 text-gray-800 dark:text-white" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="none" viewBox="0 0 24 24">
							<path stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M14 6H4a1 1 0 0 0-1 1v10a1 1 0 0 0 1 1h10a1 1 0 0 0 1-1V7a1 1 0 0 0-1-1Zm7 11-6-2V9l6-2v10Z"/>
						</svg>
						I want to stream
					</button>

				</div>

					<div className='flex flex-col my-4 justify-center'>
						<label className='block text-sm font-bold mb-2' htmlFor='streamKey'>
							Stream Key
						</label>

						<input
							className='mb-2 appearance-none border w-full py-2 px-3 leading-tight focus:outline-hidden focus:shadow-outline bg-gray-700 border-gray-700 text-white rounded-sm shadow-md placeholder-gray-200'
							id='streamKey'
							placeholder={`Insert the key you of the stream you want to ${streamType === "Share" ? 'share' : 'join'}`}
							type='text'
							onKeyUp={(e => {
								if(e.key === "Enter"){
									onStreamClick()
								}
							})}
							ref={streamKey}
							autoFocus/>

						<button
							className={`py-2 px-4 ${streamKey.current?.value.length === 0 ? "bg-gray-700" : "bg-blue-500"} text-white font-semibold rounded-lg shadow-md ${streamKey.current?.value.length === 0 ? "hover:bg-gray-600" : "hover:bg-blue-700" } focus:outline-hidden focus:ring-2 focus:ring-blue-400 focus:ring-opacity-75`}
							disabled={streamKey.current?.value.length === 0}
							type='button'
							onClick={onStreamClick}>
							{streamType === "Share" ? "Start stream" : "Join stream"}
						</button>
					</div>

				<AvailableStreams/>
			</div>
		</div>
	)
}

export default Frontpage