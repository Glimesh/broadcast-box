import React, { useContext, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import AvailableStreams from "./AvailableStreams";
import { HeaderContext } from '../../providers/HeaderProvider';
import Button from '../shared/Button';
import Toggle from '../shared/Toggle';
import Input from '../shared/Input';

const Frontpage = () => {
	const { setTitle } = useContext(HeaderContext)
	const [isDisabledSubmit, setIsDisabledSubmit] = useState(true)
	const [streamType, setStreamType] = useState<'Watch' | 'Share'>('Watch');
	const [streamKey, setStreamKey] = useState("")
	const navigate = useNavigate()
	setTitle("")

	const onSubmit = () => {
		if (!streamKey || streamKey === '') {
			return;
		}

		if (streamType === "Share") {
			navigate(`/publish/${streamKey}`)
		}

		if (streamType === "Watch") {
			navigate(`/${streamKey}`)
		}
	}
	const onKeyUp = () => {
		setIsDisabledSubmit(() => streamKey === '')
	}

	return (
		<div className='space-y-4 mx-auto max-w-2xl pt-20 md:pt-24'>

			<div className='rounded-md bg-gray-800 shadow-md p-8'>
				<h2 className="font-light leading-tight text-4xl mt-0 mb-2">Welcome to Broadcast Box</h2>
				<p>Broadcast Box is a tool that allows you to efficiently stream high-quality video in real time, using the latest in video codecs and WebRTC technology.</p>

				<div className='mt-4' />

				<Toggle
					titleLeft='I want to watch'
					onClickLeft={() => setStreamType("Watch")}
					iconLeft='People'

					titleRight='I want to stream'
					onClickRight={() => setStreamType("Share")}
					iconRight='Camera'

					selected={streamType === "Watch" ? "Left" : "Right"}

				/>

				<div className='flex flex-col mt-2 justify-center'>
					<Input
						label="Stream key"
						value={streamKey}
						setValue={setStreamKey}
						hasAutofocus={true}
						placeholder={`Insert the key of the stream you want to ${streamType === "Share" ? 'share' : 'join'}`}
						onKeyUp={onKeyUp}
						onEnterKeyUp={onSubmit}
					/>

					<Button
						title={streamType === "Share" ? "Start stream" : "Join stream"}
						center
						isDisabled={isDisabledSubmit}
						onClick={onSubmit}
						stretch
					/>
				</div>

				<AvailableStreams />
			</div>

		</div>
	)
}

export default Frontpage
