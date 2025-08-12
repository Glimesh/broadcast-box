import React, { ChangeEvent, useState } from "react";
import { MusicalNoteIcon } from "@heroicons/react/20/solid";

interface QualityComponentProps {
	layers: string[];
	layerEndpoint: string;
	hasPacketLoss: boolean;
}

const AudioLayerSelectorComponent = (props: QualityComponentProps) => {
	const audioMediaId = "2"
	const [isOpen, setIsOpen] = useState<boolean>(false);
	const [currentLayer, setCurrentLayer] = useState<string>('');

	const onLayerChange = (event: ChangeEvent<HTMLSelectElement>) => {
		fetch(props.layerEndpoint, {
			method: 'POST',
			body: JSON.stringify({ mediaId: audioMediaId, encodingId: event.target.value }),
			headers: {
				'Content-Type': 'application/json'
			}
		}).catch((err) => console.error("onLayerChange", err))
		setIsOpen(false)
		setCurrentLayer(event.target.value)
	}

	let layerList = [
		currentLayer,
		...props.layers.filter(layer => layer !== currentLayer)
	].map(layer => <option key={`layerEncodingId_${layer}`} value={layer}>{layer}</option>)
	if (layerList[0].props.value === '') {
		layerList[0] = <option key="disabled">Auto</option>
	}

	return (
		<div className="h-full flex">
			<MusicalNoteIcon
				className={props.hasPacketLoss ? "text-orange-600" : ""}
				onClick={() => setIsOpen((prev) => props.layers.length <= 1 ? false : !prev)} />

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

export default AudioLayerSelectorComponent
