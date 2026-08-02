import { SpeakerWaveIcon, SpeakerXMarkIcon } from "@heroicons/react/16/solid";
import { useState, type WheelEvent } from "react";

interface VolumeComponentProps {
	isMuted: boolean;
	volume: number;
	onStateChanged: (isMuted: boolean) => void;
	onVolumeChanged: (value: number) => void;
}

const VolumeComponent = (props: VolumeComponentProps) => {
	const { isMuted, onStateChanged, onVolumeChanged, volume } = props;
	const [showSlider, setShowSlider] = useState<boolean>(false);

	const onVolumeChange = (newValue: number) => {
		if (isMuted && newValue !== 0) {
			onStateChanged(false)
		}
		if (!isMuted && newValue === 0) {
			onStateChanged(true)
		}

		onVolumeChanged(newValue);
	}

	return <div
		onMouseEnter={() => setShowSlider(true)}
		onMouseLeave={() => setShowSlider(false)}
		className="flex justify-start max-w-42 gap-2 items-center"
	>
		{isMuted && (
			<SpeakerXMarkIcon className="w-5" onClick={() => onStateChanged(false)} />
		)}
		{!isMuted && (
			<SpeakerWaveIcon className="w-5" onClick={() => onStateChanged(true)} />
		)}

		<VolumeSlider
			isVisible={showSlider}
			volume={volume}
			onVolumeChange={onVolumeChange}
		/>

	</div>
}

interface VolumeSliderProps {
	isVisible: boolean;
	volume: number;
	onVolumeChange: (value: number) => void
}
const VolumeSlider = (props: VolumeSliderProps) => {
	const { isVisible, onVolumeChange, volume } = props;

	const onVolumeWheel = (event: WheelEvent<HTMLDivElement>) => {
		event.preventDefault()

		let newValue = volume + (event.deltaY < 0 ? 1 : -1);

		if (newValue > 100) {
			newValue = 100
		}
		if (newValue < 0) {
			newValue = 0
		}

		onVolumeChange(newValue)
	}

	return <div
		id="volumeComponentWrapper"
		onWheel={onVolumeWheel}
			className={`bg-transparent cursor-pointer h-full ${!isVisible && `invisible`} flex flex-col justify-center`}>
		<input
			id="default-range"
			type="range"
			min={0}
			max={100}
			value={volume}
			onChange={(event) => onVolumeChange(parseInt(event.target.value))}
			className={`h-2 w-18 rounded-lg appearance-none cursor-pointer dark:bg-gray-700`}
		/>

	</div>
}

export default VolumeComponent
