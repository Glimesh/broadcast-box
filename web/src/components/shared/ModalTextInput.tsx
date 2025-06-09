import React, {useRef} from "react";
import {useState} from "react";

interface Props<T extends string | number> {
	title: string;
	message: string;
	children?: React.ReactNode;
	isOpen: boolean;
	onClose?: () => void;
	onAccept?: (result: T) => void;
	onChange?: (result: T) => void;
	initialValue?: T;

	canCloseOnBackgroundClick?: boolean;
}

export default function ModalTextInput<T extends string | number>(props: Props<T>) {
	const [isOpen, setIsOpen] = useState<boolean>(props.isOpen);
	const valueRef = useRef<HTMLInputElement>(null);
	
	if(!isOpen){
		return <></>
	}

	return (
		<div className="flex justify-center items-center h-screen absolute z-100 ">
				<div
					className="fixed inset-0 bg-transparent flex items-center justify-center"
					onClick={() => props.canCloseOnBackgroundClick && setIsOpen(false)} 
				>
					<div
						className="p-6 rounded-lg shadow-lg w-1/2 bg-gray-800"
						onClick={(e) => e.stopPropagation()} 
					>
						<h2 className="text-lg font-semibold">{props.title}</h2>
						<p className="mb-2">{props.message}</p>

						<input
							className='mb-6 appearance-none border w-full py-2 px-3 leading-tight focus:outline-hidden focus:shadow-outline bg-gray-700 border-gray-700 text-white rounded-sm shadow-md placeholder-gray-200'
							type='text'
							ref={valueRef!}
							placeholder={`Insert the key you of the stream you want to add`}
							autoFocus/>

						{/*Buttons*/}
						<div className="flex flex-row justify-items-stretch gap-4 ">
							{props.onAccept !== null && (
								<button
									className="bg-green-700 text-white px-4 py-2 rounded"
									onClick={() => props.onAccept?.(valueRef.current?.value as T)}
								>
									Accept
								</button>
							)}
							<button
								onClick={() => setIsOpen(false)}
								className="bg-blue-900 hover:bg-blue-700 text-white px-4 py-2 rounded"
							>
								Close
							</button>

						</div>
					</div>
				</div>
		</div>
	);
}