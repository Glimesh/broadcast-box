import { SpeakerWaveIcon, SpeakerXMarkIcon } from "@heroicons/react/16/solid";
import { useCallback, useEffect, useRef, useState } from "react";

interface VolumeComponentProps {
	isMuted: boolean;
	isDisabled?: boolean;
	onStateChanged: (isMuted: boolean) => void;
	onVolumeChanged: (value: number) => void;
}

const VolumeComponent = (props: VolumeComponentProps) => {
	const { isDisabled, onStateChanged, onVolumeChanged } = props;
	const [isMuted, setIsMuted] = useState<boolean>(props.isMuted);
	const [showSlider, setShowSlider] = useState<boolean>(false);

	useEffect(() => {
		onStateChanged(isMuted);
	}, [isMuted, onStateChanged]);

	const onVolumeChange = (newValue: number) => {
		if (isMuted && newValue !== 0) {
			setIsMuted(false)
		}
		if (!isMuted && newValue === 0) {
			setIsMuted(true)
		}

		onVolumeChanged(newValue);
	}

	if (isDisabled) {
		return (<SpeakerXMarkIcon className="w-5 opacity-25" />)
	}

	return <div
		onMouseEnter={() => setShowSlider(true)}
		onMouseLeave={() => setShowSlider(false)}
		className="flex justify-start max-w-42 gap-2 items-center"
	>
		{isMuted && (
			<SpeakerXMarkIcon className="w-5" onClick={() => setIsMuted((prev) => !prev)} />
		)}
		{!isMuted && (
			<SpeakerWaveIcon className="w-5" onClick={() => setIsMuted((prev) => !prev)} />
		)}

		<VolumeSlider
			isVisible={showSlider}
			onVolumeChange={onVolumeChange}
		/>

	</div>
}

interface VolumeSliderProps {
	isVisible: boolean;
	onVolumeChange: (value: number) => void
}
const VolumeSlider = (props: VolumeSliderProps) => {
	const { isVisible, onVolumeChange } = props;
	const inputRef = useRef<HTMLInputElement>(null);
	const volumeRef = useRef<number>(50);
	const [currentVolume, setCurrentVolume] = useState<number>(50)

	const setVolume = useCallback((value: number) => {
		onVolumeChange(value)
		setCurrentVolume(() => value)
		volumeRef.current = value
	}, [onVolumeChange])

	useEffect(() => {
		const inputElement = inputRef.current
		const wheelHandler = (event: WheelEvent) => {
			event.preventDefault()

			let newValue = volumeRef.current + (event.deltaY < 0 ? 1 : -1);

			if (newValue > 100) {
				newValue = 100
			}
			if (newValue < 0) {
				newValue = 0
			}

			setVolume(newValue)
		}

		inputElement?.addEventListener("wheel", wheelHandler, { passive: false })

		return () => {
			inputElement?.removeEventListener("wheel", wheelHandler)
		}
	}, [setVolume])

	return <div
		id="volumeComponentWrapper"
		ref={inputRef}
			className={`bg-transparent cursor-pointer h-full ${!isVisible && `invisible`} flex flex-col justify-center`}>
		<input
			id="default-range"
			type="range"
			min={0}
			max={100}
			value={currentVolume}
			onChange={(event) => setVolume(parseInt(event.target.value))}
			className={`h-2 w-18 rounded-lg appearance-none cursor-pointer dark:bg-gray-700`}
		/>

	</div>
}

export default VolumeComponent
