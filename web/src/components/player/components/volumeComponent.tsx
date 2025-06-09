import React, {useEffect, useState} from "react";
import {SpeakerWaveIcon, SpeakerXMarkIcon} from "@heroicons/react/16/solid";

interface VolumeComponentProps {
	isMuted: boolean;
	onStateChanged: (isMuted: boolean) => void;
}

// TODO: 
// Implement volume bar
const VolumeComponent = (props: VolumeComponentProps) => {
	const [isMuted, setIsMuted] = useState<boolean>(props.isMuted);

	useEffect(() => {
		props.onStateChanged(isMuted);
	}, [isMuted]);

	if (isMuted) {
		return <SpeakerXMarkIcon onClick={() => setIsMuted((prev) => !prev)}/>
	}
	if (!isMuted) {
		return <SpeakerWaveIcon onClick={() => setIsMuted((prev) => !prev)}/>
	}
}
export default VolumeComponent
