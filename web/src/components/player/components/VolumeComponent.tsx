import React, {useEffect, useRef, useState} from "react";
import {SpeakerWaveIcon, SpeakerXMarkIcon} from "@heroicons/react/16/solid";

interface VolumeComponentProps {
	isMuted: boolean;
	onStateChanged: (isMuted: boolean) => void;
	onVolumeChanged: (value: number) => void;
}

const VolumeComponent = (props: VolumeComponentProps) => {
	const [isMuted, setIsMuted] = useState<boolean>(props.isMuted);
	const [showSlider, setShowSlider] = useState<boolean>(false);
	const volumeRef = useRef<number>(20);
	
	useEffect(() => {
		props.onStateChanged(isMuted);
	}, [isMuted]);
	
	const onVolumeChange = (newValue: number) => {
		if(isMuted && newValue !== 0){
			setIsMuted((_) => false)
		}
		if(!isMuted && newValue === 0){
			setIsMuted((_) => true)
		}
		
		props.onVolumeChanged(newValue / 100);
	}

	return <div
		onMouseEnter={() => setShowSlider(true)}
		onMouseLeave={() => setShowSlider(false)}
		className="flex justify-start max-w-42 gap-2 items-center"
	>
		{isMuted && (
			<SpeakerXMarkIcon className="w-5" onClick={() => setIsMuted((prev) => !prev)}/>
		)}
		{!isMuted && (
			<SpeakerWaveIcon className="w-5" onClick={() => setIsMuted((prev) => !prev)}/>
		)}
		<input
			id="default-range"
			type="range"
			max={100}
			defaultValue={volumeRef.current}
			onChange={(event) => onVolumeChange(parseInt(event.target.value))}
			className={
				`
					${!showSlider && `
						invisible
					`} 
				w-18 h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer dark:bg-gray-700`}/>
	</div>
}
export default VolumeComponent
