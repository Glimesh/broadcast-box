import { ChangeEvent, useState } from "react";
import { ChartBarIcon } from "@heroicons/react/16/solid";

interface QualityComponentProps {
  layers: string[];
  layerEndpoint: string;
  hasPacketLoss: boolean;
  currentLayer: string;
}

const VideoLayerSelectorComponent = (props: QualityComponentProps) => {
  const videoMediaId = "1";
  const [isOpen, setIsOpen] = useState<boolean>(false);

  const onLayerChange = (event: ChangeEvent<HTMLSelectElement>) => {
    fetch(props.layerEndpoint, {
      method: 'POST',
      body: JSON.stringify({ mediaId: videoMediaId, encodingId: event.target.value }),
      headers: {
        'Content-Type': 'application/json'
      }
    }).catch((err) => console.error("VideoLayerSelectorComponent.onLayerChange", err));
    setIsOpen(false);
  };

  const currentLayer = props.currentLayer;
  const layerList = [
    currentLayer,
    ...props.layers.filter(layer => layer !== currentLayer)
  ].map(layer => <option key={`layerEncodingId_${layer}`} value={layer}>{layer}</option>);

  if (currentLayer === '' || layerList[0].props.value === '') {
    layerList[0] = <option key="disabled">Auto</option>;
  }

  return (
    <div className="relative flex h-full items-center">
      <ChartBarIcon
        className={`video-control ${props.hasPacketLoss ? "text-orange-500" : ""} ${layerList.length === 0 ? "opacity-25" : ""}`}
        onClick={() => setIsOpen((prev) => props.layers.length <= 1 ? false : !prev)}
      />

      {isOpen && (
        <select
          onChange={onLayerChange}
          value={currentLayer}
          className="video-select"
        >
          {layerList}
        </select>
      )}
    </div>
  );
};

export default VideoLayerSelectorComponent;
