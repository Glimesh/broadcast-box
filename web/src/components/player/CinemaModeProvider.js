import React, { useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";
export const CinemaModeContext = React.createContext({
    cinemaMode: false,
    setCinemaMode: () => { },
    toggleCinemaMode: () => { }
});
export function CinemaModeProvider(props) {
    const [searchParams] = useSearchParams();
    const cinemaModeInUrl = searchParams.get("cinemaMode") === "true";
    const [cinemaMode, setCinemaMode] = useState(() => cinemaModeInUrl || localStorage.getItem("cinema-mode") === "true");
    const state = useMemo(() => ({
        cinemaMode: cinemaMode,
        setCinemaMode: setCinemaMode,
        toggleCinemaMode: () => setCinemaMode((prev) => !prev)
    }), [cinemaMode]);
    useEffect(() => localStorage.setItem("cinema-mode", cinemaMode ? "true" : "false"), [cinemaMode]);
    return (React.createElement(CinemaModeContext.Provider, { value: state }, props.children));
}
