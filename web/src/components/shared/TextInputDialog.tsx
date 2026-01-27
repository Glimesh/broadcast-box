import React, { useContext, useRef } from "react";
import { LocaleContext } from "../../providers/LocaleProvider";

interface Props<T extends string | number> {
	title: string;
	message: string;
	placeholder?: string;
	isSecret?: boolean;
	children?: React.ReactNode;
	// eslint-disable-next-line no-unused-vars
	onAccept?: (result: T) => void;
	// eslint-disable-next-line no-unused-vars
	onChange?: (result: T) => void;
	initialValue?: T;

	canCloseOnBackgroundClick?: boolean;

	buttonAcceptText?: string;
}

export default function TextInputDialog<T extends string | number>(
	props: Props<T>,
) {
	const { locale } = useContext(LocaleContext)
	const valueRef = useRef<HTMLInputElement>(null);

	return (
		<div
			className="p-6 rounded-lg shadow-lg w-1/2 bg-gray-800"
			onClick={(e) => e.stopPropagation()}
		>
			<h2 className="text-lg font-semibold">{props.title}</h2>
			<p className="mb-2">{props.message}</p>

			<input
				className="mb-6 appearance-none border w-full py-2 px-3 leading-tight focus:outline-hidden focus:shadow-outline bg-gray-700 border-gray-700 text-white rounded-sm shadow-md placeholder-gray-200"
				type={props.isSecret ? "password" : "text"}
				ref={valueRef!}
				placeholder={props.placeholder}
				autoFocus
				onKeyUp={(e) => {
					if (e.key === "Enter") {
						props.onAccept?.(valueRef.current?.value as T)
					}
				}}
			/>

			{/*Buttons*/}
			<div className="flex flex-row justify-items-stretch gap-4 ">
				{props.onAccept !== null && (
					<button
						className="bg-green-700 text-white px-4 py-2 rounded"
						onClick={() => props.onAccept?.(valueRef.current?.value as T)}
					>
						{props.buttonAcceptText !== undefined ? props.buttonAcceptText : locale.shared_component_text_input_dialog.button_accept}
					</button>
				)}
			</div>
		</div>
	);
}
