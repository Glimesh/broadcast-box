import React, {useContext, useState} from "react";
import {CinemaModeContext} from "./CinemaModeProvider";
import Player from "./Player";
import ErrorHeader from "../error-header/errorHeader";

const PlayerPage = () => {
	const {cinemaMode, toggleCinemaMode} = useContext(CinemaModeContext);
	const [peerConnectionDisconnected, setPeerConnectionDisconnected] = useState<boolean>(false)

	return (
		<div className="mt-0">
			{peerConnectionDisconnected && <ErrorHeader> WebRTC has disconnected or failed to connect at all 😭 </ErrorHeader>}
			<div className={`flex flex-col items-center ${!cinemaMode && 'mx-auto px-2 py-2 container'}`}>
				<Player 
					cinemaMode={cinemaMode}
					peerConnectionDisconnected={peerConnectionDisconnected}
					setPeerConnectionDisconnected={setPeerConnectionDisconnected}/>
				
				<button className='bg-blue-900 px-4 py-2 rounded-lg mt-6' onClick={toggleCinemaMode}>
					{cinemaMode ? "Disable cinema mode" : "Enable cinema mode"}
				</button>
			</div>
		</div>
	)
}

export default PlayerPage
