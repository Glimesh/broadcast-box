import React, { useContext, useEffect, useState } from "react";
import Player from "./Player";
import { useNavigate } from "react-router-dom";
import { CinemaModeContext } from "../../providers/CinemaModeProvider";
import ModalTextInput from "../shared/ModalTextInput";
import { StatusContext, StreamStatus } from "../../providers/StatusProvider";
import Button from "../shared/Button";
import AvailableStreams from "../selection/AvailableStreams";

const PlayerPage = () => {
  const navigate = useNavigate();
  const { cinemaMode, toggleCinemaMode } = useContext(CinemaModeContext);
  const { currentStreamStatus } = useContext(StatusContext);
  const [streamKeys, setStreamKeys] = useState<string[]>([
    window.location.pathname.substring(1),
  ]);
  const [isModalOpen, setIsModelOpen] = useState<boolean>(false);
  const [status, setStatus] = useState<StreamStatus>({
    motd: "",
    viewers: 0,
    streamKey: "",
    isOnline: false,
  });

  useEffect(() => {
    if (currentStreamStatus === undefined) {
      setStatus(() => ({
        motd: "",
        viewers: 0,
        streamKey: "",
        isOnline: false,
      }));
      return;
    }

    setStatus(() => currentStreamStatus);
  }, [currentStreamStatus]);

  const addStream = (streamKey: string) => {
    if (
      streamKeys.some(
        (key: string) => key.toLowerCase() === streamKey.toLowerCase(),
      )
    ) {
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
          placeholder={"Insert the key of the stream you want to add"}
          isOpen={isModalOpen}
          canCloseOnBackgroundClick={true}
          onClose={() => setIsModelOpen(false)}
          onAccept={(result: string) => addStream(result)}
        >
          <AvailableStreams
            showHeader={false}
            onClickOverride={(streamKey) => addStream(streamKey)}
          />
        </ModalTextInput>
      )}

      <div className={`flex flex-col w-full items-center ${!cinemaMode && "mx-auto px-2 py-2 container gap-2"}`} >
        <div className={`grid ${streamKeys.length !== 1 ? "grid-cols-2" : ""}  w-full gap-2`} >
          {streamKeys.map((streamKey) => (
            <Player
              key={`${streamKey}_player`}
              streamKey={streamKey}
              cinemaMode={cinemaMode}
              onCloseStream={
                streamKeys.length === 1
                  ? () => navigate("/")
                  : () =>
                    setStreamKeys((prev) =>
                      prev.filter((key) => key !== streamKey),
                    )
              }
            />
          ))}
        </div>

        {!cinemaMode && (
          <div className="w-full -mt-2 ml-8">
            <div className="relative h-5">
              <div className={`absolute inset-0 transition-opacity duration-300 text-gray-400 ${status?.isOnline ? "opacity-100" : "opacity-0"}`} >
                {status?.motd}
              </div>

              <div className={`absolute inset-0 transition-opacity duration-300 text-red-400 font-semibold ${!status?.isOnline ? "opacity-100" : "opacity-0"}`} >
                Offline
              </div>
            </div>
          </div>
        )}

        {/*Implement footer menu*/}
        <div className="flex flex-row gap-2">
          <Button
            title={cinemaMode ? "Disable cinema mode" : "Enable cinema mode"}
            onClick={toggleCinemaMode}
            iconRight="CodeBracketSquare"
          />

          {/*Show modal to add stream keys with*/}
          <Button
            title="Add Stream"
            color="Accept"
            onClick={() => setIsModelOpen((prev) => !prev)}
            iconRight="SquaresPlus"
          />
        </div>
      </div>
    </div>
  );
};

export default PlayerPage;
