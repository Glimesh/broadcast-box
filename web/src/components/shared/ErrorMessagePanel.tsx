import React from "react";

interface ErrorMessagePanelProps {
	title: string;
	message: string;
	className?: string;
}

export default function ErrorMessagePanel(props: ErrorMessagePanelProps) {
	const extraClassName = props.className ?? "";

	return (
		<div className={`mt-4 w-full max-w-md mx-auto rounded-xl border border-red-400 bg-gray-800 p-4 shadow-md ${extraClassName}`.trim()}>
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
					{props.title}
				</h2>
			</div>
			<p className="mt-2 pl-6 text-sm text-red-600">{props.message}</p>
		</div>
	);
}
