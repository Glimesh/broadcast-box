import React, { useContext } from "react";
import { LocaleContext } from "../../providers/LocaleProvider";

interface Props {
	title?: string;
	subTitle?: string;
	placeholder?: string;
	children?: React.ReactNode;
	classNames?: string;

	onAccept?: () => void;
	onDeny?: () => void;
}

export default function Card(props: Props) {
	const { locale } = useContext(LocaleContext)

	return (
		<div
			className={`flex flex-col p-2 rounded-lg shadow-lg bg-gray-800 h-full border-1 border-gray-700 ${props.classNames}`}
			onClick={(e) => e.stopPropagation()}
		>
			{!!props.title && (
				<h2 className="text-lg font-semibold">{props.title}</h2>
			)}
			{!!props.subTitle && <p className="mb-2">{props.subTitle}</p>}

			{props.children}

			{props.onAccept !== undefined ||
				(props.onDeny !== undefined && (
					<div className="flex flex-row justify-items-stretch gap-4 ">
						{props.onAccept !== null && (
							<button
								className="bg-green-700 text-white px-4 py-2 rounded"
								onClick={() => props.onAccept?.()}
							>
								{locale.shared_component_card.button_accept}
							</button>
						)}
					</div>
				))}
		</div>
	);
}
