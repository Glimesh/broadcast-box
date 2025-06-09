import {useEffect} from "react";

export enum ShortcutEvent {
	PlayPause,
	MuteUnmute,
	FullscreenToggle,
	CinemaMode
} 

const useKeyboardShortcuts = (onShortcut: (event: ShortcutEvent) => void) => {
	useEffect(() => {
		const handleKeyDown = (event: KeyboardEvent) => {
			if (event.key === 'f') {
				event.preventDefault()
				onShortcut(ShortcutEvent.FullscreenToggle);
			}

			if (event.code === "Space") {
				event.preventDefault()
				onShortcut(ShortcutEvent.PlayPause);
			}

			if (event.key === 'm') {
				event.preventDefault()
				onShortcut(ShortcutEvent.MuteUnmute);
			}

			if (event.key === 'c') {
				event.preventDefault()
				onShortcut(ShortcutEvent.CinemaMode);
			}
		}

		window.addEventListener("keydown", handleKeyDown, { passive: false})

		return () => {
			window.removeEventListener("keydown", handleKeyDown)
		}

	}, [onShortcut]);
}

export default useKeyboardShortcuts