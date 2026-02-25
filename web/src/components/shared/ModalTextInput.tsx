import React, { useContext, useEffect, useRef } from "react";
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
	onAccept?: (result: T) => void;
	onDeny?: () => void;
	onChange?: (result: T) => void;
	onClose?: () => void;
	initialValue?: T;

	canCloseOnBackgroundClick?: boolean;
}

export default function ModalTextInput<T extends string | number>(
	props: Props<T>,
) {
	const { locale } = useContext(LocaleContext)
	const { onClose } = props
	const valueRef = useRef<HTMLInputElement>(null);

	useEffect(() => {
		if (!props.isOpen) {
			onClose?.()
		}
	}, [onClose, props.isOpen])

	if (!props.isOpen) {
		return <></>;
	}

	return (
		<div className="fixed inset-0 z-50 flex items-center justify-center">
			<div
				className="absolute inset-0 bg-black/50"
				onClick={() => {
					if (props.canCloseOnBackgroundClick) {
						props.onDeny?.();
						props.onClose?.();
					}
				}}
			/>
			<div
				className="relative p-4 rounded-lg shadow-lg w-1/2 max-w-md bg-gray-800"
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
							props.onClose?.();
						}}
						className="bg-blue-900 hover:bg-blue-700 text-white px-4 py-2 rounded"
					>
						Close
					</button>
				</div>
			</div>
		</div>
	);
}
