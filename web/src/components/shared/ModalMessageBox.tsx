import React, { useEffect } from "react";
import { useState } from "react";

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
	const [isOpen, setIsOpen] = useState<boolean>(props.isOpen);

	useEffect(() => {
		setIsOpen(() => props.isOpen)
	}, [props.isOpen])

	if (!isOpen) {
		return <></>;
	}

	return (
		<div className="flex justify-center items-center h-screen absolute z-10">
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

					{ /*Error message*/}
					{props.errorMessage != undefined && props.errorMessage !== "" && (
						<div className="mt-4 w-full max-w-md mx-auto rounded-xl border border-red-400 bg-gray-800 p-4 shadow-md mb-4">
							<div className="flex items-center space-x-3 ">
								<svg
									className="h-6 w-6 text-red-600"
									fill="none"
									stroke="currentColor"
									strokeWidth="2"
									viewBox="0 0 24 24"
								>
									<path
										strokeLinecap="round"
										strokeLinejoin="round"
										d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
									/>
								</svg>
								<h2 className="text-lg font-semibold text-red-700">
									{props.errorTitle ?? "Error"}
								</h2>
							</div>
							<p className="mt-2 pl-6 text-sm text-red-600">{props.errorMessage}</p>
						</div>
					)}

					{/*Buttons*/}
					<div className="flex flex-row justify-items-stretch gap-4 ">
						{props.onAccept !== null && (
							<button
								className="bg-green-700 text-white px-4 py-2 rounded"
								onClick={() => props.onAccept?.()}
							>
								Accept
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
