import { useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";
import { useLocation } from "react-router-dom";
import { CinemaModeContext } from "../../providers/CinemaModeProvider";
import {
	DEFAULT_MUSIC_PLAYLIST,
	MUSIC_PROMPTED_STORAGE_KEY,
	MUSIC_STORAGE_KEY,
	MUSIC_VOLUME_STORAGE_KEY,
	type MusicPlaylist,
} from "../../config/siteExtras";

const IFRAME_ID = "bb-music-player";
const NON_STREAM_PATHS = ["/statistics", "/login", "/account", "/admin"];

const matchesPath = (pathname: string, path: string) => (
	pathname === path || pathname === `${path}/` || pathname.startsWith(`${path}/`)
);

const isStreamPath = (pathname: string) => {
	if (pathname === "/") return false;
	if (NON_STREAM_PATHS.some((path) => matchesPath(pathname, path))) return false;
	if (pathname.startsWith("/u/")) return false;
	return true;
};

const getStoredBoolean = (key: string, fallback: boolean) => {
	try {
		const stored = localStorage.getItem(key);
		if (stored === "true") return true;
		if (stored === "false") return false;
	} catch {
		// localStorage may be unavailable.
	}

	return fallback;
};

const getStoredNumber = (key: string, fallback: number) => {
	try {
		const raw = localStorage.getItem(key);
		if (raw === null) return fallback;

		const stored = Number(raw);
		if (Number.isFinite(stored) && stored >= 0 && stored <= 100) return stored;
	} catch {
		// localStorage may be unavailable.
	}

	return fallback;
};

export default function MusicPlayer() {
	const { pathname } = useLocation();
	const { cinemaMode } = useContext(CinemaModeContext);
	const [isOpen, setIsOpen] = useState(false);
	const [isEnabled, setIsEnabled] = useState(() => getStoredBoolean(MUSIC_STORAGE_KEY, true));
	const [hasPrompted, setHasPrompted] = useState(() => getStoredBoolean(MUSIC_PROMPTED_STORAGE_KEY, false));
	const [volume, setVolume] = useState(() => getStoredNumber(MUSIC_VOLUME_STORAGE_KEY, DEFAULT_MUSIC_PLAYLIST.defaultVolume));
	const [playlist, setPlaylist] = useState<MusicPlaylist>(DEFAULT_MUSIC_PLAYLIST);
	const [isReady, setIsReady] = useState(false);
	const [isPlaying, setIsPlaying] = useState(false);
	const [playAttempt, setPlayAttempt] = useState(0);
	const iframeRef = useRef<HTMLIFrameElement>(null);
	const hasInteractedRef = useRef(false);

	const isBroadcastPage = pathname === "/publish" || pathname.startsWith("/publish/");
	const isStreamPage = isStreamPath(pathname);
	const isHidden = cinemaMode || isStreamPage || isBroadcastPage;
	const showConsent = !hasPrompted && !isHidden;
	const shouldPlay = isEnabled && !isHidden && !showConsent;

	const embedUrl = useMemo(() => {
		const playlistIds = encodeURIComponent(playlist.videoIds.join(","));
		const origin = encodeURIComponent(window.location.origin);
		return `https://www.youtube-nocookie.com/embed/${playlist.videoIds[0]}?enablejsapi=1&autoplay=0&loop=1&playlist=${playlistIds}&controls=0&modestbranding=1&rel=0&playsinline=1&origin=${origin}`;
	}, [playlist]);

	const sendCommand = useCallback((func: string, args: unknown[] = []) => {
		iframeRef.current?.contentWindow?.postMessage(
			JSON.stringify({ event: "command", func, args }),
			"*",
		);
	}, []);

	useEffect(() => {
		const handleFirstInteraction = () => {
			if (hasInteractedRef.current || showConsent) return;
			hasInteractedRef.current = true;
			setPlayAttempt((attempt) => attempt + 1);

			if (isReady && isEnabled && !isHidden) {
				setIsPlaying(true);
				sendCommand("playVideo");
			}
		};

		window.addEventListener("pointerdown", handleFirstInteraction);
		window.addEventListener("keydown", handleFirstInteraction);

		return () => {
			window.removeEventListener("pointerdown", handleFirstInteraction);
			window.removeEventListener("keydown", handleFirstInteraction);
		};
	}, [isReady, isEnabled, isHidden, sendCommand, showConsent]);

	useEffect(() => {
		const handleMessage = (event: MessageEvent) => {
			const isYouTubeOrigin = event.origin === "https://www.youtube.com"
				|| event.origin === "https://www.youtube-nocookie.com";
			if (!isYouTubeOrigin || event.source !== iframeRef.current?.contentWindow) return;
			if (typeof event.data !== "string") return;

			try {
				const data = JSON.parse(event.data) as { event?: string; info?: number };

				if (data.event === "onReady") {
					setIsReady(true);
				}

				if (data.event === "onStateChange") {
					if (data.info === 1) {
						setIsPlaying(true);
					} else if (data.info === 0 || data.info === 2 || data.info === 5) {
						setIsPlaying(false);
					}
				}

				if (data.event === "onError") {
					setIsPlaying(false);
				}
			} catch {
				// Ignore messages that are not from the YouTube player.
			}
		};

		window.addEventListener("message", handleMessage);
		return () => window.removeEventListener("message", handleMessage);
	}, []);

	useEffect(() => {
		if (!isReady) return;
		sendCommand("setVolume", [volume]);
		sendCommand(shouldPlay ? "playVideo" : "pauseVideo");
	}, [isReady, playAttempt, sendCommand, shouldPlay, volume]);

	useEffect(() => {
		try {
			localStorage.setItem(MUSIC_STORAGE_KEY, String(isEnabled));
		} catch {
			// The setting still applies for this session.
		}
	}, [isEnabled]);

	useEffect(() => {
		try {
			localStorage.setItem(MUSIC_PROMPTED_STORAGE_KEY, String(hasPrompted));
		} catch {
			// The setting still applies for this session.
		}
	}, [hasPrompted]);

	useEffect(() => {
		try {
			localStorage.setItem(MUSIC_VOLUME_STORAGE_KEY, String(volume));
		} catch {
			// The setting still applies for this session.
		}
	}, [volume]);

	useEffect(() => {
		fetch("/api/music")
			.then((result) => (result.ok ? result.json() : null))
			.then((result: MusicPlaylist | null) => {
				if (result && Array.isArray(result.videoIds) && result.videoIds.length > 0) {
					setPlaylist(result);
				}
			})
			.catch(() => {});
	}, []);

	const toggleMusic = () => {
		setHasPrompted(true);

		if (isPlaying) {
			setIsEnabled(false);
			setIsPlaying(false);
			sendCommand("pauseVideo");
			return;
		}

		setIsEnabled(true);
		setIsPlaying(true);
		sendCommand("playVideo");
	};

	const enableMusic = () => {
		setHasPrompted(true);
		setIsEnabled(true);
		setIsPlaying(true);
		setPlayAttempt((attempt) => attempt + 1);
		sendCommand("playVideo");
	};

	const dismissMusicPrompt = () => {
		setHasPrompted(true);
		setIsEnabled(false);
		setIsPlaying(false);
		sendCommand("pauseVideo");
	};

	const handleIframeLoad = () => {
		const announceListening = () => {
			iframeRef.current?.contentWindow?.postMessage(
				JSON.stringify({ event: "listening", id: IFRAME_ID }),
				"*",
			);
		};

		announceListening();
		window.setTimeout(announceListening, 600);
		window.setTimeout(announceListening, 1600);
	};

	return (
		<>
			<iframe
				ref={iframeRef}
				id={IFRAME_ID}
				title="frutiger aero radio"
				className="music-iframe"
				src={embedUrl}
				allow="autoplay; encrypted-media; picture-in-picture"
				onLoad={handleIframeLoad}
				aria-hidden="true"
				tabIndex={-1}
			/>

			{showConsent && (
				<div
					className="modal-backdrop music-consent"
					role="dialog"
					aria-modal="true"
					aria-labelledby="music-consent-title"
				>
					<div className="window modal-window music-consent__window">
						<div className="window-bar">
							<span className="window-dots" aria-hidden="true">
								<span className="window-dot" />
								<span className="window-dot" />
								<span className="window-dot" />
							</span>
							<span>background tunes ♡</span>
							<span className="window-bar__spacer" />
							<span className="window-bar__code">sound ♡</span>
						</div>

						<div className="music-consent__body">
							<p className="music-consent__kicker">♫ one little thing</p>
							<h2 id="music-consent-title" className="music-consent__title">can we play some cozy tunes?</h2>
							<p className="music-consent__copy">
								we&apos;d love to play a little nintendo mix while you browse.
								browsers need one tap before audio can start, so this is the
								permission prompt. it starts at a gentle 30% by default.
							</p>

							<div className="music-consent__actions">
								<button
									type="button"
									className="pixel-btn pixel-btn--lime"
									onClick={enableMusic}
									autoFocus
								>
									yes, play ♡
								</button>
								<button
									type="button"
									className="pixel-btn pixel-btn--ghost"
									onClick={dismissMusicPrompt}
								>
									not now
								</button>
							</div>

							<p className="music-consent__note">
								we&apos;ll remember this choice. music pauses on streams.
							</p>
						</div>
					</div>
				</div>
			)}

			<div
				className={`music-player ${isHidden ? "is-hidden" : ""} ${isOpen ? "is-open" : ""}`}
				aria-hidden={isHidden}
			>
				<div className="music-player__head">
					<button
						type="button"
						className="music-player__open"
						onClick={() => setIsOpen((prev) => !prev)}
						aria-label={isOpen ? "collapse music player" : "expand music player"}
						aria-expanded={isOpen}
					>
						<span className="music-player__note" aria-hidden="true">♫</span>
						<span className="music-player__copy">
							<strong>tunes ♡</strong>
							<small>{isPlaying ? "now playing" : isEnabled ? "tap to start ♡" : "paused ♡"}</small>
						</span>
						<span className={`music-eq ${isPlaying ? "is-playing" : ""}`} aria-hidden="true">
							<i />
							<i />
							<i />
							<i />
						</span>
					</button>

					<button
						type="button"
						className="music-player__play"
						onClick={toggleMusic}
						aria-label={isPlaying ? "pause music" : "play music"}
					>
						{isPlaying ? "❚❚" : "▶"}
					</button>
				</div>

				<div className="music-player__body">
					<p className="music-player__sub">{playlist.subtitle}</p>

					<label className="music-player__volume">
						<span>vol</span>
						<input
							type="range"
							min="0"
							max="100"
							value={volume}
							onChange={(event) => setVolume(Number(event.target.value))}
							aria-label="music volume"
						/>
						<span>{volume}%</span>
					</label>

					<div className="music-player__foot">
						<a href={`https://www.youtube.com/watch_videos?video_ids=${encodeURIComponent(playlist.videoIds.join(","))}`} target="_blank" rel="noreferrer">
							open on youtube ↗
						</a>
						<span>pauses on streams</span>
					</div>

					{!isPlaying && (
						<p className="music-player__hint">
							{isEnabled ? "tap play to start the mix ♡" : "paused ♡"}
						</p>
					)}
				</div>
			</div>
		</>
	);
}
