import React, { useContext, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import AvailableStreams from "./AvailableStreams";
import { HeaderContext } from '../../providers/HeaderProvider';
import { LocaleContext } from '../../providers/LocaleProvider';
import Button from '../shared/Button';
import Toggle from '../shared/Toggle';
import Input from '../shared/Input';

const Frontpage = () => {
	const { locale } = useContext(LocaleContext)
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
	const onKeyUp = () => setIsDisabledSubmit(() => streamKey === '')

	return (
		<div className='flex mx-auto max-w-2xl pt-18 gap-2'>

			<div className='flex flex-col rounded-md bg-gray-800 shadow-md p-8 gap-2'>
				<h2 className="font-light leading-tight text-4xl mt-0 mb-2">{locale.frontpage.welcome}</h2>
				<p>{locale.frontpage.welcome_subtitle}</p>

				<div className='mt-4' />

				<Toggle
					titleLeft={locale.frontpage.toggle_watch}
					onClickLeft={() => setStreamType("Watch")}
					iconLeft='People'

					titleRight={locale.frontpage.toggle_stream}
					onClickRight={() => setStreamType("Share")}
					iconRight='Camera'

					selected={streamType === "Watch" ? "Left" : "Right"}

				/>

				<div className='flex flex-col justify-center'>
					<Input
						label="Stream key"
						value={streamKey}
						setValue={setStreamKey}
						hasAutofocus={true}
						placeholder={streamType === "Share" ?
							locale.frontpage.stream_input_placeholder_share :
							locale.frontpage.stream_input_placeholder_join}
						onKeyUp={onKeyUp}
						onEnterKeyUp={onSubmit}
					/>

					<Button
						title={streamType === "Share" ?
							locale.frontpage.stream_button_stream_start :
							locale.frontpage.stream_button_stream_join
						}
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
