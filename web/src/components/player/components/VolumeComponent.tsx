import { SpeakerWaveIcon, SpeakerXMarkIcon } from "@heroicons/react/16/solid";
import { useCallback, useEffect, useRef, useState } from "react";

interface VolumeComponentProps {
  isMuted: boolean;
  isDisabled?: boolean;
  initialVolume?: number;
  onStateChanged: (isMuted: boolean) => void;
  onVolumeChanged: (value: number) => void;
}

const VolumeComponent = (props: VolumeComponentProps) => {
  const { isDisabled, initialVolume, onStateChanged, onVolumeChanged } = props;
  const [isMuted, setIsMuted] = useState<boolean>(props.isMuted);

  useEffect(() => {
    onStateChanged(isMuted);
  }, [isMuted, onStateChanged]);

  const onVolumeChange = (newValue: number) => {
    if (isMuted && newValue !== 0) {
      setIsMuted(false);
    }
    if (!isMuted && newValue === 0) {
      setIsMuted(true);
    }

    onVolumeChanged(newValue);
  };

  if (isDisabled) {
    return <SpeakerXMarkIcon className="video-control opacity-25" />;
  }

  return (
    <div className="video-volume">
      <button
        type="button"
        className="video-control"
        title={isMuted ? "Unmute" : "Mute"}
        onClick={() => setIsMuted((prev) => !prev)}
      >
        {isMuted
          ? <SpeakerXMarkIcon className="h-5 w-5" />
          : <SpeakerWaveIcon className="h-5 w-5" />}
      </button>

      <VolumeSlider
        initialVolume={initialVolume}
        onVolumeChange={onVolumeChange}
      />
    </div>
  );
};

interface VolumeSliderProps {
  initialVolume?: number;
  onVolumeChange: (value: number) => void;
}

const VolumeSlider = (props: VolumeSliderProps) => {
  const { initialVolume = 100, onVolumeChange } = props;
  const inputRef = useRef<HTMLInputElement>(null);
  const volumeRef = useRef<number>(initialVolume);
  const [currentVolume, setCurrentVolume] = useState<number>(initialVolume);

  const setVolume = useCallback((value: number) => {
    onVolumeChange(value);
    setCurrentVolume(() => value);
    volumeRef.current = value;
  }, [onVolumeChange]);

  useEffect(() => {
    const inputElement = inputRef.current;
    const wheelHandler = (event: WheelEvent) => {
      event.preventDefault();

      let newValue = volumeRef.current + (event.deltaY < 0 ? 1 : -1);

      if (newValue > 100) {
        newValue = 100;
      }
      if (newValue < 0) {
        newValue = 0;
      }

      setVolume(newValue);
    };

    inputElement?.addEventListener("wheel", wheelHandler, { passive: false });

    return () => {
      inputElement?.removeEventListener("wheel", wheelHandler);
    };
  }, [setVolume]);

  return (
    <div id="volumeComponentWrapper" ref={inputRef} className="volume-wrapper">
      <input
        id="default-range"
        type="range"
        min={0}
        max={100}
        value={currentVolume}
        onChange={(event) => setVolume(parseInt(event.target.value))}
        className="volume-slider"
        aria-label="Volume"
      />
    </div>
  );
};

export default VolumeComponent;
