import React, {ChangeEvent, useState} from "react";
import {ChartBarIcon} from "@heroicons/react/16/solid";

interface LayerProps {
	encodingId: string;
}

interface QualityComponentProps {
	layers: LayerProps[];
	layerEndpoint: string;
}

// TODO:
// - Create popup selector
const QualitySelectorComponent = (props: QualityComponentProps) => {
	const [isOpen, setIsOpen] = useState<boolean>(false);

	const onLayerChange = (event: ChangeEvent<HTMLSelectElement>) => {
		fetch(props.layerEndpoint, {
			method: 'POST',
			body: JSON.stringify({mediaId: '1', encodingId: event.target.value}),
			headers: {
				'Content-Type': 'application/json'
			}
		}).catch((err) => console.error(err))
	}

	return (
		<ChartBarIcon onClick={() => setIsOpen((prev) => !prev)}>
			<select
				defaultValue="disabled"
				onChange={onLayerChange}
				className="appearance-none border w-full py-2 px-3 leading-tight focus:outline-hidden focus:shadow-outline bg-gray-700 border-gray-700 text-white rounded-sm shadow-md placeholder-gray-200">
				<option value="disabled" disabled={true}>Choose Quality Level</option>
				{props.layers.map(layer => <option key={`layerEndodingId_${layer.encodingId}`}
																					 value={layer.encodingId}>{layer.encodingId}</option>)}
			</select>
		</ChartBarIcon>
	)
}

export default QualitySelectorComponent