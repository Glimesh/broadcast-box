import React, { useContext, useState } from "react";
import { CinemaModeContext } from "./CinemaModeProvider";
import Player from "./Player";
import ErrorHeader from "../error-header/errorHeader";
const PlayerPage = () => {
    const { cinemaMode, toggleCinemaMode } = useContext(CinemaModeContext);
    const [peerConnectionDisconnected, setPeerConnectionDisconnected] = useState(true);
    return (React.createElement("div", { className: "mt-0" },
        peerConnectionDisconnected && React.createElement(ErrorHeader, null, " WebRTC has disconnected or failed to connect at all \uD83D\uDE2D "),
        React.createElement("div", { className: `flex flex-col items-center ${!cinemaMode && 'mx-auto px-2 py-2 container'}` },
            React.createElement(Player, { cinemaMode: cinemaMode, peerConnectionDisconnected: peerConnectionDisconnected, setPeerConnectionDisconnected: setPeerConnectionDisconnected }),
            React.createElement("button", { className: 'bg-blue-900 px-4 py-2 rounded-lg mt-6', onClick: toggleCinemaMode }, cinemaMode ? "Disable cinema mode" : "Enable cinema mode"))));
};
export default PlayerPage;
