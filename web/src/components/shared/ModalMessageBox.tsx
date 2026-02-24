import React from "react";
import Button from "./Button";
import ErrorMessagePanel from "./ErrorMessagePanel";

interface Props {
	title: string;
	errorTitle?: string;
	message: string;
	errorMessage?: string;
	children?: React.ReactNode;
	isOpen: boolean;
	onAccept?: () => void;
	onDeny?: () => void;
	onChange?: () => void;

	canCloseOnBackgroundClick?: boolean;
}

export default function ModalMessageBox(props: Props) {
	if (!props.isOpen) {
		return <></>;
	}

	return (
		<div className="flex justify-center items-center h-screen absolute z-10">
			<div
				className="fixed inset-0 bg-transparent flex items-center justify-center"
				onClick={() => props.canCloseOnBackgroundClick && props.onDeny?.()}
			>
				<div
					className="p-6 rounded-lg shadow-lg w-1/2 bg-gray-800"
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

					{/*Buttons*/}
					<div className="flex flex-row justify-items-stretch gap-4 ">
						{props.onAccept !== null && (
							<Button
								color="Accept"
								title="Accept"
								onClick={() => props.onAccept?.()}
							/>
						)}

							<Button
								title="Close"
								onClick={() => {
									props.onDeny?.();
								}}
							/>

					</div>
				</div>
			</div>
		</div>
	);
}
