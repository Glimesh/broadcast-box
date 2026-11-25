import { ChangeEvent, useState } from "react";
import { ChartBarIcon } from "@heroicons/react/16/solid";

interface QualityComponentProps {
	layers: string[];
	layerEndpoint: string;
	hasPacketLoss: boolean;
	currentLayer: string;
}

const VideoLayerSelectorComponent = (props: QualityComponentProps) => {
	const videoMediaId = "1"
	const [isOpen, setIsOpen] = useState<boolean>(false);

	const onLayerChange = (event: ChangeEvent<HTMLSelectElement>) => {
		fetch(props.layerEndpoint, {
			method: 'POST',
			body: JSON.stringify({ mediaId: videoMediaId, encodingId: event.target.value }),
			headers: {
				'Content-Type': 'application/json'
			}
		}).catch((err) => console.error("VideoLayerSelectorComponent.onLayerChange", err))
		setIsOpen(false)
	}

	const currentLayer = props.currentLayer
	const layerList = [
		currentLayer,
		...props.layers.filter(layer => layer !== currentLayer)
	].map(layer => <option key={`layerEncodingId_${layer}`} value={layer}>{layer}</option>)

	if (currentLayer === '' || layerList[0].props.value === '') {
		layerList[0] = <option key="disabled">Auto</option>
	}

	const hasLayers = props.layers.length > 1;

	return (
		<div className="h-full flex">
			<ChartBarIcon
				className={`${props.hasPacketLoss ? "text-orange-600" : ""} ${hasLayers ? "cursor-pointer" : "opacity-25 cursor-default"}`}
				onClick={() => setIsOpen((prev) => hasLayers && !prev)} />

			{isOpen && (
				<select
					onChange={onLayerChange}
					value={currentLayer}
					className="
						absolute 
						right-0
						bottom-8
						w-50
						appearance-none
						border
						py-2
						px-3
						leading-tight
						focus:outline-hidden
						focus:shadow-outline
						bg-gray-700
						border-gray-700
						text-white
						rounded-sm
						shadow-md
						placeholder-gray-200">
					{
						layerList
					}
				</select>
			)}
		</div>
	)
}

export default VideoLayerSelectorComponent
