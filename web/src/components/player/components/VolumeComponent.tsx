import { SpeakerWaveIcon, SpeakerXMarkIcon } from "@heroicons/react/16/solid";
import React, { useEffect, useRef, useState } from "react";

interface VolumeComponentProps {
	isMuted: boolean;
	isDisabled?: boolean;
	onStateChanged: (isMuted: boolean) => void;
	onVolumeChanged: (value: number) => void;
}

const VolumeComponent = (props: VolumeComponentProps) => {
	const [isMuted, setIsMuted] = useState<boolean>(props.isMuted);
	const [showSlider, setShowSlider] = useState<boolean>(false);

	useEffect(() => {
		props.onStateChanged(isMuted);
	}, [isMuted]);

	const onVolumeChange = (newValue: number) => {
		if (isMuted && newValue !== 0) {
			setIsMuted((_) => false)
		}
		if (!isMuted && newValue === 0) {
			setIsMuted((_) => true)
		}

		props.onVolumeChanged(newValue);
	}

	if (props.isDisabled) {
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
	const inputRef = useRef<HTMLInputElement>(null);
	const volumeRef = useRef<number>(50);

	// Forces UI rendering
	const [_, setCurrentVolume] = useState<number>(volumeRef.current)

	const setVolume = (value: number) => {
		props.onVolumeChange(value)
		setCurrentVolume(() => value)
		volumeRef.current = value
	}

	useEffect(() => {
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

		inputRef.current?.addEventListener("wheel", wheelHandler, { passive: false })

		return () => {
			inputRef.current?.removeEventListener("wheel", wheelHandler)
		}
	}, [])

	return <div
		id="volumeComponentWrapper"
		ref={inputRef}
		className={`bg-transparent cursor-pointer h-full ${!props.isVisible && `invisible`} flex flex-col justify-center`}>
		<input
			id="default-range"
			type="range"
			min={0}
			max={100}
			value={volumeRef.current}
			onChange={(event) => setVolume(parseInt(event.target.value))}
			className={`h-2 w-18 rounded-lg appearance-none cursor-pointer dark:bg-gray-700`}
		/>

	</div>
}

export default VolumeComponent
