import React, {useContext, useState} from "react";
import Player from "./Player";
import {useNavigate} from "react-router-dom";
import {CinemaModeContext} from "../../providers/CinemaModeProvider";
import ModalTextInput from "../shared/ModalTextInput";

const PlayerPage = () => {
  const navigate = useNavigate();
  const {cinemaMode, toggleCinemaMode} = useContext(CinemaModeContext);
  const [streamKeys, setStreamKeys] = useState<string[]>([window.location.pathname.substring(1)]);
  const [isModalOpen, setIsModelOpen] = useState<boolean>(false);

  const addStream = (streamKey: string) => {
    if (streamKeys.some((key: string) => key.toLowerCase() === streamKey.toLowerCase())) {
      return;
    }
    setStreamKeys((prev) => [...prev, streamKey]);
    setIsModelOpen((prev) => !prev);
  };

  return (
    <div>
      {isModalOpen && (
        <ModalTextInput<string>
          title="Add stream"
          message={"Insert stream key to add to multi stream"}
          isOpen={isModalOpen}
          canCloseOnBackgroundClick={false}
          onClose={() => setIsModelOpen(false)}
          onAccept={(result: string) => addStream(result)}
        />
      )}

      <div className={`flex flex-col w-full items-center ${!cinemaMode && "mx-auto px-2 py-2 container gap-2"}`}>
        <div className={`grid ${streamKeys.length !== 1 ? "grid-cols-2" : ""}  w-full gap-2`}>
          {streamKeys.map((streamKey) =>
            <Player
              key={`${streamKey}_player`}
              streamKey={streamKey}
              cinemaMode={cinemaMode}
              onCloseStream={
                streamKeys.length === 1
                  ? () => navigate('/')
                  : () => setStreamKeys((prev) => prev.filter((key) => key !== streamKey))
              }
            />
          )}
        </div>

        {/*Implement footer menu*/}
        <div className="flex flex-row p-2 gap-2">
          <button
            className="bg-blue-900 hover:bg-blue-800 px-4 py-2 rounded-lg mt-6"
            onClick={toggleCinemaMode}
          >
            {cinemaMode ? "Disable cinema mode" : "Enable cinema mode"}
          </button>

          {/*Show modal to add stream keys with*/}
          <button
            className="bg-blue-900 hover:bg-blue-800 px-4 py-2 rounded-lg mt-6"
            onClick={() => setIsModelOpen((prev) => !prev)}>
            Add Stream
          </button>
        </div>
      </div>
    </div>
  )
};

export default PlayerPage;