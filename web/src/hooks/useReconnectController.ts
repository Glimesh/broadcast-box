import { useCallback, useEffect, useMemo, useRef, useState } from "react";

interface ReconnectControllerOptions {
	baseDelayMs?: number;
	maxDelayMs?: number;
	maxAttempts?: number;
	jitterRatio?: number;
	pauseWhileHidden?: boolean;
}

interface ReconnectControllerState {
	attempt: number;
	isReconnecting: boolean;
	isExhausted: boolean;
	nextDelayMs: number | null;
}

interface ScheduleReconnectOptions {
	immediate?: boolean;
}

interface ReconnectController {
	state: ReconnectControllerState;
	scheduleReconnect: (
		onReconnect: () => void,
		options?: ScheduleReconnectOptions,
	) => boolean;
	reset: () => void;
	cancel: () => void;
}

const PAUSE_RETRY_MS = 1_000;

const clampDelay = (delayMs: number, maxDelayMs: number) => {
	if (delayMs > maxDelayMs) {
		return maxDelayMs;
	}

	return delayMs;
};

const getBackoffDelay = (
	attempt: number,
	baseDelayMs: number,
	maxDelayMs: number,
) => {
	const rawDelay = baseDelayMs * Math.pow(2, Math.max(0, attempt - 1));
	return clampDelay(rawDelay, maxDelayMs);
};

const withJitter = (delayMs: number, jitterRatio: number) => {
	const jitterRange = Math.floor(delayMs * jitterRatio);
	if (jitterRange <= 0) {
		return delayMs;
	}

	const jitter =
		Math.floor(Math.random() * (jitterRange * 2 + 1)) - jitterRange;
	return Math.max(0, delayMs + jitter);
};

const shouldPauseReconnect = (pauseWhileHidden: boolean) => {
	if (typeof window === "undefined") {
		return false;
	}

	if (typeof navigator !== "undefined" && navigator.onLine === false) {
		return true;
	}

	if (pauseWhileHidden && typeof document !== "undefined") {
		return document.visibilityState !== "visible";
	}

	return false;
};

export function useReconnectController(
	options?: ReconnectControllerOptions,
): ReconnectController {
	const {
		baseDelayMs = 500,
		maxDelayMs = 8_000,
		maxAttempts = 8,
		jitterRatio = 0.2,
		pauseWhileHidden = true,
	} = options ?? {};

	const timeoutRef = useRef<number | undefined>(undefined);
	const stoppedRef = useRef(false);
	const [attempt, setAttempt] = useState(0);
	const [isReconnecting, setIsReconnecting] = useState(false);
	const [nextDelayMs, setNextDelayMs] = useState<number | null>(null);
	const [isExhausted, setIsExhausted] = useState(false);

	const clearPendingTimer = useCallback(() => {
		clearTimeout(timeoutRef.current);
		timeoutRef.current = undefined;
	}, []);

	const cancel = useCallback(() => {
		stoppedRef.current = true;
		clearPendingTimer();
		setIsReconnecting(false);
		setNextDelayMs(null);
	}, [clearPendingTimer]);

	const reset = useCallback(() => {
		stoppedRef.current = false;
		clearPendingTimer();
		setAttempt(0);
		setIsReconnecting(false);
		setIsExhausted(false);
		setNextDelayMs(null);
	}, [clearPendingTimer]);

	const scheduleReconnect = useCallback(
		(onReconnect: () => void, scheduleOptions?: ScheduleReconnectOptions) => {
			if (stoppedRef.current) {
				return false;
			}

			setIsExhausted(false);

			let scheduled = false;
			setAttempt((currentAttempt) => {
				const nextAttempt = currentAttempt + 1;

				if (!scheduleOptions?.immediate && nextAttempt > maxAttempts) {
					setIsExhausted(true);
					setIsReconnecting(false);
					setNextDelayMs(null);
					scheduled = false;
					return currentAttempt;
				}

				clearPendingTimer();
				const delayMs = scheduleOptions?.immediate
					? 0
					: withJitter(
							getBackoffDelay(nextAttempt, baseDelayMs, maxDelayMs),
							jitterRatio,
						);

				setIsReconnecting(true);
				setNextDelayMs(delayMs);
				scheduled = true;

				const triggerReconnect = () => {
					if (stoppedRef.current) {
						return;
					}

					if (shouldPauseReconnect(pauseWhileHidden)) {
						timeoutRef.current = setTimeout(triggerReconnect, PAUSE_RETRY_MS);
						return;
					}

					setIsReconnecting(false);
					setNextDelayMs(null);
					onReconnect();
				};

				timeoutRef.current = setTimeout(triggerReconnect, delayMs);
				return nextAttempt;
			});

			return scheduled;
		},
		[
			baseDelayMs,
			clearPendingTimer,
			jitterRatio,
			maxAttempts,
			maxDelayMs,
			pauseWhileHidden,
		],
	);

	useEffect(() => {
		return () => {
			clearPendingTimer();
		};
	}, [clearPendingTimer]);

	const state = useMemo<ReconnectControllerState>(
		() => ({
			attempt,
			isReconnecting,
			isExhausted,
			nextDelayMs,
		}),
		[attempt, isExhausted, isReconnecting, nextDelayMs],
	);

	return {
		state,
		scheduleReconnect,
		reset,
		cancel,
	};
}
