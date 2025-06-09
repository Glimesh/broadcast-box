import React, {ChangeEvent, useState} from "react";
import {ChartBarIcon} from "@heroicons/react/16/solid";

interface QualityComponentProps {
	layers: string[];
	layerEndpoint: string;
}

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
		
		setIsOpen(() => false)
	}
	
	if(props.layers.length === 0){
		return <></>
	}

	return (
		<div className="h-full flex">
			<ChartBarIcon onClick={() => setIsOpen((prev) => !prev)}/>

			{isOpen && (
				
			<select
				defaultValue="disabled"
				onChange={onLayerChange}
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
					props.layers.map(layer => 
					<option key={`layerEncodingId_${layer}`} value={layer}>{layer}</option>)
				}
			</select>
			)}
		</div>
	)
}

export default QualitySelectorComponent