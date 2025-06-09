import React, {useEffect, useState} from "react";
import {PauseIcon, PlayIcon} from "@heroicons/react/16/solid";

interface PlayPauseComponentProps {
	videoRef: React.RefObject<HTMLVideoElement | null>;
}

const PlayPauseComponent = (props: PlayPauseComponentProps) => {
	const [isPaused, setIsPaused] = useState<boolean>(true);

	if (props.videoRef.current === null) {
		return <></>;
	}

	useEffect(() => {
		if (props.videoRef.current === null) {
			return;
		}

		const canPlayHandler = (_: Event) => props.videoRef.current?.play()
		const playingHandler = (_: Event) => setIsPaused(() => false)
		const pauseHandler = (_: Event) => setIsPaused(() => true);

		props.videoRef.current.addEventListener("canplay", canPlayHandler)
		props.videoRef.current.addEventListener("playing", playingHandler)
		props.videoRef.current.addEventListener("pause", pauseHandler)

		return () => {
			if (props.videoRef.current) {
				props.videoRef.current.removeEventListener("canplay", canPlayHandler);
				props.videoRef.current.removeEventListener("playing", playingHandler);
				props.videoRef.current.removeEventListener("pause", pauseHandler);
			}
		}
	}, []);

	useEffect(() => {
		if(isPaused){
			props.videoRef.current?.pause();
		}
		if(!isPaused){
			props.videoRef.current?.play().catch((err) => console.error("VideoError", err));
		}
	}, [isPaused]);

	if (isPaused) {
		return <PlayIcon onClick={() => props.videoRef.current?.play()}/>
	}
	if (!isPaused) {
		return <PauseIcon onClick={() => props.videoRef.current?.pause()}/>
	}
}

export default PlayPauseComponent