import React, { useContext, useEffect, useRef } from "react";
import { useState } from "react";
import { LocaleContext } from "../../providers/LocaleProvider";
import ErrorMessagePanel from "./ErrorMessagePanel";

interface Props<T extends string | number> {
	title: string;
	errorTitle?: string;
	message: string;
	errorMessage?: string;
	placeholder?: string;
	children?: React.ReactNode;
	isOpen: boolean;
	// eslint-disable-next-line no-unused-vars
	onAccept?: (result: T) => void;
	onDeny?: () => void;
	// eslint-disable-next-line no-unused-vars
	onChange?: (result: T) => void;
	onClose?: () => void;
	initialValue?: T;

	canCloseOnBackgroundClick?: boolean;
}

export default function ModalTextInput<T extends string | number>(
	props: Props<T>,
) {
	const { locale } = useContext(LocaleContext)
	const [isOpen, setIsOpen] = useState<boolean>(props.isOpen);
	const valueRef = useRef<HTMLInputElement>(null);

	useEffect(() => {
		setIsOpen(() => props.isOpen)
	}, [props.isOpen])

	useEffect(() => {
		if (!isOpen) {
			props.onClose?.()
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [isOpen])

	if (!isOpen) {
		return <></>;
	}

	return (
		<div className="flex justify-center items-center h-screen absolute z-10">
			<div
				className="flex fixed inset-0 bg-transparent items-center justify-center"
				onClick={() => props.canCloseOnBackgroundClick && setIsOpen(false)}
			>
				<div
					className="p-4 rounded-lg shadow-lg w-1/2 bg-gray-800"
					onClick={(e) => e.stopPropagation()}
				>
					<h2 className="text-lg font-semibold">{props.title}</h2>
					<p className="mb-2">{props.message}</p>

					{ /*Error message*/}
					{props.errorMessage != undefined && props.errorMessage !== "" && (
						<ErrorMessagePanel
							className="mb-4"
							title={props.errorTitle ?? "Error"}
							message={props.errorMessage}
						/>
					)}

					{/*Input*/}
					<input
						className="mb-6 appearance-none border w-full py-2 px-3 leading-tight focus:outline-hidden focus:shadow-outline bg-gray-700 border-gray-700 text-white rounded-sm shadow-md placeholder-gray-200"
						type="text"
						ref={valueRef!}
						placeholder={props.placeholder}
						onKeyUp={(evt) => evt.key === "Enter" ? props.onAccept?.(valueRef.current?.value as T) : null}
						autoFocus
					/>

					{/* Optional children */}
					{props.children && (<div className="mb-2">
						{props.children}
					</div>)}

					{/*Buttons*/}
					<div className="flex flex-row justify-items-stretch gap-2">
						{props.onAccept !== null && (
							<button
								className="bg-green-700 text-white px-4 py-2 rounded"
								onClick={() => props.onAccept?.(valueRef.current?.value as T)}
							>
								{locale.shared_component_text_input_modal.button_accept}
							</button>
						)}
						<button
							onClick={() => {
								props.onDeny?.();
								setIsOpen(false);
							}}
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
