import React, {ChangeEvent, useState} from "react";
import {ChartBarIcon} from "@heroicons/react/16/solid";

interface QualityComponentProps {
	layers: string[];
	layerEndpoint: string;
	hasPacketLoss: boolean;
}

const QualitySelectorComponent = (props: QualityComponentProps) => {
	const [isOpen, setIsOpen] = useState<boolean>(false);
	const [currentLayer, setCurrentLayer] = useState<string>('');

	const onLayerChange = (event: ChangeEvent<HTMLSelectElement>) => {
		fetch(props.layerEndpoint, {
			method: 'POST',
			body: JSON.stringify({mediaId: '1', encodingId: event.target.value}),
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
		layerList[0] = <option key="disabled" value="disabled">No Layer Selected</option>
	}

	return (
		<div className="h-full flex">
			<ChartBarIcon
				className={props.hasPacketLoss ? "text-orange-600" :""}
				onClick={() => setIsOpen((prev) => props.layers.length <= 1 ? false : !prev)}/>

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

export default QualitySelectorComponent