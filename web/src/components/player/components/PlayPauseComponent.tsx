import { useEffect, useState, type RefObject } from "react";
import { PauseIcon, PlayIcon } from "@heroicons/react/16/solid";

interface PlayPauseComponentProps {
  videoRef: RefObject<HTMLVideoElement | null>;
}

const PlayPauseComponent = (props: PlayPauseComponentProps) => {
  const { videoRef } = props;
  const [isPaused, setIsPaused] = useState<boolean>(true);

  useEffect(() => {
    const videoElement = videoRef.current;
    if (videoElement === null) {
      return;
    }

    const canPlayHandler = () => videoElement.play();
    const playingHandler = () => setIsPaused(() => false);
    const pauseHandler = () => setIsPaused(() => true);

    videoElement.addEventListener("canplay", canPlayHandler);
    videoElement.addEventListener("playing", playingHandler);
    videoElement.addEventListener("pause", pauseHandler);

    return () => {
      videoElement.removeEventListener("canplay", canPlayHandler);
      videoElement.removeEventListener("playing", playingHandler);
      videoElement.removeEventListener("pause", pauseHandler);
    };
  }, [videoRef]);

  useEffect(() => {
    if (isPaused) {
      videoRef.current?.pause();
    }
    if (!isPaused) {
      videoRef.current?.play().catch((err) => console.error("VideoError", err));
    }
  }, [isPaused, videoRef]);

  return (
    <button
      type="button"
      className="video-control"
      title={isPaused ? "Play" : "Pause"}
      onClick={() => {
        if (isPaused) {
          void videoRef.current?.play();
        } else {
          videoRef.current?.pause();
        }
      }}
    >
      {isPaused
        ? <PlayIcon className="h-5 w-5" />
        : <PauseIcon className="h-5 w-5" />}
    </button>
  );
};

export default PlayPauseComponent;
