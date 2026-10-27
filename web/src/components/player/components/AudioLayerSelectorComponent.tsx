import { ChangeEvent, useState } from "react";
import { MusicalNoteIcon } from "@heroicons/react/20/solid";

interface QualityComponentProps {
  layers: string[];
  layerEndpoint: string;
  hasPacketLoss: boolean;
  currentLayer: string;
}

const AudioLayerSelectorComponent = (props: QualityComponentProps) => {
  const audioMediaId = "2";
  const [isOpen, setIsOpen] = useState<boolean>(false);
  const [currentLayer, setCurrentLayer] = useState<string>(props.currentLayer);

  const onLayerChange = (event: ChangeEvent<HTMLSelectElement>) => {
    fetch(props.layerEndpoint, {
      method: 'POST',
      body: JSON.stringify({ mediaId: audioMediaId, encodingId: event.target.value }),
      headers: {
        'Content-Type': 'application/json'
      }
    }).catch((err) => console.error("onLayerChange", err));
    setIsOpen(false);
    setCurrentLayer(event.target.value);
  };

  const layerList = [
    currentLayer,
    ...props.layers.filter(layer => layer !== currentLayer)
  ].map(layer => <option key={`layerEncodingId_${layer}`} value={layer}>{layer}</option>);

  if (currentLayer === '' || layerList[0].props.value === '') {
    layerList[0] = <option key="disabled">Auto</option>;
  }

  return (
    <div className="relative flex h-full items-center">
      <MusicalNoteIcon
        className={`video-control ${props.hasPacketLoss ? "text-orange-500" : ""}`}
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

export default AudioLayerSelectorComponent;
