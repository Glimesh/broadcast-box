import React from "react";

interface PlayerHeaderProps {
	children: React.ReactNode;
	headerType: "Error" | "Warning" | "Success"
}

const PlayerHeader = (props: PlayerHeaderProps) => {
	return (
		<p className={`
		${props.headerType === "Error" && "bg-red-700"}
		${props.headerType === "Warning" && "bg-orange-500"}
		${props.headerType === "Success" && "bg-green-500"}
		text-white 
		text-lg 
		text-center
		p-4
		rounded-t-lg
		whitespace-pre-wrap`}>
			{props.children}
		</p>
	)
}

export default PlayerHeader;