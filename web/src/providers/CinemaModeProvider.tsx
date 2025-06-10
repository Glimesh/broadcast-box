import React, {Dispatch, SetStateAction, useEffect, useMemo, useState} from "react";
import {useSearchParams} from "react-router-dom";

interface CinemaModeProviderContextProps{
	cinemaMode: boolean;
	setCinemaMode: Dispatch<SetStateAction<boolean>>;
	toggleCinemaMode: () => void;
}

export const CinemaModeContext = React.createContext<CinemaModeProviderContextProps>({
	cinemaMode: false,
	setCinemaMode: () => {},
	toggleCinemaMode: () => {}
});

interface CinemaModeProviderProps {
	children: React.ReactNode;
}

export function CinemaModeProvider(props: CinemaModeProviderProps) {
	const [searchParams] = useSearchParams();
	const cinemaModeInUrl = searchParams.get("cinemaMode") === "true"
	const [cinemaMode, setCinemaMode] = useState(() => cinemaModeInUrl || localStorage.getItem("cinema-mode") === "true")
	
	const state = useMemo<CinemaModeProviderContextProps>(() => ({
		cinemaMode: cinemaMode,
		setCinemaMode: setCinemaMode,
		toggleCinemaMode: () => setCinemaMode((prev) => !prev)
	}), [cinemaMode]);

	useEffect(() => localStorage.setItem("cinema-mode", cinemaMode ? "true" : "false"), [cinemaMode]);
	return (
		<CinemaModeContext.Provider value={state}>
			{props.children}
		</CinemaModeContext.Provider>
	);
}
